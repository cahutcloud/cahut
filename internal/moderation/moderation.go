// Package moderation keeps the reports that anyone makes about the Windows
// of a Cahut, and what its moderators decide about them. The moderators are
// the administrators of the Cahut and those of its server: they read the
// reports in the console, and leave each Window as it is or withdraw it.
// Moderation is for people: no AI and no service is offered it. See
// docs/adr/0054-reporting-and-withdrawing-a-window.md.
package moderation

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/store"
)

// Bounds of a report.
const (
	MinExplanation = 10
	MaxExplanation = 2000
	MaxName        = 100
	MaxNote        = 500
)

// Categories are what a report may say is wrong with a Window, in the order
// that the page of a report offers them.
var Categories = []Category{
	{"phishing", "Phishing or a scam"},
	{"malware", "Malware"},
	{"illegal", "Illegal content"},
	{"rights", "Someone's rights: copyright, privacy"},
	{"other", "Something else"},
}

// Category is what a report says is wrong, by its name and in words.
type Category struct {
	Name, Words string
}

// The decisions of the moderators about a report.
const (
	Dismissed = "dismissed" // the moderators left the Window as it is
	Withdrawn = "withdrawn" // they withdrew it
)

// The refusals of moderation that people meet in the console, which says
// them in their language.
var (
	reportDecided = operation.NewRefusalOf[ReportRef]("report_decided", operation.ErrInvalid,
		"report {report} is decided already: list_reports gives those that wait")
	noteTooLong = operation.NewRefusalOf[NoteLimit]("report_note_too_long", operation.ErrInvalid,
		"the note holds more than {limit} characters: keep it shorter")
)

// ReportRef fills the refusals that name a report.
type ReportRef struct {
	Report int64 `json:"report" jsonschema:"the number of the report"`
}

// NoteLimit fills the refusal of a note that is too long.
type NoteLimit struct {
	Limit int `json:"limit" jsonschema:"the most characters of a note"`
}

// Register adds the operations of moderation to r. frameKey signs the
// frames of the Windows that moderators open.
func Register(r *operation.Registry, frameKey []byte) {
	operation.Register(r, &operation.Op[ReportInput, ReportReceived]{
		Name:  "report_window",
		Title: "Report a Window",
		Description: "Report a Window of this Cahut to its moderators, from the page /-/report, without an account: " +
			"its address, what is wrong with it, and an explanation.",
		PeopleOnly: true,
		Visitors:   true,
		Run:        report,
	})
	operation.Register(r, &operation.Op[ListReportsInput, Reports]{
		Name:  "list_reports",
		Title: "List the reports",
		Description: "List the reports about the Windows of this Cahut: those that wait for a decision, the oldest first, " +
			"then the latest decided. For its moderators: its administrators and those of its server.",
		Level:      operation.CahutAdmin,
		PeopleOnly: true,
		ReadOnly:   true,
		Run:        list,
	})
	operation.Register(r, &operation.Op[PreviewReportedInput, ReportedPreview]{
		Name:  "preview_reported_window",
		Title: "Open a reported Window",
		Description: "Get the frame that shows, in the console, the Window of a report that waits, whatever its access, " +
			"for four hours at most. The members of its Room read in the Logbook that a moderator opened it. " +
			"For the moderators of this Cahut.",
		Level: operation.CahutAdmin,
		// It reads, yet goes to the Logbook, as the reads of people do not.
		PeopleOnly: true,
		Run:        previewReported(frameKey),
	})
	operation.Register(r, &operation.Op[DismissInput, Report]{
		Name:  "dismiss_report",
		Title: "Dismiss a report",
		Description: "Decide a report that waits, and leave its Window as it is, with a note for the moderators. " +
			"For the moderators of this Cahut.",
		Level:      operation.CahutAdmin,
		PeopleOnly: true,
		Run:        dismiss,
	})
}

// Report is a report about a Window, as its moderators read it.
type Report struct {
	ID          int64      `json:"id"`
	Window      Reported   `json:"window"`
	Category    string     `json:"category" jsonschema:"what is wrong: phishing, malware, illegal, rights or other"`
	Explanation string     `json:"explanation"`
	Name        string     `json:"name,omitempty" jsonschema:"the name of whoever reported, until the decision"`
	Email       string     `json:"email,omitempty" jsonschema:"the address of whoever reported, until the decision"`
	CreatedAt   time.Time  `json:"created_at"`
	Decision    string     `json:"decision,omitempty" jsonschema:"dismissed or withdrawn; absent while it waits"`
	DecidedAt   *time.Time `json:"decided_at,omitempty"`
	DecidedBy   string     `json:"decided_by,omitempty" jsonschema:"the handle of the moderator who decided"`
	ByServer    bool       `json:"by_server,omitempty" jsonschema:"an administrator of the server decided"`
	Note        string     `json:"note,omitempty" jsonschema:"what the moderators noted when they left the Window as it is"`
}

// Reported is the Window that a report names.
type Reported struct {
	ID      string `json:"id" jsonschema:"the Window's id, which never changes"`
	Room    string `json:"room"`
	Name    string `json:"name"`
	Title   string `json:"title"`
	Access  string `json:"access" jsonschema:"author, room, internal, key or public"`
	Trashed bool   `json:"trashed,omitempty" jsonschema:"the Window is in the trash"`
}

// ListReportsInput is the input of list_reports.
type ListReportsInput struct {
	Waiting bool `json:"waiting,omitempty" jsonschema:"only the reports that wait for a decision"`
	Limit   int  `json:"limit,omitempty" jsonschema:"how many reports, 100 by default, at most 500"`
}

// Reports is the output of list_reports.
type Reports struct {
	Waiting int      `json:"waiting" jsonschema:"how many reports wait for a decision, in all"`
	Reports []Report `json:"reports" jsonschema:"the reports that wait, the oldest first, then the latest decided"`
}

const selectReport = `
	SELECT p.id, w.uid, r.name, w.name, w.title, w.access, w.trashed_at IS NOT NULL, p.category, p.explanation,
		coalesce(p.name, ''), coalesce(p.email, ''), p.created_at, coalesce(p.decision, ''), p.decided_at,
		coalesce(p.decided_by, ''), p.by_server, p.note
	FROM reports p JOIN windows w ON w.id = p.window_id JOIN rooms r ON r.id = w.room_id`

func list(ctx context.Context, c *operation.Call, in ListReportsInput) (Reports, error) {
	c.Log("reports", "")
	limit := in.Limit
	switch {
	case limit == 0:
		limit = 100
	case limit < 0 || limit > 500:
		return Reports{}, operation.Invalidf("limit %d: use 1 to 500", limit)
	}
	out := Reports{Reports: []Report{}}
	if err := c.Tx.QueryRowContext(ctx, "SELECT count(*) FROM reports WHERE decision IS NULL").Scan(&out.Waiting); err != nil {
		return Reports{}, err
	}
	waiting, err := query(ctx, c.Tx, selectReport+" WHERE p.decision IS NULL ORDER BY p.id LIMIT ?", limit)
	if err != nil {
		return Reports{}, err
	}
	out.Reports = append(out.Reports, waiting...)
	if in.Waiting || len(waiting) == limit {
		return out, nil
	}
	decided, err := query(ctx, c.Tx, selectReport+" WHERE p.decision IS NOT NULL ORDER BY p.decided_at DESC, p.id DESC LIMIT ?", limit-len(waiting))
	if err != nil {
		return Reports{}, err
	}
	out.Reports = append(out.Reports, decided...)
	return out, nil
}

func query(ctx context.Context, tx *sql.Tx, q string, args ...any) ([]Report, error) {
	rows, err := tx.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []Report
	for rows.Next() {
		var p Report
		var created string
		var decided sql.NullString
		err := rows.Scan(&p.ID, &p.Window.ID, &p.Window.Room, &p.Window.Name, &p.Window.Title, &p.Window.Access, &p.Window.Trashed,
			&p.Category, &p.Explanation, &p.Name, &p.Email, &created, &p.Decision, &decided, &p.DecidedBy, &p.ByServer, &p.Note)
		if err != nil {
			return nil, err
		}
		if p.CreatedAt, err = store.ParseTime(created); err != nil {
			return nil, err
		}
		if decided.Valid {
			t, err := store.ParseTime(decided.String)
			if err != nil {
				return nil, err
			}
			p.DecidedAt = &t
		}
		list = append(list, p)
	}
	return list, rows.Err()
}

// get returns the report id.
func get(ctx context.Context, tx *sql.Tx, id int64) (Report, error) {
	list, err := query(ctx, tx, selectReport+" WHERE p.id = ?", id)
	if err != nil {
		return Report{}, err
	}
	if len(list) == 0 {
		return Report{}, operation.NotFoundf("no report %d: list_reports gives the reports of this Cahut", id)
	}
	return list[0], nil
}

// DismissInput is the input of dismiss_report.
type DismissInput struct {
	Report int64  `json:"report" jsonschema:"the number of the report, which list_reports gives"`
	Note   string `json:"note,omitempty" jsonschema:"why the Window stays as it is, for the moderators, 500 characters at most"`
}

func dismiss(ctx context.Context, c *operation.Call, in DismissInput) (Report, error) {
	c.Log("reports", fmt.Sprintf("report %d", in.Report))
	note := strings.TrimSpace(in.Note)
	if utf8.RuneCountInString(note) > MaxNote {
		return Report{}, noteTooLong(NoteLimit{Limit: MaxNote})
	}
	p, err := get(ctx, c.Tx, in.Report)
	if err != nil {
		return Report{}, err
	}
	if p.Decision != "" {
		return Report{}, reportDecided(ReportRef{Report: in.Report})
	}
	c.Log(p.Window.Room+"/"+p.Window.Name, fmt.Sprintf("report %d", in.Report))
	if err := decide(ctx, c, in.Report, Dismissed, note); err != nil {
		return Report{}, err
	}
	return get(ctx, c.Tx, in.Report)
}

// decide records the decision of the moderator of c about the report id,
// and forgets the name and the address of whoever made it.
func decide(ctx context.Context, c *operation.Call, id int64, decision, note string) error {
	_, err := c.Tx.ExecContext(ctx, `
		UPDATE reports SET decision = ?, decided_at = ?, decided_by = ?, by_server = ?, note = ?, name = NULL, email = NULL
		WHERE id = ? AND decision IS NULL`,
		decision, store.FormatTime(c.Now), c.Actor.Handle, c.Actor.Kind == operation.Admin, note, id)
	return err
}

// cleanLine returns s trimmed, or false when it holds a line break or
// another control character.
func cleanLine(s string) (string, bool) {
	s = strings.TrimSpace(s)
	for _, r := range s {
		if unicode.IsControl(r) || r == utf8.RuneError {
			return "", false
		}
	}
	return s, true
}

// errNoWindow refuses an address that names no Window of the Cahut.
var errNoWindow = errors.New("this address names no Window of this Cahut")
