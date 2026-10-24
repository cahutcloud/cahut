# 25. The first start is guided in the browser, from a single-use link

- Status: Accepted, extended by [ADR 31](0031-e-mail-through-a-relay.md) (the first start asks for the address of the first person)
- Date: 2026-10-23

## Context

Once `install.sh` has run, the administrator of a new server still has work to do in a terminal: create their account with `cahut user create`, open the activation link it prints, find the screen where an AI connects, and, for the server itself, create an administrator with `cahut admin create`. The script ends with these commands. The [V0 plan](../plan-v0.md) wants instead a first start guided in the browser, opened by a single-use link that the installation prints, and a first publication by an AI within five minutes of it.

[ADR 13](0013-people-sessions-and-rights.md) keeps the administrators of the server apart from the people of each Cahut: two kinds of accounts, two sessions. A person is born from an activation link, where they choose their password; an AI acts on behalf of a person. Someone who installs Cahut for themselves needs both: a person, to have a Room and connect an AI, and an administrator of the server, to declare other Cahuts.

## Decision

- **The installation ends with a link**, `https://<address>/console/setup#<code>`. The code is random, single-use, valid for 24 hours, kept hashed in the database of the server and bound to one Cahut. It works only while that Cahut has no person. `cahut setup` prints a new link, which cancels the previous one; `install.sh` runs it when the Cahut of its domain has nobody.
- **The first start has three steps**, on a full screen of the console:
  1. **You.** A name, a handle and a password. They make the first person of the Cahut, its administrator. While the server has no administrator, they also make one, with the same handle and the same password: two accounts, as ADR 13 has it, created at once. The console then signs in as the person.
  2. **Your Cahut.** Its address. At an IP address ([ADR 26](0026-cahut-at-an-ip-address.md)), how to move it to a domain. The name of the organization, its main color and its logo, for the house style; or the neutral house style. This step and the next can be skipped.
  3. **Your first AI.** The MCP address to copy, and the steps for Claude Code, the Claude apps, ChatGPT and other MCP clients. The screen waits for the AI, and says when it has connected.
- **Three operations of the server.** `create_setup_link`, for its administrators and the command line. `get_setup`, for visitors, says whether a code still works and for which Cahut. `set_up`, for visitors and people only, takes the code, the name, the handle and the password: it creates the person with the operations that exist, the creation of a person and the activation of their account, then creates the administrator of the server when there is none, and returns the session of the person. Both go through the limits of signing in. The house style and the AI then go through the operations of every day.
- **The Logbook** records each step under its own operation: the link created, the person and the administrator created, the account activated. Never the code or the password.
- **Later**, the overview of the console offers to connect an AI, with the same steps, as long as the person has neither a connected AI nor a token. The screen AIs and access shows the same steps.

## Consequences

- A new server goes from the end of `install.sh` to a connected AI without a terminal command: the link holds the proof that the person is the one who installed it.
- The link printed at the end of the installation is worth an account of administrator for 24 hours. It shows in the terminal of the installation and nowhere else; `cahut setup` replaces it.
- The same password guards two accounts. Changing it in one leaves the other as it was: the console says so on the page of the account.
- A server that already has administrators can still give a setup link for a new Cahut: its first person goes through the same steps, without becoming an administrator of the server.
- The steps for each AI client live in the console, which changes with each release; the documentation keeps the longer version.

## Alternatives considered

- **A setup screen open to whoever comes first**, without a link, as many self-hosted applications have it. Between the start of the service and the first visit, anyone who finds the address takes the server.
- **The commands of today, with a better text.** No new code, but a person who does not live in a terminal stops there, and the plan wants the browser.
- **Only a person, with the administrator of the server made later.** One account fewer, but someone alone on their server would have to run `cahut admin create` to declare a second Cahut, which the first start is there to spare them.
- **One account for both roles.** It would undo ADR 13, which keeps the people of each Cahut sealed off from the server.
