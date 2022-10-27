package main

import (
	"math/rand"
	"time"
)

type direction int

const (
	numDirections = 4
)

// Possible directions
const (
	north direction = iota
	south
	east
	west
)

// neighbors holds information on the adjacent cities
type neighbors map[direction]*city

// city represents a single unique city instance
type city struct {
	name      string
	neighbors neighbors
}

// newCity generates a new city instance
func newCity(name string) *city {
	return &city{
		name:      name,
		neighbors: make(map[direction]*city),
	}
}

// addNeighbor adds a new neighbor to the city.
// Additionally, it overwrites the previous neighbor entry, if any
func (c *city) addNeighbor(direction direction, city *city) {
	c.neighbors[direction] = city
}

// getRandomNeighbor returns a random city neighbor that is present.
// If no neighbors are present, it returns nil
func (c *city) getRandomNeighbor() *city {
	if len(c.neighbors) == 0 {
		// There are no neighbors present
		return nil
	}

	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())

	// Get a random present neighbor
	getRandCity := func() *city {
		return c.neighbors[direction(rand.Intn(numDirections))]
	}

	randCity := getRandCity()
	for randCity == nil {
		randCity = getRandCity()
	}

	return randCity
}
