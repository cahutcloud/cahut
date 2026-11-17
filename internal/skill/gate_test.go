package skill_test

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"cahut.cloud/internal/account"
	"cahut.cloud/internal/logbook"
	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/room"
	"cahut.cloud/internal/settings"
	"cahut.cloud/internal/skill"
	"cahut.cloud/internal/store"
	"cahut.cloud/internal/windata"
	"cahut.cloud/internal/window"
)

var cli = operation.Actor{Kind: operation.Admin, Via: "cli"}

// gated is a collaborative Cahut with Windows and skills: olivier, an
// administrator in the console, camille, a member in the console, and
// Claude Code acting for camille.
type gated struct {
	r                *operation.Registry
	db               *store.DB
	olivier, camille operation.Actor
	ai               operation.Actor
}

func setupGate(t *testing.T) *gated {
	t.Helper()
	ctx := context.Background()
	dir := t.TempDir()
	db, err := store.Open(ctx, filepath.Join(dir, "cahut.db"), store.CahutSchema)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	r := operation.NewRegistry(db)
	r.Individual = func(ctx context.Context, tx *sql.Tx) (bool, error) { return settings.IsIndividual(ctx, tx) }
	r.Suspended = settings.CheckActive
	account.Register(r)
	room.Register(r)
	operation.Register(r, settings.Set)
	skill.Register(r, nil)
	data := windata.NewPool(windata.Options{})
	t.Cleanup(data.Close)
	window.NewService(filepath.Join(dir, "windows"), "https://cahut.test/", []byte("the key that signs frames in tests"), data).Register(r)

	g := &gated{r: r, db: db}
	for _, p := range []struct {
		handle string
		to     *operation.Actor
	}{{"olivier", &g.olivier}, {"camille", &g.camille}} {
		u, err := operation.Do(ctx, r, account.CreateUser, cli, account.CreateUserInput{Handle: p.handle, Name: p.handle, Admin: p.handle == "olivier"})
		if err != nil {
			t.Fatal(err)
		}
		*p.to = operation.Actor{Kind: operation.Person, Via: "api", UserID: u.ID, Handle: p.handle, Session: 1}
	}
	tok, err := operation.Do(ctx, r, account.CreateToken, cli, account.CreateTokenInput{User: "camille", Name: "Claude Code"})
	if err != nil {
		t.Fatal(err)
	}
	g.ai = operation.Actor{Kind: operation.AI, Via: "mcp", UserID: g.camille.UserID, Handle: "camille", TokenID: tok.ID, Client: "Claude Code"}
	return g
}

const charte = "---\nname: charte-acme\ndescription: The brand of Acme, for every Window.\n---\n# Acme\n\nTitles in Inter, on a white background.\n"

// brand writes the skill charte-acme of the Cahut with text, marked for
// every Window, and returns its version.
func (g *gated) brand(t *testing.T, text string) skill.Skill {
	t.Helper()
	s, err := operation.Do(context.Background(), g.r, skill.WriteSkill, g.olivier, skill.WriteInput{
		Scope: skill.ScopeCahut, Skill: "charte-acme", Text: text, EveryWindow: new(true),
		Files: []skill.FileInput{{Path: "assets/logo.svg", Content: "<svg/>"}, {Path: "references/deck.md", Content: "# Deck\n"}},
	})
	if err != nil {
		t.Fatalf("write_skill charte-acme: %v", err)
	}
	return s
}

// call calls the operation name as who, with in.
func (g *gated) call(who operation.Actor, name string, in any) (any, error) {
	raw, err := json.Marshal(in)
	if err != nil {
		return nil, err
	}
	return g.r.Invoke(context.Background(), name, who, raw)
}

// publish calls publish_window as who, with follows.
func (g *gated) publish(who operation.Actor, follows ...string) (window.Window, error) {
	out, err := g.call(who, "publish_window", window.PublishInput{
		Title: "Budget 2027", Kind: window.Deck, Follows: follows,
		Files:       []window.File{{Path: "index.html", Content: "<h1>Budget 2027</h1>"}},
		Declaration: &window.Declaration{},
	})
	w, _ := out.(window.Window)
	return w, err
}

// update calls update_window as who, with follows.
func (g *gated) update(who operation.Actor, id string, follows ...string) error {
	_, err := g.call(who, "update_window", window.UpdateInput{
		Window: id, Follows: follows,
		Files: []window.File{{Path: "index.html", Content: "<h1>Budget 2027, again</h1>"}}, Declaration: &window.Declaration{},
	})
	return err
}

// windows returns how many Windows the Cahut holds.
func (g *gated) windows(t *testing.T) int {
	t.Helper()
	var n int
	if err := g.db.Read.QueryRowContext(context.Background(), "SELECT count(*) FROM windows").Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func TestTheGateStopsAnAIWithoutFollows(t *testing.T) {
	g := setupGate(t)
	s := g.brand(t, charte)
	_, err := g.publish(g.ai)
	if !errors.Is(err, operation.ErrInvalid) {
		t.Fatalf("publish_window without follows: %v, want an invalid call", err)
	}
	if n := g.windows(t); n != 0 {
		t.Errorf("the refused call published %d Windows", n)
	}
	// The refusal carries the skill, its files, and what to give.
	text, _ := skill.RenderText(charte)
	for _, want := range []string{text, "assets/logo.svg", "references/deck.md", `follows: ["charte-acme@` + s.Version + `"]`, "charte-acme/assets/logo.svg"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the refusal does not hold %q:\n%s", want, err)
		}
	}
	// A wrong digest stops it too.
	if _, err := g.publish(g.ai, "charte-acme@00000000"); !errors.Is(err, operation.ErrInvalid) {
		t.Errorf("publish_window with another version: %v, want an invalid call", err)
	}
}

func TestTheGatePassesWithTheCurrentVersion(t *testing.T) {
	g := setupGate(t)
	s := g.brand(t, charte)
	if _, err := g.publish(g.ai, "charte-acme@"+s.Version); err != nil {
		t.Errorf("publish_window with the current version: %v", err)
	}
	// The name may take its scope as a prefix, as find_skills gives it.
	w, err := g.publish(g.ai, "cahut:charte-acme@"+s.Version)
	if err != nil {
		t.Fatalf("publish_window with the prefix of the scope: %v", err)
	}
	if err := g.update(g.ai, w.ID, "charte-acme@"+s.Version); err != nil {
		t.Errorf("update_window with the current version: %v", err)
	}
}

func TestTheGateAsksAgainAfterANewVersion(t *testing.T) {
	g := setupGate(t)
	first := g.brand(t, charte)
	w, err := g.publish(g.ai, "charte-acme@"+first.Version)
	if err != nil {
		t.Fatal(err)
	}
	second := g.brand(t, strings.Replace(charte, "white", "warm grey", 1))
	if second.Version == first.Version {
		t.Fatal("the new version has the digest of the first")
	}

	// The digest of the first version no longer passes, and the refusal
	// gives the new text.
	err = g.update(g.ai, w.ID, "charte-acme@"+first.Version)
	if !errors.Is(err, operation.ErrInvalid) || !strings.Contains(err.Error(), "warm grey") || !strings.Contains(err.Error(), "charte-acme@"+second.Version) {
		t.Errorf("update_window with the first version: %v, want the refusal of the gate with the new text", err)
	}
	if _, err := g.publish(g.ai, "charte-acme@"+first.Version); !errors.Is(err, operation.ErrInvalid) {
		t.Errorf("publish_window with the first version: %v, want an invalid call", err)
	}
	if _, err := g.publish(g.ai, "charte-acme@"+second.Version); err != nil {
		t.Errorf("publish_window with the new version: %v", err)
	}
}

func TestTheGateNamesEverySkillMissing(t *testing.T) {
	g := setupGate(t)
	brand := g.brand(t, charte)
	forms, err := operation.Do(context.Background(), g.r, skill.WriteSkill, g.olivier, skill.WriteInput{
		Scope: skill.ScopeCahut, Skill: "formulaires", EveryWindow: new(true),
		Text: "---\nname: formulaires\ndescription: The notices of forms.\n---\nSay what a form keeps.\n",
	})
	if err != nil {
		t.Fatal(err)
	}
	// One of the two is not enough, and the refusal gives both.
	_, err = g.publish(g.ai, "charte-acme@"+brand.Version)
	want := `follows: ["charte-acme@` + brand.Version + `", "formulaires@` + forms.Version + `"]`
	if !errors.Is(err, operation.ErrInvalid) || !strings.Contains(err.Error(), want) || !strings.Contains(err.Error(), "Say what a form keeps.") {
		t.Errorf("publish_window with one skill of two: %v, want the refusal with %s", err, want)
	}
	if err != nil && strings.Contains(err.Error(), "Titles in Inter") {
		t.Errorf("the refusal carries the skill that the call follows already:\n%s", err)
	}
	if _, err := g.publish(g.ai, "formulaires@"+forms.Version, "charte-acme@"+brand.Version); err != nil {
		t.Errorf("publish_window with both: %v", err)
	}
}

func TestTheGateNeverStopsAPerson(t *testing.T) {
	g := setupGate(t)
	g.brand(t, charte)
	if _, err := g.publish(g.camille); err != nil {
		t.Errorf("camille publishes in the console without follows: %v", err)
	}
}

func TestNoGateWithoutAMarkedSkill(t *testing.T) {
	g := setupGate(t)
	g.brand(t, charte)
	if _, err := operation.Do(context.Background(), g.r, skill.WriteSkill, g.olivier,
		skill.WriteInput{Scope: skill.ScopeCahut, Skill: "charte-acme", EveryWindow: new(false)}); err != nil {
		t.Fatal(err)
	}
	if _, err := g.publish(g.ai); err != nil {
		t.Errorf("publish_window without a marked skill: %v", err)
	}
}

func TestNoGateInAnIndividualCahut(t *testing.T) {
	g := setupGate(t)
	g.brand(t, charte)
	if _, err := operation.Do(context.Background(), g.r, settings.Set, cli, settings.SetCahutSettingsInput{Kind: settings.Individual}); err != nil {
		t.Fatal(err)
	}
	// What the Cahut kept from before stops nobody.
	if _, err := g.publish(g.ai); err != nil {
		t.Errorf("publish_window in an individual Cahut: %v", err)
	}
}

func TestTheLogbookKeepsFollowsAndTheGate(t *testing.T) {
	g := setupGate(t)
	ctx := context.Background()
	s := g.brand(t, charte)
	if _, err := operation.Do(ctx, g.r, skill.WriteSkill, g.camille,
		skill.WriteInput{Scope: skill.ScopePersonal, Skill: "mes-notes", Text: "---\nname: mes-notes\ndescription: My notes.\n---\nNotes.\n"}); err != nil {
		t.Fatal(err)
	}
	notes := g.findPersonal(t)
	if _, err := g.publish(g.ai); err == nil {
		t.Fatal("publish_window without follows passed the gate")
	}
	if _, err := g.publish(g.ai, "charte-acme@"+s.Version, "mes-notes@"+notes); err != nil {
		t.Fatal(err)
	}
	entries, err := logbook.List(ctx, g.db.Read, 5)
	if err != nil {
		t.Fatal(err)
	}
	published, refused := entries[0], entries[1]
	if refused.Operation != "publish_window" || refused.Outcome != logbook.Failed ||
		refused.Error != "invalid input: gate: charte-acme@"+s.Version {
		t.Errorf("the refusal of the gate is logged as %+v", refused)
	}
	if published.Operation != "publish_window" || published.Outcome != logbook.OK ||
		!strings.HasSuffix(published.Detail, "; follows charte-acme@"+s.Version+" and 1 other") {
		t.Errorf("the publication is logged as %+v", published)
	}
	// The administrators read the Logbook: it never names a personal skill,
	// nor keeps the text of a skill.
	for _, e := range entries {
		line, _ := json.Marshal(e)
		if strings.Contains(string(line), "mes-notes") || strings.Contains(string(line), "Titles in Inter") {
			t.Errorf("the Logbook keeps %s", line)
		}
	}
}

// findPersonal returns the version of the personal skill mes-notes of
// camille, as get_skill gives it to her AI.
func (g *gated) findPersonal(t *testing.T) string {
	t.Helper()
	s, err := operation.Do(context.Background(), g.r, skill.GetSkill, g.ai, skill.GetInput{Skill: "mes-notes"})
	if err != nil {
		t.Fatal(err)
	}
	return s.Version
}

func TestGetSkillSaysWhatToGiveInFollows(t *testing.T) {
	g := setupGate(t)
	s := g.brand(t, charte)
	got, err := operation.Do(context.Background(), g.r, skill.GetSkill, g.ai, skill.GetInput{Skill: "charte-acme"})
	if err != nil {
		t.Fatal(err)
	}
	if want := `give follows: ["charte-acme@` + s.Version + `"] to publish_window and update_window`; !strings.Contains(got.Note, want) {
		t.Errorf("the note of get_skill is %q, want it to say %q", got.Note, want)
	}
}
