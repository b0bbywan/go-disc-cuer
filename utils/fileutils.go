package utils

import (
	"log"
	"os"
	"path/filepath"
)

// CheckIfPlaylistExists checks if a CUE playlist file already exists at the specified path.
//
// Parameters:
//   - cueFilePath (string): The path to the CUE file that needs to be checked for existence.
//
// Returns:
//   - bool: True if the playlist file exists, false otherwise.
//   - The function also logs an informational message if the file exists.
func CheckIfPlaylistExists(cueFilePath string) bool {
	if _, err := os.Stat(cueFilePath); err == nil {
		log.Printf("info: Playlist already exists at %s", cueFilePath)
		return true
	}
	return false
}

// CreateFolderIfNeeded ensures that the folder for the playlist file exists. If the folder does not exist, it creates it.
//
// Parameters:
//   - cueFilePath (string): The path to the CUE file whose folder needs to be created if it doesn't already exist.
//
// Returns:
//   - error: Any error encountered during the folder creation process.
func CreateFolderIfNeeded(cueFilePath string) error {
	return os.MkdirAll(filepath.Dir(cueFilePath), os.ModePerm)
}

// CachePlaylistPath generates the file path where the playlist CUE file is cached based on the disc ID.
//
// Parameters:
//   - cacheLocation: The base directory for caching files.
//   - discID (string): The disc ID used to generate the path for the cached playlist CUE file.
//
// Returns:
//   - string: The generated file path for the cached CUE file.
func CachePlaylistPath(cacheLocation, discID string) string {
	return getCachePath(cacheLocation, discID, "playlist.cue")
}

// CacheCoverArtPath generates the file path where the cover art image is cached based on the disc ID.
//
// Parameters:
//   - cacheLocation: The base directory for caching files.
//   - discID (string): The disc ID used to generate the path for the cached cover art file.
//
// Returns:
//   - string: The generated file path for the cached cover art image.
func CacheCoverArtPath(cacheLocation, discID string) string {
	return getCachePath(cacheLocation, discID, "cover.jpg")
}

// getCachePath constructs a file path within the cache directory for a given disc ID and filename.
//
// Parameters:
//   - cacheLocation: The base directory for caching files.
//   - discID: The disc ID used to create a subdirectory.
//   - filename: The name of the file to be cached.
//
// Returns:
//   - string: The complete file path for the cached file.
func getCachePath(cacheLocation, discID, filename string) string {
	return filepath.Join(cacheLocation, discID, filename)
}
