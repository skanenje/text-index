package index

import (
	"encoding/binary"
	"os"
	"sort"
	"sync"

	"jamtex/internal/simhash"
)

// IndexEntry represents a single entry in the index
type IndexEntry struct {
	Hash     uint64
	Offsets  []int64
	FileName string
}

// Index represents the in-memory index structure
type Index struct {
	Entries    map[uint64]IndexEntry
	SourceFile string
	BitIndices *BitIndices // New field
}

// BitIndices tracks which hashes have specific bits set
type BitIndices struct {
	set   map[int]map[uint64]struct{} // Bits that are set (1)
	unset map[int]map[uint64]struct{} // Bits that are unset (0)
	mu    sync.RWMutex                // For concurrent access
}

// Update NewIndex to initialize BitIndices
func NewIndex(sourceFile string) *Index {
	return &Index{
		Entries:    make(map[uint64]IndexEntry),
		SourceFile: sourceFile,
		BitIndices: NewBitIndices(),
	}
}

// NewBitIndices creates a new BitIndices
func NewBitIndices() *BitIndices {
	return &BitIndices{
		set:   make(map[int]map[uint64]struct{}),
		unset: make(map[int]map[uint64]struct{}),
	}
}

// AddHash adds a hash to the bit indices
func (bi *BitIndices) AddHash(hash uint64) {
	bi.mu.Lock()
	defer bi.mu.Unlock()

	// For each bit position
	for i := 0; i < simhash.HashSize; i++ {
		bit := (hash >> uint(i)) & 1

		if bit == 1 {
			// Bit is set
			if bi.set[i] == nil {
				bi.set[i] = make(map[uint64]struct{})
			}
			bi.set[i][hash] = struct{}{}
		} else {
			// Bit is unset
			if bi.unset[i] == nil {
				bi.unset[i] = make(map[uint64]struct{})
			}
			bi.unset[i][hash] = struct{}{}
		}
	}
}

// FindCandidates finds hashes that might be within the given Hamming distance
func (bi *BitIndices) FindCandidates(hash uint64, threshold int) map[uint64]struct{} {
	bi.mu.RLock()
	defer bi.mu.RUnlock()

	candidates := make(map[uint64]struct{})

	// For each bit position
	for i := 0; i < simhash.HashSize; i++ {
		bit := (hash >> uint(i)) & 1

		// If bit is set (1), get hashes with bit unset (0)
		// If bit is unset (0), get hashes with bit set (1)
		var possibleMatches map[uint64]struct{}
		if bit == 1 {
			possibleMatches = bi.unset[i]
		} else {
			possibleMatches = bi.set[i]
		}

		// Add all possible matches to candidates
		for h := range possibleMatches {
			candidates[h] = struct{}{}
		}
	}

	return candidates
}

// AddEntry adds or updates an entry in the index
func (idx *Index) AddEntry(hash uint64, offset int64) {
	entry, exists := idx.Entries[hash]

	if !exists {
		entry = IndexEntry{
			Hash:     hash,
			Offsets:  []int64{offset},
			FileName: idx.SourceFile,
		}
	} else {
		entry.Offsets = append(entry.Offsets, offset)
	}

	idx.Entries[hash] = entry

	// Update bit indices
	idx.BitIndices.AddHash(hash)
}

// FindExact returns the index entry for an exact hash match
func (idx *Index) FindExact(hash uint64) (IndexEntry, bool) {
	entry, exists := idx.Entries[hash]
	return entry, exists
}

// FindSimilar returns entries with hashes similar to the given hash
// within the specified Hamming distance threshold
func (idx *Index) FindSimilar(hash uint64, threshold int) []IndexEntry {
	// Get candidate hashes that might be within the threshold
	candidates := idx.BitIndices.FindCandidates(hash, threshold)

	var results []IndexEntry

	// Check each candidate's actual Hamming distance
	for candidateHash := range candidates {
		distance := simhash.HammingDistance(hash, candidateHash)
		if distance <= threshold {
			entry, exists := idx.Entries[candidateHash]
			if exists {
				results = append(results, entry)
			}
		}
	}

	// Sort results by similarity (lowest distance first)
	sort.Slice(results, func(i, j int) bool {
		distI := simhash.HammingDistance(hash, results[i].Hash)
		distJ := simhash.HammingDistance(hash, results[j].Hash)
		return distI < distJ
	})

	return results
}

// SaveToFile saves the index to a file
func (idx *Index) SaveToFile(filePath string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write source file name
	fileNameBytes := []byte(idx.SourceFile)
	binary.Write(file, binary.LittleEndian, uint32(len(fileNameBytes)))
	file.Write(fileNameBytes)

	// Write number of entries
	binary.Write(file, binary.LittleEndian, uint32(len(idx.Entries)))

	// Write each entry
	for _, entry := range idx.Entries {
		binary.Write(file, binary.LittleEndian, entry.Hash)
		binary.Write(file, binary.LittleEndian, uint32(len(entry.Offsets)))
		for _, offset := range entry.Offsets {
			binary.Write(file, binary.LittleEndian, offset)
		}
	}

	return nil
}

// LoadFromFile loads an index from a file
func LoadFromFile(filePath string) (*Index, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Read source file name
	var fileNameLen uint32
	binary.Read(file, binary.LittleEndian, &fileNameLen)
	fileNameBytes := make([]byte, fileNameLen)
	file.Read(fileNameBytes)
	sourceFile := string(fileNameBytes)

	idx := NewIndex(sourceFile)

	// Read number of entries
	var numEntries uint32
	binary.Read(file, binary.LittleEndian, &numEntries)

	// Read each entry
	for i := uint32(0); i < numEntries; i++ {
		var hash uint64
		binary.Read(file, binary.LittleEndian, &hash)

		var numOffsets uint32
		binary.Read(file, binary.LittleEndian, &numOffsets)

		offsets := make([]int64, numOffsets)
		for j := uint32(0); j < numOffsets; j++ {
			binary.Read(file, binary.LittleEndian, &offsets[j])
		}

		idx.Entries[hash] = IndexEntry{
			Hash:     hash,
			Offsets:  offsets,
			FileName: sourceFile,
		}
	}
	for hash := range idx.Entries {
		idx.BitIndices.AddHash(hash)
	}

	return idx, nil
}
