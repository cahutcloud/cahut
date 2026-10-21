// Package operation is the only way to act on a home. MCP tools, the HTTP
// API and the command line are thin adapters over its registry, which checks
// who acts, runs the operation and writes the Logbook.
// See docs/adr/0001-operation-registry.md.
package operation

import (
	"bytes"
	"cmp"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"slices"
	"time"

	"cahut.cloud/internal/logbook"
	"cahut.cloud/internal/store"
)

// ActorKind tells who is acting.
type ActorKind string

// Kinds of actors.
const (
	Admin  ActorKind = "admin"  // the administrator, on the server's command line
	Person ActorKind = "person" // a person, in the console
	AI     ActorKind = "ai"     // an AI acting on behalf of a person
)

// Actor is who calls an operation.
type Actor struct {
	Kind    ActorKind
	Via     string // the adapter the call comes through: cli, mcp
	UserID  int64  // the person, or the person an AI acts for; zero for Admin
	Handle  string // the handle of that person
	TokenID int64  // the token an AI presented
	Client  string // the AI, as its token and its software name it
}

// Op is an operation: what it is, who may call it, and how it runs.
type Op[In, Out any] struct {
	// Name is stable and snake_case. It is also the name of the MCP tool.
	Name string
	// Description tells an AI what the operation does and when to use it.
	Description string
	// ReadOnly operations change nothing.
	ReadOnly bool
	// Irreversible operations cannot be undone. An AI needs the approval of
	// a person to run one: Knock first.
	Irreversible bool
	// AdminOnly operations are reserved to the administrator and are not
	// offered to AIs.
	AdminOnly bool
	// Run does the work, inside the transaction of the call. It returns an
	// error wrapping ErrInvalid, ErrNotFound or ErrForbidden when the caller
	// is at fault.
	Run func(ctx context.Context, c *Call, in In) (Out, error)
}

// Info describes an operation to the adapters.
type Info struct {
	Name         string
	Description  string
	ReadOnly     bool
	Irreversible bool
	AdminOnly    bool
	Input        reflect.Type
	Output       reflect.Type
}

// Call is one call of an operation, as Run sees it.
type Call struct {
	Actor Actor
	Now   time.Time // when the call started
	Tx    *sql.Tx   // read-only for a ReadOnly operation

	target, detail string
	afterCommit    []func()
	onFailure      []func()
}

// Log sets what the Logbook entry of the call says the operation acted on,
// with an optional detail. Never put secrets or file contents in either.
func (c *Call) Log(target, detail string) {
	c.target, c.detail = target, detail
}

// AfterCommit runs fn once the call has succeeded and its transaction is
// committed.
func (c *Call) AfterCommit(fn func()) {
	c.afterCommit = append(c.afterCommit, fn)
}

// OnFailure runs fn if the call fails after Run, for instance to remove
// files that Run wrote.
func (c *Call) OnFailure(fn func()) {
	c.onFailure = append(c.onFailure, fn)
}

// Errors that tell the caller what went wrong. Adapters show their messages
// to the caller; any other error is internal.
var (
	ErrInvalid          = errors.New("invalid input")
	ErrNotFound         = errors.New("not found")
	ErrForbidden        = errors.New("not allowed")
	ErrApprovalRequired = errors.New("needs the approval of a person")
)

// Invalidf returns an error wrapping ErrInvalid.
func Invalidf(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrInvalid, fmt.Sprintf(format, args...))
}

// NotFoundf returns an error wrapping ErrNotFound.
func NotFoundf(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrNotFound, fmt.Sprintf(format, args...))
}

// Forbiddenf returns an error wrapping ErrForbidden.
func Forbiddenf(format string, args ...any) error {
	return fmt.Errorf("%w: %s", ErrForbidden, fmt.Sprintf(format, args...))
}

// Public reports whether the message of err may be shown to the caller.
func Public(err error) bool {
	return errors.Is(err, ErrInvalid) || errors.Is(err, ErrNotFound) ||
		errors.Is(err, ErrForbidden) || errors.Is(err, ErrApprovalRequired)
}

// Registry holds the operations of a home and runs them.
type Registry struct {
	db  *store.DB
	ops map[string]registered

	// Now returns the current time. Tests replace it.
	Now func() time.Time
}

type registered interface {
	info() Info
	invoke(ctx context.Context, r *Registry, actor Actor, args json.RawMessage) (any, error)
}

// NewRegistry returns an empty registry over db.
func NewRegistry(db *store.DB) *Registry {
	return &Registry{db: db, ops: map[string]registered{}, Now: time.Now}
}

var validName = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

// Register adds op to r. It panics if the name is invalid or taken, which
// is a programming error.
func Register[In, Out any](r *Registry, op *Op[In, Out]) {
	if !validName.MatchString(op.Name) {
		panic("operation: invalid name " + op.Name)
	}
	if _, ok := r.ops[op.Name]; ok {
		panic("operation: " + op.Name + " registered twice")
	}
	r.ops[op.Name] = op
}

// Ops describes the registered operations, sorted by name.
func (r *Registry) Ops() []Info {
	infos := make([]Info, 0, len(r.ops))
	for _, op := range r.ops {
		infos = append(infos, op.info())
	}
	slices.SortFunc(infos, func(a, b Info) int { return cmp.Compare(a.Name, b.Name) })
	return infos
}

// Do runs op for actor.
func Do[In, Out any](ctx context.Context, r *Registry, op *Op[In, Out], actor Actor, in In) (Out, error) {
	var out Out
	err := r.run(ctx, op.info(), actor, func(ctx context.Context, c *Call) error {
		var err error
		out, err = op.Run(ctx, c, in)
		return err
	})
	if err != nil {
		var zero Out
		return zero, err
	}
	return out, nil
}

// Invoke runs the operation called name with arguments in JSON, for
// adapters that receive JSON such as MCP. Unknown fields are refused.
func (r *Registry) Invoke(ctx context.Context, name string, actor Actor, args json.RawMessage) (any, error) {
	op, ok := r.ops[name]
	if !ok {
		return nil, NotFoundf("no operation %q", name)
	}
	return op.invoke(ctx, r, actor, args)
}

func (op *Op[In, Out]) info() Info {
	return Info{
		Name:         op.Name,
		Description:  op.Description,
		ReadOnly:     op.ReadOnly,
		Irreversible: op.Irreversible,
		AdminOnly:    op.AdminOnly,
		Input:        reflect.TypeFor[In](),
		Output:       reflect.TypeFor[Out](),
	}
}

func (op *Op[In, Out]) invoke(ctx context.Context, r *Registry, actor Actor, args json.RawMessage) (any, error) {
	var out Out
	err := r.run(ctx, op.info(), actor, func(ctx context.Context, c *Call) error {
		var in In
		if len(bytes.TrimSpace(args)) > 0 {
			dec := json.NewDecoder(bytes.NewReader(args))
			dec.DisallowUnknownFields()
			if err := dec.Decode(&in); err != nil {
				return Invalidf("arguments: %v", err)
			}
		}
		var err error
		out, err = op.Run(ctx, c, in)
		return err
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// run checks that actor may call the operation, runs fn in a transaction
// and writes the Logbook. A change and its entry share the transaction.
// Every call by an AI is logged, reads included; reads by the
// administrator or a person are not.
func (r *Registry) run(ctx context.Context, info Info, actor Actor, fn func(context.Context, *Call) error) error {
	logged := !info.ReadOnly || actor.Kind == AI
	c := &Call{Actor: actor, Now: r.Now()}

	if err := authorize(info, actor); err != nil {
		return r.fail(ctx, info, c, logged, err)
	}
	db := r.db.Write
	if info.ReadOnly {
		db = r.db.Read
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return r.fail(ctx, info, c, logged, err)
	}
	c.Tx = tx

	err = fn(ctx, c)
	if err == nil && logged && !info.ReadOnly {
		err = logbook.Write(ctx, tx, r.entry(info, c, nil))
	}
	if err == nil {
		err = tx.Commit()
	}
	if err != nil {
		_ = tx.Rollback() // the error that matters is err
		for _, f := range c.onFailure {
			f()
		}
		return r.fail(ctx, info, c, logged, err)
	}
	if logged && info.ReadOnly {
		if err := logbook.Write(ctx, r.db.Write, r.entry(info, c, nil)); err != nil {
			return err
		}
	}
	for _, f := range c.afterCommit {
		f()
	}
	return nil
}

// fail logs a refused or failed call, outside any transaction of the call,
// and returns err.
func (r *Registry) fail(ctx context.Context, info Info, c *Call, logged bool, err error) error {
	if !logged {
		return err
	}
	if lerr := logbook.Write(ctx, r.db.Write, r.entry(info, c, err)); lerr != nil {
		return errors.Join(err, lerr)
	}
	return err
}

func (r *Registry) entry(info Info, c *Call, err error) logbook.Entry {
	e := logbook.Entry{
		At:         c.Now,
		Operation:  info.Name,
		Actor:      string(c.Actor.Kind),
		UserID:     c.Actor.UserID,
		UserHandle: c.Actor.Handle,
		TokenID:    c.Actor.TokenID,
		Client:     c.Actor.Client,
		Via:        c.Actor.Via,
		Target:     c.target,
		Detail:     c.detail,
		Outcome:    logbook.OK,
	}
	if err != nil {
		e.Outcome = logbook.Failed
		if errors.Is(err, ErrForbidden) || errors.Is(err, ErrApprovalRequired) {
			e.Outcome = logbook.Refused
		}
		e.Error = err.Error()
	}
	return e
}

func authorize(info Info, a Actor) error {
	switch a.Kind {
	case Admin:
	case Person:
		if a.UserID == 0 {
			return errors.New("operation: person without a user")
		}
	case AI:
		if a.UserID == 0 || a.TokenID == 0 {
			return errors.New("operation: AI without a user or a token")
		}
	default:
		return fmt.Errorf("operation: unknown actor kind %q", a.Kind)
	}
	if info.AdminOnly && a.Kind != Admin {
		return Forbiddenf("%s is reserved to the administrator", info.Name)
	}
	if info.Irreversible && a.Kind == AI {
		return fmt.Errorf("%w: %s cannot be undone, and approvals are not available yet", ErrApprovalRequired, info.Name)
	}
	return nil
}
