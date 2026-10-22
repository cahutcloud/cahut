// Small pieces that the screens share.
import { AppWindow, ChartColumn, Check, Copy, FileText, Globe, KeyRound, Lock, Presentation, Users } from "lucide-preact";
import type { ComponentChildren } from "preact";
import { useState } from "preact/hooks";
import type { Window } from "./api.gen";
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

/** A Window, as a card that leads to its page. */
export function WindowCard({ window: w, showRoom }: { window: Window; showRoom?: boolean }) {
	const kind = kinds[w.kind] ?? kinds.site;
	const Icon = kind.icon;
	return (
		<article class="c-card c-window">
			<Link class="c-window__thumb window-thumb" href={windowPage(w)} aria-label={w.title}>
				<Icon size={40} aria-hidden="true" />
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

/** Says a size in bytes, as in "12 KB" or "3.4 MB". */
export function bytes(n: number): string {
	if (n < 1000) return `${n} bytes`;
	if (n < 1000 * 1000) return `${Math.round(n / 1000)} KB`;
	return `${(n / 1000 / 1000).toFixed(1)} MB`;
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
