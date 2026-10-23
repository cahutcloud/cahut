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
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"cahut.cloud/internal/account"
	"cahut.cloud/internal/address"
	"cahut.cloud/internal/cahut"
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
}

// Server is the open server.
type Server struct {
	DB       *store.DB
	Registry *operation.Registry
	// AddDomain declares a domain, which creates its Cahut.
	AddDomain *operation.Op[AddDomainInput, NewCahut]
	cfg       Config
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
	s.AddDomain = &operation.Op[AddDomainInput, NewCahut]{
		Name: "add_domain",
		Description: "Declare a domain on this server, which creates the Cahut at that domain, " +
			"optionally with its first administrator and the link where they choose their password.",
		Level: operation.ServerAdmin,
		Run:   s.addDomain,
	}
	operation.Register(s.Registry, s.AddDomain)
	operation.Register(s.Registry, ListDomains)
	operation.Register(s.Registry, cahut.ListServerLogbook)
	registerAdmins(s.Registry)
	return s, nil
}

// Close closes the server's database.
func (s *Server) Close() error {
	return s.DB.Close()
}

// Info describes a Cahut of the server.
type Info struct {
	ID        int64     `json:"id"`
	Domain    string    `json:"domain"`
	CreatedAt time.Time `json:"created_at"`
}

// AddDomainInput is the input of add_domain.
type AddDomainInput struct {
	Domain string      `json:"domain" jsonschema:"the domain of the new Cahut, such as cahut.example.com"`
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

func (s *Server) addDomain(ctx context.Context, c *operation.Call, in AddDomainInput) (NewCahut, error) {
	domain := strings.TrimSuffix(strings.ToLower(strings.TrimSpace(in.Domain)), ".")
	c.Log(domain, "")
	if !address.ValidDomain(domain) {
		return NewCahut{}, operation.Invalidf("%q is not a domain name, such as cahut.example.com; IP addresses are not supported", in.Domain)
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

// lookup returns the id of the Cahut at domain, or an error wrapping
// operation.ErrNotFound.
func (s *Server) lookup(ctx context.Context, domain string) (int64, error) {
	if !address.ValidDomain(domain) {
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
	cfg := cahut.Config{Dir: s.dir(id), Domain: domain}
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
