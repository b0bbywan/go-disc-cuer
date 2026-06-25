package gnudb

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/b0bbywan/go-disc-cuer/config"
)

// gnudbStub serves the two-step GNUDB exchange: a `cddb query` returning an
// exact match, then a `cddb read data` returning the record body.
func gnudbStub(t *testing.T, queryBody, readBody string) *config.Config {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cmd := r.URL.Query().Get("cmd")
		switch {
		case strings.Contains(cmd, "query"):
			if _, err := w.Write([]byte(queryBody)); err != nil {
				t.Errorf("writing response: %v", err)
			}
		case strings.Contains(cmd, "read"):
			if _, err := w.Write([]byte(readBody)); err != nil {
				t.Errorf("writing response: %v", err)
			}
		default:
			t.Errorf("unexpected cmd: %q", cmd)
		}
	}))
	t.Cleanup(srv.Close)
	return &config.Config{
		AppName:       "disc-cuer",
		AppVersion:    "test",
		GnuHelloEmail: "tester@example.com",
		GnuDbUrl:      srv.URL,
	}
}

const gnudbReadBody = `# xmcd
DTITLE=The Artist / The Album
DYEAR=2021
DGENRE=Rock
TTITLE0=First
TTITLE1=Second
`

func TestFetchDiscInfo(t *testing.T) {
	query := "210 OK\nFound exact matches, list follows\nrock abcd1234 The Artist / The Album\n.\n"
	cfg := gnudbStub(t, query, gnudbReadBody)

	info, err := FetchDiscInfo(cfg, "abcd1234+2+150+1500+200")
	if err != nil {
		t.Fatalf("FetchDiscInfo() error = %v", err)
	}
	if info.Artist != "The Artist" || info.Title != "The Album" {
		t.Errorf("unexpected artist/title: %+v", info)
	}
	if info.ReleaseDate != "2021" || info.Genre != "Rock" {
		t.Errorf("unexpected date/genre: %+v", info)
	}
	if len(info.Tracks) != 2 || info.Tracks[0] != "First" || info.Tracks[1] != "Second" {
		t.Errorf("unexpected tracks: %v", info.Tracks)
	}
}

func TestFetchDiscInfoNoMatch(t *testing.T) {
	cfg := gnudbStub(t, "202 No match found\n", gnudbReadBody)
	if _, err := FetchDiscInfo(cfg, "deadbeef+1+150+200"); err == nil {
		t.Fatal("expected error when GNUDB reports no match, got nil")
	}
}

func TestNewGnuConfigRequiresEmail(t *testing.T) {
	if _, err := newGnuConfig(&config.Config{}); err == nil {
		t.Fatal("newGnuConfig without GnuHelloEmail expected error, got nil")
	}
}

func TestExtractGnuDBID(t *testing.T) {
	id, err := extractGnuDBID("210 OK\nrock abcd1234 Artist / Title\n")
	if err != nil {
		t.Fatalf("extractGnuDBID() error = %v", err)
	}
	if id != "abcd1234" {
		t.Errorf("extractGnuDBID() = %q, want %q", id, "abcd1234")
	}

	if _, err := extractGnuDBID("single line only"); err == nil {
		t.Fatal("expected error on malformed response, got nil")
	}
}

func TestParseGNUDBResponse(t *testing.T) {
	info, err := parseGNUDBResponse(strings.NewReader(gnudbReadBody))
	if err != nil {
		t.Fatalf("parseGNUDBResponse() error = %v", err)
	}
	if info.Title != "The Album" || len(info.Tracks) != 2 {
		t.Errorf("unexpected parse result: %+v", info)
	}

	if _, err := parseGNUDBResponse(strings.NewReader("DYEAR=2021\n")); err == nil {
		t.Fatal("expected error when title is missing, got nil")
	}
}
