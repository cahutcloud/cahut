package window

import (
	"context"
	"fmt"
	"time"

	"cahut.cloud/internal/operation"
)

// MaxPreviews is how many Windows one call of preview_windows previews.
const MaxPreviews = 100

// PreviewInput is the input of preview_windows.
type PreviewInput struct {
	Windows []string `json:"windows" jsonschema:"the Windows to preview, by their address or as room/name, at most 100"`
}

// Previews is the output of preview_windows.
type Previews struct {
	Previews []Preview `json:"previews" jsonschema:"the previews of the Windows asked for that are in your Rooms and out of the trash, in the same order"`
}

// Preview is the frame that shows a Window in the console.
type Preview struct {
	Window    string    `json:"window" jsonschema:"the Window, as room/name"`
	Frame     string    `json:"frame" jsonschema:"the path of the frame that shows the Window, on the domain of the Cahut, such as /-/frame/<capability>/"`
	ExpiresAt time.Time `json:"expires_at" jsonschema:"when the frame stops showing the Window, if the session has not ended before"`
}

// preview gives the console a frame for each Window it lists, bound to its
// session. Each frame opens one Window only, so that a Window that reads
// the address of its own frame opens no other. See
// docs/adr/0027-previews-of-windows-in-the-console.md.
func (s *Service) preview(ctx context.Context, c *operation.Call, in PreviewInput) (Previews, error) {
	c.Log(fmt.Sprintf("%d Windows", len(in.Windows)), "")
	if c.Actor.Kind != operation.Person || c.Actor.Session == 0 {
		return Previews{}, operation.Invalidf("previews show Windows in the console: sign in to it as a person of this Cahut")
	}
	if len(in.Windows) > MaxPreviews {
		return Previews{}, operation.Invalidf("%d Windows: ask for %d at most in one call", len(in.Windows), MaxPreviews)
	}
	expires := c.Now.Add(FrameTTL).Truncate(time.Second)
	out := Previews{Previews: []Preview{}}
	for _, ref := range in.Windows {
		r, ok := parseRef(ref)
		if !ok {
			return Previews{}, operation.Invalidf("window %q: give the Window's id or address, or its Room and its name as room/name", ref)
		}
		where, args := "r.name = ? AND w.name = ?", []any{r.room, r.name}
		if r.id != "" {
			where, args = "w.uid = ?", []any{r.id}
		}
		found, err := s.query(ctx, c, "w.trashed_at IS NULL AND "+where, args...)
		if err != nil {
			return Previews{}, err
		}
		if len(found) == 0 {
			continue
		}
		f := Frame{Window: found[0].id, Kind: FrameByConsole, ID: c.Actor.Session, Expires: expires.Unix()}
		out.Previews = append(out.Previews, Preview{
			Window:    found[0].Room + "/" + found[0].Name,
			Frame:     FramePrefix + f.Sign(s.frameKey) + "/",
			ExpiresAt: expires,
		})
	}
	return out, nil
}
