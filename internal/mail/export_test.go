package mail

import (
	"time"
)

// SetRetries sets when s tries again a message that a relay could not take
// for now.
func SetRetries(s *Service, retries ...time.Duration) {
	s.retries = retries
}

// SetGreeting sets how long a relay has to greet.
func SetGreeting(d time.Duration) {
	greeting = d
}

// Brand returns the colors of the brand that e-mail uses, by name, and the
// logo.
func Brand() (map[string]string, []byte) {
	return brand, logo
}

// Seal seals password for the relay at host, with the key of s.
func Seal(s *Service, password, host, user string) ([]byte, error) {
	return s.seal(password, host, user)
}

// Unseal opens what Seal sealed, with the key of s.
func Unseal(s *Service, sealed []byte, host, user string) (string, error) {
	return s.unseal(sealed, host, user)
}

// Build makes the message l of from, sent through the relay r at now.
func Build(r *Relay, from From, l Letter, now time.Time) ([]byte, error) {
	return build(r, from, l, now)
}
