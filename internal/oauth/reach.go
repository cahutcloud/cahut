package oauth

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"cahut.cloud/internal/operation"
)

// SetAIReachInput is the input of set_ai_reach.
type SetAIReachInput struct {
	ID       int64    `json:"id,omitempty" jsonschema:"the connection, from list_connected_ais; an AI leaves it out, to set its own"`
	Scope    string   `json:"scope" jsonschema:"read, or write"`
	AllRooms bool     `json:"all_rooms,omitempty" jsonschema:"every Room of the person, including those they join later"`
	Rooms    []string `json:"rooms,omitempty" jsonschema:"the names of the Rooms the AI reaches, when all_rooms is false"`
}

// SetAIReach sets the scope and the Rooms of a connected AI. A person sets
// those of their AIs, at once. An AI sets its own: narrowing acts at once,
// widening knocks first. It is how an AI that may only read asks to write.
// See docs/adr/0021-ais-knock-first.md.
var SetAIReach = &operation.Op[SetAIReachInput, Connection]{
	Name:  "set_ai_reach",
	Title: "Change what an AI reaches",
	Description: "Set what an AI connected over OAuth reaches: read or write, and every Room of its person or some of them. " +
		"A person sets the reach of their AIs. An AI sets its own, leaving out id: narrowing it acts at once, " +
		"and widening it, such as asking for write or for another Room, waits for its person to approve.",
	KnocksWhen:  "it widens the reach of an AI",
	ReadOnlyAIs: true,
	Run: func(ctx context.Context, c *operation.Call, in SetAIReachInput) (Connection, error) {
		id := in.ID
		switch c.Actor.Kind {
		case operation.AI:
			if c.Actor.Conn == 0 {
				return Connection{}, operation.Invalidf("you use a personal token, which reaches everything %s can: there is nothing to set",
					c.Actor.Handle)
			}
			if id != 0 && id != c.Actor.Conn {
				return Connection{}, operation.Forbiddenf("an AI sets its own reach only: leave out id")
			}
			id = c.Actor.Conn
		case operation.Person:
		default:
			return Connection{}, operation.Forbiddenf("only a person, for their AIs, or an AI, for itself, sets the reach of an AI")
		}
		c.Log(fmt.Sprintf("connection %d", id), "")
		list, err := connections(ctx, c.Tx, c.Now, id, "")
		if err != nil {
			return Connection{}, err
		}
		if id <= 0 || len(list) == 0 || list[0].User != c.Actor.Handle {
			return Connection{}, operation.NotFoundf("no connected AI %d among yours", id)
		}
		conn := list[0]
		switch {
		case conn.Status == Ended:
			return Connection{}, operation.Invalidf("the AI %d is disconnected: connect it again instead", id)
		case in.Scope != Read && in.Scope != Write:
			return Connection{}, operation.Invalidf("scope %q: use read or write", in.Scope)
		case in.AllRooms && len(in.Rooms) > 0:
			return Connection{}, operation.Invalidf("give all_rooms, or rooms, not both")
		case !in.AllRooms && len(in.Rooms) == 0:
			return Connection{}, operation.Invalidf("give at least one Room, or every Room with all_rooms")
		}
		reach := "every Room"
		if !in.AllRooms {
			reach = strings.Join(in.Rooms, ", ")
		}
		c.Log(fmt.Sprintf("connection %d of %s", conn.ID, conn.User), fmt.Sprintf("%s, %s, %s", conn.Client, in.Scope, reach))

		// A limited AI learns nothing of the Rooms it names beyond its own:
		// widening stops here, and the approval checks them.
		widens := in.Scope == Write && conn.Scope == Read || in.AllRooms && !conn.AllRooms ||
			!conn.AllRooms && slices.ContainsFunc(in.Rooms, func(name string) bool { return !slices.Contains(conn.Rooms, name) })
		if widens && c.Knock() {
			return conn, nil
		}
		var ids []int64
		if !in.AllRooms {
			if ids, err = roomIDs(ctx, c.Tx, c.Actor.UserID, in.Rooms); err != nil {
				return Connection{}, err
			}
		}
		if _, err := c.Tx.ExecContext(ctx, "UPDATE connections SET scope = ?, all_rooms = ? WHERE id = ?", in.Scope, in.AllRooms, conn.ID); err != nil {
			return Connection{}, err
		}
		if _, err := c.Tx.ExecContext(ctx, "DELETE FROM connection_rooms WHERE connection_id = ?", conn.ID); err != nil {
			return Connection{}, err
		}
		for _, room := range ids {
			if _, err := c.Tx.ExecContext(ctx, "INSERT INTO connection_rooms (connection_id, room_id) VALUES (?, ?)", conn.ID, room); err != nil {
				return Connection{}, err
			}
		}
		list, err = connections(ctx, c.Tx, c.Now, conn.ID, "")
		if err != nil {
			return Connection{}, err
		}
		return list[0], nil
	},
}
