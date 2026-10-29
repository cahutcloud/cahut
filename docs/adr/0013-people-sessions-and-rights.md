# 13. People sign in to their Cahut, and the registry checks their rights

- Status: Accepted, superseded in part by [ADR 21](0021-ais-knock-first.md) (an AI that calls an operation that knocks first gets a Knock, not a refusal), extended by [ADR 30](0030-kind-and-limits-of-each-cahut.md) (in an individual Cahut, only administrators list people, and a Room has one member) by [ADR 31](0031-e-mail-through-a-relay.md) (a person has an e-mail address, where their links and a forgotten password go) by [ADR 43](0043-the-console-holds-two-sessions.md) (the console holds the session of a person and that of an administrator of the server together) and by [ADR 48](0048-suspending-a-cahut.md) (only the administrators of the server act in a suspended Cahut)
- Date: 2026-10-22

## Context

Batch 1 had no sign-in. The administrator created people and personal tokens on the command line, and the registry knew three actors: the command line, a person, which nothing used yet, and an AI. One flag, `AdminOnly`, kept an operation for the command line and hid it from AIs.

Batch 2 of the [V0 plan](../plan-v0.md) brings the console, team Rooms and the two levels of administration of [ADR 8](0008-several-cahuts-per-server.md). People need passwords and sessions, the console needs an API ([ADR 14](0014-console-http-api.md)), and every operation must say who may call it.

Three constraints hold. There is no e-mail in the V0. The console has its own port and keeps its session token in the storage of its origin, without cookies ([ADR 11](0011-console-on-its-own-port.md)). A request reaches one Cahut only, found from its domain before any record is read ([ADR 8](0008-several-cahuts-per-server.md)).

## Decision

### Accounts

- The people of a Cahut sign in at their Cahut's domain, with their handle and a password. A person is an `admin` or a `member` of the Cahut, and a Cahut keeps at least one administrator.
- The server's administrators have accounts of their own in `server.db`, apart from the people of the Cahuts. From the console of any domain, they declare Cahuts and read the server's Logbook. To act in a Cahut, they sign in to that Cahut's console, where they act as its administrators do.
- No person is deleted in the V0. An administrator disables a person, which ends their sessions and suspends their tokens, and can enable them again.

### Passwords and activation

- Passwords are hashed with Argon2id (RFC 9106), with the parameters OWASP recommends: 19 MiB of memory, 2 passes, 1 lane, and a random salt of 16 bytes. The hash is stored as a PHC string, which records its parameters, so that they can rise later. A semaphore bounds the hashes computed at once, so that a burst of sign-ins cannot exhaust the memory of a small server.
- A password has at least 15 characters, as NIST SP 800-63B asks of a password used without a second factor, and at most 256. There is no rule on what it is made of.
- Without e-mail, an administrator creates an account and passes on a single-use activation link, valid 7 days, where the person chooses their password. An administrator gives a new link to a person who forgot their password; the old password works until the new link is used. The command line gives a link for any account, a server administrator's included: it is the way back in when every administrator is locked out.
- The code of a link sits in the fragment of its address, `https://<domain>:8443/activate#<code>`, so that it reaches neither the server's logs nor a `Referer` header. Only its SHA-256 hash is stored.

### Sessions

- Signing in, or activating an account, opens a session: an opaque random token from `crypto/rand`, with the prefix `cahut_ses_`, stored as a SHA-256 hash like personal tokens. The console sends it in the `Authorization` header.
- A session ends after 7 days without use, and 30 days after it opened in any case. Signing out ends it. Changing one's password ends one's other sessions, and disabling a person ends all of theirs.
- A session is good at one domain. A person's session lives in the database of their Cahut, which no other Cahut reads. A server administrator's session lives in `server.db`, bound to the domain where it opened.
- Sign-in is rate limited, by address and by handle. An unknown handle costs the same hash as a known one, so that response times do not tell which handles exist.

### Rights

- Each operation states the level it needs: any person, an administrator of the Cahut, or an administrator of the server. This replaces `AdminOnly`. The command line acts as an administrator of the server, on the server and in every Cahut.
- The registry reads the role and the state of the person in the transaction of every call, whether the person acts in the console or an AI acts for them. A disabled person, and every AI that acts for them, loses access at once.
- The operations check rights on a Room themselves, with one shared helper. The owners of a Room manage its members and its title, and its members publish in it. Any person can create a team Room, and becomes its owner. Administrators manage every Room.
- Administrators manage people and Rooms, not content. An administrator of the Cahut or of the server sees every Room, its members and the whole Logbook, but opens only the Windows of the Rooms they belong to.
- A member reads in the Logbook their own actions, those of their AIs, and every entry about their Rooms. For this, Logbook entries record the id of the Room they concern.
- The operations that prove who someone is are not offered to AIs: signing in, activating an account, changing a password, creating a personal token. They give no power over content; they are how a person, and the tokens a person hands to AIs, come to be trusted. Every other operation of a Cahut is offered to AIs, within the rights of the person they act for. An AI sees only the tools its person may use: the AI of a member does not see `create_user`.
- An operation is marked as knocking first rather than as irreversible: `Irreversible` becomes `KnockFirst`. The list of batch 5 holds operations that can be undone but change what a person may do: adding a person, changing a person's role, disabling or enabling a person, adding or removing a member of a Room, changing a member's role. Until batch 5 brings Knocks, an AI that calls such an operation is refused, with a message that asks its person to act in the console. A person who calls one in the console acts at once, after a confirmation ([ADR 1](0001-operation-registry.md)).

### Logbook

- A server administrator appears in the Logbook as `admin`, with their handle. The command line appears as `admin`, without one.
- Sign-ins, refused sign-ins, activations and changes of password are logged, under the person concerned. The Logbook never holds a password, a session token or the code of a link.

## Consequences

- Each Cahut signs in its own people, and shares no account, session or link with another, as [ADR 8](0008-several-cahuts-per-server.md) requires.
- Someone who installs Cahut alone has two accounts: server administrator, and person of their Cahut. The first start of batch 7 can create both at once.
- A server administrator who looks after several Cahuts signs in to each of them. This replaces the line of [ADR 11](0011-console-on-its-own-port.md) where a server administrator who signed in at any domain reached every Cahut.
- Without e-mail, a forgotten password goes through an administrator, or through the command line. E-mail comes after the V0.
- Argon2id needs `golang.org/x/crypto`, already in the module graph through CertMagic.
- `KnockFirst` replaces the class of [ADR 1](0001-operation-registry.md): the registry turns a call by an AI into a Knock when the operation knocks first, whether or not it can be undone.
- The tests of every operation run over a matrix of callers: a server administrator, a Cahut administrator, a Room owner, a Room member, a person outside the Room, an AI, a disabled person.
- The people created in batch 1 have no password. The migration marks them as pending, and `cahut user link` gives each of them a link.

## Alternatives considered

- **Server administrators as people of a Cahut, with a flag.** One account for someone alone, but signing in at another domain would read the records of the Cahut that holds the account.
- **Server administrators' sessions good at every domain.** The console of one domain would act in another Cahut, and one request would touch two Cahuts.
- **bcrypt.** Well tried, and in the same module, but it cuts passwords at 72 bytes and does not use memory to slow down GPUs. RFC 9106 and OWASP recommend Argon2id first.
- **Signed sessions, such as JWT.** No lookup per request, but a signed token holds until it expires, and disabling a person must take effect at once.
- **Passwords set by the administrator.** No link to pass on, but the administrator then knows the password and passes it on all the same.
- **Administrators who open every Window**, as the owners of a GitHub organization read its private repositories. Simpler queries, but a Cahut holds personal Rooms, and a host that serves several organizations should not read them. Whoever runs the command line on the server can read the disk anyway.
