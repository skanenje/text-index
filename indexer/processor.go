package indexer

import (
	"io"
	"os"
	"sync"
)

// Processor handles the concurrent processing of chunks
type Processor struct {
	chunkSize    int
	workers      int
}

// NewProcessor creates a new Processor
func NewProcessor(chunkSize, workers int) *Processor {
	return &Processor{
		chunkSize: chunkSize,
		workers: workers,
	}
}

//ProcessFile processes a file and returns an index
func (p *Processor) ProcessFile(filePath string) (*Index, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := NewChunkReader(file, p.chunkSize)
	index := NewIndex()

	// Create worker pool
	var wg sync.WaitGroup
	chunkChan := make(chan *Chunk, p.workers)
	errorChan := make(chan error, 1)

	// Start workers
	for i := 0; i < p.workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for chunk := range chunkChan {
				chunk.SimHash = chunk.ComputeSimHash()
				index.Add(chunk)
			}
		}()
	}

	// Read chunks and send to workers
	go func() {
		for {
			chunk, err := reader.NextChunk()
			if err != nil {
				close(chunkChan)
				if err != io.EOF {
					errorChan <- err
				}
				return
			}
			chunkChan <- chunk
		}
	}()

	// Wait for all workers to finish
	wg.Wait()

	// Check for errors
	select {
	case err := <-errorChan:
		return nil, err
	default:
		return index, nil
	}
}