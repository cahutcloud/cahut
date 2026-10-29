// Package plugins embeds the files of the plugins of Cahut for AI clients,
// which live in this directory so that a marketplace can install them as
// they are: the binary serves the same skill as the plugin. See
// docs/adr/0022-house-style-of-each-cahut.md and
// docs/adr/0049-no-house-style.md.
package plugins

import (
	"archive/zip"
	"bytes"
	"embed"
	"io/fs"
	"sync"
	"time"
)

// The skill of the plugin for Claude.
//
//go:embed cahut/skills/cahut
var skill embed.FS

// SkillName is the name of the skill, and of the folder that holds it.
const SkillName = "cahut"

var (
	zipOnce  sync.Once
	zipBytes []byte
	zipErr   error
)

// SkillZip returns the skill for Claude as a zip of its folder, as the
// Claude apps take an uploaded skill.
func SkillZip() ([]byte, error) {
	zipOnce.Do(func() { zipBytes, zipErr = makeZip() })
	return zipBytes, zipErr
}

func makeZip() ([]byte, error) {
	root, err := fs.Sub(skill, "cahut/skills")
	if err != nil {
		return nil, err
	}
	var b bytes.Buffer
	w := zip.NewWriter(&b)
	err = fs.WalkDir(root, SkillName, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		content, err := fs.ReadFile(root, path)
		if err != nil {
			return err
		}
		// A fixed time, so that the same skill makes the same zip.
		f, err := w.CreateHeader(&zip.FileHeader{Name: path, Method: zip.Deflate, Modified: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)})
		if err != nil {
			return err
		}
		_, err = f.Write(content)
		return err
	})
	if err != nil {
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}
