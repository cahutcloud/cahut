package cahut

import (
	"context"
	"testing"

	"cahut.cloud/internal/account"
	"cahut.cloud/internal/operation"
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
	for _, handle := range []string{"olivier", "camille"} {
		u, err := operation.Do(ctx, c.Registry, account.CreateUser, admin, account.CreateUserInput{Handle: handle, Name: handle, Admin: handle == "olivier"})
		if err != nil {
			t.Fatal(err)
		}
		people[handle] = operation.Actor{Kind: operation.Person, Via: "api", UserID: u.ID, Handle: handle, Session: 1}
		if _, err := operation.Do(ctx, c.Registry, account.CreateToken, people[handle], account.CreateTokenInput{Name: "Claude Code"}); err != nil {
			t.Fatal(err)
		}
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
	// An administrator reads everything; a member, what they did.
	if got := targets(people["olivier"]); len(got) != 4 {
		t.Errorf("the administrator reads %q, want 4 entries", got)
	}
	if got := targets(people["camille"]); len(got) != 1 || got[0] != "create_token token 2 of camille" {
		t.Errorf("the member reads %q, want their token only", got)
	}
	if _, err := operation.Do(ctx, c.Registry, ListLogbook, admin, ListLogbookInput{Limit: 1001}); err == nil {
		t.Error("list_logbook with a limit above 1000 succeeded")
	}
}
