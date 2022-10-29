package game

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

// TestCity_RemoveNeighbors makes sure neighbors are removed correctly
func TestCity_RemoveNeighbors(t *testing.T) {
	t.Parallel()

	var (
		city      = newCity("city name")
		neighbors = generateRandomCities(numDirections)
	)

	directions := []direction{north, east, west, south}

	// Add the random neighbors
	for index, neighbor := range neighbors {
		city.addNeighbor(directions[index], neighbor)
	}

	// Make sure the neighbors are added successfully
	assert.Len(t, city.neighbors, len(neighbors))

	// Remove every other neighbor
	for i := 0; i < len(neighbors); i += 2 {
		city.removeNeighbor(directions[i])
	}

	// Make sure the neighbors are removed successfully
	assert.Len(t, city.neighbors, len(neighbors)/2)
}

// TestCity_Direction makes sure the direction helper methods work fine
func TestCity_Direction(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		direction        direction
		expectedOpposite direction
	}{
		{
			north,
			south,
		},
		{
			south,
			north,
		},
		{
			east,
			west,
		},
		{
			west,
			east,
		},
	}

	for _, testCase := range testTable {
		testCase := testCase

		t.Run(
			fmt.Sprintf(
				"opposite direction of %s",
				testCase.direction.getName(),
			), func(t *testing.T) {
				t.Parallel()

				assert.Equal(
					t,
					testCase.expectedOpposite,
					testCase.direction.getOpposite(),
				)
			},
		)
	}
}

// TestCity_AddInvader makes sure invaders
// are properly added to the city
func TestCity_AddInvader(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		name            string
		initialInvaders []int
		invader         int

		shouldDestroyCity bool
		shouldAddInvader  bool
	}{
		{
			"valid invader addition",
			[]int{},
			0,
			false,
			true,
		},
		{
			"valid invader addition, with present invader",
			[]int{0},
			1,
			true,
			true,
		},
		{
			"invalid invader addition, city destroyed",
			[]int{0, 1},
			2,
			true,
			false,
		},
	}

	for _, testCase := range testTable {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Create the city
			c := newCity("city name")

			// Add initial invaders
			for invader := range testCase.initialInvaders {
				assert.True(t, c.addInvader(invader))
			}

			// Make sure all initial invaders are present
			assert.Equal(t, len(testCase.initialInvaders), c.numInvaders())

			// Attempt to add the invader
			invaderAdded := c.addInvader(testCase.invader)

			// Check if the invader was added
			assert.Equal(t, testCase.shouldAddInvader, invaderAdded)

			// Check if the city was destroyed
			assert.Equal(t, testCase.shouldDestroyCity, c.destroyed)
		})
	}
}

// TestCity_RemoveInvader makes sure invaders are properly removed
// from the city
func TestCity_RemoveInvader(t *testing.T) {
	t.Parallel()

	var (
		c        = newCity("city name")
		invaders = []int{0, 1}
	)

	// Add initial invaders
	for invader := range invaders {
		assert.True(t, c.addInvader(invader))
	}

	// Remove the first invader
	c.removeInvader(invaders[0])

	// Make sure the invader has been removed
	assert.Equal(t, c.numInvaders(), len(invaders)-1)
}

// TestCity_Accessible checks that the city is accessible
// under specific circumstances
func TestCity_Accessible(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		name      string
		invaders  []int
		destroyed bool

		shouldBeAccessible bool
	}{
		{
			"accessible city without invaders",
			[]int{},
			false,
			true,
		},
		{
			"accessible city with single invader",
			[]int{0},
			false,
			true,
		},
		{
			"non-accessible city with 2 invaders",
			[]int{0, 1},
			false,
			false,
		},
		{
			"non-accessible destroyed city",
			[]int{},
			true,
			false,
		},
	}

	for _, testCase := range testTable {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			c := newCity("city name")

			// Add initial invaders
			for invader := range testCase.invaders {
				assert.True(t, c.addInvader(invader))
			}

			c.destroyed = testCase.destroyed

			// Check if the city is accessible
			assert.Equal(t, testCase.shouldBeAccessible, c.isAccessible())
		})
	}
}

// TestCity_AccessibleNeighbors checks if neighbors
// are accessible due to their characteristics
func TestCity_AccessibleNeighbors(t *testing.T) {
	t.Parallel()

	// Create an occupied neighbor
	occupiedNeighbor := newCity("occupied")
	occupiedNeighbor.addInvader(0)
	occupiedNeighbor.addInvader(1)

	// Create a destroyed neighbor
	destroyedNeighbor := newCity("destroyed")
	destroyedNeighbor.destroyed = true

	// Create a valid neighbor
	validNeighbor := newCity("valid")

	testTable := []struct {
		name      string
		neighbors neighbors

		shouldHaveValidNeighbor bool
	}{
		{
			"no valid neighbors",
			neighbors{
				north: occupiedNeighbor,
				south: destroyedNeighbor,
			},
			false,
		},
		{
			"valid neighbor",
			neighbors{
				north: occupiedNeighbor,
				south: validNeighbor,
				west:  destroyedNeighbor,
			},
			true,
		},
	}

	for _, testCase := range testTable {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Create the initial city and add neighbors
			c := newCity("city name")
			c.neighbors = testCase.neighbors

			assert.Equal(
				t,
				testCase.shouldHaveValidNeighbor,
				c.hasAccessibleNeighbors(),
			)
		})
	}
}
