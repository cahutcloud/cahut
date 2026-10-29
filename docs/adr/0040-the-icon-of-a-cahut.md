# 40. The icon of a Cahut

- Status: Accepted, superseded in part by [ADR 49](0049-no-house-style.md) (the root of a Cahut serves the icon of Cahut, always)
- Date: 2026-10-25

## Context

AI clients such as Claude, ChatGPT and Le Chat show an icon beside each connector. A Cahut gave them none. It served nothing at `/favicon.ico`, its console added its icon by script, which the services that draw the icon of a site do not run, and its MCP server named no icon. The services that clients use answered that the domain of a Cahut has no icon, and the clients fell back on the parent domain: `my.cahut.cloud` showed the icon of `cahut.cloud`, as a favicon service had cached it, long after the brand guidelines had changed the app icon. The Cahut of an organization at `cahut.example.com` would show the icon of `example.com`, or a generic globe.

The brand guidelines draw the app icon for the web: a favicon, an icon for iOS and icons of 192 and 512 pixels. [ADR 32](0032-the-page-of-a-room.md) and [ADR 33](0033-the-brand-of-cahut-until-the-house-style-has-a-look.md) give the pages of a Cahut the brand of Cahut while its house style keeps the look of the neutral one, and never show the logo of Cahut beside an organization. The console is a screen of Cahut, in its brand whatever the house style.

## Decision

- Each Cahut serves the icons of the brand under `/-/brand/`, at the paths they have in the brand guidelines: the favicon, `logos/favicon/favicon.ico`, the icon for iOS, `logos/favicon/apple-touch-icon.png`, the app icon, `logos/cahut-icone-app.svg`, and its render of 512 pixels, `logos/favicon/icon-512.png`.
- The page of the console names its icons in its HTML, under `/-/brand/`: the favicon, the SVG favicon that follows the theme of the system, and the icon for iOS. Services that do not run its script find them, and they stay whatever the house style.
- The MCP server of a Cahut gives its title, Cahut, the address of the Cahut as its website, and two icons in its server info: the app icon in SVG, of any size, and its render of 512 pixels, both at the origin of the Cahut.
- `/favicon.ico` and `/apple-touch-icon.png`, where browsers and favicon services look for the icon of a site, serve the icon of Cahut while the house style keeps the look of the neutral one, and are not found once it has a look of its own. Every page of the domain shows them in its tab when it names no icon, Windows included.

## Consequences

- An AI client that reads the icons of an MCP server shows the icon of Cahut beside the connector, from the Cahut itself. One that asks a favicon service finds the icon of the domain of the Cahut once the service has visited it, rather than that of the parent domain.
- The icon follows the brand guidelines: the binary embeds the files of `docs/charte-graphique`, and a release carries the icon of its version.
- A Window that names no icon shows that of Cahut in its tab while the house style is neutral, as the pages of the Cahut do, and none once the organization has a look of its own.
- A favicon service that cached an older icon keeps it until it visits again; a new icon reaches clients at its pace, not at that of a release.
- This record extends ADR 32: the root of a Cahut carries the brand of Cahut on the same terms as its pages.

## Alternatives considered

- **The icon of Cahut at the root whatever the house style.** The tab of every Window of an organization would show it, beside its look: ADR 32 keeps the logo of Cahut off its pages.
- **The logo of the house style as the icon.** A logo is drawn for a header, often wide, and an icon is square and small. The house style may give an icon of its own one day.
- **The icons in the MCP server info alone.** Not every client reads them yet; most ask a favicon service for the domain.
