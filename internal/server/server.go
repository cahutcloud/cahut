// Package server keeps the records of the server as a whole: the Cahuts it
// holds, each at its own domain, and the Logbook of the operations on the
// server. It opens each Cahut from its domain, with a database of its own,
// so that nothing reaches the records of another Cahut.
// See docs/adr/0008-several-cahuts-per-server.md.
package server

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"cahut.cloud/internal/account"
	"cahut.cloud/internal/address"
	"cahut.cloud/internal/cahut"
	"cahut.cloud/internal/mail"
	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/store"
)

// Config says where the server keeps its data and how its Cahuts are
// reached.
type Config struct {
	// DataDir holds server.db, and a directory per Cahut in cahuts/.
	DataDir string
	// BaseURL returns the address of the Cahut at domain, such as
	// https://cahut.example.com/. It may be nil where addresses are not
	// needed, as on the command line.
	BaseURL func(domain string) string
	// ConsoleURL returns the address of the console of the Cahut at domain,
	// such as https://cahut.example.com/console/, where people consent to AI
	// clients. It may be nil, as BaseURL may.
	ConsoleURL func(domain string) string
	// UpdateDir is where the updater, run as root, leaves the state of
	// updates (update.DefaultStateDir). Empty, the server does not update
	// itself.
	UpdateDir string
	// MailDev writes the e-mail that no relay sends to DataDir/mail/, for
	// development on one machine.
	MailDev bool
	// Log reports the e-mail that could not be sent. Nil discards.
	Log *slog.Logger
}

// Server is the open server.
type Server struct {
	DB       *store.DB
	Registry *operation.Registry
	// AddDomain declares a domain, which creates its Cahut.
	AddDomain *operation.Op[AddDomainInput, NewCahut]
	// CreateSetupLink gives the link to the first start of a Cahut.
	CreateSetupLink *operation.Op[CreateSetupLinkInput, SetupLink]
	// Mail sends the e-mail of the server and of its Cahuts. See
	// docs/adr/0031-e-mail-through-a-relay.md.
	Mail *mail.Service
	cfg  Config
}

// Open opens the server's records in cfg.DataDir, creating them if needed.
func Open(ctx context.Context, cfg Config) (*Server, error) {
	if err := os.MkdirAll(filepath.Join(cfg.DataDir, "cahuts"), 0o750); err != nil {
		return nil, err
	}
	db, err := store.Open(ctx, filepath.Join(cfg.DataDir, "server.db"), store.ServerSchema)
	if err != nil {
		return nil, fmt.Errorf("open the server's database: %w", err)
	}
	s := &Server{DB: db, Registry: operation.NewRegistry(db), cfg: cfg}
	s.Mail = mail.New(mail.Config{DataDir: cfg.DataDir, Server: db, Dev: cfg.MailDev, Log: cfg.Log})
	s.AddDomain = &operation.Op[AddDomainInput, NewCahut]{
		Name:  "add_domain",
		Title: "Declare a Cahut",
		Description: "Declare a domain on this server, which creates the Cahut at that domain, " +
			"optionally with its first administrator and the link where they choose their password.",
		Level: operation.ServerAdmin,
		Run:   s.addDomain,
	}
	operation.Register(s.Registry, s.AddDomain)
	operation.Register(s.Registry, ListDomains)
	operation.Register(s.Registry, RenameDomain)
	operation.Register(s.Registry, cahut.ListServerLogbook)
	registerAdmins(s.Registry)
	s.registerUpdates(s.Registry)
	s.registerSetup(s.Registry)
	s.Mail.RegisterServer(s.Registry)
	return s, nil
}

// mailDrain bounds the time Close gives the e-mail in the queue to go out.
const mailDrain = 30 * time.Second

// Close sends the e-mail that waits in the queue, for a while, then closes
// the server's database.
func (s *Server) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), mailDrain)
	defer cancel()
	return errors.Join(s.Mail.Close(ctx), s.DB.Close())
}

// Info describes a Cahut of the server.
type Info struct {
	ID        int64     `json:"id"`
	Domain    string    `json:"domain"`
	CreatedAt time.Time `json:"created_at"`
}

// AddDomainInput is the input of add_domain.
type AddDomainInput struct {
	Domain string      `json:"domain" jsonschema:"the domain of the new Cahut, such as cahut.example.com, or the public IPv4 address of the server, for a Cahut without a domain"`
	Admin  *FirstAdmin `json:"admin,omitempty" jsonschema:"the first administrator of the Cahut, a person of it"`
}

// FirstAdmin is the first administrator of a new Cahut.
type FirstAdmin struct {
	Handle string `json:"handle"`
	Name   string `json:"name"`
}

// NewCahut is the output of add_domain.
type NewCahut struct {
	Info
	Admin *account.NewUser `json:"admin,omitempty" jsonschema:"the first administrator of the Cahut, with their activation link"`
}

// cleanAddress returns the address of a Cahut as it is kept: in lowercase,
// without a final dot.
func cleanAddress(s string) string {
	return strings.TrimSuffix(strings.ToLower(strings.TrimSpace(s)), ".")
}

// checkAddress returns an error that says what is wrong with the address
// of a Cahut. See docs/adr/0026-cahut-at-an-ip-address.md.
func checkAddress(given, domain string) error {
	if address.ValidAddress(domain) {
		return nil
	}
	if address.IsIP(domain) {
		return operation.Invalidf("%q is not the public IPv4 address of a server: a Cahut without a domain takes "+
			"the public IPv4 address of its server, not a private, local or IPv6 address", given)
	}
	return operation.Invalidf("%q is not a domain name, such as cahut.example.com, or the public IPv4 address of the server", given)
}

func (s *Server) addDomain(ctx context.Context, c *operation.Call, in AddDomainInput) (NewCahut, error) {
	domain := cleanAddress(in.Domain)
	c.Log(domain, "")
	if err := checkAddress(in.Domain, domain); err != nil {
		return NewCahut{}, err
	}
	n := NewCahut{Info: Info{Domain: domain, CreatedAt: c.Now}}
	err := c.Tx.QueryRowContext(ctx, "INSERT INTO cahuts (domain, created_at) VALUES (?, ?) ON CONFLICT DO NOTHING RETURNING id",
		domain, store.FormatTime(c.Now)).Scan(&n.ID)
	if errors.Is(err, sql.ErrNoRows) {
		return NewCahut{}, operation.Invalidf("%s is already the domain of a Cahut", domain)
	}
	if err != nil {
		return NewCahut{}, err
	}

	// Create the Cahut's database now, so that a failure shows at once.
	dir := s.dir(n.ID)
	c.OnFailure(func() { os.RemoveAll(dir) })
	created, err := cahut.Open(ctx, cahut.Config{Dir: dir, Domain: domain})
	if err != nil {
		return NewCahut{}, err
	}
	if in.Admin != nil {
		// The administrator of the server creates the person, in the
		// Logbook of the new Cahut.
		u, err := operation.Do(ctx, created.Registry, account.CreateUser, c.Actor,
			account.CreateUserInput{Handle: in.Admin.Handle, Name: in.Admin.Name, Admin: true})
		if err != nil {
			created.Close()
			return NewCahut{}, fmt.Errorf("first administrator: %w", err)
		}
		n.Admin = &u
		c.Log(domain, "with "+u.Handle)
	}
	return n, created.Close()
}

// RenameDomainInput is the input of rename_domain.
type RenameDomainInput struct {
	Domain string `json:"domain" jsonschema:"the domain or the address of the Cahut now"`
	To     string `json:"to" jsonschema:"its new domain, such as cahut.example.com, which must point to this server, or the public IPv4 address of the server"`
}

// RenameDomain moves a Cahut to another domain, or from the IP address of
// the server to a domain. Everything in it stays; its old address serves
// nothing any more. See docs/adr/0026-cahut-at-an-ip-address.md.
var RenameDomain = &operation.Op[RenameDomainInput, Info]{
	Name:  "rename_domain",
	Title: "Move a Cahut to a domain",
	Description: "Move a Cahut to another domain, or from the IP address of the server to a domain. " +
		"Its people, Rooms, Windows, house style and Logbook stay; the links that name its old address stop working, " +
		"and AIs connect again at the new one.",
	Level: operation.ServerAdmin,
	Run: func(ctx context.Context, c *operation.Call, in RenameDomainInput) (Info, error) {
		from, to := cleanAddress(in.Domain), cleanAddress(in.To)
		c.Log(from, "to "+to)
		if err := checkAddress(in.To, to); err != nil {
			return Info{}, err
		}
		var info Info
		var created string
		err := c.Tx.QueryRowContext(ctx, "SELECT id, created_at FROM cahuts WHERE domain = ?", from).Scan(&info.ID, &created)
		if errors.Is(err, sql.ErrNoRows) {
			return Info{}, operation.NotFoundf("no Cahut at %s: list_domains lists them", from)
		}
		if err != nil {
			return Info{}, err
		}
		if from == to {
			return Info{}, operation.Invalidf("the Cahut is at %s already", to)
		}
		res, err := c.Tx.ExecContext(ctx, "UPDATE cahuts SET domain = ? WHERE id = ? AND NOT EXISTS (SELECT 1 FROM cahuts WHERE domain = ?)",
			to, info.ID, to)
		if err != nil {
			return Info{}, err
		}
		if n, err := res.RowsAffected(); err != nil || n != 1 {
			return Info{}, errors.Join(err, operation.Invalidf("%s is already the address of a Cahut", to))
		}
		// The sessions of administrators hold at a domain: those at the
		// old one end.
		if _, err := c.Tx.ExecContext(ctx, "DELETE FROM admin_sessions WHERE domain = ?", from); err != nil {
			return Info{}, err
		}
		info.Domain = to
		info.CreatedAt, err = store.ParseTime(created)
		return info, err
	},
}

// ListDomainsInput is the input of list_domains.
type ListDomainsInput struct{}

// DomainList is the output of list_domains.
type DomainList struct {
	Cahuts []Info `json:"cahuts"`
}

// ListDomains lists the Cahuts of the server, in the order they were
// declared.
var ListDomains = &operation.Op[ListDomainsInput, DomainList]{
	Name:        "list_domains",
	Title:       "List Cahuts",
	Description: "List the Cahuts of this server and their domains.",
	Level:       operation.ServerAdmin,
	ReadOnly:    true,
	Run: func(ctx context.Context, c *operation.Call, _ ListDomainsInput) (DomainList, error) {
		rows, err := c.Tx.QueryContext(ctx, "SELECT id, domain, created_at FROM cahuts ORDER BY id")
		if err != nil {
			return DomainList{}, err
		}
		defer rows.Close()
		list := DomainList{Cahuts: []Info{}}
		for rows.Next() {
			var info Info
			var created string
			if err := rows.Scan(&info.ID, &info.Domain, &created); err != nil {
				return DomainList{}, err
			}
			if info.CreatedAt, err = store.ParseTime(created); err != nil {
				return DomainList{}, err
			}
			list.Cahuts = append(list.Cahuts, info)
		}
		return list, rows.Err()
	},
}

// Domains returns the domains of all the Cahuts, in the order they were
// declared.
func (s *Server) Domains(ctx context.Context) ([]string, error) {
	rows, err := s.DB.Read.QueryContext(ctx, "SELECT domain FROM cahuts ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var domains []string
	for rows.Next() {
		var d string
		if err := rows.Scan(&d); err != nil {
			return nil, err
		}
		domains = append(domains, d)
	}
	return domains, rows.Err()
}

// CahutID returns the id of the Cahut at domain, which stays when its
// domain changes, or an error wrapping operation.ErrNotFound.
func (s *Server) CahutID(ctx context.Context, domain string) (int64, error) {
	return s.lookup(ctx, domain)
}

// lookup returns the id of the Cahut at domain, or an error wrapping
// operation.ErrNotFound.
func (s *Server) lookup(ctx context.Context, domain string) (int64, error) {
	if !address.ValidAddress(domain) {
		return 0, operation.NotFoundf("no Cahut at %q", domain)
	}
	var id int64
	err := s.DB.Read.QueryRowContext(ctx, "SELECT id FROM cahuts WHERE domain = ?", domain).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, operation.NotFoundf("no Cahut at %s", domain)
	}
	return id, err
}

// Declared reports whether a Cahut is at domain.
func (s *Server) Declared(ctx context.Context, domain string) (bool, error) {
	_, err := s.lookup(ctx, domain)
	if errors.Is(err, operation.ErrNotFound) {
		return false, nil
	}
	return err == nil, err
}

// OpenCahut opens the Cahut at domain, which the caller closes. It returns
// an error wrapping operation.ErrNotFound when no Cahut is at domain.
func (s *Server) OpenCahut(ctx context.Context, domain string) (*cahut.Cahut, error) {
	id, err := s.lookup(ctx, domain)
	if err != nil {
		return nil, err
	}
	cfg := cahut.Config{Dir: s.dir(id), Domain: domain, Mail: s.Mail}
	if s.cfg.BaseURL != nil {
		cfg.BaseURL = s.cfg.BaseURL(domain)
	}
	if s.cfg.ConsoleURL != nil {
		cfg.ConsoleURL = s.cfg.ConsoleURL(domain)
	}
	return cahut.Open(ctx, cfg)
}

// dir returns the directory of the Cahut with the given id.
func (s *Server) dir(id int64) string {
	return filepath.Join(s.cfg.DataDir, "cahuts", strconv.FormatInt(id, 10))
}
