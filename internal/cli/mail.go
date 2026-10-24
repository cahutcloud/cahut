package cli

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"

	"cahut.cloud/internal/cahut"
	"cahut.cloud/internal/mail"
	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/server"
)

// The relay of the server, or of a Cahut with --domain. See
// docs/adr/0031-e-mail-through-a-relay.md.
var mailCommands = []command{
	{name: "show", summary: "Print the e-mail relay of the server, or of a Cahut with --domain", run: runMailShow},
	{name: "set", summary: "Set the e-mail relay of the server, or of a Cahut with --domain, once it answers", run: runMailSet},
	{name: "remove", summary: "Remove the e-mail relay of the server, or of a Cahut with --domain", run: runMailRemove},
	{name: "test", summary: "Send a test message through the relay of the server, or of a Cahut with --domain", run: runMailTest},
}

// stdin is where commands read what they take on their standard input,
// such as the password of a relay. Tests replace it.
var stdin io.Reader = os.Stdin

// mailDomainFlag adds the --domain flag of the mail commands, which act on
// the relay of the server without it.
func mailDomainFlag(fs *flag.FlagSet) *string {
	return fs.String("domain", "", "the Cahut whose relay to act on, by its domain; without it, the relay of the server")
}

// onRelay runs the operation called name, with the input in, on the relay
// of the server, or of the Cahut at domain. It returns its output and what
// the relay belongs to: "the server", or the domain of the Cahut.
func onRelay[Out any](data, domain, name string, in any, stderr io.Writer) (Out, string, bool) {
	var zero Out
	args, err := json.Marshal(in)
	if err != nil {
		fmt.Fprintf(stderr, "cahut: %v\n", err)
		return zero, "", false
	}
	invoke := func(ctx context.Context, r *operation.Registry) (Out, error) {
		out, err := r.Invoke(ctx, name, admin, args)
		if err != nil {
			return zero, err
		}
		return out.(Out), nil
	}
	if domain == "" {
		out, ok := withServer(data, stderr, func(ctx context.Context, srv *server.Server) (Out, error) {
			return invoke(ctx, srv.Registry)
		})
		return out, "the server", ok
	}
	return withCahut(data, domain, stderr, func(ctx context.Context, c *cahut.Cahut) (Out, error) {
		return invoke(ctx, c.Registry)
	})
}

func runMailShow(args []string, stdout, stderr io.Writer) int {
	fs := newFlags("mail show", stderr)
	data := dataFlag(fs)
	domain := mailDomainFlag(fs)
	if code, ok := parse(fs, args); !ok {
		return code
	}
	out, of, ok := onRelay[mail.MailRelay](*data, *domain, "get_mail_relay", mail.GetMailRelayInput{}, stderr)
	if !ok {
		return exitError
	}
	printRelay(stdout, of, out)
	return exitOK
}

func runMailSet(args []string, stdout, stderr io.Writer) int {
	fs := newFlags("mail set", stderr)
	data := dataFlag(fs)
	domain := mailDomainFlag(fs)
	var in mail.SetMailRelayInput
	fs.StringVar(&in.Host, "host", "", "the name or the address of the SMTP relay, such as smtp.example.com")
	fs.IntVar(&in.Port, "port", 0, "its port: 465 for TLS from the start, 587 for STARTTLS; 465 with tls and 587 otherwise when left out")
	fs.StringVar(&in.Security, "security", "", "tls, starttls, or none for a relay of the server on this machine; tls on port 465 and starttls otherwise when left out")
	fs.StringVar(&in.Username, "user", "", "the user name at the relay; leave it out when the relay takes no authentication")
	password := fs.Bool("password-stdin", false, "read the password at the relay on the standard input; without it, "+
		"the password kept stays when the host and the user name do not change")
	fs.StringVar(&in.From, "from", "", "the address that messages come from, one that the relay accepts")
	fs.StringVar(&in.Name, "name", "", "the name that messages come from; the name of the Cahut when left out")
	fs.Usage = func() {
		fmt.Fprintln(stderr, "Usage: cahut mail set [--data DIR] [--domain DOMAIN] --host HOST [--port PORT] [--security tls|starttls|none] "+
			"[--user USER --password-stdin] --from ADDRESS [--name NAME]")
		fs.PrintDefaults()
	}
	if code, ok := parse(fs, args); !ok {
		return code
	}
	if strings.TrimSpace(in.Host) == "" {
		fmt.Fprintln(stderr, "cahut mail set: give --host; cahut mail remove removes a relay")
		return exitUsage
	}
	if *password {
		b, err := io.ReadAll(io.LimitReader(stdin, 4096))
		if err != nil {
			fmt.Fprintf(stderr, "cahut: read the password: %v\n", err)
			return exitError
		}
		in.Password = strings.TrimRight(string(b), "\r\n")
		if in.Password == "" {
			fmt.Fprintln(stderr, "cahut mail set: --password-stdin read no password")
			return exitUsage
		}
	}
	out, of, ok := onRelay[mail.MailRelay](*data, *domain, "set_mail_relay", in, stderr)
	if !ok {
		return exitError
	}
	fmt.Fprintf(stdout, "The relay answered, and sends the e-mail of %s.\n\n", of)
	printRelay(stdout, of, out)
	return exitOK
}

func runMailRemove(args []string, stdout, stderr io.Writer) int {
	fs := newFlags("mail remove", stderr)
	data := dataFlag(fs)
	domain := mailDomainFlag(fs)
	if code, ok := parse(fs, args); !ok {
		return code
	}
	out, of, ok := onRelay[mail.MailRelay](*data, *domain, "set_mail_relay", mail.SetMailRelayInput{}, stderr)
	if !ok {
		return exitError
	}
	fmt.Fprintf(stdout, "Removed the relay of %s.\n\n", of)
	printRelay(stdout, of, out)
	return exitOK
}

func runMailTest(args []string, stdout, stderr io.Writer) int {
	fs := newFlags("mail test", stderr)
	data := dataFlag(fs)
	domain := mailDomainFlag(fs)
	var in mail.SendTestMailInput
	fs.StringVar(&in.To, "to", "", "the address to send the test to")
	if code, ok := parse(fs, args); !ok {
		return code
	}
	if in.To == "" {
		fmt.Fprintln(stderr, "cahut mail test: give --to, the address to send the test to")
		return exitUsage
	}
	out, of, ok := onRelay[mail.TestSent](*data, *domain, "send_test_mail", in, stderr)
	if !ok {
		return exitError
	}
	through := "its relay"
	if out.Server {
		through = "the relay of the server"
	}
	fmt.Fprintf(stdout, "Sent a test from %s to %s, through %s, which took it.\n", of, out.To, through)
	return exitOK
}

// printRelay prints the relay of the server, or of a Cahut, called of.
func printRelay(w io.Writer, of string, m mail.MailRelay) {
	switch {
	case m.Relay != nil:
		r := m.Relay
		user := "none: the relay takes no authentication"
		if r.Username != "" {
			user = r.Username + ", with a password"
		}
		from := r.From
		if r.Name != "" {
			from = r.Name + " <" + r.From + ">"
		}
		tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
		fmt.Fprintf(tw, "host\t%s:%d, %s\n", r.Host, r.Port, r.Security)
		fmt.Fprintf(tw, "user\t%s\n", user)
		fmt.Fprintf(tw, "from\t%s\n", from)
		fmt.Fprintf(tw, "set\t%s\n", minute(r.SetAt))
		_ = tw.Flush()
	case m.Server:
		fmt.Fprintf(w, "%s has no relay of its own: its e-mail goes through the relay of the server.\n", of)
	case of == "the server":
		fmt.Fprintln(w, "The server has no relay: the Cahuts without one of their own send no e-mail.")
	default:
		fmt.Fprintf(w, "%s has no relay, and the server has none: it sends no e-mail.\n", of)
	}
}
