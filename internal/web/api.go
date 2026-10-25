package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"cahut.cloud/internal/account"
	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/server"
)

// The HTTP API of the console calls each operation at its own route, under
// the console's path: POST /console/api/<operation> for the operations of
// the Cahut of the request's domain, and POST /console/api/server/<operation>
// for those of the server. It takes session tokens only, reads no cookie and grants no
// CORS access. See docs/adr/0014-console-http-api.md.

// MaxAPIRequestSize bounds the body of a request to the API. The console
// does not publish: Windows come from AIs, over MCP. The calls of visitors,
// who sign in, take a handle, a password or a code: MaxVisitorRequestSize.
const (
	MaxAPIRequestSize     = 1 << 20
	MaxVisitorRequestSize = 16 << 10
)

// The calls that sign in, which take no session, are limited by address: a
// burst of SignInBurst, then one every SignInEvery.
const (
	SignInBurst = 10
	SignInEvery = 6 * time.Second
)

// apiPolicy forbids a response of the API to load or frame anything.
const apiPolicy = "default-src 'none'; frame-ancestors 'none'"

// serveAPI runs an operation of the Cahut of the request.
func (s *Server) serveAPI(w http.ResponseWriter, r *http.Request) {
	s.callAPI(w, r, siteOf(r).cahut.Registry, false)
}

// serveServerAPI runs an operation of the server.
func (s *Server) serveServerAPI(w http.ResponseWriter, r *http.Request) {
	s.callAPI(w, r, s.srv.Registry, true)
}

// refuseAPI answers the requests to the API that are not a POST to an
// operation: the API answers no other method, not even a CORS preflight.
func refuseAPI(w http.ResponseWriter, r *http.Request) {
	setAPIHeaders(w)
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		apiError(w, http.StatusMethodNotAllowed, "the API takes POST requests only, with the input of the operation in JSON")
		return
	}
	apiError(w, http.StatusNotFound, "no such route: call POST /api/<operation> or POST /api/server/<operation>")
}

func setAPIHeaders(w http.ResponseWriter) {
	h := w.Header()
	h.Set("Content-Type", "application/json")
	h.Set("Cache-Control", "no-store")
	h.Set("X-Content-Type-Options", "nosniff")
	h.Set("Content-Security-Policy", apiPolicy)
	h.Set("Referrer-Policy", "no-referrer")
}

// callAPI decodes a call of the API, runs it in reg for the actor that the
// request's session names, and encodes its result. ofServer says whether
// reg is the server's, which only its administrators reach.
func (s *Server) callAPI(w http.ResponseWriter, r *http.Request, reg *operation.Registry, ofServer bool) {
	setAPIHeaders(w)
	name := r.PathValue("operation")
	info, ok := reg.Lookup(name)
	if !ok {
		apiError(w, http.StatusNotFound, fmt.Sprintf("no operation %q: the reference of the operations lists them", name))
		return
	}
	limit := int64(MaxAPIRequestSize)
	if info.Visitors {
		limit = MaxVisitorRequestSize
	}
	body, ok := readCall(w, r, limit)
	if !ok {
		return
	}

	st := siteOf(r)
	actor := operation.Actor{Kind: operation.Visitor, Via: "api", Domain: st.cahut.Domain, Address: clientKey(r)}
	var err error
	if info.Visitors {
		// Signing in takes no session; it is limited by address, on top of
		// the limits by handle that the operation applies.
		if !s.visitors.allow(clientKey(r)) {
			apiError(w, http.StatusTooManyRequests, "too many attempts from this address: wait a minute")
			return
		}
	} else {
		actor, err = s.sessionActor(r, st, ofServer)
		if errors.Is(err, account.ErrInvalidSession) {
			w.Header().Set("WWW-Authenticate", "Bearer")
			apiError(w, http.StatusUnauthorized, err.Error())
			return
		}
		if err != nil {
			s.log.Error("verify a session", "cahut", st.cahut.Domain, "error", err)
			apiError(w, http.StatusInternalServerError, "Internal error.")
			return
		}
	}

	s.invoke(w, r, reg, name, actor, body)
}

// readCall reads the body of a call, of limit bytes at most. When it
// cannot, it answers the request itself and returns false.
func readCall(w http.ResponseWriter, r *http.Request, limit int64) ([]byte, bool) {
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, limit))
	var tooLarge *http.MaxBytesError
	if errors.As(err, &tooLarge) {
		apiError(w, http.StatusRequestEntityTooLarge, fmt.Sprintf("the body holds more than %d KiB", limit>>10))
		return nil, false
	}
	if err != nil {
		apiError(w, http.StatusBadRequest, "the body could not be read")
		return nil, false
	}
	return body, true
}

// invoke runs the operation name of reg for actor, and encodes its result.
func (s *Server) invoke(w http.ResponseWriter, r *http.Request, reg *operation.Registry, name string, actor operation.Actor, body []byte) {
	out, err := reg.Invoke(r.Context(), name, actor, body)
	if err != nil {
		status, msg := apiStatus(err)
		if status == http.StatusInternalServerError {
			s.log.Error("API call failed", "cahut", actor.Domain, "operation", name, "user", actor.Handle,
				"client", actor.Client, "error", err)
			msg = "Internal error. An administrator can find it in the Logbook."
		}
		apiError(w, status, msg)
		return
	}
	if err := json.NewEncoder(w).Encode(out); err != nil {
		s.log.Warn("write an API response", "operation", name, "error", err)
	}
}

// sessionActor returns the actor of the session that the request presents:
// a person of the Cahut st, or an administrator of the server who signed in
// at its domain. Only administrators of the server reach its operations. It
// returns an error wrapping account.ErrInvalidSession when the request has
// no valid session for this call.
func (s *Server) sessionActor(r *http.Request, st *site, ofServer bool) (operation.Actor, error) {
	token, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
	token = strings.TrimSpace(token)
	actor := operation.Actor{Via: "api", Domain: st.cahut.Domain}
	now := time.Now()
	switch {
	case !ok || token == "":
		return actor, fmt.Errorf("%w: sign in first, then send the session token in the Authorization header", account.ErrInvalidSession)
	case strings.HasPrefix(token, server.SessionPrefix):
		a, err := s.srv.VerifySession(r.Context(), token, st.cahut.Domain, now)
		if err != nil {
			return actor, signInAgain(err)
		}
		actor.Kind, actor.Handle, actor.Session = operation.Admin, a.Handle, a.SessionID
	case strings.HasPrefix(token, account.SessionPrefix) && !ofServer:
		p, err := account.VerifySession(r.Context(), st.cahut.DB, token, now)
		if err != nil {
			return actor, signInAgain(err)
		}
		actor.Kind, actor.UserID, actor.Handle, actor.Session = operation.Person, p.UserID, p.Handle, p.SessionID
	case ofServer:
		return actor, fmt.Errorf("%w: the operations of the server need the session of an administrator of the server", account.ErrInvalidSession)
	default:
		return actor, fmt.Errorf("%w: the API takes session tokens; AIs use MCP with their personal tokens", account.ErrInvalidSession)
	}
	return actor, nil
}

// signInAgain says what to do about a session that is unknown or has ended.
func signInAgain(err error) error {
	if errors.Is(err, account.ErrInvalidSession) {
		return fmt.Errorf("%w: sign in again at this domain", err)
	}
	return err
}

// apiStatus returns the status code of a failed call, and the message that
// the caller may read.
func apiStatus(err error) (int, string) {
	switch {
	case errors.Is(err, operation.ErrInvalid):
		return http.StatusBadRequest, err.Error()
	case errors.Is(err, operation.ErrForbidden):
		return http.StatusForbidden, err.Error()
	case errors.Is(err, operation.ErrNotFound):
		return http.StatusNotFound, err.Error()
	case errors.Is(err, operation.ErrApprovalRequired):
		return http.StatusConflict, err.Error()
	case errors.Is(err, operation.ErrTooMany):
		return http.StatusTooManyRequests, err.Error()
	}
	return http.StatusInternalServerError, ""
}

func apiError(w http.ResponseWriter, status int, msg string) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(struct {
		Error string `json:"error"`
	}{msg})
}

// clientIP returns the address of the request's client, without its port.
func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// clientKey returns the address of the request's client as limits count
// it: an IPv4 address, or the /64 of an IPv6 address, which one machine
// usually holds whole. Counting each IPv6 address would let a single
// machine dodge every limit.
func clientKey(r *http.Request) string {
	ip, err := netip.ParseAddr(clientIP(r))
	if err != nil {
		return clientIP(r)
	}
	ip = ip.Unmap()
	if ip.Is4() {
		return ip.String()
	}
	p, _ := ip.Prefix(64) // cannot fail for an IPv6 address
	return p.String()
}

// limiter allows each address a burst of calls, then one every so often.
// It forgets the addresses it has not seen for a while.
type limiter struct {
	every time.Duration
	burst int

	mu    sync.Mutex
	seen  map[string]*visit
	swept time.Time
}

type visit struct {
	limiter *rate.Limiter
	last    time.Time
}

func newLimiter(every time.Duration, burst int) *limiter {
	return &limiter{every: every, burst: burst, seen: map[string]*visit{}}
}

// spent reports whether the address addr has no call left now, without
// making one.
func (l *limiter) spent(addr string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	v, ok := l.seen[addr]
	return ok && v.limiter.TokensAt(time.Now()) < 1
}

// full reports whether the address addr has its whole burst now: it made no
// call lately.
func (l *limiter) full(addr string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	v, ok := l.seen[addr]
	return !ok || v.limiter.TokensAt(time.Now()) >= float64(l.burst)
}

// allow reports whether the address addr may make a call now.
func (l *limiter) allow(addr string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	// An address that waited long enough to refill its burst is forgotten.
	idle := l.every * time.Duration(l.burst)
	if now.Sub(l.swept) > idle {
		for a, v := range l.seen {
			if now.Sub(v.last) > idle {
				delete(l.seen, a)
			}
		}
		l.swept = now
	}
	v, ok := l.seen[addr]
	if !ok {
		v = &visit{limiter: rate.NewLimiter(rate.Every(l.every), l.burst)}
		l.seen[addr] = v
	}
	v.last = now
	return v.limiter.AllowN(now, 1)
}
