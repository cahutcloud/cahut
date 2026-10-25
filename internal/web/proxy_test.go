package web

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/netip"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"cahut.cloud/internal/update"
)

func TestParseProxies(t *testing.T) {
	for _, tt := range []struct {
		list string
		want []string
	}{
		{"", nil},
		{"127.0.0.1", []string{"127.0.0.1/32"}},
		{" 127.0.0.1 , ::1 ,", []string{"127.0.0.1/32", "::1/128"}},
		{"10.1.2.3/8,fd00::/8", []string{"10.0.0.0/8", "fd00::/8"}},
		{"::ffff:127.0.0.1", []string{"127.0.0.1/32"}},
	} {
		got, err := ParseProxies(tt.list)
		if err != nil {
			t.Errorf("ParseProxies(%q): %v", tt.list, err)
			continue
		}
		var names []string
		for _, p := range got {
			names = append(names, p.String())
		}
		if !slices.Equal(names, tt.want) {
			t.Errorf("ParseProxies(%q) = %v, want %v", tt.list, names, tt.want)
		}
	}
	for _, list := range []string{"localhost", "10.0.0.0/33", "fe80::1%eth0", "127.0.0.1:8442", "127.0.0.1;10.0.0.1"} {
		if _, err := ParseProxies(list); err == nil || !strings.Contains(err.Error(), "IP addresses or prefixes") {
			t.Errorf("ParseProxies(%q) = %v, want a refusal", list, err)
		}
	}
}

// bytesConn is a connection that reads from a buffer, from a proxy at
// 127.0.0.1 to Cahut at 127.0.0.1:8443.
type bytesConn struct {
	net.Conn
	r io.Reader
}

func (c bytesConn) Read(b []byte) (int, error) { return c.r.Read(b) }
func (c bytesConn) RemoteAddr() net.Addr {
	return &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 40000}
}
func (c bytesConn) LocalAddr() net.Addr { return &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 8443} }

// v2Header builds a header of version 2: its version and command byte, its
// family and transport byte, then body.
func v2Header(command, family byte, body []byte) []byte {
	h := append([]byte{}, v2Signature...)
	h = append(h, command, family)
	h = binary.BigEndian.AppendUint16(h, uint16(len(body))) //nolint:gosec // the bodies of the tests are short
	return append(h, body...)
}

// v2Addresses lays out two addresses and their ports, as version 2 does.
func v2Addresses(src, dst string, sport, dport uint16) []byte {
	s, d := netip.MustParseAddr(src), netip.MustParseAddr(dst)
	b := append(s.AsSlice(), d.AsSlice()...)
	b = binary.BigEndian.AppendUint16(b, sport)
	return binary.BigEndian.AppendUint16(b, dport)
}

func TestReadProxyHeader(t *testing.T) {
	inet := v2Addresses("198.51.100.7", "203.0.113.1", 50000, 443)
	inet6 := v2Addresses("2001:db8::7", "2001:db8::1", 50000, 443)
	tlv := []byte{0x01, 0x00, 0x02, 'h', '2'} // PP2_TYPE_ALPN, which Cahut sets aside
	const proxy, cahut = "127.0.0.1:40000", "127.0.0.1:8443"
	for _, tt := range []struct {
		name          string
		in            []byte
		remote, local string
	}{
		{"version 1, TCP4", []byte("PROXY TCP4 198.51.100.7 203.0.113.1 50000 443\r\n"), "198.51.100.7:50000", "203.0.113.1:443"},
		{"version 1, TCP6", []byte("PROXY TCP6 2001:db8::7 2001:db8::1 50000 443\r\n"), "[2001:db8::7]:50000", "[2001:db8::1]:443"},
		{"version 1, UNKNOWN", []byte("PROXY UNKNOWN\r\n"), proxy, cahut},
		{"version 1, UNKNOWN with addresses", []byte("PROXY UNKNOWN ffff:: ffff:: 1 2\r\n"), proxy, cahut},
		{"version 2, INET", v2Header(0x21, 0x11, inet), "198.51.100.7:50000", "203.0.113.1:443"},
		{"version 2, INET6 with a TLV", v2Header(0x21, 0x21, append(inet6, tlv...)), "[2001:db8::7]:50000", "[2001:db8::1]:443"},
		{"version 2, an IPv4 address in IPv6", v2Header(0x21, 0x21, v2Addresses("::ffff:198.51.100.7", "::ffff:203.0.113.1", 50000, 443)),
			"198.51.100.7:50000", "203.0.113.1:443"},
		{"version 2, LOCAL", v2Header(0x20, 0x00, nil), proxy, cahut},
		{"version 2, LOCAL with addresses", v2Header(0x20, 0x11, inet), proxy, cahut},
		{"version 2, UNSPEC", v2Header(0x21, 0x00, nil), proxy, cahut},
		{"version 2, UNIX", v2Header(0x21, 0x31, make([]byte, 216)), proxy, cahut},
	} {
		const after = "GET / HTTP/1.1\r\nHost: cahut.example.com\r\n\r\n"
		pc, err := readProxyHeader(bytesConn{r: io.MultiReader(bytes.NewReader(tt.in), strings.NewReader(after))})
		if err != nil {
			t.Errorf("%s: %v", tt.name, err)
			continue
		}
		if got := pc.RemoteAddr().String(); got != tt.remote {
			t.Errorf("%s: remote address %s, want %s", tt.name, got, tt.remote)
		}
		if got := pc.LocalAddr().String(); got != tt.local {
			t.Errorf("%s: local address %s, want %s", tt.name, got, tt.local)
		}
		// What follows the header reaches the server whole, whether it was
		// read with the header or not.
		if rest, err := io.ReadAll(pc); err != nil || string(rest) != after {
			t.Errorf("%s: after the header, %q (%v), want %q", tt.name, rest, err, after)
		}
	}
}

func TestReadProxyHeaderRefuses(t *testing.T) {
	inet := v2Addresses("198.51.100.7", "203.0.113.1", 50000, 443)
	for _, tt := range []struct {
		name string
		in   []byte
		want string
	}{
		{"a TLS handshake", []byte{0x16, 0x03, 0x01, 0x02, 0x00, 0x01, 0x00, 0x01, 0xfc, 0x03, 0x03, 0x00, 0x00}, "starts with a TLS handshake"},
		{"an HTTP request", []byte("GET / HTTP/1.1\r\nHost: cahut.example.com\r\n\r\n"), "starts without a header"},
		{"a word of PROXY only", []byte("PROXYING / HTTP/1.1\r\n\r\n"), "starts without a header"},
		{"version 1, too long", []byte("PROXY TCP6 " + strings.Repeat("1", 120) + "\r\n"), "longer than 107 bytes"},
		{"version 1, LF only", []byte("PROXY TCP4 198.51.100.7 203.0.113.1 50000 443\n"), "does not end with CRLF"},
		{"version 1, unknown protocol", []byte("PROXY UDP4 198.51.100.7 203.0.113.1 50000 443\r\n"), "neither TCP4"},
		{"version 1, missing port", []byte("PROXY TCP4 198.51.100.7 203.0.113.1 50000\r\n"), "neither TCP4"},
		{"version 1, two spaces", []byte("PROXY TCP4  198.51.100.7 203.0.113.1 50000 443\r\n"), "neither TCP4"},
		{"version 1, IPv6 for TCP4", []byte("PROXY TCP4 2001:db8::7 203.0.113.1 50000 443\r\n"), "bad address"},
		{"version 1, IPv4 for TCP6", []byte("PROXY TCP6 198.51.100.7 2001:db8::1 50000 443\r\n"), "bad address"},
		{"version 1, a zone", []byte("PROXY TCP6 fe80::7%eth0 fe80::1 50000 443\r\n"), "bad address"},
		{"version 1, port out of range", []byte("PROXY TCP4 198.51.100.7 203.0.113.1 65536 443\r\n"), "bad port"},
		{"version 1, cut short", []byte("PROXY TCP4 198.51.100.7"), "ends before the end of its header"},
		{"version 2, another version", v2Header(0x11, 0x11, inet), "a header of version 1, after the signature of version 2"},
		{"version 2, unknown command", v2Header(0x22, 0x11, inet), "unknown command"},
		{"version 2, unknown family", v2Header(0x21, 0x41, inet), "unknown family"},
		{"version 2, a datagram", v2Header(0x21, 0x12, inet), "not a stream"},
		{"version 2, addresses too short", v2Header(0x21, 0x21, inet), "too short for its addresses"},
		{"version 2, cut short", v2Header(0x21, 0x11, inet)[:20], "ends before the end of its header"},
		{"a few bytes only", []byte("PRO"), "ends before the end of its header"},
	} {
		_, err := readProxyHeader(bytesConn{r: bytes.NewReader(tt.in)})
		if err == nil || !strings.Contains(err.Error(), tt.want) {
			t.Errorf("%s: %v, want an error with %q", tt.name, err, tt.want)
		}
	}
	if _, err := readProxyHeader(bytesConn{r: bytes.NewReader(nil)}); !errors.Is(err, errClosedAtOnce) {
		t.Errorf("a connection closed at once: %v, want %v", err, errClosedAtOnce)
	}
}

// serveProxied serves the address of each client over HTTP, through a
// listener that trusts the proxies of list, and returns the listener.
func serveProxied(t *testing.T, list string, log *slog.Logger) *proxyListener {
	t.Helper()
	ln, _ := listen(t)
	trusted, err := ParseProxies(list)
	if err != nil {
		t.Fatal(err)
	}
	l := newProxyListener(ln, "https", trusted, 200*time.Millisecond, log)
	s := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain")
			_, _ = io.WriteString(w, r.RemoteAddr) //nolint:gosec // an address, in plain text
		}),
		ReadHeaderTimeout: 10 * time.Second,
	}
	go s.Serve(l) //nolint:errcheck // ends with Close
	t.Cleanup(func() { s.Close() })
	return l
}

// ask opens a connection to l, sends header then a request, and returns
// the answer, or the error of a connection that the server closed.
func ask(t *testing.T, l net.Listener, header string) (string, error) {
	t.Helper()
	c, err := (&net.Dialer{}).DialContext(context.Background(), "tcp", l.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	_ = c.SetDeadline(time.Now().Add(5 * time.Second))
	if _, err := io.WriteString(c, header+"GET / HTTP/1.1\r\nHost: cahut.example.com\r\nConnection: close\r\n\r\n"); err != nil {
		return "", err
	}
	resp, err := http.ReadResponse(bufio.NewReader(c), nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	return resp.Status + " " + string(body), err
}

func TestProxyListener(t *testing.T) {
	var journal lockedBuffer
	l := serveProxied(t, "127.0.0.1", slog.New(slog.NewTextHandler(&journal, nil)))

	got, err := ask(t, l, "PROXY TCP4 198.51.100.7 203.0.113.1 50000 443\r\n")
	if err != nil || got != "200 OK 198.51.100.7:50000" {
		t.Errorf("version 1: %q (%v), want the address of the client", got, err)
	}
	header := string(v2Header(0x21, 0x21, v2Addresses("2001:db8::7", "2001:db8::1", 50001, 443)))
	if got, err := ask(t, l, header); err != nil || got != "200 OK [2001:db8::7]:50001" {
		t.Errorf("version 2: %q (%v), want the address of the client", got, err)
	}
	if got, err := ask(t, l, "PROXY UNKNOWN\r\n"); err != nil || !strings.HasPrefix(got, "200 OK 127.0.0.1:") {
		t.Errorf("UNKNOWN: %q (%v), want the address of the proxy", got, err)
	}
	if h := l.health(); h.LastTaken == nil || h.LastRefused != nil {
		t.Errorf("health = %+v, want a connection taken and none refused", h)
	}

	// Without a header, or with a bad one, the server answers nothing, and
	// says why.
	for _, header := range []string{"", "PROXY TCP4 198.51.100.7 203.0.113.1 50000\r\n"} {
		if got, err := ask(t, l, header); err == nil {
			t.Errorf("header %q: answered %q, want the connection closed", header, got)
		}
	}
	// A connection that sends nothing is refused once its time is up.
	c, err := (&net.Dialer{}).DialContext(context.Background(), "tcp", l.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	_ = c.SetDeadline(time.Now().Add(5 * time.Second))
	if n, err := c.Read(make([]byte, 1)); err == nil {
		t.Errorf("a silent connection read %d bytes, want it closed", n)
	}
	c.Close()
	// A connection closed at once, as the TCP check of a load balancer,
	// leaves no line.
	c, err = (&net.Dialer{}).DialContext(context.Background(), "tcp", l.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	c.Close()
	time.Sleep(50 * time.Millisecond)

	h := l.health()
	if h.LastRefused == nil || !strings.Contains(h.Refusal, "no header within 200ms") ||
		!slices.Equal(h.Trusted, []string{"127.0.0.1"}) || h.Listener != "https" || h.Address != l.Addr().String() {
		t.Errorf("health = %+v, want the last refusal", h)
	}
	lines := journal.String()
	for _, want := range []string{"the connection starts without a header", "neither TCP4", "no header within 200ms", "listener=https", "proxy=127.0.0.1:"} {
		if !strings.Contains(lines, want) {
			t.Errorf("the journal does not say %q:\n%s", want, lines)
		}
	}
	if n := strings.Count(lines, "refuse a connection"); n != 3 {
		t.Errorf("the journal refuses %d connections, want 3:\n%s", n, lines)
	}
}

// lockedBuffer lets the goroutines of a listener log to one buffer.
type lockedBuffer struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (lb *lockedBuffer) Write(p []byte) (int, error) {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	return lb.b.Write(p)
}

func (lb *lockedBuffer) String() string {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	return lb.b.String()
}

// From an address that it does not trust, the listener reads no header:
// the connection is served as it comes, under its own address.
func TestProxyListenerUntrusted(t *testing.T) {
	l := serveProxied(t, "192.0.2.0/24", slog.New(slog.DiscardHandler))
	if got, err := ask(t, l, ""); err != nil || !strings.HasPrefix(got, "200 OK 127.0.0.1:") {
		t.Errorf("no header: %q (%v), want the address of the connection", got, err)
	}
	if got, err := ask(t, l, "PROXY TCP4 198.51.100.7 203.0.113.1 50000 443\r\n"); err != nil || !strings.HasPrefix(got, "400 ") {
		t.Errorf("a header: %q (%v), want a bad request", got, err)
	}
	if h := l.health(); h.LastTaken != nil || h.LastRefused != nil {
		t.Errorf("health = %+v, want nothing taken or refused", h)
	}
}

// The header comes before TLS: the handshake that follows it sees the
// addresses that the header names, and a client that sends no server name
// gets the certificate of the address it called.
func TestProxyListenerBeforeTLS(t *testing.T) {
	ln, _ := listen(t)
	l := newProxyListener(ln, "https", []netip.Prefix{netip.MustParsePrefix("127.0.0.1/32")}, time.Second, slog.New(slog.DiscardHandler))
	defer l.Close()
	cert := selfSigned(t, "203.0.113.7")
	called := make(chan string, 1)
	server := tls.NewListener(l, &tls.Config{
		NextProtos: []string{"acme-tls/1"},
		GetCertificate: func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
			called <- hello.Conn.LocalAddr().String() + " from " + hello.Conn.RemoteAddr().String()
			return &cert, nil
		},
	})
	go func() {
		c, err := server.Accept()
		if err != nil {
			return
		}
		defer c.Close()
		_ = c.(*tls.Conn).HandshakeContext(context.Background())
	}()

	raw, err := (&net.Dialer{}).DialContext(context.Background(), "tcp", ln.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	defer raw.Close()
	_ = raw.SetDeadline(time.Now().Add(5 * time.Second))
	if _, err := raw.Write(v2Header(0x21, 0x11, v2Addresses("198.51.100.7", "203.0.113.7", 50000, 443))); err != nil {
		t.Fatal(err)
	}
	pool := x509.NewCertPool()
	pool.AddCert(cert.Leaf)
	c := tls.Client(raw, &tls.Config{RootCAs: pool, ServerName: "203.0.113.7", NextProtos: []string{"acme-tls/1"}})
	if err := c.HandshakeContext(context.Background()); err != nil {
		t.Fatalf("handshake after the header: %v", err)
	}
	if p := c.ConnectionState().NegotiatedProtocol; p != "acme-tls/1" {
		t.Errorf("protocol %q, want acme-tls/1", p)
	}
	if got, want := <-called, "203.0.113.7:443 from 198.51.100.7:50000"; got != want {
		t.Errorf("the handshake saw %s, want %s", got, want)
	}
}

// Behind a proxy that passes the PROXY protocol, the limits count each
// client by its own address, the logs name it, and the local health check
// answers the machine itself only.
func TestProxyProtocolCountsEachClient(t *testing.T) {
	f := setup(t)
	var journal lockedBuffer
	f.web.log = slog.New(slog.NewTextHandler(&journal, nil))
	ln, err := f.web.listen(context.Background(), "https", "127.0.0.1:0", []netip.Prefix{netip.MustParsePrefix("127.0.0.1/32")})
	if err != nil {
		t.Fatal(err)
	}
	s := &http.Server{Handler: f.web.localHealth(f.web), ReadHeaderTimeout: 10 * time.Second}
	go s.Serve(ln) //nolint:errcheck // ends with Close
	t.Cleanup(func() { s.Close() })
	// call sends a request through a connection that the proxy opens for
	// client.
	call := func(client, method, url, body string) (int, string) {
		t.Helper()
		tr := &http.Transport{DisableKeepAlives: true, DialContext: func(ctx context.Context, network, _ string) (net.Conn, error) {
			c, err := (&net.Dialer{}).DialContext(ctx, network, ln.Addr().String())
			if err == nil {
				_, err = io.WriteString(c, "PROXY TCP4 "+client+" 203.0.113.1 50000 443\r\n")
			}
			return c, err
		}}
		req, err := http.NewRequestWithContext(context.Background(), method, url, strings.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}
		resp, err := (&http.Client{Transport: tr}).Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		b, _ := io.ReadAll(resp.Body)
		return resp.StatusCode, string(b)
	}
	register := func(client string) int {
		status, _ := call(client, http.MethodPost, "http://localhost/oauth/register", `{"redirect_uris": ["http://127.0.0.1/callback"]}`)
		return status
	}
	for i := range RegisterBurst {
		if status := register("198.51.100.1"); status != http.StatusCreated {
			t.Fatalf("registration %d = %d, want 201", i+1, status)
		}
	}
	if status := register("198.51.100.1"); status != http.StatusTooManyRequests {
		t.Errorf("registration %d of one client = %d, want 429", RegisterBurst+1, status)
	}
	if status := register("198.51.100.2"); status != http.StatusCreated {
		t.Errorf("the registration of another client = %d, want 201", status)
	}
	lines := journal.String()
	for _, want := range []string{"client=198.51.100.1", "client=198.51.100.2", "listener=https", "proxies=[127.0.0.1]"} {
		if !strings.Contains(lines, want) {
			t.Errorf("the journal does not say %q:\n%s", want, lines)
		}
	}
	if strings.Contains(lines, "client=127.0.0.1") {
		t.Errorf("the journal names the proxy as a client:\n%s", lines)
	}

	if status, _ := call("198.51.100.1", http.MethodGet, "http://127.0.0.1/-/healthz", ""); status != http.StatusMisdirectedRequest {
		t.Errorf("the local health check answers a client of the internet: %d", status)
	}
	status, body := call("127.0.0.1", http.MethodGet, "http://127.0.0.1/-/healthz", "")
	var h update.Health
	if err := json.Unmarshal([]byte(body), &h); err != nil || status != http.StatusOK {
		t.Fatalf("local health = %d %s", status, body)
	}
	if len(h.ProxyProtocol) != 1 || h.ProxyProtocol[0].Listener != "https" || h.ProxyProtocol[0].LastTaken == nil ||
		!slices.Equal(h.ProxyProtocol[0].Trusted, []string{"127.0.0.1"}) {
		t.Errorf("local health = %+v, want the listener that reads the PROXY protocol", h.ProxyProtocol)
	}
}
