# 62. A sign-in refused during a wait does not prolong it, and its refusal says when the wait ends

- Status: Accepted
- Date: 2026-11-15

## Context

Refused sign-ins make the next ones wait: ten to a handle from one address within 15 minutes make that address wait, and a hundred from everywhere make everyone wait ([ADR 13](0013-people-sessions-and-rights.md)). Since [ADR 60](0060-signing-in-after-a-forgotten-password.md), the refusals before a person's last new password no longer count. Each sign-in refused during the wait still counts, though it checks no password, and so it makes the wait last.

A person of a Cahut chose a password from a link while they waited, before ADR 60, then tried again seven times in six seconds. Each try told them to wait 15 minutes, and each started the wait again. The message gave a fixed 15 minutes, true only for someone who stops trying, and did not say that a new password ends the wait. The person gave up.

The same count lets one stranger lock a person out. An address signs in 10 times at once, then once every 6 seconds: 160 times in 15 minutes. After ten wrong passwords it waits, but each refusal of its wait counts towards the hundred from everywhere, which it reaches in about nine minutes and keeps for as long as it goes on. The limits meant the opposite: a stranger who knows a handle cannot lock its person out from a few addresses.

## Decision

- Only the refused sign-ins that checked a password count towards a wait: a wrong password, or a handle or address that no account has. A sign-in refused because it must wait, or because too many came at once, checked none, and no longer counts. The count reads the Logbook as before, and leaves out the refusals whose error is of the kind too many attempts.
- The refusal of a sign-in that must wait says when the wait ends: the minutes until the oldest of the refusals that make it wait is 15 minutes old, rounded up. It speaks of the account rather than the handle, since a person signs in with their e-mail address too. Its code, `too_many_refused_sign_ins`, does not change.
- Under this refusal, the page of sign-in of the console says that a new password ends the wait, with a button to ask for one where the Cahut sends e-mail. Elsewhere, its foot already says who gives a new link.
- No e-mail tells a person that their sign-ins wait: they are in front of the page that says so.
- The same holds for the administrators of the server.

## Consequences

- A person who tries again during a wait no longer makes it last, and is told how long it lasts and how to end it sooner.
- An address still checks ten passwords at most to a handle in 15 minutes. A stranger needs ten addresses at least to make a person wait everywhere.
- The refusals already in the Logbook follow the new count as soon as it ships: no migration.

## Alternatives considered

- **Keeping the count, and saying that trying again makes the wait last.** A person who cannot sign in tries again, and one stranger would still lock a person out.
- **An e-mail when a sign-in must wait.** The person who waits reads it on the page, and anyone who knows a handle or an address could make Cahut write to its person. An e-mail when the hundred refusals from everywhere are reached, a sign that someone else is guessing, may come later, with its own record.
- **A column of the Logbook for the refusals that checked a password.** A migration, for what the error of each entry already says.
