// The menu Share of a Window. It shares its permanent address: copied,
// sent by e-mail, posted through the addresses of sharing of X, LinkedIn and
// Facebook, or given to the share sheet of the system. Its entries are plain
// links: the console loads nothing of those sites, and they learn of the
// Window only when the person posts it. See
// docs/adr/0051-sharing-a-window-and-its-preview.md.
import { ArrowUpRight, Check, Link as LinkIcon, Mail, Share2 } from "lucide-preact";
import { useEffect, useId, useRef, useState } from "preact/hooks";
import type { Window } from "./api.gen";
import { words } from "./i18n";

const t = words({
	en: {
		share: "Share",
		shareWindow: (title: string) => `Share ${title}`,
		copyLink: "Copy the link",
		copied: "Link copied",
		email: "E-mail",
		shareVia: "Share via…",
		newTab: "opens in a new tab",
		opens: {
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
		newTab: "s'ouvre dans un nouvel onglet",
		opens: {
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
export function ShareMenu({ window: w, compact }: { window: Window; compact?: boolean }) {
	const id = useId();
	const button = useRef<HTMLButtonElement>(null);
	const menu = useRef<HTMLDivElement>(null);
	const [open, setOpen] = useState(false);
	const [copied, setCopied] = useState(false);
	const outside = w.access === "public" || w.access === "key";
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
		</>
	);
}
