package window

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"cahut.cloud/internal/account"
	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/windata"
)

// The pages of a Window call its data at .data/ under its root: .data/sql
// for a SELECT of their own over what the declaration makes readable, and
// .data/queries/<name> for a query of the declaration. Each call is an
// operation for the readers of the Window, which finds the Window and its
// reader from what the address holds: the name of a public Window, or the
// capability of a frame. See docs/adr/0056-a-database-for-each-window.md.

// DataPath starts the part of the address of a Window where its pages call
// its data. No file of a Window starts with a dot.
const DataPath = ".data/"

// PageDataInput is the input of select_window_data and call_window_query.
type PageDataInput struct {
	Window string                     `json:"window,omitempty" jsonschema:"room/name of a public Window, as its address has it"`
	Frame  string                     `json:"frame,omitempty" jsonschema:"the capability of the frame that opened the Window, as its address has it"`
	SQL    string                     `json:"sql,omitempty" jsonschema:"for select_window_data: one SELECT over what the declaration makes readable"`
	Query  string                     `json:"query,omitempty" jsonschema:"for call_window_query: the name of a query of the declaration"`
	Params map[string]json.RawMessage `json:"params,omitempty" jsonschema:"the named parameters of the statement; :reader is given by Cahut"`
	// NoWrite refuses a query that writes: the web server sets it when the
	// address of the caller has written too much lately.
	NoWrite bool `json:"no_write,omitempty"`
}

// PageData is the output of select_window_data and call_window_query: rows,
// and for a query that writes, the rows it changed and the id of the last
// row it added.
type PageData struct {
	windata.Rows
	Changes int64 `json:"changes,omitempty"`
	LastID  int64 `json:"id,omitempty"`
	// Wrote says that the query wrote, for the limits of the web server.
	Wrote bool `json:"-"`
}

// Codes of the refusals of the calls of pages, which their scripts read.
const (
	DataInvalid    = windata.CodeInvalid
	DataTimeout    = windata.CodeTimeout
	DataConstraint = windata.CodeConstraint
	DataFull       = windata.CodeFull
	DataBusy       = windata.CodeBusy
	DataClosed     = "closed"
	DataForbidden  = "forbidden"
	DataNotFound   = "not_found"
	DataWithdrawn  = "withdrawn"
	DataTooMany    = "too_many"
)

// DataError refuses a call of a page, with a code that its script reads.
type DataError struct {
	Code    string
	Message string
	kind    error
}

func (e *DataError) Error() string { return e.Message }
func (e *DataError) Unwrap() error { return e.kind }

func refuse(code string, kind error, format string, args ...any) *DataError {
	return &DataError{Code: code, Message: fmt.Sprintf(format, args...), kind: kind}
}

// pageReader is who calls the data of a Window from one of its pages.
type pageReader struct {
	value    any   // :reader, or nil
	person   int64 // the person, or zero
	readOnly bool  // a moderator, who writes nothing
}

func (s *Service) registerPages(r *operation.Registry) {
	operation.Register(r, &operation.Op[PageDataInput, PageData]{
		Name:        "select_window_data",
		Title:       "Read the data of a Window",
		Description: "Run the SELECT of a page of a Window over what its declaration makes readable.",
		PeopleOnly:  true,
		Visitors:    true,
		ReadOnly:    true, // the database of the Cahut: the data of a Window lives in its own
		Run:         s.selectData,
	})
	operation.Register(r, &operation.Op[PageDataInput, PageData]{
		Name:        "call_window_query",
		Title:       "Call a query of a Window",
		Description: "Run a query of the declaration of a Window, for a page of that Window.",
		PeopleOnly:  true,
		Visitors:    true,
		ReadOnly:    true, // the database of the Cahut: the data of a Window lives in its own
		Run:         s.callQuery,
	})
}

func (s *Service) selectData(ctx context.Context, c *operation.Call, in PageDataInput) (PageData, error) {
	w, reader, err := s.pageWindow(ctx, c, in)
	if err != nil {
		return PageData{}, err
	}
	params, err := pageParams(in.Params, reader)
	if err != nil {
		return PageData{}, err
	}
	rows, err := s.data.Select(ctx, s.dataDir(w.ID), windata.ReaderLimits, in.SQL, params, w.Declaration.Data.Readable)
	if err != nil {
		return PageData{}, pageError(err)
	}
	return PageData{Rows: rows}, nil
}

func (s *Service) callQuery(ctx context.Context, c *operation.Call, in PageDataInput) (PageData, error) {
	w, reader, err := s.pageWindow(ctx, c, in)
	if err != nil {
		return PageData{}, err
	}
	data := w.Declaration.Data
	q, ok := data.Queries[in.Query]
	if !ok {
		return PageData{}, refuse(DataNotFound, operation.ErrNotFound, "this Window has no query %q", in.Query)
	}
	switch q.Who {
	case People:
		if reader.person == 0 {
			return PageData{}, refuse(DataForbidden, operation.ErrForbidden, "sign in to the Cahut to do this")
		}
	case Members:
		var member bool
		if reader.person != 0 {
			err := c.Tx.QueryRowContext(ctx, "SELECT EXISTS (SELECT 1 FROM room_members WHERE room_id = ? AND user_id = ?)",
				w.RoomID, reader.person).Scan(&member)
			if err != nil {
				return PageData{}, err
			}
		}
		if !member {
			return PageData{}, refuse(DataForbidden, operation.ErrForbidden, "only the members of the Room of this Window do this")
		}
	}
	writes := q.Kind != windata.Reads
	if writes {
		switch {
		case reader.readOnly:
			return PageData{}, refuse(DataForbidden, operation.ErrForbidden, "a moderator writes nothing in the data of a Window")
		case data.Closes != nil && !c.Now.Before(*data.Closes):
			return PageData{}, refuse(DataClosed, operation.ErrForbidden, "this Window stopped taking answers on %s", data.Closes.Format("2 January 2006 at 15:04 UTC"))
		case in.NoWrite:
			return PageData{}, refuse(DataTooMany, operation.ErrForbidden, "too many answers from this address: try again in a minute")
		}
	}
	params, err := pageParams(in.Params, reader)
	if err != nil {
		return PageData{}, err
	}
	lim, err := limitsOf(ctx, c, windata.ReaderLimits)
	if err != nil {
		return PageData{}, err
	}
	r, err := s.data.Call(ctx, s.dataDir(w.ID), lim, q.SQL, params, writes)
	if err != nil {
		return PageData{}, pageError(err)
	}
	if writes {
		s.copyDaily(w.ID, c.Now)
	}
	if r.Columns == nil {
		// A query that writes and returns no rows still answers lists.
		r.Columns, r.Rows.Rows = []string{}, [][]json.RawMessage{}
	}
	return PageData{Rows: r.Rows, Changes: r.Changes, LastID: r.LastID, Wrote: writes}, nil
}

// pageWindow finds the Window that a page calls, and its reader: anyone at
// the address of a public Window; in a frame, the person of its session, or
// its pass. The Window must keep data.
func (s *Service) pageWindow(ctx context.Context, c *operation.Call, in PageDataInput) (Served, pageReader, error) {
	var (
		w      Served
		reader pageReader
		err    error
	)
	missing := refuse(DataNotFound, operation.ErrNotFound, "no Window here keeps data")
	switch {
	case in.Window != "" && in.Frame == "":
		room, name, ok := strings.Cut(in.Window, "/")
		if !ok {
			return Served{}, pageReader{}, missing
		}
		w, err = lookupIn(ctx, c.Tx, c.Now, missing, "r.name = ? AND w.name = ?", room, name)
		if err == nil && w.Access != AccessPublic {
			return Served{}, pageReader{}, missing
		}
	case in.Frame != "" && in.Window == "":
		frame, ok := VerifyFrame(s.frameKey, in.Frame, c.Now)
		if !ok {
			return Served{}, pageReader{}, refuse(DataNotFound, operation.ErrNotFound, "this frame no longer opens its Window: open the Window at its address again")
		}
		w, err = lookupIn(ctx, c.Tx, c.Now, missing, "w.id = ?", frame.Window)
		if err != nil {
			break
		}
		var userID, pass int64
		switch frame.Kind {
		case FrameBySession, FrameByConsole:
			o, err := account.SessionIn(ctx, c.Tx, frame.ID, frame.Kind == FrameBySession, c.Now)
			if err != nil && !errors.Is(err, account.ErrInvalidSession) {
				return Served{}, pageReader{}, err
			}
			userID = o.UserID
		case FrameByPass:
			lasts, err := passLastsIn(ctx, c.Tx, w.ID, frame.ID, c.Now)
			if err != nil {
				return Served{}, pageReader{}, err
			}
			if lasts {
				pass = frame.ID
			}
		case FrameByModerator, FrameByHost:
			// A moderator sees what anyone sees, while the frame opens the
			// Window.
			opens := false
			if s.ModeratorOpens != nil {
				if opens, err = s.ModeratorOpens(ctx, frame, c.Now); err != nil {
					return Served{}, pageReader{}, err
				}
			}
			if !opens {
				return Served{}, pageReader{}, refuse(DataNotFound, operation.ErrNotFound, "this frame no longer opens its Window: open the Window at its address again")
			}
			reader.readOnly = true
		}
		if !reader.readOnly {
			how, err := w.opensIn(ctx, c.Tx, userID, pass)
			if err != nil {
				return Served{}, pageReader{}, err
			}
			switch {
			case how == Closed:
				return Served{}, pageReader{}, refuse(DataNotFound, operation.ErrNotFound, "this frame no longer opens its Window: open the Window at its address again")
			case userID != 0:
				reader.person, reader.value = userID, s.readerValue(w.ID, "person", userID)
			case pass != 0:
				reader.value = s.readerValue(w.ID, "pass", pass)
			}
		}
	default:
		return Served{}, pageReader{}, refuse(DataInvalid, operation.ErrInvalid, "name the Window by its public address or by its frame, not both")
	}
	if _, ok := errors.AsType[*Withdrawn](err); ok {
		return Served{}, pageReader{}, refuse(DataWithdrawn, operation.ErrForbidden, "this Window was withdrawn by the moderators of this Cahut")
	}
	if err != nil {
		return Served{}, pageReader{}, err
	}
	if w.Declaration == nil || w.Declaration.Data == nil {
		return Served{}, pageReader{}, missing
	}
	return w, reader, nil
}

// readerValue returns :reader for a person or a pass in the Window id: the
// same for them in this Window, another in every other Window, and nothing
// that tells who they are.
func (s *Service) readerValue(id int64, kind string, who int64) string {
	key := hmac.New(sha256.New, s.frameKey)
	key.Write([]byte("readers of the data of Windows"))
	mac := hmac.New(sha256.New, key.Sum(nil))
	fmt.Fprintf(mac, "%d %s %d", id, kind, who)
	return "r" + hex.EncodeToString(mac.Sum(nil)[:12])
}

// pageParams returns the parameters of a page with :reader, which a page
// cannot give.
func pageParams(params map[string]json.RawMessage, reader pageReader) (map[string]json.RawMessage, error) {
	if _, ok := params["reader"]; ok {
		return nil, refuse(DataInvalid, operation.ErrInvalid, "reader is given by Cahut: leave it out of params")
	}
	out := make(map[string]json.RawMessage, len(params)+1)
	for k, v := range params {
		out[k] = v
	}
	value, err := json.Marshal(reader.value)
	if err != nil {
		return nil, err
	}
	out["reader"] = value
	return out, nil
}

// pageError tells a page what went wrong in the database of its Window.
func pageError(err error) error {
	e, ok := errors.AsType[*windata.Error](err)
	if !ok {
		return err
	}
	switch e.Code {
	case windata.CodeInvalid, windata.CodeTimeout:
		return refuse(e.Code, operation.ErrInvalid, "%s", e.Message)
	case windata.CodeConstraint:
		return refuse(e.Code, operation.ErrInvalid, "%s", e.Message)
	case windata.CodeFull:
		return refuse(e.Code, operation.ErrForbidden, "the database of this Window is full")
	case windata.CodeBusy, windata.CodeUnavailable:
		return refuse(DataBusy, operation.ErrForbidden, "the database of this Window is busy: try again in a moment")
	}
	return err
}
