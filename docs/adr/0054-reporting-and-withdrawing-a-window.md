# 54. Anyone reports a Window, and the moderators of its Cahut open it and withdraw it

- Status: Accepted, extended by [ADR 56](0056-a-database-for-each-window.md) (a withdrawn Window takes no data from readers, and moderators write none)
- Date: 2026-10-30

## Context

A host opens Cahuts to the public, such as an individual Cahut where people sign up on a site of the host ([ADR 30](0030-kind-and-limits-of-each-cahut.md), [ADR 35](0035-a-program-on-the-server-manages-people-with-a-service-token.md)). Each person publishes Windows, makes one public in the console at once, or gives it to anyone with a share link ([ADR 19](0019-window-versions-access-and-declaration.md), [ADR 52](0052-share-links.md)). A service that serves pages to anyone draws abuse: pages that steal passwords, malware, illegal content.

The host can suspend a whole Cahut ([ADR 48](0048-suspending-a-cahut.md)), disable a person, or delete a person with the Rooms that are theirs alone ([ADR 47](0047-deleting-a-person.md)). Nothing stops one Window and leaves the rest. Nobody can tell the host about a Window either: the pages of a Cahut link its legal pages and its support ([ADR 42](0042-the-legal-pages-of-a-cahut.md)), but no page says how to report one, and the host learns nothing from Cahut.

The host could not look at it anyway. By [ADR 13](0013-people-sessions-and-rights.md), the administrators of a Cahut and of its server manage people and Rooms, not content: they open only the Windows of the Rooms they belong to. A public Window opens to anyone, them included; a Window with a Key or a share link does not.

In the European Union, the Digital Services Act asks every hosting service for a way to notify illegal content by electronic means, a receipt and a decision for whoever notifies and gives an address, and a statement of reasons for the person whose content is restricted (Articles 16 and 17).

An organization has the same need inside its own Cahut: a member publishes what the organization must not serve, or a Window leaks what it should not.

## Decision

### The moderators

- **The moderators of a Cahut are its administrators and the administrators of its server.** Both read the reports of the Cahut, open the Windows that the reports name, and withdraw and reinstate Windows. The administrators of the server do it from their session in the console of the Cahut ([ADR 43](0043-the-console-holds-two-sessions.md)).
- **Moderation is for people only.** No AI and no service is offered its operations: a decision of moderation belongs to a person, and reports hold what others wrote.

### Reporting a Window

- **Anyone reports a Window, without an account, at `https://<domain>/-/report`.** It is a page of Cahut, rendered on the server without a script, as the page of a Key is, in the brand of Cahut and in English ([ADR 36](0036-console-in-english-and-french.md)).
- Its form asks for:
  - the address of the Window: any address that a reader of this Cahut meets, `/<room>/<window>/` or a file under it, the permanent address `/-/w/<id>` ([ADR 50](0050-window-ids-permanent-addresses-and-names.md)), or a share link `/-/s/<token>`. The address of a Room that shows a Window at its address names that Window ([ADR 32](0032-the-page-of-a-room.md));
  - what is wrong: phishing or a scam, malware, illegal content, the rights of someone (copyright, privacy), or something else;
  - an explanation, of 10 to 2,000 characters;
  - a name and an e-mail address, both optional, for a receipt and the decision;
  - a box to tick: the report is made in good faith, and what it says is true.
- The form posts to its own address. `report_window` (`PeopleOnly`, `Visitors`) finds the Window and keeps the report. An address that names no Window of the Cahut, or a Window in the trash, is refused, with what to check. Once kept, the page says that the report was received, and nothing about the Window.
- Reports are limited by network address, as sign-in is ([ADR 13](0013-people-sessions-and-rights.md)).
- **Where readers find it.** The foot of every page of Cahut for readers links Report, beside the legal pages, also when the Cahut has none: the page of a Room, of a Key, of a share link, of what is not found and of a link that has expired. The page of a Room and the page of a Key fill in the address of their Window, with `/-/report?window=<room>/<name>`. The page of a share link does not, so that its token reaches no other address.
- **Nothing is drawn in or around Windows**, as [ADR 42](0042-the-legal-pages-of-a-cahut.md) has it: the reader of a public Window types or pastes its address. A host says in its legal notice where to report too.

### Who learns of a report

- The console of the Cahut gets a page, Moderation, for its moderators: the reports that wait come first, and the menu counts them.
- The list of the Cahuts of the server, in its console and in `cahut domain list`, says how many reports wait in each.
- When the Cahut has a relay ([ADR 34](0034-e-mail-through-the-relay-of-each-cahut-only.md)):
  - each active administrator of the Cahut who has an address receives an e-mail for each report, with its number, its category, the address of the Window and a link to the report in the console. The explanation stays in the console.
  - **the address of moderation** receives the same e-mail. The administrators of the server have no e-mail ([ADR 34](0034-e-mail-through-the-relay-of-each-cahut-only.md)), so they set an address for each Cahut, such as an abuse mailbox of the host, with `set_moderation_address`, from its page Kind and limits or with `cahut settings moderation-address`. It lives in the settings of the Cahut, beside its legal pages, and travels with it.
  - whoever reports with an address receives a receipt, then the decision: the Window was withdrawn, or the moderators left it as it is. These e-mails name the address that was reported and hold nothing that the reporter wrote, so that the form cannot send a text of one's choice to someone else.

### Opening a reported Window

- **A moderator opens the Window that an open report names, whatever its access**, in the console. `preview_reported_window` gives a frame of [ADR 27](0027-previews-of-windows-in-the-console.md), bound to the session of the moderator, of the Cahut or of the server, that opens this Window only, for the time of a preview. The Window runs in its sandbox, as in every preview ([ADR 10](0010-windows-in-a-csp-sandbox.md)).
- **Each opening goes to the Logbook in the Room of the Window**, so that its members read it, with the moderator and the number of the report.
- **Without a report, a moderator opens what anyone opens:** a public Window, at its address. The page Moderation lists the public Windows of the Cahut, the latest made public first, with their Room and the person who published them, so that moderators see what the Cahut serves to anyone without waiting for a report.
- A Window open to its author, to its Room, to the whole Cahut, with a Key or with a share link opens to no moderator without a report, beyond what ADR 13 already gives them.

### Withdrawing a Window

- **`withdraw_window` withdraws a Window** that an open report names, or a public Window, with a reason that the call must give, of 500 characters at most. The reason is written for the members of its Room. Withdrawing a withdrawn Window changes its reason.
- **A withdrawn Window keeps everything:** its versions, its access and its end, its Key, its share links and their passes, its invitations. Until it is reinstated:

| Who | What they meet |
| --- | --- |
| Readers | Its address, its permanent address, its share links and its frames answer with a page of Cahut, "This Window was withdrawn by the moderators of this Cahut.", with the status `410 Gone` and the foot of every page. The page gives neither the reason nor the title. The page of its Room no longer lists it, nor shows it at its address. |
| The members of its Room | The console marks it withdrawn, with the date, the reason, and who withdrew it: the administrators of the Cahut, or those of the server. They preview it there. Every operation that changes it or opens it to someone is refused with the code `window_withdrawn`: a new version, a rollback, its access, its Key, its name, its title and description, a share link, an invitation, the home of its Room. They put it in the trash, restore it, and it stays withdrawn, and purge it, as any other Window. |
| Their AIs | `get_window` and `list_windows` give its state and its reason. The calls refused give a message that says it was withdrawn, why, and to tell the person, since only the moderators reinstate it. |

- **`reinstate_window` gives it back as it was:** its access, its Key, its share links until their end. A Window that the administrators of the server withdrew is reinstated by them only, since the host answers for what its server serves. A Window that the administrators of the Cahut withdrew is reinstated by either.
- When the Cahut has a relay, the members of its Room who have an address receive an e-mail when it is withdrawn, with the reason and the support page of the Cahut when it has one ([ADR 42](0042-the-legal-pages-of-a-cahut.md)), and another when it is reinstated. This is the statement of reasons that the Digital Services Act asks for.
- **Withdrawing decides the reports** that are open about the Window: it was withdrawn. `dismiss_report` decides a report and leaves the Window as it is, with a note for the moderators, optional.

### Keeping reports

- A report keeps its number, its category, its explanation, its Window, its date, its decision, who decided and when.
- The name and the address of whoever reported are erased once the decision is sent, or once it is decided when the Cahut sends no e-mail. Cahut keeps neither their network address nor their browser.
- Reports go with their Window when it is purged. The Logbook keeps what was reported and what was decided.

### The Logbook

- `report_window` is logged in the Cahut, without the Room, so that only its administrators read it, with the number of the report, its category and the Window. Its actor is someone who reports, unnamed, as for the openings of share links ([ADR 52](0052-share-links.md)). The entry holds neither the explanation, nor the name, nor the address.
- `preview_reported_window`, `withdraw_window` and `reinstate_window` are logged in the Room of the Window, with the number of the report and the reason.
- `dismiss_report` is logged in the Cahut, with the number of the report.

### Operations

| Operation | Who calls it | What it does |
| --- | --- | --- |
| `report_window` | The page `/-/report`, for its visitors (`PeopleOnly`, `Visitors`) | Keeps a report about a Window |
| `list_reports` | The moderators | Lists the reports, those that wait first |
| `list_public_windows` | The moderators | Lists the public Windows, the latest made public first |
| `preview_reported_window` | The moderators | Gives a frame that opens the Window of an open report |
| `dismiss_report` | The moderators | Decides a report, and leaves its Window as it is |
| `withdraw_window` | The moderators | Withdraws a reported or public Window, with a reason |
| `reinstate_window` | The moderators; for a Window that the administrators of the server withdrew, them only | Gives a withdrawn Window back as it was |
| `set_moderation_address` | The administrators of the server | Sets or removes the address that receives the reports of the Cahut |

- Every operation but `set_moderation_address` is an operation of the Cahut at the level of its administrators, `PeopleOnly`, which the administrators of the server call too. `set_moderation_address` is at the level of the administrators of the server, as `set_legal_links` is.
- A suspended Cahut answers `/-/report` with its page, as every other address ([ADR 48](0048-suspending-a-cahut.md)).

## Consequences

- A host stops one Window and leaves the rest of the Cahut as it is, from a report or from the list of public Windows, and gives it back in one gesture. Whoever notifies gets a receipt and a decision, and the person whose Window is withdrawn gets the reason, as the Digital Services Act asks.
- People know who can read their Windows: the moderators open a Window that a report names, and the Logbook of its Room tells its members each time, with the name of the moderator. The documentation says so, and a host says it in its terms.
- An administrator of a Cahut could report a Window to open it. The Logbook of its Room shows that opening to its members, with the administrator who made it.
- A person who publishes abuse again and again is disabled, deleted ([ADR 47](0047-deleting-a-person.md)), or their Cahut suspended ([ADR 48](0048-suspending-a-cahut.md)), as today. Disabling a person leaves their public Windows open: the moderators withdraw them too.
- A Cahut that a team installs for itself gets the same: its administrators moderate it. Its pages for readers show a line with Report alone when it has no legal pages.
- Migrations give the database of every Cahut a table of reports, the state of withdrawal of each Window, and the address of moderation in its settings.
- The console says `window_withdrawn` in English and French ([ADR 37](0037-refusals-with-a-code-for-the-console.md)), and the page Moderation in both languages. The page `/-/report` speaks English, as the other pages for readers.
- The tests of every operation run over the matrix of callers of ADR 13. A test walks every address of a withdrawn Window and checks that none serves it.
- This record supersedes in part ADR 13: the moderators open a Window that a report names, beyond the Windows of their Rooms. It extends ADR 19 (a Window can be withdrawn), ADR 42 (the foot of the pages for readers links Report, and the settings of a Cahut hold its address of moderation) and ADR 52 (the share links of a withdrawn Window open nothing until it is reinstated).

## Alternatives considered

- **The administrators of the server alone.** The host decides, and there is one level less. But an organization whose Cahut runs on the server of a host could not take down what its own members publish, and in a Cahut installed for itself, they are the same people.
- **Moderators who open every Window that opens beyond its Room**, without a report. A closer watch, but a Window with a Key or a share link is shared with readers chosen one by one, and its people would have nothing that the host could not read.
- **Moderators who open public Windows only.** A Window shared by a link, which serves phishing as well as a public address, would be withdrawn unseen, on the word of a stranger.
- **A delay before a Window opens beyond its Room**, such as an hour, for a moderator to look. A delay on public Windows protects nothing while a share link opens at once, and a delay on every access would stop the sharing that people come for. A setting of the host can come later, beside this record.
- **A button Report on every Window.** Readers would find it at once, but Cahut would draw on the content of Rooms, against ADR 42, and next to the scripts of the Window.
- **An address of abuse alone**, without a page or a list. Hosts would keep track of reports by hand, and nothing would hold the receipt and the decision that the law asks for.
- **Deleting a Window rather than withdrawing it.** It cannot be undone, and loses what a mistake would need back. Its members purge it themselves if they want it gone.
- **Reports in the Logbook of the Room.** Its members would learn of a report before any decision, and read the words of a stranger.
- **The status `451 Unavailable For Legal Reasons`.** It says that a legal demand stopped the page, while a host also withdraws for its terms.
- **Moderation by AIs.** The AI of an administrator would read the reports of others and Windows that its person did not open. Withdrawing stays a decision of a person.
