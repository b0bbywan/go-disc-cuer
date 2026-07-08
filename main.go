// Command disc-cuer generates a CUE sheet for an audio CD from MusicBrainz and
// GNUDB metadata.
package main

import (
	"flag"
	"log"

	"github.com/b0bbywan/go-disc-cuer/config"
	"github.com/b0bbywan/go-disc-cuer/cue"
)

var (
	overwrite      bool
	musicbrainzID  string
	providedDiscID string
	deviceFlag     string
)

func init() {
	flag.BoolVar(&overwrite, "overwrite", false, "regenerate the CUE file even if it exists")
	flag.StringVar(&musicbrainzID, "musicbrainz", "", "MusicBrainz release id")
	flag.StringVar(&providedDiscID, "disc-id", "", "disc id (requires -musicbrainz)")
	flag.StringVar(&deviceFlag, "device", "", "disc device")
}

func getDevice(device string, cfg *config.Config) string {
	if device != "" {
		return device
	}
	return cfg.Device
}

func main() {
	flag.Parse()

	cfg, err := config.NewDefaultConfig()
	if err != nil {
		log.Fatalf("error: failed to initialize %s config: %v", config.AppName, err)
	}

	if _, err = cue.New(cfg).Generate(cue.Options{
		Device:        getDevice(deviceFlag, cfg),
		DiscID:        providedDiscID,
		MusicBrainzID: musicbrainzID,
		Overwrite:     overwrite,
	}); err != nil {
		log.Fatalf("error: failed to generate playlist: %v", err)
	}
}
