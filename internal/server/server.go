// Package server serves a home: the MCP endpoint for AIs at the home's own
// address, and each Window on an origin of its own, under it.
package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	"cahut.cloud/internal/home"
	"cahut.cloud/internal/window"
)

// Config says how a home is reached.
type Config struct {
	// Domain is the home's address, such as home.example.com. Each Window is
	// served at <name>.<Domain>.
	Domain string

	// Dev serves plain HTTP on DevAddr, for development on one machine.
	// Domain is then "localhost": browsers resolve *.localhost to the
	// machine itself.
	Dev     bool
	DevAddr string

	// For HTTPS: the addresses to listen on, and the ACME account.
	HTTPAddr  string // ":80", for ACME challenges and redirects to HTTPS
	HTTPSAddr string // ":443"
	Email     string // contact of the ACME account, optional
	CA        string // ACME directory: Let's Encrypt when empty, its staging for "staging"
	CertDir   string // where certificates and ACME keys are kept
}

// WindowURL returns the function that gives the address of a Window.
func (c Config) WindowURL() func(name string) string {
	if c.Dev {
		_, port, _ := net.SplitHostPort(c.DevAddr)
		return func(name string) string { return "http://" + name + ".localhost:" + port + "/" }
	}
	return func(name string) string { return "https://" + name + "." + c.Domain + "/" }
}

// Server serves a home.
type Server struct {
	home *home.Home
	cfg  Config
	log  *slog.Logger
	mcp  http.Handler
}

// New returns a server for h.
func New(h *home.Home, cfg Config, log *slog.Logger) *Server {
	if cfg.Dev {
		cfg.Domain = "localhost"
	}
	cfg.Domain = strings.ToLower(cfg.Domain)
	s := &Server{home: h, cfg: cfg, log: log}
	s.mcp = s.mcpHandler()
	return s
}

// MaxRequestSize bounds the body of a request: a publication of
// window.MaxSize bytes, in base64, with room for the JSON around it.
const MaxRequestSize = window.MaxSize*4/3 + 1<<20

// ServeHTTP routes a request by its host: the home itself, or a Window.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	host := hostOf(r)
	switch {
	case host == s.cfg.Domain:
		s.serveHome(w, r)
	case strings.HasSuffix(host, "."+s.cfg.Domain):
		s.serveWindow(w, r, strings.TrimSuffix(host, "."+s.cfg.Domain))
	default:
		misdirected(w)
	}
}

// hostOf returns the host a request is for, in lowercase and without port.
func hostOf(r *http.Request) string {
	host := strings.ToLower(r.Host)
	if h, _, err := net.SplitHostPort(host); err == nil {
		host = h
	}
	return host
}

func misdirected(w http.ResponseWriter) {
	http.Error(w, "This address is not served here.", http.StatusMisdirectedRequest)
}

func (s *Server) serveHome(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/mcp":
		s.mcp.ServeHTTP(w, r)
	case "/healthz":
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		fmt.Fprintln(w, "ok")
	case "/":
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		fmt.Fprintln(w, "This is a Cahut home. AIs connect to it over MCP at /mcp.")
	default:
		http.NotFound(w, r)
	}
}

// Run serves until ctx is done, then shuts down gracefully.
func (s *Server) Run(ctx context.Context) error {
	if s.cfg.Dev {
		srv := s.httpServer(s.cfg.DevAddr, s)
		s.log.Info("serving over HTTP for development", "address", "http://"+s.cfg.DevAddr+"/")
		return serve(ctx, srv, srv.ListenAndServe)
	}

	magic, issuer, err := s.certificates(ctx)
	if err != nil {
		return err
	}
	https := s.httpServer(s.cfg.HTTPSAddr, s)
	https.TLSConfig = magic.TLSConfig()
	https.TLSConfig.NextProtos = append([]string{"h2", "http/1.1"}, https.TLSConfig.NextProtos...)
	plain := s.httpServer(s.cfg.HTTPAddr, issuer.HTTPChallengeHandler(redirectToHTTPS{domain: s.cfg.Domain}))

	s.log.Info("serving over HTTPS", "domain", s.cfg.Domain)
	errs := make(chan error, 2)
	go func() { errs <- serve(ctx, plain, plain.ListenAndServe) }()
	go func() { errs <- serve(ctx, https, func() error { return https.ListenAndServeTLS("", "") }) }()
	return errors.Join(<-errs, <-errs)
}

func (s *Server) httpServer(addr string, h http.Handler) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           h,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       5 * time.Minute, // a large publication on a slow link
		IdleTimeout:       2 * time.Minute,
		ErrorLog:          slog.NewLogLogger(s.log.Handler(), slog.LevelWarn),
	}
}

// serve runs listen until ctx is done, then shuts srv down.
func serve(ctx context.Context, srv *http.Server, listen func() error) error {
	errc := make(chan error, 1)
	go func() { errc <- listen() }()
	select {
	case err := <-errc:
		return fmt.Errorf("serve %s: %w", srv.Addr, err)
	case <-ctx.Done():
		shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return srv.Shutdown(shutdown)
	}
}

// redirectToHTTPS sends requests for the home and its Windows to HTTPS.
type redirectToHTTPS struct{ domain string }

func (rd redirectToHTTPS) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	host := hostOf(r)
	if host != rd.domain && !strings.HasSuffix(host, "."+rd.domain) {
		misdirected(w)
		return
	}
	target := "https://" + host + r.URL.RequestURI()
	http.Redirect(w, r, target, http.StatusPermanentRedirect) //nolint:gosec // the host is the home or one of its Windows
}
