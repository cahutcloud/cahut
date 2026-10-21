# 4. SQLite in pure Go

- Status: Accepted, extended by [ADR 8](0008-several-cahuts-per-server.md) (one database per Cahut)
- Date: 2026-10-21

## Context

Cahut keeps its records on one machine: users, Rooms, Windows, Doors, Knocks, OAuth clients and tokens, and the Logbook. The files of each Window live on disk, next to the database. Releases are static binaries for linux/amd64 and linux/arm64, built with cgo disabled ([ADR 7](0007-module-path-repository-and-releases.md)).

## Decision

Cahut stores its records in SQLite, through [modernc.org/sqlite](https://gitlab.com/cznic/sqlite), a translation of SQLite's C code to Go, used with `database/sql`.

- WAL mode, foreign keys on and a busy timeout. One connection writes, a small pool reads.
- Schema migrations are SQL files embedded in the binary, applied in order at start, each in its own transaction.
- Plain SQL, no ORM.
- The database file and the Window files live in Cahut's data directory.

## Consequences

- No C toolchain anywhere: binaries cross-compile and stay static.
- The driver is slower than C SQLite on heavy queries, by a factor of 1.3 to 2 on measured sorts and groupings. At the scale of one home, this does not show.
- modernc.org/sqlite needs the exact `modernc.org/libc` version named in its `go.mod`. The two are updated together.
- Writes are serialized. This suits one machine, and it keeps the rule that a change and its Logbook entry share a transaction ([ADR 1](0001-operation-registry.md)) simple.
- Backups, after the V0, can use SQLite's online backup or `VACUUM INTO` while the home runs.

## Alternatives considered

- **mattn/go-sqlite3**, the reference driver. It needs cgo, and a C cross-compiler for arm64, which breaks static builds.
- **ncruces/go-sqlite3**, also without cgo, with good WAL support on Linux. Still before 1.0, and it changed its engine in March 2026. Worth another look later.
- **PostgreSQL.** A second service to install, update and back up, with no need for it at this scale. Apps will get their own PostgreSQL databases later; Cahut's records stay in SQLite.
