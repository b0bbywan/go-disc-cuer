package utils

import (
	"fmt"
	"strings"

	"go.uploadedlobster.com/discid"

	"github.com/b0bbywan/go-disc-cuer/logger"
)

// GetTocAndDiscID takes a disc object and returns the corresponding GNU TOC string, MusicBrainz disc ID, and any errors encountered.
//
// Parameters:
//   - disc (discid.Disc): The disc object containing the TOC and disc ID information.
//
// Returns:
//   - gnuToc (string): The generated GNU TOC string for the disc.
//   - discID (string): The FreeDB ID for the disc.
//   - error: Any error encountered during the process.
func GetTocAndDiscID(disc discid.Disc) (string, string, error) {
	logger.Debugf("Generating GNU TOC and disc ID from disc data")
	gnuToc, err := tocToGnu(disc)
	if err != nil {
		logger.Errorf("Failed to generate GNU TOC: %v", err)
		return "", "", fmt.Errorf("failed to generate GNU TOC: %w", err)
	}

	discID := disc.FreedbID()
	logger.Debugf("GNU TOC: %s", gnuToc)
	logger.Debugf("Disc ID (FreeDB): %s", discID)
	return gnuToc, discID, nil
}

// GetMusicBrainzTOC retrieves the TOC string for MusicBrainz from the given disc object.
//
// Parameters:
//   - disc (discid.Disc): The disc object containing the MusicBrainz TOC information.
//
// Returns:
//   - mbToc (string): The MusicBrainz TOC string for the disc.
//   - error: Any error encountered during the process.
func GetMusicBrainzTOC(disc discid.Disc) (string, error) {
	logger.Debugf("Generating MusicBrainz TOC from disc data")
	mbToc := disc.TOCString()
	logger.Debugf("MusicBrainz TOC: %s", mbToc)
	return mbToc, nil
}

// tocToGnu generates the GNU TOC string from a disc object. This string is used for querying databases like FreeDB.
//
// Parameters:
//   - disc (discid.Disc): The disc object containing track and FreeDB ID information.
//
// Returns:
//   - gnuToc (string): The generated GNU TOC string.
//   - error: Any error encountered during the process.
func tocToGnu(disc discid.Disc) (string, error) {
	// Get FreeDB ID
	freedbID := disc.FreedbID()
	// Get the number of tracks
	trackCount := disc.LastTrackNumber()
	logger.Debugf("Building GNU TOC for disc with %d tracks (FreeDB ID: %s)", trackCount, freedbID)

	// Collect track offsets
	offsets := []string{freedbID, fmt.Sprintf("%d", trackCount)}
	for i := 1; i <= trackCount; i++ {
		track, err := disc.Track(i)
		if err != nil {
			logger.Errorf("Failed to get track %d: %v", i, err)
			return "", err
		}
		offsets = append(offsets, fmt.Sprintf("%d", track.Offset))
	}

	// Append the disc duration as an integer
	discDuration := int(disc.Duration().Seconds())
	offsets = append(offsets, fmt.Sprintf("%d", discDuration))
	logger.Debugf("Disc duration: %d seconds", discDuration)

	// Join the components with spaces
	return strings.Join(offsets, " "), nil
}

// GetTrackCount retrieves the number of tracks on a given disc object.
//
// Parameters:
//   - device (string): The path to the CD-ROM device.
//
// Returns:
//   - trackCount (int): The total number of tracks on the disc
//   - error :  Any error while opening the disc.
func GetTrackCount(device string) (int, error) {
	logger.Debugf("Reading track count from device: %s", device)
	disc, err := discid.Read(device)
	if err != nil {
		logger.Errorf("Failed to read disc from device %s: %v", device, err)
		return 0, err
	}
	defer disc.Close()
	trackCount := disc.LastTrackNumber()
	logger.Debugf("Disc has %d tracks", trackCount)
	return trackCount, nil
}
