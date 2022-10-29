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
	doneCh chan<- struct{},
) {
	var (
		moveCount   = 0
		currentCity = startingCity
	)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			// Attempt to lay siege to a random neighbor
			siegedNeighbor := a.siegeRandomNeighbor(currentCity)
			if siegedNeighbor == nil {
				// No neighbor can be sieged, the alien dies
				notifyCh(ctx, doneCh)

				return
			}

			// Check if the current city can be left
			if !currentCity.removeInvader(a.id) {
				// The alien cannot leave the current city because it
				// has been killed, remove the siege from the neighbor
				siegedNeighbor.liftSiege(a.id)

				notifyCh(ctx, doneCh)

				return
			}

			currentCity = siegedNeighbor

			// Invade the sieged neighbor
			currentCity.addInvader(a.id)

			// Increase the movement counter
			moveCount++

			// Check if max moves have been reached
			if moveCount >= maxMoveCount {
				notifyCh(ctx, doneCh)

				return
			}
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

// siegeRandomNeighbor attempts to siege a random neighbor
// of the given city.
// The assumption is that if no suitable neighbor is found (alien is trapped in a city),
// the alien dies.
// Returns the sieged city, if any
func (a *alien) siegeRandomNeighbor(c *city) *city {
	if len(c.neighbors) == 0 {
		// There are no neighbors the alien can move to,
		// so the alien dies
		return nil
	}

	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())

	// While there are still valid neighbors, attempt to siege
	// them randomly
	for c.hasAccessibleNeighbors() {
		//nolint:gosec
		randNeighbor := c.neighbors[direction(rand.Intn(numDirections))]

		if randNeighbor == nil {
			// No neighbor in this direction, try again
			continue
		}

		// Attempt to lay siege to the random neighbor
		if !randNeighbor.laySiege(a.id) {
			// Unable to lay siege to the neighbor, even though
			// they are a viable candidate
			continue
		}

		return randNeighbor
	}

	// There are no suitable neighbors present to which
	// the alien can lay siege to. It is assumed that the alien dies in this
	// situation
	return nil
}
