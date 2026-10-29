---
title: Share a Window
description: The menu Share of the console gives the permanent address of a Window, to copy, to send by e-mail or to post on X, LinkedIn or Facebook. Pasted in Slack, Teams, iMessage or elsewhere, it shows a preview with a card that the Cahut draws, and that says no more than the address says to a stranger.
sidebar:
  label: Sharing
---

People share a Window, not the Cahut. The link to give is its permanent address, `https://cahut.example.com/-/w/<id>`: it opens the Window to whoever may open it, and keeps working when the Window changes its name or its author changes their handle. See [The id and the permanent address](/concepts/rooms-and-windows/#the-id-and-the-permanent-address).

## The menu Share

In the console, **Share**, at the head of the page of a Window and on each card of a Window, opens a menu:

| Entry | What it does |
| --- | --- |
| Copy the link | Copies the permanent address |
| E-mail | Opens your mail app with a message: the title of the Window as its subject, its description and the link |
| X, LinkedIn, Facebook | Open their own page of sharing in a new tab, with the link; for a public Window or a Window with a Key only |
| Share… | The share sheet of your system, where the browser has one, as on phones |

The menu says who opens the link: anyone, anyone with its Key, the people of the Cahut once signed in, or the members of its Room. X, LinkedIn and Facebook do not show for a Window open to its Room or its Cahut: their readers could not open it.

Its entries are plain links. The console loads no script of X, LinkedIn or Facebook, and they learn of the Window only when you post it. The link of a Window with a Key does not carry its Key: give it apart, or [invite people by e-mail](/concepts/access/#invitations), each with a personal link. Sharing goes to no Logbook: it changes nothing in the Cahut.

## The preview of a link

Slack, Teams, Discord, iMessage, WhatsApp, Signal, X, LinkedIn, Facebook and Mastodon show a preview under a link: a robot fetches the address, without a session or a Key, and reads the tags of [Open Graph](https://ogp.me) in the page it gets. At the permanent address of a Window, Cahut answers these robots with a page of its own, whose preview says what the address says to a stranger:

| Access of the Window | Title | Description | Card |
| --- | --- | --- | --- |
| Public | Its title | Its description, when it has one | With its title |
| Key | Its title, which the page of its Key already shows | Opens with its Key. | With its title |
| Room only, Internal | A Window at `cahut.example.com` | Sign in to open it. | Without a title |

The title and the description of a Window open to its Room or its Cahut stay out of the preview, and out of the servers of the platforms that draw it. The access that counts is the one of the moment, for the next robot that comes. A preview already drawn stays as it was in its messages: a title shared while the Window was public stays there.

The page of a robot opens nothing: it sets no cookie and gives no access. People who follow the same link get the Window, the page of its Key, or the page of what is not found, as before. Search engines are not robots of previews: they follow the link as people do.

The robots are known by the words of their `User-Agent`. A robot that Cahut does not know yet gets what a person without a session gets: the tags that the author of a public Window wrote in its page, and nothing for the others.

## The card

The picture of a preview is a card in the brand of Cahut, 1200 by 630 pixels, that each Cahut draws itself: its logo, the title of the Window, and the domain of the Cahut. A long title gets smaller, then ends with an ellipsis. The card has no title for a Window open to its Room or its Cahut, nor when the title is in a script that the font of the brand does not draw, such as Greek, Cyrillic, Arabic, Hebrew, Chinese or Japanese: the preview still gives the title in text, in the fonts of the platform. Emoji are left out of the card.

The card is at `https://cahut.example.com/-/card/<id>.png`. A new title gives it a new address in the preview, so that platforms that keep pictures take the new one.

## The address of the moment

The address of the moment, `https://cahut.example.com/<room>/<window>/`, does not change: it serves a public Window as its AI made it, with the tags of its page if it wrote some, and the page of its Key or of what is not found to the others. A link to it shows the preview that its page gives, or none.

The decision is [ADR 51](https://github.com/getcahut/cahut/blob/main/docs/adr/0051-sharing-a-window-and-its-preview.md).
