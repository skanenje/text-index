package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"jamtex/internal/chunker"
	"jamtex/internal/index"
	"jamtex/internal/simhash"
)

// Arguments stores command line arguments
type Arguments struct {
	Command    string
	InputFile  string
	ChunkSize  int
	OutputFile string
	HashValue  uint64
}

// ParseArgs parses command line arguments
func ParseArgs() (Arguments, error) {
	args := Arguments{
		ChunkSize: 4096, // Default chunk size
	}

	for i := 1; i < len(os.Args); i++ {
		switch os.Args[i] {
		case "-c":
			if i+1 < len(os.Args) {
				args.Command = os.Args[i+1]
				i++
			} else {
				return args, errors.New("missing command value")
			}
		case "-i":
			if i+1 < len(os.Args) {
				args.InputFile = os.Args[i+1]
				i++
			} else {
				return args, errors.New("missing input file")
			}
		case "-s":
			if i+1 < len(os.Args) {
				size, err := strconv.Atoi(os.Args[i+1])
				if err != nil {
					return args, errors.New("invalid chunk size")
				}
				args.ChunkSize = size
				i++
			} else {
				return args, errors.New("missing chunk size")
			}
		case "-o":
			if i+1 < len(os.Args) {
				args.OutputFile = os.Args[i+1]
				i++
			} else {
				return args, errors.New("missing output file")
			}
		case "-h":
			if i+1 < len(os.Args) {
				hash, err := strconv.ParseUint(os.Args[i+1], 10, 64)
				if err != nil {
					return args, errors.New("invalid hash value")
				}
				args.HashValue = hash
				i++
			} else {
				return args, errors.New("missing hash value")
			}
		}
	}

	// Validate required arguments
	if args.Command != "index" && args.Command != "lookup" {
		return args, errors.New("unknown command, must be 'index' or 'lookup'")
	}

	if args.InputFile == "" {
		return args, errors.New("input file is required")
	}

	if args.Command == "index" && args.OutputFile == "" {
		return args, errors.New("output file is required for index command")
	}

	if args.Command == "lookup" && args.HashValue == 0 {
		return args, errors.New("hash value is required for lookup command")
	}

	return args, nil
}

// RunIndexCommand handles the index command
func RunIndexCommand(args Arguments) error {
	// Create a chunker
	c := chunker.NewChunker(args.ChunkSize)

	// Process the file
	chunks, err := c.ProcessFile(args.InputFile)
	if err != nil {
		return fmt.Errorf("failed to process file: %v", err)
	}
	// Create a slice to store hash log entries
	var hashLog []index.HashLogEntry
	// Create a new index
	idx := index.NewIndex(args.InputFile)

	// Add each chunk to the index
	for _, chunk := range chunks {
		hash := simhash.Hash(chunk.Content)
		idx.AddEntry(hash, chunk.Position)
		// Add to hash log
		hashLog = append(hashLog, index.HashLogEntry{
			Hash:     hash,
			Offset:   chunk.Position,
			FileName: args.InputFile,
		})
	}

	// Save the index to file
	err = idx.SaveToFile(args.OutputFile)
	if err != nil {
		return fmt.Errorf("failed to save index: %v", err)
	}
	// Save the hash log to a file (same name as output file with .hashlog extension)
	hashLogFile := args.OutputFile + ".hashlog"
	err = index.SaveHashLog(hashLog, hashLogFile)
	if err != nil {
		return fmt.Errorf("failed to save hash log: %v", err)
	}
	fmt.Printf("Indexed %d chunks from %s, saved to %s\n", len(chunks), args.InputFile, args.OutputFile)
	fmt.Printf("Hash log saved to %s\n", hashLogFile)
	fmt.Printf("To test lookup, use a hash from %s with: %s -c lookup -i %s -h <hash>\n",
		hashLogFile, filepath.Base(os.Args[0]), args.OutputFile)
	return nil
}

// RunLookupCommand handles the lookup command
func RunLookupCommand(args Arguments) error {
	// Load the index
	idx, err := index.LoadFromFile(args.InputFile)
	if err != nil {
		return fmt.Errorf("failed to load index: %v", err)
	}

	// Find the entry
	entry, exists := idx.FindExact(args.HashValue)
	if !exists {
		// Try finding similar entries if exact match not found
		similar := idx.FindSimilar(args.HashValue, 3) // Threshold of 3 bits
		if len(similar) == 0 {
			return errors.New("hash not found in index")
		}

		fmt.Println("Exact hash not found. Closest matches:")
		for i, sim := range similar {
			fmt.Printf("%d. Hash: %016x, Distance: %d, Offsets: %v\n",
				i+1,
				sim.Hash,
				simhash.HammingDistance(args.HashValue, sim.Hash),
				sim.Offsets)
		}
		return nil
	}

	// Display results
	fmt.Println("Hash found:")
	fmt.Printf("Original source file: %s\n", entry.FileName)
	fmt.Printf("Positions in source file: %v\n", entry.Offsets)

	// Try to read associated text if possible
	if len(entry.Offsets) > 0 {
		file, err := os.Open(entry.FileName)
		if err == nil {
			defer file.Close()

			buffer := make([]byte, args.ChunkSize)
			_, err = file.ReadAt(buffer, entry.Offsets[0])
			if err == nil {
				fmt.Printf("Associated text: %s\n", string(buffer))
			}
		}
	}

	return nil
}
