//nolint:unused
package game

import (
	"fmt"
	"sync"

	"github.com/hashicorp/go-hclog"
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

// getOpposite returns the opposite direction for the given
// direction
func (d direction) getOpposite() direction {
	switch d {
	case north:
		return south
	case south:
		return north
	case east:
		return west
	default:
		return east
	}
}

// getName returns the name of the given direction
func (d direction) getName() string {
	switch d {
	case north:
		return "north"
	case south:
		return "south"
	case east:
		return "east"
	default:
		return "west"
	}
}

// neighbors holds information on the adjacent cities
type neighbors map[direction]*city

// city represents a single unique city instance
type city struct {
	sync.RWMutex

	name      string
	neighbors neighbors
	log       hclog.Logger

	destroyed bool
	invaders  map[int]struct{}
}

// withLogger sets a specific city logger
func withLogger(log hclog.Logger) func(*city) {
	return func(c *city) {
		c.log = log
	}
}

// newCity generates a new city instance
func newCity(name string, opts ...func(*city)) *city {
	c := &city{
		name:      name,
		neighbors: make(map[direction]*city),
		invaders:  make(map[int]struct{}),
		log:       hclog.NewNullLogger(),
	}

	for _, callback := range opts {
		callback(c)
	}

	return c
}

// addNeighbor adds a new neighbor to the city.
// Additionally, it overwrites the previous neighbor entry, if any
func (c *city) addNeighbor(direction direction, city *city) {
	c.neighbors[direction] = city
}

// removeNeighbor removes a neighboring city in the
// specified direction
func (c *city) removeNeighbor(direction direction) {
	delete(c.neighbors, direction)
}

// hasAccessibleNeighbors checks travel is possible to
// neighbors of a given city
func (c *city) hasAccessibleNeighbors() bool {
	for _, neighbor := range c.neighbors {
		if neighbor.isAccessible() {
			return true
		}
	}

	return false
}

// addInvader adds an invader to the specified city.
// It returns a flag indicating if the invader was added.
// The alien can invade a city if:
//   - the city has not already been destroyed
//   - the city doesn't have 2 invaders present
// [Thread safe]
func (c *city) addInvader(alienID int) bool {
	c.Lock()
	defer c.Unlock()

	// Check if invasion is possible
	if c.destroyed || c.numInvaders() == 2 {
		// The city is already destroyed, or there are already 2 invaders present.
		// The assumption is that there can be no more than 2 invaders at any given time in a city
		return false
	}

	// Increase the number of invaders in a city
	c.invaders[alienID] = struct{}{}

	// Check if the city is destroyed
	if c.numInvaders() == 2 {
		// Mark the city as destroyed, print the invaders
		c.destroyed = true
		c.printInvaders()
	}

	return true
}

// removeInvader removes an invader from the city [Thread safe]
func (c *city) removeInvader(alienID int) {
	c.Lock()
	defer c.Unlock()

	delete(c.invaders, alienID)
}

// numInvaders returns the number of active invaders [NOT Thread safe]
func (c *city) numInvaders() int {
	return len(c.invaders)
}

// printInvaders prints the current invaders in the city. [NOT Thread safe]
func (c *city) printInvaders() {
	invaders := make([]int, len(c.invaders))

	i := 0

	for invader := range c.invaders {
		invaders[i] = invader
		i++
	}

	c.log.Info(
		fmt.Sprintf(
			"City has been destroyed by aliens %d and %d!",
			invaders[0],
			invaders[1],
		),
	)
}

// isDestroyed returns a flag indicating if a city has been
// destroyed or if there are two invaders present [Thread safe]
func (c *city) isAccessible() bool {
	c.RLock()
	defer c.RUnlock()

	return !c.destroyed && len(c.invaders) != 2
}
