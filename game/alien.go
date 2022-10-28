package game

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-hclog"
)

// alien defines the single alien instance
type alien struct {
	id  int
	log hclog.Logger
}

// newAlien creates a new alien instance
func newAlien(id int, logger hclog.Logger) *alien {
	return &alien{
		id:  id,
		log: logger.Named(fmt.Sprintf("alien %d", id)),
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

	defer func() {
		a.log.Info("Finished!")
	}()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			// Check if max moves have been reached
			if moveCount >= maxMoveCount {
				notifyCh(ctx, movesCompletedCh)

				return
			}

			// Increase the move count
			moveCount++

			// Attempt to move the alien to the city
			if nextCity.addInvader(a.id) {
				// The city has been destroyed, and the alien killed
				notifyCh(ctx, alienKilledCh)

				return
			}

			// Get a new random neighboring city
			nextCity = nextCity.getRandomNeighbor()
			if nextCity == nil {
				// No more active neighbors, the alien is trapped with nowhere to go.
				// The assumption is that the alien dies here. An alternative would be to
				// pick a new random city for "teleportation", assuming the alien is advanced enough
				notifyCh(ctx, alienKilledCh)

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
