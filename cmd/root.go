package cmd

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
)

var (
	errInvalidAlienNumber = errors.New("invalid number of aliens provided")
	errAlienNumberMissing = errors.New("number of aliens not provided as argument")
)

type RootCommand struct {
	baseCmd *cobra.Command
}

func NewRootCommand() *RootCommand {
	rootCommand := &RootCommand{
		baseCmd: &cobra.Command{
			Short:   "A program for simulating the invasion of mad aliens on Earth",
			Args:    validateArguments,
			PreRunE: runPreRun,
			RunE:    runCommand,
		},
	}

	// Set the base flags
	setFlags(rootCommand.baseCmd)

	// Set the required flags
	setRequiredFlags(rootCommand.baseCmd, params.getRequiredFlags())

	return rootCommand
}

func (rc *RootCommand) Execute() {
	if err := rc.baseCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)

		os.Exit(1)
	}
}

// setFlags sets the base command flags
func setFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(
		&params.mapPath,
		mapPathFlag,
		"",
		"The path to the input map file of the Earth",
	)

	cmd.Flags().StringVar(
		&params.outputPath,
		outputPathFlag,
		"",
		"The path to output the Earth map after the invasion. If omitted, the output is directed to the console",
	)

	cmd.Flags().StringVar(
		&params.logLevel,
		logLevelFlag,
		"INFO",
		"The log level for the program execution",
	)
}

// validateArguments validates that the command line arguments are valid
func validateArguments(cmd *cobra.Command, args []string) error {
	// Make sure at least one argument is present (the number of aliens)
	if err := cobra.MinimumNArgs(1)(cmd, args); err != nil {
		return errAlienNumberMissing
	}

	// Make sure the number of aliens is valid
	numAliens, err := strconv.Atoi(args[0])
	if err != nil || numAliens == 0 {
		return errInvalidAlienNumber
	}

	return nil
}

// setRequiredFlags marks the specified flags as required
func setRequiredFlags(cmd *cobra.Command, requiredFlags []string) {
	for _, requiredFlag := range requiredFlags {
		_ = cmd.MarkFlagRequired(requiredFlag)
	}
}

// runPreRun instantiates the command line arguments for the runtime
func runPreRun(_ *cobra.Command, args []string) error {
	numAliens, err := strconv.Atoi(args[0])
	if err != nil || numAliens == 0 {
		return errInvalidAlienNumber
	}

	// Set the number of aliens
	params.n = numAliens

	return nil
}

// runCommand runs the root command
func runCommand(_ *cobra.Command, _ []string) error {
	fmt.Println("Hello!")

	return nil
}
