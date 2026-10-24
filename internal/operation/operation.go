// Package operation is the only way to act on a Cahut or on its server. MCP
// tools, the HTTP API and the command line are thin adapters over its
// registry, which checks who acts, runs the operation and writes the Logbook.
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
	Admin   ActorKind = "admin"   // the command line, or an administrator of the server, who has a handle
	Person  ActorKind = "person"  // a person, in the console
	AI      ActorKind = "ai"      // an AI acting on behalf of a person
	Visitor ActorKind = "visitor" // someone who has not signed in, to sign in
)

// Actor is who calls an operation.
type Actor struct {
	Kind    ActorKind
	Via     string // the adapter the call comes through: cli, mcp
	Domain  string // the domain the call came to; empty on the command line
	UserID  int64  // the person, or the person an AI acts for; zero for Admin
	Handle  string // the handle of that person, or of the administrator of the server
	Session int64  // the session a person or an administrator signed in with
	TokenID int64  // the personal token an AI presented
	Conn    int64  // the connection of an AI that connected over OAuth
	Client  string // the AI, as its token or its connection, and its software, name it
	// Address is the client's address on the network, as limits count it:
	// an IPv4 address, or the /64 of an IPv6 one. Empty on the command line.
	Address string

	// The reach that a person granted an AI. See
	// docs/adr/0016-oauth-for-ai-clients.md.
	ReadOnly bool    // the AI may only call operations that change nothing
	Rooms    []int64 // the only Rooms the AI reaches; every Room of its person when nil
}

// Reaches reports whether the actor may reach the Room id at all, as far as
// the grant of an AI goes. The rights of its person still apply.
func (a Actor) Reaches(id int64) bool {
	return a.Rooms == nil || slices.Contains(a.Rooms, id)
}

// Limited reports whether the actor is an AI that reaches some Rooms only.
func (a Actor) Limited() bool {
	return a.Rooms != nil
}

// CommandLine reports whether the actor is the command line on the server,
// which reaches everything: whoever runs it can read the disk anyway.
func (a Actor) CommandLine() bool {
	return a.Kind == Admin && a.Handle == ""
}

// Level is who may call an operation.
type Level int

// Levels, from the widest. See docs/adr/0013-people-sessions-and-rights.md.
const (
	// AnyPerson is any person of the Cahut, and the AIs that act for one.
	AnyPerson Level = iota
	// CahutAdmin is the administrators of the Cahut, and those of the server.
	CahutAdmin
	// ServerAdmin is the administrators of the server.
	ServerAdmin
	// Nobody may call the operation: in an individual Cahut, those that
	// manage the members of a Room. See Op.InIndividual.
	Nobody
)

func (l Level) String() string {
	switch l {
	case AnyPerson:
		return "any person"
	case CahutAdmin:
		return "administrators of the Cahut"
	case ServerAdmin:
		return "administrators of the server"
	case Nobody:
		return "nobody"
	}
	return fmt.Sprintf("level %d", int(l))
}

// Op is an operation: what it is, who may call it, and how it runs.
type Op[In, Out any] struct {
	// Name is stable and snake_case. It is also the name of the MCP tool.
	Name string
	// Title names the operation in a few words of plain English, such as
	// "Publish a Window". MCP clients show it to the person, as ChatGPT
	// does when it asks them to confirm a call.
	Title string
	// Description tells an AI what the operation does and when to use it.
	Description string
	// ReadOnly operations change nothing.
	ReadOnly bool
	// Level is who may call the operation: any person when it is not set.
	Level Level
	// InIndividual is who may call the operation in an individual Cahut,
	// when it is narrower than Level: the administrators of the Cahut for
	// listing its people, Nobody for managing the members of a Room. See
	// docs/adr/0030-kind-and-limits-of-each-cahut.md.
	InIndividual Level
	// PeopleOnly operations prove who someone is, such as signing in, or
	// serve the session of a person in the console, such as previews. They
	// are not offered to AIs.
	PeopleOnly bool
	// Visitors may call the operation before they sign in: it signs them in.
	Visitors bool
	// KnockFirst operations cannot be undone, or change what a person may
	// do. An AI needs the approval of a person to run one: its call makes a
	// Knock. See docs/adr/0021-ais-knock-first.md.
	KnockFirst bool
	// KnocksWhen says, for an operation that knocks first for some calls
	// only, which ones, such as "it makes a Window public". Run calls
	// Call.Knock for them.
	KnocksWhen string
	// ReadOnlyAIs operations change things, yet are offered to the AIs that
	// may only read: set_ai_reach is how they ask for more.
	ReadOnlyAIs bool
	// OpenWorld operations reach people outside the Cahut, such as
	// invite_to_window, which sends e-mail. MCP tells AI clients so.
	OpenWorld bool
	// Prepare, when set, does the slow part of a call before its
	// transaction opens, so that the call holds no connection to the
	// database meanwhile: checking a password, fetching a document. c.Tx is
	// nil; c.Read reads, and c.Keep keeps what Run needs. An error fails the
	// call. An operation that knocks first has no Prepare: a Knock runs
	// again later, without it.
	Prepare func(ctx context.Context, c *Call, in In) error
	// Run does the work, inside the transaction of the call. It returns an
	// error wrapping ErrInvalid, ErrNotFound or ErrForbidden when the caller
	// is at fault.
	Run func(ctx context.Context, c *Call, in In) (Out, error)
}

// Info describes an operation to the adapters.
type Info struct {
	Name         string
	Title        string
	Description  string
	ReadOnly     bool
	Level        Level
	InIndividual Level
	PeopleOnly   bool
	Visitors     bool
	KnockFirst   bool
	KnocksWhen   string
	ReadOnlyAIs  bool
	OpenWorld    bool
	Input        reflect.Type
	Output       reflect.Type
}

// ForAIs reports whether the operation is offered to the AIs of a person,
// who administers the Cahut or not, when they may change things or only
// read, in a collaborative Cahut.
func (i Info) ForAIs(admin, readOnly bool) bool {
	return !i.PeopleOnly && (i.Level == AnyPerson || i.Level == CahutAdmin && admin) && (i.ReadOnly || i.ReadOnlyAIs || !readOnly)
}

// ForAIsIn reports, as ForAIs does, whether the operation is offered to
// the AIs of a person, in a Cahut that is individual or not.
func (i Info) ForAIsIn(individual, admin, readOnly bool) bool {
	l := i.LevelIn(individual)
	return i.ForAIs(admin, readOnly) && (l == AnyPerson || l == CahutAdmin && admin)
}

// LevelIn returns who may call the operation in a Cahut that is individual
// or not.
func (i Info) LevelIn(individual bool) Level {
	if individual && i.InIndividual > i.Level {
		return i.InIndividual
	}
	return i.Level
}

// Call is one call of an operation, as Run sees it.
type Call struct {
	Actor Actor
	Now   time.Time // when the call started, in UTC, to the millisecond
	Tx    *sql.Tx   // read-only for a ReadOnly operation; nil in Prepare
	// Admin reports whether the actor administers the Cahut: the person who
	// acts, or for whom an AI acts, is an administrator of the Cahut, or the
	// actor administers the server.
	Admin bool

	target, detail string
	room           int64     // the Room the call concerns
	as             ActorKind // what a Visitor's call is logged as
	afterCommit    []func()
	onFailure      []func()

	db       *store.DB
	values   map[any]any // what the registry provides
	kept     any         // what Prepare kept for Run
	input    any         // the input of the call, which a Knock keeps
	knocks   bool        // the call waits for a person: an AI made it, and it knocks first
	approved int64       // the Knock that a person approved for the call to run
	knock    int64       // the Knock that the call made or answered, for the Logbook
	unlogged bool        // the call wrote the entries that the Logbook needs
}

// Log sets what the Logbook entry of the call says the operation acted on,
// with an optional detail. Never put secrets or file contents in either.
func (c *Call) Log(target, detail string) {
	c.target, c.detail = target, detail
}

// InRoom names the Room that the call concerns, so that the members of that
// Room read its Logbook entry.
func (c *Call) InRoom(id int64) {
	c.room = id
}

// Identify names who a Visitor's call concerns, such as the person who
// signs in, so that the Logbook records the call under them: a Person, with
// their id, or an Admin of the server. The handle may be empty, when nobody
// has it. A call that Identify does not name is logged as a Person's.
func (c *Call) Identify(kind ActorKind, userID int64, handle string) {
	if c.Actor.Kind == Visitor {
		c.as, c.Actor.UserID, c.Actor.Handle = kind, userID, handle
	}
}

// Read runs fn in a read-only transaction of its own, for Prepare, which runs
// before the call's transaction.
func (c *Call) Read(ctx context.Context, fn func(tx *sql.Tx) error) error {
	tx, err := c.db.Read.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }() // it only read
	return fn(tx)
}

// Value returns what the registry of the call provides under key, or nil.
func (c *Call) Value(key any) any {
	return c.values[key]
}

// Keep keeps v, from Prepare, for Run.
func (c *Call) Keep(v any) {
	c.kept = v
}

// Kept returns what Prepare kept, or nil.
func (c *Call) Kept() any {
	return c.kept
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
	ErrTooMany          = errors.New("too many attempts")
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
		errors.Is(err, ErrForbidden) || errors.Is(err, ErrApprovalRequired) ||
		errors.Is(err, ErrTooMany)
}

// Registry holds the operations of a Cahut, or of the server, and runs them.
type Registry struct {
	db  *store.DB
	ops map[string]registered

	// Now returns the current time. Tests replace it.
	Now func() time.Time

	// CheckPerson reads, in the transaction of a call, whether a person
	// administers the Cahut. It returns an error wrapping ErrForbidden when
	// the person may not act at all, as when they are disabled. Each Cahut
	// sets it; the registry of the server, which people do not call, does
	// not.
	CheckPerson func(ctx context.Context, tx *sql.Tx, userID int64) (admin bool, err error)

	// Individual reads, in the transaction of a call, whether the Cahut is
	// individual, for the operations whose InIndividual is narrower than
	// their Level. Each Cahut sets it; the registry of the server does not.
	Individual func(ctx context.Context, tx *sql.Tx) (bool, error)

	// Knock records, in tx, the Knock that a call of an AI makes, or returns
	// the one that already waits for the same call. Each Cahut sets it; a
	// registry without it refuses the calls that knock. See
	// docs/adr/0021-ais-knock-first.md.
	Knock func(ctx context.Context, tx *sql.Tx, req KnockRequest) (*Knocked, error)

	values map[any]any
}

// Provide makes v available to the operations of r, which read it with
// Call.Value(key): a service of the Cahut, such as what sends its e-mail.
// Keys follow the rules of context keys. Provide is called before r serves
// any call.
func (r *Registry) Provide(key, v any) {
	if r.values == nil {
		r.values = map[any]any{}
	}
	r.values[key] = v
}

type registered interface {
	info() Info
	invoke(ctx context.Context, r *Registry, actor Actor, args json.RawMessage) (any, error)
	runJSON(ctx context.Context, c *Call, args json.RawMessage) (any, error)
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
	if op.KnockFirst && op.KnocksWhen != "" {
		panic("operation: " + op.Name + " knocks first always and when " + op.KnocksWhen)
	}
	if op.Prepare != nil && (op.KnockFirst || op.KnocksWhen != "") {
		panic("operation: " + op.Name + " knocks first and has a Prepare, which an approved Knock would skip")
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

// Lookup describes the operation called name, when r has one.
func (r *Registry) Lookup(name string) (Info, bool) {
	op, ok := r.ops[name]
	if !ok {
		return Info{}, false
	}
	return op.info(), true
}

// Do runs op for actor.
func Do[In, Out any](ctx context.Context, r *Registry, op *Op[In, Out], actor Actor, in In) (Out, error) {
	var out Out
	err := r.run(ctx, op.info(), actor, op.prepare(in, nil), func(ctx context.Context, c *Call) error {
		var err error
		c.input = in
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
		Title:        op.Title,
		Description:  op.Description,
		ReadOnly:     op.ReadOnly,
		Level:        op.Level,
		InIndividual: op.InIndividual,
		PeopleOnly:   op.PeopleOnly,
		Visitors:     op.Visitors,
		KnockFirst:   op.KnockFirst,
		KnocksWhen:   op.KnocksWhen,
		ReadOnlyAIs:  op.ReadOnlyAIs,
		OpenWorld:    op.OpenWorld,
		Input:        reflect.TypeFor[In](),
		Output:       reflect.TypeFor[Out](),
	}
}

func (op *Op[In, Out]) invoke(ctx context.Context, r *Registry, actor Actor, args json.RawMessage) (any, error) {
	in, decodeErr := decode[In](args)
	var out any
	err := r.run(ctx, op.info(), actor, op.prepare(in, decodeErr), func(ctx context.Context, c *Call) error {
		c.input = in
		o, err := op.Run(ctx, c, in)
		out = o
		return err
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// prepare returns what runs before the transaction of a call with the
// input in: the error of decoding it, if any, then Prepare.
func (op *Op[In, Out]) prepare(in In, decodeErr error) func(context.Context, *Call) error {
	return func(ctx context.Context, c *Call) error {
		if decodeErr != nil || op.Prepare == nil {
			return decodeErr
		}
		c.input = in
		return op.Prepare(ctx, c, in)
	}
}

// decode decodes the arguments of a call in JSON, refusing unknown fields.
func decode[In any](args json.RawMessage) (In, error) {
	var in In
	if len(bytes.TrimSpace(args)) > 0 {
		dec := json.NewDecoder(bytes.NewReader(args))
		dec.DisallowUnknownFields()
		if err := dec.Decode(&in); err != nil {
			return in, Invalidf("arguments: %v", err)
		}
	}
	return in, nil
}

// runJSON decodes args, refusing unknown fields, and runs the operation in
// the call c.
func (op *Op[In, Out]) runJSON(ctx context.Context, c *Call, args json.RawMessage) (any, error) {
	in, err := decode[In](args)
	if err != nil {
		return nil, err
	}
	c.input = in
	out, err := op.Run(ctx, c, in)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// run checks that actor may call the operation, runs fn in a transaction
// and writes the Logbook. A change and its entry share the transaction.
// Every call by an AI is logged, reads included; reads by the
// administrator or a person are not. A call that knocks first runs to its
// end, to check it, and is rolled back: it makes a Knock instead.
func (r *Registry) run(ctx context.Context, info Info, actor Actor, prepare, fn func(context.Context, *Call) error) error {
	logged := !info.ReadOnly || actor.Kind == AI
	// The precision of the times the store keeps.
	c := &Call{Actor: actor, Now: r.Now().UTC().Truncate(time.Millisecond), db: r.db, values: r.values}
	c.knocks = info.KnockFirst && actor.Kind == AI

	if err := authorize(info, actor); err != nil {
		return r.fail(ctx, info, c, logged, err)
	}
	if err := prepare(ctx, c); err != nil {
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

	err = r.admit(ctx, info, c)
	if err == nil {
		err = fn(ctx, c)
	}
	if err == nil && c.knocks {
		_ = tx.Rollback() // the call only knocks: it must change nothing
		for _, f := range c.onFailure {
			f()
		}
		return r.knock(ctx, info, c)
	}
	if err == nil && logged && !info.ReadOnly && !c.unlogged {
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
	actor := c.Actor.Kind
	if actor == Visitor {
		actor = cmp.Or(c.as, Person)
	}
	e := logbook.Entry{
		At:         c.Now,
		Operation:  info.Name,
		Actor:      string(actor),
		UserID:     c.Actor.UserID,
		UserHandle: c.Actor.Handle,
		TokenID:    c.Actor.TokenID,
		Connection: c.Actor.Conn,
		Client:     c.Actor.Client,
		Via:        c.Actor.Via,
		Target:     c.target,
		Detail:     c.detail,
		RoomID:     c.room,
		KnockID:    c.knock,
		Outcome:    logbook.OK,
	}
	if c.Actor.Kind == Visitor {
		// Where a sign-in came from, which limits count.
		e.Address = c.Actor.Address
	}
	if err != nil {
		e.Outcome = logbook.Failed
		if errors.Is(err, ErrForbidden) || errors.Is(err, ErrApprovalRequired) || errors.Is(err, ErrTooMany) {
			e.Outcome = logbook.Refused
		}
		e.Error = err.Error()
	}
	return e
}

// authorize checks, before the call opens its transaction, what the kind of
// the actor alone decides.
func authorize(info Info, a Actor) error {
	switch a.Kind {
	case Admin:
	case Person:
		if a.UserID == 0 {
			return errors.New("operation: person without a user")
		}
	case AI:
		if a.UserID == 0 || a.TokenID == 0 && a.Conn == 0 {
			return errors.New("operation: AI without a user, or without a token or a connection")
		}
		if info.PeopleOnly {
			return Forbiddenf("%s is for people, not for AIs: it proves who someone is, or serves the console", info.Name)
		}
		if a.ReadOnly && !info.ReadOnly && !info.ReadOnlyAIs {
			return Forbiddenf("%s changes things, and this AI may only read: "+
				"%s can connect it again with write access, in the console of the Cahut", info.Name, a.Handle)
		}
	case Visitor:
		if !info.Visitors {
			return Forbiddenf("%s needs someone signed in", info.Name)
		}
	default:
		return fmt.Errorf("operation: unknown actor kind %q", a.Kind)
	}
	if info.Level == ServerAdmin && a.Kind != Admin {
		return Forbiddenf("%s is for the %s", info.Name, info.Level)
	}
	return nil
}

// admit checks, in the transaction of the call, what the records say: the
// role and the state of the person who acts, or for whom an AI acts. It sets
// c.Admin.
func (r *Registry) admit(ctx context.Context, info Info, c *Call) error {
	switch c.Actor.Kind {
	case Admin:
		c.Admin = true
	case Person, AI:
		if r.CheckPerson == nil {
			return fmt.Errorf("operation: %s for a person, in a registry that knows no people", info.Name)
		}
		admin, err := r.CheckPerson(ctx, c.Tx, c.Actor.UserID)
		if err != nil {
			return err
		}
		c.Admin = admin
	}
	if info.Level == CahutAdmin && !c.Admin {
		return Forbiddenf("%s is for the %s", info.Name, info.Level)
	}
	if info.InIndividual > info.Level && r.Individual != nil {
		individual, err := r.Individual(ctx, c.Tx)
		if err != nil {
			return err
		}
		switch {
		case !individual:
		case info.InIndividual == Nobody:
			return Forbiddenf("%s is not available in this Cahut, which is individual: each Room keeps one member, its owner, "+
				"and people do not see one another", info.Name)
		case info.InIndividual == CahutAdmin && !c.Admin:
			return Forbiddenf("%s is for the administrators of this Cahut, which is individual: people do not see one another", info.Name)
		}
	}
	return nil
}
