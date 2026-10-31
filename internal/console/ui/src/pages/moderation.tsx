// Moderation: the reports that anyone makes about the Windows of this Cahut,
// from the page /-/report, for its moderators, its administrators and those
// of its server. Each report waits for a decision: the moderators leave its
// Window as it is, or withdraw it. See
// docs/adr/0054-reporting-and-withdrawing-a-window.md.
import { ArrowUpRight, Eye, Flag } from "lucide-preact";
import { createContext } from "preact";
import { useContext, useState } from "preact/hooks";
import { type CahutCall, explain } from "../api";
import type { PublicWindow, Report, ReportedPreview } from "../api.gen";
import { AccessBadge, ConfirmButton, Dialog, Failure, Notice, PageHead } from "../components";
import { useConsole } from "../context";
import { locale, plural, words } from "../i18n";
import { useLoad } from "../load";

const t = words({
	en: {
		label: "This Cahut",
		title: "Moderation",
		lead: "Anyone reports a Window of this Cahut, without an account, from the page Report a Window that the pages of this Cahut link. Its administrators read the reports here, and decide.",
		onlyModerators: "Only the administrators of this Cahut, and those of its server, read its reports.",
		waiting: "Waiting for a decision",
		nothingWaits: "No report waits.",
		decided: "Decided",
		categories: {
			phishing: "Phishing or a scam",
			malware: "Malware",
			illegal: "Illegal content",
			rights: "Someone's rights",
			other: "Something else",
		} as Record<string, string>,
		report: (id: number) => `Report ${id}`,
		reportedOn: (when: string) => `Reported on ${when}`,
		from: (who: string) => `From ${who}`,
		anonymous: "Gave neither a name nor an address",
		trashed: "In the trash",
		openAtAddress: "Open at its address",
		open: "Open",
		preview: (title: string) => `The Window ${title}`,
		openedNotice: "The members of its Room read in the Logbook that you opened it, with this report. It runs in its sandbox, without its forms.",
		withdraw: "Withdraw",
		reason: "Reason, for the members of its Room",
		reasonHelp: (left: number) => `They read it in the console. ${plural(left, "character", "characters")} left.`,
		confirmWithdraw: "Withdraw the Window",
		withdrawnNow: "Withdrawn: it serves nothing to its readers until it is reinstated.",
		reinstate: "Reinstate",
		reinstateQuestion: "Give the Window back as it was?",
		dismiss: "Leave it as it is",
		note: "Note for the moderators",
		noteHelp: (left: number) => `Optional. ${plural(left, "character", "characters")} left.`,
		confirmDismiss: "Leave the Window as it is",
		cancel: "Cancel",
		dismissedBy: (who: string, when: string) => `Left as it is by ${who} on ${when}.`,
		withdrawnBy: (who: string, when: string) => `Withdrawn by ${who} on ${when}.`,
		serverAdmin: (handle: string) => (handle ? `${handle}, an administrator of the server` : "the command line of the server"),
		waitingCount: (n: number) => plural(n, "report waits", "reports wait"),
		publicTitle: "Public Windows",
		publicLead: "What this Cahut serves to anyone, the latest made public first. You open them at their address, as anyone does, without a report.",
		noPublic: "This Cahut has no public Window.",
		publicSince: (who: string, when: string) => (who ? `Published by ${who}, public since ${when}` : `Public since ${when}`),
	},
	fr: {
		label: "Cette Cahut",
		title: "Modération",
		lead: "Tout le monde signale une Window de cette Cahut, sans compte, depuis la page Report a Window que ses pages relient. Ses administrateurs lisent les signalements ici, et tranchent.",
		onlyModerators: "Seuls les administrateurs de cette Cahut, et ceux de son serveur, lisent ses signalements.",
		waiting: "En attente d'une décision",
		nothingWaits: "Aucun signalement n'attend.",
		decided: "Tranchés",
		categories: {
			phishing: "Hameçonnage ou arnaque",
			malware: "Logiciel malveillant",
			illegal: "Contenu illégal",
			rights: "Droits d'une personne",
			other: "Autre chose",
		},
		report: (id) => `Signalement ${id}`,
		reportedOn: (when) => `Signalé le ${when}`,
		from: (who) => `De ${who}`,
		anonymous: "Sans nom ni adresse",
		trashed: "À la corbeille",
		openAtAddress: "Ouvrir à son adresse",
		open: "Ouvrir",
		preview: (title) => `La Window ${title}`,
		openedNotice: "Les membres de sa Room lisent dans le Logbook que vous l'avez ouverte, avec ce signalement. Elle s'exécute dans son bac à sable, sans ses formulaires.",
		withdraw: "Retirer",
		reason: "Raison, pour les membres de sa Room",
		reasonHelp: (left) => `Ils la lisent dans la console. ${plural(left, "caractère restant", "caractères restants")}.`,
		confirmWithdraw: "Retirer la Window",
		withdrawnNow: "Retirée : ses lecteurs n'y ont plus accès jusqu'à ce qu'elle soit remise en ligne.",
		reinstate: "Remettre en ligne",
		reinstateQuestion: "Rendre la Window telle qu'elle était ?",
		dismiss: "La laisser telle quelle",
		note: "Note pour les modérateurs",
		noteHelp: (left) => `Facultative. ${plural(left, "caractère restant", "caractères restants")}.`,
		confirmDismiss: "Laisser la Window telle quelle",
		cancel: "Annuler",
		dismissedBy: (who, when) => `Laissée telle quelle par ${who} le ${when}.`,
		withdrawnBy: (who, when) => `Retirée par ${who} le ${when}.`,
		serverAdmin: (handle) => (handle ? `${handle}, administrateur du serveur` : "la ligne de commande du serveur"),
		waitingCount: (n) => plural(n, "signalement attend", "signalements attendent"),
		publicTitle: "Windows publiques",
		publicLead: "Ce que cette Cahut sert à tout le monde, de la dernière rendue publique à la plus ancienne. Vous les ouvrez à leur adresse, comme tout le monde, sans signalement.",
		noPublic: "Cette Cahut n'a aucune Window publique.",
		publicSince: (who, when) => (who ? `Publiée pour ${who}, publique depuis le ${when}` : `Publique depuis le ${when}`),
	},
});

/** The most characters of the note of a decision, and of the reason of a withdrawal. */
const maxNote = 500;

function when(at: string): string {
	return new Date(at).toLocaleString(locale, { dateStyle: "medium", timeStyle: "short" });
}

/** Who decided a report: an administrator of the Cahut, or of the server. */
function decider(p: Report): string {
	return p.by_server ? t.serverAdmin(p.decided_by ?? "") : (p.decided_by ?? "");
}

/**
 * Who moderates, as the cards of the page see it: how they call the API of
 * the Cahut, its origin, where its Windows open, whether they act for the
 * host, who alone reinstates what the administrators of the server
 * withdrew, and what to count again after a decision. See
 * docs/adr/0054-reporting-and-withdrawing-a-window.md and
 * docs/adr/0058-a-console-for-the-server.md.
 */
interface Moderator {
	call: CahutCall;
	origin: string;
	host: boolean;
	onDecided: () => void;
}

const moderator = createContext<Moderator | null>(null);

function useModerator(): Moderator {
	const m = useContext(moderator);
	if (!m) {
		throw new Error("useModerator outside a page of moderation");
	}
	return m;
}

/** A report, as a card, with what the moderators can decide while it waits. */
function ReportCard(props: { report: Report; onDecided: () => void }) {
	const { call, origin, host, onDecided } = useModerator();
	const p = props.report;
	// The decision being made: leave the Window as it is, or withdraw it.
	const [asking, setAsking] = useState<"dismiss" | "withdraw" | null>(null);
	const [note, setNote] = useState("");
	const [busy, setBusy] = useState(false);
	const [error, setError] = useState("");
	const [shown, setShown] = useState<ReportedPreview | null>(null);
	const w = p.window;
	const open = async () => {
		setError("");
		try {
			setShown(await call("preview_reported_window", { report: p.id }));
		} catch (err) {
			setError(explain(err));
		}
	};
	const address = `${origin}/${w.room}/${w.name}/`;
	const run = async (act: () => Promise<unknown>) => {
		setBusy(true);
		setError("");
		try {
			await act();
			onDecided();
			props.onDecided();
		} catch (err) {
			setError(explain(err));
		}
		setBusy(false);
	};
	const decide = (e: Event) => {
		e.preventDefault();
		run(() =>
			asking === "withdraw"
				? call("withdraw_window", { window: w.id, reason: note })
				: call("dismiss_report", { report: p.id, note }),
		);
	};
	const contact = [p.name, p.email].filter(Boolean).join(", ");
	return (
		<article class={`c-card panel report${p.decision ? " report--decided" : ""}`}>
			<div class="report__head">
				<span class={`c-badge ${p.decision ? "c-badge--neutral" : "c-badge--warning"}`}>
					<Flag class="i" size={14} aria-hidden="true" />
					{t.categories[p.category] ?? p.category}
				</span>
				<span class="c-mono c-muted report__meta">
					{t.report(p.id)} · {t.reportedOn(when(p.created_at))}
				</span>
			</div>
			<p class="report__window">
				<strong>{w.title}</strong> <span class="c-mono c-muted">{`${w.room}/${w.name}`}</span> <AccessBadge access={w.access} />
				{w.trashed && <span class="c-badge c-badge--neutral">{t.trashed}</span>}
			</p>
			<p class="report__explanation">{p.explanation}</p>
			{!p.decision && <p class="c-small c-muted">{contact ? t.from(contact) : t.anonymous}</p>}
			{p.decision && (
				<p class="c-small c-muted">
					{(p.decision === "withdrawn" ? t.withdrawnBy : t.dismissedBy)(decider(p), when(p.decided_at ?? p.created_at))}
					{p.note && ` ${p.note}`}
				</p>
			)}
			{p.decision && w.withdrawn && (
				<div class="actions">
					<span class="c-badge c-badge--danger">{t.withdrawnNow}</span>
					{/* What the administrators of the server withdrew, they alone
					    reinstate. */}
					{(host || !p.by_server) && (
						<ConfirmButton
							label={t.reinstate}
							question={t.reinstateQuestion}
							confirm={t.reinstate}
							onConfirm={() => run(() => call("reinstate_window", { window: w.id }))}
						/>
					)}
				</div>
			)}
			{error && <Failure message={error} />}
			{!p.decision && !asking && (
				<div class="actions">
					{!w.trashed && (
						<button class="c-btn c-btn--secondary c-btn--sm" type="button" onClick={open}>
							<Eye class="i" size={14} aria-hidden="true" />
							{t.open}
						</button>
					)}
					{w.access === "public" && !w.trashed && (
						<a class="c-btn c-btn--secondary c-btn--sm" href={address} target="_blank" rel="noopener noreferrer">
							<ArrowUpRight class="i" size={14} aria-hidden="true" />
							{t.openAtAddress}
						</a>
					)}
					<button class="c-btn c-btn--secondary c-btn--sm" type="button" onClick={() => setAsking("withdraw")}>
						{t.withdraw}
					</button>
					<button class="c-btn c-btn--secondary c-btn--sm" type="button" onClick={() => setAsking("dismiss")}>
						{t.dismiss}
					</button>
				</div>
			)}
			{shown && (
				<Dialog
					title={w.title}
					meta={<span class="c-mono c-small c-muted">{`${shown.window} · ${t.report(p.id)}`}</span>}
					onClose={() => setShown(null)}
				>
					<p class="c-small c-muted">{t.openedNotice}</p>
					<iframe class="report__frame" src={`${origin}${shown.frame}`} title={t.preview(w.title)} sandbox="allow-scripts" referrerpolicy="no-referrer" />
				</Dialog>
			)}
			{!p.decision && asking && (
				<form class="report__decide" onSubmit={decide}>
					<label class="c-field">
						<span class="c-field__label">{asking === "withdraw" ? t.reason : t.note}</span>
						<textarea
							class="c-input"
							required={asking === "withdraw"}
							maxLength={maxNote}
							rows={2}
							value={note}
							onInput={(e) => setNote(e.currentTarget.value)}
						/>
						<span class="c-field__help">{(asking === "withdraw" ? t.reasonHelp : t.noteHelp)(maxNote - note.length)}</span>
					</label>
					<div class="actions">
						{asking === "withdraw" ? (
							<button class="c-btn c-btn--danger c-btn--sm" type="submit" disabled={busy || note.trim() === ""}>
								{t.confirmWithdraw}
							</button>
						) : (
							<button class="c-btn c-btn--primary c-btn--sm" type="submit" disabled={busy}>
								{t.confirmDismiss}
							</button>
						)}
						<button class="c-btn c-btn--ghost c-btn--sm" type="button" onClick={() => (setAsking(null), setNote(""), setError(""))}>
							{t.cancel}
						</button>
					</div>
				</form>
			)}
		</article>
	);
}

/**
 * A public Window, as the moderators look at it without a report: at its
 * address, as anyone does, with Withdraw.
 */
function PublicCard(props: { window: PublicWindow; onWithdrawn: () => void }) {
	const { call, origin, onDecided } = useModerator();
	const w = props.window;
	const [asking, setAsking] = useState(false);
	const [reason, setReason] = useState("");
	const [busy, setBusy] = useState(false);
	const [error, setError] = useState("");
	const withdraw = async (e: Event) => {
		e.preventDefault();
		setBusy(true);
		setError("");
		try {
			await call("withdraw_window", { window: w.id, reason });
			onDecided();
			props.onWithdrawn();
		} catch (err) {
			setError(explain(err));
		}
		setBusy(false);
	};
	return (
		<article class="c-card panel report">
			<p class="report__window">
				<strong>{w.title}</strong> <span class="c-mono c-muted">{`${w.room}/${w.name}`}</span> <AccessBadge access={w.access} />
			</p>
			<p class="c-small c-muted">{t.publicSince(w.by ?? "", when(w.public_at))}</p>
			{error && <Failure message={error} />}
			{!asking ? (
				<div class="actions">
					<a class="c-btn c-btn--secondary c-btn--sm" href={`${origin}/${w.room}/${w.name}/`} target="_blank" rel="noopener noreferrer">
						<ArrowUpRight class="i" size={14} aria-hidden="true" />
						{t.openAtAddress}
					</a>
					<button class="c-btn c-btn--secondary c-btn--sm" type="button" onClick={() => setAsking(true)}>
						{t.withdraw}
					</button>
				</div>
			) : (
				<form class="report__decide" onSubmit={withdraw}>
					<label class="c-field">
						<span class="c-field__label">{t.reason}</span>
						<textarea class="c-input" required maxLength={maxNote} rows={2} value={reason} onInput={(e) => setReason(e.currentTarget.value)} />
						<span class="c-field__help">{t.reasonHelp(maxNote - reason.length)}</span>
					</label>
					<div class="actions">
						<button class="c-btn c-btn--danger c-btn--sm" type="submit" disabled={busy || reason.trim() === ""}>
							{t.confirmWithdraw}
						</button>
						<button class="c-btn c-btn--ghost c-btn--sm" type="button" onClick={() => (setAsking(false), setReason(""), setError(""))}>
							{t.cancel}
						</button>
					</div>
				</form>
			)}
		</article>
	);
}

/** The public Windows of this Cahut, which its moderators look at without a report. */
function PublicWindows() {
	const { call } = useModerator();
	const { data, error, reload } = useLoad(() => call("list_public_windows", {}));
	if (error) return <Failure message={error} />;
	if (!data) return null;
	return (
		<section class="section">
			<h2 class="c-h4">{t.publicTitle}</h2>
			<p class="c-small c-muted">{t.publicLead}</p>
			{data.windows.length === 0 && <p class="c-muted">{t.noPublic}</p>}
			<div class="reports">
				{data.windows.map((w) => (
					<PublicCard key={w.id} window={w} onWithdrawn={reload} />
				))}
			</div>
		</section>
	);
}

/**
 * The page of moderation in the console of the Cahut, for its
 * administrators. Those of the server moderate from the console of the
 * server. See docs/adr/0058-a-console-for-the-server.md.
 */
export function Moderation() {
	const api = useConsole();
	// Whether the person administers the Cahut comes after the first render.
	if (!api.isAdmin) return <Notice>{t.onlyModerators}</Notice>;
	return <ModerationOf call={api.cahut} origin={location.origin} host={false} onDecided={api.reloadReports} version={api.reports} head />;
}

/**
 * The reports of a Cahut, those that wait, then the latest decided, and its
 * public Windows, for its moderators: with the head of a page, or as a
 * section of the page of a Cahut in the console of the server.
 */
export function ModerationOf(props: Moderator & { version?: number; head?: boolean }) {
	const { call, origin, host, onDecided } = props;
	const { data, error, reload } = useLoad(() => call("list_reports", {}), [props.version]);
	if (error) return <Failure message={error} />;
	if (!data) return null;
	const waiting = data.reports.filter((p) => !p.decision);
	const decided = data.reports.filter((p) => p.decision);
	return (
		<moderator.Provider value={{ call, origin, host, onDecided }}>
			{props.head ? (
				<PageHead label={t.label} title={t.title}>
					{t.lead}
				</PageHead>
			) : (
				<h2 class="c-h3 section">{t.title}</h2>
			)}
			<section class="section">
				<h2 class="c-h4">{t.waiting}</h2>
				{waiting.length === 0 ? <p class="c-muted">{t.nothingWaits}</p> : <p class="c-small c-muted">{t.waitingCount(data.waiting)}</p>}
				<div class="reports">
					{waiting.map((p) => (
						<ReportCard key={p.id} report={p} onDecided={reload} />
					))}
				</div>
			</section>
			<PublicWindows key={props.version} />
			{decided.length > 0 && (
				<section class="section">
					<h2 class="c-h4">{t.decided}</h2>
					<div class="reports">
						{decided.map((p) => (
							<ReportCard key={p.id} report={p} onDecided={reload} />
						))}
					</div>
				</section>
			)}
		</moderator.Provider>
	);
}
