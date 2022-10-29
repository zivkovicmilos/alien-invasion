package game

import (
	"context"
	"math/rand"
	"time"
)

// alien defines the single alien instance
type alien struct {
	id int
}

// newAlien creates a new alien instance
func newAlien(id int) *alien {
	return &alien{
		id: id,
	}
}

// runAlien runs the alien's main run loop
func (a *alien) runAlien(
	ctx context.Context,
	startingCity *city,
	movesCompletedCh chan<- struct{},
	alienKilledCh chan<- struct{},
) {
	var (
		moveCount = 0
		nextCity  = startingCity
	)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			// Check if the current city has been destroyed
			if !nextCity.isAccessible() {
				notifyCh(ctx, alienKilledCh)

				return
			}

			// Check if max moves have been reached
			if moveCount >= maxMoveCount {
				notifyCh(ctx, movesCompletedCh)

				return
			}

			// Attempt to move the alien to a new random neighbor city
			if a.invadeRandomNeighbor(nextCity) {
				// The city has been destroyed, and the alien killed
				notifyCh(ctx, alienKilledCh)

				return
			}

			// Alien has not been killed, move on to the next town
			moveCount++
		}
	}
}

// notifyCh safely alerts the channel of a notification,
// while making sure the running thread is properly cancelled
func notifyCh(ctx context.Context, ch chan<- struct{}) {
	select {
	case <-ctx.Done():
		return
	case ch <- struct{}{}:
		return
	}
}

// invadeRandomNeighbor attempts to invade a random neighbor
// of the given city.
// The assumption is that if no suitable neighbor is found (alien is trapped in a city),
// the alien dies.
// Returns a flag indicating if the alien died in combat (or loneliness in trapped cities)
func (a *alien) invadeRandomNeighbor(c *city) bool {
	if len(c.neighbors) == 0 {
		// There are no neighbors the alien can move to,
		// so the alien dies
		return true
	}

	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())

	// While there are still valid neighbors, attempt to invade
	// them randomly
	for c.hasAccessibleNeighbors() {
		//nolint:gosec
		randNeighbor := c.neighbors[direction(rand.Intn(numDirections))]

		if randNeighbor == nil {
			// Invalid direction selected, try again
			continue
		}

		// Attempt to invade the random neighbor
		if randNeighbor.addInvader(a.id) {
			// Managed to invade, remove the alien from the current city
			c.removeInvader(a.id)

			// Check if the alien died in combat
			return !randNeighbor.isAccessible()
		}
	}

	// There are no suitable neighbors present to which
	// the alien can move to. It is assumed that the alien dies in this
	// situation
	return true
}
