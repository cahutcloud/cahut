package oauth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"slices"
	"strings"

	"cahut.cloud/internal/credential"
	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/store"
)

// Request is an authorization request, with the parameters of its address
// at /oauth/authorize that the console passes on.
type Request struct {
	ResponseType        string `json:"response_type,omitempty"`
	ClientID            string `json:"client_id"`
	RedirectURI         string `json:"redirect_uri,omitempty"`
	Scope               string `json:"scope,omitempty"`
	State               string `json:"state,omitempty"`
	CodeChallenge       string `json:"code_challenge,omitempty"`
	CodeChallengeMethod string `json:"code_challenge_method,omitempty"`
	Resource            string `json:"resource,omitempty"`
}

// checked is an authorization request once checked.
type checked struct {
	client   Client
	redirect string // the redirect URI, which the client registered
	scope    string // the most the client asks for
	resource string
	// refusal is set when the request is invalid but its client and its
	// redirect URI are not: the client hears why at its redirect URI.
	refusal *refusal
}

// refusal is an error of an authorization request, in the terms of RFC
// 6749.
type refusal struct{ code, description string }

var challengeFormat = regexp.MustCompile(`^[A-Za-z0-9_-]{43}$`)

// check checks an authorization request. It returns an error wrapping
// operation.ErrInvalid when the client or the redirect URI is invalid: the
// request then cannot go back to the client, which may not be who it says.
func (s *Server) check(ctx context.Context, c *operation.Call, req Request) (checked, error) {
	client, err := s.client(ctx, c, req.ClientID)
	if err != nil {
		return checked{}, err
	}
	redirect, err := client.redirectURI(req.RedirectURI)
	if err != nil {
		return checked{}, err
	}
	ch := checked{client: client, redirect: redirect, scope: Write, resource: s.cfg.Resource}
	refuse := func(code, description string) (checked, error) {
		ch.refusal = &refusal{code, description}
		return ch, nil
	}
	switch {
	case req.ResponseType != "code":
		return refuse("unsupported_response_type", "Cahut answers response_type=code only")
	case req.CodeChallenge == "":
		return refuse("invalid_request", "PKCE is required: send a code_challenge, with code_challenge_method=S256")
	case req.CodeChallengeMethod != "S256":
		return refuse("invalid_request", "code_challenge_method must be S256")
	case !challengeFormat.MatchString(req.CodeChallenge):
		return refuse("invalid_request", "code_challenge must be the base64url SHA-256 of the code verifier, 43 characters")
	case req.Resource != "" && !s.sameResource(req.Resource):
		return refuse("invalid_target", "this server grants access to "+s.cfg.Resource+" only")
	}
	if req.Scope != "" {
		// Scopes that Cahut does not know are ignored: some clients add
		// their own. A client that asks for read only gets read only.
		scopes := strings.Fields(req.Scope)
		if !slices.Contains(scopes, Write) && slices.Contains(scopes, Read) {
			ch.scope = Read
		}
	}
	return ch, nil
}

// back returns the redirect URI of a checked request, with params, the
// state of the request and the issuer (RFC 9207) in its query.
func (s *Server) back(ch checked, req Request, params url.Values) string {
	u, err := url.Parse(ch.redirect)
	if err != nil {
		return ch.redirect // checked already: cannot happen
	}
	q := u.Query()
	for k, v := range params {
		q[k] = v
	}
	if req.State != "" {
		q.Set("state", req.State)
	}
	q.Set("iss", s.cfg.Issuer)
	u.RawQuery = q.Encode()
	return u.String()
}

// redirectHost names where a redirect URI leads, for a person to read: its
// host, or its scheme and host for the private scheme of an application.
func redirectHost(uri string) string {
	uri = strings.ToLower(uri) // as the browser reads it: cIaude.ai is ciaude.ai
	u, err := url.Parse(uri)
	if err != nil {
		return uri
	}
	if u.Scheme == "http" || u.Scheme == "https" {
		return u.Host
	}
	return u.Scheme + "://" + u.Host
}

// person returns the id of the person who consents: someone who signed in
// to the console of this Cahut.
func person(c *operation.Call) (int64, error) {
	if c.Actor.Kind != operation.Person {
		return 0, operation.Forbiddenf("only a person of this Cahut, signed in to its console, connects an AI to it")
	}
	return c.Actor.UserID, nil
}

// RoomChoice is a Room that a person may grant an AI.
type RoomChoice struct {
	Name  string `json:"name"`
	Title string `json:"title"`
	Kind  string `json:"kind" jsonschema:"personal or team"`
}

// rooms returns the Rooms of the person userID, their personal Room first.
func rooms(ctx context.Context, tx *sql.Tx, userID int64) ([]RoomChoice, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT r.name, r.title, r.kind FROM rooms r JOIN room_members m ON m.room_id = r.id
		WHERE m.user_id = ? ORDER BY r.kind = 'team', r.name`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []RoomChoice{}
	for rows.Next() {
		var r RoomChoice
		if err := rows.Scan(&r.Name, &r.Title, &r.Kind); err != nil {
			return nil, err
		}
		list = append(list, r)
	}
	return list, rows.Err()
}

// DescribeInput is the input of describe_authorization.
type DescribeInput struct {
	Request Request `json:"request"`
}

// Authorization describes an authorization request to the person who
// consents.
type Authorization struct {
	Client     string `json:"client" jsonschema:"the name of the AI client, as it gives it"`
	ClientURI  string `json:"client_uri,omitempty" jsonschema:"its web page, as it gives it"`
	ClientHost string `json:"client_host,omitempty" jsonschema:"the host of its client metadata document, which vouches for it"`
	// RedirectHost is where the AI will be sent back: anyone can register
	// a client under any name, but not receive codes at another host.
	RedirectHost string       `json:"redirect_host" jsonschema:"where the browser goes back to after consent"`
	Scope        string       `json:"scope" jsonschema:"the most the client asks for: read, or write"`
	Rooms        []RoomChoice `json:"rooms" jsonschema:"the Rooms the person may grant"`
	// Redirect is set when the request is invalid: the browser goes there,
	// and the client learns why.
	Redirect string `json:"redirect,omitempty" jsonschema:"when set, the request is invalid: send the browser to this address"`
	Error    string `json:"error,omitempty" jsonschema:"why the request is invalid, when it is"`
}

func (s *Server) describeOp() *operation.Op[DescribeInput, Authorization] {
	return &operation.Op[DescribeInput, Authorization]{
		Name:  "describe_authorization",
		Title: "Describe a connection request",
		Description: "Check the request of an AI client to connect to this Cahut, and describe it to the person who consents: " +
			"the client, where it will be sent back, the scope it asks for, and the Rooms the person may grant.",
		PeopleOnly: true,
		ReadOnly:   true,
		Prepare: func(ctx context.Context, c *operation.Call, in DescribeInput) error {
			return s.prefetch(ctx, c, in.Request.ClientID)
		},
		Run: func(ctx context.Context, c *operation.Call, in DescribeInput) (Authorization, error) {
			userID, err := person(c)
			if err != nil {
				return Authorization{}, err
			}
			ch, err := s.check(ctx, c, in.Request)
			if err != nil {
				return Authorization{}, err
			}
			c.Log(ch.client.Name, redirectHost(ch.redirect))
			a := Authorization{Client: ch.client.Name, ClientURI: ch.client.URI, RedirectHost: redirectHost(ch.redirect), Scope: ch.scope}
			if ch.client.Document {
				a.ClientHost = redirectHost(ch.client.ID)
			}
			if ch.refusal != nil {
				a.Error = ch.refusal.description
				a.Redirect = s.back(ch, in.Request, url.Values{"error": {ch.refusal.code}, "error_description": {ch.refusal.description}})
				return a, nil
			}
			a.Rooms, err = rooms(ctx, c.Tx, userID)
			return a, err
		},
	}
}

// ApproveInput is the input of approve_authorization.
type ApproveInput struct {
	Request  Request  `json:"request"`
	Scope    string   `json:"scope" jsonschema:"read, or write; at most what the client asks for"`
	AllRooms bool     `json:"all_rooms,omitempty" jsonschema:"every Room of the person, including those they join later"`
	Rooms    []string `json:"rooms,omitempty" jsonschema:"the Rooms granted, by name, when all_rooms is false"`
}

// Redirect is where the browser goes once the person has answered.
type Redirect struct {
	Redirect string `json:"redirect" jsonschema:"send the browser to this address"`
}

func (s *Server) approveOp() *operation.Op[ApproveInput, Redirect] {
	return &operation.Op[ApproveInput, Redirect]{
		Name:  "approve_authorization",
		Title: "Connect an AI",
		Description: "Connect an AI client to this Cahut, for the person who consents, with the scope and the Rooms they grant. " +
			"Returns the address that sends the browser back to the client, with its authorization code.",
		PeopleOnly: true,
		Prepare: func(ctx context.Context, c *operation.Call, in ApproveInput) error {
			return s.prefetch(ctx, c, in.Request.ClientID)
		},
		Run: func(ctx context.Context, c *operation.Call, in ApproveInput) (Redirect, error) {
			userID, err := person(c)
			if err != nil {
				return Redirect{}, err
			}
			ch, err := s.check(ctx, c, in.Request)
			if err != nil {
				return Redirect{}, err
			}
			c.Log(ch.client.Name, "")
			if ch.refusal != nil {
				return Redirect{}, operation.Invalidf("%s", ch.refusal.description)
			}
			switch {
			case in.Scope != Read && in.Scope != Write:
				return Redirect{}, operation.Invalidf("scope %q: use read or write", in.Scope)
			case in.Scope == Write && ch.scope == Read:
				return Redirect{}, operation.Invalidf("the client asks for read only: grant read")
			case !in.AllRooms && len(in.Rooms) == 0:
				return Redirect{}, operation.Invalidf("grant at least one Room, or every Room with all_rooms")
			}
			var ids []int64
			if !in.AllRooms {
				if ids, err = roomIDs(ctx, c.Tx, userID, in.Rooms); err != nil {
					return Redirect{}, err
				}
			}
			now := store.FormatTime(c.Now)
			var id int64
			err = c.Tx.QueryRowContext(ctx, `
				INSERT INTO connections (user_id, client_id, client_name, redirect_host, scope, all_rooms, created_at)
				VALUES (?, ?, ?, ?, ?, ?, ?) RETURNING id`,
				userID, ch.client.ID, ch.client.Name, redirectHost(ch.redirect), in.Scope, in.AllRooms, now).Scan(&id)
			if err != nil {
				return Redirect{}, err
			}
			for _, room := range ids {
				if _, err := c.Tx.ExecContext(ctx, "INSERT INTO connection_rooms (connection_id, room_id) VALUES (?, ?)", id, room); err != nil {
					return Redirect{}, err
				}
			}
			if len(ids) == 1 {
				c.InRoom(ids[0])
			}
			if _, err := c.Tx.ExecContext(ctx, "UPDATE oauth_clients SET connected_at = ? WHERE id = ?", now, ch.client.ID); err != nil {
				return Redirect{}, err
			}
			code := credential.Token("")
			_, err = c.Tx.ExecContext(ctx, `
				INSERT INTO oauth_codes (hash, connection_id, client_id, redirect_uri, challenge, resource, expires_at)
				VALUES (?, ?, ?, ?, ?, ?, ?)`,
				credential.Hash(code), id, ch.client.ID, ch.redirect, in.Request.CodeChallenge, ch.resource,
				store.FormatTime(c.Now.Add(CodeLifetime)))
			if err != nil {
				return Redirect{}, err
			}
			reach := "every Room"
			if !in.AllRooms {
				reach = strings.Join(in.Rooms, ", ")
			}
			c.Log(fmt.Sprintf("connection %d of %s", id, c.Actor.Handle), fmt.Sprintf("%s, %s, %s", ch.client.Name, in.Scope, reach))
			return Redirect{Redirect: s.back(ch, in.Request, url.Values{"code": {code}})}, nil
		},
	}
}

// roomIDs returns the ids of the Rooms called names, which the person
// userID must be in.
func roomIDs(ctx context.Context, tx *sql.Tx, userID int64, names []string) ([]int64, error) {
	var ids []int64
	for _, name := range names {
		var id int64
		err := tx.QueryRowContext(ctx, `
			SELECT r.id FROM rooms r JOIN room_members m ON m.room_id = r.id
			WHERE r.name = ? AND m.user_id = ?`, name, userID).Scan(&id)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, operation.NotFoundf("no Room %q among yours", name)
		}
		if err != nil {
			return nil, err
		}
		if !slices.Contains(ids, id) {
			ids = append(ids, id)
		}
	}
	return ids, nil
}

// DenyInput is the input of deny_authorization.
type DenyInput struct {
	Request Request `json:"request"`
}

func (s *Server) denyOp() *operation.Op[DenyInput, Redirect] {
	return &operation.Op[DenyInput, Redirect]{
		Name:        "deny_authorization",
		Title:       "Refuse an AI",
		Description: "Refuse the request of an AI client to connect to this Cahut. Returns the address that tells the client so.",
		PeopleOnly:  true,
		Run: func(ctx context.Context, c *operation.Call, in DenyInput) (Redirect, error) {
			if _, err := person(c); err != nil {
				return Redirect{}, err
			}
			ch, err := s.check(ctx, c, in.Request)
			if err != nil {
				return Redirect{}, err
			}
			c.Log(ch.client.Name, redirectHost(ch.redirect))
			return Redirect{Redirect: s.back(ch, in.Request, url.Values{
				"error": {"access_denied"}, "error_description": {"the person refused to connect this AI"},
			})}, nil
		},
	}
}
