---
title: The console
description: Sign in to the web console of a Cahut, under /console/ of its domain, and what each person and each administrator finds there.
---

Each Cahut has a web console, under `/console/` of its domain: `https://cahut.example.com/console/`. `/admin` leads there too. People sign in there, see their Rooms and Windows, connect their AIs and read the Logbook; administrators manage people, Rooms and the server. The console uses the same [API](/reference/http-api/) as the AIs, and nothing else.

Port 443 keeps the Windows and MCP. Some company networks let only port 443 out: from there, the console is out of reach, the Windows are not.

## Sign in for the first time

Nobody chooses a password for someone else. An administrator creates the account, and Cahut gives a link that works once, for 7 days:

```text
https://cahut.example.com/console/activate#...
```

The person opens it, chooses a password of 15 characters at least, and is signed in. The code of the link stays in the browser: it never reaches the server's logs, and the console takes it out of the address once read.

An administrator creates accounts in the console, under **People**, or on the command line with [`cahut user create`](/admin/commands/#create-a-person). A person who forgot their password gets a new link the same way: **New link** in People, or `cahut user link`. Their password works until they use the new link, and using it ends their other sessions.

## What a person finds

| Screen | What it shows |
| --- | --- |
| Overview | The latest Windows of your Rooms, and your Rooms |
| Knocks | What your AIs asked that waits for you, with Approve and Decline, then what you answered and what expired; the navigation counts those that wait |
| A Knock | Where the link of an AI leads: what it asks, the details of the call, when it expires, Approve and Decline, then what the call gave back |
| A Room | Its Windows, each with its access, its trash, and its people; owners add and remove people, rename the Room, and delete a team Room for good once its name is typed |
| A Window | Where it opens; its access, its end date and its Key, which shows once, next to the address, ready to copy; its two versions, with a button that goes back; what its AI declared it does; the trash, and deleting it for good from there |
| Rooms | A new team Room, of which you become the owner |
| Logbook | What you did, what your AIs did, and what happened in your Rooms, by day, with filters |
| AIs and access | The address to give an AI client; your connected AIs, with what they may do and where, Change and Disconnect; your personal tokens: create one, shown once with the command that connects Claude Code, and revoke it in one click |
| Account | A new password, and signing out |

A change to what someone may do, such as adding a person to a Room, making a Window public, and deleting anything for good, ask you to confirm first. When your AI asks for one of these changes, it knocks first: the change waits for you under **Knocks**. See [Knock first](/concepts/knock-first-and-logbook/#knock-first).

## What an administrator of the Cahut finds

The same screens, and:

- **People**: a new person with their link, a new link, making someone an administrator or a member, disabling and enabling. A disabled person cannot sign in, and their AIs stop at once.
- **Rooms**: every Room of the Cahut, whose people they manage. They open the Windows of the Rooms they are in only.
- **Logbook**: every entry of the Cahut, from people and from AIs.
- **AIs and access**: the connected AIs and the tokens of others too, which they can disconnect and revoke.
- **House style**: see below.

### House style

The administrators set the [house style](/concepts/house-style/) of the Cahut here: the name of the organization, its logo and its version for dark backgrounds, its four colors with their contrasts, the colors of its charts, its fonts and their files, and its guidance for AIs. A preview shows the form in its colors and fonts before saving, and the three templates open in the house style that is saved. **Go back to it** restores the previous version, and **Back to neutral** starts again from the neutral house style, both after a confirmation.

The **AIs and access** screen of every person gives the commands that install the [Cahut plugin](/connect/claude-plugin/) in Claude Code, and the skill as a zip for the Claude apps.

## Administrators of the server

The administrators of the server have accounts of their own, apart from the people of the Cahuts: see [Administrators of the server](/admin/commands/#administrators-of-the-server). They sign in at **Administer the server**, under the sign-in form of any Cahut, or at `/server/sign-in`. Their first link opens at `/server/activate#...`.

They find:

- **Overview**: the Cahuts of the server, and a form that declares one with its first administrator, whose link the console shows.
- **Administrators**: the list, a new administrator with their link, and a new link.
- **Logbook of the server**: the Cahuts declared, the administrators and their sign-ins.
- In the Cahut of the domain where they signed in, **People**, **Rooms** and its **Logbook**, as its administrators have them, without its Windows.

A session holds at the domain where it opened: to look after another Cahut, sign in at its own console.

## Sessions

A session ends 7 days after its last use, and 30 days after it opened in any case. The console keeps it in the storage of its own address and sends it with every call; it sets no cookie. After 10 refused sign-ins in 15 minutes, a handle must wait; an address that tries too often must wait too.

## Readers

To open a Window that is not public, a reader signs in through the console: the link on the page of a closed Window sends them to `/console/read`, which signs them in if needed, then sends them back to the Window. See [Who can open a Window](/concepts/access/#readers-who-sign-in).

## Consent

When an AI client connects with OAuth, it sends you to `/console/oauth/authorize`. You sign in there if you need to, and choose what the AI may do and in which Rooms. See [Connect an AI client](/connect/ai-clients/#consent).
