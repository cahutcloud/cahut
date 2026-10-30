package moderation

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"cahut.cloud/internal/mail"
	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/settings"
	"cahut.cloud/internal/window"
)

// The e-mail of moderation, when the Cahut sends e-mail. Each report goes
// to the administrators of the Cahut who have an address, and to its
// address of moderation; whoever reports with an address gets a receipt,
// then the decision; the members of the Room of a Window hear that it was
// withdrawn, with the reason, and that it was reinstated. These e-mails
// name the Window by its address, and hold nothing that whoever reported
// wrote, so that the page of a report cannot send a text of one's choice
// to someone else. See
// docs/adr/0054-reporting-and-withdrawing-a-window.md.

// addressOf returns the address of the Window room/name, as the e-mail of
// the Cahut of o names it.
func addressOf(o *mail.Outbox, room, name string) string {
	return o.Base + room + "/" + name + "/"
}

// words returns what a report says is wrong, as a sentence goes on.
func words(category string) string {
	for _, k := range Categories {
		if k.Name == category {
			return strings.ToLower(k.Words[:1]) + k.Words[1:]
		}
	}
	return category
}

// tellOfReport sends the report id about the Window room/name to the
// moderators who receive e-mail, and a receipt to email, when it is given.
func tellOfReport(ctx context.Context, c *operation.Call, id int64, category, room, name, email string) error {
	o := mail.OutboxOf(c)
	if o == nil {
		return nil
	}
	if sends, err := o.Sends(ctx, c.Tx); err != nil || !sends {
		return err
	}
	address := addressOf(o, room, name)
	subject := fmt.Sprintf("Report %d about a Window of %s", id, o.Domain)
	write := func(hello, where string) mail.Letter {
		return mail.Letter{
			Subject: subject,
			Lines: []string{
				hello,
				fmt.Sprintf("Someone reported the Window %s of %s: %s.", address, where, words(category)),
				"Read the report, open the Window, and decide in the console of the Cahut: leave the Window as it is, or withdraw it.",
			},
			Action: &mail.Action{Label: "Open Moderation", URL: o.Link("moderation")},
		}
	}
	ids, err := admins(ctx, c)
	if err != nil {
		return err
	}
	for _, uid := range ids {
		l, err := o.Prepare(ctx, c.Tx, uid, mail.Always, func(p mail.Person, _, where string) mail.Letter {
			return write("Hello "+p.Name+",", where)
		})
		if err != nil {
			return err
		}
		if l != nil {
			o.Post(c, *l)
		}
	}
	host, err := settings.ModerationAddress(ctx, c.Tx)
	if err != nil {
		return err
	}
	_, where := o.Where()
	if host != "" {
		l := write("Hello,", where)
		l.To = host
		l.After = []string{"This address receives the reports of " + where + ", as the administrators of its server set it."}
		o.Post(c, l)
	}
	if email != "" {
		o.Post(c, mail.Letter{
			To:      email,
			Subject: "Your report about a Window of " + o.Domain,
			Lines: []string{
				"Hello,",
				fmt.Sprintf("The moderators of %s received your report %d about the Window %s.", where, id, address),
				"They will read it and decide whether to withdraw the Window. Their decision will come to this address.",
			},
			After: []string{"This address was given in a report on the page " + o.Base + "-/report. Cahut forgets it once the moderators decide."},
		})
	}
	return nil
}

// tellOfDecision sends the decision about the report id to the address of
// whoever made it, when they gave one.
func tellOfDecision(c *operation.Call, id int64, decision, email, room, name string) {
	o := mail.OutboxOf(c)
	if o == nil || email == "" {
		return
	}
	_, where := o.Where()
	address := addressOf(o, room, name)
	what := fmt.Sprintf("The moderators of %s read your report %d about the Window %s, and left the Window as it is.", where, id, address)
	if decision == Withdrawn {
		what = fmt.Sprintf("The moderators of %s read your report %d about the Window %s, and withdrew the Window: it no longer opens.", where, id, address)
	}
	o.Post(c, mail.Letter{
		To:      email,
		Subject: "The decision on your report about a Window of " + o.Domain,
		Lines:   []string{"Hello,", what, "Thank you for telling them."},
		After:   []string{"Cahut no longer keeps this address."},
	})
}

// tellMembers sends the letter that write makes to the people who read
// what happens to the Window t: the members of its Room, or its author
// alone for a Window that is theirs alone.
func tellMembers(ctx context.Context, c *operation.Call, t target, write func(p mail.Person, where, address string) mail.Letter) error {
	o := mail.OutboxOf(c)
	if o == nil {
		return nil
	}
	ids := []int64{t.author}
	if t.Access != window.AccessAuthor {
		var err error
		if ids, err = members(ctx, c, t.roomID); err != nil {
			return err
		}
	}
	address := addressOf(o, t.Room, t.Name)
	for _, uid := range ids {
		l, err := o.Prepare(ctx, c.Tx, uid, mail.Always, func(p mail.Person, _, where string) mail.Letter {
			return write(p, where, address)
		})
		if err != nil {
			return err
		}
		if l != nil {
			o.Post(c, *l)
		}
	}
	return nil
}

// tellOfWithdrawal tells the members of the Room of the Window t that the
// moderators withdrew it, and why: the statement of reasons that the
// Digital Services Act asks for. With changed, it was withdrawn already,
// and tells them its new reason.
func tellOfWithdrawal(ctx context.Context, c *operation.Call, t target, w window.Withdrawal, changed bool) error {
	legal, err := settings.Legal(ctx, c.Tx)
	if err != nil {
		return err
	}
	by := "The moderators"
	if byHost(c) {
		by = "The administrators of the server"
	}
	return tellMembers(ctx, c, t, func(p mail.Person, where, address string) mail.Letter {
		subject := t.Room + "/" + t.Name + " was withdrawn"
		what := fmt.Sprintf("%s of %s withdrew the Window %s, in the Room %s, on %s.", by, where, address, t.Room, mail.Day(w.At))
		if changed {
			subject = t.Room + "/" + t.Name + " is withdrawn for a new reason"
			what = fmt.Sprintf("%s of %s changed the reason why the Window %s, in the Room %s, is withdrawn. It is withdrawn since %s.",
				by, where, address, t.Room, mail.Day(w.At))
		}
		l := mail.Letter{
			Subject: subject,
			Lines: []string{
				"Hello " + p.Name + ",",
				what,
				"It keeps everything, and serves nothing to its readers until they reinstate it: its address and its share links say that it was withdrawn. " +
					"You still find it in the console, where you can put it in the trash or delete it for good.",
			},
			Quote: &mail.Quote{By: by, Text: w.Reason},
			After: []string{"To ask about this decision, or to contest it, write to the administrators of " + where + "."},
		}
		if legal.Support != "" {
			l.Action = &mail.Action{Label: "Get support", URL: legal.Support}
		} else {
			l.Action = &mail.Action{Label: "Open the Window in the console", URL: mail.OutboxOf(c).Link("rooms/" + t.Room + "/" + t.Name)}
		}
		return l
	})
}

// tellOfReinstatement tells the members of the Room of the Window t that it
// opens again.
func tellOfReinstatement(ctx context.Context, c *operation.Call, t target, since time.Time) error {
	return tellMembers(ctx, c, t, func(p mail.Person, where, address string) mail.Letter {
		return mail.Letter{
			Subject: t.Room + "/" + t.Name + " is open again",
			Lines: []string{
				"Hello " + p.Name + ",",
				fmt.Sprintf("The moderators of %s reinstated the Window %s, which they had withdrawn on %s. "+
					"It opens again as it did, to the readers of its access and of its share links.", where, address, mail.Day(since)),
			},
			Action: &mail.Action{Label: "Open the Window in the console", URL: mail.OutboxOf(c).Link("rooms/" + t.Room + "/" + t.Name)},
		}
	})
}

// admins returns the administrators of the Cahut who are not disabled.
func admins(ctx context.Context, c *operation.Call) ([]int64, error) {
	return ids(ctx, c, "SELECT id FROM users WHERE role = 'admin' AND disabled_at IS NULL ORDER BY id")
}

// members returns the members of the Room id.
func members(ctx context.Context, c *operation.Call, id int64) ([]int64, error) {
	return ids(ctx, c, "SELECT user_id FROM room_members WHERE room_id = ? ORDER BY user_id", id)
}

func ids(ctx context.Context, c *operation.Call, q string, args ...any) ([]int64, error) {
	rows, err := c.Tx.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	var list []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return nil, err
		}
		list = append(list, id)
	}
	return list, errors.Join(rows.Err(), rows.Close())
}
