// Package logbook records every action on a Cahut or on its server: which
// AI or person acted, on whose behalf, what they did and with what result.
package logbook

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"cahut.cloud/internal/store"
)

// Outcome is how a call ended.
type Outcome string

// Outcomes of a call.
const (
	OK      Outcome = "ok"
	Refused Outcome = "refused" // not allowed, or waiting for a person's approval
	Failed  Outcome = "failed"
)

// Entry is one line of the Logbook.
type Entry struct {
	ID         int64     `json:"id"`
	At         time.Time `json:"at"`
	Operation  string    `json:"operation"`
	Actor      string    `json:"actor" jsonschema:"admin, person or ai"`
	UserID     int64     `json:"user_id,omitempty" jsonschema:"the person, or the person an AI acts for"`
	UserHandle string    `json:"user,omitempty"`
	TokenID    int64     `json:"token_id,omitempty"`
	Client     string    `json:"client,omitempty" jsonschema:"the AI, as its token and its software name it"`
	Via        string    `json:"via" jsonschema:"the adapter the call came through: cli, mcp"`
	Target     string    `json:"target,omitempty" jsonschema:"what the operation acted on"`
	Detail     string    `json:"detail,omitempty"`
	Outcome    Outcome   `json:"outcome" jsonschema:"ok, refused or failed"`
	Error      string    `json:"error,omitempty"`
}

// Execer is satisfied by *sql.DB and *sql.Tx.
type Execer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

// Queryer is satisfied by *sql.DB and *sql.Tx.
type Queryer interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

// Write adds e to the Logbook. Pass the transaction of a change, so that the
// change and its entry are committed together.
func Write(ctx context.Context, db Execer, e Entry) error {
	_, err := db.ExecContext(ctx, `
		INSERT INTO logbook (at, operation, actor, user_id, user_handle, token_id,
			client, via, target, detail, outcome, error)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		store.FormatTime(e.At), e.Operation, e.Actor, nullInt(e.UserID), nullString(e.UserHandle),
		nullInt(e.TokenID), nullString(e.Client), e.Via, nullString(e.Target),
		nullString(e.Detail), string(e.Outcome), nullString(e.Error))
	if err != nil {
		return fmt.Errorf("write logbook: %w", err)
	}
	return nil
}

// List returns the latest entries, newest first.
func List(ctx context.Context, db Queryer, limit int) ([]Entry, error) {
	return list(ctx, db, "1 = 1", limit)
}

// ListOf returns the latest entries about the person userID, newest first:
// what they did, and what AIs did on their behalf.
func ListOf(ctx context.Context, db Queryer, userID int64, limit int) ([]Entry, error) {
	return list(ctx, db, "user_id = ?", userID, limit)
}

// list returns the latest entries that match where, newest first. Its last
// argument is the limit.
func list(ctx context.Context, db Queryer, where string, args ...any) ([]Entry, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id, at, operation, actor, user_id, user_handle, token_id,
			client, via, target, detail, outcome, error
		FROM logbook WHERE `+where+` ORDER BY id DESC LIMIT ?`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []Entry
	for rows.Next() {
		var (
			e                                           Entry
			at, outcome                                 string
			userID, tokenID                             sql.NullInt64
			handle, client, target, detail, errorString sql.NullString
		)
		if err := rows.Scan(&e.ID, &at, &e.Operation, &e.Actor, &userID, &handle, &tokenID,
			&client, &e.Via, &target, &detail, &outcome, &errorString); err != nil {
			return nil, err
		}
		if e.At, err = store.ParseTime(at); err != nil {
			return nil, err
		}
		e.UserID, e.TokenID = userID.Int64, tokenID.Int64
		e.UserHandle, e.Client, e.Target = handle.String, client.String, target.String
		e.Detail, e.Error = detail.String, errorString.String
		e.Outcome = Outcome(outcome)
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

func nullInt(v int64) sql.NullInt64 {
	return sql.NullInt64{Int64: v, Valid: v != 0}
}

func nullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}
