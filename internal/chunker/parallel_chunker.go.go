// internal/chunker/parallel_chunker.go
package chunker

import (
    "bufio"
    "io"
    "os"
    "sync"
    
    "jamtex/internal/simhash"
)

// ParallelChunker processes files in parallel
type ParallelChunker struct {
    ChunkSize  int
    NumWorkers int
}

// NewParallelChunker creates a new parallel chunker
func NewParallelChunker(chunkSize, numWorkers int) *ParallelChunker {
    if chunkSize <= 0 {
        chunkSize = 4096 // Default to 4KB
    }
    
    if numWorkers <= 0 {
        numWorkers = 4 // Default to 4 workers
    }
    
    return &ParallelChunker{
        ChunkSize:  chunkSize,
        NumWorkers: numWorkers,
    }
}

// ProcessFileWithHash processes a file and returns chunks with their SimHash values
func (pc *ParallelChunker) ProcessFileWithHash(filePath string) (map[uint64][]int64, error) {
    // Open the file
    file, err := os.Open(filePath)
    if err != nil {
        return nil, err
    }
    defer file.Close()
    
    // Set up channels for worker communication
    chunkCh := make(chan Chunk, pc.NumWorkers*2)
    resultCh := make(chan struct {
        Hash    uint64
        Offset  int64
    }, pc.NumWorkers*2)
    errCh := make(chan error, pc.NumWorkers)
    doneCh := make(chan struct{})
    
    // Start worker goroutines
    var wg sync.WaitGroup
    for i := 0; i < pc.NumWorkers; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            for chunk := range chunkCh {
                hash := simhash.Hash(chunk.Content)
                resultCh <- struct {
                    Hash    uint64
                    Offset  int64
                }{hash, chunk.Position}
            }
        }()
    }
    
    // Start a goroutine to close the result channel when all workers are done
    go func() {
        wg.Wait()
        close(resultCh)
        close(doneCh)
    }()
    
    // Start a goroutine to read the file
    go func() {
        reader := bufio.NewReader(file)
        position := int64(0)
        
        for {
            buffer := make([]byte, pc.ChunkSize)
            bytesRead, err := reader.Read(buffer)
            
            if err != nil && err != io.EOF {
                errCh <- err
                break
            }
            
            if bytesRead > 0 {
                chunkCh <- Chunk{
                    Content:  buffer[:bytesRead],
                    Position: position,
                    Size:     bytesRead,
                }
                position += int64(bytesRead)
            }
            
            if err == io.EOF {
                break
            }
        }
        
        close(chunkCh)
    }()
    
    // Collect results
    results := make(map[uint64][]int64)
    
    // Process results from workers
    for {
        select {
        case err := <-errCh:
            return nil, err
        case result, ok := <-resultCh:
            if !ok {
                continue
            }
            offsets, exists := results[result.Hash]
            if !exists {
                offsets = []int64{}
            }
            results[result.Hash] = append(offsets, result.Offset)
        case <-doneCh:
            return results, nil
        }
    }
}