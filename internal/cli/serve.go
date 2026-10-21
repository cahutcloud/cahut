package cli

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"cahut.cloud/internal/home"
	"cahut.cloud/internal/server"
)

func runServe(args []string, _, stderr io.Writer) int {
	fs := newFlags("serve", stderr)
	data := dataFlag(fs)
	var cfg server.Config
	fs.StringVar(&cfg.Domain, "domain", os.Getenv("CAHUT_DOMAIN"),
		"address of the home, such as home.example.com; each Window is served under it (env CAHUT_DOMAIN)")
	fs.StringVar(&cfg.Email, "email", os.Getenv("CAHUT_ACME_EMAIL"),
		"contact of the Let's Encrypt account, optional (env CAHUT_ACME_EMAIL)")
	fs.StringVar(&cfg.CA, "acme-ca", os.Getenv("CAHUT_ACME_CA"),
		`ACME directory, or "staging" for Let's Encrypt's staging environment (env CAHUT_ACME_CA)`)
	fs.StringVar(&cfg.HTTPAddr, "http", ":80", "address for HTTP, which answers ACME challenges and redirects")
	fs.StringVar(&cfg.HTTPSAddr, "https", ":443", "address for HTTPS")
	fs.BoolVar(&cfg.Dev, "dev", false, "serve plain HTTP on --dev-addr, with the Windows under localhost")
	fs.StringVar(&cfg.DevAddr, "dev-addr", "localhost:8080", "address for --dev")
	if code, ok := parse(fs, args); !ok {
		return code
	}
	if !cfg.Dev && cfg.Domain == "" {
		fmt.Fprintln(stderr, "cahut serve: set --domain or CAHUT_DOMAIN, or use --dev")
		return exitUsage
	}
	cfg.CertDir = filepath.Join(*data, "certificates")

	log := slog.New(slog.NewTextHandler(stderr, nil))
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	h, err := home.Open(ctx, home.Config{DataDir: *data, WindowURL: cfg.WindowURL()})
	if err != nil {
		log.Error("open the home", "error", err)
		return exitError
	}
	defer h.Close()
	if err := server.New(h, cfg, log).Run(ctx); err != nil {
		log.Error("serve", "error", err)
		return exitError
	}
	return exitOK
}
