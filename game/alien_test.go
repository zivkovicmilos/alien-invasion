package game

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestAlien_InvadeRandomNeighbor verifies that the alien
// can successfully siege and invade a random city neighbor
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

		expectedNeighbor *city
	}{
		{
			"No neighbors",
			&city{
				neighbors: neighbors{},
			},
			nil,
		},
		{
			"No valid neighbors",
			&city{
				neighbors: neighbors{
					north: invalidCity,
				},
			},
			nil,
		},
		{
			"Valid neighbor",
			&city{
				neighbors: neighbors{
					north: validCity,
				},
			},
			validCity,
		},
	}

	for _, testCase := range testTable {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			// Make sure the alien can siege a city
			siegedNeighbor := newAlien(alienID).siegeRandomNeighbor(testCase.refCity)
			assert.Equal(
				t,
				testCase.expectedNeighbor,
				siegedNeighbor,
			)

			// Make sure the alien is in the correct city
			if testCase.expectedNeighbor != nil {
				if siegedNeighbor == nil {
					t.Fatal("neighbor should be sieged, but isn't")
				}

				// Make sure the invader is removed from the start city
				assert.True(t, testCase.refCity.removeInvader(alienID))
				assert.Len(t, testCase.refCity.invaders, 0)
				assert.Len(t, testCase.refCity.sieges, 0)

				// Make sure the invader is added to the end city
				siegedNeighbor.addInvader(alienID)

				assert.Len(t, siegedNeighbor.invaders, 1)
				assert.Len(t, siegedNeighbor.sieges, 1)

				for invaderID := range siegedNeighbor.invaders {
					assert.Equal(t, alienID, invaderID)
				}

				for invaderID := range siegedNeighbor.sieges {
					assert.Equal(t, alienID, invaderID)
				}
			}
		})
	}
}

// TestAlien_NonSiegeableCities verifies that the alien
// cannot successfully siege any city neighbor
func TestAlien_NonSiegeableCities(t *testing.T) {
	t.Parallel()

	// Make sure all neighbors can't be sieged, but
	// are valid (not destroyed)
	neighbor := newCity("neighbor city")

	neighbor.sieges[0] = struct{}{}
	neighbor.sieges[1] = struct{}{}

	currentCity := newCity("current city")
	currentCity.neighbors = neighbors{
		north: neighbor,
	}

	var wg sync.WaitGroup

	wg.Add(1)

	go func(c *city) {
		defer func() {
			wg.Done()
		}()

		// After some time, all accessible neighbor cities become destroyed
		<-time.After(time.Second)

		c.addInvader(0)
		c.addInvader(1)
	}(neighbor)

	// Attempt to siege a random neighbor
	siegedNeighbor := newAlien(0).siegeRandomNeighbor(currentCity)

	wg.Wait()

	// Make sure no neighbor is sieged
	assert.Nil(t, siegedNeighbor)
}

// TestAlien_AlienKilled_StartingCityDestroyed verifies the main run functionality
// of the alien thread, and that it gets killed off appropriately
// when it finds itself in a destroyed starting city
func TestAlien_AlienKilled_StartingCityDestroyed(t *testing.T) {
	t.Parallel()

	var (
		wg sync.WaitGroup

		a            = newAlien(0)
		invadingCity = newCity("invading city")

		alienDone = false
		doneCh    = make(chan struct{})
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
		case <-doneCh:
			alienDone = true
		}
	}()

	// Start the main loop
	go a.runAlien(ctx, invadingCity, doneCh)

	wg.Wait()

	// Make sure the alien alerted the channel about dying
	assert.True(t, alienDone)
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

		alienDone   = false
		alienDoneCh = make(chan struct{})
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
		case <-alienDoneCh:
			alienDone = true
		}
	}()

	// Start the main loop
	go a.runAlien(ctx, invadingCity, alienDoneCh)

	wg.Wait()

	// Make sure the alien alerted the channel about dying
	assert.True(t, alienDone)

	// Make sure the cities have not been destroyed
	assert.False(t, invadingCity.destroyed)
	assert.False(t, invadingCityNeighbor.destroyed)
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

		alienDone   = false
		alienDoneCh = make(chan struct{})
	)

	// Make sure the neighbor city has at least one invader
	neighbor := newCity("neighbor with invader")

	neighbor.laySiege(1)
	neighbor.addInvader(1)

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
		case <-alienDoneCh:
			alienDone = true
		}
	}()

	// Start the main loop
	go a.runAlien(ctx, invadingCity, alienDoneCh)

	wg.Wait()

	// Make sure the alien alerted the channel about dying
	assert.True(t, alienDone)

	// Make sure the city has been destroyed
	assert.True(t, neighbor.destroyed)
}

// TestAlien_AlienKilled_CitySiegedNotInvaded verifies the main run functionality
// of the alien thread, and that it gets killed off appropriately
// when it sieges a city, but cannot leave the current one (doesn't invade it)
func TestAlien_AlienKilled_CitySiegedNotInvaded(t *testing.T) {
	t.Parallel()

	var (
		wg sync.WaitGroup

		a            = newAlien(0)
		invadingCity = newCity("invading city")

		alienDone   = false
		alienDoneCh = make(chan struct{})
	)

	// Make sure the neighbor city is valid
	neighbor := newCity("valid neighbor")

	invadingCity.neighbors = neighbors{
		north: neighbor,
	}

	// Make sure the current city the alien is in is destroyed
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
		case <-alienDoneCh:
			alienDone = true
		}
	}()

	// Start the main loop
	go a.runAlien(ctx, invadingCity, alienDoneCh)

	wg.Wait()

	// Make sure the alien alerted the channel about dying
	assert.True(t, alienDone)

	// Make sure the siege is removed
	assert.Len(t, neighbor.sieges, 0)
}
