package oauth

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/store"
)

// Connection is a connected AI: what a person granted an AI client.
type Connection struct {
	ID           int64      `json:"id"`
	User         string     `json:"user" jsonschema:"the person the AI acts for"`
	Client       string     `json:"client" jsonschema:"the name of the AI client"`
	ClientID     string     `json:"client_id"`
	RedirectHost string     `json:"redirect_host" jsonschema:"where its authorization went back to"`
	Scope        string     `json:"scope" jsonschema:"read, or write"`
	AllRooms     bool       `json:"all_rooms" jsonschema:"every Room of the person, including those they join later"`
	Rooms        []string   `json:"rooms" jsonschema:"the Rooms granted, when all_rooms is false"`
	Status       string     `json:"status" jsonschema:"active, expired after 30 days without use, or ended"`
	CreatedAt    time.Time  `json:"created_at"`
	LastUsedAt   *time.Time `json:"last_used_at,omitempty"`
	EndedAt      *time.Time `json:"ended_at,omitempty"`
	EndReason    string     `json:"end_reason,omitempty"`
}

// Statuses of a connection.
const (
	Active  = "active"
	Expired = "expired" // its refresh token expired: 30 days without use
	Ended   = "ended"
)

// ListConnectedAIsInput is the input of list_connected_ais.
type ListConnectedAIsInput struct {
	User string `json:"user,omitempty" jsonschema:"only the AIs of this person; for a person who is not an administrator, their own"`
}

// ConnectedAIs is the output of list_connected_ais.
type ConnectedAIs struct {
	Connections []Connection `json:"connections"`
}

// ListConnectedAIs lists the AIs connected over OAuth, newest first: all
// of them for an administrator, their own for anyone else.
var ListConnectedAIs = &operation.Op[ListConnectedAIsInput, ConnectedAIs]{
	Name:        "list_connected_ais",
	Description: "List the AIs connected over OAuth, newest first, with the person each acts for, its scope and its Rooms.",
	PeopleOnly:  true,
	ReadOnly:    true,
	Run: func(ctx context.Context, c *operation.Call, in ListConnectedAIsInput) (ConnectedAIs, error) {
		user := in.User
		if !c.Admin {
			if user != "" && user != c.Actor.Handle {
				return ConnectedAIs{}, operation.Forbiddenf("list_connected_ais: only administrators of the Cahut see the AIs of others")
			}
			user = c.Actor.Handle
		}
		list, err := connections(ctx, c.Tx, c.Now, 0, user)
		return ConnectedAIs{Connections: list}, err
	},
}

// connections returns the connection id, or the connections of the person
// with handle user, or all of them when both are empty, newest first.
func connections(ctx context.Context, tx *sql.Tx, now time.Time, id int64, user string) ([]Connection, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT c.id, u.handle, c.client_name, c.client_id, c.redirect_host, c.scope, c.all_rooms,
			c.created_at, c.last_used_at, c.ended_at, c.end_reason,
			EXISTS (SELECT 1 FROM oauth_tokens t WHERE t.connection_id = c.id AND t.kind = 'refresh'
				AND t.replaced_at IS NULL AND t.expires_at > ?)
		FROM connections c JOIN users u ON u.id = c.user_id
		WHERE (? = 0 OR c.id = ?) AND (? = '' OR u.handle = ?)
		ORDER BY c.id DESC`, store.FormatTime(now), id, id, user, user)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []Connection{}
	for rows.Next() {
		var (
			conn            Connection
			created         string
			lastUsed, ended sql.NullString
			reason          sql.NullString
			refreshable     bool
		)
		if err := rows.Scan(&conn.ID, &conn.User, &conn.Client, &conn.ClientID, &conn.RedirectHost, &conn.Scope, &conn.AllRooms,
			&created, &lastUsed, &ended, &reason, &refreshable); err != nil {
			return nil, err
		}
		if conn.CreatedAt, err = store.ParseTime(created); err != nil {
			return nil, err
		}
		if conn.LastUsedAt, err = nullTime(lastUsed); err != nil {
			return nil, err
		}
		if conn.EndedAt, err = nullTime(ended); err != nil {
			return nil, err
		}
		conn.EndReason = reason.String
		switch {
		case ended.Valid:
			conn.Status = Ended
		case refreshable || !lastUsed.Valid && now.Sub(conn.CreatedAt) < CodeLifetime:
			// Active, or waiting for its client to exchange its code.
			conn.Status = Active
		default:
			conn.Status = Expired
		}
		list = append(list, conn)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for i := range list {
		if list[i].Rooms, err = roomNames(ctx, tx, list[i].ID); err != nil {
			return nil, err
		}
	}
	return list, nil
}

func roomNames(ctx context.Context, tx *sql.Tx, connection int64) ([]string, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT r.name FROM connection_rooms cr JOIN rooms r ON r.id = cr.room_id
		WHERE cr.connection_id = ? ORDER BY r.kind = 'team', r.name`, connection)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	names := []string{}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		names = append(names, name)
	}
	return names, rows.Err()
}

func nullTime(s sql.NullString) (*time.Time, error) {
	if !s.Valid {
		return nil, nil
	}
	t, err := store.ParseTime(s.String)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// DisconnectAIInput is the input of disconnect_ai.
type DisconnectAIInput struct {
	ID int64 `json:"id" jsonschema:"the connection, from list_connected_ais"`
}

// DisconnectAI ends a connection: its tokens stop working at once.
var DisconnectAI = &operation.Op[DisconnectAIInput, Connection]{
	Name:        "disconnect_ai",
	Description: "Disconnect an AI connected over OAuth. Its tokens stop working at once; to use it again, the person connects it again.",
	PeopleOnly:  true,
	Run: func(ctx context.Context, c *operation.Call, in DisconnectAIInput) (Connection, error) {
		c.Log(fmt.Sprintf("connection %d", in.ID), "")
		if in.ID <= 0 {
			return Connection{}, operation.NotFoundf("no connected AI %d", in.ID)
		}
		list, err := connections(ctx, c.Tx, c.Now, in.ID, "")
		if err != nil {
			return Connection{}, err
		}
		// The connections of others do not exist for those who may not see
		// them.
		if len(list) == 0 || list[0].User != c.Actor.Handle && !c.Admin {
			return Connection{}, operation.NotFoundf("no connected AI %d", in.ID)
		}
		conn := list[0]
		c.Log(fmt.Sprintf("connection %d of %s", conn.ID, conn.User), conn.Client)
		if conn.Status == Ended {
			return Connection{}, operation.Invalidf("the AI %d is already disconnected", in.ID)
		}
		reason := "disconnected by " + c.Actor.Handle
		if c.Actor.Handle == "" {
			reason = "disconnected by an administrator of the server"
		}
		if err := endIn(ctx, c.Tx, conn.ID, c.Now, reason); err != nil {
			return Connection{}, err
		}
		now := c.Now
		conn.Status, conn.EndedAt, conn.EndReason = Ended, &now, reason
		return conn, nil
	},
}

// endIn ends the connection id in tx and deletes its codes and tokens.
func endIn(ctx context.Context, tx *sql.Tx, id int64, now time.Time, reason string) error {
	res, err := tx.ExecContext(ctx, "UPDATE connections SET ended_at = ?, end_reason = ? WHERE id = ? AND ended_at IS NULL",
		store.FormatTime(now), reason, id)
	if err != nil {
		return err
	}
	if n, err := res.RowsAffected(); err != nil || n == 0 {
		return err
	}
	if _, err := tx.ExecContext(ctx, "DELETE FROM oauth_tokens WHERE connection_id = ?", id); err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, "DELETE FROM oauth_codes WHERE connection_id = ?", id)
	return err
}

type endInput struct {
	Connection int64
	Reason     string
}

// endConnection ends a connection for its client: a revocation, or a code
// or a refresh token that came back. It is not registered: only the
// protocol's endpoints run it, so that the Logbook records it.
var endConnection = &operation.Op[endInput, struct{}]{
	Name: "end_connection",
	Run: func(ctx context.Context, c *operation.Call, in endInput) (struct{}, error) {
		c.Log(fmt.Sprintf("connection %d of %s", in.Connection, c.Actor.Handle), in.Reason)
		return struct{}{}, endIn(ctx, c.Tx, in.Connection, c.Now, in.Reason)
	},
}
