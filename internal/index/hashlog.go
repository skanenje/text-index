// internal/index/hashlog.go
package index

import (
    "encoding/json"
    "os"
)

// HashLogEntry represents a single hash entry in the log
type HashLogEntry struct {
    Hash     uint64 `json:"hash"`
    Offset   int64  `json:"offset"`
    FileName string `json:"filename"`
}

// SaveHashLog saves the generated hashes to a JSON file
func SaveHashLog(entries []HashLogEntry, logFile string) error {
    file, err := os.Create(logFile)
    if err != nil {
        return err
    }
    defer file.Close()

    encoder := json.NewEncoder(file)
    encoder.SetIndent("", "  ")
    return encoder.Encode(entries)
}