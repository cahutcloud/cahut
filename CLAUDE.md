# CLAUDE.md

Guidance for Claude Code in this repository.

## The project

Cahut is a self-hosted, open source home for what AIs make: documents, decks, dashboards and small apps, published over MCP to the user's own server and shared by link. It is a personal project by Olivier Durand, independent of any company and of any AI vendor.

The V0 is under way, batch by batch, following the V0 plan. Batch 1 is the first vertical slice: an AI publishes a Window over MCP with a personal token, Cahut serves it over HTTPS at `<domain>/<room>/<window>/` in a CSP sandbox, and the call goes to the Logbook. A server already holds several Cahuts, each declared by its domain. There are no sign-in, console or access levels yet: anyone with a Window's address can read it.

The website (`cahut.cloud`) and the documentation (`docs.cahut.cloud`) are Astro sites in [sites/](sites/README.md), built on the brand guidelines ([ADR 12](docs/adr/0012-website-and-docs-as-astro-sites.md)). The website's `/brand` page gives access to the brand.

Read before working on anything substantial:

- [README.md](README.md): the public pitch, scope, governance model and technical choices.
- [docs/brief-projet.md](docs/brief-projet.md): the full project brief, in French, with the open questions.
- [docs/plan-v0.md](docs/plan-v0.md): the development plan for the V0 (the MVP), in French: scope, batches, milestones and decisions to make.
- [docs/adr/README.md](docs/adr/README.md): the architecture decision records, in English. They settle the web server and TLS, the database, the console, OAuth, the module path and releases, then several Cahuts per server, the addresses of Rooms and Windows, the Window sandbox, the console's port, and the website and documentation.
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
| `make build` | Builds `bin/cahut` for this machine |
| `make dev` | Serves the Cahuts of `tmp/dev` on http://localhost:8080 |
| `make test` | Runs the tests with the race detector |
| `make lint` | Checks that `go.mod` is tidy, then runs golangci-lint |
| `make fmt` | Formats the code (gofmt, goimports) |
| `make check` | Lint and tests, as CI runs them: run it before pushing |
| `make snapshot` | Builds the release archives into `dist/`, without publishing |
| `make release-check` | Validates `.goreleaser.yaml` |
| `make site-dev` | Serves cahut.cloud on http://localhost:4321 |
| `make docs-dev` | Serves docs.cahut.cloud on http://localhost:4322 |
| `make docs-reference` | Writes the reference of the docs from the code: run it after changing an operation or a command |
| `make sites` | Builds both sites, as CI does: run it after changing `sites/` or the brand |

golangci-lint and GoReleaser run through `go run`, at the versions pinned in the Makefile: nothing needs installing. The sites need Node.js at the version of [.nvmrc](.nvmrc) (`nvm use`); their tasks install the npm packages on first run. The Go tasks never need Node.js.

To try a Cahut on this machine, run `make dev`, then, in another terminal:

```bash
go run ./cmd/cahut domain add --data tmp/dev localhost
go run ./cmd/cahut user create --data tmp/dev --handle you --name "Your Name" --admin
go run ./cmd/cahut token create --data tmp/dev --user you --name "Claude Code"
claude mcp add --transport http cahut-dev http://localhost:8080/mcp --header "Authorization: Bearer <token>"
```

Windows open at `http://localhost:8080/<room>/<window>/`. To try a second Cahut, declare `beta.localhost` too: Chrome, Firefox and curl resolve `*.localhost` to this machine. Once the server holds several Cahuts, the commands that act in one take `--domain`. `go run ./cmd/cahut logbook --data tmp/dev` prints the Logbook of a Cahut.

On a server, [install.sh](install.sh) installs a release, or a binary given with `--binary`, as the `cahut` systemd service, and declares the Cahut of `--domain`; run it with `--help`. Administration commands run as the service user, with the binary's full path, since sudo on AlmaLinux does not search `/usr/local/bin`: `sudo -u cahut /usr/local/bin/cahut token create ...`.

CI ([.github/workflows/ci.yml](.github/workflows/ci.yml)) runs on every pull request and on `main`: lint, a check that the reference of the docs is up to date, tests, and a GoReleaser snapshot build that checks the linux/amd64 and linux/arm64 binaries are static. [sites.yml](.github/workflows/sites.yml) builds the sites on pull requests that touch them or the brand, and on `main` uploads them to the web server with rsync over SSH, once the repository's variables name it ([sites/README.md](sites/README.md#deployment)). Pushing a `v*` tag runs [release.yml](.github/workflows/release.yml), which drafts a GitHub release with the archives and checksums; a person reviews and publishes it.

## Code conventions

- Module path `cahut.cloud`. The entry point is `cmd/cahut`; everything else lives under `internal/`.
- Binaries are static, built with `CGO_ENABLED=0`. Never add a dependency that needs cgo.
- Prefer the standard library. A new dependency needs a reason, stated in the pull request.
- Every action goes through the operation registry ([ADR 1](docs/adr/0001-operation-registry.md)). MCP tools, the HTTP API, the command line and SSH are thin adapters: decode, call the registry, encode. No business logic and no storage access in them. An operation is an `operation.Op` registered in `cahut.Open`, or in `server.Open` for the operations of the server as a whole; every operation of a Cahut that is not `AdminOnly` becomes an MCP tool, with its JSON Schemas generated from the Go input and output types.
- A Cahut never reads the records of another. Each has its own database, opened from the domain of the request before anything else.
- Errors meant for the caller wrap `operation.ErrInvalid`, `ErrNotFound` or `ErrForbidden`, with a message that tells an AI how to fix its call. Any other error is internal: adapters hide it and the Logbook keeps it.
- Never put a token, a password or file contents in a Logbook entry (`Call.Log`).
- Schema changes are new numbered files in `internal/store/migrations/cahut/` or `internal/store/migrations/server/`. Never edit a migration that has been released.
- A new architecture decision is a new record in `docs/adr/`. Never rewrite an accepted record: supersede it.
- The sites follow the same rule for npm packages: a reason in the pull request, an exact version, and no install scripts ([sites/.npmrc](sites/.npmrc)). Add them with `npm install <package> --workspace <site>` in `sites/`, never in another folder.
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
- The sites import brand files through the `@charte` alias (`@charte/tokens/cahut-tokens.css`), which points to `docs/charte-graphique`. Never copy a token, a font or a logo into `sites/`.

Changing the logo, the palette, the tokens or the vocabulary is a change to the brand guidelines: update the guidelines README, `index.html`, both token files (CSS and JSON, kept in sync by hand) when tokens change, the PNG renders, the PDF (`node .claude/skills/charte-graphique-cahut/scripts/imprimer-charte.mjs`), and bump the version in the changelog table.

## Repository layout

```
cmd/cahut/             Entry point of the cahut binary
internal/
  operation/           The registry of operations, the only way in (ADR 1)
  logbook/             The record of every action
  store/               SQLite database and its migrations
  account/             People, personal Rooms and personal tokens
  window/              Windows: publish, describe, list, files on disk
  cahut/               Assembles a Cahut: its database, registry and operations
  server/              The server as a whole: its Cahuts and their domains
  web/                 HTTP(S): MCP endpoint, Windows by domain and path, certificates
  cli/                 Command line: serve and the administration commands
  reference/           Writes the reference of the docs from the code (make docs-reference)
  version/             Build information, printed by cahut version
docs/
  adr/                 Architecture decision records (English)
  brief-projet.md      Project brief (French)
  plan-v0.md           V0 development plan (French)
  charte-graphique/    Brand guidelines with all source files (French)
sites/                 npm workspace of the sites (ADR 12)
  cahut.cloud/         The website, Astro, with the brand at /brand
  docs.cahut.cloud/    The documentation, Astro with Starlight
  brand/               @cahut/brand: the brand of docs/charte-graphique for the sites
.github/
  workflows/           CI, sites and release
  dependabot.yml       Weekly updates of Go modules, npm packages and actions
.claude/skills/
  charte-graphique-cahut/   Project skill for the brand
install.sh             Installs Cahut on a server as a systemd service
Makefile               Development tasks
.nvmrc                 Node.js version, for the sites
.golangci.yml          Linter configuration
.goreleaser.yaml       Release build
```

## Writing style

The README sets the tone: short sentences, concrete verbs, no hype. No exclamation marks, no capitals for emphasis, no promises of AI magic: Cahut hosts the work, it does not create it.

## Git

- `main` is the default branch. Work on a topic branch (`claude/<topic>` for Claude sessions) and open a pull request. CI must be green before merging.
- Commit messages in English: a short imperative subject line, then a body that says what changed and why.
- The repository is private until launch and will be opened as is, history included. Keep it free of references to other projects, employers or clients, and never commit secrets.
