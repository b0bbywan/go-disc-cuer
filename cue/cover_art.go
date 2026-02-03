package cue

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/b0bbywan/go-disc-cuer/logger"
	"github.com/b0bbywan/go-disc-cuer/types"
)

const coverArtURL = "https://coverartarchive.org/release"

// fetchAndAttachCoverArt attempts to fetch cover art and attach it to disc info
//
// Parameters:
//   - discInfo: Disc metadata to attach cover art path to
//   - coverFilePath: Path where cover art should be saved
//
// Returns:
//   - error: Error if cover art cannot be fetched (non-critical)
func fetchAndAttachCoverArt(discInfo *types.DiscInfo, coverFilePath string) error {
	if discInfo.ID == "" {
		logger.Debugf("No MusicBrainz ID available, skipping cover art fetch")
		return fmt.Errorf("no MusicBrainz ID available")
	}

	logger.Debugf("Attempting to fetch cover art for MusicBrainz ID: %s", discInfo.ID)

	err := downloadCoverArt(discInfo.ID, coverFilePath)
	if err != nil {
		logger.Warnf("Failed to fetch cover art: %v", err)
		return err
	}

	discInfo.CoverArtPath = coverFilePath
	logger.Infof("Cover art saved to: %s", coverFilePath)
	return nil
}

// downloadCoverArt downloads cover art from Cover Art Archive
//
// Parameters:
//   - mbID: MusicBrainz release ID
//   - targetFile: Local file path to save the cover art
//
// Returns:
//   - error: Error if download fails
func downloadCoverArt(mbID, targetFile string) error {
	url := fmt.Sprintf("%s/%s/front", coverArtURL, mbID)
	logger.Debugf("Fetching cover art from: %s", url)

	resp, err := http.Get(url)
	if err != nil {
		logger.Debugf("HTTP request failed for cover art: %v", err)
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		logger.Debugf("Cover art not found (status %d) for MusicBrainz ID: %s", resp.StatusCode, mbID)
		return fmt.Errorf("cover art not found (status %d)", resp.StatusCode)
	}

	logger.Debugf("Saving cover art to: %s", targetFile)
	file, err := os.Create(targetFile)
	if err != nil {
		logger.Errorf("Failed to create cover art file %s: %v", targetFile, err)
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	if _, err = io.Copy(file, resp.Body); err != nil {
		logger.Errorf("Failed to write cover art data: %v", err)
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
