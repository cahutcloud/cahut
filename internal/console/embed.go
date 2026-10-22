//go:build console

package console

import (
	"embed"
	"io/fs"
)

// dist is the console as Vite built it. make build builds it first, then
// builds the binary with the console tag.
//
//go:embed all:ui/dist
var dist embed.FS

func app() fs.FS {
	files, err := fs.Sub(dist, "ui/dist")
	if err != nil {
		panic(err) // a constant path, so a programming error
	}
	return files
}
