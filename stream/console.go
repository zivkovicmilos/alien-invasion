package stream

import (
	"fmt"
)

// ConsoleWriter outputs the data to standard output (console)
type ConsoleWriter struct {
}

func NewConsoleWriter() OutputWriter {
	return &ConsoleWriter{}
}

func (cw *ConsoleWriter) Write(s string) error {
	fmt.Print(s)

	return nil
}

func (cw *ConsoleWriter) Close() error {
	return nil
}

func (cw *ConsoleWriter) Flush() error {
	return nil
}
