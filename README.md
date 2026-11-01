<div align="center">

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="docs/charte-graphique/logos/cahut-logo-horizontal-inverse.svg">
  <img alt="Cahut" src="docs/charte-graphique/logos/cahut-logo-horizontal.svg" width="300">
</picture>

<h3>A home to publish and share what your AIs make.</h3>

Documents, slide decks, dashboards and small apps, as pages at your address<br>
on your own server, where your team works on them, whichever AI made them.

<p>
  <a href="https://github.com/cahutcloud/cahut/releases/latest"><img alt="Latest release" src="https://img.shields.io/github/v/release/cahutcloud/cahut?label=release&color=B9E1FB&labelColor=1F2A33"></a>
  <a href="https://github.com/cahutcloud/cahut/actions/workflows/ci.yml"><img alt="CI" src="https://github.com/cahutcloud/cahut/actions/workflows/ci.yml/badge.svg?branch=main"></a>
  <a href="LICENSE"><img alt="License AGPL-3.0" src="https://img.shields.io/badge/license-AGPL--3.0-D9DCDF?labelColor=1F2A33"></a>
  <a href="go.mod"><img alt="Go 1.27" src="https://img.shields.io/badge/Go-1.27-D9DCDF?logo=go&logoColor=F7F5F0&labelColor=1F2A33"></a>
  <a href="https://www.bestpractices.dev/projects/15845"><img alt="OpenSSF Best Practices" src="https://www.bestpractices.dev/projects/15845/badge"></a>
</p>

**[Website](https://cahut.cloud)** · **[Documentation](https://docs.cahut.cloud)** · **[Brand](https://cahut.cloud/brand)** · **[Contributing](CONTRIBUTING.md)** · **[Security](SECURITY.md)**

<br>

<img alt="You ask Claude for a deck, and it publishes the deck on your Cahut, at its own address. Making it public waits in the console of the Cahut, where you approve it. The deck is public." src=".github/assets/cahut-demo.gif" width="800">

</div>

<br>

Cahut installs with one command on a Linux server. Claude, ChatGPT, Mistral Vibe or any MCP client publish to it, and you share a link instead of sending a file. Your team reads it and asks the AI for the next version, in Rooms where people and their AIs work on the same Windows. The AI does on its own what can be undone, a human approves the rest, and every action is logged.

## Try it

- **See what AIs published, without an account.** [demo.cahut.cloud/examples](https://demo.cahut.cloud/examples/) holds Windows that AIs made and published, each with the prompt that made it: a small shop, a deck the team comments on, a CV, a team report, a travel journal.
- **Without a server, in two minutes.** [Sign up for My Cahut](https://account.cahut.cloud/en/signup), a free Cahut at `my.cahut.cloud` with limited seats, and connect Claude or Mistral Vibe, even on their free plans, or ChatGPT on a Business, Enterprise or Edu plan. [All about My Cahut](https://docs.cahut.cloud/start/my-cahut/).
- **On your own server, in one command.** On a fresh Ubuntu or AlmaLinux server, with ports 80 and 443 open and a domain that points to it:

  ```bash
  curl -fsSL https://cahut.cloud/install.sh | sudo sh -s -- --domain cahut.example.com
  ```

  [Installation](#installation) says what the script does and how to connect your AI.

If Cahut is useful to you, a star helps others find it.

> [!WARNING]
> Cahut is young: its first release came out on 21 October 2026. What [Features](#features) marks as available works in the latest release; apps, comments, export, single sign-on and backups come next. There are no backups yet, so keep snapshots of your server.

## Why Cahut

Today, what your AIs make stays at the AI's place. A Claude link, a ChatGPT link, a Gamma deck: each lives on the service of the AI that made it. To publish it, share it or work on it with others, you turn it back into a file.

The file is going away. Word, PowerPoint and Excel were designed for humans who write, lay out and calculate by hand. When an AI makes the content, a page at its own address does better: it updates, takes comments, keeps its versions, and exports to PDF or PowerPoint for whoever still needs a file.

Publish, share, work together: today, all three are hard.

- **Scattered**: a Claude link here, a ChatGPT link there, a Gamma deck somewhere else. Each AI keeps what it makes on its own service, so nothing is published in one place.
- **Converted back**: to share it, the result is turned back into a PowerPoint or a Word file, then attached and forwarded.
- **Uncontrolled**: who has access, until when, which version is the right one, where the data lives. Nobody really knows.
- **Alone**: each person works with their own AI. The team has no common place where its work lives and where it works on it together.
- **Unguarded**: handing a server to an AI means giving it a root shell, or going through a platform built for humans that it can barely use.

<p align="center">
  <img alt="Documents, dashboards, code and images join the cloud house of Cahut" src="docs/charte-graphique/illustrations/maison-ouverte.svg" width="560">
</p>

**Your AIs make it and publish it. You share it and work on it together.** Everything they produce lives in one place, as pages at your address, on your server, whatever the AI.

Cahut does not create content, it hosts it. It does not compete with AIs or creation tools: it is the neutral, self-hosted home where their work lives, and where your team works on it.

| Layer | Who handles it | What it brings |
| --- | --- | --- |
| Create | The AIs (Claude, ChatGPT, Mistral, Copilot...) | The content: documents, decks, dashboards, apps |
| Publish, share, work together | Cahut, on your server | A page at your address, a link opened to the right people, Rooms where the team and its AIs work, versions, comments, search, export, Logbook |

## How it works

The AIs stay outside. They connect to Cahut over MCP, drop their work through bounded operations, then disconnect. No AI model runs on the server.

```mermaid
flowchart LR
  subgraph AI["Create: the AIs"]
    C[Claude]
    G[ChatGPT]
    M[Mistral Vibe]
    P[Any MCP client]
  end
  subgraph CAHUT["Cahut, on your server"]
    H["Your Cahut: publish, share, access, versions, search, data, Logbook, irreversible actions approved by you"]
  end
  subgraph READ["Read"]
    E[Your team]
    CL[Your clients]
    PU[The public]
  end
  AI -- MCP --> CAHUT
  CAHUT -- a link --> READ
```

1. **Connect.** Add Cahut to your AI (MCP connector and skill) and authorize it once with OAuth, with the scope you choose.
2. **Create.** Ask your AI for a document, a deck, a dashboard or an app, in the style you want. The skill for Claude gives it the formats.
3. **Publish.** The AI publishes a Window to a Room, declares what it does, and gets its address back, such as `cahut.example.com/marketing/q3-review`. The previous version is kept.
4. **Share.** You, or the AI within the limits you set, choose who opens each Window: its Room only, everyone in your Cahut, anyone with its Key, or the public. Anything irreversible waits for approval in the console.
5. **Live.** Readers open the page in a browser, and a poll or a form keeps what they write in the database of its Window. The AI updates it on request. Everything goes to the Logbook.

## Features

What is marked available works in the latest release. The rest comes next, in about the order of the [roadmap](#roadmap).

| Area | What it does | Status |
| --- | --- | --- |
| Installation | One command on a fresh VPS or dedicated server (AlmaLinux or Ubuntu), a guided first start in the browser, and `cahut doctor` to check the server | Available |
| AI access | A native MCP server over HTTP, with OAuth or a personal token, for claude.ai, Claude Desktop, Claude Code, ChatGPT, Mistral Vibe and any MCP client | Available |
| Skill for Claude | A skill that teaches Claude to make Windows and publish them, packaged as a plugin for Claude Code | Available |
| Pages | Documents, decks, dashboards and static sites, each at its own address in its Room, in a sandbox that holds it to what its AI declares it loads and calls | Available |
| Sharing | Window by Window: its Room only, internal (everyone in the Cahut), a Key (a password) or public, with an optional end date; share links that open a Window without an account, until they end | Available |
| Invitations | A Window sent by e-mail, to people of the Cahut or outside it, with a personal link in place of its Key | Available |
| Versions | Two versions kept per Window, a rollback in one gesture, and a trash | Available |
| Search | Find any page, deck or dashboard of your Rooms by a word of it or a tag | Available |
| Data | A SQLite database for each Window, on your server: polls, forms, petitions and trackers keep what their readers write | Available |
| Users and Rooms | The people of each Cahut and its admins, a personal Room for each person, team Rooms with owners and members | Available |
| Sign-in | Accounts of each Cahut, with a link by e-mail for a forgotten password | Available |
| Knock first | What an AI cannot undo waits for its person, who approves or declines it in the console | Available |
| Web console | In English and French: Knocks, Rooms, Windows, people, connected AIs and the Logbook under `/console/` of each Cahut, and a console for the administrators of the server under `/server/` | Available |
| Domains and TLS | Several Cahuts per server, each at its own domain or subdomain, no wildcard needed, automatic certificates | Available |
| Limits | People, Rooms, Windows and disk for each Cahut, set by the administrators of the server, and Cahuts where each person works alone | Available |
| Moderation | Reports of Windows from anyone, without an account; the administrators of the Cahut and of its server withdraw a Window with a reason, and reinstate it | Available |
| Logbook | History of every action: who, which AI, on whose behalf, with what result | Available |
| Updates | Daily, from the signed releases on GitHub, with a snapshot of the databases and a rollback | Available |
| Comments | Reader feedback directly on the pages | Coming |
| Single sign-on | OpenID Connect: Google, Microsoft Entra, Okta... | Coming |
| Backups | Optional, to external S3 or SFTP storage, restore from the console | Coming |
| Apps | Predefined runtimes (Node, Python...) and an advanced Dockerfile mode, variables and secrets, PostgreSQL or SQLite databases, runtime logs, resource limits | Coming |
| Export | PDF and PowerPoint, for recipients still on Office | Coming |
| SSH access | The same operations over forced-command SSH, for terminal AIs | Coming |
| More AI clients | Copilot and others, with a skill packaged for each client that allows it | Coming |

What Cahut does **not** do:

- Create content: that is the AIs' job.
- Run an AI model on the server: the AIs stay outside and connect to it.
- Depend on one AI vendor, or lock work into a proprietary format.
- Give an AI free access to the server.
- Run across several machines: one server holds one or more Cahuts.

## Governance: knock first

The AI is the intended operator of Cahut, but never without guardrails.

1. **Bounded operations.** The AI acts through typed operations (publish a page, create a database, share a link), never through a free shell, over MCP as over SSH.
2. **Reversible or approved.** What can be undone, the AI does alone. What cannot waits for a human to approve it in the console, with a notification and an expiry.
3. **The AI declares what it publishes.** Each publication states what it contains and what it does (an app that collects data, calls out, opens a form).
4. **Per-agent permissions.** Each AI acts on behalf of a person, with the scope that person grants, revocable at any time.
5. **Everything is logged.** Which AI, on whose behalf, which action, which result: the Logbook is readable in the console.
6. **Data stays home.** Everything lives on your server. Backups only go to the external storage you choose.
7. **Isolation.** Each Window runs in a sandbox: its scripts cannot read the other Windows, the console or the reader's session. Each app runs in an unprivileged container with its own resource limits. The Cahuts of one server share nothing but the machine.

| The AI does it alone (reversible) | A human approves (irreversible) |
| --- | --- |
| Publish or update a Window (the previous version is kept) | Permanently delete a Window, an app or a database |
| Restrict access, extend a lifetime | Make a Window public |
| Deploy an app within the set limits | Add a person to a Cahut |
| Roll back to the previous version | Change the permissions of a user or an AI |
| Run a backup | Erase backups or change their destination |

Apps and backups are coming: their rows say the rule they will follow.

## FAQ

### Why not share a Claude artifact or a ChatGPT link?

Each lives on the service of the AI that made it, and is shared on that service's terms. Cahut keeps the work of every AI in one place, on your server, at your address. Each Window opens to the people you choose, keeps its previous version and can keep what its readers write, and every action goes to the Logbook. The two go together: ask Claude to publish its artifact to your Cahut.

### Why not Gamma?

Gamma makes decks and documents with its own AI and hosts them on its service. Cahut makes nothing: it hosts what your own AIs make, a deck or anything else they write as a web page, on your server.

### Why not Netlify, Vercel or GitHub Pages?

They deploy sites for developers, from a repository or a command line. In Cahut, the AI publishes over MCP, with no Git, deploy key or build step. Each page opens to the people you choose, by Room, Key or share link, and what cannot be undone waits for a person.

### Why not Coolify, Dokku or another self-hosted PaaS?

They run the apps that developers deploy, from Git and in containers. Cahut hosts the pages that AIs publish themselves, through typed operations that never give them a shell, and lets people share each one by link. It does not run apps yet: they are on the [roadmap](#roadmap).

### Why not Nextcloud?

Nextcloud stores, syncs and shares files, and edits office documents. Cahut serves pages, not files: each Window opens in a browser at its own address, runs its scripts in a sandbox, and can keep what its readers write in a database of its own.

### What can a Window do, and what can it not?

A Window is HTML and its files: scripts, styles, images, fonts and data. No code of a Window runs on the server, and its pages load and call only the origins its AI declared. [What Cahut can and cannot do](https://docs.cahut.cloud/start/can-and-cannot/) gives the detail and the limits.

## Vocabulary

Cahut speaks the language of the house, in plain English.

| Word | What it means |
| --- | --- |
| Cahut | An instance on your server, at its own domain. One server can hold several |
| Room | A personal or team space, with its members, at `<domain>/<room>` |
| Window | A living page: document, deck, dashboard or app, at `<domain>/<room>/<window>` |
| Key | The password that opens a protected Window |
| Knock first | The governance rule: the AI knocks before any irreversible action |
| Logbook | The record of every action |

Each Window opens to its Room only, to everyone in the Cahut, to anyone with its Key, or to the public.

## Requirements

- One server, a VPS or a dedicated machine, for one or more Cahuts.
- Minimum 2 vCPU, 4 GB of RAM, 40 GB of disk.
- AlmaLinux or Ubuntu, from the provider's standard image.
- One domain or subdomain per Cahut, pointing to the server; until it has one, a Cahut can live at the server's public IPv4 address. Ports 80 and 443 open.

## Installation

On a fresh AlmaLinux or Ubuntu server, with ports 80 and 443 open and a domain that points to it, run:

```bash
curl -fsSL https://cahut.cloud/install.sh | sudo sh -s -- --domain cahut.example.com
```

The script checks the signature of the release and installs Cahut as a systemd service. It opens the firewall, sets up the daily updates and runs `cahut doctor`. It ends with a single-use link: open it to finish in the browser, where you choose your account and connect your first AI. Without `--domain`, Cahut takes the server's public IPv4 address, with a certificate for that address. You can move it to a domain later, from the console.

Then give your AI the address of the MCP server, `https://cahut.example.com/mcp`. It sends you to the console to sign in and choose what it may do:

| AI client | Where |
| --- | --- |
| claude.ai, Claude Desktop | Settings, Connectors, add a custom connector, on every plan |
| Claude Code | `/plugin marketplace add cahutcloud/cahut`, then `/plugin install cahut@cahut`, which also brings the skill for Claude |
| Mistral Vibe, formerly Le Chat | Connectors, add a custom MCP connector, on every plan |
| ChatGPT | On the web, Plugins, Add custom MCP server, with OAuth, on the Business, Enterprise and Edu plans only: not on Free, Plus or Pro |
| Others | Any MCP client over HTTP, with OAuth or a personal token |

Ask it for a deck: it publishes a Window and gives you its address. You choose who opens it in the console.

Cahut updates itself every day from the signed releases, patch releases by default. It snapshots the databases first, and goes back on its own if the new version does not start. `sudo /usr/local/bin/cahut doctor` says what is wrong with the server, and `sudo /usr/local/bin/cahut update` updates at once. The [documentation](https://docs.cahut.cloud/install/server/) covers the rest: the options of the installer, a Cahut without a domain, several Cahuts on one server, and updates.

## How Cahut is built

Cahut is written by one person, with a lot of help from Claude Code. These rules hold for every change, whoever or whatever wrote it:

- Each change is a pull request, merged once CI is green. Olivier answers for every change that gets in, and each commit carries his sign-off.
- CI runs golangci-lint, govulncheck and the tests with the race detector, replays the acceptance of the first version on a test server, and installs Cahut on a fresh Ubuntu machine, then updates it and rolls it back.
- Each technical choice is an [architecture decision record](docs/adr/README.md), with its reasons and the options set aside, and each batch of work has a written plan.
- The first version had a [security review](docs/revue-securite-v0.md) in five areas (in French): what was found sound, each finding and what was done about it.
- Releases are signed with Ed25519, and each server checks the signature before it installs one.

Contributions made with an AI are welcome under the same rules: [Working with an AI](CONTRIBUTING.md#working-with-an-ai) says them.

## Development

Cahut is written in **Go**: a new, clean codebase built for Cahut, which one person must be able to carry end to end. Go gives a single static binary that is simple to install and update on a standard Linux, and a standard library that covers HTTP, TLS and cryptography.

### Technical choices

| Topic | Choice | Status |
| --- | --- | --- |
| Nature | A platform on a standard Linux, not an operating system | Decided |
| Hosting | One server holds one or more Cahuts, each with its own domain, people and data | Decided |
| Language | Go | Decided |
| AI interface | MCP over HTTP with OAuth, forced-command SSH for the same operations | Decided |
| Web console | A Preact app embedded in the binary, served under `/console/` of each Cahut, and under `/server/` for the administrators of the server, using the same API as the AIs: no capability is reserved for it | Decided |
| Web and TLS | Built into the binary, automatic certificates with CertMagic, Windows at paths of their Cahut's domain, each in a CSP sandbox | Decided |
| Database | SQLite in pure Go, so the binary stays static, one database per Cahut, and one per Window for its data, run in a confined process | Decided |
| Authorization | Cahut's own minimal OAuth 2.1 server, built for MCP clients | Decided |
| Updates | Automatic, latest version checked on GitHub, rollback possible | Decided |
| Backups | Optional, to external S3 or SFTP | Decided |
| Website and docs | Astro sites in this repository, built on the brand guidelines, served from the project's own server | Decided |
| Apps | Predefined runtimes and Dockerfile mode, on an unprivileged container engine (Podman or Docker) | Open |
| License | AGPL-3.0 for the repository; the name and the logo stay apart | Decided |

The reasons behind each choice are in the [architecture decision records](docs/adr/README.md).

### Repository layout

```
cmd/cahut/   Entry point of the cahut binary
internal/    The code of Cahut, private to this module
install.sh   Installs Cahut on a server as a systemd service
docs/        Project documents: architecture decision records in English; the brief, the V0 plan and the brand guidelines in French for now
sites/       The website and the documentation, Astro sites built on the brand guidelines
```

### Building

Building Cahut needs Go 1.27 or later, make, and Node.js at the version of `.nvmrc` for the web console, which the binary embeds. `make build` builds `bin/cahut`, and `make check` runs the linters, govulncheck and the tests, as CI does, without Node.js. Run `make` for the full list of tasks. The sites need Node.js too: see [sites/README.md](sites/README.md).

Releases are static binaries for linux/amd64 and linux/arm64, built by GoReleaser when a version tag is pushed.

## Roadmap

Batches 0 to 11 are done. On 27 October 2026, the first version met its five success criteria on a fresh Ubuntu server, with Claude and ChatGPT ([acceptance](docs/recette-v0.md), in French):

| Criterion | Threshold |
| --- | --- |
| Install on a fresh VPS | Under 10 minutes, one command |
| First publication by an AI | Under 5 minutes, without touching the server |
| Sharing by a non-technical person | A deck shared by link in a few minutes |
| Supported AIs | The same Cahut works with at least two different AI clients |
| Governance | Every AI action is in the Logbook, and is either reversible or was approved |

Next, in about this order:

1. Comments, single sign-on, and backups and restore.
2. Apps in unprivileged containers: runtimes, a Dockerfile mode, databases, secrets, logs and resource limits.
3. Export to PDF and PowerPoint, SSH access limited to the same operations, and more AI clients.

The [roadmap on cahut.cloud](https://cahut.cloud/project/roadmap/) tells each batch, and what works today.

## Documentation

- [docs.cahut.cloud](https://docs.cahut.cloud): install, connect an AI, publish, the concepts, administration, and the reference of the MCP tools, the command line and the HTTP API.
- [Project brief](docs/brief-projet.md) (French): vision, positioning, scope, governance, technical choices and open questions.
- [V0 plan](docs/plan-v0.md) (French): scope, batches and milestones of the first version.
- [Governance and project life](docs/gouvernance.md) (French): how the project is run and how it lives: roles, contributions, support, channels, releases, security, name, license and money, with the decisions still open.
- [Architecture decision records](docs/adr/README.md): each technical choice, its reasons and the options set aside.
- [Brand guidelines](docs/charte-graphique/README.md) (French): logo, colors, typography, icons, illustrations, photography, interface kit and templates, with all source files. A visual version opens from [docs/charte-graphique/index.html](docs/charte-graphique/index.html).

## Contributing and community

Cahut is maintained by one person, in the open. [GOVERNANCE.md](GOVERNANCE.md) says how the project is run, [CONTRIBUTING.md](CONTRIBUTING.md) what helps and how a change gets in, [SUPPORT.md](SUPPORT.md) where to ask for help, and [SECURITY.md](SECURITY.md) how to report a vulnerability, privately. Questions and ideas go to the [Discussions](https://github.com/cahutcloud/cahut/discussions). Everyone who takes part follows the [code of conduct](CODE_OF_CONDUCT.md).

## License

Cahut is free software under the GNU Affero General Public License, version 3 ([LICENSE](LICENSE)). You can run it, study it, change it and share it. If you change it and let others use your version over a network, you give them your changes under the same license. Installing or distributing Cahut as it is asks nothing more than keeping the license with it.

The name Cahut and its logo identify this project and are not covered by this license. Say that something runs on Cahut or works with Cahut; do not use the name or the logo for a product, a service or a domain of your own, and give a fork its own name. [TRADEMARK.md](TRADEMARK.md) says this in full.

[ADR 15](docs/adr/0015-license.md) records the decision and the options set aside.

## Author

Cahut is a personal project by Olivier Durand, independent of any company and of any AI vendor.

<br>

<div align="center">

<picture>
  <source media="(prefers-color-scheme: dark)" srcset="docs/charte-graphique/logos/cahut-symbole-inverse.svg">
  <img alt="" src="docs/charte-graphique/logos/cahut-symbole.svg" width="64">
</picture>

<sub><a href="https://cahut.cloud">cahut.cloud</a> · <a href="https://docs.cahut.cloud">docs.cahut.cloud</a></sub>

</div>
