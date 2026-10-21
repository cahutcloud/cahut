package cli

import (
	"context"
	"fmt"
	"io"
	"text/tabwriter"
	"time"

	"cahut.cloud/internal/account"
	"cahut.cloud/internal/home"
	"cahut.cloud/internal/logbook"
	"cahut.cloud/internal/operation"
)

// admin is the actor of every command run on the server: whoever can run
// them can already read the home's data directory.
var admin = operation.Actor{Kind: operation.Admin, Via: "cli"}

var userCommands = []command{
	{name: "create", summary: "Create a person, with their personal Room", run: runUserCreate},
}

var tokenCommands = []command{
	{name: "create", summary: "Create a personal token for an AI", run: runTokenCreate},
	{name: "list", summary: "List personal tokens", run: runTokenList},
	{name: "revoke", summary: "Revoke a personal token", run: runTokenRevoke},
}

// do opens the home in dir, runs op as the administrator and closes the
// home. It reports errors on stderr.
func do[In, Out any](dir string, op *operation.Op[In, Out], in In, stderr io.Writer) (Out, bool) {
	ctx := context.Background()
	var zero Out
	h, err := home.Open(ctx, home.Config{DataDir: dir})
	if err != nil {
		fmt.Fprintf(stderr, "cahut: %v\n", err)
		return zero, false
	}
	defer h.Close()
	out, err := operation.Do(ctx, h.Registry, op, admin, in)
	if err != nil {
		fmt.Fprintf(stderr, "cahut: %s: %v\n", op.Name, err)
		return zero, false
	}
	return out, true
}

func runUserCreate(args []string, stdout, stderr io.Writer) int {
	fs := newFlags("user create", stderr)
	data := dataFlag(fs)
	var in account.CreateUserInput
	fs.StringVar(&in.Handle, "handle", "", "short name: 2 to 32 lowercase letters, digits and hyphens")
	fs.StringVar(&in.Name, "name", "", "full name")
	fs.BoolVar(&in.Admin, "admin", false, "make the person an administrator of the home")
	if code, ok := parse(fs, args); !ok {
		return code
	}
	u, ok := do(*data, account.CreateUser, in, stderr)
	if !ok {
		return exitError
	}
	fmt.Fprintf(stdout, "Created %s (%s), %s, with a personal Room.\n", u.Handle, u.Name, u.Role)
	return exitOK
}

func runTokenCreate(args []string, stdout, stderr io.Writer) int {
	fs := newFlags("token create", stderr)
	data := dataFlag(fs)
	var in account.CreateTokenInput
	fs.StringVar(&in.User, "user", "", "handle of the person the AI acts for")
	fs.StringVar(&in.Name, "name", "", "what the token is for, such as \"Claude Code on my laptop\"")
	if code, ok := parse(fs, args); !ok {
		return code
	}
	t, ok := do(*data, account.CreateToken, in, stderr)
	if !ok {
		return exitError
	}
	fmt.Fprintf(stdout, "Token %d for %s (%s). It is shown once: copy it now.\n\n%s\n\n", t.ID, t.User, t.Name, t.Token)
	fmt.Fprintln(stdout, "The AI sends it in the Authorization header: Bearer <token>.")
	return exitOK
}

func runTokenList(args []string, stdout, stderr io.Writer) int {
	fs := newFlags("token list", stderr)
	data := dataFlag(fs)
	var in account.ListTokensInput
	fs.StringVar(&in.User, "user", "", "only the tokens of this person")
	if code, ok := parse(fs, args); !ok {
		return code
	}
	list, ok := do(*data, account.ListTokens, in, stderr)
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
	var in account.RevokeTokenInput
	fs.Int64Var(&in.ID, "id", 0, "the token to revoke, as token list shows it")
	if code, ok := parse(fs, args); !ok {
		return code
	}
	t, ok := do(*data, account.RevokeToken, in, stderr)
	if !ok {
		return exitError
	}
	fmt.Fprintf(stdout, "Revoked token %d of %s (%s).\n", t.ID, t.User, t.Name)
	return exitOK
}

func runLogbook(args []string, stdout, stderr io.Writer) int {
	fs := newFlags("logbook", stderr)
	data := dataFlag(fs)
	var in home.ListLogbookInput
	fs.IntVar(&in.Limit, "limit", 20, "how many entries")
	if code, ok := parse(fs, args); !ok {
		return code
	}
	page, ok := do(*data, home.ListLogbook, in, stderr)
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
