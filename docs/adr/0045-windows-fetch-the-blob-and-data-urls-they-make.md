# 45. The pages of a Window fetch the blob: and data: URLs they make

- Status: Accepted
- Date: 2026-10-27

## Context

[ADR 19](0019-window-versions-access-and-declaration.md) makes the Content Security Policy of each file of a Window from its declaration. Its `connect-src` lists the origin of the Cahut and the declared origins, and nothing else, while `img-src` and `media-src` allow `data:` and `blob:`, and `worker-src` allows `blob:`.

Libraries read in-memory data with `fetch`. The `GLTFLoader` of three.js, the usual way to show a 3D model in a page, does it twice:

- A `.glb` file holds its textures in its binary buffer. The loader wraps each one in a `Blob`, makes a `blob:` URL for it, and reads it with `fetch`, through `ImageBitmapLoader`: in Chrome, in Firefox since version 98 and in Safari since version 17.
- A `.gltf` file may hold its buffers and textures as `data:` URLs, which the loader reads with `fetch` too.

Under the policy of ADR 19, `connect-src` blocks both reads. The model loads without its textures, and the console of the browser names `connect-src`. An AI that published a 3D furniture configurator on demo.cahut.cloud had to force three.js to load its textures through `<img>`, which `img-src blob:` allows. Every AI that publishes a textured model meets the same wall, and the declaration cannot help: it takes `https` and `wss` origins only.

## Decision

- The policy of each file of a declared Window allows `data:` and `blob:` in `connect-src`, after the origin of the Cahut:

  ```
  connect-src <self> data: blob: <origins>;
  ```

- The rest of the policy of ADR 19 stays as it is. A declaration still lists `https` and `wss` origins only: `data:` and `blob:` need no declaration, as the files of the Window need none.
- The versions published before declarations keep the sandbox alone, as ADR 19 has it.

### Why it is safe

`connect-src` bounds where a page sends requests and what it reads from elsewhere. A `data:` or `blob:` URL is neither:

- Fetching one makes no request on the network. A `data:` URL carries its bytes in the URL itself, which the page wrote. A `blob:` URL names a `Blob` in the memory of the browser. Nothing leaves the reader's browser, so the page cannot send anything to an origin it did not declare, and nothing comes in that the page did not already hold.
- `fetch` reads only the `blob:` URLs of the page's own origin. The origin of a `blob:` URL is that of the page that made it, and a `fetch` in `cors` mode of a URL of another origin, whose scheme is not HTTP(S), is a network error ([Fetch](https://fetch.spec.whatwg.org/#main-fetch), [URL](https://url.spec.whatwg.org/#concept-url-origin)). The sandbox of [ADR 10](0010-windows-in-a-csp-sandbox.md) gives every Window a new opaque origin, so a Window reads the `blob:` URLs that it made, or that its workers made, and no other: not those of another Window, which has an opaque origin of its own, and not those of the console or the pages of Cahut, which have the origin of the Cahut.

Chrome 156 bears this out, with a Window that loads a textured `.glb` and a `.gltf` with `data:` URLs through three.js 0.170, on a local Cahut:

| From a Window | Before | After |
| --- | --- | --- |
| `fetch` of a `blob:` URL that it made | Blocked by `connect-src` | Read |
| `fetch` of a `data:` URL | Blocked by `connect-src` | Read |
| The textures of the `.glb` and of the `.gltf` | Missing | Loaded |
| `fetch` of an undeclared origin | Blocked by `connect-src` | Blocked by `connect-src` |
| `fetch` of the `blob:` URL of another Window, open in another tab | Blocked by `connect-src` | Network error |
| `fetch` of a `blob:` URL made by a page of the Cahut | Blocked by `connect-src` | Network error |

## Consequences

- three.js loads textured models as it does on any other site, and so do the libraries that read a `Blob` or a `data:` URL with `fetch`, such as a page that turns the `data:` URL of a canvas into a file to download. AIs need no workaround and declare nothing more.
- The policy is made at each request from the declaration of the version: every declared Window gets the change on its next request, without a new version and without a migration. The `<img>` workaround of the demo Window keeps working, and may go.
- A Window could already use `data:` and `blob:` URLs as images and media, and `blob:` URLs as workers. It now reads them with `fetch` too: the same bytes, which it made itself.

## Alternatives considered

- **Keep the policy, and tell AIs to load textures through `<img>`.** Nothing changes in Cahut, but every AI that publishes a 3D model must know the workaround, and every library that reads a `Blob` with `fetch` needs one of its own.
- **Let the declaration list `data:` and `blob:`.** The policy would allow them to the Windows that ask, but these URLs call nothing, and AIs would have to learn when to list them, only to load what their pages already hold.
- **`blob:` alone.** It covers `.glb` files, not the `.gltf` files that hold their buffers as `data:` URLs.
