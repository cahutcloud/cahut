package operation

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"cahut.cloud/internal/logbook"
)

// An AI that calls an operation that knocks first does not run it: the
// registry checks the call, rolls it back and records a Knock, which the
// person the AI acts for approves or denies in the console. Approving runs
// the call. See docs/adr/0021-ais-knock-first.md.

// KnockRequest is a call of an AI that knocks first, for the registry's
// Knock to record.
type KnockRequest struct {
	Operation string
	Input     json.RawMessage // the input of the call
	Actor     Actor           // the AI that called
	Target    string          // what the call acts on, as its Logbook entry says
	Detail    string
	Room      int64 // the Room the call concerns, or zero
	At        time.Time
}

// Knocked is the error of a call that knocks first: it waits for a person
// to approve it. Its message tells the AI what to do. It wraps
// ErrApprovalRequired.
type Knocked struct {
	ID        int64
	Operation string
	Person    string // the name of the person who answers
	URL       string // where they answer, in the console
	ExpiresAt time.Time
}

func (k *Knocked) Error() string {
	where := "Give them this link, where they approve or deny it: " + k.URL + "."
	if k.URL == "" {
		where = fmt.Sprintf("They approve or deny it among the Knocks of the console of the Cahut, where it is number %d.", k.ID)
	}
	return fmt.Sprintf("Waiting for %s to approve this call of %s: it cannot be undone, or it changes what a person may do. "+
		"%s It expires on %s. Call get_knock with knock %d to learn the outcome; calling %s the same way again returns the same Knock.",
		k.Person, k.Operation, where, k.ExpiresAt.UTC().Format("2 January 2006 at 15:04 MST"), k.ID, k.Operation)
}

func (k *Knocked) Unwrap() error {
	return ErrApprovalRequired
}

// Knock makes the call knock first when an AI makes it, for an operation
// that knocks for some calls only (KnocksWhen). It reports whether the call
// only knocks: once Run returns, the registry rolls the call back and
// records a Knock. Run may return at once, or go on to check what the AI may
// learn before a person answers. For a person, and for a call that a person
// approved, Knock does nothing and reports false.
func (c *Call) Knock() bool {
	if c.Actor.Kind != AI || c.approved != 0 {
		return false
	}
	c.knocks = true
	return true
}

// Approved reports whether the call runs because a person approved the
// Knock of an AI. Its output then goes to that person, not to the AI.
func (c *Call) Approved() bool {
	return c.approved != 0
}

// ForKnock names the Knock that the call answers, for its Logbook entry.
func (c *Call) ForKnock(id int64) {
	c.knock = id
}

// knock records the Knock of the call c, which knocks first, with its
// Logbook entry, and returns it as the error of the call.
func (r *Registry) knock(ctx context.Context, info Info, c *Call) error {
	if r.Knock == nil {
		return r.fail(ctx, info, c, true, fmt.Errorf("%w: %s needs the approval of a person, which cannot be asked for here",
			ErrApprovalRequired, info.Name))
	}
	input, err := json.Marshal(c.input)
	if err != nil {
		return r.fail(ctx, info, c, true, err)
	}
	tx, err := r.db.Write.BeginTx(ctx, nil)
	if err != nil {
		return r.fail(ctx, info, c, true, err)
	}
	k, err := r.Knock(ctx, tx, KnockRequest{
		Operation: info.Name,
		Input:     input,
		Actor:     c.Actor,
		Target:    c.target,
		Detail:    c.detail,
		Room:      c.room,
		At:        c.Now,
	})
	if err == nil {
		c.knock = k.ID
		e := r.entry(info, c, nil)
		e.Outcome = logbook.Knocked
		err = logbook.Write(ctx, tx, e)
	}
	if err == nil {
		err = tx.Commit()
	}
	if err != nil {
		_ = tx.Rollback() // the error that matters is err
		return r.fail(ctx, info, c, true, err)
	}
	return k
}

// RunApproved runs, in the call c that approves the Knock knock, the call
// that the Knock holds: the operation called name, with its input, for
// actor, the AI that knocked. The registry checks the call again as things
// stand, runs it in a savepoint of c's transaction, and logs it with the
// Knock, in place of c. When the operation fails, its changes are undone and
// its entry says why: the error is returned as failed, and c goes on. err
// is an error of c itself. Only the operation that approves Knocks calls
// RunApproved.
func (r *Registry) RunApproved(ctx context.Context, c *Call, name string, actor Actor, input json.RawMessage, knock int64) (
	out any, failed, err error) {
	op, ok := r.ops[name]
	if !ok {
		return nil, nil, fmt.Errorf("operation: knock %d holds an unknown operation %q", knock, name)
	}
	info := op.info()
	run := &Call{Actor: actor, Now: c.Now, Tx: c.Tx, approved: knock, knock: knock, db: r.db, values: r.values}
	c.unlogged = true

	if _, err := c.Tx.ExecContext(ctx, "SAVEPOINT knock"); err != nil {
		return nil, nil, err
	}
	failed = authorize(info, actor)
	if failed == nil {
		failed = r.admit(ctx, info, run)
	}
	if failed == nil {
		out, failed = op.runJSON(ctx, run, input)
	}
	if failed != nil {
		if _, err := c.Tx.ExecContext(ctx, "ROLLBACK TO knock"); err != nil {
			return nil, nil, err
		}
		for _, f := range run.onFailure {
			f()
		}
		out = nil
	} else {
		c.afterCommit = append(c.afterCommit, run.afterCommit...)
		c.onFailure = append(c.onFailure, run.onFailure...)
	}
	if err := logbook.Write(ctx, c.Tx, r.entry(info, run, failed)); err != nil {
		return nil, nil, err
	}
	if _, err := c.Tx.ExecContext(ctx, "RELEASE knock"); err != nil {
		return nil, nil, err
	}
	return out, failed, nil
}
