package cmd

// Define the present flags for the base program
const (
	mapPathFlag    = "map-path"
	outputPathFlag = "output-path"
	logLevelFlag   = "log-level"
)

var (
	params = rootParams{}
)

// rootParams defines the storage for the
// base program arguments
type rootParams struct {
	n          int
	mapPath    string
	outputPath string
	logLevel   string
}

// getRequiredFlags returns the required flags
func (r *rootParams) getRequiredFlags() []string {
	return []string{
		mapPathFlag,
	}
}
