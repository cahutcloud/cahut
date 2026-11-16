package skill

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"testing/fstest"

	"cahut.cloud/internal/account"
	"cahut.cloud/internal/logbook"
	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/settings"
	"cahut.cloud/internal/store"
	"cahut.cloud/plugins"
)

var cli = operation.Actor{Kind: operation.Admin, Via: "cli"}

// fixture is a collaborative Cahut with olivier, an administrator, and
// camille and bruno, members, each in the console.
type fixture struct {
	r                       *operation.Registry
	db                      *store.DB
	olivier, camille, bruno operation.Actor
}

func setup(t *testing.T) *fixture {
	t.Helper()
	ctx := context.Background()
	db, err := store.Open(ctx, filepath.Join(t.TempDir(), "cahut.db"), store.CahutSchema)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	f := &fixture{r: operation.NewRegistry(db), db: db}
	f.r.Individual = func(ctx context.Context, tx *sql.Tx) (bool, error) { return settings.IsIndividual(ctx, tx) }
	f.r.Suspended = settings.CheckActive
	account.Register(f.r)
	operation.Register(f.r, settings.Set)
	Register(f.r, plugins.SkillFS())
	for _, p := range []struct {
		handle string
		to     *operation.Actor
	}{{"olivier", &f.olivier}, {"camille", &f.camille}, {"bruno", &f.bruno}} {
		u, err := operation.Do(ctx, f.r, account.CreateUser, cli, account.CreateUserInput{Handle: p.handle, Name: p.handle, Admin: p.handle == "olivier"})
		if err != nil {
			t.Fatal(err)
		}
		*p.to = operation.Actor{Kind: operation.Person, Via: "api", UserID: u.ID, Handle: p.handle, Session: 1}
	}
	return f
}

func textOf(name, description, body string) string {
	return "---\nname: " + name + "\ndescription: " + description + "\n---\n" + body
}

const relance = "---\nname: relance-facture\ndescription: L'e-mail de relance d'une facture impayée: objet, délai et signature.\n---\n" +
	"# Relance\n\nL'objet commence par [Lumen · Relance 1].\n"

func (f *fixture) write(t *testing.T, who operation.Actor, in WriteInput) Skill {
	t.Helper()
	s, err := operation.Do(context.Background(), f.r, WriteSkill, who, in)
	if err != nil {
		t.Fatalf("write_skill %s %s: %v", in.Scope, in.Skill, err)
	}
	return s
}

func (f *fixture) get(t *testing.T, who operation.Actor, ref string) Skill {
	t.Helper()
	s, err := operation.Do(context.Background(), f.r, GetSkill, who, GetInput{Skill: ref})
	if err != nil {
		t.Fatalf("get_skill %s: %v", ref, err)
	}
	return s
}

func (f *fixture) find(t *testing.T, who operation.Actor, query string) []string {
	t.Helper()
	found, err := operation.Do(context.Background(), f.r, FindSkills, who, FindInput{Query: query})
	if err != nil {
		t.Fatalf("find_skills %q: %v", query, err)
	}
	got := []string{}
	for _, s := range found.Skills {
		got = append(got, string(s.Scope)+" "+s.Skill)
	}
	return got
}

func (f *fixture) revision(t *testing.T) int64 {
	t.Helper()
	var n int64
	if err := f.db.Read.QueryRow("SELECT revision FROM skill_catalog").Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func refusal(err error) string {
	var r *operation.Refusal
	if errors.As(err, &r) {
		return r.Code
	}
	return ""
}

func TestWriteAndGetASkill(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	text := textOf("charte-acme", "The brand of Acme: for every deck.", "# Acme\n\nTitles in Inter.\n")
	files := []FileInput{
		{Path: "references/deck.md", Content: "# Deck\n"},
		{Path: "assets/logo.png", Content: base64.StdEncoding.EncodeToString([]byte{0x89, 'P', 'N', 'G', 0xff}), Encoding: "base64"},
	}
	w := f.write(t, f.olivier, WriteInput{Scope: ScopeCahut, Skill: "charte-acme", Text: text, Files: files})
	digest := Digest(text, []File{
		{Path: "references/deck.md", Content: []byte("# Deck\n")},
		{Path: "assets/logo.png", Content: []byte{0x89, 'P', 'N', 'G', 0xff}},
	})
	if w.Skill != "charte-acme" || w.Scope != ScopeCahut || w.Number != 1 || w.Version != digest || w.Text != "" || w.Previous != 0 {
		t.Errorf("write_skill = %+v, want version 1 of digest %s, without its text", w, digest)
	}

	// Every person of the Cahut reads it.
	g := f.get(t, f.camille, "charte-acme")
	rendered, _ := RenderText(text)
	if g.Text != rendered || g.Number != 1 || g.Version != digest || g.WrittenBy != "olivier" || g.Via != "console" ||
		g.WrittenAt == nil || g.Description != "The brand of Acme: for every deck." || g.Note == "" {
		t.Errorf("get_skill = %+v", g)
	}
	if got := fmt.Sprint(g.Files); got != fmt.Sprint([]FileInfo{
		{Path: "assets/logo.png", Size: 5, Type: "image/png"},
		{Path: "references/deck.md", Size: 7, Type: "text/markdown; charset=utf-8"},
	}) {
		t.Errorf("files = %s", got)
	}
	if !strings.Contains(g.Text, `description: "The brand of Acme: for every deck."`) {
		t.Errorf("the text that AIs read does not quote its description: %q", g.Text)
	}

	// And each of its files.
	file, err := operation.Do(ctx, f.r, GetSkill, f.bruno, GetInput{Skill: "charte-acme", File: "references/deck.md"})
	if err != nil || file.File != "references/deck.md" || file.Content != "# Deck\n" || file.Encoding != "utf-8" || file.Text != "" {
		t.Errorf("get_skill of a text file = %+v, %v", file, err)
	}
	file, err = operation.Do(ctx, f.r, GetSkill, f.bruno, GetInput{Skill: "charte-acme", File: "assets/logo.png"})
	if err != nil || file.Encoding != "base64" || file.Content != files[1].Content {
		t.Errorf("get_skill of a binary file = %+v, %v", file, err)
	}
	if _, err := operation.Do(ctx, f.r, GetSkill, f.bruno, GetInput{Skill: "charte-acme", File: "assets/none.png"}); !errors.Is(err, operation.ErrNotFound) {
		t.Errorf("get_skill of a file that is not there: err = %v, want ErrNotFound", err)
	}
	if _, err := operation.Do(ctx, f.r, GetSkill, f.bruno, GetInput{Skill: "charte-zeta"}); !errors.Is(err, operation.ErrNotFound) {
		t.Errorf("get_skill of a skill that is not there: err = %v, want ErrNotFound", err)
	}

	// A new version.
	text2 := strings.Replace(text, "Inter", "Plus Jakarta Sans", 1)
	w = f.write(t, f.olivier, WriteInput{Scope: ScopeCahut, Skill: "charte-acme", Text: text2})
	if w.Number != 2 || w.Previous != 1 || w.Version == digest {
		t.Errorf("second version = %+v", w)
	}
	// The same text again makes no version.
	if again := f.write(t, f.olivier, WriteInput{Scope: ScopeCahut, Skill: "charte-acme", Text: text2}); again.Number != 2 || again.Version != w.Version {
		t.Errorf("the same text again = %+v, want version 2 as it was", again)
	}

	// A person writes a skill of their own.
	mine := f.write(t, f.camille, WriteInput{Scope: ScopePersonal, Skill: "notes", Text: textOf("notes", "How I take notes.", "Short.\n")})
	if mine.Scope != ScopePersonal || mine.Number != 1 {
		t.Errorf("personal skill = %+v", mine)
	}

	for _, c := range []struct {
		in   WriteInput
		want error
	}{
		{WriteInput{Scope: ScopeCahut, Skill: "nouveau"}, operation.ErrInvalid},                                                 // a new skill needs its text
		{WriteInput{Scope: ScopeCahut, Skill: "autre", Text: textOf("charte-acme", "Another name.", "")}, operation.ErrInvalid}, // the header names another skill
		{WriteInput{Scope: "team", Skill: "x", Text: textOf("x", "X.", "")}, operation.ErrInvalid},
		{WriteInput{Scope: ScopeBuiltin, Skill: "x", Text: textOf("x", "X.", "")}, operation.ErrInvalid},
		{WriteInput{Skill: "x", Text: textOf("x", "X.", "")}, operation.ErrInvalid},
		{WriteInput{Scope: ScopeCahut, Skill: "charte-acme", Files: []FileInput{{Path: "a.md", Content: "not base64", Encoding: "base64"}}}, operation.ErrInvalid},
		{WriteInput{Scope: ScopeCahut, Skill: "charte-acme", Files: []FileInput{{Path: "a.md"}, {Path: "A.md"}}}, operation.ErrInvalid},
		{WriteInput{Scope: ScopeCahut, Skill: "charte-acme", Files: []FileInput{{Path: "references", Content: "x"}}}, operation.ErrInvalid}, // a file where a folder is
	} {
		if _, err := operation.Do(ctx, f.r, WriteSkill, f.olivier, c.in); !errors.Is(err, c.want) {
			t.Errorf("write_skill %+v: err = %v, want %v", c.in, err, c.want)
		}
	}
	if _, err := operation.Do(ctx, f.r, WriteSkill, f.olivier, WriteInput{Scope: ScopeCahut, Skill: "Charte", Text: textOf("charte", "X.", "")}); refusal(err) != "skill_name_invalid" {
		t.Errorf("an invalid name: err = %v", err)
	}
	if _, err := operation.Do(ctx, f.r, WriteSkill, f.olivier, WriteInput{Scope: ScopeCahut, Skill: "x", Text: "no header"}); refusal(err) != "skill_header_invalid" {
		t.Errorf("an invalid header: err = %v", err)
	}
	if _, err := operation.Do(ctx, f.r, WriteSkill, f.olivier, WriteInput{Scope: ScopeCahut, Skill: "charte-acme",
		Files: []FileInput{{Path: "scripts/run.sh", Content: "echo"}}}); refusal(err) != "skill_has_scripts" {
		t.Errorf("a script: err = %v", err)
	}
}

func TestWriteKeepsTheFilesItDoesNotName(t *testing.T) {
	f := setup(t)
	text := textOf("charte-acme", "The brand of Acme.", "v1\n")
	f.write(t, f.olivier, WriteInput{Scope: ScopeCahut, Skill: "charte-acme", Text: text,
		Files: []FileInput{{Path: "a.md", Content: "a1"}, {Path: "b.md", Content: "b1"}}})
	w := f.write(t, f.olivier, WriteInput{Scope: ScopeCahut, Skill: "charte-acme", Files: []FileInput{{Path: "b.md", Content: "b2"}, {Path: "c/d.md", Content: "d"}}})
	if w.Number != 2 {
		t.Errorf("version = %d, want 2", w.Number)
	}
	g := f.get(t, f.bruno, "charte-acme")
	if rendered, _ := RenderText(text); g.Text != rendered {
		t.Errorf("text = %q, want the text of version 1", g.Text)
	}
	if got := fmt.Sprint(g.Files); got != "[{a.md 2 text/markdown; charset=utf-8} {b.md 2 text/markdown; charset=utf-8} {c/d.md 1 text/markdown; charset=utf-8}]" {
		t.Errorf("files = %s", got)
	}
	b, err := operation.Do(context.Background(), f.r, GetSkill, f.bruno, GetInput{Skill: "charte-acme", File: "b.md"})
	if err != nil || b.Content != "b2" {
		t.Errorf("b.md = %+v, %v", b, err)
	}
}

func TestRemoveAFile(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	f.write(t, f.olivier, WriteInput{Scope: ScopeCahut, Skill: "charte-acme", Text: textOf("charte-acme", "The brand of Acme.", ""),
		Files: []FileInput{{Path: "a.md", Content: "a"}, {Path: "b.md", Content: "b"}}})
	f.write(t, f.olivier, WriteInput{Scope: ScopeCahut, Skill: "charte-acme", Remove: []string{"a.md"}})
	if g := f.get(t, f.bruno, "charte-acme"); len(g.Files) != 1 || g.Files[0].Path != "b.md" {
		t.Errorf("files = %v, want b.md alone", g.Files)
	}
	if _, err := operation.Do(ctx, f.r, WriteSkill, f.olivier, WriteInput{Scope: ScopeCahut, Skill: "charte-acme", Remove: []string{"a.md"}}); !errors.Is(err, operation.ErrNotFound) {
		t.Errorf("removing a file that is not there: err = %v, want ErrNotFound", err)
	}
	if _, err := operation.Do(ctx, f.r, WriteSkill, f.olivier, WriteInput{Scope: ScopeCahut, Skill: "charte-acme",
		Files: []FileInput{{Path: "b.md", Content: "b2"}}, Remove: []string{"b.md"}}); !errors.Is(err, operation.ErrInvalid) {
		t.Errorf("writing and removing the same file: err = %v, want ErrInvalid", err)
	}
}

func TestRollBackASkill(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	f.write(t, f.olivier, WriteInput{Scope: ScopeCahut, Skill: "charte-acme", Text: textOf("charte-acme", "Version one.", ""),
		Files: []FileInput{{Path: "a.md", Content: "a"}}})
	if _, err := operation.Do(ctx, f.r, RollbackSkill, f.olivier, RollbackInput{Skill: "charte-acme"}); !errors.Is(err, operation.ErrInvalid) {
		t.Errorf("rolling back a skill with one version: err = %v, want ErrInvalid", err)
	}
	f.write(t, f.olivier, WriteInput{Scope: ScopeCahut, Skill: "charte-acme", Text: textOf("charte-acme", "Version two.", ""), Remove: []string{"a.md"}})

	back, err := operation.Do(ctx, f.r, RollbackSkill, f.olivier, RollbackInput{Skill: "charte-acme"})
	if err != nil || back.Number != 1 || back.Previous != 2 {
		t.Fatalf("rollback_skill = %+v, %v; want version 1, then 2", back, err)
	}
	if g := f.get(t, f.bruno, "charte-acme"); g.Description != "Version one." || len(g.Files) != 1 {
		t.Errorf("after the rollback = %+v", g)
	}
	if found := f.find(t, f.bruno, "two"); slices.Contains(found, "cahut charte-acme") {
		t.Errorf("find_skills two after the rollback = %q, without charte-acme", found)
	}
	// A rollback goes back and forth, and a write after it keeps the version
	// that was current.
	back, err = operation.Do(ctx, f.r, RollbackSkill, f.olivier, RollbackInput{Skill: "charte-acme"})
	if err != nil || back.Number != 2 || back.Previous != 1 {
		t.Errorf("rollback_skill again = %+v, %v; want version 2, then 1", back, err)
	}
	w := f.write(t, f.olivier, WriteInput{Scope: ScopeCahut, Skill: "charte-acme", Text: textOf("charte-acme", "Version three.", "")})
	if w.Number != 3 || w.Previous != 2 {
		t.Errorf("a write after a rollback = %+v, want version 3, then 2", w)
	}
	var versions int
	if err := f.db.Read.QueryRow("SELECT count(*) FROM skill_versions").Scan(&versions); err != nil || versions != 2 {
		t.Errorf("%d versions kept, %v; want 2", versions, err)
	}
}

func TestDeleteASkill(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	f.write(t, f.olivier, WriteInput{Scope: ScopeCahut, Skill: "charte-acme", Text: textOf("charte-acme", "The brand of Acme.", ""),
		Files: []FileInput{{Path: "a.md", Content: "a"}}})
	f.write(t, f.olivier, WriteInput{Scope: ScopeCahut, Skill: "charte-acme", Text: textOf("charte-acme", "The brand of Acme, again.", "")})
	d, err := operation.Do(ctx, f.r, DeleteSkill, f.olivier, DeleteInput{Skill: "charte-acme"})
	if err != nil || d.Skill != "charte-acme" || !d.Deleted {
		t.Errorf("delete_skill = %+v, %v", d, err)
	}
	if _, err := operation.Do(ctx, f.r, GetSkill, f.bruno, GetInput{Skill: "charte-acme"}); !errors.Is(err, operation.ErrNotFound) {
		t.Errorf("get_skill after delete_skill: err = %v, want ErrNotFound", err)
	}
	if found := f.find(t, f.bruno, "acme"); len(found) != 0 {
		t.Errorf("find_skills after delete_skill = %q", found)
	}
	for _, table := range []string{"skills", "skill_versions", "skill_files", "skill_search"} {
		var n int
		if err := f.db.Read.QueryRow("SELECT count(*) FROM " + table).Scan(&n); err != nil || n != 0 {
			t.Errorf("%s holds %d rows, %v", table, n, err)
		}
	}
	if _, err := operation.Do(ctx, f.r, DeleteSkill, f.olivier, DeleteInput{Skill: "charte-acme"}); !errors.Is(err, operation.ErrNotFound) {
		t.Errorf("delete_skill twice: err = %v, want ErrNotFound", err)
	}
	if _, err := operation.Do(ctx, f.r, DeleteSkill, f.olivier, DeleteInput{Skill: "cahut"}); !errors.Is(err, operation.ErrForbidden) {
		t.Errorf("deleting the skill of the plugin: err = %v, want ErrForbidden", err)
	}
}

func TestOnlyAdministratorsWriteTheSkillsOfTheCahut(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	text := textOf("charte-acme", "The brand of Acme.", "")
	if _, err := operation.Do(ctx, f.r, WriteSkill, f.camille, WriteInput{Scope: ScopeCahut, Skill: "charte-acme", Text: text}); !errors.Is(err, operation.ErrForbidden) {
		t.Errorf("write_skill of the Cahut by a member: err = %v, want ErrForbidden", err)
	}
	zip, err := WriteZip("charte-acme", text, nil)
	if err != nil {
		t.Fatal(err)
	}
	encoded := base64.StdEncoding.EncodeToString(zip)
	if _, err := operation.Do(ctx, f.r, ImportSkill, f.camille, ImportInput{Scope: ScopeCahut, Zip: encoded}); !errors.Is(err, operation.ErrForbidden) {
		t.Errorf("import_skill of the Cahut by a member: err = %v, want ErrForbidden", err)
	}
	f.write(t, f.olivier, WriteInput{Scope: ScopeCahut, Skill: "charte-acme", Text: text})
	f.write(t, f.olivier, WriteInput{Scope: ScopeCahut, Skill: "charte-acme", Text: textOf("charte-acme", "Again.", "")})
	if _, err := operation.Do(ctx, f.r, RollbackSkill, f.camille, RollbackInput{Skill: "charte-acme"}); !errors.Is(err, operation.ErrForbidden) {
		t.Errorf("rollback_skill of the Cahut by a member: err = %v, want ErrForbidden", err)
	}
	if _, err := operation.Do(ctx, f.r, DeleteSkill, f.camille, DeleteInput{Skill: "charte-acme"}); !errors.Is(err, operation.ErrForbidden) {
		t.Errorf("delete_skill of the Cahut by a member: err = %v, want ErrForbidden", err)
	}
	// A member exports what they read, and writes their own.
	if _, err := operation.Do(ctx, f.r, ExportSkill, f.camille, ExportInput{Skill: "charte-acme"}); err != nil {
		t.Errorf("export_skill by a member: %v", err)
	}
	f.write(t, f.camille, WriteInput{Scope: ScopePersonal, Skill: "charte-acme", Text: text})
	if _, err := operation.Do(ctx, f.r, ImportSkill, f.camille, ImportInput{Scope: ScopePersonal, Zip: encoded}); err != nil {
		t.Errorf("import_skill of their own by a member: %v", err)
	}
	// The command line writes the skills of the Cahut; it has no skills of
	// its own.
	if _, err := operation.Do(ctx, f.r, WriteSkill, cli, WriteInput{Scope: ScopeCahut, Skill: "charte-cli", Text: textOf("charte-cli", "From the command line.", "")}); err != nil {
		t.Errorf("write_skill by the command line: %v", err)
	}
	if _, err := operation.Do(ctx, f.r, WriteSkill, cli, WriteInput{Scope: ScopePersonal, Skill: "x", Text: textOf("x", "X.", "")}); !errors.Is(err, operation.ErrInvalid) {
		t.Errorf("a personal skill of the command line: err = %v, want ErrInvalid", err)
	}
}

func TestNoSkillsOfTheCahutInAnIndividualCahut(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	f.write(t, f.olivier, WriteInput{Scope: ScopeCahut, Skill: "charte-acme", Text: textOf("charte-acme", "The brand of Acme.", "")})
	if _, err := operation.Do(ctx, f.r, settings.Set, cli, settings.SetCahutSettingsInput{Kind: settings.Individual}); err != nil {
		t.Fatal(err)
	}
	_, err := operation.Do(ctx, f.r, WriteSkill, f.olivier, WriteInput{Scope: ScopeCahut, Skill: "charte-zeta", Text: textOf("charte-zeta", "Z.", "")})
	if refusal(err) != "skill_scope_individual" || !errors.Is(err, operation.ErrForbidden) {
		t.Errorf("write_skill of the Cahut in an individual Cahut: err = %v, want skill_scope_individual", err)
	}
	// What the Cahut kept from before is neither listed nor read.
	if got := f.find(t, f.camille, ""); !slices.Equal(got, []string{"builtin cahut"}) {
		t.Errorf("find_skills in an individual Cahut = %q", got)
	}
	if _, err := operation.Do(ctx, f.r, GetSkill, f.camille, GetInput{Skill: "charte-acme"}); !errors.Is(err, operation.ErrNotFound) {
		t.Errorf("get_skill of the Cahut in an individual Cahut: err = %v, want ErrNotFound", err)
	}
	f.write(t, f.camille, WriteInput{Scope: ScopePersonal, Skill: "notes", Text: textOf("notes", "My notes.", "")})
	if got := f.find(t, f.camille, ""); !slices.Equal(got, []string{"personal notes", "builtin cahut"}) {
		t.Errorf("find_skills by camille = %q", got)
	}
}

func TestAdministratorsDoNotReadThePersonalSkillsOfOthers(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	f.write(t, f.camille, WriteInput{Scope: ScopePersonal, Skill: "plan-social", Text: textOf("plan-social", "Mon plan.", "")})
	for _, ref := range []string{"plan-social", "personal:plan-social"} {
		if _, err := operation.Do(ctx, f.r, GetSkill, f.olivier, GetInput{Skill: ref}); !errors.Is(err, operation.ErrNotFound) {
			t.Errorf("get_skill %s by an administrator: err = %v, want ErrNotFound", ref, err)
		}
		if _, err := operation.Do(ctx, f.r, ExportSkill, f.olivier, ExportInput{Skill: ref}); !errors.Is(err, operation.ErrNotFound) {
			t.Errorf("export_skill %s by an administrator: err = %v, want ErrNotFound", ref, err)
		}
		if _, err := operation.Do(ctx, f.r, DeleteSkill, f.olivier, DeleteInput{Skill: ref}); !errors.Is(err, operation.ErrNotFound) {
			t.Errorf("delete_skill %s by an administrator: err = %v, want ErrNotFound", ref, err)
		}
	}
	if got := f.find(t, f.olivier, "plan"); len(got) != 0 {
		t.Errorf("find_skills plan by an administrator = %q", got)
	}
	// The administrator writes a skill of the same name among their own,
	// which tells them nothing of camille's.
	f.write(t, f.olivier, WriteInput{Scope: ScopePersonal, Skill: "plan-social", Text: textOf("plan-social", "Le mien.", "")})
	if g := f.get(t, f.camille, "plan-social"); g.Description != "Mon plan." {
		t.Errorf("camille reads %+v", g)
	}
}

func TestTheSameNameInTwoScopes(t *testing.T) {
	f := setup(t)
	f.write(t, f.olivier, WriteInput{Scope: ScopeCahut, Skill: "relance-facture", Text: relance})
	f.write(t, f.camille, WriteInput{Scope: ScopePersonal, Skill: "relance-facture", Text: textOf("relance-facture", "Ma relance.", "")})
	if g := f.get(t, f.camille, "cahut:relance-facture"); g.Scope != ScopeCahut {
		t.Errorf("cahut:relance-facture = %+v", g)
	}
	if g := f.get(t, f.camille, "personal:relance-facture"); g.Scope != ScopePersonal || g.Skill != "relance-facture" {
		t.Errorf("personal:relance-facture = %+v", g)
	}
	// bruno has no skill of that name: the name alone is enough.
	if g := f.get(t, f.bruno, "relance-facture"); g.Scope != ScopeCahut {
		t.Errorf("relance-facture for bruno = %+v", g)
	}
	// find_skills names each with its scope, for the person who has both.
	if got := f.find(t, f.camille, "relance"); !slices.Equal(got, []string{"cahut cahut:relance-facture", "personal personal:relance-facture"}) {
		t.Errorf("find_skills by camille = %q", got)
	}
	if got := f.find(t, f.bruno, "relance"); !slices.Equal(got, []string{"cahut relance-facture"}) {
		t.Errorf("find_skills by bruno = %q", got)
	}
}

func TestAnAmbiguousNameIsRefused(t *testing.T) {
	f := setup(t)
	f.write(t, f.olivier, WriteInput{Scope: ScopeCahut, Skill: "relance-facture", Text: relance})
	f.write(t, f.camille, WriteInput{Scope: ScopePersonal, Skill: "relance-facture", Text: textOf("relance-facture", "Ma relance.", "")})
	_, err := operation.Do(context.Background(), f.r, GetSkill, f.camille, GetInput{Skill: "relance-facture"})
	if !errors.Is(err, operation.ErrInvalid) || !strings.Contains(err.Error(), "cahut:relance-facture") || !strings.Contains(err.Error(), "personal:relance-facture") {
		t.Errorf("get_skill of a name in two scopes: err = %v, want ErrInvalid naming both", err)
	}
	if _, err := operation.Do(context.Background(), f.r, GetSkill, f.camille, GetInput{Skill: "team:relance-facture"}); !errors.Is(err, operation.ErrInvalid) {
		t.Errorf("an unknown scope: err = %v, want ErrInvalid", err)
	}
}

func TestTheBuiltinSkill(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	if got := f.find(t, f.bruno, ""); !slices.Equal(got, []string{"builtin cahut"}) {
		t.Errorf("find_skills = %q", got)
	}
	g := f.get(t, f.bruno, "cahut")
	if g.Scope != ScopeBuiltin || g.Number != 1 || len(g.Version) != 8 || g.WrittenBy != "" || g.WrittenAt != nil || g.Via != "" ||
		!strings.Contains(g.Text, "# Cahut") || !strings.Contains(g.Text, `name: "cahut"`) {
		t.Errorf("get_skill cahut = %+v", g)
	}
	if len(g.Files) != 1 || g.Files[0].Path != "references/formats.md" {
		t.Errorf("files = %v", g.Files)
	}
	if b := f.get(t, f.bruno, "builtin:cahut"); b.Version != g.Version {
		t.Errorf("builtin:cahut = %+v", b)
	}
	file, err := operation.Do(ctx, f.r, GetSkill, f.bruno, GetInput{Skill: "cahut", File: "references/formats.md"})
	if err != nil || file.Encoding != "utf-8" || file.Content == "" {
		t.Errorf("a file of the skill of the plugin = %+v, %v", file, err)
	}
	for _, scope := range []Scope{ScopeCahut, ScopePersonal} {
		_, err := operation.Do(ctx, f.r, WriteSkill, f.olivier, WriteInput{Scope: scope, Skill: "cahut", Text: textOf("cahut", "Mine.", "")})
		if refusal(err) != "skill_name_reserved" {
			t.Errorf("write_skill %s cahut: err = %v, want skill_name_reserved", scope, err)
		}
	}
	if _, err := operation.Do(ctx, f.r, ExportSkill, f.bruno, ExportInput{Skill: "cahut"}); err != nil {
		t.Errorf("export_skill cahut: %v", err)
	}
	if _, err := operation.Do(ctx, f.r, RollbackSkill, f.olivier, RollbackInput{Skill: "cahut"}); !errors.Is(err, operation.ErrForbidden) {
		t.Errorf("rollback_skill cahut: err = %v, want ErrForbidden", err)
	}

	// A test gives a skill of the plugin of its own.
	db, err := store.Open(ctx, filepath.Join(t.TempDir(), "cahut.db"), store.CahutSchema)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	r := operation.NewRegistry(db)
	account.Register(r)
	Register(r, fstest.MapFS{"SKILL.md": {Data: []byte(textOf("cahut", "A test.", ""))}})
	u, err := operation.Do(ctx, r, account.CreateUser, cli, account.CreateUserInput{Handle: "ana", Name: "Ana"})
	if err != nil {
		t.Fatal(err)
	}
	ana := operation.Actor{Kind: operation.Person, Via: "api", UserID: u.ID, Handle: "ana", Session: 1}
	if s, err := operation.Do(ctx, r, GetSkill, ana, GetInput{Skill: "cahut"}); err != nil || s.Description != "A test." || len(s.Files) != 0 {
		t.Errorf("get_skill of a skill from a test = %+v, %v", s, err)
	}
}

func TestFindSkillsWhateverTheCaseAndAccents(t *testing.T) {
	f := setup(t)
	f.write(t, f.olivier, WriteInput{Scope: ScopeCahut, Skill: "relance-facture", Text: relance})
	f.write(t, f.olivier, WriteInput{Scope: ScopeCahut, Skill: "charte-acme", Text: textOf("charte-acme", "The brand of Acme.", "Les titres en Plümage.\n")})
	for query, want := range map[string][]string{
		"IMPAYEE":        {"cahut relance-facture"},
		"impayée":        {"cahut relance-facture"},
		"factur":         {"cahut relance-facture"},
		"lumen relance":  {"cahut relance-facture"},
		"PLUMAGE":        {"cahut charte-acme"},
		"acme brand":     {"cahut charte-acme"},
		"rien du tout":   {},
		"publish window": {"builtin cahut"},
		"":               {"cahut charte-acme", "cahut relance-facture", "builtin cahut"},
		"* \" ) :":       {"cahut charte-acme", "cahut relance-facture", "builtin cahut"},
	} {
		if got := f.find(t, f.bruno, query); !slices.Equal(got, want) {
			t.Errorf("find_skills %q = %q, want %q", query, got, want)
		}
	}
	found, err := operation.Do(context.Background(), f.r, FindSkills, f.bruno, FindInput{Query: "relance"})
	if err != nil || len(found.Skills) != 1 {
		t.Fatalf("find_skills relance = %+v, %v", found, err)
	}
	if s := found.Skills[0]; s.Description != "L'e-mail de relance d'une facture impayée: objet, délai et signature." || s.Number != 1 || len(s.Version) != 8 {
		t.Errorf("found = %+v", s)
	}
}

// asAI registers a copy of the operation op that AIs may call, in a
// registry of the fixture's database, as the batch of the skills offers
// find_skills and get_skill to AIs later.
func asAI[In, Out any](t *testing.T, f *fixture, op *operation.Op[In, Out]) (*operation.Registry, *operation.Op[In, Out]) {
	t.Helper()
	r := operation.NewRegistry(f.db)
	r.Individual = f.r.Individual
	account.Register(r)
	Register(r, plugins.SkillFS())
	copied := *op
	copied.Name += "_by_ai"
	copied.PeopleOnly = false
	operation.Register(r, &copied)
	return r, &copied
}

func ai(p operation.Actor) operation.Actor {
	return operation.Actor{Kind: operation.AI, Via: "mcp", UserID: p.UserID, Handle: p.Handle, TokenID: 1, Client: "Claude Code"}
}

func TestTheLogbookKeepsNoQuery(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	f.write(t, f.camille, WriteInput{Scope: ScopePersonal, Skill: "plan-de-licenciement", Text: textOf("plan-de-licenciement", "Le plan.", "")})
	r, find := asAI(t, f, FindSkills)
	if _, err := operation.Do(ctx, r, find, ai(f.camille), FindInput{Query: "licenciement"}); err != nil {
		t.Fatal(err)
	}
	if _, err := operation.Do(ctx, r, find, ai(f.camille), FindInput{Query: "licenciement \"("}); err != nil {
		t.Fatal(err)
	}
	r, get := asAI(t, f, GetSkill)
	if _, err := operation.Do(ctx, r, get, ai(f.bruno), GetInput{Skill: "cahut", File: "references/formats.md"}); err != nil {
		t.Fatal(err)
	}
	entries, err := logbook.List(ctx, f.db.Read, 50)
	if err != nil {
		t.Fatal(err)
	}
	var searches, reads int
	for _, e := range entries {
		switch e.Operation {
		case "find_skills_by_ai":
			searches++
			if e.Detail != "a search, 1 found" {
				t.Errorf("the entry of a search says %q, want a search, 1 found", e.Detail)
			}
		case "get_skill_by_ai":
			reads++
			if e.Target != "cahut" || e.Detail != "references/formats.md" {
				t.Errorf("the entry of a read = %q, %q", e.Target, e.Detail)
			}
		}
		if strings.Contains(e.Detail+e.Target+e.Error, "licenciement") && e.Operation != "write_skill" {
			t.Errorf("the Logbook keeps the words of a search: %+v", e)
		}
	}
	if searches != 2 || reads != 1 {
		t.Errorf("%d searches and %d reads in the Logbook, want 2 and 1", searches, reads)
	}
}

func TestTheRevisionFollowsTheSkillsOfTheCahut(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	steps := []struct {
		what string
		do   func() error
		want int64
	}{
		{"a write of the Cahut", func() error {
			_, err := operation.Do(ctx, f.r, WriteSkill, f.olivier, WriteInput{Scope: ScopeCahut, Skill: "charte-acme", Text: textOf("charte-acme", "One.", "")})
			return err
		}, 1},
		{"a personal write", func() error {
			_, err := operation.Do(ctx, f.r, WriteSkill, f.camille, WriteInput{Scope: ScopePersonal, Skill: "notes", Text: textOf("notes", "Notes.", "")})
			return err
		}, 1},
		{"a second version", func() error {
			_, err := operation.Do(ctx, f.r, WriteSkill, f.olivier, WriteInput{Scope: ScopeCahut, Skill: "charte-acme", Text: textOf("charte-acme", "Two.", "")})
			return err
		}, 2},
		{"the same text again", func() error {
			_, err := operation.Do(ctx, f.r, WriteSkill, f.olivier, WriteInput{Scope: ScopeCahut, Skill: "charte-acme", Text: textOf("charte-acme", "Two.", "")})
			return err
		}, 2},
		{"a refused write", func() error {
			_, err := operation.Do(ctx, f.r, WriteSkill, f.olivier, WriteInput{Scope: ScopeCahut, Skill: "charte-acme", Text: "no header"})
			if err == nil {
				return errors.New("not refused")
			}
			return nil
		}, 2},
		{"a rollback", func() error {
			_, err := operation.Do(ctx, f.r, RollbackSkill, f.olivier, RollbackInput{Skill: "charte-acme"})
			return err
		}, 3},
		{"a personal rollback", func() error {
			if _, err := operation.Do(ctx, f.r, WriteSkill, f.camille, WriteInput{Scope: ScopePersonal, Skill: "notes", Text: textOf("notes", "More notes.", "")}); err != nil {
				return err
			}
			_, err := operation.Do(ctx, f.r, RollbackSkill, f.camille, RollbackInput{Skill: "notes"})
			return err
		}, 3},
		{"an import", func() error {
			zip, err := WriteZip("relance-facture", relance, nil)
			if err != nil {
				return err
			}
			_, err = operation.Do(ctx, f.r, ImportSkill, f.olivier, ImportInput{Scope: ScopeCahut, Zip: base64.StdEncoding.EncodeToString(zip)})
			return err
		}, 4},
		{"a deletion", func() error {
			_, err := operation.Do(ctx, f.r, DeleteSkill, f.olivier, DeleteInput{Skill: "charte-acme"})
			return err
		}, 5},
		{"a personal deletion", func() error {
			_, err := operation.Do(ctx, f.r, DeleteSkill, f.camille, DeleteInput{Skill: "notes"})
			return err
		}, 5},
	}
	for _, s := range steps {
		if err := s.do(); err != nil {
			t.Fatalf("%s: %v", s.what, err)
		}
		if got := f.revision(t); got != s.want {
			t.Errorf("after %s, the revision is %d, want %d", s.what, got, s.want)
		}
	}
}

func TestImportAndExport(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	zip, err := WriteZip("relance-facture", relance, []File{
		{Path: "references/exemple.md", Content: []byte("Bonjour,\n")},
		{Path: "assets/signature.png", Content: []byte{1, 2, 3}},
	})
	if err != nil {
		t.Fatal(err)
	}
	imported, err := operation.Do(ctx, f.r, ImportSkill, f.olivier, ImportInput{Scope: ScopeCahut, Zip: base64.StdEncoding.EncodeToString(zip)})
	if err != nil || imported.Skill != "relance-facture" || imported.Number != 1 || len(imported.Files) != 2 {
		t.Fatalf("import_skill = %+v, %v", imported, err)
	}
	if g := f.get(t, f.bruno, "relance-facture"); g.Via != "import" {
		t.Errorf("via = %q, want import", g.Via)
	}
	exported, err := operation.Do(ctx, f.r, ExportSkill, f.bruno, ExportInput{Skill: "relance-facture"})
	if err != nil || exported.Name != "relance-facture" {
		t.Fatalf("export_skill = %+v, %v", exported, err)
	}
	data, err := base64.StdEncoding.DecodeString(exported.Zip)
	if err != nil {
		t.Fatal(err)
	}
	text, files, err := ReadZip(data)
	if err != nil {
		t.Fatal(err)
	}
	if want, _ := RenderText(relance); text != want || len(files) != 2 || !bytes.Equal(files[0].Content, []byte{1, 2, 3}) {
		t.Errorf("the exported zip holds %q and %v", text, files)
	}
	// The exported zip imports again as it is, and exports the same.
	again, err := operation.Do(ctx, f.r, ImportSkill, f.bruno, ImportInput{Scope: ScopePersonal, Zip: exported.Zip})
	if err != nil {
		t.Fatal(err)
	}
	if again.Version != Digest(text, files) {
		t.Errorf("the digest of the zip imported again = %s, want %s", again.Version, Digest(text, files))
	}
	twice, err := operation.Do(ctx, f.r, ExportSkill, f.bruno, ExportInput{Skill: "personal:relance-facture"})
	if err != nil || twice.Zip != exported.Zip {
		t.Errorf("the zip exported twice differs, %v", err)
	}
	// An import replaces the files of the skill: it keeps none it does not
	// hold.
	small, err := WriteZip("relance-facture", relance, nil)
	if err != nil {
		t.Fatal(err)
	}
	replaced, err := operation.Do(ctx, f.r, ImportSkill, f.olivier, ImportInput{Scope: ScopeCahut, Zip: base64.StdEncoding.EncodeToString(small)})
	if err != nil || replaced.Number != 2 || len(replaced.Files) != 0 {
		t.Errorf("import_skill again = %+v, %v; want version 2 without files", replaced, err)
	}
	for _, bad := range []string{"not base64!", base64.StdEncoding.EncodeToString([]byte("not a zip"))} {
		if _, err := operation.Do(ctx, f.r, ImportSkill, f.olivier, ImportInput{Scope: ScopeCahut, Zip: bad}); !errors.Is(err, operation.ErrInvalid) {
			t.Errorf("import_skill of %q: err = %v, want ErrInvalid", bad, err)
		}
	}
}

func TestTheLimits(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	big := func(n int) string { return base64.StdEncoding.EncodeToString(make([]byte, n)) }
	write := func(in WriteInput) error {
		_, err := operation.Do(ctx, f.r, WriteSkill, f.olivier, in)
		return err
	}
	text := textOf("charte-acme", "The brand of Acme.", "")
	for what, in := range map[string]WriteInput{
		"a long SKILL.md": {Scope: ScopeCahut, Skill: "charte-acme", Text: text + strings.Repeat("x", MaxText)},
		"a large file":    {Scope: ScopeCahut, Skill: "charte-acme", Text: text, Files: []FileInput{{Path: "a.bin", Content: big(MaxFile + 1), Encoding: "base64"}}},
		"too much in all": {Scope: ScopeCahut, Skill: "charte-acme", Text: text, Files: []FileInput{
			{Path: "a.bin", Content: big(MaxFile), Encoding: "base64"}, {Path: "b.bin", Content: big(MaxFile), Encoding: "base64"}}},
	} {
		if err := write(in); refusal(err) != "skill_too_large" {
			t.Errorf("%s: err = %v, want skill_too_large", what, err)
		}
	}
	var many []FileInput
	for i := range MaxFiles + 1 {
		many = append(many, FileInput{Path: fmt.Sprintf("references/%d.md", i), Content: "x"})
	}
	if err := write(WriteInput{Scope: ScopeCahut, Skill: "charte-acme", Text: text, Files: many}); refusal(err) != "skill_too_large" {
		t.Errorf("101 files: err = %v, want skill_too_large", err)
	}
	if err := write(WriteInput{Scope: ScopeCahut, Skill: "charte-acme", Text: text, Files: many[:MaxFiles]}); err != nil {
		t.Errorf("100 files: %v", err)
	}
	// Files added to a skill count with those it keeps.
	if err := write(WriteInput{Scope: ScopeCahut, Skill: "charte-acme", Files: many[MaxFiles:]}); refusal(err) != "skill_too_large" {
		t.Errorf("a 101st file: err = %v, want skill_too_large", err)
	}
	for i := 1; i < MaxPerScope; i++ {
		name := fmt.Sprintf("skill-%d", i)
		if err := write(WriteInput{Scope: ScopeCahut, Skill: name, Text: textOf(name, "One of many.", "")}); err != nil {
			t.Fatalf("skill %d: %v", i, err)
		}
	}
	err := write(WriteInput{Scope: ScopeCahut, Skill: "one-too-many", Text: textOf("one-too-many", "One too many.", "")})
	if refusal(err) != "skill_too_many" || !errors.Is(err, operation.ErrForbidden) {
		t.Errorf("a 51st skill: err = %v, want skill_too_many", err)
	}
	// A new version of a skill is not a new skill, and a person has 50 of
	// their own.
	if err := write(WriteInput{Scope: ScopeCahut, Skill: "skill-1", Text: textOf("skill-1", "A new version.", "")}); err != nil {
		t.Errorf("a new version of the 50th skill: %v", err)
	}
	if _, err := operation.Do(ctx, f.r, WriteSkill, f.olivier, WriteInput{Scope: ScopePersonal, Skill: "mine", Text: textOf("mine", "Mine.", "")}); err != nil {
		t.Errorf("a personal skill beside 50 of the Cahut: %v", err)
	}
}

func TestTheDiskCountsSkills(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	use := func() int64 {
		tx, err := f.db.Read.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
		if err != nil {
			t.Fatal(err)
		}
		defer tx.Rollback() //nolint:errcheck // it only read
		u, err := settings.Use(ctx, tx)
		if err != nil {
			t.Fatal(err)
		}
		return u.Disk
	}
	text := textOf("charte-acme", "The brand of Acme.", "")
	f.write(t, f.olivier, WriteInput{Scope: ScopeCahut, Skill: "charte-acme", Text: text, Files: []FileInput{{Path: "a.md", Content: "12345"}}})
	f.write(t, f.camille, WriteInput{Scope: ScopePersonal, Skill: "notes", Text: textOf("notes", "N.", "")})
	want := int64(len(text)+5) + int64(len(textOf("notes", "N.", "")))
	if got := use(); got != want {
		t.Errorf("the disk counts %d bytes, want %d", got, want)
	}
	// Both versions count.
	f.write(t, f.olivier, WriteInput{Scope: ScopeCahut, Skill: "charte-acme", Files: []FileInput{{Path: "b.md", Content: "123"}}})
	want += int64(len(text) + 5 + 3)
	if got := use(); got != want {
		t.Errorf("the disk counts %d bytes after a second version, want %d", got, want)
	}
	limit := want + 10
	if _, err := operation.Do(ctx, f.r, settings.Set, cli, settings.SetCahutSettingsInput{Limits: settings.Limits{Disk: &limit}}); err != nil {
		t.Fatal(err)
	}
	_, err := operation.Do(ctx, f.r, WriteSkill, f.olivier, WriteInput{Scope: ScopeCahut, Skill: "charte-zeta", Text: textOf("charte-zeta", "Z.", ""),
		Files: []FileInput{{Path: "a.md", Content: strings.Repeat("x", 20)}}})
	if !errors.Is(err, operation.ErrForbidden) || !strings.Contains(err.Error(), "delete a skill") {
		t.Errorf("a skill past the limit of the disk: err = %v, want ErrForbidden", err)
	}
	// A rollback adds nothing, and a deletion frees both versions.
	if _, err := operation.Do(ctx, f.r, RollbackSkill, f.olivier, RollbackInput{Skill: "charte-acme"}); err != nil {
		t.Errorf("a rollback at the limit: %v", err)
	}
	if _, err := operation.Do(ctx, f.r, DeleteSkill, f.olivier, DeleteInput{Skill: "charte-acme"}); err != nil {
		t.Fatal(err)
	}
	if got := use(); got != int64(len(textOf("notes", "N.", ""))) {
		t.Errorf("after the deletion, the disk counts %d bytes", got)
	}
}

func TestDeletingAPersonTakesTheirSkills(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	u, err := operation.Do(ctx, f.r, account.CreateUser, cli, account.CreateUserInput{Handle: "dana", Name: "Dana", Admin: true})
	if err != nil {
		t.Fatal(err)
	}
	dana := operation.Actor{Kind: operation.Person, Via: "api", UserID: u.ID, Handle: "dana", Session: 1}
	f.write(t, dana, WriteInput{Scope: ScopeCahut, Skill: "charte-acme", Text: textOf("charte-acme", "The brand of Acme.", "")})
	f.write(t, dana, WriteInput{Scope: ScopePersonal, Skill: "notes", Text: textOf("notes", "Notes of Dana.", ""), Files: []FileInput{{Path: "a.md", Content: "a"}}})
	// As delete_user does, in internal/window.
	if _, err := f.db.Write.ExecContext(ctx, "DELETE FROM users WHERE id = ?", u.ID); err != nil {
		t.Fatal(err)
	}
	var skills, search, files int
	if err := f.db.Read.QueryRow("SELECT (SELECT count(*) FROM skills), (SELECT count(*) FROM skill_search), (SELECT count(*) FROM skill_files)").
		Scan(&skills, &search, &files); err != nil {
		t.Fatal(err)
	}
	if skills != 1 || search != 1 || files != 0 {
		t.Errorf("%d skills, %d rows of the search and %d files left, want 1, 1 and 0", skills, search, files)
	}
	if g := f.get(t, f.bruno, "charte-acme"); g.WrittenBy != "" {
		t.Errorf("the skill of the Cahut keeps %q as its author", g.WrittenBy)
	}
}
