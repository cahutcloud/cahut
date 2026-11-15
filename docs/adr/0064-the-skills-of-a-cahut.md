# 64. A Cahut keeps the skills of its team and of each person, and every AI connected to it finds them in its list of tools

- Status: Accepted
- Date: 2026-11-15

## Context

A skill is a folder of instructions that an AI follows for some tasks, in the format of Agent Skills: a `SKILL.md` with a YAML header that gives its `name` and its `description`, then its instructions in Markdown, and files beside it, such as references, fonts, logos and templates. The AI clients that connect to Cahut all read this format, and each keeps skills its own way: in the account, in a plugin, in a folder of the computer, in a workspace. A team whose people use different AIs keeps no common set: each change asks each person to import the skill again, into each AI.

[ADR 49](0049-no-house-style.md) removed the house style, and left the brand of an organization to a module of skills. The brand is one case among many: a team wants its tasks done its way, an e-mail that reminds a client of an invoice, a report, a deck, whether the task publishes a Window or nothing at all.

What reaches the model of each client without anyone asking:

- **The list of tools**, in every client. Claude Code loads the names of the tools first, then the description of a tool when it has a reason to look at it.
- **The instructions of the MCP server**, in Claude, ChatGPT and Codex, cut short: claude.ai keeps 2,048 characters, Codex 512. Mistral Vibe 2.1.0 does not give them to its model at all.
- ChatGPT keeps a copy of the list of tools, and holds back a new tool until a person approves it. The list must stay fixed: a tool for each skill would not reach it.

Three trials, on 14 and 15 November 2026, with Mistral Vibe 2.1.0 and Claude Code 2.1.322, each in a blank folder, with skills kept by a Cahut:

- **A task that publishes nothing**, an e-mail that reminds a client of an invoice, the skill for it among others:
  - with nothing that points to the skills, neither AI read one: 0 of 10 for each;
  - with a pointer in the instructions and in the description of a tool `find_skills`, Claude Code read it 10 times of 10, and Mistral Vibe never;
  - with the catalogue of the skills, the name and the description of each, in the description of `find_skills`: Mistral Vibe 50 of 50 and Claude Code 30 of 30, with twenty skills of which five close to the task, and with a vaguer request;
  - with the skills installed in each client, for comparison: 10 of 10 for each.
- Every skill that an AI read, it followed: 130 times of 130. No AI read a skill that did not fit, and none called the Cahut when no skill applied.
- **A deck to publish**: a refusal of `publish_window` that carried the brand of the organization stopped Mistral Vibe 10 times of 10, and each deck it then published followed the brand.

Thirty successes of thirty say that the rate is above 90%, not that it is 100%. The brand of an organization is not left to chance where it need not be.

The list of tools that MCP gives an administrator who may write stays under 40 KB ([ADR 57](0057-the-list-of-tools-under-40-kb.md)). On 15 November 2026 it weighs 37,336 bytes, for 41 tools.

## Decision

### A skill, and where it lives

- A skill of a Cahut is a skill in the format of Agent Skills, as it is. Cahut reads `name` and `description` from its header, keeps the other keys without acting on them, and keeps its files under any path but `scripts/`.
- **Three scopes**:
  - `cahut`: the skills of the team, which the administrators of the Cahut write, for every person of the Cahut and their AIs, those that only read included;
  - `personal`: the skills of a person, which they alone write and read, with their AIs. The administrators neither read nor change them, as they do not open the Windows of the Rooms of others ([ADR 13](0013-people-sessions-and-rights.md));
  - `builtin`: the skill `cahut` of the plugin for Claude ([ADR 22](0022-house-style-of-each-cahut.md)), served from the binary and read only, so that the AIs that install no zip read it too.
- The skills of team Rooms come later, in a record of their own.
- **An individual Cahut** ([ADR 30](0030-kind-and-limits-of-each-cahut.md)) has no skills of the Cahut, only the skills of each person. Its administrators host the service: they do not give instructions to the AIs of the people who signed up.
- **A server has no skills of its own** for all its Cahuts: a Cahut reads nothing of another ([ADR 8](0008-several-cahuts-per-server.md)).
- **Names.** A name is unique in its scope, and does not change: a skill that needs another name is written under it, and the old one deleted. The same name may live in two scopes, since refusing it would tell an administrator the names of the personal skills of someone. A name that two scopes hold takes its scope as a prefix, `cahut:charte-acme`, and the refusal of a name without it says so. The name `cahut` is kept for the skill of the plugin.
- **Two versions**, as a Window has ([ADR 19](0019-window-versions-access-and-declaration.md)). Each write makes a new version, the previous one stays, and a rollback brings it back. Each version has a number, for people, and a digest, for AIs: the first 8 hexadecimal digits of the SHA-256 of its `SKILL.md` and its files, by path.
- **Limits**: a `SKILL.md` of 64 KiB; a file of 2 MiB; 100 files and 4 MiB for a skill; 50 skills in a scope. The disk of a Cahut counts both versions of each skill.
- **No `scripts/`.** A script in a skill of the Cahut is code that some clients run on the computer of each member, and that an administrator changes from afar. Cahut refuses a skill that holds one, and says why.
- **The header.** Cahut reads it with `go.yaml.in/yaml/v3`, and keeps it as it came. When Cahut gives a `SKILL.md`, to an AI or in an export, it writes `name` and `description` in double quotes: Mistral Vibe drops without a word a skill whose description holds a colon and a space unquoted.
- Cahut runs nothing of a skill. It gives its files to the AI, which does with them what its client allows.
- Skills live in the database of each Cahut, their files included: a write is undone with its transaction, and the snapshots of updates ([ADR 24](0024-signed-releases-and-updates.md)) carry them.

### How AIs find skills

- AIs get two tools, both read only, offered to every AI, those that only read included:
  - `find_skills` lists the skills that the person reads: those of the Cahut, their own, and the skill of the plugin, each with its name, scope, description, version and mark. With `query`, it finds those whose name, description or text hold some words, whatever their case and accents.
  - `get_skill` gives a `SKILL.md` whole, its scope, the number and the digest of its version, who wrote it, when and from where, and the list of its files. A short note gives the rule of priority, `from_skill`, and for a skill marked for every Window, what to give in `follows`. With `file`, it gives that file.
- **The catalogue.** The description of `find_skills` carries a line for each skill of the Cahut: its name, then its description cut at 160 characters on a word. The skills marked for every Window come first, then the others by name, within a fixed budget of bytes. The lines that do not fit give way to a single one: "and 12 more skills: find_skills lists them". The console shows administrators the catalogue as AIs receive it, and the skills left out.
- The catalogue leaves out the personal skills and the skill of the plugin, which `find_skills` gives: a catalogue for each person would mean a list of tools for each person.
- **The catalogue follows the skills.** Each Cahut keeps a revision of its catalogue, which each write, rollback, mark or deletion of a skill of the Cahut raises in its transaction; a personal skill leaves it as it is. The MCP server of a Cahut is kept for each set of tools and each revision, and each request reads the revision, as it reads the kind of the Cahut. The servers of past revisions are forgotten. A session opened before a change keeps the list it had.
- An individual Cahut has no catalogue: the description of `find_skills` is the same there as in a Cahut without skills.
- **A pointer in the instructions.** The second paragraph of the instructions of the MCP server opens with a sentence that sends an AI to `find_skills` before a task, any task. It lies within the 2,048 characters that claude.ai keeps, and the first paragraph, on what to publish, does not change ([ADR 29](0029-chatgpt-second-ai-client.md)). Claude Code needs it: without it, it never looked at the description of `find_skills`.
- **Two ways that do not depend on the AI.** A person who names a skill ("follow the skill relance-facture") gets it from `get_skill`, in every AI. The console also gives a line to paste into the instructions of an AI that does not think of the skills by itself.

### The rule of priority

`get_skill` gives it with each skill, and the skill of the plugin says it:

1. what the person asks in the conversation;
2. their personal skill;
3. the skill of the Cahut;
4. a skill for the same use installed in the client: it is a copy, and the skill of the Cahut has the last word.

### The gate of Windows

- The administrators of a Cahut mark up to three of its skills **for every Window**, such as its brand, the notices of its forms and its conventions. The `SKILL.md` of a marked skill holds 16 KiB at most, since a refusal carries it whole.
- `publish_window` and `update_window` take `follows`: the skills that the Window follows, each as `get_skill` gives it, its name and the digest of its version, `charte-acme@3f9a2c1b`.
- A call of an AI whose `follows` lacks the current version of a marked skill is refused, as an invalid call, and publishes nothing. The refusal carries, for each skill missing, its `SKILL.md`, the list of its files and the value to give in `follows`, and asks the AI to make the Window again as the skill says. The next call, with `follows`, passes.
- **A digest, not a memory of the connection.** A token or a connection over OAuth lasts weeks and serves many conversations: a gate that remembered what a connection received would let its next conversation through without the skill in front of it. The digest comes only in the answers of `get_skill` and in the refusal, so an AI that gives it received the skill. The gate keeps nothing between two calls, and holds for each conversation.
- A new version of a marked skill changes its digest: the next publication is refused once, and receives the new text.
- The gate stops AIs only. A person who publishes is never stopped.
- It makes sure that the AI had the skill in front of it, not that the Window follows it.

### The files of a skill in a Window

- A file given to `publish_window` or `update_window` may name `from_skill` in place of its `content`: the name of a skill, with its scope as a prefix if need be, then the path of a file in it, `charte-acme/assets/fonts/inter.woff2`.
- Cahut copies the bytes of the current version of the skill into the version of the Window, in the transaction of the publication, if the AI may read the skill. The limits of Windows count the copied files.
- What ADR 49 decided holds: a Window carries its own files, and Cahut serves nothing under `/-/` for Windows to link. A published Window does not change when its skill changes: an AI publishes it again to follow the new version.

### AIs read skills, and do not write them

- `write_skill`, `import_skill`, `export_skill`, `rollback_skill` and `delete_skill` serve the console only (`PeopleOnly`): people write skills, AIs read them.
- A skill is a lasting instruction, which every AI of a person, or of a whole Cahut, then follows. An AI misled by a Window, a web page or an e-mail must not be able to leave one behind. A Knock ([ADR 21](0021-ais-knock-first.md)) would make that harder without preventing it, since people approve quickly what they do not read. An AI still drafts a `SKILL.md`, which a person imports.
- The rule of one API holds as ADR 57 reads it: an operation that a person alone needs stays out of MCP. Writing by AIs, behind a Knock that shows the whole text and what changes, is left to a later record.
- `import_skill` takes a call of 6 MiB on the API of the console, for a zip of 4 MiB in base64, where other operations take 1 MiB. Its registration says so, as it says which operations serve visitors.
- Cahut reads a zip with its `SKILL.md` at its root or in a single folder, as clients export them. It checks each path and each limit as it reads, refuses `..`, absolute paths, links and duplicates, and never holds more than the limit of a skill.
- No command line in this batch: the console is enough until a need comes.

### The Logbook

- A write, an import, a rollback, a mark or a deletion keeps the scope, the name, the version, and the number and weight of the files, never their content.
- `find_skills` keeps that a search took place and the number of results, never its words. Administrators read the whole Logbook, and searches such as "redundancy plan" would tell them the tasks of others.
- `get_skill` keeps the skill and the file read: administrators see which skills serve, and which AIs read them.
- A publication keeps the skills of `follows` and the files copied with `from_skill`. A refusal of the gate keeps the skills that were missing.

### The list of tools

- The skills add about 2 KB to the list of an administrator who may write: `find_skills` and `get_skill`, then `follows` and `from_skill` in `publish_window` and `update_window`. Under 40 KB, the catalogue has a budget of about 1,200 bytes, about six skills.
- As ADR 57 asks, the list is tried at 48 KB with claude.ai, Claude Code, ChatGPT, Codex and Mistral Vibe before its bound rises. If each of them passes, a later record raises the bound to 48 KB, and the catalogue takes 6,000 bytes, about 33 skills. If one fails, the bound stays at 40 KB, and the other skills are found with `find_skills`.

### Records

- This record extends [ADR 49](0049-no-house-style.md): the brand of an organization comes from its skills, which the gate and `from_skill` bring to its Windows.
- It extends [ADR 30](0030-kind-and-limits-of-each-cahut.md): the disk of a Cahut counts its skills, and an individual Cahut has no skills of the Cahut.

## Consequences

- A team keeps one set of skills, whatever the AIs of its people. A change reaches the next conversation of each AI, and nobody imports anything again.
- A skill exported from a Cahut installs as it is in the clients that read Agent Skills, and a skill exported from Claude imports as it is.
- **What is certain**: a skill that the person names, and the gate for the skills marked for every Window. **What is measured, not certain**: an AI that thinks of the right skill of the Cahut by itself, as often as of a skill installed in it. claude.ai, ChatGPT and Codex are tried at the acceptance of the batch.
- ChatGPT keeps its copy of the list of tools: a new skill may wait until a person refreshes the connector. The acceptance tells, and the documentation says it.
- The gate costs a Window made for nothing, once in a conversation that had not read the skill: about 10 KB and 15 seconds in the trial. The catalogue and the pointer avoid it most of the time, since `get_skill`, read first, gives the digest.
- A skill of the Cahut gives instructions to the AIs of everyone in it, and the account of an administrator can be stolen. Every person sees the skills of the Cahut and who wrote them, in the console and in `get_skill`, and the Logbook keeps every write. AIs receive a skill as the result of a tool. This does not guard against an administrator who means harm, who can already add people and change their rights.
- Every client that loads the list whole carries the catalogue, in each conversation. Its budget is fixed, and the test of ADR 57 measures the list with a full catalogue.

## Alternatives considered

- **The list of skills in the instructions of the MCP server.** claude.ai cuts them at 2,048 characters and Codex at 512, and they do not reach the model of Mistral Vibe.
- **An MCP server for each person**, with a catalogue of their own skills. More servers to keep, for personal skills alone. It can come later, if a client needs it.
- **A tool for each skill.** ChatGPT keeps its copy of the list, and holds back each new tool until a person approves it.
- **The prompts and resources of MCP.** A person starts a prompt, and ChatGPT, Codex and Mistral Vibe do not show them; no client puts a resource before its model by itself.
- **The extension of MCP for skills.** Its specification is final, but almost no client reads it yet. Cahut can serve it once Claude or ChatGPT do.
- **A gate that remembers each connection**, as the trial had it. A connection serves many conversations.
- **The line to paste as the main way.** It works, but asks each person for a step in each AI, and the catalogue does as well without it.
- **Skills installed in each client, fed by the Cahut**, through a marketplace of plugins or a git mirror. They do not trigger better than the catalogue, and each vendor has its own mechanism, often on its paid plans. They can come later, as a comfort.
- **Writing by AIs, behind a Knock.** The Knock would show a long text that people approve without reading. It waits for a record of its own.
- **A skill as a new kind of Window.** A Window is a page served to readers in a sandbox: `publish_window`, the access of Windows, share links, their data and their cards would all need exceptions.
- **Windows that link the files of a skill**, under `/-/skills/`. A Window that moves would lose its look, and a change of the skill would change every published Window at once. ADR 49 removed such links.
