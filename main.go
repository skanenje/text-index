package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"textindexer/indexer"
	"time"
)

func main() {
	// Command line flag
	filePath := flag.String("file", "", "Path to the text file to index")
	chunkSize := flag.Int("chunk-size", 4096, "Size of each chunk in bytes")
	workers := flag.Int("workers", runtime.NumCPU(), "Number of worker threads")
	flag.Parse()

	if *filePath == "" {
		fmt.Fprintf(os.Stderr, "Error: Please provide a file path using -file flag\n")
		os.Exit(1)
	}

	// Create new processor
	processor := indexer.NewProcessor(*chunkSize, *workers)

	// Start timing
	start := time.Now()

	// Process the file
	idx, err := processor.ProcessFile(*filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error processing file: %v\n", err)
		os.Exit(1)
	}

	// Print statistics
	elapsed := time.Since(start)
	fmt.Printf("Indexing completed in %v\n", elapsed)
	fmt.Printf("Total chunks processed: %d\n", idx.Size())
	fmt.Printf("Memory usage: %d\n", idx.MemoryUsage())

	// Keep the program running for demo purposes
	fmt.Println("Press Ctrl+C to exit")
	for {
		time.Sleep(time.Second)
	}
}