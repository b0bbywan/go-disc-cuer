package cue

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"go.uploadedlobster.com/discid"

	"github.com/b0bbywan/go-disc-cuer/config"
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
//                              the disc ID is determined automatically.
//   - musicbrainzID (string): A MusicBrainz release ID for fetching metadata. If empty,
//                              GNUDB is used as the fallback metadata source.
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
//   1. If a `providedDiscID` or `musicbrainzID` is provided, fetch corresponding disc info.
//   2. If `discID` is not determined, read the disc from the drive and compute its ID and TOC.
//   3. Check if a cached CUE file exists. If so, return it unless `overwrite` is true.
//   4. If `discInfo` and `discID` are both valid, finalize the CUE file generation.
//   5. If necessary, fetch metadata concurrently from GNUDB and MusicBrainz.
//   6. Ensure necessary directories exist, then create and save the CUE file.
//
// Notes:
// - This function is used internally by both `GenerateFromDisc` and `GenerateWithOptions`.
// - Fetching metadata from GNUDB and MusicBrainz occurs concurrently to improve efficiency.
//
// Returns:
//   - string: The path to the generated CUE file.
//   - error: Any error encountered during the operation.
func generate(device string, cuerConfig *config.Config, providedDiscID, musicbrainzID string, overwrite bool) (string, error) {
	if cuerConfig == nil {
		return "", fmt.Errorf("Failed to generate cue file: empty config")
	}
	discInfo, discID, err := fetchDiscInfoFromFlags(providedDiscID, musicbrainzID)
	if err != nil {
		return "", err
	}

	var disc discid.Disc
	var gnuToc string
	if discID == "" {
		disc, err = discid.Read(device)
		if err != nil {
			return "", err
		}
		defer disc.Close()
		if gnuToc, discID, err = utils.GetTocAndDiscID(disc); err != nil {
			return "", err
		}
	}
	cacheLocation := cuerConfig.GetCacheLocation()
	cueFilePath := utils.CachePlaylistPath(cacheLocation, discID)

	if utils.CheckIfPlaylistExists(cueFilePath) && !overwrite {
		return cueFilePath, nil
	}

	if discInfo != nil && discID != "" {
		return finalizeIfSuccess(discInfo, cacheLocation, cueFilePath)
	}
	var mbToc string
	if mbToc, err = utils.GetMusicBrainzTOC(disc); err != nil {
		return "", fmt.Errorf("Failed to get musicbrainz TOC: %w", err)
	}

	if err = utils.CreateFolderIfNeeded(cueFilePath); err != nil {
		return "", fmt.Errorf("Failed to create %s folder: %w", cueFilePath, err)
	}

	// Fetch DiscInfo concurrently
	if discInfo, err = fetchDiscInfoConcurrently(cuerConfig, gnuToc, mbToc); err != nil {
		return "", fmt.Errorf("Failed to get disc metadata: %w", err)
	}

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
		discInfo, err := musicbrainz.FetchReleaseByID(musicbrainzID)
		if err != nil {
			return nil, "", fmt.Errorf("Failed to get MusicBrainz %s Release: %w", musicbrainzID, err)
		}
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
    if err := fetchCoverArtIfNeeded(discInfo, cacheLocation, cueFilePath); err != nil {
        log.Printf("Error fetching cover art: %v", err)
    }
	// Generate the CUE file and save
	if err := generateCueFile(discInfo, cacheLocation, cueFilePath); err != nil {
		return "", fmt.Errorf("Failed To Generate cue file %s: %w", cueFilePath, err)
	}
	log.Printf("info: Playlist generated at %s", cueFilePath)
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
	file, err := os.Create(cueFilePath)
	if err != nil {
		return fmt.Errorf("Failed to create cue file %s: %w", cueFilePath, err)
	}
	defer file.Close()

	if info.CoverArtPath == "" {
		discID := filepath.Base(filepath.Dir(cueFilePath))
		coverFilePath := utils.CacheCoverArtPath(cacheLocation, discID)
		if err := fetchCoverArt(info.ID, coverFilePath); err == nil {
			info.CoverArtPath = coverFilePath
		} else {
			log.Printf("error getting cover: %v", err)
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
