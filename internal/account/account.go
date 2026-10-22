// Package account manages the people of a Cahut: their accounts and roles,
// how they sign in, and the personal tokens that their AIs present.
// See docs/adr/0013-people-sessions-and-rights.md.
package account

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"cahut.cloud/internal/address"
	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/store"
)

// Register adds the account operations to r, and lets r check people.
func Register(r *operation.Registry) {
	r.CheckPerson = Check
	operation.Register(r, CreateUser)
	operation.Register(r, ListUsers)
	operation.Register(r, GetAccount)
	operation.Register(r, SetUserRole)
	operation.Register(r, DisableUser)
	operation.Register(r, EnableUser)
	operation.Register(r, CreateActivationLink)
	operation.Register(r, ActivateAccount)
	operation.Register(r, SignIn)
	operation.Register(r, SignOut)
	operation.Register(r, ChangePassword)
	operation.Register(r, SignInToWindows)
	operation.Register(r, CreateToken)
	operation.Register(r, ListTokens)
	operation.Register(r, RevokeToken)
}

// Roles of a person in a Cahut.
const (
	Admin  = "admin"
	Member = "member"
)

// States of an account.
const (
	Pending  = "pending" // no password yet: the person has not used their activation link
	Active   = "active"
	Disabled = "disabled"
)

// User is a person of a Cahut.
type User struct {
	ID           int64  `json:"id"`
	Handle       string `json:"handle"`
	Name         string `json:"name"`
	Role         string `json:"role" jsonschema:"admin or member of the Cahut"`
	Status       string `json:"status" jsonschema:"pending until the person chooses a password, then active, or disabled"`
	PersonalRoom string `json:"personal_room" jsonschema:"name of the person's personal Room, the handle"`
}

// person is a person as the operations read them.
type person struct {
	User
	passwordHash string
}

// findUser returns the person with handle, or an error wrapping
// operation.ErrNotFound.
func findUser(ctx context.Context, tx *sql.Tx, handle string) (person, error) {
	var p person
	var hash, disabled sql.NullString
	err := tx.QueryRowContext(ctx, "SELECT id, handle, name, role, password_hash, disabled_at FROM users WHERE handle = ?",
		handle).Scan(&p.ID, &p.Handle, &p.Name, &p.Role, &hash, &disabled)
	if errors.Is(err, sql.ErrNoRows) {
		return person{}, operation.NotFoundf("no person %q in this Cahut", handle)
	}
	if err != nil {
		return person{}, err
	}
	p.passwordHash = hash.String
	p.Status = status(hash, disabled)
	p.PersonalRoom = p.Handle
	return p, nil
}

func status(passwordHash, disabledAt sql.NullString) string {
	switch {
	case disabledAt.Valid:
		return Disabled
	case !passwordHash.Valid:
		return Pending
	}
	return Active
}

// Check tells the registry whether the person userID administers the Cahut,
// and refuses a disabled person. It is the registry's CheckPerson.
func Check(ctx context.Context, tx *sql.Tx, userID int64) (bool, error) {
	var role string
	var disabled sql.NullString
	err := tx.QueryRowContext(ctx, "SELECT role, disabled_at FROM users WHERE id = ?", userID).Scan(&role, &disabled)
	if errors.Is(err, sql.ErrNoRows) {
		return false, operation.Forbiddenf("this person is not in the Cahut")
	}
	if err != nil {
		return false, err
	}
	if disabled.Valid {
		return false, operation.Forbiddenf("this person is disabled: an administrator of the Cahut can enable them again")
	}
	return role == Admin, nil
}

// self returns the person the actor is or acts for, or an error when the
// actor is the administrator on the command line, who is not a person.
func self(c *operation.Call, op string) (int64, error) {
	if c.Actor.UserID == 0 {
		return 0, operation.Invalidf("%s acts for a person, and the administrator is not one", op)
	}
	return c.Actor.UserID, nil
}

// CreateUserInput is the input of create_user.
type CreateUserInput struct {
	Handle string `json:"handle"`
	Name   string `json:"name"`
	Admin  bool   `json:"admin,omitempty"`
}

// NewUser is the output of create_user.
type NewUser struct {
	User
	Activation *Activation `json:"activation,omitempty" jsonschema:"the link where the person chooses their password; an AI never gets it"`
}

// A handle is also the name of the person's personal Room, so it is a valid
// Room name too.
var validHandle = regexp.MustCompile(`^[a-z][a-z0-9-]{0,30}[a-z0-9]$`)

// ValidHandle reports whether s can be a handle: 2 to 32 lowercase letters,
// digits and hyphens, starting with a letter and ending with a letter or a
// digit. The administrators of the server have handles of the same form.
func ValidHandle(s string) bool {
	return validHandle.MatchString(s)
}

// CreateUser creates a person and their personal Room, named after their
// handle, with an activation link.
var CreateUser = &operation.Op[CreateUserInput, NewUser]{
	Name:        "create_user",
	Description: "Create a person, with their personal Room, and the link where they choose their password.",
	Level:       operation.CahutAdmin,
	KnockFirst:  true,
	Run: func(ctx context.Context, c *operation.Call, in CreateUserInput) (NewUser, error) {
		if !validHandle.MatchString(in.Handle) {
			return NewUser{}, operation.Invalidf("handle %q: use 2 to 32 lowercase letters, digits and hyphens, "+
				"starting with a letter and ending with a letter or a digit", in.Handle)
		}
		if address.Reserved(in.Handle) {
			return NewUser{}, operation.Invalidf("handle %q is reserved: it names a page of every Cahut", in.Handle)
		}
		name := strings.TrimSpace(in.Name)
		if name == "" {
			return NewUser{}, operation.Invalidf("name is empty")
		}
		u := NewUser{User: User{Handle: in.Handle, Name: name, Role: Member, Status: Pending, PersonalRoom: in.Handle}}
		if in.Admin {
			u.Role = Admin
		}
		c.Log(u.Handle, u.Role)

		var taken bool
		err := c.Tx.QueryRowContext(ctx, `
			SELECT EXISTS (SELECT 1 FROM users WHERE handle = ?) OR EXISTS (SELECT 1 FROM rooms WHERE name = ?)`,
			u.Handle, u.Handle).Scan(&taken)
		if err != nil {
			return NewUser{}, err
		}
		if taken {
			return NewUser{}, operation.Invalidf("handle %q is taken, by a person or a Room", u.Handle)
		}
		now := store.FormatTime(c.Now)
		err = c.Tx.QueryRowContext(ctx, "INSERT INTO users (handle, name, role, created_at) VALUES (?, ?, ?, ?) RETURNING id",
			u.Handle, u.Name, u.Role, now).Scan(&u.ID)
		if err != nil {
			return NewUser{}, err
		}
		var room int64
		err = c.Tx.QueryRowContext(ctx, "INSERT INTO rooms (name, kind, title, created_at) VALUES (?, 'personal', ?, ?) RETURNING id",
			u.PersonalRoom, u.Name, now).Scan(&room)
		if err != nil {
			return NewUser{}, err
		}
		_, err = c.Tx.ExecContext(ctx, "INSERT INTO room_members (room_id, user_id, role) VALUES (?, ?, 'owner')",
			room, u.ID)
		if err != nil {
			return NewUser{}, err
		}
		c.InRoom(room) // the person reads how their account began
		// An AI never gets the code: it could choose the password itself.
		if c.Actor.Kind != operation.AI {
			a, err := newActivation(ctx, c, u.ID, u.Handle)
			if err != nil {
				return NewUser{}, err
			}
			u.Activation = &a
		}
		return u, nil
	},
}

// ListUsersInput is the input of list_users.
type ListUsersInput struct{}

// UserList is the output of list_users.
type UserList struct {
	Users []User `json:"users"`
}

// ListUsers lists the people of the Cahut, by handle.
var ListUsers = &operation.Op[ListUsersInput, UserList]{
	Name:        "list_users",
	Description: "List the people of the Cahut, with their handles, names, roles and states.",
	ReadOnly:    true,
	Run: func(ctx context.Context, c *operation.Call, _ ListUsersInput) (UserList, error) {
		rows, err := c.Tx.QueryContext(ctx, "SELECT id, handle, name, role, password_hash, disabled_at FROM users ORDER BY handle")
		if err != nil {
			return UserList{}, err
		}
		defer rows.Close()
		list := UserList{Users: []User{}}
		for rows.Next() {
			var u User
			var hash, disabled sql.NullString
			if err := rows.Scan(&u.ID, &u.Handle, &u.Name, &u.Role, &hash, &disabled); err != nil {
				return UserList{}, err
			}
			u.Status, u.PersonalRoom = status(hash, disabled), u.Handle
			list.Users = append(list.Users, u)
		}
		c.Log("", fmt.Sprintf("%d people", len(list.Users)))
		return list, rows.Err()
	},
}

// GetAccountInput is the input of get_account.
type GetAccountInput struct{}

// Account is the output of get_account.
type Account struct {
	User
	Rooms []Membership `json:"rooms" jsonschema:"the Rooms of the person, their personal Room first"`
}

// Membership is a Room of a person.
type Membership struct {
	Room  string `json:"room" jsonschema:"the Room's name, first part of its address"`
	Title string `json:"title"`
	Kind  string `json:"kind" jsonschema:"personal or team"`
	Role  string `json:"role" jsonschema:"owner or member of the Room"`
}

// GetAccount describes the person the actor is, or acts for.
var GetAccount = &operation.Op[GetAccountInput, Account]{
	Name:        "get_account",
	Description: "Describe the person you act for: their handle, name, role in the Cahut and Rooms.",
	ReadOnly:    true,
	Run: func(ctx context.Context, c *operation.Call, _ GetAccountInput) (Account, error) {
		if _, err := self(c, "get_account"); err != nil {
			return Account{}, err
		}
		c.Log(c.Actor.Handle, "")
		p, err := findUser(ctx, c.Tx, c.Actor.Handle)
		if err != nil {
			return Account{}, err
		}
		rows, err := c.Tx.QueryContext(ctx, `
			SELECT r.id, r.name, r.title, r.kind, m.role FROM rooms r JOIN room_members m ON m.room_id = r.id
			WHERE m.user_id = ? ORDER BY r.kind = 'team', r.name`, p.ID)
		if err != nil {
			return Account{}, err
		}
		defer rows.Close()
		a := Account{User: p.User, Rooms: []Membership{}}
		for rows.Next() {
			var m Membership
			var id int64
			if err := rows.Scan(&id, &m.Room, &m.Title, &m.Kind, &m.Role); err != nil {
				return Account{}, err
			}
			if !c.Actor.Reaches(id) {
				continue
			}
			a.Rooms = append(a.Rooms, m)
		}
		return a, rows.Err()
	},
}

// SetUserRoleInput is the input of set_user_role.
type SetUserRoleInput struct {
	User string `json:"user" jsonschema:"handle of the person"`
	Role string `json:"role" jsonschema:"admin or member"`
}

// SetUserRole makes a person an administrator of the Cahut, or a member.
var SetUserRole = &operation.Op[SetUserRoleInput, User]{
	Name:        "set_user_role",
	Description: "Make a person an administrator of the Cahut, or a member. The Cahut keeps at least one administrator.",
	Level:       operation.CahutAdmin,
	KnockFirst:  true,
	Run: func(ctx context.Context, c *operation.Call, in SetUserRoleInput) (User, error) {
		c.Log(in.User, in.Role)
		if in.Role != Admin && in.Role != Member {
			return User{}, operation.Invalidf("role %q: use admin or member", in.Role)
		}
		p, err := findUser(ctx, c.Tx, in.User)
		if err != nil {
			return User{}, err
		}
		if p.Role == in.Role {
			return p.User, nil
		}
		if p.Role == Admin {
			if err := keepAnAdmin(ctx, c.Tx, p.ID); err != nil {
				return User{}, err
			}
		}
		if _, err := c.Tx.ExecContext(ctx, "UPDATE users SET role = ? WHERE id = ?", in.Role, p.ID); err != nil {
			return User{}, err
		}
		p.Role = in.Role
		return p.User, nil
	},
}

// keepAnAdmin returns an error wrapping operation.ErrInvalid when the person
// userID is the last administrator of the Cahut who is not disabled.
func keepAnAdmin(ctx context.Context, tx *sql.Tx, userID int64) error {
	var others int
	err := tx.QueryRowContext(ctx, "SELECT count(*) FROM users WHERE role = 'admin' AND disabled_at IS NULL AND id != ?",
		userID).Scan(&others)
	if err != nil {
		return err
	}
	if others == 0 {
		return operation.Invalidf("the Cahut needs an administrator: make another person one first")
	}
	return nil
}

// UserInput names a person, by handle.
type UserInput struct {
	User string `json:"user" jsonschema:"handle of the person"`
}

// DisableUser disables a person: they can no longer sign in, and their AIs
// can no longer act.
var DisableUser = &operation.Op[UserInput, User]{
	Name: "disable_user",
	Description: "Disable a person: their sessions end, and neither they nor their AIs can act until a person enables them again. " +
		"Their Rooms and Windows stay.",
	Level:      operation.CahutAdmin,
	KnockFirst: true,
	Run: func(ctx context.Context, c *operation.Call, in UserInput) (User, error) {
		c.Log(in.User, "")
		p, err := findUser(ctx, c.Tx, in.User)
		if err != nil {
			return User{}, err
		}
		if p.Status == Disabled {
			return User{}, operation.Invalidf("%s is already disabled", p.Handle)
		}
		if p.ID == c.Actor.UserID {
			return User{}, operation.Invalidf("you cannot disable yourself")
		}
		if p.Role == Admin {
			if err := keepAnAdmin(ctx, c.Tx, p.ID); err != nil {
				return User{}, err
			}
		}
		if _, err := c.Tx.ExecContext(ctx, "UPDATE users SET disabled_at = ? WHERE id = ?", store.FormatTime(c.Now), p.ID); err != nil {
			return User{}, err
		}
		if err := endSessions(ctx, c.Tx, p.ID, 0); err != nil {
			return User{}, err
		}
		if _, err := c.Tx.ExecContext(ctx, "DELETE FROM activations WHERE user_id = ?", p.ID); err != nil {
			return User{}, err
		}
		p.Status = Disabled
		return p.User, nil
	},
}

// EnableUser enables a disabled person again.
var EnableUser = &operation.Op[UserInput, User]{
	Name:        "enable_user",
	Description: "Enable a disabled person again: they can sign in, and their AIs can act with their tokens.",
	Level:       operation.CahutAdmin,
	KnockFirst:  true,
	Run: func(ctx context.Context, c *operation.Call, in UserInput) (User, error) {
		c.Log(in.User, "")
		p, err := findUser(ctx, c.Tx, in.User)
		if err != nil {
			return User{}, err
		}
		if p.Status != Disabled {
			return User{}, operation.Invalidf("%s is not disabled", p.Handle)
		}
		if _, err := c.Tx.ExecContext(ctx, "UPDATE users SET disabled_at = NULL WHERE id = ?", p.ID); err != nil {
			return User{}, err
		}
		p.Status = status(sql.NullString{String: p.passwordHash, Valid: p.passwordHash != ""}, sql.NullString{})
		return p.User, nil
	},
}

func parseNullTime(s sql.NullString) (*time.Time, error) {
	if !s.Valid {
		return nil, nil
	}
	t, err := store.ParseTime(s.String)
	if err != nil {
		return nil, err
	}
	return &t, nil
}
