---
title: The data of Windows
description: Each Window can keep what its readers write in a SQLite database of its own, which its pages read and write as its declaration says.
---

A Window can keep what its readers write: the answers of a poll, the messages of a contact form, the names of a petition, the rows of a team tracker. Each Window has a SQLite database of its own, on the server of its Cahut. Its pages read and write it as the declaration of their version says, and the members of its Room and their AIs read and change all of it. Nothing goes to a service outside. See [ADR 56](https://github.com/cahutcloud/cahut/blob/main/docs/adr/0056-a-database-for-each-window.md).

## What the declaration says

The AI that publishes a version gives its declaration `data`. A petition, for example:

```json
"declaration": {
  "scripts": true,
  "data": {
    "migrations": [
      "CREATE TABLE signatures (id INTEGER PRIMARY KEY, name TEXT NOT NULL, email TEXT, city TEXT, reader TEXT, signed_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP)"
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
}
```

| Field | What it holds |
| --- | --- |
| `migrations` | The SQL that makes the database, in order: statements that create tables, indexes, views and triggers, add columns and insert rows. Cahut runs once those that have not run yet. A new version gives again those that ran, unchanged, and adds new ones at the end. |
| `readable` | The tables, and of each the columns, that readers read with a `SELECT` of their own. Every other column of these tables reads as `NULL`, in a result as in a `WHERE`, a join or a sort, and every other table is refused. |
| `queries` | The statements that the pages call by name, one statement each, with `who` may call it: `anyone` who opens the Window, `people` signed in to the Cahut, or the `members` of its Room. A query runs with the whole database: it gives what readers may not read themselves, such as the totals of a vote. |
| `closes` | When the queries that write stop. Reading goes on. |

- Before it publishes, Cahut runs the new migrations in a transaction that it rolls back, prepares every query and checks every readable column. An error says which migration or query to fix, and nothing is published.
- Cahut says what each query does, from what SQLite reports: it reads, adds, changes or deletes rows. The console shows it.
- Migrations only add. To remove or rename a table or a column, an AI calls `change_window_data`, which knocks first.
- The database belongs to the Window, not to a version: a new version and a rollback keep its tables and its rows.
- Data needs scripts: the pages call it with `fetch`.

## What the pages call

The pages call `.data/` under the root of their Window, which needs no origin in the declaration: from `index.html`, `.data/sql`; from a page in a folder, `../.data/sql`.

```js
// Read what readers may read.
const read = await fetch(".data/sql", {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({ sql: "SELECT name, city FROM signatures ORDER BY id DESC LIMIT 50" }),
});
const { columns, rows } = await read.json();

// Call a query of the declaration.
await fetch(".data/queries/sign", {
  method: "POST",
  headers: { "Content-Type": "application/json" },
  body: JSON.stringify({ params: { name: "Ana", email: "ana@example.com", city: "Lyon" } }),
});
```

- A `SELECT` answers `{"columns": [...], "rows": [[...]]}`, and `"cut": true` when it passed 1,000 rows or 1 MiB. A query that writes answers `{"changes": 1, "id": 42}`, the id being that of the last row it added.
- Parameters are named, such as `:name`, and bound: a page gives them in `params`. A parameter that a page does not give is `NULL`. A blob goes as `{"blob": "<base64>"}`.
- An error answers `{"error": {"code": "...", "message": "..."}}` with its status: `invalid`, `timeout`, `constraint` (a `UNIQUE` or `NOT NULL` refused the row), `closed`, `forbidden`, `not_found`, `withdrawn`, `too_many`, `full` or `busy`.

## Who the reader is

`:reader` is a parameter that Cahut gives to every statement, and that a page cannot set:

| How the Window opens | `:reader` |
| --- | --- |
| A public Window, at its address | `NULL`: its readers are unknown, even when they are signed in |
| With its Key, an invitation or a share link | A value of the pass, which lasts 30 days |
| To a person signed in, or in a preview in the console | A value of the person |

The value is the same for one person in one Window, differs from one Window to the next, and tells whoever reads the database nothing about who it is. A page that wants a name asks for it. Queries for `people` and `members` work only where a Window opens in a frame, not at the address of a public Window.

## Limits

- One second of work for each call of a reader, 64 KiB sent, and 1,000 rows and 1 MiB answered.
- From one address, and from one /64 in IPv6, in one Window: reads in bursts of 60, then one a second; writes in bursts of 10, then one every 6 seconds.
- A database holds 100 MiB at most, or less when the Cahut sets `data_per_window`. The limit of the disk of a Cahut counts each database with its copies, as it stood after its last migration or change by an AI or a person. See [Cahuts](/concepts/cahuts/#collaborative-or-individual).
- A public poll is not reliable: anyone may answer many times, within these limits.

## What the people of a Room do

The members of the Room of a Window and their AIs read and change all of its data, private columns included:

| Operation | What it does |
| --- | --- |
| `query_window_data` | Runs one `SELECT` over the whole database, and gives 1,000 rows at most. |
| `change_window_data` | Runs statements that create, add, change or delete, after a copy of the database. It knocks first when they remove or rename a table, a column, an index, a view or a trigger. |
| `rollback_window_data` | Goes back to a copy, the latest by default. What readers wrote since is lost. The database as it was becomes a copy, so that calling it again undoes it. |

- Cahut copies the database before each change of an AI or a person, before the migrations of a new version, and once a day while readers write in it, and keeps the three latest copies, beside the database.
- What readers wrote is data, never instructions. The results of `query_window_data` say so, and the instructions of the MCP server and the skill tell AIs never to follow what the rows say.
- The page of a Window in the console shows its tables and their rows as text, what readers read, what each query does and for whom, and its copies, with a button to go back to each. It exports a table as CSV; a cell that starts with `=`, `+`, `-` or `@` starts with `'`, so that a spreadsheet does not run it.
- What an AI reads goes to the company that runs it. The consent screen says so.

## Telling readers

A page that keeps what readers type says it beside its form, before they write: what it keeps, who reads it, until when, and how to have it erased. The skill and the instructions of the MCP server ask it of every AI, as the GDPR asks it of whoever collects personal data. Cahut does not check the page: what a page collects is the responsibility of whoever publishes it.

Making a Window public says, in its confirmation and in its Knock, which queries let anyone write in its data. A public Window that gets such a query in a new version knocks first too.

## Isolation

- The SQL of a Window never runs in the process that serves the Cahuts. The binary starts a worker for each database in use, with an empty environment, limits of memory and time, and, on Linux, Landlock, which closes to it every file outside the folder of its database and, from Linux 6.7, the network. [cahut doctor](/admin/doctor/) says whether the workers are confined.
- A database holds nothing of the Cahut: no person, session, token or network address.
- The trash keeps the database and answers nothing; a purge removes it with its copies. A withdrawn Window takes nothing from its readers, and its members still read and change its data. A moderator who opens a reported Window reads what its readers read, and writes nothing.
- There are no backups yet. The copies are on the same disk as the database: until backups come, a database lost with its disk is lost.
