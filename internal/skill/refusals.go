package skill

import "cahut.cloud/internal/operation"

// Named is the value of a refusal that names a skill.
type Named struct {
	Name operation.Quoted `json:"name"`
}

// HeaderProblem is the value of the refusal of a header: what is wrong
// with it, in English, which the console leaves out.
type HeaderProblem struct {
	Problem string `json:"problem"`
}

// AtPath is the value of a refusal that names a file of a skill.
type AtPath struct {
	Path operation.Quoted `json:"path"`
}

// The refusals that people meet when they write or import a skill. See
// docs/adr/0037-refusals-with-a-code-for-the-console.md.
var (
	errNameInvalid = operation.NewRefusalOf[Named]("skill_name_invalid", operation.ErrInvalid,
		"{name} is not the name of a skill: use 1 to 64 lowercase letters, digits and hyphens, such as charte-acme, "+
			"without a hyphen at either end or two in a row")
	errHeaderInvalid = operation.NewRefusalOf[HeaderProblem]("skill_header_invalid", operation.ErrInvalid,
		"the header of SKILL.md {problem}: a SKILL.md opens with a line ---, then name and description on lines of their own, "+
			"then a line ---, as Agent Skills have it")
	errHasScripts = operation.NewRefusalOf[AtPath]("skill_has_scripts", operation.ErrInvalid,
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
