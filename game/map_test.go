package game

import (
	"fmt"
	"testing"

	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
	"github.com/zivkovicmilos/alien-invasion/stream"
)

// arrayReader is a simple city array input reader used for testing
type arrayReader struct {
	cityArray []string
	index     int
}

func newArrayReader(cityArray []string) stream.InputReader {
	return &arrayReader{
		cityArray: cityArray,
		index:     0,
	}
}

func (ar *arrayReader) HasMoreCities() bool {
	return ar.index < len(ar.cityArray)
}

func (ar *arrayReader) ReadCity() string {
	line := ar.cityArray[ar.index]
	ar.index++

	return line
}

func (ar *arrayReader) Close() error {
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

func (aw *arrayWriter) Write(s string) error {
	aw.outputArray = append(aw.outputArray, s)

	return nil
}

func (aw *arrayWriter) Flush() error {
	return nil
}

func (aw *arrayWriter) Close() error {
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
	earthMap := NewEarthMap(hclog.NewNullLogger())

	// Initialize the earth map using the reader
	earthMap.InitMap(reader)

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
	earthMap := NewEarthMap(hclog.NewNullLogger())

	// Initialize the earth map using the reader
	earthMap.InitMap(reader)

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
	earthMap := NewEarthMap(hclog.NewNullLogger())

	// Initialize the earth map using the reader
	earthMap.InitMap(reader)

	// Make sure the cities are properly added
	assert.Len(t, earthMap.cityMap, 2)

	// Create a mock output writer
	writer := newArrayWriter()

	// Write the output
	assert.NoError(t, earthMap.WriteOutput(writer))

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

// TestMap_GetRandomCities makes sure random cities are properly sampled
// from the earth map
func TestMap_GetRandomCities(t *testing.T) {
	t.Parallel()

	cityInputs := []string{
		"Foo",
		"Bar",
		"Baz",
	}

	// Create a mock input reader
	reader := newArrayReader(cityInputs)

	// Create an instance of the earth map
	earthMap := NewEarthMap(hclog.NewNullLogger())

	// Initialize the earth map using the reader
	earthMap.InitMap(reader)

	// Get the random cities
	randomCount := 10
	randomCities := earthMap.getRandomCities(randomCount)

	// Make sure the random cities are valid
	assert.Len(t, randomCities, randomCount)

	for _, randomCity := range randomCities {
		assert.Equal(t, earthMap.cityMap[randomCity.name], randomCity)
	}
}

// TestMap_PruneDestroyedCities verifies the city pruning
// functionality from the earth map
func TestMap_PruneDestroyedCities(t *testing.T) {
	t.Parallel()

	// Create two neighboring cities, one destroyed,
	// and one in-tact
	var (
		cityFoo = newCity("Foo")
		cityBar = newCity("Bar")
	)

	cityFoo.destroyed = true
	cityFoo.neighbors = neighbors{
		north: cityBar,
	}

	cityBar.neighbors = neighbors{
		south: cityFoo,
	}

	testTable := []struct {
		name   string
		cities []*city

		expectedCities     []*city
		expectedPruneCount int
	}{
		{
			"no cities to prune",
			[]*city{
				{
					name: "name",
				},
			},
			[]*city{
				{
					name: "name",
				},
			},
			0,
		},
		{
			"valid city prune",
			[]*city{
				cityFoo, // should get pruned out
				cityBar,
			},
			[]*city{
				{
					name:      "Bar",
					neighbors: neighbors{}, // no neighbors
				},
			},
			1,
		},
	}

	for _, testCase := range testTable {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			m := NewEarthMap(hclog.NewNullLogger())

			// Add initial cities
			for _, city := range testCase.cities {
				m.cityMap[city.name] = city
			}

			// Make sure all initial cities are present
			assert.Len(t, m.cityMap, len(testCase.cities))

			// Prune the destroyed cities
			assert.Equal(t, testCase.expectedPruneCount, m.pruneDestroyedCities())

			// Make sure the earth map matches the expected one
			// after pruning
			assert.Len(t, m.cityMap, len(testCase.expectedCities))

			for _, expectedCity := range testCase.expectedCities {
				city := m.cityMap[expectedCity.name]

				// Make sure it exists in the earth map
				if city == nil {
					t.Fatalf("City with name %s not present", expectedCity.name)
				}

				// Make sure the neighbors are the same
				assert.Len(t, city.neighbors, len(expectedCity.neighbors))

				for direction, neighbor := range city.neighbors {
					assert.Equal(t, expectedCity.neighbors[direction].name, neighbor.name)
				}
			}
		})
	}
}
