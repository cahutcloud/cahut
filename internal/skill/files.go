package skill

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"cahut.cloud/internal/operation"
)

// aPersonalFile is what FileOf gives the Logbook for a file of a personal
// skill, which the administrators do not read.
const aPersonalFile = "a file of a personal skill"

// FileOf returns, in the transaction of c, a file of the current version
// of a skill that the actor of c may read, named as the skill, with its
// scope as a prefix if need be, then the path of the file in the skill,
// such as charte-acme/assets/logo.svg. It returns too what the Logbook
// keeps of ref, which never names a file of a personal skill, for Copied.
// See docs/adr/0064-the-skills-of-a-cahut.md.
func FileOf(ctx context.Context, c *operation.Call, ref string) (File, string, error) {
	name, file, _ := strings.Cut(ref, "/")
	if name == "" || file == "" {
		return File{}, "", operation.Invalidf("from_skill %q: give the name of a skill, then the path of one of its files, such as charte-acme/assets/logo.svg", ref)
	}
	t, err := resolve(ctx, c, name)
	if err != nil {
		return File{}, "", err
	}
	var f File
	var found bool
	if t.builtin != nil {
		f, found = t.builtin.file(file)
	} else if f, found, err = fileOf(ctx, c.Tx, t.row.current, file); err != nil {
		return File{}, "", err
	}
	if !found {
		return File{}, "", private(t.scope(), operation.NotFoundf("the skill %s has no file %q: get_skill gives the list of its files", t.name(), file))
	}
	if t.scope() == ScopePersonal {
		return f, aPersonalFile, nil
	}
	return f, t.name() + "/" + f.Path, nil
}

// Copied returns what the Logbook keeps of the files that a publication
// copied from skills, as FileOf gives them: the files of the skills of the
// Cahut and of the plugin by name, and how many of personal skills.
func Copied(logged []string) string {
	var named []string
	personal := 0
	for _, l := range logged {
		switch {
		case l == aPersonalFile:
			personal++
		case !slices.Contains(named, l):
			named = append(named, l)
		}
	}
	switch {
	case personal == 1:
		named = append(named, aPersonalFile)
	case personal > 1:
		named = append(named, fmt.Sprintf("%d files of personal skills", personal))
	}
	if len(named) == 0 {
		return ""
	}
	return "copied " + andList(named)
}
