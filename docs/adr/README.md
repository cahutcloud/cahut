# Architecture decision records

Each record states one architecture decision for Cahut: its context, what was decided, what follows from it, and the options set aside. Once accepted, a record changes only for typos and links. A new record replaces it, and the old one links to the new one.

| Record | Decision | Status |
| --- | --- | --- |
| [1](0001-operation-registry.md) | One registry of typed operations is the only way in | Accepted |
| [2](0002-built-in-https-with-certmagic.md) | The binary serves HTTPS itself, with CertMagic | Accepted |
| [3](0003-one-origin-per-window.md) | Each Window gets its own web origin | Accepted |
| [4](0004-sqlite-in-pure-go.md) | SQLite in pure Go | Accepted |
| [5](0005-preact-console-embedded.md) | The console is a Preact app embedded in the binary | Accepted |
| [6](0006-in-house-oauth-server.md) | Cahut runs its own minimal OAuth authorization server | Accepted |
| [7](0007-module-path-repository-and-releases.md) | Module path, repository and releases | Accepted |

## Writing a record

- Copy the layout of an existing record: a title that states the decision, then Status, Date, Context, Decision, Consequences and Alternatives considered.
- Number records in order and never reuse a number. The file name is the number on four digits, then the title in a few words.
- Write in English: the records are public at launch.
- Add the record to the table above in the same pull request.
