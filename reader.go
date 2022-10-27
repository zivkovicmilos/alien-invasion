package main

import (
	"bufio"
	"fmt"
	"os"
)

// mapReader defines the base map reader interface
type mapReader interface {
	// readCity reads a single city line from the map.
	// Returns "" if no input cities are present
	readCity() string

	// close closes the map reader
	close() error
}

// fileReader implements the map reader interface for
// reading the map from an input file
type fileReader struct {
	mapFile     *os.File
	fileScanner *bufio.Scanner
}

// newFileReader creates a new instance of the file reader
func newFileReader(filePath string) (mapReader, error) {
	mapFile, err := os.Open(filePath)

	if err != nil {
		return nil, fmt.Errorf("unable to open file, %v", err)
	}

	fileScanner := bufio.NewScanner(mapFile)
	fileScanner.Split(bufio.ScanLines)

	return &fileReader{
		mapFile:     mapFile,
		fileScanner: fileScanner,
	}, nil
}

func (fr *fileReader) readCity() string {
	// Check if there are leftover lines to read
	if fr.fileScanner.Scan() {
		return fr.fileScanner.Text()
	}

	// End of file is reached
	return ""
}

func (fr *fileReader) close() error {
	return fr.mapFile.Close()
}
