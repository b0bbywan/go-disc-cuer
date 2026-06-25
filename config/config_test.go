package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
)

// resetViper isolates each test from the global viper singleton that NewConfig
// mutates. Tests touching it must not run in parallel.
func resetViper(t *testing.T) {
	t.Helper()
	viper.Reset()
	t.Cleanup(viper.Reset)
}

func TestNewConfigDefaults(t *testing.T) {
	resetViper(t)
	home := t.TempDir()
	t.Setenv("HOME", home)

	cfg, err := NewConfig("", "", "")
	if err != nil {
		t.Fatalf("NewConfig() error = %v", err)
	}

	if cfg.AppName != AppName {
		t.Errorf("AppName = %q, want %q", cfg.AppName, AppName)
	}
	if cfg.AppVersion != AppVersion {
		t.Errorf("AppVersion = %q, want %q", cfg.AppVersion, AppVersion)
	}
	if cfg.GnuDbUrl != "https://gnudb.gnudb.org" {
		t.Errorf("GnuDbUrl = %q, want default", cfg.GnuDbUrl)
	}
	if cfg.Device != "/dev/sr0" {
		t.Errorf("Device = %q, want /dev/sr0", cfg.Device)
	}
	if want := filepath.Join(home, ".cache", AppName); cfg.CacheLocation != want {
		t.Errorf("CacheLocation = %q, want %q", cfg.CacheLocation, want)
	}
}

func TestNewConfigBaseCacheFolderOverride(t *testing.T) {
	resetViper(t)
	t.Setenv("HOME", t.TempDir())

	cfg, err := NewConfig("myapp", "1.2", "/custom/cache")
	if err != nil {
		t.Fatalf("NewConfig() error = %v", err)
	}
	if cfg.AppName != "myapp" || cfg.AppVersion != "1.2" {
		t.Errorf("app name/version not honored: %+v", cfg)
	}
	if cfg.CacheLocation != "/custom/cache" {
		t.Errorf("CacheLocation = %q, want /custom/cache", cfg.CacheLocation)
	}
}

func TestNewConfigEnvOverride(t *testing.T) {
	resetViper(t)
	t.Setenv("HOME", t.TempDir())
	// SetEnvPrefix("DISC_CUER") + AutomaticEnv binds these.
	t.Setenv("DISC_CUER_DEVICE", "/dev/custom")
	t.Setenv("DISC_CUER_GNUHELLOEMAIL", "tester@example.com")

	cfg, err := NewConfig("", "", "")
	if err != nil {
		t.Fatalf("NewConfig() error = %v", err)
	}
	if cfg.Device != "/dev/custom" {
		t.Errorf("Device = %q, want /dev/custom (from env)", cfg.Device)
	}
	if cfg.GnuHelloEmail != "tester@example.com" {
		t.Errorf("GnuHelloEmail = %q, want env value", cfg.GnuHelloEmail)
	}
}

func TestNewConfigReadsConfigFile(t *testing.T) {
	resetViper(t)
	home := t.TempDir()
	t.Setenv("HOME", home)

	dir := filepath.Join(home, ".config", AppName)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("creating config dir: %v", err)
	}
	yaml := "device: /dev/fromfile\ngnuHelloEmail: file@example.com\n"
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(yaml), 0o644); err != nil {
		t.Fatalf("writing config file: %v", err)
	}

	cfg, err := NewConfig("", "", "")
	if err != nil {
		t.Fatalf("NewConfig() error = %v", err)
	}
	if cfg.Device != "/dev/fromfile" {
		t.Errorf("Device = %q, want /dev/fromfile (from file)", cfg.Device)
	}
	if cfg.GnuHelloEmail != "file@example.com" {
		t.Errorf("GnuHelloEmail = %q, want file value", cfg.GnuHelloEmail)
	}
}

func TestGetCacheFolder(t *testing.T) {
	if got := getCacheFolder("/explicit", "myapp"); got != "/explicit" {
		t.Errorf("getCacheFolder with base = %q, want /explicit", got)
	}

	home := t.TempDir()
	t.Setenv("HOME", home)
	if got, want := getCacheFolder("", "myapp"), filepath.Join(home, ".cache", "myapp"); got != want {
		t.Errorf("getCacheFolder default = %q, want %q", got, want)
	}
}

func TestGetDefaultCacheFolderNoHome(t *testing.T) {
	t.Setenv("HOME", "") // os.UserHomeDir then errors, triggering the /var/cache fallback
	if got, want := getDefaultCacheFolder("myapp"), filepath.Join("/var", "cache", "myapp"); got != want {
		t.Errorf("getDefaultCacheFolder = %q, want %q", got, want)
	}
}

func TestGetCacheLocation(t *testing.T) {
	cfg := &Config{CacheLocation: "/some/path"}
	if cfg.GetCacheLocation() != "/some/path" {
		t.Errorf("GetCacheLocation() = %q, want /some/path", cfg.GetCacheLocation())
	}
}
