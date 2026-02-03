// The main package handles command-line flags and orchestrates the generation of CUE files,
// either from MusicBrainz release IDs or directly provided disc IDs, with an option to overwrite
// existing files.
package main

import (
	"flag"

	"github.com/b0bbywan/go-disc-cuer/config"
	"github.com/b0bbywan/go-disc-cuer/cue"
	"github.com/b0bbywan/go-disc-cuer/logger"
)

// Command-line flags
var (
	// overwrite specifies whether to force regenerate the CUE file even if it exists.
	overwrite bool

	// musicbrainzID specifies the MusicBrainz release ID for fetching the release information.
	musicbrainzID string

	// providedDiscID specifies the disc ID to be used for generating the CUE file.
	providedDiscID string

	// deviceFlag specifies the drive to read data from
	deviceFlag string
)

// init initializes the command-line flags and their descriptions.
func init() {
	// -overwrite flag to force CUE file regeneration
	flag.BoolVar(&overwrite, "overwrite", false, "force regenerating the CUE file even if it exists")

	// -musicbrainz flag to specify the MusicBrainz release ID
	flag.StringVar(&musicbrainzID, "musicbrainz", "", "specify MusicBrainz release ID directly")

	// -disc-id flag to specify a direct disc ID
	flag.StringVar(&providedDiscID, "disc-id", "", "specify disc ID directly")

	flag.StringVar(&deviceFlag, "device", "", "Disc Device")
}

func getDevice(device string, cuerConfig *config.Config) string {
	if device != "" {
		return device
	}
	return cuerConfig.Device
}

// main is the entry point for the program. It parses the flags and generates a CUE file
// based on the provided MusicBrainz ID, disc ID, and overwrite flag.
func main() {
	flag.Parse()

	cuerConfig, err := config.NewDefaultConfig()
	if err != nil {
		logger.Fatalf("Failed to initialize %s config: %v", config.AppName, err)
	}

	// Initialize logger with configured level
	logger.SetLevel(logger.ParseLogLevel(cuerConfig.LogLevel))
	logger.Debugf("Starting %s version %s", cuerConfig.AppName, cuerConfig.AppVersion)
	logger.Debugf("Configuration loaded: LogLevel=%s, Device=%s, CacheLocation=%s",
		cuerConfig.LogLevel, cuerConfig.Device, cuerConfig.CacheLocation)

	device := getDevice(deviceFlag, cuerConfig)
	logger.Debugf("Using device: %s", device)

	if _, err = cue.GenerateWithOptions(device, cuerConfig, providedDiscID, musicbrainzID, overwrite); err != nil {
		logger.Fatalf("Failed to generate playlist: %v", err)
	}
}
