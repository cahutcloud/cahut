package mail_test

import (
	"reflect"
	"testing"

	cmail "cahut.cloud/internal/mail"
)

func TestParseNote(t *testing.T) {
	type run = cmail.NoteRun
	b := func(text string) run { return run{Text: text, Bold: true} }
	i := func(text string) run { return run{Text: text, Italic: true} }
	bi := func(text string) run { return run{Text: text, Bold: true, Italic: true} }
	w := func(text string) run { return run{Text: text} }
	line := func(runs ...run) [][]run { return [][]run{runs} }

	for _, c := range []struct {
		line string
		want []run
	}{
		{"Have a look", []run{w("Have a look")}},
		{"**Friday** at *noon*", []run{b("Friday"), w(" at "), i("noon")}},
		{"***both***", []run{bi("both")}},
		{"**bold *and italic* inside**", []run{b("bold "), bi("and italic"), b(" inside")}},
		{"**a***b*", []run{b("a"), i("b")}},
		{"*a***b**", []run{i("a"), b("b")}},
		// A star next to a space, or that nothing closes, stays a star.
		{"5 * 3 * 2", []run{w("5 * 3 * 2")}},
		{"**not closed", []run{w("**not closed")}},
		{"*a **b*", []run{i("a **b")}},
		{"** spaced **", []run{w("** spaced **")}},
		{"****", []run{w("****")}},
		// A backslash keeps a mark as it is.
		{`\*not italic\*`, []run{w("*not italic*")}},
		{`**a\***`, []run{b("a*")}},
		{`C:\path and \\`, []run{w(`C:\path and \`)}},
		// Markup of another kind is words.
		{"<b>x</b> [link](https://example.com) _x_", []run{w("<b>x</b> [link](https://example.com) _x_")}},
	} {
		got := cmail.ParseNote(c.line)
		if want := []cmail.NoteBlock{{Lines: [][]run{c.want}}}; !reflect.DeepEqual(got, want) {
			t.Errorf("ParseNote(%q) = %+v, want %+v", c.line, got, want)
		}
	}

	note := "Here is the plan:\n- the **deck**\n* the notes\n\n3. first\n4) second\nAfter\nthe list\n\n\n\\- not a list\n1\\. nor this"
	want := []cmail.NoteBlock{
		{Lines: line(w("Here is the plan:"))},
		{List: cmail.Bullets, Lines: [][]run{{w("the "), b("deck")}, {w("the notes")}}},
		{List: cmail.Numbers, Start: 3, Lines: [][]run{{w("first")}, {w("second")}}},
		{Lines: [][]run{{w("After")}, {w("the list")}}},
		{Lines: [][]run{{w("- not a list")}, {w("1. nor this")}}},
	}
	if got := cmail.ParseNote(note); !reflect.DeepEqual(got, want) {
		t.Errorf("ParseNote(%q) =\n%+v\nwant\n%+v", note, got, want)
	}
	if got := cmail.ParseNote(""); len(got) != 0 {
		t.Errorf("ParseNote of nothing = %+v", got)
	}
}
