package cue

import (
	"fmt"
	"strings"

	"go.uploadedlobster.com/discid"

	"github.com/b0bbywan/go-disc-cuer/logger"
	"github.com/b0bbywan/go-disc-cuer/utils"
)

// readDiscData reads disc data from a physical device and returns DiscData
//
// Parameters:
//   - device: Path to the CD-ROM device
//
// Returns:
//   - *DiscData: Disc identifiers and TOC information
//   - error: Any error encountered during disc reading
func readDiscData(device string) (*DiscData, error) {
	logger.Debugf("Reading disc from device: %s", device)

	disc, err := discid.Read(device)
	if err != nil {
		logger.Errorf("Failed to read disc from device %s: %v", device, err)
		return nil, fmt.Errorf("failed to read disc: %w", err)
	}
	defer disc.Close()

	// Get GNU TOC and FreeDB disc ID
	gnuToc, discID, err := utils.GetTocAndDiscID(disc)
	if err != nil {
		logger.Errorf("Failed to get TOC and disc ID: %v", err)
		return nil, fmt.Errorf("failed to get TOC: %w", err)
	}

	// Get MusicBrainz TOC
	mbToc, err := utils.GetMusicBrainzTOC(disc)
	if err != nil {
		logger.Errorf("Failed to get MusicBrainz TOC: %v", err)
		return nil, fmt.Errorf("failed to get MusicBrainz TOC: %w", err)
	}

	logger.Infof("Disc ID calculated: %s", discID)
	logger.Debugf("GNU TOC: %s", gnuToc)
	logger.Debugf("MusicBrainz TOC: %s", mbToc)

	return &DiscData{
		DiscID:         discID,
		GnuTOC:         strings.ReplaceAll(gnuToc, " ", "+"),
		MusicBrainzTOC: strings.ReplaceAll(mbToc, " ", "+"),
	}, nil
}
