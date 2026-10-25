# 42. The pages of a Cahut link its legal pages, which the administrators of its server give

- Status: Accepted
- Date: 2026-10-25

## Context

A Cahut serves pages of its own. Its console holds the sign-in, the activation links, the confirmation of an address, invitations and the consent screen where a person connects an AI client ([ADR 16](0016-oauth-for-ai-clients.md), [ADR 20](0020-console-under-a-path-of-port-443.md)). Its pages for readers hold the page of a Room, the page that asks for a Key, what is not found and the links that have expired ([ADR 32](0032-the-page-of-a-room.md), [ADR 38](0038-the-root-of-a-cahut-opens-its-console.md)). Everything else that a Cahut serves is its Windows, the content of its Rooms, each in its sandbox ([ADR 10](0010-windows-in-a-csp-sandbox.md)).

A host who runs Cahuts for the public, such as individual Cahuts where people sign up on a site of the host ([ADR 30](0030-kind-and-limits-of-each-cahut.md), [ADR 35](0035-a-program-on-the-server-manages-people-with-a-service-token.md)), publishes a legal notice, terms of use and a privacy policy on that site. The law of several countries asks a service open to the public to say who publishes it and who hosts it, from every page. The directories of AI clients ask a connector for its privacy policy and its terms, and check that the pages of the service reach them. No page of a Cahut links them: a person who arrives on the consent screen from an AI client, or a reader who opens the page of a Room, cannot find them.

A Cahut that a team installs for itself has no such pages, and needs none of this.

## Decision

### The links

- Each Cahut may link four pages, each optional: its legal notice (`legal_notice`), its terms (`terms`), its privacy policy (`privacy`) and where people get help (`support`).
- A link is the address of a page that the host publishes elsewhere, and starts with `https://`. Cahut serves none of these pages: it links them.
- A link may give another address in French, for the console when it speaks French ([ADR 36](0036-console-in-english-and-french.md)). The French address takes the place of the other in French only, and needs it: every page that the Cahut links is reachable in every language.
- The links are among the settings of the Cahut, in its own database, beside its kind and limits, and travel with it.

### Who sets them

- The administrators of the server set them, through `set_legal_links`, an operation of the Cahut at their level ([ADR 1](0001-operation-registry.md)), from the page Kind and limits of its console, or with `cahut settings links` on the command line. The host answers for these pages, not the organization that uses the Cahut; whoever runs their own server administers it too. No AI is offered the operation.
- `set_legal_links` takes every link: a link left out is removed. Each change goes to the Logbook of the Cahut, with the links before and after. Setting the old links again undoes it.
- `get_cahut` gives the links to people and to AIs, beside the kind and limits, and `describe_sign_in` to whoever has not signed in, for the pages of the console that open before sign-in. Both already exist: no page of the console makes a call more than it did, but for the pages of a link, such as an activation, which ask `describe_sign_in` once. These pages are open to anyone, and the links are public, as on the site of the host.

### Where they show

- In a discreet line of small links, in the text color for what is secondary: under the card of the pages of the console that open before sign-in, under the consent screen, and at the foot of every screen of the console once signed in; at the foot of the page of a Room, of a Key, of what is not found and of a link that has expired.
- Each link reads in the language of its page: Legal notice, Terms, Privacy and Support in English; Mentions légales, Conditions d'utilisation, Confidentialité and Assistance in French. The pages for readers speak English, as ADR 36 has it, and link the addresses that are not in French.
- They open in a new tab, so that a form being filled, such as a sign-in, a Key or a consent, stays as it is.
- The line follows the look of its page: the brand of Cahut in the console, and on the pages for readers the brand of Cahut or the house style, as [ADR 33](0033-the-brand-of-cahut-until-the-house-style-has-a-look.md) has it. It names no host and adds no logo.
- Not in Windows, nor around them: a Window is the content of its Room, and the frame that Cahut puts around a protected Window, or around the Window that a Room shows at its address, draws nothing of its own.
- A Cahut without links shows nothing new: no line, and no space where it would be.

## Consequences

- A host shows its legal notice, terms, privacy policy and support from every page that a Cahut serves itself, in the language of the page, without a release of its own: from the console of the Cahut or its command line.
- The Cahuts that people install for themselves change in nothing.
- The consent screen that an AI client opens links the privacy policy and the terms, as the directories of AI clients ask.
- A Room that shows a Window at its address fills the page with it, without the links. The Window links them if its Room needs them.
- A migration gives the database of every Cahut a table of links, empty.
- A third language of the console adds its addresses beside the French ones, and its words to the line.
- This record extends ADR 30: the settings of a Cahut hold its legal pages too.

## Alternatives considered

- **Legal pages that Cahut serves itself**, with a text that the host writes, at `/-/legal` and alike. Cahut would keep legal texts, their format, their languages and their versions, for pages that every host already publishes on its own site, next to its offer.
- **One set of links for the whole server**, in its configuration. A server may hold the Cahuts of several organizations, which answer to different publishers: a host may run a Cahut for an organization that publishes its own terms. Each Cahut keeps its own, as [ADR 8](0008-several-cahuts-per-server.md) has it, and moves with them.
- **The links in the house style**, which the administrators of the Cahut set. The house style is the brand of the organization, and goes to AIs; the legal pages of a hosted service are those of the host, who answers for them.
- **The links in `set_cahut_settings`**, beside the kind and limits. One operation less, but each change of a limit would have to send every link again, since a value left out is lifted, and a script that sets limits would read the links first.
- **One address per page, whatever the language.** Simpler, but a host who publishes its pages in both languages would link its English terms from a French console.
- **A French address alone, without the other.** The page would be missing from the console in English and from the pages for readers, which the law reads as missing.
- **Other schemes, such as `mailto:` for support.** A page of support gives an address, a form or both; an address alone opens a mail program that the person may not have, and HTTPS keeps every link the same kind of thing.
- **A new operation, `get_legal_links`, for the console.** The pages before sign-in would make one more call open to visitors, which counts against the calls of an address that signs in ([ADR 13](0013-people-sessions-and-rights.md)): people behind one address would spend their sign-ins reading links. `describe_sign_in` and `get_cahut` already answer every page.
