# The formats of Windows

What makes a good Window of each kind. The look is the user's choice, or that of the skills of their organization; this says what to put in it and how to build it.

## Every kind

- One style sheet for the Window, `css/style.css` or a `<style>` in the page, with its colors and fonts as CSS variables at the top, so that a change of look is one edit.
- The fonts of the system (`system-ui`, `Georgia`) need no file. A font of the brand goes among the files, in WOFF2, with an `@font-face` rule, and only when its license allows it.
- Text that reads: body text at 4.5:1 against its background at least, headings and links in the color that the user asks for.
- `<meta name="viewport" content="width=device-width, initial-scale=1">`, and a layout that holds from a phone to a large screen.

## Document

A report, a memo, a brief, a guide: something people read from top to bottom.

- Open with a kicker (the kind of document), a title that says the subject, and a lead of one or two sentences that says what the reader learns and why it matters.
- Put the conclusion first, in a callout of three findings at most, then the reasons.
- Make headings say something: "Revenue grew in the North", not "Results".
- Short paragraphs, one idea each, about 70 characters to a line. Lists for parallel items, tables for figures.
- In tables, align numbers right with tabular figures (`font-variant-numeric: tabular-nums`), give units in the header, and round to what matters.
- Give each figure a caption that says what it shows and where the data comes from.
- End with the next steps, or the decisions to make, and who makes them.

## Deck

A presentation, shown on a screen or read alone.

- 6 to 15 slides. One idea per slide, and a title that states it: "North leads, West slows down".
- Open with a title slide, a section slide for each part of a long deck, and close on the one message to keep, or on the next steps.
- 3 to 5 short points on a slide, never paragraphs. The details go in a document, linked from the last slide.
- Show numbers big: two or three key numbers, or a chart for a trend or a comparison.
- Make each slide a `<section>` of 16:9 that scales with the screen, its text sized in container query units or `em`, and check that every slide holds its content without overflowing.
- A small script of the deck's own shows one slide at a time: the arrow keys, Space, Page Up and Page Down, Home and End move between them, buttons do the same on touch screens, and the address carries the number of the slide, such as `#5`. The Window then declares that it runs scripts. Without the script, the slides follow one another down the page, and print one per page.
- Speaker notes go in an `<aside>` that the screen hides.
- An image needs a license that allows it, and goes among the files of the Window.

## Dashboard

Figures to watch, read at a glance.

- 3 to 6 key numbers at the top, each with its label, its value and its change over a stated period.
- Then panels, one question each, with a title that names it: "Revenue by region".
- Pick the chart that the question needs: bars to compare, a line for a trend over time, a table for exact values. One axis per chart; two measures of different scales make two charts.
- Give series distinct colors, in a fixed order, never cycled, readable by people with color vision deficiencies, and label them in a legend or directly.
- Say the date of the data and its source, in the header or the footer.
- Charts with scripts declare their library's origin. Without scripts, use inline SVG or CSS bars.

## Site

A few pages linked together, such as a guide or an event page.

- Make each page as a document, with the same header, footer and style sheet.
- `index.html` at the root opens first. Link the pages with relative paths: `about.html`, `guide/setup.html` and, from there, `../index.html`.
- Put a short navigation in the header of every page.
