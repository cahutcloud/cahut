package settings

import (
	"cmp"
	"context"
	"database/sql"
	"strings"

	"cahut.cloud/internal/mail"
	"cahut.cloud/internal/operation"
)

// The address of moderation of a Cahut receives each report about its
// Windows, beside its administrators who have an address: an abuse mailbox
// of its host, since the administrators of the server have no e-mail. They
// set it, and it travels with the Cahut, as its legal pages do. See
// docs/adr/0054-reporting-and-withdrawing-a-window.md.

// ModerationAddress returns the address of moderation of the Cahut, or ""
// when it has none.
func ModerationAddress(ctx context.Context, q querier) (string, error) {
	var a sql.NullString
	err := q.QueryRowContext(ctx, "SELECT moderation_email FROM settings WHERE id = 1").Scan(&a)
	return a.String, err
}

// SetModerationAddressInput is the input of set_moderation_address.
type SetModerationAddressInput struct {
	Email string `json:"email,omitempty" jsonschema:"the address that receives each report about the Windows of this Cahut, such as abuse@host.example; empty removes it"`
}

// ModerationAddressSet is the output of set_moderation_address.
type ModerationAddressSet struct {
	Email string `json:"email,omitempty" jsonschema:"the address of moderation of the Cahut; absent when it has none"`
}

// SetModerationAddress sets the address of moderation of the Cahut, for
// the administrators of the server.
var SetModerationAddress = &operation.Op[SetModerationAddressInput, ModerationAddressSet]{
	Name:  "set_moderation_address",
	Title: "Set the address of moderation",
	Description: "Set the address that receives each report about the Windows of this Cahut, beside its administrators, " +
		"such as an abuse mailbox of its host, or remove it with an empty address. It receives them when the Cahut sends e-mail. " +
		"For the administrators of the server.",
	Level: operation.ServerAdmin,
	Run: func(ctx context.Context, c *operation.Call, in SetModerationAddressInput) (ModerationAddressSet, error) {
		email := strings.TrimSpace(in.Email)
		c.Log("cahut", "none")
		if email != "" {
			var err error
			if email, err = mail.Address(email); err != nil {
				return ModerationAddressSet{}, operation.Invalidf("%v", err)
			}
			c.Log("cahut", email)
		}
		was, err := ModerationAddress(ctx, c.Tx)
		if err != nil {
			return ModerationAddressSet{}, err
		}
		_, err = c.Tx.ExecContext(ctx, "UPDATE settings SET moderation_email = ? WHERE id = 1", sql.NullString{String: email, Valid: email != ""})
		if err != nil {
			return ModerationAddressSet{}, err
		}
		c.Log("cahut", "from "+cmp.Or(was, "none")+" to "+cmp.Or(email, "none"))
		return ModerationAddressSet{Email: email}, nil
	},
}
