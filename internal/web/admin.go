package web

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"cahut.cloud/internal/account"
	"cahut.cloud/internal/operation"
)

// The admin API serves the operations of every Cahut of the server that are
// open to services, to the programs of the machine itself that hold a
// service token: POST /api/<operation>, the host of the request naming the
// Cahut, as for the API of the console. It listens at an address of the
// machine only, and takes service tokens only. See
// docs/adr/0035-a-program-on-the-server-manages-people-with-a-service-token.md.

// CheckAdminAPIAddr returns an error unless addr is an address of the
// machine itself: a loopback address, or localhost, with a port.
func CheckAdminAPIAddr(addr string) error {
	host, port, err := net.SplitHostPort(addr)
	if err != nil || port == "" {
		return fmt.Errorf("admin API at %q: give an address and a port, such as 127.0.0.1:8090", addr)
	}
	if ip := net.ParseIP(host); host != "localhost" && (ip == nil || !ip.IsLoopback()) {
		return fmt.Errorf("admin API at %q: it listens on the machine itself only, at 127.0.0.1, [::1] or localhost", addr)
	}
	return nil
}

// AdminHandler returns the handler of the admin API.
func (s *Server) AdminHandler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/{operation}", s.serveAdminAPI)
	mux.HandleFunc("/", refuseAdminAPI)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setAPIHeaders(w)
		if !fromMachine(r) {
			apiError(w, http.StatusForbidden, "the admin API serves the machine itself only")
			return
		}
		st, err := s.site(r.Context(), hostOf(r))
		if errors.Is(err, operation.ErrNotFound) {
			apiError(w, http.StatusMisdirectedRequest,
				fmt.Sprintf("no Cahut at %q: the host of the request names the Cahut, such as Host: my.example.org", hostOf(r)))
			return
		}
		if err != nil {
			s.log.Error("open a Cahut", "host", hostOf(r), "error", err)
			apiError(w, http.StatusInternalServerError, "Internal error.")
			return
		}
		mux.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), siteKey{}, st)))
	})
}

// fromMachine reports whether r comes from the machine itself, whatever its
// host names.
func fromMachine(r *http.Request) bool {
	ip := net.ParseIP(clientIP(r))
	return ip != nil && ip.IsLoopback()
}

// refuseAdminAPI answers the requests to the admin API that are not a POST
// to an operation.
func refuseAdminAPI(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		apiError(w, http.StatusMethodNotAllowed, "the admin API takes POST requests only, with the input of the operation in JSON")
		return
	}
	apiError(w, http.StatusNotFound, "no such route: call POST /api/<operation>")
}

// serveAdminAPI runs an operation of the Cahut of the request, for the
// service whose token the request presents.
func (s *Server) serveAdminAPI(w http.ResponseWriter, r *http.Request) {
	st := siteOf(r)
	reg := st.cahut.Registry
	name := r.PathValue("operation")
	if _, ok := reg.Lookup(name); !ok {
		apiError(w, http.StatusNotFound, fmt.Sprintf("no operation %q: the reference of the admin API lists those open to services", name))
		return
	}
	body, ok := readCall(w, r, MaxAPIRequestSize)
	if !ok {
		return
	}
	token, _ := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
	o, err := account.VerifyService(r.Context(), st.cahut.DB, strings.TrimSpace(token), time.Now())
	if errors.Is(err, account.ErrInvalidToken) {
		w.Header().Set("WWW-Authenticate", "Bearer")
		apiError(w, http.StatusUnauthorized,
			"the admin API takes the service tokens of this Cahut only, in the Authorization header: Bearer cahut_svc_…")
		return
	}
	if err != nil {
		s.log.Error("verify a service token", "cahut", st.cahut.Domain, "error", err)
		apiError(w, http.StatusInternalServerError, "Internal error.")
		return
	}
	actor := o.Actor(st.cahut.Domain)
	actor.Address = clientKey(r)
	s.invoke(w, r, reg, name, actor, body)
}
