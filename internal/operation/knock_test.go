package operation

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"cahut.cloud/internal/logbook"
)

// knocks records the Knocks of a test registry.
type knocks struct {
	requests []KnockRequest
	err      error
}

func (k *knocks) knock(_ context.Context, _ *sql.Tx, req KnockRequest) (*Knocked, error) {
	if k.err != nil {
		return nil, k.err
	}
	k.requests = append(k.requests, req)
	return &Knocked{ID: int64(len(k.requests)), Operation: req.Operation, Person: "Olivier",
		URL: "https://cahut.test/console/knocks/1", ExpiresAt: req.At.Add(24 * time.Hour)}, nil
}

func knockingRegistry(t *testing.T) (*Registry, *knocks) {
	t.Helper()
	r, _ := testRegistry(t)
	k := &knocks{}
	r.Knock = k.knock
	return r, k
}

func TestKnockFirstChecksAndChangesNothing(t *testing.T) {
	r, db := testRegistry(t)
	k := &knocks{}
	r.Knock = k.knock
	op := createRoom(nil)
	op.KnockFirst = true
	ran, cleaned, committed := false, false, false
	run := op.Run
	op.Run = func(ctx context.Context, c *Call, in roomIn) (roomOut, error) {
		ran = true
		c.OnFailure(func() { cleaned = true })
		c.AfterCommit(func() { committed = true })
		c.InRoom(4)
		return run(ctx, c, in)
	}
	Register(r, op)

	_, err := r.Invoke(context.Background(), "create_room", ai, json.RawMessage(`{"name": "Team"}`))
	var knocked *Knocked
	if !errors.As(err, &knocked) || !errors.Is(err, ErrApprovalRequired) || !Public(err) {
		t.Fatalf("err = %v, want a Knock", err)
	}
	if !strings.Contains(err.Error(), "Waiting for Olivier to approve") || !strings.Contains(err.Error(), knocked.URL) ||
		!strings.Contains(err.Error(), "22 October 2026 at 09:00 UTC") {
		t.Errorf("message = %q", err)
	}
	if !ran || rooms(t, db) != 0 || !cleaned || committed {
		t.Errorf("ran = %v, rooms = %d, OnFailure ran = %v, AfterCommit ran = %v: the call must run, then change nothing",
			ran, rooms(t, db), cleaned, committed)
	}
	if len(k.requests) != 1 {
		t.Fatalf("requests = %+v", k.requests)
	}
	req := k.requests[0]
	if req.Operation != "create_room" || string(req.Input) != `{"name":"Team"}` || req.Target != "Team" ||
		req.Detail != "a detail" || req.Room != 4 || req.Actor.TokenID != ai.TokenID || !req.At.Equal(r.Now()) {
		t.Errorf("request = %+v", req)
	}
	got := entries(t, db)
	if len(got) != 1 || got[0].Outcome != logbook.Knocked || got[0].KnockID != 1 || got[0].Actor != "ai" ||
		got[0].Target != "Team" || got[0].RoomID != 4 || got[0].Error != "" {
		t.Errorf("entries = %+v", got)
	}

	// A person acts at once.
	if _, err := Do(context.Background(), r, op, member, roomIn{Name: "Team"}); err != nil || rooms(t, db) != 1 {
		t.Errorf("a person: err = %v, rooms = %d", err, rooms(t, db))
	}
}

func TestKnockFirstRefusesAWrongCallAtOnce(t *testing.T) {
	r, k := knockingRegistry(t)
	op := createRoom(Invalidf("bad name"))
	op.KnockFirst = true
	Register(r, op)

	_, err := Do(context.Background(), r, op, ai, roomIn{Name: "Team"})
	if !errors.Is(err, ErrInvalid) || errors.Is(err, ErrApprovalRequired) {
		t.Errorf("err = %v, want ErrInvalid", err)
	}
	if len(k.requests) != 0 {
		t.Errorf("a wrong call knocked: %+v", k.requests)
	}
}

func TestKnockThatCannotBeRecorded(t *testing.T) {
	r, k := knockingRegistry(t)
	k.err = errors.Join(ErrTooMany, errors.New("20 Knocks wait"))
	op := createRoom(nil)
	op.KnockFirst = true
	Register(r, op)

	_, err := Do(context.Background(), r, op, ai, roomIn{Name: "Team"})
	if !errors.Is(err, ErrTooMany) {
		t.Fatalf("err = %v, want ErrTooMany", err)
	}
	got := entries(t, r.db)
	if len(got) != 1 || got[0].Outcome != logbook.Refused {
		t.Errorf("entries = %+v", got)
	}
}

func TestKnocksWhen(t *testing.T) {
	r, k := knockingRegistry(t)
	op := createRoom(nil)
	op.KnocksWhen = "the Room is called Public"
	var knocks []bool
	run := op.Run
	op.Run = func(ctx context.Context, c *Call, in roomIn) (roomOut, error) {
		if in.Name == "Public" {
			knocks = append(knocks, c.Knock())
		}
		return run(ctx, c, in)
	}
	Register(r, op)
	ctx := context.Background()

	if _, err := Do(ctx, r, op, ai, roomIn{Name: "Team"}); err != nil || rooms(t, r.db) != 1 {
		t.Errorf("a call that does not knock: err = %v", err)
	}
	if _, err := Do(ctx, r, op, ai, roomIn{Name: "Public"}); !errors.Is(err, ErrApprovalRequired) || rooms(t, r.db) != 1 {
		t.Errorf("a call that knocks: err = %v", err)
	}
	if _, err := Do(ctx, r, op, member, roomIn{Name: "Public"}); err != nil || rooms(t, r.db) != 2 {
		t.Errorf("the same call by a person: err = %v", err)
	}
	if len(knocks) != 2 || !knocks[0] || knocks[1] || len(k.requests) != 1 {
		t.Errorf("Knock() = %v, requests = %d", knocks, len(k.requests))
	}

	defer func() {
		if recover() == nil {
			t.Error("an operation that knocks first always and when registered")
		}
	}()
	Register(r, &Op[struct{}, int]{Name: "both", KnockFirst: true, KnocksWhen: "never"})
}

func TestKnockWithoutKnocks(t *testing.T) {
	r, db := testRegistry(t)
	op := createRoom(nil)
	op.KnockFirst = true
	Register(r, op)

	_, err := Do(context.Background(), r, op, ai, roomIn{Name: "Team"})
	if !errors.Is(err, ErrApprovalRequired) || rooms(t, db) != 0 {
		t.Errorf("err = %v, rooms = %d", err, rooms(t, db))
	}
}

// approval is what RunApproved returned.
type approval struct {
	out    any
	failed error
}

// approve returns an operation that approves the Knock of ai for the
// operation called name, with input, and records what RunApproved returns.
func approve(r *Registry, name, input string, got *approval) *Op[struct{}, string] {
	return &Op[struct{}, string]{
		Name:       "approve",
		PeopleOnly: true,
		Run: func(ctx context.Context, c *Call, _ struct{}) (string, error) {
			// The approval itself changes something, which commits
			// whatever the operation does.
			if _, err := c.Tx.ExecContext(ctx, "INSERT INTO secrets (name, value) VALUES ('answered', x'00')"); err != nil {
				return "", err
			}
			a := ai
			a.Via = c.Actor.Via
			var err error
			got.out, got.failed, err = r.RunApproved(ctx, c, name, a, json.RawMessage(input), 12)
			return "", err
		},
	}
}

func TestRunApproved(t *testing.T) {
	r, _ := knockingRegistry(t)
	op := createRoom(nil)
	op.KnockFirst = true
	var approved, knocks bool
	run := op.Run
	op.Run = func(ctx context.Context, c *Call, in roomIn) (roomOut, error) {
		approved, knocks = c.Approved(), c.Knock()
		return run(ctx, c, in)
	}
	Register(r, op)
	var got approval
	Register(r, approve(r, "create_room", `{"name": "Team"}`, &got))

	if _, err := r.Invoke(context.Background(), "approve", member, nil); err != nil {
		t.Fatalf("approve: %v", err)
	}
	if got.failed != nil || got.out.(roomOut).ID == 0 || rooms(t, r.db) != 1 || !approved || knocks {
		t.Errorf("out = %+v, failed = %v, rooms = %d, Approved() = %v, Knock() = %v", got.out, got.failed, rooms(t, r.db), approved, knocks)
	}
	// The entry of the operation replaces that of the approval.
	e := entries(t, r.db)
	if len(e) != 1 || e[0].Operation != "create_room" || e[0].Actor != "ai" || e[0].TokenID != ai.TokenID ||
		e[0].Via != "api" || e[0].KnockID != 12 || e[0].Outcome != logbook.OK || e[0].Target != "Team" {
		t.Errorf("entries = %+v", e)
	}
}

func TestRunApprovedThatFails(t *testing.T) {
	r, _ := knockingRegistry(t)
	op := createRoom(Invalidf("bad name"))
	op.KnockFirst = true
	cleaned := false
	run := op.Run
	op.Run = func(ctx context.Context, c *Call, in roomIn) (roomOut, error) {
		c.OnFailure(func() { cleaned = true })
		return run(ctx, c, in)
	}
	Register(r, op)
	var got approval
	Register(r, approve(r, "create_room", `{"name": "Team"}`, &got))

	if _, err := r.Invoke(context.Background(), "approve", member, nil); err != nil {
		t.Fatalf("approve: %v", err)
	}
	if !errors.Is(got.failed, ErrInvalid) || got.out != nil || rooms(t, r.db) != 0 || !cleaned {
		t.Errorf("out = %+v, failed = %v, rooms = %d, OnFailure ran = %v", got.out, got.failed, rooms(t, r.db), cleaned)
	}
	// The approval commits, without the changes of the operation.
	var answered int
	if err := r.db.Read.QueryRowContext(context.Background(), "SELECT count(*) FROM secrets WHERE name = 'answered'").Scan(&answered); err != nil || answered != 1 {
		t.Errorf("the approval did not commit: %d, %v", answered, err)
	}
	e := entries(t, r.db)
	if len(e) != 1 || e[0].Operation != "create_room" || e[0].Outcome != logbook.Failed || e[0].KnockID != 12 ||
		e[0].Error != "invalid input: bad name" {
		t.Errorf("entries = %+v", e)
	}
}

func TestRunApprovedChecksAgain(t *testing.T) {
	r, _ := knockingRegistry(t)
	op := createRoom(nil)
	op.KnockFirst = true
	op.Level = CahutAdmin // olivier, whom ai acts for, is not an administrator
	Register(r, op)
	var got approval
	Register(r, approve(r, "create_room", `{"name": "Team"}`, &got))

	if _, err := r.Invoke(context.Background(), "approve", member, nil); err != nil {
		t.Fatalf("approve: %v", err)
	}
	if !errors.Is(got.failed, ErrForbidden) || rooms(t, r.db) != 0 {
		t.Errorf("failed = %v, rooms = %d", got.failed, rooms(t, r.db))
	}
	if e := entries(t, r.db); len(e) != 1 || e[0].Outcome != logbook.Refused || e[0].KnockID != 12 {
		t.Errorf("entries = %+v", e)
	}
}
