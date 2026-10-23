package window

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/room"
)

// Purging a Window and deleting a Room cannot be undone: an AI that calls
// them knocks first. Their files go once the call commits, never before, so
// that a call that only knocks leaves them. See
// docs/adr/0021-ais-knock-first.md.

func (s *Service) purge(ctx context.Context, c *operation.Call, in RefInput) (Window, error) {
	w, err := s.target(ctx, c, in.Window)
	if err != nil {
		return Window{}, err
	}
	if w.TrashedAt == nil {
		return Window{}, operation.Invalidf("%s/%s is not in the trash: put it there with trash_window first", w.Room, w.Name)
	}
	if _, err := c.Tx.ExecContext(ctx, "DELETE FROM windows WHERE id = ?", w.id); err != nil {
		return Window{}, err
	}
	c.Log(w.Room+"/"+w.Name, fmt.Sprintf("%s %q, versions %s", w.Kind, w.Title, versions(w.Window)))
	c.AfterCommit(func() { s.remove(w.id) })
	return w.Window, nil
}

// versions names the versions that a Window keeps.
func versions(w Window) string {
	if w.Previous == 0 {
		return strconv.Itoa(w.Version)
	}
	return fmt.Sprintf("%d and %d", w.Version, w.Previous)
}

// remove deletes the files of the Window id, every version.
func (s *Service) remove(id int64) {
	_ = os.RemoveAll(filepath.Join(s.dir, strconv.FormatInt(id, 10))) // a leftover directory serves nothing
}

// DeleteRoomInput is the input of delete_room.
type DeleteRoomInput struct {
	Room string `json:"room" jsonschema:"the name of the team Room"`
}

// DeletedRoom is the output of delete_room.
type DeletedRoom struct {
	Room    string `json:"room" jsonschema:"the name of the Room, now free"`
	Title   string `json:"title"`
	Windows int    `json:"windows" jsonschema:"how many Windows were deleted with it, those in its trash included"`
}

// deleteRoom deletes a team Room, its members and its Windows. It lives
// with the Windows, whose files it deletes.
func (s *Service) deleteRoom(ctx context.Context, c *operation.Call, in DeleteRoomInput) (DeletedRoom, error) {
	c.Log(in.Room, "")
	a, err := room.Require(ctx, c, in.Room, room.Owner)
	if err != nil {
		return DeletedRoom{}, err
	}
	c.InRoom(a.ID)
	if a.Kind == room.Personal {
		return DeletedRoom{}, operation.Invalidf("%s is a personal Room: it keeps its person, and is never deleted", a.Name)
	}
	rows, err := c.Tx.QueryContext(ctx, "SELECT id FROM windows WHERE room_id = ?", a.ID)
	if err != nil {
		return DeletedRoom{}, err
	}
	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return DeletedRoom{}, err
		}
		ids = append(ids, id)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return DeletedRoom{}, err
	}
	// The versions and the passes of the Windows, the members of the Room
	// and its place in the reach of AIs go with them.
	if _, err := c.Tx.ExecContext(ctx, "DELETE FROM windows WHERE room_id = ?", a.ID); err != nil {
		return DeletedRoom{}, err
	}
	if _, err := c.Tx.ExecContext(ctx, "DELETE FROM rooms WHERE id = ?", a.ID); err != nil {
		return DeletedRoom{}, err
	}
	c.Log(a.Name, fmt.Sprintf("%q, %d Windows", a.Title, len(ids)))
	c.AfterCommit(func() {
		for _, id := range ids {
			s.remove(id)
		}
	})
	return DeletedRoom{Room: a.Name, Title: a.Title, Windows: len(ids)}, nil
}
