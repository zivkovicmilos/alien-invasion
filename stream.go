//nolint:unused
package main

import (
	"bufio"
	"fmt"
	"os"
)

// inputReader defines the base map reader interface
type inputReader interface {
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
func newFileReader(filePath string) (inputReader, error) {
	mapFile, err := os.Open(filePath)

	if err != nil {
		return nil, fmt.Errorf("unable to open file, %w", err)
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

type outputWriter interface {
	write(string) error
	flush() error
	close() error
}

type fileWriter struct {
	outputFile     *os.File
	bufferedWriter *bufio.Writer
}

func newFileWriter(filePath string) (outputWriter, error) {
	file, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("unable to create file, %w", err)
	}

	bw := bufio.NewWriter(file)

	return &fileWriter{
		outputFile:     file,
		bufferedWriter: bw,
	}, nil
}

func (fw *fileWriter) write(s string) error {
	_, err := fw.bufferedWriter.WriteString(s)

	return err
}

func (fw *fileWriter) close() error {
	return fw.outputFile.Close()
}

func (fw *fileWriter) flush() error {
	return fw.bufferedWriter.Flush()
}
