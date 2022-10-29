package game

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestAlien_InvadeRandomNeighbor verifies that the alien
// can successfully invade a random city neighbor
func TestAlien_InvadeRandomNeighbor(t *testing.T) {
	t.Parallel()

	var (
		invalidCity = newCity("invalid city")
		validCity   = newCity("valid city")
		alienID     = 10
	)

	invalidCity.destroyed = true

	testTable := []struct {
		name    string
		refCity *city

		shouldDie    bool
		shouldInvade bool
	}{
		{
			"No neighbors",
			&city{
				neighbors: neighbors{},
			},
			true,
			false,
		},
		{
			"No valid neighbors",
			&city{
				neighbors: neighbors{
					north: invalidCity,
				},
			},
			true,
			false,
		},
		{
			"Valid neighbor",
			&city{
				neighbors: neighbors{
					north: validCity,
				},
			},
			false, // the valid city doesn't have any invaders / is not destroyed
			true,
		},
	}

	for _, testCase := range testTable {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Make sure the alien is properly killed off
			assert.Equal(
				t,
				testCase.shouldDie,
				newAlien(alienID).invadeRandomNeighbor(testCase.refCity),
			)

			// Make sure the alien is in the correct city
			if testCase.shouldInvade {
				// Make sure the invader is removed from the start city
				assert.Len(t, testCase.refCity.invaders, 0)

				// Make sure the invader is added to the end city
				for _, neighbor := range testCase.refCity.neighbors {
					assert.Len(t, neighbor.invaders, 1)

					for invaderID := range neighbor.invaders {
						assert.Equal(t, alienID, invaderID)
					}
				}
			}
		})
	}
}

// TestAlien_AlienKilled_CityNotAccessible verifies the main run functionality
// of the alien thread, and that it gets killed off appropriately
// when it finds itself in a destroyed starting city
func TestAlien_AlienKilled_CityNotAccessible(t *testing.T) {
	t.Parallel()

	var (
		wg sync.WaitGroup

		a            = newAlien(0)
		invadingCity = newCity("invading city")

		alienKilled   = false
		alienKilledCh = make(chan struct{})
	)

	// Mark the starting city as destroyed
	invadingCity.destroyed = true

	ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFn()

	// Create a listener thread
	wg.Add(1)
	go func() {
		defer func() {
			wg.Done()
		}()

		select {
		case <-ctx.Done():
		case <-alienKilledCh:
			alienKilled = true
		}
	}()

	// Start the main loop
	go a.runAlien(ctx, invadingCity, make(chan struct{}), alienKilledCh)

	wg.Wait()

	// Make sure the alien alerted the channel about dying
	assert.True(t, alienKilled)
}

// TestAlien_AlienKilled_MaxMovesReached verifies the main run functionality
// of the alien thread, and that it gets killed off appropriately
// when it reaches the maximum number of moves
func TestAlien_AlienKilled_MaxMovesReached(t *testing.T) {
	t.Parallel()

	var (
		wg sync.WaitGroup

		a                    = newAlien(0)
		invadingCity         = newCity("invading city")
		invadingCityNeighbor = newCity("invading city neighbor")

		maxMovesReached   = false
		maxMovesReachedCh = make(chan struct{})
	)

	// Create 2 cities that the alien will move through
	// until it reaches max moves
	invadingCity.neighbors = neighbors{
		north: invadingCityNeighbor,
	}

	invadingCityNeighbor.neighbors = neighbors{
		south: invadingCity,
	}

	ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFn()

	// Create a listener thread
	wg.Add(1)
	go func() {
		defer func() {
			wg.Done()
		}()

		select {
		case <-ctx.Done():
		case <-maxMovesReachedCh:
			maxMovesReached = true
		}
	}()

	// Start the main loop
	go a.runAlien(ctx, invadingCity, maxMovesReachedCh, make(chan struct{}))

	wg.Wait()

	// Make sure the alien alerted the channel about dying
	assert.True(t, maxMovesReached)
}

// TestAlien_AlienKilled_CityInvaded verifies the main run functionality
// of the alien thread, and that it gets killed off appropriately
// when it invades a city and encounters another alien
func TestAlien_AlienKilled_CityInvaded(t *testing.T) {
	t.Parallel()

	var (
		wg sync.WaitGroup

		a            = newAlien(0)
		invadingCity = newCity("invading city")

		alienKilled   = false
		alienKilledCh = make(chan struct{})
	)

	// Make sure the neighbor city has at least one invader
	neighbor := newCity("neighbor with invader")
	neighbor.invaders[1] = struct{}{}

	invadingCity.neighbors = neighbors{
		north: neighbor,
	}

	ctx, cancelFn := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFn()

	// Create a listener thread
	wg.Add(1)
	go func() {
		defer func() {
			wg.Done()
		}()

		select {
		case <-ctx.Done():
		case <-alienKilledCh:
			alienKilled = true
		}
	}()

	// Start the main loop
	go a.runAlien(ctx, invadingCity, make(chan struct{}), alienKilledCh)

	wg.Wait()

	// Make sure the alien alerted the channel about dying
	assert.True(t, alienKilled)
}
