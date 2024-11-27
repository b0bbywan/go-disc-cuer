// The main package handles command-line flags and orchestrates the generation of CUE files,
// either from MusicBrainz release IDs or directly provided disc IDs, with an option to overwrite
// existing files.
package main

import (
	"flag"
	"log"

	"github.com/b0bbywan/go-disc-cuer/config"
	"github.com/b0bbywan/go-disc-cuer/cue"
)

// Command-line flags
var (
	// overwrite specifies whether to force regenerate the CUE file even if it exists.
	overwrite      bool

	// musicbrainzID specifies the MusicBrainz release ID for fetching the release information.	
	musicbrainzID  string

	// providedDiscID specifies the disc ID to be used for generating the CUE file.
	providedDiscID string

)

// init initializes the command-line flags and their descriptions.
func init() {
	// -overwrite flag to force CUE file regeneration
	flag.BoolVar(&overwrite, "overwrite", false, "force regenerating the CUE file even if it exists")

	// -musicbrainz flag to specify the MusicBrainz release ID
	flag.StringVar(&musicbrainzID, "musicbrainz", "", "specify MusicBrainz release ID directly")

	// -disc-id flag to specify a direct disc ID
	flag.StringVar(&providedDiscID, "disc-id", "", "specify disc ID directly")
}

// main is the entry point for the program. It parses the flags and generates a CUE file
// based on the provided MusicBrainz ID, disc ID, and overwrite flag.
func main() {
	flag.Parse()
	cuerConfig, err := config.NewDefaultConfig()
	if err != nil {
		log.Fatalf("error: Failed to initialize %s config: %v", config.AppName, err)
	}
	if _, err = cue.GenerateWithOptions(cuerConfig.Device, cuerConfig, musicbrainzID, providedDiscID, overwrite); err != nil {
		log.Fatalf("error: Failed to generate playlist from both GNUDB and MusicBrainz: %v", err)
	}
}
