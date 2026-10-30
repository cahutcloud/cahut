// The page of a Window: its title and its description, where it opens, at
// its address and at its permanent address, who opens it and until when, its two versions, what its AI
// declared it does, the menu that shares it, its share links, and the trash. See
// docs/adr/0019-window-versions-access-and-declaration.md,
// docs/adr/0050-window-ids-permanent-addresses-and-names.md,
// docs/adr/0051-sharing-a-window-and-its-preview.md and
// docs/adr/0052-share-links.md.
import { ArrowUpRight, Link2, RotateCcw, Send, Undo2 } from "lucide-preact";
import type { ComponentChildren } from "preact";
import { useId, useState } from "preact/hooks";
import { explain } from "../api";
import type { AccessSet, Declaration, WindowDetail } from "../api.gen";
import {
	AccessBadge,
	accesses,
	ago,
	agoInText,
	bytes,
	ConfirmButton,
	CopyField,
	DescriptionField,
	Failure,
	kindName,
	Notice,
	PageHead,
	plural,
	TagsField,
	tagsOf,
} from "../components";
import { useConsole } from "../context";
import { de, lang, locale, words } from "../i18n";
import { useLoad } from "../load";
import { NoteEditor, noteLength } from "../note";
import { Link, navigate } from "../router";
import { ShareLinkDialog, ShareMenu } from "../share";

/** The most characters of a note, as the server counts them. */
const maxNote = 500;

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
		text: "Title, description and tags",
		textHelp:
			"They change without a new version, and the address stays. Readers see the description on the page of the Room. The tags help you and your AIs find the Window among many, and only the people of the Room see them.",
		titleLabel: "Title",
		descriptionLabel: "Description",
		descriptionHelp: "A sentence or two that say what the Window is.",
		tagsLabel: "Tags",
		tagsHelp: "Press Enter or type a comma after each. Cahut writes them in lower case, with hyphens between their words.",
		saveText: "Save",
		textSaved: "Saved.",
		trash: "Trash",
		trashHelp: "A Window in the trash no longer opens, and keeps its versions and its access until you restore it.",
		putInTrash: "Put in the trash",
		putInTrashQuestion: "It will no longer open.",
		access: "Access",
		canOpen: (who: string, access: string) => (access === "author" ? "Only you can open it" : `${capitalize(who)} can open it`),
		until: (day: string) => `, until ${day}. Then only its Room can.`,
		shareTitle: "Share the address and the Key",
		shareHelp: "Copy the Key now: Cahut keeps only a trace of it, and shows it this once.",
		theAddress: "the address",
		theKey: "the Key",
		addresses: "Addresses",
		addressLabel: "Its address",
		permanentLabel: "Its permanent address",
		thePermanentAddress: "the permanent address",
		permanentHelp: "The permanent address opens this Window to whoever may open it, even once its name has changed. Give it for a link that lasts.",
		nameLabel: "Name in the address",
		nameHelp: "1 to 63 lowercase letters, digits and hyphens. The title stays.",
		changeName: "Change the name",
		nameQuestion: (old: string) =>
			`${old} will stop working at once, and its name will be free for another Window. The permanent address will keep working. Change the name?`,
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
		authorHelp:
			"Nobody else in its Room sees it any more, nor their AIs, until you open it to the Room again. Its share links still open it, to whoever holds them.",
		publicQuestion: "Anyone with its address will open it, and may keep a copy. Make it public?",
		makePublic: "Make it public",
		cancel: "Cancel",
		saveAccess: "Save the access",
		versions: "Versions",
		versionsHelp: "Cahut keeps the current version and the one before. Going back can be undone: the version you leave becomes the previous one.",
		version: (n: number) => `Version ${n}`,
		current: "Current",
		// by is empty once the person it was published for is deleted.
		versionMeta: (ago: string, by: string, files: string, size: string) => `${ago}${by ? ` · for ${by}` : ""} · ${files}, ${size}`,
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
		sentTo: (to: string) => `Sent to ${to}.`,
		unreachable: (who: string[]) => `${who.join(", ")} ${who.length === 1 ? "has" : "have"} no e-mail address: nothing went to them.`,
		how: {
			author: "This Window opens to you alone: nobody is invited to read it. To give it to one reader, make a share link that opens once, from the menu Share.",
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
		noteHelp: "Optional. The message shows it as yours, with its bold, italics and lists.",
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
		linksToo: (n: number) =>
			n === 1 ? "A share link opens it too, to whoever holds it." : `${n} share links open it too, to whoever holds them.`,
		links: "Share links",
		linksHelp:
			"Each link opens this Window to whoever holds it, without an account or a Key, until it ends, besides its access. Taken back, a link opens nothing more, to those who opened it too.",
		linksPublic: "This Window is public: anyone opens it at its permanent address. The links that stand open it too, until they end.",
		createLink: "Create a share link",
		linkLabel: "Label",
		noLabel: "No label",
		ends: "Until",
		openings: "Openings",
		openingsOf: (n: number, most?: number) => (most ? `${n} of ${most}` : `${n}`),
		stands: "Opens",
		linkEnded: "Ended",
		spent: "Used up",
		takenBack: "Taken back",
		takeBackLink: "Take back this link? It opens nothing more.",
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
		text: "Titre, description et tags",
		textHelp:
			"Ils changent sans nouvelle version, et l'adresse reste la même. Les lecteurs voient la description sur la page de la Room. Les tags vous aident, vous et vos IA, à retrouver la Window parmi d'autres, et seules les personnes de la Room les voient.",
		titleLabel: "Titre",
		descriptionLabel: "Description",
		descriptionHelp: "Une phrase ou deux qui disent ce qu'est la Window.",
		tagsLabel: "Tags",
		tagsHelp: "Appuyez sur Entrée ou tapez une virgule après chacun. Cahut les écrit en minuscules, avec des tirets entre leurs mots.",
		saveText: "Enregistrer",
		textSaved: "Enregistré.",
		trash: "Corbeille",
		trashHelp: "Une Window à la corbeille ne s'ouvre plus, et garde ses versions et son accès jusqu'à ce que vous la restauriez.",
		putInTrash: "Mettre à la corbeille",
		putInTrashQuestion: "Elle ne s'ouvrira plus.",
		access: "Accès",
		canOpen: (who, access) =>
			access === "author" ? "Personne d'autre que vous ne peut l'ouvrir" : `${capitalize(who)} ${access === "room" ? "peuvent" : "peut"} l'ouvrir`,
		until: (day) => `, jusqu'au ${day}. Ensuite, seuls les membres de sa Room le peuvent.`,
		shareTitle: "Partagez l'adresse et la Key",
		shareHelp: "Copiez la Key maintenant : Cahut n'en garde qu'une empreinte, et ne la montre que cette fois.",
		theAddress: "l'adresse",
		theKey: "la Key",
		addresses: "Adresses",
		addressLabel: "Son adresse",
		permanentLabel: "Son adresse permanente",
		thePermanentAddress: "l'adresse permanente",
		permanentHelp:
			"L'adresse permanente ouvre cette Window à qui peut l'ouvrir, même une fois son nom changé. Donnez-la pour un lien qui dure.",
		nameLabel: "Nom dans l'adresse",
		nameHelp: "De 1 à 63 lettres minuscules, chiffres et tirets. Le titre ne change pas.",
		changeName: "Changer le nom",
		nameQuestion: (old) =>
			`${old} cessera aussitôt de fonctionner, et son nom pourra servir à une autre Window. L'adresse permanente continuera de fonctionner. Changer le nom ?`,
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
		authorHelp:
			"Personne d'autre dans sa Room ne la voit plus, ni ses IA, jusqu'à ce que vous la rouvriez à la Room. Ses liens de partage l'ouvrent toujours, à qui les détient.",
		publicQuestion: "Quiconque a son adresse l'ouvrira, et pourra en garder une copie. La rendre publique ?",
		makePublic: "La rendre publique",
		cancel: "Annuler",
		saveAccess: "Enregistrer l'accès",
		versions: "Versions",
		versionsHelp:
			"Cahut garde la version actuelle et celle d'avant. Revenir en arrière s'annule : la version que vous quittez devient la précédente.",
		version: (n) => `Version ${n}`,
		current: "Actuelle",
		versionMeta: (ago, by, files, size) => `${ago}${by ? ` · pour ${by}` : ""} · ${files}, ${size}`,
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
		sentTo: (to) => `Message envoyé à ${to}.`,
		unreachable: (who) =>
			who.length === 1
				? `${who[0]} n'a pas d'adresse e-mail : rien ne lui a été envoyé.`
				: `${who.join(", ")} n'ont pas d'adresse e-mail : rien ne leur a été envoyé.`,
		how: {
			author:
				"Cette Window ne s'ouvre qu'à vous : personne n'est invité à la lire. Pour la donner à une personne, faites un lien de partage qui s'ouvre une fois, depuis le menu Partager.",
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
		noteHelp: "Facultative. Le message l'affiche comme venant de vous, avec son gras, son italique et ses listes.",
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
		linksToo: (n) =>
			n === 1 ? "Un lien de partage l'ouvre aussi, à qui le détient." : `${n} liens de partage l'ouvrent aussi, à qui les détient.`,
		links: "Liens de partage",
		linksHelp:
			"Chaque lien ouvre cette Window à qui le détient, sans compte ni Key, jusqu'à sa fin, en plus de son accès. Retiré, un lien n'ouvre plus rien, même à ceux qui l'ont ouvert.",
		linksPublic:
			"Cette Window est publique : n'importe qui l'ouvre à son adresse permanente. Les liens encore actifs l'ouvrent aussi, jusqu'à leur fin.",
		createLink: "Créer un lien de partage",
		linkLabel: "Étiquette",
		noLabel: "Sans étiquette",
		ends: "Jusqu'au",
		openings: "Ouvertures",
		openingsOf: (n, most) => (most ? `${n} sur ${most}` : `${n}`),
		stands: "Actif",
		linkEnded: "Terminé",
		spent: "Épuisé",
		takenBack: "Retiré",
		takeBackLink: "Retirer ce lien ? Il n'ouvrira plus rien.",
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
				lead={w.description}
				actions={
					!w.trashed_at && (
						<>
							<ShareMenu window={w} onLinks={reload} />
							<a class="c-btn c-btn--secondary" href={w.url} target="_blank" rel="noopener noreferrer">
								<ArrowUpRight class="i" size={16} aria-hidden="true" />
								{t.open}
							</a>
						</>
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
						<TextPanel
							window={w}
							onSave={async (input) => {
								await api.cahut("retitle_window", { window: w.id, ...input });
								api.reloadRooms();
								reload();
							}}
						/>
						<AccessPanel
							// Its form starts again from the access, when a link that opens once changes it.
							key={w.access}
							window={w}
							shared={shared}
							onSet={async (input) => {
								setShared(null);
								const out = await api.cahut("set_window_access", input);
								if (out.key) setShared(out);
								reload();
							}}
						/>
						<ShareLinks window={w} onChange={reload} />
						<Versions window={w} onRollback={() => act(() => api.cahut("rollback_window", { window: `${w.room}/${w.name}` }))} />
						<Invitations window={w} />
					</section>
					<aside class="aside">
						<Addresses
							window={w}
							onRename={async (name) => {
								const out = await api.cahut("rename_window", { window: w.id, name });
								api.reloadRooms();
								navigate(`/rooms/${out.room}/${out.name}`);
							}}
						/>
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

/**
 * The title, the description and the tags of a Window, which change in
 * place, without a new version. The field of its tags suggests those of the
 * other Windows of the person.
 */
function TextPanel(props: { window: WindowDetail; onSave: (input: { title?: string; description?: string; tags?: string[] }) => Promise<void> }) {
	const w = props.window;
	const api = useConsole();
	const id = useId();
	const [title, setTitle] = useState(w.title);
	const [description, setDescription] = useState(w.description ?? "");
	const [tags, setTags] = useState(w.tags);
	const [busy, setBusy] = useState(false);
	const [said, setSaid] = useState("");
	const [error, setError] = useState("");
	const others = useLoad(() => api.cahut("list_windows", {}).catch(() => null));
	const suggestions = tagsOf(others.data?.windows ?? []).map(([tag]) => tag);
	const newTitle = title.trim() !== w.title;
	const newDescription = description.trim() !== (w.description ?? "");
	const newTags = tags.join(" ") !== w.tags.join(" ");
	const save = async (e: Event) => {
		e.preventDefault();
		setBusy(true);
		setError("");
		setSaid("");
		try {
			await props.onSave({ ...(newTitle && { title }), ...(newDescription && { description }), ...(newTags && { tags }) });
			setSaid(t.textSaved);
		} catch (err) {
			setError(explain(err));
		} finally {
			setBusy(false);
		}
	};
	return (
		<div class="c-card panel aside">
			<h2 class="c-h4">{t.text}</h2>
			<p class="c-small c-muted">{t.textHelp}</p>
			<form class="form" onSubmit={save}>
				<label class="c-field">
					<span class="c-field__label">{t.titleLabel}</span>
					<input class="c-input" required value={title} onInput={(e) => (setTitle(e.currentTarget.value), setSaid(""))} />
				</label>
				<div class="c-field">
					<label class="c-field__label" for={id}>
						{t.descriptionLabel}
					</label>
					<DescriptionField id={id} value={description} help={t.descriptionHelp} onInput={(v) => (setDescription(v), setSaid(""))} />
				</div>
				<div class="c-field">
					<label class="c-field__label" for={`${id}-tags`}>
						{t.tagsLabel}
					</label>
					<TagsField
						id={`${id}-tags`}
						value={tags}
						suggestions={suggestions}
						help={t.tagsHelp}
						onChange={(v) => (setTags([...v].sort()), setSaid(""))}
					/>
				</div>
				{error && <Failure message={error} />}
				{said && (
					<div class="c-alert c-alert--success" role="status">
						{said}
					</div>
				)}
				<div>
					<button class="c-btn c-btn--secondary" type="submit" disabled={busy || (!newTitle && !newDescription && !newTags)}>
						{t.saveText}
					</button>
				</div>
			</form>
		</div>
	);
}

type AccessInput = { window: string; access: string; ends_at?: string; key?: string; new_key?: boolean };

/** The access of a Window, and the form that changes it. */
function AccessPanel(props: { window: WindowDetail; shared: AccessSet | null; onSet: (input: AccessInput) => Promise<void> }) {
	const { individual, session } = useConsole();
	const w = props.window;
	// Only its author keeps a Window for themself.
	const author = w.created_by === session.handle;
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
		if (access !== "room" && access !== "author" && ends) {
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
			{w.share_links > 0 && <p class="c-small c-muted">{t.linksToo(w.share_links)}</p>}
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
						.filter(([value]) => value !== "author" || author || w.access === "author")
						.map(([value, a]) => (
						<label class="c-check" key={value}>
							<input type="radio" name="access" checked={access === value} onChange={() => setAccess(value)} />
							<span>{t.option(a.name, a.who)}</span>
						</label>
					))}
					{access === "author" && w.access !== "author" && <span class="c-field__help">{t.authorHelp}</span>}
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
				{access !== "room" && access !== "author" && (
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

/**
 * The addresses of a Window: the one of the moment, which its name makes,
 * and the permanent one, which its id makes and which never changes; and its
 * name, which changes after a warning: the old address stops working at
 * once. See docs/adr/0050-window-ids-permanent-addresses-and-names.md.
 */
function Addresses({ window: w, onRename }: { window: WindowDetail; onRename: (name: string) => Promise<void> }) {
	const id = useId();
	const [name, setName] = useState(w.name);
	const [asking, setAsking] = useState(false);
	const [busy, setBusy] = useState(false);
	const [error, setError] = useState("");
	const prefix = w.url.slice(0, w.url.length - w.name.length - 1);
	const rename = async () => {
		setBusy(true);
		setError("");
		try {
			await onRename(name.trim());
		} catch (err) {
			setError(explain(err));
			setAsking(false);
		} finally {
			setBusy(false);
		}
	};
	return (
		<div class="c-card panel aside">
			<h2 class="c-h4">{t.addresses}</h2>
			<div class="c-field">
				<span class="c-field__label">{t.addressLabel}</span>
				<CopyField label={t.theAddress} value={w.url} />
			</div>
			<div class="c-field">
				<span class="c-field__label">{t.permanentLabel}</span>
				<CopyField label={t.thePermanentAddress} value={w.permanent_url} />
			</div>
			<p class="c-small c-muted">{t.permanentHelp}</p>
			<form
				class="form"
				onSubmit={(e) => {
					e.preventDefault();
					setAsking(true);
				}}
			>
				<div class="c-field">
					<label class="c-field__label" for={id}>
						{t.nameLabel}
					</label>
					<span class="window-name">
						<span class="window-name__prefix c-mono">{prefix.replace(/^https?:\/\//, "")}</span>
						<input
							id={id}
							class="c-input c-mono"
							required
							maxLength={63}
							pattern="[a-z0-9]([a-z0-9\-]*[a-z0-9])?"
							autocapitalize="none"
							spellcheck={false}
							value={name}
							onInput={(e) => (setName(e.currentTarget.value), setAsking(false), setError(""))}
						/>
					</span>
					<span class="c-field__help">{t.nameHelp}</span>
				</div>
				{error && <Failure message={error} />}
				{asking ? (
					<span class="confirm" role="group">
						<span class="c-small">{t.nameQuestion(w.url)}</span>
						<button class="c-btn c-btn--primary c-btn--sm" type="button" disabled={busy} onClick={rename}>
							{t.changeName}
						</button>
						<button class="c-btn c-btn--secondary c-btn--sm" type="button" onClick={() => setAsking(false)}>
							{t.cancel}
						</button>
					</span>
				) : (
					<div>
						<button class="c-btn c-btn--secondary" type="submit" disabled={name.trim() === w.name}>
							{t.changeName}
						</button>
					</div>
				)}
			</form>
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
	const [sent, setSent] = useState(0);
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
			setSent((n) => n + 1);
			reload();
		} catch (err) {
			setError(explain(err));
		}
		setBusy(false);
	};
	const how = t.how;
	const tooLong = noteLength(note) > maxNote;
	if (!api.sendsMail && !data?.invitations.length) return null;
	return (
		<div class="c-card panel section">
			<h2 class="c-h4">{api.sendsMail ? t.invitePeople : t.invitations}</h2>
			<p class="c-small c-muted">
				{api.sendsMail
					? (how[w.access] ?? how.room)
					: t.noMail}
			</p>
			{api.sendsMail && w.access !== "author" && (
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
					<NoteEditor
						key={sent}
						label={t.note}
						max={maxNote}
						help={t.noteHelp}
						onChange={setNote}
					/>
					{error && <Failure message={error} />}
					{said && (
						<div class="c-alert c-alert--success" role="status">
							{said}
						</div>
					)}
					<div>
						<button class="c-btn c-btn--primary" type="submit" disabled={busy || tooLong}>
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

/**
 * The share links of a Window: the links that stand, then those that ended,
 * with their label, who made them, their end and their openings, and a
 * button that takes back each link that still opens the Window to someone.
 * See docs/adr/0052-share-links.md.
 */
function ShareLinks({ window: w, onChange }: { window: WindowDetail; onChange: () => void }) {
	const api = useConsole();
	const [making, setMaking] = useState(false);
	if (w.access === "public" && w.links.length === 0) return null;
	const now = new Date();
	const day = (at: string) => new Date(at).toLocaleDateString(lang === "en" ? "en-US" : locale, { month: "short", day: "numeric", year: "numeric" });
	return (
		<div class="c-card panel section">
			<h2 class="c-h4">{t.links}</h2>
			<p class="c-small c-muted">{w.access === "public" ? t.linksPublic : t.linksHelp}</p>
			{w.access !== "public" && (
				<div>
					<button class="c-btn c-btn--secondary" type="button" onClick={() => setMaking(true)}>
						<Link2 class="i" size={16} aria-hidden="true" />
						{t.createLink}
					</button>
				</div>
			)}
			{making && <ShareLinkDialog window={w} onMade={onChange} onClose={() => setMaking(false)} />}
			{w.links.length > 0 && (
				<div class="table-wrap">
					<table class="table">
						<thead>
							<tr>
								<th>{t.linkLabel}</th>
								<th>{t.ends}</th>
								<th>{t.openings}</th>
								<th>{t.state}</th>
								<th />
							</tr>
						</thead>
						<tbody>
							{w.links.map((l) => {
								const ended = new Date(l.expires_at) <= now;
								// A link used up still opens the Window to those who opened it.
								const lasts = !l.revoked_at && !ended;
								return (
									<tr key={l.id}>
										<td class="c-small">
											<div>{l.label || <span class="c-muted">{t.noLabel}</span>}</div>
											<div class="c-muted">
												{ago(l.created_at)}
												{l.by && t.by(l.by)}
												{l.ai && t.through(l.ai)}
											</div>
										</td>
										<td class="c-small">{day(l.expires_at)}</td>
										<td class="c-small">{t.openingsOf(l.openings, l.max_openings)}</td>
										<td class="c-small">{l.revoked_at ? t.takenBack : ended ? t.linkEnded : l.stands ? t.stands : t.spent}</td>
										<td class="table__actions">
											{lasts && (
												<ConfirmButton
													label={t.takeBack}
													question={t.takeBackLink}
													confirm={t.takeBack}
													danger
													onConfirm={async () => {
														await api.cahut("revoke_share_link", { link: l.id });
														onChange();
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
