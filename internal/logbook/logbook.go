// Package logbook records every action on a Cahut or on its server: which
// AI or person acted, on whose behalf, what they did and with what result.
package logbook

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"cahut.cloud/internal/store"
)

// Outcome is how a call ended.
type Outcome string

// Outcomes of a call.
const (
	OK      Outcome = "ok"
	Knocked Outcome = "knocked" // waiting for a person's approval: Knock first
	Refused Outcome = "refused" // not allowed
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
	TokenID    int64     `json:"token_id,omitempty" jsonschema:"the personal token the AI presented"`
	Connection int64     `json:"connection_id,omitempty" jsonschema:"the connection of the AI, when it connected over OAuth"`
	Client     string    `json:"client,omitempty" jsonschema:"the AI, as its token or its connection, and its software, name it"`
	Via        string    `json:"via" jsonschema:"the adapter the call came through: cli, api, mcp or oauth"`
	Target     string    `json:"target,omitempty" jsonschema:"what the operation acted on"`
	Detail     string    `json:"detail,omitempty"`
	Outcome    Outcome   `json:"outcome" jsonschema:"ok, knocked (waiting for a person's approval), refused or failed"`
	Error      string    `json:"error,omitempty"`
	KnockID    int64     `json:"knock_id,omitempty" jsonschema:"the Knock that the call made, answered or ran for"`
	Address    string    `json:"address,omitempty" jsonschema:"where a call that signs in came from: an IPv4 address, or the /64 of an IPv6 one"`
	RoomID     int64     `json:"-"` // the Room the entry concerns, whose members read it
}

// Execer is satisfied by *sql.DB and *sql.Tx.
type Execer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

// Queryer is satisfied by *sql.DB and *sql.Tx.
type Queryer interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

// The most an entry keeps of what a caller can make long: an AI or a
// visitor chooses the name it acts on and the input an error repeats.
// Beyond, the text is cut, so that a call cannot write rows of megabytes
// that every reading of the Logbook then loads.
const (
	MaxTarget = 512
	MaxDetail = 1024
	MaxError  = 1024
	MaxClient = 512
)

// Write adds e to the Logbook. Pass the transaction of a change, so that the
// change and its entry are committed together.
func Write(ctx context.Context, db Execer, e Entry) error {
	_, err := db.ExecContext(ctx, `
		INSERT INTO logbook (at, operation, actor, user_id, user_handle, token_id,
			client, via, target, detail, outcome, error, room_id, connection_id, knock_id, address)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		store.FormatTime(e.At), e.Operation, e.Actor, nullInt(e.UserID), nullString(Clip(e.UserHandle, MaxTarget)),
		nullInt(e.TokenID), nullString(Clip(e.Client, MaxClient)), e.Via, nullString(Clip(e.Target, MaxTarget)),
		nullString(Clip(e.Detail, MaxDetail)), string(e.Outcome), nullString(Clip(e.Error, MaxError)), nullInt(e.RoomID),
		nullInt(e.Connection), nullInt(e.KnockID), nullString(e.Address))
	if err != nil {
		return fmt.Errorf("write logbook: %w", err)
	}
	return nil
}

// Clip returns s cut to at most n bytes, on a character boundary, with an
// ellipsis when it was longer.
func Clip(s string, n int) string {
	if len(s) <= n {
		return s
	}
	const ellipsis = "…"
	cut := n - len(ellipsis)
	for cut > 0 && !utf8.RuneStart(s[cut]) {
		cut--
	}
	return s[:cut] + ellipsis
}

// List returns the latest entries, newest first.
func List(ctx context.Context, db Queryer, limit int) ([]Entry, error) {
	return list(ctx, db, "1 = 1", limit)
}

// ListOf returns the latest entries that the person userID reads, newest
// first: what they did, what AIs did on their behalf, and what happened in
// their Rooms. Only the Logbook of a Cahut has Rooms.
func ListOf(ctx context.Context, db Queryer, userID int64, limit int) ([]Entry, error) {
	return list(ctx, db, "user_id = ? OR room_id IN (SELECT room_id FROM room_members WHERE user_id = ?)", userID, userID, limit)
}

// ListIn returns the latest entries about the Rooms rooms, newest first.
// When member is not zero, only the Rooms that the person member is in
// count.
func ListIn(ctx context.Context, db Queryer, rooms []int64, member int64, limit int) ([]Entry, error) {
	if len(rooms) == 0 {
		return nil, nil
	}
	args := make([]any, 0, len(rooms)+2)
	for _, id := range rooms {
		args = append(args, id)
	}
	where := "room_id IN (" + strings.Repeat("?, ", len(rooms)-1) + "?)"
	if member != 0 {
		where += " AND room_id IN (SELECT room_id FROM room_members WHERE user_id = ?)"
		args = append(args, member)
	}
	return list(ctx, db, where, append(args, limit)...)
}

// list returns the latest entries that match where, newest first. Its last
// argument is the limit.
func list(ctx context.Context, db Queryer, where string, args ...any) ([]Entry, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id, at, operation, actor, user_id, user_handle, token_id,
			client, via, target, detail, outcome, error, room_id, connection_id, knock_id, address
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
			userID, tokenID, roomID, connection, knock  sql.NullInt64
			handle, client, target, detail, errorString sql.NullString
			address                                     sql.NullString
		)
		if err := rows.Scan(&e.ID, &at, &e.Operation, &e.Actor, &userID, &handle, &tokenID,
			&client, &e.Via, &target, &detail, &outcome, &errorString, &roomID, &connection, &knock, &address); err != nil {
			return nil, err
		}
		if e.At, err = store.ParseTime(at); err != nil {
			return nil, err
		}
		e.UserID, e.TokenID, e.RoomID, e.Connection, e.KnockID = userID.Int64, tokenID.Int64, roomID.Int64, connection.Int64, knock.Int64
		e.UserHandle, e.Client, e.Target = handle.String, client.String, target.String
		e.Detail, e.Error, e.Address = detail.String, errorString.String, address.String
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
