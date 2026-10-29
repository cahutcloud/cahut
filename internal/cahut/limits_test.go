package cahut

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"cahut.cloud/internal/account"
	"cahut.cloud/internal/logbook"
	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/room"
	"cahut.cloud/internal/settings"
)

// limits is a Cahut whose limits the tests set, with the command line as
// the administrator of the server.
type limits struct {
	t     *testing.T
	ctx   context.Context
	c     *Cahut
	admin operation.Actor
}

func openLimits(t *testing.T) *limits {
	t.Helper()
	ctx := context.Background()
	c, err := Open(ctx, Config{Dir: t.TempDir(), Domain: "cahut.example.com"})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { c.Close() })
	return &limits{t: t, ctx: ctx, c: c, admin: operation.Actor{Kind: operation.Admin, Via: "cli"}}
}

func n(v int64) *int64 { return &v }

func (l *limits) set(lim settings.Limits) settings.Settings {
	l.t.Helper()
	s, err := operation.Do(l.ctx, l.c.Registry, settings.Set, l.admin, settings.SetCahutSettingsInput{Limits: lim})
	if err != nil {
		l.t.Fatalf("set_cahut_settings: %v", err)
	}
	return s
}

func (l *limits) about(who operation.Actor) About {
	l.t.Helper()
	a, err := operation.Do(l.ctx, l.c.Registry, l.c.About, who, GetCahutInput{})
	if err != nil {
		l.t.Fatalf("get_cahut: %v", err)
	}
	return a
}

// person creates a person and returns them as they act in the console.
func (l *limits) person(handle string, isAdmin bool) operation.Actor {
	l.t.Helper()
	u, err := l.createUser(l.admin, handle, isAdmin)
	if err != nil {
		l.t.Fatalf("create_user %s: %v", handle, err)
	}
	return operation.Actor{Kind: operation.Person, Via: "api", UserID: u.ID, Handle: handle, Session: 1}
}

func (l *limits) createUser(who operation.Actor, handle string, isAdmin bool) (account.NewUser, error) {
	return operation.Do(l.ctx, l.c.Registry, account.CreateUser, who, account.CreateUserInput{Handle: handle, Name: handle, Admin: isAdmin})
}

func (l *limits) invoke(who operation.Actor, name string, args any) (any, error) {
	l.t.Helper()
	raw, err := json.Marshal(args)
	if err != nil {
		l.t.Fatal(err)
	}
	return l.c.Registry.Invoke(l.ctx, name, who, raw)
}

// publish publishes a Window of size bytes in room, as who.
func (l *limits) publish(who operation.Actor, room, title string, size int) error {
	l.t.Helper()
	_, err := l.invoke(who, "publish_window", map[string]any{
		"title": title, "kind": "document", "room": room, "declaration": map[string]any{"scripts": false},
		"files": []map[string]string{{"path": "index.html", "content": strings.Repeat("x", size)}},
	})
	return err
}

func (l *limits) update(who operation.Actor, window string, size int) error {
	l.t.Helper()
	_, err := l.invoke(who, "update_window", map[string]any{
		"window": window, "declaration": map[string]any{"scripts": false},
		"files": []map[string]string{{"path": "index.html", "content": strings.Repeat("x", size)}},
	})
	return err
}

func (l *limits) refused(err error, about string) {
	l.t.Helper()
	if !errors.Is(err, operation.ErrForbidden) || !strings.Contains(err.Error(), about) {
		l.t.Errorf("err = %v, want a refusal that mentions %q", err, about)
	}
}

func (l *limits) ok(err error, what string) {
	l.t.Helper()
	if err != nil {
		l.t.Errorf("%s: %v", what, err)
	}
}

func TestSetCahutSettings(t *testing.T) {
	l := openLimits(t)
	olivier := l.person("olivier", true)

	// A Cahut has no limit until the administrators of its server set some.
	a := l.about(olivier)
	if a.Limits != (settings.Limits{}) || a.Usage.People != 1 || a.Usage.Admins != 1 {
		t.Errorf("get_cahut of a new Cahut = %+v", a.Settings)
	}

	s := l.set(settings.Limits{People: n(10), Admins: n(2), Disk: n(5_000_000_000)})
	if *s.Limits.People != 10 || *s.Limits.Admins != 2 || *s.Limits.Disk != 5_000_000_000 || s.Limits.Rooms != nil || s.Usage.People != 1 {
		t.Errorf("set_cahut_settings = %+v", s)
	}
	if a := l.about(olivier); a.Limits.People == nil || *a.Limits.People != 10 {
		t.Errorf("get_cahut after set = %+v", a.Settings)
	}

	// Every change goes to the Logbook of the Cahut, which its
	// administrators read.
	entries, err := logbook.List(l.ctx, l.c.DB.Read, 1)
	if err != nil {
		t.Fatal(err)
	}
	if e := entries[0]; e.Operation != "set_cahut_settings" || e.Target != "settings" ||
		e.Detail != "people none → 10, admins none → 2, disk none → 5 GB" {
		t.Errorf("logbook = %+v", e)
	}

	// Setting the limits again replaces them: a limit left out is lifted.
	s = l.set(settings.Limits{Rooms: n(3)})
	if s.Limits.People != nil || s.Limits.Disk != nil || *s.Limits.Rooms != 3 {
		t.Errorf("set_cahut_settings again = %+v", s.Limits)
	}

	// Only the administrators of the server set them: not those of the
	// Cahut, and no AI, which is never offered the operation.
	_, err = operation.Do(l.ctx, l.c.Registry, settings.Set, olivier, settings.SetCahutSettingsInput{})
	l.refused(err, "administrators of the server")
	info, _ := l.c.Registry.Lookup("set_cahut_settings")
	if info.ForAIs(true, false) {
		t.Error("set_cahut_settings is offered to the AIs of an administrator of the Cahut")
	}

	for _, bad := range []settings.Limits{{People: n(0)}, {Admins: n(0)}, {Windows: n(-1)}, {Disk: n(-5)}} {
		if _, err := operation.Do(l.ctx, l.c.Registry, settings.Set, l.admin, settings.SetCahutSettingsInput{Limits: bad}); !errors.Is(err, operation.ErrInvalid) {
			t.Errorf("set_cahut_settings %+v: err = %v, want invalid", bad, err)
		}
	}
}

func TestLimitsOfPeople(t *testing.T) {
	l := openLimits(t)
	olivier := l.person("olivier", true)
	l.set(settings.Limits{People: n(3), Admins: n(1)})
	l.person("camille", false)

	// A second administrator crosses the limit of administrators, as a new
	// one or as a member who becomes one.
	_, err := l.createUser(olivier, "bruno", true)
	l.refused(err, "1 administrators, its limit")
	_, err = operation.Do(l.ctx, l.c.Registry, account.SetUserRole, olivier, account.SetUserRoleInput{User: "camille", Role: account.Admin})
	l.refused(err, "make one of them a member first")

	l.person("bruno", false)
	_, err = l.createUser(olivier, "denise", false)
	l.refused(err, "3 people, its limit")
	if !strings.Contains(err.Error(), "ask the administrators of the server") {
		t.Errorf("an administrator of the Cahut is told %q, want to ask the administrators of the server", err)
	}

	// Disabled people do not count: disabling one makes room, and enabling
	// them again counts them again.
	_, err = operation.Do(l.ctx, l.c.Registry, account.DisableUser, olivier, account.UserInput{User: "bruno"})
	l.ok(err, "disable_user bruno")
	_, err = l.createUser(olivier, "denise", false)
	l.ok(err, "create_user denise after disabling bruno")
	_, err = operation.Do(l.ctx, l.c.Registry, account.EnableUser, olivier, account.UserInput{User: "bruno"})
	l.refused(err, "3 people, its limit")

	// An AI that would cross a limit is refused at once, without a Knock.
	ai := olivier
	ai.Kind, ai.Via, ai.TokenID, ai.Client = operation.AI, "mcp", 1, "Claude Code"
	_, err = l.createUser(ai, "emile", false)
	l.refused(err, "3 people, its limit")

	// Lowering a limit below what the Cahut holds deletes nobody.
	l.set(settings.Limits{People: n(1)})
	if a := l.about(olivier); a.Usage.People != 3 {
		t.Errorf("people after lowering the limit = %d, want 3", a.Usage.People)
	}
	l.set(settings.Limits{})
	_, err = operation.Do(l.ctx, l.c.Registry, account.EnableUser, olivier, account.UserInput{User: "bruno"})
	l.ok(err, "enable_user bruno once the limit is lifted")
}

func TestLimitOfRooms(t *testing.T) {
	l := openLimits(t)
	l.person("olivier", true)
	camille := l.person("camille", false)
	l.set(settings.Limits{Rooms: n(1)})

	create := func(title string) error {
		_, err := operation.Do(l.ctx, l.c.Registry, room.CreateRoom, camille, room.CreateRoomInput{Title: title})
		return err
	}
	l.ok(create("Marketing"), "create_room Marketing")
	err := create("Sales")
	l.refused(err, "1 team Rooms, its limit")
	if !strings.Contains(err.Error(), "ask an administrator of the Cahut") {
		t.Errorf("a member is told %q, want to ask an administrator of the Cahut", err)
	}
	// Personal Rooms come with their person and do not count.
	l.person("bruno", false)
	if a := l.about(camille); a.Usage.Rooms != 1 {
		t.Errorf("team Rooms = %d, want 1", a.Usage.Rooms)
	}
}

func TestLimitsOfWindows(t *testing.T) {
	l := openLimits(t)
	camille := l.person("camille", false)
	_, err := operation.Do(l.ctx, l.c.Registry, room.CreateRoom, camille, room.CreateRoomInput{Title: "Marketing"})
	l.ok(err, "create_room")
	l.set(settings.Limits{Windows: n(2), WindowsPerRoom: n(1)})

	l.ok(l.publish(camille, "camille", "One", 10), "publish One")
	l.refused(l.publish(camille, "camille", "Two", 10), "the Room camille holds 1 Windows")
	l.ok(l.publish(camille, "marketing", "Two", 10), "publish Two in marketing")
	l.set(settings.Limits{Windows: n(2)})
	l.refused(l.publish(camille, "camille", "Three", 10), "this Cahut holds 2 Windows out of the trash")

	// A Window in the trash does not count, and restoring it counts it again.
	_, err = l.invoke(camille, "trash_window", map[string]string{"window": "camille/one"})
	l.ok(err, "trash_window")
	l.ok(l.publish(camille, "camille", "Three", 10), "publish Three once One is in the trash")
	_, err = l.invoke(camille, "restore_window", map[string]string{"window": "camille/one"})
	l.refused(err, "this Cahut holds 2 Windows out of the trash")
	if a := l.about(camille); a.Usage.Windows != 2 {
		t.Errorf("Windows = %d, want 2", a.Usage.Windows)
	}
}

func TestLimitOfDisk(t *testing.T) {
	l := openLimits(t)
	olivier := l.person("olivier", true)
	l.set(settings.Limits{Disk: n(1000)})

	l.ok(l.publish(olivier, "", "Report", 600), "publish 600 bytes")
	l.refused(l.publish(olivier, "", "Notes", 500), "keeps 600 bytes, and 500 bytes more would take it past its limit of 1 kB")

	// An update keeps the version before: the disk holds both.
	l.ok(l.update(olivier, "olivier/report", 300), "update to 300 bytes")
	if a := l.about(olivier); a.Usage.Disk != 900 {
		t.Errorf("disk = %d, want 900", a.Usage.Disk)
	}
	// The next one drops the first version and adds the new one.
	l.refused(l.update(olivier, "olivier/report", 800), "keeps 900 bytes, and 200 bytes more")
	l.ok(l.update(olivier, "olivier/report", 650), "update to 650 bytes, which drops the version of 600")
	if a := l.about(olivier); a.Usage.Disk != 950 {
		t.Errorf("disk = %d, want 950", a.Usage.Disk)
	}

	// A Window in the trash still takes its room on the disk, until it is
	// purged.
	_, err := l.invoke(olivier, "trash_window", map[string]string{"window": "olivier/report"})
	l.ok(err, "trash_window")
	l.refused(l.publish(olivier, "", "Notes", 100), "purge Windows from the trash")
	_, err = l.invoke(olivier, "purge_window", map[string]string{"window": "olivier/report"})
	l.ok(err, "purge_window")
	l.ok(l.publish(olivier, "", "Notes", 100), "publish once Report is purged")

	// The files of the house style count too.
	l.set(settings.Limits{Disk: n(150)})
	logo := `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 10 10"><rect width="10" height="10"/></svg>`
	_, err = l.invoke(olivier, "set_house_style", map[string]any{"logo": map[string]string{"content": logo}})
	l.refused(err, "keeps 100 bytes, and")
	_, err = l.invoke(olivier, "set_house_style", map[string]any{"name": "Example Co"})
	l.ok(err, "set_house_style without files")

	// Lowering the limit below what the Cahut keeps deletes nothing, and a
	// call that frees bytes still runs.
	l.ok(l.update(olivier, "olivier/notes", 40), "update Notes to 40 bytes")
	l.set(settings.Limits{Disk: n(10)})
	if a := l.about(olivier); a.Usage.Disk != 140 || a.Usage.Windows != 1 {
		t.Errorf("after lowering the limit, usage = %+v", a.Usage)
	}
	l.ok(l.update(olivier, "olivier/notes", 30), "an update that drops 100 bytes for 30")
	l.refused(l.update(olivier, "olivier/notes", 120), fmt.Sprintf("its limit of %s", settings.Bytes(10)))
}

func TestBytes(t *testing.T) {
	for n, want := range map[int64]string{0: "0 bytes", 999: "999 bytes", 1000: "1 kB", 1500: "1.5 kB", 5_000_000_000: "5 GB", 12_345_678: "12.3 MB"} {
		if got := settings.Bytes(n); got != want {
			t.Errorf("Bytes(%d) = %q, want %q", n, got, want)
		}
	}
}

func TestNear(t *testing.T) {
	for _, c := range []struct {
		use, limit int64
		want       bool
	}{
		{89, 100, false}, {90, 100, true}, {100, 100, true}, {120, 100, true},
		{8, 10, false}, {9, 10, true}, {13, 15, false}, {14, 15, true},
		{0, 1, false}, {1, 1, true}, {0, 0, false}, {3, 0, false},
		{8_999_999_999, 10_000_000_000, false}, {9_000_000_000, 10_000_000_000, true},
	} {
		if got := settings.Near(c.use, c.limit); got != c.want {
			t.Errorf("Near(%d, %d) = %v, want %v", c.use, c.limit, got, c.want)
		}
	}
}
