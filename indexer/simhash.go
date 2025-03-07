package indexer

import (
	"hash/fnv"
	"strings"
)

const numFeatures = 64

// Generates a 64-bit hash for similarity comparison
func (c *Chunk) ComputeSimHash() uint64 {
	// Initialize feature vector
	features := make([]int, numFeatures)

	// Split content into words
	words := strings.Fields(string(c.Content))

	// Process each word
	for _, word := range words {
		hash := computeWordHash(word)

		// Update feature vector based on hash bits
		for i := 0; i < numFeatures; i++ {
			if (hash & (1 << uint(i))) != 0 {
				features[i]++
			} else {
				features[i]--
			}
		}
	}

	// Generate final SimHash
	var simHash uint64
	for i := 0; i < numFeatures; i++ {
		if features[i] > 0 {
			simHash |= (1 << uint(i))
		}
	}

	return simHash
}

// generates a hash for a single word
func computeWordHash(word string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(word))
	return h.Sum64()
}

// Calculates the number of differing bits between two hashes
func HammingDistance(hash1, hash2 uint64) int {
	xor := hash1 ^ hash2
	distance := 0
	for xor != 0 {
		distance += int(xor & 1)
		xor >>= 1
	}
	return distance
}