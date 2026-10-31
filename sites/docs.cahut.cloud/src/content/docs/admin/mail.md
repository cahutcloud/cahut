---
title: E-mail
description: How a Cahut sends e-mail through the SMTP relay that its administrators set, what it sends, and how to set and test the relay.
---

A Cahut sends e-mail through the SMTP relay that its administrators set: the links where people choose their password, invitations to read a Window, and notices. It never delivers messages to the servers of their recipients itself.

## Each Cahut, its own relay

Only a Cahut has a relay. Its administrators set it, in the console under **E-mail**, or with `cahut mail set`. The server has none, and nothing falls back to one: each Cahut sends its e-mail from an address of its own, with the account of its own provider.

A Cahut without a relay sends no e-mail, and its console offers nothing that would: no **Forgot your password?** on the page of sign-in, no address on the page of an account or under **People**, no invitation on the page of a Window, no notice to turn off. Administrators pass links on by hand, as before. The addresses already given stay, and serve again once a relay is set.

The relay of a Cahut speaks TLS, from the start or after STARTTLS, and is on the internet: Cahut refuses a relay at a loopback, private or reserved address, such as a Postfix on the server itself.

Cahut connects to a relay and signs in before it keeps it. The answer of the relay tells you what is wrong: a host that does not answer, a relay that offers no STARTTLS, a refused password.

## Set a relay

In the console, the page **E-mail** asks for what your provider of e-mail gives:

| Field | What to give |
| --- | --- |
| Host | The name of the relay, such as `smtp.example.com`, without a scheme or a port |
| Port | 587 for STARTTLS, 465 for TLS from the start |
| Security | STARTTLS or TLS |
| User name and password | Those of the account at the relay; empty when it takes no authentication |
| From | An address that the relay accepts, such as `cahut@example.com` |
| Name | The name messages come from; the domain of the Cahut when empty |

On the command line, the password comes on the standard input, never as an argument:

```bash
printf '%s' "$SMTP_PASSWORD" | sudo -u cahut /usr/local/bin/cahut mail set --host smtp.example.com --port 587 \
  --user cahut@example.com --password-stdin --from cahut@example.com
```

```text
The relay answered, and sends the e-mail of cahut.example.com.

host  smtp.example.com:587, starttls
user  cahut@example.com, with a password
from  cahut@example.com
set   2026-10-24 14:05
```

On a server that holds several Cahuts, `--domain cahut.example.com` names the Cahut.

Setting the relay again without a password keeps the password, as long as the host and the user name do not change. `cahut mail show` prints the relay, `cahut mail remove` removes it: the Cahut then sends no e-mail.

## The password of a relay

Cahut keeps the password of a relay sealed with AES-256-GCM, under a key of the server in `mail.key`, in the data directory beside `server.db`, readable by the service only. The password is never shown again, and never goes to the Logbook. A copy of the directory of a Cahut, a backup for instance, does not carry it: a Cahut moved to another server needs its password typed again.

## Test

**Send a test to your address**, on the page **E-mail**, sends a message to the address of your account, and says whether the relay took it. On the command line, which has no address, name one:

```bash
sudo -u cahut /usr/local/bin/cahut mail test --domain cahut.example.com --to you@example.com
```

The page also shows when a message last went out, and the last one that failed, with the answer of the relay. A relay that answers with a code 4xx, such as a busy server, gets the message again after 1, 5 and 10 minutes; a code 5xx is final. Messages wait in the memory of the service, never on disk: one lost to a restart means asking for its link again.

## Deliverability

Whether messages arrive depends on the relay and on the domain of the sender. Set SPF, DKIM and DMARC for that domain with the provider of the relay, as it explains: Cahut signs nothing itself. Messages are marked `Auto-Submitted: auto-generated`, so that auto-replies do not answer them, and carry no tracking pixel or rewritten link: their logo travels with them.

## What a Cahut sends

| Message | To |
| --- | --- |
| The link where a new person chooses their password | The person, when an administrator gives an address |
| A link to choose a new password | A person who forgot theirs, from **Forgot your password?** on the page of sign-in |
| A link to confirm a new address | The new address, when a person changes theirs |
| An invitation to read a Window | The people named, by handle or by address. See [Invitations](/concepts/access/#invitations) |
| A Knock that waits | The person its AI acts for. See [Knock first](/concepts/knock-first-and-logbook/#knock-first) |
| A password, an address, a role changed; a personal token created; an AI connected | The person |
| Added to a team Room; the access of a Window that ends in three days | The person, unless they turned these notices off on the page of their account |
| A limit of the Cahut at 90% or reached | Its administrators |
| A report about a Window | Its administrators, and its address of moderation. See [Moderation](/admin/moderation/#e-mail) |
| The receipt of a report, then the decision | Whoever reported, when they gave an address |
| A Window withdrawn, with the reason, each new reason, then reinstated | The members of its Room, or its author for a Window that is theirs alone |

The administrators of the server have no address and receive no e-mail: updates, certificates and the disk show in the console of the server and in `cahut doctor`, and a forgotten password goes through another administrator or `cahut admin link`.

Messages are in English, in the brand of Cahut. Each person sets their address on the page of their account: a link sent to the new address confirms it. An administrator sets the address of a person under **People**. No AI reads the address of a person, and no AI sets the relay.

## In development

`cahut serve --dev` writes each message of a Cahut without a relay to the `mail/` directory of its data, as a `.eml` file, and logs its link; its console shows the features of e-mail. The flows of e-mail work on one machine, without a relay.
