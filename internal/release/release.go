// Package release checks the releases of Cahut before a server installs
// them: an Ed25519 key signs the checksum file of each release, which names
// the SHA-256 of every archive. The public keys are embedded here and
// written in install.sh. See docs/adr/0024-signed-releases-and-updates.md.
package release

import (
	"bufio"
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"crypto/subtle"
	"crypto/x509"
	_ "embed"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"strings"
)

// Files of a release, next to its archives.
const (
	ChecksumsFile = "checksums.txt"
	SignatureFile = "checksums.txt.sig"
)

// keysPEM holds the public keys that sign releases, one PEM block each. A
// new key is added here, in a release signed by the old one, before it
// signs anything.
//
//go:embed release-keys.pem
var keysPEM []byte

// KeysPEM returns the public keys as they are embedded, for install.sh.
func KeysPEM() []byte { return keysPEM }

// Keys returns the public keys that sign releases.
func Keys() []ed25519.PublicKey {
	keys, err := ParseKeys(keysPEM)
	if err != nil {
		panic(err) // the embedded file is checked by the tests
	}
	return keys
}

// ParseKeys reads Ed25519 public keys from PEM blocks of type PUBLIC KEY.
func ParseKeys(data []byte) ([]ed25519.PublicKey, error) {
	var keys []ed25519.PublicKey
	for {
		var block *pem.Block
		block, data = pem.Decode(data)
		if block == nil {
			break
		}
		if block.Type != "PUBLIC KEY" {
			return nil, fmt.Errorf("a PEM block of type %q, not PUBLIC KEY", block.Type)
		}
		key, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		ed, ok := key.(ed25519.PublicKey)
		if !ok {
			return nil, fmt.Errorf("a public key of type %T, not Ed25519", key)
		}
		keys = append(keys, ed)
	}
	if len(keys) == 0 {
		return nil, errors.New("no public key")
	}
	return keys, nil
}

// ErrSignature is the error of a checksum file that no key signed.
var ErrSignature = errors.New("the signature of the release does not match the keys of Cahut")

// Verify checks that one of keys signed checksums with sig.
func Verify(keys []ed25519.PublicKey, checksums, sig []byte) error {
	if len(sig) != ed25519.SignatureSize {
		return fmt.Errorf("%w: a signature has %d bytes, not %d", ErrSignature, ed25519.SignatureSize, len(sig))
	}
	for _, key := range keys {
		if ed25519.Verify(key, checksums, sig) {
			return nil
		}
	}
	return ErrSignature
}

// ErrChecksum is the error of a file whose SHA-256 is not the one that the
// checksum file gives.
var ErrChecksum = errors.New("checksum mismatch")

// Check checks that content is the file name of the checksum file, which
// lines such as "<sha256 in hex>  <name>" make.
func Check(checksums []byte, name string, content []byte) error {
	want, err := sum(checksums, name)
	if err != nil {
		return err
	}
	got := sha256.Sum256(content)
	if subtle.ConstantTimeCompare(got[:], want) != 1 {
		return fmt.Errorf("%w for %s", ErrChecksum, name)
	}
	return nil
}

func sum(checksums []byte, name string) ([]byte, error) {
	lines := bufio.NewScanner(bytes.NewReader(checksums))
	for lines.Scan() {
		fields := strings.Fields(lines.Text())
		if len(fields) != 2 || strings.TrimPrefix(fields[1], "*") != name {
			continue
		}
		s, err := hex.DecodeString(fields[0])
		if err != nil || len(s) != sha256.Size {
			return nil, fmt.Errorf("the checksum of %s is not a SHA-256 in hex", name)
		}
		return s, nil
	}
	if err := lines.Err(); err != nil {
		return nil, err
	}
	return nil, fmt.Errorf("the checksum file does not name %s", name)
}
