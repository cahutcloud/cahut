# CLAUDE.md

Guidance for Claude Code in this repository.

## The project

Cahut is a self-hosted, open source home for what AIs make: documents, decks, dashboards and small apps, published over MCP to the user's own server and shared by link. It is a personal project by Olivier Durand, independent of any company and of any AI vendor.

The V0 is under way, batch by batch, following the V0 plan. Batch 1 is the first vertical slice: an AI publishes a Window over MCP with a personal token, Cahut serves it over HTTPS at `<domain>/<room>/<window>/` in a CSP sandbox, and the call goes to the Logbook. A server already holds several Cahuts, each declared by its domain. Batch 2 ([docs/plan-lot-2.md](docs/plan-lot-2.md)) brought people who sign in, administrators of each Cahut and of the server, team Rooms with their rights, and the web console with its HTTP API, served under `/console/` on port 443 since [ADR 20](docs/adr/0020-console-under-a-path-of-port-443.md). Batch 3 ([docs/plan-lot-3.md](docs/plan-lot-3.md)) brought OAuth for AI clients: each Cahut is its own authorization server, on port 443 for the servers of AI clients, and a person grants an AI some Rooms and a scope on a consent screen in the console ([ADR 16](docs/adr/0016-oauth-for-ai-clients.md), [ADR 17](docs/adr/0017-oauth-endpoints-on-port-443.md)). Batch 4 ([docs/plan-lot-4.md](docs/plan-lot-4.md)) brought two versions per Window with a rollback, the trash, the access of each Window (Room only by default, internal, Key or public, until a date), readers who sign in on port 443 through the console, and the declaration of what a Window loads and calls, held by its CSP ([ADR 19](docs/adr/0019-window-versions-access-and-declaration.md)). Batch 5 ([docs/plan-lot-5.md](docs/plan-lot-5.md)) brought Knock first: an AI that calls an operation that cannot be undone, or that changes what a person may do, gets a Knock, which the person it acts for approves or declines in the console; Windows in the trash can be purged, team Rooms deleted, and the reach of a connected AI changed ([ADR 21](docs/adr/0021-ais-knock-first.md)). Batch 6 ([docs/plan-lot-6.md](docs/plan-lot-6.md)) brought the house style of each Cahut: its name, colors, fonts, logo and guidance for AIs, set in the console or by an AI and undone in one click, served under `/-/` to the Windows that link it and to the pages that readers see, and given to AIs by `get_house_style`, with a template for each kind of Window and a deck engine; and a skill for Claude, packaged as a plugin in a marketplace of this repository and served by each Cahut as a zip ([ADR 22](docs/adr/0022-house-style-of-each-cahut.md)). An organization with guidelines of its own makes the house style inactive for AIs, which then get nothing of it ([ADR 23](docs/adr/0023-house-style-inactive-for-ais.md)). Batch 7 ([docs/plan-lot-7.md](docs/plan-lot-7.md)) brings the installation, the first start and the updates: signed releases that each server installs with a health check and a rollback ([ADR 24](docs/adr/0024-signed-releases-and-updates.md)), a first start guided in the browser from a single-use link ([ADR 25](docs/adr/0025-first-start-in-the-browser.md)), and a Cahut at the IP address of its server until it has a domain ([ADR 26](docs/adr/0026-cahut-at-an-ip-address.md)).

The website (`cahut.cloud`) and the documentation (`docs.cahut.cloud`) are Astro sites in [sites/](sites/README.md), built on the brand guidelines ([ADR 12](docs/adr/0012-website-and-docs-as-astro-sites.md)). The website's `/brand` page gives access to the brand, and its `/project` section is built from the community files at the root of the repository ([ADR 18](docs/adr/0018-project-pages-from-community-files.md)).

Read before working on anything substantial:

- [README.md](README.md): the public pitch, scope, governance model and technical choices.
- [docs/brief-projet.md](docs/brief-projet.md): the full project brief, in French, with the open questions.
- [docs/plan-v0.md](docs/plan-v0.md): the development plan for the V0 (the MVP), in French: scope, batches, milestones and decisions to make.
- [docs/plan-lot-2.md](docs/plan-lot-2.md): the detailed plan of batch 2 (people, team Rooms and the console), in French: operations, screens and pull requests.
- [docs/plan-lot-3.md](docs/plan-lot-3.md): the detailed plan of batch 3 (OAuth for AI clients), in French.
- [docs/plan-lot-4.md](docs/plan-lot-4.md): the detailed plan of batch 4 (versions, access and declaration of Windows), in French.
- [docs/plan-lot-5.md](docs/plan-lot-5.md): the detailed plan of batch 5 (Knock first), in French.
- [docs/plan-lot-6.md](docs/plan-lot-6.md): the detailed plan of batch 6 (house style and skill for Claude), in French.
- [docs/plan-lot-7.md](docs/plan-lot-7.md): the detailed plan of batch 7 (installation, first start and updates), in French.
- [docs/gouvernance.md](docs/gouvernance.md): how the project is governed and how it lives, in French: roles, contributions, channels, releases, security, name, license and money, and the decisions still open.
- [docs/adr/README.md](docs/adr/README.md): the architecture decision records, in English. They settle the web server and TLS, the database, the console, OAuth, the module path and releases, then several Cahuts per server, the addresses of Rooms and Windows, the Window sandbox, the console's port, the website and documentation, and then people, sessions and rights, the console's HTTP API, the license, OAuth for AI clients, the port of its endpoints, the project pages of the website, the versions, access and declaration of Windows, the console under a path of port 443, Knock first, the house style of each Cahut, a house style inactive for AIs, signed releases and updates, the first start in the browser, and a Cahut at an IP address.
- [docs/charte-graphique/README.md](docs/charte-graphique/README.md): the brand guidelines, in French.

The implementation language is **Go** (decided, see the README). The brief still lists Go or Rust as open: the README and the ADRs are the current reference.

## Languages

- Talk with Olivier in French.
- Write in English everything that will be public at launch: README, code, comments, commit messages, pull requests, product copy, interface text, the website and the documentation.
- The website has a French version too, under `/fr/`: its words live in `sites/cahut.cloud/src/i18n/fr.ts`, next to the English ones in `en.ts`. Change both together.
- `docs/` stays in French for now (brief, V0 plan and brand guidelines), except `docs/adr/`, in English because the records are public at launch. Use correct French spelling with all accents.

## Build and test

The Go version is set by the `toolchain` line of `go.mod`; the `go` command downloads that toolchain when the local one is older. Run `make` to list the tasks:

| Task | What it does |
| --- | --- |
| `make build` | Builds the console, then `bin/cahut` for this machine with the console embedded |
| `make dev` | Serves the Cahuts of `tmp/dev` on http://localhost:8080, and their console at http://localhost:8080/console/ |
| `make console` | Builds the console into `internal/console/ui/dist` |
| `make console-dev` | Serves the console with live reload on http://localhost:5173/console/, against the API of `make dev` |
| `make test` | Runs the tests with the race detector |
| `make lint` | Checks that `go.mod` is tidy, then runs golangci-lint |
| `make fmt` | Formats the code (gofmt, goimports) |
| `make check` | Lint and tests, as CI runs them: run it before pushing |
| `make snapshot` | Builds the release archives into `dist/`, without publishing |
| `make release-check` | Validates `.goreleaser.yaml` |
| `make site-dev` | Serves cahut.cloud on http://localhost:4321 |
| `make docs-dev` | Serves docs.cahut.cloud on http://localhost:4322 |
| `make docs-reference` | Writes the reference of the docs from the code: run it after changing an operation or a command |
| `make console-types` | Writes the TypeScript types of the console's API from the code: run it after changing an operation |
| `make sites` | Builds both sites, as CI does: run it after changing `sites/` or the brand |

golangci-lint and GoReleaser run through `go run`, at the versions pinned in the Makefile: nothing needs installing. The sites and the console need Node.js at the version of [.nvmrc](.nvmrc) (`nvm use`); their tasks install the npm packages on first run. `make build`, `make dev` and `make snapshot` build the console, so they need Node.js; `make test`, `make lint` and `make check` do not. The binary embeds the console only with the `console` build tag: a plain `go build` serves a notice in its place ([internal/console](internal/console/console.go)).

To try a Cahut on this machine, run `make dev`, then, in another terminal:

```bash
go run ./cmd/cahut domain add --data tmp/dev localhost
go run ./cmd/cahut user create --data tmp/dev --handle you --name "Your Name" --admin
claude mcp add --transport http cahut-dev http://localhost:8080/mcp
claude mcp login cahut-dev
```

Open the link that `user create` prints to choose a password. `claude mcp login` then sends you to the consent screen of the console, over OAuth; `--no-browser` prints its address instead. A personal token works too: `go run ./cmd/cahut token create --data tmp/dev --user you --name "Claude Code"`, then `claude mcp add` with `--header "Authorization: Bearer <token>"`. When port 8080 is taken, `go run -tags console ./cmd/cahut serve --dev --data tmp/dev --dev-addr localhost:8090` serves elsewhere, OAuth included; the links that the commands print still name port 8080.

Windows open at `http://localhost:8080/<room>/<window>/`, to the members of their Room by default: a reader signs in through the console's `/console/read` page, and the page of each Window in the console sets its access and its Key. An AI that asks for something that knocks first gets the link of its Knock, `http://localhost:8080/console/knocks/<id>`, where its person approves it. The link that `user create` prints opens the console at `http://localhost:8080/console/activate#…`, and its API answers at `http://localhost:8080/console/api/<operation>` ([ADR 14](docs/adr/0014-console-http-api.md)). To try a second Cahut, declare `beta.localhost` too: Chrome, Firefox and curl resolve `*.localhost` to this machine. Once the server holds several Cahuts, the commands that act in one take `--domain`. `go run ./cmd/cahut logbook --data tmp/dev` prints the Logbook of a Cahut.

On a server, [install.sh](install.sh) installs a release, or a binary given with `--binary`, as the `cahut` systemd service, and declares the Cahut of `--domain`; run it with `--help`. Administration commands run as the service user, with the binary's full path, since sudo on AlmaLinux does not search `/usr/local/bin`: `sudo -u cahut /usr/local/bin/cahut token create ...`.

CI ([.github/workflows/ci.yml](.github/workflows/ci.yml)) runs on every pull request and on `main`: lint, a check that the reference of the docs is up to date, tests, and a GoReleaser snapshot build that checks the linux/amd64 and linux/arm64 binaries are static. [sites.yml](.github/workflows/sites.yml) builds the sites on pull requests that touch them or the brand, and on `main` uploads them to the web server with rsync over SSH, once the repository's variables name it ([sites/README.md](sites/README.md#deployment)). Pushing a `v*` tag runs [release.yml](.github/workflows/release.yml), which drafts a GitHub release with the archives, their checksums and the Ed25519 signature of the checksums, made with the `CAHUT_RELEASE_KEY` secret ([ADR 24](docs/adr/0024-signed-releases-and-updates.md)); a person reviews and publishes it. The public keys live in `internal/release/release-keys.pem` and in `install.sh`, which a test keeps the same.

## Code conventions

- Module path `cahut.cloud`. The entry point is `cmd/cahut`; everything else lives under `internal/`.
- Binaries are static, built with `CGO_ENABLED=0`. Never add a dependency that needs cgo.
- Prefer the standard library. A new dependency needs a reason, stated in the pull request.
- Every action goes through the operation registry ([ADR 1](docs/adr/0001-operation-registry.md)). MCP tools, the HTTP API, the command line and SSH are thin adapters: decode, call the registry, encode. No business logic and no storage access in them. An operation is an `operation.Op` registered in `cahut.Open`, or in `server.Open` for the operations of the server as a whole. It states its `Level` (any person, Cahut administrators, server administrators), whether it is `PeopleOnly` (it proves who someone is, such as `sign_in`) and whether it is `KnockFirst` ([ADR 13](docs/adr/0013-people-sessions-and-rights.md)), or knocks for some calls only (`KnocksWhen`, with `Call.Knock` in its `Run`). Every operation of a Cahut that is not `PeopleOnly` becomes an MCP tool, offered to the AIs whose person has its level, with its JSON Schemas generated from the Go input and output types. The registry reads the person's role and state in the transaction of each call: rights on a Room are checked by the operations.
- A Cahut never reads the records of another. Each has its own database, opened from the domain of the request before anything else.
- Errors meant for the caller wrap `operation.ErrInvalid`, `ErrNotFound` or `ErrForbidden`, with a message that tells an AI how to fix its call. Any other error is internal: adapters hide it and the Logbook keeps it.
- Never put a token, a password or file contents in a Logbook entry (`Call.Log`).
- A call of an AI that knocks first runs to its end to be checked, then is rolled back, and makes a Knock that its person approves in the console ([ADR 21](docs/adr/0021-ais-knock-first.md)). An operation that knocks keeps its effects in its transaction and touches the disk only after the commit (`Call.AfterCommit`), takes no secret in its input, and gives its output to a person only (`Call.Approved`).
- Schema changes are new numbered files in `internal/store/migrations/cahut/` or `internal/store/migrations/server/`. Never edit a migration that has been released. A migration that makes again a table that others refer to starts with the line `-- foreign keys off`, as SQLite asks ([store.go](internal/store/store.go)).
- A new architecture decision is a new record in `docs/adr/`. Never rewrite an accepted record: supersede it.
- The classes, the variables, the markup of the deck engine and its keys in `internal/housestyle/kit/` are a contract with the Windows already published: add to them, refine how they look, and never rename or remove one before a major version ([ADR 22](docs/adr/0022-house-style-of-each-cahut.md)). A test checks that the guide for AIs names every class of the sheet.
- Never name a file `claude.md`, in any case: macOS ignores case, and Claude Code reads such a file as instructions, like `CLAUDE.md`.
- The sites and the console follow the same rule for npm packages: a reason in the pull request, an exact version, and no install scripts ([sites/.npmrc](sites/.npmrc), [internal/console/ui/.npmrc](internal/console/ui/.npmrc)). Add them with `npm install <package> --workspace <site>` in `sites/`, or with `npm install <package>` in `internal/console/ui/`, never in another folder.
- Versions to keep in step: golangci-lint and GoReleaser are pinned both in the Makefile and in the workflows. Actions are pinned by commit hash, with the version in a comment; Dependabot updates them, and the npm packages of the sites too. Raise the `toolchain` line of `go.mod` when Go ships a patch release. `.nvmrc` holds the Node.js LTS line for the whole repository.

## Product rules that shape the code

These are decided and should hold in every design and every line of code:

- **Bounded operations.** AIs act through typed operations (publish a page, create a database, share a link), over MCP or forced-command SSH. Never a free shell.
- **Knock first.** What can be undone, the AI does alone. What cannot (permanent deletion, making a Window public, adding a person, changing permissions, erasing backups) waits for a human to approve it in the console, with a notification and an expiry.
- **Everything is logged.** Every action records which AI, on whose behalf, what it did and with what result.
- **One API.** The web console uses the same API as the AIs: no capability is reserved for it.
- **No AI on the server.** The AIs stay outside and connect over MCP.
- **One server, one or more Cahuts**, on a standard AlmaLinux or Ubuntu. Each Cahut has its own domain, people and database, and shares nothing else with the others ([ADR 8](docs/adr/0008-several-cahuts-per-server.md)). Data stays on that server; backups go only to the external S3 or SFTP storage the admin chooses.
- **Isolation.** Every file of a Window is served in a CSP sandbox ([ADR 10](docs/adr/0010-windows-in-a-csp-sandbox.md)): its scripts cannot read other Windows, the console or the reader's session. Each app runs in an unprivileged container with resource limits.

## Vocabulary

Product text speaks the language of the house, in plain English. Use these words consistently in the interface, the docs and the site:

| Word | Meaning |
| --- | --- |
| Cahut | An instance on the user's server, at its own domain; one server can hold several |
| Room | A personal or team space, with its members, at `<domain>/<room>` |
| Window | A living page: document, deck, dashboard or app, at `<domain>/<room>/<window>` |
| Key | The password that opens a protected Window |
| House style | The organization's brand, applied by every AI |
| Knock first | The rule that the AI asks before any irreversible action |
| Logbook | The record of every action |

Each Window opens to its Room only (the default), to everyone signed in to its Cahut (internal), to anyone with its Key, or to the public. The word Door is gone: speak of a Window's access.

Cahut, Room, Window and Key take a capital in running text ("3 Rooms in your Cahut", "This Key no longer opens Q3 review."). An instance is a Cahut, with a capital, like the product. Never write "CAHUT"; lowercase "cahut" exists only in the logo and in commands.

## Brand

For anything visual (site, console, mockup, slide, social image, e-mail, illustration) or any product copy, use the project skill **`charte-graphique-cahut`** in [.claude/skills/charte-graphique-cahut/](.claude/skills/charte-graphique-cahut/SKILL.md). The short version:

- Take colors, type, spacing, radii and shadows from [docs/charte-graphique/tokens/cahut-tokens.css](docs/charte-graphique/tokens/cahut-tokens.css). Never retype hex values.
- Use the logo SVG files as they are. Never redraw, recolor or rebuild the logotype.
- Interface components come from [docs/charte-graphique/ui/cahut-ui.css](docs/charte-graphique/ui/cahut-ui.css) (`c-` classes).
- Pages that Cahut serves for an organization carry that organization's house style, not Cahut's. Cahut's own brand applies to its own screens: console, sign-in, e-mails, site.
- The sites and the console import brand files through the `@charte` alias (`@charte/tokens/cahut-tokens.css`), which points to `docs/charte-graphique`. Never copy a token, a font or a logo into `sites/` or `internal/console/ui/`.

Changing the logo, the palette, the tokens or the vocabulary is a change to the brand guidelines: update the guidelines README, `index.html`, both token files (CSS and JSON, kept in sync by hand) when tokens change, the PNG renders, the PDF (`node .claude/skills/charte-graphique-cahut/scripts/imprimer-charte.mjs`), and bump the version in the changelog table.

## Repository layout

```
cmd/cahut/             Entry point of the cahut binary
internal/
  operation/           The registry of operations, the only way in (ADR 1)
  logbook/             The record of every action
  store/               SQLite database and its migrations
  credential/          Passwords (Argon2id), random tokens and codes, and their hashes
  account/             People, roles, activation links, sessions, sign-in and personal tokens
  room/                Rooms: personal and team Rooms, their owners and members, rights in a Room
  oauth/               The OAuth authorization server of each Cahut: clients, consent, tokens, connected AIs (ADR 16)
  knock/               Knocks: the calls of AIs that wait for their person, who approves or denies them (ADR 21)
  housestyle/          The house style of each Cahut, its sheet and logos under /-/, and its operations (ADR 22)
    kit/               The classes, the deck engine, the templates and the guide for AIs, under Apache-2.0
  window/              Windows: publish, describe, list, files on disk
  cahut/               Assembles a Cahut: its database, registry and operations
  server/              The server as a whole: its Cahuts and their domains, its administrators
  web/                 HTTP(S): MCP endpoint, Windows by domain and path, the console's API, certificates
  cli/                 Command line: serve and the administration commands
  reference/           Writes the reference of the docs from the code (make docs-reference)
  apitypes/            Writes the TypeScript types of the console's API from the code (make console-types)
  console/             The web console, embedded with the console build tag (ADR 5)
    ui/                Its Preact app, built by Vite; src/api.gen.ts comes from make console-types
  version/             Build information, printed by cahut version
  release/             The keys that sign releases, and the check of a release's signature and checksums (ADR 24)
  releasesign/         Signs the checksum file of a release, run by GoReleaser
docs/
  adr/                 Architecture decision records (English)
  brief-projet.md      Project brief (French)
  plan-v0.md           V0 development plan (French)
  plan-lot-2.md        Detailed plan of batch 2 (French)
  plan-lot-3.md        Detailed plan of batch 3 (French)
  plan-lot-4.md        Detailed plan of batch 4 (French)
  plan-lot-5.md        Detailed plan of batch 5 (French)
  plan-lot-6.md        Detailed plan of batch 6 (French)
  plan-lot-7.md        Detailed plan of batch 7 (French)
  gouvernance.md       Governance and project life (French)
  charte-graphique/    Brand guidelines with all source files (French)
plugins/               Plugins for AI clients, under Apache-2.0; plugins.go embeds the skill that each Cahut serves (ADR 22)
  cahut/               The Cahut plugin for Claude: the skill, and the MCP server of a Cahut
.claude-plugin/        marketplace.json: this repository as a marketplace of Claude Code plugins
sites/                 npm workspace of the sites (ADR 12)
  cahut.cloud/         The website, Astro, with the brand at /brand
  docs.cahut.cloud/    The documentation, Astro with Starlight
  brand/               @cahut/brand: the brand of docs/charte-graphique for the sites
.github/
  workflows/           CI, sites and release
  ISSUE_TEMPLATE/      The bug report and accepted feature forms; config.yml sends questions to the Discussions
  PULL_REQUEST_TEMPLATE.md   What a pull request says
  CODEOWNERS           Who reviews each path: the lead maintainer, for now
  dco.yml              Settings of the DCO app, which checks the sign-offs
  rulesets/            The protection of main and of release tags, to apply with the API
  FUNDING.yml          The Sponsor button
  dependabot.yml       Weekly updates of Go modules, npm packages and actions
.claude/skills/
  charte-graphique-cahut/   Project skill for the brand
GOVERNANCE.md          How the project is run: the public summary of docs/gouvernance.md
CONTRIBUTING.md        What helps, how a change gets in, the DCO and the rule on AI-assisted work
CODE_OF_CONDUCT.md     The Contributor Covenant 3.0
SECURITY.md            How to report a vulnerability
SUPPORT.md             Where to ask for help
TRADEMARK.md           Using the name and the logo
LICENSE                The AGPL-3.0 (ADR 15)
install.sh             Installs Cahut on a server as a systemd service
Makefile               Development tasks
.nvmrc                 Node.js version, for the sites
.golangci.yml          Linter configuration
.goreleaser.yaml       Release build
```

## Writing style

The README sets the tone: short sentences, concrete verbs, no hype. No exclamation marks, no capitals for emphasis, no promises of AI magic: Cahut hosts the work, it does not create it.

## Git

- `main` is the default branch. Work on a topic branch (`claude/<topic>` for Claude sessions) and open a pull request. CI must be green before merging. Pull requests are merged by squash, the only method the repository allows.
- Sign off every commit with `git commit -s`, as [CONTRIBUTING.md](CONTRIBUTING.md#sign-your-commits) asks: the DCO app checks it on every pull request. Olivier asks Claude sessions to sign off in his name, with the git identity of this checkout; he answers for what he merges.
- Commit messages in English: a short imperative subject line, then a body that says what changed and why.
- The repository is private until launch and will be opened as is, history included. Keep it free of references to other projects, employers or clients, and never commit secrets.
