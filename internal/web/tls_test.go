package web

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"log/slog"
	"math/big"
	"net"
	"testing"
	"time"
)

// selfSigned returns a certificate for name, a domain or an IP address.
func selfSigned(t *testing.T, name string) tls.Certificate {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	tmpl := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: name},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(72 * time.Hour),
	}
	if ip := net.ParseIP(name); ip != nil {
		tmpl.IPAddresses = []net.IP{ip}
	} else {
		tmpl.DNSNames = []string{name}
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	leaf, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatal(err)
	}
	return tls.Certificate{Certificate: [][]byte{der}, PrivateKey: key, Leaf: leaf}
}

// conn is a connection that only knows its addresses.
type conn struct {
	net.Conn
	local net.Addr
}

func (c conn) LocalAddr() net.Addr { return c.local }
func (c conn) RemoteAddr() net.Addr {
	return &net.TCPAddr{IP: net.ParseIP("198.51.100.9"), Port: 50000}
}

func TestCertificateByNameOrAddress(t *testing.T) {
	f := setup(t)
	s := New(f.srv, Config{CertDir: t.TempDir(), CA: "https://acme.invalid/directory"}, slog.New(slog.DiscardHandler))
	ctx := context.Background()
	c, err := s.certificates(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.names.CacheUnmanagedTLSCertificate(ctx, selfSigned(t, "cahut.example.com"), nil); err != nil {
		t.Fatal(err)
	}
	if _, err := c.ips.CacheUnmanagedTLSCertificate(ctx, selfSigned(t, "203.0.113.7"), nil); err != nil {
		t.Fatal(err)
	}
	cfg := c.tlsConfig()
	for _, tt := range []struct {
		name  string
		hello *tls.ClientHelloInfo
		want  string
	}{
		{"by its name", &tls.ClientHelloInfo{ServerName: "cahut.example.com"}, "cahut.example.com"},
		{"without a name", &tls.ClientHelloInfo{Conn: conn{local: &net.TCPAddr{IP: net.ParseIP("203.0.113.7"), Port: 443}}}, "203.0.113.7"},
	} {
		cert, err := cfg.GetCertificate(tt.hello)
		if err != nil {
			t.Errorf("%s: %v", tt.name, err)
			continue
		}
		leaf, err := x509.ParseCertificate(cert.Certificate[0])
		if err != nil {
			t.Fatal(err)
		}
		if leaf.Subject.CommonName != tt.want {
			t.Errorf("%s: the certificate of %s, want %s", tt.name, leaf.Subject.CommonName, tt.want)
		}
	}
	if cfg.NextProtos[0] != "h2" {
		t.Errorf("NextProtos = %v", cfg.NextProtos)
	}
}
