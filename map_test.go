package main

import (
	"testing"

	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
)

// arrayReader is a simple city array input reader used for testing
type arrayReader struct {
	cityArray []string
	index     int
}

func newArrayReader(cityArray []string) mapReader {
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

// TestMap_InitMap makes sure the earth city map
// is properly initialized using an input stream
func TestMap_InitMap(t *testing.T) {
	t.Parallel()

	var (
		cityInputs = []string{
			"Foo north=Bar west=Baz south=Qu-ux",
			"Bar south=Foo west=Bee",
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
