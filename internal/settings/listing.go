package settings

import (
	"cmp"
	"context"
	"database/sql"
	"net/url"
	"strings"

	"cahut.cloud/internal/operation"
)

// A Cahut that the directory of connectors of Claude lists says so: the
// console then leads its people to the listing, where they connect Claude
// in one step, rather than to its address. The administrators of the
// server set it, as the legal pages of a Cahut: the host lists the
// service, not its people. See docs/adr/0061-a-guide-to-the-first-ai.md.

// claudeDirectory is where every listing of the directory of connectors of
// Claude starts.
const claudeDirectory = "https://claude.ai/directory/"

// Listings are the listings of a Cahut in the directories of AI clients.
type Listings struct {
	Claude string `json:"claude,omitempty" jsonschema:"the address of the listing of the Cahut in the directory of connectors of Claude, where a person connects it in one step; absent when it has none"`
}

// ReadListings returns the listings of the Cahut.
func ReadListings(ctx context.Context, q querier) (Listings, error) {
	var claude sql.NullString
	err := q.QueryRowContext(ctx, "SELECT claude_listing FROM settings WHERE id = 1").Scan(&claude)
	return Listings{Claude: claude.String}, err
}

// SetListingsInput is the input of set_listings.
type SetListingsInput struct {
	Claude string `json:"claude,omitempty" jsonschema:"the address of the listing of this Cahut in the directory of connectors of Claude, which starts with https://claude.ai/directory/; empty removes it"`
}

// Listing names a listing of set_listings, as its input names it.
type Listing struct {
	Listing string `json:"listing"` // such as claude
}

var badListing = operation.NewRefusalOf[Listing]("listing_invalid", operation.ErrInvalid,
	"{listing}: give the address of the listing of this Cahut in the directory, such as https://claude.ai/directory/my-cahut, or leave it out to remove it")

// SetListings sets the listings of the Cahut in the directories of AI
// clients, for the administrators of the server.
var SetListings = &operation.Op[SetListingsInput, Listings]{
	Name:  "set_listings",
	Title: "Set the listings of the Cahut",
	Description: "Set the address of the listing of this Cahut in the directory of connectors of Claude, which starts with " +
		"https://claude.ai/directory/, or remove it with an empty address. The console then leads people to the listing, " +
		"where they connect Claude in one step. For the administrators of the server.",
	Level: operation.ServerAdmin,
	Run: func(ctx context.Context, c *operation.Call, in SetListingsInput) (Listings, error) {
		c.Log("listings", "")
		claude := strings.TrimSpace(in.Claude)
		if claude != "" && !validListing(claude, claudeDirectory) {
			return Listings{}, badListing(Listing{Listing: "claude"})
		}
		was, err := ReadListings(ctx, c.Tx)
		if err != nil {
			return Listings{}, err
		}
		_, err = c.Tx.ExecContext(ctx, "UPDATE settings SET claude_listing = ? WHERE id = 1", sql.NullString{String: claude, Valid: claude != ""})
		if err != nil {
			return Listings{}, err
		}
		if was.Claude == claude {
			c.Log("listings", "unchanged")
		} else {
			c.Log("listings", "claude "+cmp.Or(was.Claude, "none")+" → "+cmp.Or(claude, "none"))
		}
		return Listings{Claude: claude}, nil
	},
}

// validListing reports whether s is the address of a listing under
// directory: a page of its own below it, in HTTPS, without a query.
func validListing(s, directory string) bool {
	if !validLink(s) || !strings.HasPrefix(s, directory) || len(s) == len(directory) {
		return false
	}
	u, err := url.Parse(s)
	return err == nil && u.RawQuery == "" && u.Fragment == "" && !strings.Contains(u.Path, "..")
}
