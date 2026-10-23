package release

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"os"
	"strings"
	"testing"
)

func TestEmbeddedKeys(t *testing.T) {
	if n := len(Keys()); n == 0 {
		t.Fatal("no embedded key")
	}
}

// The keys written in install.sh are the embedded ones, so that the script
// and the updater trust the same releases.
func TestInstallScriptHasTheKeys(t *testing.T) {
	script, err := os.ReadFile("../../install.sh")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(script, KeysPEM()) {
		t.Errorf("install.sh does not hold the keys of internal/release/release-keys.pem:\n%s", KeysPEM())
	}
}

func newKey(t *testing.T) (ed25519.PublicKey, ed25519.PrivateKey) {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	return pub, priv
}

func TestVerify(t *testing.T) {
	pub, priv := newKey(t)
	other, otherPriv := newKey(t)
	sums := []byte("0123  cahut_0.4.0_linux_amd64.tar.gz\n")
	sig := ed25519.Sign(priv, sums)

	if err := Verify([]ed25519.PublicKey{pub}, sums, sig); err != nil {
		t.Errorf("a good signature: %v", err)
	}
	if err := Verify([]ed25519.PublicKey{other, pub}, sums, sig); err != nil {
		t.Errorf("a good signature by the second key: %v", err)
	}
	for name, c := range map[string]struct {
		keys []ed25519.PublicKey
		sums []byte
		sig  []byte
	}{
		"another key":     {[]ed25519.PublicKey{other}, sums, sig},
		"changed file":    {[]ed25519.PublicKey{pub}, append([]byte("f"), sums...), sig},
		"short signature": {[]ed25519.PublicKey{pub}, sums, sig[:10]},
		"signed by other": {[]ed25519.PublicKey{pub}, sums, ed25519.Sign(otherPriv, sums)},
	} {
		if err := Verify(c.keys, c.sums, c.sig); !errors.Is(err, ErrSignature) {
			t.Errorf("%s: err = %v, want ErrSignature", name, err)
		}
	}
}

func TestParseKeys(t *testing.T) {
	a, _ := newKey(t)
	b, _ := newKey(t)
	var data []byte
	for _, key := range []ed25519.PublicKey{a, b} {
		der, err := x509.MarshalPKIXPublicKey(key)
		if err != nil {
			t.Fatal(err)
		}
		data = append(data, pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der})...)
	}
	keys, err := ParseKeys(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(keys) != 2 || !keys[0].Equal(a) || !keys[1].Equal(b) {
		t.Errorf("ParseKeys returned %d keys, not the two given", len(keys))
	}
	for _, bad := range []string{"", "not pem", "-----BEGIN PRIVATE KEY-----\nAAAA\n-----END PRIVATE KEY-----\n"} {
		if _, err := ParseKeys([]byte(bad)); err == nil {
			t.Errorf("ParseKeys(%q) succeeded", bad)
		}
	}
}

func TestCheck(t *testing.T) {
	archive := []byte("the archive")
	s := sha256.Sum256(archive)
	sums := []byte(hex.EncodeToString(s[:]) + "  cahut_0.4.0_linux_amd64.tar.gz\n" +
		strings.Repeat("0", 64) + "  cahut_0.4.0_linux_arm64.tar.gz\n")

	if err := Check(sums, "cahut_0.4.0_linux_amd64.tar.gz", archive); err != nil {
		t.Errorf("the right archive: %v", err)
	}
	if err := Check(sums, "cahut_0.4.0_linux_arm64.tar.gz", archive); !errors.Is(err, ErrChecksum) {
		t.Errorf("another archive: err = %v, want ErrChecksum", err)
	}
	if err := Check(sums, "cahut_0.4.0_linux_riscv64.tar.gz", archive); err == nil || errors.Is(err, ErrChecksum) {
		t.Errorf("an archive the file does not name: err = %v", err)
	}
	if err := Check([]byte("zz  a.tar.gz\n"), "a.tar.gz", archive); err == nil {
		t.Error("a checksum that is not hex passed")
	}
}
