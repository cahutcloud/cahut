# 61. The console guides each person to a first AI, a first Window and a first share, and leads to the listing of its Cahut in the directory of Claude

- Status: Accepted
- Date: 2026-11-05

## Context

Once a person has chosen their password, the overview of the console shows how to connect an AI, as long as they have neither a connected AI nor a token ([ADR 25](0025-first-start-in-the-browser.md)). It is a panel of five tabs, one per client, each a paragraph of instructions, Claude Code first. Nothing says where to start, which tab is theirs, or whether the AI is connected once they are done. People who sign up to an individual Cahut ([ADR 30](0030-kind-and-limits-of-each-cahut.md)) do not know what to do there.

They often arrive on a phone: the link that confirms their address comes by e-mail, and they open it where they read it. Five tabs do not fit the width of a phone, Claude Code and ChatGPT cannot be connected from one, and the steps of Claude ask to copy an address into an app.

Once the person approves an AI on the consent screen ([ADR 16](0016-oauth-for-ai-clients.md)), the console sends the browser back to the client at once. On a phone, they land in the AI with nothing that says what to ask it.

Some Cahuts are listed in the directory of connectors of Claude, as my.cahut.cloud is. A person connects it from its listing in one step, without an address to copy. The console does not know that its Cahut is listed, and gives the address.

## Decision

### First steps, in three steps

- The console has a guide, **First steps**, at `/console/start`, in three steps: connect an AI, ask it for a first Window, share it.
- It opens once, right after the person chooses their password, when they have neither a connected AI nor a token. The overview then shows a card with the step they are at, instead of the panel of tabs, until the guide is done or dismissed. **AIs and access** keeps the steps of every client.
- The first step starts with one question: which AI? The answers are Claude, the app and claude.ai together, since a connector belongs to the account of the person and serves in both; Mistral Vibe; ChatGPT; Claude Code; another MCP client; and no AI yet, which points to the two clients that publish on their free plans.
- ChatGPT asks for the plan first, and with a plan that cannot publish points to Claude and Mistral Vibe instead.
- Each client then has its own page, at `/console/start/<client>`, with its steps one action per line: the address to copy, a link to the page of the client where it goes, and the film of its steps where there is one.
- On a phone, the clients that need a computer, Claude Code and ChatGPT, come after the others, under their own heading. Choosing one says to go on from a computer, where the guide opens at the same step.
- The guide sees the AI arrive: it reads `list_connected_ais` and `list_tokens` every three seconds while it is visible, and as soon as it is visible again. The first step is done when the person has an active connection or a token. The second step proposes requests to copy into the AI, and is done when a Window of the person appears in `list_windows`. The third shows that Window, opens it, and shares it with a share link ([ADR 52](0052-share-links.md)), through the menu of the phone where there is one.
- What the guide shows comes from these calls: a person who connects their AI on a phone finds the guide at the same step on a computer. Only its end, finished or dismissed, is kept in the browser, which no other device sees. The guide adds no operation and no record.

### After the consent

- When the person approves their first AI, one that they have neither connected before nor reached with a token, the consent screen says that it is connected and gives a first request to copy, before it sends them back to the client with a button.
- The client exchanges its code within five minutes ([ADR 16](0016-oauth-for-ai-clients.md)). The screen therefore sends the browser back by itself after four minutes.
- On a phone, the buttons of the consent screen sit at its foot, across its width.

### The listing of a Cahut in the directory of Claude

- `set_listings` sets the address of the listing of the Cahut in the directory of connectors of Claude: an address that starts with `https://claude.ai/directory/`. An empty address removes it. It is for the administrators of the server, as the legal pages of a Cahut are ([ADR 42](0042-the-legal-pages-of-a-cahut.md)): the host lists the service, not its people. The console of the server sets it with the other settings of a Cahut, and the command line with `cahut settings listing`.
- `get_cahut` gives it to everyone in the Cahut, as `listings.claude`.
- When the Cahut has one, the page of Claude in the guide opens the listing first, with a single link, and keeps the address for those who add the connector themselves.
- An address that is not that of a listing is refused with the code `listing_invalid` ([ADR 37](0037-refusals-with-a-code-for-the-console.md)).

## Consequences

- A person who has just chosen their password sees one question, then the steps of their AI alone, then their first Window.
- The console tells them when their AI is connected and when it has published, so they know that it worked.
- On a phone, a person who uses Claude connects it without leaving the guide, and goes back to it with a first request to copy.
- The guide calls three operations every three seconds while it is open. Those reads are light, and stop when the page is hidden.
- A person who dismisses the guide on a phone sees it again on a computer, until they dismiss it there too.
- The steps of the clients live in the console. When a client moves its menus, the guide and **AIs and access** change together.
- A host that lists its Cahut in another directory, such as that of ChatGPT, needs another field in `set_listings`.

## Alternatives considered

- **Keeping the panel of tabs, with the tab of the person's client first.** The console cannot tell which client a person uses before they say so, and the panel still asks them to read every step at once.
- **Two answers for Claude, the app and claude.ai.** They lead to the same steps: the connector belongs to the account, not to the device.
- **Recognizing my.cahut.cloud by its domain.** The code of Cahut would name one host. The setting serves any Cahut that a directory lists.
- **Keeping the end of the guide in the Cahut**, so that it follows the person from device to device. It needs an operation and a record for a convenience; what matters, the connected AI and the first Window, already follows them.
- **Sending the browser back to the client at once, as before, and giving the first request on the tab of the guide only.** On a phone, the person lands in the AI and rarely goes back to the tab of their browser.
- **Telling the person what to ask through the skill of Cahut.** The skill speaks once the person asks something; it cannot open the conversation.
