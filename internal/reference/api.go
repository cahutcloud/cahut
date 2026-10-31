package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

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
	m.block(fmt.Sprintf("Each Cahut serves the API of its console under %s of its domain, next to the console itself at %s. "+
		"Every domain of the server serves the API of the console of the server under %s, next to that console at %s. "+
		"The consoles use these APIs and nothing else, and AIs reach the same operations over [MCP](/reference/mcp-tools/): "+
		"there is one registry of [operations](/reference/operations/).",
		code(web.ConsolePath+"api/"), code(web.ConsolePath), code(web.ServerPath+"api/"), code(web.ServerPath)))

	m.block("## Calls")
	m.table([]string{"Route", "What it runs"}, [][]string{
		{code("POST " + web.ConsolePath + "api/<operation>"), "An operation of the Cahut of the domain"},
		{code("POST " + web.ServerPath + "api/<operation>"), "An operation of the server, for its administrators"},
		{code("POST " + web.ServerPath + "api/cahuts/<domain>/<operation>"), "An operation that serves the server, in the Cahut of that domain, for the administrators of the server"},
	})
	m.block("The body of the request is the input of the operation, in JSON, and the body of the response its output, " +
		"with the same names and schemas as the MCP tools. Fields that the operation does not know are refused. " +
		"The API takes `POST` only, reads included, and answers no CORS preflight.")
	m.code("bash", `curl -X POST https://cahut.example.com`+web.ConsolePath+`api/list_rooms \
  -H "Authorization: Bearer <session token>" -d '{}'`)

	m.block("## Sessions")
	m.block("The operations that sign in take no session: " + code("sign_in") + " and " + code("activate_account") +
		", at both routes. They return a session token, which every other call sends in the header " +
		code("Authorization: Bearer <token>") + ". The API reads no cookie.")
	m.table([]string{"Token", "Whose session", "Where it holds"}, [][]string{
		{code(account.SessionPrefix + "…"), "A person of the Cahut", "The Cahut that holds the person"},
		{code(server.SessionPrefix + "…"), "An administrator of the server", "The domain where it opened, for the server and for every Cahut that it names"},
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
			"for one handle, %d refused from one address or %d from everywhere in %d minutes. An IPv6 address counts with its /64",
			web.SignInBurst, int(web.SignInEvery.Seconds()), credential.MaxRefusals, credential.MaxHandleRefusals, int(credential.RefusalWindow.Minutes()))},
		{status(http.StatusInternalServerError), "Anything else"},
	})
	m.block("A refusal that people meet in the console gives its code too, which never changes, and the values of its message, as in " +
		code(fmt.Sprintf(`{"error": "invalid input: the password has 9 characters: use at least %d", "code": "password_too_short", "values": {"length": 9, "limit": %d}}`,
			credential.MinPassword, credential.MinPassword)) +
		". The console says it in the language of its person. A program tells these refusals apart by their code: " +
		"the wording of a message may change.")
	m.table([]string{"Code", "Values", "Message"}, refusalRows(operation.Refusals()))

	m.block("## Operations of a Cahut")
	m.table([]string{"Route", "Who calls it", "Session"}, apiRows(web.ConsolePath+"api/", s.cahutOps))
	m.block("## Operations of the server")
	m.table([]string{"Route", "Who calls it", "Session"}, apiRows(web.ServerPath+"api/", s.serverOps))
	m.block("## Operations that serve the server in a Cahut")
	m.block("The administrators of the server call these operations in every Cahut of the server, " +
		"from the domain where they signed in: the Cahut is the one that the route names.")
	var forServer []operation.Info
	for _, op := range s.cahutOps {
		if op.ServesServer() {
			forServer = append(forServer, op)
		}
	}
	m.table([]string{"Route", "Who calls it", "Session"}, apiRows(web.ServerPath+"api/cahuts/<domain>/", forServer))

	m.block("## Admin API")
	m.block("A program on the server, such as a site where people sign up, manages the people of a Cahut through the admin API, " +
		"with a service token, and without a Knock. The API is off unless " + code("cahut serve --admin-api <address>") +
		", or " + code("CAHUT_ADMIN_API") + ", gives it an address of the machine itself, such as " + code("127.0.0.1:8090") +
		": Cahut refuses to start with another one, and the API refuses every client that is not on the machine.")
	m.table([]string{"Route", "What it runs"}, [][]string{
		{code("POST /api/<operation>"), "An operation of the Cahut that the host of the request names, among those open to services"},
	})
	m.block("Calls, inputs, outputs and errors are those of the API of the console, with a service token in the header " +
		code("Authorization: Bearer <token>") + " in place of a session. A service token starts with " +
		code(account.ServiceTokenPrefix) + ". It belongs to one Cahut, and opens nothing else: neither the console, nor MCP, " +
		"nor the Windows. The administrators of the server create one with " + code("cahut service-token create") +
		", see them with " + code("cahut service-token list") + " and revoke one with " + code("cahut service-token revoke") +
		", which ends its access at once.")
	m.code("bash", `curl -X POST http://127.0.0.1:8090/api/check_handle \
  -H "Host: cahut.example.com" -H "Authorization: Bearer <service token>" -d '{"handle": "ada"}'`)
	var services []operation.Info
	for _, op := range s.cahutOps {
		if op.Services {
			services = append(services, op)
		}
	}
	m.table([]string{"Route", "What it does"}, serviceRows(services))
	return p
}

func refusalRows(refusals []operation.RefusalInfo) [][]string {
	var rows [][]string
	for _, r := range refusals {
		var values []string
		if r.Values != nil {
			for f := range r.Values.Fields() {
				if name, _, _ := strings.Cut(f.Tag.Get("json"), ","); f.IsExported() && name != "" && name != "-" {
					values = append(values, code(name))
				}
			}
		}
		rows = append(rows, []string{code(r.Code), strings.Join(values, ", "), text(capitalize(r.Message))})
	}
	return rows
}

func serviceRows(ops []operation.Info) [][]string {
	var rows [][]string
	for _, op := range ops {
		rows = append(rows, []string{code("POST /api/" + op.Name), text(op.Description)})
	}
	return rows
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
