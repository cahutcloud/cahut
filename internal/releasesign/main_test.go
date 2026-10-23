package main

import (
	"crypto/ed25519"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"cahut.cloud/internal/release"
)

func privatePEM(t *testing.T, key ed25519.PrivateKey) string {
	t.Helper()
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der}))
}

func TestSign(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	file := filepath.Join(dir, "checksums.txt")
	if err := os.WriteFile(file, []byte("0123  cahut_0.4.0_linux_amd64.tar.gz\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := run(privatePEM(t, priv), file, file+".sig", []ed25519.PublicKey{pub}); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(file)
	sig, err := os.ReadFile(file + ".sig")
	if err != nil {
		t.Fatal(err)
	}
	if err := release.Verify([]ed25519.PublicKey{pub}, data, sig); err != nil {
		t.Errorf("the signature does not verify: %v", err)
	}
}

func TestSignRefusesAnUntrustedKey(t *testing.T) {
	_, priv, _ := ed25519.GenerateKey(nil)
	other, _, _ := ed25519.GenerateKey(nil)
	file := filepath.Join(t.TempDir(), "checksums.txt")
	if err := os.WriteFile(file, []byte("sums\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	err := run(privatePEM(t, priv), file, file+".sig", []ed25519.PublicKey{other})
	if err == nil || !strings.Contains(err.Error(), "release-keys.pem") {
		t.Errorf("err = %v, want a key that servers would refuse", err)
	}
	if _, err := os.Stat(file + ".sig"); !os.IsNotExist(err) {
		t.Error("a signature was written with an untrusted key")
	}
}

func TestSignNeedsAKey(t *testing.T) {
	for _, key := range []string{"", "not pem", "-----BEGIN PUBLIC KEY-----\nAAAA\n-----END PUBLIC KEY-----\n"} {
		if err := run(key, "checksums.txt", "checksums.txt.sig", release.Keys()); err == nil {
			t.Errorf("run with key %q succeeded", key)
		}
	}
}
