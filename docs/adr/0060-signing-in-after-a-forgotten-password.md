# 60. People sign in with their handle or their e-mail address, and a new password ends the wait of refused sign-ins

- Status: Accepted
- Date: 2026-11-05

## Context

The people of a Cahut sign in with their handle and a password ([ADR 13](0013-people-sessions-and-rights.md)). Since [ADR 31](0031-e-mail-through-a-relay.md), an account has an e-mail address, and a person who forgot their password asks for a link with their handle or with that address. A person who signs up to a Cahut through the app of a host gives an address and a handle, and remembers the address better.

Sign-in is rate limited, by address and by handle ([ADR 13](0013-people-sessions-and-rights.md)): ten refused sign-ins to a handle from one address within 15 minutes, or a hundred from everywhere, make the next ones wait. Each sign-in refused during the wait counts too.

A person signed in with their e-mail address and their password, and was told that the handle or the password was wrong. They chose a new password from a link, twice, and were told the same each time. They then signed in with their handle, were refused ten times, chose a new password once more, and were told to wait 15 minutes: the refusals for their old password still counted against the new one. For them, the new password was never saved.

## Decision

- A person signs in with their handle, or with the e-mail address of their account, and their password. Case and the spaces around do not matter, as for a forgotten password. Handles hold no @, and an address belongs to one account of a Cahut at most, so each names one person.
- An unknown address is refused as an unknown handle is: the same message, the same cost of a password check, and nothing of it in the Logbook, which notes an unknown account.
- The page of sign-in of the console asks for a handle or an e-mail address in a Cahut that sends e-mail, and for a handle elsewhere, where people do not see their address ([ADR 34](0034-e-mail-through-the-relay-of-each-cahut-only.md)).
- The refused sign-ins before a person's password last changed, from a link or in the console, no longer count. A person who chose a new password from a link signs in with it at once; the refusals after it count as before. The count still reads the Logbook: the refusals of the person since their last `activate_account` or `change_password` that succeeded, within the 15 minutes.
- The same holds for the administrators of the server, who sign in with their handle only: they have no address ([ADR 34](0034-e-mail-through-the-relay-of-each-cahut-only.md)).

## Consequences

- A person signs in with what they remember. Sign-in tells no one which addresses have an account, as `request_password_link` does not.
- Only a link sent to the address of the account or given by an administrator, or the current password, ends the wait. The guesses of a stranger before the change were for a password that is gone, and those after it count again.
- A person who tries again during a wait still makes it last; a new password ends it.
- The console speaks of a handle or an address on the page of sign-in, as it already did on the page of a forgotten password.

## Alternatives considered

- **Keeping the handle only, and saying so on the page.** People sign up with an address, and the page of a forgotten password already takes one: two rules for two forms of the same page confuse more than they protect.
- **Ending the wait after a sign-in that succeeds.** It does not help the person who cannot sign in.
- **A column for the time of the last change of password.** A migration, for what the Logbook already records.
