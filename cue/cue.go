package cue

import (
	"fmt"
	"os"
	"path/filepath"

	"go.uploadedlobster.com/discid"

	"github.com/b0bbywan/go-disc-cuer/config"
	"github.com/b0bbywan/go-disc-cuer/logger"
	"github.com/b0bbywan/go-disc-cuer/musicbrainz"
	"github.com/b0bbywan/go-disc-cuer/types"
	"github.com/b0bbywan/go-disc-cuer/utils"
)

// GenerateFromDefaultDisc generates a CUE file for the currently inserted audio CD
// using the default behavior. It does not rely on any pre-provided disc ID or
// MusicBrainz release ID. This function assumes a disc is present and accessible
// in the drive. Use Device from config (defaut to "/dev/sr0")
//
// Parameters:
//   - cuerConfig: The Config instance to use for generating the CUE file.
//
// Returns:
//   - string: The path to the generated CUE file, or an existing file.
//   - error: Any error encountered during the process, such as failure to read the disc or generate the file.
func GenerateFromDefaultDisc(cuerConfig *config.Config) (string, error) {
	return generate(cuerConfig.Device, cuerConfig, "", "", false)
}

// GenerateFromDefaultDisc generates a CUE file for the currently inserted audio CD
// using the default behavior. It does not rely on any pre-provided disc ID or
// MusicBrainz release ID. This function assumes a disc is present and accessible
// in the given drive.
//
// Parameters:
//   - device: The path to the disc drive (e.g., "/dev/sr0").
//   - cuerConfig: The Config instance to use for generating the CUE file.
//
// Returns:
//   - string: The path to the generated CUE file, or an existing file.
//   - error: Any error encountered during the process, such as failure to read the disc or generate the file.
func GenerateDefaultFromDisc(device string, cuerConfig *config.Config) (string, error) {
	return generate(device, cuerConfig, "", "", false)
}

// GenerateWithOptions generates a CUE file with additional options, allowing the user
// to specify a disc ID or a MusicBrainz release ID, and control whether to overwrite
// existing CUE files.
//
// Parameters:
//   - device (string): The path to the CD-ROM device.
//   - cuerConfig: The Config instance to use for generating the CUE file.
//   - providedDiscID (string): A user-supplied disc ID to bypass detection. If empty,
//     the disc ID is determined automatically.
//   - musicbrainzID (string): A MusicBrainz release ID for fetching metadata. If empty,
//     GNUDB is used as the fallback metadata source.
//   - overwrite (bool): If true, forces regeneration of the CUE file even if it already exists.
//
// Returns:
//   - string: The path to the generated, or an existing file if overwrite is not set.
//   - error: Any error encountered during the process, such as metadata fetch or file write failure.
func GenerateWithOptions(device string, cuerConfig *config.Config, providedDiscID, musicbrainzID string, overwrite bool) (string, error) {
	return generate(device, cuerConfig, providedDiscID, musicbrainzID, overwrite)
}

// generate is the core function responsible for creating a CUE file. It handles
// disc ID calculation, metadata retrieval, and file creation or update.
//
// Parameters:
//   - device (string): The path to the CD-ROM device.
//   - cuerConfig: The Config instance to use for generating the CUE file.
//   - providedDiscID (string): A user-supplied disc ID (optional).
//   - musicbrainzID (string): A MusicBrainz release ID for metadata (optional).
//   - overwrite (bool): Whether to overwrite an existing CUE file.
//
// Returns:
//   - string: The path to the generated or updated CUE file.
//   - error: Any error encountered during the process.
//
// Workflow:
//  1. If a `providedDiscID` or `musicbrainzID` is provided, fetch corresponding disc info.
//  2. If `discID` is not determined, read the disc from the drive and compute its ID and TOC.
//  3. Check if a cached CUE file exists. If so, return it unless `overwrite` is true.
//  4. If `discInfo` and `discID` are both valid, finalize the CUE file generation.
//  5. If necessary, fetch metadata concurrently from GNUDB and MusicBrainz.
//  6. Ensure necessary directories exist, then create and save the CUE file.
//
// Notes:
// - This function is used internally by both `GenerateFromDisc` and `GenerateWithOptions`.
// - Fetching metadata from GNUDB and MusicBrainz occurs concurrently to improve efficiency.
//
// Returns:
//   - string: The path to the generated CUE file.
//   - error: Any error encountered during the operation.
func generate(device string, cuerConfig *config.Config, providedDiscID, musicbrainzID string, overwrite bool) (string, error) {
	logger.Debugf("Starting CUE generation with device=%s, providedDiscID=%s, musicbrainzID=%s, overwrite=%v",
		device, providedDiscID, musicbrainzID, overwrite)

	if cuerConfig == nil {
		return "", fmt.Errorf("Failed to generate cue file: empty config")
	}
	discInfo, discID, err := fetchDiscInfoFromFlags(providedDiscID, musicbrainzID)
	if err != nil {
		logger.Errorf("Failed to fetch disc info from flags: %v", err)
		return "", err
	}

	var disc discid.Disc
	var gnuToc string
	if discID == "" {
		logger.Debugf("Reading disc from device: %s", device)
		disc, err = discid.Read(device)
		if err != nil {
			logger.Errorf("Failed to read disc from device %s: %v", device, err)
			return "", err
		}
		defer disc.Close()
		if gnuToc, discID, err = utils.GetTocAndDiscID(disc); err != nil {
			logger.Errorf("Failed to get TOC and disc ID: %v", err)
			return "", err
		}
		logger.Infof("Disc ID calculated: %s", discID)
	} else {
		logger.Debugf("Using provided disc ID: %s", discID)
	}
	cacheLocation := cuerConfig.GetCacheLocation()
	cueFilePath := utils.CachePlaylistPath(cacheLocation, discID)
	logger.Debugf("CUE file path: %s", cueFilePath)

	if utils.CheckIfPlaylistExists(cueFilePath) && !overwrite {
		logger.Infof("CUE file already exists: %s (use -overwrite to regenerate)", cueFilePath)
		return cueFilePath, nil
	}

	if discInfo != nil && discID != "" {
		logger.Debugf("Disc info already available, finalizing CUE generation")
		return finalizeIfSuccess(discInfo, cacheLocation, cueFilePath)
	}
	var mbToc string
	if mbToc, err = utils.GetMusicBrainzTOC(disc); err != nil {
		logger.Errorf("Failed to get MusicBrainz TOC: %v", err)
		return "", fmt.Errorf("Failed to get musicbrainz TOC: %w", err)
	}
	logger.Debugf("MusicBrainz TOC: %s", mbToc)

	if err = utils.CreateFolderIfNeeded(cueFilePath); err != nil {
		logger.Errorf("Failed to create folder for %s: %v", cueFilePath, err)
		return "", fmt.Errorf("Failed to create %s folder: %w", cueFilePath, err)
	}

	// Fetch DiscInfo concurrently
	logger.Infof("Fetching disc metadata from GNUDB and MusicBrainz...")
	if discInfo, err = fetchDiscInfoConcurrently(cuerConfig, gnuToc, mbToc); err != nil {
		logger.Errorf("Failed to get disc metadata: %v", err)
		return "", fmt.Errorf("Failed to get disc metadata: %w", err)
	}
	logger.Infof("Successfully fetched disc metadata: %s - %s", discInfo.Artist, discInfo.Title)

	return finalizeIfSuccess(discInfo, cacheLocation, cueFilePath)
}

// fetchDiscInfoFromFlags returns DiscInfo, disc ID, and an error based on provided options.
func fetchDiscInfoFromFlags(musicbrainzID, providedDiscID string) (*types.DiscInfo, string, error) {
	// Enforce --musicbrainz with --disc-id
	if providedDiscID != "" && musicbrainzID == "" {
		return nil, "", fmt.Errorf("error: --disc-id option requires --musicbrainz to be set")
	}

	// If --musicbrainz is provided, fetch DiscInfo directly from MusicBrainz
	if musicbrainzID != "" {
		logger.Infof("Fetching release info from MusicBrainz ID: %s", musicbrainzID)
		discInfo, err := musicbrainz.FetchReleaseByID(musicbrainzID)
		if err != nil {
			logger.Errorf("Failed to fetch MusicBrainz release %s: %v", musicbrainzID, err)
			return nil, "", fmt.Errorf("Failed to get MusicBrainz %s Release: %w", musicbrainzID, err)
		}
		logger.Infof("Successfully fetched MusicBrainz release: %s - %s", discInfo.Artist, discInfo.Title)
		return discInfo, providedDiscID, nil
	}
	return nil, "", nil
}

// finalizeIfSuccess finalizes the creation of a CUE file and saves associated metadata.
//
// Parameters:
//   - discInfo: Metadata about the disc to include in the CUE file.
//   - cacheLocation: The cache directory path.
//   - cueFilePath: The path to save the CUE file.
//
// Returns:
//   - string: The path to the finalized CUE file.
//   - error: Any error encountered during the operation.
func finalizeIfSuccess(discInfo *types.DiscInfo, cacheLocation, cueFilePath string) (string, error) {
	logger.Debugf("Finalizing CUE file generation for: %s - %s", discInfo.Artist, discInfo.Title)
	if err := fetchCoverArtIfNeeded(discInfo, cacheLocation, cueFilePath); err != nil {
		logger.Warnf("Error fetching cover art: %v", err)
	}
	// Generate the CUE file and save
	if err := generateCueFile(discInfo, cacheLocation, cueFilePath); err != nil {
		logger.Errorf("Failed to generate CUE file %s: %v", cueFilePath, err)
		return "", fmt.Errorf("Failed To Generate cue file %s: %w", cueFilePath, err)
	}
	logger.Infof("Playlist successfully generated at: %s", cueFilePath)
	return cueFilePath, nil
}

// generateCueFile generates and writes a CUE file based on disc metadata.
//
// Parameters:
//   - info: Metadata about the disc.
//   - cacheLocation: The cache directory path.
//   - cueFilePath: The path to save the CUE file.
//
// Returns:
//   - error: Any error encountered during file creation.
func generateCueFile(info *types.DiscInfo, cacheLocation, cueFilePath string) error {
	logger.Debugf("Generating CUE file: %s", cueFilePath)
	file, err := os.Create(cueFilePath)
	if err != nil {
		return fmt.Errorf("Failed to create cue file %s: %w", cueFilePath, err)
	}
	defer file.Close()

	if info.CoverArtPath == "" {
		discID := filepath.Base(filepath.Dir(cueFilePath))
		coverFilePath := utils.CacheCoverArtPath(cacheLocation, discID)
		logger.Debugf("Attempting to fetch cover art for MusicBrainz ID: %s", info.ID)
		if err := fetchCoverArt(info.ID, coverFilePath); err == nil {
			info.CoverArtPath = coverFilePath
			logger.Infof("Cover art saved to: %s", coverFilePath)
		} else {
			logger.Warnf("Failed to fetch cover art: %v", err)
		}
	}

	var content string
	if info.ReleaseDate != "" {
		content += fmt.Sprintf("REM DATE \"%s\"\n", info.ReleaseDate)
	}
	if info.Genre != "" {
		content += fmt.Sprintf("REM GENRE \"%s\"\n", info.Genre)
	}
	if info.CoverArtPath != "" {
		content += fmt.Sprintf("REM COVER \"%s\"\n", info.CoverArtPath)
	}
	content += fmt.Sprintf("PERFORMER \"%s\"\nTITLE \"%s\"\n", info.Artist, info.Title)

	for i, track := range info.Tracks {
		content += fmt.Sprintf("FILE \"cdda:///%d\" WAVE\n  TRACK %02d AUDIO\n    TITLE \"%s\"\n",
			i+1, i+1, track)
	}

	_, err = file.WriteString(content)
	return err
}
