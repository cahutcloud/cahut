package mail

import (
	"bytes"
	"cmp"
	"crypto/rand"
	_ "embed" // the logo of Cahut
	"encoding/base64"
	"encoding/hex"
	"fmt"
	htmltemplate "html/template"
	"io"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net/mail"
	"net/textproto"
	"strings"
	"text/template"
	"time"
)

// Letter is what a message says. The layout of Cahut sets it in plain text
// and in HTML: its words are plain text, never markup.
type Letter struct {
	To      string
	Subject string
	// Lines are the paragraphs before the action.
	Lines []string
	// Quote holds words that a person or an AI wrote, which the message
	// shows apart from those of Cahut.
	Quote *Quote
	// Action is the link the message is for, shown as a button and as an
	// address.
	Action *Action
	// After are the paragraphs after the action, such as when its link
	// expires.
	After []string
}

// Quote is words that someone wrote, and who.
type Quote struct {
	By   string
	Text string
}

// Action is a link, and what following it does.
type Action struct {
	Label string
	URL   string
}

// The brand of Cahut, as e-mail takes it: literal values, since mail
// clients ignore the variables of CSS. A test keeps them equal to the
// tokens of docs/charte-graphique/tokens/cahut-tokens.json, and the logo to
// docs/charte-graphique/logos/png/cahut-logo-horizontal-800px.png. See
// docs/adr/0012-website-and-docs-as-astro-sites.md.
var brand = map[string]string{
	"paper": "#F7F5F0", // color.brand.paper: the background
	"ink":   "#1F2A33", // color.brand.ink: text, and the button
	"muted": "#5C6872", // color.ink.600: secondary text
	"link":  "#245D83", // color.sky.700: links
	"rule":  "#D9DCDF", // color.brand.light-grey: borders
	"white": "#FFFFFF", // color.neutral.white: the card
}

//go:embed logo.png
var logo []byte

// logoID is the Content-ID of the logo in a message.
const logoID = "logo@cahut"

// The layout of a message, in HTML: tables and inline styles, 600 pixels
// wide at most, as mail clients read it.
var htmlLayout = htmltemplate.Must(htmltemplate.New("html").Funcs(htmltemplate.FuncMap{"lines": paragraphs}).Parse(colors(`<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<meta name="color-scheme" content="light">
<title>{{.Subject}}</title>
</head>
<body style="margin:0;padding:0;background:{paper}">
<table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="background:{paper};border-collapse:collapse">
<tr><td align="center" style="padding:32px 16px">
<table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="max-width:600px;border-collapse:collapse;font-family:'Plus Jakarta Sans',Arial,sans-serif;color:{ink}">
<tr><td style="padding:0 0 24px 0"><img src="cid:` + logoID + `" width="160" height="32" alt="Cahut" style="display:block;border:0;width:160px;height:auto"></td></tr>
<tr><td style="background:{white};border:1px solid {rule};border-radius:16px;padding:32px">
{{- range .Lines}}
<p style="margin:0 0 16px 0;font-size:16px;line-height:24px">{{.}}</p>
{{- end}}
{{- with .Quote}}
<table role="presentation" width="100%" cellpadding="0" cellspacing="0" style="margin:0 0 16px 0;border-collapse:collapse"><tr><td style="border-left:3px solid {rule};padding:4px 0 4px 16px">
<p style="margin:0 0 4px 0;font-size:14px;line-height:20px;color:{muted}">{{.By}} wrote:</p>
<p style="margin:0;font-size:16px;line-height:24px">{{range $i, $l := lines .Text}}{{if $i}}<br>{{end}}{{$l}}{{end}}</p>
</td></tr></table>
{{- end}}
{{- with .Action}}
<table role="presentation" cellpadding="0" cellspacing="0" style="margin:8px 0 24px 0;border-collapse:separate"><tr><td style="background:{ink};border-radius:10px">
<a href="{{.URL}}" style="display:inline-block;padding:12px 20px;font-size:16px;line-height:24px;font-weight:600;color:{paper};text-decoration:none">{{.Label}}</a>
</td></tr></table>
<p style="margin:0 0 16px 0;font-size:14px;line-height:20px;color:{muted}">Or open this address: <a href="{{.URL}}" style="color:{link};font-family:'IBM Plex Mono',Consolas,monospace;word-break:break-all">{{.URL}}</a></p>
{{- end}}
{{- range .After}}
<p style="margin:0 0 16px 0;font-size:14px;line-height:20px;color:{muted}">{{.}}</p>
{{- end}}
</td></tr>
<tr><td style="padding:24px 0 0 0;font-size:12px;line-height:16px;color:{muted}">{{.Footer}}</td></tr>
</table>
</td></tr>
</table>
</body>
</html>
`)))

// The layout of a message, in plain text.
var textLayout = template.Must(template.New("text").Funcs(template.FuncMap{"lines": paragraphs}).Parse(`
{{- range .Lines}}{{.}}

{{end}}
{{- with .Quote}}{{.By}} wrote:

{{range lines .Text}}> {{.}}
{{end}}
{{end}}
{{- with .Action}}{{.Label}}: {{.URL}}

{{end}}
{{- range .After}}{{.}}

{{end}}--
{{.Footer}}
`))

// colors puts the colors of the brand in a layout, where it names them in
// braces, such as {ink}.
func colors(layout string) string {
	pairs := make([]string, 0, 2*len(brand))
	for name, value := range brand {
		pairs = append(pairs, "{"+name+"}", value)
	}
	return strings.NewReplacer(pairs...).Replace(layout)
}

// paragraphs splits text into its lines.
func paragraphs(text string) []string {
	return strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
}

// page is what the layouts read.
type page struct {
	Letter
	Footer string
}

// build makes the message l of from, sent through the relay r at now: its
// headers, its text and HTML, and the logo.
func build(r *Relay, from From, l Letter, now time.Time) ([]byte, error) {
	to, err := Address(l.To)
	if err != nil {
		return nil, err
	}
	p := page{Letter: l}
	p.Subject = oneLine(l.Subject)
	p.Footer = "Sent by the Cahut at " + from.Domain + "."
	var text, html bytes.Buffer
	if err := textLayout.Execute(&text, p); err != nil {
		return nil, err
	}
	if err := htmlLayout.Execute(&html, p); err != nil {
		return nil, err
	}

	id := make([]byte, 16)
	if _, err := rand.Read(id); err != nil {
		return nil, err
	}
	_, host, _ := strings.Cut(r.From, "@")
	name := oneLine(cmp.Or(r.Name, from.Name))
	var msg bytes.Buffer
	related := multipart.NewWriter(&msg)
	inner := multipart.NewWriter(io.Discard).Boundary() // of the text and the HTML, inside
	header := []struct{ key, value string }{
		{"From", (&mail.Address{Name: name, Address: r.From}).String()},
		{"To", to},
		{"Subject", mime.QEncoding.Encode("utf-8", p.Subject)},
		{"Date", now.Format(time.RFC1123Z)},
		{"Message-ID", "<" + hex.EncodeToString(id) + "@" + host + ">"},
		{"MIME-Version", "1.0"},
		{"Auto-Submitted", "auto-generated"},
		{"Content-Type", fmt.Sprintf(`multipart/related; boundary="%s"; type="multipart/alternative"`, related.Boundary())},
	}
	for _, h := range header {
		fmt.Fprintf(&msg, "%s: %s\r\n", h.key, h.value)
	}
	msg.WriteString("\r\n")

	part, err := related.CreatePart(textproto.MIMEHeader{
		"Content-Type": {fmt.Sprintf(`multipart/alternative; boundary="%s"`, inner)},
	})
	if err != nil {
		return nil, err
	}
	alternative := multipart.NewWriter(part)
	if err := alternative.SetBoundary(inner); err != nil {
		return nil, err
	}
	for _, body := range []struct {
		typ  string
		data []byte
	}{{"text/plain", text.Bytes()}, {"text/html", html.Bytes()}} {
		w, err := alternative.CreatePart(textproto.MIMEHeader{
			"Content-Type":              {body.typ + "; charset=utf-8"},
			"Content-Transfer-Encoding": {"quoted-printable"},
		})
		if err != nil {
			return nil, err
		}
		qp := quotedprintable.NewWriter(w)
		if _, err := qp.Write(body.data); err != nil {
			return nil, err
		}
		if err := qp.Close(); err != nil {
			return nil, err
		}
	}
	if err := alternative.Close(); err != nil {
		return nil, err
	}
	w, err := related.CreatePart(textproto.MIMEHeader{
		"Content-Type":              {"image/png"},
		"Content-Transfer-Encoding": {"base64"},
		"Content-ID":                {"<" + logoID + ">"},
		"Content-Disposition":       {`inline; filename="cahut.png"`},
	})
	if err != nil {
		return nil, err
	}
	encoded := base64.StdEncoding.EncodeToString(logo)
	for len(encoded) > 76 {
		fmt.Fprintf(w, "%s\r\n", encoded[:76])
		encoded = encoded[76:]
	}
	fmt.Fprintf(w, "%s\r\n", encoded)
	if err := related.Close(); err != nil {
		return nil, err
	}
	return msg.Bytes(), nil
}

// oneLine keeps s on one line: a header holds no line break.
func oneLine(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
