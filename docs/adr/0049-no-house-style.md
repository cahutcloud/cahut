# 49. Cahut has no house style, until a module of skills brings the brand of an organization

- Status: Accepted
- Date: 2026-10-29

## Context

[ADR 22](0022-house-style-of-each-cahut.md) gave each Cahut a house style: the name of its organization, its colors, its fonts, its logo and guidance for AIs, set in the console or by an AI and undone in one click. Cahut served it under `/-/` as a style sheet, logos and fonts, with a deck engine and a template for each kind of Window, and gave it to AIs with `get_house_style`. [ADR 23](0023-house-style-inactive-for-ais.md) let an organization with guidelines of its own make it inactive for AIs, and [ADR 32](0032-the-page-of-a-room.md) and [ADR 33](0033-the-brand-of-cahut-until-the-house-style-has-a-look.md) gave the pages that readers see the brand of Cahut until the house style had a look of its own.

The house style holds one brand, in one shape: a few colors, two fonts, a logo and three templates, the same for every organization. Organizations that care about their brand already keep it elsewhere, often in a skill for their AIs, which says far more than a style sheet can: a tone of voice, layouts, the rules of their charts. ADR 23 was made for them, and the staging Cahut ran with its house style inactive. An AI that follows the skill of an organization does not need the templates of Cahut, and the classes of the kit are a contract with every Window that links them, which Cahut would maintain for brands that do not use it.

Cahut will instead hold the skills of an organization, which its AIs read before they make a Window: a module of skills, which a later record settles. Until then, each Window carries its own look.

## Decision

### What goes

- The operations `get_house_style`, `set_house_style` and `rollback_house_style`, and with them the tools that AIs see.
- What Cahut served under `/-/` for the house style: the style sheet `/-/house-style.css`, the logos `/-/house-style/logo` and `/-/house-style/logo-inverse`, the fonts under `/-/house-style/fonts/`, the deck engine `/-/deck.js` and the templates `/-/templates/<kind>`. These addresses are not found.
- The kit in `internal/housestyle/kit/`: its classes, the deck engine, the templates and the guide for AIs, under Apache-2.0. Their contract with published Windows ends with them, before a major version: Cahut is in its 0.x versions.
- The screen House style of the console, and the step Your Cahut of the first start, which set the name of the organization, its main color and its logo. The first start creates the account of the first person, then connects their first AI.
- The tables `house_styles` and `house_style_files`, which a migration drops. The update that installs it takes a snapshot of the databases first ([ADR 24](0024-signed-releases-and-updates.md)).
- The word House style, from the vocabulary of the interface, the docs, the website and the brand guidelines, with its pictogram.

### What stays, and what changes

- **AIs make each Window in the style that the person asks for**, with its CSS, its fonts and its scripts among its own files, linked by relative paths ([ADR 9](0009-rooms-and-windows-as-paths.md)). A deck carries its own engine, or uses a library whose origin its declaration lists ([ADR 19](0019-window-versions-access-and-declaration.md)).
- **The instructions of the MCP server** say first how to publish a Window: a complete `index.html`, relative links, a declaration, the address given to the person, and Knock first ([ADR 29](0029-chatgpt-second-ai-client.md)). They name no house style.
- **The skill for Claude stays**: its plugin `cahut` in the marketplace of this repository, and its zip at `/-/cahut-skill.zip` ([ADR 22](0022-house-style-of-each-cahut.md)). It says how to make and publish a Window, without a house style, and its plugin takes a new version.
- **The pages that readers see carry the brand of Cahut**, always: the page of a Room, of a Key, of what is not found, of an expired link, of a suspended Cahut, and the sign-in of readers. `/favicon.ico` and `/apple-touch-icon.png` serve the icon of Cahut, so a Window that names no icon shows it in its tab ([ADR 40](0040-the-icon-of-a-cahut.md)).
- **A Cahut is named by its domain.** E-mails speak of "the Cahut at cahut.example.com".
- **The disk of a Cahut** counts every version of every Window, those in the trash included ([ADR 30](0030-kind-and-limits-of-each-cahut.md)).
- **The Logbook keeps its past entries**, those of `set_house_style` and `rollback_house_style` included: it records what happened, and nothing rewrites it.
- **Rule of the brand**: Cahut's own brand is for its own screens, the console, the pages of a Cahut, its e-mails and its sites. A Window carries the look that its AI gave it, never the brand of Cahut.

### Records

- This record supersedes [ADR 22](0022-house-style-of-each-cahut.md), but for the skill for Claude, its plugin and its zip, [ADR 23](0023-house-style-inactive-for-ais.md) and [ADR 33](0033-the-brand-of-cahut-until-the-house-style-has-a-look.md).
- It supersedes in part [ADR 19](0019-window-versions-access-and-declaration.md) (the pages of Cahut carry its brand), [ADR 25](0025-first-start-in-the-browser.md) (the first start has no step for the house style), [ADR 29](0029-chatgpt-second-ai-client.md) (the instructions of the server name no house style), [ADR 30](0030-kind-and-limits-of-each-cahut.md) (the disk counts Windows only), [ADR 32](0032-the-page-of-a-room.md) (the pages of a Room carry the brand of Cahut, always) and [ADR 40](0040-the-icon-of-a-cahut.md) (the root of a Cahut serves the icon of Cahut, always).
- What other records say of the house style in passing, such as what a Cahut keeps ([ADR 8](0008-several-cahuts-per-server.md), [ADR 48](0048-suspending-a-cahut.md)), no longer applies.

## Consequences

- Published Windows that link `/-/house-style.css` or `/-/deck.js` lose their look: their text reads in the fonts and colors of the browser, and the slides of a deck follow one another down the page. Publishing them again with their own CSS gives them back a look, which no change of Cahut can take away.
- A Window holds all it needs, and keeps its look when it moves to another Cahut, or to another server.
- Cahut maintains no deck engine and no classes, and no promise to keep them.
- An organization gives its brand to its AIs through their own skills until Cahut holds them.
- The administrators of a Cahut have one screen less, and the first start one step less.
- The pages that readers see look the same on every Cahut: the brand of Cahut, light or dark as their system is.

## Alternatives considered

- **Keep the style sheet and the deck engine, frozen**, without operations, screen or storage, so that published Windows keep their look until a major version. Cahut would keep serving a kit that no AI is told about. Cahut is in its 0.x versions, its first Cahuts are days old, and a Window published again carries its own look.
- **Keep the name of the organization in the settings of the Cahut**, for its e-mails. The domain already names a Cahut, and the brand of an organization will come with its skills.
- **Make the house style inactive by default**, and leave the rest. A feature that nobody is pointed to still costs its contract, its storage and its screen.
- **Wait for the module of skills**, and remove the house style when it comes. Every Window made until then would link a sheet bound to go.
