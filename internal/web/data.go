package web

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/window"
)

// The pages of a Window call its data at .data/ under its root, from a
// sandbox whose origin is opaque: these calls carry no credentials, and the
// address says which Window and which reader. See
// docs/adr/0056-a-database-for-each-window.md.

// The calls of pages are limited by address and by Window, reads and writes
// each with a budget of their own.
const (
	DataReadBurst  = 60
	DataReadEvery  = time.Second
	DataWriteBurst = 10
	DataWriteEvery = 6 * time.Second
)

// maxDataRequest bounds what a page sends in one call.
const maxDataRequest = 64 << 10

// pageCall is the body of a call of a page.
type pageCall struct {
	SQL    string                     `json:"sql"`
	Params map[string]json.RawMessage `json:"params"`
}

// serveData answers a call of a page of a public Window, at its address:
// what follows the address starts with .data/.
func (s *Server) serveData(w http.ResponseWriter, r *http.Request) {
	rest, ok := strings.CutPrefix(r.PathValue("rest"), window.DataPath)
	if !ok {
		notAllowed(w)
		return
	}
	r.SetPathValue("rest", rest)
	s.answerData(w, r, window.PageDataInput{Window: r.PathValue("room") + "/" + r.PathValue("window")})
}

// serveFrameData answers a call of a page in a frame.
func (s *Server) serveFrameData(w http.ResponseWriter, r *http.Request) {
	s.answerData(w, r, window.PageDataInput{Frame: r.PathValue("capability")})
}

// notAllowed answers a method that the files of a Window do not take, as
// the routes of the server would without the data of Windows.
func notAllowed(w http.ResponseWriter) {
	w.Header().Set("Allow", "GET, HEAD")
	http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
}

// preflightData answers the preflight of a call of a page, which a sandbox
// makes for a POST of JSON.
func preflightData(w http.ResponseWriter, r *http.Request) {
	if r.PathValue("capability") == "" && !strings.HasPrefix(r.PathValue("rest"), window.DataPath) {
		notAllowed(w)
		return
	}
	h := w.Header()
	h.Set("Access-Control-Allow-Origin", "*")
	h.Set("Access-Control-Allow-Methods", "POST")
	h.Set("Access-Control-Allow-Headers", "Content-Type")
	h.Set("Access-Control-Max-Age", "600")
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) answerData(w http.ResponseWriter, r *http.Request, in window.PageDataInput) {
	h := w.Header()
	h.Set("Access-Control-Allow-Origin", "*")
	h.Set("Content-Type", "application/json")
	h.Set("Cache-Control", "no-store")
	h.Set("X-Content-Type-Options", "nosniff")

	op := "select_window_data"
	rest := r.PathValue("rest")
	switch {
	case rest == "sql":
	case strings.HasPrefix(rest, "queries/") && !strings.Contains(strings.TrimPrefix(rest, "queries/"), "/"):
		op, in.Query = "call_window_query", strings.TrimPrefix(rest, "queries/")
	default:
		pageFailure(w, http.StatusNotFound, window.DataNotFound, "call .data/sql or .data/queries/<name>")
		return
	}
	key := clientKey(r) + " " + in.Window + in.Frame
	if !s.dataReads.allow(key) {
		pageFailure(w, http.StatusTooManyRequests, window.DataTooMany, "too many calls from this address: try again in a moment")
		return
	}
	var body pageCall
	r.Body = http.MaxBytesReader(w, r.Body, maxDataRequest)
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
			pageFailure(w, http.StatusRequestEntityTooLarge, window.DataInvalid, "a call sends 64 KiB at most")
			return
		}
		pageFailure(w, http.StatusBadRequest, window.DataInvalid, `send JSON, such as {"sql": "SELECT …", "params": {…}}`)
		return
	}
	in.SQL, in.Params = body.SQL, body.Params
	in.NoWrite = s.dataWrites.spent(key)

	c := siteOf(r).cahut
	actor := operation.Actor{Kind: operation.Visitor, Via: "web", Domain: c.Domain, Address: clientKey(r)}
	out, err := c.Registry.Invoke(r.Context(), op, actor, mustJSON(in))
	if err != nil {
		var refusal *window.DataError
		switch {
		case errors.As(err, &refusal):
			pageFailure(w, dataStatus(refusal.Code), refusal.Code, refusal.Message)
		case operation.Public(err):
			pageFailure(w, http.StatusBadRequest, window.DataInvalid, err.Error())
		default:
			s.log.Error("answer a call of a page", "cahut", c.Domain, "op", op, "error", err)
			pageFailure(w, http.StatusInternalServerError, "internal", "Internal error.")
		}
		return
	}
	data := out.(window.PageData)
	if data.Wrote {
		s.dataWrites.allow(key)
	}
	_ = json.NewEncoder(w).Encode(data)
}

// dataStatus returns the status of the response to a refused call.
func dataStatus(code string) int {
	switch code {
	case window.DataNotFound:
		return http.StatusNotFound
	case window.DataWithdrawn:
		return http.StatusGone
	case window.DataForbidden, window.DataClosed:
		return http.StatusForbidden
	case window.DataConstraint:
		return http.StatusConflict
	case window.DataTooMany:
		return http.StatusTooManyRequests
	case window.DataFull:
		return http.StatusInsufficientStorage
	case window.DataBusy:
		return http.StatusServiceUnavailable
	}
	return http.StatusBadRequest
}

// pageFailure answers a call of a page with an error that its script reads.
func pageFailure(w http.ResponseWriter, status int, code, message string) {
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{"error": map[string]string{"code": code, "message": message}})
}

func mustJSON(v any) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}
