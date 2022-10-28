package game

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/go-hclog"
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

// earthMap keeps track of all active Earth cities
type earthMap struct {
	log hclog.Logger

	cityMap map[string]*city
}

// newEarthMap creates a new instance of the earth map
func newEarthMap(log hclog.Logger) *earthMap {
	return &earthMap{
		log:     log,
		cityMap: make(map[string]*city),
	}
}

// initMap initializes the city map using the specified reader
func (m *earthMap) initMap(reader inputReader) {
	directions := []direction{north, south, east, west}

	// Read each city from the input stream, until it is depleted
	for reader.hasMoreCities() {
		cityLine := reader.readCity()

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
		city := newCity(cityNameMatch[0])

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
func (m *earthMap) getCity(name string) *city {
	return m.cityMap[name]
}

// addCity appends a city to the city map
func (m *earthMap) addCity(newCity *city) {
	m.cityMap[newCity.name] = newCity
}

// removeCity removes the city from the city map
func (m *earthMap) removeCity(name string) {
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
func (m *earthMap) getOrAddCity(name string) *city {
	city := m.getCity(name)

	if city == nil {
		// City not created yet, add it
		city = newCity(name)

		m.addCity(city)
	}

	return city
}

// writeOutput writes the current map layout to the specified
// output stream. It assumes that the output order is not important
func (m *earthMap) writeOutput(writer outputWriter) error {
	// Each city has an output format:
	// CityName direction=CityName...
	for _, city := range m.cityMap {
		var sb strings.Builder

		// Write the city name
		sb.WriteString(fmt.Sprintf("%s ", city.name))

		// For each direction, write the neighbor with the direction
		for direction, neighbor := range city.neighbors {
			sb.WriteString(
				fmt.Sprintf(
					"%s=%s",
					direction.getName(),
					neighbor.name,
				),
			)
		}

		if err := writer.write(fmt.Sprintf("%s\n", sb.String())); err != nil {
			return fmt.Errorf("unable to write to output stream, %w", err)
		}
	}

	return nil
}
