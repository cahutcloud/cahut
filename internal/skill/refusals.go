package skill

import (
	"errors"
	"unicode/utf8"

	"cahut.cloud/internal/operation"
)

// HeaderProblem is the value of the refusal of a header: what is wrong
// with it, in English, which the console leaves out.
type HeaderProblem struct {
	Problem string `json:"problem"`
}

// ScriptPath is the value of the refusal of a file under scripts/.
type ScriptPath struct {
	Path operation.Quoted `json:"path"`
}

// The refusals that people meet when they write or import a skill. See
// docs/adr/0037-refusals-with-a-code-for-the-console.md.
var (
	errNameInvalid = operation.NewRefusal("skill_name_invalid", operation.ErrInvalid,
		"the name of the skill is not valid: use 1 to 64 lowercase letters, digits and hyphens, such as charte-acme, "+
			"without a hyphen at either end or two in a row")
	errHeaderInvalid = operation.NewRefusalOf[HeaderProblem]("skill_header_invalid", operation.ErrInvalid,
		"the header of SKILL.md {problem}: a SKILL.md opens with a line ---, then name and description on lines of their own, "+
			"then a line ---, as Agent Skills have it")
	errHasScripts = operation.NewRefusalOf[ScriptPath]("skill_has_scripts", operation.ErrInvalid,
		"{path} is under scripts/, and a skill of a Cahut holds no scripts: some AI clients run them on the computer of each person, "+
			"and a script would be code that someone changes from afar. Leave out scripts/, and say in SKILL.md what the AI should do instead")
)

// ZipProblem is the value of the refusal of a zip: what is wrong with it,
// in English, which the console leaves out.
type ZipProblem struct {
	Problem string `json:"problem"`
}

var (
	errZipInvalid = operation.NewRefusalOf[ZipProblem]("skill_zip_invalid", operation.ErrInvalid,
		"the zip {problem}: give the zip of a skill, with its SKILL.md at its root or in a single folder, as AI clients export them")
	errTooLarge = operation.NewRefusal("skill_too_large", operation.ErrInvalid,
		"this skill is too large: a SKILL.md holds 64 KiB at most, each of its files 2 MiB, and a skill 100 files and 4 MiB in all")
)

var (
	errNameReserved = operation.NewRefusal("skill_name_reserved", operation.ErrInvalid,
		"the name cahut is kept for the skill of the plugin for Claude, which every AI of this Cahut reads: give your skill another name")
	errTooMany = operation.NewRefusal("skill_too_many", operation.ErrForbidden,
		"this scope holds 50 skills, the most it may: delete one before you add another")
	errScopeIndividual = operation.NewRefusal("skill_scope_individual", operation.ErrForbidden,
		"this Cahut is individual, and has no skills of the Cahut: its people share the Cahut, not their work. "+
			"Write the skill among your own, with the scope personal")
)

// FilePath is the value of a refusal about the path of a file of a skill.
type FilePath struct {
	Path operation.Quoted `json:"path"`
}

// The refusals that people meet on the page of the console that writes a
// skill: the header of a SKILL.md that names another skill, and the paths
// of its files.
var (
	errNameDiffers = operation.NewRefusal("skill_name_differs", operation.ErrInvalid,
		"the name in the header of SKILL.md is not the name of this skill: keep its name in the header, or write a new skill under the other one")
	errPathInvalid = operation.NewRefusalOf[FilePath]("skill_path_invalid", operation.ErrInvalid,
		"{path} is not the path of a file of a skill: use a relative path of printable characters with / between folders, such as references/deck.md, "+
			"of 255 bytes at most, with no part that starts with a dot, and not SKILL.md, which is the text of the skill")
	errPathConflict = operation.NewRefusalOf[FilePath]("skill_path_conflict", operation.ErrInvalid,
		"{path} clashes with another file of the skill: give each path once, whatever its case, never as both a file and a folder, "+
			"and never as both written and removed")
)

// pathOf is the path p as a refusal names it: a long path is cut.
func pathOf(p string) FilePath {
	if len(p) > 64 {
		cut := 60
		for cut > 0 && !utf8.RuneStart(p[cut]) {
			cut--
		}
		p = p[:cut] + "..."
	}
	return FilePath{Path: operation.Quoted(p)}
}

// The refusals of the mark for every Window, which administrators meet on
// the page of a skill of the Cahut.
var (
	errEveryWindowTooMany = operation.NewRefusal("every_window_too_many", operation.ErrForbidden,
		"three skills of this Cahut are for every Window, the most it may: take the mark off one of them before you give it to another")
	errEveryWindowTooLong = operation.NewRefusal("every_window_too_long", operation.ErrInvalid,
		"a skill for every Window holds a SKILL.md of 16 KiB at most, since AIs receive it whole when they publish a Window: "+
			"shorten it, and move the details into its files, which it names")
)

var (
	errNotFound = operation.NewRefusal("skill_not_found", operation.ErrNotFound,
		"there is no skill of this name that you may read: find_skills lists them")
	errNotAdmin = operation.NewRefusal("skill_admins_only", operation.ErrForbidden,
		"only the administrators of this Cahut write its skills: write the skill among your own, with the scope personal")
)

// private returns err, the refusal of a write or an import in scope, which
// the Logbook keeps without its message for a personal skill: the message
// may name a file of the skill, and the administrators read the whole
// Logbook. The person reads it whole.
func private(scope Scope, err error) error {
	if scope != ScopePersonal || !operation.Public(err) {
		return err
	}
	logged := "refused"
	for _, kind := range []error{operation.ErrInvalid, operation.ErrNotFound, operation.ErrForbidden, operation.ErrTooMany} {
		if errors.Is(err, kind) {
			logged = kind.Error()
			break
		}
	}
	if r, ok := errors.AsType[*operation.Refusal](err); ok {
		logged += ": " + r.Code
	}
	return operation.LogAs(err, logged+", about a personal skill")
}
