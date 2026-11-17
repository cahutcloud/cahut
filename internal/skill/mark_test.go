package skill

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"testing"

	"cahut.cloud/internal/logbook"
	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/settings"
)

func encode(b []byte) string { return base64.StdEncoding.EncodeToString(b) }

// rendered is text as Cahut keeps it, and as AIs read it.
func (f *fixture) rendered(t *testing.T, text string) string {
	t.Helper()
	r, err := RenderText(text)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

// mark writes the mark of the skill name of the Cahut, as the console does,
// without its text.
func (f *fixture) mark(who operation.Actor, scope Scope, name string, every bool) (Skill, error) {
	return operation.Do(context.Background(), f.r, WriteSkill, who, WriteInput{Scope: scope, Skill: name, EveryWindow: new(every)})
}

func TestMarkASkillForEveryWindow(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	f.write(t, f.olivier, WriteInput{Scope: ScopeCahut, Skill: "charte-acme", Text: textOf("charte-acme", "The brand of Acme.", "# Acme\n")})
	before := f.revision(t)

	// The mark makes no new version, and raises the revision of the
	// catalog, whose lines it orders.
	s, err := f.mark(f.olivier, ScopeCahut, "charte-acme", true)
	if err != nil {
		t.Fatalf("mark: %v", err)
	}
	if !s.EveryWindow || s.Number != 1 {
		t.Errorf("write_skill = %+v, want version 1 for every Window", s)
	}
	if f.revision(t) != before+1 {
		t.Errorf("the revision is %d, want %d", f.revision(t), before+1)
	}
	if g := f.get(t, f.camille, "charte-acme"); !g.EveryWindow {
		t.Errorf("get_skill = %+v, want it for every Window", g)
	}
	found, err := operation.Do(ctx, f.r, FindSkills, f.camille, FindInput{})
	if err != nil || len(found.Skills) == 0 || !found.Skills[0].EveryWindow {
		t.Errorf("find_skills = %+v, %v", found, err)
	}

	// A new version without every_window keeps the mark.
	s = f.write(t, f.olivier, WriteInput{Scope: ScopeCahut, Skill: "charte-acme", Text: textOf("charte-acme", "The brand of Acme, again.", "# Acme\n")})
	if !s.EveryWindow || s.Number != 2 {
		t.Errorf("a new version = %+v, want version 2, still for every Window", s)
	}

	// The same mark again changes nothing.
	at := f.revision(t)
	if _, err := f.mark(f.olivier, ScopeCahut, "charte-acme", true); err != nil || f.revision(t) != at {
		t.Errorf("the same mark again: %v, revision %d, want %d", err, f.revision(t), at)
	}

	// The mark comes off.
	if s, err = f.mark(f.olivier, ScopeCahut, "charte-acme", false); err != nil || s.EveryWindow || s.Number != 2 {
		t.Errorf("unmark = %+v, %v", s, err)
	}
	if f.revision(t) != at+1 {
		t.Errorf("the revision is %d, want %d", f.revision(t), at+1)
	}

	// A new skill takes its mark as it is written.
	formulaires := textOf("formulaires", "The notices of forms.", "")
	s = f.write(t, f.olivier, WriteInput{Scope: ScopeCahut, Skill: "formulaires", Text: formulaires, EveryWindow: new(true)})
	if !s.EveryWindow {
		t.Errorf("a new skill for every Window = %+v", s)
	}

	entries, err := logbook.List(ctx, f.db.Read, 20)
	if err != nil {
		t.Fatal(err)
	}
	var marks []string
	for _, e := range entries {
		if e.Operation == "write_skill" && strings.Contains(e.Detail, "every Window") {
			marks = append(marks, e.Target+": "+e.Detail)
		}
	}
	want := []string{
		fmt.Sprintf("formulaires: cahut, version 1, 0 files, %s, for every Window", settings.Bytes(int64(len(f.rendered(t, formulaires))))),
		"charte-acme: cahut, no longer for every Window",
		"charte-acme: cahut, for every Window",
	}
	if strings.Join(marks, "\n") != strings.Join(want, "\n") {
		t.Errorf("the Logbook keeps\n%s\nwant\n%s", strings.Join(marks, "\n"), strings.Join(want, "\n"))
	}
}

func TestAtMostThreeSkillsForEveryWindow(t *testing.T) {
	f := setup(t)
	for _, name := range []string{"alpha", "beta", "gamma", "delta"} {
		f.write(t, f.olivier, WriteInput{Scope: ScopeCahut, Skill: name, Text: textOf(name, "A skill.", "")})
	}
	for _, name := range []string{"alpha", "beta", "gamma"} {
		if _, err := f.mark(f.olivier, ScopeCahut, name, true); err != nil {
			t.Fatalf("mark %s: %v", name, err)
		}
	}
	// A skill that has the mark keeps it.
	if _, err := f.mark(f.olivier, ScopeCahut, "beta", true); err != nil {
		t.Errorf("mark beta again: %v", err)
	}
	before := f.revision(t)
	_, err := f.mark(f.olivier, ScopeCahut, "delta", true)
	if refusal(err) != "every_window_too_many" || !errors.Is(err, operation.ErrForbidden) {
		t.Errorf("a fourth mark: %v, want the refusal every_window_too_many", err)
	}
	_, err = operation.Do(context.Background(), f.r, WriteSkill, f.olivier,
		WriteInput{Scope: ScopeCahut, Skill: "epsilon", Text: textOf("epsilon", "A skill.", ""), EveryWindow: new(true)})
	if refusal(err) != "every_window_too_many" {
		t.Errorf("a new skill for every Window: %v, want the refusal every_window_too_many", err)
	}
	if got := f.find(t, f.olivier, "epsilon"); len(got) != 0 {
		t.Errorf("the refused skill was written: %v", got)
	}
	if f.revision(t) != before {
		t.Errorf("the refusals raised the revision to %d, from %d", f.revision(t), before)
	}

	// Once a mark comes off, another may take it.
	if _, err := f.mark(f.olivier, ScopeCahut, "alpha", false); err != nil {
		t.Fatal(err)
	}
	if _, err := f.mark(f.olivier, ScopeCahut, "delta", true); err != nil {
		t.Errorf("mark delta once alpha is unmarked: %v", err)
	}
}

func TestASkillForEveryWindowStaysShort(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	long := textOf("charte-acme", "The brand of Acme.", strings.Repeat("Titles in Inter. ", 1000)) // about 17 KB
	short := textOf("charte-acme", "The brand of Acme.", "Titles in Inter.\n")
	tooLong := func(what string, err error) {
		t.Helper()
		if refusal(err) != "every_window_too_long" || !errors.Is(err, operation.ErrInvalid) {
			t.Errorf("%s: %v, want the refusal every_window_too_long", what, err)
		}
	}

	// A skill that no Window follows may be long, and then takes no mark.
	f.write(t, f.olivier, WriteInput{Scope: ScopeCahut, Skill: "charte-acme", Text: long})
	_, err := f.mark(f.olivier, ScopeCahut, "charte-acme", true)
	tooLong("the mark of a long skill", err)
	if g := f.get(t, f.olivier, "charte-acme"); g.EveryWindow {
		t.Errorf("the refused mark was kept: %+v", g)
	}

	// A marked skill takes no long version, written or imported.
	f.write(t, f.olivier, WriteInput{Scope: ScopeCahut, Skill: "charte-acme", Text: short})
	if _, err := f.mark(f.olivier, ScopeCahut, "charte-acme", true); err != nil {
		t.Fatalf("mark: %v", err)
	}
	_, err = operation.Do(ctx, f.r, WriteSkill, f.olivier, WriteInput{Scope: ScopeCahut, Skill: "charte-acme", Text: long})
	tooLong("a long version of a marked skill", err)
	zip, err := WriteZip("charte-acme", long, nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = operation.Do(ctx, f.r, ImportSkill, f.olivier, ImportInput{Scope: ScopeCahut, Zip: encode(zip)})
	tooLong("an import of a long version of a marked skill", err)

	// Nor does it go back to a long version.
	_, err = operation.Do(ctx, f.r, RollbackSkill, f.olivier, RollbackInput{Skill: "charte-acme"})
	tooLong("a rollback to a long version of a marked skill", err)
	if g := f.get(t, f.olivier, "charte-acme"); g.Text != f.rendered(t, short) || !g.EveryWindow {
		t.Errorf("after the refusals, get_skill = %+v", g)
	}

	// A new skill for every Window is short too.
	_, err = operation.Do(ctx, f.r, WriteSkill, f.olivier,
		WriteInput{Scope: ScopeCahut, Skill: "formulaires", Text: strings.Replace(long, "charte-acme", "formulaires", 1), EveryWindow: new(true)})
	tooLong("a new long skill for every Window", err)
}

func TestOnlyAdministratorsMarkASkill(t *testing.T) {
	f := setup(t)
	f.write(t, f.olivier, WriteInput{Scope: ScopeCahut, Skill: "charte-acme", Text: textOf("charte-acme", "The brand of Acme.", "")})
	f.write(t, f.camille, WriteInput{Scope: ScopePersonal, Skill: "notes", Text: textOf("notes", "My notes.", "")})

	_, err := f.mark(f.camille, ScopeCahut, "charte-acme", true)
	if refusal(err) != "skill_admins_only" {
		t.Errorf("a member marks a skill of the Cahut: %v, want the refusal skill_admins_only", err)
	}
	if g := f.get(t, f.olivier, "charte-acme"); g.EveryWindow {
		t.Errorf("the member's mark was kept: %+v", g)
	}
	// A personal skill is never for every Window, an administrator's
	// included.
	for _, who := range []operation.Actor{f.camille, f.olivier} {
		if who.Handle == "olivier" {
			f.write(t, f.olivier, WriteInput{Scope: ScopePersonal, Skill: "notes", Text: textOf("notes", "My notes.", "")})
		}
		_, err := f.mark(who, ScopePersonal, "notes", true)
		if !errors.Is(err, operation.ErrInvalid) || !strings.Contains(err.Error(), "skill of the Cahut") {
			t.Errorf("%s marks a personal skill: %v, want an invalid call", who.Handle, err)
		}
	}
}
