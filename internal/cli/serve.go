package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"cahut.cloud/internal/server"
	"cahut.cloud/internal/update"
	"cahut.cloud/internal/web"
)

func runServe(args []string, _, stderr io.Writer) int {
	fs := newFlags("serve", stderr)
	data := dataFlag(fs)
	var cfg web.Config
	fs.StringVar(&cfg.Email, "email", os.Getenv("CAHUT_ACME_EMAIL"),
		"contact of the Let's Encrypt account, optional (env CAHUT_ACME_EMAIL)")
	fs.StringVar(&cfg.CA, "acme-ca", os.Getenv("CAHUT_ACME_CA"),
		`ACME directory, or "staging" for Let's Encrypt's staging environment (env CAHUT_ACME_CA)`)
	fs.StringVar(&cfg.HTTPAddr, "http", ":80", "address for HTTP, which answers ACME challenges and redirects")
	fs.StringVar(&cfg.HTTPSAddr, "https", ":443", "address for HTTPS")
	fs.BoolVar(&cfg.Dev, "dev", false, "serve plain HTTP on --dev-addr, for Cahuts at localhost and *.localhost")
	fs.StringVar(&cfg.DevAddr, "dev-addr", fmt.Sprintf("localhost:%d", web.DevPort), "address for --dev")
	updates := fs.String("update-state", envOr("CAHUT_UPDATE_STATE", update.DefaultStateDir),
		"directory where the updater leaves the state of updates (env CAHUT_UPDATE_STATE)")
	if code, ok := parse(fs, args); !ok {
		return code
	}
	cfg.CertDir = filepath.Join(*data, "certificates")

	log := slog.New(slog.NewTextHandler(stderr, nil))
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	srv, err := server.Open(ctx, server.Config{DataDir: *data, BaseURL: cfg.BaseURL, ConsoleURL: cfg.ConsoleURL, UpdateDir: *updates})
	if err != nil {
		log.Error("open the server", "error", err)
		return exitError
	}
	watched := make(chan struct{})
	go func() {
		defer close(watched)
		srv.WatchUpdates(ctx, log)
	}()
	w := web.New(srv, cfg, log)
	err = w.Run(ctx)
	stop()
	<-watched
	err = errors.Join(err, w.Close(), srv.Close())
	if err != nil {
		log.Error("serve", "error", err)
		return exitError
	}
	return exitOK
}

// envOr returns the environment variable name, or else value.
func envOr(name, value string) string {
	if v := os.Getenv(name); v != "" {
		return v
	}
	return value
}
