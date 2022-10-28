package main

import (
	"github.com/zivkovicmilos/alien-invasion/cmd"
)

func main() {
	// Run the base command
	cmd.NewRootCommand().Execute()
}
