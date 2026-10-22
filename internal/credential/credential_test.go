package credential

import (
	"context"
	"encoding/base64"
	"errors"
	"regexp"
	"strings"
	"testing"

	"golang.org/x/crypto/argon2"

	"cahut.cloud/internal/operation"
)

func TestPasswords(t *testing.T) {
	ctx := context.Background()
	const password = "correct horse battery staple"
	hash, err := HashPassword(ctx, password)
	if err != nil {
		t.Fatal(err)
	}
	if !regexp.MustCompile(`^\$argon2id\$v=19\$m=19456,t=2,p=1\$[A-Za-z0-9+/]{22}\$[A-Za-z0-9+/]{43}$`).MatchString(hash) {
		t.Errorf("hash = %q", hash)
	}
	if again, _ := HashPassword(ctx, password); again == hash {
		t.Error("two hashes of a password are equal: the salt is not random")
	}
	for pw, want := range map[string]bool{password: true, password + " ": false, "": false} {
		ok, err := CheckPassword(ctx, hash, pw)
		if err != nil || ok != want {
			t.Errorf("CheckPassword(%q) = %v, %v, want %v", pw, ok, err, want)
		}
	}

	// A hash made with other parameters still checks.
	salt, b64 := []byte("saltsaltsaltsalt"), base64.RawStdEncoding
	weaker := "$argon2id$v=19$m=8,t=1,p=1$" + b64.EncodeToString(salt) + "$" +
		b64.EncodeToString(argon2.IDKey([]byte("x"), salt, 1, 8, 1, 16))
	if ok, err := CheckPassword(ctx, weaker, "x"); !ok || err != nil {
		t.Errorf("CheckPassword with other parameters = %v, %v", ok, err)
	}

	for _, bad := range []string{"", "$2a$10$abc", "$argon2i$v=19$m=8,t=1,p=1$c2FsdA$a2V5", "$argon2id$v=16$m=8,t=1,p=1$c2FsdA$a2V5",
		"$argon2id$v=19$m=8$c2FsdA$a2V5", "$argon2id$v=19$m=8,t=1,p=1$!!$a2V5", "$argon2id$v=19$m=8,t=1,p=1$c2FsdA$"} {
		if _, err := CheckPassword(ctx, bad, password); err == nil {
			t.Errorf("CheckPassword(%q) accepted a malformed hash", bad)
		}
	}
	if err := CheckNoPassword(ctx, password); err != nil {
		t.Errorf("CheckNoPassword: %v", err)
	}

	canceled, cancel := context.WithCancel(ctx)
	cancel()
	for range cap(slots) {
		slots <- struct{}{}
	}
	_, err = HashPassword(canceled, password)
	for range cap(slots) {
		<-slots
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("HashPassword with every slot taken and a canceled context: %v", err)
	}
}

func TestCheckPolicy(t *testing.T) {
	for pw, ok := range map[string]bool{
		strings.Repeat("a", 14):  false,
		strings.Repeat("a", 15):  true,
		strings.Repeat("é", 15):  true, // characters, not bytes
		strings.Repeat("é", 14):  false,
		strings.Repeat("a", 256): true,
		strings.Repeat("a", 257): false,
	} {
		err := CheckPolicy(pw)
		if (err == nil) != ok || err != nil && !errors.Is(err, operation.ErrInvalid) {
			t.Errorf("CheckPolicy(%d characters) = %v", len([]rune(pw)), err)
		}
	}
}

func TestTokens(t *testing.T) {
	tok := Token("cahut_x_")
	if !regexp.MustCompile(`^cahut_x_[a-z2-7]{52}$`).MatchString(tok) {
		t.Errorf("Token = %q", tok)
	}
	if a, b := Token(""), Token(""); a == b {
		t.Error("two tokens are equal")
	}
	if h := Hash(tok); len(h) != 32 || string(h) != string(Hash(tok)) || string(h) == string(Hash(tok+"x")) {
		t.Errorf("Hash = %x", h)
	}
}
