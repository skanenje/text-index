package indexer

import (
	"bufio"
	"fmt"
	"os"
)

func Indexer(filename string) {
	file, err := os.Open("input_file")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error opening file: %v\n", err)
		return
	}
	defer file.Close()

	
}
