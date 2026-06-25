package musicbrainz

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/b0bbywan/go-disc-cuer/types"
)

// mbURL is the MusicBrainz web service base. It is a var rather than a const so
// tests can redirect it to a local stub server.
var mbURL = "https://musicbrainz.org/ws/2"

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
	var release types.MBRelease
	if err := fetchJSON(url, &release); err != nil {
		return nil, err
	}
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
	var result types.ReleaseResult
	if err := fetchJSON(url, &result); err != nil {
		return nil, err
	}

	if len(result.Releases) == 0 {
		return nil, errors.New("no release data found")
	}

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
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error: failed to fetch from URL %s, status code: %d", url, resp.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(target)
}
