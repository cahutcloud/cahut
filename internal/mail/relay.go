package mail

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"net/mail"
	"os"
	"path/filepath"
	"strings"
	"time"

	"cahut.cloud/internal/address"
	"cahut.cloud/internal/store"
)

// Modes of TLS of the connection to a relay.
const (
	// TLS speaks TLS from the start, usually on port 465.
	TLS = "tls"
	// StartTLS connects in clear, then asks for TLS before anything else,
	// usually on port 587.
	StartTLS = "starttls"
	// Plain does without TLS: only the relay of the server may, on its
	// loopback address, where the message does not leave the machine.
	Plain = "none"
)

// Relay is an SMTP relay, with its password in clear once unsealed.
type Relay struct {
	Host     string
	Port     int
	Security string
	Username string
	Password string
	From     string // the address messages come from
	Name     string // the name they come from; empty for the name of the Cahut
	SetAt    time.Time
}

// querier runs a query, in a transaction or not.
type querier interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// stored is the relay kept in a database, with its password sealed.
type stored struct {
	Relay
	sealed []byte
}

// load returns the relay kept in the database that q reads, or nil.
func load(ctx context.Context, q querier) (*stored, error) {
	var r stored
	var at string
	err := q.QueryRowContext(ctx, `
		SELECT host, port, security, username, password, sender, sender_name, set_at FROM mail_relay WHERE id = 1`).
		Scan(&r.Host, &r.Port, &r.Security, &r.Username, &r.sealed, &r.From, &r.Name, &at)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	r.SetAt, err = store.ParseTime(at)
	return &r, err
}

// save keeps r in the database of tx, with its password sealed.
func save(ctx context.Context, tx *sql.Tx, r *stored) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO mail_relay (id, host, port, security, username, password, sender, sender_name, set_at)
		VALUES (1, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (id) DO UPDATE SET host = excluded.host, port = excluded.port, security = excluded.security,
		  username = excluded.username, password = excluded.password, sender = excluded.sender,
		  sender_name = excluded.sender_name, set_at = excluded.set_at`,
		r.Host, r.Port, r.Security, r.Username, r.sealed, r.From, r.Name, store.FormatTime(r.SetAt))
	return err
}

// remove removes the relay kept in the database of tx.
func remove(ctx context.Context, tx *sql.Tx) error {
	_, err := tx.ExecContext(ctx, "DELETE FROM mail_relay")
	return err
}

// relayOf returns the relay that sends the e-mail of from, with its
// password unsealed: that of its Cahut, or else that of the server, or nil
// when there is none. own reports whether it is the Cahut's.
func (s *Service) relayOf(ctx context.Context, from From) (r *Relay, own bool, err error) {
	if from.DB != nil {
		st, err := load(ctx, from.DB.Read)
		if err != nil {
			return nil, false, err
		}
		if st != nil {
			r, err := s.open(st)
			return r, true, err
		}
	}
	st, err := load(ctx, s.cfg.Server.Read)
	if err != nil || st == nil {
		return nil, false, err
	}
	r, err = s.open(st)
	return r, from.DB == nil, err
}

// open returns the relay st with its password unsealed.
func (s *Service) open(st *stored) (*Relay, error) {
	r := st.Relay
	if st.sealed == nil {
		return &r, nil
	}
	var err error
	r.Password, err = s.unseal(st.sealed, st.Host, st.Username)
	return &r, err
}

// check returns an error that says what is wrong with r, as given to
// set_mail_relay, once cleaned. A Cahut's relay must speak TLS.
func check(r *Relay, cahut bool) error {
	r.Host = strings.TrimSuffix(strings.ToLower(strings.TrimSpace(r.Host)), ".")
	r.Username = strings.TrimSpace(r.Username)
	r.Name = strings.TrimSpace(r.Name)
	if !address.ValidDomain(r.Host) && !address.IsIP(r.Host) && r.Host != "localhost" {
		return fmt.Errorf("host %q: give the name or the address of the relay alone, such as smtp.example.com, without a scheme or a port", r.Host)
	}
	if r.Security == "" {
		r.Security = StartTLS
		if r.Port == 465 {
			r.Security = TLS
		}
	}
	if r.Port == 0 {
		r.Port = 587
		if r.Security == TLS {
			r.Port = 465
		}
	}
	switch {
	case r.Port < 1 || r.Port > 65535:
		return fmt.Errorf("port %d: give a port from 1 to 65535, usually 465 or 587", r.Port)
	case r.Security == Plain && cahut:
		return errors.New("security none: the relay of a Cahut speaks TLS, from the start (tls) or after STARTTLS (starttls)")
	case r.Security != TLS && r.Security != StartTLS && r.Security != Plain:
		return fmt.Errorf("security %q: use tls, starttls or, for a relay of the server on this machine, none", r.Security)
	case r.Username == "" && r.Password != "":
		return errors.New("a password without a user name: give the user name at the relay too")
	case len(r.Username) > 256 || len(r.Password) > 1024:
		return errors.New("the user name or the password is too long for a relay")
	case strings.ContainsAny(r.Username+r.Password, "\x00\r\n"):
		return errors.New("the user name or the password holds a line break or a null character")
	case len(r.Name) > 100 || strings.ContainsAny(r.Name, "\r\n"):
		return errors.New("name: give a name of 100 characters at most, on one line")
	}
	from, err := Address(r.From)
	if err != nil {
		return fmt.Errorf("from: %w", err)
	}
	r.From = from
	return nil
}

// Address returns the e-mail address s, cleaned, or an error that says
// what is wrong with it: one address alone, such as jane@example.com,
// without a name or angle brackets.
func Address(s string) (string, error) {
	s = strings.TrimSpace(s)
	a, err := mail.ParseAddress(s)
	if err != nil || a.Name != "" || a.Address != s || len(s) > 254 || !strings.Contains(s, "@") {
		return "", fmt.Errorf("%q is not an e-mail address: give one address alone, such as jane@example.com", s)
	}
	local, domain, _ := strings.Cut(s, "@")
	return local + "@" + strings.ToLower(domain), nil
}

// keyFile holds the key that seals the passwords of relays, in the data
// directory of the server, beside server.db and outside the directory of
// every Cahut: a copy of a Cahut does not carry the password of its relay.
const keyFile = "mail.key"

// sealKey returns the key that seals the passwords of relays, drawn on
// first use when create is set.
func (s *Service) sealKey(create bool) ([]byte, error) {
	s.keyMu.Lock()
	defer s.keyMu.Unlock()
	if s.key != nil {
		return s.key, nil
	}
	root, err := os.OpenRoot(s.cfg.DataDir)
	if err != nil {
		return nil, err
	}
	defer root.Close()
	key, err := root.ReadFile(keyFile)
	if errors.Is(err, fs.ErrNotExist) && create {
		key = make([]byte, 32)
		if _, err := rand.Read(key); err != nil {
			return nil, err
		}
		f, err := root.OpenFile(keyFile, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
		if errors.Is(err, fs.ErrExist) {
			// Another process drew it first.
			key, err = root.ReadFile(keyFile)
		} else if err == nil {
			_, err = f.Write(key)
			err = errors.Join(err, f.Close())
		}
		if err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}
	if len(key) != 32 {
		return nil, fmt.Errorf("%s holds %d bytes, not 32", filepath.Join(s.cfg.DataDir, keyFile), len(key))
	}
	s.key = key
	return key, nil
}

// sealed binds a password to the host and the user name of its relay: a
// sealed password moved to another relay in the database does not open.
func sealedFor(host, user string) []byte {
	return []byte("cahut relay\x00" + host + "\x00" + user)
}

// seal seals the password of the relay at host, for user.
func (s *Service) seal(password, host, user string) ([]byte, error) {
	key, err := s.sealKey(true)
	if err != nil {
		return nil, err
	}
	aead, err := gcm(key)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	return aead.Seal(nonce, nonce, []byte(password), sealedFor(host, user)), nil
}

// errUnsealed means that the password of a relay cannot be read here.
var errUnsealed = errors.New("the password of the relay cannot be read on this server, which may not be the one where it was set: set the relay again")

// unseal opens the sealed password of the relay at host, for user.
func (s *Service) unseal(sealed []byte, host, user string) (string, error) {
	key, err := s.sealKey(false)
	if errors.Is(err, fs.ErrNotExist) {
		return "", errUnsealed
	}
	if err != nil {
		return "", err
	}
	aead, err := gcm(key)
	if err != nil {
		return "", err
	}
	if len(sealed) < aead.NonceSize() {
		return "", errUnsealed
	}
	pw, err := aead.Open(nil, sealed[:aead.NonceSize()], sealed[aead.NonceSize():], sealedFor(host, user))
	if err != nil {
		return "", errUnsealed
	}
	return string(pw), nil
}

func gcm(key []byte) (cipher.AEAD, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}
