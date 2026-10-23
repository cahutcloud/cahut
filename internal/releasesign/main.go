// Command releasesign signs the checksum file of a release with the private
// key of the releases, which it reads, in PEM, from an environment variable.
// It writes the raw 64 bytes of the Ed25519 signature next to the file, then
// checks the signature against the public keys embedded in Cahut, so that a
// key that servers would not trust signs nothing. See
// docs/adr/0024-signed-releases-and-updates.md.
//
//	CAHUT_RELEASE_KEY="$(cat release-key.pem)" go run ./internal/releasesign dist/checksums.txt
//
// GoReleaser runs it in the release workflow, which reads the key from the
// secret of the same name. A key pair is made with OpenSSL:
//
//	openssl genpkey -algorithm ed25519 -out release-key.pem
//	openssl pkey -in release-key.pem -pubout >> internal/release/release-keys.pem
package main

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"flag"
	"fmt"
	"os"

	"cahut.cloud/internal/release"
)

func main() {
	keyEnv := flag.String("key-env", "CAHUT_RELEASE_KEY", "environment variable that holds the private key, in PEM")
	output := flag.String("output", "", "file of the signature (default: the file, then .sig)")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, "Usage: go run ./internal/releasesign [-key-env NAME] [-output FILE] <checksums.txt>")
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(2)
	}
	file := flag.Arg(0)
	if *output == "" {
		*output = file + ".sig"
	}
	if err := run(os.Getenv(*keyEnv), file, *output, release.Keys()); err != nil {
		fmt.Fprintf(os.Stderr, "releasesign: %v\n", err)
		os.Exit(1)
	}
}

// run signs file with the private key in keyPEM, writes the signature to
// output, and checks it against trusted.
func run(keyPEM, file, output string, trusted []ed25519.PublicKey) error {
	if keyPEM == "" {
		return errors.New("no private key: set the environment variable named by -key-env")
	}
	key, err := parseKey([]byte(keyPEM))
	if err != nil {
		return err
	}
	data, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	sig := ed25519.Sign(key, data)
	if err := release.Verify(trusted, data, sig); err != nil {
		return errors.New("the private key is not one of internal/release/release-keys.pem: servers would refuse what it signs")
	}
	return os.WriteFile(output, sig, 0o644) //nolint:gosec // a signature is public
}

// parseKey reads an Ed25519 private key in PKCS #8, as OpenSSL writes it.
func parseKey(data []byte) (ed25519.PrivateKey, error) {
	block, _ := pem.Decode(data)
	if block == nil || block.Type != "PRIVATE KEY" {
		return nil, errors.New("the private key is not a PEM block of type PRIVATE KEY")
	}
	key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("read the private key: %w", err)
	}
	ed, ok := key.(ed25519.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("a private key of type %T, not Ed25519", key)
	}
	return ed, nil
}
