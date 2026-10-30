package moderation

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/store"
	"cahut.cloud/internal/window"
)

// A moderator opens the Window that a report names, whatever its access,
// while the report waits: in the console, in a frame bound to their
// session, of the Cahut or of the server, which opens that Window only.
// Each opening goes to the Logbook of its Room, so that its members read
// it. See docs/adr/0054-reporting-and-withdrawing-a-window.md.

// PreviewReportedInput is the input of preview_reported_window.
type PreviewReportedInput struct {
	Report int64 `json:"report" jsonschema:"the number of a report that waits, which list_reports gives"`
}

// ReportedPreview is the output of preview_reported_window.
type ReportedPreview struct {
	Window    string    `json:"window" jsonschema:"the Window, as room/name"`
	Frame     string    `json:"frame" jsonschema:"the path of the frame that shows the Window, on the domain of the Cahut, such as /-/frame/<capability>/"`
	ExpiresAt time.Time `json:"expires_at" jsonschema:"when the frame stops showing the Window, if the session or the report have not ended before"`
}

// previewReported gives the moderator who calls it a frame of the Window
// that the report names.
func previewReported(frameKey []byte) func(ctx context.Context, c *operation.Call, in PreviewReportedInput) (ReportedPreview, error) {
	return func(ctx context.Context, c *operation.Call, in PreviewReportedInput) (ReportedPreview, error) {
		c.Log("reports", fmt.Sprintf("report %d", in.Report))
		var kind byte
		switch {
		case c.Actor.Kind == operation.Person && c.Actor.Session != 0:
			kind = window.FrameByModerator
		case c.Actor.Kind == operation.Admin && c.Actor.Session != 0:
			kind = window.FrameByHost
		default:
			return ReportedPreview{}, operation.Invalidf("a reported Window opens in the console: sign in to it as a moderator of this Cahut")
		}
		p, err := get(ctx, c.Tx, in.Report)
		if err != nil {
			return ReportedPreview{}, err
		}
		if p.Decision != "" {
			return ReportedPreview{}, reportDecided(ReportRef(in))
		}
		var id, roomID, author int64
		err = c.Tx.QueryRowContext(ctx, `
			SELECT w.id, w.room_id, coalesce(w.created_by, 0) FROM reports p JOIN windows w ON w.id = p.window_id WHERE p.id = ?`,
			in.Report).Scan(&id, &roomID, &author)
		if err != nil {
			return ReportedPreview{}, err
		}
		at := p.Window.Room + "/" + p.Window.Name
		// The members of its Room read that a moderator opened it, or its
		// author alone when the Window is theirs alone.
		c.Log(at, fmt.Sprintf("report %d", in.Report))
		c.InRoom(roomID)
		if p.Window.Access == window.AccessAuthor {
			c.ReadBy(author)
		}
		if p.Window.Trashed {
			return ReportedPreview{}, operation.Invalidf("%s is in the trash: it opens to nobody, and its members may restore it", at)
		}
		expires := c.Now.Add(window.FrameTTL).Truncate(time.Second)
		f := window.Frame{Window: id, Kind: kind, ID: c.Actor.Session, Expires: expires.Unix()}
		return ReportedPreview{Window: at, Frame: window.FramePrefix + f.Sign(frameKey) + "/", ExpiresAt: expires}, nil
	}
}

// Opens reports whether the frame of a moderator opens the Window id: a
// report about it waits, and the person userID still administers the
// Cahut, when the frame is theirs. A frame of an administrator of the
// server gives no person: its session is checked in the database of the
// server.
func Opens(ctx context.Context, db *store.DB, id, userID int64) (bool, error) {
	var reported bool
	err := db.Read.QueryRowContext(ctx, "SELECT EXISTS (SELECT 1 FROM reports WHERE window_id = ? AND decision IS NULL)", id).Scan(&reported)
	if err != nil || !reported || userID == 0 {
		return reported, err
	}
	var admin bool
	err = db.Read.QueryRowContext(ctx, "SELECT role = 'admin' FROM users WHERE id = ? AND disabled_at IS NULL", userID).Scan(&admin)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return admin, err
}
