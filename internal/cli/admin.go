package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
	"time"

	"cahut.cloud/internal/account"
	"cahut.cloud/internal/cahut"
	"cahut.cloud/internal/logbook"
	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/server"
	"cahut.cloud/internal/web"
)

// admin is the actor of every command run on the server: whoever can run
// them can already read the server's data directory.
var admin = operation.Actor{Kind: operation.Admin, Via: "cli"}

var domainCommands = []command{
	{name: "add", summary: "Declare a domain, which creates its Cahut", run: runDomainAdd},
	{name: "list", summary: "List the Cahuts of this server", run: runDomainList},
}

var userCommands = []command{
	{name: "create", summary: "Create a person, with their personal Room and an activation link", run: runUserCreate},
	{name: "list", summary: "List the people of a Cahut", run: runUserList},
	{name: "link", summary: "Give a person a new link where they choose their password", run: runUserLink},
}

var tokenCommands = []command{
	{name: "create", summary: "Create a personal token for an AI", run: runTokenCreate},
	{name: "list", summary: "List personal tokens", run: runTokenList},
	{name: "revoke", summary: "Revoke a personal token", run: runTokenRevoke},
}

// onServer opens the server in dir, runs op there as the administrator and
// closes the server. It reports errors on stderr.
func onServer[In, Out any](dir string, op *operation.Op[In, Out], in In, stderr io.Writer) (Out, bool) {
	return withServer(dir, stderr, func(ctx context.Context, srv *server.Server) (Out, error) {
		return operation.Do(ctx, srv.Registry, op, admin, in)
	})
}

// inCahut opens the server in dir and its Cahut at domain, runs op in that
// Cahut as the administrator and closes both. The domain may be empty when
// the server holds a single Cahut. It reports errors on stderr.
func inCahut[In, Out any](dir, domain string, op *operation.Op[In, Out], in In, stderr io.Writer) (Out, string, bool) {
	var at string
	out, ok := withServer(dir, stderr, func(ctx context.Context, srv *server.Server) (Out, error) {
		var zero Out
		if domain == "" {
			domains, err := srv.Domains(ctx)
			if err != nil {
				return zero, err
			}
			switch len(domains) {
			case 0:
				return zero, errors.New("this server holds no Cahut yet: declare one with cahut domain add <domain>")
			case 1:
				domain = domains[0]
			default:
				return zero, fmt.Errorf("this server holds %d Cahuts: choose one with --domain (cahut domain list shows them)", len(domains))
			}
		}
		c, err := srv.OpenCahut(ctx, strings.ToLower(domain))
		if err != nil {
			return zero, err
		}
		defer c.Close()
		at = c.Domain
		return operation.Do(ctx, c.Registry, op, admin, in)
	})
	return out, at, ok
}

func withServer[Out any](dir string, stderr io.Writer, fn func(context.Context, *server.Server) (Out, error)) (Out, bool) {
	ctx := context.Background()
	var zero Out
	srv, err := server.Open(ctx, server.Config{DataDir: dir})
	if err != nil {
		fmt.Fprintf(stderr, "cahut: %v\n", err)
		return zero, false
	}
	defer srv.Close()
	out, err := fn(ctx, srv)
	if err != nil {
		fmt.Fprintf(stderr, "cahut: %v\n", err)
		return zero, false
	}
	return out, true
}

// domainFlag adds the --domain flag, which chooses the Cahut a command acts
// in.
func domainFlag(fs *flag.FlagSet) *string {
	return fs.String("domain", "", "the Cahut to act in, by its domain; optional when the server holds a single Cahut")
}

func runDomainAdd(args []string, stdout, stderr io.Writer) int {
	fs := newFlags("domain add", stderr)
	data := dataFlag(fs)
	fs.Usage = func() {
		fmt.Fprintln(stderr, "Usage: cahut domain add [--data DIR] <domain>")
		fs.PrintDefaults()
	}
	rest, code, ok := parseWithArgs(fs, args, 1)
	if !ok {
		return code
	}
	info, ok := withServer(*data, stderr, func(ctx context.Context, srv *server.Server) (server.Info, error) {
		return operation.Do(ctx, srv.Registry, srv.AddDomain, admin, server.AddDomainInput{Domain: rest[0]})
	})
	if !ok {
		return exitError
	}
	fmt.Fprintf(stdout, "Declared %s: its Cahut is ready. Point its DNS to this server, then serve it.\n", info.Domain)
	return exitOK
}

func runDomainList(args []string, stdout, stderr io.Writer) int {
	fs := newFlags("domain list", stderr)
	data := dataFlag(fs)
	if code, ok := parse(fs, args); !ok {
		return code
	}
	list, ok := onServer(*data, server.ListDomains, server.ListDomainsInput{}, stderr)
	if !ok {
		return exitError
	}
	tw := tabwriter.NewWriter(stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(tw, "ID\tDOMAIN\tCREATED")
	for _, c := range list.Cahuts {
		fmt.Fprintf(tw, "%d\t%s\t%s\n", c.ID, c.Domain, day(c.CreatedAt))
	}
	if err := tw.Flush(); err != nil {
		return exitError
	}
	return exitOK
}

func runUserCreate(args []string, stdout, stderr io.Writer) int {
	fs := newFlags("user create", stderr)
	data := dataFlag(fs)
	domain := domainFlag(fs)
	var in account.CreateUserInput
	fs.StringVar(&in.Handle, "handle", "", "short name, which also names the personal Room: 2 to 32 lowercase letters, digits and hyphens")
	fs.StringVar(&in.Name, "name", "", "full name")
	fs.BoolVar(&in.Admin, "admin", false, "make the person an administrator of the Cahut")
	if code, ok := parse(fs, args); !ok {
		return code
	}
	u, at, ok := inCahut(*data, *domain, account.CreateUser, in, stderr)
	if !ok {
		return exitError
	}
	fmt.Fprintf(stdout, "Created %s (%s), %s of %s, with a personal Room at %s/%s.\n\n", u.Handle, u.Name, u.Role, at, at, u.PersonalRoom)
	printActivation(stdout, at, *u.Activation)
	return exitOK
}

func runUserList(args []string, stdout, stderr io.Writer) int {
	fs := newFlags("user list", stderr)
	data := dataFlag(fs)
	domain := domainFlag(fs)
	if code, ok := parse(fs, args); !ok {
		return code
	}
	list, _, ok := inCahut(*data, *domain, account.ListUsers, account.ListUsersInput{}, stderr)
	if !ok {
		return exitError
	}
	tw := tabwriter.NewWriter(stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(tw, "HANDLE\tNAME\tROLE\tSTATUS")
	for _, u := range list.Users {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", u.Handle, u.Name, u.Role, u.Status)
	}
	if err := tw.Flush(); err != nil {
		return exitError
	}
	return exitOK
}

func runUserLink(args []string, stdout, stderr io.Writer) int {
	fs := newFlags("user link", stderr)
	data := dataFlag(fs)
	domain := domainFlag(fs)
	var in account.UserInput
	fs.StringVar(&in.User, "handle", "", "the person who chooses a new password")
	if code, ok := parse(fs, args); !ok {
		return code
	}
	a, at, ok := inCahut(*data, *domain, account.CreateActivationLink, in, stderr)
	if !ok {
		return exitError
	}
	printActivation(stdout, at, a)
	fmt.Fprintln(stdout, "Their password, if they have one, works until they use the link.")
	return exitOK
}

// printActivation prints the link of an activation, in the console of the
// Cahut at domain.
func printActivation(w io.Writer, domain string, a account.Activation) {
	fmt.Fprintf(w, "Send %s this link, where they choose their password. It works once, until %s, and is shown only now.\n\n",
		a.User, day(a.ExpiresAt))
	fmt.Fprintf(w, "%sactivate#%s\n\n", web.ConsoleURL(domain), a.Code)
}

func runTokenCreate(args []string, stdout, stderr io.Writer) int {
	fs := newFlags("token create", stderr)
	data := dataFlag(fs)
	domain := domainFlag(fs)
	var in account.CreateTokenInput
	fs.StringVar(&in.User, "user", "", "handle of the person the AI acts for")
	fs.StringVar(&in.Name, "name", "", "what the token is for, such as \"Claude Code on my laptop\"")
	if code, ok := parse(fs, args); !ok {
		return code
	}
	t, at, ok := inCahut(*data, *domain, account.CreateToken, in, stderr)
	if !ok {
		return exitError
	}
	fmt.Fprintf(stdout, "Token %d for %s (%s) in %s. It is shown once: copy it now.\n\n%s\n\n", t.ID, t.User, t.Name, at, t.Token)
	fmt.Fprintln(stdout, "The AI sends it in the Authorization header: Bearer <token>.")
	return exitOK
}

func runTokenList(args []string, stdout, stderr io.Writer) int {
	fs := newFlags("token list", stderr)
	data := dataFlag(fs)
	domain := domainFlag(fs)
	var in account.ListTokensInput
	fs.StringVar(&in.User, "user", "", "only the tokens of this person")
	if code, ok := parse(fs, args); !ok {
		return code
	}
	list, _, ok := inCahut(*data, *domain, account.ListTokens, in, stderr)
	if !ok {
		return exitError
	}
	tw := tabwriter.NewWriter(stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(tw, "ID\tUSER\tNAME\tCREATED\tLAST USED\tSTATE")
	for _, t := range list.Tokens {
		state := "active"
		if t.RevokedAt != nil {
			state = "revoked " + day(*t.RevokedAt)
		}
		lastUsed := "never"
		if t.LastUsedAt != nil {
			lastUsed = minute(*t.LastUsedAt)
		}
		fmt.Fprintf(tw, "%d\t%s\t%s\t%s\t%s\t%s\n", t.ID, t.User, t.Name, day(t.CreatedAt), lastUsed, state)
	}
	if err := tw.Flush(); err != nil {
		return exitError
	}
	return exitOK
}

func runTokenRevoke(args []string, stdout, stderr io.Writer) int {
	fs := newFlags("token revoke", stderr)
	data := dataFlag(fs)
	domain := domainFlag(fs)
	var in account.RevokeTokenInput
	fs.Int64Var(&in.ID, "id", 0, "the token to revoke, as token list shows it")
	if code, ok := parse(fs, args); !ok {
		return code
	}
	t, _, ok := inCahut(*data, *domain, account.RevokeToken, in, stderr)
	if !ok {
		return exitError
	}
	fmt.Fprintf(stdout, "Revoked token %d of %s (%s).\n", t.ID, t.User, t.Name)
	return exitOK
}

func runLogbook(args []string, stdout, stderr io.Writer) int {
	fs := newFlags("logbook", stderr)
	data := dataFlag(fs)
	domain := domainFlag(fs)
	ofServer := fs.Bool("server", false, "print the Logbook of the server, not of a Cahut")
	var in cahut.ListLogbookInput
	fs.IntVar(&in.Limit, "limit", 20, "how many entries")
	if code, ok := parse(fs, args); !ok {
		return code
	}
	var page cahut.LogbookPage
	var ok bool
	if *ofServer {
		page, ok = onServer(*data, cahut.ListServerLogbook, in, stderr)
	} else {
		page, _, ok = inCahut(*data, *domain, cahut.ListLogbook, in, stderr)
	}
	if !ok {
		return exitError
	}
	tw := tabwriter.NewWriter(stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(tw, "TIME\tOPERATION\tWHO\tVIA\tTARGET\tOUTCOME")
	for _, e := range page.Entries {
		outcome := string(e.Outcome)
		if e.Error != "" {
			outcome += ": " + e.Error
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\n", minute(e.At), e.Operation, who(e), e.Via, e.Target, outcome)
	}
	if err := tw.Flush(); err != nil {
		return exitError
	}
	return exitOK
}

// who says who acted, as in "Claude Code for olivier".
func who(e logbook.Entry) string {
	switch {
	case e.Actor == string(operation.AI):
		return e.Client + " for " + e.UserHandle
	case e.UserHandle != "":
		return e.UserHandle
	}
	return e.Actor
}

func day(t time.Time) string    { return t.Local().Format("2006-01-02") }
func minute(t time.Time) string { return t.Local().Format("2006-01-02 15:04") }
