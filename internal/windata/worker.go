package windata

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

// WorkerCommand is the hidden command of the binary that runs a worker:
// cahut window-data-worker <folder of the database>.
const WorkerCommand = "window-data-worker"

// File is the name of the database of a Window, in its folder.
const File = "data.db"

// readers is the number of connections that read, beside the one that
// writes.
const readers = 3

// heapLimit bounds the memory of SQLite in a worker.
const heapLimit = 256 << 20

// RunWorker runs the worker of the database in the folder args[0]: it
// confines itself to that folder, then answers the requests of its standard
// input on its standard output until the input ends. It returns the exit
// code of the process.
func RunWorker(args []string, stderr io.Writer) int {
	if len(args) != 1 || !filepath.IsAbs(args[0]) {
		fmt.Fprintf(stderr, "cahut %s: give the absolute path of the folder of a database\n", WorkerCommand)
		return 2
	}
	dir := args[0]
	if err := confine(dir); err != nil {
		fmt.Fprintf(stderr, "cahut %s: confine the worker: %v\n", WorkerCommand, err)
		return 1
	}
	if err := Serve(dir, os.Stdin, os.Stdout); err != nil {
		fmt.Fprintf(stderr, "cahut %s: %v\n", WorkerCommand, err)
		return 1
	}
	return 0
}

// Serve opens the database in dir and answers the requests that come on in
// with responses on out, a line of JSON each, until in ends. Requests run at
// once, as many as the connections allow, and answer in the order they end.
func Serve(dir string, in io.Reader, out io.Writer) error {
	ctx := context.Background()
	d, err := openDatabase(ctx, dir)
	if err != nil {
		return err
	}
	defer d.close()

	dec := json.NewDecoder(in)
	enc := json.NewEncoder(out)
	var mu sync.Mutex
	var wg sync.WaitGroup
	defer wg.Wait()
	for {
		var req request
		if err := dec.Decode(&req); err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return fmt.Errorf("read a request: %w", err)
		}
		wg.Go(func() {
			resp := d.handle(ctx, req)
			resp.ID = req.ID
			mu.Lock()
			defer mu.Unlock()
			_ = enc.Encode(resp) // the service sees a worker that stops answering
		})
	}
}

// database is the database of a Window, opened by its worker: one
// connection that writes, and a few that read.
type database struct {
	dir     string
	write   *sql.DB
	read    *sql.DB
	writer  chan *conn
	readers chan *conn
}

// conn is a connection of a worker, with the policy of the statement that
// it prepares.
type conn struct {
	c      *sql.Conn
	policy atomic.Pointer[policy]
}

func openDatabase(ctx context.Context, dir string) (*database, error) {
	path := filepath.Join(dir, File)
	// No temporary files: Landlock leaves the worker nothing but its folder.
	common := "_pragma=busy_timeout(5000)&_pragma=trusted_schema(0)&_pragma=temp_store(2)&_pragma=foreign_keys(1)&_defensive=1"
	write, err := sql.Open("sqlite", path+"?"+common+"&_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)&_txlock=immediate")
	if err != nil {
		return nil, err
	}
	read, err := sql.Open("sqlite", path+"?"+common+"&_query_only=1")
	if err != nil {
		write.Close()
		return nil, err
	}
	d := &database{dir: dir, write: write, read: read, writer: make(chan *conn, 1), readers: make(chan *conn, readers)}
	w, err := d.connect(ctx, write)
	if err != nil {
		d.close()
		return nil, err
	}
	if _, err := w.c.ExecContext(ctx, fmt.Sprintf("PRAGMA hard_heap_limit = %d", heapLimit)); err != nil {
		d.close()
		return nil, err
	}
	d.writer <- w
	for range readers {
		r, err := d.connect(ctx, read)
		if err != nil {
			d.close()
			return nil, err
		}
		d.readers <- r
	}
	return d, nil
}

// limits are those of sqlite3_limit on every connection of a worker.
var limits = map[int]int{
	sqlite3.SQLITE_LIMIT_LENGTH:              1 << 20,
	sqlite3.SQLITE_LIMIT_SQL_LENGTH:          100_000,
	sqlite3.SQLITE_LIMIT_COLUMN:              200,
	sqlite3.SQLITE_LIMIT_EXPR_DEPTH:          100,
	sqlite3.SQLITE_LIMIT_COMPOUND_SELECT:     50,
	sqlite3.SQLITE_LIMIT_FUNCTION_ARG:        32,
	sqlite3.SQLITE_LIMIT_ATTACHED:            1, // for VACUUM INTO, which copies; the authorizer refuses ATTACH
	sqlite3.SQLITE_LIMIT_LIKE_PATTERN_LENGTH: 1000,
	sqlite3.SQLITE_LIMIT_VARIABLE_NUMBER:     100,
	sqlite3.SQLITE_LIMIT_TRIGGER_DEPTH:       10,
	sqlite3.SQLITE_LIMIT_WORKER_THREADS:      0,
}

// connect opens a connection of db, sets its limits and installs its
// authorizer, which follows the policy of the statement it prepares.
func (d *database) connect(ctx context.Context, db *sql.DB) (*conn, error) {
	c, err := db.Conn(ctx)
	if err != nil {
		return nil, err
	}
	cn := &conn{c: c}
	for id, v := range limits {
		if _, err := sqlite.Limit(c, id, v); err != nil {
			c.Close()
			return nil, err
		}
	}
	err = c.Raw(func(dc any) error {
		r, ok := dc.(sqlite.AuthorizerRegisterer)
		if !ok {
			return errors.New("the driver has no authorizer")
		}
		return r.RegisterAuthorizer(func(action sqlite.AuthorizerActionCode, a1, a2, db, trigger string) sqlite.AuthorizerReturnCode {
			return cn.policy.Load().authorize(action, a1, a2, db, trigger)
		})
	})
	if err != nil {
		c.Close()
		return nil, err
	}
	// The pages of SQLite are read and written at once: a check that the
	// database is readable, on which every later statement relies.
	if _, err := c.ExecContext(ctx, "SELECT count(*) FROM sqlite_schema"); err != nil {
		c.Close()
		return nil, err
	}
	return cn, nil
}

func (d *database) close() {
	for _, ch := range []chan *conn{d.writer, d.readers} {
		for {
			select {
			case c := <-ch:
				c.c.Close()
				continue
			default:
			}
			break
		}
	}
	d.write.Close()
	d.read.Close()
}

// take waits for a connection of ch, as long as ctx lasts.
func take(ctx context.Context, ch chan *conn) (*conn, *Error) {
	select {
	case c := <-ch:
		return c, nil
	case <-ctx.Done():
		return nil, &Error{Code: CodeBusy, Message: "the database of this Window is busy: try again in a moment"}
	}
}

// handle answers a request, within its limits.
func (d *database) handle(ctx context.Context, req request) response {
	millis := req.Limits.Millis
	if millis <= 0 {
		millis = ReaderLimits.Millis
	}
	ctx, cancel := context.WithTimeout(ctx, time.Duration(millis)*time.Millisecond)
	defer cancel()
	var resp response
	var err *Error
	switch req.Op {
	case opSelect, opQuery:
		err = d.selectRows(ctx, req, &resp)
	case opCall:
		err = d.call(ctx, req, &resp)
	case opExec:
		err = d.exec(ctx, req, &resp)
	case opCheck, opMigrate:
		err = d.migrate(ctx, req, &resp)
	case opCopy:
		var c Copy
		c, err = d.copy(ctx, req.Keep)
		resp.Copy = &c
	case opRestore:
		err = d.restore(ctx, req, &resp)
	case opSchema:
		var s Schema
		s, err = d.schema(ctx)
		resp.Schema = &s
	default:
		err = &Error{Code: CodeInternal, Message: "unknown request " + req.Op}
	}
	if err != nil {
		if err.Code == CodeTimeout {
			err.Message = fmt.Sprintf("the statement ran for more than %s, and was stopped: make it do less", time.Duration(millis)*time.Millisecond)
		}
		return response{Error: err}
	}
	return resp
}

// one returns the single statement of sql, whose first word is one of
// firsts.
func one(sql string, firsts ...string) (statement, *Error) {
	stmts, err := split(sql)
	if err != nil {
		return statement{}, err
	}
	if len(stmts) != 1 {
		return statement{}, invalidf("give one statement, not %d", len(stmts))
	}
	if !slices.Contains(firsts, stmts[0].first) {
		return statement{}, invalidf("a statement here starts with %s, not %q", strings.Join(firsts, ", "), stmts[0].first)
	}
	return stmts[0], nil
}

// args binds the parameters that s names, from params: a parameter not
// given is NULL.
func args(s statement, params map[string]json.RawMessage) ([]any, *Error) {
	out := make([]any, 0, len(s.params))
	for _, name := range s.params {
		var v any
		if raw, ok := params[name]; ok {
			var err error
			if v, err = decodeValue(raw); err != nil {
				return nil, invalidf("parameter %s: %v", name, err)
			}
		}
		out = append(out, sql.Named(name, v))
	}
	return out, nil
}

var reads = []string{"SELECT", "WITH", "VALUES"}

func (d *database) selectRows(ctx context.Context, req request, resp *response) *Error {
	s, err := one(req.SQL, reads...)
	if err != nil {
		return err
	}
	a, err := args(s, req.Params)
	if err != nil {
		return err
	}
	c, err := take(ctx, d.readers)
	if err != nil {
		return err
	}
	defer func() { d.readers <- c }()
	p := &policy{op: req.Op, readable: readableOf(req.Readable)}
	if req.Op == opSelect {
		if p.tables, err = tables(ctx, c.c); err != nil {
			return err
		}
	}
	rows, err := c.query(ctx, c.c, p, s.text, a, req.Limits)
	if err != nil {
		return err
	}
	resp.Columns, resp.Rows, resp.Cut = rows.Columns, rows.Rows, rows.Cut
	return nil
}

// querier is what runs a statement: a connection, or its transaction.
type querier interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

// query runs a statement that returns rows under p, and reads them within
// lim.
func (c *conn) query(ctx context.Context, q querier, p *policy, text string, a []any, lim Limits) (Rows, *Error) {
	c.policy.Store(p)
	defer c.policy.Store(nil)
	rows, err := q.QueryContext(ctx, text, a...)
	if err != nil {
		return Rows{}, sqlError(ctx, err, p)
	}
	defer rows.Close()
	out, err := collect(rows, lim)
	if err != nil {
		return Rows{}, sqlError(ctx, err, p)
	}
	return out, nil
}

// run runs a statement that returns no rows under p.
func (c *conn) run(ctx context.Context, q querier, p *policy, text string, a []any) (sql.Result, *Error) {
	c.policy.Store(p)
	defer c.policy.Store(nil)
	res, err := q.ExecContext(ctx, text, a...)
	if err != nil {
		return nil, sqlError(ctx, err, p)
	}
	return res, nil
}

// collect reads rows within lim, and marks them cut beyond it.
func collect(rows *sql.Rows, lim Limits) (Rows, error) {
	cols, err := rows.Columns()
	if err != nil {
		return Rows{}, err
	}
	out := Rows{Columns: cols, Rows: [][]json.RawMessage{}}
	var size int64
	vals := make([]any, len(cols))
	ptrs := make([]any, len(cols))
	for i := range vals {
		ptrs[i] = &vals[i]
	}
	for rows.Next() {
		if lim.Rows > 0 && len(out.Rows) >= lim.Rows {
			out.Cut = true
			break
		}
		if err := rows.Scan(ptrs...); err != nil {
			return Rows{}, err
		}
		row := make([]json.RawMessage, len(cols))
		for i, v := range vals {
			row[i] = encodeValue(v)
			size += int64(len(row[i]) + 1)
		}
		if lim.Bytes > 0 && size > lim.Bytes {
			out.Cut = true
			break
		}
		out.Rows = append(out.Rows, row)
	}
	return out, rows.Err()
}

var calls = []string{"SELECT", "WITH", "VALUES", "INSERT", "REPLACE", "UPDATE", "DELETE"}

// call runs a query of a declaration, with the connection that writes when
// it writes.
func (d *database) call(ctx context.Context, req request, resp *response) *Error {
	s, err := one(req.SQL, calls...)
	if err != nil {
		return err
	}
	a, err := args(s, req.Params)
	if err != nil {
		return err
	}
	p := &policy{op: opCall}
	if !req.Write {
		c, err := take(ctx, d.readers)
		if err != nil {
			return err
		}
		defer func() { d.readers <- c }()
		rows, err := c.query(ctx, c.c, p, s.text, a, req.Limits)
		if err != nil {
			return err
		}
		resp.Columns, resp.Rows, resp.Cut = rows.Columns, rows.Rows, rows.Cut
		return nil
	}
	c, err := take(ctx, d.writer)
	if err != nil {
		return err
	}
	defer func() { d.writer <- c }()
	return d.inTx(ctx, c, req.Limits, false, func(tx *sql.Tx) *Error {
		if s.returning {
			rows, err := c.query(ctx, tx, p, s.text, a, req.Limits)
			if err != nil {
				return err
			}
			resp.Columns, resp.Rows, resp.Cut = rows.Columns, rows.Rows, rows.Cut
			resp.Changes = int64(len(rows.Rows))
			return nil
		}
		res, err := c.run(ctx, tx, p, s.text, a)
		if err != nil {
			return err
		}
		resp.Changes, _ = res.RowsAffected()
		if p.kind() == Adds {
			resp.LastID, _ = res.LastInsertId()
		}
		return nil
	})
}

// inTx runs fn in a transaction of the connection that writes, with the
// size of the database bounded by lim, and commits it unless dry or fn
// fails.
func (d *database) inTx(ctx context.Context, c *conn, lim Limits, dry bool, fn func(*sql.Tx) *Error) *Error {
	if lim.MaxSize > 0 {
		// The pages of a database are 4,096 bytes, as SQLite makes them.
		if _, err := c.c.ExecContext(ctx, fmt.Sprintf("PRAGMA max_page_count = %d", max(lim.MaxSize/4096, 1))); err != nil {
			return sqlError(ctx, err, nil)
		}
	}
	tx, err := c.c.BeginTx(ctx, nil)
	if err != nil {
		return sqlError(ctx, err, nil)
	}
	if e := fn(tx); e != nil || dry {
		_ = tx.Rollback() // nothing of fn must stay
		return e
	}
	if err := tx.Commit(); err != nil {
		return sqlError(ctx, err, nil)
	}
	return nil
}

// exec runs the statements of an AI or a person, and says what they remove
// or rename, by comparing the schema before and after. A dry run rolls
// them back.
func (d *database) exec(ctx context.Context, req request, resp *response) *Error {
	stmts, err := split(req.SQL)
	if err != nil {
		return err
	}
	if len(stmts) == 0 {
		return invalidf("give at least one statement")
	}
	if len(stmts) > 100 {
		return invalidf("give 100 statements at most, not %d", len(stmts))
	}
	for _, s := range stmts {
		if s.first == "VACUUM" || s.first == "ATTACH" || s.first == "DETACH" {
			return invalidf("%s is not allowed: Cahut keeps the database of a Window itself", s.first)
		}
	}
	c, err := take(ctx, d.writer)
	if err != nil {
		return err
	}
	defer func() { d.writer <- c }()
	return d.inTx(ctx, c, req.Limits, req.Dry, func(tx *sql.Tx) *Error {
		before, err := snapshot(ctx, tx)
		if err != nil {
			return err
		}
		for _, s := range stmts {
			a, err := args(s, req.Params)
			if err != nil {
				return err
			}
			res, err := c.run(ctx, tx, &policy{op: opExec}, s.text, a)
			if err != nil {
				return err
			}
			n, _ := res.RowsAffected()
			resp.Changes += n
		}
		after, err := snapshot(ctx, tx)
		if err != nil {
			return err
		}
		resp.Statements = len(stmts)
		resp.Removes = removed(before, after)
		return nil
	})
}

// migrate runs the migrations that have not run yet. A check runs them,
// then the queries and what is readable, and rolls everything back.
func (d *database) migrate(ctx context.Context, req request, resp *response) *Error {
	c, err := take(ctx, d.writer)
	if err != nil {
		return err
	}
	defer func() { d.writer <- c }()
	check := req.Op == opCheck
	return d.inTx(ctx, c, req.Limits, check, func(tx *sql.Tx) *Error {
		if _, err := tx.ExecContext(ctx, "CREATE TABLE IF NOT EXISTS cahut_migrations (n INTEGER PRIMARY KEY, hash TEXT NOT NULL, ran_at TEXT NOT NULL)"); err != nil {
			return sqlError(ctx, err, nil)
		}
		ran, err := ranMigrations(ctx, tx)
		if err != nil {
			return err
		}
		for i, h := range ran {
			if i >= len(req.Migrations) {
				break
			}
			if hash(req.Migrations[i]) != h {
				return invalidf("migration %d changed since it ran: leave the migrations that ran as they are, and add a new one at the end", i+1)
			}
		}
		before, err := snapshot(ctx, tx)
		if err != nil {
			return err
		}
		for i := len(ran); i < len(req.Migrations); i++ {
			stmts, err := split(req.Migrations[i])
			if err != nil {
				return invalidf("migration %d: %s", i+1, err.Message)
			}
			for _, s := range stmts {
				if s.first == "VACUUM" || s.first == "ATTACH" || s.first == "DETACH" {
					return invalidf("migration %d: %s is not allowed", i+1, s.first)
				}
				if len(s.params) > 0 {
					return invalidf("migration %d: a migration binds no parameter", i+1)
				}
				if _, err := c.run(ctx, tx, &policy{op: opMigrate}, s.text, nil); err != nil {
					return invalidf("migration %d: %s", i+1, err.Message)
				}
			}
			if _, err := tx.ExecContext(ctx, "INSERT INTO cahut_migrations (n, hash, ran_at) VALUES (?, ?, ?)",
				i+1, hash(req.Migrations[i]), time.Now().UTC().Format(time.RFC3339)); err != nil {
				return sqlError(ctx, err, nil)
			}
			resp.Ran++
		}
		after, err := snapshot(ctx, tx)
		if err != nil {
			return err
		}
		if gone := removed(before, after); len(gone) > 0 {
			return invalidf("the migrations remove or rename %s: migrations only add; remove or rename with change_window_data", strings.Join(gone, ", "))
		}
		if !check {
			return nil
		}
		if err := checkReadable(after, req.Readable); err != nil {
			return err
		}
		resp.Kinds = map[string]string{}
		for name, text := range req.Queries {
			s, err := one(text, calls...)
			if err != nil {
				return invalidf("query %s: %s", name, err.Message)
			}
			p := &policy{op: opCall}
			c.policy.Store(p)
			stmt, perr := tx.PrepareContext(ctx, s.text)
			c.policy.Store(nil)
			if perr != nil {
				return invalidf("query %s: %s", name, sqlError(ctx, perr, p).Message)
			}
			stmt.Close()
			resp.Kinds[name] = p.kind()
		}
		return nil
	})
}

// tables returns the names of the tables and views of the database, in
// lowercase.
func tables(ctx context.Context, q querier) (map[string]bool, *Error) {
	rows, err := q.QueryContext(ctx, "SELECT lower(name) FROM sqlite_schema WHERE type IN ('table', 'view')")
	if err != nil {
		return nil, sqlError(ctx, err, nil)
	}
	defer rows.Close()
	out := map[string]bool{}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, sqlError(ctx, err, nil)
		}
		out[name] = true
	}
	if err := rows.Err(); err != nil {
		return nil, sqlError(ctx, err, nil)
	}
	return out, nil
}

func hash(migration string) string {
	sum := sha256.Sum256([]byte(migration))
	return hex.EncodeToString(sum[:])
}

func ranMigrations(ctx context.Context, q querier) ([]string, *Error) {
	rows, err := q.QueryContext(ctx, "SELECT hash FROM cahut_migrations ORDER BY n")
	if err != nil {
		return nil, sqlError(ctx, err, nil)
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var h string
		if err := rows.Scan(&h); err != nil {
			return nil, sqlError(ctx, err, nil)
		}
		out = append(out, h)
	}
	if err := rows.Err(); err != nil {
		return nil, sqlError(ctx, err, nil)
	}
	return out, nil
}

// checkReadable checks that what the declaration makes readable names
// tables and their columns.
func checkReadable(s shape, readable map[string][]string) *Error {
	for t, cols := range readable {
		key := "table " + strings.ToLower(t)
		if internal(t) {
			return invalidf("readable: the table %s belongs to SQLite or to Cahut", t)
		}
		if _, ok := s[key]; !ok {
			if _, view := s["view "+strings.ToLower(t)]; view {
				return invalidf("readable: %s is a view; list the tables and the columns it reads", t)
			}
			return invalidf("readable: there is no table %s", t)
		}
		for _, col := range cols {
			if _, ok := s["column "+strings.ToLower(t)+"."+strings.ToLower(col)]; !ok {
				return invalidf("readable: the table %s has no column %s", t, col)
			}
		}
	}
	return nil
}

// shape is the schema of a database as a set: "table t", "column t.c",
// "index i", "view v" and "trigger x", in lowercase, each with its name as
// written.
type shape map[string]string

// snapshot reads the shape of the database.
func snapshot(ctx context.Context, q querier) (shape, *Error) {
	rows, err := q.QueryContext(ctx, `SELECT type, name FROM sqlite_schema
		WHERE type IN ('table', 'index', 'view', 'trigger') AND name NOT LIKE 'sqlite\_%' ESCAPE '\' AND name NOT LIKE 'cahut\_%' ESCAPE '\'`)
	if err != nil {
		return nil, sqlError(ctx, err, nil)
	}
	s := shape{}
	var tables []string
	for rows.Next() {
		var typ, name string
		if err := rows.Scan(&typ, &name); err != nil {
			rows.Close()
			return nil, sqlError(ctx, err, nil)
		}
		s[typ+" "+strings.ToLower(name)] = name
		if typ == "table" {
			tables = append(tables, name)
		}
	}
	rows.Close()
	for _, t := range tables {
		cols, err := q.QueryContext(ctx, "SELECT name FROM pragma_table_info(?)", t)
		if err != nil {
			return nil, sqlError(ctx, err, nil)
		}
		for cols.Next() {
			var name string
			if err := cols.Scan(&name); err != nil {
				cols.Close()
				return nil, sqlError(ctx, err, nil)
			}
			s["column "+strings.ToLower(t)+"."+strings.ToLower(name)] = t + "." + name
		}
		cols.Close()
	}
	return s, nil
}

// removed lists what before holds and after does not, sorted.
func removed(before, after shape) []string {
	var out []string
	for k, name := range before {
		if _, ok := after[k]; !ok {
			kind, _, _ := strings.Cut(k, " ")
			out = append(out, kind+" "+name)
		}
	}
	slices.Sort(out)
	return out
}

// copyName matches the name of a copy: copy-<Unix time in milliseconds>.db.
var copyName = regexp.MustCompile(`^copy-[0-9]{13}\.db$`)

// copy copies the database beside it, and keeps the keep latest copies.
func (d *database) copy(ctx context.Context, keep int) (Copy, *Error) {
	c, err := take(ctx, d.writer)
	if err != nil {
		return Copy{}, err
	}
	defer func() { d.writer <- c }()
	return d.copyWith(ctx, c, keep)
}

func (d *database) copyWith(ctx context.Context, c *conn, keep int) (Copy, *Error) {
	ms := time.Now().UnixMilli()
	name := fmt.Sprintf("copy-%013d.db", ms)
	for {
		if _, err := os.Stat(filepath.Join(d.dir, name)); errors.Is(err, os.ErrNotExist) { //nolint:gosec // a name that copyName matches
			break
		}
		ms++
		name = fmt.Sprintf("copy-%013d.db", ms)
	}
	path := filepath.Join(d.dir, name)
	if _, err := c.c.ExecContext(ctx, "VACUUM INTO ?", path); err != nil {
		return Copy{}, sqlError(ctx, err, nil)
	}
	info, err := os.Stat(path) //nolint:gosec // a name that copyName matches
	if err != nil {
		return Copy{}, &Error{Code: CodeInternal, Message: err.Error()}
	}
	if keep <= 0 {
		keep = 3
	}
	copies, _ := listCopies(d.dir)
	for len(copies) > keep {
		_ = os.Remove(filepath.Join(d.dir, copies[0].Name)) //nolint:gosec // the oldest, which copyName matched: a copy left behind only takes room
		copies = copies[1:]
	}
	return Copy{Name: name, Size: info.Size()}, nil
}

// listCopies lists the copies in dir, the oldest first.
func listCopies(dir string) ([]Copy, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var out []Copy
	for _, e := range entries {
		if !e.Type().IsRegular() || !copyName.MatchString(e.Name()) {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		out = append(out, Copy{Name: e.Name(), Size: info.Size()})
	}
	slices.SortFunc(out, func(a, b Copy) int { return strings.Compare(a.Name, b.Name) })
	return out, nil
}

// restore goes back to the copy that the request names: it first copies
// the database as it is, so that a restore undoes itself, then puts the
// copy in its place and removes it.
func (d *database) restore(ctx context.Context, req request, resp *response) *Error {
	if !copyName.MatchString(req.Copy) {
		return invalidf("%q is not the name of a copy", req.Copy)
	}
	src := filepath.Join(d.dir, req.Copy)
	if _, err := os.Stat(src); err != nil {
		return invalidf("there is no copy %s", req.Copy)
	}
	c, err := take(ctx, d.writer)
	if err != nil {
		return err
	}
	defer func() { d.writer <- c }()
	keep := req.Keep
	if keep <= 0 {
		keep = 3
	}
	// The copy that keeps the database as it is must not push out the one
	// it goes back to.
	now, err := d.copyWith(ctx, c, keep+1)
	if err != nil {
		return err
	}
	rerr := c.c.Raw(func(dc any) error {
		r, ok := dc.(interface {
			NewRestore(string) (*sqlite.Backup, error)
		})
		if !ok {
			return errors.New("the driver cannot restore")
		}
		b, err := r.NewRestore(src)
		if err != nil {
			return err
		}
		if _, err := b.Step(-1); err != nil {
			_ = b.Finish()
			return err
		}
		return b.Finish()
	})
	if rerr != nil {
		_ = os.Remove(filepath.Join(d.dir, now.Name)) // the database did not change
		return sqlError(ctx, rerr, nil)
	}
	_ = os.Remove(src) // the database is that copy now
	resp.Copy = &now
	return nil
}

// schema describes the database.
func (d *database) schema(ctx context.Context) (Schema, *Error) {
	c, err := take(ctx, d.readers)
	if err != nil {
		return Schema{}, err
	}
	defer func() { d.readers <- c }()
	s := Schema{Tables: []Table{}}
	rows, qerr := c.c.QueryContext(ctx, `SELECT type, name FROM sqlite_schema
		WHERE name NOT LIKE 'sqlite\_%' ESCAPE '\' AND name NOT LIKE 'cahut\_%' ESCAPE '\' ORDER BY type, name`)
	if qerr != nil {
		return Schema{}, sqlError(ctx, qerr, nil)
	}
	for rows.Next() {
		var typ, name string
		if err := rows.Scan(&typ, &name); err != nil {
			rows.Close()
			return Schema{}, sqlError(ctx, err, nil)
		}
		switch typ {
		case "table":
			s.Tables = append(s.Tables, Table{Name: name, Columns: []Column{}})
		case "view":
			s.Views = append(s.Views, name)
		case "index":
			s.Indexes = append(s.Indexes, name)
		case "trigger":
			s.Triggers = append(s.Triggers, name)
		}
	}
	rows.Close()
	for i := range s.Tables {
		t := &s.Tables[i]
		cols, err := c.c.QueryContext(ctx, `SELECT name, type, "notnull", pk FROM pragma_table_info(?)`, t.Name)
		if err != nil {
			return Schema{}, sqlError(ctx, err, nil)
		}
		for cols.Next() {
			var col Column
			var notNull, pk int
			if err := cols.Scan(&col.Name, &col.Type, &notNull, &pk); err != nil {
				cols.Close()
				return Schema{}, sqlError(ctx, err, nil)
			}
			col.NotNull, col.PrimaryKey = notNull != 0, pk != 0
			t.Columns = append(t.Columns, col)
		}
		cols.Close()
		if err := c.c.QueryRowContext(ctx, `SELECT count(*) FROM "`+strings.ReplaceAll(t.Name, `"`, `""`)+`"`).Scan(&t.Rows); err != nil {
			return Schema{}, sqlError(ctx, err, nil)
		}
	}
	var hasMigrations int
	if err := c.c.QueryRowContext(ctx, "SELECT count(*) FROM sqlite_schema WHERE name = 'cahut_migrations'").Scan(&hasMigrations); err != nil {
		return Schema{}, sqlError(ctx, err, nil)
	}
	if hasMigrations > 0 {
		if err := c.c.QueryRowContext(ctx, "SELECT count(*) FROM cahut_migrations").Scan(&s.Migrations); err != nil {
			return Schema{}, sqlError(ctx, err, nil)
		}
	}
	var pages, size int64
	if err := c.c.QueryRowContext(ctx, "SELECT page_count, page_size FROM pragma_page_count(), pragma_page_size()").Scan(&pages, &size); err != nil {
		return Schema{}, sqlError(ctx, err, nil)
	}
	s.Size = pages * size
	if info, err := os.Stat(filepath.Join(d.dir, File+"-wal")); err == nil {
		s.Size += info.Size()
	}
	return s, nil
}

// sqlError tells what went wrong with a statement prepared under p.
func sqlError(ctx context.Context, err error, p *policy) *Error {
	if errors.Is(err, context.DeadlineExceeded) || ctx.Err() != nil {
		return &Error{Code: CodeTimeout, Message: "the statement ran out of time"}
	}
	se, ok := errors.AsType[*sqlite.Error](err)
	if !ok {
		return &Error{Code: CodeInternal, Message: err.Error()}
	}
	msg := strings.TrimSpace(se.Error())
	if i := strings.LastIndex(msg, " ("); i > 0 && strings.HasSuffix(msg, ")") {
		msg = msg[:i]
	}
	switch se.Code() & 0xff {
	case sqlite3.SQLITE_AUTH:
		why := "it does what this call may not do"
		if p != nil && p.why() != "" {
			why = p.why()
		}
		return invalidf("not allowed: %s", why)
	case sqlite3.SQLITE_FULL:
		return &Error{Code: CodeFull, Message: "the database of this Window is full"}
	case sqlite3.SQLITE_INTERRUPT:
		return &Error{Code: CodeTimeout, Message: "the statement ran out of time"}
	case sqlite3.SQLITE_CONSTRAINT:
		return &Error{Code: CodeConstraint, Message: msg}
	case sqlite3.SQLITE_TOOBIG:
		return invalidf("a value or the statement is too long: %s", msg)
	case sqlite3.SQLITE_NOMEM:
		return invalidf("the statement needs more memory than a Window has")
	case sqlite3.SQLITE_ERROR, sqlite3.SQLITE_RANGE, sqlite3.SQLITE_MISMATCH:
		return invalidf("%s", msg)
	}
	return &Error{Code: CodeInternal, Message: msg}
}
