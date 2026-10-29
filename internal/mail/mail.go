// Package mail sends the e-mail of the Cahuts of a server. It hands each
// message to the SMTP relay of its Cahut, which its administrators set,
// with TLS; a Cahut without a relay sends nothing, and the server has none.
// It never delivers to the servers of recipients itself. Messages leave
// from a queue in memory, once the call that makes them is committed, and
// nothing of them waits on disk. See docs/adr/0031-e-mail-through-a-relay.md
// and docs/adr/0034-e-mail-through-the-relay-of-each-cahut-only.md.
package mail

import (
	"cmp"
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"cahut.cloud/internal/store"
)

// Config says how the Cahuts of a server send e-mail.
type Config struct {
	// DataDir is the data directory of the server. It holds mail.key, the
	// key that seals the passwords of relays, and, for development, the
	// messages that no relay sends, in mail/.
	DataDir string
	// Dev writes each message to DataDir/mail/ when no relay sends it, and
	// logs its link, for development on one machine.
	Dev bool
	// Log reports the messages that could not be sent, and in development
	// those written to disk. Nil discards.
	Log *slog.Logger
}

// Service sends the e-mail of the Cahuts of a server.
type Service struct {
	cfg   Config
	log   *slog.Logger
	helo  string
	roots *x509.CertPool // nil: the system's
	// dialAnywhere lets the relays of Cahuts be on any address: tests only.
	dialAnywhere bool

	keyMu sync.Mutex
	key   []byte

	mu      sync.Mutex
	status  map[string]*Status // by the domain of a Cahut
	jobs    chan *job
	timers  map[*time.Timer]struct{}
	closed  bool
	work    sync.WaitGroup
	stop    context.Context
	halt    context.CancelFunc
	retries []time.Duration
}

// workers is how many messages go out at once.
const workers = 2

// queued is how many messages wait in the queue at most. A message that
// finds it full is not sent.
const queued = 500

// New returns the service that sends the e-mail of the server described by
// cfg, and starts its queue.
func New(cfg Config) *Service {
	log := cfg.Log
	if log == nil {
		log = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	s := &Service{
		cfg:     cfg,
		log:     log,
		helo:    heloName(),
		status:  map[string]*Status{},
		jobs:    make(chan *job, queued),
		timers:  map[*time.Timer]struct{}{},
		retries: []time.Duration{time.Minute, 5 * time.Minute, 10 * time.Minute},
	}
	s.stop, s.halt = context.WithCancel(context.Background())
	for range workers {
		s.work.Add(1)
		go func() {
			defer s.work.Done()
			for j := range s.jobs {
				s.attempt(j)
			}
		}()
	}
	return s
}

// Close sends the messages in the queue, until ctx is done, then stops the
// queue. The messages that wait to be tried again are given up.
func (s *Service) Close(ctx context.Context) error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil
	}
	s.closed = true
	for t := range s.timers {
		t.Stop()
	}
	close(s.jobs)
	s.mu.Unlock()
	done := make(chan struct{})
	go func() {
		s.work.Wait()
		close(done)
	}()
	select {
	case <-done:
		s.halt()
		return nil
	case <-ctx.Done():
		s.halt()
		<-done
		return errors.New("some e-mail was not sent before the server stopped")
	}
}

// UseRoots makes s trust the certificates that roots holds, in place of
// those of the system. Only tests call it, for a relay with a certificate
// of their own.
func (s *Service) UseRoots(roots *x509.CertPool) {
	s.roots = roots
}

// DialAnywhere lets the relays of Cahuts be on any address, loopback
// included. Only tests call it, to reach a relay on this machine; the
// other checks stay.
func (s *Service) DialAnywhere() {
	s.dialAnywhere = true
}

// Status says how the e-mail of a Cahut went out lately.
// The server keeps it in memory: it starts empty with the service.
type Status struct {
	LastSent    *time.Time `json:"last_sent,omitempty" jsonschema:"when a message last went out"`
	LastError   string     `json:"last_error,omitempty" jsonschema:"why the last message that failed did not go out"`
	LastErrorAt *time.Time `json:"last_error_at,omitempty" jsonschema:"when it failed"`
}

// statusOf returns how the e-mail of sender went out lately.
func (s *Service) statusOf(sender string) Status {
	s.mu.Lock()
	defer s.mu.Unlock()
	if st, ok := s.status[sender]; ok {
		return *st
	}
	return Status{}
}

// record notes that a message of sender went out, or failed with err.
func (s *Service) record(sender string, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.note(sender, err)
}

// note is record, with s.mu held.
func (s *Service) note(sender string, err error) {
	now := time.Now().UTC().Truncate(time.Second)
	st, ok := s.status[sender]
	if !ok {
		st = &Status{}
		s.status[sender] = st
	}
	if err == nil {
		st.LastSent = &now
		return
	}
	st.LastError, st.LastErrorAt = cut(err.Error(), 300), &now
}

// From says which Cahut a message comes from.
type From struct {
	// DB is the database of the Cahut, which keeps its relay.
	DB *store.DB
	// Domain is the domain of the Cahut.
	Domain string
	// Name is the name messages come from when their relay names none: the
	// domain of the Cahut.
	Name string
}

// ErrNoRelay means that the Cahut has no relay: the message does not go out.
var ErrNoRelay = errors.New("this Cahut sends no e-mail: its administrators set its relay, on the page E-mail of the console")

// Post hands l to the queue, which sends it through the relay of from. It
// returns ErrNoRelay when there is none, and in development writes the
// message to disk instead. A call posts its messages once it is committed,
// with operation.Call.AfterCommit, so that a call that fails or knocks
// sends nothing.
func (s *Service) Post(ctx context.Context, from From, l Letter) error {
	r, err := s.relayOf(ctx, from)
	if err != nil {
		return err
	}
	if r == nil {
		if s.cfg.Dev {
			return s.write(from, l)
		}
		return ErrNoRelay
	}
	msg, err := build(r, from, l, time.Now())
	if err != nil {
		return err
	}
	j := &job{sender: from.Domain, relay: *r, to: l.To, msg: msg}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return errors.New("the server is stopping: the message was not sent")
	}
	select {
	case s.jobs <- j:
		return nil
	default:
		err := fmt.Errorf("%d messages already wait to go out: this one was not sent", queued)
		s.note(from.Domain, err)
		return err
	}
}

// job is a message in the queue.
type job struct {
	sender string // the domain of the Cahut
	relay  Relay
	to     string
	msg    []byte
	tries  int
}

// attempt sends j, and puts it back in the queue later when the relay
// could not take it for now.
func (s *Service) attempt(j *job) {
	ctx, cancel := context.WithTimeout(s.stop, timeout)
	defer cancel()
	err := s.deliver(ctx, &j.relay, j.to, j.msg)
	s.record(j.sender, err)
	if err == nil {
		return
	}
	who := j.sender
	if !temporary(err) || j.tries >= len(s.retries) {
		s.log.Warn("send e-mail", "for", who, "error", err)
		return
	}
	s.log.Info("send e-mail, to try again", "for", who, "error", err)
	delay := s.retries[j.tries]
	j.tries++
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return
	}
	var t *time.Timer
	t = time.AfterFunc(delay, func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		delete(s.timers, t)
		if s.closed {
			return
		}
		select {
		case s.jobs <- j:
		default:
			s.log.Warn("send e-mail", "for", who, "error", "the queue is full")
		}
	})
	s.timers[t] = struct{}{}
}

// write writes the message l of from to the data directory, in
// development, and logs its link.
func (s *Service) write(from From, l Letter) error {
	r := &Relay{From: "cahut@" + cmp.Or(from.Domain, "localhost")}
	msg, err := build(r, from, l, time.Now())
	if err != nil {
		return err
	}
	dir := filepath.Join(s.cfg.DataDir, "mail")
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return err
	}
	name := filepath.Join(dir, time.Now().UTC().Format("20060102-150405.000000000")+".eml")
	if err := os.WriteFile(name, msg, 0o600); err != nil {
		return err
	}
	args := []any{"to", l.To, "subject", l.Subject, "file", name}
	if l.Action != nil {
		args = append(args, "link", l.Action.URL)
	}
	s.log.Info("e-mail written, since no relay sends it", args...)
	return nil
}

// heloName returns the name the server gives itself to relays: its host
// name when it is a full name, or else localhost.
func heloName() string {
	name, err := os.Hostname()
	if err != nil || !strings.Contains(name, ".") || net.ParseIP(name) != nil {
		return "localhost"
	}
	return name
}

// cut cuts s to n bytes at most, on a rune.
func cut(s string, n int) string {
	if len(s) <= n {
		return s
	}
	for n > 0 && !utf8.RuneStart(s[n]) {
		n--
	}
	return s[:n] + "…"
}
