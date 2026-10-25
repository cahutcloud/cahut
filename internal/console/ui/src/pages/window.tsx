// The page of a Window: where it opens, who opens it and until when, its
// two versions, what its AI declared it does, and the trash. See
// docs/adr/0019-window-versions-access-and-declaration.md.
import { ArrowUpRight, RotateCcw, Send, Trash2, Undo2 } from "lucide-preact";
import type { ComponentChildren } from "preact";
import { useState } from "preact/hooks";
import { explain } from "../api";
import type { AccessSet, Declaration, WindowDetail } from "../api.gen";
import { AccessBadge, accesses, ago, agoInText, bytes, ConfirmButton, CopyField, Failure, kindName, Notice, PageHead, plural } from "../components";
import { useConsole } from "../context";
import { de, lang, locale, words } from "../i18n";
import { useLoad } from "../load";
import { Link, navigate } from "../router";

const t = words({
	en: {
		open: "Open",
		head: (room: ComponentChildren, version: number, updated: string) => (
			<>
				In {room} · version {version} · updated {updated} ·{" "}
			</>
		),
		inTrash: "This Window is in the trash: it does not open, and it is in no list. Restore it to open it again, as it was, or delete it for good.",
		restore: "Restore",
		deleteForGood: "Delete for good",
		deleteQuestion: "Delete it with its versions and its files? This cannot be undone.",
		trash: "Trash",
		trashHelp: "A Window in the trash no longer opens, and keeps its versions and its access until you restore it.",
		putInTrash: "Put in the trash",
		putInTrashQuestion: "It will no longer open.",
		access: "Access",
		canOpen: (who: string, _access: string) => `${capitalize(who)} can open it`,
		until: (day: string) => `, until ${day}. Then only its Room can.`,
		shareTitle: "Share the address and the Key",
		shareHelp: "Copy the Key now: Cahut keeps only a trace of it, and shows it this once.",
		theAddress: "the address",
		theKey: "the Key",
		whoOpens: "Who opens it",
		option: (name: string, who: string) => `${name}: ${who}`,
		key: "Key",
		keepKey: "Keep its Key",
		newKey: "A new Key, drawn by Cahut: the old one stops working",
		drawnKey: "A Key drawn by Cahut",
		chosenKey: "A Key I choose",
		keyLabel: "The Key",
		keyHelp: "10 characters at least. Readers type it to open the Window.",
		untilLabel: "Until",
		untilHelp: "Leave empty for no end. After this day, only its Room opens the Window.",
		publicQuestion: "Anyone with its address will open it, and may keep a copy. Make it public?",
		makePublic: "Make it public",
		cancel: "Cancel",
		saveAccess: "Save the access",
		versions: "Versions",
		versionsHelp: "Cahut keeps the current version and the one before. Going back can be undone: the version you leave becomes the previous one.",
		version: (n: number) => `Version ${n}`,
		current: "Current",
		versionMeta: (ago: string, by: string, files: string, size: string) => `${ago} · for ${by} · ${files}, ${size}`,
		file: "file",
		files: "files",
		goBack: "Go back to it",
		whatItDoes: "What it does",
		notDeclared:
			"Not declared: this version was published before declarations, and may load and call any address. Its next version will declare what it does.",
		declared: "As its AI declared it. The pages can load and call nothing else.",
		scripts: "Scripts",
		yes: "Yes",
		no: "No",
		loads: "Loads and calls",
		nothingOutside: "Nothing outside this Cahut",
		forms: "Forms send to",
		nowhere: "Nowhere",
		trashedSince: (ago: string) => `In the trash since ${ago}`,
		sentTo: (to: string) => `Sent to ${to}.`,
		unreachable: (who: string[]) => `${who.join(", ")} ${who.length === 1 ? "has" : "have"} no e-mail address: nothing went to them.`,
		how: {
			key: "People of the Cahut, by handle, and anyone else, by address. Each address gets a personal link that opens it for them alone, until you take it back; the Key never goes.",
			public: "Anyone, by handle or by address: the message gives its address.",
			internal: "People of the Cahut, by handle: they sign in to read it. To invite someone else by address, set its access to key first.",
			room: "The members of its Room, by handle. To invite someone else, set its access to internal or key first.",
		} as Record<string, string>,
		invitePeople: "Invite people",
		invitations: "Invitations",
		noMail: "This Cahut no longer sends e-mail. The personal links that went out still open this Window, until you take them back.",
		to: "To",
		toHelp: "Handles or e-mail addresses, 20 at most, separated by commas.",
		note: "Note",
		noteHelp: "Optional, in plain text. The message shows it as yours.",
		send: "Send the invitation",
		sentAt: "Sent",
		sent: "Sent",
		state: "State",
		by: (who: string) => `, by ${who}`,
		through: (ai: string) => ` through ${ai}`,
		ended: "Ended",
		expired: "Expired",
		personalLink: "Personal link",
		takeBack: "Take back",
		takeBackQuestion: (to: string) => `Take back the link of ${to}? It no longer opens the Window.`,
	},
	fr: {
		open: "Ouvrir",
		head: (room, version, updated) => (
			<>
				Dans {room} · version {version} · mise à jour {updated} ·{" "}
			</>
		),
		inTrash:
			"Cette Window est à la corbeille : elle ne s'ouvre pas, et ne figure dans aucune liste. Restaurez-la pour l'ouvrir de nouveau, telle qu'elle était, ou supprimez-la définitivement.",
		restore: "Restaurer",
		deleteForGood: "Supprimer définitivement",
		deleteQuestion: "La supprimer avec ses versions et ses fichiers ? Cela ne s'annule pas.",
		trash: "Corbeille",
		trashHelp: "Une Window à la corbeille ne s'ouvre plus, et garde ses versions et son accès jusqu'à ce que vous la restauriez.",
		putInTrash: "Mettre à la corbeille",
		putInTrashQuestion: "Elle ne s'ouvrira plus.",
		access: "Accès",
		canOpen: (who, access) => `${capitalize(who)} ${access === "room" ? "peuvent" : "peut"} l'ouvrir`,
		until: (day) => `, jusqu'au ${day}. Ensuite, seuls les membres de sa Room le peuvent.`,
		shareTitle: "Partagez l'adresse et la Key",
		shareHelp: "Copiez la Key maintenant : Cahut n'en garde qu'une empreinte, et ne la montre que cette fois.",
		theAddress: "l'adresse",
		theKey: "la Key",
		whoOpens: "Qui l'ouvre",
		option: (name, who) => `${name} : ${who}`,
		key: "Key",
		keepKey: "Garder sa Key",
		newKey: "Une nouvelle Key, tirée au sort par Cahut : l'ancienne ne fonctionne plus",
		drawnKey: "Une Key tirée au sort par Cahut",
		chosenKey: "Une Key que je choisis",
		keyLabel: "La Key",
		keyHelp: "10 caractères au moins. Les lecteurs la tapent pour ouvrir la Window.",
		untilLabel: "Jusqu'au",
		untilHelp: "Laissez vide pour ne fixer aucune fin. Après ce jour, seule sa Room ouvre la Window.",
		publicQuestion: "Quiconque a son adresse l'ouvrira, et pourra en garder une copie. La rendre publique ?",
		makePublic: "La rendre publique",
		cancel: "Annuler",
		saveAccess: "Enregistrer l'accès",
		versions: "Versions",
		versionsHelp:
			"Cahut garde la version actuelle et celle d'avant. Revenir en arrière s'annule : la version que vous quittez devient la précédente.",
		version: (n) => `Version ${n}`,
		current: "Actuelle",
		versionMeta: (ago, by, files, size) => `${ago} · pour ${by} · ${files}, ${size}`,
		file: "fichier",
		files: "fichiers",
		goBack: "Y revenir",
		whatItDoes: "Ce qu'elle fait",
		notDeclared:
			"Pas de déclaration : cette version a été publiée avant les déclarations, et peut charger et appeler n'importe quelle adresse. Sa prochaine version déclarera ce qu'elle fait.",
		declared: "Telle que son IA l'a déclarée. Ses pages ne peuvent rien charger ni appeler d'autre.",
		scripts: "Scripts",
		yes: "Oui",
		no: "Non",
		loads: "Charge et appelle",
		nothingOutside: "Rien hors de cette Cahut",
		forms: "Les formulaires envoient à",
		nowhere: "Nulle part",
		trashedSince: (ago) => `Mise à la corbeille ${ago}`,
		sentTo: (to) => `Message envoyé à ${to}.`,
		unreachable: (who) =>
			who.length === 1
				? `${who[0]} n'a pas d'adresse e-mail : rien ne lui a été envoyé.`
				: `${who.join(", ")} n'ont pas d'adresse e-mail : rien ne leur a été envoyé.`,
		how: {
			key: "Les personnes de la Cahut, par identifiant, et toute autre personne, par adresse. Chaque adresse reçoit un lien personnel qui ouvre la Window pour elle seule, jusqu'à ce que vous le retiriez ; la Key ne part jamais.",
			public: "N'importe qui, par identifiant ou par adresse : le message donne son adresse.",
			internal:
				"Les personnes de la Cahut, par identifiant : elles se connectent pour la lire. Pour inviter quelqu'un d'autre par adresse, passez d'abord son accès à Key.",
			room: "Les membres de sa Room, par identifiant. Pour inviter quelqu'un d'autre, passez d'abord son accès à Interne ou à Key.",
		},
		invitePeople: "Inviter des personnes",
		invitations: "Invitations",
		noMail: "Cette Cahut n'envoie plus d'e-mail. Les liens personnels déjà partis ouvrent toujours cette Window, jusqu'à ce que vous les retiriez.",
		to: "À",
		toHelp: "Des identifiants ou des adresses e-mail, 20 au plus, séparés par des virgules.",
		note: "Note",
		noteHelp: "Facultative, en texte brut. Le message l'affiche comme venant de vous.",
		send: "Envoyer l'invitation",
		sentAt: "Envoi",
		sent: "Envoyée",
		state: "État",
		by: (who) => `, par ${who}`,
		through: (ai) => ` via ${ai}`,
		ended: "Terminée",
		expired: "Expirée",
		personalLink: "Lien personnel",
		takeBack: "Retirer",
		takeBackQuestion: (to) => `Retirer le lien ${de(to)} ? Il n'ouvrira plus la Window.`,
	},
});

export function WindowPage({ room, name }: { room: string; name: string }) {
	const api = useConsole();
	const { data, error, reload } = useLoad(() => api.cahut("get_window", { window: `${room}/${name}` }), [room, name]);
	const [failure, setFailure] = useState("");
	// The Key that the last change drew or set, shown this once.
	const [shared, setShared] = useState<AccessSet | null>(null);
	if (error) return <Failure message={error} />;
	if (!data) return null;
	const w = data;
	const act = async (run: () => Promise<unknown>) => {
		setFailure("");
		try {
			await run();
			api.reloadRooms();
			reload();
		} catch (err) {
			setFailure(explain(err));
		}
	};
	return (
		<>
			<PageHead
				label={kindName(w.kind)}
				title={w.title}
				actions={
					!w.trashed_at && (
						<a class="c-btn c-btn--secondary" href={w.url} target="_blank" rel="noopener noreferrer">
							<ArrowUpRight class="i" size={16} aria-hidden="true" />
							{t.open}
						</a>
					)
				}
			>
				{t.head(
					<Link class="c-link" href={`/rooms/${w.room}`}>
						{w.room}
					</Link>,
					w.version,
					agoInText(w.updated_at),
				)}
				<span class="c-mono">{w.url}</span>
			</PageHead>
			{failure && <Failure message={failure} />}
			{w.trashed_at ? (
				<section class="section">
					<Notice>{t.inTrash}</Notice>
					<div class="actions">
						<button class="c-btn c-btn--primary" type="button" onClick={() => act(() => api.cahut("restore_window", { window: `${w.room}/${w.name}` }))}>
							<Undo2 class="i" size={16} aria-hidden="true" />
							{t.restore}
						</button>
						<ConfirmButton
							label={t.deleteForGood}
							question={t.deleteQuestion}
							confirm={t.deleteForGood}
							danger
							onConfirm={() =>
								act(async () => {
									await api.cahut("purge_window", { window: `${w.room}/${w.name}` });
									navigate(`/rooms/${w.room}`);
								})
							}
						/>
					</div>
				</section>
			) : (
				<div class="layout">
					<section>
						<AccessPanel
							window={w}
							shared={shared}
							onSet={async (input) => {
								setShared(null);
								const out = await api.cahut("set_window_access", input);
								if (out.key) setShared(out);
								reload();
							}}
						/>
						<Versions window={w} onRollback={() => act(() => api.cahut("rollback_window", { window: `${w.room}/${w.name}` }))} />
						<Invitations window={w} />
					</section>
					<aside class="aside">
						<DeclarationPanel declaration={w.declaration} />
						<div class="c-card panel aside">
							<h2 class="c-h4">{t.trash}</h2>
							<p class="c-small c-muted">{t.trashHelp}</p>
							<div>
								<ConfirmButton
									label={t.putInTrash}
									question={t.putInTrashQuestion}
									confirm={t.putInTrash}
									danger
									onConfirm={() => act(() => api.cahut("trash_window", { window: `${w.room}/${w.name}` }))}
								/>
							</div>
						</div>
					</aside>
				</div>
			)}
		</>
	);
}

type AccessInput = { window: string; access: string; ends_at?: string; key?: string; new_key?: boolean };

/** The access of a Window, and the form that changes it. */
function AccessPanel(props: { window: WindowDetail; shared: AccessSet | null; onSet: (input: AccessInput) => Promise<void> }) {
	const { individual } = useConsole();
	const w = props.window;
	const [access, setAccess] = useState(w.access);
	const [ends, setEnds] = useState(w.access_ends_at ? w.access_ends_at.slice(0, 10) : "");
	const [keyChoice, setKeyChoice] = useState<"keep" | "draw" | "choose">(w.access === "key" ? "keep" : "draw");
	const [key, setKey] = useState("");
	const [asking, setAsking] = useState(false);
	const [error, setError] = useState("");
	const [busy, setBusy] = useState(false);

	const save = async () => {
		setError("");
		setBusy(true);
		const input: AccessInput = { window: `${w.room}/${w.name}`, access };
		if (access !== "room" && ends) {
			// The access ends at the end of the chosen day, where the person is.
			input.ends_at = new Date(`${ends}T23:59:59`).toISOString();
		}
		if (access === "key" && keyChoice === "draw" && w.access === "key") input.new_key = true;
		if (access === "key" && keyChoice === "choose") input.key = key;
		try {
			await props.onSet(input);
			setAsking(false);
			setKey("");
			setKeyChoice("keep");
		} catch (err) {
			setError(explain(err));
		} finally {
			setBusy(false);
		}
	};
	const submit = (e: Event) => {
		e.preventDefault();
		// Opening a Window to the public asks first: once read, it cannot be taken back.
		if (access === "public" && w.access !== "public") {
			setAsking(true);
			return;
		}
		save();
	};
	const today = new Date();
	const tomorrow = new Date(today.getFullYear(), today.getMonth(), today.getDate() + 1).toLocaleDateString("en-CA");

	return (
		<div class="c-card panel aside">
			<div class="access-head">
				<h2 class="c-h4">{t.access}</h2>
				<AccessBadge access={w.access} />
			</div>
			<p class="c-small c-muted">
				{t.canOpen((accesses[w.access] ?? accesses.room).who, w.access in accesses ? w.access : "room")}
				{w.access_ends_at
					? t.until(
							new Date(w.access_ends_at).toLocaleDateString(lang === "en" ? "en-US" : locale, { month: "short", day: "numeric", year: "numeric" }),
						)
					: "."}
			</p>
			{props.shared?.key && (
				<div class="c-alert c-alert--success share" role="status">
					<div>
						<strong>{t.shareTitle}</strong>
						{t.shareHelp}
						<CopyField label={t.theAddress} value={props.shared.url} />
						<CopyField label={t.theKey} value={props.shared.key} />
					</div>
				</div>
			)}
			<form class="form" onSubmit={submit}>
				<fieldset class="consent__group">
					<legend class="c-field__label">{t.whoOpens}</legend>
					{Object.entries(accesses)
						// An individual Cahut opens no Window to everyone signed in to it.
						.filter(([value]) => !individual || value !== "internal" || w.access === "internal")
						.map(([value, a]) => (
						<label class="c-check" key={value}>
							<input type="radio" name="access" checked={access === value} onChange={() => setAccess(value)} />
							<span>{t.option(a.name, a.who)}</span>
						</label>
					))}
				</fieldset>
				{access === "key" && (
					<fieldset class="consent__group">
						<legend class="c-field__label">{t.key}</legend>
						{w.access === "key" && (
							<label class="c-check">
								<input type="radio" name="key" checked={keyChoice === "keep"} onChange={() => setKeyChoice("keep")} />
								<span>{t.keepKey}</span>
							</label>
						)}
						<label class="c-check">
							<input type="radio" name="key" checked={keyChoice === "draw"} onChange={() => setKeyChoice("draw")} />
							<span>{w.access === "key" ? t.newKey : t.drawnKey}</span>
						</label>
						<label class="c-check">
							<input type="radio" name="key" checked={keyChoice === "choose"} onChange={() => setKeyChoice("choose")} />
							<span>{t.chosenKey}</span>
						</label>
						{keyChoice === "choose" && (
							<label class="c-field consent__rooms">
								<span class="c-field__label">{t.keyLabel}</span>
								<input class="c-input c-mono" required minLength={10} autocomplete="off" value={key} onInput={(e) => setKey(e.currentTarget.value)} />
								<span class="c-field__help">{t.keyHelp}</span>
							</label>
						)}
					</fieldset>
				)}
				{access !== "room" && (
					<label class="c-field">
						<span class="c-field__label">{t.untilLabel}</span>
						<input class="c-input" type="date" min={tomorrow} value={ends} onInput={(e) => setEnds(e.currentTarget.value)} />
						<span class="c-field__help">{t.untilHelp}</span>
					</label>
				)}
				{error && <Failure message={error} />}
				{asking ? (
					<span class="confirm" role="group">
						<span class="c-small">{t.publicQuestion}</span>
						<button class="c-btn c-btn--primary c-btn--sm" type="button" disabled={busy} onClick={save}>
							{t.makePublic}
						</button>
						<button class="c-btn c-btn--ghost c-btn--sm" type="button" onClick={() => setAsking(false)}>
							{t.cancel}
						</button>
					</span>
				) : (
					<div>
						<button class="c-btn c-btn--primary" type="submit" disabled={busy}>
							{t.saveAccess}
						</button>
					</div>
				)}
			</form>
		</div>
	);
}

function capitalize(s: string): string {
	return s.charAt(0).toUpperCase() + s.slice(1);
}

/** The current and the previous version of a Window. */
function Versions({ window: w, onRollback }: { window: WindowDetail; onRollback: () => void }) {
	return (
		<div class="c-card panel aside">
			<h2 class="c-h4">{t.versions}</h2>
			<p class="c-small c-muted">{t.versionsHelp}</p>
			<ul class="versions">
				{w.versions.map((v) => (
					<li key={v.number} class="versions__row">
						<span>
							<b>{t.version(v.number)}</b> {v.current && <span class="c-badge c-badge--success">{t.current}</span>}
							<span class="c-small c-muted versions__meta">
								{t.versionMeta(ago(v.published_at), v.published_by, plural(v.file_count, t.file, t.files), bytes(v.size))}
							</span>
						</span>
						{!v.current && (
							<button class="c-btn c-btn--secondary c-btn--sm" type="button" onClick={onRollback}>
								<RotateCcw class="i" size={14} aria-hidden="true" />
								{t.goBack}
							</button>
						)}
					</li>
				))}
			</ul>
		</div>
	);
}

/** What the AI declared that the pages of the current version do. */
function DeclarationPanel({ declaration: d }: { declaration?: Declaration }) {
	return (
		<div class="c-card panel aside">
			<h2 class="c-h4">{t.whatItDoes}</h2>
			{!d ? (
				<p class="c-small c-muted">{t.notDeclared}</p>
			) : (
				<>
					<p class="c-small c-muted">{t.declared}</p>
					<dl class="declaration">
						<dt>{t.scripts}</dt>
						<dd>{d.scripts ? t.yes : t.no}</dd>
						<dt>{t.loads}</dt>
						<dd>{d.origins?.length ? d.origins.map((o) => <div class="c-mono" key={o}>{o}</div>) : t.nothingOutside}</dd>
						<dt>{t.forms}</dt>
						<dd>{d.forms?.length ? d.forms.map((o) => <div class="c-mono" key={o}>{o}</div>) : t.nowhere}</dd>
					</dl>
				</>
			)}
		</div>
	);
}

/** The Windows of a Room in the trash, when it has some. */
export function Trash({ room }: { room: string }) {
	const api = useConsole();
	const { data } = useLoad(() => api.cahut("list_windows", { room, trashed: true }).catch(() => null), [room]);
	if (!data || data.windows.length === 0) return null;
	return (
		<section class="section">
			<h2 class="c-h4">
				<Trash2 class="i" size={16} aria-hidden="true" /> {t.trash}
			</h2>
			<ul class="versions">
				{data.windows.map((w) => (
					<li key={w.name} class="versions__row">
						<Link class="c-link" href={`/rooms/${w.room}/${w.name}`}>
							{w.title}
						</Link>
						<span class="c-small c-muted">{w.trashed_at ? t.trashedSince(agoInText(w.trashed_at)) : ""}</span>
					</li>
				))}
			</ul>
		</section>
	);
}

/**
 * The invitations to read a Window, sent by e-mail, and the form that sends
 * them: people of the Cahut by handle, anyone else by address, with a
 * personal link for a Window with a Key. A Cahut without a relay offers no
 * form, and shows only the invitations it sent before, to take them back.
 * See docs/adr/0031-e-mail-through-a-relay.md and
 * docs/adr/0034-e-mail-through-the-relay-of-each-cahut-only.md.
 */
function Invitations({ window: w }: { window: WindowDetail }) {
	const api = useConsole();
	const ref = `${w.room}/${w.name}`;
	const { data, reload } = useLoad(() => api.cahut("list_invitations", { window: ref }), [ref]);
	const [to, setTo] = useState("");
	const [note, setNote] = useState("");
	const [said, setSaid] = useState("");
	const [error, setError] = useState("");
	const [busy, setBusy] = useState(false);
	const send = async (e: Event) => {
		e.preventDefault();
		setBusy(true);
		setError("");
		setSaid("");
		try {
			const out = await api.cahut("invite_to_window", { window: ref, to: to.split(/[\s,;]+/).filter(Boolean), note });
			let msg = out.invitations.length > 0 ? t.sentTo(out.invitations.map((i) => i.to).join(", ")) : "";
			if (out.unreachable?.length) {
				msg += ` ${t.unreachable(out.unreachable)}`;
			}
			setSaid(msg.trim());
			setTo("");
			setNote("");
			reload();
		} catch (err) {
			setError(explain(err));
		}
		setBusy(false);
	};
	const how = t.how;
	if (!api.sendsMail && !data?.invitations.length) return null;
	return (
		<div class="c-card panel section">
			<h2 class="c-h4">{api.sendsMail ? t.invitePeople : t.invitations}</h2>
			<p class="c-small c-muted">
				{api.sendsMail
					? (how[w.access] ?? how.room)
					: t.noMail}
			</p>
			{api.sendsMail && (
				<form class="form" onSubmit={send}>
					<label class="c-field">
						<span class="c-field__label">{t.to}</span>
						<input
							class="c-input"
							required
							autocapitalize="none"
							placeholder="ana, joe@example.com"
							value={to}
							onInput={(e) => setTo(e.currentTarget.value)}
						/>
						<span class="c-field__help">{t.toHelp}</span>
					</label>
					<label class="c-field">
						<span class="c-field__label">{t.note}</span>
						<textarea class="c-input" rows={3} maxLength={500} value={note} onInput={(e) => setNote(e.currentTarget.value)} />
						<span class="c-field__help">{t.noteHelp}</span>
					</label>
					{error && <Failure message={error} />}
					{said && (
						<div class="c-alert c-alert--success" role="status">
							{said}
						</div>
					)}
					<div>
						<button class="c-btn c-btn--primary" type="submit" disabled={busy}>
							<Send class="i" size={16} aria-hidden="true" />
							{t.send}
						</button>
					</div>
				</form>
			)}
			{data && data.invitations.length > 0 && (
				<div class="table-wrap">
					<table class="table">
						<thead>
							<tr>
								<th>{t.to}</th>
								<th>{t.sentAt}</th>
								<th>{t.state}</th>
								<th />
							</tr>
						</thead>
						<tbody>
							{data.invitations.map((i) => {
								const open = !i.ended_at && (!i.expires_at || new Date(i.expires_at) > new Date());
								return (
									<tr key={i.id}>
										<td>
											<span class="c-mono c-small">{i.to}</span>
										</td>
										<td class="c-small">
											{ago(i.created_at)}
											{i.by && t.by(i.by)}
											{i.ai && t.through(i.ai)}
										</td>
										<td class="c-small">{i.ended_at ? t.ended : !open ? t.expired : i.link ? t.personalLink : t.sent}</td>
										<td class="table__actions">
											{open && i.link && (
												<ConfirmButton
													label={t.takeBack}
													question={t.takeBackQuestion(i.to)}
													confirm={t.takeBack}
													danger
													onConfirm={async () => {
														await api.cahut("revoke_invitation", { invitation: i.id });
														reload();
													}}
												/>
											)}
										</td>
									</tr>
								);
							})}
						</tbody>
					</table>
				</div>
			)}
		</div>
	);
}
