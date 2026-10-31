// Package address checks the names that make the address of a Window,
// <domain>/<room>/<window>/: the domain of its Cahut, then the names of its
// Room and of the Window. See docs/adr/0009-rooms-and-windows-as-paths.md.
package address

import (
	"regexp"
	"strings"
)

var validName = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$`)

// ValidName reports whether s can name a Room or a Window: 1 to 63 lowercase
// ASCII letters, digits and hyphens, starting and ending with a letter or a
// digit.
func ValidName(s string) bool {
	return validName.MatchString(s)
}

// reserved are the names that no Room can take, because clients expect them
// at the root of a Cahut's domain or Cahut may need them there: server is
// the console of the server. See docs/adr/0058-a-console-for-the-server.md.
var reserved = map[string]bool{"mcp": true, "api": true, "admin": true, "console": true, "help": true, "oauth": true, "server": true}

// Reserved reports whether no Room can take the name s.
func Reserved(s string) bool {
	return reserved[s]
}

var validDomain = regexp.MustCompile(`^([a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?\.)*[a-z]([a-z0-9-]{0,61}[a-z0-9])?$`)

// ValidDomain reports whether s can be the domain of a Cahut: a host name in
// lowercase, such as cahut.example.com or localhost. Its last label starts
// with a letter, which rules out IP addresses.
func ValidDomain(s string) bool {
	return len(s) <= 253 && validDomain.MatchString(s)
}

// folds turns the accented letters of Latin scripts into ASCII for names.
var folds = map[rune]string{
	'à': "a", 'á': "a", 'â': "a", 'ã': "a", 'ä': "a", 'å': "a", 'æ': "ae",
	'ç': "c", 'è': "e", 'é': "e", 'ê': "e", 'ë': "e",
	'ì': "i", 'í': "i", 'î': "i", 'ï': "i", 'ñ': "n",
	'ò': "o", 'ó': "o", 'ô': "o", 'õ': "o", 'ö': "o", 'ø': "o", 'œ': "oe",
	'ù': "u", 'ú': "u", 'û': "u", 'ü': "u", 'ý': "y", 'ÿ': "y", 'ß': "ss",
}

// Slug turns a title into a name: at most 40 lowercase ASCII letters, digits
// and hyphens. A title with none of them gives fallback.
func Slug(title, fallback string) string {
	var b strings.Builder
	hyphen := false
	for _, r := range strings.ToLower(title) {
		s, ok := folds[r]
		switch {
		case ok:
		case r >= 'a' && r <= 'z' || r >= '0' && r <= '9':
			s = string(r)
		default:
			if b.Len() > 0 && !hyphen {
				b.WriteByte('-')
				hyphen = true
			}
			continue
		}
		if b.Len()+len(s) > 40 {
			break
		}
		b.WriteString(s)
		hyphen = false
	}
	if s := strings.Trim(b.String(), "-"); s != "" {
		return s
	}
	return fallback
}
