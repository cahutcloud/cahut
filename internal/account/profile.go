package account

import (
	"context"
	"errors"
	"strings"

	"cahut.cloud/internal/address"
	"cahut.cloud/internal/mail"
	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/settings"
	"cahut.cloud/internal/store"
)

// The longest name of a person: the longest title of a Room, since it
// titles their personal Room.
const maxName = 200

// The refusals of a name and of a new handle, which the console says in the
// language of its person.
var (
	nameEmpty   = operation.NewRefusal("name_empty", operation.ErrInvalid, "name is empty")
	nameTooLong = operation.NewRefusal("name_too_long", operation.ErrInvalid, "name is longer than 200 bytes")
	knocksWait  = operation.NewRefusal("handle_knocks_waiting", operation.ErrInvalid,
		"Knocks wait for you: answer them before you change your handle, since they name the addresses of your personal Room")
	roomCannotMove = operation.NewRefusalOf[MovedRoom]("handle_room_cannot_move", operation.ErrInvalid,
		"your Room {room} would be named {name}, which another Room holds or which is too long: choose another handle")
)

// MovedRoom fills the refusal of a Room that cannot take a new handle.
type MovedRoom struct {
	Room operation.Quoted `json:"room" jsonschema:"the name of the Room now"`
	Name operation.Quoted `json:"name" jsonschema:"the name it would take"`
}

// PersonName trims the name of a person, which is neither empty nor longer
// than the title of a Room. The administrators of the server have names of
// the same kind.
func PersonName(s string) (string, error) {
	name := strings.TrimSpace(s)
	if name == "" {
		return "", nameEmpty
	}
	if len(name) > maxName {
		return "", nameTooLong
	}
	return name, nil
}

// ChangeNameInput is the input of change_name.
type ChangeNameInput struct {
	Name string `json:"name" jsonschema:"your name, as the people of the Cahut see it"`
}

// ChangeName changes the name of the person who calls. Their personal Room
// takes it too, unless they gave it a title of its own. No AI changes it,
// as no AI changes an e-mail address.
var ChangeName = &operation.Op[ChangeNameInput, User]{
	Name:        "change_name",
	Title:       "Change your name",
	Description: "Change your name, as the people of the Cahut see it. Your personal Room takes it too, unless you gave it a title of its own.",
	PeopleOnly:  true,
	Run: func(ctx context.Context, c *operation.Call, in ChangeNameInput) (User, error) {
		if _, err := self(c, "change_name"); err != nil {
			return User{}, err
		}
		c.Log(c.Actor.Handle, "")
		name, err := PersonName(in.Name)
		if err != nil {
			return User{}, err
		}
		p, err := findUser(ctx, c.Tx, c.Actor.Handle)
		if err != nil {
			return User{}, err
		}
		c.Log(p.Handle, name)
		if _, err := c.Tx.ExecContext(ctx, "UPDATE users SET name = ? WHERE id = ?", name, p.ID); err != nil {
			return User{}, err
		}
		_, err = c.Tx.ExecContext(ctx, "UPDATE rooms SET title = ? WHERE name = ? AND kind = 'personal' AND title = ?",
			name, p.PersonalRoom, p.Name)
		if err != nil {
			return User{}, err
		}
		u := p.User
		u.Name = name
		return u, nil
	},
}

// ChangeHandleInput is the input of change_handle.
type ChangeHandleInput struct {
	Handle string `json:"handle" jsonschema:"your new handle, which names your personal Room too"`
}

// ChangeHandle changes the handle of the person who calls, with which they
// sign in. Their personal Room takes it as its name, so the addresses of
// its Windows change and the old ones stop working, and the old handle is
// free for another person or Room. In an individual Cahut, the names of
// their other Rooms start with their handle too, and take the new one. A
// Knock that waits for the person names the old addresses: they answer it
// first. Sessions, tokens and connected AIs hold the person and their
// Rooms, not their names, and keep working. See
// docs/adr/0039-people-change-their-name-and-handle.md.
var ChangeHandle = &operation.Op[ChangeHandleInput, User]{
	Name:  "change_handle",
	Title: "Change your handle",
	Description: "Change your handle, with which you sign in. Your personal Room takes it as its name: the addresses of its Windows change, " +
		"the old ones stop working, and your old handle is free for another person or Room. In an individual Cahut, your other Rooms, " +
		"whose names start with your handle, take the new one too. Answer the Knocks that wait for you first.",
	PeopleOnly: true,
	Run: func(ctx context.Context, c *operation.Call, in ChangeHandleInput) (User, error) {
		if _, err := self(c, "change_handle"); err != nil {
			return User{}, err
		}
		c.Log(c.Actor.Handle, in.Handle)
		p, err := findUser(ctx, c.Tx, c.Actor.Handle)
		if err != nil {
			return User{}, err
		}
		if in.Handle == p.Handle {
			return p.User, nil
		}
		if !validHandle.MatchString(in.Handle) {
			return User{}, InvalidHandle(in.Handle)
		}
		if address.Reserved(in.Handle) {
			return User{}, handleReserved(Handle{Handle: operation.Quoted(in.Handle)})
		}
		taken, err := handleTaken(ctx, c.Tx, in.Handle)
		if err != nil {
			return User{}, err
		}
		if taken {
			return User{}, handleIsTaken(Handle{Handle: operation.Quoted(in.Handle)})
		}
		var waiting int
		err = c.Tx.QueryRowContext(ctx, "SELECT count(*) FROM knocks WHERE user_id = ? AND state = 'waiting' AND expires_at > ?",
			p.ID, store.FormatTime(c.Now)).Scan(&waiting)
		if err != nil {
			return User{}, err
		}
		if waiting > 0 {
			return User{}, knocksWait
		}
		if _, err := c.Tx.ExecContext(ctx, "UPDATE users SET handle = ? WHERE id = ?", in.Handle, p.ID); err != nil {
			return User{}, err
		}
		var room int64
		err = c.Tx.QueryRowContext(ctx, "UPDATE rooms SET name = ? WHERE name = ? AND kind = 'personal' RETURNING id",
			in.Handle, p.PersonalRoom).Scan(&room)
		if err != nil {
			return User{}, err
		}
		c.InRoom(room)
		if err := moveOwnRooms(ctx, c, p.ID, p.Handle, in.Handle); err != nil {
			return User{}, err
		}
		u := p.User
		u.Handle, u.PersonalRoom = in.Handle, in.Handle
		return u, mail.OutboxOf(c).Notify(ctx, c, p.ID, mail.Always, handleChangedLetter(p.Handle))
	},
}

// moveOwnRooms gives the new handle to the Rooms of a person in an
// individual Cahut, whose names start with their handle and a hyphen. See
// docs/adr/0030-kind-and-limits-of-each-cahut.md.
func moveOwnRooms(ctx context.Context, c *operation.Call, userID int64, old, handle string) error {
	individual, err := settings.IsIndividual(ctx, c.Tx)
	if err != nil || !individual {
		return err
	}
	prefix := old + "-"
	rows, err := c.Tx.QueryContext(ctx, `SELECT r.id, r.name FROM rooms r JOIN room_members m ON m.room_id = r.id
		WHERE m.user_id = ? AND m.role = 'owner' AND r.kind = 'team' AND substr(r.name, 1, ?) = ? ORDER BY r.name`,
		userID, len(prefix), prefix)
	if err != nil {
		return err
	}
	type room struct {
		id         int64
		name, next string
	}
	var rooms []room
	for rows.Next() {
		var r room
		if err := rows.Scan(&r.id, &r.name); err != nil {
			rows.Close()
			return err
		}
		r.next = handle + "-" + strings.TrimPrefix(r.name, prefix)
		rooms = append(rooms, r)
	}
	if err := errors.Join(rows.Err(), rows.Close()); err != nil {
		return err
	}
	for _, r := range rooms {
		taken := !address.ValidName(r.next)
		if !taken {
			if err := c.Tx.QueryRowContext(ctx, "SELECT EXISTS (SELECT 1 FROM rooms WHERE name = ?)", r.next).Scan(&taken); err != nil {
				return err
			}
		}
		if taken {
			return roomCannotMove(MovedRoom{Room: operation.Quoted(r.name), Name: operation.Quoted(r.next)})
		}
		if _, err := c.Tx.ExecContext(ctx, "UPDATE rooms SET name = ? WHERE id = ?", r.next, r.id); err != nil {
			return err
		}
	}
	return nil
}

func handleChangedLetter(old string) func(p mail.Person, name, where string) mail.Letter {
	return func(p mail.Person, name, where string) mail.Letter {
		return mail.Letter{
			Subject: "Your handle changed",
			Lines: []string{"Hello " + p.Name + ",", "Your handle on " + where + " is now " + p.Handle + ", in place of " + old +
				": sign in with it. Your personal Room took it as its name, and the links to its Windows under " + old + " no longer work."},
			After: []string{"If you did not change it, ask an administrator of " + name + " now."},
		}
	}
}
