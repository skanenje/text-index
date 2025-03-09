package simhash

import (
    "hash/fnv"
    "strings"
)

const (
    HashSize = 64
)

// Config defines options for the SimHash calculation
type Config struct {
    NGramSize int  // Size of n-grams (1 = unigrams, 2 = bigrams, etc.)
    UseWords  bool // True for word-based, false for character-based tokenization
}

// Hash calculates the SimHash fingerprint for a chunk of text with optional configuration
func Hash(content []byte, config ...Config) uint64 {
    // Default config
    cfg := Config{NGramSize: 1, UseWords: true}
    if len(config) > 0 {
        cfg = config[0]
    }

    // Ensure NGramSize is at least 1
    if cfg.NGramSize < 1 {
        cfg.NGramSize = 1
    }

    text := string(content)
    vector := make([]int, HashSize)
    var tokens []string

    if cfg.UseWords {
        // Word-based tokenization
        words := strings.Fields(text) // Splits on whitespace
        if len(words) == 0 {
            return 0 // Return 0 hash for empty input
        }
        tokens = generateNGrams(words, cfg.NGramSize)
    } else {
        // Character-based tokenization
        chars := strings.Split(text, "") // Split into individual characters
        if len(chars) == 0 {
            return 0 // Return 0 hash for empty input
        }
        tokens = generateNGrams(chars, cfg.NGramSize)
    }

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

// generateNGrams creates n-grams from a slice of tokens
func generateNGrams(tokens []string, n int) []string {
    if n <= 0 {
        n = 1 // Minimum n-gram size
    }
    if len(tokens) == 0 {
        return nil
    }
    if n > len(tokens) {
        n = len(tokens) // Cap n at the number of tokens
    }

    var ngrams []string
    for i := 0; i <= len(tokens)-n; i++ {
        ngrams = append(ngrams, strings.Join(tokens[i:i+n], " "))
    }
    return ngrams
}

// HammingDistance calculates the Hamming distance between two SimHash values
func HammingDistance(hash1, hash2 uint64) int {
    xor := hash1 ^ hash2
    distance := 0
    for xor != 0 {
        distance++
        xor &= xor - 1
    }
    return distance
}

// hashToken hashes a single token
func hashToken(token string) uint64 {
    h := fnv.New64a()
    h.Write([]byte(token))
    return h.Sum64()
}