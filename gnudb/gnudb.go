package gnudb

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/b0bbywan/go-disc-cuer/config"
	"github.com/b0bbywan/go-disc-cuer/logger"
	"github.com/b0bbywan/go-disc-cuer/types"
)

const (
	// Response keys for GNUDB
	keyTitle = "DTITLE="
	keyYear  = "DYEAR="
	keyGenre = "DGENRE="
	keyTrack = "TTITLE"
)

type gnuConfig struct {
	GnuHello string
	GnudbURL string
}

// newGnuConfig initializes the GNUDB configuration based on the provided application configuration.
//
// Parameters:
//   - cuerConfig: The Config instance containing application settings.
//
// Returns:
//   - *gnuConfig: A GNUDB-specific configuration instance.
//   - error: Any error encountered during initialization.
func newGnuConfig(cuerConfig *config.Config) (*gnuConfig, error) {
	if cuerConfig.GnuHelloEmail == "" {
		return nil, fmt.Errorf("gnuHelloEmail is required in config.yaml or via environment variable to use gnuDB")
	}
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown-host"
	}
	gnuHello := fmt.Sprintf("%s+%s+%s+%s", cuerConfig.GnuHelloEmail, hostname, cuerConfig.AppName, cuerConfig.AppVersion)
	gnudbURL := fmt.Sprintf("%s/~cddb/cddb.cgi", cuerConfig.GnuDbUrl)
	return &gnuConfig{
		GnuHello: gnuHello,
		GnudbURL: gnudbURL,
	}, nil
}

// FetchDiscInfo queries GNUDB to retrieve metadata about a disc.
//
// Parameters:
//   - cuerConfig: The Config instance containing GNUDB settings.
//   - gnuToc: The table of contents (TOC) of the disc.
//
// Returns:
//   - *types.DiscInfo: Metadata about the disc.
//   - error: Any error encountered during the operation.
func FetchDiscInfo(cuerConfig *config.Config, gnuToc string) (*types.DiscInfo, error) {
	logger.Debugf("Fetching disc info from GNUDB with TOC: %s", gnuToc)
	gnuConfig, err := newGnuConfig(cuerConfig)
	if err != nil {
		logger.Errorf("Failed to initialize GNUDB config: %v", err)
		return nil, fmt.Errorf("Invalid GNUConfig: %w", err)
	}
	logger.Debugf("GNUDB config initialized: %s", gnuConfig.GnuHello)
	client := &http.Client{}

	// First, query GNUDB for a match
	gnudbID, err := queryGNUDB(client, gnuConfig, gnuToc)
	if err != nil {
		logger.Debugf("GNUDB query failed for TOC %s: %v", gnuToc, err)
		return nil, fmt.Errorf("Failed to query %s on gnuDB: %w", gnuToc, err)
	}
	logger.Debugf("Found GNUDB ID: %s", gnudbID)

	// Fetch the full metadata from GNDB
	discInfo, err := fetchFullMetadata(client, gnuConfig, gnudbID)
	if err != nil {
		logger.Errorf("Failed to fetch GNUDB metadata for ID %s: %v", gnudbID, err)
		return nil, fmt.Errorf("Failed to fetch %s (%s) metadata on gnuDB: %w", gnudbID, gnuToc, err)
	}
	logger.Infof("Successfully fetched GNUDB metadata: %s - %s", discInfo.Artist, discInfo.Title)

	return discInfo, nil
}

// queryGNUDB performs the initial query to GNUDB to find a matching record for the given TOC.
//
// Parameters:
//   - client (*http.Client): HTTP client for making requests.
//   - gnuConfig (*gnuConfig): The gnuConfig instance containing GNUDB settings.
//   - gnuToc (string): The disc's TOC, formatted for GNUDB queries.
//
// Returns:
//   - string: The GNUDB ID of the matching record.
//   - error: An error if the query fails, the response cannot be read, or no match is found.
func queryGNUDB(client *http.Client, gnuConfig *gnuConfig, gnuToc string) (string, error) {
	if gnuConfig == nil {
		return "", fmt.Errorf("Failed to query gnudb: empty config")
	}
	queryURL := fmt.Sprintf("%s?cmd=cddb+query+%s&hello=%s&proto=6", gnuConfig.GnudbURL, gnuToc, gnuConfig.GnuHello)
	logger.Debugf("Querying GNUDB: %s", queryURL)
	resp, err := makeGnuRequest(client, queryURL)
	if err != nil {
		logger.Debugf("GNUDB request failed: %v", err)
		return "", fmt.Errorf("Failed GnuRequest (%s): %w", queryURL, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Debugf("Failed to read GNUDB response body: %v", err)
		return "", fmt.Errorf("Failed to read response body: %w", err)
	}
	logger.Debugf("GNUDB query response: %s", string(body))
	if !strings.Contains(string(body), "Found exact matches") {
		logger.Debugf("No exact match found in GNUDB response")
		return "", fmt.Errorf("No exact match found in GNUDB: %s", string(body))
	}
	return extractGnuDBID(string(body))
}

// extractGnuDBID extracts the GNUDB ID from a successful query response.
//
// Parameters:
//   - response (string): The raw response string from the GNUDB query.
//
// Returns:
//   - string: The extracted GNUDB ID.
//   - error: An error if the response format is invalid.
func extractGnuDBID(response string) (string, error) {
	lines := strings.Split(response, "\n")
	if len(lines) < 2 {
		return "", fmt.Errorf("invalid GNUDB response format: %s", response)
	}
	return strings.Fields(lines[1])[1], nil
}

// fetchFullMetadata retrieves detailed disc metadata from GNUDB using the record's ID.
//
// Parameters:
//   - client (*http.Client): HTTP client for making requests.
//   - gnuConfig (*gnuConfig): The gnuConfig instance containing GNUDB settings.
//   - gnudbID (string): The ID of the record in GNUDB.
//
// Returns:
//   - *types.DiscInfo: A struct containing the disc's metadata (artist, title, tracks, etc.).
//   - error: An error if the metadata cannot be retrieved or parsed.
func fetchFullMetadata(client *http.Client, gnuConfig *gnuConfig, gnudbID string) (*types.DiscInfo, error) {
	if gnuConfig == nil {
		return nil, fmt.Errorf("Failed to fetch gnudb metadata: empty config")
	}
	readURL := fmt.Sprintf("%s?cmd=cddb+read+data+%s&hello=%s&proto=6", gnuConfig.GnudbURL, gnudbID, gnuConfig.GnuHello)
	logger.Debugf("Fetching full metadata from GNUDB: %s", readURL)
	resp, err := makeGnuRequest(client, readURL)
	if err != nil {
		logger.Debugf("GNUDB metadata request failed: %v", err)
		return nil, fmt.Errorf("Failed GnuRequest (%s): %w", readURL, err)
	}
	defer resp.Body.Close()

	logger.Debugf("Parsing GNUDB response...")
	return parseGNUDBResponse(resp.Body)
}

// parseGNUDBResponse parses the response from a detailed metadata request to GNUDB.
//
// Parameters:
//   - body (io.Reader): The response body from the GNUDB read command.
//
// Returns:
//   - *types.DiscInfo: A struct containing the parsed disc metadata.
//   - error: An error if parsing fails or if required fields (e.g., title) are missing.
func parseGNUDBResponse(body io.Reader) (*types.DiscInfo, error) {
	scanner := bufio.NewScanner(body)
	discInfo := &types.DiscInfo{}
	trackCount := 0

	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, keyTitle):
			titleLine := strings.TrimPrefix(line, keyTitle)
			parts := strings.SplitN(titleLine, " / ", 2)
			if len(parts) == 2 {
				discInfo.Artist, discInfo.Title = parts[0], parts[1]
				logger.Debugf("Parsed GNUDB title: %s - %s", discInfo.Artist, discInfo.Title)
			}
		case strings.HasPrefix(line, keyYear):
			discInfo.ReleaseDate = strings.TrimPrefix(line, keyYear)
			logger.Debugf("Parsed GNUDB year: %s", discInfo.ReleaseDate)
		case strings.HasPrefix(line, keyGenre):
			discInfo.Genre = strings.TrimPrefix(line, keyGenre)
			logger.Debugf("Parsed GNUDB genre: %s", discInfo.Genre)
		case strings.HasPrefix(line, keyTrack):
			track := strings.SplitN(line, "=", 2)
			discInfo.Tracks = append(discInfo.Tracks, track[1])
			trackCount++
		}
	}

	logger.Debugf("Parsed %d tracks from GNUDB", trackCount)

	if err := scanner.Err(); err != nil {
		logger.Errorf("Failed to scan GNUDB response body: %v", err)
		return nil, fmt.Errorf("Failed to scan body: %w", err)
	}

	if discInfo.Title == "" {
		logger.Errorf("No valid title found in GNUDB data")
		return nil, fmt.Errorf("error: no valid title in GNUDB data")
	}

	return discInfo, nil
}

// makeGnuRequest performs an HTTP GET request with a predefined User-Agent header.
//
// Parameters:
//   - client (*http.Client): HTTP client for making requests.
//   - url (string): The URL to send the GET request to.
//
// Returns:
//   - *http.Response: The HTTP response object.
//   - error: An error if the request fails.
func makeGnuRequest(client *http.Client, url string) (*http.Response, error) {
	userAgent := "curl/8.9.1"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)

	return client.Do(req)
}
