---
title: House style
description: The brand of the organization in each Cahut, the style sheet that Windows link at a stable address, the templates and the deck engine, and how administrators change it and go back.
---

The house style is the brand of the organization, applied by every AI: its name, its colors, its fonts, its logo, and what it asks of AIs. Each Cahut has its own. An AI reads it before it makes a Window, starts from a template in it, and links its style sheet, so that a document, a deck or a dashboard comes out in the colors of the organization. The pages that Cahut shows readers, such as the page of a Key, carry it too. The decisions are [ADR 22](https://github.com/getcahut/cahut/blob/main/docs/adr/0022-house-style-of-each-cahut.md) and [ADR 23](https://github.com/getcahut/cahut/blob/main/docs/adr/0023-house-style-inactive-for-ais.md).

## What it holds

| Part | What it is |
| --- | --- |
| Name | The name of the organization, at the top of the pages of Cahut and in the templates. The domain of the Cahut until it has one. |
| Colors | Text, background, primary and secondary, as `#rrggbb`. Cahut derives the others: muted text, lines, the background of cards, and the text on the primary and secondary colors, white or the text color, whichever reads better. |
| Colors of charts | Up to eight colors for the series of charts, in order. Without them, the series start with the primary and the secondary colors, then take colors from a palette checked for readers with color vision deficiencies. |
| Fonts | One for headings and one for body text: a family, a generic family to fall back on (sans serif, serif or monospace), and its files in WOFF2, WOFF, TrueType or OpenType, each with its weight and style. A family without files is one that readers' systems have, such as Georgia. |
| Logo | For light backgrounds, and a version for dark backgrounds, such as slides in the primary color. SVG, PNG or WebP. |
| Guidance for AIs | What the organization asks of AIs, such as its tone of voice or its rules for charts, in 4,000 characters at most. AIs read it; readers never see it. |

Body text must contrast at 4.5:1 at least on the background, as WCAG AA asks: a house style below that is refused. A primary or secondary color that reads poorly on the background, or under the text on it, is accepted with a warning. A house style takes 512 KB per image, 1 MB per font file, 16 font files and 8 MB in all.

Until an administrator sets one, a Cahut has the neutral house style: the fonts of the system, dark text on white, a blue primary color, no logo. It is not the brand of Cahut, which never appears on the pages of an organization.

## How Windows link it

Cahut serves the house style of each Cahut at stable addresses, under `/-/`:

| Address | What it serves |
| --- | --- |
| `/-/house-style.css` | The style sheet: the variables of the house style, such as `--hs-primary` and `--hs-font-heading`, the `@font-face` rules of its fonts, and the classes of the templates |
| `/-/house-style/logo` | The logo. Without one, the name of the organization, set in SVG |
| `/-/house-style/logo-inverse` | The logo for dark backgrounds. Without one, the name in light letters |
| `/-/house-style/fonts/<file>` | The files of the fonts |
| `/-/deck.js` | The deck engine |
| `/-/templates/document`, `/-/templates/deck`, `/-/templates/dashboard` | Each template, filled with sample content, in the house style of the Cahut |

A Window links them from the root of its domain, such as `<link rel="stylesheet" href="/-/house-style.css">`: the one exception to the relative links of a Window, and it needs no declaration. The same path works at the address of the Window and in the frame of a protected Window.

These addresses are public: the house style of an organization is in every Window it makes public, and on the page of a Key, which readers see before they sign in. Its guidance is not served. Browsers check them again at each visit, so a change reaches every reader at once: a new logo or a new color shows on every Window that links the sheet, without publishing it again. A Window that must keep its look copies the style sheet into its own files.

## Templates and classes

Cahut has a template for each kind of Window: a document, a deck and a dashboard. A site starts from the document's. `get_house_style` gives an AI the template of the kind it makes, filled with the name of the organization, and a guide to the classes of the sheet. The [House style sheet](/reference/house-style-sheet/) reference gives that guide in full.

- The classes start with `hs-`, and style elements only inside a root: `<body class="hs-document">`, `hs-deck` or `hs-dashboard`. A page that links the sheet for its variables alone keeps its own look.
- They have no specificity over the CSS of a Window: its own rules win.
- They are a contract with the Windows already published. A release of Cahut adds classes and refines how they look; it renames and removes none before a major version.
- The templates, the classes and the deck engine are under the Apache License 2.0, since they end up in the pages of organizations.

## Decks

A deck is a `<body class="hs-deck">` of `<section class="hs-slide">`, in 16:9. Text scales with the slide. A slide opens the deck (`hs-slide--title`), opens a part in the primary color (`hs-slide--section`), holds one message (`hs-slide--statement`) or covers a picture (`hs-slide--image`); each shows the logo and its number at its foot.

With `<script src="/-/deck.js" defer></script>`, and `scripts` declared, the deck shows one slide at a time, as large as the screen:

| Key | What it does |
| --- | --- |
| Right arrow, Down arrow, Space, Page Down | Next slide |
| Left arrow, Up arrow, Page Up | Previous slide |
| Home, End | First and last slide |
| F | Full screen |
| O | Every slide down the page, and back |

A swipe on a touch screen moves too, and buttons show when the pointer moves. The address carries the number of the slide, such as `#5`, so a link opens a deck at a slide. Without the engine, the slides follow one another down the page. Printing gives one slide per page, to save a deck as PDF.

## Change it

The administrators of a Cahut change its house style in the console, on the [House style](/admin/console/#house-style) screen, or through their AI, with `set_house_style`. An AI gives only what changes, such as a new primary color, and the rest stays. An AI granted some Rooms only does not change the look of the whole Cahut: it reads the house style, and that is all.

Cahut keeps the house style before each change as the previous one, and `rollback_house_style`, or **Go back to it** in the console, restores it; the one before that is deleted. Since a change can be undone, it does not knock first. **Back to neutral** starts again from the neutral house style, which then becomes a change like any other.

The Logbook records each change: the name, the colors, the families of the fonts and the number of their files, the size of the logos. Never the contents of a file.

## Turn it off for AIs

The house style of Cahut suits an organization without brand guidelines of its own: in a few minutes, it gives its Windows a look that holds together. An organization that has full guidelines, with more colors, visuals and icons, often gives them to its AIs already, as a skill of their own. It makes the house style inactive, with the switch **AIs use this house style** of the console, or `set_house_style` with `"active": false`:

- `get_house_style` still answers an AI, but gives nothing of the house style: no colors, no fonts, no logo, no guidance, no template. It says that the house style is inactive, and the AI makes the Window in the style the user asks for, with its own CSS.
- The administrators still see the whole house style in the console, to change it or to turn it on again.
- The style sheet, the logos and the fonts are still served: the Windows that link the sheet keep their look, and the pages that readers see, such as the page of a Key, keep the name, the logo and the colors of the house style. An organization can thus give Cahut its name and its logo for these pages, and leave the rest to its own guidelines.

Turning it off or on is a change like any other: the house style before becomes the previous one, and the Logbook records it.

## Fonts and their license

The browsers of readers download the fonts of the house style from the Cahut, like those of any website. Upload only fonts whose license allows it on the web. Fonts are not loaded from Google Fonts or another service: the readers of a Window call no one but the Cahut.
