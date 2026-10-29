package mail

import (
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

// A note, such as that of an invitation, is plain text with a few marks of
// Markdown: **bold**, *italics*, and lists whose lines start with -, * or +,
// or with a number and a period. A blank line ends a paragraph or a list, and
// a backslash before a mark keeps it as it is. Nothing else is markup: no
// link, heading, image or HTML. See
// docs/adr/0046-styled-notes-in-invitations.md.

// NoteBlock is a paragraph or a list of a note.
type NoteBlock struct {
	List  string      `json:"list,omitempty" jsonschema:"bullets or numbers for a list, empty for a paragraph"`
	Start int         `json:"start,omitempty" jsonschema:"the number of the first item of a numbered list"`
	Lines [][]NoteRun `json:"lines" jsonschema:"the lines of a paragraph, or the items of a list"`
}

// NoteRun is words of a line in the same style.
type NoteRun struct {
	Text   string `json:"text"`
	Bold   bool   `json:"bold,omitempty"`
	Italic bool   `json:"italic,omitempty"`
}

// The kinds of lists.
const (
	Bullets = "bullets"
	Numbers = "numbers"
)

var (
	bulletLine = regexp.MustCompile(`^ {0,3}[-*+][ \t]+(.*)$`)
	numberLine = regexp.MustCompile(`^ {0,3}(\d{1,9})[.)][ \t]+(.*)$`)
)

// ParseNote returns the paragraphs and lists of a note.
func ParseNote(text string) []NoteBlock {
	var blocks []NoteBlock
	var cur *NoteBlock
	for _, line := range strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n") {
		kind, start, words := Bullets, 0, ""
		if m := bulletLine.FindStringSubmatch(line); m != nil {
			words = m[1]
		} else if m := numberLine.FindStringSubmatch(line); m != nil {
			kind, words = Numbers, m[2]
			start, _ = strconv.Atoi(m[1])
		} else if strings.TrimSpace(line) == "" {
			cur = nil
			continue
		} else {
			kind, words = "", line
		}
		if cur == nil || cur.List != kind {
			blocks = append(blocks, NoteBlock{List: kind, Start: start})
			cur = &blocks[len(blocks)-1]
		}
		cur.Lines = append(cur.Lines, inline(strings.TrimSpace(words)))
	}
	return blocks
}

// noteLines returns a note as the plain text of a message shows it: its
// paragraphs and lists, a blank line between them, and the marks of its
// styles, but no backslash.
func noteLines(text string) []string {
	var lines []string
	for i, b := range ParseNote(text) {
		if i > 0 {
			lines = append(lines, "")
		}
		for j, l := range b.Lines {
			var s strings.Builder
			switch b.List {
			case Bullets:
				s.WriteString("- ")
			case Numbers:
				s.WriteString(strconv.Itoa(b.Start+j) + ". ")
			}
			for _, r := range l {
				marks := strings.Repeat("*", 2*boolInt(r.Bold)+boolInt(r.Italic))
				s.WriteString(marks + r.Text + marks)
			}
			lines = append(lines, s.String())
		}
	}
	return lines
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// The styles of a run, as indexes.
const (
	bold = iota
	italic
)

// mark is a piece of a line: words, or a run of stars that opens or closes
// bold and italics.
type mark struct {
	text          string
	stars         int
	canOpen       bool
	canClose      bool
	opens, closes [2]bool
}

// inline returns the runs of a line. Two stars open or close bold and one
// star italics, three both, left to right. A star opens before words and
// closes after them, never next to a space, and a star that nothing closes
// stays a star.
func inline(line string) []NoteRun {
	var marks []mark
	var words strings.Builder
	flush := func() {
		if words.Len() > 0 {
			marks = append(marks, mark{text: words.String()})
			words.Reset()
		}
	}
	rs := []rune(line)
	for i := 0; i < len(rs); i++ {
		switch r := rs[i]; {
		case r == '\\' && i+1 < len(rs) && isPunct(rs[i+1]):
			words.WriteRune(rs[i+1])
			i++
		case r == '*':
			j := i
			for j < len(rs) && rs[j] == '*' {
				j++
			}
			if j-i > 3 {
				words.WriteString(string(rs[i:j]))
			} else {
				flush()
				marks = append(marks, mark{
					stars:    j - i,
					canOpen:  j < len(rs) && !unicode.IsSpace(rs[j]),
					canClose: i > 0 && !unicode.IsSpace(rs[i-1]),
				})
			}
			i = j - 1
		default:
			words.WriteRune(r)
		}
	}
	flush()

	var opener [2]int // the index of the mark that opened each style, plus one
	for i := range marks {
		m := &marks[i]
		styles := stylesOf(m.stars)
		for _, s := range styles {
			if opener[s] != 0 && m.canClose {
				m.closes[s], opener[s] = true, 0
			}
		}
		for _, s := range styles {
			if !m.closes[s] && opener[s] == 0 && m.canOpen {
				m.opens[s], opener[s] = true, i+1
			}
		}
	}
	for s, o := range opener {
		if o != 0 {
			marks[o-1].opens[s] = false
		}
	}

	var runs []NoteRun
	var on [2]bool
	add := func(text string) {
		if text == "" {
			return
		}
		if n := len(runs); n > 0 && runs[n-1].Bold == on[bold] && runs[n-1].Italic == on[italic] {
			runs[n-1].Text += text
			return
		}
		runs = append(runs, NoteRun{Text: text, Bold: on[bold], Italic: on[italic]})
	}
	for _, m := range marks {
		if m.stars == 0 {
			add(m.text)
			continue
		}
		left := m.stars
		for s := range on {
			if m.closes[s] {
				on[s] = false
				left -= width(s)
			}
		}
		for s := range on {
			if m.opens[s] {
				left -= width(s)
			}
		}
		add(strings.Repeat("*", left))
		for s := range on {
			if m.opens[s] {
				on[s] = true
			}
		}
	}
	return runs
}

// stylesOf returns the styles that a run of stars opens or closes.
func stylesOf(stars int) []int {
	switch stars {
	case 1:
		return []int{italic}
	case 2:
		return []int{bold}
	}
	return []int{bold, italic}
}

// width returns the stars of the mark of a style.
func width(style int) int {
	if style == bold {
		return 2
	}
	return 1
}

// isPunct reports whether a backslash before r keeps it as it is: ASCII
// punctuation, as in Markdown.
func isPunct(r rune) bool {
	return strings.ContainsRune("!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~", r)
}
