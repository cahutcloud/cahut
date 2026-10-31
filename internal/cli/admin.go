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
	"cahut.cloud/internal/address"
	"cahut.cloud/internal/cahut"
	"cahut.cloud/internal/logbook"
	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/server"
	"cahut.cloud/internal/settings"
	"cahut.cloud/internal/web"
)

// admin is the actor of every command run on the server: whoever can run
// them can already read the server's data directory.
var admin = operation.Actor{Kind: operation.Admin, Via: "cli"}

var domainCommands = []command{
	{name: "add", summary: "Declare a domain, which creates its Cahut", run: runDomainAdd},
	{name: "list", summary: "List the Cahuts of this server", run: runDomainList},
	{name: "rename", summary: "Move a Cahut to another domain, or from the IP address of the server to a domain", run: runDomainRename},
	{name: "suspend", summary: "Suspend a Cahut, which keeps everything and serves nothing until it is resumed", run: runDomainSuspend},
	{name: "resume", summary: "Resume a suspended Cahut", run: runDomainResume},
}

var adminCommands = []command{
	{name: "create", summary: "Create an administrator of the server, with an activation link", run: runAdminCreate},
	{name: "list", summary: "List the administrators of the server", run: runAdminList},
	{name: "rename", summary: "Change the name of an administrator of the server", run: runAdminRename},
	{name: "link", summary: "Give an administrator of the server a new link where they choose their password", run: runAdminLink},
	{name: "disable", summary: "Disable an administrator of the server: their sessions end, and they cannot sign in", run: runAdminDisable},
	{name: "enable", summary: "Enable a disabled administrator of the server again", run: runAdminEnable},
}

var userCommands = []command{
	{name: "create", summary: "Create a person, with their personal Room and an activation link", run: runUserCreate},
	{name: "list", summary: "List the people of a Cahut", run: runUserList},
	{name: "link", summary: "Give a person a new link where they choose their password", run: runUserLink},
	{name: "email", summary: "Give a person an e-mail address, where their links go, or remove it", run: runUserEmail},
}

var tokenCommands = []command{
	{name: "create", summary: "Create a personal token for an AI", run: runTokenCreate},
	{name: "list", summary: "List personal tokens", run: runTokenList},
	{name: "revoke", summary: "Revoke a personal token", run: runTokenRevoke},
}

var serviceTokenCommands = []command{
	{name: "create", summary: "Create a service token for a program on this server", run: runServiceTokenCreate},
	{name: "list", summary: "List service tokens", run: runServiceTokenList},
	{name: "revoke", summary: "Revoke a service token", run: runServiceTokenRevoke},
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
	return withCahut(dir, domain, stderr, func(ctx context.Context, c *cahut.Cahut) (Out, error) {
		return operation.Do(ctx, c.Registry, op, admin, in)
	})
}

// withCahut opens the server in dir and its Cahut at domain, runs fn with
// that Cahut and closes both, as inCahut does. It returns the domain of the
// Cahut.
func withCahut[Out any](dir, domain string, stderr io.Writer, fn func(context.Context, *cahut.Cahut) (Out, error)) (Out, string, bool) {
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
		return fn(ctx, c)
	})
	return out, at, ok
}

func withServer[Out any](dir string, stderr io.Writer, fn func(context.Context, *server.Server) (Out, error)) (Out, bool) {
	ctx := context.Background()
	var zero Out
	// The links that e-mail carries name the console of a Cahut, as those
	// that commands print do.
	srv, err := server.Open(ctx, server.Config{DataDir: dir, ConsoleURL: web.ConsoleURL})
	if err != nil {
		fmt.Fprintf(stderr, "cahut: %v\n", err)
		return zero, false
	}
	defer srv.Close()
	if mailForTests != nil {
		mailForTests(srv.Mail)
	}
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
	var first server.FirstAdmin
	fs.StringVar(&first.Handle, "admin-handle", "", "handle of the first administrator of the Cahut, a person of it, who gets an activation link")
	fs.StringVar(&first.Name, "admin-name", "", "full name of the first administrator")
	fs.Usage = func() {
		fmt.Fprintln(stderr, "Usage: cahut domain add [--data DIR] [--admin-handle HANDLE --admin-name NAME] <domain>")
		fs.PrintDefaults()
	}
	rest, code, ok := parseWithArgs(fs, args, 1)
	if !ok {
		return code
	}
	in := server.AddDomainInput{Domain: rest[0]}
	if first != (server.FirstAdmin{}) {
		in.Admin = &first
	}
	n, ok := withServer(*data, stderr, func(ctx context.Context, srv *server.Server) (server.NewCahut, error) {
		return operation.Do(ctx, srv.Registry, srv.AddDomain, admin, in)
	})
	if !ok {
		return exitError
	}
	if address.IsIP(n.Domain) {
		fmt.Fprintf(stdout, "Declared a Cahut at %s, without a domain. Move it to one later with cahut domain rename.\n", n.Domain)
	} else {
		fmt.Fprintf(stdout, "Declared %s: its Cahut is ready. Point its DNS to this server, then serve it.\n", n.Domain)
	}
	if u := n.Admin; u != nil {
		fmt.Fprintf(stdout, "\nCreated %s (%s), admin of %s, with a personal Room at %s/%s.\n\n", u.Handle, u.Name, n.Domain, n.Domain, u.PersonalRoom)
		printActivation(stdout, web.ActivationURL(n.Domain), *u.Activation)
	}
	return exitOK
}

func runDomainRename(args []string, stdout, stderr io.Writer) int {
	fs := newFlags("domain rename", stderr)
	data := dataFlag(fs)
	fs.Usage = func() {
		fmt.Fprintln(stderr, "Usage: cahut domain rename [--data DIR] <domain> <new domain>")
		fs.PrintDefaults()
	}
	rest, code, ok := parseWithArgs(fs, args, 2)
	if !ok {
		return code
	}
	info, ok := onServer(*data, server.RenameDomain, server.RenameDomainInput{Domain: rest[0], To: rest[1]}, stderr)
	if !ok {
		return exitError
	}
	fmt.Fprintf(stdout, "The Cahut of %s is now at https://%s/. Its links to %s stop working; AIs connect again at https://%s/mcp.\n",
		rest[0], info.Domain, rest[0], info.Domain)
	return exitOK
}

func runAdminCreate(args []string, stdout, stderr io.Writer) int {
	fs := newFlags("admin create", stderr)
	data := dataFlag(fs)
	var in server.CreateAdminInput
	fs.StringVar(&in.Handle, "handle", "", "short name of the administrator: 2 to 32 lowercase letters, digits and hyphens")
	fs.StringVar(&in.Name, "name", "", "full name")
	domain := fs.String("domain", "", "the Cahut whose console the link opens; the first one declared when empty")
	if code, ok := parse(fs, args); !ok {
		return code
	}
	type created struct {
		admin  server.NewAdmin
		domain string
	}
	c, ok := withServer(*data, stderr, func(ctx context.Context, srv *server.Server) (created, error) {
		d, err := consoleDomain(ctx, srv, *domain)
		if err != nil {
			return created{}, err
		}
		a, err := operation.Do(ctx, srv.Registry, server.CreateAdmin, admin, in)
		return created{admin: a, domain: d}, err
	})
	if !ok {
		return exitError
	}
	fmt.Fprintf(stdout, "Created %s (%s), administrator of this server.\n\n", c.admin.Handle, c.admin.Name)
	printAdminActivation(stdout, c.domain, c.admin.Activation)
	return exitOK
}

func runAdminList(args []string, stdout, stderr io.Writer) int {
	fs := newFlags("admin list", stderr)
	data := dataFlag(fs)
	if code, ok := parse(fs, args); !ok {
		return code
	}
	list, ok := onServer(*data, server.ListAdmins, server.ListAdminsInput{}, stderr)
	if !ok {
		return exitError
	}
	tw := tabwriter.NewWriter(stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(tw, "HANDLE\tNAME\tSTATUS\tCREATED")
	for _, a := range list.Admins {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n", a.Handle, a.Name, a.Status, day(a.CreatedAt))
	}
	if err := tw.Flush(); err != nil {
		return exitError
	}
	return exitOK
}

func runAdminRename(args []string, stdout, stderr io.Writer) int {
	fs := newFlags("admin rename", stderr)
	data := dataFlag(fs)
	var in server.RenameAdminInput
	fs.StringVar(&in.Admin, "handle", "", "the administrator of the server, whose handle stays")
	fs.StringVar(&in.Name, "name", "", "their new full name")
	if code, ok := parse(fs, args); !ok {
		return code
	}
	a, ok := onServer(*data, server.RenameAdmin, in, stderr)
	if !ok {
		return exitError
	}
	fmt.Fprintf(stdout, "Renamed %s (%s).\n", a.Handle, a.Name)
	return exitOK
}

func runAdminLink(args []string, stdout, stderr io.Writer) int {
	fs := newFlags("admin link", stderr)
	data := dataFlag(fs)
	var in server.AdminInput
	fs.StringVar(&in.Admin, "handle", "", "the administrator who chooses a new password")
	domain := fs.String("domain", "", "the Cahut whose console the link opens; the first one declared when empty")
	if code, ok := parse(fs, args); !ok {
		return code
	}
	type link struct {
		activation account.Activation
		domain     string
	}
	l, ok := withServer(*data, stderr, func(ctx context.Context, srv *server.Server) (link, error) {
		d, err := consoleDomain(ctx, srv, *domain)
		if err != nil {
			return link{}, err
		}
		a, err := operation.Do(ctx, srv.Registry, server.CreateActivationLink, admin, in)
		return link{activation: a, domain: d}, err
	})
	if !ok {
		return exitError
	}
	printAdminActivation(stdout, l.domain, l.activation)
	fmt.Fprintln(stdout, "Their password, if they have one, works until they use the link.")
	return exitOK
}

func runAdminDisable(args []string, stdout, stderr io.Writer) int {
	return runAdminSwitch("admin disable", server.DisableAdmin, "Disabled %s (%s): their sessions ended, and they cannot sign in.\n", args, stdout, stderr)
}

func runAdminEnable(args []string, stdout, stderr io.Writer) int {
	return runAdminSwitch("admin enable", server.EnableAdmin, "Enabled %s (%s) again.\n", args, stdout, stderr)
}

// runAdminSwitch runs op, which disables or enables the administrator that
// --handle names, and says what it did with done.
func runAdminSwitch(name string, op *operation.Op[server.AdminInput, server.Admin], done string, args []string, stdout, stderr io.Writer) int {
	fs := newFlags(name, stderr)
	data := dataFlag(fs)
	var in server.AdminInput
	fs.StringVar(&in.Admin, "handle", "", "the administrator of the server")
	if code, ok := parse(fs, args); !ok {
		return code
	}
	a, ok := onServer(*data, op, in, stderr)
	if !ok {
		return exitError
	}
	fmt.Fprintf(stdout, done, a.Handle, a.Name)
	return exitOK
}

// consoleDomain returns the domain whose console an administrator of the
// server opens a link at: domain, which a Cahut must have, or the domain of
// the first Cahut declared, or "" when there is none.
func consoleDomain(ctx context.Context, srv *server.Server, domain string) (string, error) {
	if domain != "" {
		domain = strings.ToLower(domain)
		ok, err := srv.Declared(ctx, domain)
		if err != nil {
			return "", err
		}
		if !ok {
			return "", fmt.Errorf("no Cahut at %s: cahut domain list shows them", domain)
		}
		return domain, nil
	}
	domains, err := srv.Domains(ctx)
	if err != nil || len(domains) == 0 {
		return "", err
	}
	return domains[0], nil
}

// runDomainSuspend suspends a Cahut. See
// docs/adr/0048-suspending-a-cahut.md.
func runDomainSuspend(args []string, stdout, stderr io.Writer) int {
	fs := newFlags("domain suspend", stderr)
	data := dataFlag(fs)
	reason := fs.String("reason", "", "why the Cahut is suspended, which its administrators read: an unpaid invoice, a decision of moderation")
	fs.Usage = func() {
		fmt.Fprintln(stderr, "Usage: cahut domain suspend [--data DIR] --reason TEXT <domain>")
		fs.PrintDefaults()
	}
	rest, code, ok := parseWithArgs(fs, args, 1)
	if !ok {
		return code
	}
	s, at, ok := inCahut(*data, rest[0], settings.Suspend, settings.SuspendCahutInput{Reason: *reason}, stderr)
	if !ok {
		return exitError
	}
	fmt.Fprintf(stdout, "The Cahut of %s is suspended since %s: it keeps everything, and serves nothing but to the administrators of the server.\n",
		at, minute(s.Suspension.Since))
	fmt.Fprintf(stdout, "Its administrators hear of it by e-mail when it has a relay. Resume it with cahut domain resume %s.\n", at)
	return exitOK
}

// runDomainResume resumes a suspended Cahut.
func runDomainResume(args []string, stdout, stderr io.Writer) int {
	fs := newFlags("domain resume", stderr)
	data := dataFlag(fs)
	fs.Usage = func() {
		fmt.Fprintln(stderr, "Usage: cahut domain resume [--data DIR] <domain>")
		fs.PrintDefaults()
	}
	rest, code, ok := parseWithArgs(fs, args, 1)
	if !ok {
		return code
	}
	_, at, ok := inCahut(*data, rest[0], settings.Resume, settings.ResumeCahutInput{}, stderr)
	if !ok {
		return exitError
	}
	fmt.Fprintf(stdout, "The Cahut of %s is open again: its Windows, its people and their AIs find everything as it was.\n", at)
	return exitOK
}

func runDomainList(args []string, stdout, stderr io.Writer) int {
	fs := newFlags("domain list", stderr)
	data := dataFlag(fs)
	if code, ok := parse(fs, args); !ok {
		return code
	}
	list, ok := withServer(*data, stderr, func(ctx context.Context, srv *server.Server) (server.DomainList, error) {
		return operation.Do(ctx, srv.Registry, srv.ListDomains, admin, server.ListDomainsInput{})
	})
	if !ok {
		return exitError
	}
	tw := tabwriter.NewWriter(stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(tw, "ID\tDOMAIN\tCREATED\tREPORTS\tSTATE")
	for _, c := range list.Cahuts {
		state := "active"
		if s := c.Suspension; s != nil {
			state = "suspended since " + day(s.Since) + ": " + s.Reason
		}
		// The reports that wait for the moderators of the Cahut. See
		// docs/adr/0054-reporting-and-withdrawing-a-window.md.
		fmt.Fprintf(tw, "%d\t%s\t%s\t%d\t%s\n", c.ID, c.Domain, day(c.CreatedAt), c.Reports, state)
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
	fs.StringVar(&in.Email, "email", "", "e-mail address of the person, where their link goes when the Cahut sends e-mail")
	if code, ok := parse(fs, args); !ok {
		return code
	}
	u, at, ok := inCahut(*data, *domain, account.CreateUser, in, stderr)
	if !ok {
		return exitError
	}
	fmt.Fprintf(stdout, "Created %s (%s), %s of %s, with a personal Room at %s/%s.\n\n", u.Handle, u.Name, u.Role, at, at, u.PersonalRoom)
	printActivation(stdout, web.ActivationURL(at), *u.Activation)
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
	fmt.Fprintln(tw, "HANDLE\tNAME\tROLE\tSTATUS\tEMAIL")
	for _, u := range list.Users {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n", u.Handle, u.Name, u.Role, u.Status, u.Email)
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
	var in account.ActivationLinkInput
	fs.StringVar(&in.User, "handle", "", "the person who chooses a new password")
	if code, ok := parse(fs, args); !ok {
		return code
	}
	a, at, ok := inCahut(*data, *domain, account.CreateActivationLink, in, stderr)
	if !ok {
		return exitError
	}
	printActivation(stdout, web.ActivationURL(at), a)
	fmt.Fprintln(stdout, "Their password, if they have one, works until they use the link.")
	return exitOK
}

// printActivation prints the link of an activation: base, the page of the
// console where it opens, then its code.
func printActivation(w io.Writer, base string, a account.Activation) {
	if a.Emailed {
		fmt.Fprintf(w, "The link where %s chooses their password went to their e-mail address. It works once, until %s; "+
			"here it is too, shown only now.\n\n", a.User, day(a.ExpiresAt))
	} else {
		fmt.Fprintf(w, "Send %s this link, where they choose their password. It works once, until %s, and is shown only now.\n\n",
			a.User, day(a.ExpiresAt))
	}
	fmt.Fprintf(w, "%s%s\n\n", base, a.Code)
}

func runUserEmail(args []string, stdout, stderr io.Writer) int {
	fs := newFlags("user email", stderr)
	data := dataFlag(fs)
	domain := domainFlag(fs)
	var in account.SetUserEmailInput
	fs.StringVar(&in.User, "handle", "", "the person")
	fs.StringVar(&in.Email, "email", "", "their e-mail address; empty removes it")
	if code, ok := parse(fs, args); !ok {
		return code
	}
	u, at, ok := inCahut(*data, *domain, account.SetUserEmail, in, stderr)
	if !ok {
		return exitError
	}
	if u.Email == "" {
		fmt.Fprintf(stdout, "%s of %s has no e-mail address now.\n", u.Handle, at)
	} else {
		fmt.Fprintf(stdout, "The e-mail address of %s of %s is %s now.\n", u.Handle, at, u.Email)
	}
	return exitOK
}

// printAdminActivation prints the link of an administrator's activation, at
// the console of the server at domain, or at any domain of the server when
// it holds no Cahut yet.
func printAdminActivation(w io.Writer, domain string, a account.Activation) {
	if domain == "" {
		fmt.Fprintln(w, "This server holds no Cahut yet. Once you declare one, put its domain in the link: "+
			"the console of the server opens at the domain of any Cahut of the server.")
		printActivation(w, "https://<domain>"+web.ServerPath+"activate#", a)
		return
	}
	printActivation(w, web.AdminActivationURL(domain), a)
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

func runServiceTokenCreate(args []string, stdout, stderr io.Writer) int {
	fs := newFlags("service-token create", stderr)
	data := dataFlag(fs)
	domain := domainFlag(fs)
	var in account.CreateServiceTokenInput
	fs.StringVar(&in.Name, "name", "", "what uses the token, such as \"account\" for the site where people sign up")
	if code, ok := parse(fs, args); !ok {
		return code
	}
	t, at, ok := inCahut(*data, *domain, account.CreateServiceToken, in, stderr)
	if !ok {
		return exitError
	}
	fmt.Fprintf(stdout, "Service token %d (%s) in %s. It is shown once: copy it now.\n\n%s\n\n", t.ID, t.Name, at, t.Token)
	fmt.Fprintf(stdout, "The program calls the admin API of cahut serve --admin-api, with Host: %s, "+
		"and sends the token in the Authorization header: Bearer <token>.\n", at)
	return exitOK
}

func runServiceTokenList(args []string, stdout, stderr io.Writer) int {
	fs := newFlags("service-token list", stderr)
	data := dataFlag(fs)
	domain := domainFlag(fs)
	if code, ok := parse(fs, args); !ok {
		return code
	}
	list, _, ok := inCahut(*data, *domain, account.ListServiceTokens, account.ListServiceTokensInput{}, stderr)
	if !ok {
		return exitError
	}
	tw := tabwriter.NewWriter(stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(tw, "ID\tNAME\tCREATED\tLAST USED\tSTATE")
	for _, t := range list.Tokens {
		state := "active"
		if t.RevokedAt != nil {
			state = "revoked " + day(*t.RevokedAt)
		}
		lastUsed := "never"
		if t.LastUsedAt != nil {
			lastUsed = minute(*t.LastUsedAt)
		}
		fmt.Fprintf(tw, "%d\t%s\t%s\t%s\t%s\n", t.ID, t.Name, day(t.CreatedAt), lastUsed, state)
	}
	if err := tw.Flush(); err != nil {
		return exitError
	}
	return exitOK
}

func runServiceTokenRevoke(args []string, stdout, stderr io.Writer) int {
	fs := newFlags("service-token revoke", stderr)
	data := dataFlag(fs)
	domain := domainFlag(fs)
	var in account.RevokeServiceTokenInput
	fs.Int64Var(&in.ID, "id", 0, "the token to revoke, as service-token list shows it")
	if code, ok := parse(fs, args); !ok {
		return code
	}
	t, _, ok := inCahut(*data, *domain, account.RevokeServiceToken, in, stderr)
	if !ok {
		return exitError
	}
	fmt.Fprintf(stdout, "Revoked service token %d (%s).\n", t.ID, t.Name)
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
	case e.Actor == string(operation.Person):
		// A call that came without signing in, and names no account, such
		// as a reader who opens an invitation.
		return "visitor"
	}
	return e.Actor
}

func day(t time.Time) string    { return t.Local().Format("2006-01-02") }
func minute(t time.Time) string { return t.Local().Format("2006-01-02 15:04") }
