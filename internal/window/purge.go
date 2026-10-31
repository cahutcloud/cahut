package window

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"cahut.cloud/internal/account"
	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/room"
)

// Purging a Window, deleting a Room and deleting a person cannot be undone: an AI that calls
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
	c.Log(w.Room+"/"+w.Name, fmt.Sprintf("%s %q, %s", w.Kind, w.Title, versions(w.Window)))
	c.AfterCommit(func() { s.remove(w.id) })
	return w.Window, nil
}

// versions names the versions that a Window keeps.
func versions(w Window) string {
	if w.Previous == 0 {
		return fmt.Sprintf("version %d", w.Version)
	}
	return fmt.Sprintf("versions %d and %d", w.Version, w.Previous)
}

// remove deletes the files of the Window id, every version, and its
// database with its copies, once its worker has stopped.
func (s *Service) remove(id int64) {
	s.data.Stop(s.dataDir(id))
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

// DeletedUser is the output of delete_user.
type DeletedUser struct {
	User    string   `json:"user" jsonschema:"the handle of the person, now free"`
	Name    string   `json:"name"`
	Rooms   []string `json:"rooms" jsonschema:"the Rooms deleted with the person: their personal Room, and the team Rooms they were the only owner of"`
	Windows int      `json:"windows" jsonschema:"how many Windows were deleted with these Rooms, those in their trash included"`
}

// deleteUser deletes a person, their personal Room and the team Rooms they
// are the only owner of, with the Windows of these Rooms. The Windows they
// published in the Rooms that stay keep no author. See
// docs/adr/0047-deleting-a-person.md.
func (s *Service) deleteUser(ctx context.Context, c *operation.Call, in account.UserInput) (DeletedUser, error) {
	c.Log(in.User, "")
	u, err := account.Deletable(ctx, c, in.User)
	if err != nil {
		return DeletedUser{}, err
	}
	rows, err := c.Tx.QueryContext(ctx, `
		SELECT r.id, r.name FROM rooms r JOIN room_members m ON m.room_id = r.id AND m.user_id = ?
		WHERE r.kind = 'personal' OR m.role = 'owner' AND NOT EXISTS (
			SELECT 1 FROM room_members o WHERE o.room_id = r.id AND o.role = 'owner' AND o.user_id != m.user_id)
		ORDER BY r.kind = 'personal' DESC, r.name`, u.ID)
	if err != nil {
		return DeletedUser{}, err
	}
	out := DeletedUser{User: u.Handle, Name: u.Name, Rooms: []string{}}
	var roomIDs []int64
	for rows.Next() {
		var id int64
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			rows.Close()
			return DeletedUser{}, err
		}
		roomIDs = append(roomIDs, id)
		out.Rooms = append(out.Rooms, name)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return DeletedUser{}, err
	}
	var ids []int64
	for _, room := range roomIDs {
		rows, err := c.Tx.QueryContext(ctx, "DELETE FROM windows WHERE room_id = ? RETURNING id", room)
		if err != nil {
			return DeletedUser{}, err
		}
		for rows.Next() {
			var id int64
			if err := rows.Scan(&id); err != nil {
				rows.Close()
				return DeletedUser{}, err
			}
			ids = append(ids, id)
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return DeletedUser{}, err
		}
		if _, err := c.Tx.ExecContext(ctx, "DELETE FROM rooms WHERE id = ?", room); err != nil {
			return DeletedUser{}, err
		}
	}
	// What the person did elsewhere stays, without their name: the Windows
	// and versions they published lose their author with the foreign keys,
	// and their invitations and share links lose it here. The Windows they
	// kept for themselves open to their Room, which no author holds any
	// more. The rest of what is theirs goes with them. See
	// docs/adr/0053-windows-for-their-author-alone.md.
	for _, q := range []string{
		"UPDATE windows SET access = 'room' WHERE created_by = ? AND access = 'author'",
		"UPDATE invitations SET invited_by = NULL WHERE invited_by = ?",
		"UPDATE share_links SET created_by = NULL WHERE created_by = ?",
		"DELETE FROM users WHERE id = ?",
	} {
		if _, err := c.Tx.ExecContext(ctx, q, u.ID); err != nil {
			return DeletedUser{}, err
		}
	}
	out.Windows = len(ids)
	c.Log(u.Handle, fmt.Sprintf("%q, Rooms %s, %d Windows", u.Name, strings.Join(out.Rooms, " "), len(ids)))
	c.AfterCommit(func() {
		for _, id := range ids {
			s.remove(id)
		}
	})
	return out, nil
}
