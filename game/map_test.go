package game

import (
	"fmt"
	"testing"

	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
	"github.com/zivkovicmilos/alien-invasion"
)

// arrayReader is a simple city array input reader used for testing
type arrayReader struct {
	cityArray []string
	index     int
}

func newArrayReader(cityArray []string) main.inputReader {
	return &arrayReader{
		cityArray: cityArray,
		index:     0,
	}
}

func (ar *arrayReader) hasMoreCities() bool {
	return ar.index < len(ar.cityArray)
}

func (ar *arrayReader) readCity() string {
	line := ar.cityArray[ar.index]
	ar.index++

	return line
}

func (ar *arrayReader) close() error {
	return nil
}

type arrayWriter struct {
	outputArray []string
}

func newArrayWriter() *arrayWriter {
	return &arrayWriter{
		outputArray: make([]string, 0),
	}
}

func (aw *arrayWriter) write(s string) error {
	aw.outputArray = append(aw.outputArray, s)

	return nil
}

func (aw *arrayWriter) flush() error {
	return nil
}

func (aw *arrayWriter) close() error {
	return nil
}

// TestMap_InitMap makes sure the earth city map
// is properly initialized using an input stream
func TestMap_InitMap(t *testing.T) {
	t.Parallel()

	var (
		cityInputs = []string{
			"Foo north=Bar west=Baz south=Qu-ux",
			"Bar south=Foo west=Bee",
			"", // invalid input line
		}

		expectedCities = []struct {
			name      string
			neighbors neighbors
		}{
			{
				"Foo",
				neighbors{
					north: newCity("Bar"),
					west:  newCity("Baz"),
					south: newCity("Qu-ux"),
				},
			},
			{
				"Bar",
				neighbors{
					south: newCity("Foo"),
					west:  newCity("Bee"),
				},
			},
			{
				"Baz",
				neighbors{
					east: newCity("Foo"),
				},
			},
			{
				"Qu-ux",
				neighbors{
					north: newCity("Foo"),
				},
			},
			{
				"Bee",
				neighbors{
					east: newCity("Bar"),
				},
			},
		}
	)

	// Create a mock input reader
	reader := newArrayReader(cityInputs)

	// Create an instance of the earth map
	earthMap := newEarthMap(hclog.NewNullLogger())

	// Initialize the earth map using the reader
	earthMap.initMap(reader)

	// Make sure the cities are properly added
	assert.Len(t, earthMap.cityMap, len(expectedCities))

	// Make sure the cities are present in the city map,
	// and their neighbors are correct
	for _, expectedCity := range expectedCities {
		// Make sure the city is present
		city := earthMap.getCity(expectedCity.name)
		if city == nil {
			t.Fatalf("city %s not present in city map", expectedCity.name)
		}

		// Make sure the city's neighbors are correct
		assert.Len(t, city.neighbors, len(expectedCity.neighbors))

		for expectedDirection, expectedNeighbor := range expectedCity.neighbors {
			assert.Equal(t, expectedNeighbor.name, city.neighbors[expectedDirection].name)
		}
	}
}

// TestMap_RemoveCity makes sure cities are properly removed
func TestMap_RemoveCity(t *testing.T) {
	t.Parallel()

	var (
		cityInputs = []string{
			"Foo north=Bar",
			"Bar south=Foo",
		}

		expectedCities = []struct {
			name      string
			neighbors neighbors
		}{
			{
				"Bar",
				neighbors{}, // no neighbors as Foo should be removed
			},
		}
	)

	// Create a mock input reader
	reader := newArrayReader(cityInputs)

	// Create an instance of the earth map
	earthMap := newEarthMap(hclog.NewNullLogger())

	// Initialize the earth map using the reader
	earthMap.initMap(reader)

	// Make sure the cities are properly added
	assert.Len(t, earthMap.cityMap, 2)

	// Remove a valid city
	earthMap.removeCity("Foo")

	// Attempt to remove an invalid city (no effect)
	earthMap.removeCity("Foo 2")

	// Make sure the city was removed
	assert.Len(t, earthMap.cityMap, 1)

	cityBar := earthMap.getCity(expectedCities[0].name)
	if cityBar == nil {
		t.Fatalf("city %s not present in city map", expectedCities[0].name)
	}

	// Make sure the city's neighbors are correct
	assert.Len(t, cityBar.neighbors, len(expectedCities[0].neighbors))
}

// TestMap_WriteOutput checks that the map output is valid
func TestMap_WriteOutput(t *testing.T) {
	t.Parallel()

	cityInputs := []string{
		"Foo north=Bar",
		"Bar south=Foo",
	}

	// Create a mock input reader
	reader := newArrayReader(cityInputs)

	// Create an instance of the earth map
	earthMap := newEarthMap(hclog.NewNullLogger())

	// Initialize the earth map using the reader
	earthMap.initMap(reader)

	// Make sure the cities are properly added
	assert.Len(t, earthMap.cityMap, 2)

	// Create a mock output writer
	writer := newArrayWriter()

	// Write the output
	assert.NoError(t, earthMap.writeOutput(writer))

	// Make sure the output is the same as the input
	// in this test case
	assert.Len(t, writer.outputArray, len(cityInputs))

	for _, outputLine := range writer.outputArray {
		// Make sure the output exactly matches one of the inputs
		// as nothing is unchanged in the map
		matchFound := false

		for _, input := range cityInputs {
			if fmt.Sprintf("%s\n", input) == outputLine {
				matchFound = true

				break
			}
		}

		// Check if a match has been found in the output
		if !matchFound {
			t.Fatalf("input line is not present in output, %s", outputLine)
		}
	}
}
