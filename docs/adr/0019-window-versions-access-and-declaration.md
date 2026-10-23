# 19. Windows keep two versions, open to an access, and declare what they call

- Status: Accepted, extended by [ADR 21](0021-ais-knock-first.md) (an AI that makes a Window public knocks first, and a Window in the trash can be purged) and by [ADR 22](0022-house-style-of-each-cahut.md) (the pages of Cahut carry the house style of the Cahut)
- Date: 2026-10-22

## Context

Until batch 4 of the [V0 plan](../plan-v0.md), a Window has one version and cannot change, and anyone who has its address reads it. The V0 asks for more:

- an AI updates a Window and keeps the previous version, and one step goes back to it;
- each Window opens to its Room only, to everyone signed in to its Cahut, to anyone with its Key, or to the public, with an end date that can be pushed back;
- the AI declares what a Window does, and a Content Security Policy holds it to that declaration: a Window that declares no call can make none;
- a Window goes to a trash it can come back from.

[ADR 10](0010-windows-in-a-csp-sandbox.md) settled how a protected Window is served: a page of Cahut checks its reader and frames the Window in a sandboxed iframe, whose address under `/-/` carries a capability. [ADR 11](0011-console-on-its-own-port.md) settled how a reader signs in on port 443: through the console, which hands a single-use code to `/-/session`. This record settles the rest.

## Decision

### Versions

- `update_window` replaces every file of a Window. The new version becomes the current one, the version it replaces becomes the previous one, and the version before that is deleted, files included. The address stays, and the title may change.
- `rollback_window` swaps the current and the previous version, so that a rollback is undone by another rollback. It changes no file.
- Versions are numbered from 1 and a number is never reused: after a rollback from 3 to 2, the next update is 4, and 3 is deleted.
- Each version records who published it and when, its size and its number of files, and the declaration of its AI.

### Trash

- `trash_window` puts a Window in the trash: it no longer opens, it leaves the lists, and its name stays taken. `restore_window` brings it back as it was, with its versions and its access. A Window in the trash cannot be updated.
- Purging a Window cannot be undone: it comes with Knock first, in batch 5.

### Access

- A Window opens to one of:

  | Access | Who opens it |
  | --- | --- |
  | `room` | The members of its Room, signed in. The default. |
  | `internal` | Everyone signed in to its Cahut. |
  | `key` | Anyone who types its Key, and the members of its Room. |
  | `public` | Anyone who has its address. |

- An access other than `room` may end at a date. After that date, the Window opens to its Room only, until a person sets an access again. Setting it again moves the date or removes it.
- `set_window_access` sets the access, its end and the Key. Those who publish in the Room set it: its members. Administrators manage people and Rooms, not content, and do not set the access of a Window outside their Rooms ([ADR 13](0013-people-sessions-and-rights.md)).
- An AI sets every access but `public`. Opening a Window to its Cahut or to the holders of a Key keeps it among people the organization chose, and can be undone at once. Making it public waits for Knock first: until batch 5, an AI that asks gets an error that sends its person to the console.
- The Windows published before this record open to their Room only once the migration runs.

### Keys

- A Key opens one Window. A person chooses it, from 10 to 256 characters, or Cahut draws one: twelve characters in three groups, from an alphabet without look-alike characters, about 59 bits. Cahut keeps it hashed with Argon2id, with the parameters of passwords, and shows it once, to whoever set it. The Logbook never holds it.
- Setting the access to `key` again keeps the Key, unless a new one is asked for. A new Key, or an access other than `key`, ends every pass of the old one.
- A reader without access who opens the address of a Key Window gets a page of Cahut that asks for the Key. The form posts to `/-/key`. A right Key gives the reader a pass: a random token, stored hashed, in a cookie of its own for that Window, `__Host-cahut-key-<id>`, valid 30 days, and never beyond the end of the access.
- Attempts are limited by address and by Window, and each one costs a hash, under the semaphore of passwords.

### Readers

- A reader signs in on port 443 through the console. `/-/sign-in` sends them to the console's page `/read`, where they sign in if needed. The console calls `sign_in_to_windows`, an operation for people only, which returns a single-use code valid 60 seconds, and sends the browser to `/-/session` with it. Cahut exchanges the code for a reader session in the cookie `__Host-cahut-reader`, then sends the browser back to the path it came from, on the same domain.
- A reader session lives in the table of sessions, marked as a reader's, with the same lifetimes as a session of the console: 7 days without use, 30 days in any case. Signing out of the console does not end it; disabling the person and changing their password do. `/-/sign-out` ends it.
- Every cookie of Cahut on port 443 is `HttpOnly`, `Secure`, `SameSite=Lax`, on the path `/`, with the `__Host-` prefix. A server for development, in plain HTTP on `localhost`, drops the prefix and `Secure`.
- The administrators of the server are not people of a Cahut: they do not sign in as readers.

### Serving

- A public Window is served at its address, as [ADR 10](0010-windows-in-a-csp-sandbox.md) has it.
- For a protected Window, its address answers with a page of Cahut that frames it: an iframe, sandboxed with the same flags as the policy of its files, that allows fullscreen, whose source is `/-/w/<capability>/` followed by the path and the query of the request. A small script of Cahut, `/-/frame.js`, passes the fragment of the address on to the frame.
- A capability names a Window, the reader session or the pass that opened it, and an expiry four hours later, signed with HMAC-SHA256 under a random secret of the Cahut, kept in its database. Each file requested under it checks the signature, the expiry, and that its session or pass still opens the Window. A change of access, a new Key, a removed member or a Window in the trash takes effect on the next file.
- A Window that its reader may not open answers like a Window that does not exist, `404 Not Found`, with a link to sign in, but for a Key Window, which asks for its Key.
- The address of a Room lists the Windows that its reader may open, and answers `404 Not Found` to a reader outside the Room who may open none. The address of the Cahut lists the Rooms of a reader signed in, and offers to sign in otherwise.
- The pages of Cahut on port 443 are rendered on the server, with no inline script, under `default-src 'none'`, with only the style sheet, the frame script and the frames they need, and `frame-ancestors 'none'`. They carry the house style of the Cahut, neutral until batch 6.

### Declaration

- `publish_window` and `update_window` require a declaration of what the pages of the new version do:

  | Field | Meaning |
  | --- | --- |
  | `scripts` | Whether the pages run scripts |
  | `origins` | The origins they load anything from or call, such as `https://cdn.jsdelivr.net`; `https://*.example.com` for the subdomains of a domain |
  | `forms` | The origins their forms send to |

  Origins are `https` or `wss`, with a host and an optional port, and at most 20 in each list.
- Each file of the Window is served with a policy made from the declaration of its version. `<self>` is the origin of the Cahut, such as `https://cahut.example.com`, written out rather than `'self'`, which an opaque origin may not match:

  ```
  sandbox allow-forms allow-modals allow-popups allow-popups-to-escape-sandbox allow-downloads [allow-scripts];
  default-src 'none';
  script-src <self> 'unsafe-inline' 'unsafe-eval' <origins>   (only when scripts is true)
  style-src <self> 'unsafe-inline' <origins>;
  img-src <self> data: blob: <origins>;
  font-src <self> data: <origins>;
  media-src <self> data: blob: <origins>;
  connect-src <self> <origins>;
  frame-src <self> <origins>;
  worker-src <self> blob:;
  form-action <forms> or 'none';
  base-uri 'none'
  ```

- A Window declared without scripts loses `allow-scripts` from its sandbox as well.
- The declaration is returned by `get_window`, shown in the console, and written to the Logbook with the publication.
- The versions published before this record have no declaration. They keep the policy of ADR 10, and the console says they are not declared. Their next update declares.

## Consequences

- Updating a Window costs the disk of two versions at most. A third version deletes the first, which cannot be undone; the V0 accepts it, as the plan asks for two versions.
- A protected Window cannot be read without its reader's credentials, and a capability copied out of the page stops opening it as soon as its reader loses access, and four hours later in any case.
- A file of a protected Window costs a few reads of the database: the capability, the session or the pass, and the membership. SQLite answers them from memory.
- An AI that forgets a library's origin gets a page that fails to load it. The MCP instructions, and later the skills, ask for the declaration with the pages.
- The declaration limits what pages load and call, not where their reader goes: a link the reader follows, or a script that navigates its own frame, still leaves. A Window without scripts cannot navigate on its own.
- The reader of a protected Window sees the address of the Window in the address bar, which does not follow the pages of a multi-page site.
- The open question of the V0 plan, whether a Window opened to its Cahut or with a Key by an AI should knock first, is answered: it does not.

## Alternatives considered

- **Every version kept.** A history, and a rollback to any version, but disk without bound and a purge to design. Two versions are the V0's scope.
- **A rollback that deletes the current version.** Simpler, but it cannot be undone, and would have to knock first.
- **Capabilities stored in the database.** Revocable one by one, but a write at every view. The signature and the checks of each file revoke them as well.
- **Keys kept in clear text,** so that the console shows them again. A leak of the database would open every Key Window.
- **One cookie of passes for every Window.** Fewer cookies, but a value that grows with each Key, and that Cahut must rewrite.
- **An optional declaration, permissive by default.** Nothing to change in AIs, but then a Window that declares nothing can call anything, which is what the V0 plan rules out.
- **A list of origins for each directive.** Finer, but AIs would get it wrong, and loading from an origin is a call to it all the same.
