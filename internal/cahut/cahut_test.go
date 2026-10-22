package cahut

import (
	"context"
	"slices"
	"testing"

	"cahut.cloud/internal/account"
	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/room"
)

func TestListLogbook(t *testing.T) {
	ctx := context.Background()
	c, err := Open(ctx, Config{Dir: t.TempDir(), Domain: "cahut.example.com"})
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	admin := operation.Actor{Kind: operation.Admin, Via: "cli"}
	people := map[string]operation.Actor{}
	for _, handle := range []string{"olivier", "camille", "bruno"} {
		u, err := operation.Do(ctx, c.Registry, account.CreateUser, admin, account.CreateUserInput{Handle: handle, Name: handle, Admin: handle == "olivier"})
		if err != nil {
			t.Fatal(err)
		}
		people[handle] = operation.Actor{Kind: operation.Person, Via: "api", UserID: u.ID, Handle: handle, Session: 1}
		if _, err := operation.Do(ctx, c.Registry, account.CreateToken, people[handle], account.CreateTokenInput{Name: "Claude Code"}); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := operation.Do(ctx, c.Registry, room.CreateRoom, people["camille"], room.CreateRoomInput{Title: "Marketing"}); err != nil {
		t.Fatal(err)
	}
	if _, err := operation.Do(ctx, c.Registry, room.AddMember, people["camille"], room.MemberInput{Room: "marketing", User: "bruno"}); err != nil {
		t.Fatal(err)
	}

	targets := func(who operation.Actor) []string {
		t.Helper()
		page, err := operation.Do(ctx, c.Registry, ListLogbook, who, ListLogbookInput{})
		if err != nil {
			t.Fatal(err)
		}
		var got []string
		for _, e := range page.Entries {
			got = append(got, e.Operation+" "+e.Target)
		}
		return got
	}
	// An administrator reads everything; anyone else, what they did, what
	// their AIs did, and what happened in their Rooms.
	if got := targets(people["olivier"]); len(got) != 8 {
		t.Errorf("the administrator reads %q, want 8 entries", got)
	}
	want := []string{"add_room_member marketing", "create_room marketing", "create_token token 3 of bruno", "create_user bruno"}
	if got := targets(people["bruno"]); !slices.Equal(got, want) {
		t.Errorf("bruno reads %q, want %q", got, want)
	}
	if _, err := operation.Do(ctx, c.Registry, ListLogbook, admin, ListLogbookInput{Limit: 1001}); err == nil {
		t.Error("list_logbook with a limit above 1000 succeeded")
	}
}

func TestGetCahut(t *testing.T) {
	ctx := context.Background()
	c, err := Open(ctx, Config{Dir: t.TempDir(), Domain: "cahut.example.com", BaseURL: "https://cahut.example.com/"})
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	admin := operation.Actor{Kind: operation.Admin, Via: "cli"}
	u, err := operation.Do(ctx, c.Registry, account.CreateUser, admin, account.CreateUserInput{Handle: "olivier", Name: "Olivier"})
	if err != nil {
		t.Fatal(err)
	}
	ai := operation.Actor{Kind: operation.AI, Via: "mcp", UserID: u.ID, Handle: "olivier", TokenID: 1, Client: "Claude"}
	out, err := c.Registry.Invoke(ctx, "get_cahut", ai, nil)
	want := About{Domain: "cahut.example.com", URL: "https://cahut.example.com/", MCP: "https://cahut.example.com/mcp"}
	if err != nil || out != want {
		t.Errorf("get_cahut = %+v, %v; want %+v", out, err, want)
	}
}
