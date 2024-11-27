package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

const (
	AppName = "disc-cuer"
	AppVersion = "0.3"
)

type Config struct {
	AppName			string
	AppVersion		string
	GnuHelloEmail	string
	GnuDbUrl		string
	CacheLocation	string
	Device			string
}

// NewDefaultConfig creates a Config struct with default application settings.
// It is useful when no specific customization is required.
//
// Returns:
//   - *Config: A configuration instance with default values.
//   - error: Any error encountered during initialization.
func NewDefaultConfig() (*Config, error) {
	return NewConfig(AppName, AppVersion, "")
}

// NewConfig initializes the Config struct for the specified application with custom settings.
//
// Parameters:
//   - appName: The name of the application. Defaults to the package-level AppName if empty.
//   - appVersion: The version of the application. Defaults to the package-level AppVersion if empty.
//   - baseCacheFolder: The base folder for caching. Uses system defaults if empty.
//
// Returns:
//   - *Config: A configuration instance populated with the given settings.
//   - error: Any error encountered during initialization.
func NewConfig(appName, appVersion, baseCacheFolder string) (*Config, error) {
	if appName == "" {
		appName = AppName
	}
	if appVersion == "" {
		appVersion = AppVersion
	}
	cacheLocation := getCacheFolder(baseCacheFolder, appName)
	viper.SetDefault("cacheLocation", cacheLocation)
	viper.SetDefault("gnuHelloEmail", "")
	viper.SetDefault("gnuDbUrl", "https://gnudb.gnudb.org")
	viper.SetDefault("device", "/dev/sr0")

	// Load configuration paths and environment variables
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(filepath.Join("/etc", appName))
	if home, err := os.UserHomeDir(); err == nil {
		viper.AddConfigPath(filepath.Join(home, ".config", appName))
	}
	viper.SetEnvPrefix(strings.ToUpper(strings.ReplaceAll(appName, "-", "_")))
	viper.AutomaticEnv()

	// Attempt to read configuration file
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %v", err)
		}
	}

	// Populate the Config struct
	config := &Config{
		AppName:		appName,
		AppVersion:		appVersion,
		CacheLocation:	cacheLocation,
		GnuHelloEmail:	viper.GetString("gnuHelloEmail"),
		GnuDbUrl:		viper.GetString("gnuDbUrl"),
		Device:			viper.GetString("device"),
	}

	// Validate required fields
	if config.GnuHelloEmail == "" {
		fmt.Fprintf(os.Stderr, "Warning: gnuHelloEmail is required for gnuDB operations.\n")
	}

	return config, nil
}

// GetCacheLocation retrieves the cache folder path for the current configuration.
//
// Returns:
//   - string: The cache folder path.
func (c *Config) GetCacheLocation() string {
	return c.CacheLocation
}

func getCacheFolder(baseCacheFolder, appName string) string {
	if baseCacheFolder == "" {
		return getDefaultCacheFolder(appName)
	}
	return baseCacheFolder
}

func getDefaultCacheFolder(appName string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join("/var", "cache", appName)
	}
	return filepath.Join(home, ".cache", appName)
}
