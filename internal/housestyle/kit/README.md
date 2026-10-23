# The kit of the house style

The classes of the house style sheet (`kit.css`), the deck engine (`deck.js`), the templates of each kind of Window (`templates/`) and the guide that `get_house_style` gives to AIs (`guide.md`). Cahut serves them to the Windows of every Cahut, which take the look of their organization from the variables of its house style. See [ADR 22](../../../docs/adr/0022-house-style-of-each-cahut.md).

These files end up in the pages of organizations, so they are under the [Apache License 2.0](LICENSE), unlike the rest of the repository ([ADR 15](../../../docs/adr/0015-license.md)).

Their classes, their variables, the markup of the engine and its keys are a contract with the Windows already published: a release adds classes and refines how they look, and renames or removes none before a major version.
