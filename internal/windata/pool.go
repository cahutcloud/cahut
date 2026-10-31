package windata

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"sync"
	"testing"
	"time"
)

// Options configure a Pool.
type Options struct {
	// Command returns the command that runs the worker of the database in
	// dir: by default this binary, with WorkerCommand, an empty environment
	// and no file of the service.
	Command func(dir string) (*exec.Cmd, error)
	// Inline runs the workers in this process, through the same protocol,
	// without their confinement: for the tests of other packages, where it
	// is the default unless Command is set, since a test binary knows no
	// WorkerCommand.
	Inline bool
	// MaxWorkers bounds the workers that run at once: 32 by default.
	MaxWorkers int
	// Idle is how long a worker waits for a call before it stops: a minute
	// by default.
	Idle time.Duration
	// Calls bounds the calls that one worker runs at once: 4 by default,
	// one that writes and three that read.
	Calls int
	// Log receives what goes wrong with workers.
	Log *slog.Logger
}

// grace is how long a worker may answer after the time of its request,
// before the pool stops it.
const grace = 2 * time.Second

// wait is how long a call waits for a worker that is busy.
const wait = 2 * time.Second

// Pool runs the workers of the databases of Windows, one for each
// database in use, named by its folder.
type Pool struct {
	opts    Options
	mu      sync.Mutex
	workers map[string]*worker
	closed  bool
	stop    chan struct{}
	janitor sync.WaitGroup
}

// NewPool returns a pool of workers.
func NewPool(o Options) *Pool {
	if o.Command == nil {
		o.Command = command
		o.Inline = o.Inline || testing.Testing()
	}
	if o.MaxWorkers <= 0 {
		o.MaxWorkers = 32
	}
	if o.Idle <= 0 {
		o.Idle = time.Minute
	}
	if o.Calls <= 0 {
		o.Calls = 1 + readers
	}
	if o.Log == nil {
		o.Log = slog.New(slog.DiscardHandler)
	}
	p := &Pool{opts: o, workers: map[string]*worker{}, stop: make(chan struct{})}
	p.janitor.Go(p.sweep)
	return p
}

// command runs this binary as the worker of the database in dir.
func command(dir string) (*exec.Cmd, error) {
	exe, err := os.Executable()
	if err != nil {
		return nil, err
	}
	// The worker outlives the call that starts it: the pool stops it.
	cmd := exec.Command(exe, WorkerCommand, dir) //nolint:noctx // the pool stops it
	cmd.Env = []string{}
	return cmd, nil
}

// worker is a worker process, or a worker in this process for Inline.
type worker struct {
	dir     string
	cmd     *exec.Cmd
	in      io.WriteCloser
	out     io.Closer
	stderr  *tail
	mu      sync.Mutex // guards enc, pending and next
	enc     *json.Encoder
	pending map[uint64]chan response
	next    uint64
	calls   chan struct{}
	done    chan struct{} // closed once it stopped
	// guarded by the mutex of the pool
	inflight int
	used     time.Time
}

func (w *worker) stopped() bool {
	select {
	case <-w.done:
		return true
	default:
		return false
	}
}

// start starts the worker of the database in dir.
func (p *Pool) start(dir string) (*worker, error) {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, err
	}
	w := &worker{dir: dir, pending: map[uint64]chan response{}, calls: make(chan struct{}, p.opts.Calls),
		done: make(chan struct{}), stderr: &tail{}}
	var out io.ReadCloser
	if p.opts.Inline {
		inR, inW := io.Pipe()
		outR, outW := io.Pipe()
		go func() {
			err := Serve(dir, inR, outW)
			if err == nil {
				err = io.EOF
			}
			outW.CloseWithError(err)
		}()
		w.in, out = inW, outR
	} else {
		cmd, err := p.opts.Command(dir)
		if err != nil {
			return nil, err
		}
		if cmd.SysProcAttr == nil {
			detach(cmd)
		}
		cmd.Stderr = w.stderr
		if w.in, err = cmd.StdinPipe(); err != nil {
			return nil, err
		}
		if out, err = cmd.StdoutPipe(); err != nil {
			return nil, err
		}
		if err := cmd.Start(); err != nil {
			return nil, err
		}
		w.cmd = cmd
	}
	w.out = out
	w.enc = json.NewEncoder(w.in)
	go p.read(w, out)
	return w, nil
}

// read gives each response of w to the call that waits for it, until w
// stops, and then fails the calls that still wait.
func (p *Pool) read(w *worker, out io.Reader) {
	dec := json.NewDecoder(out)
	for {
		var r response
		if err := dec.Decode(&r); err != nil {
			break
		}
		w.mu.Lock()
		ch := w.pending[r.ID]
		delete(w.pending, r.ID)
		w.mu.Unlock()
		if ch != nil {
			ch <- r
		}
	}
	w.mu.Lock()
	for id, ch := range w.pending {
		ch <- response{ID: id, Error: &Error{Code: CodeUnavailable, Message: "the database of this Window stopped answering: try again"}}
		delete(w.pending, id)
	}
	w.mu.Unlock()
	if w.cmd != nil {
		if err := w.cmd.Wait(); err != nil {
			p.opts.Log.Warn("a worker of the data of a Window stopped", "folder", w.dir, "error", err, "output", w.stderr.String())
		}
	}
	close(w.done)
	p.mu.Lock()
	if p.workers[w.dir] == w {
		delete(p.workers, w.dir)
	}
	p.mu.Unlock()
}

// acquire returns the worker of dir, which it starts if need be, once it
// may take one more call.
func (p *Pool) acquire(ctx context.Context, dir string) (*worker, error) {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return nil, &Error{Code: CodeUnavailable, Message: "the service is stopping"}
	}
	w := p.workers[dir]
	if w == nil || w.stopped() {
		if len(p.workers) >= p.opts.MaxWorkers && !p.evict() {
			p.mu.Unlock()
			return nil, &Error{Code: CodeBusy, Message: "the server runs as many databases of Windows as it can: try again in a moment"}
		}
		var err error
		if w, err = p.start(dir); err != nil {
			p.mu.Unlock()
			return nil, err
		}
		p.workers[dir] = w
	}
	w.inflight++
	w.used = time.Now()
	p.mu.Unlock()

	timer := time.NewTimer(wait)
	defer timer.Stop()
	select {
	case w.calls <- struct{}{}:
		return w, nil
	case <-ctx.Done():
	case <-timer.C:
	}
	p.mu.Lock()
	w.inflight--
	p.mu.Unlock()
	return nil, &Error{Code: CodeBusy, Message: "the database of this Window is busy: try again in a moment"}
}

func (p *Pool) release(w *worker) {
	<-w.calls
	p.mu.Lock()
	w.inflight--
	w.used = time.Now()
	p.mu.Unlock()
}

// evict stops the worker that waited the longest without a call, and
// reports whether there was one. p.mu is held.
func (p *Pool) evict() bool {
	var oldest *worker
	for _, w := range p.workers {
		if w.inflight == 0 && (oldest == nil || w.used.Before(oldest.used)) {
			oldest = w
		}
	}
	if oldest == nil {
		return false
	}
	delete(p.workers, oldest.dir)
	_ = oldest.in.Close() // the worker ends with its input
	return true
}

// sweep stops the workers that have waited for a call longer than Idle.
func (p *Pool) sweep() {
	tick := time.NewTicker(max(p.opts.Idle/2, time.Second))
	defer tick.Stop()
	for {
		select {
		case <-p.stop:
			return
		case <-tick.C:
		}
		p.mu.Lock()
		for dir, w := range p.workers {
			if w.inflight == 0 && time.Since(w.used) > p.opts.Idle {
				delete(p.workers, dir)
				_ = w.in.Close() // the worker ends with its input
			}
		}
		p.mu.Unlock()
	}
}

// kill stops w at once, for a request it did not answer in time.
func (p *Pool) kill(w *worker) {
	p.mu.Lock()
	if p.workers[w.dir] == w {
		delete(p.workers, w.dir)
	}
	p.mu.Unlock()
	if w.cmd != nil {
		_ = w.cmd.Process.Kill() // it is stuck: what it was doing rolls back
	} else {
		_ = w.in.Close()
		_ = w.out.Close()
	}
}

// Stop stops the worker of dir, if it runs, and waits for it: before the
// folder of its database is removed.
func (p *Pool) Stop(dir string) {
	p.mu.Lock()
	w := p.workers[dir]
	delete(p.workers, dir)
	p.mu.Unlock()
	if w == nil {
		return
	}
	_ = w.in.Close() // the worker ends with its input
	select {
	case <-w.done:
	case <-time.After(5 * time.Second):
		p.kill(w)
		<-w.done
	}
}

// Close stops every worker.
func (p *Pool) Close() {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return
	}
	p.closed = true
	dirs := make([]string, 0, len(p.workers))
	for dir := range p.workers {
		dirs = append(dirs, dir)
	}
	p.mu.Unlock()
	close(p.stop)
	p.janitor.Wait()
	for _, dir := range dirs {
		p.Stop(dir)
	}
}

// do sends req to the worker of dir and waits for its response, at most
// the time of the request and a grace, after which it stops the worker.
func (p *Pool) do(ctx context.Context, dir string, req request) (response, error) {
	w, err := p.acquire(ctx, dir)
	if err != nil {
		return response{}, err
	}
	defer p.release(w)

	ch := make(chan response, 1)
	w.mu.Lock()
	w.next++
	req.ID = w.next
	w.pending[req.ID] = ch
	err = w.enc.Encode(req)
	w.mu.Unlock()
	if err != nil {
		p.kill(w)
		return response{}, &Error{Code: CodeUnavailable, Message: "the database of this Window stopped answering: try again"}
	}
	timer := time.NewTimer(time.Duration(req.Limits.Millis)*time.Millisecond + grace)
	defer timer.Stop()
	select {
	case r := <-ch:
		if r.Error != nil {
			return r, r.Error
		}
		return r, nil
	case <-timer.C:
		p.opts.Log.Warn("a worker of the data of a Window did not answer in time, and was stopped", "folder", dir, "op", req.Op)
		p.kill(w)
		return response{}, &Error{Code: CodeTimeout, Message: "the statement ran out of time"}
	case <-ctx.Done():
		w.mu.Lock()
		delete(w.pending, req.ID)
		w.mu.Unlock()
		return response{}, ctx.Err()
	}
}

// Select runs a SELECT of a reader over what readable lists: every other
// column of a table it lists reads as NULL, and every other table is
// refused.
func (p *Pool) Select(ctx context.Context, dir string, lim Limits, sql string, params map[string]json.RawMessage, readable map[string][]string) (Rows, error) {
	r, err := p.do(ctx, dir, request{Op: opSelect, SQL: sql, Params: params, Readable: readable, Limits: lim})
	return Rows{Columns: r.Columns, Rows: r.Rows, Cut: r.Cut}, err
}

// Query runs a SELECT over the whole database, for its AIs and its people.
func (p *Pool) Query(ctx context.Context, dir string, lim Limits, sql string, params map[string]json.RawMessage) (Rows, error) {
	r, err := p.do(ctx, dir, request{Op: opQuery, SQL: sql, Params: params, Limits: lim})
	return Rows{Columns: r.Columns, Rows: r.Rows, Cut: r.Cut}, err
}

// Call runs a query of a declaration, which writes when write is true.
func (p *Pool) Call(ctx context.Context, dir string, lim Limits, sql string, params map[string]json.RawMessage, write bool) (Result, error) {
	r, err := p.do(ctx, dir, request{Op: opCall, SQL: sql, Params: params, Write: write, Limits: lim})
	return Result{Rows: Rows{Columns: r.Columns, Rows: r.Rows, Cut: r.Cut}, Changes: r.Changes, LastID: r.LastID}, err
}

// Exec runs the statements of an AI or a person, with their named
// parameters, or runs them and rolls them back when dry, and says what
// they did.
func (p *Pool) Exec(ctx context.Context, dir string, lim Limits, sql string, params map[string]json.RawMessage, dry bool) (Change, error) {
	r, err := p.do(ctx, dir, request{Op: opExec, SQL: sql, Params: params, Dry: dry, Limits: lim})
	return Change{Statements: r.Statements, Changes: r.Changes, Removes: r.Removes}, err
}

// Check runs the migrations of a declaration that have not run yet,
// prepares its queries and checks what it makes readable, then rolls
// everything back.
func (p *Pool) Check(ctx context.Context, dir string, c Check) (Checked, error) {
	r, err := p.do(ctx, dir, request{Op: opCheck, Migrations: c.Migrations, Queries: c.Queries, Readable: c.Readable, Limits: AILimits})
	return Checked{Kinds: r.Kinds, New: r.Ran}, err
}

// Migrate runs the migrations that have not run yet, and returns how many
// it ran.
func (p *Pool) Migrate(ctx context.Context, dir string, lim Limits, migrations []string) (int, error) {
	r, err := p.do(ctx, dir, request{Op: opMigrate, Migrations: migrations, Limits: lim})
	return r.Ran, err
}

// slow are the limits of copying, restoring and describing a database.
var slow = Limits{Millis: 30000}

// Copy copies the database beside it, and keeps the keep latest copies.
func (p *Pool) Copy(ctx context.Context, dir string, keep int) (Copy, error) {
	r, err := p.do(ctx, dir, request{Op: opCopy, Keep: keep, Limits: slow})
	if err != nil {
		return Copy{}, err
	}
	return *r.Copy, nil
}

// Restore goes back to the copy called name. The database as it was
// becomes a copy, which it returns, so that a restore undoes itself.
func (p *Pool) Restore(ctx context.Context, dir, name string, keep int) (Copy, error) {
	r, err := p.do(ctx, dir, request{Op: opRestore, Copy: name, Keep: keep, Limits: slow})
	if err != nil {
		return Copy{}, err
	}
	return *r.Copy, nil
}

// Schema describes the database.
func (p *Pool) Schema(ctx context.Context, dir string) (Schema, error) {
	r, err := p.do(ctx, dir, request{Op: opSchema, Limits: slow})
	if err != nil {
		return Schema{}, err
	}
	return *r.Schema, nil
}

// Copies lists the copies of the database in dir, the oldest first.
func Copies(dir string) ([]Copy, error) {
	c, err := listCopies(dir)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	return c, err
}

// tail keeps the last few kilobytes that a worker writes on its standard
// error, for the log.
type tail struct {
	mu  sync.Mutex
	buf []byte
}

func (t *tail) Write(b []byte) (int, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.buf = append(t.buf, b...)
	if len(t.buf) > 4096 {
		t.buf = t.buf[len(t.buf)-4096:]
	}
	return len(b), nil
}

func (t *tail) String() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return string(t.buf)
}
