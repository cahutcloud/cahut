# 37. The refusals that people meet carry a code, which the console says in its language

- Status: Accepted
- Date: 2026-10-25

## Context

The console speaks English and French, and what the server writes stays in English, the messages of errors among them ([ADR 36](0036-console-in-english-and-french.md)). A person who signs in to a French console with a wrong password reads "Wrong handle or password.", in English, under French labels. The pages that open before and around sign-in are those where people meet refusals most: sign-in, an activation link, a forgotten password, the confirmation of an address, an invitation, a session that ended.

The message of a refusal is written for AIs and people alike, in English, and tells the caller how to fix its call. AIs read it over MCP, the Logbook keeps it, the command line prints it, and a program on the server reads it through the admin API ([ADR 35](0035-a-program-on-the-server-manages-people-with-a-service-token.md)). The console receives it in the body of a failed call, `{"error": "<message>"}` ([ADR 14](0014-console-http-api.md)), and shows it as it comes, without the prefix of its kind.

The language of the console lives in the browser: the person's choice, kept in the storage of the Cahut's origin, or else the browser's languages. The server does not know it. Each screen keeps its words in both languages with `words()`, whose types require every French word.

## Decision

### A code for the refusals that people meet

- A refusal that people meet in the console has a code: a stable name in snake_case, such as `wrong_handle_or_password`. It is declared once, in Go, with its kind (`ErrInvalid`, `ErrForbidden`, `ErrNotFound`, `ErrTooMany`, or none for a refusal of the API itself, such as a session that ended), its English message and the values that fill it: `operation.NewRefusal` for a message that holds no value, `operation.NewRefusalOf` for one that does, such as `password_too_short`, whose `{length}` and `{limit}` come from the JSON names of a struct.
- The English message stays what AIs, the Logbook and the command line read. The code adds to it, and changes nothing that they see.
- The API gives the code of a refusal, and its values when it has some, next to its message: `{"error": "invalid input: the password has 9 characters: use at least 15", "code": "password_too_short", "values": {"length": 9, "limit": 15}}`. MCP gives the message only, as before.
- A code is a contract with the console and with the programs that read the API. It is never renamed, nor given to another refusal; its values may gain a field, and lose none before a major version. The reference of the HTTP API lists the codes, from the code.
- The first codes are those of the pages before and around sign-in: the refusals of `sign_in`, `activate_account`, `change_password`, `request_password_link`, `confirm_email`, `get_invitation` and `open_invitation`, the limits of sign-in, and the session that ended, at both routes of the API. Other refusals get a code when a screen of the console meets them. Until then, the console shows their English message.

### The console says them

- `make console-types` writes the codes into the types of the console, each with the type of its values. The console keeps a sentence for each code in both languages, with `words()`, in `refusals.ts`: a code that the server declares without its two sentences fails the build of the console.
- `explain()` says the sentence of the refusal's code, in the language of the console, or else the English message of the API, as before.
- The sentences of the console may say more, or less, than the message: they speak to a person on a screen, where the message speaks to an AI that fixes its call.

## Consequences

- A French console says in French why a sign-in, an activation, a link, an invitation or a session failed.
- What AIs, the Logbook and the command line read does not change.
- Each code is said three times: by the English message of the server, and by the two sentences of the console. A change of what a refusal means changes all three, in the same pull request.
- A program that calls the API tells refusals apart by their code, rather than by their wording, which may change.
- A new code takes a declaration in Go, `make console-types` and two sentences: CI checks that the types are up to date, and the build of the console that the sentences are there.
- A third language adds a key to `refusals.ts`, as to each screen.

## Alternatives considered

- **The API reads `Accept-Language`, or a language that the console sends, and writes its messages in it.** One place for each message, and any client of the API would read it in its language. But the server would keep words in French, in Go, far from the screens, beside a second set of them in the console, and each refusal would keep its English message for AIs and the Logbook as well. The console would send its own choice with each call, since the browser's header does not know it.
- **The console recognizes the English messages**, by their text or by a pattern. Nothing changes on the server, but a change of wording, or a value in the message, would turn the sentence back to English, and no build would notice.
- **A code for every refusal at once.** Some 260 refusals, most of which only AIs meet: their codes would be a contract that nobody reads. The codes come where people meet the refusals.
- **The values in the order of the message's format.** Shorter to write, but the English message would bind the console: rewording it would swap the values of the French sentence.
- **The English sentences of the console taken from the message.** One English wording, but a sentence of the console could no longer speak of a screen, and English would follow another path than French in `explain()`.
