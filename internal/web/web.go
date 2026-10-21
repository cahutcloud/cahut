// Package web serves the Cahuts of a server over HTTP(S). Each Cahut answers
// at its own domain: its MCP endpoint for AIs at /mcp, and its Windows under
// the paths of their Rooms, <domain>/<room>/<window>/.
// See docs/adr/0008-several-cahuts-per-server.md and
// docs/adr/0009-rooms-and-windows-as-paths.md.
package web

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"cahut.cloud/internal/cahut"
	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/server"
	"cahut.cloud/internal/window"
)

// Config says how the Cahuts are reached.
type Config struct {
	// Dev serves plain HTTP on DevAddr, for development on one machine.
	// Browsers resolve localhost and *.localhost to the machine itself.
	Dev     bool
	DevAddr string

	// For HTTPS: the addresses to listen on, and the ACME account.
	HTTPAddr  string // ":80", for ACME challenges and redirects to HTTPS
	HTTPSAddr string // ":443"
	Email     string // contact of the ACME account, optional
	CA        string // ACME directory: Let's Encrypt when empty, its staging for "staging"
	CertDir   string // where certificates and ACME keys are kept
}

// BaseURL returns the address of the Cahut at domain.
func (c Config) BaseURL(domain string) string {
	if c.Dev {
		_, port, _ := net.SplitHostPort(c.DevAddr)
		return "http://" + domain + ":" + port + "/"
	}
	return "https://" + domain + "/"
}

// Server serves the Cahuts of a server.
type Server struct {
	srv *server.Server
	cfg Config
	log *slog.Logger
	mux *http.ServeMux

	// streams is canceled when the server shuts down, to end the
	// standalone streams of MCP sessions.
	streams    context.Context
	endStreams context.CancelFunc

	mu    sync.Mutex
	sites map[string]*site // by domain
}

// site is an open Cahut and its MCP endpoint.
type site struct {
	cahut *cahut.Cahut
	mcp   http.Handler
}

// New returns a web server for the Cahuts of srv.
func New(srv *server.Server, cfg Config, log *slog.Logger) *Server {
	s := &Server{srv: srv, cfg: cfg, log: log, sites: map[string]*site{}}
	s.streams, s.endStreams = context.WithCancel(context.Background())
	s.mux = http.NewServeMux()
	s.mux.HandleFunc("GET /{$}", s.serveRoot)
	s.mux.HandleFunc("/mcp", s.serveMCP)
	s.mux.HandleFunc("GET /-/healthz", serveHealth)
	s.mux.HandleFunc("GET /{room}/{window}", redirectToWindow)
	s.mux.HandleFunc("GET /{room}/{window}/{file...}", s.serveWindow)
	return s
}

// Close ends the MCP streams and closes the Cahuts that the server opened.
func (s *Server) Close() error {
	s.endStreams()
	s.mu.Lock()
	defer s.mu.Unlock()
	var errs []error
	for domain, st := range s.sites {
		errs = append(errs, st.cahut.Close())
		delete(s.sites, domain)
	}
	return errors.Join(errs...)
}

// MaxRequestSize bounds the body of a request: a publication of
// window.MaxSize bytes, in base64, with room for the JSON around it.
const MaxRequestSize = window.MaxSize*4/3 + 1<<20

type siteKey struct{}

func siteOf(r *http.Request) *site {
	return r.Context().Value(siteKey{}).(*site)
}

// ServeHTTP finds the Cahut of the request's host, then routes the request
// by its path.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	st, err := s.site(r.Context(), hostOf(r))
	if errors.Is(err, operation.ErrNotFound) {
		misdirected(w)
		return
	}
	if err != nil {
		s.log.Error("open a Cahut", "host", hostOf(r), "error", err)
		http.Error(w, "Internal error.", http.StatusInternalServerError)
		return
	}
	s.mux.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), siteKey{}, st)))
}

// site returns the open Cahut at domain, opening it on first use.
func (s *Server) site(ctx context.Context, domain string) (*site, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if st, ok := s.sites[domain]; ok {
		return st, nil
	}
	c, err := s.srv.OpenCahut(ctx, domain)
	if err != nil {
		return nil, err
	}
	st := &site{cahut: c, mcp: s.mcpHandler(c)}
	s.sites[domain] = st
	return st, nil
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

func (s *Server) serveRoot(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintf(w, "This is %s, a Cahut. AIs connect to it over MCP at /mcp.\n", siteOf(r).cahut.Domain)
}

// serveMCP passes the request to the MCP endpoint of its Cahut. A GET opens
// the standalone stream of an MCP session, which stays open as long as the
// client does: it ends when the server shuts down, or the shutdown would wait
// for it until its deadline.
func (s *Server) serveMCP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()
		stop := context.AfterFunc(s.streams, cancel)
		defer stop()
		r = r.WithContext(ctx)
	}
	siteOf(r).mcp.ServeHTTP(w, r)
}

func serveHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintln(w, "ok")
}

// Run serves until ctx is done, then shuts down gracefully.
func (s *Server) Run(ctx context.Context) error {
	domains, err := s.srv.Domains(ctx)
	if err != nil {
		return err
	}
	if len(domains) == 0 {
		s.log.Warn("this server holds no Cahut yet: declare one with cahut domain add <domain>")
	}
	if s.cfg.Dev {
		srv := s.httpServer(s.cfg.DevAddr, s)
		s.log.Info("serving over HTTP for development", "address", "http://"+s.cfg.DevAddr+"/", "cahuts", domains)
		return serve(ctx, srv, srv.ListenAndServe)
	}

	magic, issuer, err := s.certificates(ctx, domains)
	if err != nil {
		return err
	}
	https := s.httpServer(s.cfg.HTTPSAddr, s)
	https.TLSConfig = magic.TLSConfig()
	https.TLSConfig.NextProtos = append([]string{"h2", "http/1.1"}, https.TLSConfig.NextProtos...)
	plain := s.httpServer(s.cfg.HTTPAddr, issuer.HTTPChallengeHandler(redirectToHTTPS{srv: s.srv}))

	s.log.Info("serving over HTTPS", "cahuts", domains)
	errs := make(chan error, 2)
	go func() { errs <- serve(ctx, plain, plain.ListenAndServe) }()
	go func() { errs <- serve(ctx, https, func() error { return https.ListenAndServeTLS("", "") }) }()
	return errors.Join(<-errs, <-errs)
}

func (s *Server) httpServer(addr string, h http.Handler) *http.Server {
	srv := &http.Server{
		Addr:              addr,
		Handler:           h,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       5 * time.Minute, // a large publication on a slow link
		IdleTimeout:       2 * time.Minute,
		ErrorLog:          slog.NewLogLogger(s.log.Handler(), slog.LevelWarn),
	}
	// The shutdown waits for the requests in progress, but ends the MCP
	// streams, which would never finish on their own.
	srv.RegisterOnShutdown(s.endStreams)
	return srv
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

// redirectToHTTPS sends requests for the domains of the Cahuts to HTTPS.
type redirectToHTTPS struct{ srv *server.Server }

func (rd redirectToHTTPS) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	host := hostOf(r)
	if ok, err := rd.srv.Declared(r.Context(), host); !ok || err != nil {
		misdirected(w)
		return
	}
	target := "https://" + host + r.URL.RequestURI()
	http.Redirect(w, r, target, http.StatusPermanentRedirect) //nolint:gosec // the host is the domain of a Cahut
}
