//go:build !console

package console

import "io/fs"

// app returns no console: this binary was built without the console tag,
// which needs the console built by Node.js first.
func app() fs.FS {
	return nil
}
