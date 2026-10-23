package server

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"cahut.cloud/internal/account"
	"cahut.cloud/internal/credential"
	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/store"
)

// The first start of a Cahut: the installation ends with a single-use link,
// where the first person of the Cahut chooses their handle, their name and
// their password, and becomes its administrator, and that of the server
// while it has none. See docs/adr/0025-first-start-in-the-browser.md.

// SetupTTL is how long a link to the first start works.
const SetupTTL = 24 * time.Hour

// CreateSetupLinkInput is the input of create_setup_link.
type CreateSetupLinkInput struct {
	Domain string `json:"domain" jsonschema:"the Cahut, by its domain or its address"`
}

// SetupLink is the output of create_setup_link. Code is shown once and
// never stored.
type SetupLink struct {
	Domain    string    `json:"domain"`
	Code      string    `json:"code" jsonschema:"the first person opens the console of the Cahut at /setup#<code>"`
	ExpiresAt time.Time `json:"expires_at"`
}

// SetupInput is the input of get_setup.
type SetupInput struct {
	Code string `json:"code" jsonschema:"the code of the link to the first start"`
}

// Setup is the output of get_setup.
type Setup struct {
	Domain      string    `json:"domain"`
	ExpiresAt   time.Time `json:"expires_at"`
	ServerAdmin bool      `json:"server_admin" jsonschema:"whether the first person also becomes the administrator of the server, which has none"`
}

// SetUpInput is the input of set_up.
type SetUpInput struct {
	Code     string `json:"code" jsonschema:"the code of the link to the first start"`
	Handle   string `json:"handle"`
	Name     string `json:"name"`
	Password string `json:"password" jsonschema:"of 15 characters at least"`
}

// SetUp is the output of set_up.
type SetUp struct {
	Session     account.Session `json:"session" jsonschema:"the session of the first person, at the console of the Cahut"`
	ServerAdmin bool            `json:"server_admin" jsonschema:"whether they are also the administrator of the server, with the same handle and password"`
}

func (s *Server) registerSetup(r *operation.Registry) {
	s.CreateSetupLink = &operation.Op[CreateSetupLinkInput, SetupLink]{
		Name: "create_setup_link",
		Description: "Give the link to the first start of a Cahut that has nobody yet, single-use and valid for 24 hours, " +
			"where its first person chooses their handle and password. It replaces the previous link.",
		Level:      operation.ServerAdmin,
		PeopleOnly: true,
		Run:        s.createSetupLink,
	}
	operation.Register(r, s.CreateSetupLink)
	operation.Register(r, &operation.Op[SetupInput, Setup]{
		Name:        "get_setup",
		Description: "Say whether a link to the first start still works at this domain.",
		PeopleOnly:  true,
		Visitors:    true,
		ReadOnly:    true,
		Run: func(ctx context.Context, c *operation.Call, in SetupInput) (Setup, error) {
			link, err := s.findSetupLink(ctx, c, in.Code)
			if err != nil {
				return Setup{}, err
			}
			return Setup{Domain: link.domain, ExpiresAt: link.expiresAt, ServerAdmin: link.noAdmin}, nil
		},
	})
	operation.Register(r, &operation.Op[SetUpInput, SetUp]{
		Name: "set_up",
		Description: "Use a link to the first start: create the first person of the Cahut, its administrator, " +
			"and the administrator of the server if it has none, with the same handle and password; then sign the person in.",
		PeopleOnly: true,
		Visitors:   true,
		Run:        s.setUp,
	})
}

func (s *Server) createSetupLink(ctx context.Context, c *operation.Call, in CreateSetupLinkInput) (SetupLink, error) {
	domain := cleanAddress(in.Domain)
	c.Log(domain, "")
	var id int64
	err := c.Tx.QueryRowContext(ctx, "SELECT id FROM cahuts WHERE domain = ?", domain).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return SetupLink{}, operation.NotFoundf("no Cahut at %s: list_domains lists them", domain)
	}
	if err != nil {
		return SetupLink{}, err
	}
	if n, err := s.people(ctx, domain); err != nil || n > 0 {
		return SetupLink{}, errors.Join(err, operation.Invalidf("the Cahut of %s has people already: its first start is done, "+
			"and its administrators add people at its console", domain))
	}
	link := SetupLink{Domain: domain, Code: credential.Token(""), ExpiresAt: c.Now.Add(SetupTTL)}
	_, err = c.Tx.ExecContext(ctx, `
		INSERT INTO setup_links (cahut_id, hash, created_at, expires_at) VALUES (?, ?, ?, ?)
		ON CONFLICT (cahut_id) DO UPDATE SET hash = excluded.hash, created_at = excluded.created_at, expires_at = excluded.expires_at`,
		id, credential.Hash(link.Code), store.FormatTime(c.Now), store.FormatTime(link.ExpiresAt))
	return link, err
}

// people counts the people of the Cahut at domain.
func (s *Server) people(ctx context.Context, domain string) (int, error) {
	ch, err := s.OpenCahut(ctx, domain)
	if err != nil {
		return 0, err
	}
	defer ch.Close()
	list, err := operation.Do(ctx, ch.Registry, account.ListUsers, operation.Actor{Kind: operation.Admin, Via: "setup"}, account.ListUsersInput{})
	return len(list.Users), err
}

type setupLink struct {
	cahutID   int64
	domain    string
	expiresAt time.Time
	noAdmin   bool // the server has no administrator yet
}

// findSetupLink returns the link of code, if it works at the domain of the
// call, which is that of its Cahut.
func (s *Server) findSetupLink(ctx context.Context, c *operation.Call, code string) (setupLink, error) {
	gone := operation.NotFoundf("this link to the first start does not work any more: it was used, it expired, or it is for another address; " +
		"cahut setup, on the server, gives a new one")
	var link setupLink
	var expires string
	err := c.Tx.QueryRowContext(ctx, `
		SELECT l.cahut_id, c.domain, l.expires_at FROM setup_links l JOIN cahuts c ON c.id = l.cahut_id
		WHERE l.hash = ?`, credential.Hash(strings.TrimSpace(code))).Scan(&link.cahutID, &link.domain, &expires)
	if errors.Is(err, sql.ErrNoRows) {
		return setupLink{}, gone
	}
	if err != nil {
		return setupLink{}, err
	}
	if link.expiresAt, err = store.ParseTime(expires); err != nil {
		return setupLink{}, err
	}
	if !c.Now.Before(link.expiresAt) || link.domain != c.Actor.Domain {
		return setupLink{}, gone
	}
	var admins int
	if err := c.Tx.QueryRowContext(ctx, "SELECT count(*) FROM admins").Scan(&admins); err != nil {
		return setupLink{}, err
	}
	link.noAdmin = admins == 0
	return link, nil
}

func (s *Server) setUp(ctx context.Context, c *operation.Call, in SetUpInput) (SetUp, error) {
	handle := strings.ToLower(strings.TrimSpace(in.Handle))
	link, err := s.findSetupLink(ctx, c, in.Code)
	if err != nil {
		return SetUp{}, err
	}
	c.Log(handle, "")
	c.Identify(operation.Person, 0, handle)
	if err := credential.CheckPolicy(in.Password); err != nil {
		return SetUp{}, err
	}
	// The link is taken in the transaction of the call, which holds the
	// database of the server: a second call with the same code waits, then
	// finds it gone.
	if _, err := c.Tx.ExecContext(ctx, "DELETE FROM setup_links WHERE cahut_id = ?", link.cahutID); err != nil {
		return SetUp{}, err
	}
	if link.noAdmin {
		// The administrator of the server is checked first: it is the part
		// that this transaction can undo.
		if !account.ValidHandle(handle) {
			return SetUp{}, operation.Invalidf("handle %q: use 2 to 32 lowercase letters, digits and hyphens, "+
				"starting with a letter and ending with a letter or a digit", in.Handle)
		}
		hash, err := credential.HashPassword(ctx, in.Password)
		if err != nil {
			return SetUp{}, err
		}
		_, err = c.Tx.ExecContext(ctx, "INSERT INTO admins (handle, name, password_hash, created_at) VALUES (?, ?, ?, ?)",
			handle, strings.TrimSpace(in.Name), hash, store.FormatTime(c.Now))
		if err != nil {
			return SetUp{}, err
		}
		c.Log(handle, "with the administrator of the server")
		c.Identify(operation.Admin, 0, handle)
	}

	ch, err := s.OpenCahut(ctx, link.domain)
	if err != nil {
		return SetUp{}, err
	}
	defer ch.Close()
	setup := operation.Actor{Kind: operation.Admin, Via: "setup", Domain: link.domain}
	list, err := operation.Do(ctx, ch.Registry, account.ListUsers, setup, account.ListUsersInput{})
	if err != nil {
		return SetUp{}, err
	}
	if len(list.Users) > 0 {
		return SetUp{}, operation.Invalidf("the Cahut of %s has people already: sign in at its console", link.domain)
	}
	person, err := operation.Do(ctx, ch.Registry, account.CreateUser, setup,
		account.CreateUserInput{Handle: handle, Name: in.Name, Admin: true})
	if err != nil {
		return SetUp{}, err
	}
	visitor := operation.Actor{Kind: operation.Visitor, Via: "setup", Domain: link.domain}
	session, err := operation.Do(ctx, ch.Registry, account.ActivateAccount, visitor,
		account.ActivateInput{Code: person.Activation.Code, Password: in.Password})
	if err != nil {
		return SetUp{}, err
	}
	return SetUp{Session: session, ServerAdmin: link.noAdmin}, nil
}
