# 31. Each Cahut sends e-mail through its own relay or that of its server

- Status: Accepted, superseded in part by [ADR 34](0034-e-mail-through-the-relay-of-each-cahut-only.md) (only a Cahut has a relay, without fallback, and the administrators of the server have no e-mail)
- Date: 2026-10-24

## Context

The V0 sends no e-mail ([ADR 13](0013-people-sessions-and-rights.md)). An administrator creates an account and passes the activation link on by hand, and a person who forgot their password asks an administrator for a new one. A Knock shows only in the console, and waits 24 hours for a person who may not open it ([ADR 21](0021-ais-knock-first.md)). Whoever sets a Key passes it on through another channel ([ADR 19](0019-window-versions-access-and-declaration.md)). Each of these records left e-mail for after the V0.

A person of a Cahut has a handle and a name, and no e-mail address. Codes of links are kept hashed, and the Key of a Window is kept hashed and shown once. Each Cahut has a single write connection to its database, and slow work stays out of the transaction of a call.

A server can hold Cahuts for several organizations ([ADR 30](0030-kind-and-limits-of-each-cahut.md)). There, the administrator of a Cahut is trusted with that Cahut, not with the server or the networks it reaches. Whoever controls the relay of a Cahut also reads every link that leaves it, those that reset a password included.

## Decision

### Relays

- Cahut never delivers e-mail to the servers of its recipients. It hands each message to an SMTP relay, with TLS: implicit TLS, or STARTTLS, with a verified certificate. It authenticates when the relay asks for it, and never sends a password over a connection without TLS.
- **The relay of the server** is set by the administrators of the server, in the console or on the command line, and kept in `server.db`. It may be at any address, the server's loopback and local networks included, since they run the machine; on the loopback address it may go without TLS, since the message does not leave the machine. It sends the e-mails of the server, and those of every Cahut that has no relay of its own.
- **The relay of a Cahut** is set by the administrators of that Cahut, and kept in its own database. When it is set, the e-mails of the Cahut go through it. It must be at a public address: Cahut resolves its host and refuses a loopback, private or reserved address, on every connection, with the guard that the documents of OAuth clients already go through ([ADR 16](0016-oauth-for-ai-clients.md)). A relay on the server or on its networks is the relay of the server.
- A relay has a host, a port, a mode of TLS, a user name and a password, and the address and the name of the sender. The e-mails of a Cahut take the name of the Cahut as the name of their sender when its relay names none.
- `set_mail_relay`, `get_mail_relay` and `send_test_mail` exist in each Cahut, for its administrators, and on the server, for its administrators. `set_mail_relay` connects to the relay in its `Prepare`, outside the transaction, and refuses one that does not answer, does not offer TLS or refuses the credentials, with the answer of the relay. A relay given empty is removed, and a Cahut goes back to the relay of the server. `get_mail_relay` gives everything but the password, with the last message sent and the last error. `send_test_mail` sends a message to the caller's own address.
- No AI is offered these operations: they are `PeopleOnly`. Whoever sets a relay reads the links that go through it, and can take the accounts they open.
- The password of a relay is never given back and never logged. It is encrypted with AES-256-GCM, under a key of the server kept in a file of its data directory, beside `server.db` and outside the directory of every Cahut: a copy of a Cahut, such as a backup, does not carry the password of its relay.

### Addresses of people

- A person may have one e-mail address, unique within their Cahut, compared without regard to case. A person without one works as today.
- An administrator gives it when they create the person, with `create_user`, or later, with `set_user_email`, which knocks first for an AI: the address decides where the links of a person go. Links go to it at once, and using one verifies it.
- A person sets their own address on the page of their account, with `change_email`, which no AI is offered, as no AI changes a password. The new address replaces the old one only once the person uses a link sent to it, valid 24 hours, and the old address gets a notice.
- An address is read by its person and by the administrators of the Cahut, and by no one else. No operation gives an address to an AI: what an AI reads goes to its vendor.
- The administrators of the server have addresses too, in `server.db`, under the same rules. The first start ([ADR 25](0025-first-start-in-the-browser.md)) asks for the address of the first person, and does not require it.

### Links by e-mail

- `create_user` with an address sends the activation link to it, once the call is committed: for an AI, once its person approves the Knock. `create_activation_link` sends the link to the address of the person, when they have one. Both still give the link to the administrator, so that an e-mail that does not arrive costs nothing.
- **Forgotten passwords.** The sign-in page of the console asks for a handle or an address, and `request_password_link`, for visitors only, sends a link to the address of the account. Its answer is the same whether the account exists, has an address or is disabled, and nothing goes to a disabled account. It is rate limited by address and by account, as sign-in is, and an account gets at most one such e-mail every 5 minutes.
- That link is valid one hour and replaces any earlier link of the person. Their password works until they use the link, and using it ends their other sessions, as changing a password does. The administrators of the server have the same on their own accounts. The command line stays the way back in.
- The code of every link sits in the fragment of its address, as in ADR 13, and opening a link does nothing until the person acts on the page. Mail filters that open links to check them use up nothing.

### Invitations to a Window

- `invite_to_window` sends the address of a Window to people, with a note of up to 500 characters in plain text. The members of the Room of the Window call it, as they set its access (ADR 19), and their AIs call it without knocking.
- **People of the Cahut** are named by their handle. They must already open the Window: members of its Room, or anyone signed in when it opens to the Cahut. Cahut finds their address, which the caller never learns, and the e-mail gives the address of the Window, where they sign in.
- **Anyone else** is named by their address, and the Window must open with its Key or to the public. A public Window gets its address. A Window with a Key gets a personal invitation link for each address, `https://<domain>/console/invitation#<code>`, whose code is kept hashed. A Window open to its Room or to the Cahut is refused, with a message that offers its Key or an invitation by handle.
- In an individual Cahut, invitations go by address only, and the same rules apply whether or not a person of the Cahut holds that address: no member names another person (ADR 30).
- An invitation link opens a page of the console that names the Window and who invites, and opens the Window on a click. It gives a pass, as the right Key does (ADR 19), which the console hands to the Windows with a single-use code, as it does for the sign-in of readers ([ADR 20](0020-console-under-a-path-of-port-443.md)). The link works again on another device until the invitation ends.
- An invitation stands for its Key, for one person. It is valid 30 days, and never beyond the end of the access. A new Key, or an access other than `key`, ends it as it ends the passes of the old Key. `revoke_invitation` ends one invitation and its passes, and closes the Window to that recipient alone. It undoes `invite_to_window` for the test of governance; the e-mail stays sent, and its link opens nothing more.
- The Key itself never goes by e-mail: Cahut keeps it hashed, and shows it once to whoever set it.
- An AI calls `invite_to_window` without knocking. It already sets the Key of a Window and reads it (ADR 19), and could pass it on anywhere; an invitation gives no more, names its recipients in the Logbook, and can be revoked one by one. A call names 20 recipients at most, and a person, their AIs included, sends 100 invitations a day at most.
- The e-mail names the person who invites, and the AI when an AI sent it. The note shows apart from the words of Cahut, as the words of whoever wrote it.
- `list_invitations` gives the invitations of a Window to the members of its Room: the recipient, who invited, when, and whether the invitation still stands.

### Notices

| When | Who gets an e-mail |
| --- | --- |
| A Knock waits | The person it acts for, with the title of the operation, what it acts on, its expiry and its address in the console. Never text written by the AI. |
| A password changes, an address changes, a personal token is created, an AI connects, a role in the Cahut changes | The person, at their address; the old address for a change of address |
| A person is added to a team Room | That person |
| The access of a Window ends in three days | The person who set it last |
| A limit of ADR 30 reaches 90%, or is reached | The administrators of the Cahut and of the server; in an individual Cahut, the person, for a limit of their own |
| An update fails and is rolled back, or waits for a person ([ADR 24](0024-signed-releases-and-updates.md)); a certificate cannot be renewed; the disk of the server is 90% full | The administrators of the server |

- A person can turn off the notices that are neither about their account nor about a Knock. The same notice of a limit, an update, a certificate or the disk goes out once a day at most.
- In an individual Cahut, no e-mail to a member names another member.

### Sending

- A call builds its messages and hands them over once it is committed (`Call.AfterCommit`), never in its transaction: a call that fails, or that turns into a Knock, sends nothing, and a slow relay holds no write connection.
- A queue in the memory of the server sends them, tries again for about a quarter of an hour, then gives up. Nothing waits on disk: a message may hold a link whose code Cahut keeps hashed only. After a restart, a lost message means asking for a new link.
- Messages carry a part in plain text and a part in HTML, in English, as the console speaks. They carry the brand of Cahut, as [ADR 12](0012-website-and-docs-as-astro-sites.md) has it: like the console, they are Cahut's own screens. Their images travel with them, and nothing loads from the internet: no tracking pixel, no rewritten link. They are marked `Auto-Submitted: auto-generated` (RFC 3834), so that auto-replies do not answer them.
- Cahut uses the standard library of Go: `net/smtp`, `crypto/tls`, `mime/multipart` and `mime/quotedprintable`. No dependency is added.
- `get_cahut` says whether the Cahut sends e-mail, so that an AI knows before it invites. Without a relay, a Cahut works as in the V0: the console shows the links, the sign-in page tells a person who forgot their password to ask an administrator, and `invite_to_window` is refused with a message that says so.

### Logbook

- Setting a relay is logged with its host, its port and its sender, and says when the password changes, never what it is.
- An e-mail is not an action. The operation that sends it is logged as it is today; `invite_to_window` names its recipients, by handle or by address. No code goes to the Logbook.
- A message that the relay refuses, or that the queue gives up, shows beside the relay in the console, for the administrators who set it. The Logbook does not hold it, as it does not hold the expiry of a Knock.

### What Cahut does not send

- No operation sends free e-mail. AIs reach people through invitations and Knocks, in the words of Cahut.
- No Window sends e-mail: a form goes only where its declaration says (ADR 19), and Cahut relays nothing for it.
- Newsletters and digests are not part of this record.

## Consequences

- A person with an address gets back in on their own after forgetting their password. Administrators no longer pass links on by hand, and still can.
- A Knock reaches its person away from the console, and its expiry of 24 hours becomes workable.
- A Window with a Key opens to people chosen one by one, and closes to each of them alone. The Key no longer needs to travel.
- Cahut keeps the addresses of people: personal data, read by the person and the administrators of their Cahut only. A person removes theirs, and then gets no link by e-mail.
- Migrations add the address of each person and the relay to the database of every Cahut, and the invitations and the notices that each person turned off, and the relay and the addresses of administrators to `server.db`.
- Whether e-mail arrives depends on the relay and on the domain of its sender: SPF, DKIM and DMARC are set with the provider of the relay. The documentation says how.
- The tests run a relay in memory. The tests of the new operations run over the matrix of callers of ADR 13, and the test of ADR 30 that walks an individual Cahut reads the e-mails too.
- The house style in e-mails, which [the plan of batch 6](../plan-lot-6.md) foresaw, comes with a later record. E-mails in other languages come with a console in other languages.
- This record extends ADR 13 (addresses of people, and forgotten passwords by e-mail), ADR 19 (invitations to a Window), ADR 21 (Knocks by e-mail) and ADR 25 (the address of the first person).

## Alternatives considered

- **Delivery to the servers of recipients.** No relay to set, but many hosts close port 25, a new server has no reputation, and signing with DKIM would make Cahut keep keys in the DNS of each domain.
- **A relay of the server only**, set in `/etc/cahut/cahut.env`. Familiar to those who run servers, but outside the console, and an organization on a host could not send from its own domain.
- **A relay of each Cahut only.** Someone who installs Cahut for themselves would set it once all the same, but a host could not give one relay to all its Cahuts.
- **A relay of a Cahut at any address.** An organization could use a relay on its own network, but the server would connect wherever the administrator of a Cahut points it, its own loopback and the networks of the host included.
- **The APIs of e-mail providers**, over HTTPS. Better reports of what bounced, but one integration per provider, where every one of them offers SMTP.
- **A queue of messages in the database.** It would survive a restart, but would keep the codes of links in clear, which Cahut keeps hashed everywhere else. A link lost to a restart is asked for again.
- **The password of a relay in clear in the database of its Cahut.** Simpler, but every backup of the Cahut would carry it.
- **The Key by e-mail.** It would have to be kept in clear, which ADR 19 set aside, and one Key shared by many cannot be taken back from one of them.
- **Invitations that knock first for an AI.** An AI already reads the Key it sets and could pass it on: a Knock would add a wait without a guard. Invitations are bounded, logged and revoked one by one instead.
- **The house style in e-mails now.** An invitation would look like the organization, but the style sheet of the house style is made for pages, not for mail clients. For now, e-mails carry the brand of Cahut.
- **A tool that sends any e-mail.** AIs could write to anyone in any words, and a Cahut would become a relay of spam and phishing in the name of its organization.
