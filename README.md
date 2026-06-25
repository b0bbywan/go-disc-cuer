<p align="center">
  <a href="https://odio.love"><img src="https://odio.love/logo.png" alt="odio" width="160" /></a>   
  </p>
  <h1 align="center">go-disc-cuer</h1>
  <p align="center"><em>Generate CUE files from audio CDs with MusicBrainz and GNUDB metadata.</em></p>
  <p align="center">
  <a href="https://github.com/b0bbywan/go-disc-cuer/releases"><img src="https://img.shields.io/github/v/release/b0bbywan/go-disc-cuer?include_prereleases" alt="Release" /></a>
  <a href="https://github.com/b0bbywan/go-disc-cuer/blob/main/LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue" alt="License" /></a>
  <a href="https://goreportcard.com/report/github.com/b0bbywan/go-disc-cuer"><img src="https://goreportcard.com/badge/github.com/b0bbywan/go-disc-cuer" alt="Go Report Card" /></a>
  <a href="https://pkg.go.dev/github.com/b0bbywan/go-disc-cuer"><img src="https://pkg.go.dev/badge/github.com/b0bbywan/go-disc-cuer.svg" alt="Go Reference" /></a>
  <a href="https://github.com/sponsors/b0bbywan"><img src="https://img.shields.io/github/sponsors/b0bbywan?label=Sponsor&logo=GitHub" alt="GitHub Sponsors" /></a>   
  </p>
  <p align="center">
  <a href="https://docs.odio.love/guides/audio-cd/"><img src="https://img.shields.io/badge/Audio%20CD-F18D00" alt="Audio CD" /></a>
  <a href="https://docs.odio.love/disc-player/disc-cuer/"><img src="https://img.shields.io/badge/CUE%20generation-5B21B6" alt="CUE generation" /></a>
  <a href="https://musicbrainz.org/"><img src="https://img.shields.io/badge/MusicBrainz-BA478F?logo=musicbrainz&logoColor=white" alt="MusicBrainz" /></a>
  <a href="https://gnudb.org/"><img src="https://img.shields.io/badge/GNUDB-A42E2B" alt="GNUDB" /></a>   
  </p>
  <p align="center">   
  Part of the <a href="https://odio.love">odio</a> project — <a href="https://docs.odio.love/disc-player/disc-cuer/">full documentation</a>.
  </p>
  <p align="center">
  <a href="https://go.dev/"><img src="https://img.shields.io/badge/Go-00ADD8?logo=go&logoColor=white" alt="Go" /></a>
  <a href="https://musicbrainz.org/doc/libdiscid"><img src="https://img.shields.io/badge/libdiscid-2E6DB4" alt="libdiscid" /></a>
  </p>

## Features

- **Disc ID Calculation**: Uses `libdiscid` to compute MusicBrainz and GNUDB compatible disc IDs.
- **Metadata Integration**: Fetch track and album metadata from GNUDB or MusicBrainz.
- **Fix Incorrect CUE Files**: Force the use of a specific MusicBrainz release to correct or regenerate CUE files.
- **Configurable**: Allows configuration through files, environment variables, and command-line flags.

## Installation

### From Releases
Download the prebuilt `linux/amd64` binary attached to the
[latest release](https://github.com/b0bbywan/go-disc-cuer/releases):

```bash
curl -L -o disc-cuer \
  https://github.com/b0bbywan/go-disc-cuer/releases/latest/download/disc-cuer-linux-amd64
chmod +x disc-cuer
sudo mv disc-cuer /usr/local/bin/
```

You still need `libdiscid` installed at runtime (see the dependency step below).

### From Source
1. Clone the repository:
    ```bash
    git clone https://github.com/b0bbywan/go-disc-cuer.git
    cd go-disc-cuer
    ```
2. Install dependencies (`libdiscid`):
    ```bash
    make deps
    # or manually:
    #   Debian:  sudo apt install libdiscid0 libdiscid-dev
    #   Fedora:  sudo dnf install libdiscid libdiscid-devel
    ```

3. Build the binary (the version is stamped from `git describe`):
    ```bash
    make build
    ```

4. (Optional) Install globally:
    ```bash
    sudo mv disc-cuer /usr/local/bin/
    ```

Other Makefile targets: `make test` (race + coverage), `make lint`, `make dist`
(linux/amd64 release binary into `dist/`), `make clean`.

## Usage
1. Basic Command
Generate a CUE file for the current CD:
    ```bash
    disc-cuer
    ```

2. Options
- `--overwrite`: Regenerate the CUE file even if it exists.
- `--musicbrainz <release_id>`: Specify a MusicBrainz release ID to fetch album metadata.
- `--disc-id <disc_id>`: Provide a custom disc ID. This requires --musicbrainz to associate metadata with the ID.
- `--device <device>`: Specify the disc drive device to read from (overrides config or default)

3. Configuration
The tool loads configurations in the following order of priority:

- Command-line flags.
- Environment variables prefixed with DISC_CUER_.
- Configuration files located at:
   - `/etc/disc-cuer/config.yml`
   - `~/.config/disc-cuer/config.yaml`

4. Example Configuration

    **Please note that gnuHelloEmail is mandatory to use gnudb source**

    ```yaml
    gnuHelloEmail: "your-email@example.com"  # (no default)
    gnuDbUrl: "https://gnudb.gnudb.org"      # (default)
    cacheLocation: "/var/cache/disc-cuer"    # (root default, else ~/.cache/disc-cuer)
    device: "/dev/sr0"                       # (default)
    ```

    ```bash
    DISC_CUER_GNUHELLOEMAIL="your-email@example.com" DISC_CUER_GNUDBURL="https://gnudb.gnudb.org" DISC_CUER_CACHELOCATION="/var/cache/disc-cuer" DISC_CUER_DEVICE="/dev/sr0" disc-cuer --disc-id <id> --musicbrainz <release_id> --overwrite
    ```

## Examples
- Fetch Metadata from music brainz and force Generate CUE for current disc
    ```bash
    disc-cuer --musicbrainz <release_id> --overwrite
    ```
- Force Custom Disc ID
    ```bash
    disc-cuer --disc-id <disc_id> --musicbrainz <release_id> --overwrite
    ```

## Use as a library

`go-disc-cuer` can be embedded in another program (for example
[go-mpd-discplayer](https://github.com/b0bbywan/go-mpd-discplayer)). Build a
`*config.Config` with your own app name and version, then drive the flow through
`cue.New(cfg).Generate(...)`:

```go
import (
    "github.com/b0bbywan/go-disc-cuer/config"
    "github.com/b0bbywan/go-disc-cuer/cue"
)

// Your app's identity is what the GNUDB hello reports; pass "" to fall back to
// disc-cuer's own defaults. The third argument overrides the cache base folder.
cfg, err := config.NewConfig("my-app", "1.2.3", "")
if err != nil {
    return err
}

path, err := cue.New(cfg).Generate(cue.Options{
    Device:    "/dev/sr0", // read this drive (omit DiscID to read the disc)
    Overwrite: true,
})
```

`Options` also accepts `DiscID` + `MusicBrainzID` to bypass the drive/TOC lookup
and target a specific release.

## Project Structure
- `main.go`: Entry point and CLI flag parsing.
- `cue/`: Flow orchestration (`Generate`), CUE rendering, and cover-art fetching.
- `gnudb/`: GNUDB (CDDB protocol) integration.
- `musicbrainz/`: MusicBrainz web-service integration.
- `config/`: Configuration package built on github.com/spf13/viper.
- `types/`: Shared `DiscInfo` model and MusicBrainz JSON structs.
- `utils/`: TOC/disc-id computation and cache-path helpers.


## Contributing
Contributions are welcome! Feel free to submit issues, feature requests, or pull requests.

- Fork the repository.
- Create a feature branch.
- Commit your changes and open a pull request.
