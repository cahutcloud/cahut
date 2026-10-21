# 5. The console is a Preact app embedded in the binary

- Status: Accepted, superseded in part by [ADR 11](0011-console-on-its-own-port.md) (the console's own port)
- Date: 2026-10-21

## Context

The web console is where people see what is published, share it, approve Knocks, read the Logbook and run the home. One product rule governs it: the console uses the same API as the AIs, and no capability is reserved for it. Its look comes from the brand guidelines: tokens from `cahut-tokens.css` and components from `cahut-ui.css`.

## Decision

The console is a single-page app written in TypeScript with [Preact](https://preactjs.com), built with Vite.

- The built files are embedded in the binary with `go:embed` and served at the home's main origin.
- The console calls only the public HTTP API, that is, the operations of the registry ([ADR 1](0001-operation-registry.md)). There are no server-rendered console pages and no private endpoints.
- It uses `cahut-ui.css` and the brand tokens, with no other component library.
- Node.js is needed to build Cahut, never to run it. The build output is not committed: the Makefile, CI and the release build compile the console before the binary.
- Pages served on a Window's origin, such as the Key page of a password Door, are not part of the console. They are small pages rendered by the server in the organization's house style.

## Consequences

- The one-API rule checks itself: whatever the console needs, the API has, and so the AIs have it too.
- The binary stays self-contained, with no CDN and no separate front-end server.
- A JavaScript toolchain enters the repository, with its own dependency updates.
- A plain `go build` skips the console. Batch 2 of the [V0 plan](../plan-v0.md) makes such a binary serve a short notice instead of failing.

## Alternatives considered

- **Svelte.** Concise, with light compiled output, but its own component language. Preact keeps the React idioms that AIs and contributors write fluently, with about 4 KB of runtime.
- **React.** The same idioms as Preact, with a larger runtime.
- **Go templates with htmx.** No JavaScript build, but the handlers would sit next to the internals, and the one-API rule would rest on discipline alone.
