package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/server"
	"cahut.cloud/internal/web"
)

// runSetup prints the link to the first start of a Cahut that has nobody
// yet. See docs/adr/0025-first-start-in-the-browser.md.
func runSetup(args []string, stdout, stderr io.Writer) int {
	fs := newFlags("setup", stderr)
	data := dataFlag(fs)
	domain := domainFlag(fs)
	if code, ok := parse(fs, args); !ok {
		return code
	}
	link, ok := withServer(*data, stderr, func(ctx context.Context, srv *server.Server) (server.SetupLink, error) {
		d := strings.ToLower(*domain)
		if d == "" {
			domains, err := srv.Domains(ctx)
			if err != nil {
				return server.SetupLink{}, err
			}
			switch len(domains) {
			case 0:
				return server.SetupLink{}, errors.New("this server holds no Cahut yet: declare one with cahut domain add <domain>")
			case 1:
				d = domains[0]
			default:
				return server.SetupLink{}, fmt.Errorf("this server holds %d Cahuts: choose one with --domain (cahut domain list shows them)", len(domains))
			}
		}
		return operation.Do(ctx, srv.Registry, srv.CreateSetupLink, admin, server.CreateSetupLinkInput{Domain: d})
	})
	if !ok {
		return exitError
	}
	fmt.Fprintf(stdout, "Open this link to start your Cahut. It works once, until %s, and is shown only now.\n\n", day(link.ExpiresAt))
	fmt.Fprintf(stdout, "  %ssetup#%s\n\n", web.ConsoleURL(link.Domain), link.Code)
	return exitOK
}
