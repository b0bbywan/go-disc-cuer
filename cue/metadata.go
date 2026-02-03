package cue

import (
	"fmt"
	"sync"

	"github.com/b0bbywan/go-disc-cuer/config"
	"github.com/b0bbywan/go-disc-cuer/gnudb"
	"github.com/b0bbywan/go-disc-cuer/logger"
	"github.com/b0bbywan/go-disc-cuer/musicbrainz"
	"github.com/b0bbywan/go-disc-cuer/types"
)

// fetchMetadataFromMusicBrainzID fetches disc info using a MusicBrainz release ID
//
// Parameters:
//   - musicBrainzID: The MusicBrainz release ID
//
// Returns:
//   - *types.DiscInfo: Disc metadata
//   - error: Any error encountered
func fetchMetadataFromMusicBrainzID(musicBrainzID string) (*types.DiscInfo, error) {
	logger.Infof("Fetching release info from MusicBrainz ID: %s", musicBrainzID)

	discInfo, err := musicbrainz.FetchReleaseByID(musicBrainzID)
	if err != nil {
		logger.Errorf("Failed to fetch MusicBrainz release %s: %v", musicBrainzID, err)
		return nil, fmt.Errorf("failed to fetch MusicBrainz release: %w", err)
	}

	logger.Infof("Successfully fetched MusicBrainz release: %s - %s", discInfo.Artist, discInfo.Title)
	return discInfo, nil
}

// fetchMetadataConcurrently fetches metadata from both GNUDB and MusicBrainz concurrently
//
// Parameters:
//   - cfg: Application configuration
//   - discData: Disc TOC data for both services
//
// Returns:
//   - *types.DiscInfo: Consolidated disc metadata
//   - error: Error if both sources fail
func fetchMetadataConcurrently(cfg *config.Config, discData *DiscData) (*types.DiscInfo, error) {
	logger.Infof("Fetching disc metadata from GNUDB and MusicBrainz...")
	logger.Debugf("Using GNU TOC: %s, MusicBrainz TOC: %s", discData.GnuTOC, discData.MusicBrainzTOC)

	var wg sync.WaitGroup
	var gnudbInfo, mbInfo *types.DiscInfo
	var gnudbErr, mbErr error

	wg.Add(2)

	// Fetch from GNUDB
	go func() {
		defer wg.Done()
		logger.Debugf("Starting GNUDB fetch...")
		gnudbInfo, gnudbErr = gnudb.FetchDiscInfo(cfg, discData.GnuTOC)
		if gnudbErr != nil {
			logger.Debugf("GNUDB fetch failed: %v", gnudbErr)
		} else {
			logger.Debugf("GNUDB fetch successful: %s - %s", gnudbInfo.Artist, gnudbInfo.Title)
		}
	}()

	// Fetch from MusicBrainz
	go func() {
		defer wg.Done()
		logger.Debugf("Starting MusicBrainz fetch...")
		mbInfo, mbErr = musicbrainz.FetchReleaseByToc(discData.MusicBrainzTOC)
		if mbErr != nil {
			logger.Debugf("MusicBrainz fetch failed: %v", mbErr)
		} else {
			logger.Debugf("MusicBrainz fetch successful: %s - %s", mbInfo.Artist, mbInfo.Title)
		}
	}()

	wg.Wait()

	return mergeDiscInfo(gnudbInfo, gnudbErr, mbInfo, mbErr)
}

// mergeDiscInfo selects and merges disc info from GNUDB and MusicBrainz
//
// Strategy:
//   - Prioritizes GNUDB data for artist, title, tracks
//   - Always uses MusicBrainz ID if available (for cover art)
//
// Parameters:
//   - gnudbInfo: Metadata from GNUDB
//   - gnudbErr: Error from GNUDB fetch
//   - mbInfo: Metadata from MusicBrainz
//   - mbErr: Error from MusicBrainz fetch
//
// Returns:
//   - *types.DiscInfo: Merged metadata
//   - error: Error if both sources failed
func mergeDiscInfo(gnudbInfo *types.DiscInfo, gnudbErr error, mbInfo *types.DiscInfo, mbErr error) (*types.DiscInfo, error) {
	// If both failed, return error
	if gnudbErr != nil && mbErr != nil {
		logger.Errorf("Both GNUDB and MusicBrainz fetches failed")
		return nil, fmt.Errorf("failed to fetch from both sources: GNUDB: %w; MusicBrainz: %w", gnudbErr, mbErr)
	}

	finalInfo := &types.DiscInfo{}

	// Prioritize GNUDB data
	if gnudbErr == nil {
		*finalInfo = *gnudbInfo
		logger.Infof("Using GNUDB metadata: %s - %s", gnudbInfo.Artist, gnudbInfo.Title)
	} else if mbErr == nil {
		*finalInfo = *mbInfo
		logger.Infof("Using MusicBrainz metadata: %s - %s", mbInfo.Artist, mbInfo.Title)
	}

	// Always use MusicBrainz ID if available (needed for cover art)
	if mbInfo != nil && mbInfo.ID != "" {
		finalInfo.ID = mbInfo.ID
		logger.Debugf("Using MusicBrainz ID: %s", mbInfo.ID)
	}

	return finalInfo, nil
}
