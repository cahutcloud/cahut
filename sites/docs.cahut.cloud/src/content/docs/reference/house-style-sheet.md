---
title: "House style sheet"
description: "The variables and the classes of /-/house-style.css, and the markup of the deck engine, as get_house_style gives them to AIs."
sidebar:
  order: 5
---

<!-- Generated from the code by make docs-reference: do not edit. -->

The guide that `get_house_style` gives an AI with the house style of a Cahut. [House style](/concepts/house-style/) says what the house style holds and how it changes.

Link the sheet from the root of the domain, the one path of a Window that is not relative: `<link rel="stylesheet" href="/-/house-style.css">`. It needs no declaration. Start from the template of the kind of Window, keep its structure and classes, and replace its sample content. Your own CSS comes after the sheet and wins over it.

## Variables

Colors: `--hs-text`, `--hs-background`, `--hs-primary`, `--hs-on-primary` (text on the primary color), `--hs-secondary`, `--hs-on-secondary`, `--hs-muted` (secondary text), `--hs-line` (borders), `--hs-surface` (cards and panels), and `--hs-chart-1` to `--hs-chart-8` for the series of charts. Fonts: `--hs-font-heading`, `--hs-font-body`, `--hs-font-mono`. Logos: `--hs-logo` and `--hs-logo-inverse`, as `url()`. Use them in your own CSS rather than colors and fonts of your own.

## Roots

Put one root class on `<body>`: `hs-document` for a document or a site, `hs-deck` for a deck, `hs-dashboard` for a dashboard. The sheet styles elements only inside a root: headings, paragraphs, lists, links, tables, code, quotes and figures.

## Parts for every root

- `<img class="hs-logo" src="/-/house-style/logo" alt="Name">`: the logo, 2rem high. `/-/house-style/logo-inverse` is its version for dark backgrounds. Without a logo, both show the name of the organization.
- `hs-header` and `hs-footer`: a bar at the top and at the foot, with the logo and `hs-meta` (a date, an author).
- `hs-kicker`: a short label above a title. `hs-lead`: the paragraph that opens a text. `hs-muted`: secondary text.
- `hs-callout`, or `hs-callout hs-callout--secondary`: a block that stands out, such as a summary.
- `hs-card`: a framed block. `hs-badge`, `hs-badge--primary`, `hs-badge--secondary`: a short tag.
- `hs-columns`: columns that wrap on narrow screens; `hs-columns--2` and `hs-columns--3` set their number on slides.
- `hs-stat` with `hs-stat__value` and `hs-stat__label`: a key number and what it counts.
- `hs-num` on a table cell: a number, aligned right, in figures of equal width.

## Charts

- Give series their colors in order, never cycled: `hs-series-1`, `hs-series-2`, `hs-series-3`, `hs-series-4`, `hs-series-5`, `hs-series-6`, `hs-series-7` and `hs-series-8` set `--hs-series` to `--hs-chart-1` and the next ones. Past eight series, group the rest as Other.
- `hs-bars`: horizontal bars without scripts. `<ul class="hs-bars"><li class="hs-bar" style="--value: 0.62"><span>North</span><span>62%</span></li></ul>`, where `--value` goes from 0 to 1. Add `hs-series-2` to a bar for another color.
- `hs-legend`: `<ul class="hs-legend"><li class="hs-series-1">This year</li></ul>`, a swatch before each label.
- `hs-chart` on an inline `<svg>`: its text in the body font and muted color; `hs-series-1` on a shape fills it; `hs-line hs-series-1` on a polyline draws a line; `hs-grid` on a line draws a rule. For a chart library loaded from a CDN, declare its origin, and pass it the colors that `chart_colors` lists.

## Document

`hs-document` sets a column of reading width, with room around it. `hs-wide` on a child of `<body>` or `<main>` widens it, for a large table or figure.

## Deck

- `<body class="hs-deck">` holds one `<section class="hs-slide">` per slide, 16:9. Text scales with the slide, so give sizes in `em`, not pixels.
- `hs-slide--title` opens the deck, `hs-slide--section` opens a part in the primary color, `hs-slide--statement` holds one large message, and `hs-slide--image` covers the slide with its first `<img>`, with text at its foot. A plain `hs-slide` has a title (`h2`) and its content.
- Each slide shows the logo and its number at its foot. `<aside class="hs-notes">` holds notes for the speaker, which the slides do not show.
- The engine: `<script src="/-/deck.js" defer></script>` in the head, and declare that the pages run scripts. It shows one slide at a time, as large as the screen: arrows, Space, Page Up and Page Down, Home and End, swipes, F for full screen, O to see every slide. The address carries the number of the slide, such as `#5`. It sets `hs-live`, `hs-current`, `hs-idle`, `hs-progress` and `hs-controls` itself: never write them. Without the engine, the slides follow one another down the page. Printing gives one slide per page.

## Dashboard

- `hs-dashboard` sets a wide page. `hs-kpis` is a row of `hs-kpi`, each with `hs-kpi__label`, `hs-kpi__value` and `hs-kpi__delta`, `hs-kpi__delta--up` or `hs-kpi__delta--down`, which add an arrow.
- `hs-panels` is a grid of `hs-panel`, each with an `h2` and a chart or a table; `hs-panel--wide` takes a whole row.
