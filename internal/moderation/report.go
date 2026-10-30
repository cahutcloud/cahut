package moderation

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"slices"
	"strings"
	"unicode/utf8"

	"cahut.cloud/internal/address"
	"cahut.cloud/internal/credential"
	"cahut.cloud/internal/mail"
	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/store"
	"cahut.cloud/internal/window"
)

// ReportInput is the input of report_window.
type ReportInput struct {
	Window      string `json:"window" jsonschema:"the address of the Window, as its reader meets it: its address, a file under it, its permanent address or a share link, such as https://cahut.example.com/room/window/"`
	Category    string `json:"category" jsonschema:"what is wrong: phishing, malware, illegal, rights or other"`
	Explanation string `json:"explanation" jsonschema:"what is wrong with the Window, from 10 to 2000 characters"`
	Name        string `json:"name,omitempty" jsonschema:"the name of whoever reports, for the moderators"`
	Email       string `json:"email,omitempty" jsonschema:"where the receipt and the decision go"`
	GoodFaith   bool   `json:"good_faith" jsonschema:"whoever reports says that they do it in good faith, and that what they say is true"`
}

// ReportReceived is the output of report_window.
type ReportReceived struct {
	Report int64 `json:"report" jsonschema:"the number of the report"`
}

// report keeps a report about a Window. Whoever reports is not named in
// the Logbook, which keeps neither their name, nor their address, nor the
// explanation, nor their address on the network; only the moderators read
// the entry.
func report(ctx context.Context, c *operation.Call, in ReportInput) (ReportReceived, error) {
	c.Anonymize()
	// The address may be a share link: never log it as it was given.
	c.Log("", "")
	if !slices.ContainsFunc(Categories, func(k Category) bool { return k.Name == in.Category }) {
		return ReportReceived{}, operation.Invalidf("say what is wrong with the Window: phishing, malware, illegal, rights or other")
	}
	explanation := strings.TrimSpace(in.Explanation)
	if n := utf8.RuneCountInString(explanation); n < MinExplanation || n > MaxExplanation {
		return ReportReceived{}, operation.Invalidf("explain what is wrong with the Window in %d to %d characters", MinExplanation, MaxExplanation)
	}
	name, ok := cleanLine(in.Name)
	if !ok || utf8.RuneCountInString(name) > MaxName {
		return ReportReceived{}, operation.Invalidf("write your name on one line, in %d characters at most", MaxName)
	}
	email := strings.TrimSpace(in.Email)
	if email != "" {
		var err error
		// The error of mail.Address quotes the address, which the Logbook
		// must not keep.
		if email, err = mail.Address(email); err != nil {
			return ReportReceived{}, operation.Invalidf("give one e-mail address alone, such as jane@example.com, or none")
		}
	}
	if !in.GoodFaith {
		return ReportReceived{}, operation.Invalidf("say that you report in good faith, and that what you say is true")
	}
	w, err := find(ctx, c.Tx, c.Actor.Domain, in.Window)
	if errors.Is(err, errNoWindow) {
		return ReportReceived{}, operation.NotFoundf("%s. Give the address of the Window as your browser shows it, or the link you received", err.Error())
	}
	if err != nil {
		return ReportReceived{}, err
	}
	var id int64
	err = c.Tx.QueryRowContext(ctx, `
		INSERT INTO reports (window_id, category, explanation, name, email, created_at) VALUES (?, ?, ?, ?, ?, ?) RETURNING id`,
		w.id, in.Category, explanation, nullString(name), nullString(email), store.FormatTime(c.Now)).Scan(&id)
	if err != nil {
		return ReportReceived{}, err
	}
	c.Log(w.room+"/"+w.name, fmt.Sprintf("report %d, %s", id, in.Category))
	if err := tellOfReport(ctx, c, id, in.Category, w.room, w.name, email); err != nil {
		return ReportReceived{}, err
	}
	return ReportReceived{Report: id}, nil
}

// found is the Window that an address names.
type found struct {
	id         int64
	room, name string
}

// find returns the Window out of the trash that the address s names, in
// the Cahut at domain: its address, a file under it, its permanent address,
// a share link, or the address of a Room that shows a Window at its own.
// The address may come whole, with the domain of the Cahut, or as a path.
// It returns errNoWindow when s names no such Window.
func find(ctx context.Context, tx *sql.Tx, domain, s string) (found, error) {
	s = strings.TrimSpace(s)
	p := s
	if strings.Contains(s, "://") {
		u, err := url.Parse(s)
		if err != nil || u.Scheme != "https" && u.Scheme != "http" || domain != "" && !strings.EqualFold(u.Hostname(), hostOf(domain)) {
			return found{}, errNoWindow
		}
		p = u.Path
	} else if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	var where string
	var args []any
	parts := strings.Split(strings.TrimPrefix(p, "/"), "/")
	switch {
	case strings.HasPrefix(p, window.LinkPrefix):
		token := strings.TrimSuffix(strings.TrimPrefix(p, window.LinkPrefix), "/")
		if token == "" || strings.Contains(token, "/") {
			return found{}, errNoWindow
		}
		where, args = "w.id = (SELECT window_id FROM share_links WHERE hash = ?)", []any{credential.Hash(token)}
	case strings.HasPrefix(p, window.PermanentPrefix):
		id := strings.ToLower(parts[2])
		if !address.ValidID(id) {
			return found{}, errNoWindow
		}
		where, args = "w.uid = ?", []any{id}
	case len(parts) >= 2 && parts[1] != "" && address.ValidName(parts[0]) && address.ValidName(parts[1]):
		where, args = "r.name = ? AND w.name = ?", []any{parts[0], parts[1]}
	case len(parts) <= 2 && address.ValidName(parts[0]) && (len(parts) == 1 || parts[1] == ""):
		// The address of a Room that shows a Window at its own.
		where, args = "w.id = (SELECT home_window_id FROM rooms WHERE name = ?)", []any{parts[0]}
	default:
		return found{}, errNoWindow
	}
	var f found
	err := tx.QueryRowContext(ctx, `
		SELECT w.id, r.name, w.name FROM windows w JOIN rooms r ON r.id = w.room_id
		WHERE w.trashed_at IS NULL AND `+where, args...).Scan(&f.id, &f.room, &f.name)
	if errors.Is(err, sql.ErrNoRows) {
		return found{}, errNoWindow
	}
	return f, err
}

// hostOf returns the host of a domain that may carry a port, as a Cahut in
// development does.
func hostOf(domain string) string {
	if u, err := url.Parse("https://" + domain); err == nil && u.Hostname() != "" {
		return u.Hostname()
	}
	return domain
}

func nullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}
