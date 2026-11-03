package settings

import (
	"context"

	"cahut.cloud/internal/operation"
)

// Whether the people of a Cahut delete their own account is up to its
// administrators: every Cahut allows it until they turn it off, as an
// organization may keep the Windows of the people who leave. See
// docs/adr/0059-people-delete-their-own-account.md.

// AccountDeletion reports whether the people of the Cahut delete their own
// account.
func AccountDeletion(ctx context.Context, q querier) (bool, error) {
	var allowed bool
	err := q.QueryRowContext(ctx, "SELECT account_deletion FROM settings WHERE id = 1").Scan(&allowed)
	return allowed, err
}

// SetAccountDeletionInput is the input of set_account_deletion.
type SetAccountDeletionInput struct {
	Allowed bool `json:"allowed" jsonschema:"whether the people of this Cahut delete their own account; false turns it off"`
}

// AccountDeletionSet is the output of set_account_deletion.
type AccountDeletionSet struct {
	Allowed bool `json:"allowed" jsonschema:"whether the people of this Cahut delete their own account"`
}

// SetAccountDeletion lets the people of the Cahut delete their own account,
// or not, for its administrators.
var SetAccountDeletion = &operation.Op[SetAccountDeletionInput, AccountDeletionSet]{
	Name:  "set_account_deletion",
	Title: "Let people delete their account",
	Description: "Let the people of this Cahut delete their own account, or not. It is allowed by default. " +
		"When it is off, a person who wants to leave asks an administrator, who deletes the account. " +
		"For the administrators of the Cahut.",
	Level:      operation.CahutAdmin,
	PeopleOnly: true,
	Run: func(ctx context.Context, c *operation.Call, in SetAccountDeletionInput) (AccountDeletionSet, error) {
		was, err := AccountDeletion(ctx, c.Tx)
		if err != nil {
			return AccountDeletionSet{}, err
		}
		if _, err := c.Tx.ExecContext(ctx, "UPDATE settings SET account_deletion = ? WHERE id = 1", in.Allowed); err != nil {
			return AccountDeletionSet{}, err
		}
		c.Log("cahut", "from "+onOff(was)+" to "+onOff(in.Allowed))
		return AccountDeletionSet(in), nil
	},
}

func onOff(on bool) string {
	if on {
		return "on"
	}
	return "off"
}
