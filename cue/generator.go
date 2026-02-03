package cue

import (
	"fmt"
	"path/filepath"

	"github.com/b0bbywan/go-disc-cuer/config"
	"github.com/b0bbywan/go-disc-cuer/logger"
	"github.com/b0bbywan/go-disc-cuer/types"
	"github.com/b0bbywan/go-disc-cuer/utils"
)

// GenerateFromDefaultDisc generates a CUE file from the default disc device
//
// Parameters:
//   - cuerConfig: Application configuration
//
// Returns:
//   - string: Path to the generated CUE file
//   - error: Any error encountered
func GenerateFromDefaultDisc(cuerConfig *config.Config) (string, error) {
	opts := &GenerationOptions{
		Device: cuerConfig.Device,
		Config: cuerConfig,
	}
	return Generate(opts)
}

// GenerateDefaultFromDisc generates a CUE file from a specific disc device
//
// Parameters:
//   - device: CD-ROM device path
//   - cuerConfig: Application configuration
//
// Returns:
//   - string: Path to the generated CUE file
//   - error: Any error encountered
func GenerateDefaultFromDisc(device string, cuerConfig *config.Config) (string, error) {
	opts := &GenerationOptions{
		Device: device,
		Config: cuerConfig,
	}
	return Generate(opts)
}

// GenerateWithOptions generates a CUE file with custom options
//
// Parameters:
//   - device: CD-ROM device path
//   - cuerConfig: Application configuration
//   - providedDiscID: Optional user-supplied disc ID
//   - musicbrainzID: Optional MusicBrainz release ID
//   - overwrite: Force regeneration if file exists
//
// Returns:
//   - string: Path to the generated CUE file
//   - error: Any error encountered
func GenerateWithOptions(device string, cuerConfig *config.Config, providedDiscID, musicbrainzID string, overwrite bool) (string, error) {
	opts := &GenerationOptions{
		Device:         device,
		Config:         cuerConfig,
		ProvidedDiscID: providedDiscID,
		MusicBrainzID:  musicbrainzID,
		Overwrite:      overwrite,
	}
	return Generate(opts)
}

// Generate is the main CUE file generation orchestrator
//
// Workflow:
//  1. Validate options
//  2. Determine disc data (from flags or physical disc)
//  3. Check cache (return early if exists and overwrite=false)
//  4. Fetch metadata (from MusicBrainz ID or concurrent GNUDB+MB)
//  5. Fetch cover art (if MusicBrainz ID available)
//  6. Write CUE file
//
// Parameters:
//   - opts: Generation options
//
// Returns:
//   - string: Path to the generated CUE file
//   - error: Any error encountered
func Generate(opts *GenerationOptions) (string, error) {
	logger.Debugf("Starting CUE generation with device=%s, providedDiscID=%s, musicbrainzID=%s, overwrite=%v",
		opts.Device, opts.ProvidedDiscID, opts.MusicBrainzID, opts.Overwrite)

	// Step 1: Validate options
	if err := validateOptions(opts); err != nil {
		return "", err
	}

	// Step 2: Determine disc data and metadata source
	discData, discInfo, err := determineDiscDataAndMetadata(opts)
	if err != nil {
		return "", err
	}

	// Step 3: Check cache
	cacheLocation := opts.Config.GetCacheLocation()
	cueFilePath := utils.CachePlaylistPath(cacheLocation, discData.DiscID)
	logger.Debugf("CUE file path: %s", cueFilePath)

	if utils.CheckIfPlaylistExists(cueFilePath) && !opts.Overwrite {
		logger.Infof("CUE file already exists: %s (use -overwrite to regenerate)", cueFilePath)
		return cueFilePath, nil
	}

	// Step 4: Ensure cache directory exists
	if err := utils.CreateFolderIfNeeded(cueFilePath); err != nil {
		logger.Errorf("Failed to create folder for %s: %v", cueFilePath, err)
		return "", fmt.Errorf("failed to create cache folder: %w", err)
	}

	// Step 5: Fetch metadata if not already available
	if discInfo == nil {
		logger.Infof("Fetching disc metadata from GNUDB and MusicBrainz...")
		discInfo, err = fetchMetadataConcurrently(opts.Config, discData)
		if err != nil {
			logger.Errorf("Failed to get disc metadata: %v", err)
			return "", fmt.Errorf("failed to get disc metadata: %w", err)
		}
		logger.Infof("Successfully fetched disc metadata: %s - %s", discInfo.Artist, discInfo.Title)
	}

	// Step 6: Fetch cover art
	coverFilePath := utils.CacheCoverArtPath(cacheLocation, discData.DiscID)
	if err := fetchAndAttachCoverArt(discInfo, coverFilePath); err != nil {
		// Non-critical error, just log it
		logger.Debugf("Cover art fetch failed (non-critical): %v", err)
	}

	// Step 7: Write CUE file
	if err := writeCueFile(discInfo, cueFilePath); err != nil {
		logger.Errorf("Failed to write CUE file %s: %v", cueFilePath, err)
		return "", fmt.Errorf("failed to write CUE file: %w", err)
	}

	logger.Infof("Playlist successfully generated at: %s", cueFilePath)
	return cueFilePath, nil
}

// validateOptions validates generation options
func validateOptions(opts *GenerationOptions) error {
	if opts.Config == nil {
		return fmt.Errorf("config is required")
	}

	// --disc-id requires --musicbrainz
	if opts.ProvidedDiscID != "" && opts.MusicBrainzID == "" {
		return fmt.Errorf("--disc-id option requires --musicbrainz to be set")
	}

	return nil
}

// determineDiscDataAndMetadata determines disc data and optionally fetches metadata
//
// Returns:
//   - *DiscData: Disc identifiers and TOCs
//   - *types.DiscInfo: Disc metadata (if fetched from MusicBrainz ID)
//   - error: Any error encountered
func determineDiscDataAndMetadata(opts *GenerationOptions) (*DiscData, *types.DiscInfo, error) {
	var discData *DiscData
	var discInfo *types.DiscInfo
	var err error

	// Case 1: MusicBrainz ID provided
	if opts.MusicBrainzID != "" {
		discInfo, err = fetchMetadataFromMusicBrainzID(opts.MusicBrainzID)
		if err != nil {
			return nil, nil, err
		}

		// Use provided disc ID or generate one from physical disc
		if opts.ProvidedDiscID != "" {
			logger.Debugf("Using provided disc ID: %s", opts.ProvidedDiscID)
			discData = &DiscData{DiscID: opts.ProvidedDiscID}
		} else {
			// Still need to read disc to get the ID
			discData, err = readDiscData(opts.Device)
			if err != nil {
				return nil, nil, err
			}
		}

		return discData, discInfo, nil
	}

	// Case 2: No MusicBrainz ID - read from physical disc
	discData, err = readDiscData(opts.Device)
	if err != nil {
		return nil, nil, err
	}

	return discData, nil, nil
}
