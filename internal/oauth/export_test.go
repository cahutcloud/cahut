package oauth

import (
	"context"
	"net/http"
	"net/netip"
)

// UseDocumentTransport makes s fetch client metadata documents through rt,
// so that tests reach a server on a loopback address. The other guards
// stay.
func UseDocumentTransport(s *Server, rt http.RoundTripper) {
	s.docs.client.Transport = rt
}

// FetchDocument fetches the metadata document at id, without the cache.
func FetchDocument(s *Server, id string) (Client, error) {
	body, err := s.docs.fetch(context.Background(), id)
	if err != nil {
		return Client{}, err
	}
	return parseDocument(id, body)
}

// ForgetDocuments empties the memory of documents, as a restart does.
func ForgetDocuments(s *Server) {
	s.docs.mu.Lock()
	defer s.docs.mu.Unlock()
	clear(s.docs.docs)
}

// Public reports whether ip is an address on the internet.
func Public(ip string) bool {
	return public(netip.MustParseAddr(ip))
}
