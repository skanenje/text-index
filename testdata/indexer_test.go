package test

import (
    "io/ioutil"
    "os"
    "path/filepath"
    "testing"
    
    "jamtex/internal/chunker"
    "jamtex/internal/index"
    "jamtex/internal/simhash"
)

func TestIndexing(t *testing.T) {
    // Create a temporary file for testing
    content := "This is a test file with some content.\n" +
               "It has multiple lines and can be used for testing the chunker.\n" +
               "The SimHash algorithm should handle this text properly.\n" +
               "We want to ensure that our indexing system works correctly."
    
    tmpdir, err := ioutil.TempDir("", "textindex-test")
    if err != nil {
        t.Fatalf("Failed to create temp dir: %v", err)
    }
    defer os.RemoveAll(tmpdir)
    
    filePath := filepath.Join(tmpdir, "test.txt")
    err = ioutil.WriteFile(filePath, []byte(content), 0644)
    if err != nil {
        t.Fatalf("Failed to write test file: %v", err)
    }
    
    // Test chunking
    c := chunker.NewChunker(32) // Small chunk size for testing
    chunks, err := c.ProcessFile(filePath)
    if err != nil {
        t.Fatalf("Failed to process file: %v", err)
    }
    
    if len(chunks) == 0 {
        t.Fatal("Expected at least one chunk, got none")
    }
    
    // Test indexing
    idx := index.NewIndex(filePath)
    for _, chunk := range chunks {
        hash := simhash.Hash(chunk.Content)
        idx.AddEntry(hash, chunk.Position)
    }
    
    // Test index serialization
    indexPath := filepath.Join(tmpdir, "test.idx")
    err = idx.SaveToFile(indexPath)
    if err != nil {
        t.Fatalf("Failed to save index: %v", err)
    }
    
    // Test index loading
    loadedIdx, err := index.LoadFromFile(indexPath)
    if err != nil {
        t.Fatalf("Failed to load index: %v", err)
    }
    
    // Verify index data
    if len(loadedIdx.Entries) != len(idx.Entries) {
        t.Errorf("Loaded index has %d entries, expected %d", len(loadedIdx.Entries), len(idx.Entries))
    }
    
    // Test lookup
    for hash := range idx.Entries {
        entry, exists := loadedIdx.FindExact(hash)
        if !exists {
            t.Errorf("Hash %016x not found in loaded index", hash)
            continue
        }
        
        originalEntry := idx.Entries[hash]
        if len(entry.Offsets) != len(originalEntry.Offsets) {
            t.Errorf("Hash %016x has %d offsets, expected %d", 
                hash, len(entry.Offsets), len(originalEntry.Offsets))
        }
    }
}

func TestSimHash(t *testing.T) {
    // Test similar strings have similar hashes
    text1 := []byte("This is a sample text for testing SimHash algorithm")
    text2 := []byte("This is a simple text for testing SimHash algorithm")
    
    hash1 := simhash.Hash(text1)
    hash2 := simhash.Hash(text2)
    
    distance := simhash.HammingDistance(hash1, hash2)
    
    // Similar strings should have a small Hamming distance
    if distance > 10 {
        t.Errorf("Hamming distance between similar strings is too large: %d", distance)
    }
    
    // Test different strings have different hashes
    text3 := []byte("This text is completely different from the others")
    hash3 := simhash.Hash(text3)
    
    distance13 := simhash.HammingDistance(hash1, hash3)
    
    // Different strings should have a larger Hamming distance
    if distance13 <= distance {
        t.Errorf("Hamming distance between different strings is smaller than similar strings")
    }
}