package cahut

import (
	"errors"
	"strings"
	"testing"

	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/settings"
)

// The administrators of the server set the listing of a Cahut in the
// directory of connectors of Claude, which get_cahut gives to everyone in
// the Cahut, for the console to lead people to it. See
// docs/adr/0061-a-guide-to-the-first-ai.md.
func TestSetListings(t *testing.T) {
	l := openLimits(t)
	olivier := l.person("olivier", true)
	ana := l.person("ana", false)
	set := func(claude string) (settings.Listings, error) {
		return operation.Do(l.ctx, l.c.Registry, settings.SetListings, l.admin, settings.SetListingsInput{Claude: claude})
	}

	// A Cahut has no listing until they set it.
	if a := l.about(ana); a.Listings != (settings.Listings{}) {
		t.Errorf("the listings of a new Cahut = %+v", a.Listings)
	}

	out, err := set(" https://claude.ai/directory/my-cahut ")
	if err != nil || out.Claude != "https://claude.ai/directory/my-cahut" {
		t.Fatalf("set_listings = %+v %v", out, err)
	}
	for _, who := range []operation.Actor{olivier, ana} {
		if a := l.about(who); a.Listings.Claude != "https://claude.ai/directory/my-cahut" {
			t.Errorf("get_cahut for %s after set_listings = %+v", who.Handle, a.Listings)
		}
	}
	if e := l.lastEntry(); e.Operation != "set_listings" || e.Target != "listings" || e.Detail != "claude none → https://claude.ai/directory/my-cahut" {
		t.Errorf("logbook = %+v", e)
	}
	if _, err := set("https://claude.ai/directory/my-cahut"); err != nil {
		t.Fatal(err)
	}
	if e := l.lastEntry(); e.Detail != "unchanged" {
		t.Errorf("logbook of the same listing = %+v", e)
	}

	// Only the address of a listing in the directory of Claude.
	for _, bad := range []string{
		"http://claude.ai/directory/my-cahut",
		"https://claude.ai/directory/",
		"https://claude.ai/customize/connectors",
		"https://example.com/directory/my-cahut",
		"https://claude.ai.example.com/directory/my-cahut",
		"https://claude.ai/directory/my-cahut?ref=x",
		"https://claude.ai/directory/../customize",
		"https://claude.ai/directory/my cahut",
		"javascript:alert(1)",
	} {
		_, err := set(bad)
		var r *operation.Refusal
		if !errors.Is(err, operation.ErrInvalid) || !errors.As(err, &r) || r.Code != "listing_invalid" || !strings.HasPrefix(err.Error(), "invalid input: claude: ") {
			t.Errorf("set_listings %q: err = %v, want listing_invalid", bad, err)
		}
	}
	if a := l.about(ana); a.Listings.Claude != "https://claude.ai/directory/my-cahut" {
		t.Errorf("the listing after refused calls = %+v", a.Listings)
	}

	// An empty address removes it.
	if out, err := set(""); err != nil || out != (settings.Listings{}) {
		t.Errorf("set_listings to nothing = %+v %v", out, err)
	}
	if a := l.about(ana); a.Listings != (settings.Listings{}) {
		t.Errorf("the listings once removed = %+v", a.Listings)
	}
	if e := l.lastEntry(); e.Detail != "claude https://claude.ai/directory/my-cahut → none" {
		t.Errorf("logbook = %+v", e)
	}

	// Only the administrators of the server set it: not those of the
	// Cahut, and no AI, which is never offered the operation.
	_, err = operation.Do(l.ctx, l.c.Registry, settings.SetListings, olivier, settings.SetListingsInput{})
	l.refused(err, "administrators of the server")
	info, _ := l.c.Registry.Lookup("set_listings")
	if info.ForAIs(true, false) {
		t.Error("set_listings is offered to the AIs of an administrator of the Cahut")
	}
}
