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
	"cahut.cloud/internal/console"
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

	// DevConsoleAddr serves the console and its API in plain HTTP, for
	// development, on another port than DevAddr, as in production.
	DevConsoleAddr string

	// For HTTPS: the addresses to listen on, and the ACME account.
	HTTPAddr    string // ":80", for ACME challenges and redirects to HTTPS
	HTTPSAddr   string // ":443"
	ConsoleAddr string // ":8443", for the console and its API
	Email       string // contact of the ACME account, optional
	CA          string // ACME directory: Let's Encrypt when empty, its staging for "staging"
	CertDir     string // where certificates and ACME keys are kept
}

// The console of every Cahut listens on its own port of the Cahut's domain:
// ConsolePort, or DevConsolePort for development. See
// docs/adr/0011-console-on-its-own-port.md.
const (
	ConsolePort    = 8443
	DevConsolePort = 8081
)

// ConsoleURL returns the address of the console of the Cahut at domain.
// Only a server for development serves localhost and *.localhost, which no
// certificate authority certifies.
func ConsoleURL(domain string) string {
	if domain == "localhost" || strings.HasSuffix(domain, ".localhost") {
		return fmt.Sprintf("http://%s:%d/", domain, DevConsolePort)
	}
	return fmt.Sprintf("https://%s:%d/", domain, ConsolePort)
}

// ActivationURL is the address where a person of the Cahut at domain uses
// an activation link: the code follows it.
func ActivationURL(domain string) string {
	return ConsoleURL(domain) + "activate#"
}

// AdminActivationURL is the address where an administrator of the server
// uses an activation link, at the console of the Cahut at domain: the code
// follows it.
func AdminActivationURL(domain string) string {
	return ConsoleURL(domain) + "server/activate#"
}

// BaseURL returns the address of the Cahut at domain.
func (c Config) BaseURL(domain string) string {
	if c.Dev {
		_, port, _ := net.SplitHostPort(c.DevAddr)
		return "http://" + domain + ":" + port + "/"
	}
	return "https://" + domain + "/"
}

// ConsoleURL returns the address of the console of the Cahut at domain, as
// this server serves it.
func (c Config) ConsoleURL(domain string) string {
	if c.Dev {
		_, port, _ := net.SplitHostPort(c.DevConsoleAddr)
		return "http://" + domain + ":" + port + "/"
	}
	return ConsoleURL(domain)
}

// Server serves the Cahuts of a server.
type Server struct {
	srv *server.Server
	cfg Config
	log *slog.Logger
	mux *http.ServeMux // port 443: Windows and MCP

	console       *http.ServeMux // port 8443: the console, its API and the OAuth authorization server
	visitors      *limiter       // calls that sign in, by address
	registrations *limiter       // registrations of OAuth clients, by address

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
	for _, method := range []string{"GET", "OPTIONS"} {
		s.mux.HandleFunc(method+" /.well-known/oauth-protected-resource", s.serveResourceMetadata)
		s.mux.HandleFunc(method+" /.well-known/oauth-protected-resource/mcp", s.serveResourceMetadata)
	}
	s.mux.HandleFunc("GET /{room}/{window}", redirectToWindow)
	s.mux.HandleFunc("GET /{room}/{window}/{file...}", s.serveWindow)

	s.visitors = newLimiter(SignInEvery, SignInBurst)
	s.registrations = newLimiter(RegisterEvery, RegisterBurst)
	s.console = http.NewServeMux()
	s.console.HandleFunc("/.well-known/oauth-authorization-server", s.serveAuthorizationMetadata)
	s.console.HandleFunc("/oauth/token", s.serveToken)
	s.console.HandleFunc("/oauth/register", s.serveRegister)
	s.console.HandleFunc("/oauth/revoke", s.serveRevoke)
	s.console.HandleFunc("POST /api/server/{operation}", s.serveServerAPI)
	s.console.HandleFunc("POST /api/{operation}", s.serveAPI)
	s.console.HandleFunc("/api/", refuseAPI)
	s.console.Handle("/", console.Handler())
	return s
}

// Console returns the handler of the console's port, which serves the
// console and its API at the domain of every Cahut. See
// docs/adr/0011-console-on-its-own-port.md.
func (s *Server) Console() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.route(w, r, s.console)
	})
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
// by its path. A host that no Cahut declares gets 421: this is also the
// guard against DNS rebinding, since a page that points its own name at this
// server still sends its own name as the host.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.route(w, r, s.mux)
}

// route serves r with mux, as the Cahut of its host.
func (s *Server) route(w http.ResponseWriter, r *http.Request, mux *http.ServeMux) {
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
	mux.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), siteKey{}, st)))
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
		srv, console := s.httpServer(s.cfg.DevAddr, s), s.httpServer(s.cfg.DevConsoleAddr, s.Console())
		s.log.Info("serving over HTTP for development", "address", "http://"+s.cfg.DevAddr+"/",
			"console", "http://"+s.cfg.DevConsoleAddr+"/", "cahuts", domains)
		return serveAll(ctx,
			func(ctx context.Context) error { return serve(ctx, srv, srv.ListenAndServe) },
			func(ctx context.Context) error { return serve(ctx, console, console.ListenAndServe) })
	}

	magic, issuer, err := s.certificates(ctx, domains)
	if err != nil {
		return err
	}
	https, console := s.httpServer(s.cfg.HTTPSAddr, s), s.httpServer(s.cfg.ConsoleAddr, s.Console())
	for _, srv := range []*http.Server{https, console} {
		srv.TLSConfig = magic.TLSConfig()
		srv.TLSConfig.NextProtos = append([]string{"h2", "http/1.1"}, srv.TLSConfig.NextProtos...)
	}
	plain := s.httpServer(s.cfg.HTTPAddr, issuer.HTTPChallengeHandler(redirectToHTTPS{srv: s.srv}))

	s.log.Info("serving over HTTPS", "cahuts", domains, "console", s.cfg.ConsoleAddr)
	return serveAll(ctx,
		func(ctx context.Context) error { return serve(ctx, plain, plain.ListenAndServe) },
		func(ctx context.Context) error {
			return serve(ctx, https, func() error { return https.ListenAndServeTLS("", "") })
		},
		func(ctx context.Context) error {
			return serve(ctx, console, func() error { return console.ListenAndServeTLS("", "") })
		})
}

// serveAll runs every server until ctx is done or one of them fails, as
// when its port is taken: then it shuts the others down, so that the service
// stops, and says why, instead of running with a port missing.
func serveAll(ctx context.Context, servers ...func(context.Context) error) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	errs := make(chan error, len(servers))
	for _, run := range servers {
		go func() {
			err := run(ctx)
			if err != nil {
				cancel()
			}
			errs <- err
		}()
	}
	var all []error
	for range servers {
		all = append(all, <-errs)
	}
	return errors.Join(all...)
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
