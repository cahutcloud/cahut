# 23. The house style can be inactive for AIs

- Status: Superseded by [ADR 49](0049-no-house-style.md)
- Date: 2026-10-23

## Context

[ADR 22](0022-house-style-of-each-cahut.md) gives each Cahut a house style: a name, four colors and those of charts, two fonts, two logos and guidance for AIs, with templates and a deck engine. The instructions of the MCP server and the skill for Claude tell every AI to read it with `get_house_style` before it makes a Window, and to start from its templates.

That small core suits an organization without brand guidelines of its own: Cahut gives it a look that holds together, which it can make in a few minutes. An organization that has full guidelines, with more colors, visuals, icons and layouts, already gives them to its AIs, often as a skill of its own. For it, the house style of Cahut competes with its guidelines: the AI is told to start from Cahut's templates, and mixes the two.

## Decision

- **A house style is active or inactive.** It is active by default, as ADR 22 has it. The administrators of the Cahut, and their AIs that reach every Room, make it inactive or active again with `set_house_style` (`active`), or with a switch in the console. Turning it off or on is a change like any other: the house style before becomes the previous one, and it does not knock first.
- **Inactive, AIs do not use it.** `get_house_style` still answers an AI, so that it learns where it stands, but it gives nothing of the house style: no colors, no fonts, no logo, no guidance, no template and no guide. It says that the house style is inactive, and that the Window takes the style that the user asks for. The instructions of the MCP server and the skill say so.
- **People still see it.** In the console, the administrators keep the whole house style, to change it or to turn it on again.
- **What is served stays.** The style sheet, the logos, the fonts and the templates under `/-/` keep being served. The Windows that link the sheet keep their look, and the pages of Cahut, such as the page of a Key, keep the name, the logo and the colors of the house style. An organization with guidelines of its own can thus give Cahut its name, logo and colors for these pages, and leave the rest to its own skill.
- **The Logbook** records turning it off and on, under `set_house_style`.

## Consequences

- An organization chooses between two ways: the house style of Cahut for every AI, or its own guidelines, given to its AIs as it sees fit. The second way holds only for the AIs that have those guidelines; the others make Windows in their own style.
- The tool stays in the list of every AI. Removing it, and the paragraph of the instructions on the house style, would mean making the tools and the instructions of each Cahut again at each change, and the MCP sessions already open would keep the old ones.
- A way to make a house style quickly, for those who have none, belongs with the guided first start of batch 7.

## Alternatives considered

- **Removing the tool and the instructions when the house style is inactive.** The AI would not even know that Cahut has a house style. The tools and the instructions of a Cahut are made once, when it opens: each change would have to make them again, and the sessions already open would not follow.
- **Extending the house style** with colors in any number, a library of visuals and icons, and longer guidelines. It would bring more organizations in, at the cost of a model that never covers every set of guidelines. Organizations with full guidelines keep them; the house style stays a small core for the others. It may still grow after the V0.
- **Neutral as the off state.** Going back to the neutral house style still gives AIs a style to follow, which is what an organization with its own guidelines does not want.
