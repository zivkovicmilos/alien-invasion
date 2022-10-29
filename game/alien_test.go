package game

import (
	"testing"

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
