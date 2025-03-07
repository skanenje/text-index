package indexer

import (
	"bytes"
	"testing"
)

func TestChunkReader(t *testing.T) {
	data := []byte("This is a test content for chunking")
	reader := NewChunkReader(bytes.NewReader(data), 10)

	chunk, err := reader.NextChunk()
	if err != nil {
		t.Fatalf("Failed to read chunk: %v\n", err)
	}

	if chunk.Size != 10 {
		t.Errorf("Expected chunk size 10, got %d", chunk.Size)
	}
}

func TestSimHash(t *testing.T) {
	chunk := &Chunk{
		Content: []byte("This is a test content"),
	}

	hash := chunk.ComputeSimHash()
	if hash == 0 {
		t.Error("SimHash should not be 0")
	}
}

func TestIndex(t *testing.T) {
	index := NewIndex()
	chunk := &Chunk {
		Content:   []byte("Test content"),
		Offset:    0,
		Size:      11,
		SimHash:   123456,
	}

	index.Add(chunk)
	if index.Size() != 1 {
		t.Errorf("Expected index size 1, got %d", index.Size())
	}

	similar := index.FindSimilar(123456, 0)
	if len(similar) != 1 {
		t.Errorf("Expected 1 similar chunk, got %d", len(similar))
	}
}

func BenchmarkSimHash(b *testing.B) {
	chunk := &Chunk {
		Content:  []byte("This is a test content for benchmarking SimHash performance"),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		chunk.ComputeSimHash()
	}
}

func BenchmarkIndexAdd(b *testing.B) {
	index := NewIndex()
	chunk := &Chunk {
		Content:  []byte("Test content"),
		Offset:   0,
		Size:     11,
		SimHash:  123456,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		index.Add(chunk)
	}
}