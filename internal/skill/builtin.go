package skill

import (
	"io/fs"
	"strings"

	"cahut.cloud/internal/operation"
)

type builtinKey struct{}

// builtin is the skill of the plugin for Claude, which every Cahut gives
// from its binary, read only. See docs/adr/0022-house-style-of-each-cahut.md.
type builtin struct {
	text, rendered, description, digest string
	files                               []File
	words                               []string // the words of its name, its description and its text, folded
}

// loadBuiltin reads the skill of the plugin from fsys, its SKILL.md at the
// root.
func loadBuiltin(fsys fs.FS) (*builtin, error) {
	text, err := fs.ReadFile(fsys, "SKILL.md")
	if err != nil {
		return nil, err
	}
	b := &builtin{text: string(text)}
	h, err := ParseText(b.text)
	if err != nil {
		return nil, err
	}
	if h.Name != BuiltinName {
		return nil, operation.Invalidf("its name is %s", h.Name)
	}
	if b.rendered, err = RenderText(b.text); err != nil {
		return nil, err
	}
	b.description = h.Description
	err = fs.WalkDir(fsys, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil || !d.Type().IsRegular() || p == "SKILL.md" {
			return err
		}
		content, err := fs.ReadFile(fsys, p)
		if err != nil {
			return err
		}
		b.files = append(b.files, File{Path: p, Type: TypeOf(p), Content: content})
		return nil
	})
	if err != nil {
		return nil, err
	}
	b.digest = Digest(b.text, b.files)
	b.words = strings.FieldsFunc(fold(h.Name+" "+h.Description+" "+b.text), notInWord)
	return b, nil
}

// builtinOf returns the skill of the plugin that the registry of c
// provides, or nil.
func builtinOf(c *operation.Call) *builtin {
	b, _ := c.Value(builtinKey{}).(*builtin)
	return b
}

func (b *builtin) infos() []FileInfo {
	var infos []FileInfo
	for _, f := range b.files {
		infos = append(infos, FileInfo{Path: f.Path, Size: int64(len(f.Content)), Type: f.Type})
	}
	return infos
}

func (b *builtin) file(p string) (File, bool) {
	for _, f := range b.files {
		if f.Path == p {
			return f, true
		}
	}
	return File{}, false
}

// matches reports whether the skill holds each of words at the start of a
// word, whatever their case and accents, as the index finds the others.
func (b *builtin) matches(words []string) bool {
	for _, w := range words {
		w = fold(w)
		found := false
		for _, have := range b.words {
			if strings.HasPrefix(have, w) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// fold writes s in lowercase, without the accents of the Latin letters, as
// the index of the search reads words.
var folded = strings.NewReplacer(
	"à", "a", "á", "a", "â", "a", "ã", "a", "ä", "a", "å", "a", "æ", "ae",
	"ç", "c", "è", "e", "é", "e", "ê", "e", "ë", "e",
	"ì", "i", "í", "i", "î", "i", "ï", "i", "ñ", "n",
	"ò", "o", "ó", "o", "ô", "o", "õ", "o", "ö", "o", "œ", "oe", "ø", "o",
	"ù", "u", "ú", "u", "û", "u", "ü", "u", "ý", "y", "ÿ", "y", "ß", "ss",
)

func fold(s string) string {
	return folded.Replace(strings.ToLower(s))
}
