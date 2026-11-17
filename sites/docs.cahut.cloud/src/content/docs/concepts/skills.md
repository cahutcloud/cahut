---
title: Skills
description: The skills of a Cahut tell every AI connected to it how the team does some tasks, such as its brand or the reminder of an invoice. People write them once in the console, and each person's AI can read them, whatever AI it is.
---

A skill tells an AI how to do a task: the brand of a team, the reminder of an unpaid invoice, a way to check a contract. Each Cahut keeps the skills of its team, and those of each person, and gives them to every AI connected to it over MCP: Claude, ChatGPT, Mistral Vibe or another. Nobody installs anything in their AI, and the tools give the current version of each skill at once. See [ADR 64](https://github.com/cahutcloud/cahut/blob/main/docs/adr/0064-the-skills-of-a-cahut.md).

## What a skill is

A skill is a folder in the format of Agent Skills, which AI clients read: a `SKILL.md`, whose header gives its name and its description, then its instructions in Markdown, and files beside it, such as references, fonts, logos and templates.

```markdown
---
name: relance-facture
description: The e-mail that reminds a client of an unpaid invoice: subject, tone, delay and signature. To follow for any reminder of a payment.
---

# Reminder of an unpaid invoice

The subject starts with [Lumen · Reminder 1]. Give the reference of the invoice, its date and its amount, ...
```

- The name is 1 to 64 lowercase letters, digits and hyphens, such as `charte-acme`. It does not change: a skill that needs another name is written under it, and the old one deleted.
- The description says what the skill does and when to follow it, in 1,024 characters at most. AIs choose a skill by its description: write it for them.
- A skill of a Cahut holds no `scripts/` folder. Some AI clients run its scripts on the computer of each person, and a script would be code that someone changes from afar: say in `SKILL.md` what the AI should do instead.
- Cahut runs nothing of a skill. It gives its files to the AI.

## Whose skills

The **Skills** screen of [the console](/admin/console/) shows three kinds:

| Part | Who writes them | Who reads them |
| --- | --- | --- |
| Skills of this Cahut | Its administrators | Every person of the Cahut, and their AIs |
| My skills | Each person | That person and their AIs alone, not even the administrators |
| Comes with Cahut | Nobody: it comes with each version of Cahut | Everyone. It is the skill `cahut` of the [plugin for Claude](/connect/claude-plugin/), which tells AIs how to publish Windows |

An [individual Cahut](/concepts/cahuts/#collaborative-or-individual) has no skills of the Cahut, only those of each person: its people share the Cahut, not their work.

The same name may live in two parts, such as a skill `charte-acme` of the Cahut and one of your own. The AI then names each with its part as a prefix, `cahut:charte-acme` or `personal:charte-acme`.

## Write, import and export

On the **Skills** screen, for the skills of the Cahut if you administer it, and for your own:

- **Write a skill** takes its `SKILL.md`, typed in the page, and its files, added from your computer: what the AI reads in `references/`, and what it copies, such as fonts or a logo, in `assets/`. The page sends 1 MiB at most: import a larger skill as a zip.
- **Import a zip** takes a skill as AI clients export them, with its `SKILL.md` at the root of the zip or in a single folder. A skill exported from Claude imports as it is, unless it holds a `scripts/` folder or hidden files.
- Each write or import that changes the skill makes a new version. Cahut keeps two, the current one and the one before.

On the page of a skill:

- **Write a new version**, and the button that rolls back to the previous one.
- **Export** gives a zip of the skill, in a folder of its name, which installs as it is in Claude Code, Mistral Vibe and the other clients that read Agent Skills.
- **Delete** takes the skill and both its versions, for good.

AIs read skills, and never write them. A skill is a lasting instruction, which every AI of a person, or of a whole Cahut, then follows: an AI misled by a Window, a web page or an e-mail must not leave one behind. Ask your AI to draft a `SKILL.md`, then paste it in **Write a skill**, or import it in a zip.

## How AIs find them

Every AI connected to a Cahut gets two tools, both read only: `find_skills` lists the skills that its person reads, and finds them by their words, and `get_skill` reads one whole, or one of its files.

- **The catalogue.** The description of `find_skills` carries a line for each skill of the Cahut: its name, then its description cut at 160 characters. Their client receives it from the start of each conversation, before they call anything. The skills for every Window come first, then the others by name, within 6,000 bytes: about thirty skills. Those that do not fit give way to a line that says how many more `find_skills` lists.
- **What AIs see.** On the **Skills** screen, administrators unfold **What AIs see**: the lines as AIs receive them, the bytes they take, and the skills left out.
- The catalogue names the skills of the Cahut only. `find_skills` lists the personal skills, and the skill that comes with Cahut.
- In a Cahut that has skills of its own, the instructions of the MCP server send each AI to `find_skills` before a task, any task, even one that publishes nothing. Elsewhere, they tell it to read a skill when the person names one.
- `find_skills`, `get_skill` and the gate give the current version of each skill at once. The catalogue follows a change at the next session of the client: a conversation keeps the list of tools it began with. ChatGPT keeps its own copy of the list of tools: a new skill may wait in its catalogue until you refresh the app of Cahut in its settings, though `find_skills` lists it already.

Two ways work in every AI, whatever it does by itself:

- **Name the skill.** "Follow the skill relance-facture": the AI reads it with `get_skill`.
- **Paste a line in the instructions of your AI**, which the **Skills** screen gives, ready to copy: "Before each task, call the find_skills tool of Cahut to see whether my team has a skill that applies to it, then follow it." In an individual Cahut, the line names your own skills.

## What each AI does

Whether an AI thinks of a skill by itself, for a task that does not name Cahut, depends on its client. In November 2026, with a Cahut that held fifty skills and a request to write the reminder of an unpaid invoice, which one of them describes. The catalogue of that trial named 48 of the fifty, where the catalogue of 6,000 bytes names about thirty: an AI finds the others only if it calls `find_skills`.

| Client | Without naming the skill |
| --- | --- |
| Claude Code | Reads the skill at its first call, and follows it |
| Mistral Vibe, in the terminal | Reads the skill at its first call, and follows it |
| ChatGPT | Reads it when you pick the app of Cahut in your message; otherwise calls no app |
| claude.ai | Writes the reminder without calling Cahut: it reads the description of a tool only after it searches its tools |
| Codex | Writes the reminder without calling Cahut |

With claude.ai, Codex, and ChatGPT when you do not pick Cahut, name the skill, or paste the line. Clients change: these trials say what they did, not what they will do.

## The rule of priority

`get_skill` gives it with each skill, and the skill that comes with Cahut says it too:

1. what the person asks in the conversation;
2. their personal skill;
3. the skill of the Cahut;
4. a skill for the same use installed in the client: it is a copy, and the skill of the Cahut has the last word.

## Skills for every Window

The administrators of a Cahut mark up to three of its skills **For every Window**, on the page of each skill: its brand, the notices of its forms, its conventions. AIs must then read them before they publish or update a Window.

- `publish_window` and `update_window` take `follows`: the skills that the Window follows, each as its name and the version that `get_skill` gives, such as `charte-acme@3f9a2c1b`.
- A call of an AI whose `follows` lacks the current version of a marked skill is refused, and publishes nothing. The refusal carries the skill whole, the list of its files and what to give in `follows`. The AI makes the Window again as the skill says, and its next call passes.
- The version comes only with the text of the skill, from `get_skill` or from the refusal: an AI that gives it received the skill. Cahut keeps nothing between two calls, so the gate holds for each conversation.
- A new version of a marked skill changes its version: the next publication is refused once, and receives the new text. A Window already published does not change: the AI publishes it again to follow the new version.
- The gate stops AIs only. A person who publishes is never stopped, and an individual Cahut has no gate.
- A marked skill holds a `SKILL.md` of 16 KiB at most, since the refusal carries it whole.

The gate makes sure that the AI had the skill in front of it, not that the Window follows it. In a trial of 40 decks with Claude Code and Mistral Vibe, against the brand of a cooperative marked for every Window, no Window was published without `follows`, and each deck carried the five marks of the brand that the trial counted: its green, its background, the font of its titles, its footer and its last slide. Claude Code read the brand first each time. Mistral Vibe called `publish_window` without it in 19 tries of 25, was stopped, and made the deck again as the brand says. One of its decks left out the logo.

## The files of a skill in a Window

A file given to `publish_window` or `update_window` may name `from_skill` in place of its `content`: the name of a skill, with its part as a prefix if need be, then the path of one of its files.

```json
"files": [
  { "path": "index.html", "content": "<!doctype html>..." },
  { "path": "img/logo.svg", "from_skill": "charte-acme/assets/logo.svg" },
  { "path": "fonts/inter.woff2", "from_skill": "charte-acme/assets/fonts/inter.woff2" }
]
```

- Cahut copies the bytes of the current version of the skill into the version of the Window, if the person the AI acts for may read the skill. The AI sends neither the logo nor the font.
- The limits of a Window count the files it copies.
- A Window carries its copy, and Cahut serves nothing under `/-/` for Windows to link: when the skill changes, the Window keeps the files it had until its AI publishes it again.

## Limits

| What | At most |
| --- | --- |
| `SKILL.md` | 64 KiB, or 16 KiB for a skill for every Window |
| A file of a skill | 2 MiB |
| Files of a skill | 100, and 4 MiB in all |
| Skills of the Cahut | 50 |
| Skills of each person | 50 |
| Skills for every Window | 3 |
| Description | 1,024 characters |
| Versions of a skill | 2: the current one and the one before |

The disk of the Cahut counts both versions of each skill, as it counts the Windows.

## The Logbook

The [Logbook](/concepts/knock-first-and-logbook/#the-logbook) keeps:

- each write, import, rollback, mark and deletion, with its part, its name, its version, and the number and weight of its files, never their content;
- that a search took place, and how many skills it found, never its words: administrators read the whole Logbook, and the words of a search would tell them the tasks of others;
- each skill and each file that an AI reads, so that administrators see which skills serve;
- the skills of the Cahut that a publication follows, the files it copied, and the skills that a refusal of the gate found missing.

It never names a personal skill, nor one of its files: administrators read the whole Logbook, and do not read the skills of others.
