package game

import (
	"context"
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/zivkovicmilos/alien-invasion/stream"
)

// Predefined regexes for reading the input line
// from the map file
var (
	cityNameRegex = regexp.MustCompile(`^[^ ]+`)

	northRegex = regexp.MustCompile(`north=([^ ]+)`)
	southRegex = regexp.MustCompile(`south=([^ ]+)`)
	eastRegex  = regexp.MustCompile(`east=([^ ]+)`)
	westRegex  = regexp.MustCompile(`west=([^ ]+)`)
)

// Defines the max move count for each alien on the map
const (
	maxMoveCount = 10000
)

// getDirectionRegex returns the specific direction regex for the input line
func getDirectionRegex(direction direction) *regexp.Regexp {
	switch direction {
	case north:
		return northRegex
	case south:
		return southRegex
	case east:
		return eastRegex
	default:
		return westRegex
	}
}

// EarthMap keeps track of all active Earth cities
type EarthMap struct {
	log hclog.Logger

	cityMap map[string]*city
}

// NewEarthMap creates a new instance of the earth map
func NewEarthMap(log hclog.Logger) *EarthMap {
	return &EarthMap{
		log:     log.Named("earth-map"),
		cityMap: make(map[string]*city),
	}
}

// InitMap initializes the city map using the specified reader
func (m *EarthMap) InitMap(reader stream.InputReader) {
	directions := []direction{north, south, east, west}

	// Read each city from the input stream, until it is depleted
	for reader.HasMoreCities() {
		cityLine := reader.ReadCity()

		// Grab the city name
		cityNameMatch := cityNameRegex.FindStringSubmatch(cityLine)
		if len(cityNameMatch) == 0 {
			// The assumption is that invalid city lines are skipped
			m.log.Error(
				fmt.Sprintf("Invalid city input line: %s", cityLine),
			)

			continue
		}

		// Create a new instance of a city
		cityName := cityNameMatch[0]
		city := newCity(cityName, withLogger(m.log.Named(cityName)))

		// Add the current city to the earth map
		m.addCity(city)

		// Check if there are neighboring cities from the input line
		for _, direction := range directions {
			match := getDirectionRegex(direction).FindStringSubmatch(cityLine)

			if len(match) == 0 {
				// No neighbors found for this direction
				continue
			}

			// Grab the neighbor from the city map if it's present, otherwise create it
			neighbor := m.getOrAddCity(match[1])

			// Add the current city as a new neighbor
			neighbor.addNeighbor(direction.getOpposite(), city)

			// Add the new neighbor to the current city
			city.addNeighbor(direction, neighbor)

			m.log.Debug(
				fmt.Sprintf(
					"Added %s as a %s neighbor of %s",
					neighbor.name,
					direction.getName(),
					city.name,
				),
			)
		}
	}

	m.log.Info(
		fmt.Sprintf("Map initialized with %d cities", len(m.cityMap)),
	)
}

// getCity fetches a city from the city map.
// If the city is not present, nil is returned
func (m *EarthMap) getCity(name string) *city {
	return m.cityMap[name]
}

// addCity appends a city to the city map
func (m *EarthMap) addCity(newCity *city) {
	m.cityMap[newCity.name] = newCity
}

// removeCity removes the city from the city map
func (m *EarthMap) removeCity(name string) {
	// Grab the city
	city := m.getCity(name)
	if city == nil {
		m.log.Warn(
			fmt.Sprintf("Attempted to remove a non-existing city, %s", name),
		)

		return
	}

	// Grab the neighbors
	neighbors := city.neighbors

	// Delete the city from the lookup reference
	delete(m.cityMap, name)

	// Remove the city from the reference of all neighbors
	for direction, neighbor := range neighbors {
		neighbor.removeNeighbor(direction.getOpposite())
	}
}

// getOrAddCity attempts to fetch a city from the city map.
// If the city is not present, it is created, appended to the city map
// and returned
func (m *EarthMap) getOrAddCity(name string) *city {
	city := m.getCity(name)

	if city == nil {
		// City not created yet, add it
		city = newCity(name, withLogger(m.log.Named(name)))

		m.addCity(city)
	}

	return city
}

// WriteOutput writes the current map layout to the specified
// output stream. It assumes that the output order is not important
func (m *EarthMap) WriteOutput(writer stream.OutputWriter) error {
	// Check if there are any cities left to output
	if len(m.cityMap) == 0 {
		m.log.Info("All cities were destroyed by mad aliens")
	}

	// Each city has an output format:
	// CityName direction=CityName...
	for _, city := range m.cityMap {
		var sb strings.Builder

		// Write the city name
		sb.WriteString(city.name)

		// For each direction, write the neighbor with the direction
		for direction, neighbor := range city.neighbors {
			sb.WriteString(
				fmt.Sprintf(
					" %s=%s",
					direction.getName(),
					neighbor.name,
				),
			)
		}

		if err := writer.Write(fmt.Sprintf("%s\n", sb.String())); err != nil {
			return fmt.Errorf("unable to write to output stream, %w", err)
		}
	}

	return writer.Flush()
}

// SimulateInvasion starts the invasion simulation using the provided number of aliens.
// The invasion consists of a few steps:
// 1. Randomly assign starting positions for aliens
// 2. Set the aliens loose on the Earth map
// 3. Wait until the program terminates (either):
//    - all aliens are dead
//    - all aliens moved at least 10k times (solves the "trapped" scenarios)
//    - the user terminated the program with an exit signal (CTRL-C)
// 4. Prune out destroyed cities from the map
func (m *EarthMap) SimulateInvasion(ctx context.Context, numAliens int) {
	// Check if there are cities on the map for the invasion
	if len(m.cityMap) == 0 {
		// There are no cities on the earth map for aliens
		// to destroy, so the simulation terminates
		m.log.Error("There are no cities for the mad aliens to invade")

		return
	}

	// Randomly assign starting positions for aliens
	randomCities := m.getRandomCities(numAliens)

	// Set the aliens loose on the Earth map
	var (
		aliensLeft  = numAliens
		alienDoneCh = make(chan struct{})

		wg sync.WaitGroup
	)

	workerContext, cancelFn := context.WithCancel(ctx)

	// Cleanup
	defer func() {
		// Close off the alien routines, and wait
		// for them to complete gracefully
		cancelFn()
		wg.Wait()

		close(alienDoneCh)

		// Prune out the destroyed cities
		m.log.Info(
			fmt.Sprintf(
				"A total of %d cities were destroyed",
				m.pruneDestroyedCities(),
			),
		)
	}()

	// For each random city, attempt to add an invader,
	// and kick off the invasion process for that alien
	for id, randomCity := range randomCities {
		// Attempt to add the alien as an invader
		if !randomCity.laySiege(id) {
			// The alien could not be added, because the city
			// is not accessible. The assumption is that aliens that cannot
			// be added to their initially assigned cities are not accounted for.
			// An alternative approach would be to grab a new random city for each alien
			// in this situation (reassign them to a new random city)
			aliensLeft--

			continue
		}

		randomCity.addInvader(id)

		wg.Add(1)

		// Start the alien run loop
		go func(ctx context.Context, id int, startingCity *city) {
			defer func() {
				wg.Done()
			}()

			newAlien(id).runAlien(
				workerContext,
				startingCity,
				alienDoneCh,
			)
		}(workerContext, id, randomCity)
	}

	// Wait until the program terminates
	for {
		select {
		case <-ctx.Done():
			// User stopped the program
			m.log.Info("Shutdown signal caught...")

			return
		case <-alienDoneCh:
			aliensLeft--

			if aliensLeft == 0 {
				m.log.Info("The final alien has finished")

				return
			}
		}
	}
}

// getRandomCities fetches random cities from the earth map
func (m *EarthMap) getRandomCities(numCities int) []*city {
	// Seed the random number generator
	rand.Seed(time.Now().UnixNano())

	// Gather the cities (keys)
	var (
		totalCities = len(m.cityMap)
		cities      = make([]string, totalCities)
		index       = 0
	)

	for city := range m.cityMap {
		cities[index] = city
		index++
	}

	// Randomly distribute the cities
	randomCities := make([]*city, numCities)
	for i := 0; i < numCities; i++ {
		//nolint:gosec
		randomCities[i] = m.cityMap[cities[rand.Intn(totalCities)]]
	}

	return randomCities
}

// pruneDestroyedCities removes destroyed cities from the earth map.
// Returns the number of pruned destroyed cities
func (m *EarthMap) pruneDestroyedCities() int {
	destroyed := 0
	for _, city := range m.cityMap {
		// Prune out any destroyed cities
		if city.destroyed {
			m.removeCity(city.name)
			destroyed++
		}
	}

	return destroyed
}
