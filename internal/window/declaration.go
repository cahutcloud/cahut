package window

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"cahut.cloud/internal/operation"
)

// Declaration is what the AI declares that the pages of a version do. Cahut
// makes the Content Security Policy of their files from it: a Window that
// declares no call can make none. See
// docs/adr/0019-window-versions-access-and-declaration.md.
type Declaration struct {
	Scripts bool     `json:"scripts" jsonschema:"whether the pages run scripts, inline or in files"`
	Origins []string `json:"origins,omitempty" jsonschema:"every origin the pages load anything from or call, such as https://cdn.jsdelivr.net or https://fonts.gstatic.com; https://*.example.com for the subdomains of example.com. The files of the Window need none, nor do the blob: and data: URLs that the pages make"`
	Forms   []string `json:"forms,omitempty" jsonschema:"the origins that the forms of the pages send to, if any"`
}

// MaxOrigins bounds each list of a declaration.
const MaxOrigins = 20

// validOrigin matches an origin of a declaration: https or wss, a host
// that may start with *. for its subdomains, and an optional port.
var validOrigin = regexp.MustCompile(`^(https|wss)://(\*\.)?([a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?\.)*[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?(:[0-9]{1,5})?$`)

// check returns the declaration with its origins in lowercase and without
// a final slash, or an error wrapping operation.ErrInvalid.
func (d *Declaration) check() (*Declaration, error) {
	if d == nil {
		return nil, operation.Invalidf("declaration is missing: say whether the pages run scripts, " +
			"and list the origins they load from or call, and those their forms send to")
	}
	out := &Declaration{Scripts: d.Scripts}
	var err error
	if out.Origins, err = origins("origins", d.Origins); err != nil {
		return nil, err
	}
	if out.Forms, err = origins("forms", d.Forms); err != nil {
		return nil, err
	}
	return out, nil
}

func origins(field string, list []string) ([]string, error) {
	if len(list) > MaxOrigins {
		return nil, operation.Invalidf("declaration.%s lists %d origins, more than the %d allowed", field, len(list), MaxOrigins)
	}
	var out []string
	seen := map[string]bool{}
	for _, o := range list {
		o = strings.TrimSuffix(strings.ToLower(strings.TrimSpace(o)), "/")
		if !validOrigin.MatchString(o) {
			return nil, operation.Invalidf("declaration.%s: %q is not an origin: give https:// or wss://, a host and an optional port, "+
				"such as https://cdn.jsdelivr.net, without a path", field, o)
		}
		if !seen[o] {
			seen[o] = true
			out = append(out, o)
		}
	}
	return out, nil
}

// String sums the declaration up for the Logbook.
func (d *Declaration) String() string {
	if d == nil {
		return "not declared"
	}
	s := "no scripts"
	if d.Scripts {
		s = "scripts"
	}
	list := func(name string, origins []string) string {
		if len(origins) == 0 {
			return "; no " + name
		}
		return fmt.Sprintf("; %s %s", name, strings.Join(origins, " "))
	}
	return s + list("origins", d.Origins) + list("forms", d.Forms)
}

// encode returns the declaration as the store keeps it.
func (d *Declaration) encode() (string, error) {
	b, err := json.Marshal(d)
	return string(b), err
}

// decodeDeclaration reads a declaration as the store keeps it: nil when the
// version was published before declarations.
func decodeDeclaration(s *string) (*Declaration, error) {
	if s == nil {
		return nil, nil
	}
	d := &Declaration{}
	return d, json.Unmarshal([]byte(*s), d)
}

// Sandbox returns the flags of the sandbox of the pages that d declares:
// without allow-scripts when they run none.
func (d *Declaration) Sandbox() string {
	flags := "allow-forms allow-modals allow-popups allow-popups-to-escape-sandbox allow-downloads"
	if d == nil || d.Scripts {
		flags = "allow-scripts " + flags
	}
	return flags
}

// Policy returns the Content Security Policy of the files of a Window that
// d declares, served at the origin self, such as https://cahut.example.com.
// A Window published before declarations keeps the sandbox alone. Pages may
// fetch the blob: and data: URLs they make: these hold what the page already
// has, and the browser resolves a blob: URL only for the opaque origin that
// made it. See docs/adr/0045-windows-fetch-the-blob-and-data-urls-they-make.md.
func (d *Declaration) Policy(self string) string {
	sandbox := "sandbox " + d.Sandbox()
	if d == nil {
		return sandbox
	}
	from := func(sources ...string) string {
		return strings.Join(append(append([]string{self}, sources...), d.Origins...), " ")
	}
	directives := []string{sandbox, "default-src 'none'"}
	if d.Scripts {
		directives = append(directives, "script-src "+from("'unsafe-inline'", "'unsafe-eval'"))
	}
	forms := "'none'"
	if len(d.Forms) > 0 {
		forms = strings.Join(d.Forms, " ")
	}
	directives = append(directives,
		"style-src "+from("'unsafe-inline'"),
		"img-src "+from("data:", "blob:"),
		"font-src "+from("data:"),
		"media-src "+from("data:", "blob:"),
		"connect-src "+from("data:", "blob:"),
		"frame-src "+from(),
		"worker-src "+self+" blob:",
		"form-action "+forms,
		"base-uri 'none'",
	)
	return strings.Join(directives, "; ")
}
