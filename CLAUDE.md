# CLAUDE.md

Guidance for Claude Code in this repository.

## The project

Cahut is a self-hosted, open source home for what AIs make: documents, decks, dashboards and small apps, published over MCP to the user's own server and shared by link. It is a personal project by Olivier Durand, independent of any company and of any AI vendor.

The project is at the brief and brand stage: there is no code yet. Build, run and test instructions will be added here with the first prototype.

Read before working on anything substantial:

- [README.md](README.md): the public pitch, scope, governance model and technical choices.
- [docs/brief-projet.md](docs/brief-projet.md): the full project brief, in French, with the open questions.
- [docs/charte-graphique/README.md](docs/charte-graphique/README.md): the brand guidelines, in French.

The implementation language is **Go** (decided, see the README). The brief still lists Go or Rust as open: the README is the current reference.

## Languages

- Talk with Olivier in French.
- Write in English everything that will be public at launch: README, code, comments, commit messages, pull requests, product copy, interface text.
- `docs/` stays in French for now (brief and brand guidelines). Use correct French spelling with all accents.

## Product rules that shape the code

These are decided and should hold in every design and every line of code:

- **Bounded operations.** AIs act through typed operations (publish a page, create a database, share a link), over MCP or forced-command SSH. Never a free shell.
- **Knock first.** What can be undone, the AI does alone. What cannot (permanent deletion, making internal content public, inviting an outsider, changing permissions, erasing backups) waits for a human to approve it in the console, with a notification and an expiry.
- **Everything is logged.** Every action records which AI, on whose behalf, what it did and with what result.
- **One API.** The web console uses the same API as the AIs: no capability is reserved for it.
- **No AI on the server.** The AIs stay outside and connect over MCP.
- **One machine per instance**, on a standard AlmaLinux or Ubuntu. Data stays on that server; backups go only to the external S3 or SFTP storage the admin chooses.
- **Isolation.** Each app runs in an unprivileged container with resource limits, on its own web origin when a wildcard domain exists.

## Vocabulary

Product text speaks the language of the home, in plain English. Use these words consistently in the interface, the docs and the site:

| Word | Meaning |
| --- | --- |
| Home | The Cahut instance, on the user's server |
| Room | A personal or team space, with its members |
| Window | A living page: document, deck, dashboard or app |
| Door | The share link and its access level (public, password, internal, invited) |
| Key | The password or invitation that opens a protected page |
| House style | The organization's brand, applied by every AI |
| Knock first | The rule that the AI asks before any irreversible action |
| Logbook | The record of every action |

Room, Window, Door and Key take a capital in running text ("This Door closed yesterday."). Write the product name **Cahut**, never "CAHUT"; lowercase "cahut" exists only in the logo.

## Brand

For anything visual (site, console, mockup, slide, social image, e-mail, illustration) or any product copy, use the project skill **`charte-graphique-cahut`** in [.claude/skills/charte-graphique-cahut/](.claude/skills/charte-graphique-cahut/SKILL.md). The short version:

- Take colors, type, spacing, radii and shadows from [docs/charte-graphique/tokens/cahut-tokens.css](docs/charte-graphique/tokens/cahut-tokens.css). Never retype hex values.
- Use the logo SVG files as they are. Never redraw, recolor or rebuild the logotype.
- Interface components come from [docs/charte-graphique/ui/cahut-ui.css](docs/charte-graphique/ui/cahut-ui.css) (`c-` classes).
- Pages that Cahut serves for an organization carry that organization's house style, not Cahut's. Cahut's own brand applies to its own screens: console, sign-in, e-mails, site.

Changing the logo, the palette or the tokens is a change to the brand guidelines: update the guidelines README, `index.html`, both token files (CSS and JSON, kept in sync by hand), the PDF, and bump the version in the changelog table.

## Repository layout

```
docs/
  brief-projet.md      Project brief (French)
  charte-graphique/    Brand guidelines with all source files (French)
.claude/skills/
  charte-graphique-cahut/   Project skill for the brand
```

## Writing style

The README sets the tone: short sentences, concrete verbs, no hype. No exclamation marks, no capitals for emphasis, no promises of AI magic: Cahut hosts the work, it does not create it.

## Git

- `main` is the default branch. Work on a topic branch (`claude/<topic>` for Claude sessions) and open a pull request.
- Commit messages in English: a short imperative subject line, then a body that says what changed and why.
- The repository is private until launch and will be opened as is, history included. Keep it free of references to other projects, employers or clients, and never commit secrets.
