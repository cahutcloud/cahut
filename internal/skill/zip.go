package skill

import (
	"archive/zip"
	"bytes"
	"cmp"
	"errors"
	"io"
	"mime"
	"path"
	"slices"
	"strings"
	"time"
)

// ReadZip reads a skill from a zip, with its SKILL.md at the root or in a
// single folder, as AI clients export them. It checks every path and every
// limit as it reads, and never holds more than MaxSize.
func ReadZip(data []byte) (text string, files []File, err error) {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", nil, errZipInvalid(ZipProblem{Problem: "cannot be read as a zip"})
	}
	var entries []*zip.File
	for _, f := range r.File {
		if strings.HasSuffix(f.Name, "/") && f.Mode().IsDir() || leftByFinder(f.Name) {
			continue
		}
		if !f.Mode().IsRegular() {
			return "", nil, errZipInvalid(ZipProblem{Problem: "holds a link or another file that is not a plain file"})
		}
		if p := f.Name; p == "" || path.IsAbs(p) || strings.Contains(p, `\`) || path.Clean(p) != p || p == ".." || strings.HasPrefix(p, "../") {
			return "", nil, errZipInvalid(ZipProblem{Problem: "holds a path that is not a clean relative path"})
		}
		entries = append(entries, f)
	}
	root, err := rootOf(entries)
	if err != nil {
		return "", nil, err
	}
	if len(entries) > MaxFiles+1 {
		return "", nil, errTooLarge
	}

	var size int64
	seen := map[string]bool{}
	found := false
	for _, f := range entries {
		rel := strings.TrimPrefix(f.Name, root)
		if rel != "SKILL.md" {
			if err := CheckPath(rel); err != nil {
				return "", nil, err
			}
		}
		key := strings.ToLower(rel)
		if seen[key] {
			return "", nil, errZipInvalid(ZipProblem{Problem: "holds " + rel + " twice"})
		}
		seen[key] = true
		limit := int64(MaxFile)
		if rel == "SKILL.md" {
			limit = MaxText
		}
		content, err := readAtMost(f, min(limit, MaxSize-size))
		if err != nil {
			return "", nil, err
		}
		size += int64(len(content))
		if rel == "SKILL.md" {
			text, found = string(content), true
			continue
		}
		files = append(files, File{Path: rel, Type: TypeOf(rel), Content: content})
	}
	if !found {
		return "", nil, errZipInvalid(ZipProblem{Problem: "has no SKILL.md at its root or in a single folder"})
	}
	for _, f := range files {
		for dir := path.Dir(f.Path); dir != "."; dir = path.Dir(dir) {
			if seen[strings.ToLower(dir)] {
				return "", nil, errZipInvalid(ZipProblem{Problem: "holds " + dir + " both as a file and as a folder"})
			}
		}
	}
	if _, err := ParseText(text); err != nil {
		return "", nil, err
	}
	slices.SortFunc(files, func(a, b File) int { return cmp.Compare(a.Path, b.Path) })
	return text, files, nil
}

// leftByFinder reports whether the zip entry name is what the Finder of
// macOS adds to a zip, which a skill does not hold.
func leftByFinder(name string) bool {
	return strings.HasPrefix(name, "__MACOSX/") || path.Base(name) == ".DS_Store"
}

// rootOf returns the folder of the zip where its SKILL.md is: "" at the
// root, or the single folder that holds every entry.
func rootOf(entries []*zip.File) (string, error) {
	var top string
	for i, f := range entries {
		if f.Name == "SKILL.md" {
			return "", nil
		}
		first, _, _ := strings.Cut(f.Name, "/")
		if i > 0 && first != top {
			top = ""
			break
		}
		top = first
	}
	if top != "" {
		for _, f := range entries {
			if f.Name == top+"/SKILL.md" {
				return top + "/", nil
			}
		}
	}
	return "", errZipInvalid(ZipProblem{Problem: "has no SKILL.md at its root or in a single folder"})
}

// readAtMost reads the entry f, refusing it when it holds more than limit
// bytes, whatever its header says: a zip bomb stops at the limit.
func readAtMost(f *zip.File, limit int64) ([]byte, error) {
	if f.UncompressedSize64 > uint64(limit) {
		return nil, errTooLarge
	}
	rc, err := f.Open()
	if err != nil {
		return nil, errZipInvalid(ZipProblem{Problem: "holds an entry that cannot be read"})
	}
	defer rc.Close()
	content, err := io.ReadAll(io.LimitReader(rc, limit+1))
	if err != nil {
		if errors.Is(err, zip.ErrChecksum) || errors.Is(err, zip.ErrFormat) || errors.Is(err, io.ErrUnexpectedEOF) {
			return nil, errZipInvalid(ZipProblem{Problem: "holds an entry that cannot be read"})
		}
		return nil, err
	}
	if int64(len(content)) > limit {
		return nil, errTooLarge
	}
	return content, nil
}

// WriteZip writes a skill as a zip, in a folder of its name, its SKILL.md
// rendered by RenderText.
func WriteZip(name, text string, files []File) ([]byte, error) {
	rendered, err := RenderText(text)
	if err != nil {
		return nil, err
	}
	var b bytes.Buffer
	w := zip.NewWriter(&b)
	write := func(p string, content []byte) error {
		f, err := w.CreateHeader(&zip.FileHeader{Name: name + "/" + p, Method: zip.Deflate, Modified: time.Unix(0, 0).UTC()})
		if err != nil {
			return err
		}
		_, err = f.Write(content)
		return err
	}
	if err := write("SKILL.md", []byte(rendered)); err != nil {
		return nil, err
	}
	for _, f := range slices.SortedFunc(slices.Values(files), func(a, b File) int { return cmp.Compare(a.Path, b.Path) }) {
		if err := write(f.Path, f.Content); err != nil {
			return nil, err
		}
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// TypeOf returns the media type of a file of a skill, from its extension.
func TypeOf(p string) string {
	switch strings.ToLower(path.Ext(p)) {
	case ".md", ".markdown":
		return "text/markdown; charset=utf-8"
	case ".txt":
		return "text/plain; charset=utf-8"
	}
	if t := mime.TypeByExtension(path.Ext(p)); t != "" {
		return t
	}
	return "application/octet-stream"
}
