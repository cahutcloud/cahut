// The Rooms: those of the person, every Room for an administrator, a new
// team Room, and the page of one Room: its head, with its address and what
// that address shows, its Windows and its trash, the people of a team Room,
// and a dialog for each change.
import { ArrowUpRight, Check, ChevronRight, Copy, House, Link2, Plus, Settings, Trash2, Undo2 } from "lucide-preact";
import type { ComponentChildren } from "preact";
import { useId, useState } from "preact/hooks";
import { ApiError, explain } from "../api";
import type { Person, Room, RoomDetail, Window } from "../api.gen";
import {
	agoInText,
	ConfirmButton,
	DescriptionField,
	Dialog,
	Failure,
	initials,
	Notice,
	Outcome,
	PageHead,
	plural,
	Setting,
	WindowCards,
	windowPage,
} from "../components";
import { useConsole } from "../context";
import { de, words } from "../i18n";
import { useLoad } from "../load";
import { Pictogram, type PictogramName } from "../pictograms";
import { Link, navigate } from "../router";

const t = words({
	en: {
		member: "member",
		members: "members",
		owner: "Owner",
		memberRole: "Member",
		individualText:
			"A Room holds the Windows of its owner, at its own address. In this Cahut, each person works alone: a Room keeps one member.",
		teamText: "A Room gathers its people and their Windows, at its own address. Windows published in a team Room are for its members.",
		yourRooms: "Your Rooms",
		everyRoom: "Every Room of this Cahut",
		asAdmin: "As an administrator, you manage the people of every Room, and open the Windows of yours only.",
		newRoom: "New Room",
		newTeamRoom: "New team Room",
		title: "Title",
		name: "Name, in its address",
		nameHelpIndividual: (handle?: string) =>
			`From the title when empty, after ${handle === undefined ? "the handle of its owner" : `${handle}-`}. It cannot change later.`,
		nameHelp: "From the title when empty. It cannot change later.",
		ownerLabel: "Owner",
		choosePerson: "Choose a person",
		createRoom: "Create the Room",
		personalRoom: "Personal Room",
		teamRoom: "Team Room",
		room: "Room",
		youOwn: "you are an owner",
		youMember: "you are a member",
		youOutside: "you are not in it",
		copyAddress: "Copy the address of the Room",
		copied: "Copied",
		shows: (what: ComponentChildren) => <>Shows {what}</>,
		theList: "the list of its Windows",
		open: "Open",
		manage: "Manage the Room",
		person: "person",
		persons: "people",
		peopleIn: (n: string) => `${n} in this Room`,
		notInRoom: "You are not in this Room. As an administrator, you manage its people, not its Windows.",
		emptyTitle: "No Window in this Room yet",
		emptyText: (name: string) => (
			<>
				An AI you connect publishes them. Ask it, for example, to publish a report in the Room <span class="c-mono">{name}</span>.
			</>
		),
		connectAI: "Connect an AI",
		trash: "Trash",
		trashedSince: (ago: string) => `In the trash since ${ago}`,
		restore: "Restore",
		people: "People",
		personColumn: "Person",
		roleColumn: "Role",
		actions: "Actions",
		you: "You",
		manageShort: "Manage",
		manageLabel: (name: string) => `Manage ${name}`,
		addPerson: "Add a person",
		leave: "Leave this Room",
		leaveQuestion: "Leave this Room? You will no longer open its Windows.",
		leaveConfirm: "Leave",
		leaveHelp: "You will no longer open its Windows. Its other owners keep it.",
		// The Windows of a person for themself open to the Room when they
		// leave it. See docs/adr/0053-windows-for-their-author-alone.md.
		yoursOpen: (n: number) =>
			n === 1 ? "The Window you keep for yourself here will open to its members." : `The ${n} Windows you keep for yourself here will open to its members.`,
		theirsOpen: (name: string, n: number) =>
			n === 1
				? `The Window that ${name} keeps for themself here will open to its members.`
				: `The ${n} Windows that ${name} keeps for themself here will open to its members.`,
		roleOf: (role: string, title: string) => `${role} of ${title}`,
		role: "Role",
		memberHelp: "Opens and publishes the Windows of the Room.",
		ownerHelp: "Also manages its people, its title and its page.",
		makeMember: "Make member",
		makeOwner: "Make owner",
		makeQuestion: (name: string, owner: boolean) => `Make ${name} ${owner ? "an owner" : "a member"}?`,
		roleDone: (name: string, owner: boolean) => `${name} is now ${owner ? "an owner" : "a member"} of this Room.`,
		removeTitle: "Remove from the Room",
		removeHelp: (name: string) => `${name} will no longer open the Windows of this Room.`,
		remove: "Remove",
		removeQuestion: (name: string) => `Remove ${name} from this Room?`,
		lastOwner: "The only owner of this Room stays: make another person owner first.",
		inRoom: (title: string) => `In ${title}`,
		candidatesHelp: "The people of this Cahut who are not in this Room yet.",
		everyoneIn: "Everyone in this Cahut is already in this Room.",
		willOpen: (name: string) => `${name} will open and publish the Windows of this Room.`,
		add: "Add",
		cancel: "Cancel",
		addressStays: "The address of the Room stays.",
		rename: "Rename",
		renamed: "Renamed.",
		description: "Description",
		descriptionHelp: "A sentence or two that say what the Room is. Readers see it on its page, and AIs read it.",
		saveDescription: "Save",
		described: "Saved.",
		pageOfRoom: "Page of the Room",
		homeHelp: (address: string) => (
			<>
				What <span class="c-mono">{address}</span> shows to its readers.
			</>
		),
		homeShows: "What the Room shows",
		homeList: "The list of its Windows",
		homeListHelp: "Each reader sees the Windows they may open, with a preview of each.",
		homeOne: "One Window",
		homeOneHelp: "It opens at the address of the Room. Readers who may not open it see the list.",
		shownWindow: "Window shown",
		save: "Save",
		saved: "Saved.",
		deleteRoom: "Delete this Room",
		deleteHelp: (windows: string) =>
			`Its people, ${windows}, those in its trash, those that its people keep for themselves, and their files. It cannot be undone.`,
		deleteAsk: "Delete…",
		typeToConfirm: (name: string) => (
			<>
				Type <span class="c-mono">{name}</span> to confirm
			</>
		),
		deleteForGood: "Delete for good",
	},
	fr: {
		member: "membre",
		members: "membres",
		owner: "Propriétaire",
		memberRole: "Membre",
		individualText:
			"Une Room garde les Windows de son propriétaire, à sa propre adresse. Dans cette Cahut, chacun travaille seul : une Room n'a qu'un membre.",
		teamText:
			"Une Room réunit ses membres et leurs Windows, à sa propre adresse. Les Windows publiées dans une Room d'équipe sont pour ses membres.",
		yourRooms: "Vos Rooms",
		everyRoom: "Toutes les Rooms de cette Cahut",
		asAdmin: "En tant qu'administrateur, vous gérez les personnes de chaque Room, et n'ouvrez que les Windows des vôtres.",
		newRoom: "Nouvelle Room",
		newTeamRoom: "Nouvelle Room d'équipe",
		title: "Titre",
		name: "Nom, dans son adresse",
		nameHelpIndividual: (handle) =>
			`Tiré du titre s'il est vide, après ${handle === undefined ? "l'identifiant de son propriétaire" : `${handle}-`}. Il ne pourra plus changer.`,
		nameHelp: "Tiré du titre s'il est vide. Il ne pourra plus changer.",
		ownerLabel: "Propriétaire",
		choosePerson: "Choisir une personne",
		createRoom: "Créer la Room",
		personalRoom: "Room personnelle",
		teamRoom: "Room d'équipe",
		room: "Room",
		youOwn: "vous êtes propriétaire",
		youMember: "vous êtes membre",
		youOutside: "vous n'y êtes pas",
		copyAddress: "Copier l'adresse de la Room",
		copied: "Copiée",
		shows: (what) => <>Montre {what}</>,
		theList: "la liste de ses Windows",
		open: "Ouvrir",
		manage: "Gérer la Room",
		person: "personne",
		persons: "personnes",
		peopleIn: (n) => `${n} dans cette Room`,
		notInRoom: "Vous n'êtes pas dans cette Room. En tant qu'administrateur, vous gérez ses personnes, pas ses Windows.",
		emptyTitle: "Pas encore de Window dans cette Room",
		emptyText: (name) => (
			<>
				Une IA que vous connectez les publie. Demandez-lui, par exemple, de publier un rapport dans la Room <span class="c-mono">{name}</span>.
			</>
		),
		connectAI: "Connecter une IA",
		trash: "Corbeille",
		trashedSince: (ago) => `Mise à la corbeille ${ago}`,
		restore: "Restaurer",
		people: "Personnes",
		personColumn: "Personne",
		roleColumn: "Rôle",
		actions: "Actions",
		you: "Vous",
		manageShort: "Gérer",
		manageLabel: (name) => `Gérer ${name}`,
		addPerson: "Ajouter une personne",
		leave: "Quitter cette Room",
		leaveQuestion: "Quitter cette Room ? Vous n'ouvrirez plus ses Windows.",
		leaveConfirm: "Quitter",
		leaveHelp: "Vous n'ouvrirez plus ses Windows. Ses autres propriétaires la gardent.",
		yoursOpen: (n) =>
			n === 1 ? "La Window que vous gardez pour vous ici s'ouvrira à ses membres." : `Les ${n} Windows que vous gardez pour vous ici s'ouvriront à ses membres.`,
		theirsOpen: (name, n) =>
			n === 1 ? `La Window gardée ici pour soi par ${name} s'ouvrira à ses membres.` : `Les ${n} Windows gardées ici pour soi par ${name} s'ouvriront à ses membres.`,
		roleOf: (role, title) => `${role} ${de(title)}`,
		role: "Rôle",
		memberHelp: "Ouvre et publie les Windows de la Room.",
		ownerHelp: "Gère aussi ses personnes, son titre et sa page.",
		makeMember: "Rendre membre",
		makeOwner: "Rendre propriétaire",
		makeQuestion: (name, owner) => `Faire ${de(name)} ${owner ? "un propriétaire" : "un membre"} ?`,
		roleDone: (name, owner) => `${name} est maintenant ${owner ? "propriétaire" : "membre"} de cette Room.`,
		removeTitle: "Retirer de la Room",
		removeHelp: (name) => `${name} n'ouvrira plus les Windows de cette Room.`,
		remove: "Retirer",
		removeQuestion: (name) => `Retirer ${name} de cette Room ?`,
		lastOwner: "Le seul propriétaire de cette Room reste : rendez d'abord une autre personne propriétaire.",
		inRoom: (title) => `Dans ${title}`,
		candidatesHelp: "Les personnes de cette Cahut qui ne sont pas encore dans cette Room.",
		everyoneIn: "Toutes les personnes de cette Cahut sont déjà dans cette Room.",
		willOpen: (name) => `${name} pourra ouvrir et publier les Windows de cette Room.`,
		add: "Ajouter",
		cancel: "Annuler",
		addressStays: "L'adresse de la Room ne change pas.",
		rename: "Renommer",
		renamed: "Renommée.",
		description: "Description",
		descriptionHelp: "Une phrase ou deux qui disent ce qu'est la Room. Les lecteurs la voient sur sa page, et les IA la lisent.",
		saveDescription: "Enregistrer",
		described: "Enregistrée.",
		pageOfRoom: "Page de la Room",
		homeHelp: (address) => (
			<>
				Ce que <span class="c-mono">{address}</span> montre à ses lecteurs.
			</>
		),
		homeShows: "Ce que montre la Room",
		homeList: "La liste de ses Windows",
		homeListHelp: "Chaque lecteur voit les Windows qu'il peut ouvrir, avec un aperçu de chacune.",
		homeOne: "Une Window",
		homeOneHelp: "Elle s'ouvre à l'adresse de la Room. Les lecteurs qui ne peuvent pas l'ouvrir voient la liste.",
		shownWindow: "Window montrée",
		save: "Enregistrer",
		saved: "Enregistré.",
		deleteRoom: "Supprimer cette Room",
		deleteHelp: (windows) =>
			`Ses personnes, ${windows}, celles de sa corbeille, celles que ses personnes gardent pour elles, et leurs fichiers. Cela ne s'annule pas.`,
		deleteAsk: "Supprimer…",
		typeToConfirm: (name) => (
			<>
				Tapez <span class="c-mono">{name}</span> pour confirmer
			</>
		),
		deleteForGood: "Supprimer définitivement",
	},
});

/**
 * The pictogram of a Room: one person on its floor for a personal Room, two
 * for a team Room, the Room alone for a Room of one's own in an individual Cahut.
 */
export function roomPictogram(r: Pick<Room, "kind">, individual: boolean): PictogramName {
	if (r.kind === "personal") return "roomPersonal";
	return individual ? "room" : "roomTeam";
}

/** Rooms as cards, each a link to its page. */
export function RoomCards({ rooms }: { rooms: Room[] }) {
	const { individual } = useConsole();
	return (
		<div class="cards">
			{rooms.map((r) => (
				<Link key={r.name} class="c-card card card--link" href={`/rooms/${r.name}`}>
					<span class="c-window__title">
						<Pictogram name={roomPictogram(r, individual)} />
						{r.title}
					</span>
					<span class="c-mono c-small c-muted">/{r.name}/</span>
					<span class="c-window__meta">
						<span>
							{plural(r.windows, "Window", "Windows")} · {plural(r.members, t.member, t.members)}
						</span>
						{r.role && <span class="c-badge c-badge--neutral">{r.role === "owner" ? t.owner : t.memberRole}</span>}
					</span>
				</Link>
			))}
		</div>
	);
}

export function RoomsPage() {
	const api = useConsole();
	const asAdmin = api.isAdmin;
	const { data, error } = useLoad(
		() => Promise.all([api.cahut("list_rooms", {}), asAdmin ? api.cahut("list_rooms", { all: true }) : Promise.resolve(null)]),
		[api.roomsVersion],
	);
	if (error) return <Failure message={error} />;
	if (!data) return null;
	const [mine, all] = data;
	return (
		<>
			<PageHead label="Rooms" title="Rooms">
				{api.individual ? t.individualText : t.teamText}
			</PageHead>
			{mine.rooms.length > 0 && (
				<section class="section">
					<h2 class="c-h4">{t.yourRooms}</h2>
					<RoomCards rooms={mine.rooms} />
				</section>
			)}
			{all && (
				<section class="section">
					<h2 class="c-h4">{t.everyRoom}</h2>
					<p class="c-small c-muted">{t.asAdmin}</p>
					<RoomCards rooms={all.rooms} />
				</section>
			)}
			<NewRoom />
		</>
	);
}

/** The form that creates a team Room, or in an individual Cahut a Room of one's own. */
function NewRoom() {
	const api = useConsole();
	const [title, setTitle] = useState("");
	const [name, setName] = useState("");
	const [owner, setOwner] = useState("");
	const [error, setError] = useState("");
	// The administrator of the server is no person of the Cahut: they name the owner.
	const people = useLoad(() => (api.session.kind === "admin" ? api.cahut("list_users", {}) : Promise.resolve(null)));
	const submit = async (e: Event) => {
		e.preventDefault();
		setError("");
		try {
			const r = await api.cahut("create_room", { title, name: name || undefined, owner: owner || undefined });
			api.reloadRooms();
			navigate(`/rooms/${r.name}`);
		} catch (err) {
			setError(explain(err));
		}
	};
	return (
		<section class="section c-card panel">
			<h2 class="c-h4">{api.individual ? t.newRoom : t.newTeamRoom}</h2>
			<form class="form" onSubmit={submit}>
				<div class="form__row">
					<label class="c-field">
						<span class="c-field__label">{t.title}</span>
						<input class="c-input" required placeholder="Marketing" value={title} onInput={(e) => setTitle(e.currentTarget.value)} />
					</label>
					<label class="c-field">
						<span class="c-field__label">{t.name}</span>
						<input class="c-input c-mono" placeholder="marketing" autocapitalize="none" value={name} onInput={(e) => setName(e.currentTarget.value)} />
						<span class="c-field__help">
							{api.individual ? t.nameHelpIndividual(api.session.kind === "admin" ? undefined : api.session.handle) : t.nameHelp}
						</span>
					</label>
				</div>
				{people.data && (
					<label class="c-field">
						<span class="c-field__label">{t.ownerLabel}</span>
						<select class="c-input" required value={owner} onChange={(e) => setOwner(e.currentTarget.value)}>
							<option value="">{t.choosePerson}</option>
							{people.data.users.map((u) => (
								<option key={u.handle} value={u.handle}>
									{u.name} ({u.handle})
								</option>
							))}
						</select>
					</label>
				)}
				{error && <Failure message={error} />}
				<div>
					<button class="c-btn c-btn--primary" type="submit">
						<Plus class="i" size={16} aria-hidden="true" />
						{t.createRoom}
					</button>
				</div>
			</form>
		</section>
	);
}

/**
 * The page of a Room: its head, its Windows as wide as the page with its
 * trash folded under them, then the people of a team Room. The changes open
 * dialogs, as on the People screen.
 */
export function RoomPage({ name }: { name: string }) {
	const api = useConsole();
	const room = useLoad(() => api.cahut("get_room", { room: name }), [name, api.roomsVersion]);
	const windows = useLoad(
		() =>
			api.cahut("list_windows", { room: name }).catch((err) => {
				// An administrator manages a Room that is not theirs, without its Windows.
				if (err instanceof ApiError && err.status === 403) return null;
				throw err;
			}),
		[name],
	);
	const [dialog, setDialog] = useState<"manage" | "add" | null>(null);
	const [managing, setManaging] = useState<string | null>(null);
	const [leaveError, setLeaveError] = useState("");
	if (room.error) return <Failure message={room.error} />;
	if (!room.data) return null;
	const r = room.data;
	const team = r.kind === "team";
	// The owners of a team Room, and the administrators, manage its people, its title and its end.
	const manage = team && (r.role === "owner" || api.isAdmin);
	// In an individual Cahut, a Room keeps one member: nobody joins or leaves it.
	const members = !api.individual;
	const owners = r.people.filter((p) => p.role === "owner").length;
	const list = windows.data?.windows;
	// What the address shows is set by those who own the Room, from its Windows.
	const home = (r.role === "owner" || api.isAdmin) && list !== undefined;
	const leave = team && members && r.role === "owner" && owners > 1;
	// Its owners, personal Rooms included, and the administrators describe it.
	const describe = r.role === "owner" || api.isAdmin;
	const reload = () => {
		api.reloadRooms();
		room.reload();
	};
	const person = r.people.find((p) => p.handle === managing);
	const yours = r.people.find((p) => p.handle === api.session.handle)?.kept ?? 0;
	return (
		<>
			<RoomHead
				room={r}
				windows={list}
				onPeople={team && members ? () => document.getElementById("people")?.scrollIntoView({ behavior: "smooth" }) : undefined}
				onManage={manage || home || leave || describe ? () => setDialog("manage") : undefined}
			/>
			<section class="room-section">
				<div class="room-section__head">
					<h2 class="c-h4">
						Windows <span class="room-section__count">{r.windows.toLocaleString()}</span>
					</h2>
				</div>
				{windows.error && <Failure message={windows.error} />}
				{windows.data === null && <Notice>{t.notInRoom}</Notice>}
				{list && list.length === 0 && <EmptyRoom name={r.name} />}
				{list && list.length > 0 && <WindowCards windows={list} home={r.home} onChanged={windows.reload} />}
				{list && (
					<RoomTrash
						room={r.name}
						onRestore={() => {
							windows.reload();
							reload();
						}}
					/>
				)}
			</section>
			{team && members && (
				<section class="room-section" id="people">
					<div class="room-section__head">
						<h2 class="c-h4">
							{t.people} <span class="room-section__count">{r.people.length.toLocaleString()}</span>
						</h2>
						{manage ? (
							<button class="c-btn c-btn--secondary c-btn--sm" type="button" onClick={() => setDialog("add")}>
								<Plus class="i" size={14} aria-hidden="true" />
								{t.addPerson}
							</button>
						) : (
							r.role === "member" && (
								<ConfirmButton
									label={t.leave}
									question={withKept(t.leaveQuestion, yours, t.yoursOpen)}
									confirm={t.leaveConfirm}
									danger
									onConfirm={async () => {
										setLeaveError("");
										try {
											await api.cahut("remove_room_member", { room: r.name, user: api.session.handle });
											api.reloadRooms();
											navigate("/rooms");
										} catch (err) {
											setLeaveError(explain(err));
										}
									}}
								/>
							)
						)}
					</div>
					{leaveError && <Failure message={leaveError} />}
					<PeopleTable room={r} manage={manage} onManage={setManaging} />
				</section>
			)}
			{dialog === "manage" && (
				<ManageRoom
					room={r}
					windows={home ? list : undefined}
					manage={manage}
					describe={describe}
					leave={leave}
					onChanged={reload}
					onClose={() => setDialog(null)}
				/>
			)}
			{dialog === "add" && <AddPerson room={r} onAdded={reload} onClose={() => setDialog(null)} />}
			{person && (
				<ManageMember key={person.handle} room={r} person={person} owners={owners} onChanged={reload} onClose={() => setManaging(null)} />
			)}
		</>
	);
}

/** The head of a Room: what it is, its address and what that shows, who is in it, and its two actions. */
function RoomHead(props: { room: RoomDetail; windows?: Window[]; onPeople?: () => void; onManage?: () => void }) {
	const { individual } = useConsole();
	const r = props.room;
	const label = r.kind === "personal" ? t.personalRoom : individual ? t.room : t.teamRoom;
	const role = r.role === "owner" ? t.youOwn : r.role === "member" ? t.youMember : t.youOutside;
	const shown = r.home ? (props.windows?.find((w) => w.name === r.home)?.title ?? r.home) : null;
	const people = plural(r.people.length, t.person, t.persons);
	return (
		<div class="room-head">
			<div class="room-head__text">
				<span class="c-label">{label}</span>
				<h1 class="c-h1">{r.title}</h1>
				{r.description && <p class="page-head__lead">{r.description}</p>}
				<p class="c-muted">
					{plural(r.windows, "Window", "Windows")}
					{r.kind === "team" && !individual && ` · ${role}`}
				</p>
				<div class="room-address">
					<span class="room-address__copy">
						<span class="room-address__url">
							<Link2 class="i" size={16} aria-hidden="true" />
							<span>
								{location.host}/<wbr />
								{r.name}/
							</span>
						</span>
						<CopyAddress value={`${location.origin}/${r.name}/`} />
					</span>
					<span class="room-address__shows">
						<House class="i" size={16} aria-hidden="true" />
						<span>{t.shows(<strong>{shown ?? t.theList}</strong>)}</span>
					</span>
				</div>
			</div>
			<div class="room-head__actions">
				{props.onPeople && (
					<button class="room-people" type="button" aria-label={t.peopleIn(people)} onClick={props.onPeople}>
						<span class="c-avatars" aria-hidden="true">
							{r.people.slice(0, 3).map((p) => (
								<span key={p.handle} class="c-avatar">
									{initials(p.name)}
								</span>
							))}
							{r.people.length > 3 && <span class="c-avatar">+{r.people.length - 3}</span>}
						</span>
						{people}
					</button>
				)}
				<a class="c-btn c-btn--secondary" href={`/${r.name}/`} target="_blank" rel="noopener">
					<ArrowUpRight class="i" size={16} aria-hidden="true" />
					{t.open}
				</a>
				{props.onManage && (
					<button class="c-btn c-btn--secondary" type="button" onClick={props.onManage}>
						<Settings class="i" size={16} aria-hidden="true" />
						{t.manage}
					</button>
				)}
			</div>
		</div>
	);
}

/** Copies the address of a Room, from a button that says so. */
function CopyAddress({ value }: { value: string }) {
	const [copied, setCopied] = useState(false);
	const copy = async () => {
		await navigator.clipboard.writeText(value);
		setCopied(true);
		setTimeout(() => setCopied(false), 2000);
	};
	return (
		<button class="c-btn c-btn--ghost c-btn--sm icon-button" type="button" aria-label={copied ? t.copied : t.copyAddress} title={copied ? t.copied : t.copyAddress} onClick={copy}>
			{copied ? <Check class="i" size={14} aria-hidden="true" /> : <Copy class="i" size={14} aria-hidden="true" />}
		</button>
	);
}

/** A Room without a Window yet: who publishes them, and where to connect an AI. */
function EmptyRoom({ name }: { name: string }) {
	return (
		<div class="room-empty">
			<Pictogram name="window" large />
			<h3 class="c-h4">{t.emptyTitle}</h3>
			<p class="c-muted">{t.emptyText(name)}</p>
			<Link class="c-btn c-btn--secondary c-btn--sm" href="/access">
				<Pictogram name="aiAccess" />
				{t.connectAI}
			</Link>
		</div>
	);
}

/** The Windows of a Room in its trash, folded under the others, each with its button to restore it. */
function RoomTrash({ room, onRestore }: { room: string; onRestore: () => void }) {
	const api = useConsole();
	const { data, reload } = useLoad(() => api.cahut("list_windows", { room, trashed: true }).catch(() => null), [room]);
	const [error, setError] = useState("");
	if (!data || data.windows.length === 0) return null;
	const restore = async (w: Window) => {
		setError("");
		try {
			await api.cahut("restore_window", { window: `${w.room}/${w.name}` });
			reload();
			onRestore();
		} catch (err) {
			setError(explain(err));
		}
	};
	return (
		<details class="trash">
			<summary>
				<ChevronRight class="i trash__chevron" size={14} aria-hidden="true" />
				<Trash2 class="i" size={14} aria-hidden="true" />
				{t.trash} · {plural(data.windows.length, "Window", "Windows")}
			</summary>
			<ul class="trash__list">
				{data.windows.map((w) => (
					<li key={w.name} class="trash__row">
						<Link class="c-link" href={windowPage(w)}>
							{w.title}
						</Link>
						<span class="c-small c-muted">{w.trashed_at ? t.trashedSince(agoInText(w.trashed_at)) : ""}</span>
						<button class="c-btn c-btn--secondary c-btn--sm" type="button" onClick={() => restore(w)}>
							<Undo2 class="i" size={14} aria-hidden="true" />
							{t.restore}
						</button>
					</li>
				))}
			</ul>
			{error && <Failure message={error} />}
		</details>
	);
}

/** The people of a team Room, owners first, with a Manage button each for those who manage it. */
function PeopleTable({ room, manage, onManage }: { room: RoomDetail; manage: boolean; onManage: (handle: string) => void }) {
	const { session } = useConsole();
	return (
		<div class="table-wrap">
			<table class="table people-table">
				<thead>
					<tr>
						<th>{t.personColumn}</th>
						<th class="people-table__wide">{t.roleColumn}</th>
						{manage && (
							<th>
								<span class="visually-hidden">{t.actions}</span>
							</th>
						)}
					</tr>
				</thead>
				<tbody>
					{room.people.map((p) => (
						<tr key={p.handle}>
							<td>
								<span class="person">
									<span class="c-avatar" aria-hidden="true">
										{initials(p.name)}
									</span>
									<span class="people__who">
										<span class="people__name">
											{p.name}
											{p.handle === session.handle && <span class="c-badge c-badge--neutral">{t.you}</span>}
										</span>
										<span class="c-mono c-muted people__handle">{p.handle}</span>
										<span class="people-table__narrow">
											<span class="c-small c-muted">{roleWord(p)}</span>
										</span>
									</span>
								</span>
							</td>
							<td class="people-table__wide people-table__role">{roleWord(p)}</td>
							{manage && (
								<td class="table__actions">
									{p.handle !== session.handle && (
										<button class="c-btn c-btn--secondary c-btn--sm" type="button" aria-label={t.manageLabel(p.name)} onClick={() => onManage(p.handle)}>
											{t.manageShort}
										</button>
									)}
								</td>
							)}
						</tr>
					))}
				</tbody>
			</table>
		</div>
	);
}

function roleWord(p: Pick<Person, "role">): string {
	return p.role === "owner" ? t.owner : t.memberRole;
}

/** The dialog that manages the Room itself: its title, its description, what its address shows, leaving it, deleting it. */
function ManageRoom(props: {
	room: RoomDetail;
	/** The Windows to choose from for the page of the Room; none when the person may not set it. */
	windows?: Window[];
	manage: boolean;
	describe: boolean;
	leave: boolean;
	onChanged: () => void;
	onClose: () => void;
}) {
	const api = useConsole();
	const { individual } = api;
	const r = props.room;
	const label = r.kind === "personal" ? t.personalRoom : individual ? t.room : t.teamRoom;
	return (
		<Dialog
			title={r.title}
			meta={
				<>
					<span>{label}</span>
					<span aria-hidden="true">·</span>
					<span class="c-mono">
						{location.host}/{r.name}/
					</span>
				</>
			}
			onClose={props.onClose}
		>
			<div class="settings">
				{props.manage && (
					<TitleSetting
						room={r}
						onSave={async (title) => {
							await api.cahut("update_room", { room: r.name, title });
							props.onChanged();
						}}
					/>
				)}
				{props.describe && (
					<DescriptionSetting
						room={r}
						onSave={async (description) => {
							await api.cahut("update_room", { room: r.name, description });
							props.onChanged();
						}}
					/>
				)}
				{props.windows && (
					<HomeSetting
						room={r}
						windows={props.windows}
						onSave={async (window) => {
							await api.cahut("set_room_home", { room: r.name, window });
							props.onChanged();
						}}
					/>
				)}
				{props.leave && (
					<Setting
						title={t.leave}
						help={withKept(t.leaveHelp, r.people.find((p) => p.handle === api.session.handle)?.kept ?? 0, t.yoursOpen)}
						action={t.leaveConfirm}
						ask={t.leaveQuestion}
						onAct={async () => {
							await api.cahut("remove_room_member", { room: r.name, user: api.session.handle });
							api.reloadRooms();
							navigate("/rooms");
							return null;
						}}
					/>
				)}
				{props.manage && <DeleteSetting room={r} />}
			</div>
		</Dialog>
	);
}

/** The title of a team Room: its address stays. */
function TitleSetting({ room, onSave }: { room: RoomDetail; onSave: (title: string) => Promise<void> }) {
	const id = useId();
	const [title, setTitle] = useState(room.title);
	const [busy, setBusy] = useState(false);
	const [said, setSaid] = useState("");
	const [error, setError] = useState("");
	const save = async (e: Event) => {
		e.preventDefault();
		setBusy(true);
		setError("");
		setSaid("");
		try {
			await onSave(title.trim());
			setSaid(t.renamed);
		} catch (err) {
			setError(explain(err));
		} finally {
			setBusy(false);
		}
	};
	return (
		<form class="setting" onSubmit={save}>
			<div class="setting__text">
				<label class="setting__title" for={id}>
					{t.title}
				</label>
				<p class="c-small c-muted">{t.addressStays}</p>
			</div>
			<div class="setting__more setting__inline">
				<input id={id} class="c-input" required value={title} onInput={(e) => (setTitle(e.currentTarget.value), setSaid(""))} />
				<button class="c-btn c-btn--secondary" type="submit" disabled={busy || title.trim() === room.title}>
					{t.rename}
				</button>
			</div>
			{said && <Outcome>{said}</Outcome>}
			{error && (
				<div class="setting__more">
					<Failure message={error} />
				</div>
			)}
		</form>
	);
}

/** The description of a Room: a sentence or two, on its page and for AIs. */
function DescriptionSetting({ room, onSave }: { room: RoomDetail; onSave: (description: string) => Promise<void> }) {
	const id = useId();
	const [description, setDescription] = useState(room.description ?? "");
	const [busy, setBusy] = useState(false);
	const [said, setSaid] = useState("");
	const [error, setError] = useState("");
	const save = async (e: Event) => {
		e.preventDefault();
		setBusy(true);
		setError("");
		setSaid("");
		try {
			await onSave(description);
			setSaid(t.described);
		} catch (err) {
			setError(explain(err));
		} finally {
			setBusy(false);
		}
	};
	return (
		<form class="setting" onSubmit={save}>
			<div class="setting__text">
				<label class="setting__title" for={id}>
					{t.description}
				</label>
				<p class="c-small c-muted">{t.descriptionHelp}</p>
			</div>
			<div class="setting__more c-field">
				<DescriptionField id={id} value={description} onInput={(v) => (setDescription(v), setSaid(""))} />
			</div>
			<div class="setting__more">
				<button class="c-btn c-btn--secondary" type="submit" disabled={busy || description.trim() === (room.description ?? "")}>
					{t.saveDescription}
				</button>
			</div>
			{said && <Outcome>{said}</Outcome>}
			{error && (
				<div class="setting__more">
					<Failure message={error} />
				</div>
			)}
		</form>
	);
}

/**
 * What the address of a Room shows: the list of its Windows, or one of
 * them, for its readers who open it. See docs/adr/0032-the-page-of-a-room.md.
 */
function HomeSetting({ room, windows, onSave }: { room: RoomDetail; windows: Window[]; onSave: (window?: string) => Promise<void> }) {
	const name = useId();
	const current = room.home ?? "";
	const [one, setOne] = useState(current !== "");
	const [home, setHome] = useState(current || (windows[0]?.name ?? ""));
	const [busy, setBusy] = useState(false);
	const [said, setSaid] = useState("");
	const [error, setError] = useState("");
	const chosen = one ? home : "";
	const save = async (e: Event) => {
		e.preventDefault();
		setBusy(true);
		setError("");
		setSaid("");
		try {
			await onSave(chosen || undefined);
			setSaid(t.saved);
		} catch (err) {
			setError(explain(err));
		} finally {
			setBusy(false);
		}
	};
	const pick = (value: boolean) => {
		setOne(value);
		setSaid("");
	};
	return (
		<form class="setting" onSubmit={save}>
			<div class="setting__text">
				<span class="setting__title">{t.pageOfRoom}</span>
				<p class="c-small c-muted">{t.homeHelp(`${location.host}/${room.name}/`)}</p>
			</div>
			<div class="setting__more home-form">
				<fieldset class="choices choices--stacked">
					<legend class="visually-hidden">{t.homeShows}</legend>
					<label class="choice">
						<input type="radio" name={name} checked={!one} onChange={() => pick(false)} />
						<span class="choice__name">{t.homeList}</span>
						<span class="choice__help">{t.homeListHelp}</span>
					</label>
					<label class="choice">
						<input type="radio" name={name} checked={one} disabled={windows.length === 0} onChange={() => pick(true)} />
						<span class="choice__name">{t.homeOne}</span>
						<span class="choice__help">{t.homeOneHelp}</span>
						{windows.length > 0 && (
							<select
								class="c-input choice__more"
								aria-label={t.shownWindow}
								value={home}
								onChange={(e) => (setHome(e.currentTarget.value), pick(true))}
							>
								{windows.map((w) => (
									<option key={w.name} value={w.name}>
										{w.title}
									</option>
								))}
							</select>
						)}
					</label>
				</fieldset>
				<div>
					<button class="c-btn c-btn--secondary" type="submit" disabled={busy || chosen === current}>
						{t.save}
					</button>
				</div>
			</div>
			{said && <Outcome>{said}</Outcome>}
			{error && (
				<div class="setting__more">
					<Failure message={error} />
				</div>
			)}
		</form>
	);
}

/** Deletes a team Room for good, once its name is typed. */
function DeleteSetting({ room }: { room: RoomDetail }) {
	const api = useConsole();
	const [asking, setAsking] = useState(false);
	const [typed, setTyped] = useState("");
	const [busy, setBusy] = useState(false);
	const [error, setError] = useState("");
	const remove = async (e: Event) => {
		e.preventDefault();
		setBusy(true);
		setError("");
		try {
			await api.cahut("delete_room", { room: room.name });
			api.reloadRooms();
			navigate("/rooms");
		} catch (err) {
			setError(explain(err));
			setBusy(false);
		}
	};
	return (
		<div class="setting">
			<div class="setting__text">
				<span class="setting__title">{t.deleteRoom}</span>
				<p class="c-small c-muted">{t.deleteHelp(plural(room.windows, "Window", "Windows"))}</p>
			</div>
			{!asking && (
				<div class="setting__action">
					<button class="c-btn c-btn--danger" type="button" onClick={() => setAsking(true)}>
						{t.deleteAsk}
					</button>
				</div>
			)}
			{asking && (
				<form class="setting__more setting__confirm" onSubmit={remove}>
					<label class="c-field">
						<span class="c-field__label">{t.typeToConfirm(room.name)}</span>
						<input
							class="c-input c-mono"
							autoComplete="off"
							autocapitalize="none"
							spellcheck={false}
							value={typed}
							onInput={(e) => setTyped(e.currentTarget.value)}
						/>
					</label>
					{error && <Failure message={error} />}
					<div class="actions">
						<button class="c-btn c-btn--danger c-btn--sm" type="submit" disabled={busy || typed.trim() !== room.name}>
							{t.deleteForGood}
						</button>
						<button class="c-btn c-btn--ghost c-btn--sm" type="button" onClick={() => (setAsking(false), setTyped(""))}>
							{t.cancel}
						</button>
					</div>
				</form>
			)}
		</div>
	);
}

/** The dialog that manages a person of a team Room: their role in it, and removing them. */
function ManageMember(props: { room: RoomDetail; person: Person; owners: number; onChanged: () => void; onClose: () => void }) {
	const api = useConsole();
	const { room: r, person: p } = props;
	const toOwner = p.role !== "owner";
	// A Room keeps an owner: the last one can neither become a member nor leave it.
	const last = p.role === "owner" && props.owners === 1;
	return (
		<Dialog
			title={p.name}
			meta={
				<>
					<span class="c-mono">{p.handle}</span>
					<span aria-hidden="true">·</span>
					<span>{t.roleOf(roleWord(p), r.title)}</span>
				</>
			}
			onClose={props.onClose}
		>
			{last ? (
				<Notice>{t.lastOwner}</Notice>
			) : (
				<div class="settings">
					<Setting
						title={t.role}
						help={`${roleWord(p)}. ${p.role === "owner" ? t.ownerHelp : t.memberHelp}`}
						action={toOwner ? t.makeOwner : t.makeMember}
						ask={t.makeQuestion(p.name, toOwner)}
						onAct={async () => {
							await api.cahut("set_room_member_role", { room: r.name, user: p.handle, role: toOwner ? "owner" : "member" });
							props.onChanged();
							return t.roleDone(p.name, toOwner);
						}}
					/>
					<Setting
						title={t.removeTitle}
						help={withKept(t.removeHelp(p.name), p.kept ?? 0, (n) => t.theirsOpen(p.name, n))}
						action={t.remove}
						danger
						ask={t.removeQuestion(p.name)}
						onAct={async () => {
							await api.cahut("remove_room_member", { room: r.name, user: p.handle });
							props.onChanged();
							props.onClose();
							return null;
						}}
					/>
				</div>
			)}
		</Dialog>
	);
}

/** The dialog that adds a person of the Cahut to a team Room, as a member or an owner. */
function AddPerson(props: { room: RoomDetail; onAdded: () => void; onClose: () => void }) {
	const api = useConsole();
	const r = props.room;
	const people = useLoad(() => api.cahut("list_users", {}));
	const name = useId();
	const [handle, setHandle] = useState("");
	const [role, setRole] = useState<"member" | "owner">("member");
	const [busy, setBusy] = useState(false);
	const [error, setError] = useState("");
	const inRoom = new Set(r.people.map((p) => p.handle));
	const candidates = (people.data?.users ?? []).filter((u) => !inRoom.has(u.handle) && u.status !== "disabled");
	const chosen = candidates.find((u) => u.handle === handle);
	const submit = async (e: Event) => {
		e.preventDefault();
		setBusy(true);
		setError("");
		try {
			await api.cahut("add_room_member", { room: r.name, user: handle, role });
			props.onAdded();
			props.onClose();
		} catch (err) {
			setError(explain(err));
			setBusy(false);
		}
	};
	return (
		<Dialog title={t.addPerson} meta={<span>{t.inRoom(r.title)}</span>} onClose={props.onClose}>
			{people.error && <Failure message={people.error} />}
			{people.data && candidates.length === 0 && <Notice>{t.everyoneIn}</Notice>}
			{candidates.length > 0 && (
				<form class="form" onSubmit={submit}>
					<label class="c-field">
						<span class="c-field__label">{t.personColumn}</span>
						<select class="c-input" required value={handle} onChange={(e) => setHandle(e.currentTarget.value)}>
							<option value="">{t.choosePerson}</option>
							{candidates.map((u) => (
								<option key={u.handle} value={u.handle}>
									{u.name} ({u.handle})
								</option>
							))}
						</select>
						<span class="c-field__help">{t.candidatesHelp}</span>
					</label>
					<fieldset class="choices">
						<legend class="c-field__label">{t.role}</legend>
						{(["member", "owner"] as const).map((v) => (
							<label key={v} class="choice">
								<input type="radio" name={name} checked={role === v} onChange={() => setRole(v)} />
								<span class="choice__name">{v === "owner" ? t.owner : t.memberRole}</span>
								<span class="choice__help">{v === "owner" ? t.ownerHelp : t.memberHelp}</span>
							</label>
						))}
					</fieldset>
					{chosen && <p class="c-small dialog__intro">{t.willOpen(chosen.name)}</p>}
					{error && <Failure message={error} />}
					<div class="dialog__foot">
						<button class="c-btn c-btn--secondary" type="button" onClick={props.onClose}>
							{t.cancel}
						</button>
						<button class="c-btn c-btn--primary" type="submit" disabled={busy || !chosen}>
							{t.add}
						</button>
					</div>
				</form>
			)}
		</Dialog>
	);
}

/** A sentence, and what happens to the Windows that a person keeps for themself, when they keep any. */
function withKept(sentence: string, kept: number, open: (n: number) => string): string {
	return kept > 0 ? `${sentence} ${open(kept)}` : sentence;
}
