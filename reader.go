package main

import (
	"bufio"
	"fmt"
	"os"
)

// mapReader defines the base map reader interface
type mapReader interface {
	// hasMoreCities returns a status indicating if there are more cities
	// to parse
	hasMoreCities() bool

	// readCity reads a single city line from the map
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

func (fr *fileReader) hasMoreCities() bool {
	return fr.fileScanner.Scan()
}

func (fr *fileReader) readCity() string {
	return fr.fileScanner.Text()
}

func (fr *fileReader) close() error {
	return fr.mapFile.Close()
}
