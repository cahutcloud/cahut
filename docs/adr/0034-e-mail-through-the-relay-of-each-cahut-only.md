# 34. E-mail goes through the relay of each Cahut only

- Status: Accepted
- Date: 2026-10-24

## Context

[ADR 31](0031-e-mail-through-a-relay.md) gave e-mail two relays. The relay of the server sent the e-mail of the server and of every Cahut without a relay of its own; the relay of a Cahut, set by its administrators, took precedence for that Cahut. Batch 9 built both.

The maintainer decided otherwise, once both ran. A Cahut is the sender of its own e-mail. On a server that holds Cahuts for several organizations, a fallback would send in the name of one organization from the domain of the host, and the host would answer for what it sends. A Cahut that never chose to send e-mail would send some all the same. And a feature that depends on e-mail should not show where no relay sends it.

## Decision

- **Only a Cahut has a relay.** Its administrators set it, as ADR 31 has it: at a public address, with TLS, and kept only once Cahut has connected to it and signed in. The server has no relay, and nothing falls back to one. A migration removes the relay of the server from `server.db`.
- **Without a relay, a Cahut sends no e-mail, and shows nothing that would.** Its console offers no "Forgot your password?", no address on the page of the account or in People, no invitation on the page of a Window, and no notice to turn off. Its operations refuse what would send e-mail: `invite_to_window`, `change_email` and `request_password_link` say that the Cahut sends no e-mail. `get_cahut` says whether the Cahut sends e-mail, to people and to AIs. `describe_sign_in` tells the page of sign-in, before anyone signs in. Addresses that an administrator gave stay, and serve once a relay is set.
- **The administrators of the server have no e-mail.** They have no address, and receive no link and no notice. A forgotten password goes through another administrator, or through the command line, `cahut admin link`, as before batch 9. Failed updates, releases that wait, certificates and the disk show in the console of the server and in `cahut doctor`. The notices of the limits of a Cahut go to its administrators only.
- **A test goes to the address of the person who asks.** The command line, and an administrator of the server at the console of a Cahut, who has no address, name one.
- **In development**, a Cahut without a relay writes its messages to disk, as before, and its console shows the features of e-mail, so that they can be tried on one machine.

This record supersedes, in ADR 31, the relay of the server, the fallback to it, and the e-mail of the administrators of the server. The rest of ADR 31 holds.

## Consequences

- An organization that wants e-mail sets the relay of its Cahut, with the SMTP account of its own provider. Someone who installs Cahut for themselves does the same, for their one Cahut.
- A Postfix on the server itself cannot be the relay of a Cahut: a relay of a Cahut is at a public address. The relay of the provider of e-mail does it.
- A host gives no e-mail to the Cahuts it runs, and sends nothing in their name.
- The administrators of the server watch the console of the server and `cahut doctor`, as they did before batch 9.
- The tests run the relay in memory as the relay of a Cahut, through the hook that lets tests reach a relay on this machine.

## Alternatives considered

- **The relay of the server as a fallback**, as ADR 31 had it. One setting for a server of one Cahut, but a Cahut would send without having chosen to, from the domain of its host.
- **The relay of the server for the server alone**, without fallback. The administrators of the server would keep their notices, at the cost of a second setting, and of a relay that the maintainer does not want on the server.
- **The features of e-mail shown, and refused when used.** Simpler screens, but a person would find a "Forgot your password?" that cannot work.
