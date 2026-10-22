# 18. The project pages are built from the community files of the repository

- Status: Accepted
- Date: 2026-10-22

## Context

How the project is run is written down in [docs/gouvernance.md](../gouvernance.md), in French, and summed up for GitHub and for contributors in the community files at the root of the repository, in English and short: `GOVERNANCE.md`, `CONTRIBUTING.md`, `CODE_OF_CONDUCT.md`, `SECURITY.md`, `SUPPORT.md` and `TRADEMARK.md`. GitHub shows them where they matter: when someone opens an issue or a pull request, in the Security tab, in the sidebar of the repository.

The website should say the same things to the people who have not opened the repository: how Cahut is run, how to help, where to report a vulnerability, which rules apply, what may be done with the name, and where the project stands. [ADR 12](0012-website-and-docs-as-astro-sites.md) put the site in this repository, as an Astro site with a French version under `/fr/`, and gave it the brand through the `@charte` import alias: the brand has one source, and the site never copies it.

Two sources for one text drift apart, and one person keeps this project. The question is where the project pages take their words.

## Decision

- The site has a project section at `/project/`, not a subdomain: `/project/` for how Cahut is run, then `/project/contribute/`, `/project/security/`, `/project/code-of-conduct/`, `/project/support/` and `/project/roadmap/`. The policy on the name and the logo is a section of the brand page, at `/brand/#name-and-logo`. The header and the Project column of the footer lead to them.
- The community files at the root of the repository are the single source of these pages in English. The site imports them at build time, as Markdown, through the `@repo` import alias, which points to the root of the repository as `@charte` points to the brand. Nothing is copied into `sites/`.
- A plugin of the site's Markdown processor adapts each file to the page it becomes: the title of the file becomes the title of the page, a link to another community file leads to its page, and a link to any other file of the repository opens that file on GitHub. Under its content, each page names the file it comes from.
- The French pages under `/fr/project/` are translations kept in the site, in `sites/cahut.cloud/src/project/fr/`, one Markdown file per community file, with the same name, written like the original and rendered the same way. The English file is the reference; a translation follows it, and may lag behind it.
- The code of conduct in French is the official translation of the Contributor Covenant 3.0, with the reporting paragraph of the English file. It is under CC BY-SA 4.0, as the English text is.
- The roadmap page is written in the site, in both languages, from the batches of the [V0 plan](../plan-v0.md) and from the status page of the documentation, without dates: neither is a public English file at the root.
- The brand page stays in English, so the policy on the name and the logo has no French version.

## Consequences

- A change to a community file reaches the site at its next build, in the pull request that changes it, with no page to edit. The French translation changes in the same pull request when the change is more than wording, or follows later.
- The community files keep their shape for GitHub: Markdown, links as paths from the root of the repository, the title first. The site adapts to them, not the other way round.
- The Markdown processor of the site is configured, so `@astrojs/markdown-satteri`, the processor that Astro ships, is a dependency of the site, at the version Astro pins. It brings no other package.
- A news page, built from the releases of the repository, is not part of this record. It would be built the same way, from GitHub, at the build that a release triggers.

## Alternatives considered

- **A subdomain, `community.cahut.cloud`.** A second site to keep, with its own header, footer and deployment, for six pages of text. A subdomain makes sense for a forum one day, not for these pages.
- **Pages written for the site.** Pages in the site's own words, pointing to the files. The nicest wording, and two texts to keep in step by hand: the first change to a file that forgets the site makes them disagree.
- **Links to GitHub only.** The site links to the files and stops there. Readers of a website rarely cross over to a repository, and there would be no French version.
- **A content collection.** Astro's content collections could load the files from the root. They bring a schema, slugs and queries for six files that are imported by name: more machinery than the pages need.
