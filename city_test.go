package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// generateRandomCities generates random cities
func generateRandomCities(count int) []*city {
	cities := make([]*city, count)

	for i := 0; i < count; i++ {
		cities[i] = newCity(fmt.Sprintf("city %d", i))
	}

	return cities
}

// TestCity_AddNeighbor makes sure neighbors are added correctly
func TestCity_AddNeighbor(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		name       string
		neighbors  []*city
		directions []direction
	}{
		{
			"single neighbor",
			generateRandomCities(1),
			[]direction{north},
		},
		{
			"multiple neighbors",
			generateRandomCities(4),
			[]direction{north, south, east, west},
		},
		{
			"multiple neighbors with overwrites",
			generateRandomCities(5),
			[]direction{north, south, east, west, north},
		},
	}

	for _, testCase := range testTable {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Create a reference city
			city := newCity("city name")

			for index, neighbor := range testCase.neighbors {
				// Grab the direction
				direction := testCase.directions[index]

				// Add the neighbor
				city.addNeighbor(direction, neighbor)

				// Make sure the neighbor is added
				assert.Equal(t, neighbor.name, city.neighbors[direction].name)
			}

			expectedNeighbors := len(testCase.neighbors)
			if expectedNeighbors > numDirections {
				// There can be no more than 4 neighbors
				expectedNeighbors = numDirections
			}

			assert.Len(t, city.neighbors, expectedNeighbors)
		})
	}
}

func TestCity_GetRandomNeighbor(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		name         string
		neighbors    []*city
		expectResult bool
	}{
		{
			"no valid neighbors",
			generateRandomCities(0),
			false,
		},
		{
			"valid random neighbor",
			generateRandomCities(numDirections),
			true,
		},
	}

	for _, testCase := range testTable {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Create a reference city
			city := newCity("city name")

			for index, neighbor := range testCase.neighbors {
				// Grab a random direction
				direction := direction(index % numDirections)

				// Add the neighbor
				city.addNeighbor(direction, neighbor)
			}

			// Get a random neighbor
			randomNeighbor := city.getRandomNeighbor()

			if testCase.expectResult {
				assert.NotNil(t, randomNeighbor)
			} else {
				assert.Nil(t, randomNeighbor)
			}
		})
	}
}
