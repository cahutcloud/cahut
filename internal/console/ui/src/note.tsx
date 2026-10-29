// The note of an invitation: the small editor that writes it, and the page
// that shows it. A note is plain text with a few marks of Markdown, **bold**,
// *italics* and lists, which the server reads into paragraphs and lists for
// the e-mail and for the page of the invitation. The editor shows the styles
// as they will read, and writes the marks. See
// docs/adr/0046-styled-notes-in-invitations.md.
import { Bold, Italic, List, ListOrdered } from "lucide-preact";
import { type ComponentChildren, Fragment } from "preact";
import { useEffect, useId, useRef, useState } from "preact/hooks";
import type { NoteBlock, NoteRun } from "./api.gen";
import { words } from "./i18n";

const t = words({
	en: {
		formatting: "Formatting",
		bold: "Bold",
		italic: "Italic",
		bullets: "Bulleted list",
		numbers: "Numbered list",
		placeholder: "A few words for them",
		count: (n: number, max: number) => `${n} of ${max} characters`,
		tooLong: (n: number, max: number) => `${n} characters, and a note holds ${max} at most: shorten it to send it.`,
	},
	fr: {
		formatting: "Mise en forme",
		bold: "Gras",
		italic: "Italique",
		bullets: "Liste à puces",
		numbers: "Liste numérotée",
		placeholder: "Quelques mots pour les personnes invitées",
		count: (n, max) => `${n} caractères sur ${max}`,
		tooLong: (n, max) => `${n} caractères, et une note en tient ${max} au plus : raccourcissez-la pour l'envoyer.`,
	},
});

/** The length of a note, as the server counts it: in characters. */
export function noteLength(note: string): number {
	return [...note].length;
}

type Tool = "bold" | "italic" | "bullets" | "numbers";

const tools: { tool: Tool; command: string; icon: typeof Bold; key: string }[] = [
	{ tool: "bold", command: "bold", icon: Bold, key: "B" },
	{ tool: "italic", command: "italic", icon: Italic, key: "I" },
	{ tool: "bullets", command: "insertUnorderedList", icon: List, key: "Shift+8" },
	{ tool: "numbers", command: "insertOrderedList", icon: ListOrdered, key: "Shift+7" },
];

const mac = /Mac|iPhone|iPad/.test(navigator.platform);

/**
 * The editor of a note: a toolbar for bold, italics and lists, and the
 * text, which takes what is pasted as plain text. It hands the note on each
 * change, says when it is longer than max, and empties with a new key.
 */
export function NoteEditor({
	label,
	help,
	max,
	onChange,
}: {
	label: string;
	help: ComponentChildren;
	max: number;
	onChange: (note: string) => void;
}) {
	const id = useId();
	const box = useRef<HTMLDivElement>(null);
	const bar = useRef<HTMLDivElement>(null);
	const saved = useRef<Range | null>(null);
	const [on, setOn] = useState<Record<Tool, boolean>>({ bold: false, italic: false, bullets: false, numbers: false });
	const [length, setLength] = useState(0);
	const [empty, setEmpty] = useState(true);
	const [current, setCurrent] = useState(0);

	// The buttons show the styles where the caret is, and the editor keeps
	// the selection, which a click on the toolbar would lose.
	useEffect(() => {
		const update = () => {
			const s = document.getSelection();
			if (!s || s.rangeCount === 0 || !box.current?.contains(s.anchorNode)) return;
			saved.current = s.getRangeAt(0).cloneRange();
			setOn({
				bold: document.queryCommandState("bold"),
				italic: document.queryCommandState("italic"),
				bullets: document.queryCommandState("insertUnorderedList"),
				numbers: document.queryCommandState("insertOrderedList"),
			});
		};
		document.addEventListener("selectionchange", update);
		return () => document.removeEventListener("selectionchange", update);
	}, []);

	const changed = () => {
		if (!box.current) return;
		const note = toNote(box.current);
		setLength(noteLength(note));
		setEmpty(!box.current.textContent && !box.current.querySelector("li"));
		onChange(note);
	};

	const apply = (command: string) => {
		const el = box.current;
		if (!el) return;
		if (document.activeElement !== el) {
			el.focus();
			let r = saved.current;
			if (!r) {
				r = document.createRange();
				r.selectNodeContents(el);
				r.collapse(false);
			}
			const s = document.getSelection();
			s?.removeAllRanges();
			s?.addRange(r);
		}
		document.execCommand(command);
		changed();
	};

	const shortcut = (e: KeyboardEvent) => {
		if (!(mac ? e.metaKey : e.ctrlKey) || e.altKey) return;
		const key = e.key.toLowerCase();
		const tool = e.shiftKey
			? tools.find((x) => x.key === `Shift+${e.code.replace("Digit", "")}`)
			: tools.find((x) => x.key.toLowerCase() === key);
		if (tool) {
			e.preventDefault();
			apply(tool.command);
		} else if (key === "u" && !e.shiftKey) {
			e.preventDefault(); // a note has no underline
		}
	};

	// Arrows move between the buttons of the toolbar, which is one stop of
	// the Tab key.
	const move = (e: KeyboardEvent) => {
		const last = tools.length - 1;
		const next = { ArrowRight: current + 1, ArrowLeft: current - 1, Home: 0, End: last }[e.key];
		if (next === undefined) return;
		e.preventDefault();
		const i = (next + tools.length) % tools.length;
		setCurrent(i);
		bar.current?.querySelectorAll("button")[i]?.focus();
	};

	return (
		<div class="c-field">
			<span class="c-field__label" id={`${id}-label`} onClick={() => box.current?.focus()}>
				{label}
			</span>
			<div class="editor">
				<div class="editor__bar" role="toolbar" aria-label={t.formatting} aria-controls={id} ref={bar} onKeyDown={move}>
					{tools.map(({ tool, command, icon: Icon, key }, i) => (
						<Fragment key={tool}>
							{i === 2 && <span class="editor__rule" aria-hidden="true" />}
							<button
								type="button"
								class="editor__tool"
								aria-label={t[tool]}
								aria-pressed={on[tool]}
								aria-keyshortcuts={`${mac ? "Meta" : "Control"}+${key}`}
								title={`${t[tool]} (${mac ? "⌘" : "Ctrl+"}${key.replace("Shift+", mac ? "⇧" : "Shift+")})`}
								tabIndex={i === current ? 0 : -1}
								onFocus={() => setCurrent(i)}
								onMouseDown={(e) => e.preventDefault()}
								onClick={() => apply(command)}
							>
								<Icon size={16} aria-hidden="true" />
							</button>
						</Fragment>
					))}
					<span class={`editor__count${length > max ? " editor__count--over" : ""}`} title={t.count(length, max)}>
						{length}/{max}
					</span>
				</div>
				<div
					id={id}
					ref={box}
					class="editor__text"
					contentEditable
					role="textbox"
					aria-multiline="true"
					aria-labelledby={`${id}-label`}
					aria-describedby={`${id}-help`}
					aria-invalid={length > max}
					data-placeholder={t.placeholder}
					data-empty={empty}
					onInput={changed}
					onKeyDown={shortcut}
					onBeforeInput={(e) => {
						// Only the styles of a note, from the menus of the browser too.
						if (e.inputType.startsWith("format") && e.inputType !== "formatBold" && e.inputType !== "formatItalic") e.preventDefault();
					}}
					onPaste={(e) => {
						e.preventDefault();
						document.execCommand("insertText", false, e.clipboardData?.getData("text/plain") ?? "");
					}}
					onDrop={(e) => e.preventDefault()}
				/>
			</div>
			<span class={`c-field__help${length > max ? " editor__over" : ""}`} id={`${id}-help`} aria-live="polite">
				{length > max ? t.tooLong(length, max) : help}
			</span>
		</div>
	);
}

/** A note, in its paragraphs and lists. */
export function Note({ blocks }: { blocks: NoteBlock[] }) {
	return (
		<div class="note">
			{blocks.map((b, i) =>
				b.list === "bullets" ? (
					<ul key={i}>
						{b.lines.map((l, j) => (
							<li key={j}>
								<Runs runs={l} />
							</li>
						))}
					</ul>
				) : b.list === "numbers" ? (
					<ol key={i} start={b.start ?? 0}>
						{b.lines.map((l, j) => (
							<li key={j}>
								<Runs runs={l} />
							</li>
						))}
					</ol>
				) : (
					<p key={i}>
						{b.lines.map((l, j) => (
							<Fragment key={j}>
								{j > 0 && <br />}
								<Runs runs={l} />
							</Fragment>
						))}
					</p>
				),
			)}
		</div>
	);
}

function Runs({ runs }: { runs: NoteRun[] }) {
	return (
		<>
			{runs.map((r, i) => {
				let words: ComponentChildren = r.text;
				if (r.italic) words = <em>{words}</em>;
				if (r.bold) words = <strong>{words}</strong>;
				return <Fragment key={i}>{words}</Fragment>;
			})}
		</>
	);
}

// What the editor holds, in the format of notes.

type Piece = { text: string; bold: boolean; italic: boolean };
type Line = { list: Element | null; pieces: Piece[] };

/**
 * toNote writes the text of the editor as a note: its lines, the items of
 * its lists, and the bold and italics that the browser shows, whatever
 * markup it made for them. Spaces stay single, a blank line ends a
 * paragraph, and a star, a backslash or the start of a line that would read
 * as a mark gets a backslash.
 */
export function toNote(root: HTMLElement): string {
	const lines: Line[] = [];
	let line: Line | null = null;
	let list: Element | null = null;
	const end = () => {
		if (line) lines.push(line);
		line = null;
	};
	const open = () => (line ??= { list, pieces: [] });
	const walk = (node: Node) => {
		for (const n of Array.from(node.childNodes)) {
			if (n instanceof Text) {
				const style = getComputedStyle(n.parentElement ?? root);
				open().pieces.push({ text: n.data, bold: Number(style.fontWeight) >= 600, italic: style.fontStyle !== "normal" });
			} else if (n instanceof HTMLBRElement) {
				open();
				end();
			} else if (n instanceof HTMLUListElement || n instanceof HTMLOListElement) {
				end();
				const outer = list;
				list = n;
				walk(n);
				end();
				list = outer;
			} else if (n instanceof HTMLElement) {
				const block = !getComputedStyle(n).display.startsWith("inline");
				if (block) end();
				walk(n);
				if (block) end();
			}
		}
	};
	walk(root);
	end();

	const out: string[] = [];
	let prev: Line | null = null;
	let number = 0;
	for (const l of lines) {
		const text = marks(l.pieces);
		if (l.list) {
			if (!text) continue; // an empty item
			if (prev?.list && prev.list !== l.list) out.push(""); // two lists, one after the other
			number = prev?.list === l.list ? number + 1 : 1;
			out.push(l.list instanceof HTMLOListElement ? `${number}. ${text}` : `- ${text}`);
		} else {
			out.push(text.replace(/^([-+])(?= )/, "\\$1").replace(/^(\d{1,9})([.)])(?= )/, "$1\\$2"));
		}
		prev = l;
	}
	return out.join("\n").replace(/\n{3,}/g, "\n\n").trim();
}

/**
 * marks writes the pieces of a line with the stars of their styles: they
 * open before a word and close after one, never next to a space, so that
 * the server reads them as the editor showed them.
 */
function marks(pieces: Piece[]): string {
	const stars = (bold: boolean, italic: boolean) => (bold && italic ? "***" : bold ? "**" : italic ? "*" : "");
	let out = "";
	let space = "";
	let bold = false;
	let italic = false;
	for (const p of pieces) {
		for (const part of p.text.replace(/[\s\u00a0]+/g, " ").match(/ +|[^ ]+/g) ?? []) {
			if (part.startsWith(" ")) {
				if (out) space = " ";
				continue;
			}
			out += stars(bold && !p.bold, italic && !p.italic) + space + stars(p.bold && !bold, p.italic && !italic);
			out += part.replace(/[\\*]/g, "\\$&");
			space = "";
			bold = p.bold;
			italic = p.italic;
		}
	}
	return out + stars(bold, italic);
}
