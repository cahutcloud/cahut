# The formats of Windows

What makes a good Window of each kind. The template of each kind, which `get_house_style` gives, has the structure; this says what to put in it.

## Document

A report, a memo, a brief, a guide: something people read from top to bottom.

- Open with a kicker (the kind of document), a title that says the subject, and a lead of one or two sentences that says what the reader learns and why it matters.
- Put the conclusion first, in a callout of three findings at most, then the reasons.
- Make headings say something: "Revenue grew in the North", not "Results".
- Short paragraphs, one idea each. Lists for parallel items, tables for figures.
- In tables, align numbers right with `hs-num`, give units in the header, and round to what matters.
- Give each figure a caption that says what it shows and where the data comes from.
- End with the next steps, or the decisions to make, and who makes them.

## Deck

A presentation, shown on a screen or read alone.

- 6 to 15 slides. One idea per slide, and a title that states it: "North leads, West slows down".
- Open with a title slide (`hs-slide--title`), and a section slide (`hs-slide--section`) for each part of a long deck.
- 3 to 5 short points on a slide, never paragraphs. The details go in a document, linked from the last slide.
- Show numbers big: `hs-stat` for two or three key numbers, a chart for a trend or a comparison.
- Close on a statement slide (`hs-slide--statement`) with the one message to keep, or on the next steps.
- Speaker notes go in `<aside class="hs-notes">`.
- Text scales with the slide: give sizes in `em`, and check that every slide holds its content without overflowing.
- An image slide (`hs-slide--image`) needs an image whose license allows it, among the files of the Window.

## Dashboard

Figures to watch, read at a glance.

- 3 to 6 key numbers at the top (`hs-kpis`), each with its label, its value and its change over a stated period.
- Then panels, one question each, with a title that names it: "Revenue by region".
- Pick the chart that the question needs: bars to compare, a line for a trend over time, a table for exact values. One axis per chart; two measures of different scales make two charts.
- Give series the colors of the house style in their order (`hs-series-1`, then the next), never cycled, and label them in a legend or directly.
- Say the date of the data and its source, in the header or the footer.
- Charts with scripts declare their library's origin. Without scripts, use inline SVG or the bars of the house style.

## Site

A few pages linked together, such as a guide or an event page.

- Start each page from the template of a document, with the same header and footer.
- `index.html` at the root opens first. Link the pages with relative paths: `about.html`, `guide/setup.html` and, from there, `../index.html`.
- Put a short navigation in the header of every page.
