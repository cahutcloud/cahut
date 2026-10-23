# 22. Each Cahut has a house style, served to its Windows and given to AIs

- Status: Accepted, extended by [ADR 23](0023-house-style-inactive-for-ais.md) (the house style can be inactive for AIs)
- Date: 2026-10-23

## Context

The house style is the organization's brand, applied by every AI: its colors, fonts, logo and templates. [ADR 8](0008-several-cahuts-per-server.md) gives each Cahut its own. [ADR 9](0009-rooms-and-windows-as-paths.md) keeps `/-/` for Cahut's own pages and endpoints, the house style sheet among them. [ADR 19](0019-window-versions-access-and-declaration.md) says that the pages of Cahut on port 443, such as the page of a Key, carry the house style of the Cahut, neutral until batch 6, and its policy for the files of a Window admits the origin of their Cahut, so that a Window loads what Cahut serves there without declaring it. [ADR 15](0015-license.md) puts the templates and the skills under Apache-2.0.

Batch 6 of the [V0 plan](../plan-v0.md) brings the house style: colors, fonts, logo, an HTML template for each kind of Window with a small deck engine served by Cahut, an operation that gives it to AIs, since not every client reads MCP resources, a style sheet at a stable address under `/-/`, a screen in the console, a neutral house style by default, distinct from the brand of Cahut, and a skill for Claude packaged as a plugin. Its milestone: the same brief gives a deck in the colors of the organization, published in a single request.

It must settle what a house style holds, who changes it and how it is undone, how Windows and readers get it, what the templates are and what they promise, and how AIs learn it.

## Decision

### What a house style holds

- **A name**, the organization's, which the pages of Cahut show. Until it has one, they show the domain.
- **Four colors**: `text`, `background`, `primary` and `secondary`, each as `#rrggbb`, and up to eight colors for the series of charts. Cahut derives the rest: the muted text and the lines, from the text and the background; the text on the primary color, white or the text color, whichever reads better. Without colors for charts, the series start with the primary and the secondary colors.
- **Readable colors.** Text on the background must contrast at 4.5:1 at least, as WCAG AA asks for body text: below that, the call is refused. A primary color below 3:1 on the background, or the text on it below 4.5:1, is accepted with a warning in the output, which the console shows.
- **Two fonts**, for headings and for body text. Each is a family, a generic fallback (`sans-serif`, `serif` or `monospace`) and, optionally, its files, in WOFF2, WOFF, TrueType or OpenType, each with its weight and its style. A family without files is one that readers' systems have, such as Georgia. The body font is the system's until set; the heading font is the body font until set.
- **A logo**, and a version of it for dark backgrounds, in SVG, PNG or WebP. Cahut checks that each file is what it claims to be.
- **Guidance for AIs**: free text of 4,000 characters at most, such as the tone of voice or the rules for charts. It goes to AIs, never to readers.
- **Limits**: 512 KiB for each image, 1 MiB for each font file, 16 font files, 8 MiB in all.
- **A neutral house style** until an administrator sets one: system fonts, dark text on white, a blue primary color, no logo. It is not the brand of Cahut, which never appears on the pages of an organization.

### Changing it

- `set_house_style` changes what it names and keeps the rest: an AI asked to change the primary color sends that color alone. A logo given empty removes it. `neutral` starts again from the neutral house style before applying the rest.
- The house style before a change becomes the previous one, which `rollback_house_style` restores, as for Windows ([ADR 19](0019-window-versions-access-and-declaration.md)); the one before that is deleted. A change can be undone, so it does not knock first: an AI makes it, its person looks at the result and goes back if needed.
- The administrators of the Cahut change it, and their AIs that reach every Room. An AI granted some Rooms only acts on those Rooms, not on the look of the whole Cahut: it reads the house style, and does not change it.
- `get_house_style` is for anyone of the Cahut, and every AI, those that may only read included.

### Served under `/-/`

- `/-/house-style.css`: the house style as CSS custom properties, prefixed `--hs-`, the `@font-face` rules of its fonts, and the classes of the templates.
- `/-/house-style/logo` and `/-/house-style/logo-inverse`: the logo and its version for dark backgrounds. Without a logo, Cahut sets the name of the organization in SVG, so that a Window that shows the logo never shows a broken image, and shows the logo once there is one.
- `/-/house-style/fonts/<file>`: the files of the fonts.
- `/-/deck.js`: the deck engine.
- `/-/templates/<kind>`: each template filled with sample content, in the house style of the Cahut, for people to look at. They are served in the sandbox of a Window that runs scripts and calls nothing.
- All of it is public. The house style of an organization is in every Window it makes public, and on the page of a Key, which readers see before they sign in. Its guidance is not served.
- Every response carries `Access-Control-Allow-Origin: *`, since the pages of Windows have an opaque origin ([ADR 10](0010-windows-in-a-csp-sandbox.md)) and fetch fonts and module scripts in CORS mode, and `X-Content-Type-Options: nosniff`. Browsers revalidate them each time (`Cache-Control: no-cache` with an `ETag`), so a change reaches every reader at once. They keep the default sandbox policy of port 443 ([ADR 20](0020-console-under-a-path-of-port-443.md)): an SVG logo opened on its own runs nothing.
- A Window links them by root-relative paths, such as `/-/house-style.css`: the one exception to the relative links of [ADR 9](0009-rooms-and-windows-as-paths.md). The same path works from the address of the Window and from the frame of a protected Window. The Window declares nothing for them; a deck that uses the engine declares that it runs scripts.
- The address is stable: changing the house style changes every Window that links it, at once. A Window that must keep its look copies the style sheet into its own files.

### Templates and classes

- **One template for each kind**: document, deck and dashboard. A site starts from the document's. The templates are in the binary, the same in every Cahut, and take the look of the organization from its style sheet. Templates of its own for each Cahut come after the V0.
- **Classes prefixed `hs-`.** The sheet styles elements only inside the root classes `hs-document`, `hs-deck` and `hs-dashboard`, so that a page that links it for its variables alone keeps its own look.
- **A contract.** The classes, the variables, the markup of the engine and its keys are a contract with the Windows already published. A release adds classes and refines how they look; it renames and removes none before a major version.
- **The deck engine.** A deck is a `main.hs-deck` of `section.hs-slide`. Slides are 16:9 and scale to the screen with container query units. The engine shows them one at a time, with the keyboard, buttons and swipes, puts the number of the slide in the address (`#5`), which the frame of a protected Window passes on, goes full screen, and prints one slide per page. Without scripts, the slides follow one another down the page.
- **Apache-2.0.** The templates, the classes of the sheet and the engine end up in the pages of organizations, and carry Apache-2.0, stated in their directory ([ADR 15](0015-license.md)).

### For AIs

- **`get_house_style`** gives the name, the colors with those that Cahut derives, the fonts as CSS stacks, the addresses of the sheet, the logo and the engine, the guidance, a short guide to the classes, and the template of the kind it is asked for. It is an operation, so an MCP tool, rather than an MCP resource: not every client reads resources.
- **The instructions of the MCP server** tell every AI to call `get_house_style` before it makes a Window, to link the sheet, to start from the template of its kind, and to follow the guidance.
- **A skill for Claude**, the same for every Cahut: it carries the formats, the access, the relative links, the sandbox and Knock first, and leaves the house style to `get_house_style`. It is packaged as a plugin, `cahut`, in a marketplace of this repository: `.claude-plugin/marketplace.json` at its root, and the plugin in `plugins/cahut/`, under Apache-2.0. The plugin declares the MCP server of the Cahut whose domain the user gives when installing it.
- **The skill as a zip**, for the Claude apps where skills are uploaded: each Cahut serves it at `/-/cahut-skill.zip`, built from the same files, which the binary embeds. The console links it.

### The pages of Cahut

- The home of the Cahut, the page of a Room, the page of a Key, the page of what is not found and the page of an expired link carry the house style: the logo, or else the name, at the top, its colors and its fonts. They follow its single palette, without a dark variant.
- The console keeps the brand of Cahut: it is Cahut's own screen.

### Logbook

- `set_house_style` and `rollback_house_style` record what changed: the name, the colors, the families of the fonts and the number of their files, the size of the logos. Never the contents of a file.
- Readers who fetch the sheet or the logo are not logged: they read a public file, as they read a public Window.

## Consequences

- The brand of an organization lives in one place. A new logo or a new color reaches every Window that links the sheet, without publishing them again.
- A Window that links the sheet depends on its Cahut: copied elsewhere, it loses its look. Windows live on their Cahut, and one that must travel copies the sheet.
- The classes become a public contract. Changing how they look takes care, since published Windows change with them. The guide that `get_house_style` gives lists every class, and a test checks it against the sheet.
- Uploaded fonts are served to anyone, as the fonts of any website are. The organization must hold a license that allows it, and the console says so.
- Cahut maintains a deck engine in JavaScript, small and without dependencies, served to Windows.
- A brand whose text does not read on its background is refused. Such a palette fails readers anyway.
- The MCP server of the plugin is named after it. A user who already added the Cahut to Claude Code by hand removes that entry, or keeps it and leaves the plugin's server off.
- The repository holds a plugin outside `cmd/` and `internal/`, with a Go file in `plugins/` that embeds the files of the skill: the binary serves the same skill as the plugin, and nothing is copied by hand.

## Alternatives considered

- **MCP resources and prompts.** The natural place for templates in MCP, but not every client reads them, and an AI would have to know to look. They may come in addition later.
- **AIs copy the house style into each Window.** Each Window would keep its look, but a change of brand would reach none of them, and each would carry the same kilobytes. An AI may still copy it.
- **Fonts named from Google Fonts.** No files to upload, but the browser of every reader would call Google for every Window, the data would no longer stay on the server, and each Window would have to declare those origins.
- **Knock first for each change, without versions.** The person would approve a change of colors and logo without seeing it. With the previous version kept, they see the result on their own Windows, and go back.
- **A dark variant of the house style.** Few brands define one, and a deck keeps the look it was made in. After the V0.
- **A house style for each Room**, for teams with a brand of their own. After the V0.
- **The console in the house style.** The console is Cahut's screen, the same for every organization, and its brand must stay recognizable where a person approves Knocks.
- **Templates uploaded by each Cahut.** More freedom, but the templates of Cahut already take the brand from the sheet, and a template uploaded by hand would follow none of the contract. After the V0.
- **The skill in each Cahut's MCP instructions only.** They reach every client, and stay: the skill adds what instructions cannot hold, the formats in detail and when to use each.
