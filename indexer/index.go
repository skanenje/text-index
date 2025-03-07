package indexer

import "sync"

//Stores the mapping between SimHash values and chunk information
type Index struct {
	mu      sync.RWMutex
	chunks  map[uint64][]ChunkInfo
	size    int
}

// Stores metadata about a chunk
type ChunkInfo struct {
	Offset  int64
	Size    int
}

func NewIndex() *Index {
	return &Index{
		chunks: make(map[uint64][]ChunkInfo),
	}
}

// Adds a chunk to the index
func (idx *Index) Add(chunk *Chunk) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	info := ChunkInfo{
		Offset: chunk.Offset,
		Size: chunk.Size,
	}

	idx.chunks[chunk.SimHash] = append(idx.chunks[chunk.SimHash], info)
	idx.size++
}

func (idx *Index) FindSimilar(simHash uint64, maxDistance int) []ChunkInfo {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	var results []ChunkInfo
	for hash, chunks := range idx.chunks {
		if HammingDistance(hash, simHash) <= maxDistance {
			results = append(results, chunks...)
		}
	}
	return results
}

// Returns the total number of indexed chunks
func (idx *Index) Size() int {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	return idx.size
}

// Returns appropriate memory usage
func (idx *Index) MemoryUsage() int64 {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	var usage int64
	usage += int64(len(idx.chunks) * 8) // Map overhead
	for _, chunks := range idx.chunks {
		usage += int64(len(chunks) * 16)
	}
	return usage
}