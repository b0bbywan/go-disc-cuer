package cue

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/b0bbywan/go-disc-cuer/config"
	"github.com/b0bbywan/go-disc-cuer/gnudb"
	"github.com/b0bbywan/go-disc-cuer/musicbrainz"
	"github.com/b0bbywan/go-disc-cuer/types"
	"github.com/b0bbywan/go-disc-cuer/utils"

)

const (
	coverArtURL   = "https://coverartarchive.org/release"
)

// fetchCoverArtIfNeeded ensures that cover art is available for the given disc.
// If the cover art is missing, it attempts to fetch it from the Cover Art Archive
// based on the MusicBrainz ID of the disc and saves it in the appropriate cache folder.
//
// Parameters:
//   - discInfo (*types.DiscInfo): Metadata for the disc, including its MusicBrainz ID and cover art path.
//   - cueFilePath (string): The path to the CUE file, used to determine the cache directory.
//
// Returns:
//   - error: An error if the cover art cannot be fetched or saved; nil otherwise.
func fetchCoverArtIfNeeded(discInfo *types.DiscInfo, cacheLocation, cueFilePath string) error {
    if discInfo.CoverArtPath == "" {
        coverFilePath := utils.CacheCoverArtPath(cacheLocation, filepath.Base(filepath.Dir(cueFilePath)))
        if err := fetchCoverArt(discInfo.ID, coverFilePath); err == nil {
            discInfo.CoverArtPath = coverFilePath
        } else {
            return fmt.Errorf("error getting cover: %v", err)
        }
    }
    return nil
}


// fetchCoverArt downloads cover art from the Cover Art Archive using a MusicBrainz ID.
//
// Parameters:
//   - mbID (string): The MusicBrainz release ID for the disc.
//   - coverFile (string): The file path where the cover art will be saved.
//
// Returns:
//   - error: An error if the HTTP request fails, the response status is not OK,
//            or the file cannot be saved; nil otherwise.
func fetchCoverArt(mbID, coverFile string) error {
	url := fmt.Sprintf("%s/%s/front", coverArtURL, mbID)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("failed to fetch cover art: received status code %d", resp.StatusCode)
	}

	file, err := os.Create(coverFile)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	return err
}

// fetchDiscInfoConcurrently fetches metadata about a disc from both GNUDB and MusicBrainz concurrently.
// This function uses goroutines and a WaitGroup to perform the operations in parallel.
//
// Parameters:
//   - gnuToc (string): The disc's TOC formatted for GNUDB queries.
//   - mbToc (string): The disc's TOC formatted for MusicBrainz queries.
//
// Returns:
//   - *types.DiscInfo: Consolidated metadata about the disc, prioritizing GNUDB data when available.
//   - error: An error if both sources fail to provide valid data; nil otherwise.
// Function to fetch disc info from both services using goroutines and WaitGroup
func fetchDiscInfoConcurrently(cuerConfig *config.Config, gnuToc, mbToc string) (*types.DiscInfo, error) {
	if cuerConfig == nil {
		return nil, fmt.Errorf("Failed to fetch disc info: empty config")
	}
	var wg sync.WaitGroup
	var gndbDiscInfo, mbDiscInfo *types.DiscInfo
	var gndbErr, mbErr error
	formattedGnuTOC := strings.ReplaceAll(gnuToc, " ", "+")
	formattedMBTOC := strings.ReplaceAll(mbToc, " ", "+")

	wg.Add(2)

	// Fetch from GNUDB
	go func() {
		defer wg.Done()
		gndbDiscInfo, gndbErr = gnudb.FetchDiscInfo(cuerConfig, formattedGnuTOC)
	}()

	// Fetch from MusicBrainz
	go func() {
		defer wg.Done()
		mbDiscInfo, mbErr = musicbrainz.FetchReleaseByToc(formattedMBTOC)
	}()

	// Wait for both fetches to complete
	wg.Wait()

	return selectDiscInfo(gndbDiscInfo, gndbErr, mbDiscInfo, mbErr)
}

// selectDiscInfo determines the final disc metadata to use based on the results from GNUDB and MusicBrainz.
//
// Parameters:
//   - gndbDiscInfo (*types.DiscInfo): Metadata fetched from GNUDB (if available).
//   - gndbErr (error): Any error encountered during the GNUDB fetch.
//   - mbDiscInfo (*types.DiscInfo): Metadata fetched from MusicBrainz (if available).
//   - mbErr (error): Any error encountered during the MusicBrainz fetch.
//
// Returns:
//   - *types.DiscInfo: The chosen disc metadata, prioritizing GNUDB data when both sources are successful.
//   - error: An error if both sources fail, containing details about both failures.
func selectDiscInfo(gndbDiscInfo *types.DiscInfo, gndbErr error, mbDiscInfo *types.DiscInfo, mbErr error) (*types.DiscInfo, error) {

	// Decide on the final discInfo, prioritizing GNUDB data where available
	finalDiscInfo := &types.DiscInfo{}
	if gndbErr == nil {
		*finalDiscInfo = *gndbDiscInfo
	} else if mbErr == nil {
		*finalDiscInfo = *mbDiscInfo
	}

	// Use MusicBrainz ID regardless of source priority
	if mbDiscInfo != nil {
		finalDiscInfo.ID = mbDiscInfo.ID
	}

	// If both failed, return an error
	if gndbErr != nil && mbErr != nil {
		return nil, fmt.Errorf("failed to fetch from both sources: GNUDB error: %v; MusicBrainz error: %v", gndbErr, mbErr)
	}

	return finalDiscInfo, nil
}
