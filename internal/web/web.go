// Package web serves the Cahuts of a server over HTTP(S). Each Cahut answers
// at its own domain: its MCP endpoint for AIs at /mcp, and its Windows under
// the paths of their Rooms, <domain>/<room>/<window>/.
// See docs/adr/0008-several-cahuts-per-server.md and
// docs/adr/0009-rooms-and-windows-as-paths.md.
package web

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/netip"
	"regexp"
	"strings"
	"sync"
	"time"

	"cahut.cloud/internal/cahut"
	"cahut.cloud/internal/card"
	"cahut.cloud/internal/console"
	"cahut.cloud/internal/notice"
	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/server"
	"cahut.cloud/internal/settings"
	"cahut.cloud/internal/update"
	"cahut.cloud/internal/version"
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

	// The proxies whose PROXY protocol Cahut reads on HTTPS and on HTTP:
	// none when empty. See
	// docs/adr/0041-the-proxy-protocol-from-trusted-proxies.md.
	HTTPSProxies, HTTPProxies []netip.Prefix

	// AdminAPIAddr, when set, serves the admin API there, to the programs
	// of the machine itself. See
	// docs/adr/0035-a-program-on-the-server-manages-people-with-a-service-token.md.
	AdminAPIAddr string
}

// DevPort is the port of a server for development, unless it is told
// another one.
const DevPort = 8080

// ConsolePath is where the console of every Cahut is served, on the port
// of its Windows. See docs/adr/0020-console-under-a-path-of-port-443.md.
const ConsolePath = "/console/"

// ConsoleURL returns the address of the console of the Cahut at domain.
// Only a server for development serves localhost and *.localhost, which no
// certificate authority certifies.
func ConsoleURL(domain string) string {
	if domain == "localhost" || strings.HasSuffix(domain, ".localhost") {
		return fmt.Sprintf("http://%s:%d%s", domain, DevPort, ConsolePath)
	}
	return "https://" + domain + ConsolePath
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
	return c.BaseURL(domain) + strings.TrimPrefix(ConsolePath, "/")
}

// Server serves the Cahuts of a server.
type Server struct {
	srv *server.Server
	cfg Config
	log *slog.Logger
	mux *http.ServeMux // port 443: Windows, MCP and OAuth

	// console serves the paths under ConsolePath: the console, its API and
	// the consent screen of OAuth. It has a mux of its own, since its paths
	// would overlap those of the Rooms.
	console       *http.ServeMux
	visitors      *limiter // calls that sign in, by address
	registrations *limiter // registrations of OAuth clients, by address
	tokens        *limiter // calls to the token and revocation endpoints, by address
	mcpSessions   *limiter // MCP sessions opened, by token
	// bigRequests holds a place for each large request to /mcp served at
	// once, such as a publication with images.
	bigRequests chan struct{}
	keyTries    *limiter // Keys typed, by address
	keyFailures *limiter // wrong Keys, by Window
	// wrong Keys, by Window and address
	keyFailuresHere *limiter

	// cards are the cards of Windows drawn last, for the previews of their
	// links.
	cards *card.Cache

	// streams is canceled when the server shuts down, to end the
	// standalone streams of MCP sessions.
	streams    context.Context
	endStreams context.CancelFunc

	mu      sync.Mutex
	sites   map[string]*site // by domain
	proxied []*proxyListener // the listeners that read the PROXY protocol
}

// site is an open Cahut and its MCP endpoint.
type site struct {
	id    int64 // of the Cahut, which stays when its domain changes
	cahut *cahut.Cahut
	mcp   http.Handler
}

// New returns a web server for the Cahuts of srv.
func New(srv *server.Server, cfg Config, log *slog.Logger) *Server {
	s := &Server{srv: srv, cfg: cfg, log: log, sites: map[string]*site{}, cards: card.NewCache(64)}
	// The operations of the server read the Cahuts that this server holds
	// open, rather than open them again.
	srv.Opened = func(ctx context.Context, domain string) (*cahut.Cahut, error) {
		st, err := s.site(ctx, domain)
		if err != nil {
			return nil, err
		}
		return st.cahut, nil
	}
	s.streams, s.endStreams = context.WithCancel(context.Background())
	s.mux = http.NewServeMux()
	s.mux.HandleFunc("GET /{$}", s.serveRoot)
	s.mux.HandleFunc("/mcp", s.serveMCP)
	s.mux.HandleFunc("GET /-/healthz", serveHealth)
	// The OAuth endpoints that the servers of AI clients call are on this
	// port, which every client reaches. See
	// docs/adr/0017-oauth-endpoints-on-port-443.md.
	for _, method := range []string{"GET", "OPTIONS"} {
		s.mux.HandleFunc(method+" /.well-known/oauth-protected-resource", s.logged(s.serveResourceMetadata))
		s.mux.HandleFunc(method+" /.well-known/oauth-protected-resource/mcp", s.logged(s.serveResourceMetadata))
		s.mux.HandleFunc(method+" /.well-known/oauth-authorization-server", s.logged(s.serveAuthorizationMetadata))
		s.mux.HandleFunc(method+" /.well-known/openid-configuration", s.logged(serveUnknownMetadata))
	}
	for _, method := range []string{"POST", "OPTIONS"} {
		s.mux.HandleFunc(method+" /oauth/token", s.logged(s.serveToken))
		s.mux.HandleFunc(method+" /oauth/register", s.logged(s.serveRegister))
		s.mux.HandleFunc(method+" /oauth/revoke", s.logged(s.serveRevoke))
	}
	// The pages of Cahut for readers, and the frames of protected Windows.
	// See docs/adr/0019-window-versions-access-and-declaration.md.
	s.mux.HandleFunc("GET /-/pages.css", serveAsset("pages.css"))
	s.mux.HandleFunc("GET /-/frame.js", serveAsset("frame.js"))
	s.mux.HandleFunc("GET /-/pages.js", serveAsset("pages.js"))
	s.mux.HandleFunc("GET "+BrandPath+"{file...}", serveBrand)
	// The icon of Cahut, where browsers look for that of a site. See
	// docs/adr/0040-the-icon-of-a-cahut.md.
	for p := range rootIcons {
		s.mux.HandleFunc("GET "+p, serveRootIcon)
	}
	// The skill for Claude. See docs/adr/0022-house-style-of-each-cahut.md.
	s.mux.HandleFunc("GET "+SkillPath, s.serveSkill)
	s.mux.HandleFunc("GET /-/sign-in", s.serveSignIn)
	s.mux.HandleFunc("GET /-/session", s.serveSession)
	s.mux.HandleFunc("GET /-/invitation", s.serveInvitation)
	s.mux.HandleFunc("POST /-/sign-out", s.serveSignOut)
	s.mux.HandleFunc("POST /-/key", s.serveKey)
	s.mux.HandleFunc("GET /-/frame/{capability}/{file...}", s.serveFrame)
	s.mux.HandleFunc("GET "+window.PermanentPrefix+"{id}", s.servePermanent)
	s.mux.HandleFunc("GET "+window.PermanentPrefix+"{id}/{file...}", s.servePermanent)
	s.mux.HandleFunc("GET "+CardPath+"{file}", s.serveCard)
	s.mux.HandleFunc("/{room}", s.redirectToRoom)
	s.mux.HandleFunc("GET /{room}/{$}", s.serveRoom)
	s.mux.HandleFunc("GET /{room}/{window}", redirectToWindow)
	s.mux.HandleFunc("GET /{room}/{window}/{file...}", s.serveWindow)

	s.visitors = newLimiter(SignInEvery, SignInBurst)
	s.registrations = newLimiter(RegisterEvery, RegisterBurst)
	s.tokens = newLimiter(TokenEvery, TokenBurst)
	s.mcpSessions = newLimiter(MCPSessionEvery, MCPSessionBurst)
	s.bigRequests = make(chan struct{}, BigRequests)
	s.keyTries = newLimiter(KeyEvery, KeyBurst)
	s.keyFailures = newLimiter(KeyWindowEvery, KeyWindowBurst)
	s.keyFailuresHere = newLimiter(KeyHereEvery, KeyHereBurst)
	// The console, its API and the consent screen of OAuth. See
	// docs/adr/0020-console-under-a-path-of-port-443.md.
	s.console = http.NewServeMux()
	pages := http.StripPrefix(strings.TrimSuffix(ConsolePath, "/"), console.Handler())
	s.console.HandleFunc("GET "+ConsolePath+"oauth/authorize", s.logged(pages.ServeHTTP))
	s.console.HandleFunc("POST "+ConsolePath+"api/server/{operation}", s.serveServerAPI)
	s.console.HandleFunc("POST "+ConsolePath+"api/{operation}", s.serveAPI)
	s.console.HandleFunc(ConsolePath+"api/", refuseAPI)
	s.console.Handle(ConsolePath, pages)
	return s
}

// redirectToConsole sends /console, and /admin, which people type out of
// habit, to the console.
func redirectToConsole(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, ConsolePath, http.StatusMovedPermanently)
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
//
// Every response starts sandboxed: only the handlers of Cahut's own pages
// replace the policy. See docs/adr/0020-console-under-a-path-of-port-443.md.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Security-Policy", defaultPolicy)
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
	r = r.WithContext(context.WithValue(r.Context(), siteKey{}, st))
	switch p := r.URL.Path; {
	case strings.HasPrefix(p, ConsolePath):
		s.console.ServeHTTP(w, r)
	case p == strings.TrimSuffix(ConsolePath, "/"), p == "/admin", strings.HasPrefix(p, "/admin/"):
		redirectToConsole(w, r)
	case !servedWhileSuspended(p) && s.serveSuspended(w, r):
	default:
		s.mux.ServeHTTP(w, r)
	}
}

// defaultPolicy sandboxes what no handler gives a policy of its own: an
// error, a redirect, a path that no route names. A file of a Window served
// by mistake without its policy still gets an opaque origin, and reaches
// neither the reader's session nor the storage of the domain.
const defaultPolicy = "sandbox; default-src 'none'; base-uri 'none'; form-action 'none'"

// site returns the open Cahut at domain, opening it on first use. It reads
// the domains of the server each time, so that a Cahut moved to another
// domain, even from the command line, no longer answers at the old one.
// See docs/adr/0026-cahut-at-an-ip-address.md.
func (s *Server) site(ctx context.Context, domain string) (*site, error) {
	id, err := s.srv.CahutID(ctx, domain)
	s.mu.Lock()
	defer s.mu.Unlock()
	st, ok := s.sites[domain]
	switch {
	case err != nil:
		if ok && errors.Is(err, operation.ErrNotFound) {
			s.retire(domain, st)
		}
		return nil, err
	case ok && st.id == id:
		return st, nil
	}
	// The Cahut is open once: under its old domain, it goes.
	for d, other := range s.sites {
		if other.id == id || d == domain {
			s.retire(d, other)
		}
	}
	c, err := s.srv.OpenCahut(ctx, domain)
	if err != nil {
		return nil, err
	}
	st = &site{id: id, cahut: c, mcp: s.mcpHandler(c)}
	s.sites[domain] = st
	return st, nil
}

// retire forgets the open Cahut at domain, and closes it once the requests
// it serves are likely done. s.mu is held.
func (s *Server) retire(domain string, st *site) {
	delete(s.sites, domain)
	time.AfterFunc(time.Minute, func() {
		if err := st.cahut.Close(); err != nil {
			s.log.Warn("close a Cahut that moved", "cahut", domain, "error", err)
		}
	})
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

// serveMCP passes the request to the MCP endpoint of its Cahut. A GET opens
// the standalone stream of an MCP session, and subscriptions/listen the
// stream of notifications of the protocol of 2026-07-28. Each stays open as
// long as the client does: it ends when the server shuts down, or the
// shutdown would wait for it until its deadline.
func (s *Server) serveMCP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet || r.Header.Get("Mcp-Method") == "subscriptions/listen" {
		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()
		stop := context.AfterFunc(s.streams, cancel)
		defer stop()
		r = r.WithContext(ctx)
	}
	rec := &recorder{ResponseWriter: w, status: http.StatusOK}
	peek := &peeker{ReadCloser: r.Body}
	r.Body = peek
	siteOf(r).mcp.ServeHTTP(rec, r)
	s.logMCP(r, rec, peek)
}

// rpcMethod finds the JSON-RPC method at the start of a request's body.
var rpcMethod = regexp.MustCompile(`"method"\s*:\s*"([^"]{1,64})"`)

// logMCP logs a request to the MCP endpoint: its JSON-RPC method, and how
// Cahut answered. A request without a valid token starts the sign-in of an
// AI client, and the OAuth endpoints log what follows; a refused request
// says why, with the scheme and the prefix of its token, never the token.
// The tool calls themselves go to the Logbook.
func (s *Server) logMCP(r *http.Request, rec *recorder, peek *peeker) {
	attrs := []any{"cahut", hostOf(r), "method", r.Method}
	if peek.head.Len() == 0 && rec.status >= 400 {
		// A request refused before its body was read, such as one without
		// a valid token, still says what it asked for.
		_, _ = io.CopyN(&peek.head, peek.ReadCloser, 512)
	}
	if m := rpcMethod.FindSubmatch(peek.head.Bytes()); m != nil {
		attrs = append(attrs, "rpc", string(m[1]))
	}
	if v := r.Header.Get("Mcp-Protocol-Version"); v != "" {
		attrs = append(attrs, "protocol", v)
	}
	attrs = append(attrs, "status", rec.status, "client", clientIP(r), "user_agent", r.UserAgent())
	if rec.status >= 400 {
		attrs = append(attrs, "credential", credentialKind(r.Header.Get("Authorization")),
			"answer", strings.TrimSpace(rec.body.String()))
	}
	s.log.Info("mcp", attrs...)
}

// credentialKind describes the Authorization header of a request without
// revealing it: its scheme and the prefix of its token, such as
// "Bearer cahut_oat_".
func credentialKind(header string) string {
	if header == "" {
		return "none"
	}
	scheme, token, _ := strings.Cut(header, " ")
	token = strings.TrimSpace(token)
	if i := strings.LastIndex(token, "_"); i > 0 && i < 16 {
		return scheme + " " + token[:i+1]
	}
	return fmt.Sprintf("%s, %d characters", scheme, len(token))
}

// peeker keeps the first bytes of a request's body as it is read.
type peeker struct {
	io.ReadCloser
	head bytes.Buffer
}

func (p *peeker) Read(b []byte) (int, error) {
	n, err := p.ReadCloser.Read(b)
	if room := 512 - p.head.Len(); room > 0 && n > 0 {
		p.head.Write(b[:min(n, room)])
	}
	return n, err
}

func serveHealth(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintln(w, "ok")
}

// localHealth answers /-/healthz on plain HTTP to the machine itself, at a
// loopback address, with the version and the state of every Cahut, and
// passes every other request to next. The updater checks a new version so,
// and cahut doctor reads it. See docs/adr/0024-signed-releases-and-updates.md.
func (s *Server) localHealth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/-/healthz" && fromLoopback(r) {
			s.serveLocalHealth(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// fromLoopback reports whether r comes from the machine itself, for a
// loopback address.
func fromLoopback(r *http.Request) bool {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return false
	}
	remote, to := net.ParseIP(host), net.ParseIP(strings.Trim(hostOf(r), "[]"))
	return remote != nil && remote.IsLoopback() && to != nil && to.IsLoopback()
}

// serveLocalHealth opens every Cahut, which runs its migrations, and says
// how it went.
func (s *Server) serveLocalHealth(w http.ResponseWriter, r *http.Request) {
	h := update.Health{Version: version.Get().Version, OK: true, Cahuts: []update.CahutHealth{}}
	domains, err := s.srv.Domains(r.Context())
	if err != nil {
		h.OK = false
	}
	for _, d := range domains {
		c := update.CahutHealth{Domain: d, OK: true}
		st, err := s.site(r.Context(), d)
		var suspension *settings.Suspension
		if err == nil {
			suspension, err = settings.SuspensionOf(r.Context(), st.cahut.DB.Read)
		}
		if err != nil {
			c.OK, c.Error, h.OK = false, err.Error(), false
		}
		c.Suspended = suspension != nil
		h.Cahuts = append(h.Cahuts, c)
	}
	s.mu.Lock()
	for _, l := range s.proxied {
		h.ProxyProtocol = append(h.ProxyProtocol, l.health())
	}
	s.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	if !h.OK {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	_ = json.NewEncoder(w).Encode(h)
}

// Run serves until ctx is done, then shuts down gracefully.
func (s *Server) Run(ctx context.Context) error {
	var admin []func(context.Context) error
	if addr := s.cfg.AdminAPIAddr; addr != "" {
		if err := CheckAdminAPIAddr(addr); err != nil {
			return err
		}
		srv := s.httpServer(addr, s.AdminHandler())
		admin = append(admin, func(ctx context.Context) error { return serve(ctx, srv, srv.ListenAndServe) })
		s.log.Info("serving the admin API to the machine itself", "address", "http://"+addr+"/api/")
	}
	domains, err := s.srv.Domains(ctx)
	if err != nil {
		return err
	}
	if len(domains) == 0 {
		s.log.Warn("this server holds no Cahut yet: declare one with cahut domain add <domain>")
	}
	go s.watchNotices(ctx)
	if s.cfg.Dev {
		srv := s.httpServer(s.cfg.DevAddr, s.localHealth(s))
		s.log.Info("serving over HTTP for development", "address", "http://"+s.cfg.DevAddr+"/",
			"console", "http://"+s.cfg.DevAddr+ConsolePath, "cahuts", domains)
		return serveAll(ctx, append(admin, func(ctx context.Context) error { return serve(ctx, srv, srv.ListenAndServe) })...)
	}

	c, err := s.certificates(ctx, domains)
	if err != nil {
		return err
	}
	https := s.httpServer(s.cfg.HTTPSAddr, s)
	https.TLSConfig = c.tlsConfig()
	plain := s.httpServer(s.cfg.HTTPAddr, s.localHealth(c.issuer.HTTPChallengeHandler(redirectToHTTPS{srv: s.srv})))
	secure, err := s.listen(ctx, "https", s.cfg.HTTPSAddr, s.cfg.HTTPSProxies)
	if err != nil {
		return err
	}
	insecure, err := s.listen(ctx, "http", s.cfg.HTTPAddr, s.cfg.HTTPProxies)
	if err != nil {
		_ = secure.Close()
		return err
	}

	s.log.Info("serving over HTTPS", "cahuts", domains, "console", ConsolePath)
	return serveAll(ctx, append(admin,
		func(ctx context.Context) error {
			return serve(ctx, plain, func() error { return plain.Serve(insecure) })
		},
		func(ctx context.Context) error {
			return serve(ctx, https, func() error { return https.ServeTLS(secure, "", "") })
		})...)
}

// listen listens on addr for the listener called name, and reads the PROXY
// protocol of the proxies it trusts, when there are some.
func (s *Server) listen(ctx context.Context, name, addr string, proxies []netip.Prefix) (net.Listener, error) {
	if addr == "" {
		addr = ":" + name
	}
	ln, err := (&net.ListenConfig{}).Listen(ctx, "tcp", addr)
	if err != nil || len(proxies) == 0 {
		return ln, err
	}
	l := newProxyListener(ln, name, proxies, proxyHeaderTimeout, s.log)
	s.mu.Lock()
	s.proxied = append(s.proxied, l)
	s.mu.Unlock()
	s.log.Info("reading the PROXY protocol of trusted proxies", "listener", name, "address", ln.Addr().String(), "proxies", proxyNames(proxies))
	return l, nil
}

// watchNotices sends, every hour until ctx is done, the notices that time
// brings: the end of an access, the limits of each Cahut. See
// docs/adr/0031-e-mail-through-a-relay.md.
func (s *Server) watchNotices(ctx context.Context) {
	tick := time.NewTicker(notice.Every)
	defer tick.Stop()
	first := time.After(time.Minute)
	for {
		select {
		case <-ctx.Done():
			return
		case <-first:
		case <-tick.C:
		}
		s.SendNotices(ctx, time.Now())
	}
}

// SendNotices sends the notices that are due at now, for every Cahut.
func (s *Server) SendNotices(ctx context.Context, now time.Time) {
	domains, err := s.srv.Domains(ctx)
	if err != nil {
		s.log.Warn("send notices", "error", err)
		return
	}
	for _, d := range domains {
		st, err := s.site(ctx, d)
		if err == nil {
			err = notice.Cahut(ctx, st.cahut, now)
		}
		if err != nil {
			s.log.Warn("send notices", "cahut", d, "error", err)
		}
	}
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
