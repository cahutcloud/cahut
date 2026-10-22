# Contributing to Cahut

Thank you for your interest. Cahut is maintained by one person, in the open. This page says what helps, what to discuss first, and how a change gets in. [GOVERNANCE.md](GOVERNANCE.md) says how the project is run; the [code of conduct](CODE_OF_CONDUCT.md) applies everywhere.

## Welcome right away

Open a pull request without asking first for:

- a bug fix, with a test that shows the bug;
- a better page of the documentation, or a fixed typo in the site or in a message;
- a translation of the site: the English words are in `sites/cahut.cloud/src/i18n/en.ts` and the French ones in `fr.ts`, changed together, and the French pages of the project section are in `sites/cahut.cloud/src/project/fr/`;
- a clearer error message, one that tells an AI how to fix its call;
- a dependency update.

## Discuss first

Open a Discussion in *Ideas*, or an issue, before writing code for:

- a feature;
- a new operation, and so a new MCP tool;
- a change to the interface or to the vocabulary;
- a new dependency, with its reason; a dependency must not need cgo;
- anything that touches security, the sandbox or knock first.

A thousand-line pull request that arrives unannounced is closed kindly, whatever its quality. This protects your time as much as ours.

## What Cahut will not do

The README says it: create content, run an AI model on the server, give an AI a shell, depend on one vendor, span several machines. Nor will it ship a reduced or closed edition, send telemetry, or reserve a feature to a paid offer. The documentation keeps these decisions and their reasons, so that the question is not asked twice.

## How a pull request gets in

1. One branch per topic, one pull request per topic, small when possible.
2. `make check` passes: lint and tests, as CI runs them.
3. The code follows [CLAUDE.md](CLAUDE.md): English everywhere, imperative commit subjects, every action through the operation registry, no business logic in the adapters, the standard library first. That file is written for the AIs working in this repository and applies to everyone.
4. The description says what changes, why, and how it was tested. For a feature, it links the discussion.
5. The maintainer reviews in a weekly slot. He asks for changes, merges, or explains why not.
6. A pull request without an answer from its author for thirty days is closed. Reopen it any time.

### Sign your commits

Cahut uses the [Developer Certificate of Origin](https://developercertificate.org/). Each commit carries the line `Signed-off-by: Your Name <you@example.com>`, added by `git commit -s`, by which you certify that you have the right to contribute this code under the project's license. Nothing is assigned, there is nothing else to sign, and no lawyer is involved. There is no contributor license agreement, and there will be none: that is how the license stays what it is ([ADR 15](docs/adr/0015-license.md)).

### Working with an AI

Cahut is written with Claude Code, and [CLAUDE.md](CLAUDE.md) is public. Contributions made with an AI are welcome under five rules:

1. **Say so.** Name the tool and how much it did, in the pull request or in a commit trailer such as `Assisted-by: Claude Code`.
2. **Understand all of it.** The person who signs has read, understood, run and tested the change, and answers reviews in person. A generated answer to a review closes the pull request.
3. **The sign-off is yours.** `Signed-off-by` certifies what you did, not what the tool did: an agent adds it only on your instruction and in your name, and no agent opens an issue or a pull request without a person having read it.
4. **Reproduce before you report.** An issue written by an AI without a reproduction is closed. A fabricated security report gets its author banned.
5. **Visuals come from the brand.** No generated image or illustration in the site or in the documentation.

## Reporting a bug

Use the bug report form. Give the version from `cahut version`, the system, how Cahut was installed, what you expected, what happened, and the lines of the Logbook that go with it, without any token. Questions go to the Discussions ([SUPPORT.md](SUPPORT.md)). Vulnerabilities go through [SECURITY.md](SECURITY.md), never through an issue.

## Recognition

The release notes name the contributors of each release. The labels `good first issue` and `help wanted` mark where to start. There is no bounty program.
