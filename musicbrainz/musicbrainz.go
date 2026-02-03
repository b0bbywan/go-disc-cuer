package musicbrainz

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/b0bbywan/go-disc-cuer/logger"
	"github.com/b0bbywan/go-disc-cuer/types"
)

const (
	mbURL = "https://musicbrainz.org/ws/2"
)

// FetchReleaseByID fetches a MusicBrainz release's information based on its release ID.
//
// Parameters:
//   - releaseID (string): The MusicBrainz release ID (e.g., `ab123456-7890-1234-5678-abcdef123456`).
//
// Returns:
//   - *types.DiscInfo: A struct containing the release's metadata (artist, title, tracks, etc.).
//   - error: An error if the release data cannot be fetched or parsed.
func FetchReleaseByID(releaseID string) (*types.DiscInfo, error) {
	url := fmt.Sprintf("%s/release/%s?inc=artists+recordings&fmt=json", mbURL, releaseID)
	logger.Debugf("Fetching MusicBrainz release by ID: %s", releaseID)
	var release types.MBRelease
	if err := fetchJSON(url, &release); err != nil {
		logger.Errorf("Failed to fetch MusicBrainz release %s: %v", releaseID, err)
		return nil, err
	}
	logger.Debugf("Successfully fetched MusicBrainz release: %s", release.Title)
	return convertReleaseToDiscInfo(release)

}

// FetchReleaseByToc fetches a MusicBrainz release's information based on its TOC (Table of Contents).
//
// Parameters:
//   - mbToc (string): The TOC of the disc in MusicBrainz format (e.g., `12345678`).
//
// Returns:
//   - *types.DiscInfo: A struct containing the release's metadata (artist, title, tracks, etc.).
//   - error: An error if no release data is found or if the request fails.
func FetchReleaseByToc(mbToc string) (*types.DiscInfo, error) {
	url := fmt.Sprintf("%s/discid/-?toc=%s&inc=artists+recordings&fmt=json", mbURL, mbToc)
	logger.Debugf("Fetching MusicBrainz release by TOC: %s", mbToc)
	var result types.ReleaseResult
	if err := fetchJSON(url, &result); err != nil {
		logger.Debugf("Failed to fetch MusicBrainz data by TOC: %v", err)
		return nil, err
	}

	if len(result.Releases) == 0 {
		logger.Debugf("No MusicBrainz releases found for TOC: %s", mbToc)
		return nil, errors.New("no release data found")
	}

	logger.Debugf("Found %d MusicBrainz release(s), using the first one", len(result.Releases))
	release := result.Releases[0]
	return convertReleaseToDiscInfo(release)
}

// convertReleaseToDiscInfo converts a MusicBrainz release object to a DiscInfo object.
//
// Parameters:
//   - release (types.MBRelease): A MusicBrainz release object containing the metadata.
//
// Returns:
//   - *types.DiscInfo: A struct with the converted disc information (artist, title, release date, tracks).
//   - error: An error if any data is missing or cannot be converted.
func convertReleaseToDiscInfo(release types.MBRelease) (*types.DiscInfo, error) {
	tracks := make([]string, len(release.Media[0].Tracks))
	for i, track := range release.Media[0].Tracks {
		tracks[i] = track.Title
	}

	return &types.DiscInfo{
		ID:          release.ID,
		Title:       release.Title,
		Artist:      release.ArtistCredit[0].Name,
		ReleaseDate: release.Date,
		Tracks:      tracks,
	}, nil
}

// fetchJSON performs an HTTP GET request to fetch JSON data from a URL and decodes it into the target structure.
//
// Parameters:
//   - url (string): The URL to fetch the JSON data from.
//   - target (interface{}): A pointer to the target structure where the JSON response will be decoded.
//
// Returns:
//   - error: An error if the request fails or if the response cannot be parsed.
func fetchJSON(url string, target interface{}) error {
	logger.Debugf("Making HTTP request to: %s", url)
	resp, err := http.Get(url)
	if err != nil {
		logger.Debugf("HTTP request failed: %v", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.Debugf("HTTP request returned status code: %d", resp.StatusCode)
		return fmt.Errorf("error: failed to fetch from URL %s, status code: %d", url, resp.StatusCode)
	}

	logger.Debugf("Successfully received response, decoding JSON...")
	return json.NewDecoder(resp.Body).Decode(target)
}
