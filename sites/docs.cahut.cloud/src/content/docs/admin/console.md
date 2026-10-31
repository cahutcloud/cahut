---
title: The console
description: Sign in to the web console of a Cahut, under /console/ of its domain, and to that of the server, under /server/, and what each person and each administrator finds there.
---

Each Cahut has a web console, under `/console/` of its domain: `https://cahut.example.com/console/`. The root of the domain, `https://cahut.example.com/`, and `/admin` lead there too: the console opens on its overview once you have signed in, and on its page of sign-in until then. People sign in there, see their Rooms and Windows, connect their AIs and read the Logbook; its administrators manage its people and its Rooms. The administrators of the server have [a console of their own](#the-console-of-the-server), under `/server/`. The console uses the same [API](/reference/http-api/) as the AIs, and nothing else.

Port 443 keeps the Windows and MCP. Some company networks let only port 443 out: from there, the console is out of reach, the Windows are not.

## Sign in for the first time

The first person of a Cahut starts from the link that the installation prints, at `/console/setup#...`: see [The first start](/install/first-start/). After them, nobody chooses a password for someone else. An administrator creates the account, and Cahut gives a link that works once, for 7 days:

```text
https://cahut.example.com/console/activate#...
```

The person opens it, chooses a password of 15 characters at least, and is signed in. The code of the link stays in the browser: it never reaches the server's logs, and the console takes it out of the address once read.

An administrator creates accounts in the console, under **People**, or on the command line with [`cahut user create`](/admin/commands/#create-a-person). When the Cahut [sends e-mail](/admin/mail/) and the person has an address, the link goes there too. In a Cahut that sends e-mail, a person who forgot their password asks for a link from **Forgot your password?**, on the page of sign-in: it goes to the address of their account and works for one hour. Without an address, or without e-mail, an administrator gives a new link: **Manage**, then **Give a new link**, in People, or `cahut user link`. The password works until the new link is used, and using it ends the other sessions.

## What a person finds

| Screen | What it shows |
| --- | --- |
| Overview | The latest Windows of your Rooms, each with a preview and a button **Share**, your Rooms, and the limits of the Cahut, or yours in an individual Cahut, when there are some |
| Search | From the field at the head of every page, or of the menu on a narrow screen: the Windows of your Rooms whose title, description, tags or pages hold your words, the best first, each with the words of its pages around them, and their tags, which narrow them. See [Search](/concepts/rooms-and-windows/#search) |
| Knocks | What your AIs asked that waits for you, with Approve and Decline, then what you answered and what expired; the navigation counts those that wait |
| A Knock | Where the link of an AI leads: what it asks, the details of the call, when it expires, Approve and Decline, then what the call gave back |
| A Room | Its address, ready to copy, and what it shows; its Windows, each with a preview, its tags, its access and a button **Share**, the one its address shows marked, the tags over them, which show only the Windows with one, and its trash folded under them, where a Window comes back with **Restore**; the people of a team Room. Owners add a person with **Add a person** and change their role or remove them with **Manage**; **Manage the Room** chooses what its address shows, its Windows or one of them, renames a team Room, changes its description, leaves it, and deletes it for good once its name is typed |
| A Window | **Share**, which copies its permanent address or sends it by e-mail, to X, LinkedIn or Facebook, and creates [share links](/concepts/sharing/#share-links), or a link that opens it [once](/concepts/sharing/#share-once) for its author; its title, its description and its tags, which change without a new version; where it opens, at its address and at its permanent address, and its name in the address, which changes after a warning; its access, Only you for its author, its end date and its Key, which shows once, next to the address, ready to copy; its share links, with **Take back**; its two versions, with a button that goes back; what its AI declared it does; the trash, and deleting it for good from there |
| Rooms | A new team Room, of which you become the owner |
| Logbook | What you did, what your AIs did, and what happened in your Rooms, by day, with filters |
| AIs and access | The address to give an AI client; your connected AIs, with what they may do and where, Change and Disconnect; your personal tokens: create one, shown once with the command that connects Claude Code, and revoke it in one click |
| Account | At the foot of the navigation, and behind your name in the top bar: your name and your handle; your e-mail address and notices, when the Cahut sends e-mail; the language of the console; a new password, and signing out |

The **AIs and access** screen also gives the commands that install the [Cahut plugin](/connect/claude-plugin/) in Claude Code, and the skill as a zip for the Claude apps.

A change to what someone may do, such as adding a person to a Room, making a Window public, and deleting anything for good, ask you to confirm first. When your AI asks for one of these changes, it knocks first: the change waits for you under **Knocks**. See [Knock first](/concepts/knock-first-and-logbook/#knock-first).

## What an administrator of the Cahut finds

The same screens, and:

- **People**: everyone in the Cahut, with their role and state, and their e-mail address when the Cahut sends e-mail. **New person** adds someone and gives their link. **Manage** opens what can change for a person, each with what it does: their e-mail address, making them an administrator or a member, a new link, disabling and enabling. A change of role and disabling ask to confirm first. A disabled person cannot sign in, and their AIs stop at once.
- **Rooms**: every Room of the Cahut, whose people they manage. They open the Windows of the Rooms they are in only.
- **Logbook**: every entry of the Cahut, from people and from AIs.
- **AIs and access**: the connected AIs and the tokens of others too, which they can disconnect and revoke.
- **Kind and limits**: whether the Cahut is collaborative or individual, its [limits](/concepts/cahuts/#limits) with what it holds, and the links to its [legal pages](/concepts/cahuts/#legal-pages), which they read: the administrators of the server set them, in [their console](#the-console-of-the-server).
- **Moderation**: the reports that anyone makes about the Windows of the Cahut, those that wait first, which the menu counts, and its public Windows; a reported Window opens there whatever its access, and is withdrawn or left as it is. See [Moderation](/admin/moderation/).

In an [individual Cahut](/concepts/cahuts/#collaborative-or-individual), the page of a Room no longer adds or removes people, and the access of a Window no longer offers Internal.

## The console of the server

The administrators of the server have accounts of their own, apart from the people of the Cahuts: see [Administrators of the server](/admin/commands/#administrators-of-the-server). They have a console of their own too, under `/server/` of any domain of the server: `https://cahut.example.com/server/`. Its top bar is Ink, Paper in the dark theme, and says **Server**. The console of a Cahut does not link to it, so that its people are not shown the way to the server. Its first link opens at `/server/activate#...`, and its administrators sign in at `/server/sign-in`.

From the domain where they signed in, they look after every Cahut of the server, without signing in again. They find:

- **Cahuts**: the Cahuts of the server, with their kind, how many reports wait in each, and their suspension, since when and why; a form that declares one with its first administrator, whose link the console shows; and on each Cahut **Move to a domain** or **Change its domain**. When the Cahut of the domain where you signed in moves, the console sends you to sign in again at its new address.
- **The page of a Cahut**, from that list:
  - its kind and its [limits](/concepts/cahuts/#limits), with what it holds: a field left empty lifts a limit, and the disk counts in gigabytes of a billion bytes;
  - the addresses of its [legal pages](/concepts/cahuts/#legal-pages), each with another address in French where the page in French differs; a field left empty removes its link;
  - its address of moderation;
  - its [suspension](/concepts/cahuts/#suspension), with a reason, which they change, and its resumption;
  - its [moderation](/admin/moderation/): the reports that wait, its public Windows, the preview of a reported Window, its withdrawal and its reinstatement;
  - its service tokens, which a program on the server uses for the [admin API](/reference/http-api/#admin-api): **Create a token** shows it once, and **Revoke** ends its access at once;
  - while the Cahut has nobody, a link of [first start](/install/first-start/).

  Each change goes to the Logbook of the Cahut.
- **Administrators**: the list, with the day each was added and their state. **New administrator** adds one and gives their link. **Manage** changes the name of an administrator, theirs or one's own, gives another administrator a new link, or disables and enables them; disabling asks to confirm first. A disabled administrator loses their sessions and cannot sign in; the server keeps at least one who is not disabled.
- **Logbook**: the Cahuts declared and moved, the administrators and their sign-ins, and the updates.
- **Updates**: the version running, the newest release, the last update and how it ended, with **Check now** and **Update**. See [Updates](/admin/updates/).
- **Account**: the language of the console, the password of the server, and signing out of the server.

The people, the Rooms, the e-mail and the Logbook of a Cahut are for its own administrators: an administrator of the server who looks after them needs an account of a person there, as one of its administrators. Someone alone on their server has both accounts, which the first start makes with one password.

## Sessions

A session ends 7 days after its last use, and 30 days after it opened in any case. The console keeps it in the storage of its own address and sends it with every call; it sets no cookie. A browser holds the session of a person, for the console of the Cahut, and that of an administrator of the server, for the console of the server: each console signs out of its own, and a session that ends asks to sign in again with its own account. Signing out also ends the reader sessions that the console opened in that browser, and **Sign out** on the pages of the Cahut ends the session of the console that opened them.

After 10 refused sign-ins to one handle from one address in 15 minutes, that address must wait; after 100 from everywhere, the handle waits everywhere. An address that tries too often must wait too, and an IPv6 address counts with its whole /64. The Logbook keeps the address of every sign-in.

On a Knock, **Approve** and, on the consent screen of an AI, **Connect** become active once the page has been in front for a moment and you have moved the pointer or pressed a key: a page that would slip them under a double-click gets nowhere.

## Name and handle

Anyone changes their name and their handle under **Account**. Their personal Room takes the new name as its title, unless they gave it another one.

A new handle must be free, held by no person and no Room. It becomes the name of the personal Room and, in an [individual Cahut](/concepts/cahuts/#collaborative-or-individual), the start of the name of each Room of the person, such as `bob-projects`: the addresses of their Windows change, and the old ones stop working, the links already given out included. The old handle is free for someone else. The console shows the old and the new addresses, and asks again before the change. Answer the Knocks that wait for you first, since they name the old addresses. Your sessions, tokens and connected AIs go on as before, and you sign in with the new handle.

## Language

The console speaks English and French. It follows the languages of the browser, and anyone can choose one at the foot of the navigation, on every page, under **Account**, or under the card of the page of sign-in: the choice stays in that browser, for that Cahut. The words of the house, such as Room, Window and Knock, stay in English, and so do the Logbook, e-mail and most messages of the server. The refusals that people meet most, such as a wrong password, a link that expired, a session that ended, a handle or a name of a Room already taken, a Key too short or an invitation that cannot go, come in the language of the console.

## Readers

To open a Window that is not public, a reader signs in through the console: the link on the page of a closed Window sends them to `/console/read`, which signs them in if needed, then sends them back to the Window. See [Who can open a Window](/concepts/access/#readers-who-sign-in).

## Consent

When an AI client connects with OAuth, it sends you to `/console/oauth/authorize`. You sign in there if you need to, and choose what the AI may do and in which Rooms. See [Connect an AI client](/connect/ai-clients/#consent).
