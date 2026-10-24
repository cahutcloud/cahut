---
title: E-mail
description: How a Cahut sends e-mail through an SMTP relay, its own or that of its server, what it sends, and how to set and test a relay.
---

A Cahut sends e-mail through an SMTP relay: the links where people choose their password, invitations to read a Window, and notices. It never delivers messages to the servers of their recipients itself. Without a relay, a Cahut works as before: administrators pass links on by hand, and nothing else goes out.

## Two relays

| Relay | Who sets it | What it sends |
| --- | --- | --- |
| The relay of the server | Its administrators, in the console under **E-mail of the server**, or with `cahut mail set` | The e-mail of the server, and that of every Cahut without a relay of its own |
| The relay of a Cahut | Its administrators, in the console under **E-mail**, or with `cahut mail set --domain` | The e-mail of that Cahut, from an address of its own |

The relay of a Cahut speaks TLS, from the start or after STARTTLS, and is on the internet: Cahut refuses a relay of a Cahut at a loopback, private or reserved address. A relay on the server or on its network is the relay of the server, which may also do without TLS on the loopback address of the machine, where the message does not leave it.

Cahut connects to a relay and signs in before it keeps it. The answer of the relay tells you what is wrong: a host that does not answer, a relay that offers no STARTTLS, a refused password.

## Set a relay

In the console, the page **E-mail** asks for what your provider of e-mail gives:

| Field | What to give |
| --- | --- |
| Host | The name of the relay, such as `smtp.example.com`, without a scheme or a port |
| Port | 587 for STARTTLS, 465 for TLS from the start |
| Security | STARTTLS or TLS; for the relay of the server on this machine, none |
| User name and password | Those of the account at the relay; empty when it takes no authentication |
| From | An address that the relay accepts, such as `cahut@example.com` |
| Name | The name messages come from; the name of the Cahut when empty |

On the command line, the password comes on the standard input, never as an argument:

```bash
printf '%s' "$SMTP_PASSWORD" | sudo -u cahut /usr/local/bin/cahut mail set --host smtp.example.com --port 587 \
  --user cahut@example.com --password-stdin --from cahut@example.com
```

```text
The relay answered, and sends the e-mail of the server.

host  smtp.example.com:587, starttls
user  cahut@example.com, with a password
from  cahut@example.com
set   2026-10-24 14:05
```

`--domain cahut.example.com` sets the relay of that Cahut instead. A Postfix on the server itself, which takes no authentication, is a relay of the server: `--host 127.0.0.1 --port 25 --security none`.

Setting the relay again without a password keeps the password, as long as the host and the user name do not change. `cahut mail show` prints the relay, `cahut mail remove` removes it: a Cahut without a relay then sends through that of the server.

## The password of a relay

Cahut keeps the password of a relay sealed with AES-256-GCM, under a key of the server in `mail.key`, in the data directory beside `server.db`, readable by the service only. The password is never shown again, and never goes to the Logbook. A copy of the directory of a Cahut, a backup for instance, does not carry it: a Cahut moved to another server needs its password typed again.

## Test

**Send a test to your address**, on the page **E-mail**, sends a message to the address of your account, and says whether the relay took it. On the command line, name the address:

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
| A limit of the Cahut at 90% or reached | Its administrators, and those of the server |
| A failed update, a release that waits, a certificate that cannot be renewed, a disk 90% full | The administrators of the server |

Messages are in English, in the brand of Cahut. Each person sets their address on the page of their account: a link sent to the new address confirms it. An administrator sets the address of a person under **People**. No AI reads the address of a person, and no AI sets the relay.

## In development

`cahut serve --dev` writes each message that no relay sends to the `mail/` directory of its data, as a `.eml` file, and logs its link: the flows of e-mail work on one machine, without a relay.
