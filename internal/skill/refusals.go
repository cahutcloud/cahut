package skill

import "cahut.cloud/internal/operation"

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
