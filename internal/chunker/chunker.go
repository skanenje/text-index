package chunker

import (
    "bufio"
    "io"
    "os"
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

    var chunks []Chunk
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