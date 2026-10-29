# 33. The pages of Cahut carry its brand until the house style has a look of its own

- Status: Superseded by [ADR 49](0049-no-house-style.md)
- Date: 2026-10-24

## Context

[ADR 32](0032-the-page-of-a-room.md) gives the pages that readers see the brand of Cahut while the Cahut has the neutral house style, and the house style once an administrator sets one. The first version read "neutral" as get_house_style does: a house style that no administrator has changed in any way.

A house style holds more than a look. Its guidance goes to AIs and never to readers, and [ADR 23](0023-house-style-inactive-for-ais.md) lets an administrator make it inactive for AIs, for an organization that gives them guidelines of its own. The staging Cahut has its house style inactive for AIs, and nothing else: its pages lost the brand of Cahut for the neutral look, in the fonts of the system, though nothing a reader sees had changed.

## Decision

- The pages that readers see carry the brand of Cahut while the house style keeps the look of the neutral one: no name, no colors, chart colors included, no fonts and no logo.
- The guidance for AIs, and whether AIs use the house style, change nothing for these pages.
- As soon as the house style has a name, colors, fonts or a logo, the pages carry it, as ADR 32 has it.
- get_house_style keeps its `neutral`, which says whether an administrator has changed the house style at all: AIs need to know that a house style is inactive, whatever its look.

## Consequences

- A Cahut whose house style only speaks to AIs, or keeps them away from it, shows its readers the brand of Cahut.
- An organization that gives Cahut only its name sees its name at the top of the pages, in the neutral look, without the brand of Cahut; it gives its colors and its logo to look like itself.
- This record extends ADR 32.

## Alternatives considered

- **The neutral house style as get_house_style reads it**, as the first version did. A switch for AIs would change what readers see.
- **The name of the organization beside the brand of Cahut.** The brand guidelines never put the logo of Cahut beside the name or the logo of an organization: a name is the start of a look of its own.
