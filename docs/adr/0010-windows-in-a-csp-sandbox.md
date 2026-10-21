# 10. Windows run in a CSP sandbox

- Status: Accepted
- Date: 2026-10-21

## Context

All the Windows of a Cahut share the origin `https://<domain>` ([ADR 9](0009-rooms-and-windows-as-paths.md)). Browsers isolate pages by origin, not by path. A Window holds HTML and scripts written by an AI, which nobody reviews line by line. Without more:

- the script of a Window could fetch or frame any other Window its reader can open, with the reader's cookies, and read it. A member who publishes a booby-trapped Window would read the Rooms of every person who opens it;
- every Window would share `localStorage`, IndexedDB and cookies with all the others;
- a Window could set a cookie on the domain, which browsers also send to the console on port 8443, because cookies ignore ports ([ADR 11](0011-console-on-its-own-port.md)).

## Decision

- Every response that serves a file of a Window, whatever its type, carries the CSP directive `sandbox`, without `allow-same-origin`:

  ```
  Content-Security-Policy: sandbox allow-scripts allow-forms allow-modals allow-popups allow-popups-to-escape-sandbox allow-downloads
  ```

  The page gets an opaque origin. It cannot read any other page of the Cahut, its requests carry `Origin: null`, and `localStorage`, `sessionStorage`, IndexedDB and `document.cookie` throw `SecurityError`. Pages keep their scripts, forms, dialogs, downloads, and links that open in a new tab.
- The files of a Window are served with `Access-Control-Allow-Origin: *`, never with credentials. The sandbox turns module scripts, `fetch` calls and web fonts into cross-origin requests, and this header keeps them working.
- Cahut's own pages on the domain, at the root, at the address of a Room and under `/-/`, are not sandboxed. They run only Cahut's code, under a strict CSP with `frame-ancestors 'none'`.
- Only Cahut sets cookies on the domain: the reader's session and the Keys they have entered, with the `__Host-` prefix, `HttpOnly`, `Secure` and `SameSite=Lax`. Sandboxed pages can neither read nor set them.
- A public Window is served at its address. A protected Window, open to its Room, to the Cahut or with a Key, is served at its address through a page of Cahut's own. That page checks the reader's access and frames the Window in a sandboxed iframe, whose address under `/-/` carries a capability: a random value bound to that reader and that Window, valid a few hours. The files of the Window then load without cookies, which browsers do not always send from an opaque origin. They carry `Referrer-Policy: no-referrer`, so that the capability does not leak to other sites.

## Consequences

- No Window can read another Window, the console or the reader's session, and none can hand the console a cookie.
- AI-made pages that keep their state in `localStorage` or IndexedDB lose it: the call throws, and a page that does not catch the error stops. The MCP instructions and the skills tell AIs to keep state in memory or in the address. Storage provided by Cahut can come later.
- A protected Window runs in an iframe. The address bar does not follow the navigation between the pages of a multi-page site, and the iframe must allow fullscreen for decks. Batch 4 of the [V0 plan](../plan-v0.md) builds it and tests it in Chrome, Firefox and Safari.
- The declaration of what a Window does (batch 4) adds its own directives to this policy, such as `connect-src` and `form-action`.
- Apps in containers, after the V0, need an origin of their own for their sessions. They get their own decision before their batch.
- This record replaces the isolation part of [ADR 3](0003-one-origin-per-window.md).

## Alternatives considered

- **No sandbox.** Pages keep their storage, but any Window reads what its reader can open.
- **A sandbox that a person can lift for one Window.** The lifted Window reads everything its reader can open: the risk moves, it does not go away.
- **Cookies with `SameSite=None` for protected Windows, without an iframe.** The address bar stays exact, but an opaque origin sends no cookie with module scripts, `fetch` or fonts, and browsers that block third-party cookies may send none at all.
- **A second registrable domain for Windows**, as GitHub serves pages from github.io. It protects the console without a sandbox, but not the Windows from each other, and asks every Cahut for a second domain.
