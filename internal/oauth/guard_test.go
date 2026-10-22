package oauth

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestGuard checks that the client of metadata documents refuses to
// connect to this machine, whatever name it is given.
func TestGuard(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	defer srv.Close()
	for _, addr := range []string{srv.URL, strings.Replace(srv.URL, "127.0.0.1", "localhost", 1)} {
		resp, err := guardedClient().Get(addr) //nolint:noctx // the client has its own timeout
		if err == nil {
			resp.Body.Close()
		}
		if err == nil || !strings.Contains(err.Error(), "is not a public address") {
			t.Errorf("GET %s: %v, want a refusal of the address", addr, err)
		}
	}
}
