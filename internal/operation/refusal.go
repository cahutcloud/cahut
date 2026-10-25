package operation

import (
	"cmp"
	"fmt"
	"maps"
	"reflect"
	"regexp"
	"slices"
	"strings"
	"sync"
)

// A Refusal is an error meant for the caller that people meet in the
// console, which says it in the language of its person. Code names it, and
// never changes; Values fill the sentence of the console. Its message, in
// English, is what AIs, the command line and the Logbook read, as for any
// other error meant for the caller. The API gives all three. See
// docs/adr/0037-refusals-with-a-code-for-the-console.md.
type Refusal struct {
	Code   string
	Values any // a struct that JSON encodes; nil when the message holds no value
	kind   error
	text   string
}

func (r *Refusal) Error() string {
	if r.kind == nil {
		return r.text
	}
	return r.kind.Error() + ": " + r.text
}

// Unwrap returns the kind of the refusal, such as ErrInvalid.
func (r *Refusal) Unwrap() error {
	return r.kind
}

// RefusalInfo describes a refusal to the generators of the console's types
// and of the reference.
type RefusalInfo struct {
	Code    string
	Kind    error        // ErrInvalid, ErrForbidden, ErrNotFound or ErrTooMany; nil for one of an adapter
	Message string       // the English message, with {name} where each value goes
	Values  reflect.Type // the struct of the values; nil when the message holds none
}

var (
	refusalsMu sync.Mutex
	refusals   = map[string]RefusalInfo{}
)

var (
	codePattern = regexp.MustCompile(`^[a-z][a-z0-9]*(_[a-z0-9]+)*$`)
	placeholder = regexp.MustCompile(`\{([a-z_]+)\}`)
)

// declare records the refusal of info, and panics when its code is not
// snake_case or is taken: a code names one refusal only.
func declare(info RefusalInfo) {
	if !codePattern.MatchString(info.Code) {
		panic(fmt.Sprintf("operation: refusal code %q is not snake_case", info.Code))
	}
	refusalsMu.Lock()
	defer refusalsMu.Unlock()
	if _, ok := refusals[info.Code]; ok {
		panic(fmt.Sprintf("operation: refusal %q declared twice", info.Code))
	}
	refusals[info.Code] = info
}

// NewRefusal declares the refusal code, of the kind ErrInvalid,
// ErrForbidden, ErrNotFound or ErrTooMany, whose English message holds no
// value, and returns it. A nil kind is for a refusal of an adapter, such as
// a session that ended, which no operation returns. Declare each refusal
// once, in a variable of its package.
func NewRefusal(code string, kind error, message string) error {
	declare(RefusalInfo{Code: code, Kind: kind, Message: message})
	return &Refusal{Code: code, kind: kind, text: message}
}

// NewRefusalOf declares the refusal code, as NewRefusal does, whose English
// message holds the values of V: a struct whose fields JSON encodes, each
// named in the message as {name}. It returns the function that makes the
// refusal from its values. It panics when the message and V do not name
// the same values.
func NewRefusalOf[V any](code string, kind error, message string) func(V) error {
	t := reflect.TypeFor[V]()
	fields := valueFields(t)
	names := slices.Sorted(maps.Keys(fields))
	var used []string
	for _, m := range placeholder.FindAllStringSubmatch(message, -1) {
		used = append(used, m[1])
	}
	slices.Sort(used)
	if !slices.Equal(names, slices.Compact(used)) {
		panic(fmt.Sprintf("operation: refusal %q names %v in its message, and its values are %v", code, used, names))
	}
	declare(RefusalInfo{Code: code, Kind: kind, Message: message, Values: t})
	return func(v V) error {
		rv := reflect.ValueOf(v)
		text := placeholder.ReplaceAllStringFunc(message, func(m string) string {
			return fmt.Sprint(rv.Field(fields[m[1:len(m)-1]]).Interface())
		})
		return &Refusal{Code: code, Values: v, kind: kind, text: text}
	}
}

// valueFields returns the index of each field of the struct t that JSON
// encodes, by its JSON name.
func valueFields(t reflect.Type) map[string]int {
	if t.Kind() != reflect.Struct {
		panic(fmt.Sprintf("operation: the values of a refusal are a struct, not %s", t))
	}
	fields := map[string]int{}
	for f := range t.Fields() {
		name, _, _ := strings.Cut(f.Tag.Get("json"), ",")
		if f.IsExported() && name != "" && name != "-" {
			fields[name] = f.Index[0]
		}
	}
	return fields
}

// Refusals returns the refusals that the packages of the program declare,
// sorted by code.
func Refusals() []RefusalInfo {
	refusalsMu.Lock()
	defer refusalsMu.Unlock()
	out := make([]RefusalInfo, 0, len(refusals))
	for _, r := range refusals {
		out = append(out, r)
	}
	slices.SortFunc(out, func(a, b RefusalInfo) int { return cmp.Compare(a.Code, b.Code) })
	return out
}
