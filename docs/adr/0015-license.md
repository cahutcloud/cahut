# 15. License

- Status: Accepted
- Date: 2026-10-22

## Context

Cahut is a self-hosted server whose argument is control: the data stays on the user's server, and the code can be read. The README has carried "Apache-2.0 or AGPL-3.0" since the brief, and without a LICENSE file everything is all rights reserved. The repository opens before the public launch, so the license must be settled first.

The project makes three promises, laid out in [docs/gouvernance.md](../gouvernance.md): the same code for everyone, self-hosted or hosted by someone else; no closed edition; a license that does not change. It takes contributions under the Developer Certificate of Origin, without a contributor license agreement. A hosted offer run by the author may come later, with the same binary. Hosting partners must be able to install and distribute Cahut as it is.

## Decision

- **The GNU Affero General Public License, version 3 only** (SPDX `AGPL-3.0-only`), for the whole repository: the server, its command line, the install script, the sites and the documents, unless a directory states another license.
- **The text of the license, verbatim, in `LICENSE`.** The README says what it means in plain words.
- **The name and the logo are not licensed.** They identify the project. The README says so, and a trademark policy will say it in full before the repository opens. A fork takes another name.
- **Files meant to be copied into an AI client**, the skills and the templates of the house style when they come (batch 6 of the V0 plan), carry Apache-2.0, stated in their directory: what people copy into their own tools must not bind those tools.
- **No contributor license agreement.** Contributions are certified with the DCO and stay under the AGPL-3.0, with their authors' copyright. This makes the promise structural: once outside contributions are in, nobody can relicense Cahut without the agreement of every contributor, the author included.
- **Version 3 only.** The terms are those of the text in the repository, not those of a future version of the license.

## Consequences

- Anyone can run, study, change and share Cahut. Whoever changes it and lets others use the changed version over a network gives them the changed source under the same license. Running or distributing Cahut as it is asks nothing more than keeping the license with it: hosting partners and one-click installers are free, as they are for Nextcloud.
- A closed fork, or a closed improved service built on Cahut, is not possible. Neither is a dual-license model where the author sells exceptions to the AGPL: it needs the whole copyright, which the project does not collect.
- Dependencies must be compatible with the AGPL-3.0. MIT, BSD, Apache-2.0 and MPL-2.0 are; code under a license the GPL does not accept is not. The direct dependencies of today, CertMagic, the MCP Go SDK, jsonschema-go, zap, the `golang.org/x` modules and modernc.org/sqlite, are under Apache-2.0, MIT or BSD licenses.
- Some organizations ban AGPL code by policy, mostly for libraries and internal services. Cahut is a server installed as it is, which those policies rarely cover; the README's plain-words paragraph answers the usual question.
- The documentation and the sites are under the same license for now. A Creative Commons license for the documentation is a decision for the opening of the repository, together with the trademark policy.
- Release archives carry the LICENSE file: GoReleaser includes it by default.

## Alternatives considered

- **Apache-2.0.** Permissive, with an explicit patent grant and no reserve anywhere. It would let anyone, the author included, ship a closed edition or a closed improved service, which is the opposite of the project's promise. Kept for the skills and the templates, where it is the right fit.
- **MIT.** Shorter than Apache-2.0, without the patent grant; the same objection.
- **GPL-3.0.** Copyleft without the network clause: a service running a changed Cahut would owe its users nothing. The network clause is the point for a server.
- **AGPL-3.0-or-later.** The FSF's recommendation, and Nextcloud's choice. It would let a future version of the license change the terms; the project prefers terms that only its contributors can change.
- **Source-available licenses** (BSL, SSPL, FSL). They forbid competing services and are not open source. HashiCorp, Redis and Elastic moved to them against their communities and got forks; two of the three have since added the AGPL.
