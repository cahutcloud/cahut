package window

import (
	"errors"
	"strings"
	"testing"
	"time"

	"cahut.cloud/internal/operation"
)

// frameKey signs the frames of the tests.
var frameKey = []byte("the key that signs frames in tests")

func TestPreview(t *testing.T) {
	f := setup(t)
	if _, err := f.invoke(t, f.olivier, "publish_window", deck); err != nil {
		t.Fatal(err)
	}
	if _, err := f.invoke(t, f.camille, "publish_window", deck); err != nil {
		t.Fatal(err)
	}
	olivier := person(f.olivier)
	olivier.Session = 42
	before := time.Now()
	out, err := f.invoke(t, olivier, "preview_windows",
		`{"windows": ["https://cahut.test/olivier/revue-du-3e-trimestre/", "camille/revue-du-3e-trimestre", "olivier/nope"]}`)
	if err != nil {
		t.Fatalf("preview_windows: %v", err)
	}

	// A Window of another Room, or that does not exist, has no preview.
	previews := out.(Previews).Previews
	if len(previews) != 1 || previews[0].Window != "olivier/revue-du-3e-trimestre" {
		t.Fatalf("previews = %+v", previews)
	}
	p := previews[0]
	signed, ok := strings.CutPrefix(p.Frame, FramePrefix)
	if !ok || !strings.HasSuffix(signed, "/") {
		t.Fatalf("frame = %q", p.Frame)
	}
	frame, ok := VerifyFrame(frameKey, strings.TrimSuffix(signed, "/"), time.Now())
	id, _, _ := current(t.Context(), f.db, "olivier", "revue-du-3e-trimestre")
	if !ok || frame.Window != id || frame.Kind != FrameByConsole || frame.ID != 42 || frame.Expires != p.ExpiresAt.Unix() {
		t.Errorf("frame = %+v, %v", frame, ok)
	}
	if ends := p.ExpiresAt.Sub(before); ends < FrameTTL-2*time.Second || ends > FrameTTL+time.Second {
		t.Errorf("the frame lasts %v", ends)
	}
	if _, ok := VerifyFrame([]byte("another key"), strings.TrimSuffix(signed, "/"), time.Now()); ok {
		t.Error("a frame verifies with another key")
	}

	// A Window in the trash has none either.
	if _, err := f.invoke(t, f.olivier, "trash_window", `{"window": "olivier/revue-du-3e-trimestre"}`); err != nil {
		t.Fatal(err)
	}
	out, err = f.invoke(t, olivier, "preview_windows", `{"windows": ["olivier/revue-du-3e-trimestre"]}`)
	if err != nil || len(out.(Previews).Previews) != 0 {
		t.Errorf("preview of a Window in the trash = %+v, %v", out, err)
	}

	// Previews are for the console: AIs have none, and a call needs a session.
	if _, err := f.invoke(t, f.olivier, "preview_windows", `{"windows": []}`); !errors.Is(err, operation.ErrForbidden) {
		t.Errorf("preview_windows by an AI: %v", err)
	}
	olivier.Session = 0
	if _, err := f.invoke(t, olivier, "preview_windows", `{"windows": []}`); !errors.Is(err, operation.ErrInvalid) {
		t.Errorf("preview_windows without a session: %v", err)
	}
	olivier.Session = 42
	many := `{"windows": ["olivier/a"` + strings.Repeat(`, "olivier/a"`, MaxPreviews) + `]}`
	if _, err := f.invoke(t, olivier, "preview_windows", many); !errors.Is(err, operation.ErrInvalid) {
		t.Errorf("preview_windows of %d Windows: %v", MaxPreviews+1, err)
	}
	if _, err := f.invoke(t, olivier, "preview_windows", `{"windows": ["nope"]}`); !errors.Is(err, operation.ErrInvalid) {
		t.Errorf("preview_windows of a bad address: %v", err)
	}
}

func TestVerifyFrame(t *testing.T) {
	now := time.Now()
	f := Frame{Window: 7, Kind: FrameByConsole, ID: 3, Expires: now.Add(time.Hour).Unix()}
	signed := f.Sign(frameKey)
	if got, ok := VerifyFrame(frameKey, signed, now); !ok || got != f {
		t.Errorf("VerifyFrame = %+v, %v", got, ok)
	}
	tampered := []byte(signed)
	tampered[5] ^= 1
	for name, s := range map[string]string{
		"tampered":     string(tampered),
		"expired":      Frame{Window: 7, Kind: FrameByConsole, ID: 3, Expires: now.Unix()}.Sign(frameKey),
		"of no kind":   Frame{Window: 7, Kind: 'x', ID: 3, Expires: now.Add(time.Hour).Unix()}.Sign(frameKey),
		"not base64":   "a/b",
		"too short":    "abc",
		"another key":  f.Sign([]byte("another key")),
		"with nothing": "",
	} {
		if got, ok := VerifyFrame(frameKey, s, now); ok {
			t.Errorf("a frame %s verifies: %+v", name, got)
		}
	}
}
