package doctor

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io"
	"io/fs"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"strconv"
	"strings"
	"syscall"
	"time"

	"cahut.cloud/internal/windata"
)

// Machine is the System of the machine that runs cahut doctor.
type Machine struct{}

var _ System = Machine{}

// Command runs name with args.
func (Machine) Command(ctx context.Context, name string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, name, args...).Output()
	return strings.TrimSpace(string(out)), err
}

// ReadFile reads a file.
func (Machine) ReadFile(name string) ([]byte, error) { return os.ReadFile(name) }

// Stat describes a file.
func (Machine) Stat(name string) (fs.FileInfo, error) { return os.Stat(name) }

// Owner returns the name of the user that owns name.
func (Machine) Owner(name string) (string, error) {
	fi, err := os.Stat(name)
	if err != nil {
		return "", err
	}
	st, ok := fi.Sys().(*syscall.Stat_t)
	if !ok {
		return "", errors.New("no owner")
	}
	id := strconv.FormatUint(uint64(st.Uid), 10)
	if u, err := user.LookupId(id); err == nil {
		return u.Username, nil
	}
	return id, nil
}

// Label returns the SELinux label of name.
func (Machine) Label(name string) (string, error) { return label(name) }

// FreeBytes returns the space left to unprivileged users on the file
// system of path.
func (Machine) FreeBytes(path string) (uint64, error) { return freeBytes(path) }

// LocalAddrs returns the addresses of the interfaces of the machine.
func (Machine) LocalAddrs() ([]net.IP, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}
	var ips []net.IP
	for _, a := range addrs {
		if n, ok := a.(*net.IPNet); ok {
			ips = append(ips, n.IP)
		}
	}
	return ips, nil
}

// LookupIP resolves host.
func (Machine) LookupIP(ctx context.Context, host string) ([]net.IP, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	return net.DefaultResolver.LookupIP(ctx, "ip", host)
}

// Dial connects to addr over TCP.
func (Machine) Dial(ctx context.Context, addr string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	var d net.Dialer
	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return err
	}
	return conn.Close()
}

// Get reads url, checking certificates.
func (Machine) Get(ctx context.Context, url string) (int, []byte, error) {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("User-Agent", "cahut-doctor")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	return resp.StatusCode, body, err
}

// Certificate connects to addr over TLS, asking for name, and returns the
// certificate it serves, with whether browsers would trust it for name.
func (Machine) Certificate(ctx context.Context, addr, name string) (Certificate, error) {
	ctx, cancel := context.WithTimeout(ctx, 20*time.Second) // a first certificate may be obtained on the way
	defer cancel()
	d := tls.Dialer{Config: &tls.Config{ServerName: name, InsecureSkipVerify: true}} //nolint:gosec // the certificate is checked below, to say what is wrong with it
	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return Certificate{}, err
	}
	defer conn.Close()
	certs := conn.(*tls.Conn).ConnectionState().PeerCertificates
	if len(certs) == 0 {
		return Certificate{}, errors.New("no certificate")
	}
	leaf := certs[0]
	inter := x509.NewCertPool()
	for _, c := range certs[1:] {
		inter.AddCert(c)
	}
	_, verr := leaf.Verify(x509.VerifyOptions{DNSName: name, Intermediates: inter})
	return Certificate{NotBefore: leaf.NotBefore, NotAfter: leaf.NotAfter, Issuer: issuer(leaf), Trusted: verr}, nil
}

func issuer(c *x509.Certificate) string {
	if len(c.Issuer.Organization) > 0 {
		return c.Issuer.Organization[0]
	}
	return c.Issuer.CommonName
}

// Root reports whether cahut doctor runs as root.
func (Machine) Root() bool { return os.Geteuid() == 0 }

// Landlock returns the version of Landlock that the kernel offers, or 0.
func (Machine) Landlock() int { return windata.LandlockABI() }

// Now returns the time.
func (Machine) Now() time.Time { return time.Now() }
