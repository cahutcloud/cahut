# 44. The console shows the server and the Cahut in two bars, and each signs out of its own session

- Status: Accepted
- Date: 2026-10-27

## Context

The console of a domain holds the session of a person of its Cahut and that of an administrator of the server together ([ADR 43](0043-the-console-holds-two-sessions.md)). With both, the console is that of the person, with a section of its sidebar for the server, and the top bar names the person. With the session of the server alone, the top bar names the administrator, followed by "· server" in grey. Signing out ends both sessions; leaving the server, from the page of the account, ends its session only.

In use, an administrator of the server gets lost. Nothing on the screen says that the browser holds the session of the server, but a section of the sidebar or a word in grey. The domain of the Cahut appears on its overview only, while an administrator of the server looks after several Cahuts, each at its own console ([ADR 13](0013-people-sessions-and-rights.md)). And nothing says whether a page of the Cahut acts as the person or as the administrator of the server.

## Decision

### Two bars

- When the browser holds the session of an administrator of the server, a bar of the server sits above the top bar, in the color of the primary button: Ink, or Paper in the dark theme. It says that this is the administration of the server, names the Cahut of the domain, and names the administrator of the server, with a button that signs out of the server.
- The name of the Cahut opens the list of the Cahuts of the server. Another Cahut opens at its own console, where the administrator signs in to the server again: a session is good at one domain ([ADR 13](0013-people-sessions-and-rights.md)).
- The top bar is that of the Cahut, and says who acts in it: the person, with their name and the button that signs them out; or, without a person, the administrator of the server, in words, with a link to sign in as a person.
- A browser that holds the session of a person alone shows the top bar alone, as before.
- On a narrow screen, the bar of the server drops its words but keeps its mark, the name of the Cahut, the avatar of its administrator and the button that signs out of the server.
- Who acts does not change: the pages of the server, and the kind, the limits and the legal pages of the Cahut, call the API with the session of the server; every other page calls it with that of the person, when the browser holds one ([ADR 43](0043-the-console-holds-two-sessions.md)).

### Each bar signs out of its own session

- Signing out, in the top bar, ends the session of the person only. If the browser holds the session of the server, the console stays open for its administrator, and the top bar then says that they act in the Cahut.
- Signing out of the server, in its bar or from the page of the account, ends the session of the server only. If the browser holds the session of a person, the console stays open for them, without the bar of the server.
- When the browser holds one session, signing out ends it, and the console goes back to its sign-in.
- This replaces the rule of [ADR 43](0043-the-console-holds-two-sessions.md) under which signing out ended both sessions.

## Consequences

- An administrator of the server sees on every page that the browser holds their session, in which Cahut they are, and who acts in it.
- Someone alone on their server signs out as a person without leaving the server, and the other way round. A member of the Cahut who administers the server signs out as a person to act in the Cahut as its administrators do, which [ADR 43](0043-the-console-holds-two-sessions.md) asked them to do by signing out of both.
- On a computer that others use, signing out in the top bar no longer closes everything: the session of the server stays until its bar signs it out, or until it ends by itself. The bar of the server stays in sight while it is open.
- The bar of the server takes 48 pixels of height on every page that an administrator of the server sees.
- The words of the two bars come in English and French, as every screen of the console ([ADR 36](0036-console-in-english-and-french.md)).

## Alternatives considered

- **One top bar in Ink for an administrator of the server.** As visible, and no height taken, but one bar names one account, while the browser may hold two.
- **The server and the Cahut as a trail in the top bar, with the role under the name.** Nothing added above the page, but easy to miss, and it still names one account.
- **Signing out ends both sessions, as in ADR 43.** Safer on a computer that others use, but a bar would not answer for its own session, and a person who signs out to let the administrator of the server act in the Cahut would leave the server too.
