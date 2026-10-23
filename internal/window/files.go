package window

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io/fs"
	"mime"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"unicode"
	"unicode/utf8"

	"cahut.cloud/internal/operation"
)

// MinFreeSpace is the space that a publication leaves free on the disk of
// the server, which all its Cahuts and their databases share: below it, a
// Cahut refuses to publish, rather than let a full disk stop them all.
// Quotas for each Cahut come after the V0.
const MinFreeSpace = 1 << 30

// freeSpace returns the bytes free for Cahut on the disk that holds dir,
// or on that of its nearest folder that exists. Tests replace it.
var freeSpace = func(dir string) (uint64, error) {
	for {
		var st syscall.Statfs_t
		err := syscall.Statfs(dir, &st)
		if err == nil {
			// The size of a block is an int64 on Linux and a uint32 on
			// macOS: its text reads the same on both.
			block, err := strconv.ParseUint(fmt.Sprint(st.Bsize), 10, 64)
			return st.Bavail * block, err
		}
		if parent := filepath.Dir(dir); parent != dir && errors.Is(err, fs.ErrNotExist) {
			dir = parent
			continue
		}
		return 0, err
	}
}

// checkSpace refuses a publication of size bytes when the disk is short of
// space.
func (s *Service) checkSpace(size int64) error {
	free, err := freeSpace(s.dir)
	if err != nil {
		return err
	}
	if free < MinFreeSpace+uint64(size) { //nolint:gosec // size is positive
		return operation.Invalidf("the disk of the server is almost full, with %d MiB free: nothing was published. "+
			"Ask its administrator to make room, or purge Windows from the trash", free>>20)
	}
	return nil
}

// Limits of a publication.
const (
	MaxFiles = 500
	MaxSize  = 32 << 20 // bytes, all files decoded
)

type decoded struct {
	path string
	data []byte
}

// decode checks the files of a publication and decodes their contents.
func decode(files []File) ([]decoded, int64, error) {
	if len(files) == 0 {
		return nil, 0, operation.Invalidf("files is empty: publish at least index.html")
	}
	if len(files) > MaxFiles {
		return nil, 0, operation.Invalidf("%d files, more than the %d allowed", len(files), MaxFiles)
	}
	out := make([]decoded, 0, len(files))
	seen := map[string]bool{} // lowercase paths, so that no two files collide on a case-insensitive disk
	var size int64
	for _, f := range files {
		if err := checkPath(f.Path); err != nil {
			return nil, 0, err
		}
		key := strings.ToLower(f.Path)
		if seen[key] {
			return nil, 0, operation.Invalidf("file %q appears twice", f.Path)
		}
		seen[key] = true

		var data []byte
		switch f.Encoding {
		case "", "utf-8", "utf8":
			data = []byte(f.Content)
		case "base64":
			var err error
			if data, err = base64.StdEncoding.DecodeString(f.Content); err != nil {
				return nil, 0, operation.Invalidf("file %q: content is not valid base64", f.Path)
			}
		default:
			return nil, 0, operation.Invalidf("file %q: encoding %q, use utf-8 or base64", f.Path, f.Encoding)
		}
		size += int64(len(data))
		if size > MaxSize {
			return nil, 0, operation.Invalidf("files total more than %d MiB", MaxSize>>20)
		}
		out = append(out, decoded{path: f.Path, data: data})
	}
	if !seen["index.html"] {
		return nil, 0, operation.Invalidf("index.html is missing: it is the page that opens")
	}
	for _, f := range out {
		for dir := path.Dir(f.path); dir != "."; dir = path.Dir(dir) {
			if seen[strings.ToLower(dir)] {
				return nil, 0, operation.Invalidf("%q is both a file and the folder of %q", dir, f.path)
			}
		}
	}
	return out, size, nil
}

// checkPath accepts clean, relative paths with / as separator, and no
// hidden or special parts.
func checkPath(p string) error {
	switch {
	case p == "":
		return operation.Invalidf("a file has an empty path")
	case len(p) > 255:
		return operation.Invalidf("path %.40q... is longer than 255 bytes", p)
	case !utf8.ValidString(p) || strings.ContainsFunc(p, unicode.IsControl) || strings.Contains(p, `\`):
		return operation.Invalidf("path %q: use printable characters and / as separator", p)
	case path.IsAbs(p) || path.Clean(p) != p:
		return operation.Invalidf("path %q: use a clean relative path, such as img/logo.svg", p)
	}
	for _, part := range strings.Split(p, "/") {
		if strings.HasPrefix(part, ".") {
			return operation.Invalidf("path %q: no part may start with a dot", p)
		}
	}
	return nil
}

// writeFiles writes files into dir, which must not exist yet. The writes go
// through an os.Root, so that no path can leave dir.
func writeFiles(dir string, files []decoded) error {
	if err := os.MkdirAll(filepath.Dir(dir), 0o750); err != nil {
		return err
	}
	if err := os.Mkdir(dir, 0o750); err != nil {
		return err
	}
	root, err := os.OpenRoot(dir)
	if err != nil {
		return err
	}
	defer root.Close()
	for _, f := range files {
		if d := path.Dir(f.path); d != "." {
			if err := root.MkdirAll(d, 0o750); err != nil {
				return err
			}
		}
		if err := root.WriteFile(f.path, f.data, 0o640); err != nil {
			return err
		}
	}
	return nil
}

// listFiles lists the files under dir, with their contents if asked. Text
// comes back as UTF-8, anything else in base64.
func listFiles(dir string, content bool) ([]FileInfo, error) {
	root, err := os.OpenRoot(dir)
	if err != nil {
		return nil, err
	}
	defer root.Close()
	fsys := root.FS()

	files := []FileInfo{}
	err = fs.WalkDir(fsys, ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil || !d.Type().IsRegular() {
			return err
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		f := FileInfo{Path: p, Size: info.Size(), Type: mediaType(p)}
		if content {
			data, err := fs.ReadFile(fsys, p)
			if err != nil {
				return err
			}
			if utf8.Valid(data) && isText(f.Type) {
				f.Content, f.Encoding = string(data), "utf-8"
			} else {
				f.Content, f.Encoding = base64.StdEncoding.EncodeToString(data), "base64"
			}
		}
		files = append(files, f)
		return nil
	})
	return files, err
}

func mediaType(p string) string {
	if t := mime.TypeByExtension(path.Ext(p)); t != "" {
		return t
	}
	return "application/octet-stream"
}

func isText(mediaType string) bool {
	t, _, _ := strings.Cut(mediaType, ";")
	switch t {
	case "application/json", "application/javascript", "application/xml", "image/svg+xml":
		return true
	}
	return strings.HasPrefix(t, "text/")
}
