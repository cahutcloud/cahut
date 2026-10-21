package web

import (
	"context"
	"fmt"

	"github.com/caddyserver/certmagic"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// certificates sets up CertMagic for the domains of the Cahuts (ADR 2 and
// ADR 8): one certificate per Cahut, obtained at start for the domains
// already declared, then renewed. A domain declared while the server runs
// gets its certificate at its first TLS handshake. Failures are retried in
// the background.
func (s *Server) certificates(ctx context.Context, domains []string) (*certmagic.Config, *certmagic.ACMEIssuer, error) {
	var magic *certmagic.Config
	log := certLogger()
	cache := certmagic.NewCache(certmagic.CacheOptions{
		GetConfigForCert: func(certmagic.Certificate) (*certmagic.Config, error) { return magic, nil },
		Logger:           log,
	})
	magic = certmagic.New(cache, certmagic.Config{
		Storage: &certmagic.FileStorage{Path: s.cfg.CertDir},
		Logger:  log,
		// Without a decision function, CertMagic would obtain a certificate
		// for any name a client sends.
		OnDemand: &certmagic.OnDemandConfig{DecisionFunc: s.declared},
	})
	ca := s.cfg.CA
	switch ca {
	case "":
		ca = certmagic.LetsEncryptProductionCA
	case "staging":
		ca = certmagic.LetsEncryptStagingCA
	}
	issuer := certmagic.NewACMEIssuer(magic, certmagic.ACMEIssuer{CA: ca, Email: s.cfg.Email, Agreed: true, Logger: log})
	magic.Issuers = []certmagic.Issuer{issuer}

	if err := magic.ManageAsync(ctx, domains); err != nil {
		return nil, nil, err
	}
	return magic, issuer, nil
}

// declared accepts the domains of the Cahuts only.
func (s *Server) declared(ctx context.Context, name string) error {
	ok, err := s.srv.Declared(ctx, name)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("%s is not the domain of a Cahut", name)
	}
	return nil
}

// certLogger returns a logger for CertMagic that writes lines like those of
// log/slog's text handler, with readable times, to stderr.
func certLogger() *zap.Logger {
	cfg := zap.NewProductionConfig()
	cfg.Encoding = "console"
	cfg.DisableCaller = true
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	log, err := cfg.Build()
	if err != nil {
		return zap.NewNop()
	}
	return log.Named("certificates")
}
