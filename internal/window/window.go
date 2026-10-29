// Package window publishes Windows, the living pages that AIs make, and
// keeps their files on disk. A Window is at <domain>/<room>/<window>/, under
// the domain of its Cahut, and at its permanent address, <domain>/-/w/<id>,
// whose id never changes. It keeps its current version and the one before,
// and can go to the trash and come back. See
// docs/adr/0009-rooms-and-windows-as-paths.md,
// docs/adr/0019-window-versions-access-and-declaration.md and
// docs/adr/0050-window-ids-permanent-addresses-and-names.md.
package window

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"cahut.cloud/internal/account"
	"cahut.cloud/internal/address"
	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/room"
	"cahut.cloud/internal/settings"
	"cahut.cloud/internal/store"
)

// Kinds of Windows.
const (
	Document  = "document"
	Deck      = "deck"
	Dashboard = "dashboard"
	Site      = "site"
)

// Service runs the Window operations and knows where their files live.
type Service struct {
	dir      string
	base     string
	frameKey []byte
}

// NewService returns a service that keeps Window files under dir, gives
// Window addresses under base, the address of the Cahut with its final
// slash, or none when base is empty, and signs the frames of previews with
// frameKey.
func NewService(dir, base string, frameKey []byte) *Service {
	return &Service{dir: dir, base: base, frameKey: frameKey}
}

// PermanentPrefix starts the path of the permanent address of every
// Window, /-/w/<id>.
const PermanentPrefix = "/-/w/"

// url returns the address of the Window called name in the Room room.
func (s *Service) url(room, name string) string {
	if s.base == "" {
		return ""
	}
	return s.base + room + "/" + name + "/"
}

// permanentURL returns the permanent address of the Window with the given
// id.
func (s *Service) permanentURL(id string) string {
	if s.base == "" {
		return ""
	}
	return s.base + strings.TrimPrefix(PermanentPrefix, "/") + id
}

// Register adds the Window operations to r.
func (s *Service) Register(r *operation.Registry) {
	operation.Register(r, &operation.Op[PublishInput, Window]{
		Name:  "publish_window",
		Title: "Publish a Window",
		Description: "Publish a new Window: a document, deck, dashboard or site made of HTML and its files, " +
			"in a Room of the person you act for, by default their personal Room. Returns the address where it opens.",
		Run: s.publish,
	})
	operation.Register(r, &operation.Op[UpdateInput, Window]{
		Name:  "update_window",
		Title: "Update a Window",
		Description: "Publish a new version of a Window: give all its files again, as for publish_window. " +
			"Its address stays. Its current version becomes the previous one, which rollback_window goes back to; " +
			"the version before that is deleted.",
		Destructive: true,
		Run:         s.update,
	})
	operation.Register(r, &operation.Op[RetitleInput, Window]{
		Name:  "retitle_window",
		Title: "Retitle and describe a Window",
		Description: "Change the title of a Window, its description, or both, without publishing a new version. " +
			"Its address stays. Give the old title and description again to undo it.",
		Run: s.retitle,
	})
	operation.Register(r, &operation.Op[RenameInput, Window]{
		Name:  "rename_window",
		Title: "Rename a Window in its address",
		Description: "Change the name of a Window, the last part of its address, <domain>/<room>/<name>/. Its title, its id and its permanent address stay. " +
			"The old address stops working at once, and its name is free in the Room: rename_window back to it undoes the change, while no other Window takes it.",
		Run: s.rename,
	})
	operation.Register(r, &operation.Op[RefInput, Window]{
		Name:  "rollback_window",
		Title: "Roll back a Window",
		Description: "Go back to the previous version of a Window. The version you leave becomes the previous one, " +
			"so calling rollback_window again returns to it.",
		Run: s.rollback,
	})
	operation.Register(r, &operation.Op[RefInput, Window]{
		Name:        "trash_window",
		Title:       "Put a Window in the trash",
		Description: "Put a Window in the trash: it no longer opens and leaves the lists, until restore_window brings it back as it was.",
		Run:         s.trash,
	})
	operation.Register(r, &operation.Op[RefInput, Window]{
		Name:        "restore_window",
		Title:       "Restore a Window",
		Description: "Bring a Window back from the trash, with its versions.",
		Run:         s.restore,
	})
	operation.Register(r, &operation.Op[RefInput, Window]{
		Name:  "purge_window",
		Title: "Delete a Window for good",
		Description: "Delete a Window in the trash for good, with its versions and its files: this cannot be undone. " +
			"Its address becomes free. Put the Window in the trash first, with trash_window.",
		KnockFirst: true,
		Run:        s.purge,
	})
	operation.Register(r, &operation.Op[DeleteRoomInput, DeletedRoom]{
		Name:  "delete_room",
		Title: "Delete a team Room",
		Description: "Delete a team Room for good, with its members and its Windows, those in its trash included, and their files: " +
			"this cannot be undone. Its address becomes free. Its owners and the administrators of the Cahut delete it; " +
			"a personal Room is never deleted.",
		KnockFirst: true,
		Run:        s.deleteRoom,
	})
	operation.Register(r, &operation.Op[account.UserInput, DeletedUser]{
		Name:  "delete_user",
		Title: "Delete a person for good",
		Description: "Delete a person for good: their account, their tokens and the AIs they connected, their personal Room " +
			"and the team Rooms they are the only owner of, with the Windows of these Rooms and their files. This cannot be undone. " +
			"The Windows they published in other Rooms stay, without an author. Their handle becomes free. " +
			"No one deletes themself, nor the last administrator of the Cahut; disable_user stops a person and keeps what is theirs.",
		Level:      operation.CahutAdmin,
		KnockFirst: true,
		Run:        s.deleteUser,
	})
	operation.Register(r, &operation.Op[AccessInput, AccessSet]{
		Name:  "set_window_access",
		Title: "Set who opens a Window",
		Description: "Set who opens a Window: room, the members of its Room (the default); internal, everyone signed in to the Cahut, except in an individual Cahut; " +
			"key, anyone who types its Key; or public, anyone with its address. The access may end at a date, " +
			"after which the Window opens to its Room only. For key, Cahut draws a Key unless you give one, " +
			"and returns it once: give it to the person. Making a Window public waits for the approval of a person.",
		KnocksWhen:  "it makes a Window public",
		Destructive: true,
		OpenWorld:   true,
		Run:         s.setAccess,
	})
	operation.Register(r, &operation.Op[InviteInput, Invited]{
		Name:  "invite_to_window",
		Title: "Invite people to read a Window",
		Description: "Send the address of a Window by e-mail, with a note: to people of the Cahut, named by handle, who sign in to read it, " +
			"or to anyone else, named by address, when the Window opens with its Key or to the public. For a Window with a Key, " +
			"each address gets a personal link that opens it for them alone, until revoke_invitation takes it back; the Key itself never goes. " +
			"20 people a call, and 100 invitations a day. get_cahut says whether the Cahut sends e-mail.",
		OpenWorld: true,
		Run:       s.invite,
	})
	operation.Register(r, &operation.Op[ListInvitationsInput, InvitationList]{
		Name:        "list_invitations",
		Title:       "List the invitations to a Window",
		Description: "List the invitations to a Window, the latest first: who got one, who invited, and whether it still stands.",
		ReadOnly:    true,
		Run:         s.listInvitations,
	})
	operation.Register(r, &operation.Op[RevokeInvitationInput, Invitation]{
		Name:  "revoke_invitation",
		Title: "Take back an invitation",
		Description: "Take back an invitation to a Window: its personal link, and the passes it gave, no longer open the Window. " +
			"The e-mail stays sent.",
		Destructive: true,
		Run:         s.revokeInvitation,
	})
	operation.Register(r, &operation.Op[ShareLinkInput, ShareLinkMade]{
		Name:  "create_share_link",
		Title: "Share a Window with a link",
		Description: "Make a share link: whoever holds it opens the Window, without an account or a Key, until it ends, besides its access. " +
			"It ends in 7 days unless you give an end, 30 days at most, and may open to a number of readers only. " +
			"It is returned this once: give it to the person. revoke_share_link takes it back. Not for a public Window.",
		OpenWorld: true,
		Run:       s.createLink,
	})
	operation.Register(r, &operation.Op[RevokeShareLinkInput, ShareLink]{
		Name:        "revoke_share_link",
		Title:       "Take back a share link",
		Description: "Take back a share link: it no longer opens its Window, to those who opened it too.",
		Destructive: true,
		Run:         s.revokeLink,
	})
	operation.Register(r, &operation.Op[OpenShareLinkInput, ShareLinkOpening]{
		Name:        "open_share_link",
		Title:       "Open a share link",
		Description: "Open the Window of a share link, from the page of the link: spend one of its openings, and give a pass.",
		PeopleOnly:  true,
		Visitors:    true,
		Run:         openLink,
	})
	operation.Register(r, &operation.Op[InvitationCodeInput, InvitationPage]{
		Name:        "get_invitation",
		Title:       "Describe an invitation",
		Description: "Describe the invitation of a personal link: the Window, who invites, and their note.",
		PeopleOnly:  true,
		Visitors:    true,
		ReadOnly:    true,
		Run:         getInvitation,
	})
	operation.Register(r, &operation.Op[InvitationCodeInput, InvitationOpening]{
		Name:        "open_invitation",
		Title:       "Open an invitation",
		Description: "Open the Window of the invitation of a personal link: a code, valid a minute, for /-/invitation, which gives a pass.",
		PeopleOnly:  true,
		Visitors:    true,
		Run:         openInvitation,
	})
	operation.Register(r, &operation.Op[GetInput, Detail]{
		Name:        "get_window",
		Title:       "Describe a Window",
		Description: "Describe a Window and list its files, optionally with their contents.",
		ReadOnly:    true,
		Run:         s.get,
	})
	operation.Register(r, &operation.Op[ListInput, List]{
		Name:        "list_windows",
		Title:       "List Windows",
		Description: "List the Windows in the Rooms of the person you act for, or in one of them, most recently updated first.",
		ReadOnly:    true,
		Run:         s.list,
	})
	operation.Register(r, &operation.Op[PreviewInput, Previews]{
		Name:  "preview_windows",
		Title: "Preview Windows",
		Description: "Get the frames that show Windows of your Rooms in the console, as previews: " +
			"each opens one Window, while your session lasts and four hours at most.",
		ReadOnly:   true,
		PeopleOnly: true,
		Run:        s.preview,
	})
}

// Window describes a published Window.
type Window struct {
	ID           string    `json:"id" jsonschema:"the Window's id, which never changes: name the Window by it to find it whatever its name"`
	Name         string    `json:"name" jsonschema:"the Window's name, last part of its address"`
	Title        string    `json:"title"`
	Description  string    `json:"description,omitempty" jsonschema:"what the Window is, in a sentence or two; absent when it has none"`
	Kind         string    `json:"kind"`
	URL          string    `json:"url" jsonschema:"where the Window opens"`
	PermanentURL string    `json:"permanent_url" jsonschema:"the Window's permanent address, which sends whoever may open the Window to its address, whatever its name: the link that lasts"`
	Room         string    `json:"room" jsonschema:"the name of its Room, the part of its address before its name"`
	Version      int       `json:"version" jsonschema:"the number of the version that readers see"`
	Previous     int       `json:"previous_version,omitempty" jsonschema:"the number of the version that rollback_window goes back to; absent when there is none"`
	Size         int64     `json:"size" jsonschema:"total size of the files of the current version, in bytes"`
	FileCount    int       `json:"file_count"`
	CreatedBy    string    `json:"created_by" jsonschema:"handle of the person it was first published for, empty once that person is deleted"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Access       string    `json:"access" jsonschema:"who opens the Window: room (the members of its Room), internal (everyone signed in to the Cahut), key (anyone with its Key) or public"`
	// The times below are pointers, so that a Window without them omits
	// them.
	AccessEndsAt *time.Time   `json:"access_ends_at,omitempty" jsonschema:"when the access ends, after which the Window opens to its Room only; absent when it does not end"`
	TrashedAt    *time.Time   `json:"trashed_at,omitempty" jsonschema:"when the Window went to the trash; absent when it is not in it"`
	Declaration  *Declaration `json:"declaration,omitempty" jsonschema:"what the pages of the current version do, as declared when it was published; absent for a Window published before declarations"`
}

// PublishInput is the input of publish_window.
type PublishInput struct {
	Title       string       `json:"title" jsonschema:"title of the Window, also used in its address"`
	Description string       `json:"description,omitempty" jsonschema:"what the Window is, in a sentence or two of plain text, 300 characters at most: readers see it on the page of its Room, and AIs read it to find the Window"`
	Kind        string       `json:"kind" jsonschema:"one of document, deck, dashboard or site"`
	Room        string       `json:"room,omitempty" jsonschema:"name of the Room to publish in; the personal Room of the person you act for when empty"`
	Name        string       `json:"name,omitempty" jsonschema:"its name in its address, 1 to 63 lowercase letters, digits and hyphens; from the title when empty"`
	Files       []File       `json:"files" jsonschema:"the files of the Window; index.html at the root is the page that opens. Link them with relative paths, such as css/deck.css"`
	Declaration *Declaration `json:"declaration" jsonschema:"what the pages do: Cahut lets them load and call the origins listed here, and nothing else"`
}

// File is a file of a publication.
type File struct {
	Path     string `json:"path" jsonschema:"relative path with / as separator, such as index.html or img/logo.svg"`
	Content  string `json:"content"`
	Encoding string `json:"encoding,omitempty" jsonschema:"utf-8 (the default) or base64, for binary files"`
}

func (s *Service) publish(ctx context.Context, c *operation.Call, in PublishInput) (Window, error) {
	c.Log(strings.TrimSpace(in.Title), "")
	title, err := room.CheckTitle(in.Title)
	if err != nil {
		return Window{}, err
	}
	description, err := room.CheckDescription(in.Description)
	if err != nil {
		return Window{}, err
	}
	switch in.Kind {
	case Document, Deck, Dashboard, Site:
	default:
		return Window{}, operation.Invalidf("kind %q: use document, deck, dashboard or site", in.Kind)
	}
	files, size, err := decode(in.Files)
	if err != nil {
		return Window{}, err
	}
	if err := s.checkSpace(size); err != nil {
		return Window{}, err
	}
	declared, err := in.Declaration.check()
	if err != nil {
		return Window{}, err
	}
	if c.Actor.UserID == 0 {
		return Window{}, operation.Invalidf("publish_window acts for a person, and the administrator is not one")
	}

	// A person publishes in their personal Room, named after their handle,
	// or in a Room they are in.
	name := in.Room
	if name == "" {
		name = c.Actor.Handle
	}
	r, err := room.Require(ctx, c, name, room.Member)
	if err != nil {
		return Window{}, err
	}
	c.InRoom(r.ID)
	roomID := r.ID
	if err := settings.CheckWindow(ctx, c, roomID, r.Name); err != nil {
		return Window{}, err
	}
	if err := settings.CheckDisk(ctx, c, roomID, size); err != nil {
		return Window{}, err
	}
	w := Window{ID: address.NewID(), Title: title, Description: description, Kind: in.Kind, Room: r.Name, Version: 1, Size: size, FileCount: len(files), Access: AccessRoom,
		CreatedBy: c.Actor.Handle, CreatedAt: c.Now, UpdatedAt: c.Now, Declaration: declared}
	if in.Name != "" {
		w.Name = strings.TrimSpace(in.Name)
		if err := checkName(ctx, c.Tx, roomID, r.Name, w.Name); err != nil {
			return Window{}, err
		}
	} else if w.Name, err = freeName(ctx, c.Tx, roomID, title); err != nil {
		return Window{}, err
	}
	w.URL, w.PermanentURL = s.url(w.Room, w.Name), s.permanentURL(w.ID)
	now := store.FormatTime(c.Now)
	var id int64
	err = c.Tx.QueryRowContext(ctx, `
		INSERT INTO windows (uid, room_id, name, title, description, kind, version, size, file_count, created_by, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) RETURNING id`,
		w.ID, roomID, w.Name, w.Title, w.Description, w.Kind, w.Version, w.Size, w.FileCount, c.Actor.UserID, now, now).Scan(&id)
	if err != nil {
		return Window{}, err
	}
	if err := insertVersion(ctx, c, id, w); err != nil {
		return Window{}, err
	}
	c.Log(w.Room+"/"+w.Name, fmt.Sprintf("%s %q, id %s, %d files, %d bytes; %s", w.Kind, w.Title, w.ID, w.FileCount, w.Size, w.Declaration))

	c.OnFailure(func() { os.RemoveAll(filepath.Join(s.dir, strconv.FormatInt(id, 10))) })
	if err := writeFiles(s.Dir(id, w.Version), files); err != nil {
		return Window{}, err
	}
	return w, nil
}

// UpdateInput is the input of update_window.
type UpdateInput struct {
	Window string `json:"window" jsonschema:"the Window's id or address, or its Room and its name as room/name"`
	Title  string `json:"title,omitempty" jsonschema:"a new title; the title stays when empty. The address does not change"`
	// A pointer, so that an empty description removes it and an absent one
	// keeps it.
	Description *string      `json:"description,omitempty" jsonschema:"a new description, a sentence or two of plain text; empty to remove it, and it stays when absent"`
	Files       []File       `json:"files" jsonschema:"every file of the new version, as for publish_window; files left out are not in it"`
	Declaration *Declaration `json:"declaration" jsonschema:"what the pages do: Cahut lets them load and call the origins listed here, and nothing else"`
}

func (s *Service) update(ctx context.Context, c *operation.Call, in UpdateInput) (Window, error) {
	w, err := s.target(ctx, c, in.Window)
	if err != nil {
		return Window{}, err
	}
	if err := w.outOfTrash(); err != nil {
		return Window{}, err
	}
	if strings.TrimSpace(in.Title) != "" {
		if w.Title, err = room.CheckTitle(in.Title); err != nil {
			return Window{}, err
		}
	}
	if in.Description != nil {
		if w.Description, err = room.CheckDescription(*in.Description); err != nil {
			return Window{}, err
		}
	}
	files, size, err := decode(in.Files)
	if err != nil {
		return Window{}, err
	}
	if err := s.checkSpace(size); err != nil {
		return Window{}, err
	}
	if w.Declaration, err = in.Declaration.check(); err != nil {
		return Window{}, err
	}

	// The current version becomes the previous one, and the previous one
	// goes, files included, once the call has succeeded.
	dropped := w.Previous
	var droppedSize int64
	if dropped != 0 {
		err := c.Tx.QueryRowContext(ctx, "SELECT size FROM window_versions WHERE window_id = ? AND number = ?", w.id, dropped).Scan(&droppedSize)
		if err != nil {
			return Window{}, err
		}
	}
	if err := settings.CheckDisk(ctx, c, w.roomID, size-droppedSize); err != nil {
		return Window{}, err
	}
	w.Previous, w.Version = w.Version, max(w.Version, w.Previous)+1
	w.Size, w.FileCount, w.UpdatedAt = size, len(files), c.Now
	now := store.FormatTime(c.Now)
	if err := insertVersion(ctx, c, w.id, w.Window); err != nil {
		return Window{}, err
	}
	if _, err := c.Tx.ExecContext(ctx, "DELETE FROM window_versions WHERE window_id = ? AND number = ?", w.id, dropped); err != nil {
		return Window{}, err
	}
	_, err = c.Tx.ExecContext(ctx, `
		UPDATE windows SET title = ?, description = ?, version = ?, previous = ?, size = ?, file_count = ?, updated_at = ? WHERE id = ?`,
		w.Title, w.Description, w.Version, w.Previous, w.Size, w.FileCount, now, w.id)
	if err != nil {
		return Window{}, err
	}
	c.Log(w.Room+"/"+w.Name, fmt.Sprintf("version %d, %q, %d files, %d bytes; %s", w.Version, w.Title, w.FileCount, w.Size, w.Declaration))

	// No version ever had this number, but a crash may have left its files.
	dir := s.Dir(w.id, w.Version)
	if err := os.RemoveAll(dir); err != nil {
		return Window{}, err
	}
	c.OnFailure(func() { os.RemoveAll(dir) })
	if err := writeFiles(dir, files); err != nil {
		return Window{}, err
	}
	if dropped != 0 {
		c.AfterCommit(func() { os.RemoveAll(s.Dir(w.id, dropped)) })
	}
	return w.Window, nil
}

// insertVersion records the current version of the Window w, whose id is
// id, as the call publishes it.
func insertVersion(ctx context.Context, c *operation.Call, id int64, w Window) error {
	declared, err := w.Declaration.encode()
	if err != nil {
		return err
	}
	_, err = c.Tx.ExecContext(ctx, `
		INSERT INTO window_versions (window_id, number, size, file_count, published_by, published_at, declaration)
		VALUES (?, ?, ?, ?, ?, ?, ?)`, id, w.Version, w.Size, w.FileCount, c.Actor.UserID, store.FormatTime(c.Now), declared)
	return err
}

// RetitleInput is the input of retitle_window.
type RetitleInput struct {
	Window string `json:"window" jsonschema:"the Window's id or address, or its Room and its name as room/name"`
	Title  string `json:"title,omitempty" jsonschema:"the new title; the title stays when empty. The address does not change"`
	// A pointer, so that an empty description removes it and an absent one
	// keeps it.
	Description *string `json:"description,omitempty" jsonschema:"the new description, a sentence or two of plain text, 300 characters at most; empty to remove it, and it stays when absent"`
}

// retitle changes the title of a Window, its description, or both, in
// place: no new version, and the time of its last version stays.
func (s *Service) retitle(ctx context.Context, c *operation.Call, in RetitleInput) (Window, error) {
	if strings.TrimSpace(in.Title) == "" && in.Description == nil {
		c.Log(in.Window, "")
		return Window{}, operation.Invalidf("give a new title, a new description, or both")
	}
	w, err := s.target(ctx, c, in.Window)
	if err != nil {
		return Window{}, err
	}
	if err := w.outOfTrash(); err != nil {
		return Window{}, err
	}
	var changes []string
	if strings.TrimSpace(in.Title) != "" {
		if w.Title, err = room.CheckTitle(in.Title); err != nil {
			return Window{}, err
		}
		changes = append(changes, fmt.Sprintf("title %q", w.Title))
	}
	if in.Description != nil {
		if w.Description, err = room.CheckDescription(*in.Description); err != nil {
			return Window{}, err
		}
		changes = append(changes, fmt.Sprintf("description %q", w.Description))
	}
	if _, err := c.Tx.ExecContext(ctx, "UPDATE windows SET title = ?, description = ? WHERE id = ?", w.Title, w.Description, w.id); err != nil {
		return Window{}, err
	}
	c.Log(w.Room+"/"+w.Name, strings.Join(changes, "; "))
	return w.Window, nil
}

// RenameInput is the input of rename_window.
type RenameInput struct {
	Window string `json:"window" jsonschema:"the Window's id or address, or its Room and its name as room/name"`
	Name   string `json:"name" jsonschema:"the new name, 1 to 63 lowercase letters, digits and hyphens"`
}

// rename changes the name of a Window in its address, and nothing else:
// the files, the passes, the frames, the invitations and the home of its
// Room hold it by its id. The old address stops working, without a
// redirect, and its name is free. See
// docs/adr/0050-window-ids-permanent-addresses-and-names.md.
func (s *Service) rename(ctx context.Context, c *operation.Call, in RenameInput) (Window, error) {
	w, err := s.target(ctx, c, in.Window)
	if err != nil {
		return Window{}, err
	}
	if err := w.outOfTrash(); err != nil {
		return Window{}, err
	}
	name := strings.TrimSpace(in.Name)
	if name == w.Name {
		return Window{}, operation.Invalidf("%s/%s already has this name", w.Room, w.Name)
	}
	if err := checkName(ctx, c.Tx, w.roomID, w.Room, name); err != nil {
		return Window{}, err
	}
	if _, err := c.Tx.ExecContext(ctx, "UPDATE windows SET name = ? WHERE id = ?", name, w.id); err != nil {
		return Window{}, err
	}
	c.Log(w.Room+"/"+w.Name, fmt.Sprintf("to %s/%s, id %s", w.Room, name, w.ID))
	w.Name, w.URL = name, s.url(w.Room, name)
	return w.Window, nil
}

// RefInput names a Window, for the operations that act on one.
type RefInput struct {
	Window string `json:"window" jsonschema:"the Window's id or address, or its Room and its name as room/name"`
}

func (s *Service) rollback(ctx context.Context, c *operation.Call, in RefInput) (Window, error) {
	w, err := s.target(ctx, c, in.Window)
	if err != nil {
		return Window{}, err
	}
	if err := w.outOfTrash(); err != nil {
		return Window{}, err
	}
	if w.Previous == 0 {
		return Window{}, operation.Invalidf("%s/%s has a single version: there is none to go back to", w.Room, w.Name)
	}
	w.Version, w.Previous = w.Previous, w.Version
	w.UpdatedAt = c.Now
	var declared *string
	err = c.Tx.QueryRowContext(ctx, "SELECT size, file_count, declaration FROM window_versions WHERE window_id = ? AND number = ?",
		w.id, w.Version).Scan(&w.Size, &w.FileCount, &declared)
	if err != nil {
		return Window{}, err
	}
	if w.Declaration, err = decodeDeclaration(declared); err != nil {
		return Window{}, err
	}
	_, err = c.Tx.ExecContext(ctx, "UPDATE windows SET version = ?, previous = ?, size = ?, file_count = ?, updated_at = ? WHERE id = ?",
		w.Version, w.Previous, w.Size, w.FileCount, store.FormatTime(c.Now), w.id)
	if err != nil {
		return Window{}, err
	}
	c.Log(w.Room+"/"+w.Name, fmt.Sprintf("version %d, from version %d", w.Version, w.Previous))
	return w.Window, nil
}

func (s *Service) trash(ctx context.Context, c *operation.Call, in RefInput) (Window, error) {
	w, err := s.target(ctx, c, in.Window)
	if err != nil {
		return Window{}, err
	}
	if w.TrashedAt != nil {
		return Window{}, operation.Invalidf("%s/%s is already in the trash", w.Room, w.Name)
	}
	if _, err := c.Tx.ExecContext(ctx, "UPDATE windows SET trashed_at = ? WHERE id = ?", store.FormatTime(c.Now), w.id); err != nil {
		return Window{}, err
	}
	w.TrashedAt = &c.Now
	return w.Window, nil
}

func (s *Service) restore(ctx context.Context, c *operation.Call, in RefInput) (Window, error) {
	w, err := s.target(ctx, c, in.Window)
	if err != nil {
		return Window{}, err
	}
	if w.TrashedAt == nil {
		return Window{}, operation.Invalidf("%s/%s is not in the trash", w.Room, w.Name)
	}
	if err := settings.CheckWindow(ctx, c, w.roomID, w.Room); err != nil {
		return Window{}, err
	}
	if _, err := c.Tx.ExecContext(ctx, "UPDATE windows SET trashed_at = NULL WHERE id = ?", w.id); err != nil {
		return Window{}, err
	}
	w.TrashedAt = nil
	return w.Window, nil
}

// target returns the Window that ref names, for a call that changes it: the
// caller must publish in its Room.
func (s *Service) target(ctx context.Context, c *operation.Call, ref string) (found, error) {
	roomName, name, err := s.locate(ctx, c, ref)
	if err != nil {
		return found{}, err
	}
	r, err := room.Require(ctx, c, roomName, room.Member)
	if err != nil {
		return found{}, err
	}
	c.InRoom(r.ID)
	w, err := s.query(ctx, c, "w.room_id = ? AND w.name = ?", r.ID, name)
	if err != nil {
		return found{}, err
	}
	if len(w) == 0 {
		return found{}, operation.NotFoundf("no Window %s/%s", roomName, name)
	}
	return w[0], nil
}

// outOfTrash returns an error wrapping operation.ErrInvalid when the Window
// is in the trash.
func (w found) outOfTrash() error {
	if w.TrashedAt != nil {
		return operation.Invalidf("%s/%s is in the trash: bring it back with restore_window first", w.Room, w.Name)
	}
	return nil
}

// Dir returns the directory that holds the files of a version of the Window
// with the given id.
func (s *Service) Dir(id int64, version int) string {
	return filepath.Join(s.dir, strconv.FormatInt(id, 10), strconv.Itoa(version))
}

// GetInput is the input of get_window.
type GetInput struct {
	Window         string `json:"window" jsonschema:"the Window's id or address, or its Room and its name as room/name"`
	IncludeContent bool   `json:"include_content,omitempty" jsonschema:"also return the content of each file"`
}

// Detail is the output of get_window.
type Detail struct {
	Window
	Versions []Version  `json:"versions" jsonschema:"the current version, then the previous one if there is one"`
	Files    []FileInfo `json:"files" jsonschema:"the files of the current version"`
	// Share links open the Window besides its access. See
	// docs/adr/0052-share-links.md.
	ShareLinks int         `json:"share_links" jsonschema:"how many share links open the Window to someone new, besides its access"`
	Links      []ShareLink `json:"links" jsonschema:"its share links, the latest first, without the links themselves"`
}

// Version describes a version of a Window.
type Version struct {
	Number      int       `json:"number"`
	Current     bool      `json:"current" jsonschema:"whether readers see this version"`
	Size        int64     `json:"size" jsonschema:"total size of its files, in bytes"`
	FileCount   int       `json:"file_count"`
	PublishedBy string    `json:"published_by" jsonschema:"handle of the person it was published for, empty once that person is deleted"`
	PublishedAt time.Time `json:"published_at"`
}

// FileInfo describes a file of a Window.
type FileInfo struct {
	Path     string `json:"path"`
	Size     int64  `json:"size"`
	Type     string `json:"type" jsonschema:"media type"`
	Content  string `json:"content,omitempty"`
	Encoding string `json:"encoding,omitempty" jsonschema:"utf-8 or base64"`
}

func (s *Service) get(ctx context.Context, c *operation.Call, in GetInput) (Detail, error) {
	room, name, err := s.locate(ctx, c, in.Window)
	if err != nil {
		return Detail{}, err
	}
	w, err := s.query(ctx, c, "r.name = ? AND w.name = ?", room, name)
	if err != nil {
		return Detail{}, err
	}
	if len(w) == 0 {
		return Detail{}, operation.NotFoundf("no Window %s/%s in your Rooms", room, name)
	}
	c.InRoom(w[0].roomID)
	versions, err := versionsOf(ctx, c.Tx, w[0])
	if err != nil {
		return Detail{}, err
	}
	files, err := listFiles(s.Dir(w[0].id, w[0].Version), in.IncludeContent)
	if err != nil {
		return Detail{}, err
	}
	links, standing, err := linksOf(ctx, c.Tx, w[0].id, c.Now)
	if err != nil {
		return Detail{}, err
	}
	return Detail{Window: w[0].Window, Versions: versions, Files: files, ShareLinks: standing, Links: links}, nil
}

// versionsOf returns the versions of the Window w, the current one first.
func versionsOf(ctx context.Context, tx *sql.Tx, w found) ([]Version, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT v.number, v.size, v.file_count, coalesce(u.handle, ''), v.published_at
		FROM window_versions v LEFT JOIN users u ON u.id = v.published_by
		WHERE v.window_id = ? ORDER BY v.number = ? DESC`, w.id, w.Version)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	versions := []Version{}
	for rows.Next() {
		var v Version
		var at string
		if err := rows.Scan(&v.Number, &v.Size, &v.FileCount, &v.PublishedBy, &at); err != nil {
			return nil, err
		}
		if v.PublishedAt, err = store.ParseTime(at); err != nil {
			return nil, err
		}
		v.Current = v.Number == w.Version
		versions = append(versions, v)
	}
	return versions, rows.Err()
}

// ListInput is the input of list_windows.
type ListInput struct {
	Room    string `json:"room,omitempty" jsonschema:"only the Windows of this Room, by its name"`
	Trashed bool   `json:"trashed,omitempty" jsonschema:"list the Windows in the trash instead"`
}

// List is the output of list_windows.
type List struct {
	Windows []Window `json:"windows"`
}

func (s *Service) list(ctx context.Context, c *operation.Call, in ListInput) (List, error) {
	where, args := "w.trashed_at IS NULL", []any{}
	if in.Trashed {
		where = "w.trashed_at IS NOT NULL"
	}
	if in.Room != "" {
		r, err := room.Require(ctx, c, in.Room, room.Member)
		if err != nil {
			return List{}, err
		}
		c.InRoom(r.ID)
		where, args = where+" AND w.room_id = ?", []any{r.ID}
	}
	found, err := s.query(ctx, c, where, args...)
	if err != nil {
		return List{}, err
	}
	c.Log(in.Room, fmt.Sprintf("%d Windows", len(found)))
	list := List{Windows: make([]Window, len(found))}
	for i, w := range found {
		list.Windows[i] = w.Window
	}
	return list, nil
}

// found is a Window that query found, with its id and that of its Room,
// and whether it has a Key.
type found struct {
	Window
	id, roomID int64
	hasKey     bool
}

// query returns the Windows matching where that the actor may see: all of
// them for the command line, those of their Rooms for anyone else, and of
// the Rooms it was granted for an AI.
// Administrators manage people and Rooms, not content: an administrator of
// the server, who belongs to no Room, sees none.
func (s *Service) query(ctx context.Context, c *operation.Call, where string, args ...any) ([]found, error) {
	q := `
		SELECT w.id, w.uid, w.room_id, w.name, w.title, w.description, w.kind, r.name, w.version, w.previous, w.size, w.file_count,
			coalesce(u.handle, ''), w.created_at, w.updated_at, w.trashed_at, w.access, w.access_ends_at, w.key_hash IS NOT NULL, v.declaration
		FROM windows w JOIN rooms r ON r.id = w.room_id LEFT JOIN users u ON u.id = w.created_by
		JOIN window_versions v ON v.window_id = w.id AND v.number = w.version
		WHERE ` + where
	if !c.Actor.CommandLine() {
		q += " AND w.room_id IN (SELECT room_id FROM room_members WHERE user_id = ?)"
		args = append(args, c.Actor.UserID)
	}
	rows, err := c.Tx.QueryContext(ctx, q+" ORDER BY w.updated_at DESC, w.id DESC", args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	windows := []found{}
	for rows.Next() {
		var w found
		var previous sql.NullInt64
		var created, updated string
		var trashed, ends sql.NullString
		var declared *string
		if err := rows.Scan(&w.id, &w.ID, &w.roomID, &w.Name, &w.Title, &w.Description, &w.Kind, &w.Room, &w.Version, &previous, &w.Size, &w.FileCount,
			&w.CreatedBy, &created, &updated, &trashed, &w.Access, &ends, &w.hasKey, &declared); err != nil {
			return nil, err
		}
		if w.Declaration, err = decodeDeclaration(declared); err != nil {
			return nil, err
		}
		if w.Access, w.AccessEndsAt, err = effective(w.Access, ends, c.Now); err != nil {
			return nil, err
		}
		w.Previous = int(previous.Int64)
		if trashed.Valid {
			at, err := store.ParseTime(trashed.String)
			if err != nil {
				return nil, err
			}
			w.TrashedAt = &at
		}
		if w.CreatedAt, err = store.ParseTime(created); err != nil {
			return nil, err
		}
		if w.UpdatedAt, err = store.ParseTime(updated); err != nil {
			return nil, err
		}
		if !c.Actor.Reaches(w.roomID) {
			continue
		}
		w.URL, w.PermanentURL = s.url(w.Room, w.Name), s.permanentURL(w.ID)
		windows = append(windows, w)
	}
	return windows, rows.Err()
}

// The refusals of a name that a person chooses for a Window.
var (
	badName = operation.NewRefusalOf[GivenName]("window_name_invalid", operation.ErrInvalid,
		"name {name}: use 1 to 63 lowercase letters, digits and hyphens, starting and ending with a letter or a digit")
	nameTaken = operation.NewRefusalOf[NameInRoom]("window_name_taken", operation.ErrInvalid,
		"name {name} is taken in the Room {room}, by a Window or one in its trash: choose another")
)

// GivenName fills the refusal of a name that has not the form of one.
type GivenName struct {
	Name operation.Quoted `json:"name" jsonschema:"the name"`
}

// NameInRoom fills the refusal of a name that a Window of the Room has.
type NameInRoom struct {
	Name operation.Quoted `json:"name" jsonschema:"the name"`
	Room string           `json:"room" jsonschema:"the name of the Room"`
}

// checkName returns an error wrapping operation.ErrInvalid when name cannot
// be the name of a Window in the Room roomID, called room: when it has not
// the form of a name, or when a Window of that Room, in the trash or not,
// has it.
func checkName(ctx context.Context, tx *sql.Tx, roomID int64, room, name string) error {
	if !address.ValidName(name) {
		return badName(GivenName{Name: operation.Quoted(name)})
	}
	var taken bool
	if err := tx.QueryRowContext(ctx, "SELECT EXISTS (SELECT 1 FROM windows WHERE room_id = ? AND name = ?)", roomID, name).Scan(&taken); err != nil {
		return err
	}
	if taken {
		return nameTaken(NameInRoom{Name: operation.Quoted(name), Room: room})
	}
	return nil
}

// freeName returns a name for a new Window in the Room roomID: its title as
// a name, then -2, -3 and so on when the Room already has a Window of that
// name.
func freeName(ctx context.Context, tx *sql.Tx, roomID int64, title string) (string, error) {
	base := address.Slug(title, "window")
	for n := 1; n <= 1000; n++ {
		name := base
		if n > 1 {
			name += "-" + strconv.Itoa(n)
		}
		var taken bool
		err := tx.QueryRowContext(ctx, "SELECT EXISTS (SELECT 1 FROM windows WHERE room_id = ? AND name = ?)", roomID, name).Scan(&taken)
		if err != nil {
			return "", err
		}
		if !taken {
			return name, nil
		}
	}
	return "", operation.Invalidf("the Room has too many Windows called %q: choose another title", title)
}

// reference is a Window as a call names it: by its id, or by its Room and
// its name.
type reference struct {
	id, room, name string
}

// parseRef reads a Window from its id; from its permanent address, such as
// https://cahut.example.com/-/w/<id>; from its address, such as
// https://cahut.example.com/marketing/q3-review/index.html; or from
// room/name. It reports false when s is none of them.
func parseRef(s string) (reference, bool) {
	p := strings.TrimSpace(s)
	if id := strings.ToLower(p); address.ValidID(id) {
		return reference{id: id}, true
	}
	if strings.Contains(p, "://") {
		u, err := url.Parse(p)
		if err != nil {
			return reference{}, false
		}
		p = u.Path
	}
	parts := strings.Split(strings.Trim(p, "/"), "/")
	if len(parts) >= 3 && "/"+parts[0]+"/"+parts[1]+"/" == PermanentPrefix {
		id := strings.ToLower(parts[2])
		return reference{id: id}, address.ValidID(id)
	}
	if len(parts) < 2 || !address.ValidName(parts[0]) || !address.ValidName(parts[1]) {
		return reference{}, false
	}
	return reference{room: parts[0], name: parts[1]}, true
}

// locate returns the Room and the name of the Window that ref names, and
// logs it as room/name. A Window named by its id is found in the Rooms
// that the actor reaches only, as get_window finds it.
func (s *Service) locate(ctx context.Context, c *operation.Call, ref string) (room, name string, err error) {
	r, ok := parseRef(ref)
	if !ok {
		c.Log(ref, "")
		return "", "", operation.Invalidf("window %q: give the Window's id or address, or its Room and its name as room/name", ref)
	}
	if r.id == "" {
		c.Log(r.room+"/"+r.name, "")
		return r.room, r.name, nil
	}
	w, err := s.query(ctx, c, "w.uid = ?", r.id)
	if err != nil {
		return "", "", err
	}
	if len(w) == 0 {
		c.Log(r.id, "")
		return "", "", operation.NotFoundf("no Window %s in your Rooms", r.id)
	}
	c.Log(w[0].Room+"/"+w[0].Name, "")
	return w[0].Room, w[0].Name, nil
}
