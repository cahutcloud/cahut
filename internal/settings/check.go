package settings

import (
	"context"
	"fmt"

	"cahut.cloud/internal/operation"
)

// The checks below run in the transaction of the call that would add to
// the Cahut, before it writes. A Cahut has a single write connection, so
// two calls at once cannot cross a limit together. A limit lowered below
// what the Cahut holds refuses what would add to it, and deletes nothing.
//
// In an individual Cahut, a refusal for a limit of the whole Cahut gives a
// member no count, which would tell them how much the others hold.

// CheckPerson returns an error wrapping operation.ErrForbidden when one more
// person, an administrator when admin is true, would cross a limit of the
// Cahut: when a person is created, or enabled again.
func CheckPerson(ctx context.Context, c *operation.Call, admin bool) error {
	_, l, err := get(ctx, c.Tx)
	if err != nil || l.People == nil && (l.Admins == nil || !admin) {
		return err
	}
	u, err := Use(ctx, c.Tx)
	if err != nil {
		return err
	}
	if l.People != nil && u.People >= *l.People {
		return operation.Forbiddenf("this Cahut has %d people, its limit, set by the administrators of its server: "+
			"disable a person who no longer needs an account, since disabled people do not count, or %s", u.People, askForMore(c))
	}
	if admin {
		return admins(c, l, u)
	}
	return nil
}

// CheckAdmin returns an error wrapping operation.ErrForbidden when one more
// administrator, among the people who already count, would cross the limit
// of the Cahut: when a person becomes an administrator.
func CheckAdmin(ctx context.Context, c *operation.Call) error {
	_, l, err := get(ctx, c.Tx)
	if err != nil || l.Admins == nil {
		return err
	}
	u, err := Use(ctx, c.Tx)
	if err != nil {
		return err
	}
	return admins(c, l, u)
}

func admins(c *operation.Call, l Limits, u Usage) error {
	if l.Admins != nil && u.Admins >= *l.Admins {
		return operation.Forbiddenf("this Cahut has %d administrators, its limit, set by the administrators of its server: "+
			"make one of them a member first, or %s", u.Admins, askForMore(c))
	}
	return nil
}

// CheckRoom returns an error wrapping operation.ErrForbidden when one more
// team Room, owned by the person owner, would cross a limit of the Cahut.
func CheckRoom(ctx context.Context, c *operation.Call, owner int64) error {
	kind, l, err := get(ctx, c.Tx)
	if err != nil {
		return err
	}
	if l.Rooms != nil {
		u, err := Use(ctx, c.Tx)
		if err != nil {
			return err
		}
		if u.Rooms >= *l.Rooms {
			return operation.Forbiddenf("%s: delete a Room that is no longer used, or %s",
				full(c, kind, "Rooms", fmt.Sprintf("this Cahut has %d team Rooms, its limit", u.Rooms)), askForMore(c))
		}
	}
	if kind == Individual && l.RoomsPerPerson != nil {
		u, err := UseOf(ctx, c.Tx, owner)
		if err != nil {
			return err
		}
		if u.Rooms >= *l.RoomsPerPerson {
			who, err := holder(ctx, c, owner)
			if err != nil {
				return err
			}
			return operation.Forbiddenf("%s %d Rooms beside the personal one, the limit of each person in this Cahut: "+
				"delete a Room that is no longer used, or %s", who.has, u.Rooms, askForMore(c))
		}
	}
	return nil
}

// CheckWindow returns an error wrapping operation.ErrForbidden when one more
// Window out of the trash, in the Room roomID called room, would cross a
// limit of the Cahut: when a Window is published, or restored from the
// trash.
func CheckWindow(ctx context.Context, c *operation.Call, roomID int64, room string) error {
	kind, l, err := get(ctx, c.Tx)
	if err != nil {
		return err
	}
	if l.Windows != nil {
		u, err := Use(ctx, c.Tx)
		if err != nil {
			return err
		}
		if u.Windows >= *l.Windows {
			return operation.Forbiddenf("%s: put a Window that is no longer needed in the trash first, or %s",
				full(c, kind, "Windows", fmt.Sprintf("this Cahut holds %d Windows out of the trash, its limit", u.Windows)), askForMore(c))
		}
	}
	if l.WindowsPerRoom != nil {
		var n int64
		err := c.Tx.QueryRowContext(ctx, "SELECT count(*) FROM windows WHERE room_id = ? AND trashed_at IS NULL", roomID).Scan(&n)
		if err != nil {
			return err
		}
		if n >= *l.WindowsPerRoom {
			return operation.Forbiddenf("the Room %s holds %d Windows out of the trash, the limit of each Room in this Cahut: "+
				"put one of them in the trash first, use another Room, or %s", room, n, askForMore(c))
		}
	}
	if kind == Individual && l.WindowsPerPerson != nil {
		owner, err := ownerOf(ctx, c, roomID)
		if err != nil {
			return err
		}
		u, err := UseOf(ctx, c.Tx, owner)
		if err != nil {
			return err
		}
		if u.Windows >= *l.WindowsPerPerson {
			who, err := holder(ctx, c, owner)
			if err != nil {
				return err
			}
			return operation.Forbiddenf("%s %d Windows out of the trash, the limit of each person in this Cahut: "+
				"put one of them in the trash first, or %s", who.holds, u.Windows, askForMore(c))
		}
	}
	return nil
}

// CheckDisk returns an error wrapping operation.ErrForbidden when adding
// added bytes to a Window of the Room roomID would take the Cahut past a
// limit of its disk. A call that adds nothing, or frees bytes, always
// passes.
func CheckDisk(ctx context.Context, c *operation.Call, roomID, added int64) error {
	if added <= 0 {
		return nil
	}
	kind, l, err := get(ctx, c.Tx)
	if err != nil {
		return err
	}
	if l.Disk != nil {
		used, err := disk(ctx, c.Tx)
		if err != nil {
			return err
		}
		if used+added > *l.Disk {
			return operation.Forbiddenf("%s: purge Windows from the trash, since they still count, or %s",
				full(c, kind, Bytes(added)+" more",
					fmt.Sprintf("this Cahut keeps %s, and %s more would take it past its limit of %s", Bytes(used), Bytes(added), Bytes(*l.Disk))),
				askForMore(c))
		}
	}
	if kind == Individual && l.DiskPerPerson != nil {
		owner, err := ownerOf(ctx, c, roomID)
		if err != nil {
			return err
		}
		u, err := UseOf(ctx, c.Tx, owner)
		if err != nil {
			return err
		}
		if u.Disk+added > *l.DiskPerPerson {
			who, err := holder(ctx, c, owner)
			if err != nil {
				return err
			}
			return operation.Forbiddenf("%s %s, and %s more would take %s past the limit of each person in this Cahut, %s: "+
				"purge Windows from the trash, since they still count, or %s",
				who.keeps, Bytes(u.Disk), Bytes(added), who.them, Bytes(*l.DiskPerPerson), askForMore(c))
		}
	}
	return nil
}

// full says that the Cahut has no room for more of what, with the count
// that says so, except to a member of an individual Cahut, who gets no
// count.
func full(c *operation.Call, kind, what, counted string) string {
	if kind == Individual && !c.Admin {
		return fmt.Sprintf("this Cahut has no room for %s, under the limit that the administrators of its server set", what)
	}
	return counted + ", set by the administrators of its server"
}

// ownerOf returns the owner of the Room roomID, who in an individual Cahut
// is its only member.
func ownerOf(ctx context.Context, c *operation.Call, roomID int64) (int64, error) {
	var id int64
	err := c.Tx.QueryRowContext(ctx, "SELECT user_id FROM room_members WHERE room_id = ? AND role = 'owner' ORDER BY user_id LIMIT 1", roomID).Scan(&id)
	return id, err
}

// person names the person whose limit a refusal speaks of: the caller, or
// another, such as the owner of a Room that an administrator creates.
type person struct{ has, holds, keeps, them string }

func holder(ctx context.Context, c *operation.Call, userID int64) (person, error) {
	if userID == c.Actor.UserID {
		return person{has: "you have", holds: "you hold", keeps: "your Windows keep", them: "them"}, nil
	}
	var handle string
	if err := c.Tx.QueryRowContext(ctx, "SELECT handle FROM users WHERE id = ?", userID).Scan(&handle); err != nil {
		return person{}, err
	}
	return person{has: handle + " has", holds: handle + " holds", keeps: "the Windows of " + handle + " keep", them: "them"}, nil
}

// askForMore says how the caller gets a higher limit: an administrator of
// the server raises it, an administrator of the Cahut asks them, and anyone
// else asks an administrator of the Cahut.
func askForMore(c *operation.Call) string {
	if c.Actor.Kind == operation.Admin {
		return "raise the limit with set_cahut_settings"
	}
	if c.Admin {
		return "ask the administrators of the server for a higher limit"
	}
	return "ask an administrator of the Cahut about a higher limit"
}
