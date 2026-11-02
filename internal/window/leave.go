package window

import (
	"context"
	"database/sql"

	"cahut.cloud/internal/account"
	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/room"
	"cahut.cloud/internal/settings"
)

// A person who deletes their own account takes with them the Rooms that are
// theirs alone, and no other. check_account_deletion says which, and what
// stops them; leaving is the one place that decides, for it and for
// delete_account. See docs/adr/0059-people-delete-their-own-account.md.

// LeavingRoom is a Room of the person who would delete their account.
type LeavingRoom struct {
	Room    string `json:"room" jsonschema:"the name of the Room"`
	Title   string `json:"title"`
	Kind    string `json:"kind" jsonschema:"personal or team"`
	Windows int    `json:"windows" jsonschema:"how many Windows the Room holds, those in its trash included"`
	Members int    `json:"members,omitempty" jsonschema:"for a Room that stops the deletion: how many other members it has"`
}

// CheckAccountDeletionInput is the input of check_account_deletion.
type CheckAccountDeletionInput struct{}

// AccountDeletion is the output of check_account_deletion.
type AccountDeletion struct {
	Allowed   bool          `json:"allowed" jsonschema:"whether the administrators of the Cahut let people delete their own account; when false, nothing else is listed"`
	LastAdmin bool          `json:"last_admin,omitempty" jsonschema:"true when the person is the last administrator of the Cahut who is not disabled, which stops the deletion"`
	Rooms     []LeavingRoom `json:"rooms" jsonschema:"the Rooms that would be deleted: the personal Room, and the team Rooms the person is the only member of"`
	Shared    []LeavingRoom `json:"shared" jsonschema:"the team Rooms that the person is the only owner of, and that have other members: they stop the deletion until the person gives each another owner, or deletes it"`
	Left      []LeavingRoom `json:"left" jsonschema:"the team Rooms that have another owner: they stay, without the person"`
	Drafts    int           `json:"drafts" jsonschema:"how many Windows the person keeps for themself in the Rooms that stay: they would be deleted"`
	Published int           `json:"published" jsonschema:"how many other Windows the person published in the Rooms that stay: they would stay, without an author"`
	Windows   int           `json:"windows" jsonschema:"how many Windows would be deleted in all: those of the Rooms that go, and the drafts"`
}

func (s *Service) checkAccountDeletion(ctx context.Context, c *operation.Call, _ CheckAccountDeletionInput) (AccountDeletion, error) {
	if c.Actor.UserID == 0 {
		return AccountDeletion{}, operation.Invalidf("check_account_deletion acts for a person, and the administrator is not one")
	}
	d, _, err := s.leaving(ctx, c.Tx, c.Actor.UserID)
	return d, err
}

// leaving says what deleting the account of the person userID would do, and
// the ids of the Rooms that would go, in the order of AccountDeletion.Rooms.
// A Cahut that does not let people delete their own account lists nothing.
//
// A team Room goes with its only owner when no one else is in it, as in
// ADR 47. When others are, it stops the deletion: a team does not lose a
// Room because one person left. A team Room that has another owner stays.
func (s *Service) leaving(ctx context.Context, tx *sql.Tx, userID int64) (AccountDeletion, []int64, error) {
	out := AccountDeletion{Rooms: []LeavingRoom{}, Shared: []LeavingRoom{}, Left: []LeavingRoom{}}
	allowed, err := settings.AccountDeletion(ctx, tx)
	if err != nil {
		return AccountDeletion{}, nil, err
	}
	if !allowed {
		return out, nil, nil
	}
	out.Allowed = true
	if out.LastAdmin, err = account.IsLastAdmin(ctx, tx, userID); err != nil {
		return AccountDeletion{}, nil, err
	}

	// The Rooms of the person, sorted as deleteUser sorts the ones it
	// deletes, with the others each holds.
	rows, err := tx.QueryContext(ctx, `
		SELECT r.id, r.name, r.title, r.kind, m.role = 'owner',
			(SELECT count(*) FROM room_members o WHERE o.room_id = r.id AND o.user_id != m.user_id),
			(SELECT count(*) FROM room_members o WHERE o.room_id = r.id AND o.role = 'owner' AND o.user_id != m.user_id),
			(SELECT count(*) FROM windows w WHERE w.room_id = r.id)
		FROM rooms r JOIN room_members m ON m.room_id = r.id AND m.user_id = ?
		ORDER BY r.kind = 'personal' DESC, r.name`, userID)
	if err != nil {
		return AccountDeletion{}, nil, err
	}
	defer rows.Close()
	var ids []int64
	listed := map[int64]bool{} // the Rooms of Rooms and Shared
	for rows.Next() {
		var (
			id                  int64
			l                   LeavingRoom
			owner               bool
			others, otherOwners int
		)
		if err := rows.Scan(&id, &l.Room, &l.Title, &l.Kind, &owner, &others, &otherOwners, &l.Windows); err != nil {
			return AccountDeletion{}, nil, err
		}
		switch {
		case l.Kind == room.Personal, owner && otherOwners == 0 && others == 0:
			out.Rooms = append(out.Rooms, l)
			out.Windows += l.Windows
			ids = append(ids, id)
			listed[id] = true
		case owner && otherOwners == 0:
			l.Members = others
			out.Shared = append(out.Shared, l)
			listed[id] = true
		default:
			out.Left = append(out.Left, l)
		}
	}
	if err := rows.Err(); err != nil {
		return AccountDeletion{}, nil, err
	}
	rows.Close()

	// The Windows of the person in the other Rooms: the drafts they keep for
	// themself go, as no one else has seen them; what they published stays,
	// without an author.
	rows, err = tx.QueryContext(ctx, `
		SELECT room_id, access = 'author', count(*) FROM windows WHERE created_by = ? GROUP BY room_id, access = 'author'`, userID)
	if err != nil {
		return AccountDeletion{}, nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var (
			id    int64
			draft bool
			n     int
		)
		if err := rows.Scan(&id, &draft, &n); err != nil {
			return AccountDeletion{}, nil, err
		}
		switch {
		case listed[id]:
		case draft:
			out.Drafts += n
		default:
			out.Published += n
		}
	}
	if err := rows.Err(); err != nil {
		return AccountDeletion{}, nil, err
	}
	out.Windows += out.Drafts
	return out, ids, nil
}
