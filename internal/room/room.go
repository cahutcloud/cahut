// Package room manages the Rooms of a Cahut: the personal Room of each
// person, and team Rooms with their owners and members. A Room is at
// <domain>/<room>/. See docs/adr/0009-rooms-and-windows-as-paths.md and
// docs/adr/0013-people-sessions-and-rights.md.
package room

import (
	"cmp"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"cahut.cloud/internal/address"
	"cahut.cloud/internal/mail"
	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/settings"
	"cahut.cloud/internal/store"
)

// Kinds of Rooms, and roles in a Room.
const (
	Personal = "personal"
	Team     = "team"

	Owner  = "owner"
	Member = "member"
)

// Register adds the Room operations to r.
func Register(r *operation.Registry) {
	operation.Register(r, CreateRoom)
	operation.Register(r, ListRooms)
	operation.Register(r, GetRoom)
	operation.Register(r, UpdateRoom)
	operation.Register(r, SetHome)
	operation.Register(r, AddMember)
	operation.Register(r, SetMemberRole)
	operation.Register(r, RemoveMember)
}

// Access is a Room as a call reaches it.
type Access struct {
	ID    int64
	Name  string
	Title string
	Kind  string
	Role  string // the role in the Room of the person the call acts for; empty when they are not in it
}

// Find returns the Room called name, with the role in it of the person the
// call acts for. A Room that the caller may not see answers as one that does
// not exist. Administrators see every Room.
func Find(ctx context.Context, c *operation.Call, name string) (Access, error) {
	var a Access
	var role sql.NullString
	err := c.Tx.QueryRowContext(ctx, `
		SELECT r.id, r.name, r.title, r.kind, m.role FROM rooms r
		LEFT JOIN room_members m ON m.room_id = r.id AND m.user_id = ?
		WHERE r.name = ?`, c.Actor.UserID, name).Scan(&a.ID, &a.Name, &a.Title, &a.Kind, &role)
	if errors.Is(err, sql.ErrNoRows) || err == nil && !role.Valid && !c.Admin {
		return Access{}, operation.NotFoundf("no Room %q among yours", name)
	}
	if err != nil {
		return Access{}, err
	}
	if !c.Actor.Reaches(a.ID) {
		return Access{}, operation.NotFoundf("no Room %q among those you were granted: "+
			"%s can grant you more Rooms by connecting you again", name, c.Actor.Handle)
	}
	a.Role = role.String
	return a, nil
}

// Require returns the Room called name when the caller may act in it as
// role: a Member publishes, an Owner manages the Room. Administrators manage
// every Room, but publish only in their own.
func Require(ctx context.Context, c *operation.Call, name, role string) (Access, error) {
	a, err := Find(ctx, c, name)
	if err != nil {
		return Access{}, err
	}
	switch {
	case role == Owner && (a.Role == Owner || c.Admin):
	case role == Member && a.Role != "":
	case a.Role == "":
		return Access{}, operation.Forbiddenf("you are not in the Room %s: its owners can add you", a.Name)
	default:
		return Access{}, operation.Forbiddenf("only the owners of the Room %s, and the administrators of the Cahut, can do this", a.Name)
	}
	return a, nil
}

// Of returns the Rooms of the person userID, their personal Room first,
// for the pages that readers open outside of any operation.
func Of(ctx context.Context, db *store.DB, userID int64) ([]Access, error) {
	rows, err := db.Read.QueryContext(ctx, `
		SELECT r.id, r.name, r.title, r.kind, m.role FROM rooms r JOIN room_members m ON m.room_id = r.id
		WHERE m.user_id = ? ORDER BY r.kind = 'team', r.name`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var rooms []Access
	for rows.Next() {
		var a Access
		if err := rows.Scan(&a.ID, &a.Name, &a.Title, &a.Kind, &a.Role); err != nil {
			return nil, err
		}
		rooms = append(rooms, a)
	}
	return rooms, rows.Err()
}

// Room describes a Room.
type Room struct {
	Name      string    `json:"name" jsonschema:"the Room's name, first part of its address"`
	Title     string    `json:"title"`
	Kind      string    `json:"kind" jsonschema:"personal or team"`
	Role      string    `json:"role,omitempty" jsonschema:"your role in the Room, owner or member; empty when you are not in it"`
	Members   int       `json:"members" jsonschema:"how many people are in the Room"`
	Windows   int       `json:"windows" jsonschema:"how many Windows the Room holds, out of the trash"`
	Home      string    `json:"home,omitempty" jsonschema:"the Window shown at the Room's address, in place of the list of its Windows; empty when the address lists them"`
	CreatedAt time.Time `json:"created_at"`
}

// homeName is the name of the home Window of the Room r, out of the trash,
// for the queries that describe Rooms.
const homeName = "(SELECT name FROM windows WHERE id = r.home_window_id AND trashed_at IS NULL)"

// CreateRoomInput is the input of create_room.
type CreateRoomInput struct {
	Title string `json:"title" jsonschema:"title of the Room, such as Marketing"`
	Name  string `json:"name,omitempty" jsonschema:"name of the Room, the first part of its address; from the title when empty"`
	Owner string `json:"owner,omitempty" jsonschema:"handle of its first owner; you when empty. Only administrators of the Cahut name someone else"`
}

// CreateRoom creates a team Room, whose first owner is the person who calls
// it, or the one an administrator names.
var CreateRoom = &operation.Op[CreateRoomInput, Room]{
	Name:  "create_room",
	Title: "Create a team Room",
	Description: "Create a team Room, at <domain>/<name>/. You become its owner, and can then add members. " +
		"Windows published in it are for its members. In an individual Cahut, you stay its only member, " +
		"and its name starts with your handle and a hyphen.",
	Run: func(ctx context.Context, c *operation.Call, in CreateRoomInput) (Room, error) {
		title := strings.TrimSpace(in.Title)
		if title == "" {
			return Room{}, operation.Invalidf("title is empty")
		}
		if len(title) > 200 {
			return Room{}, operation.Invalidf("title is longer than 200 bytes")
		}
		name := in.Name
		if name == "" {
			name = address.Slug(title, "")
		}
		c.Log(name, title)
		if c.Actor.Limited() {
			return Room{}, operation.Forbiddenf("you were granted some Rooms only, and a new Room would be outside them: "+
				"%s can create it in the console", c.Actor.Handle)
		}
		owner, err := owner(ctx, c, in.Owner)
		if err != nil {
			return Room{}, err
		}
		individual, err := settings.IsIndividual(ctx, c.Tx)
		if err != nil {
			return Room{}, err
		}
		if individual {
			name = ownName(cmp.Or(in.Owner, c.Actor.Handle), in.Name, title)
			c.Log(name, title)
		}
		switch {
		case !address.ValidName(name):
			return Room{}, operation.Invalidf("name %q: use 1 to 63 lowercase letters, digits and hyphens, "+
				"starting and ending with a letter or a digit", name)
		case address.Reserved(name):
			return Room{}, operation.Invalidf("name %q is reserved: it names a page of every Cahut", name)
		}
		var taken, theirs bool
		err = c.Tx.QueryRowContext(ctx, `
			SELECT EXISTS (SELECT 1 FROM rooms WHERE name = ?1) OR EXISTS (SELECT 1 FROM users WHERE handle = ?1),
			  EXISTS (SELECT 1 FROM rooms r JOIN room_members m ON m.room_id = r.id WHERE r.name = ?1 AND m.user_id = ?2)`,
			name, owner).Scan(&taken, &theirs)
		if err != nil {
			return Room{}, err
		}
		switch {
		case taken && individual && theirs:
			return Room{}, operation.Invalidf("you already have a Room %q: choose another name", name)
		case taken && individual:
			// People of an individual Cahut do not learn what others hold.
			return Room{}, operation.Invalidf("name %q is not available: choose another", name)
		case taken:
			return Room{}, operation.Invalidf("name %q is taken, by a Room or a person: choose another", name)
		}
		if err := settings.CheckRoom(ctx, c, owner); err != nil {
			return Room{}, err
		}
		r := Room{Name: name, Title: title, Kind: Team, Members: 1, CreatedAt: c.Now}
		var id int64
		err = c.Tx.QueryRowContext(ctx, "INSERT INTO rooms (name, kind, title, created_at) VALUES (?, 'team', ?, ?) RETURNING id",
			r.Name, r.Title, store.FormatTime(c.Now)).Scan(&id)
		if err != nil {
			return Room{}, err
		}
		if _, err := c.Tx.ExecContext(ctx, "INSERT INTO room_members (room_id, user_id, role) VALUES (?, ?, 'owner')", id, owner); err != nil {
			return Room{}, err
		}
		if owner == c.Actor.UserID {
			r.Role = Owner
		}
		c.InRoom(id)
		return r, nil
	},
}

// ownName returns the name of a Room of the person handle in an individual
// Cahut, where it starts with their handle and a hyphen: name with the
// prefix added when it lacks it, or one made from the title. See
// docs/adr/0030-kind-and-limits-of-each-cahut.md.
func ownName(handle, name, title string) string {
	prefix := handle + "-"
	if name != "" {
		if strings.HasPrefix(name, prefix) {
			return name
		}
		return prefix + name
	}
	slug := address.Slug(title, "")
	if slug == "" {
		return ""
	}
	const longest = 63 // the longest name, as address.ValidName has it
	if room := longest - len(prefix); len(slug) > room {
		slug = strings.TrimRight(slug[:room], "-")
	}
	return prefix + slug
}

// owner returns the id of the person who becomes the first owner of a Room:
// the one named by handle, or the person the call acts for.
func owner(ctx context.Context, c *operation.Call, handle string) (int64, error) {
	if handle == "" {
		if c.Actor.UserID == 0 {
			return 0, operation.Invalidf("owner is empty: name the person who owns the Room")
		}
		return c.Actor.UserID, nil
	}
	if handle != c.Actor.Handle && !c.Admin {
		return 0, operation.Forbiddenf("only administrators of the Cahut create Rooms for others")
	}
	return userID(ctx, c.Tx, handle)
}

func userID(ctx context.Context, tx *sql.Tx, handle string) (int64, error) {
	var id int64
	err := tx.QueryRowContext(ctx, "SELECT id FROM users WHERE handle = ?", handle).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, operation.NotFoundf("no person %q in this Cahut", handle)
	}
	return id, err
}

// ListRoomsInput is the input of list_rooms.
type ListRoomsInput struct {
	All bool `json:"all,omitempty" jsonschema:"every Room of the Cahut, for an administrator; your Rooms when false"`
}

// List is the output of list_rooms.
type List struct {
	Rooms []Room `json:"rooms"`
}

// ListRooms lists the Rooms of the person the call acts for, or every Room
// for an administrator.
var ListRooms = &operation.Op[ListRoomsInput, List]{
	Name:        "list_rooms",
	Title:       "List Rooms",
	Description: "List your Rooms, your personal Room first, with your role and how many members and Windows each has.",
	ReadOnly:    true,
	Run: func(ctx context.Context, c *operation.Call, in ListRoomsInput) (List, error) {
		if in.All && !c.Admin {
			return List{}, operation.Forbiddenf("only administrators of the Cahut list every Room")
		}
		where := "m.role IS NOT NULL"
		if in.All {
			where = "1 = 1"
		}
		rows, err := c.Tx.QueryContext(ctx, `
			SELECT r.id, r.name, r.title, r.kind, m.role, r.created_at,
				(SELECT count(*) FROM room_members WHERE room_id = r.id),
				(SELECT count(*) FROM windows WHERE room_id = r.id AND trashed_at IS NULL), `+homeName+`
			FROM rooms r LEFT JOIN room_members m ON m.room_id = r.id AND m.user_id = ?
			WHERE `+where+`
			ORDER BY r.kind = 'team', m.role IS NULL, r.name`, c.Actor.UserID)
		if err != nil {
			return List{}, err
		}
		defer rows.Close()
		list := List{Rooms: []Room{}}
		for rows.Next() {
			var r Room
			var id int64
			var role, home sql.NullString
			var created string
			if err := rows.Scan(&id, &r.Name, &r.Title, &r.Kind, &role, &created, &r.Members, &r.Windows, &home); err != nil {
				return List{}, err
			}
			if !c.Actor.Reaches(id) {
				continue
			}
			r.Role, r.Home = role.String, home.String
			if r.CreatedAt, err = store.ParseTime(created); err != nil {
				return List{}, err
			}
			list.Rooms = append(list.Rooms, r)
		}
		c.Log("", fmt.Sprintf("%d Rooms", len(list.Rooms)))
		return list, rows.Err()
	},
}

// GetInput is the input of get_room.
type GetInput struct {
	Room string `json:"room" jsonschema:"the Room's name, first part of its address"`
}

// Detail is the output of get_room.
type Detail struct {
	Room
	People []Person `json:"people" jsonschema:"the people in the Room, owners first"`
}

// Person is a person in a Room.
type Person struct {
	Handle string `json:"handle"`
	Name   string `json:"name"`
	Role   string `json:"role" jsonschema:"owner or member of the Room"`
}

// GetRoom describes a Room and lists its members.
var GetRoom = &operation.Op[GetInput, Detail]{
	Name:        "get_room",
	Title:       "Describe a Room",
	Description: "Describe a Room: its title, its owners and members, and how many Windows it holds.",
	ReadOnly:    true,
	Run: func(ctx context.Context, c *operation.Call, in GetInput) (Detail, error) {
		c.Log(in.Room, "")
		a, err := Find(ctx, c, in.Room)
		if err != nil {
			return Detail{}, err
		}
		c.InRoom(a.ID)
		return detail(ctx, c, a)
	},
}

// detail describes the Room a.
func detail(ctx context.Context, c *operation.Call, a Access) (Detail, error) {
	d := Detail{Room: Room{Name: a.Name, Title: a.Title, Kind: a.Kind, Role: a.Role}, People: []Person{}}
	var created string
	var home sql.NullString
	err := c.Tx.QueryRowContext(ctx, `
		SELECT r.created_at, (SELECT count(*) FROM windows WHERE room_id = r.id AND trashed_at IS NULL), `+homeName+`
		FROM rooms r WHERE r.id = ?`, a.ID).Scan(&created, &d.Windows, &home)
	if err != nil {
		return Detail{}, err
	}
	d.Home = home.String
	if d.CreatedAt, err = store.ParseTime(created); err != nil {
		return Detail{}, err
	}
	rows, err := c.Tx.QueryContext(ctx, `
		SELECT u.handle, u.name, m.role FROM room_members m JOIN users u ON u.id = m.user_id
		WHERE m.room_id = ? ORDER BY m.role = 'member', u.handle`, a.ID)
	if err != nil {
		return Detail{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var p Person
		if err := rows.Scan(&p.Handle, &p.Name, &p.Role); err != nil {
			return Detail{}, err
		}
		d.People = append(d.People, p)
	}
	d.Members = len(d.People)
	return d, rows.Err()
}

// UpdateRoomInput is the input of update_room.
type UpdateRoomInput struct {
	Room  string `json:"room" jsonschema:"the Room's name"`
	Title string `json:"title" jsonschema:"the new title"`
}

// UpdateRoom changes the title of a Room.
var UpdateRoom = &operation.Op[UpdateRoomInput, Detail]{
	Name:        "update_room",
	Title:       "Retitle a Room",
	Description: "Change the title of a Room. Its name, and so its address, stays.",
	Run: func(ctx context.Context, c *operation.Call, in UpdateRoomInput) (Detail, error) {
		title := strings.TrimSpace(in.Title)
		c.Log(in.Room, title)
		if title == "" {
			return Detail{}, operation.Invalidf("title is empty")
		}
		if len(title) > 200 {
			return Detail{}, operation.Invalidf("title is longer than 200 bytes")
		}
		a, err := Require(ctx, c, in.Room, Owner)
		if err != nil {
			return Detail{}, err
		}
		c.InRoom(a.ID)
		if _, err := c.Tx.ExecContext(ctx, "UPDATE rooms SET title = ? WHERE id = ?", title, a.ID); err != nil {
			return Detail{}, err
		}
		a.Title = title
		return detail(ctx, c, a)
	},
}

// SetHomeInput is the input of set_room_home.
type SetHomeInput struct {
	Room   string `json:"room" jsonschema:"the Room's name"`
	Window string `json:"window,omitempty" jsonschema:"name of a Window of the Room, shown at the Room's address; empty to go back to the page that lists the Room's Windows"`
}

// SetHome chooses the Window that a Room shows at its address, in place of
// the page that lists its Windows, or goes back to that page. Readers who
// may not open the Window still get the list of those they may open. See
// docs/adr/0032-the-page-of-a-room.md.
var SetHome = &operation.Op[SetHomeInput, Detail]{
	Name:  "set_room_home",
	Title: "Choose the home page of a Room",
	Description: "Show a Window of a Room at the Room's address, <domain>/<room>/, in place of the page that lists its Windows, " +
		"such as a landing page made for the Room. Readers who may not open that Window still get the list of the Windows they may open. " +
		"An empty window goes back to the list. The Window keeps its own address and its access.",
	Run: func(ctx context.Context, c *operation.Call, in SetHomeInput) (Detail, error) {
		c.Log(in.Room, cmp.Or(in.Window, "the list of its Windows"))
		a, err := Require(ctx, c, in.Room, Owner)
		if err != nil {
			return Detail{}, err
		}
		c.InRoom(a.ID)
		var home sql.NullInt64
		if in.Window != "" {
			err := c.Tx.QueryRowContext(ctx, "SELECT id FROM windows WHERE room_id = ? AND name = ? AND trashed_at IS NULL",
				a.ID, in.Window).Scan(&home)
			if errors.Is(err, sql.ErrNoRows) {
				return Detail{}, operation.NotFoundf("no Window %q in the Room %s, out of the trash: list_windows gives the names of its Windows", in.Window, a.Name)
			}
			if err != nil {
				return Detail{}, err
			}
		}
		if _, err := c.Tx.ExecContext(ctx, "UPDATE rooms SET home_window_id = ? WHERE id = ?", home, a.ID); err != nil {
			return Detail{}, err
		}
		return detail(ctx, c, a)
	},
}

// MemberInput names a person in a Room, with a role.
type MemberInput struct {
	Room string `json:"room" jsonschema:"the Room's name"`
	User string `json:"user" jsonschema:"handle of the person"`
	Role string `json:"role,omitempty" jsonschema:"owner or member; member when empty"`
}

// AddMember adds a person to a team Room.
var AddMember = &operation.Op[MemberInput, Detail]{
	Name:         "add_room_member",
	Title:        "Add a member to a Room",
	Description:  "Add a person to a team Room, as a member or an owner. They can then open and publish its Windows.",
	KnockFirst:   true,
	InIndividual: operation.Nobody,
	Run: func(ctx context.Context, c *operation.Call, in MemberInput) (Detail, error) {
		role := in.Role
		if role == "" {
			role = Member
		}
		a, id, err := member(ctx, c, in.Room, in.User, role)
		if err != nil {
			return Detail{}, err
		}
		var already bool
		err = c.Tx.QueryRowContext(ctx, "SELECT EXISTS (SELECT 1 FROM room_members WHERE room_id = ? AND user_id = ?)", a.ID, id).Scan(&already)
		if err != nil {
			return Detail{}, err
		}
		if already {
			return Detail{}, operation.Invalidf("%s is already in the Room %s", in.User, a.Name)
		}
		if _, err := c.Tx.ExecContext(ctx, "INSERT INTO room_members (room_id, user_id, role) VALUES (?, ?, ?)", a.ID, id, role); err != nil {
			return Detail{}, err
		}
		if err := notifyAdded(ctx, c, a, id, role); err != nil {
			return Detail{}, err
		}
		return detail(ctx, c, a)
	},
}

// SetMemberRole makes a person of a team Room an owner or a member.
var SetMemberRole = &operation.Op[MemberInput, Detail]{
	Name:         "set_room_member_role",
	Title:        "Change a role in a Room",
	Description:  "Make a person of a team Room an owner or a member of it. A Room keeps at least one owner.",
	KnockFirst:   true,
	InIndividual: operation.Nobody,
	Run: func(ctx context.Context, c *operation.Call, in MemberInput) (Detail, error) {
		a, id, err := member(ctx, c, in.Room, in.User, in.Role)
		if err != nil {
			return Detail{}, err
		}
		current, err := roleOf(ctx, c.Tx, a, id, in.User)
		if err != nil {
			return Detail{}, err
		}
		if current == Owner && in.Role != Owner {
			if err := keepAnOwner(ctx, c.Tx, a); err != nil {
				return Detail{}, err
			}
		}
		if _, err := c.Tx.ExecContext(ctx, "UPDATE room_members SET role = ? WHERE room_id = ? AND user_id = ?", in.Role, a.ID, id); err != nil {
			return Detail{}, err
		}
		return detail(ctx, c, a)
	},
}

// RemoveMemberInput is the input of remove_room_member.
type RemoveMemberInput struct {
	Room string `json:"room" jsonschema:"the Room's name"`
	User string `json:"user" jsonschema:"handle of the person; yourself to leave the Room"`
}

// RemoveMember removes a person from a team Room. Anyone can leave a Room.
var RemoveMember = &operation.Op[RemoveMemberInput, Detail]{
	Name:  "remove_room_member",
	Title: "Remove a member from a Room",
	Description: "Remove a person from a team Room, or leave it yourself. They no longer open or publish its Windows. " +
		"A Room keeps at least one owner.",
	KnockFirst:   true,
	InIndividual: operation.Nobody,
	Run: func(ctx context.Context, c *operation.Call, in RemoveMemberInput) (Detail, error) {
		c.Log(in.Room, in.User)
		need := Owner
		if in.User != "" && in.User == c.Actor.Handle {
			need = Member // leaving
		}
		a, err := Require(ctx, c, in.Room, need)
		if err != nil {
			return Detail{}, err
		}
		c.InRoom(a.ID)
		if a.Kind == Personal {
			return Detail{}, operation.Invalidf("%s is a personal Room: it keeps its person", a.Name)
		}
		id, err := userID(ctx, c.Tx, in.User)
		if err != nil {
			return Detail{}, err
		}
		current, err := roleOf(ctx, c.Tx, a, id, in.User)
		if err != nil {
			return Detail{}, err
		}
		if current == Owner {
			if err := keepAnOwner(ctx, c.Tx, a); err != nil {
				return Detail{}, err
			}
		}
		if _, err := c.Tx.ExecContext(ctx, "DELETE FROM room_members WHERE room_id = ? AND user_id = ?", a.ID, id); err != nil {
			return Detail{}, err
		}
		if id == c.Actor.UserID && !c.Admin {
			a.Role = ""
			return Detail{Room: Room{Name: a.Name, Title: a.Title, Kind: a.Kind}, People: []Person{}}, nil
		}
		return detail(ctx, c, a)
	},
}

// member checks a change to the people of the team Room called room by its
// owners or the administrators: the role, then the Room. It returns the Room
// and the id of the person with handle user.
func member(ctx context.Context, c *operation.Call, room, user, role string) (Access, int64, error) {
	c.Log(room, user+" "+role)
	if role != Owner && role != Member {
		return Access{}, 0, operation.Invalidf("role %q: use owner or member", role)
	}
	a, err := Require(ctx, c, room, Owner)
	if err != nil {
		return Access{}, 0, err
	}
	c.InRoom(a.ID)
	if a.Kind == Personal {
		return Access{}, 0, operation.Invalidf("%s is a personal Room: it has its person only; create a team Room to work with others", a.Name)
	}
	id, err := userID(ctx, c.Tx, user)
	return a, id, err
}

// roleOf returns the role of the person id in the Room a.
func roleOf(ctx context.Context, tx *sql.Tx, a Access, id int64, handle string) (string, error) {
	var role string
	err := tx.QueryRowContext(ctx, "SELECT role FROM room_members WHERE room_id = ? AND user_id = ?", a.ID, id).Scan(&role)
	if errors.Is(err, sql.ErrNoRows) {
		return "", operation.NotFoundf("%s is not in the Room %s", handle, a.Name)
	}
	return role, err
}

// keepAnOwner returns an error wrapping operation.ErrInvalid when the Room a
// has a single owner.
func keepAnOwner(ctx context.Context, tx *sql.Tx, a Access) error {
	var owners int
	if err := tx.QueryRowContext(ctx, "SELECT count(*) FROM room_members WHERE room_id = ? AND role = 'owner'", a.ID).Scan(&owners); err != nil {
		return err
	}
	if owners < 2 {
		return operation.Invalidf("the Room %s needs an owner: make another person one first", a.Name)
	}
	return nil
}

// notifyAdded tells the person userID, unless they turned such notices
// off, that they are now in the Room a, with role, and who added them.
func notifyAdded(ctx context.Context, c *operation.Call, a Access, userID int64, role string) error {
	o := mail.OutboxOf(c)
	if o == nil {
		return nil
	}
	by := "An administrator of the server"
	if c.Actor.UserID != 0 {
		if err := c.Tx.QueryRowContext(ctx, "SELECT name FROM users WHERE id = ?", c.Actor.UserID).Scan(&by); err != nil {
			return err
		}
	}
	return o.Notify(ctx, c, userID, mail.Optional, func(p mail.Person, _, where string) mail.Letter {
		return mail.Letter{
			Subject: "You were added to " + a.Title,
			Lines:   []string{"Hello " + p.Name + ",", by + " added you to the Room " + a.Title + " on " + where + ", as " + role + "."},
			Action:  &mail.Action{Label: "Open " + a.Title, URL: o.Base + a.Name + "/"},
			After:   []string{"You can turn off these notices on the page of your account, in the console."},
		}
	})
}
