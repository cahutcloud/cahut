package skill

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/settings"
)

// marked is a skill of the Cahut for every Window, in its current version.
type marked struct {
	name, digest, text string
	files              []FileInfo
}

// ref is the skill as follows gives it: its name and the digest of its
// version.
func (m marked) ref() string { return m.name + "@" + m.digest }

// Gate checks, in the transaction of c, that a call of an AI that makes a
// Window follows the current version of every skill of the Cahut marked for
// every Window. Otherwise it logs the skills that are missing and returns an
// error that wraps operation.ErrInvalid and carries the SKILL.md of each,
// its files, and what to give in follows. It lets every call of a person
// through. See docs/adr/0064-the-skills-of-a-cahut.md.
//
// The gate keeps nothing between two calls: the digest comes only in the
// answers of get_skill and in its refusal, so an AI that gives it received
// the skill, in this conversation.
func Gate(ctx context.Context, c *operation.Call, follows []string) error {
	if c.Actor.Kind != operation.AI {
		return nil
	}
	individual, err := settings.IsIndividual(ctx, c.Tx)
	if err != nil || individual {
		return err
	}
	skills, err := markedOf(ctx, c.Tx)
	if err != nil {
		return err
	}
	given := map[string]bool{}
	for _, f := range follows {
		given[bareRef(f)] = true
	}
	var missing []marked
	refs := make([]string, len(skills))
	for i, m := range skills {
		refs[i] = strconv.Quote(m.ref())
		if !given[m.ref()] {
			missing = append(missing, m)
		}
	}
	if len(missing) == 0 {
		return nil
	}

	names := make([]string, len(missing))
	logged := make([]string, len(missing))
	for i, m := range missing {
		names[i], logged[i] = m.name, m.ref()
	}
	var b strings.Builder
	if len(missing) == 1 {
		fmt.Fprintf(&b, "every Window of this Cahut follows the skill %s, whose current version this call does not give in follows: nothing was published. "+
			"Read the skill below, make the Window again as it says, then call again with follows: [%s].", names[0], strings.Join(refs, ", "))
	} else {
		fmt.Fprintf(&b, "every Window of this Cahut follows the skills %s, whose current versions this call does not give in follows: nothing was published. "+
			"Read them below, make the Window again as they say, then call again with follows: [%s].", andList(names), strings.Join(refs, ", "))
	}
	for _, m := range missing {
		fmt.Fprintf(&b, "\n\nThe skill %s, version %s:\n\n%s", m.name, m.digest, strings.TrimRight(m.text, "\n"))
		if len(m.files) > 0 {
			fmt.Fprintf(&b, "\n\nIts files, which get_skill reads with file, and which a file of the Window copies with from_skill in place of content, such as %s/%s:",
				m.name, m.files[0].Path)
			for _, f := range m.files {
				fmt.Fprintf(&b, "\n- %s (%s, %s)", f.Path, settings.Bytes(f.Size), f.Type)
			}
		}
	}
	return operation.LogAs(operation.Invalidf("%s", b.String()), operation.ErrInvalid.Error()+": gate: "+strings.Join(logged, ", "))
}

// Followed returns what the Logbook keeps of follows: the skills of the
// Cahut that it names, and how many others. It never names a personal
// skill, since the administrators read the whole Logbook.
func Followed(ctx context.Context, c *operation.Call, follows []string) (string, error) {
	if len(follows) == 0 {
		return "", nil
	}
	rows, err := c.Tx.QueryContext(ctx, "SELECT name FROM skills WHERE owner_id IS NULL")
	if err != nil {
		return "", err
	}
	defer rows.Close()
	ofTheCahut := map[string]bool{}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return "", err
		}
		ofTheCahut[name] = true
	}
	if err := rows.Err(); err != nil {
		return "", err
	}
	var named []string
	others := 0
	for _, f := range follows {
		ref := bareRef(f)
		if name, _, _ := strings.Cut(ref, "@"); ofTheCahut[name] {
			named = append(named, ref)
		} else {
			others++
		}
	}
	switch {
	case len(named) == 0:
		return fmt.Sprintf("follows %d %s", others, plural(others, "skill", "skills")), nil
	case others == 0:
		return "follows " + strings.Join(named, ", "), nil
	}
	return fmt.Sprintf("follows %s and %d %s", strings.Join(named, ", "), others, plural(others, "other", "others")), nil
}

// bareRef is an entry of follows without its spaces, nor the prefix of the
// scope of the Cahut, which find_skills gives when the person has a skill
// of the same name.
func bareRef(f string) string {
	return strings.TrimPrefix(strings.TrimSpace(f), string(ScopeCahut)+":")
}

// markedOf returns the skills of the Cahut for every Window, by name, in
// their current versions.
func markedOf(ctx context.Context, tx *sql.Tx) ([]marked, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT s.name, v.id, v.digest, v.text FROM skills s JOIN skill_versions v ON v.id = s.current_version
		WHERE s.owner_id IS NULL AND s.every_window = 1 ORDER BY s.name`)
	if err != nil {
		return nil, err
	}
	var skills []marked
	var versions []int64
	for rows.Next() {
		var m marked
		var id int64
		if err := rows.Scan(&m.name, &id, &m.digest, &m.text); err != nil {
			rows.Close()
			return nil, err
		}
		skills, versions = append(skills, m), append(versions, id)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for i := range skills {
		if skills[i].text, err = RenderText(skills[i].text); err != nil {
			return nil, err
		}
		if skills[i].files, err = infosOf(ctx, tx, versions[i]); err != nil {
			return nil, err
		}
	}
	return skills, nil
}

// andList joins names as English does: a, b and c.
func andList(names []string) string {
	if len(names) < 2 {
		return strings.Join(names, "")
	}
	return strings.Join(names[:len(names)-1], ", ") + " and " + names[len(names)-1]
}

func plural(n int, one, many string) string {
	if n == 1 {
		return one
	}
	return many
}
