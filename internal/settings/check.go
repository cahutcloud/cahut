package settings

import (
	"context"

	"cahut.cloud/internal/operation"
)

// The checks below run in the transaction of the call that would add to
// the Cahut, before it writes. A Cahut has a single write connection, so
// two calls at once cannot cross a limit together. A limit lowered below
// what the Cahut holds refuses what would add to it, and deletes nothing.

// CheckPerson returns an error wrapping operation.ErrForbidden when one more
// person, an administrator when admin is true, would cross a limit of the
// Cahut: when a person is created, or enabled again.
func CheckPerson(ctx context.Context, c *operation.Call, admin bool) error {
	l, err := Load(ctx, c.Tx)
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
	l, err := Load(ctx, c.Tx)
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
// team Room would cross the limit of the Cahut.
func CheckRoom(ctx context.Context, c *operation.Call) error {
	l, err := Load(ctx, c.Tx)
	if err != nil || l.Rooms == nil {
		return err
	}
	u, err := Use(ctx, c.Tx)
	if err != nil {
		return err
	}
	if u.Rooms >= *l.Rooms {
		return operation.Forbiddenf("this Cahut has %d team Rooms, its limit, set by the administrators of its server: "+
			"delete a Room that is no longer used, or %s", u.Rooms, askForMore(c))
	}
	return nil
}

// CheckWindow returns an error wrapping operation.ErrForbidden when one more
// Window out of the trash, in the Room roomID called room, would cross a
// limit of the Cahut: when a Window is published, or restored from the
// trash.
func CheckWindow(ctx context.Context, c *operation.Call, roomID int64, room string) error {
	l, err := Load(ctx, c.Tx)
	if err != nil {
		return err
	}
	if l.Windows != nil {
		u, err := Use(ctx, c.Tx)
		if err != nil {
			return err
		}
		if u.Windows >= *l.Windows {
			return operation.Forbiddenf("this Cahut holds %d Windows out of the trash, its limit, set by the administrators of its server: "+
				"put a Window that is no longer needed in the trash first, or %s", u.Windows, askForMore(c))
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
	return nil
}

// CheckDisk returns an error wrapping operation.ErrForbidden when adding
// added bytes would take the Cahut past the limit of its disk. A call that
// adds nothing, or frees bytes, always passes.
func CheckDisk(ctx context.Context, c *operation.Call, added int64) error {
	if added <= 0 {
		return nil
	}
	l, err := Load(ctx, c.Tx)
	if err != nil || l.Disk == nil {
		return err
	}
	used, err := disk(ctx, c.Tx)
	if err != nil {
		return err
	}
	if used+added > *l.Disk {
		return operation.Forbiddenf("this Cahut keeps %s, and %s more would take it past its limit of %s, set by the administrators of its server: "+
			"purge Windows from the trash, since they still count, or %s", Bytes(used), Bytes(added), Bytes(*l.Disk), askForMore(c))
	}
	return nil
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
