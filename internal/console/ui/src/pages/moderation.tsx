// Moderation: the reports that anyone makes about the Windows of this Cahut,
// from the page /-/report, for its moderators, its administrators and those
// of its server. Each report waits for a decision: the moderators leave its
// Window as it is, or withdraw it. See
// docs/adr/0054-reporting-and-withdrawing-a-window.md.
import { ArrowUpRight, Eye, Flag } from "lucide-preact";
import { useState } from "preact/hooks";
import { explain } from "../api";
import type { Report, ReportedPreview } from "../api.gen";
import { AccessBadge, ConfirmButton, Dialog, Failure, Notice, PageHead } from "../components";
import { useConsole } from "../context";
import { locale, plural, words } from "../i18n";
import { useLoad } from "../load";

const t = words({
	en: {
		label: "This Cahut",
		title: "Moderation",
		lead: "Anyone reports a Window of this Cahut, without an account, from the page Report a Window that the pages of this Cahut link. Its administrators and those of its server read the reports here, and decide.",
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
	},
	fr: {
		label: "Cette Cahut",
		title: "Modération",
		lead: "Tout le monde signale une Window de cette Cahut, sans compte, depuis la page Report a Window que ses pages relient. Ses administrateurs et ceux de son serveur lisent les signalements ici, et tranchent.",
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
 * Calls the API as the moderator: the person, when they administer this
 * Cahut, or else the administrator of the server whose session the browser
 * holds too. See docs/adr/0043-the-console-holds-two-sessions.md.
 */
export function useModerator() {
	const api = useConsole();
	if (api.session.kind !== "person" || api.isAdmin) {
		return { moderates: true, call: api.cahut };
	}
	return { moderates: Boolean(api.serverSession), call: api.cahutAsServer };
}

/** A report, as a card, with what the moderators can decide while it waits. */
function ReportCard(props: { report: Report; onDecided: () => void }) {
	const { call } = useModerator();
	const api = useConsole();
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
	const address = `${location.origin}/${w.room}/${w.name}/`;
	const run = async (act: () => Promise<unknown>) => {
		setBusy(true);
		setError("");
		try {
			await act();
			api.reloadReports();
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
					<ConfirmButton
						label={t.reinstate}
						question={t.reinstateQuestion}
						confirm={t.reinstate}
						onConfirm={() => run(() => call("reinstate_window", { window: w.id }))}
					/>
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
					<iframe class="report__frame" src={shown.frame} title={t.preview(w.title)} sandbox="allow-scripts" referrerpolicy="no-referrer" />
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

/** The reports of this Cahut: those that wait, then the latest decided. */
export function Moderation() {
	const { moderates, call } = useModerator();
	const api = useConsole();
	const { data, error, reload } = useLoad(() => (moderates ? call("list_reports", {}) : Promise.resolve(null)), [api.reports]);
	if (!moderates) return <Notice>{t.onlyModerators}</Notice>;
	if (error) return <Failure message={error} />;
	if (!data) return null;
	const waiting = data.reports.filter((p) => !p.decision);
	const decided = data.reports.filter((p) => p.decision);
	return (
		<>
			<PageHead label={t.label} title={t.title}>
				{t.lead}
			</PageHead>
			<section class="section">
				<h2 class="c-h4">{t.waiting}</h2>
				{waiting.length === 0 ? <p class="c-muted">{t.nothingWaits}</p> : <p class="c-small c-muted">{t.waitingCount(data.waiting)}</p>}
				<div class="reports">
					{waiting.map((p) => (
						<ReportCard key={p.id} report={p} onDecided={reload} />
					))}
				</div>
			</section>
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
		</>
	);
}
