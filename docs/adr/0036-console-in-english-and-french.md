# 36. The console speaks English and French, in the language of each browser

- Status: Accepted, superseded in part by [ADR 37](0037-refusals-with-a-code-for-the-console.md) (the console says in its language the refusals of the server that carry a code)
- Date: 2026-10-25

## Context

The console is where every person of a Cahut signs in, answers the Knocks of their AIs, manages their Rooms and Windows and reads the Logbook. Its words were in English only, the language of the launch. Its first users speak French, as the website already does under `/fr/` ([ADR 12](0012-website-and-docs-as-astro-sites.md)).

The console has some 25 screens, a few of them before sign-in: sign-in, activation links, the first start ([ADR 25](0025-first-start-in-the-browser.md)), the consent screen of OAuth ([ADR 16](0016-oauth-for-ai-clients.md)) and invitations. Several of them open from a link with a code in its fragment, which the console takes out of the address once read.

Other words reach people from the server itself: the messages of errors and some answers of the API ([ADR 14](0014-console-http-api.md)), written for AIs and people alike, the entries of the Logbook, e-mail ([ADR 31](0031-e-mail-through-a-relay.md)) and the pages that readers see.

## Decision

### Which language

- The console speaks English and French.
- It speaks the language that the person chose in this browser, or else the first of the browser's languages that it knows, or else English. The choice lives in the storage of the Cahut's origin, beside the session, under `cahut.lang`.
- The person chooses it on their Account page, and under the card of the pages that open before sign-in. The choice reloads the console; the code of a link goes back into the address first, so that the page keeps it.
- The page sets `lang` on its document, and dates and numbers follow the locale of the language.

### Where the words live

- Each screen keeps its words in both languages next to its code, with `words()` of `internal/console/ui/src/i18n.ts`: `words({ en: { … }, fr: { … } })`. Its types require every French word for every English one, with the same arguments: a word missing in French fails the build.
- A sentence that holds values is a function in both languages, and a passage with markup gives the same markup in both, so that each language orders its sentence as it needs.
- French sets a no-break space before its high punctuation and after a number, as the website does: `words()` sets them, and the French words are written with plain spaces. Small helpers elide `de` and `que` before a name.
- The words of the house (Cahut, Room, Window, Key, House style, Knock first, Logbook) stay in English, as on the website and in the brand guidelines.

### What stays in English

- What the server writes: the messages of errors, the Logbook, the names of operations, and the e-mail. The console shows them as they come. An answer that is always the same, such as that of `request_password_link`, the console says again in French.
- The pages that readers see, outside the console, and the documentation.

## Consequences

- A person who reads French uses the console in French from the first page they open, with no setting, and keeps the choice they make.
- A change of the words of a screen changes both languages, in the same file and the same pull request.
- In a French console, a refusal of the API still reads in English, such as a wrong password at sign-in. Saying those messages in the language of the person is a change of the server, for a later record.
- The choice belongs to a browser and a Cahut: another browser, or another Cahut, follows its own languages until the person chooses there too.
- A third language adds a key to each `words()` and a line to the list of languages.

## Alternatives considered

- **One dictionary per language, as the website has.** One file for a translator, but every word far from the screen that shows it, and two large files to keep in step with some 25 screens that change often.
- **A library of internationalization**, such as i18next or Lingui. A dependency, and a step of the build, for two languages that plain TypeScript types keep in step.
- **The language in the account of each person, on the server.** It would follow the person from browser to browser and could choose the language of their e-mail, but it needs a migration and an operation, and says nothing to the pages before sign-in. It can come with e-mail in French, and the browser keeps answering before sign-in.
- **One language per Cahut, set by its administrators.** A Cahut holds people of several languages; the browser knows the language of each.
- **A path per language, such as `/console/fr/`.** Every link that Cahut gives, for activation, a Knock, consent or an invitation, would need the language of its reader.
