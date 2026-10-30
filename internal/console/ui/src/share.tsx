// The menu Share of a Window. It shares its permanent address: copied,
// sent by e-mail, posted through the addresses of sharing of X, LinkedIn and
// Facebook, or given to the share sheet of the system. Its entries are plain
// links: the console loads nothing of those sites, and they learn of the
// Window only when the person posts it. For a Window that is not public, it
// also makes share links, which open it to whoever holds them, without an
// account or a Key, until they end. Its author also shares it once: a link
// that opens to a single reader, to a Window that they keep for themself.
// See docs/adr/0051-sharing-a-window-and-its-preview.md,
// docs/adr/0052-share-links.md and
// docs/adr/0053-windows-for-their-author-alone.md.
import { ArrowUpRight, Check, Link as LinkIcon, Link2, Mail, Share2, UserLock } from "lucide-preact";
import { useEffect, useId, useRef, useState } from "preact/hooks";
import { explain } from "./api";
import type { ShareLinkInput, ShareLinkMade, Window } from "./api.gen";
import { CopyField, Dialog, Failure } from "./components";
import { useConsole } from "./context";
import { locale, plural, words } from "./i18n";

/** The most readers of a share link, and the most characters of its label, as the server counts them. */
const maxReaders = 1000;
const maxLabel = 100;

const t = words({
	en: {
		share: "Share",
		shareWindow: (title: string) => `Share ${title}`,
		copyLink: "Copy the link",
		copied: "Link copied",
		email: "E-mail",
		shareVia: "Share via…",
		createLink: "Create a share link…",
		shareOnce: "Share once…",
		newTab: "opens in a new tab",
		linkTitle: "Create a share link",
		linkIntro:
			"Whoever holds the link opens this Window, without an account or a Key, until the link ends, besides its access. Make one link for each sharing, and take it back on the page of the Window.",
		label: "Label",
		labelExample: "For Acme",
		labelHelp: "Optional. A few words that tell your links apart, such as For Acme. Only the members of the Room see it.",
		lasts: "It opens the Window",
		days: (n: number) => (n === 1 ? "For a day" : `For ${n} days`),
		untilDay: "Until a day I choose",
		day: "The last day",
		readers: "Readers",
		anyReaders: "Any number of readers",
		someReaders: "A limited number of readers",
		readersLabel: "How many",
		readersHelp: "Each reader who opens it spends one. Once they are spent, it opens nobody new, and those who opened it read on until it ends.",
		create: "Create the link",
		cancel: "Cancel",
		made: "The link is ready",
		theLink: "the link",
		once: "Copy it now: Cahut keeps only a trace of it, and shows it this once.",
		opensUntil: (until: string, readers: number | undefined) =>
			`It opens the Window until ${until}${readers ? `, to ${plural(readers, "reader", "readers")}` : ""}.`,
		onceTitle: "Share once",
		onceIntro:
			"The link opens this Window once: whoever opens it first reads it in their browser for 4 hours, and nobody after them. Robots of previews and filters of mail that only follow the link spend nothing.",
		keep: "Keep it for me alone",
		keepHelp: "Nobody else in its Room sees it any more, nor their AIs. You open it to the Room again on its page.",
		opensOnce: (until: string) => `It opens the Window once, until ${until}.`,
		done: "Done",
		opens: {
			author: "Only you open it at this address. To give it to one reader, share it once.",
			public: "Anyone with the link opens it. Its preview shows its title and its description.",
			key: "Readers need its Key too, which you give them apart. Its preview shows its title.",
			internal: "People of this Cahut open it once signed in. Its preview shows no title.",
			room: "The members of its Room open it once signed in. Its preview shows no title.",
		} as Record<string, string>,
	},
	fr: {
		share: "Partager",
		shareWindow: (title) => `Partager ${title}`,
		copyLink: "Copier le lien",
		copied: "Lien copié",
		email: "E-mail",
		shareVia: "Partager via…",
		createLink: "Créer un lien de partage…",
		shareOnce: "Partager une fois…",
		newTab: "s'ouvre dans un nouvel onglet",
		linkTitle: "Créer un lien de partage",
		linkIntro:
			"Qui détient le lien ouvre cette Window, sans compte ni Key, jusqu'à la fin du lien, en plus de son accès. Créez un lien pour chaque partage, et retirez-le sur la page de la Window.",
		label: "Étiquette",
		labelExample: "Pour Acme",
		labelHelp: "Facultative. Quelques mots pour distinguer vos liens, comme Pour Acme. Seuls les membres de la Room la voient.",
		lasts: "Il ouvre la Window",
		days: (n) => (n === 1 ? "Pendant un jour" : `Pendant ${n} jours`),
		untilDay: "Jusqu'à un jour que je choisis",
		day: "Le dernier jour",
		readers: "Lecteurs",
		anyReaders: "Sans limite de lecteurs",
		someReaders: "Un nombre limité de lecteurs",
		readersLabel: "Combien",
		readersHelp:
			"Chaque lecteur qui l'ouvre en consomme un. Une fois tous consommés, il n'ouvre plus à personne de nouveau, et ceux qui l'ont ouvert lisent jusqu'à sa fin.",
		create: "Créer le lien",
		cancel: "Annuler",
		made: "Le lien est prêt",
		theLink: "le lien",
		once: "Copiez-le maintenant : Cahut n'en garde qu'une empreinte, et ne le montre que cette fois.",
		opensUntil: (until, readers) => `Il ouvre la Window jusqu'au ${until}${readers ? `, à ${plural(readers, "lecteur", "lecteurs")}` : ""}.`,
		onceTitle: "Partager une fois",
		onceIntro:
			"Le lien ouvre cette Window une fois : la première personne qui l'ouvre la lit dans son navigateur pendant 4 heures, et personne après elle. Les robots d'aperçu et les filtres de messagerie qui suivent seulement le lien ne consomment rien.",
		keep: "Me la réserver",
		keepHelp: "Personne d'autre dans sa Room ne la voit plus, ni ses IA. Vous la rouvrez à la Room sur sa page.",
		opensOnce: (until) => `Il ouvre la Window une fois, jusqu'au ${until}.`,
		done: "Terminé",
		opens: {
			author: "Personne d'autre que vous ne l'ouvre à cette adresse. Pour la donner à une personne, partagez-la une fois.",
			public: "Quiconque a le lien l'ouvre. Son aperçu montre son titre et sa description.",
			key: "Il faut aussi sa Key, que vous donnez à part. Son aperçu montre son titre.",
			internal: "Les personnes de cette Cahut l'ouvrent une fois connectées. Son aperçu ne montre pas de titre.",
			room: "Les membres de sa Room l'ouvrent une fois connectés. Son aperçu ne montre pas de titre.",
		},
	},
});

/** The addresses of sharing of the networks, for a link and its title. */
const networks: { name: string; href: (url: string, title: string) => string }[] = [
	{ name: "X", href: (url, title) => `https://x.com/intent/tweet?text=${encodeURIComponent(title)}&url=${encodeURIComponent(url)}` },
	// LinkedIn reads the link only, and the rest from its preview.
	{ name: "LinkedIn", href: (url) => `https://www.linkedin.com/sharing/share-offsite/?url=${encodeURIComponent(url)}` },
	{ name: "Facebook", href: (url) => `https://www.facebook.com/sharer/sharer.php?u=${encodeURIComponent(url)}` },
];

/** A message with the title, the description and the link, and no words of Cahut. */
export function mailto(w: Pick<Window, "title" | "description" | "permanent_url">): string {
	const body = w.description ? `${w.description}\r\n\r\n${w.permanent_url}` : w.permanent_url;
	return `mailto:?subject=${encodeURIComponent(w.title)}&body=${encodeURIComponent(body)}`;
}

/**
 * A button that opens the menu Share of a Window, under it, over the page:
 * a popover, which no card cuts. It says who opens the link; X, LinkedIn
 * and Facebook show for a public Window and a Window with a Key only, since
 * the others open to nobody outside the Cahut.
 */
export function ShareMenu({ window: w, compact, onLinks }: { window: Window; compact?: boolean; onLinks?: () => void }) {
	const { session } = useConsole();
	const id = useId();
	const button = useRef<HTMLButtonElement>(null);
	const menu = useRef<HTMLDivElement>(null);
	const [open, setOpen] = useState(false);
	const [copied, setCopied] = useState(false);
	const [making, setMaking] = useState<"link" | "once" | null>(null);
	const outside = w.access === "public" || w.access === "key";
	// Its author shares it once, and keeps it for themself.
	const author = w.created_by === session.handle;
	const native = typeof navigator.share === "function";

	// The menu stays under its button, its right edge under the button's
	// within the screen, or over it near the bottom of the screen, while the
	// page scrolls.
	const place = () => {
		const b = button.current?.getBoundingClientRect();
		const m = menu.current;
		if (!b || !m) return;
		const gap = 8;
		const width = document.documentElement.clientWidth;
		m.style.right = `${Math.max(gap, Math.min(width - b.right, width - m.offsetWidth - gap))}px`;
		const below = b.bottom + gap + m.offsetHeight <= innerHeight || b.top < m.offsetHeight + gap;
		m.style.top = below ? `${b.bottom + gap}px` : "auto";
		m.style.bottom = below ? "auto" : `${innerHeight - b.top + gap}px`;
	};
	useEffect(() => {
		if (!open) return;
		addEventListener("scroll", place, true);
		addEventListener("resize", place);
		return () => {
			removeEventListener("scroll", place, true);
			removeEventListener("resize", place);
		};
	}, [open]);

	const copy = async () => {
		await navigator.clipboard.writeText(w.permanent_url);
		setCopied(true);
		setTimeout(() => setCopied(false), 2000);
	};
	const shareVia = async () => {
		menu.current?.hidePopover();
		try {
			await navigator.share({ title: w.title, text: w.description || undefined, url: w.permanent_url });
		} catch {
			// The person closed the share sheet.
		}
	};
	return (
		<>
			<button
				ref={button}
				class={compact ? "c-btn c-btn--secondary c-btn--sm icon-button" : "c-btn c-btn--secondary"}
				type="button"
				popovertarget={id}
				aria-label={compact ? t.shareWindow(w.title) : undefined}
				title={compact ? t.shareWindow(w.title) : undefined}
			>
				<Share2 class="i" size={16} aria-hidden="true" />
				{!compact && t.share}
			</button>
			<div
				ref={menu}
				id={id}
				popover="auto"
				class="c-card share-menu"
				aria-label={t.shareWindow(w.title)}
				// It shows once placed: under its button, or over it when the
				// screen ends first, which its height says only once it shows.
				onBeforeToggle={(e) => {
					if (e.newState === "open" && menu.current) menu.current.style.visibility = "hidden";
				}}
				onToggle={(e) => {
					if (e.newState === "open" && menu.current) {
						place();
						menu.current.style.visibility = "";
					}
					setOpen(e.newState === "open");
					setCopied(false);
				}}
			>
				<p class="share-menu__note">{t.opens[w.access] ?? t.opens.room}</p>
				<button class="share-menu__item" type="button" onClick={copy}>
					{copied ? <Check class="i" size={18} aria-hidden="true" /> : <LinkIcon class="i" size={18} aria-hidden="true" />}
					<span aria-live="polite">{copied ? t.copied : t.copyLink}</span>
				</button>
				{w.access !== "public" && (
					<button
						class="share-menu__item"
						type="button"
						onClick={() => {
							menu.current?.hidePopover();
							setMaking("link");
						}}
					>
						<Link2 class="i" size={18} aria-hidden="true" />
						<span>{t.createLink}</span>
					</button>
				)}
				{w.access !== "public" && author && (
					<button
						class="share-menu__item"
						type="button"
						onClick={() => {
							menu.current?.hidePopover();
							setMaking("once");
						}}
					>
						<UserLock class="i" size={18} aria-hidden="true" />
						<span>{t.shareOnce}</span>
					</button>
				)}
				<a class="share-menu__item" href={mailto(w)}>
					<Mail class="i" size={18} aria-hidden="true" />
					<span>{t.email}</span>
				</a>
				{outside &&
					networks.map((n) => (
						<a key={n.name} class="share-menu__item" href={n.href(w.permanent_url, w.title)} target="_blank" rel="noopener noreferrer">
							<span aria-hidden="true" />
							<span>
								{n.name}
								<span class="visually-hidden">, {t.newTab}</span>
							</span>
							<ArrowUpRight class="i share-menu__hint" size={14} aria-hidden="true" />
						</a>
					))}
				{native && (
					<button class="share-menu__item" type="button" onClick={shareVia}>
						<Share2 class="i" size={18} aria-hidden="true" />
						<span>{t.shareVia}</span>
					</button>
				)}
				<code class="share-menu__url c-mono">{w.permanent_url}</code>
			</div>
			{making && <ShareLinkDialog window={w} once={making === "once"} onMade={onLinks} onClose={() => setMaking(null)} />}
		</>
	);
}

/** How long a share link lasts: a number of days, or until a day chosen. */
type Lasts = "1" | "7" | "30" | "day";

/**
 * The dialog that makes a share link to a Window, with a label, an end and
 * a number of readers, then shows the link this once. A link that opens
 * once keeps the Window for its author first, unless they say otherwise.
 */
export function ShareLinkDialog({
	window: w,
	once,
	onMade,
	onClose,
}: { window: Window; once?: boolean; onMade?: () => void; onClose: () => void }) {
	const api = useConsole();
	const labelId = useId();
	const [keep, setKeep] = useState(true);
	const [label, setLabel] = useState("");
	const [lasts, setLasts] = useState<Lasts>("7");
	const [day, setDay] = useState("");
	const [limited, setLimited] = useState(false);
	const [readers, setReaders] = useState("1");
	const [made, setMade] = useState<ShareLinkMade | null>(null);
	const [error, setError] = useState("");
	const [busy, setBusy] = useState(false);
	// A day of the next 29: the link ends at its end, where the person is,
	// within the 30 days that a link lasts at most.
	const today = new Date();
	const dayIn = (n: number) => new Date(today.getFullYear(), today.getMonth(), today.getDate() + n).toLocaleDateString("en-CA");

	const submit = async (e: Event) => {
		e.preventDefault();
		setError("");
		setBusy(true);
		const input: ShareLinkInput = { window: w.id };
		if (label.trim()) input.label = label;
		if (lasts === "1" || lasts === "30") input.ends_at = new Date(Date.now() + Number(lasts) * 24 * 3600 * 1000).toISOString();
		if (lasts === "day") input.ends_at = new Date(`${day}T23:59:59`).toISOString();
		if (limited) input.max_openings = Number(readers);
		if (once) input.max_openings = 1;
		try {
			if (once && keep && w.access !== "author") {
				await api.cahut("set_window_access", { window: w.id, access: "author" });
				onMade?.();
			}
			setMade(await api.cahut("create_share_link", input));
			onMade?.();
		} catch (err) {
			setError(explain(err));
		} finally {
			setBusy(false);
		}
	};

	const title = once ? t.onceTitle : t.linkTitle;
	if (made) {
		const until = new Date(made.expires_at).toLocaleString(locale, { dateStyle: "long", timeStyle: "short" });
		return (
			<Dialog title={title} meta={w.title} onClose={onClose}>
				<div class="c-alert c-alert--success" role="status">
					<Check class="i" size={16} aria-hidden="true" />
					<div>
						<strong>{t.made}</strong> {t.once}
					</div>
				</div>
				<CopyField label={t.theLink} value={made.link} />
				<p class="c-small c-muted dialog__intro">{once ? t.opensOnce(until) : t.opensUntil(until, made.max_openings)}</p>
				<div class="dialog__foot">
					<button class="c-btn c-btn--primary" type="button" onClick={onClose}>
						{t.done}
					</button>
				</div>
			</Dialog>
		);
	}
	return (
		<Dialog title={title} meta={w.title} onClose={onClose}>
			<p class="c-small c-muted dialog__intro">{once ? t.onceIntro : t.linkIntro}</p>
			<form class="form" onSubmit={submit}>
				<div class="c-field">
					<label class="c-field__label" for={labelId}>
						{t.label}
					</label>
					<input id={labelId} class="c-input" maxLength={maxLabel} placeholder={t.labelExample} value={label} onInput={(e) => setLabel(e.currentTarget.value)} />
					<span class="c-field__help">{t.labelHelp}</span>
				</div>
				<fieldset class="consent__group">
					<legend class="c-field__label">{t.lasts}</legend>
					{(["1", "7", "30"] as const).map((n) => (
						<label class="c-check" key={n}>
							<input type="radio" name="lasts" checked={lasts === n} onChange={() => setLasts(n)} />
							<span>{t.days(Number(n))}</span>
						</label>
					))}
					<label class="c-check">
						<input type="radio" name="lasts" checked={lasts === "day"} onChange={() => setLasts("day")} />
						<span>{t.untilDay}</span>
					</label>
					{lasts === "day" && (
						<label class="c-field">
							<span class="c-field__label">{t.day}</span>
							<input class="c-input" type="date" required min={dayIn(0)} max={dayIn(29)} value={day} onInput={(e) => setDay(e.currentTarget.value)} />
						</label>
					)}
				</fieldset>
				{once && w.access !== "author" && (
					<div class="c-field">
						<label class="c-check">
							<input type="checkbox" checked={keep} onChange={(e) => setKeep(e.currentTarget.checked)} />
							<span>{t.keep}</span>
						</label>
						<span class="c-field__help">{t.keepHelp}</span>
					</div>
				)}
				{!once && (
					<fieldset class="consent__group">
						<legend class="c-field__label">{t.readers}</legend>
						<label class="c-check">
							<input type="radio" name="readers" checked={!limited} onChange={() => setLimited(false)} />
							<span>{t.anyReaders}</span>
						</label>
						<label class="c-check">
							<input type="radio" name="readers" checked={limited} onChange={() => setLimited(true)} />
							<span>{t.someReaders}</span>
						</label>
						{limited && (
							<label class="c-field">
								<span class="c-field__label">{t.readersLabel}</span>
								<input
									class="c-input"
									type="number"
									required
									min={1}
									max={maxReaders}
									value={readers}
									onInput={(e) => setReaders(e.currentTarget.value)}
								/>
								<span class="c-field__help">{t.readersHelp}</span>
							</label>
						)}
					</fieldset>
				)}
				{error && <Failure message={error} />}
				<div class="dialog__foot">
					<button class="c-btn c-btn--secondary" type="button" onClick={onClose}>
						{t.cancel}
					</button>
					<button class="c-btn c-btn--primary" type="submit" disabled={busy}>
						<Link2 class="i" size={16} aria-hidden="true" />
						{t.create}
					</button>
				</div>
			</form>
		</Dialog>
	);
}
