package web

import (
	"context"
	"crypto/tls"
	"fmt"

	"github.com/caddyserver/certmagic"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"cahut.cloud/internal/address"
)

// certs are the certificates of the Cahuts: those of domains, and those of
// IP addresses, which Let's Encrypt issues under its short-lived profile
// only. Both share one cache and one storage, so the issuer of domains
// answers the HTTP challenges of both.
type certs struct {
	names, ips *certmagic.Config
	issuer     *certmagic.ACMEIssuer // of domains
}

// shortLived is the ACME profile of certificates for IP addresses, valid
// 160 hours. See docs/adr/0026-cahut-at-an-ip-address.md.
const shortLived = "shortlived"

// certificates sets up CertMagic for the Cahuts (ADR 2, ADR 8 and ADR 26):
// one certificate per Cahut, obtained at start for those already declared,
// then renewed. A Cahut declared while the server runs gets its certificate
// at its first TLS handshake. Failures are retried in the background.
func (s *Server) certificates(ctx context.Context, domains []string) (*certs, error) {
	c := &certs{}
	log := certLogger()
	cache := certmagic.NewCache(certmagic.CacheOptions{
		GetConfigForCert: func(cert certmagic.Certificate) (*certmagic.Config, error) {
			for _, name := range cert.Names {
				if address.IsIP(name) {
					return c.ips, nil
				}
			}
			return c.names, nil
		},
		Logger: log,
	})
	var names, ips []string
	for _, d := range domains {
		if address.IsIP(d) {
			ips = append(ips, d)
		} else {
			names = append(names, d)
		}
	}
	ca := s.cfg.CA
	switch ca {
	case "":
		ca = certmagic.LetsEncryptProductionCA
	case "staging":
		ca = certmagic.LetsEncryptStagingCA
	}
	newConfig := func(profile string) (*certmagic.Config, *certmagic.ACMEIssuer) {
		cfg := certmagic.New(cache, certmagic.Config{
			Storage: &certmagic.FileStorage{Path: s.cfg.CertDir},
			Logger:  log,
			// Without a decision function, CertMagic would obtain a
			// certificate for any name a client sends.
			OnDemand: &certmagic.OnDemandConfig{DecisionFunc: s.declared},
		})
		issuer := certmagic.NewACMEIssuer(cfg, certmagic.ACMEIssuer{
			CA: ca, Email: s.cfg.Email, Agreed: true, Logger: log,
			Profile: profile,
			// An IP address is checked over HTTP only: TLS-ALPN would name
			// it in reverse DNS, which the configuration of domains serves.
			DisableTLSALPNChallenge: profile == shortLived,
		})
		cfg.Issuers = []certmagic.Issuer{issuer}
		return cfg, issuer
	}
	c.names, c.issuer = newConfig("")
	c.ips, _ = newConfig(shortLived)
	// A client that opens an IP address sends no server name. Behind NAT,
	// the local address of the connection is not the public one: the
	// address of the Cahut stands for it.
	if len(ips) > 0 {
		c.ips.DefaultServerName = ips[0]
	}
	if err := c.names.ManageAsync(ctx, names); err != nil {
		return nil, err
	}
	if err := c.ips.ManageAsync(ctx, ips); err != nil {
		return nil, err
	}
	return c, nil
}

// tlsConfig serves the certificate of a domain by the server name that the
// client sends, and that of the IP address when it sends none.
func (c *certs) tlsConfig() *tls.Config {
	cfg := c.names.TLSConfig()
	cfg.GetCertificate = func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
		if hello.ServerName == "" || address.IsIP(hello.ServerName) {
			return c.ips.GetCertificate(hello)
		}
		return c.names.GetCertificate(hello)
	}
	cfg.NextProtos = append([]string{"h2", "http/1.1"}, cfg.NextProtos...)
	return cfg
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
