package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"

	"github.com/hashicorp/go-hclog"
	"github.com/spf13/cobra"
	"github.com/zivkovicmilos/alien-invasion/game"
	"github.com/zivkovicmilos/alien-invasion/stream"
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
	// Create an instance of the file reader
	fileReader, err := stream.NewFileReader(params.mapPath)
	if err != nil {
		return fmt.Errorf("unable to create a file reader, %w", err)
	}

	// Create an instance of the logger
	logger := hclog.New(&hclog.LoggerOptions{
		Name:  "alien-invasion",
		Level: hclog.LevelFromString(params.logLevel),
	})

	// Create an instance of the Earth map
	earthMap := game.NewEarthMap(logger)

	// Init the map from the map file
	earthMap.InitMap(fileReader)

	// Simulate the invasion
	var (
		wg                 sync.WaitGroup
		simulationComplete = make(chan struct{})
	)

	// The assumption is that very large invasion simulations
	// can take an arbitrary amount of time, depending on the map size
	// and alien count. In order to possibly prevent this, system-wide cancel
	// signals are monitored (CTRL-C, etc)
	simulationCtx, cancelSimulation := context.WithCancel(context.Background())
	defer cancelSimulation()

	wg.Add(1)

	go func() {
		defer func() {
			wg.Done()
		}()

		earthMap.SimulateInvasion(simulationCtx, params.n)
		close(simulationComplete)
	}()

	// Wait for either the simulation to complete,
	// or the user to exit
	select {
	// Get the system-wide signal handler
	case <-getTerminationSignalCh():
		// Shut down the simulation
		cancelSimulation()
	// Wait for the simulation to complete
	case <-simulationComplete:
	}

	// Wait for the simulation to gracefully exit
	wg.Wait()

	// Set up the output writer
	writer, err := getOutputWriter()
	if err != nil {
		return err
	}

	// Write the invasion output to the file
	if err := earthMap.WriteOutput(writer); err != nil {
		return fmt.Errorf("unable to write output to file, %w", err)
	}

	logger.Info("Invasion completed successfully!")

	return nil
}

// getOutputWriter returns the appropriate output writer
// based on user preferences
func getOutputWriter() (stream.OutputWriter, error) {
	var (
		err error

		writer = stream.NewConsoleWriter()
	)

	if params.outputPath != "" {
		// Output file is set, make sure it is valid
		writer, err = stream.NewFileWriter(params.outputPath)

		if err != nil {
			return nil, fmt.Errorf("unable to create an output file, %w", err)
		}
	}

	return writer, nil
}

// getTerminationSignalCh returns a listen channel for
// system-wide stop signals
func getTerminationSignalCh() <-chan os.Signal {
	signalCh := make(chan os.Signal, 1)
	signal.Notify(
		signalCh,
		os.Interrupt,
		syscall.SIGTERM,
		syscall.SIGHUP,
	)

	return signalCh
}
