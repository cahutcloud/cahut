package skill

import (
	"bytes"
	"cmp"
	"context"
	"encoding/base64"
	"fmt"
	"io/fs"
	"path"
	"slices"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/settings"
	"cahut.cloud/internal/store"
)

// Register adds the operations of skills to r, with the skill of the
// plugin for Claude from the files of builtin, its SKILL.md at the root.
// See docs/adr/0064-the-skills-of-a-cahut.md.
func Register(r *operation.Registry, builtin fs.FS) {
	if builtin != nil {
		b, err := loadBuiltin(builtin)
		if err != nil {
			panic("skill: the skill of the plugin: " + err.Error()) // it is embedded: only a programming error fails
		}
		r.Provide(builtinKey{}, b)
	}
	operation.Register(r, FindSkills)
	operation.Register(r, GetSkill)
	operation.Register(r, WriteSkill)
	operation.Register(r, ImportSkill)
	operation.Register(r, ExportSkill)
	operation.Register(r, RollbackSkill)
	operation.Register(r, DeleteSkill)
	operation.Register(r, PreviewSkillCatalog)
}

// The note that get_skill gives with each skill: the rule of priority.
const note = "Follow this skill for the tasks it describes. Where skills differ, what the person asks in the conversation comes first, " +
	"then their personal skill, then the skill of the Cahut, then a skill for the same use installed in your client, " +
	"which is a copy: the skill of the Cahut has the last word over it."

// Summary is a skill, as find_skills lists it.
type Summary struct {
	Skill       string     `json:"skill" jsonschema:"its name, to give to get_skill; with its scope as a prefix, such as cahut:charte-acme, when the person has a skill of the same name in another scope"`
	Scope       Scope      `json:"scope" jsonschema:"cahut for a skill of the team, personal for one of the person, builtin for the skill of the plugin"`
	Description string     `json:"description" jsonschema:"what the skill does, and when to follow it"`
	Version     string     `json:"version" jsonschema:"the digest of its current version"`
	Number      int        `json:"number" jsonschema:"the number of its current version"`
	WrittenAt   *time.Time `json:"written_at,omitempty" jsonschema:"when its current version was written"`
	EveryWindow bool       `json:"every_window,omitempty" jsonschema:"whether every Window of this Cahut follows it"`
}

// FindInput is the input of find_skills.
type FindInput struct {
	Query string `json:"query,omitempty" jsonschema:"words that the name, the description or the text of the skills hold, whatever their case and accents; every skill without it"`
}

// Found is the output of find_skills.
type Found struct {
	Skills []Summary `json:"skills"`
}

// FileInfo describes a file of a skill.
type FileInfo struct {
	Path string `json:"path"`
	Size int64  `json:"size"`
	Type string `json:"type" jsonschema:"its media type"`
}

// Skill is a skill, as get_skill gives it, or a file of a skill.
type Skill struct {
	Skill       string     `json:"skill"`
	Scope       Scope      `json:"scope,omitempty" jsonschema:"cahut for a skill of the team, personal for one of the person, builtin for the skill of the plugin"`
	Description string     `json:"description,omitempty"`
	Version     string     `json:"version,omitempty" jsonschema:"the digest of its current version"`
	Number      int        `json:"number,omitempty" jsonschema:"the number of its current version"`
	Previous    int        `json:"previous,omitempty" jsonschema:"the number of its previous version, which rollback_skill brings back"`
	WrittenBy   string     `json:"written_by,omitempty" jsonschema:"the handle of the person who wrote the current version, empty once that person is deleted"`
	WrittenAt   *time.Time `json:"written_at,omitempty"`
	Via         string     `json:"via,omitempty" jsonschema:"console or import: how the current version came"`
	EveryWindow bool       `json:"every_window,omitempty" jsonschema:"whether every Window of this Cahut follows it"`
	Text        string     `json:"text,omitempty" jsonschema:"its SKILL.md"`
	Files       []FileInfo `json:"files,omitempty"`
	Note        string     `json:"note,omitempty"`
	// A file of the skill, read with file.
	File     string `json:"file,omitempty"`
	Content  string `json:"content,omitempty"`
	Encoding string `json:"encoding,omitempty" jsonschema:"utf-8 for text, base64 for anything else"`
}

// GetInput is the input of get_skill.
type GetInput struct {
	Skill string `json:"skill" jsonschema:"the name of the skill as find_skills gives it, such as relance-facture, with its scope as a prefix if need be, such as cahut:charte-acme"`
	File  string `json:"file,omitempty" jsonschema:"the path of one of its files, such as references/deck.md, to read that file"`
}

// FileInput is a file that write_skill adds or replaces.
type FileInput struct {
	Path     string `json:"path" jsonschema:"relative path with / as separator, such as references/deck.md or assets/logo.svg"`
	Content  string `json:"content"`
	Encoding string `json:"encoding,omitempty" jsonschema:"utf-8 (the default) or base64, for binary files"`
}

// WriteInput is the input of write_skill.
type WriteInput struct {
	Scope  Scope       `json:"scope" jsonschema:"cahut for a skill of the team, which the administrators of the Cahut write, or personal for one of your own"`
	Skill  string      `json:"skill" jsonschema:"its name: 1 to 64 lowercase letters, digits and hyphens, the name in the header of its SKILL.md"`
	Text   string      `json:"text,omitempty" jsonschema:"its SKILL.md; a new skill needs one, and a skill keeps its text without it"`
	Files  []FileInput `json:"files,omitempty" jsonschema:"the files to add or replace; the skill keeps those it does not name"`
	Remove []string    `json:"remove,omitempty" jsonschema:"the paths of the files to remove"`
	// A pointer, so that an absent mark stays as it is.
	EveryWindow *bool `json:"every_window,omitempty" jsonschema:"for a skill of the Cahut, whether every Window follows it: AIs then read it before they publish or update a Window. Three skills at most, each with a SKILL.md of 16 KiB at most; the mark stays as it is when absent"`
}

// ImportInput is the input of import_skill.
type ImportInput struct {
	Scope Scope  `json:"scope" jsonschema:"cahut for a skill of the team, which the administrators of the Cahut write, or personal for one of your own"`
	Zip   string `json:"zip" jsonschema:"the zip of the skill in base64, with its SKILL.md at its root or in a single folder; it replaces every file of a skill of the same name"`
}

// ExportInput is the input of export_skill.
type ExportInput struct {
	Skill string `json:"skill" jsonschema:"the name of the skill, with its scope as a prefix if need be"`
}

// Export is the output of export_skill.
type Export struct {
	Name string `json:"name"`
	Zip  string `json:"zip" jsonschema:"the zip of the skill in base64, in a folder of its name, as AI clients install it"`
}

// RollbackInput is the input of rollback_skill.
type RollbackInput struct {
	Skill string `json:"skill" jsonschema:"the name of the skill, with its scope as a prefix if need be"`
}

// DeleteInput is the input of delete_skill.
type DeleteInput struct {
	Skill string `json:"skill" jsonschema:"the name of the skill, with its scope as a prefix if need be"`
}

// Deleted is the output of delete_skill.
type Deleted struct {
	Skill   string `json:"skill"`
	Deleted bool   `json:"deleted"`
}

// FindSkills lists the skills that the person reads. See
// docs/adr/0064-the-skills-of-a-cahut.md.
var FindSkills = &operation.Op[FindInput, Found]{
	Name:  "find_skills",
	Title: "Find skills",
	Description: "The skills of this Cahut: how the user's team wants some tasks done, whatever the task, such as an e-mail, a report or a deck. " +
		"Before a task, read with get_skill the skill whose description fits it, and follow it. " +
		"Without query, list every skill that the user reads, their own included; with query, those whose name, description or text hold some words.",
	ReadOnly: true,
	Run:      find,
}

// GetSkill gives a skill, or one of its files.
var GetSkill = &operation.Op[GetInput, Skill]{
	Name:  "get_skill",
	Title: "Read a skill",
	Description: "Read a skill whole: its SKILL.md, its version, who wrote it, and the list of its files. " +
		"With file, read that file.",
	ReadOnly: true,
	Run:      get,
}

// WriteSkill writes a skill, as a new version.
var WriteSkill = &operation.Op[WriteInput, Skill]{
	Name:  "write_skill",
	Title: "Write a skill",
	Description: "Write a skill of the Cahut, for its administrators, or one of your own, as a new version: its SKILL.md, " +
		"the files to add or replace, and those to remove. The previous version stays, and rollback_skill brings it back.",
	PeopleOnly: true,
	Run: func(ctx context.Context, c *operation.Call, in WriteInput) (Skill, error) {
		files, err := decodeFiles(in.Files)
		if err != nil {
			return Skill{}, private(in.Scope, err)
		}
		s, err := apply(ctx, c, change{scope: in.Scope, name: in.Skill, text: in.Text, files: files, remove: in.Remove, everyWindow: in.EveryWindow, via: "console"})
		return s, private(in.Scope, err)
	},
}

// imported is what import_skill reads from its zip, in its Prepare.
type imported struct {
	name, text string
	files      []File
}

// ImportSkill writes a skill from its zip.
var ImportSkill = &operation.Op[ImportInput, Skill]{
	Name:  "import_skill",
	Title: "Import a skill",
	Description: "Import a skill from its zip, as AI clients export them, into the skills of the Cahut, for its administrators, or among your own. " +
		"A skill of the same name gets a new version, with the files of the zip only.",
	PeopleOnly:   true,
	LargeRequest: true,
	Prepare: func(_ context.Context, c *operation.Call, in ImportInput) error {
		data, err := base64.StdEncoding.DecodeString(in.Zip)
		if err != nil {
			return errZipInvalid(ZipProblem{Problem: "is not in base64"})
		}
		text, files, err := ReadZip(data)
		if err != nil {
			return private(in.Scope, err)
		}
		h, err := ParseText(text)
		if err != nil {
			return private(in.Scope, err)
		}
		c.Keep(imported{name: h.Name, text: text, files: files})
		return nil
	},
	Run: func(ctx context.Context, c *operation.Call, in ImportInput) (Skill, error) {
		z := c.Kept().(imported)
		s, err := apply(ctx, c, change{scope: in.Scope, name: z.name, text: z.text, files: z.files, replace: true, via: "import"})
		return s, private(in.Scope, err)
	},
}

// ExportSkill gives a skill as a zip.
var ExportSkill = &operation.Op[ExportInput, Export]{
	Name:        "export_skill",
	Title:       "Export a skill",
	Description: "Export a skill as a zip, in a folder of its name, as AI clients install it.",
	ReadOnly:    true,
	PeopleOnly:  true,
	Run: func(ctx context.Context, c *operation.Call, in ExportInput) (Export, error) {
		t, err := resolve(ctx, c, in.Skill)
		if err != nil {
			return Export{}, err
		}
		logSkill(c, t.scope(), t.name(), "")
		text, files := "", []File(nil)
		if t.builtin != nil {
			text, files = t.builtin.text, t.builtin.files
		} else {
			v, err := versionOf(ctx, c.Tx, t.row.current)
			if err != nil {
				return Export{}, err
			}
			if files, err = filesOf(ctx, c.Tx, t.row.current); err != nil {
				return Export{}, err
			}
			text = v.text
		}
		zip, err := WriteZip(t.name(), text, files)
		if err != nil {
			return Export{}, err
		}
		return Export{Name: t.name(), Zip: base64.StdEncoding.EncodeToString(zip)}, nil
	},
}

// RollbackSkill brings back the previous version of a skill.
var RollbackSkill = &operation.Op[RollbackInput, Skill]{
	Name:        "rollback_skill",
	Title:       "Roll back a skill",
	Description: "Bring back the previous version of a skill, which keeps the current one as its previous version.",
	PeopleOnly:  true,
	Run: func(ctx context.Context, c *operation.Call, in RollbackInput) (Skill, error) {
		t, err := resolve(ctx, c, in.Skill)
		if err != nil {
			return Skill{}, err
		}
		if err := writable(c, t); err != nil {
			return Skill{}, err
		}
		r := t.row
		logSkill(c, r.scope, r.name, "")
		if r.previous == 0 {
			return Skill{}, operation.Invalidf("this skill has one version: there is none to go back to")
		}
		if r.everyWindow {
			previous, err := versionOf(ctx, c.Tx, r.previous)
			if err != nil {
				return Skill{}, err
			}
			if len(previous.text) > MaxEveryWindowText {
				return Skill{}, errEveryWindowTooLong
			}
		}
		if _, err := c.Tx.ExecContext(ctx, "UPDATE skills SET current_version = previous_version, previous_version = current_version WHERE id = ?", r.id); err != nil {
			return Skill{}, err
		}
		if err := changed(ctx, c, r); err != nil {
			return Skill{}, err
		}
		r.current, r.previous = r.previous, r.current
		s, err := describe(ctx, c, t, false)
		if err == nil {
			logSkill(c, r.scope, r.name, fmt.Sprintf("back to version %d", s.Number))
		}
		return s, err
	},
}

// DeleteSkill deletes a skill, both of its versions.
var DeleteSkill = &operation.Op[DeleteInput, Deleted]{
	Name:        "delete_skill",
	Title:       "Delete a skill",
	Description: "Delete a skill for good, both of its versions.",
	Destructive: true,
	PeopleOnly:  true,
	Run: func(ctx context.Context, c *operation.Call, in DeleteInput) (Deleted, error) {
		t, err := resolve(ctx, c, in.Skill)
		if err != nil {
			return Deleted{}, err
		}
		if err := writable(c, t); err != nil {
			return Deleted{}, err
		}
		r := t.row
		logSkill(c, r.scope, r.name, "both versions")
		if _, err := c.Tx.ExecContext(ctx, "DELETE FROM skills WHERE id = ?", r.id); err != nil {
			return Deleted{}, err
		}
		if r.scope == ScopeCahut {
			if err := raise(ctx, c.Tx); err != nil {
				return Deleted{}, err
			}
		}
		return Deleted{Skill: r.name, Deleted: true}, nil
	},
}

// target is a skill that a call names: a skill of the Cahut or of the
// person, or the skill of the plugin.
type target struct {
	row     *row
	builtin *builtin
}

func (t target) scope() Scope {
	if t.builtin != nil {
		return ScopeBuiltin
	}
	return t.row.scope
}

// logSkill sets what the Logbook keeps of a call on the skill name of
// scope: its name, its scope and detail. It never names a personal skill:
// the administrators read the whole Logbook, and do not read the skills of
// others.
func logSkill(c *operation.Call, scope Scope, name, detail string) {
	switch {
	case scope == ScopePersonal:
		c.Log("a personal skill", detail)
	case detail == "":
		c.Log(name, string(scope))
	default:
		c.Log(name, string(scope)+", "+detail)
	}
}

func (t target) name() string {
	if t.builtin != nil {
		return BuiltinName
	}
	return t.row.name
}

// resolve returns the skill that ref names, as the person who acts may read
// it: its name, with its scope as a prefix if need be.
func resolve(ctx context.Context, c *operation.Call, ref string) (target, error) {
	scope, name, prefixed := strings.Cut(ref, ":")
	if !prefixed {
		scope, name = "", ref
	}
	switch Scope(scope) {
	case "", ScopeCahut, ScopePersonal, ScopeBuiltin:
	default:
		// The Logbook, which administrators read, keeps neither the name nor
		// the prefix that an AI tried.
		return target{}, operation.LogAs(operation.Invalidf("%q: give the name of a skill, alone or after cahut:, personal: or builtin:", ref),
			"invalid input: an unknown scope")
	}
	individual, err := settings.IsIndividual(ctx, c.Tx)
	if err != nil {
		return target{}, err
	}
	var found []target
	if b := builtinOf(c); b != nil && name == BuiltinName && (scope == "" || Scope(scope) == ScopeBuiltin) {
		found = append(found, target{builtin: b})
	}
	if !individual && (scope == "" || Scope(scope) == ScopeCahut) {
		r, err := lookup(ctx, c.Tx, 0, name)
		if err != nil {
			return target{}, err
		}
		if r != nil {
			found = append(found, target{row: r})
		}
	}
	if c.Actor.UserID != 0 && (scope == "" || Scope(scope) == ScopePersonal) {
		r, err := lookup(ctx, c.Tx, c.Actor.UserID, name)
		if err != nil {
			return target{}, err
		}
		if r != nil {
			found = append(found, target{row: r})
		}
	}
	switch len(found) {
	case 0:
		return target{}, errNotFound
	case 1:
		return found[0], nil
	}
	// The Logbook does not tell administrators that a person has a skill of
	// the name of one of the Cahut.
	return target{}, operation.LogAs(operation.Invalidf("%s names a skill of this Cahut and one of your own: give cahut:%s or personal:%s", name, name, name),
		"invalid input: a name in two scopes")
}

// writable returns an error unless the person who acts may change the skill
// t: the administrators of the Cahut its skills, each person their own.
func writable(c *operation.Call, t target) error {
	switch {
	case t.builtin != nil:
		return operation.Forbiddenf("the skill %s comes with Cahut, and nobody changes it", BuiltinName)
	case t.row.scope == ScopeCahut && !c.Admin:
		return errNotAdmin
	}
	return nil
}

// ownerFor returns the person whose skill a write in scope makes, zero for
// the Cahut, or an error when the person who acts may not write there.
func ownerFor(ctx context.Context, c *operation.Call, scope Scope) (int64, error) {
	switch scope {
	case ScopeCahut:
		individual, err := settings.IsIndividual(ctx, c.Tx)
		if err != nil {
			return 0, err
		}
		if individual {
			return 0, errScopeIndividual
		}
		if !c.Admin {
			return 0, errNotAdmin
		}
		return 0, nil
	case ScopePersonal:
		if c.Actor.UserID == 0 {
			return 0, operation.Invalidf("a personal skill belongs to a person: the command line writes the skills of the Cahut")
		}
		return c.Actor.UserID, nil
	case ScopeBuiltin:
		return 0, operation.Invalidf("the skill %s comes with Cahut, and nobody writes in its scope: use cahut or personal", BuiltinName)
	}
	return 0, operation.Invalidf("scope %q: use cahut for a skill of the team, or personal for one of your own", scope)
}

// change is what a write or an import makes of a skill.
type change struct {
	scope   Scope
	name    string
	text    string   // empty keeps the text of the current version
	files   []File   // to add or replace
	remove  []string // the paths of the files to remove
	replace bool     // files are every file of the skill, as an import gives them
	// everyWindow is the mark that the write gives the skill; nil keeps the
	// mark as it is.
	everyWindow *bool
	via         string
}

// apply writes ch as a new version of its skill, or as a new skill.
func apply(ctx context.Context, c *operation.Call, ch change) (Skill, error) {
	logSkill(c, ch.scope, ch.name, "")
	owner, err := ownerFor(ctx, c, ch.scope)
	if err != nil {
		return Skill{}, err
	}
	if ch.everyWindow != nil && *ch.everyWindow && ch.scope != ScopeCahut {
		return Skill{}, operation.Invalidf("only a skill of the Cahut is for every Window: a personal skill serves you alone")
	}
	if err := CheckName(ch.name); err != nil {
		return Skill{}, err
	}
	if ch.name == BuiltinName {
		return Skill{}, errNameReserved
	}
	r, err := lookup(ctx, c.Tx, owner, ch.name)
	if err != nil {
		return Skill{}, err
	}
	var current *version
	var files []File
	if r != nil {
		v, err := versionOf(ctx, c.Tx, r.current)
		if err != nil {
			return Skill{}, err
		}
		current = &v
		if !ch.replace {
			if files, err = filesOf(ctx, c.Tx, r.current); err != nil {
				return Skill{}, err
			}
		}
	}

	text := ch.text
	if text == "" {
		if current == nil {
			return Skill{}, operation.Invalidf("this is a new skill: give its text, a SKILL.md whose header gives its name and its description")
		}
		text = current.text
	}
	if len(text) > MaxText {
		return Skill{}, errTooLarge
	}
	h, err := ParseText(text)
	if err != nil {
		return Skill{}, err
	}
	if h.Name != ch.name {
		return Skill{}, errNameDiffers
	}
	// Cahut keeps the text that AIs read, its name and description in
	// double quotes: the same text written back makes no new version.
	if text, err = RenderText(text); err != nil {
		return Skill{}, err
	}
	if files, err = merge(files, ch.files, ch.remove); err != nil {
		return Skill{}, err
	}
	size := int64(len(text))
	for _, f := range files {
		size += int64(len(f.Content))
	}
	if len(files) > MaxFiles || size > MaxSize {
		return Skill{}, errTooLarge
	}

	// A skill for every Window stays short, since a refusal of the gate
	// carries its SKILL.md whole, and three skills at most take the mark.
	marked := r != nil && r.everyWindow
	mark := marked
	if ch.everyWindow != nil {
		mark = *ch.everyWindow
	}
	if mark && len(text) > MaxEveryWindowText {
		return Skill{}, errEveryWindowTooLong
	}
	if mark && !marked {
		n, err := countMarked(ctx, c.Tx)
		if err != nil {
			return Skill{}, err
		}
		if n >= MaxEveryWindow {
			return Skill{}, errEveryWindowTooMany
		}
	}

	digest := Digest(text, files)
	if current != nil && digest == current.digest {
		if mark != marked {
			return remark(ctx, c, r, mark)
		}
		logSkill(c, ch.scope, ch.name, "unchanged")
		return describe(ctx, c, target{row: r}, false)
	}
	number, added := 1, size
	if r == nil {
		n, err := count(ctx, c.Tx, owner)
		if err != nil {
			return Skill{}, err
		}
		if n >= MaxPerScope {
			return Skill{}, errTooMany
		}
		id, err := insert(ctx, c.Tx, ch.scope, owner, ch.name, c.Now)
		if err != nil {
			return Skill{}, err
		}
		r = &row{id: id, scope: ch.scope, owner: owner, name: ch.name}
	} else {
		number = current.number + 1
		if r.previous != 0 {
			// The previous version goes, and frees its bytes.
			previous, gone, err := numberOf(ctx, c.Tx, r.previous)
			if err != nil {
				return Skill{}, err
			}
			number, added = max(number, previous+1), size-gone
		}
	}
	if err := settings.CheckSkillDisk(ctx, c, owner, added); err != nil {
		return Skill{}, err
	}
	v := version{number: number, digest: digest, text: text, description: h.Description, size: size, writtenAt: c.Now, via: ch.via}
	vid, err := addVersion(ctx, c.Tx, r.id, v, c.Actor.UserID, files)
	if err != nil {
		return Skill{}, err
	}
	if r.current != 0 {
		r.previous = r.current
	}
	r.current = vid
	if mark != r.everyWindow {
		if err := setMark(ctx, c.Tx, r, mark); err != nil {
			return Skill{}, err
		}
	}
	if err := changed(ctx, c, r); err != nil {
		return Skill{}, err
	}
	logSkill(c, ch.scope, ch.name, fmt.Sprintf("version %d, %d files, %s", number, len(files), settings.Bytes(size))+markDetail(mark, marked, ", "))
	return describe(ctx, c, target{row: r}, false)
}

// remark gives the skill r of the Cahut the mark, or takes it off, without
// a new version. The mark orders the lines of the catalog, whose revision
// it raises.
func remark(ctx context.Context, c *operation.Call, r *row, mark bool) (Skill, error) {
	if err := setMark(ctx, c.Tx, r, mark); err != nil {
		return Skill{}, err
	}
	if err := raise(ctx, c.Tx); err != nil {
		return Skill{}, err
	}
	logSkill(c, r.scope, r.name, markDetail(mark, !mark, ""))
	return describe(ctx, c, target{row: r}, false)
}

// markDetail is what the Logbook says of a mark that a write changes from
// was to mark, after sep; nothing when it stays.
func markDetail(mark, was bool, sep string) string {
	switch {
	case mark && !was:
		return sep + "for every Window"
	case was && !mark:
		return sep + "no longer for every Window"
	}
	return ""
}

// changed indexes the current version of the skill r, and raises the
// revision of the catalog for a skill of the Cahut.
func changed(ctx context.Context, c *operation.Call, r *row) error {
	if err := reindex(ctx, c.Tx, r.id); err != nil {
		return err
	}
	if r.scope == ScopeCahut {
		return raise(ctx, c.Tx)
	}
	return nil
}

// decodeFiles checks the files that write_skill gives and decodes their
// contents.
func decodeFiles(in []FileInput) ([]File, error) {
	files := make([]File, 0, len(in))
	for _, f := range in {
		if err := CheckPath(f.Path); err != nil {
			return nil, err
		}
		var content []byte
		switch f.Encoding {
		case "", "utf-8", "utf8":
			content = []byte(f.Content)
		case "base64":
			var err error
			if content, err = base64.StdEncoding.DecodeString(f.Content); err != nil {
				return nil, operation.Invalidf("file %q: content is not valid base64", f.Path)
			}
		default:
			return nil, operation.Invalidf("file %q: encoding %q, use utf-8 or base64", f.Path, f.Encoding)
		}
		if len(content) > MaxFile {
			return nil, errTooLarge
		}
		files = append(files, File{Path: f.Path, Type: TypeOf(f.Path), Content: content})
	}
	return files, nil
}

// merge returns the files of a skill once added and remove apply to those
// it holds, by path. Paths match whatever their case, as on the disks where
// clients install skills.
func merge(held, added []File, remove []string) ([]File, error) {
	byPath := map[string]File{}
	for _, f := range held {
		byPath[strings.ToLower(f.Path)] = f
	}
	removed := map[string]bool{}
	for _, p := range remove {
		key := strings.ToLower(p)
		if _, ok := byPath[key]; !ok {
			return nil, operation.NotFoundf("the skill has no file %q to remove", p)
		}
		removed[key] = true
		delete(byPath, key)
	}
	seen := map[string]bool{}
	for _, f := range added {
		key := strings.ToLower(f.Path)
		switch {
		case removed[key]:
			return nil, errPathConflict(pathOf(f.Path))
		case seen[key]:
			return nil, errPathConflict(pathOf(f.Path))
		}
		seen[key] = true
		byPath[key] = f
	}
	files := make([]File, 0, len(byPath))
	for _, f := range byPath {
		files = append(files, f)
	}
	slices.SortFunc(files, func(a, b File) int { return cmp.Compare(a.Path, b.Path) })
	for _, f := range files {
		for dir := path.Dir(f.Path); dir != "."; dir = path.Dir(dir) {
			if _, ok := byPath[strings.ToLower(dir)]; ok {
				return nil, errPathConflict(pathOf(dir))
			}
		}
	}
	return files, nil
}

// describe returns the skill t, with its text if asked.
func describe(ctx context.Context, c *operation.Call, t target, withText bool) (Skill, error) {
	if b := t.builtin; b != nil {
		s := Skill{Skill: BuiltinName, Scope: ScopeBuiltin, Description: b.description, Version: b.digest, Number: 1, Files: b.infos()}
		if withText {
			s.Text = b.rendered
		}
		return s, nil
	}
	r := t.row
	v, err := versionOf(ctx, c.Tx, r.current)
	if err != nil {
		return Skill{}, err
	}
	s := Skill{Skill: r.name, Scope: r.scope, Description: v.description, Version: v.digest, Number: v.number,
		WrittenBy: v.writtenBy, WrittenAt: &v.writtenAt, Via: v.via, EveryWindow: r.everyWindow}
	if r.previous != 0 {
		if s.Previous, _, err = numberOf(ctx, c.Tx, r.previous); err != nil {
			return Skill{}, err
		}
	}
	if s.Files, err = infosOf(ctx, c.Tx, r.current); err != nil {
		return Skill{}, err
	}
	if withText {
		if s.Text, err = RenderText(v.text); err != nil {
			return Skill{}, err
		}
	}
	return s, nil
}

func get(ctx context.Context, c *operation.Call, in GetInput) (Skill, error) {
	t, err := resolve(ctx, c, in.Skill)
	if err != nil {
		return Skill{}, err
	}
	file := in.File
	if file != "" && t.scope() == ScopePersonal {
		file = "a file"
	}
	logSkill(c, t.scope(), t.name(), file)
	if in.File == "" {
		s, err := describe(ctx, c, t, true)
		s.Note = note
		if len(s.Files) > 0 {
			s.Note += fmt.Sprintf(" To put one of its files in a Window, give from_skill: %q in place of its content.", strings.TrimSpace(in.Skill)+"/"+s.Files[0].Path)
		}
		if s.EveryWindow {
			s.Note += fmt.Sprintf(" Every Window of this Cahut follows this skill: give follows: [%q] to publish_window and update_window.", s.Skill+"@"+s.Version)
		}
		return s, err
	}
	var f File
	var ok bool
	if t.builtin != nil {
		f, ok = t.builtin.file(in.File)
	} else if f, ok, err = fileOf(ctx, c.Tx, t.row.current, in.File); err != nil {
		return Skill{}, err
	}
	if !ok {
		return Skill{}, private(t.scope(), operation.NotFoundf("this skill has no file %q: get_skill without file lists its files", in.File))
	}
	s := Skill{Skill: t.name(), File: f.Path, Content: string(f.Content), Encoding: "utf-8"}
	if !textual(f) {
		s.Content, s.Encoding = base64.StdEncoding.EncodeToString(f.Content), "base64"
	}
	return s, nil
}

// textual reports whether the file f reads as text: its bytes are UTF-8,
// without a NUL byte, and its type is not that of an image, but SVG, of a
// sound, a video or a font. Its bytes decide, since the media types that
// each server knows differ: .yaml and .py have none on some.
func textual(f File) bool {
	if !utf8.Valid(f.Content) || bytes.IndexByte(f.Content, 0) >= 0 {
		return false
	}
	t, _, _ := strings.Cut(f.Type, ";")
	if t == "image/svg+xml" {
		return true
	}
	family, _, _ := strings.Cut(t, "/")
	switch family {
	case "image", "audio", "video", "font":
		return false
	}
	return true
}

func find(ctx context.Context, c *operation.Call, in FindInput) (Found, error) {
	individual, err := settings.IsIndividual(ctx, c.Tx)
	if err != nil {
		return Found{}, err
	}
	words := wordsOf(in.Query)
	query := `
		SELECT s.name, s.scope, s.every_window, v.number, v.digest, v.description, v.written_at
		FROM skills s JOIN skill_versions v ON v.id = s.current_version
		WHERE (s.owner_id IS NULL AND ?1 = 0 OR s.owner_id = ?2)`
	args := []any{individual, c.Actor.UserID}
	if len(words) > 0 {
		query += " AND s.id IN (SELECT rowid FROM skill_search WHERE skill_search MATCH ?3)"
		args = append(args, matchOf(words))
	}
	rows, err := c.Tx.QueryContext(ctx, query, args...)
	if err != nil {
		return Found{}, err
	}
	defer rows.Close()
	skills := []Summary{}
	for rows.Next() {
		var s Summary
		var at string
		if err := rows.Scan(&s.Skill, &s.Scope, &s.EveryWindow, &s.Number, &s.Version, &s.Description, &at); err != nil {
			return Found{}, err
		}
		written, err := store.ParseTime(at)
		if err != nil {
			return Found{}, err
		}
		s.WrittenAt = &written
		skills = append(skills, s)
	}
	if err := rows.Err(); err != nil {
		return Found{}, err
	}

	if b := builtinOf(c); b != nil && b.matches(words) {
		skills = append(skills, Summary{Skill: BuiltinName, Scope: ScopeBuiltin, Description: b.description, Version: b.digest, Number: 1})
	}
	// The skills of each scope go by their names, before any takes the
	// prefix of its scope, those for every Window first.
	rank := map[Scope]int{ScopeCahut: 0, ScopePersonal: 1, ScopeBuiltin: 2}
	slices.SortStableFunc(skills, func(a, b Summary) int {
		if d := cmp.Compare(rank[a.Scope], rank[b.Scope]); d != 0 {
			return d
		}
		if a.EveryWindow != b.EveryWindow {
			if a.EveryWindow {
				return -1
			}
			return 1
		}
		return cmp.Compare(a.Skill, b.Skill)
	})
	// A name that the person reads in two scopes takes its scope as a
	// prefix, though the other may not hold the words of the query.
	twice := make([]bool, len(skills))
	for i, s := range skills {
		other, err := lookupOther(ctx, c, individual, s)
		if err != nil {
			return Found{}, err
		}
		twice[i] = other
	}
	for i, s := range skills {
		if twice[i] {
			skills[i].Skill = string(s.Scope) + ":" + s.Skill
		}
	}
	// The Logbook keeps that a search took place, never its words: the
	// administrators read it all.
	if len(words) > 0 {
		c.Log("skills", fmt.Sprintf("a search, %d found", len(skills)))
	} else {
		c.Log("skills", fmt.Sprintf("%d found", len(skills)))
	}
	return Found{Skills: skills}, nil
}

// lookupOther reports whether the person reads a skill of the name of s in
// the other scope than that of s.
func lookupOther(ctx context.Context, c *operation.Call, individual bool, s Summary) (bool, error) {
	var owner int64
	switch {
	case s.Scope == ScopeCahut && c.Actor.UserID != 0:
		owner = c.Actor.UserID
	case s.Scope == ScopePersonal && !individual:
	default:
		return false, nil
	}
	r, err := lookup(ctx, c.Tx, owner, s.Skill)
	return r != nil, err
}

// MaxWords is how many words of a query count.
const MaxWords = 10

// wordsOf returns the words of a query: runs of letters and digits, as the
// index cuts them.
func wordsOf(s string) []string {
	ws := strings.FieldsFunc(s, notInWord)
	if len(ws) > MaxWords {
		ws = ws[:MaxWords]
	}
	return ws
}

// notInWord reports whether r cuts words, as the index cuts them.
func notInWord(r rune) bool {
	return !unicode.IsLetter(r) && !unicode.IsDigit(r) && !unicode.IsMark(r)
}

// matchOf returns the query of FTS5 that finds the skills holding each
// word, at the start of a word: "relance"* "factur"*.
func matchOf(words []string) string {
	quoted := make([]string, len(words))
	for i, w := range words {
		quoted[i] = `"` + w + `"*`
	}
	return strings.Join(quoted, " ")
}
