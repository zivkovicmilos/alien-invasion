package stream

// InputReader defines the base map reader interface
type InputReader interface {
	// HasMoreCities returns a status indicating if there are more cities
	// to parse
	HasMoreCities() bool

	// ReadCity reads a single city line from the map
	ReadCity() string

	// Close closes the map reader
	Close() error
}

// OutputWriter defines the base map writer interface
type OutputWriter interface {
	// Write writes a single output line to the output stream
	Write(string) error

	// Flush flushes the output lines to the output stream
	Flush() error

	// Close closes the output writer
	Close() error
}
