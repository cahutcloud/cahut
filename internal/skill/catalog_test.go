package skill

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"unicode/utf8"

	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/settings"
)

func (f *fixture) catalog(t *testing.T, budget int) (string, int64) {
	t.Helper()
	lines, revision, err := Catalog(context.Background(), f.db.Read, budget)
	if err != nil {
		t.Fatal(err)
	}
	return lines, revision
}

func (f *fixture) markEveryWindow(t *testing.T, name string) {
	t.Helper()
	if _, err := f.db.Write.ExecContext(context.Background(), "UPDATE skills SET every_window = 1 WHERE owner_id IS NULL AND name = ?", name); err != nil {
		t.Fatal(err)
	}
}

func TestCatalogOrder(t *testing.T) {
	f := setup(t)
	f.write(t, f.olivier, WriteInput{Scope: ScopeCahut, Skill: "zeta", Text: textOf("zeta", "Z.", "")})
	f.write(t, f.olivier, WriteInput{Scope: ScopeCahut, Skill: "alpha", Text: textOf("alpha", "A.", "")})
	f.write(t, f.olivier, WriteInput{Scope: ScopeCahut, Skill: "charte-acme", Text: textOf("charte-acme", "The brand of Acme.", "")})
	f.markEveryWindow(t, "charte-acme")
	lines, revision := f.catalog(t, CatalogBytes)
	want := "The skills of this Cahut today:\n- charte-acme: The brand of Acme.\n- alpha: A.\n- zeta: Z."
	if lines != want {
		t.Errorf("catalog =\n%s\nwant\n%s", lines, want)
	}
	if revision != f.revision(t) || revision == 0 {
		t.Errorf("revision = %d, want %d", revision, f.revision(t))
	}
	if r, err := Revision(context.Background(), f.db.Read); err != nil || r != revision {
		t.Errorf("Revision = %d, %v, want %d", r, err, revision)
	}
}

func TestCatalogCutsDescriptionsOnAWord(t *testing.T) {
	f := setup(t)
	long := strings.Repeat("Relancer une facture impayée ", 8) // 29 characters, eight times
	f.write(t, f.olivier, WriteInput{Scope: ScopeCahut, Skill: "relance-facture", Text: textOf("relance-facture", long, "")})
	exact := strings.Repeat("é", 160)
	f.write(t, f.olivier, WriteInput{Scope: ScopeCahut, Skill: "exact", Text: textOf("exact", exact, "")})
	folded := "---\nname: folded\ndescription: >\n  The brand\n  of Acme.\n---\n"
	f.write(t, f.olivier, WriteInput{Scope: ScopeCahut, Skill: "folded", Text: folded})
	lines, _ := f.catalog(t, 10_000)
	got := map[string]string{}
	for _, l := range strings.Split(lines, "\n")[1:] {
		name, description, _ := strings.Cut(strings.TrimPrefix(l, "- "), ": ")
		got[name] = description
	}
	// 160 characters end in the sixth time, within "facture": the cut goes
	// back to the end of the word before, at 157 characters.
	if want := strings.Repeat("Relancer une facture impayée ", 5) + "Relancer une…"; got["relance-facture"] != want {
		t.Errorf("relance-facture = %q, want %q", got["relance-facture"], want)
	}
	if got["exact"] != exact {
		t.Errorf("a description of 160 characters is cut: %q", got["exact"])
	}
	if got["folded"] != "The brand of Acme." {
		t.Errorf("a description over lines = %q, want it on one line", got["folded"])
	}
	for name, d := range got {
		if n := utf8.RuneCountInString(d); n > 161 {
			t.Errorf("%s: %d characters", name, n)
		}
	}
}

func TestCatalogWithinItsBudget(t *testing.T) {
	f := setup(t)
	for i := range 50 {
		name := fmt.Sprintf("skill-%02d", i)
		description := string([]rune(strings.Repeat(fmt.Sprintf("Le skill %02d dit «comment» <faire> & quoi. ", i), 25))[:1024])
		f.write(t, f.olivier, WriteInput{Scope: ScopeCahut, Skill: name, Text: textOf(name, `"`+strings.ReplaceAll(description, `"`, `\"`)+`"`, "")})
	}
	lines, _ := f.catalog(t, CatalogBytes)
	encoded, _ := json.Marshal(lines)
	if len(encoded)-2 > CatalogBytes {
		t.Errorf("the catalog weighs %d bytes of JSON, more than %d", len(encoded)-2, CatalogBytes)
	}
	all := strings.Split(lines, "\n")
	shown := len(all) - 2 // the heading, and the last line
	if shown < 1 {
		t.Fatalf("no skill in the catalog:\n%s", lines)
	}
	if last, want := all[len(all)-1], fmt.Sprintf("and %d more skills: find_skills lists them", 50-shown); last != want {
		t.Errorf("last line = %q, want %q", last, want)
	}
	if !strings.HasPrefix(all[1], "- skill-00: ") {
		t.Errorf("first line = %q", all[1])
	}

	// One more skill than fits: the last line counts it alone.
	f = setup(t)
	f.write(t, f.olivier, WriteInput{Scope: ScopeCahut, Skill: "alpha", Text: textOf("alpha", "A.", "")})
	f.write(t, f.olivier, WriteInput{Scope: ScopeCahut, Skill: "beta", Text: textOf("beta", strings.Repeat("B", 100), "")})
	whole, _ := f.catalog(t, 10_000)
	cut := "The skills of this Cahut today:\n- alpha: A.\nand 1 more skill: find_skills lists it"
	if lines, _ = f.catalog(t, jsonBytes(whole)-1); lines != cut {
		t.Errorf("catalog =\n%s\nwant\n%s", lines, cut)
	}
	if lines, _ = f.catalog(t, jsonBytes(cut)-1); lines != "The skills of this Cahut today:\nand 2 more skills: find_skills lists them" {
		t.Errorf("catalog of a budget too small for one line =\n%s", lines)
	}
	if lines, _ = f.catalog(t, jsonBytes(whole)); lines != whole {
		t.Errorf("a catalog that fits its budget exactly is cut:\n%s", lines)
	}
}

func TestNoCatalogInAnIndividualCahut(t *testing.T) {
	f := setup(t)
	f.write(t, f.olivier, WriteInput{Scope: ScopeCahut, Skill: "charte-acme", Text: textOf("charte-acme", "The brand of Acme.", "")})
	if _, err := operation.Do(context.Background(), f.r, settings.Set, cli, settings.SetCahutSettingsInput{Kind: settings.Individual}); err != nil {
		t.Fatal(err)
	}
	if lines, _ := f.catalog(t, CatalogBytes); lines != "" {
		t.Errorf("catalog of an individual Cahut =\n%s", lines)
	}
}

func TestCatalogLeavesOutPersonalSkills(t *testing.T) {
	f := setup(t)
	f.write(t, f.camille, WriteInput{Scope: ScopePersonal, Skill: "notes", Text: textOf("notes", "My notes.", "")})
	before := f.revision(t)
	if lines, _ := f.catalog(t, CatalogBytes); lines != "" {
		t.Errorf("catalog with a personal skill only =\n%s", lines)
	}
	f.write(t, f.olivier, WriteInput{Scope: ScopeCahut, Skill: "charte-acme", Text: textOf("charte-acme", "The brand of Acme.", "")})
	lines, revision := f.catalog(t, CatalogBytes)
	if want := "The skills of this Cahut today:\n- charte-acme: The brand of Acme."; lines != want {
		t.Errorf("catalog =\n%s\nwant\n%s", lines, want)
	}
	if revision != before+1 {
		t.Errorf("revision = %d, want %d", revision, before+1)
	}
}
