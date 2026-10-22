// Package credential makes and checks the secrets that prove who calls a
// Cahut: passwords, hashed with Argon2id, and random tokens and codes, of
// which only a hash is kept. See docs/adr/0013-people-sessions-and-rights.md.
package credential

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base32"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"sync"
	"unicode/utf8"

	"golang.org/x/crypto/argon2"

	"cahut.cloud/internal/operation"
)

// Limits of a password. NIST SP 800-63B asks for 15 characters at least when
// a password is used without a second factor.
const (
	MinPassword = 15
	MaxPassword = 256
)

// The Argon2id parameters that OWASP recommends: 19 MiB of memory, 2 passes
// and 1 lane, with a salt of 16 bytes and a key of 32.
const (
	memory  = 19 * 1024 // KiB
	passes  = 2
	lanes   = 1
	saltLen = 16
	keyLen  = 32
)

// slots bounds the hashes computed at once, so that a burst of sign-ins
// cannot exhaust the memory of a small server: each takes 19 MiB.
var slots = make(chan struct{}, 4)

// CheckPolicy returns an error wrapping operation.ErrInvalid when password
// is too short or too long. There is no rule on what it is made of.
func CheckPolicy(password string) error {
	switch n := utf8.RuneCountInString(password); {
	case n < MinPassword:
		return operation.Invalidf("the password has %d characters: use at least %d", n, MinPassword)
	case n > MaxPassword:
		return operation.Invalidf("the password has %d characters: use at most %d", n, MaxPassword)
	}
	return nil
}

// HashPassword hashes password with Argon2id and a random salt, as a PHC
// string that records the parameters, such as
// $argon2id$v=19$m=19456,t=2,p=1$<salt>$<key>.
func HashPassword(ctx context.Context, password string) (string, error) {
	salt := Random(saltLen)
	key, err := derive(ctx, password, salt, memory, passes, lanes, keyLen)
	if err != nil {
		return "", err
	}
	b64 := base64.RawStdEncoding
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, memory, passes, lanes, b64.EncodeToString(salt), b64.EncodeToString(key)), nil
}

// CheckPassword reports whether password matches hash, a PHC string from
// HashPassword. It reads the parameters from hash, so that hashes made with
// other parameters still check.
func CheckPassword(ctx context.Context, hash, password string) (bool, error) {
	var version int
	var m, t uint32
	var p uint8
	parts := strings.Split(hash, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return false, errors.New("credential: not an Argon2id hash")
	}
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil || version != argon2.Version {
		return false, fmt.Errorf("credential: Argon2 version %q", parts[2])
	}
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &m, &t, &p); err != nil {
		return false, fmt.Errorf("credential: Argon2 parameters %q: %w", parts[3], err)
	}
	b64 := base64.RawStdEncoding
	salt, err := b64.DecodeString(parts[4])
	if err != nil {
		return false, fmt.Errorf("credential: salt: %w", err)
	}
	want, err := b64.DecodeString(parts[5])
	if err != nil {
		return false, fmt.Errorf("credential: key: %w", err)
	}
	if len(want) == 0 || len(want) > 1024 {
		return false, fmt.Errorf("credential: key of %d bytes", len(want))
	}
	got, err := derive(ctx, password, salt, m, t, p, uint32(len(want))) //nolint:gosec // at most 1024, checked above
	if err != nil {
		return false, err
	}
	return subtle.ConstantTimeCompare(got, want) == 1, nil
}

// decoy is the hash that CheckNoPassword checks against.
var decoy = sync.OnceValues(func() (string, error) {
	return HashPassword(context.Background(), string(Random(MinPassword)))
})

// CheckNoPassword spends the time of CheckPassword for someone who has no
// password, or no account, so that response times do not tell them apart.
func CheckNoPassword(ctx context.Context, password string) error {
	hash, err := decoy()
	if err != nil {
		return err
	}
	_, err = CheckPassword(ctx, hash, password)
	return err
}

func derive(ctx context.Context, password string, salt []byte, m, t uint32, p uint8, n uint32) ([]byte, error) {
	select {
	case slots <- struct{}{}:
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	defer func() { <-slots }()
	return argon2.IDKey([]byte(password), salt, t, m, p, n), nil
}

// Token returns a new random token: prefix, then 52 lowercase letters and
// digits that encode 32 random bytes. The prefix lets people and secret
// scanners recognize a token; it may be empty.
func Token(prefix string) string {
	return prefix + strings.ToLower(base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(Random(32)))
}

// Hash returns the SHA-256 hash of a token, which is all a Cahut keeps of it.
// Tokens hold 256 random bits, so they need no slow hash.
func Hash(token string) []byte {
	h := sha256.Sum256([]byte(token))
	return h[:]
}

// Random returns n random bytes.
func Random(n int) []byte {
	b := make([]byte, n)
	_, _ = rand.Read(b) // crypto/rand.Read never fails
	return b
}
