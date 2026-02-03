package cue

import (
	"fmt"
	"os"

	"github.com/b0bbywan/go-disc-cuer/logger"
	"github.com/b0bbywan/go-disc-cuer/types"
)

// writeCueFile generates and writes a CUE file from disc metadata
//
// Parameters:
//   - discInfo: Disc metadata to write
//   - cueFilePath: Path where the CUE file should be written
//
// Returns:
//   - error: Error if file cannot be created or written
func writeCueFile(discInfo *types.DiscInfo, cueFilePath string) error {
	logger.Debugf("Writing CUE file to: %s", cueFilePath)

	file, err := os.Create(cueFilePath)
	if err != nil {
		logger.Errorf("Failed to create CUE file %s: %v", cueFilePath, err)
		return fmt.Errorf("failed to create CUE file: %w", err)
	}
	defer file.Close()

	content := buildCueContent(discInfo)

	if _, err = file.WriteString(content); err != nil {
		logger.Errorf("Failed to write CUE file content: %v", err)
		return fmt.Errorf("failed to write CUE file: %w", err)
	}

	logger.Infof("CUE file successfully written: %s", cueFilePath)
	return nil
}

// buildCueContent constructs the CUE file content from disc metadata
//
// Parameters:
//   - info: Disc metadata
//
// Returns:
//   - string: Complete CUE file content
func buildCueContent(info *types.DiscInfo) string {
	content := ""

	// Add optional metadata
	if info.ReleaseDate != "" {
		content += fmt.Sprintf("REM DATE \"%s\"\n", info.ReleaseDate)
	}
	if info.Genre != "" {
		content += fmt.Sprintf("REM GENRE \"%s\"\n", info.Genre)
	}
	if info.CoverArtPath != "" {
		content += fmt.Sprintf("REM COVER \"%s\"\n", info.CoverArtPath)
	}

	// Add required metadata
	content += fmt.Sprintf("PERFORMER \"%s\"\n", info.Artist)
	content += fmt.Sprintf("TITLE \"%s\"\n", info.Title)

	// Add tracks
	for i, track := range info.Tracks {
		trackNum := i + 1
		content += fmt.Sprintf("FILE \"cdda:///%d\" WAVE\n", trackNum)
		content += fmt.Sprintf("  TRACK %02d AUDIO\n", trackNum)
		content += fmt.Sprintf("    TITLE \"%s\"\n", track)
	}

	return content
}
