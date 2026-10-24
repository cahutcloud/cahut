package server

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"cahut.cloud/internal/logbook"
	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/release"
	"cahut.cloud/internal/update"
	"cahut.cloud/internal/version"
)

// The updates of the server: root installs them, the service reads what
// root left and asks for a check or an update. See
// docs/adr/0024-signed-releases-and-updates.md.

// UpdatesInput is the input of get_updates.
type UpdatesInput struct{}

// Updates is the output of get_updates.
type Updates struct {
	Running   string        `json:"running" jsonschema:"the version of Cahut running"`
	Managed   bool          `json:"managed" jsonschema:"whether the server updates itself, as install.sh sets it up"`
	Policy    update.Policy `json:"policy,omitempty" jsonschema:"what the daily run does: patch (installs the patch releases of the minor version running), all, check or off"`
	Check     *update.Check `json:"check,omitempty"`
	Last      *update.Run   `json:"last,omitempty"`
	Requested string        `json:"requested,omitempty" jsonschema:"what an administrator asked for, check or a version, which the updater has not taken yet"`
}

// RequestUpdateInput is the input of request_update.
type RequestUpdateInput struct {
	Version string `json:"version,omitempty" jsonschema:"a published release newer than the version running, such as 0.4.1; empty to check for new releases only"`
}

// RequestedUpdate is the output of request_update.
type RequestedUpdate struct {
	Requested string `json:"requested" jsonschema:"check, or the version to install"`
}

func (s *Server) registerUpdates(r *operation.Registry) {
	operation.Register(r, &operation.Op[UpdatesInput, Updates]{
		Name:        "get_updates",
		Title:       "Describe updates",
		Description: "Say which version of Cahut runs, the newest release, and how the last check and the last update went.",
		Level:       operation.ServerAdmin,
		ReadOnly:    true,
		Run: func(context.Context, *operation.Call, UpdatesInput) (Updates, error) {
			return s.updates()
		},
	})
	operation.Register(r, &operation.Op[RequestUpdateInput, RequestedUpdate]{
		Name:  "request_update",
		Title: "Ask for an update",
		Description: "Ask the updater of the server to check for new releases, or to install a release newer than the one running. " +
			"It runs as root within a minute, checks the signature, and goes back to the version running if the new one fails.",
		Level: operation.ServerAdmin,
		Run:   s.requestUpdate,
	})
}

func (s *Server) updates() (Updates, error) {
	u := Updates{Running: version.Get().Version}
	if s.cfg.UpdateDir == "" {
		return u, nil
	}
	st, managed, err := update.ReadStatus(s.cfg.UpdateDir)
	if err != nil || !managed {
		return u, err
	}
	u.Managed, u.Policy, u.Check, u.Last = true, st.Policy, st.Check, st.Last
	what, asked, err := update.ReadRequest(s.cfg.DataDir)
	if asked && err == nil {
		u.Requested = what
		if what == "" {
			u.Requested = "check"
		}
	}
	return u, nil
}

func (s *Server) requestUpdate(_ context.Context, c *operation.Call, in RequestUpdateInput) (RequestedUpdate, error) {
	out := RequestedUpdate{Requested: "check"}
	if in.Version != "" {
		v, err := release.ParseVersion(in.Version)
		if err != nil {
			return RequestedUpdate{}, operation.Invalidf("version %q: give a release such as 0.4.1, or nothing to check for releases", in.Version)
		}
		if running, err := release.ParseVersion(version.Get().Version); err == nil && v.Compare(running) <= 0 {
			return RequestedUpdate{}, operation.Invalidf("%s is not newer than %s, the version running: an update never goes back", v, running)
		}
		out.Requested = v.String()
	}
	c.Log(out.Requested, "")
	if _, managed, _ := update.ReadStatus(s.cfg.UpdateDir); s.cfg.UpdateDir == "" || !managed {
		return RequestedUpdate{}, operation.Invalidf("this server does not update itself: install.sh sets up the updater; " +
			"run install.sh again, or update with sudo /usr/local/bin/cahut update")
	}
	want := ""
	if out.Requested != "check" {
		want = out.Requested
	}
	if err := update.WriteRequest(s.cfg.DataDir, want); err != nil {
		return RequestedUpdate{}, err
	}
	return out, nil
}

// WatchUpdates writes to the Logbook of the server, every minute until ctx
// is done, the updates that the updater made, each once. The updater runs
// as root and never opens the databases.
func (s *Server) WatchUpdates(ctx context.Context, log *slog.Logger) {
	if s.cfg.UpdateDir == "" {
		return
	}
	tick := time.NewTicker(time.Minute)
	defer tick.Stop()
	for {
		if err := s.LogUpdates(ctx); err != nil {
			log.Warn("record the last update in the Logbook", "error", err)
		}
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
		}
	}
}

// LogUpdates writes the updates that the updater made to the Logbook of the
// server, those that are not there already.
func (s *Server) LogUpdates(ctx context.Context) error {
	st, managed, err := update.ReadStatus(s.cfg.UpdateDir)
	if err != nil || !managed {
		return err
	}
	for _, run := range st.Runs() {
		if err := s.logUpdate(ctx, run); err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) logUpdate(ctx context.Context, run update.Run) error {
	tx, err := s.DB.Write.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck // a no-op after Commit
	var seen int
	err = tx.QueryRowContext(ctx, "SELECT 1 FROM updates_logged WHERE id = ?", run.ID).Scan(&seen)
	if err == nil {
		return nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return err
	}
	e := logbook.Entry{
		At:        run.Ended,
		Operation: "update",
		Actor:     string(operation.Admin),
		Via:       map[string]string{update.ByConsole: "api", update.ByTimer: "timer"}[run.By],
		Target:    fmt.Sprintf("%s to %s", run.From, run.To),
		Outcome:   logbook.OK,
	}
	if e.Via == "" {
		e.Via = "cli"
	}
	switch run.Result {
	case update.Updated:
		e.Detail = "updated"
	case update.RolledBack:
		e.Detail, e.Outcome, e.Error = "rolled back", logbook.Failed, run.Reason
	default:
		e.Detail, e.Outcome, e.Error = run.Result, logbook.Failed, run.Reason
	}
	if err := logbook.Write(ctx, tx, e); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, "INSERT INTO updates_logged (id) VALUES (?)", run.ID); err != nil {
		return err
	}
	return tx.Commit()
}
