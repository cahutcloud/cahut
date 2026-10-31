// Package windata runs the SQL of the database of each Window, in a worker
// process of its own for each database: the binary started again, which
// holds no key, opens nothing but the folder of its database, and has
// limits of memory and time. The service talks to its workers through a
// Pool, one line of JSON per message. See
// docs/adr/0056-a-database-for-each-window.md.
package windata

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
)

// Kinds of request a worker answers.
const (
	opSelect  = "select"  // a SELECT of a reader, over what is readable
	opQuery   = "query"   // a SELECT of an AI or a person, over everything
	opCall    = "call"    // a query of the declaration
	opExec    = "exec"    // statements of an AI or a person
	opCheck   = "check"   // migrations, queries and what is readable, in a transaction rolled back
	opMigrate = "migrate" // the migrations that have not run yet
	opCopy    = "copy"    // a copy of the database, beside it
	opRestore = "restore" // back to a copy
	opSchema  = "schema"  // the tables and their columns, rows and size
)

// request is a message from the service to a worker.
type request struct {
	ID         uint64                     `json:"id"`
	Op         string                     `json:"op"`
	SQL        string                     `json:"sql,omitempty"`
	Params     map[string]json.RawMessage `json:"params,omitempty"`
	Readable   map[string][]string        `json:"readable,omitempty"`
	Migrations []string                   `json:"migrations,omitempty"`
	Queries    map[string]string          `json:"queries,omitempty"`
	Write      bool                       `json:"write,omitempty"` // the query writes: it needs the write connection
	Dry        bool                       `json:"dry,omitempty"`   // run, then roll back
	Copy       string                     `json:"copy,omitempty"`  // the name of a copy
	Keep       int                        `json:"keep,omitempty"`  // the copies to keep
	Limits     Limits                     `json:"limits"`
}

// response is a message from a worker to the service, for the request of
// the same id.
type response struct {
	ID         uint64              `json:"id"`
	Error      *Error              `json:"error,omitempty"`
	Columns    []string            `json:"columns,omitempty"`
	Rows       [][]json.RawMessage `json:"rows,omitempty"`
	Cut        bool                `json:"cut,omitempty"`
	Changes    int64               `json:"changes,omitempty"`
	LastID     int64               `json:"last_id,omitempty"`
	Kinds      map[string]string   `json:"kinds,omitempty"`
	Ran        int                 `json:"ran,omitempty"`
	Removes    []string            `json:"removes,omitempty"`
	Schema     *Schema             `json:"schema,omitempty"`
	Copy       *Copy               `json:"copy,omitempty"`
	Statements int                 `json:"statements,omitempty"`
}

// Limits bound one request.
type Limits struct {
	Millis  int64 `json:"millis"`   // of work; the statement is interrupted beyond
	Rows    int   `json:"rows"`     // answered at most
	Bytes   int64 `json:"bytes"`    // answered at most, as JSON
	MaxSize int64 `json:"max_size"` // of the database, for a request that writes
}

// Limits of the calls of readers and of AIs.
var (
	ReaderLimits = Limits{Millis: 1000, Rows: 1000, Bytes: 1 << 20, MaxSize: MaxSize}
	AILimits     = Limits{Millis: 10000, Rows: 1000, Bytes: 1 << 20, MaxSize: MaxSize}
)

// MaxSize bounds the database of every Window.
const MaxSize = 100 << 20

// Codes of the errors of a request.
const (
	CodeInvalid     = "invalid"     // the SQL or its parameters are at fault: the message says why
	CodeConstraint  = "constraint"  // a constraint of the database refused the row
	CodeFull        = "full"        // the database has reached its size
	CodeTimeout     = "timeout"     // the statement ran out of time
	CodeBusy        = "busy"        // too many calls at once
	CodeUnavailable = "unavailable" // the worker stopped
	CodeInternal    = "internal"
)

// Error is an error of a request, with a code that says whose fault it is.
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *Error) Error() string { return e.Message }

// CodeOf returns the code of err, or CodeInternal.
func CodeOf(err error) string {
	if e, ok := errors.AsType[*Error](err); ok {
		return e.Code
	}
	return CodeInternal
}

func invalidf(format string, args ...any) *Error {
	return &Error{Code: CodeInvalid, Message: fmt.Sprintf(format, args...)}
}

// Rows is the result of a SELECT: its columns, its rows as JSON values,
// and whether the limits cut it.
type Rows struct {
	Columns []string            `json:"columns"`
	Rows    [][]json.RawMessage `json:"rows"`
	Cut     bool                `json:"cut,omitempty"`
}

// Result is the result of a query of a declaration: rows for a query that
// reads, the rows changed and the id of the last row added otherwise.
type Result struct {
	Rows
	Changes int64 `json:"changes"`
	LastID  int64 `json:"id,omitempty"`
}

// Change is what statements of an AI or a person did, or would do when
// they ran dry: their count, the rows they changed, and the tables,
// columns, indexes, views and triggers that they remove or rename.
type Change struct {
	Statements int      `json:"statements"`
	Changes    int64    `json:"changes"`
	Removes    []string `json:"removes,omitempty"`
}

// Schema is what a database holds.
type Schema struct {
	Tables     []Table  `json:"tables"`
	Views      []string `json:"views,omitempty"`
	Indexes    []string `json:"indexes,omitempty"`
	Triggers   []string `json:"triggers,omitempty"`
	Migrations int      `json:"migrations"`
	Size       int64    `json:"size"`
}

// Table is a table of a database.
type Table struct {
	Name    string   `json:"name"`
	Columns []Column `json:"columns"`
	Rows    int64    `json:"rows"`
}

// Column is a column of a table.
type Column struct {
	Name       string `json:"name"`
	Type       string `json:"type,omitempty"`
	NotNull    bool   `json:"not_null,omitempty"`
	PrimaryKey bool   `json:"primary_key,omitempty"`
}

// Copy is a copy of a database, beside it.
type Copy struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
}

// Check is what a declaration asks of the database of its Window.
type Check struct {
	Migrations []string
	Queries    map[string]string
	Readable   map[string][]string
}

// Checked is what checking a declaration found: what each query does,
// and how many migrations have not run yet.
type Checked struct {
	Kinds map[string]string
	New   int
}

// What a query does, from what SQLite reports when it prepares it.
const (
	Reads   = "reads"
	Adds    = "adds"
	Changes = "changes"
	Deletes = "deletes"
)

// decodeValue reads a parameter: null, a boolean as 1 or 0, a number, a
// string, {"blob": "<base64>"}, or another object or array as its JSON
// text.
func decodeValue(raw json.RawMessage) (any, error) {
	s := strings.TrimSpace(string(raw))
	switch {
	case s == "" || s == "null":
		return nil, nil
	case s == "true":
		return int64(1), nil
	case s == "false":
		return int64(0), nil
	case s[0] == '"':
		var v string
		err := json.Unmarshal(raw, &v)
		return v, err
	case s[0] == '{':
		var obj map[string]json.RawMessage
		if err := json.Unmarshal(raw, &obj); err != nil {
			return nil, err
		}
		var blob string
		if len(obj) == 1 && json.Unmarshal(obj["blob"], &blob) == nil {
			return base64.StdEncoding.DecodeString(blob)
		}
		return s, nil
	case s[0] == '[':
		return s, nil
	}
	var n json.Number
	if err := json.Unmarshal(raw, &n); err != nil {
		return nil, err
	}
	if i, err := n.Int64(); err == nil {
		return i, nil
	}
	return n.Float64()
}

// encodeValue writes a value that SQLite returned as JSON: a blob as
// {"blob": "<base64>"}, and an infinite number as null.
func encodeValue(v any) json.RawMessage {
	switch x := v.(type) {
	case nil:
		return json.RawMessage("null")
	case []byte:
		b, _ := json.Marshal(map[string]string{"blob": base64.StdEncoding.EncodeToString(x)})
		return b
	case float64:
		if math.IsInf(x, 0) || math.IsNaN(x) {
			return json.RawMessage("null")
		}
	case time.Time:
		// The driver reads the text of a DATE, DATETIME or TIMESTAMP column
		// as a time: write it back as SQLite writes CURRENT_TIMESTAMP.
		layout := "2006-01-02 15:04:05.999999999"
		if x.Location() != time.UTC {
			layout += "-07:00"
		}
		v = x.Format(layout)
	}
	b, err := json.Marshal(v)
	if err != nil {
		return json.RawMessage("null")
	}
	return b
}
