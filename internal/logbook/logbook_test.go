package logbook

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"cahut.cloud/internal/store"
)

func TestClip(t *testing.T) {
	for _, tt := range []struct {
		in   string
		n    int
		want string
	}{
		{"olivier/q3", 512, "olivier/q3"},
		{"abcdefgh", 6, "abc…"},
		{"ééééé", 7, "éé…"}, // never half a character
	} {
		if got := Clip(tt.in, tt.n); got != tt.want || len(got) > tt.n {
			t.Errorf("Clip(%q, %d) = %q, want %q", tt.in, tt.n, got, tt.want)
		}
	}
}

// TestEntriesStaySmall checks that a caller cannot write rows of megabytes,
// which every reading of the Logbook would then load.
func TestEntriesStaySmall(t *testing.T) {
	ctx := context.Background()
	db, err := store.Open(ctx, filepath.Join(t.TempDir(), "cahut.db"), store.CahutSchema)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	huge := strings.Repeat("x", 4<<20)
	if err := Write(ctx, db.Write, Entry{At: time.Now(), Operation: "get_window", Actor: "ai", Via: "mcp",
		Client: huge, Target: huge, Detail: huge, Outcome: Failed, Error: huge}); err != nil {
		t.Fatal(err)
	}
	var size int
	if err := db.Read.QueryRowContext(ctx, "SELECT length(client) + length(target) + length(detail) + length(error) FROM logbook").Scan(&size); err != nil {
		t.Fatal(err)
	}
	if size > MaxClient+MaxTarget+MaxDetail+MaxError {
		t.Errorf("an entry keeps %d bytes", size)
	}
}
