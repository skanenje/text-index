package simhash

import (
    "hash/fnv"
    "strings"
)

// We'll use a 64-bit SimHash implementation
const (
    HashSize = 64
)

// Hash calculates the SimHash fingerprint for a chunk of text
func Hash(content []byte) uint64 {
    // Convert to string for tokenization
    text := string(content)
    
    // Initialize vector for feature hashing
    vector := make([]int, HashSize)
    
    // Simple tokenization by splitting on whitespace (can be improved)
    tokens := strings.Fields(text)
    
    // Hash each token and update the feature vector
    for _, token := range tokens {
        h := hashToken(token)
        for i := 0; i < HashSize; i++ {
            bit := (h >> uint(i)) & 1
            if bit == 1 {
                vector[i]++
            } else {
                vector[i]--
            }
        }
    }
    
    // Create the fingerprint from the feature vector
    var fingerprint uint64
    for i := 0; i < HashSize; i++ {
        if vector[i] > 0 {
            fingerprint |= 1 << uint(i)
        }
    }
    
    return fingerprint
}

// Calculate Hamming distance between two SimHash values
func HammingDistance(hash1, hash2 uint64) int {
    xor := hash1 ^ hash2
    distance := 0
    
    // Count the number of set bits
    for xor != 0 {
        distance++
        xor &= xor - 1
    }
    
    return distance
}

// Helper function to hash a token
func hashToken(token string) uint64 {
    h := fnv.New64a()
    h.Write([]byte(token))
    return h.Sum64()
}