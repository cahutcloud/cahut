package web

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/netip"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"cahut.cloud/internal/update"
)

// The PROXY protocol, which a proxy in front of Cahut speaks to pass the
// address of each client, in a header before the first byte of the
// connection: a line of text in version 1, a binary block in version 2.
// Cahut reads it from the proxies it trusts only, and from them always. See
// docs/adr/0041-the-proxy-protocol-from-trusted-proxies.md and
// https://www.haproxy.org/download/3.2/doc/proxy-protocol.txt.

// ParseProxies reads a list of trusted proxies: IP addresses or prefixes,
// separated by commas, such as "127.0.0.1" or "10.0.0.0/8". An empty list
// trusts no proxy.
func ParseProxies(list string) ([]netip.Prefix, error) {
	var proxies []netip.Prefix
	for entry := range strings.SplitSeq(list, ",") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		p, ok := parseProxy(entry)
		if !ok {
			return nil, fmt.Errorf("proxy %q: give IP addresses or prefixes, separated by commas, such as 127.0.0.1 or 10.0.0.0/8", entry)
		}
		proxies = append(proxies, p)
	}
	return proxies, nil
}

// parseProxy reads an IP address, without a zone, or a prefix.
func parseProxy(entry string) (netip.Prefix, bool) {
	if strings.Contains(entry, "/") {
		p, err := netip.ParsePrefix(entry)
		return p.Masked(), err == nil
	}
	a, err := netip.ParseAddr(entry)
	if err != nil || a.Zone() != "" {
		return netip.Prefix{}, false
	}
	a = a.Unmap()
	return netip.PrefixFrom(a, a.BitLen()), true
}

// proxyNames writes proxies as people write them: an address alone, or a
// prefix.
func proxyNames(proxies []netip.Prefix) []string {
	names := []string{}
	for _, p := range proxies {
		if p.IsSingleIP() {
			names = append(names, p.Addr().String())
		} else {
			names = append(names, p.String())
		}
	}
	return names
}

// proxyHeaderTimeout is how long a trusted proxy has to send its header.
const proxyHeaderTimeout = 10 * time.Second

// proxyListener reads the header of the PROXY protocol on the connections of
// the proxies it trusts, each in a goroutine of its own, so that a slow one
// holds back no other. It passes the connections of other addresses on as
// they come.
type proxyListener struct {
	net.Listener
	name    string // of the listener, such as https
	trusted []netip.Prefix
	timeout time.Duration // for the header
	log     *slog.Logger

	ready   chan accepted
	done    chan struct{}
	closing sync.Once

	mu                     sync.Mutex
	lastTaken, lastRefused time.Time
	refusal                string // why the last connection was refused
}

// accepted is what Accept returns.
type accepted struct {
	conn net.Conn
	err  error
}

// newProxyListener reads the PROXY protocol on the connections of ln that
// come from the trusted proxies, which have timeout to send their header.
func newProxyListener(ln net.Listener, name string, trusted []netip.Prefix, timeout time.Duration, log *slog.Logger) *proxyListener {
	l := &proxyListener{Listener: ln, name: name, trusted: trusted, timeout: timeout, log: log,
		ready: make(chan accepted), done: make(chan struct{})}
	go l.accept()
	return l
}

func (l *proxyListener) Accept() (net.Conn, error) {
	select {
	case a := <-l.ready:
		return a.conn, a.err
	case <-l.done:
		return nil, net.ErrClosed
	}
}

func (l *proxyListener) Close() error {
	l.closing.Do(func() { close(l.done) })
	return l.Listener.Close()
}

// accept takes the connections of the listener until it closes. Its errors
// go to Accept, which net/http retries after a pause when they are
// temporary, as when the process has no file left.
func (l *proxyListener) accept() {
	for {
		c, err := l.Listener.Accept()
		switch {
		case err != nil:
			if !l.deliver(accepted{err: err}) || errors.Is(err, net.ErrClosed) {
				return
			}
		case l.trusts(c.RemoteAddr()):
			go l.handshake(c)
		default:
			if !l.deliver(accepted{conn: c}) {
				return
			}
		}
	}
}

// deliver hands a to Accept, or closes its connection when the listener
// closes first.
func (l *proxyListener) deliver(a accepted) bool {
	select {
	case l.ready <- a:
		return true
	case <-l.done:
		if a.conn != nil {
			_ = a.conn.Close()
		}
		return false
	}
}

// trusts reports whether addr is that of a trusted proxy.
func (l *proxyListener) trusts(addr net.Addr) bool {
	tcp, ok := addr.(*net.TCPAddr)
	if !ok {
		return false
	}
	ip := tcp.AddrPort().Addr().Unmap()
	for _, p := range l.trusted {
		if p.Contains(ip) {
			return true
		}
	}
	return false
}

// handshake reads the header of a connection from a trusted proxy, and
// hands the connection on with the addresses it names. Without a valid
// header, it closes the connection.
func (l *proxyListener) handshake(c net.Conn) {
	_ = c.SetReadDeadline(time.Now().Add(l.timeout))
	pc, err := readProxyHeader(c)
	if err == nil {
		err = c.SetReadDeadline(time.Time{})
	}
	if errors.Is(err, os.ErrDeadlineExceeded) {
		err = fmt.Errorf("no header within %s", l.timeout)
	}
	if err != nil {
		if !errors.Is(err, errClosedAtOnce) {
			l.refused(c.RemoteAddr(), err)
		}
		_ = c.Close()
		return
	}
	l.mu.Lock()
	l.lastTaken = time.Now()
	l.mu.Unlock()
	l.deliver(accepted{conn: pc})
}

// refused notes and logs a connection of a trusted proxy that came without
// a valid header.
func (l *proxyListener) refused(proxy net.Addr, err error) {
	l.mu.Lock()
	l.lastRefused, l.refusal = time.Now(), err.Error()
	l.mu.Unlock()
	l.log.Warn("refuse a connection of a trusted proxy, without a valid PROXY protocol header",
		"listener", l.name, "proxy", proxy.String(), "error", err)
}

// health says what the listener trusts, and how its last connections went.
func (l *proxyListener) health() update.ProxyHealth {
	h := update.ProxyHealth{Listener: l.name, Address: l.Addr().String(), Trusted: proxyNames(l.trusted)}
	l.mu.Lock()
	defer l.mu.Unlock()
	if !l.lastTaken.IsZero() {
		h.LastTaken = new(l.lastTaken.UTC())
	}
	if !l.lastRefused.IsZero() {
		h.LastRefused, h.Refusal = new(l.lastRefused.UTC()), l.refusal
	}
	return h
}

// proxiedConn is a connection with the addresses that its header names.
type proxiedConn struct {
	net.Conn
	rest          []byte // read with the header, and not yet served
	remote, local net.Addr
}

func (c *proxiedConn) Read(b []byte) (int, error) {
	if len(c.rest) > 0 {
		n := copy(b, c.rest)
		c.rest = c.rest[n:]
		return n, nil
	}
	return c.Conn.Read(b)
}

func (c *proxiedConn) RemoteAddr() net.Addr { return c.remote }
func (c *proxiedConn) LocalAddr() net.Addr  { return c.local }

// CloseWrite lets net/http end its side of a connection, as it does on TCP.
func (c *proxiedConn) CloseWrite() error {
	if cw, ok := c.Conn.(interface{ CloseWrite() error }); ok {
		return cw.CloseWrite()
	}
	return nil
}

var (
	// errClosedAtOnce is a connection closed before its first byte, as the
	// TCP checks of load balancers do.
	errClosedAtOnce = errors.New("closed before its first byte")

	v1Prefix    = []byte("PROXY ")
	v2Signature = []byte("\r\n\r\n\x00\r\nQUIT\n")
)

// v1MaxLine is the longest header of version 1, with its CRLF.
const v1MaxLine = 107

// readProxyHeader reads the header at the start of c, version 1 or 2.
func readProxyHeader(c net.Conn) (*proxiedConn, error) {
	br := bufio.NewReaderSize(c, 512)
	pc := &proxiedConn{Conn: c, remote: c.RemoteAddr(), local: c.LocalAddr()}
	head, err := br.Peek(len(v2Signature))
	switch {
	case len(head) == 0 && errors.Is(err, io.EOF):
		return nil, errClosedAtOnce
	case bytes.Equal(head, v2Signature):
		err = readV2(br, pc)
	case bytes.HasPrefix(head, v1Prefix):
		err = readV1(br, pc)
	case err != nil:
	case head[0] == 0x16: // a TLS record of the handshake
		return nil, errors.New("the connection starts with a TLS handshake, without a header")
	default:
		return nil, errors.New("the connection starts without a header")
	}
	switch {
	case errors.Is(err, io.EOF), errors.Is(err, io.ErrUnexpectedEOF):
		return nil, errors.New("the connection ends before the end of its header")
	case err != nil:
		return nil, err
	}
	if n := br.Buffered(); n > 0 {
		rest, _ := br.Peek(n)
		pc.rest = bytes.Clone(rest)
	}
	return pc, nil
}

// readV1 reads a header of version 1, such as
// "PROXY TCP4 198.51.100.7 203.0.113.1 50000 443\r\n".
func readV1(br *bufio.Reader, pc *proxiedConn) error {
	var line []byte
	for !bytes.HasSuffix(line, []byte("\n")) {
		if len(line) == v1MaxLine {
			return fmt.Errorf("a header of version 1 longer than %d bytes", v1MaxLine)
		}
		b, err := br.ReadByte()
		if err != nil {
			return err
		}
		line = append(line, b)
	}
	text, ok := strings.CutSuffix(string(line), "\r\n")
	if !ok {
		return errors.New("a header of version 1 that does not end with CRLF")
	}
	fields := strings.Split(text, " ")
	switch {
	case len(fields) >= 2 && fields[1] == "UNKNOWN":
		return nil // the proxy speaks for itself
	case len(fields) != 6 || fields[1] != "TCP4" && fields[1] != "TCP6":
		return fmt.Errorf("a header of version 1 that is neither TCP4, TCP6 nor UNKNOWN: %q", text)
	}
	v4 := fields[1] == "TCP4"
	var addrs [2]netip.AddrPort
	for i := range addrs {
		ip, err := netip.ParseAddr(fields[2+i])
		if err != nil || ip.Zone() != "" || ip.Is4() != v4 {
			return fmt.Errorf("a header of version 1 with a bad address for %s: %q", fields[1], fields[2+i])
		}
		port, err := strconv.ParseUint(fields[4+i], 10, 16)
		if err != nil {
			return fmt.Errorf("a header of version 1 with a bad port: %q", fields[4+i])
		}
		addrs[i] = netip.AddrPortFrom(ip.Unmap(), uint16(port))
	}
	pc.remote, pc.local = net.TCPAddrFromAddrPort(addrs[0]), net.TCPAddrFromAddrPort(addrs[1])
	return nil
}

// readV2 reads a header of version 2: its signature, its version and
// command, its family and transport, the length of what follows, then the
// addresses and the TLVs, which Cahut sets aside.
func readV2(br *bufio.Reader, pc *proxiedConn) error {
	var h [16]byte
	if _, err := io.ReadFull(br, h[:]); err != nil {
		return err
	}
	if v := h[12] >> 4; v != 2 {
		return fmt.Errorf("a header of version %d, after the signature of version 2", v)
	}
	command, family, transport := h[12]&0x0f, h[13]>>4, h[13]&0x0f
	body := make([]byte, binary.BigEndian.Uint16(h[14:]))
	if _, err := io.ReadFull(br, body); err != nil {
		return err
	}
	switch command {
	case 0x0: // LOCAL: the proxy speaks for itself
		return nil
	case 0x1: // PROXY
	default:
		return fmt.Errorf("a header of version 2 with the unknown command %#x", command)
	}
	var size int
	switch family {
	case 0x0, 0x3: // UNSPEC, UNIX: no IP address to take
		return nil
	case 0x1: // INET
		size = 4
	case 0x2: // INET6
		size = 16
	default:
		return fmt.Errorf("a header of version 2 with the unknown family %#x", family)
	}
	if transport != 0x1 {
		return fmt.Errorf("a header of version 2 for the transport %#x, not a stream", transport)
	}
	if len(body) < 2*size+4 {
		return fmt.Errorf("a header of version 2 too short for its addresses: %d bytes", len(body))
	}
	var addrs [2]netip.AddrPort
	for i := range addrs {
		ip, _ := netip.AddrFromSlice(body[i*size : (i+1)*size])
		port := binary.BigEndian.Uint16(body[2*size+2*i:])
		addrs[i] = netip.AddrPortFrom(ip.Unmap(), port)
	}
	pc.remote, pc.local = net.TCPAddrFromAddrPort(addrs[0]), net.TCPAddrFromAddrPort(addrs[1])
	return nil
}
