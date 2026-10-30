// Package knock keeps the Knocks of a Cahut. An AI that calls an operation
// that cannot be undone, or that changes what a person may do, does not run
// it: the registry checks the call, rolls it back and has this package
// record a Knock, which waits for the person the AI acts for. That person,
// and no one else, approves it in the console, which runs the call, or
// denies it. See docs/adr/0021-ais-knock-first.md.
package knock

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"cahut.cloud/internal/mail"
	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/store"
)

const (
	// Lifetime is how long a Knock waits for an answer.
	Lifetime = 24 * time.Hour
	// MaxWaiting is how many Knocks of one AI may wait at once.
	MaxWaiting = 20
)

// States of a Knock. A Knock that waits past its expiry has expired, and
// one whose AI lost access before an answer is canceled.
const (
	Waiting  = "waiting"
	Done     = "done"   // approved, and the operation ran
	Failed   = "failed" // approved, and the operation failed
	Denied   = "denied"
	Expired  = "expired"
	Canceled = "canceled"
)

// Knock is a call of an AI that waits, or waited, for its person.
type Knock struct {
	ID         int64          `json:"id"`
	Operation  string         `json:"operation" jsonschema:"the operation that the AI called"`
	Input      map[string]any `json:"input" jsonschema:"the arguments of the call"`
	Target     string         `json:"target,omitempty" jsonschema:"what the call acts on"`
	Detail     string         `json:"detail,omitempty"`
	Room       string         `json:"room,omitempty" jsonschema:"the name of the Room the call concerns, if any"`
	User       string         `json:"user" jsonschema:"handle of the person the AI acts for, who answers the Knock"`
	Client     string         `json:"client" jsonschema:"the AI that called, as its token or its connection, and its software, name it"`
	State      string         `json:"state" jsonschema:"waiting, done (approved and run), failed (approved, and the operation failed), denied, expired or canceled (the AI lost access)"`
	CreatedAt  time.Time      `json:"created_at"`
	ExpiresAt  time.Time      `json:"expires_at" jsonschema:"when the Knock expires, if no one answers it"`
	AnsweredAt *time.Time     `json:"answered_at,omitempty"`
	Error      string         `json:"error,omitempty" jsonschema:"why the operation failed, once approved"`
	URL        string         `json:"url,omitempty" jsonschema:"where the person approves or denies it, in the console"`
}

// Service records and answers the Knocks of a Cahut.
type Service struct {
	reg *operation.Registry
	url func(id int64) string
}

// Register adds the operations on Knocks to r, and has r record its Knocks
// with them. url gives the address of a Knock in the console; it may return
// an empty string, as on the command line.
func Register(r *operation.Registry, url func(id int64) string) {
	s := &Service{reg: r, url: url}
	r.Knock = s.record
	operation.Register(r, &operation.Op[AnswerInput, Answer]{
		Name:  "approve_knock",
		Title: "Approve a Knock",
		Description: "Approve a Knock of one of your AIs: the call runs now, as the AI, " +
			"and what it returns comes to you. It may still fail, if things changed since the AI called.",
		PeopleOnly: true,
		Run:        s.approve,
	})
	operation.Register(r, &operation.Op[AnswerInput, Knock]{
		Name:        "deny_knock",
		Title:       "Deny a Knock",
		Description: "Deny a Knock of one of your AIs: the call never runs. The AI learns it from get_knock.",
		PeopleOnly:  true,
		Run:         s.deny,
	})
	operation.Register(r, &operation.Op[ListInput, List]{
		Name:  "list_knocks",
		Title: "List approval requests (Knocks)",
		Description: "List the Knocks of this Cahut, those that wait first. A Knock is a call of an AI that waits for a person to approve it, " +
			"since it cannot be undone or changes what a person may do. An AI lists the Knocks it made; a person, those of their AIs.",
		ReadOnly: true,
		Run:      s.list,
	})
	operation.Register(r, &operation.Op[GetInput, Knock]{
		Name:  "get_knock",
		Title: "Check an approval request (Knock)",
		Description: "Check a Knock of this Cahut, a call of an AI that waits for a person to approve it: whether it waits, was approved and run, " +
			"failed, was denied, expired or was canceled. Call it to learn what became of a call that knocked.",
		ReadOnly: true,
		Run:      s.get,
	})
}

// record records the Knock of the call req, or finds the one that waits for
// the same call. It is the registry's Knock.
func (s *Service) record(ctx context.Context, tx *sql.Tx, req operation.KnockRequest) (*operation.Knocked, error) {
	a := req.Actor
	at := store.FormatTime(req.At)
	var id int64
	var expires string
	err := tx.QueryRowContext(ctx, `
		SELECT id, expires_at FROM knocks
		WHERE user_id = ? AND token_id IS ? AND connection_id IS ? AND operation = ? AND input = ?
			AND state = 'waiting' AND expires_at > ?
		ORDER BY id DESC LIMIT 1`,
		a.UserID, nullInt(a.TokenID), nullInt(a.Conn), req.Operation, string(req.Input), at).Scan(&id, &expires)
	fresh := errors.Is(err, sql.ErrNoRows)
	if fresh {
		var waiting int
		err = tx.QueryRowContext(ctx, `
			SELECT count(*) FROM knocks
			WHERE user_id = ? AND token_id IS ? AND connection_id IS ? AND state = 'waiting' AND expires_at > ?`,
			a.UserID, nullInt(a.TokenID), nullInt(a.Conn), at).Scan(&waiting)
		if err != nil {
			return nil, err
		}
		if waiting >= MaxWaiting {
			return nil, fmt.Errorf("%w: %d Knocks of this AI already wait for %s: they answer some in the console first",
				operation.ErrTooMany, waiting, a.Handle)
		}
		expires = store.FormatTime(req.At.Add(Lifetime))
		err = tx.QueryRowContext(ctx, `
			INSERT INTO knocks (user_id, token_id, connection_id, client, operation, input, target, detail, room_id,
				created_at, expires_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?) RETURNING id`,
			a.UserID, nullInt(a.TokenID), nullInt(a.Conn), a.Client, req.Operation, string(req.Input),
			nullString(req.Target), nullString(req.Detail), nullInt(req.Room), at, expires).Scan(&id)
	}
	if err != nil {
		return nil, err
	}
	var name string
	if err := tx.QueryRowContext(ctx, "SELECT name FROM users WHERE id = ?", a.UserID).Scan(&name); err != nil {
		return nil, err
	}
	k := &operation.Knocked{ID: id, Operation: req.Operation, Person: name, URL: s.url(id)}
	if k.ExpiresAt, err = store.ParseTime(expires); err != nil {
		return nil, err
	}
	if fresh && k.URL != "" {
		// The person hears of a new Knock by e-mail, once it is recorded.
		o := mail.OutboxIn(s.reg)
		info, _ := s.reg.Lookup(req.Operation)
		l, err := o.Prepare(ctx, tx, a.UserID, mail.Always, knockLetter(a.Client, info.Title, req.Target, k.URL, k.ExpiresAt))
		if err != nil {
			return nil, err
		}
		if l != nil {
			k.Then = func() { o.Send(*l) }
		}
	}
	return k, nil
}

// knockLetter writes the message that tells a person of a new Knock of
// their AI client: what it asks, as the operation names it, and where to
// answer. It holds no text that the AI wrote.
func knockLetter(client, title, target, url string, expires time.Time) func(p mail.Person, name, where string) mail.Letter {
	client = strings.Join(strings.Fields(client), " ")
	if utf8.RuneCountInString(client) > 60 {
		client = string([]rune(client)[:60]) + "…"
	}
	what := title
	if r, n := utf8.DecodeRuneInString(title); n > 0 {
		what = string(unicode.ToLower(r)) + title[n:]
	}
	if target != "" {
		what += ": " + target
	}
	return func(p mail.Person, _, where string) mail.Letter {
		return mail.Letter{
			Subject: client + " is waiting for your answer",
			Lines: []string{"Hello " + p.Name + ",", client + " asks for your approval before it acts for you on " + where + ".",
				"It asks to " + what + "."},
			Action: &mail.Action{Label: "Open the Knock", URL: url},
			After:  []string{"It waits until " + mail.Day(expires) + ". If no one answers, it expires, and nothing happens."},
		}
	}
}

// AnswerInput is the input of approve_knock and deny_knock.
type AnswerInput struct {
	Knock int64 `json:"knock" jsonschema:"the number of the Knock"`
}

// Answer is the output of approve_knock.
type Answer struct {
	Knock  Knock `json:"knock"`
	Output any   `json:"output,omitempty" jsonschema:"what the operation returned, when it ran"`
}

func (s *Service) approve(ctx context.Context, c *operation.Call, in AnswerInput) (Answer, error) {
	k, err := s.answerable(ctx, c, in.Knock)
	if err != nil {
		return Answer{}, err
	}
	ai, err := s.actor(ctx, c, k)
	if err != nil {
		return Answer{}, err
	}
	out, failed, err := s.reg.RunApproved(ctx, c, k.Operation, ai, k.input, k.ID)
	if err != nil {
		return Answer{}, err
	}
	k.State = Done
	if failed != nil {
		k.State, k.Error = Failed, failed.Error()
		if !operation.Public(failed) {
			k.Error = "Internal error. The administrator of the Cahut can find it in the Logbook."
		}
	}
	if err := s.answer(ctx, c, &k.Knock); err != nil {
		return Answer{}, err
	}
	return Answer{Knock: k.Knock, Output: out}, nil
}

func (s *Service) deny(ctx context.Context, c *operation.Call, in AnswerInput) (Knock, error) {
	k, err := s.answerable(ctx, c, in.Knock)
	if err != nil {
		return Knock{}, err
	}
	k.State = Denied
	if err := s.answer(ctx, c, &k.Knock); err != nil {
		return Knock{}, err
	}
	return k.Knock, nil
}

// answerable returns the Knock id of an AI of the person who calls, if it
// waits.
func (s *Service) answerable(ctx context.Context, c *operation.Call, id int64) (row, error) {
	c.Log(fmt.Sprintf("knock %d", id), "")
	c.ForKnock(id)
	if c.Actor.Kind != operation.Person {
		return row{}, operation.Forbiddenf("only the person an AI acts for answers its Knocks, in the console")
	}
	ks, err := s.find(ctx, c, "id = ? AND user_id = ?", 1, id, c.Actor.UserID)
	if err != nil {
		return row{}, err
	}
	if len(ks) == 0 {
		return row{}, operation.NotFoundf("none of your AIs made a Knock %d", id)
	}
	k := ks[0]
	c.Log(fmt.Sprintf("knock %d", id), k.Operation)
	if k.State != Waiting {
		return row{}, operation.Invalidf("knock %d is %s: only a Knock that waits can be answered", id, k.State)
	}
	return k, nil
}

// answer records that the person who calls answered the Knock k, as its
// state says.
func (s *Service) answer(ctx context.Context, c *operation.Call, k *Knock) error {
	now := c.Now
	res, err := c.Tx.ExecContext(ctx, "UPDATE knocks SET state = ?, answered_at = ?, error = ? WHERE id = ? AND state = 'waiting'",
		k.State, store.FormatTime(now), nullString(k.Error), k.ID)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return fmt.Errorf("knock: knock %d was answered meanwhile", k.ID)
	}
	k.AnsweredAt = &now
	return nil
}

// actor returns the AI that made the Knock k, with what it reaches now.
func (s *Service) actor(ctx context.Context, c *operation.Call, k row) (operation.Actor, error) {
	a := operation.Actor{
		Kind:    operation.AI,
		Via:     c.Actor.Via,
		UserID:  c.Actor.UserID,
		Handle:  k.User,
		TokenID: k.token,
		Conn:    k.conn,
		Client:  k.Client,
	}
	if k.conn == 0 {
		return a, nil // a personal token reaches everything its person can
	}
	var scope string
	var all bool
	err := c.Tx.QueryRowContext(ctx, "SELECT scope, all_rooms FROM connections WHERE id = ?", k.conn).Scan(&scope, &all)
	if err != nil {
		return a, err
	}
	a.ReadOnly = scope == "read"
	if all {
		return a, nil
	}
	a.Rooms = []int64{}
	rows, err := c.Tx.QueryContext(ctx, "SELECT room_id FROM connection_rooms WHERE connection_id = ?", k.conn)
	if err != nil {
		return a, err
	}
	defer rows.Close()
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return a, err
		}
		a.Rooms = append(a.Rooms, id)
	}
	return a, rows.Err()
}

// ListInput is the input of list_knocks.
type ListInput struct {
	Waiting bool `json:"waiting,omitempty" jsonschema:"only the Knocks that wait"`
	Limit   int  `json:"limit,omitempty" jsonschema:"how many Knocks, 50 by default, at most 500"`
}

// List is the output of list_knocks.
type List struct {
	Knocks  []Knock `json:"knocks"`
	Waiting int     `json:"waiting" jsonschema:"how many Knocks wait, in all"`
}

func (s *Service) list(ctx context.Context, c *operation.Call, in ListInput) (List, error) {
	limit := in.Limit
	switch {
	case limit == 0:
		limit = 50
	case limit < 0 || limit > 500:
		return List{}, operation.Invalidf("limit %d: use 1 to 500", limit)
	}
	l := List{Knocks: []Knock{}}
	where, args := visible(c.Actor)
	err := c.Tx.QueryRowContext(ctx, "SELECT count(*) FROM ("+knocks+") WHERE state = 'waiting' AND ("+where+")",
		append([]any{store.FormatTime(c.Now)}, args...)...).Scan(&l.Waiting)
	if err != nil {
		return List{}, err
	}
	if in.Waiting {
		where = "state = 'waiting' AND (" + where + ")"
	}
	ks, err := s.find(ctx, c, where, limit, args...)
	if err != nil {
		return List{}, err
	}
	for _, k := range ks {
		l.Knocks = append(l.Knocks, k.Knock)
	}
	return l, nil
}

// GetInput is the input of get_knock.
type GetInput struct {
	Knock int64 `json:"knock" jsonschema:"the number of the Knock"`
}

func (s *Service) get(ctx context.Context, c *operation.Call, in GetInput) (Knock, error) {
	c.Log(fmt.Sprintf("knock %d", in.Knock), "")
	c.ForKnock(in.Knock)
	where, args := visible(c.Actor)
	ks, err := s.find(ctx, c, "id = ? AND ("+where+")", 1, append([]any{in.Knock}, args...)...)
	if err != nil {
		return Knock{}, err
	}
	if len(ks) == 0 {
		if c.Actor.Kind == operation.AI {
			return Knock{}, operation.NotFoundf("you made no Knock %d", in.Knock)
		}
		return Knock{}, operation.NotFoundf("none of your AIs made a Knock %d", in.Knock)
	}
	return ks[0].Knock, nil
}

// visible returns the condition on the Knocks that actor reads: an AI reads
// those it made, a person those of their AIs, and the command line every
// one.
func visible(a operation.Actor) (string, []any) {
	switch {
	case a.CommandLine():
		return "1 = 1", nil
	case a.Kind == operation.AI:
		return "user_id = ? AND token_id IS ? AND connection_id IS ?", []any{a.UserID, nullInt(a.TokenID), nullInt(a.Conn)}
	default:
		return "user_id = ?", []any{a.UserID}
	}
}

// row is a Knock as find reads it, with what approving it needs.
type row struct {
	Knock
	input       json.RawMessage
	token, conn int64
}

// knocks selects every Knock with its state as it stands, for its first
// argument, the current time: a Knock that waits is canceled once its AI
// lost access, and expired after its expiry.
const knocks = `
	SELECT k.id, k.user_id, k.token_id, k.connection_id, k.operation, k.input, k.target, k.detail,
		r.name AS room, u.handle, k.client,
		CASE
			WHEN k.state != 'waiting' THEN k.state
			WHEN t.revoked_at IS NOT NULL OR n.ended_at IS NOT NULL OR u.disabled_at IS NOT NULL THEN 'canceled'
			WHEN k.expires_at <= ? THEN 'expired'
			ELSE 'waiting'
		END AS state,
		k.created_at, k.expires_at, k.answered_at, k.error
	FROM knocks k JOIN users u ON u.id = k.user_id
		LEFT JOIN tokens t ON t.id = k.token_id
		LEFT JOIN connections n ON n.id = k.connection_id
		LEFT JOIN rooms r ON r.id = k.room_id`

// find returns at most limit Knocks that match where, those that wait
// first, then the newest.
func (s *Service) find(ctx context.Context, c *operation.Call, where string, limit int, args ...any) ([]row, error) {
	//nolint:gosec // knocks and where are made of constants of this package; values go as arguments
	rows, err := c.Tx.QueryContext(ctx, `
		SELECT id, operation, input, target, detail, room, handle, client, state,
			created_at, expires_at, answered_at, error, token_id, connection_id
		FROM (`+knocks+`) WHERE `+where+` ORDER BY state = 'waiting' DESC, id DESC LIMIT ?`,
		append(append([]any{store.FormatTime(c.Now)}, args...), limit)...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ks []row
	for rows.Next() {
		var (
			k                                    row
			input, created, expires              string
			target, detail, room, answered, fail sql.NullString
			token, conn                          sql.NullInt64
		)
		if err := rows.Scan(&k.ID, &k.Operation, &input, &target, &detail, &room, &k.User, &k.Client, &k.State,
			&created, &expires, &answered, &fail, &token, &conn); err != nil {
			return nil, err
		}
		k.input, k.token, k.conn = json.RawMessage(input), token.Int64, conn.Int64
		if err := json.Unmarshal(k.input, &k.Input); err != nil {
			return nil, fmt.Errorf("knock: the input of knock %d: %w", k.ID, err)
		}
		k.Target, k.Detail, k.Room, k.Error = target.String, detail.String, room.String, fail.String
		if k.CreatedAt, err = store.ParseTime(created); err != nil {
			return nil, err
		}
		if k.ExpiresAt, err = store.ParseTime(expires); err != nil {
			return nil, err
		}
		if answered.Valid {
			at, err := store.ParseTime(answered.String)
			if err != nil {
				return nil, err
			}
			k.AnsweredAt = &at
		}
		k.URL = s.url(k.ID)
		ks = append(ks, k)
	}
	return ks, rows.Err()
}

func nullInt(v int64) sql.NullInt64 {
	return sql.NullInt64{Int64: v, Valid: v != 0}
}

func nullString(v string) sql.NullString {
	return sql.NullString{String: v, Valid: strings.TrimSpace(v) != ""}
}
