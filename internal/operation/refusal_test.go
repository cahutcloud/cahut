package operation

import (
	"errors"
	"slices"
	"testing"
)

// The refusals of the tests, declared once, as packages declare theirs.
var (
	errRoomClosed  = NewRefusal("test_room_closed", ErrForbidden, "this room is closed")
	errSessionGone = NewRefusal("test_session_gone", nil, "the session ended")
	roomFull       = NewRefusalOf[fullRoom]("test_room_full", ErrInvalid, "the room {room} holds {count} people: {count} at most")
	noRoom         = NewRefusalOf[namedRoom]("test_no_room", ErrNotFound, "no room {room} here")
)

type namedRoom struct {
	Room Quoted `json:"room"`
}

type fullRoom struct {
	Room   string `json:"room"`
	Count  int    `json:"count"`
	hidden int
}

func TestNewRefusal(t *testing.T) {
	var r *Refusal
	if !errors.As(errRoomClosed, &r) || r.Code != "test_room_closed" || r.Values != nil {
		t.Fatalf("refusal = %#v", errRoomClosed)
	}
	if got := errRoomClosed.Error(); got != "not allowed: this room is closed" {
		t.Errorf("message = %q", got)
	}
	if !errors.Is(errRoomClosed, ErrForbidden) || !Public(errRoomClosed) {
		t.Error("the refusal is not a public ErrForbidden")
	}

	// A refusal of an adapter has no kind, and its message alone.
	if got := errSessionGone.Error(); got != "the session ended" || Public(errSessionGone) {
		t.Errorf("message = %q, public = %v", got, Public(errSessionGone))
	}

	// The values fill the message, and go with the refusal.
	err := roomFull(fullRoom{Room: "team", Count: 12, hidden: 3})
	if got := err.Error(); got != "invalid input: the room team holds 12 people: 12 at most" {
		t.Errorf("message = %q", got)
	}
	if !errors.As(err, &r) || r.Code != "test_room_full" || r.Values != (fullRoom{Room: "team", Count: 12, hidden: 3}) {
		t.Errorf("refusal = %#v", err)
	}
	if !errors.Is(err, ErrInvalid) {
		t.Error("the refusal is not an ErrInvalid")
	}

	// A quoted value is quoted in the message, as %q does, and given as it
	// is with the refusal.
	err = noRoom(namedRoom{Room: `the "big" one`})
	if got := err.Error(); got != `not found: no room "the \"big\" one" here` {
		t.Errorf("message = %q", got)
	}
	if !errors.As(err, &r) || r.Values.(namedRoom).Room != `the "big" one` {
		t.Errorf("refusal = %#v", err)
	}
}

func TestRefusalsDeclared(t *testing.T) {
	var codes []string
	for _, r := range Refusals() {
		codes = append(codes, r.Code)
		if r.Code == "test_room_full" && (r.Values == nil || r.Message != "the room {room} holds {count} people: {count} at most") {
			t.Errorf("info = %+v", r)
		}
	}
	if !slices.IsSorted(codes) {
		t.Errorf("codes = %v, want them sorted", codes)
	}
	for _, want := range []string{"test_room_closed", "test_room_full", "test_session_gone"} {
		if !slices.Contains(codes, want) {
			t.Errorf("codes = %v, want %s", codes, want)
		}
	}
}

func TestRefusalMisdeclared(t *testing.T) {
	type values struct {
		Room string `json:"room"`
	}
	for name, declare := range map[string]func(){
		"a code taken":              func() { _ = NewRefusal("test_room_closed", ErrForbidden, "closed again") },
		"a code not snake_case":     func() { _ = NewRefusal("RoomClosed", ErrForbidden, "closed") },
		"a value not named":         func() { NewRefusalOf[values]("test_unnamed", ErrInvalid, "the room is full") },
		"a name with no value":      func() { NewRefusalOf[values]("test_unknown", ErrInvalid, "{room} holds {count}") },
		"values that are no struct": func() { NewRefusalOf[string]("test_string", ErrInvalid, "the room is full") },
	} {
		func() {
			defer func() {
				if recover() == nil {
					t.Errorf("%s: no panic", name)
				}
			}()
			declare()
		}()
	}
}
