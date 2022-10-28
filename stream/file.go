package stream

import (
	"bufio"
	"fmt"
	"os"
)

// FileReader implements the map reader interface for
// reading the map from an input file
type FileReader struct {
	mapFile     *os.File
	fileScanner *bufio.Scanner
}

// NewFileReader creates a new instance of the file reader
func NewFileReader(filePath string) (InputReader, error) {
	mapFile, err := os.Open(filePath)

	if err != nil {
		return nil, fmt.Errorf("unable to open file, %w", err)
	}

	fileScanner := bufio.NewScanner(mapFile)
	fileScanner.Split(bufio.ScanLines)

	return &FileReader{
		mapFile:     mapFile,
		fileScanner: fileScanner,
	}, nil
}

func (fr *FileReader) HasMoreCities() bool {
	return fr.fileScanner.Scan()
}

func (fr *FileReader) ReadCity() string {
	return fr.fileScanner.Text()
}

func (fr *FileReader) Close() error {
	return fr.mapFile.Close()
}

type FileWriter struct {
	outputFile     *os.File
	bufferedWriter *bufio.Writer
}

func NewFileWriter(filePath string) (OutputWriter, error) {
	file, err := os.Create(filePath)
	if err != nil {
		return nil, fmt.Errorf("unable to create file, %w", err)
	}

	bw := bufio.NewWriter(file)

	return &FileWriter{
		outputFile:     file,
		bufferedWriter: bw,
	}, nil
}

func (fw *FileWriter) Write(s string) error {
	_, err := fw.bufferedWriter.WriteString(s)

	return err
}

func (fw *FileWriter) Close() error {
	return fw.outputFile.Close()
}

func (fw *FileWriter) Flush() error {
	return fw.bufferedWriter.Flush()
}
