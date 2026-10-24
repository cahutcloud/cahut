// Package notice sends the notices that time brings, rather than a call:
// the access of a Window that ends within three days, a limit of a Cahut at
// 90% or reached, the disk of the server almost full. The server runs it
// every hour. See docs/adr/0031-e-mail-through-a-relay.md.
package notice

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"cahut.cloud/internal/cahut"
	"cahut.cloud/internal/mail"
	"cahut.cloud/internal/server"
	"cahut.cloud/internal/settings"
	"cahut.cloud/internal/store"
)

// Every is how often the server looks for notices to send.
const Every = time.Hour

// Lead is how long before the end of an access its person hears of it.
const Lead = 3 * 24 * time.Hour

// Cahut sends the notices of the Cahut c that are due at now: to the person
// who set the access of a Window that ends within three days, once for that
// end; to the administrators of the Cahut, and of the server, when a limit
// is at 90% or reached, once a day; and in an individual Cahut, to a person
// whose own limit is.
func Cahut(ctx context.Context, srv *server.Server, c *cahut.Cahut, now time.Time) error {
	if c.Outbox == nil {
		return nil
	}
	return errors.Join(accessEnds(ctx, c, now), limits(ctx, srv, c, now))
}

// once sends l, made for the person userID as Prepare decides, when the
// notice of kind about key may go now, at most once every every, or once
// for good when every is zero. The Cahut records it as sent.
func once(ctx context.Context, c *cahut.Cahut, userID int64, kind, key string, every time.Duration, now time.Time,
	write func(p mail.Person, name, where string) mail.Letter) error {
	tx, err := c.DB.Write.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck // a no-op after Commit
	l, err := c.Outbox.Prepare(ctx, tx, userID, mail.Optional, write)
	if err != nil || l == nil {
		return err
	}
	if ok, err := mail.Once(ctx, tx, kind, key, every, now); err != nil || !ok {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	c.Outbox.Send(*l)
	return nil
}

// accessEnds tells the person who set the access of a Window, other than
// its Room, that it ends within three days.
func accessEnds(ctx context.Context, c *cahut.Cahut, now time.Time) error {
	rows, err := c.DB.Read.QueryContext(ctx, `
		SELECT w.id, r.name, w.name, w.title, w.access_ends_at, w.access_set_by
		FROM windows w JOIN rooms r ON r.id = w.room_id
		WHERE w.access != 'room' AND w.trashed_at IS NULL AND w.access_set_by IS NOT NULL
			AND w.access_ends_at > ? AND w.access_ends_at <= ?`,
		store.FormatTime(now), store.FormatTime(now.Add(Lead)))
	if err != nil {
		return err
	}
	type ending struct {
		id, by            int64
		room, name, title string
		ends              string
	}
	var list []ending
	for rows.Next() {
		var e ending
		if err := rows.Scan(&e.id, &e.room, &e.name, &e.title, &e.ends, &e.by); err != nil {
			rows.Close()
			return err
		}
		list = append(list, e)
	}
	if err := errors.Join(rows.Err(), rows.Close()); err != nil {
		return err
	}
	var errs []error
	for _, e := range list {
		ends, err := store.ParseTime(e.ends)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		key := strconv.FormatInt(e.id, 10) + " " + e.ends
		errs = append(errs, once(ctx, c, e.by, "access_ends", key, 0, now, func(p mail.Person, _, where string) mail.Letter {
			return mail.Letter{
				Subject: e.title + " closes to its readers on " + ends.UTC().Format("2 January"),
				Lines: []string{"Hello " + p.Name + ",", "The access that you set for " + e.title + ", on " + where + ", ends on " +
					mail.Day(ends) + ". After that, it opens to its Room only."},
				Action: &mail.Action{Label: "Open " + e.title, URL: c.Outbox.Base + e.room + "/" + e.name + "/"},
				After:  []string{"To keep it open, set its access again in the console, or ask your AI. You can turn off these notices on the page of your account."},
			}
		}))
	}
	return errors.Join(errs...)
}

// level says how close a use is to its limit: "reached", "90%", or empty.
func level(use, limit int64) string {
	switch {
	case use >= limit:
		return "reached"
	case use*10 >= limit*9:
		return "90%"
	}
	return ""
}

// limit is a limit of a Cahut, as a notice names it.
type limit struct {
	name, what string // such as windows, and Windows out of the trash
	use, limit int64
	bytes      bool
}

func (l limit) show(n int64) string {
	if l.bytes {
		return settings.Bytes(n)
	}
	return strconv.FormatInt(n, 10)
}

// limitLetter writes the notice of the limit l at level, to someone called
// hello, about the Cahut called where.
func limitLetter(l limit, lvl, hello, name, where string, own bool) mail.Letter {
	subject := name + " is at 90% of its limit of " + l.what
	state := where + " holds " + l.show(l.use) + " of the " + l.show(l.limit) + " " + l.what + " it may have."
	if own {
		subject = "You are at 90% of your limit of " + l.what + " on " + name
		state = "You hold " + l.show(l.use) + " of the " + l.show(l.limit) + " " + l.what + " you may have on " + where + "."
	}
	if lvl == "reached" {
		subject = name + " reached its limit of " + l.what
		if own {
			subject = "You reached your limit of " + l.what + " on " + name
		}
		state += " It refuses what would add to it."
	}
	return mail.Letter{
		Subject: subject,
		Lines:   []string{"Hello " + hello + ",", state},
		After:   []string{"The administrators of the server set the limits: ask them for more, or make room, for instance by purging the trash."},
	}
}

// limits tells the administrators of the Cahut, and of the server, of a
// limit of the Cahut at 90% or reached; in an individual Cahut, each person
// of their own limits.
func limits(ctx context.Context, srv *server.Server, c *cahut.Cahut, now time.Time) error {
	tx, err := c.DB.Read.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return err
	}
	s, err := settings.Read(ctx, tx, 0, true)
	if err == nil && s.Kind == settings.Individual {
		err = people(ctx, c, tx, s.Limits, now)
	}
	_ = tx.Rollback() // it only read
	if err != nil || s.Usage == nil {
		return err
	}
	var all []limit
	add := func(name, what string, use int64, most *int64, bytes bool) {
		if most != nil {
			all = append(all, limit{name: name, what: what, use: use, limit: *most, bytes: bytes})
		}
	}
	u, l := *s.Usage, s.Limits
	add("people", "people", u.People, l.People, false)
	add("admins", "administrators", u.Admins, l.Admins, false)
	add("rooms", "Rooms", u.Rooms, l.Rooms, false)
	add("windows", "Windows", u.Windows, l.Windows, false)
	add("disk", "disk", u.Disk, l.Disk, true)
	admins, err := cahutAdmins(ctx, c)
	if err != nil {
		return err
	}
	var errs []error
	for _, lim := range all {
		lvl := level(lim.use, lim.limit)
		if lvl == "" {
			continue
		}
		key := lim.name + " " + lvl
		for _, id := range admins {
			errs = append(errs, once(ctx, c, id, "limit", key+" "+strconv.FormatInt(id, 10), 24*time.Hour, now,
				func(p mail.Person, name, where string) mail.Letter {
					return limitLetter(lim, lvl, p.Name, name, where, false)
				}))
		}
		name, where, err := c.Outbox.Where(ctx)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		errs = append(errs, srv.NotifyAdmins(ctx, "limit", c.Domain+" "+key, 24*time.Hour, func(a server.Admin) mail.Letter {
			return limitLetter(lim, lvl, a.Name, name, where, false)
		}))
	}
	return errors.Join(errs...)
}

// cahutAdmins returns the administrators of the Cahut c who are not
// disabled.
func cahutAdmins(ctx context.Context, c *cahut.Cahut) ([]int64, error) {
	rows, err := c.DB.Read.QueryContext(ctx, "SELECT id FROM users WHERE role = 'admin' AND disabled_at IS NULL ORDER BY id")
	if err != nil {
		return nil, err
	}
	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, errors.Join(rows.Err(), rows.Close())
}

// people tells each person of an individual Cahut of their own limits at
// 90% or reached.
func people(ctx context.Context, c *cahut.Cahut, tx *sql.Tx, l settings.Limits, now time.Time) error {
	if l.RoomsPerPerson == nil && l.WindowsPerPerson == nil && l.DiskPerPerson == nil {
		return nil
	}
	rows, err := tx.QueryContext(ctx, "SELECT id FROM users WHERE disabled_at IS NULL ORDER BY id")
	if err != nil {
		return err
	}
	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return err
		}
		ids = append(ids, id)
	}
	if err := errors.Join(rows.Err(), rows.Close()); err != nil {
		return err
	}
	var errs []error
	for _, id := range ids {
		u, err := settings.UseOf(ctx, tx, id)
		if err != nil {
			return err
		}
		for _, lim := range []limit{
			{name: "rooms_per_person", what: "Rooms", use: u.Rooms, limit: deref(l.RoomsPerPerson)},
			{name: "windows_per_person", what: "Windows", use: u.Windows, limit: deref(l.WindowsPerPerson)},
			{name: "disk_per_person", what: "disk", use: u.Disk, limit: deref(l.DiskPerPerson), bytes: true},
		} {
			if lim.limit < 0 {
				continue
			}
			lvl := level(lim.use, lim.limit)
			if lvl == "" {
				continue
			}
			key := fmt.Sprintf("%s %s %d", lim.name, lvl, id)
			errs = append(errs, once(ctx, c, id, "limit", key, 24*time.Hour, now, func(p mail.Person, name, where string) mail.Letter {
				return limitLetter(lim, lvl, p.Name, name, where, true)
			}))
		}
	}
	return errors.Join(errs...)
}

// deref returns the limit l, or -1 when there is none.
func deref(l *int64) int64 {
	if l == nil {
		return -1
	}
	return *l
}

// Full is how full the disk of the data of the server may be before its
// administrators hear of it.
const Full = 0.9

// Disk measures the file system of a directory: its size, and the space
// left on it, as doctor.DiskUsage does.
type Disk func(dir string) (total, free uint64, err error)

// Server sends the notices of the server that are due: its disk almost
// full, as disk measures it, once a day.
func Server(ctx context.Context, srv *server.Server, disk Disk) error {
	total, free, err := disk(srv.DataDir())
	if err != nil {
		return err
	}
	if total == 0 || float64(total-free) < Full*float64(total) {
		return nil
	}
	host := srv.Mail.Host()
	pct := int(100 * float64(total-free) / float64(total))
	return srv.NotifyAdmins(ctx, "disk", "data", 24*time.Hour, func(a server.Admin) mail.Letter {
		return mail.Letter{
			Subject: "The disk of " + host + " is almost full",
			Lines: []string{"Hello " + a.Name + ",", fmt.Sprintf("The disk that holds the data of the Cahut server %s is %d%% full, with %s free.",
				host, pct, settings.Bytes(int64(free)))}, //nolint:gosec // free space fits in an int64
			After: []string{"Below 1 GB free, no Window can be published. Make room, or give the server more disk: cahut doctor says more."},
		}
	})
}

// Certificate tells the administrators of the server, once a day, that the
// certificate of domain could not be renewed, and why.
func Certificate(ctx context.Context, srv *server.Server, domain string, reason error) error {
	host := srv.Mail.Host()
	return srv.NotifyAdmins(ctx, "certificate", domain, 24*time.Hour, func(a server.Admin) mail.Letter {
		return mail.Letter{
			Subject: "The certificate of " + domain + " cannot be renewed",
			Lines: []string{"Hello " + a.Name + ",", "The Cahut server " + host + " could not renew the certificate of " + domain +
				": " + reason.Error() + "."},
			After: []string{"It tries again on its own. If the domain no longer points to the server, or its ports 80 and 443 are closed, " +
				"readers will soon see an error: cahut doctor says more."},
		}
	})
}
