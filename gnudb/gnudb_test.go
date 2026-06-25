package gnudb

import (
	"net/http"
	"net/http/httptest"
	"os"
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

func TestRedactHello(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			name: "hello between params",
			in:   "http://h/cddb.cgi?cmd=cddb+query+x&hello=me@example.com+host+app+1&proto=6",
			want: "http://h/cddb.cgi?cmd=cddb+query+x&hello=***&proto=6",
		},
		{
			name: "hello at end",
			in:   "http://h/cddb.cgi?cmd=x&hello=me@example.com+host+app+1",
			want: "http://h/cddb.cgi?cmd=x&hello=***",
		},
		{
			name: "no hello",
			in:   "http://h/cddb.cgi?cmd=x&proto=6",
			want: "http://h/cddb.cgi?cmd=x&proto=6",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := redactHello(tt.in); got != tt.want {
				t.Errorf("redactHello() = %q, want %q", got, tt.want)
			}
		})
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

func TestParseTrackLineContinuation(t *testing.T) {
	// Two lines sharing TTITLE0 must concatenate into a single track.
	body := "DTITLE=Artist / Album\nTTITLE0=Long title part one \nTTITLE0=and part two\n"
	info, err := parseGNUDBResponse(strings.NewReader(body))
	if err != nil {
		t.Fatalf("parseGNUDBResponse() error = %v", err)
	}
	if len(info.Tracks) != 1 {
		t.Fatalf("got %d tracks, want 1: %q", len(info.Tracks), info.Tracks)
	}
	if want := "Long title part one and part two"; info.Tracks[0] != want {
		t.Errorf("track 0 = %q, want %q", info.Tracks[0], want)
	}
}

// TestParseGNUDBResponseIssue9 replays the real GNUDB record from issue #9, where
// two titles (TTITLE13 and TTITLE19) are wrapped across lines. The wrapped lines
// must rejoin into their own track rather than spawning extra ones.
func TestParseGNUDBResponseIssue9(t *testing.T) {
	f, err := os.Open("testdata/issue9_read.txt")
	if err != nil {
		t.Fatalf("opening fixture: %v", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			t.Errorf("closing fixture: %v", err)
		}
	}()

	info, err := parseGNUDBResponse(f)
	if err != nil {
		t.Fatalf("parseGNUDBResponse() error = %v", err)
	}

	if len(info.Tracks) != 21 {
		t.Fatalf("got %d tracks, want 21 (wrapped lines must not add tracks):\n%s",
			len(info.Tracks), strings.Join(info.Tracks, "\n"))
	}

	// TTITLE19 is wrapped mid-word (`"Th` + `e Biz"`); it must rejoin cleanly.
	const wantSatisfaction = `Benny Benassi Presents "The Biz" / Benny Benassi Presents "The Biz" - Satisfaction (2003)`
	if info.Tracks[19] != wantSatisfaction {
		t.Errorf("track 19 (satisfaction) =\n%q\nwant\n%q", info.Tracks[19], wantSatisfaction)
	}

	// TTITLE13 is also wrapped across two lines.
	if got := info.Tracks[13]; !strings.HasPrefix(got, "Jonatan Cerrada / ") || !strings.Contains(got, "T'attends (2003)") {
		t.Errorf("track 13 not rejoined: %q", got)
	}

	// Last track confirms nothing shifted.
	if want := `Diam's / Diam's - Dj (2003)`; info.Tracks[20] != want {
		t.Errorf("track 20 = %q, want %q", info.Tracks[20], want)
	}
}
