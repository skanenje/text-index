// cmd/textindex/main.go
package main

import (
    "fmt"
    "os"
    "runtime"
    
    "jamtex/internal/cli"
)

func main() {
    // Set GOMAXPROCS to use all available CPU cores
    runtime.GOMAXPROCS(runtime.NumCPU())
    
    // Parse command-line arguments
    args, err := cli.ParseArgs()
    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        printUsage()
        os.Exit(1)
    }
    
    // Execute the appropriate command
    var cmdErr error
    switch args.Command {
    case "index":
        cmdErr = cli.RunIndexCommand(args)
    case "lookup":
        cmdErr = cli.RunLookupCommand(args)
    default:
        cmdErr = fmt.Errorf("unknown command: %s", args.Command)
    }
    
    if cmdErr != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", cmdErr)
        os.Exit(1)
    }
}

func printUsage() {
    fmt.Println("Usage:")
    fmt.Println("  Indexing:  textindex -c index -i <input_file.txt> -s <chunk_size> -o <index_file.idx>")
    fmt.Println("  Lookup:    textindex -c lookup -i <index_file.idx> -h <simhash_value>")
    fmt.Println("")
    fmt.Println("Arguments:")
    fmt.Println("  -c <command>    Command to execute (index or lookup)")
    fmt.Println("  -i <file>       Input file (text file for index, index file for lookup)")
    fmt.Println("  -s <size>       Chunk size in bytes (default: 4096)")
    fmt.Println("  -o <file>       Output index file")
    fmt.Println("  -h <hash>       SimHash value for lookup")
}