# 12. The website and the docs are Astro sites in this repository

- Status: Accepted
- Date: 2026-10-21

## Context

Cahut needs a website at `cahut.cloud`, documentation at `docs.cahut.cloud`, and a place where people can see and download the brand. One person carries the project, so these should not drift apart from the code and from each other.

The brand guidelines live in `docs/charte-graphique`, and the binary uses them too: the console ([ADR 5](0005-preact-console-embedded.md)), the sign-in pages and the e-mails take their tokens and components. The documentation describes the code: the MCP tools come from the operation registry ([ADR 1](0001-operation-registry.md)), with JSON Schemas generated from Go types, and the commands come from the command line.

The project already has its own web server, which answers for `cahut.cloud`. The domain's DNS stays at the registrar.

## Decision

- The website and the docs live in this repository, under `sites/`, as one npm workspace:
  - `sites/cahut.cloud`: the website, an [Astro](https://astro.build) site. Its `/brand` page gives access to the brand: logos, colors, type, rules of use, and the full guidelines to download.
  - `sites/docs.cahut.cloud`: the documentation, an Astro site with [Starlight](https://starlight.astro.build), which brings search, navigation, and light and dark themes.
  - `sites/brand`: an Astro integration that both sites use.
- The brand has one source, `docs/charte-graphique`. The sites read it through the `@charte/` import alias and never copy it. The shared integration lets the development server read the brand files and serves the brand's favicons at the root of each site.
- The repository has one Node.js version, in `.nvmrc` at its root: Node 24 LTS, also used to build the console. Packages come from npm, with a committed lockfile. Install scripts are disabled. Dependabot updates the packages every week, once a release is 7 days old.
- The sites are written in English, in the voice of the README.
- Both sites are static files on the project's own web server, which serves them over HTTPS. `cahut.cloud` is the canonical address, and `www.cahut.cloud` redirects to it.
- The `Sites` workflow builds both sites for every pull request that touches `sites/` or the brand. On `main`, it uploads them to the server with rsync over SSH, with a key used for nothing else and a pinned host key. The upload stays off until the repository's variables name the server.

## Consequences

- A change to the documentation goes in the same pull request as the code it describes, and the reference can be generated from the registry and checked in CI.
- The brand changes in one place, and both sites follow at their next build.
- Node.js is needed to build the sites, never to run Cahut. The binary, its build and its CI do not depend on `sites/`.
- There is no preview per pull request: sites are previewed on the developer's machine with `make site-dev` and `make docs-dev`.
- One server and no CDN, which is enough for a project at launch. Static files can move behind a CDN later without changing the build.
- Until the launch, the upload stays off, or the server protects both domains with a password.

## Alternatives considered

- **Hugo.** A single Go binary that `go run` can fetch, with no Node.js. But search, navigation and themes for the docs would have to be written by hand, where Starlight brings them.
- **One repository per site.** The brand would need a published package with its own versions, the docs would drift from the code, and the instructions and skills for Claude would be copied into each repository.
- **Cloudflare Pages or Netlify.** A CDN and a preview per pull request, but one more provider between readers and the project. Cloudflare also needs the domain's DNS to serve the bare domain.
- **GitHub Pages.** One site per repository, where the project needs two domains.
