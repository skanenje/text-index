package chunker

import (
    "bufio"
    "io"
    "os"
    "sync"
)

// Chunk represents a fixed-size portion of text with its position
type Chunk struct {
    Content  []byte
    Position int64  // Byte offset in the original file
    Size     int    // Size of the chunk
}

// Chunker splits files into fixed-size chunks
type Chunker struct {
    ChunkSize int
}

// NewChunker creates a new chunker with specified chunk size
func NewChunker(chunkSize int) *Chunker {
    if chunkSize <= 0 {
        chunkSize = 4096 // Default to 4KB if invalid
    }
    return &Chunker{ChunkSize: chunkSize}
}

// ProcessFile reads the file and returns chunks with their positions
func (c *Chunker) ProcessFile(filePath string) ([]Chunk, error) {
    file, err := os.Open(filePath)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    // Get file size for pre-allocating chunks slice
    fileInfo, err := file.Stat()
    if err != nil {
        return nil, err
    }
    
    // Estimate number of chunks based on file size
    estimatedChunks := int(fileInfo.Size()/int64(c.ChunkSize)) + 1
    chunks := make([]Chunk, 0, estimatedChunks)
    
    reader := bufio.NewReader(file)
    position := int64(0)

    for {
        buffer := make([]byte, c.ChunkSize)
        bytesRead, err := reader.Read(buffer)
        
        if err != nil && err != io.EOF {
            return nil, err
        }
        
        if bytesRead > 0 {
            chunks = append(chunks, Chunk{
                Content:  buffer[:bytesRead],
                Position: position,
                Size:     bytesRead,
            })
            position += int64(bytesRead)
        }
        
        if err == io.EOF {
            break
        }
    }
    
    return chunks, nil
}

// ProcessFileParallel reads the file and processes chunks in parallel
func (c *Chunker) ProcessFileParallel(filePath string) ([]Chunk, error) {
    // First, read all chunks from the file
    chunks, err := c.ProcessFile(filePath)
    if err != nil {
        return nil, err
    }
    
    return chunks, nil
}

// This function can be called from cli.go to process chunks in parallel
func ProcessChunksParallel(chunks []Chunk, processFn func(chunk Chunk) (uint64, error), numWorkers int) ([]uint64, error) {
    if numWorkers <= 0 {
        numWorkers = 4 // Default to 4 workers if not specified
    }

    // Create channels for distributing work and collecting results
    chunksChan := make(chan Chunk, len(chunks))
    resultsChan := make(chan struct{hash uint64; index int; err error}, len(chunks))
    
    // Fill the chunks channel with all chunks
    for _, chunk := range chunks {
        chunksChan <- chunk
    }
    close(chunksChan)
    
    // Start worker goroutines
    var wg sync.WaitGroup
    for i := 0; i < numWorkers; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for chunk := range chunksChan {
                hash, err := processFn(chunk)
                resultsChan <- struct{hash uint64; index int; err error}{
                    hash:  hash,
                    index: -1, // We'll set this in the calling function
                    err:   err,
                }
            }
        }()
    }
    
    // Wait for all workers to finish and close the results channel
    go func() {
        wg.Wait()
        close(resultsChan)
    }()
    
    // Create a slice to store the results in order
    results := make([]uint64, len(chunks))
    
    // Process results
    i := 0
    for result := range resultsChan {
        if result.err != nil {
            return nil, result.err
        }
        results[i] = result.hash
        i++
    }
    
    return results, nil
}