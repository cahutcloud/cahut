# 65. The list of tools stays under 48 KB

- Status: Accepted
- Date: 2026-11-16

## Context

[ADR 57](0057-the-list-of-tools-under-40-kb.md) keeps the list of tools that MCP gives an administrator who may write under 40 KB. Before the bound rises again, it asks that the list be tried with the AI clients at its new weight.

[ADR 64](0064-the-skills-of-a-cahut.md) puts the catalogue of the skills of a Cahut in the description of `find_skills`. On 16 November 2026 the list weighs 38,817 bytes without skills, for 43 tools. Under 40 KB, with 700 bytes kept for `follows` and `from_skill`, the catalogue has 1,200 bytes: six skills of fifty. An AI finds the others only if it calls `find_skills`.

On 16 November 2026, a build that is not a release served a list of 49,008 bytes from a Cahut on a test server. Its catalogue weighed 8,890 bytes and named 48 of its 50 skills. Five clients were tried, as an administrator who may write: Claude Code 2.1.322, Mistral Vibe 2.1.0, Codex 0.159, claude.ai and ChatGPT on its free plan. Each of them:

- received the 43 tools and the whole description of `find_skills`, 8,826 characters, and copied its last line word for word;
- called the first tool of the list and the last.

None cut the list or the description.

Each client was then asked five times to write the reminder of an unpaid invoice, which a skill of the Cahut describes:

- Claude Code and Mistral Vibe read the skill each time, at their first call, and followed it.
- ChatGPT read it each time when the person picked Cahut in the message. When no one picked it, ChatGPT called no app.
- claude.ai and Codex wrote a reminder of their own, without calling Cahut. Codex did the same with a list under 40 KB. claude.ai gets the tools of an MCP server by their names, and reads a description only after it searches its tools.

## Decision

- **The list of tools stays under 48 KB**, 49,152 bytes, for an administrator who may write, and the test holds it there. It replaces the bound of 40 KB in ADR 57, and nothing else of it.
- The catalogue of the skills of a Cahut takes 6,000 bytes of it, and 700 bytes stay for `follows` and `from_skill`. A test keeps the list under 48 KB less 700 bytes with fifty skills, each with the longest description.
- Before the bound rises again, the list is tried at its new weight with these five clients.

## Consequences

- The catalogue names about 32 skills where it named six, with descriptions like those of the trial.
- With fifty skills, each with the longest description, the list weighs 44,630 bytes.
- 48 KB stays well under the list of 87 KB that Codex cut short ([ADR 29](0029-chatgpt-second-ai-client.md)).
- The weight of the list does not decide whether an AI reads a skill for a task that does not touch Cahut: its client does. The line that the Skills screen gives to paste into the instructions of an AI stays the sure way for claude.ai, for Codex, and for ChatGPT when the person does not pick Cahut. The documentation tells the people who use these clients to paste it.
- A new tool still weighs on every client that loads the list whole, as ADR 57 says.

## Alternatives considered

- **Keeping 40 KB.** Six skills in the catalogue, and the others only through `find_skills`, which clients call less often than they read a description.
- **A catalogue of 9,000 bytes, as in the trial.** It names about 48 skills, but leaves no room under 48 KB for `follows` and `from_skill`, nor for the next tool.
- **A higher bound, such as 64 KB.** More room, without a client tried at that weight.
