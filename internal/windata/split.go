package windata

import (
	"slices"
	"strings"
)

// statement is one statement of SQL, without its final semicolon: its
// text, the names of the parameters it binds, and its first word, in
// capitals.
type statement struct {
	text      string
	params    []string
	first     string
	returning bool // it returns rows, with RETURNING
}

// split cuts sql into its statements, as SQLite reads them: a semicolon
// inside a string, a quoted name or a comment ends nothing, nor one inside
// the body of a trigger, which ends at "; END;". Parameters are named, as
// :name, @name or $name; a ? is refused, so that every value has a name.
func split(sql string) ([]statement, *Error) {
	var (
		out     []statement
		start   int
		params  []string
		returns bool
		words   []string // the first three words of the statement
		trigger bool     // the statement creates a trigger
		last    string   // the last token: a word in capitals, ";" or ""
		before  string   // the token before it
	)
	end := func(i int) {
		text := strings.TrimSpace(sql[start:i])
		if text != "" {
			first := ""
			if len(words) > 0 {
				first = words[0]
			}
			out = append(out, statement{text: text, params: params, first: first, returning: returns})
		}
		start, params, words, trigger, returns = i+1, nil, nil, false, false
	}
	token := func(t string) {
		before, last = last, t
	}
	for i := 0; i < len(sql); i++ {
		c := sql[i]
		switch {
		case c == '-' && i+1 < len(sql) && sql[i+1] == '-':
			for i < len(sql) && sql[i] != '\n' {
				i++
			}
		case c == '/' && i+1 < len(sql) && sql[i+1] == '*':
			j := strings.Index(sql[i+2:], "*/")
			if j < 0 {
				return nil, invalidf("a comment /* is not closed by */")
			}
			i += j + 3
		case c == '\'' || c == '"' || c == '`':
			j, ok := closing(sql, i, c)
			if !ok {
				return nil, invalidf("a quote %c is not closed", c)
			}
			i = j
			token("")
		case c == '[':
			j := strings.IndexByte(sql[i:], ']')
			if j < 0 {
				return nil, invalidf("a name in [ is not closed by ]")
			}
			i += j
			token("")
		case c == '?':
			return nil, invalidf("name the parameters, such as :name, rather than ?")
		case (c == ':' || c == '@' || c == '$') && i+1 < len(sql) && isWord(sql[i+1]):
			j := i + 1
			for j < len(sql) && isWord(sql[j]) {
				j++
			}
			if name := sql[i+1 : j]; !slices.Contains(params, name) {
				params = append(params, name)
			}
			i = j - 1
			token("")
		case isWord(c) && !isDigit(c):
			j := i
			for j < len(sql) && (isWord(sql[j]) || sql[j] == '$') {
				j++
			}
			w := strings.ToUpper(sql[i:j])
			returns = returns || w == "RETURNING"
			if len(words) < 3 {
				words = append(words, w)
				trigger = trigger || isTrigger(words)
			}
			token(w)
			i = j - 1
		case c == ';':
			if trigger && (last != "END" || before != ";") {
				token(";")
				continue
			}
			end(i)
			last, before = "", ""
		case c == ' ' || c == '\t' || c == '\n' || c == '\r':
		default:
			token("")
		}
	}
	end(len(sql))
	return out, nil
}

// closing returns the index of the quote that closes the one at i, which a
// doubled quote does not.
func closing(sql string, i int, q byte) (int, bool) {
	for j := i + 1; j < len(sql); j++ {
		if sql[j] == q {
			if j+1 < len(sql) && sql[j+1] == q {
				j++
				continue
			}
			return j, true
		}
	}
	return 0, false
}

// isTrigger reports whether the first words of a statement create a
// trigger: CREATE TRIGGER, or CREATE TEMP TRIGGER.
func isTrigger(words []string) bool {
	if len(words) < 2 || words[0] != "CREATE" {
		return false
	}
	if words[1] == "TRIGGER" {
		return true
	}
	return len(words) == 3 && (words[1] == "TEMP" || words[1] == "TEMPORARY") && words[2] == "TRIGGER"
}

func isWord(c byte) bool {
	return c == '_' || isDigit(c) || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= 0x80
}

func isDigit(c byte) bool { return c >= '0' && c <= '9' }
