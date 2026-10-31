---
title: Moderation
description: How anyone reports a Window of a Cahut, and how its moderators read the reports, open the Window, withdraw it and reinstate it.
---

Anyone reports a Window of a Cahut to its moderators, without an account. The moderators of a Cahut are its administrators and the administrators of its server. They read the reports in the console of the Cahut, or of the server, and decide.

## Reporting a Window

Every Cahut serves a page where anyone reports one of its Windows, at `https://<domain>/-/report`. The foot of the pages that Cahut serves to readers links it as **Report a Window**: the page of a Room, the page that asks for a Key, the page of a share link, and the page of an address where nothing is found. The page that asks for a Key fills in the address of its Window. On the page of a Room, each Window links **Report** too, with its address filled in, for the readers outside the Room.

Cahut draws nothing in a Window or around it: the reader of a public Window types or pastes its address. A host who opens Cahuts to the public says in its legal notice where to report too.

The page asks for:

- **The address of the Window**, as its reader meets it: its address, `https://<domain>/<room>/<window>/`, or a file under it; its permanent address, `https://<domain>/-/w/<id>`; a share link; or the address of a Room that shows a Window at its own.
- **What is wrong**: phishing or a scam, malware, illegal content, someone's rights, such as copyright or privacy, or something else.
- **An explanation**, of 10 to 2,000 characters.
- **A name and an e-mail address**, both optional.
- **That the report is made in good faith**, and that what it says is true.

The page then says that the report was received, and nothing about the Window. Reports are limited by network address: after 20 in a row, one every 3 minutes.

## Reading the reports

The administrators of a Cahut find its reports under **Moderation**, in its console, and the administrators of its server on the page of the Cahut in the [console of the server](/admin/console/#the-console-of-the-server). The menu of the console of the Cahut counts those that wait, and so does the list of the Cahuts of the server, in the console of the server and in [`cahut domain list`](/admin/commands/#declare-a-cahut). Each report gives its number, what is wrong, the Window it names with its Room and its access, the explanation, and the name and address of whoever reported, if they gave them.

## Opening a reported Window

**Open**, on a report that waits, shows its Window in the console, whatever its access: open to its Room, to its author alone, with a Key or with a share link. It runs in its sandbox, as every Window does, without its forms. The frame opens that Window only, for four hours at most, and stops when the session of the moderator ends, when they no longer administer the Cahut, or when the report is decided.

Each opening goes to the Logbook in the Room of the Window, under the moderator and with the number of the report, so that the members of the Room read who opened it and why; for a Window that its author keeps for themself, the author reads it. Without a report, a moderator opens what anyone opens: a public Window. Under **Public Windows**, the page lists those of the Cahut, the latest made public first, with who published them, so that moderators look at what the Cahut serves to anyone without waiting for a report: each opens at its address, and **Withdraw** withdraws it.

## Deciding

**Leave it as it is** decides a report and leaves its Window as it is, with a note for the moderators if they want one.

**Withdraw** withdraws its Window, with a reason of 500 characters at most, which the members of its Room read. The Window keeps everything, and serves nothing to its readers until it is reinstated: see [A withdrawn Window](/concepts/rooms-and-windows/#a-withdrawn-window). Withdrawing decides every report that waits about the Window. A moderator also withdraws a public Window that nobody reported, with `withdraw_window`; withdrawing a withdrawn Window changes its reason.

Cahut then forgets the name and the address of whoever reported. Decided reports stay on the page, under **Decided**, with who decided and when.

## Reinstating

**Reinstate**, on a decided report whose Window is still withdrawn, gives the Window back as it was: its access, its Key, its share links until their end. A Window that the administrators of the server withdrew is theirs: only they change its reason or reinstate it, from the console of the server; the console of the Cahut shows it withdrawn by the server, without **Reinstate**.

For a person who publishes abuse again and again, disable them or delete them for good under **People**, or have the administrators of the server [suspend the Cahut](/concepts/cahuts/#suspension). Disabling a person leaves their public Windows open: withdraw them too.

Moderation is for people: no AI and no service is offered its operations, [`report_window`, `list_reports`, `list_public_windows`, `preview_reported_window`, `dismiss_report`, `withdraw_window` and `reinstate_window`](/reference/operations/). The administrators of the server set the address of moderation with `set_moderation_address`.

## E-mail

When the Cahut sends e-mail, through its [relay](/admin/mail/):

- **Each report** goes to its administrators who have an address, and to its **address of moderation**, which the administrators of the server set on the page of the Cahut in the [console of the server](/admin/console/#the-console-of-the-server), or with [`cahut settings moderation-address`](/admin/commands/#set-the-address-of-moderation-of-a-cahut): an abuse mailbox of the host, since they have no address of their own. The message gives the number of the report, what is wrong and the address of the Window, and links **Moderation**. The explanation stays in the console.
- **Whoever reported** with an address gets a receipt, then the decision: the Window was withdrawn, or the moderators left it as it is. Cahut then forgets their address.
- **The members of the Room** of a Window, or its author for a Window that is theirs alone, hear that it was withdrawn, with the reason and the link to the support page of the Cahut when it has one, then each new reason that a moderator gives, and that it was reinstated.

These messages name the Window by its address, and hold nothing that whoever reported wrote: the page of a report cannot send a text of one's choice to someone else. Without a relay, the Cahut sends none of them.

## What the Logbook keeps

Each report goes to the Logbook of the Cahut, under someone who reports a Window, with its number, what is wrong and the Window it names. Only the administrators of the Cahut read it: the members of the Room of the Window do not. The entry holds neither the explanation, nor the name, nor the address of whoever reported, nor their address on the network. Each opening of a reported Window, each withdrawal with its reason and each reinstatement go to the Logbook of its Room, so that its members read them, under the moderator who made them. A report left as it is goes to the Logbook of the Cahut.

The record of these decisions is [ADR 54](https://github.com/cahutcloud/cahut/blob/main/docs/adr/0054-reporting-and-withdrawing-a-window.md).
