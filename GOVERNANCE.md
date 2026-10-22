# How Cahut is governed

Cahut is a personal open source project by Olivier Durand, independent of any company and of any AI vendor. One person maintains it today. This page says how decisions are made, who does what, what the project promises, and how this changes as the project grows. The reasons are in [docs/gouvernance.md](docs/gouvernance.md), in French.

## Principles

1. **The maintainer decides, and writes why.** Olivier has the final say. Every decision that matters becomes a record in [docs/adr](docs/adr/README.md), with the options set aside.
2. **A no is temporary, a yes is forever.** A feature taken in is maintained for years. A feature declined can come back with better arguments.
3. **Open to read, bounded to contribute.** Everything can be read, run, tested and reported. Fixes come in directly, features are discussed first, and some things will not be done. [CONTRIBUTING.md](CONTRIBUTING.md) says which.
4. **Time is the scarce resource, and we say so.** One fixed slot a week, an answer within a week, no promise beyond that. [SUPPORT.md](SUPPORT.md) says what to expect.
5. **Honest about the state of the software.** The version number, the status page of the documentation and the release notes say what works, what breaks and what changes. Cahut is 0.x: anything may change before 1.0.
6. **The license does not change, and there is no closed edition.** Cahut is under the AGPL-3.0 ([ADR 15](docs/adr/0015-license.md)). The same code serves everyone, self-hosted or hosted by someone else. There is no contributor license agreement: contributors keep their copyright, so nobody, the author included, can relicense Cahut once outside contributions are in.
7. **The name is protected, the code is free.** "Cahut" and its logo identify this project ([TRADEMARK.md](TRADEMARK.md)). A fork takes another name.
8. **Security comes first.** A vulnerability report interrupts everything else ([SECURITY.md](SECURITY.md)).
9. **Nothing leaves your server quietly.** The only outgoing request of the software is the check for new releases on GitHub, documented and optional. There is no telemetry.
10. **AIs are welcome, humans answer.** A contribution made with an AI is disclosed, understood and tested by the person who signs it.
11. **The core is not up for a vote.** The product rules of the README (bounded operations, knock first, one API, no AI on the server, everything logged, isolation) and principles 6 to 9 change only with the agreement of the lead maintainer and of every maintainer, through a decision record.

## Roles

| Role | Who | What they do | How one gets there | How one leaves |
| --- | --- | --- | --- | --- |
| User | Anyone who installs Cahut or reads a Window | Read, report, propose, help others | By showing up | |
| Contributor | Anyone whose work was taken in: code, documentation, translation, steady help with triage | All of the above; named in the release notes | A merged pull request, or steady help | |
| Triager | A regular contributor who helps sort issues | Label, close duplicates, ask for details, answer in the Discussions, with GitHub's triage permission | Invited by the lead maintainer after a few months of help | Six months without activity: removed, with thanks |
| Maintainer | A person who reviews and merges | Review, merge, publish releases, a voice in decisions, with write permission; listed in MAINTAINERS.md once there are two | Invited after six months of steady contributions and the trust of the lead maintainer | May step down and become emeritus; a year without activity: emeritus |
| Lead maintainer | Olivier Durand ([@odura](https://github.com/odura)) | All of the above; breaks ties; holds the name, the domain and the GitHub organization | Founder | Plans the succession, below |

Roles are offered, never claimed. Asking for one insistently is the surest way not to get it.

## Decisions

- **Day to day**, the maintainer decides in issues and pull requests, and says why in a sentence.
- **Anything that commits the architecture, the product or the vocabulary** goes through a decision record. Anyone may propose one, as a pull request with the status *Proposed*. It stays open to comments for one to two weeks when it changes something important. Then the lead maintainer accepts or declines it, with the options set aside.
- **The roadmap is public and has no dates**: the milestones of this repository, one per batch of the V0 plan, and the status page of the documentation. Votes on ideas inform; they do not decide.
- **With several maintainers**, ordinary decisions pass by lazy consensus: a proposal stays open for seven days, and passes without objection. An objection says why and proposes a way out. The lead maintainer breaks ties. Every maintainer declares their conflicts of interest, such as working for a hosting company.
- **The core** (principle 11) changes only with everyone's agreement, through a decision record.

## Succession

A project of one person must survive that person. A trusted person, who need not write code, is a second owner of the GitHub organization and holds emergency access to the domain and to the release signing key, with written instructions. If the lead maintainer is unreachable for six months, or asks for it, that person publishes a notice in the repository, names a maintainer among the known contributors, or hands the project to a structure, as the instructions say. The trusted person is named here once chosen.

Trust is earned, never granted under pressure. Two-factor authentication is required on the organization, `main` is protected, nobody receives rights for insisting, and every promotion follows a period of observation. Once there are two maintainers, a change that touches security or releases is reviewed by two people.

## How the project grows

| Phase | Starts when | What changes |
| --- | --- | --- |
| 1 · The repository is open | The repository opens | A weekly slot for triage and reviews, a release per batch, news on the site. The lead maintainer decides alone. No chat, and no support beyond the Discussions |
| 2 · Regular contributors | Two people with three merged pull requests in three months, or triage that overflows the slot | Triagers, MAINTAINERS.md, the trusted person confirmed, a second maintainer for sensitive reviews and for the code of conduct; perhaps a chat, with two people to keep it |
| 3 · A team | Three maintainers active for six months, or money above the costs, or a hosted offer | Lazy consensus, a second version of this page, a structure to hold the name and the domain, a clear line between the project and any commercial activity |

What does not change: the core, the license, and the time each person gives.

## Money

Donations pay the domain, a test server, the trademark and a few tools. They do not pay a salary, and we say so. The project takes them through GitHub Sponsors, and through Liberapay once its account exists. A hosted offer or paid support, if they come, run the same binary as everyone else, with no reserved feature, and stay apart from the project.

## Changing this page

This page changes through a pull request, and through a decision record for anything beyond wording. The French version, with the reasons and the sources, is [docs/gouvernance.md](docs/gouvernance.md).
