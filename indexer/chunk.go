package indexer

import (
	"bufio"
	"io"
)

type Chunk struct {
	Content    []byte
	Offset     int64
	Size       int
	SimHash    uint64
}

// Handles the chunking of input files
type ChunkReader struct {
	reader      *bufio.Reader
	chunkSize   int
	offset      int64
}

// Create a new chunk reader
func NewChunkReader(r io.Reader, chunkSize int) *ChunkReader {
	return &ChunkReader{
		reader: bufio.NewReader(r),
		chunkSize: chunkSize,
		offset: 0,
	}
}

// Reads the next chunk from the input
func (cr *ChunkReader) NextChunk() (*Chunk, error) {
	buf := make([]byte, cr.chunkSize)
	n, err := cr.reader.Read(buf)

	if err != nil && err != io.EOF {
		return nil, err
	}

	if n == 0 {
		return nil, io.EOF
	}

	chunk := &Chunk{
		Content: buf[:n],
		Offset: cr.offset,
		Size: n,
	}

	cr.offset += int64(n)
	return chunk, nil
}