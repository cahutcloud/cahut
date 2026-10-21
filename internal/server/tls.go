package server

import (
	"context"

	"github.com/caddyserver/certmagic"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"cahut.cloud/internal/window"
)

// certificates sets up CertMagic for the home and its Windows (ADR 2). Every
// name is managed: its certificate is obtained when the home starts or the
// Window is published, then renewed. Failures are retried in the background.
func (s *Server) certificates(ctx context.Context) (*certmagic.Config, *certmagic.ACMEIssuer, error) {
	var magic *certmagic.Config
	log := certLogger()
	cache := certmagic.NewCache(certmagic.CacheOptions{
		GetConfigForCert: func(certmagic.Certificate) (*certmagic.Config, error) { return magic, nil },
		Logger:           log,
	})
	magic = certmagic.New(cache, certmagic.Config{
		Storage: &certmagic.FileStorage{Path: s.cfg.CertDir},
		Logger:  log,
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

	names, err := s.windowNames(ctx)
	if err != nil {
		return nil, nil, err
	}
	if err := magic.ManageAsync(ctx, append([]string{s.cfg.Domain}, names...)); err != nil {
		return nil, nil, err
	}
	s.home.Windows.Published = func(name string) {
		if err := magic.ManageAsync(ctx, []string{name + "." + s.cfg.Domain}); err != nil {
			s.log.Error("manage the certificate of a new Window", "window", name, "error", err)
		}
	}
	return magic, issuer, nil
}

// windowNames returns the host names of all the Windows.
func (s *Server) windowNames(ctx context.Context) ([]string, error) {
	names, err := window.Names(ctx, s.home.DB)
	if err != nil {
		return nil, err
	}
	for i, n := range names {
		names[i] = n + "." + s.cfg.Domain
	}
	return names, nil
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
