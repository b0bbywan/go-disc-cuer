package utils

import (
	"fmt"
	"log"
	"strings"

	"go.uploadedlobster.com/discid"
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
func GetTocAndDiscID(disc *discid.Disc) (string, string, error) {
	if disc == nil {
		return "", "", fmt.Errorf("no disc provided")
	}
	gnuToc, err := tocToGnu(disc)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate GNU TOC: %w", err)
	}

	discID := disc.FreedbID()
	log.Printf("GNU TOC: %s", gnuToc)
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
func GetMusicBrainzTOC(disc *discid.Disc) (string, error) {
	if disc == nil {
		return "", fmt.Errorf("no disc provided")
	}
	mbToc := disc.TOCString()
	log.Printf("MusicBrainz TOC: %s", mbToc)
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
func tocToGnu(disc *discid.Disc) (string, error) {
	if disc == nil {
		return "", fmt.Errorf("no disc provided")
	}
	// Get FreeDB ID
	freedbID := disc.FreedbID()
	// Get the number of tracks
	trackCount := disc.LastTrackNumber()

	// Collect track offsets
	offsets := []string{freedbID, fmt.Sprintf("%d", trackCount)}
	for i := 1; i <= trackCount; i++ {
		track, err := disc.Track(i)
		if err != nil {
			return "", err
		}
		offsets = append(offsets, fmt.Sprintf("%d", track.Offset))
	}

	// Append the disc duration as an integer
	offsets = append(offsets, fmt.Sprintf("%d", int(disc.Duration().Seconds())))

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
	disc, err := discid.Read(device)
	if err != nil {
		return 0, err
	}
	defer disc.Close()
	return disc.LastTrackNumber(), nil
}
