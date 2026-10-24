package cahut

import (
	"encoding/json"
	"errors"
	"regexp"
	"strings"
	"testing"

	"cahut.cloud/internal/account"
	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/room"
	"cahut.cloud/internal/settings"
)

func (l *limits) setKind(kind string, lim settings.Limits) (settings.Settings, error) {
	return operation.Do(l.ctx, l.c.Registry, settings.Set, l.admin, settings.SetCahutSettingsInput{Kind: kind, Limits: lim})
}

func (l *limits) createRoom(who operation.Actor, in room.CreateRoomInput) (room.Room, error) {
	return operation.Do(l.ctx, l.c.Registry, room.CreateRoom, who, in)
}

func (l *limits) access(who operation.Actor, window, access string) error {
	l.t.Helper()
	_, err := l.invoke(who, "set_window_access", map[string]string{"window": window, "access": access})
	return err
}

func asAI(p operation.Actor) operation.Actor {
	p.Kind, p.Via, p.Session, p.TokenID, p.Client = operation.AI, "mcp", 0, 1, "Claude Code"
	return p
}

func TestBecomeIndividual(t *testing.T) {
	l := openLimits(t)
	l.person("olivier", true)
	alice := l.person("alice", false)
	l.person("bob", false)
	_, err := l.createRoom(alice, room.CreateRoomInput{Title: "Marketing"})
	l.ok(err, "create_room")
	_, err = operation.Do(l.ctx, l.c.Registry, room.AddMember, alice, room.MemberInput{Room: "marketing", User: "bob"})
	l.ok(err, "add_room_member")
	l.ok(l.publish(alice, "alice", "Plan", 10), "publish")
	l.ok(l.access(alice, "alice/plan", "internal"), "set_window_access internal")

	// A Cahut becomes individual only once no Room is shared and no Window
	// opens to the whole Cahut: the refusal names them.
	_, err = l.setKind(settings.Individual, settings.Limits{})
	if !errors.Is(err, operation.ErrInvalid) || !strings.Contains(err.Error(), "these Rooms have more than one member: marketing") ||
		!strings.Contains(err.Error(), "these Windows open to everyone signed in to the Cahut: alice/plan") {
		t.Errorf("set_cahut_settings individual: %v", err)
	}
	_, err = operation.Do(l.ctx, l.c.Registry, room.RemoveMember, alice, room.RemoveMemberInput{Room: "marketing", User: "bob"})
	l.ok(err, "remove_room_member")
	l.ok(l.access(alice, "alice/plan", "room"), "set_window_access room")
	s, err := l.setKind(settings.Individual, settings.Limits{WindowsPerPerson: n(5)})
	if err != nil || s.Kind != settings.Individual || *s.Limits.WindowsPerPerson != 5 {
		t.Fatalf("set_cahut_settings individual = %+v, %v", s, err)
	}

	// The kind stays when it is left out, and limits per person bound the
	// people of an individual Cahut only.
	if s, err := l.setKind("", settings.Limits{}); err != nil || s.Kind != settings.Individual {
		t.Errorf("set_cahut_settings without a kind = %+v, %v", s, err)
	}
	if _, err := l.setKind(settings.Collaborative, settings.Limits{RoomsPerPerson: n(2)}); !errors.Is(err, operation.ErrInvalid) {
		t.Errorf("limits per person in a collaborative Cahut: %v", err)
	}
	if _, err := l.setKind("shared", settings.Limits{}); !errors.Is(err, operation.ErrInvalid) {
		t.Errorf("an unknown kind: %v", err)
	}
	if s, err := l.setKind(settings.Collaborative, settings.Limits{}); err != nil || s.Kind != settings.Collaborative {
		t.Errorf("back to collaborative = %+v, %v", s, err)
	}
}

func TestIndividualCahut(t *testing.T) {
	l := openLimits(t)
	olivier := l.person("olivier", true)
	alice := l.person("alice", false)
	bob := l.person("bob", false)
	if _, err := l.setKind(settings.Individual, settings.Limits{}); err != nil {
		t.Fatal(err)
	}

	// The name of a Room starts with the handle of its owner.
	for _, tt := range []struct {
		who        operation.Actor
		in         room.CreateRoomInput
		name, role string
	}{
		{alice, room.CreateRoomInput{Title: "Projects"}, "alice-projects", room.Owner},
		{alice, room.CreateRoomInput{Title: "Drafts", Name: "drafts"}, "alice-drafts", room.Owner},
		{alice, room.CreateRoomInput{Title: "Notes", Name: "alice-notes"}, "alice-notes", room.Owner},
		{alice, room.CreateRoomInput{Title: "Bob"}, "alice-bob", room.Owner},
		{bob, room.CreateRoomInput{Title: "Projects"}, "bob-projects", room.Owner},
		{olivier, room.CreateRoomInput{Title: "Shared", Owner: "bob"}, "bob-shared", ""},
	} {
		r, err := l.createRoom(tt.who, tt.in)
		if err != nil || r.Name != tt.name || r.Role != tt.role {
			t.Errorf("create_room %+v by %s = %+v, %v; want %s", tt.in, tt.who.Handle, r, err, tt.name)
		}
	}
	long := strings.Repeat("Very long title ", 6)
	if r, err := l.createRoom(alice, room.CreateRoomInput{Title: long}); err != nil || len(r.Name) > 63 || !strings.HasPrefix(r.Name, "alice-very-long") {
		t.Errorf("create_room with a long title = %+v, %v", r, err)
	}
	_, err := l.createRoom(alice, room.CreateRoomInput{Title: "Projects"})
	if !errors.Is(err, operation.ErrInvalid) || !strings.Contains(err.Error(), `you already have a Room "alice-projects"`) {
		t.Errorf("create_room again: %v", err)
	}

	// Nobody is added to a Room, and AIs are not offered the operations
	// that manage members.
	_, err = operation.Do(l.ctx, l.c.Registry, room.AddMember, alice, room.MemberInput{Room: "alice-projects", User: "bob"})
	l.refused(err, "individual")
	_, err = operation.Do(l.ctx, l.c.Registry, room.AddMember, olivier, room.MemberInput{Room: "bob-shared", User: "alice"})
	l.refused(err, "each Room keeps one member")
	for _, name := range []string{"add_room_member", "set_room_member_role", "remove_room_member"} {
		info, _ := l.c.Registry.Lookup(name)
		if info.ForAIsIn(true, true, false) {
			t.Errorf("%s is offered to AIs in an individual Cahut", name)
		}
	}

	// Only administrators list the people.
	_, err = operation.Do(l.ctx, l.c.Registry, account.ListUsers, alice, account.ListUsersInput{})
	l.refused(err, "for the administrators of this Cahut")
	if list, err := operation.Do(l.ctx, l.c.Registry, account.ListUsers, olivier, account.ListUsersInput{}); err != nil || len(list.Users) != 3 {
		t.Errorf("list_users by an administrator = %+v, %v", list, err)
	}
	if info, _ := l.c.Registry.Lookup("list_users"); info.ForAIsIn(true, false, false) || !info.ForAIsIn(true, true, false) {
		t.Error("list_users: offered to the AIs of members, or not to those of administrators, in an individual Cahut")
	}

	// No Window opens to the whole Cahut.
	l.ok(l.publish(alice, "", "Plan", 10), "publish")
	if err := l.access(alice, "alice/plan", "internal"); !errors.Is(err, operation.ErrInvalid) || !strings.Contains(err.Error(), "individual") {
		t.Errorf("set_window_access internal: %v", err)
	}
	l.ok(l.access(alice, "alice/plan", "key"), "set_window_access key")

	// A member reads the limits per person and their own use only.
	a := l.about(alice)
	if a.Kind != settings.Individual || a.Usage != nil || a.Yours == nil || a.Yours.Rooms != 5 || a.Yours.Windows != 1 {
		t.Errorf("get_cahut of a member = %+v", a.Settings)
	}
	if a := l.about(olivier); a.Usage == nil || a.Usage.People != 3 || a.Yours == nil {
		t.Errorf("get_cahut of an administrator = %+v", a.Settings)
	}
}

func TestLimitsPerPerson(t *testing.T) {
	l := openLimits(t)
	olivier := l.person("olivier", true)
	alice := l.person("alice", false)
	bob := l.person("bob", false)
	if _, err := l.setKind(settings.Individual, settings.Limits{RoomsPerPerson: n(1), WindowsPerPerson: n(1), DiskPerPerson: n(100)}); err != nil {
		t.Fatal(err)
	}

	_, err := l.createRoom(alice, room.CreateRoomInput{Title: "Projects"})
	l.ok(err, "create_room")
	_, err = l.createRoom(alice, room.CreateRoomInput{Title: "Drafts"})
	l.refused(err, "you have 1 Rooms beside the personal one, the limit of each person")
	_, err = l.createRoom(olivier, room.CreateRoomInput{Title: "Drafts", Owner: "alice"})
	l.refused(err, "alice has 1 Rooms beside the personal one")
	_, err = l.createRoom(bob, room.CreateRoomInput{Title: "Projects"})
	l.ok(err, "create_room by bob, whose limit is his own")

	// The Windows of a person count over all their Rooms.
	l.ok(l.publish(alice, "", "Plan", 10), "publish")
	l.refused(l.publish(alice, "alice-projects", "Notes", 10), "you hold 1 Windows out of the trash, the limit of each person")
	l.ok(l.publish(bob, "", "Plan", 80), "publish by bob")
	if _, err := l.setKind("", settings.Limits{DiskPerPerson: n(100)}); err != nil {
		t.Fatal(err)
	}
	l.refused(l.publish(bob, "bob-projects", "Notes", 30), "your Windows keep 80 bytes, and 30 bytes more would take them past the limit of each person")
	l.ok(l.publish(bob, "bob-projects", "Notes", 20), "publish 20 bytes, up to the limit")

	// A refusal for a limit of the whole Cahut gives a member no count.
	if _, err := l.setKind("", settings.Limits{Windows: n(3)}); err != nil {
		t.Fatal(err)
	}
	err = l.publish(alice, "", "Notes", 10)
	l.refused(err, "this Cahut has no room for Windows")
	if regexp.MustCompile(`\d`).MatchString(err.Error()) {
		t.Errorf("a member learns a count: %v", err)
	}
	_, err = l.invoke(olivier, "publish_window", map[string]any{
		"title": "Notes", "kind": "document", "declaration": map[string]any{"scripts": false},
		"files": []map[string]string{{"path": "index.html", "content": "x"}},
	})
	l.refused(err, "this Cahut holds 3 Windows out of the trash, its limit")
}

// TestIndividualCahutHidesPeople calls, as a member of an individual Cahut,
// every operation that a person or their AI may call, and fails when a
// result or an error names another person or what they hold.
func TestIndividualCahutHidesPeople(t *testing.T) {
	l := openLimits(t)
	l.person("olivier", true)
	alice := l.person("alice", false)
	zoe := l.person("zoe", false)
	if _, err := l.setKind(settings.Individual, settings.Limits{Windows: n(100)}); err != nil {
		t.Fatal(err)
	}
	if _, err := l.createRoom(zoe, room.CreateRoomInput{Title: "Hideout"}); err != nil {
		t.Fatal(err)
	}
	l.ok(l.publish(zoe, "zoe-hideout", "Clandestine", 10), "publish by zoe")
	l.ok(l.publish(alice, "", "Plan", 10), "publish by alice")
	if _, err := l.createRoom(alice, room.CreateRoomInput{Title: "Projects"}); err != nil {
		t.Fatal(err)
	}

	secrets := regexp.MustCompile(`zoe|hideout|clandestine|olivier`)
	for _, who := range []operation.Actor{alice, asAI(alice)} {
		for _, op := range l.c.Registry.Ops() {
			// Every operation that a member may call in a collaborative
			// Cahut: those that an individual Cahut keeps from them must
			// refuse without naming anyone.
			if op.Level != operation.AnyPerson || who.Kind == operation.AI && !op.ForAIs(false, false) {
				continue
			}
			for _, args := range []string{`{}`, `{"all": true}`, `{"limit": 100}`} {
				out, err := l.c.Registry.Invoke(l.ctx, op.Name, who, json.RawMessage(args))
				var b []byte
				if err != nil {
					b = []byte(err.Error())
				} else if b, err = json.Marshal(out); err != nil {
					t.Fatal(err)
				}
				if m := secrets.Find(b); m != nil {
					t.Errorf("%s %s by the %s of alice names %q: %s", op.Name, args, who.Kind, m, b)
				}
			}
		}
	}
}
