// Small pieces that the screens share.
import { AppWindow, ArrowUpRight, ChartColumn, Check, Copy, FileText, Globe, House, KeyRound, Lock, Presentation, Users, X } from "lucide-preact";
import { type ComponentChildren, Fragment } from "preact";
import { useEffect, useId, useRef, useState } from "preact/hooks";
import { cahut, explain } from "./api";
import type { Activation, LegalLinks, LegalPages, SignInPage, Window } from "./api.gen";
import { useConsole } from "./context";
import { de, lang, langs, locale, setLang, words } from "./i18n";
import { useLoad } from "./load";
import { href, Link } from "./router";

export { plural } from "./i18n";

const t = words({
	en: {
		copy: "Copy",
		copied: "Copied",
		copyLabel: (label: string) => `Copy ${label}`,
		close: "Close",
		cancel: "Cancel",
		active: "Active",
		pending: "No password yet",
		disabled: "Disabled",
		emailed: (name: string, email: string | undefined, until: string) =>
			`The link where ${name} chooses their password went to ${email ?? "their address"}. It works once, until ${until}. Here it is too, shown only now.`,
		send: (name: string, until: string) =>
			`Send ${name} this link, where they choose their password. It works once, until ${until}, and is shown only now.`,
		activationLink: "the activation link",
		document: "Document",
		deck: "Deck",
		dashboard: "Dashboard",
		site: "Site",
		roomOnly: "Room only",
		roomWho: "the members of its Room",
		internal: "Internal",
		internalWho: "everyone signed in to this Cahut",
		key: "Key",
		keyWho: "anyone with its Key",
		public: "Public",
		publicWho: "anyone with its address",
		roomPage: "Page of the Room",
		preview: (title: string) => `Preview of ${title}`,
		openWindow: (title: string) => `Open ${title} in a new tab`,
		justNow: "Just now",
		minutesAgo: (n: number) => `${n} min ago`,
		hoursAgo: (n: number) => `${n} h ago`,
		yesterday: "Yesterday",
		bytes: "bytes",
		kb: "KB",
		mb: "MB",
		gb: "GB",
		language: "Language",
		legalPages: "Legal pages",
		legalNotice: "Legal notice",
		terms: "Terms",
		privacy: "Privacy",
		support: "Support",
	},
	fr: {
		copy: "Copier",
		copied: "Copié",
		copyLabel: (label) => `Copier ${label}`,
		close: "Fermer",
		cancel: "Annuler",
		active: "Actif",
		pending: "Pas encore de mot de passe",
		disabled: "Désactivé",
		emailed: (name, email, until) =>
			`Le lien où ${name} choisira son mot de passe est parti à ${email ?? "son adresse"}. Il ne sert qu'une fois, jusqu'au ${until}. Le voici aussi, affiché cette fois seulement.`,
		send: (name, until) =>
			`Envoyez ce lien à ${name}, qui y choisira son mot de passe. Il ne sert qu'une fois, jusqu'au ${until}, et ne s'affiche que maintenant.`,
		activationLink: "le lien d'activation",
		document: "Document",
		deck: "Présentation",
		dashboard: "Tableau de bord",
		site: "Site",
		roomOnly: "Room seulement",
		roomWho: "les membres de sa Room",
		internal: "Interne",
		internalWho: "toute personne connectée à cette Cahut",
		key: "Key",
		keyWho: "quiconque a sa Key",
		public: "Public",
		publicWho: "quiconque a son adresse",
		roomPage: "Page de la Room",
		preview: (title) => `Aperçu ${de(title)}`,
		openWindow: (title) => `Ouvrir ${title} dans un nouvel onglet`,
		justNow: "À l'instant",
		minutesAgo: (n) => `il y a ${n} min`,
		hoursAgo: (n) => `il y a ${n} h`,
		yesterday: "Hier",
		bytes: "octets",
		kb: "ko",
		mb: "Mo",
		gb: "Go",
		language: "Langue",
		legalPages: "Pages légales",
		legalNotice: "Mentions légales",
		terms: "Conditions d'utilisation",
		privacy: "Confidentialité",
		support: "Assistance",
	},
});

/** The head of a screen: a label, a title and a line under it. */
export function PageHead(props: { label: string; title: ComponentChildren; children?: ComponentChildren; actions?: ComponentChildren }) {
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

/** A domain that goes on a second line before one of its dots, my.example / .com, when it is longer than its line. */
export function Domain(props: { name: string }) {
	return (
		<>
			{props.name.split(".").map((part, i) => (
				<Fragment key={i}>
					{i > 0 && <wbr />}
					{i > 0 && "."}
					{part}
				</Fragment>
			))}
		</>
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

/**
 * The languages of the console, under the card of a page that opens before
 * sign-in and at the foot of the navigation once signed in: each in its own
 * words, the current one without a button.
 */
export function Languages() {
	return (
		<nav class="languages c-small" aria-label={t.language}>
			{langs.map((l) =>
				l.lang === lang ? (
					<span key={l.lang} lang={l.lang} aria-current="true">
						{l.name}
					</span>
				) : (
					<button key={l.lang} class="c-link link-button" type="button" lang={l.lang} onClick={() => setLang(l.lang)}>
						{l.name}
					</button>
				),
			)}
		</nav>
	);
}

/** The legal pages that a Cahut may link, in the order that the foot of a page links them, each with its name. */
export const legalPages: { page: keyof LegalPages; name: string }[] = [
	{ page: "legal_notice", name: t.legalNotice },
	{ page: "terms", name: t.terms },
	{ page: "privacy", name: t.privacy },
	{ page: "support", name: t.support },
];

/**
 * The links to the legal pages of the Cahut, in a discreet line at the foot
 * of a page, in the language of the console: the address in French of a
 * page, when the console speaks French and the page has one, or else its
 * other address. Each opens in a new tab, so that a form being filled stays
 * as it is. Nothing at all when the Cahut links none. See
 * docs/adr/0042-the-legal-pages-of-a-cahut.md.
 */
export function Legal({ links }: { links?: LegalLinks }) {
	const shown = legalPages.flatMap(({ page, name }) => {
		const url = (lang === "fr" ? links?.fr?.[page] : undefined) || links?.[page];
		return url && /^https:\/\//i.test(url) ? [{ page, name, url }] : [];
	});
	if (shown.length === 0) return null;
	return (
		<footer class="legal">
			<nav class="legal__links c-small" aria-label={t.legalPages}>
				{shown.map((l) => (
					<a key={l.page} href={l.url} target="_blank" rel="noopener">
						{l.name}
					</a>
				))}
			</nav>
		</footer>
	);
}

let signInPage: Promise<SignInPage> | undefined;

/**
 * What the pages before sign-in offer, from describe_sign_in: asked once
 * for each load of the console, since every call of a visitor counts
 * against the sign-ins of its address.
 */
export function describeSignIn(): Promise<SignInPage> {
	signInPage ??= cahut("describe_sign_in", {}).catch((err: unknown) => {
		signInPage = undefined;
		throw err;
	});
	return signInPage;
}

/** The links to the legal pages of the Cahut, at the foot of a page that opens before sign-in. */
export function LegalBeforeSignIn() {
	const [links, setLinks] = useState<LegalLinks>();
	useEffect(() => {
		describeSignIn().then(
			(p) => setLinks(p.legal),
			() => undefined,
		);
	}, []);
	return <Legal links={links} />;
}

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
			<button class="c-btn c-btn--secondary c-btn--sm" type="button" onClick={copy} aria-label={t.copyLabel(label)}>
				{copied ? <Check class="i" size={14} aria-hidden="true" /> : <Copy class="i" size={14} aria-hidden="true" />}
				{copied ? t.copied : t.copy}
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
				<button class="c-btn c-btn--ghost c-btn--sm" type="button" aria-label={t.close} onClick={() => ref.current?.close()}>
					<X class="i" size={16} aria-hidden="true" />
				</button>
			</div>
			<div class="dialog__body">{props.children}</div>
		</dialog>
	);
}

const statuses: Record<string, { word: string; badge: string }> = {
	active: { word: t.active, badge: "c-badge--success" },
	pending: { word: t.pending, badge: "c-badge--warning" },
	disabled: { word: t.disabled, badge: "c-badge--neutral" },
};

/** The state of a person or of an administrator of the server, as a badge. */
export function StatusBadge({ status }: { status: string }) {
	const s = statuses[status] ?? statuses.active;
	return <span class={`c-badge ${s.badge}`}>{s.word}</span>;
}

/** A handle drawn from a name, as in "ana" from "Ana Lima": the one who adds them may change it. */
export function handleFrom(name: string): string {
	const first = name.trim().split(/\s+/)[0] ?? "";
	return first
		.normalize("NFD")
		.toLowerCase()
		.replace(/[^a-z0-9-]/g, "")
		.replace(/^[^a-z]+/, "")
		.slice(0, 32)
		.replace(/-+$/, "");
}

/**
 * The link where someone chooses their password, shown once, at page: /activate
 * for the people of a Cahut, /server/activate for the administrators of the server.
 */
export function ActivationResult(props: { name: string; email?: string; activation: Activation; page: string }) {
	const { name, email, activation } = props;
	const until = new Date(activation.expires_at).toLocaleDateString(locale, { dateStyle: "long" });
	return (
		<div class="result">
			<p class="c-small">
				{activation.emailed ? t.emailed(name, email, until) : t.send(name, until)}
			</p>
			<CopyField label={t.activationLink} value={`${location.origin}${href(props.page)}#${activation.code}`} />
		</div>
	);
}

/** What came of a change: a line, or more, such as a link. */
export function Outcome({ children }: { children: ComponentChildren }) {
	if (typeof children !== "string") return <div class="setting__more">{children}</div>;
	return (
		<p class="setting__more setting__done" role="status">
			<Check class="i" size={16} aria-hidden="true" />
			{children}
		</p>
	);
}

/**
 * One change, in a dialog that manages someone: what it does, its button,
 * which may ask first, and what came of it.
 */
export function Setting(props: {
	title: string;
	help: string;
	action: string;
	danger?: boolean;
	/** The question to answer before the change, if it asks first. */
	ask?: string;
	onAct: () => Promise<ComponentChildren>;
}) {
	const [asking, setAsking] = useState(false);
	const [busy, setBusy] = useState(false);
	const [outcome, setOutcome] = useState<ComponentChildren>(null);
	const [error, setError] = useState("");
	const ask = () => {
		setError("");
		setOutcome(null);
		setAsking(true);
	};
	const act = async () => {
		setBusy(true);
		setError("");
		setOutcome(null);
		try {
			setOutcome(await props.onAct());
			setAsking(false);
		} catch (err) {
			setError(explain(err));
		} finally {
			setBusy(false);
		}
	};
	return (
		<div class="setting">
			<div class="setting__text">
				<span class="setting__title">{props.title}</span>
				<p class={asking ? "c-small" : "c-small c-muted"}>{asking ? props.ask : props.help}</p>
			</div>
			<div class="setting__action">
				{asking && (
					<button class="c-btn c-btn--ghost" type="button" onClick={() => setAsking(false)}>
						{t.cancel}
					</button>
				)}
				{/* The answer takes the place of the button that asks: the second click of a double-click acts on nothing. */}
				<button
					class={`c-btn ${props.danger ? "c-btn--danger" : asking ? "c-btn--primary" : "c-btn--secondary"}`}
					type="button"
					disabled={busy}
					onClick={(e) => {
						if (e.detail > 1) return;
						if (props.ask && !asking) ask();
						else act();
					}}
				>
					{props.action}
				</button>
			</div>
			{outcome && <Outcome>{outcome}</Outcome>}
			{error && (
				<div class="setting__more">
					<Failure message={error} />
				</div>
			)}
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
				{t.cancel}
			</button>
		</span>
	);
}

const kinds: Record<string, { icon: typeof FileText; name: string }> = {
	document: { icon: FileText, name: t.document },
	deck: { icon: Presentation, name: t.deck },
	dashboard: { icon: ChartColumn, name: t.dashboard },
	site: { icon: AppWindow, name: t.site },
};

/** The access levels of a Window, with their badge. */
export const accesses: Record<string, { icon: typeof Lock; name: string; who: string }> = {
	room: { icon: Lock, name: t.roomOnly, who: t.roomWho },
	internal: { icon: Users, name: t.internal, who: t.internalWho },
	key: { icon: KeyRound, name: t.key, who: t.keyWho },
	public: { icon: Globe, name: t.public, who: t.publicWho },
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
export function WindowCards({ windows, showRoom, home }: { windows: Window[]; showRoom?: boolean; home?: string }) {
	const api = useConsole();
	const names = windows.slice(0, maxPreviews).map((w) => `${w.room}/${w.name}`);
	// Without previews, the cards keep the icon of their kind.
	const { data } = useLoad(() => api.cahut("preview_windows", { windows: names }).catch(() => null), [names.join(" ")]);
	const frames = new Map(data?.previews.map((p) => [p.window, p.frame]));
	return (
		<div class="cards cards--windows">
			{windows.map((w) => (
				<WindowCard key={w.url} window={w} showRoom={showRoom} home={w.name === home} preview={frames.get(`${w.room}/${w.name}`)} />
			))}
		</div>
	);
}

/**
 * A Window, as a card that leads to its page, with a button that opens the
 * Window itself in a new tab; the Window that the address of its Room shows
 * says so.
 */
export function WindowCard(props: { window: Window; showRoom?: boolean; home?: boolean; preview?: string }) {
	const { window: w, showRoom, home, preview } = props;
	const kind = kinds[w.kind] ?? kinds.site;
	const Icon = kind.icon;
	return (
		<article class="c-card c-window window-card">
			<Link class="c-window__thumb window-thumb" href={windowPage(w)} aria-label={home ? `${w.title}, ${t.roomPage}` : w.title}>
				<Icon size={40} aria-hidden="true" />
				{preview && <WindowPreview src={preview} title={w.title} />}
				{home && (
					<span class="c-badge window-flag" aria-hidden="true">
						<House class="i" size={14} aria-hidden="true" />
						{t.roomPage}
					</span>
				)}
			</Link>
			<a class="c-btn c-btn--secondary c-btn--sm icon-button window-open" href={w.url} target="_blank" rel="noopener noreferrer" aria-label={t.openWindow(w.title)} title={t.openWindow(w.title)}>
				<ArrowUpRight class="i" size={16} aria-hidden="true" />
			</a>
			<div class="c-window__body">
				<Link class="c-window__title window-title" href={windowPage(w)}>
					<Icon class="i" size={16} aria-hidden="true" />
					{w.title}
				</Link>
				<div class="c-window__meta">
					<span>
						{[kind.name, ago(w.updated_at), showRoom ? w.room : w.created_by].filter(Boolean).join(" · ")}
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
			title={t.preview(title)}
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
	if (minutes < 1) return t.justNow;
	if (minutes < 60) return t.minutesAgo(minutes);
	if (minutes < 24 * 60) return t.hoursAgo(Math.round(minutes / 60));
	if (minutes < 48 * 60) return t.yesterday;
	return then.toLocaleDateString(lang === "en" ? "en-US" : locale, { month: "short", day: "numeric" });
}

/**
 * Says how long ago a time was, within a sentence: "2 min ago", "yesterday",
 * "oct 12"; in French "il y a 2 min", "hier", "le 12 oct.".
 */
export function agoInText(at: string): string {
	const said = ago(at);
	return lang === "fr" && /^\d/.test(said) ? `le ${said}` : said.toLowerCase();
}

/** Says a size in bytes, as in "12 KB", "3.4 MB" or "10 GB", in units of 1000 bytes. */
export function bytes(n: number): string {
	const size = (v: number, unit: string, digits = 0, fixed = false) =>
		`${v.toLocaleString(locale, { minimumFractionDigits: fixed ? digits : 0, maximumFractionDigits: digits })}${lang === "fr" ? "\u00a0" : " "}${unit}`;
	if (n < 1000) return size(n, t.bytes);
	if (n < 1000 * 1000) return size(Math.round(n / 1000), t.kb);
	if (n < 1000 * 1000 * 1000) return size(n / 1000 / 1000, t.mb, 1, true);
	return size(n / 1000 / 1000 / 1000, t.gb, 1);
}

/** The initials of a name, for its avatar. */
export function initials(name: string): string {
	const words = name.trim().split(/\s+/);
	return (words[0].charAt(0) + (words.length > 1 ? words[words.length - 1].charAt(0) : "")).toUpperCase();
}
