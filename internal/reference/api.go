package main

import (
	"fmt"
	"net/http"
	"strconv"

	"cahut.cloud/internal/account"
	"cahut.cloud/internal/credential"
	"cahut.cloud/internal/operation"
	"cahut.cloud/internal/server"
	"cahut.cloud/internal/web"
)

// apiPage describes the HTTP API of the console: its routes, its sessions,
// its errors and its operations.
func apiPage(s sample) page {
	p := page{
		file:        "http-api.md",
		title:       "HTTP API",
		description: "The API of the console: the operations of a Cahut and of its server over HTTP, with a session.",
		order:       4,
	}
	m := &p.body
	m.block(fmt.Sprintf("Each Cahut serves the API of its console on port %d of its domain, next to the console itself. "+
		"The console uses this API and nothing else, and AIs reach the same operations over [MCP](/reference/mcp-tools/): "+
		"there is one registry of [operations](/reference/operations/).", web.ConsolePort))

	m.block("## Calls")
	m.table([]string{"Route", "What it runs"}, [][]string{
		{code("POST /api/<operation>"), "An operation of the Cahut of the domain"},
		{code("POST /api/server/<operation>"), "An operation of the server, for its administrators"},
	})
	m.block("The body of the request is the input of the operation, in JSON, and the body of the response its output, " +
		"with the same names and schemas as the MCP tools. Fields that the operation does not know are refused. " +
		"The API takes `POST` only, reads included, and answers no CORS preflight.")
	m.code("bash", fmt.Sprintf(`curl -X POST https://cahut.example.com:%d/api/list_rooms \
  -H "Authorization: Bearer <session token>" -d '{}'`, web.ConsolePort))

	m.block("## Sessions")
	m.block("The operations that sign in take no session: " + code("sign_in") + " and " + code("activate_account") +
		", at both routes. They return a session token, which every other call sends in the header " +
		code("Authorization: Bearer <token>") + ". The API reads no cookie.")
	m.table([]string{"Token", "Whose session", "Where it holds"}, [][]string{
		{code(account.SessionPrefix + "…"), "A person of the Cahut", "The Cahut that holds the person"},
		{code(server.SessionPrefix + "…"), "An administrator of the server", "The domain where it opened, for the server and for that Cahut"},
	})
	m.block("A session ends 7 days after its last use, and 30 days after it opened in any case. " +
		"Personal tokens, which start with " + code(account.TokenPrefix) + ", are for AIs over MCP: the API refuses them.")

	m.block("## Errors")
	m.block("A failed call answers with a status and a body such as " + code(`{"error": "invalid input: title is empty"}`) +
		". The message says what to change. An internal error says only that it happened: the Logbook keeps the details.")
	m.table([]string{"Status", "When"}, [][]string{
		{status(http.StatusBadRequest), "The input is invalid"},
		{status(http.StatusUnauthorized), "The session is missing, unknown or ended, or holds at another domain"},
		{status(http.StatusForbidden), "The caller may not do this"},
		{status(http.StatusNotFound), "The operation, or what it names, does not exist, or the caller may not see it"},
		{status(http.StatusMethodNotAllowed), "The request is not a " + code("POST")},
		{status(http.StatusConflict), "The operation needs the approval of a person, which only AIs meet"},
		{status(http.StatusRequestEntityTooLarge), "The body is larger than " + mebibytes(web.MaxAPIRequestSize)},
		{status(http.StatusMisdirectedRequest), "The domain is not that of a Cahut of the server"},
		{status(http.StatusTooManyRequests), fmt.Sprintf("Too many sign-ins: from one address, %d at once, then one every %d seconds; "+
			"for one handle, %d refused in %d minutes", web.SignInBurst, int(web.SignInEvery.Seconds()),
			credential.MaxRefusals, int(credential.RefusalWindow.Minutes()))},
		{status(http.StatusInternalServerError), "Anything else"},
	})

	m.block("## Operations of a Cahut")
	m.table([]string{"Route", "Who calls it", "Session"}, apiRows("/api/", s.cahutOps))
	m.block("## Operations of the server")
	m.table([]string{"Route", "Who calls it", "Session"}, apiRows("/api/server/", s.serverOps))
	return p
}

func apiRows(prefix string, ops []operation.Info) [][]string {
	var rows [][]string
	for _, op := range ops {
		who, session := capitalize(op.Level.String()), "yes"
		if op.Visitors {
			who, session = "Anyone, to sign in", "no"
		}
		rows = append(rows, []string{code("POST " + prefix + op.Name), who, session})
	}
	return rows
}

func status(code int) string {
	return strconv.Itoa(code) + " " + http.StatusText(code)
}
