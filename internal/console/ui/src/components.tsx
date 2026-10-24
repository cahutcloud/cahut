// Small pieces that the screens share.
import { AppWindow, ChartColumn, Check, Copy, FileText, Globe, KeyRound, Lock, Presentation, Users, X } from "lucide-preact";
import type { ComponentChildren } from "preact";
import { useEffect, useId, useRef, useState } from "preact/hooks";
import type { Window } from "./api.gen";
import { useConsole } from "./context";
import { useLoad } from "./load";
import { Link } from "./router";

/** The head of a screen: a label, a title and a line under it. */
export function PageHead(props: { label: string; title: string; children?: ComponentChildren; actions?: ComponentChildren }) {
	return (
		<div class="page-head">
			<div class="page-head__text">
				<span class="c-label">{props.label}</span>
				<h1 class="c-h1">{props.title}</h1>
				{props.children && <p class="c-muted">{props.children}</p>}
			</div>
			{props.actions && <div class="page-head__actions">{props.actions}</div>}
		</div>
	);
}

export function Failure({ message }: { message: string }) {
	return (
		<div class="c-alert c-alert--danger" role="alert">
			{message}
		</div>
	);
}

export function Notice({ children }: { children: ComponentChildren }) {
	return (
		<div class="c-alert c-alert--info" role="status">
			<div>{children}</div>
		</div>
	);
}

/** A value to copy, such as a link or a command, with its button. */
/** How long a page stays in front before its buttons that grant something arm. */
const armDelay = 800;

/**
 * Arms the buttons that grant something, such as Approve or Connect, once
 * the page has been in front, focused, for a moment, and the person has
 * moved the pointer or pressed a key since: a page that opens this one
 * under a double-click, or switches what lies under the cursor, cannot land
 * the second click on them. The page disarms when it loses the focus.
 */
export function useArmed(): boolean {
	const [armed, setArmed] = useState(false);
	useEffect(() => {
		let since = 0;
		const watch = () => {
			if (document.visibilityState === "visible" && document.hasFocus()) {
				since ||= Date.now();
				return;
			}
			since = 0;
			setArmed(false);
		};
		const act = () => {
			if (since && Date.now() - since >= armDelay) setArmed(true);
		};
		watch();
		const timer = setInterval(watch, 100);
		window.addEventListener("blur", watch);
		window.addEventListener("pointermove", act);
		window.addEventListener("keydown", act);
		return () => {
			clearInterval(timer);
			window.removeEventListener("blur", watch);
			window.removeEventListener("pointermove", act);
			window.removeEventListener("keydown", act);
		};
	}, []);
	return armed;
}

export function CopyField({ value, label }: { value: string; label: string }) {
	const [copied, setCopied] = useState(false);
	const copy = async () => {
		await navigator.clipboard.writeText(value);
		setCopied(true);
		setTimeout(() => setCopied(false), 2000);
	};
	return (
		<div class="copy">
			<code class="copy__value c-mono">{value}</code>
			<button class="c-btn c-btn--secondary c-btn--sm" type="button" onClick={copy} aria-label={`Copy ${label}`}>
				{copied ? <Check class="i" size={14} aria-hidden="true" /> : <Copy class="i" size={14} aria-hidden="true" />}
				{copied ? "Copied" : "Copy"}
			</button>
		</div>
	);
}

/**
 * A dialog over the page, open as long as it renders: it keeps the focus
 * until its close button or the Escape key closes it, and then calls onClose.
 */
export function Dialog(props: { title: string; meta?: ComponentChildren; onClose: () => void; children: ComponentChildren }) {
	const ref = useRef<HTMLDialogElement>(null);
	const id = useId();
	useEffect(() => {
		const dialog = ref.current;
		dialog?.showModal();
		return () => dialog?.close();
	}, []);
	return (
		<dialog ref={ref} class="dialog" aria-labelledby={id} onClose={props.onClose}>
			<div class="dialog__head">
				<div class="dialog__title">
					<h2 class="c-h4" id={id}>
						{props.title}
					</h2>
					{props.meta && <div class="dialog__meta">{props.meta}</div>}
				</div>
				<button class="c-btn c-btn--ghost c-btn--sm" type="button" aria-label="Close" onClick={() => ref.current?.close()}>
					<X class="i" size={16} aria-hidden="true" />
				</button>
			</div>
			<div class="dialog__body">{props.children}</div>
		</dialog>
	);
}

/**
 * A button for a change that cannot be undone, or that changes what a person
 * may do: it asks first, on the spot.
 */
export function ConfirmButton(props: {
	label: string;
	question: string;
	confirm: string;
	danger?: boolean;
	onConfirm: () => Promise<void> | void;
}) {
	const [asking, setAsking] = useState(false);
	const [busy, setBusy] = useState(false);
	if (!asking) {
		return (
			<button class="c-btn c-btn--secondary c-btn--sm" type="button" onClick={() => setAsking(true)}>
				{props.label}
			</button>
		);
	}
	const go = async () => {
		setBusy(true);
		try {
			await props.onConfirm();
		} finally {
			setBusy(false);
			setAsking(false);
		}
	};
	return (
		<span class="confirm" role="group" aria-label={props.question}>
			<span class="c-small">{props.question}</span>
			<button class={`c-btn c-btn--sm ${props.danger ? "c-btn--danger" : "c-btn--primary"}`} type="button" disabled={busy} onClick={go}>
				{props.confirm}
			</button>
			<button class="c-btn c-btn--ghost c-btn--sm" type="button" onClick={() => setAsking(false)}>
				Cancel
			</button>
		</span>
	);
}

const kinds: Record<string, { icon: typeof FileText; name: string }> = {
	document: { icon: FileText, name: "Document" },
	deck: { icon: Presentation, name: "Deck" },
	dashboard: { icon: ChartColumn, name: "Dashboard" },
	site: { icon: AppWindow, name: "Site" },
};

/** The access levels of a Window, with their badge. */
export const accesses: Record<string, { icon: typeof Lock; name: string; who: string }> = {
	room: { icon: Lock, name: "Room only", who: "the members of its Room" },
	internal: { icon: Users, name: "Internal", who: "everyone signed in to this Cahut" },
	key: { icon: KeyRound, name: "Key", who: "anyone with its Key" },
	public: { icon: Globe, name: "Public", who: "anyone with its address" },
};

/** The access of a Window, as a badge: an icon and a word, never a color alone. */
export function AccessBadge({ access }: { access: string }) {
	const a = accesses[access] ?? accesses.room;
	const Icon = a.icon;
	return (
		<span class={`c-badge c-badge--${access in accesses ? access : "room"}`}>
			<Icon class="i" size={14} aria-hidden="true" />
			{a.name}
		</span>
	);
}

/** The page of a Window in the console. */
export function windowPage(w: Pick<Window, "room" | "name">): string {
	return `/rooms/${w.room}/${w.name}`;
}

export function kindName(kind: string): string {
	return (kinds[kind] ?? kinds.site).name;
}

/** The most Windows that one call of preview_windows previews. */
const maxPreviews = 100;

/**
 * Windows, as cards, each with a preview of the Window itself when the API
 * gives one. See docs/adr/0027-previews-of-windows-in-the-console.md.
 */
export function WindowCards({ windows, showRoom }: { windows: Window[]; showRoom?: boolean }) {
	const api = useConsole();
	const names = windows.slice(0, maxPreviews).map((w) => `${w.room}/${w.name}`);
	// Without previews, the cards keep the icon of their kind.
	const { data } = useLoad(() => api.cahut("preview_windows", { windows: names }).catch(() => null), [names.join(" ")]);
	const frames = new Map(data?.previews.map((p) => [p.window, p.frame]));
	return (
		<div class="cards cards--windows">
			{windows.map((w) => (
				<WindowCard key={w.url} window={w} showRoom={showRoom} preview={frames.get(`${w.room}/${w.name}`)} />
			))}
		</div>
	);
}

/** A Window, as a card that leads to its page. */
export function WindowCard({ window: w, showRoom, preview }: { window: Window; showRoom?: boolean; preview?: string }) {
	const kind = kinds[w.kind] ?? kinds.site;
	const Icon = kind.icon;
	return (
		<article class="c-card c-window">
			<Link class="c-window__thumb window-thumb" href={windowPage(w)} aria-label={w.title}>
				<Icon size={40} aria-hidden="true" />
				{preview && <WindowPreview src={preview} title={w.title} />}
			</Link>
			<div class="c-window__body">
				<Link class="c-window__title window-title" href={windowPage(w)}>
					<Icon class="i" size={16} aria-hidden="true" />
					{w.title}
				</Link>
				<div class="c-window__meta">
					<span>
						{kind.name} · {ago(w.updated_at)} · {showRoom ? w.room : w.created_by}
					</span>
					<AccessBadge access={w.access} />
				</div>
			</div>
		</article>
	);
}

/** How wide a preview draws its Window, before it shrinks to its card. */
const previewWidth = 1280;

/**
 * The Window itself, drawn at the width of a laptop and shrunk to its card,
 * over the icon of its kind once it loads. Its frame keeps the sandbox of
 * the Window's files, and drops forms, pop-ups, dialogs and downloads; it
 * takes neither clicks nor focus.
 */
function WindowPreview({ src, title }: { src: string; title: string }) {
	const frame = useRef<HTMLIFrameElement>(null);
	const [scale, setScale] = useState(0);
	const [loaded, setLoaded] = useState(false);
	useEffect(() => {
		const thumb = frame.current?.parentElement;
		if (!thumb) return;
		const observer = new ResizeObserver(([entry]) => setScale(entry.contentRect.width / previewWidth));
		observer.observe(thumb);
		return () => observer.disconnect();
	}, []);
	return (
		<iframe
			ref={frame}
			class={`window-preview${loaded ? " window-preview--shown" : ""}`}
			style={{ "--preview-scale": scale }}
			src={src}
			title={`Preview of ${title}`}
			sandbox="allow-scripts"
			loading="lazy"
			referrerpolicy="no-referrer"
			tabIndex={-1}
			aria-hidden="true"
			onLoad={() => setLoaded(true)}
		/>
	);
}

/** Says how long ago a time was, as in "2 min ago", "Yesterday" or "Oct 12". */
export function ago(at: string): string {
	const then = new Date(at);
	const minutes = Math.round((Date.now() - then.getTime()) / 60000);
	if (minutes < 1) return "Just now";
	if (minutes < 60) return `${minutes} min ago`;
	if (minutes < 24 * 60) return `${Math.round(minutes / 60)} h ago`;
	if (minutes < 48 * 60) return "Yesterday";
	return then.toLocaleDateString("en-US", { month: "short", day: "numeric" });
}

/** Says a size in bytes, as in "12 KB", "3.4 MB" or "10 GB", in units of 1000 bytes. */
export function bytes(n: number): string {
	if (n < 1000) return `${n} bytes`;
	if (n < 1000 * 1000) return `${Math.round(n / 1000)} KB`;
	if (n < 1000 * 1000 * 1000) return `${(n / 1000 / 1000).toFixed(1)} MB`;
	return `${(n / 1000 / 1000 / 1000).toFixed(1).replace(/\.0$/, "")} GB`;
}

/** Says a count with its noun, as in "1 Window" or "3 Windows". */
export function plural(n: number, one: string, many: string): string {
	return `${n} ${n === 1 ? one : many}`;
}

/** The initials of a name, for its avatar. */
export function initials(name: string): string {
	const words = name.trim().split(/\s+/);
	return (words[0].charAt(0) + (words.length > 1 ? words[words.length - 1].charAt(0) : "")).toUpperCase();
}
