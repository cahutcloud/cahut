package server

import (
	"context"
	"errors"
	"time"

	"cahut.cloud/internal/mail"
	"cahut.cloud/internal/release"
	"cahut.cloud/internal/update"
	"cahut.cloud/internal/version"
)

// The notices of the server go to its administrators who have an address,
// through the relay of the server. See docs/adr/0031-e-mail-through-a-relay.md.

// NotifyAdmins sends to every administrator of the server who has an
// address the letter that write makes, at most once every every for the
// notice of kind about key, or once for good when every is zero. Nothing is
// recorded while no administrator has an address or the server sends no
// e-mail, so that the notice goes once one can receive it.
func (s *Server) NotifyAdmins(ctx context.Context, kind, key string, every time.Duration, write func(a Admin) mail.Letter) error {
	o := mail.OutboxIn(s.Registry)
	if o == nil {
		return nil
	}
	rows, err := s.DB.Read.QueryContext(ctx, "SELECT handle, name, email FROM admins WHERE email IS NOT NULL AND disabled_at IS NULL ORDER BY id")
	if err != nil {
		return err
	}
	var admins []Admin
	for rows.Next() {
		var a Admin
		if err := rows.Scan(&a.Handle, &a.Name, &a.Email); err != nil {
			rows.Close()
			return err
		}
		admins = append(admins, a)
	}
	if err := errors.Join(rows.Err(), rows.Close()); err != nil || len(admins) == 0 {
		return err
	}
	tx, err := s.DB.Write.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck // a no-op after Commit
	if ok, err := o.Sends(ctx, tx); err != nil || !ok {
		return err
	}
	if ok, err := mail.Once(ctx, tx, kind, key, every, time.Now().UTC()); err != nil || !ok {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	for _, a := range admins {
		l := write(a)
		l.To = a.Email
		o.Send(l)
	}
	return nil
}

// noticeFailedUpdate tells the administrators that the update run failed,
// once.
func (s *Server) noticeFailedUpdate(ctx context.Context, run update.Run) error {
	host := s.Mail.Host()
	return s.NotifyAdmins(ctx, "update_failed", run.ID, 0, func(a Admin) mail.Letter {
		return mail.Letter{
			Subject: "An update failed on " + host,
			Lines: []string{"Hello " + a.Name + ",", "The update of the Cahut server " + host + " from " + run.From + " to " + run.To +
				" failed: " + run.Reason + ". The server went back to " + run.From + "."},
			After: []string{"cahut doctor, on the server, says more."},
		}
	})
}

// noticeRelease tells the administrators, once for each release, that a
// release newer than the one running waits for them: their policy of
// updates does not install it on its own.
func (s *Server) noticeRelease(ctx context.Context) error {
	return s.noticeReleaseOver(ctx, version.Get().Version)
}

// noticeReleaseOver is noticeRelease, with the server running the version
// current.
func (s *Server) noticeReleaseOver(ctx context.Context, current string) error {
	if s.cfg.UpdateDir == "" {
		return nil
	}
	st, managed, err := update.ReadStatus(s.cfg.UpdateDir)
	if err != nil || !managed || st.Check == nil || st.Check.Latest == "" || st.Policy == update.All || st.Policy == update.Off {
		return err
	}
	latest, err := release.ParseVersion(st.Check.Latest)
	if err != nil {
		return nil //nolint:nilerr // a release that is not a version waits for no one
	}
	running, err := release.ParseVersion(current)
	if err != nil || latest.Compare(running) <= 0 || st.Policy == update.Patch && latest.SameMinor(running) {
		return nil //nolint:nilerr // a build that is not a release, or a release that installs on its own
	}
	host := s.Mail.Host()
	return s.NotifyAdmins(ctx, "release", latest.String(), 0, func(a Admin) mail.Letter {
		l := mail.Letter{
			Subject: "Cahut " + latest.String() + " is ready to install on " + host,
			Lines: []string{"Hello " + a.Name + ",", "Cahut " + latest.String() + " is out, and the server " + host + " runs " + running.String() +
				". Its policy of updates leaves this one to you."},
			After: []string{"Install it from the console, under Updates, or with sudo /usr/local/bin/cahut update on the server."},
		}
		if st.Check.URL != "" {
			l.Action = &mail.Action{Label: "Read what changes", URL: st.Check.URL}
		}
		return l
	})
}
