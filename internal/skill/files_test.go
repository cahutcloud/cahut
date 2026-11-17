package skill_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"strings"
	"testing"

	"cahut.cloud/internal/logbook"
	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/settings"
	"cahut.cloud/internal/skill"
	"cahut.cloud/internal/window"
)

// publishFiles calls publish_window as who, with files and follows.
func (g *gated) publishFiles(who operation.Actor, files []window.File, follows ...string) (window.Window, error) {
	out, err := g.call(who, "publish_window", window.PublishInput{
		Title: "Budget 2027", Kind: window.Deck, Follows: follows, Files: files, Declaration: &window.Declaration{},
	})
	w, _ := out.(window.Window)
	return w, err
}

// fileOf returns the bytes of the file at path of the Window id, as
// get_window gives them to who.
func (g *gated) fileOf(t *testing.T, who operation.Actor, id, path string) []byte {
	t.Helper()
	out, err := g.call(who, "get_window", map[string]any{"window": id, "include_content": true})
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range out.(window.Detail).Files {
		if f.Path != path {
			continue
		}
		if f.Encoding == "base64" {
			b, err := base64.StdEncoding.DecodeString(f.Content)
			if err != nil {
				t.Fatal(err)
			}
			return b
		}
		return []byte(f.Content)
	}
	t.Fatalf("the Window has no file %s", path)
	return nil
}

var page = window.File{Path: "index.html", Content: `<h1>Budget 2027</h1><img src="img/logo.svg">`}

// font is a file of a skill that is not text.
var font = []byte{'w', 'O', 'F', '2', 0, 1, 0, 0, 0xff, 0xfe}

// fonts writes the skill polices of the Cahut, which no Window follows,
// with a font.
func (g *gated) fonts(t *testing.T) skill.Skill {
	t.Helper()
	s, err := operation.Do(context.Background(), g.r, skill.WriteSkill, g.olivier, skill.WriteInput{
		Scope: skill.ScopeCahut, Skill: "polices", Text: "---\nname: polices\ndescription: The fonts of Acme.\n---\nTitles in Inter.\n",
		Files: []skill.FileInput{{Path: "assets/inter.woff2", Content: base64.StdEncoding.EncodeToString(font), Encoding: "base64"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestFromSkillCopiesTheFile(t *testing.T) {
	g := setupGate(t)
	s := g.brand(t, charte)
	g.fonts(t)
	w, err := g.publishFiles(g.ai, []window.File{page,
		{Path: "img/logo.svg", FromSkill: "charte-acme/assets/logo.svg"},
		{Path: "fonts/inter.woff2", FromSkill: "cahut:polices/assets/inter.woff2"},
	}, "charte-acme@"+s.Version)
	if err != nil {
		t.Fatalf("publish_window with from_skill: %v", err)
	}
	if got := g.fileOf(t, g.camille, w.ID, "img/logo.svg"); string(got) != "<svg/>" {
		t.Errorf("img/logo.svg = %q, want the logo of the skill", got)
	}
	if got := g.fileOf(t, g.camille, w.ID, "fonts/inter.woff2"); !bytes.Equal(got, font) {
		t.Errorf("fonts/inter.woff2 = %q, want the font of the skill", got)
	}
	if want := int64(len(page.Content) + len("<svg/>") + len(font)); w.Size != want || w.FileCount != 3 {
		t.Errorf("the Window has %d files and %d bytes, want 3 and %d", w.FileCount, w.Size, want)
	}

	// update_window copies too, and a person copies as an AI does.
	_, err = g.call(g.camille, "update_window", window.UpdateInput{Window: w.ID, Declaration: &window.Declaration{},
		Files: []window.File{page, {Path: "img/logo.svg", FromSkill: "charte-acme/references/deck.md"}}})
	if err != nil {
		t.Fatalf("update_window with from_skill: %v", err)
	}
	if got := g.fileOf(t, g.camille, w.ID, "img/logo.svg"); string(got) != "# Deck\n" {
		t.Errorf("after update_window, img/logo.svg = %q", got)
	}
}

func TestFromSkillNeedsTheRightToRead(t *testing.T) {
	g := setupGate(t)
	ctx := context.Background()
	if _, err := operation.Do(ctx, g.r, skill.WriteSkill, g.olivier, skill.WriteInput{
		Scope: skill.ScopePersonal, Skill: "mes-modeles", Text: "---\nname: mes-modeles\ndescription: My templates.\n---\nMine.\n",
		Files: []skill.FileInput{{Path: "deck.html", Content: "<p>olivier's</p>"}},
	}); err != nil {
		t.Fatal(err)
	}
	for _, ref := range []string{"mes-modeles/deck.html", "personal:mes-modeles/deck.html"} {
		_, err := g.publishFiles(g.ai, []window.File{page, {Path: "deck.html", FromSkill: ref}})
		if !errors.Is(err, operation.ErrNotFound) {
			t.Errorf("camille's AI copies %s of olivier: %v, want ErrNotFound", ref, err)
		}
	}

	// In an individual Cahut, what the Cahut kept from before is read by
	// no one.
	g.fonts(t)
	if _, err := operation.Do(ctx, g.r, settings.Set, cli, settings.SetCahutSettingsInput{Kind: settings.Individual}); err != nil {
		t.Fatal(err)
	}
	_, err := g.publishFiles(g.ai, []window.File{page, {Path: "inter.woff2", FromSkill: "polices/assets/inter.woff2"}})
	if !errors.Is(err, operation.ErrNotFound) {
		t.Errorf("a copy from a skill of an individual Cahut: %v, want ErrNotFound", err)
	}
	if n := g.windows(t); n != 0 {
		t.Errorf("the refused copies published %d Windows", n)
	}
}

func TestFromSkillOfAFileThatIsNotThere(t *testing.T) {
	g := setupGate(t)
	g.fonts(t)
	for ref, want := range map[string]error{
		"polices/assets/none.woff2": operation.ErrNotFound,
		"polices/SKILL.md":          operation.ErrNotFound,
		"nothing/assets/a.woff2":    operation.ErrNotFound,
		"polices":                   operation.ErrInvalid,
		"polices/":                  operation.ErrInvalid,
		"":                          operation.ErrInvalid,
	} {
		files := []window.File{page, {Path: "a.woff2", FromSkill: ref}}
		if ref == "" {
			files[1].FromSkill, files[1].Content = "", ""
		}
		_, err := g.publishFiles(g.ai, files)
		if !errors.Is(err, want) {
			t.Errorf("from_skill %q: %v, want %v", ref, err, want)
		}
	}
	_, err := g.publishFiles(g.ai, []window.File{page, {Path: "a.woff2", FromSkill: "polices/assets/none.woff2"}})
	if err == nil || !strings.Contains(err.Error(), "get_skill") {
		t.Errorf("the refusal of a file that is not there does not say how to find the others: %v", err)
	}
}

func TestContentAndFromSkillTogether(t *testing.T) {
	g := setupGate(t)
	g.fonts(t)
	_, err := g.publishFiles(g.ai, []window.File{page, {Path: "a.woff2", Content: "x", FromSkill: "polices/assets/inter.woff2"}})
	if !errors.Is(err, operation.ErrInvalid) || !strings.Contains(err.Error(), "a.woff2") {
		t.Errorf("content and from_skill together: %v, want an invalid call that names the file", err)
	}
}

func TestNeitherContentNorFromSkill(t *testing.T) {
	g := setupGate(t)
	_, err := g.publishFiles(g.ai, []window.File{page, {Path: "css/deck.css"}})
	if !errors.Is(err, operation.ErrInvalid) || !strings.Contains(err.Error(), "css/deck.css") || !strings.Contains(err.Error(), "from_skill") {
		t.Errorf("a file without content: %v, want an invalid call that names the file and from_skill", err)
	}
}

func TestWindowLimitsCountCopiedFiles(t *testing.T) {
	g := setupGate(t)
	big := bytes.Repeat([]byte{0xff}, skill.MaxFile)
	if _, err := operation.Do(context.Background(), g.r, skill.WriteSkill, g.olivier, skill.WriteInput{
		Scope: skill.ScopeCahut, Skill: "photos", Text: "---\nname: photos\ndescription: The photos of Acme.\n---\nPhotos.\n",
		Files: []skill.FileInput{{Path: "assets/hall.jpg", Content: base64.StdEncoding.EncodeToString(big), Encoding: "base64"}},
	}); err != nil {
		t.Fatal(err)
	}
	// Each copy weighs on the Window: seventeen copies of 2 MiB are more
	// than the 32 MiB of a Window.
	files := []window.File{page}
	for i := range 17 {
		files = append(files, window.File{Path: "img/hall-" + string(rune('a'+i)) + ".jpg", FromSkill: "photos/assets/hall.jpg"})
	}
	_, err := g.publishFiles(g.ai, files)
	if !errors.Is(err, operation.ErrInvalid) || !strings.Contains(err.Error(), "MiB") {
		t.Errorf("seventeen copies of 2 MiB: %v, want the limit of a Window", err)
	}
	if _, err := g.publishFiles(g.ai, files[:16]); err != nil {
		t.Errorf("fifteen copies of 2 MiB: %v", err)
	}
}

func TestTheLogbookKeepsTheCopiedFiles(t *testing.T) {
	g := setupGate(t)
	ctx := context.Background()
	g.fonts(t)
	if _, err := operation.Do(ctx, g.r, skill.WriteSkill, g.camille, skill.WriteInput{
		Scope: skill.ScopePersonal, Skill: "mes-modeles", Text: "---\nname: mes-modeles\ndescription: My templates.\n---\nMine.\n",
		Files: []skill.FileInput{{Path: "secret-plan.html", Content: "<p>mine</p>"}},
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := g.publishFiles(g.ai, []window.File{page,
		{Path: "inter.woff2", FromSkill: "polices/assets/inter.woff2"},
		{Path: "plan.html", FromSkill: "mes-modeles/secret-plan.html"},
	}); err != nil {
		t.Fatal(err)
	}
	// A file of a personal skill that is not there is refused, and the
	// Logbook, which the administrators read, does not name it either.
	if _, err := g.publishFiles(g.ai, []window.File{page, {Path: "plan.html", FromSkill: "mes-modeles/other-plan.html"}}); !errors.Is(err, operation.ErrNotFound) {
		t.Fatalf("a file of a personal skill that is not there: %v", err)
	}
	entries, err := logbook.List(ctx, g.db.Read, 2)
	if err != nil {
		t.Fatal(err)
	}
	refused, published := entries[0], entries[1]
	if !strings.HasSuffix(published.Detail, "; copied polices/assets/inter.woff2 and a file of a personal skill") {
		t.Errorf("the publication is logged as %q", published.Detail)
	}
	if refused.Operation != "publish_window" || refused.Outcome != logbook.Failed {
		t.Errorf("the refusal is logged as %+v", refused)
	}
	for _, e := range entries {
		for _, s := range []string{"mes-modeles", "secret-plan", "other-plan"} {
			if strings.Contains(e.Target+e.Detail+e.Error, s) {
				t.Errorf("the Logbook names %q, of a personal skill: %+v", s, e)
			}
		}
	}
}

func TestGetSkillSaysHowToCopyAFile(t *testing.T) {
	g := setupGate(t)
	g.fonts(t)
	got, err := operation.Do(context.Background(), g.r, skill.GetSkill, g.ai, skill.GetInput{Skill: "polices"})
	if err != nil {
		t.Fatal(err)
	}
	if want := `from_skill: "polices/assets/inter.woff2"`; !strings.Contains(got.Note, want) {
		t.Errorf("the note of get_skill is %q, want it to say %q", got.Note, want)
	}
}

// A from_skill that names a personal skill and no file, or a file the skill
// does not have, tells the AI which file of the Window is wrong; the
// Logbook, which the administrators read, keeps neither the name nor the
// path.
func TestTheLogbookKeepsNoMalformedPersonalRef(t *testing.T) {
	g := setupGate(t)
	ctx := context.Background()
	if _, err := operation.Do(ctx, g.r, skill.WriteSkill, g.camille, skill.WriteInput{
		Scope: skill.ScopePersonal, Skill: "plan-social", Text: "---\nname: plan-social\ndescription: My plan.\n---\nMine.\n",
		Files: []skill.FileInput{{Path: "lettre.html", Content: "<p>mine</p>"}},
	}); err != nil {
		t.Fatal(err)
	}
	refs := []string{"personal:plan-social", "plan-social", "plan-social/", "/plan-social/lettre.html", "plan-social/autre.html", "personal:plan-social/autre.html"}
	for _, ref := range refs {
		_, err := g.publishFiles(g.ai, []window.File{page, {Path: "lettre.html", FromSkill: ref}})
		if err == nil {
			t.Errorf("from_skill %q: published", ref)
		}
	}
	entries, err := logbook.List(ctx, g.db.Read, len(refs))
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if strings.Contains(e.Target+e.Detail+e.Error, "plan-social") || strings.Contains(e.Target+e.Detail+e.Error, "autre.html") {
			t.Errorf("the Logbook names a personal skill: %+v", e)
		}
	}
}

// With several files copied from skills, the refusal names the one that
// is wrong.
func TestFromSkillSaysWhichFileIsWrong(t *testing.T) {
	g := setupGate(t)
	g.fonts(t)
	for _, ref := range []string{"polices/assets/none.woff2", "nothing/assets/a.woff2", "polices"} {
		_, err := g.publishFiles(g.ai, []window.File{page,
			{Path: "fonts/inter.woff2", FromSkill: "polices/assets/inter.woff2"},
			{Path: "fonts/other.woff2", FromSkill: ref},
		})
		if err == nil || !strings.Contains(err.Error(), "fonts/other.woff2") || !strings.Contains(err.Error(), ref) {
			t.Errorf("from_skill %q: %v, want a refusal that names fonts/other.woff2 and %s", ref, err, ref)
		}
	}
	// Spaces around the reference do not count.
	if _, err := g.publishFiles(g.ai, []window.File{page, {Path: "inter.woff2", FromSkill: " polices/assets/inter.woff2 "}}); err != nil {
		t.Errorf("from_skill with spaces: %v", err)
	}
}
