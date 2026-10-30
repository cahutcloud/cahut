# 56. Each Window has a database of its own, which its pages read and write as its declaration says

- Status: Accepted
- Date: 2026-10-30

## Context

A Window cannot keep what its readers do. Its pages run in a sandbox where `localStorage`, `sessionStorage`, IndexedDB and cookies throw errors ([ADR 10](0010-windows-in-a-csp-sandbox.md)), which said that storage provided by Cahut could come later. Their Content Security Policy lets them call the origin of their Cahut ([ADR 19](0019-window-versions-access-and-declaration.md)), but nothing there answers. A poll, a contact form, a petition, a sign-up sheet, a guest book or the tracker of a team needs somewhere to write. Today the only way is a form that posts to a service outside, which the declaration allows, and the data of readers then leaves the server, against a rule of the product.

Several things constrain the answer:

- The pages of a Window are written by an AI, and nobody reviews them line by line. Anyone who opens a public Window can read its scripts, and call what they call with curl.
- One process serves every Cahut of a server, and holds their keys and their databases ([ADR 8](0008-several-cahuts-per-server.md)). The database of a Cahut holds its people, their sessions and their tokens, and has a single write connection.
- What cannot be undone knocks first ([ADR 21](0021-ais-knock-first.md)), and every call of an AI goes to the Logbook.
- In the European Union, whoever collects personal data tells the person, when it is collected, who collects it, why, for how long, and how to have it erased (Article 13 of the GDPR).
- An AI that reads what strangers wrote can be steered by it.

This changes what a Window is, and comes with Cahut 1.0.

## Decision

### One database for each Window

- **Each Window has a SQLite database of its own**, a file in the folder of its Cahut, `windows/<id>/data/data.db`, beside the folders of its versions and never among the files that Cahut serves. Cahut creates it with the first migration of the Window. A Window without data has none.
- **It belongs to the Window, not to a version.** A new version and a rollback of the Window keep its tables and its rows.
- **It holds nothing of the Cahut:** no person, no session, no token, no network address. A reader appears in it only as `:reader`, a value that Cahut gives (below). Cahut keeps in the database itself how many of its migrations ran, and a hash of them, so that the database and its copies carry their own state.
- **Cahut opens only the database of the Window that a request names**, by its address, its frame or its id. No SQL attaches another database, loads an extension or sets a pragma.
- **It follows its Window.** In the trash, it is kept and answers nothing. Restored, it answers again. Purged, it is erased with its copies. A withdrawn Window ([ADR 54](0054-reporting-and-withdrawing-a-window.md)) and a suspended Cahut ([ADR 48](0048-suspending-a-cahut.md)) take nothing from readers and give nothing.

### The SQL of Windows runs in processes of its own

- **Cahut never runs the SQL of a Window in the process that serves the Cahuts.** The binary starts itself again, as a worker for one database. The worker gets an empty environment and no file of its parent. It holds no key, and opens no file but those of its database's folder: on Linux, Landlock closes the rest of the file system to it, and the network when the kernel restricts it. `cahut doctor` says whether the workers are confined.
- A worker serves one database, while it is in use, and stops after a minute without calls. A server runs a bounded number of workers at once. When all are busy, a call waits a moment, then is refused as unavailable.
- Each worker has limits of memory and of time. One that crashes or passes them is stopped, and only the calls to its Window fail.
- **Each database has a single write connection, in its worker,** and reads beside it, in WAL mode.
- In the worker, SQLite runs in defensive mode, with `trusted_schema` off, and with limits on the length of values, the depth of expressions and the size of queries. An authorizer allows each statement only the actions and the functions of its kind. A query is interrupted when its time is up.
- A bug of SQLite that a stranger reaches with SQL thus stays in one process, with the database of one Window, and cannot read the keys or the databases of the Cahuts.

### What the declaration says

The declaration of each version ([ADR 19](0019-window-versions-access-and-declaration.md)) gets `data`. A petition, for example:

```json
"data": {
  "migrations": [
    "CREATE TABLE signatures (id INTEGER PRIMARY KEY, name TEXT NOT NULL, email TEXT NOT NULL, city TEXT, reader TEXT, signed_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP)"
  ],
  "readable": { "signatures": ["name", "city", "signed_at"] },
  "queries": {
    "sign": {
      "sql": "INSERT INTO signatures (name, email, city, reader) VALUES (:name, :email, :city, :reader)",
      "who": "anyone"
    }
  },
  "closes": "2027-03-31T23:59:59Z"
}
```

- **`migrations`** is the list of the SQL that makes the database, in order. Cahut runs those that have not run yet, together, after a copy of the database. The list starts with those that already ran, unchanged: a change to one of them is refused, with the advice to add a migration.
- **Migrations only add:** they create tables, indexes, views and triggers, add columns, and insert rows. A migration that removes or renames anything, or changes or deletes rows, is refused, with the advice to call `change_window_data`, which knocks first when it removes or renames.
- **`readable` lists the tables and the columns that readers read** with SQL of their own. Everything else is private: a statement of a reader that names a table not listed is refused, and a column not listed reads as `NULL`, in its result as in a `WHERE`, a join, a sort or a view, so that nothing can be learned from it.
- **`queries` are the statements that the pages call by name**, one statement each, with `who` may call it: `anyone` who opens the Window, `people` signed in to the Cahut, or the `members` of its Room. They run with the whole database, so that a query gives what readers may not read themselves, such as the totals of a secret vote. Cahut says what each one does, from what SQLite reports when it prepares it: it reads, adds, changes or deletes rows.
- **Their parameters are named and bound**, never put in the text. Pages give them all except `:reader`, which Cahut gives and a page cannot set.
- **`closes`**, optional, is when the queries that write stop. Reading goes on.
- **`publish_window` and `update_window` check it all before they publish:** they run the new migrations on a copy, prepare every query, and check every table and column that `readable` lists. An error tells the AI what to fix, and nothing is published or run.
- `rollback_window` gives back the previous pages and their declaration. The database keeps what the newer migrations added, which the previous pages do not use.

### What the pages call

- **The pages call `.data/` under the root of their Window**: under its address for a public Window, under `/-/frame/<capability>/` in a frame. No file of a Window starts with a dot, so no file takes it. A page at the root calls `.data/sql`; a page in a folder calls `../.data/sql`.
- **`POST .data/sql`**, with `{"sql": "…", "params": {…}}`, runs one `SELECT` of the reader over what `readable` lists, and answers `{"columns": […], "rows": [[…], …]}`.
- **`POST .data/queries/<name>`**, with `{"params": {…}}`, runs a query of the declaration, and answers its rows, or `{"changes": 1, "id": 42}` when it writes.
- These calls take and give JSON, without credentials, with `Access-Control-Allow-Origin: *`, and Cahut answers their preflight. An error answers with its status, a message for the person who reads the page and a code.
- Each call is an operation of the registry ([ADR 1](0001-operation-registry.md)) for the readers of a Window (`PeopleOnly`, `Visitors`), as `report_window` is: `select_window_data` and `call_window_query`. They check the Window and its reader in a read transaction of the Cahut, then give the SQL to the worker: readers never take the write connection of the Cahut.

### Who the reader is

| How the Window opens | `:reader` | The queries it may call |
| --- | --- | --- |
| A public Window, at its address | `NULL` | `anyone` |
| A frame opened by a pass: its Key, an invitation or a share link | The value of the pass, for its 30 days | `anyone` |
| A frame opened by a person: a session of a reader, or a preview in the console | The value of the person | `anyone` and `people`, and `members` for the members of its Room |
| A frame of a moderator ([ADR 54](0054-reporting-and-withdrawing-a-window.md)) | `NULL` | The queries for `anyone` that read: moderators see what readers see, and write nothing |

- **`:reader` is a hash of the Window and of the person or the pass, with a key of the Cahut.** It stays the same for one person in one Window, differs from one Window to the next, and tells whoever reads the database nothing about who it is.
- A public Window opens at its address without a frame ([ADR 19](0019-window-versions-access-and-declaration.md)), so its readers are unknown to its data, even when they are signed in. Queries for `people` or `members` work only where a Window opens in a frame.
- The page never gets the name, the handle or the address of its reader. A page that needs a name asks for it.

### Limits

- **For each call of a reader:** one second of work, 64 KiB sent, and 1,000 rows and 1 MiB answered, with a mark when the answer was cut.
- **For each network address**, as the PROXY protocol gives it ([ADR 41](0041-the-proxy-protocol-from-trusted-proxies.md)), and by /64 in IPv6, in each Window: reads come in bursts of 60, then one a second; writes in bursts of 10, then one every 6 seconds. The addresses stay in memory, as for sign-in, and never reach the database.
- **For each Window:** a few calls at once, and a database of 100 MiB at most. A database that is full refuses writes, with the code `window_data_full`, and keeps being read.
- **For each Cahut:** the limit `disk` of [ADR 30](0030-kind-and-limits-of-each-cahut.md) counts the databases and their copies, and a new optional limit, `data_per_window`, bounds each database below 100 MiB. Writes are refused when the disk of the server is almost full, as publishing is.
- The Logbook of the Room records, once a day at most, a Window that reaches one of its limits.

### What people and AIs do with the data

**The members of the Room of a Window and their AIs read and change all of its data**, as they read and change the Window.

| Operation | Who calls it | What it does |
| --- | --- | --- |
| `query_window_data` | The members of its Room and their AIs | Runs one `SELECT` over the whole database, with parameters, and gives 1,000 rows at most, with a mark when it cut |
| `change_window_data` | The members of its Room and their AIs | Runs statements that create, add, change or delete, after a copy of the database; knocks first when they remove or rename a table, a column, an index, a view or a trigger |
| `rollback_window_data` | The members of its Room and their AIs | Goes back to a copy of the database, and says that what readers wrote since is gone; the database as it was becomes a copy, so that a rollback undoes itself |
| `export_window_data` | The members of its Room, in the console (`PeopleOnly`) | Gives a table as CSV, or the whole database as a SQLite file |
| `select_window_data` | Readers (`PeopleOnly`, `Visitors`) | Runs the `SELECT` of a page over what its declaration makes readable |
| `call_window_query` | Readers (`PeopleOnly`, `Visitors`) | Runs a query of the declaration, for those it names |

- `get_window` gives the data of a Window: its tables, their columns and their rows counted, its size, the migrations that ran, what readers read, each query with what it does and who calls it, when writing closes, and the copies.
- **What readers wrote is data, not instructions.** The results of `query_window_data` say so beside the rows, and the instructions of the MCP server and the skill tell AIs never to follow what the rows say.
- The console shows the tables of each Window and their rows as text, and exports them. The CSV puts `'` before a cell that starts with `=`, `+`, `-` or `@`, so that a spreadsheet does not run it.

### Copies

- **Cahut copies the database before each change of an AI or a person**, by `change_window_data` or by the migrations of a new version, and once a day while readers write in it. The copies stay in its folder, made with the backup API of SQLite. Cahut keeps the three latest.
- `rollback_window_data` goes back to the latest copy, or to one that the call names by its date.

### Knock first

- `change_window_data` knocks first when its statements remove or rename a table, a column, an index, a view or a trigger. It finds out by running them in a transaction that it rolls back, and comparing the schema before and after. The Knock says what would go. The copy undoes the rest.
- The Knock that makes a Window public, and the confirmation that the console asks of a person, say that the Window keeps what its readers write, and which of its queries add, change or delete rows for anyone.
- `update_window` knocks first when a public Window gets a query that changes or deletes rows for `anyone`, for the same reason.

### Telling readers

- **The page says it itself**, beside its form, before a reader writes: what it keeps, who reads it, until when, and how to have it erased. Cahut cannot add a line to a public Window, which it serves as it is ([ADR 42](0042-the-legal-pages-of-a-cahut.md)).
- **The instructions of the MCP server and the skill ask it of every AI**, for every Window whose queries write what readers type. Cahut does not check the page. The console shows what each Window keeps and who writes in it, and the Knock that makes it public reminds the person. What a page collects is the responsibility of whoever publishes it, as its content is.
- The skill also asks the pages to show what readers wrote as text, never as HTML, and to say that a public poll counts visits, not people.

### Moderation

- A withdrawn Window takes nothing and gives nothing at `.data/`, which answers `410 Gone`. Its members still read and change its data, to clear what made it withdrawn.
- A moderator who opens a reported Window sees what its readers see: its frame reads what is readable and calls the queries for `anyone` that read, and writes nothing. Rows that are not readable stay with the members of its Room.

### The Logbook

- `query_window_data`, `change_window_data`, `rollback_window_data`, `export_window_data` and the migrations of a version go to the Logbook of the Room, with what they do, the tables and the number of rows, and never the values.
- The calls of readers are not entries of the Logbook one by one, as the opening of a public Window is not: the database holds what they wrote.

## Consequences

- An AI publishes a poll, a form, a petition, a sign-up sheet or the tracker of a team as a Window, without a service outside, and what readers write stays on the server.
- What readers write is private until the AI declares it readable. An AI that forgets a column leaks nothing; one that declares too much publishes it, row by row, to anyone who opens the Window.
- A public poll can be filled by robots within its limits, and a vote whose choice is readable is not secret. The skill and the documentation say so.
- What readers write can steer the AI that reads it. Knock first stops what cannot be undone, and the Logbook and the copies undo what an AI does alone, once a person sees it.
- Data that an AI reads goes to the vendor of that AI. The consent screen ([ADR 16](0016-oauth-for-ai-clients.md)) says that the AI reads and changes the data of the Windows of the Rooms it reaches.
- A rollback of the data loses what readers wrote since the copy, and says so.
- There are no backups yet: the copies are on the same disk. Until backups come, and hold the databases of Windows, the documentation says so.
- The snapshot of `cahut update` leaves the databases of Windows, as it leaves their files ([ADR 24](0024-signed-releases-and-updates.md)): releases do not migrate them, and each carries its own state.
- Each database in use costs a process, and each call a round trip to it. On macOS and on a Linux without Landlock, the workers run with their limits but unconfined: `cahut doctor` says so.
- The console gets the data of each Window. The refusals that people meet, such as `window_data_full` and `window_data_closed`, carry codes that the console says in English and French ([ADR 37](0037-refusals-with-a-code-for-the-console.md)).
- Tests cover the private columns through every form of `SELECT`, view and subquery; a list of hostile SQL (recursive queries without end, huge strings, deep nesting) that must end in time and within memory; the confinement of the workers on Linux; and the matrix of callers of [ADR 13](0013-people-sessions-and-rights.md). The test of governance names the undo of each new operation.
- This record extends ADR 4 (the SQL of Windows runs in processes of its own), ADR 10 (Cahut keeps the data of Windows, on the server and not in the browser), ADR 19 (the declaration says what the data holds and who writes it, and the database of a Window outlives its versions), ADR 21 (`change_window_data` and `update_window` knock first for some calls), ADR 30 (the disk counts the databases, and a limit bounds each one) and ADR 54 (a withdrawn Window takes no data from readers, its members still change it, and moderators write none).

## Alternatives considered

- **Tables in the database of the Cahut.** One file to back up, but the SQL of pages would run beside the people, sessions and tokens of the Cahut, and every reader who writes would take its single write connection.
- **One database for each Room, or for each Cahut.** Fewer files and processes, but the SQL of one Window could read the data of another.
- **A store of keys and values, or of JSON documents.** Simpler to secure, but polls and trackers count, sort and filter, and AIs write SQL well.
- **Storage in the browser**, with `allow-same-origin`. It would undo the sandbox ([ADR 10](0010-windows-in-a-csp-sandbox.md)), and what a reader writes would stay in their browser.
- **Queries declared for reading too, and no SQL of readers.** A smaller surface, but each view of a page would wait for a new version. Readers keep their own `SELECT`, held by what is readable and by a process of its own.
- **Every column readable unless declared private.** A single column forgotten would publish the addresses of every reader.
- **The SQL of Windows in the process of Cahut**, held by the authorizer and limits alone. Simpler, but a bug of SQLite that a stranger reaches would run beside the keys and the databases of every Cahut, and a crash would stop them all.
- **Readers who write with SQL of their own.** One reader could change or delete what the others wrote.
- **Migrations that remove or rename without a Knock**, since a copy undoes them. A rollback to the copy would lose what readers wrote after it, and the pages that use what went would break meanwhile.
- **A notice that Cahut draws in or around each Window.** Cahut serves public Windows as they are, and does not draw on their content ([ADR 42](0042-the-legal-pages-of-a-cahut.md)).
- **The name of the reader given to the page.** The database would hold the people of the Cahut, and a Window would learn who reads it.
- **Each call of a reader in the Logbook.** A popular poll would fill the Logbook of its Room.
- **The databases of Windows in the snapshot of `cahut update`.** Releases do not migrate them, and copying hundreds of databases would slow every update.
