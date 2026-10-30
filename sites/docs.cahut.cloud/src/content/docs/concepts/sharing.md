---
title: Share a Window
description: The menu Share of the console gives the permanent address of a Window, to copy, to send by e-mail or to post on X, LinkedIn or Facebook, and makes share links, which open a Window to whoever holds them, without an account or a Key, until they end. Pasted in Slack, Teams, iMessage or elsewhere, a link shows a preview with a card that the Cahut draws.
sidebar:
  label: Sharing
---

People share a Window, not the Cahut. The link to give is its permanent address, `https://cahut.example.com/-/w/<id>`: it opens the Window to whoever may open it, and keeps working when the Window changes its name or its author changes their handle. See [The id and the permanent address](/concepts/rooms-and-windows/#the-id-and-the-permanent-address).

## The menu Share

In the console, **Share**, at the head of the page of a Window and on each card of a Window, opens a menu:

| Entry | What it does |
| --- | --- |
| Copy the link | Copies the permanent address |
| Create a share link… | Makes a [share link](#share-links), which opens the Window without an account or a Key; not for a public Window |
| Share once… | Makes a link that opens the Window [once](#share-once), and keeps it for you alone; for its author, and not for a public Window |
| E-mail | Opens your mail app with a message: the title of the Window as its subject, its description and the link |
| X, LinkedIn, Facebook | Open their own page of sharing in a new tab, with the link; for a public Window or a Window with a Key only |
| Share… | The share sheet of your system, where the browser has one, as on phones |

The menu says who opens the link: anyone, anyone with its Key, the people of the Cahut once signed in, or the members of its Room. X, LinkedIn and Facebook do not show for a Window open to its Room or its Cahut: their readers could not open it.

Its entries are plain links. The console loads no script of X, LinkedIn or Facebook, and they learn of the Window only when you post it. The link of a Window with a Key does not carry its Key: give it apart, or [invite people by e-mail](/concepts/access/#invitations), each with a personal link. Sharing goes to no Logbook: it changes nothing in the Cahut.

## Share links

A share link opens a Window to whoever holds it, without an account or a Key, until it ends: a link to paste in a message, in Slack or in a document, for a Window that stays open to its Room or its Cahut. It looks like `https://cahut.example.com/-/s/<token>`, with a token of 52 characters that nobody can guess.

**Create a share link…**, in the menu Share, or **Create a share link** on the page of the Window, asks for:

| Field | What it sets |
| --- | --- |
| Label | A few words that tell your links apart, such as *For Acme*, which only the members of the Room see; optional |
| It opens the Window | For a day, 7 days, 30 days, or until a day you choose, within 30 days; 7 days by default |
| Readers | Any number, or a number of readers from 1 to 1000 |

The console then shows the link, once: Cahut keeps only a trace of it, as of a Key, and nobody can read it again. Make a link for each sharing, so that you can take one back without the others.

- **A link adds to the access.** The Window keeps its access, and opens besides to whoever holds the link. Its page in the console says how many links open it. A public Window takes no link: its permanent address opens it to anyone.
- **Opening is a click.** The link opens a page of the Cahut that names the Window and says until when the link opens it, with a button, **Open**. The visit counts nothing: the robots of previews and the filters of mail that follow links spend nothing. The click spends one opening, and gives the reader a pass, as the right Key does: the Window, its pictures, its styles and its scripts then open until the link ends. A reader who opens the Window already, a member signed in or a reader who clicked before, goes straight to it.
- **A number of readers.** Once the openings are spent, the link opens nobody new, and those who opened it read on until it ends.
- **Taking a link back.** The page of the Window lists its links: their label, who made them, their end and their openings. **Take back** ends a link and closes the Window at once to those who opened it. A link that ends does the same.
- **The Logbook** records each link made, each opening, as *Someone with a share link*, and each link taken back. It keeps neither the reader, nor their address, nor their browser.
- **The life of the Window.** A link follows the Window when it changes its name. In the trash, the Window opens to no link; restored, it opens again to those that have not ended. Deleted for good, it takes its links with it. A new Key, or another access, does not end them.
- **Bounds.** A Window holds 20 links that stand, and a person makes 100 links a day, their AIs included.

An AI makes a link with `create_share_link` without [knocking first](/concepts/knock-first-and-logbook/#knock-first), as it sets a Key: it gets the link once, to give to you. `get_window` lists the links of a Window, and `revoke_share_link` takes one back.

## Share once

A Window can open once: its link works for the first reader who opens it, and for nobody after. It is a [Window for its author alone](/concepts/access/#only-you) with a share link that opens once: the Room never sees it, the reader reads it, and you keep it after.

**Share once…**, in the menu Share of a Window you are the author of, asks for a label and an end, as for any link, and keeps the Window for you alone first, unless you uncheck **Keep it for me alone**. The console then shows the link, once.

- **The page of the link says it opens once**, and that its reader reads the Window in their browser for 4 hours. Robots of previews and filters of mail that only follow the link spend nothing; its preview says *Opens once, until* its end.
- **The click** spends the opening and gives a pass for a session of reading: the browser drops it when it closes, and Cahut ends it 4 hours after the click. Until then, the reader reloads the Window, follows its pages and comes back to the link. After that, they get a page that says the link opens nothing.
- **After the opening**, only the link dies. The Window stays, for you alone, with the link shown as opened, and its opening in the Logbook. Make another link for another reader, open the Window to the Room, or put it in the trash.

The same holds for every link that opens once, to any Window: its pass lasts 4 hours. An AI makes such a Window in two calls: `publish_window` with the access `author`, then `create_share_link` with `max_openings` 1.

Once is about the link, not about what the reader does: in those 4 hours, they may save, print or copy what they read.

## The preview of a link

Slack, Teams, Discord, iMessage, WhatsApp, Signal, X, LinkedIn, Facebook and Mastodon show a preview under a link: a robot fetches the address, without a session or a Key, and reads the tags of [Open Graph](https://ogp.me) in the page it gets. At the permanent address of a Window, Cahut answers these robots with a page of its own, whose preview says what the address says to a stranger:

| Access of the Window | Title | Description | Card |
| --- | --- | --- | --- |
| Public | Its title | Its description, when it has one | With its title |
| Key | Its title, which the page of its Key already shows | Opens with its Key. | With its title |
| Only you, Room only, Internal | A Window at `cahut.example.com` | Sign in to open it. | Without a title |

The title and the description of a Window open to its Room or its Cahut stay out of the preview, and out of the servers of the platforms that draw it. The access that counts is the one of the moment, for the next robot that comes. A preview already drawn stays as it was in its messages: a title shared while the Window was public stays there.

The page of a robot opens nothing: it sets no cookie and gives no access. People who follow the same link get the Window, the page of its Key, or the page of what is not found, as before. Search engines are not robots of previews: they follow the link as people do.

The robots are known by the words of their `User-Agent`. A robot that Cahut does not know yet gets what a person without a session gets: the tags that the author of a public Window wrote in its page, and nothing for the others.

A share link gives its preview on its own page, which every visitor gets, robots and people, and which opens nothing. Since the link opens the Window to whoever holds it, its preview shows the title of the Window, `Opens with this link until <date>.`, or `Opens once, until <date>.`, and the card of the Window. A link that has ended shows no preview.

## The card

The picture of a preview is a card in the brand of Cahut, 1200 by 630 pixels, that each Cahut draws itself: its logo, the title of the Window, and the domain of the Cahut. A long title gets smaller, then ends with an ellipsis. The card has no title for a Window open to its Room or its Cahut, nor when the title is in a script that the font of the brand does not draw, such as Greek, Cyrillic, Arabic, Hebrew, Chinese or Japanese: the preview still gives the title in text, in the fonts of the platform. Emoji are left out of the card.

The card is at `https://cahut.example.com/-/card/<id>.png`. A new title gives it a new address in the preview, so that platforms that keep pictures take the new one.

## The address of the moment

The address of the moment, `https://cahut.example.com/<room>/<window>/`, does not change: it serves a public Window as its AI made it, with the tags of its page if it wrote some, and the page of its Key or of what is not found to the others. A link to it shows the preview that its page gives, or none.

The decisions are [ADR 51](https://github.com/cahutcloud/cahut/blob/main/docs/adr/0051-sharing-a-window-and-its-preview.md), for the menu and the preview, [ADR 52](https://github.com/cahutcloud/cahut/blob/main/docs/adr/0052-share-links.md), for share links, and [ADR 53](https://github.com/cahutcloud/cahut/blob/main/docs/adr/0053-windows-for-their-author-alone.md), for the Windows that open to their author alone, and once to a single reader.
