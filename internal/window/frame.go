package window

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"time"
)

// FrameTTL is how long the address of the frame of a Window lasts.
const FrameTTL = 4 * time.Hour

// FramePrefix starts the path of every frame, such as /-/w/<capability>/.
const FramePrefix = "/-/w/"

// What opened the frame of a Window.
const (
	FrameBySession byte = 's' // a reader session, at the address of a protected Window
	FrameByPass    byte = 'p' // a pass, given for the Key of a Window
	FrameByConsole byte = 'c' // a session of the console, for the preview of a Window
)

// Frame is the capability in the address of the frame of a Window: it names
// the Window, the session or the pass that opened it, and when it expires.
// It is signed with the frame key of the Cahut. See
// docs/adr/0019-window-versions-access-and-declaration.md and
// docs/adr/0027-previews-of-windows-in-the-console.md.
type Frame struct {
	Window  int64
	Kind    byte
	ID      int64
	Expires int64 // Unix time
}

const (
	frameSize = 8 + 1 + 8 + 8
	macSize   = 16
)

// Sign returns the capability and its signature, in base64url.
func (f Frame) Sign(key []byte) string {
	b := make([]byte, 0, frameSize+macSize)
	b = binary.BigEndian.AppendUint64(b, uint64(f.Window)) //nolint:gosec // ids and times are positive
	b = append(b, f.Kind)
	b = binary.BigEndian.AppendUint64(b, uint64(f.ID))      //nolint:gosec // ids and times are positive
	b = binary.BigEndian.AppendUint64(b, uint64(f.Expires)) //nolint:gosec // ids and times are positive
	return base64.RawURLEncoding.EncodeToString(append(b, mac(key, b)...))
}

// VerifyFrame returns the capability that s signs with key, when the
// signature holds and it has not expired.
func VerifyFrame(key []byte, s string, now time.Time) (Frame, bool) {
	b, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil || len(b) != frameSize+macSize {
		return Frame{}, false
	}
	data, sig := b[:frameSize], b[frameSize:]
	if !hmac.Equal(sig, mac(key, data)) {
		return Frame{}, false
	}
	f := Frame{
		Window:  int64(binary.BigEndian.Uint64(data[0:8])), //nolint:gosec // signed by this Cahut
		Kind:    data[8],
		ID:      int64(binary.BigEndian.Uint64(data[9:17])),  //nolint:gosec // signed by this Cahut
		Expires: int64(binary.BigEndian.Uint64(data[17:25])), //nolint:gosec // signed by this Cahut
	}
	if now.Unix() >= f.Expires || f.Kind != FrameBySession && f.Kind != FrameByPass && f.Kind != FrameByConsole {
		return Frame{}, false
	}
	return f, true
}

func mac(key, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)[:macSize]
}
