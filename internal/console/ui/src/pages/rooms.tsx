// The Rooms: those of the person, every Room for an administrator, a new
// team Room, and the page of one Room with its Windows, its people and
// what its address shows.
import { Folder, Plus, User, Users } from "lucide-preact";
import type { ComponentChildren } from "preact";
import { useState } from "preact/hooks";
import { ApiError, explain } from "../api";
import type { Room, RoomDetail, Window } from "../api.gen";
import { ConfirmButton, Failure, initials, Notice, PageHead, plural, WindowCards } from "../components";
import { useConsole } from "../context";
import { de, words } from "../i18n";
import { useLoad } from "../load";
import { Link, navigate } from "../router";
import { Trash } from "./window";

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
		room: "Room",
		notInRoom: "You are not in this Room. As an administrator, you manage its people, not its Windows.",
		noWindow: "No Window in this Room yet.",
		publishedByOwner: "An AI of its owner publishes them.",
		publishedByMember: "An AI of one of its members publishes them.",
		people: "People",
		makeMember: "Make member",
		makeOwner: "Make owner",
		makeQuestion: (name: string, owner: boolean) => `Make ${name} ${owner ? "an owner" : "a member"}?`,
		yes: "Yes",
		remove: "Remove",
		removeQuestion: (name: string) => `Remove ${name} from this Room?`,
		leave: "Leave this Room",
		leaveQuestion: "Leave this Room? You will no longer open its Windows.",
		leaveConfirm: "Leave",
		addPerson: "Add a person",
		as: "As",
		willOpen: (name: string) => `${name} will open and publish the Windows of this Room.`,
		add: "Add",
		cancel: "Cancel",
		addressStays: "The address of the Room stays.",
		rename: "Rename",
		pageOfRoom: "Page of the Room",
		homeText: (link: ComponentChildren) => <>What {link} shows to its readers.</>,
		show: "Show",
		roomWindows: "The Windows of the Room",
		homeHelp: "Readers who may not open this Window see the list of the Windows they open.",
		listHelp: "Each reader sees the Windows they may open, with a preview of each.",
		save: "Save",
		deleteRoom: "Delete this Room",
		deleteText: (title: string, windows: string) =>
			`This deletes the Room ${title} for good: its people, ${windows}, those in its trash, and their files. It cannot be undone.`,
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
		room: "Room",
		notInRoom: "Vous n'êtes pas dans cette Room. En tant qu'administrateur, vous gérez ses personnes, pas ses Windows.",
		noWindow: "Pas encore de Window dans cette Room.",
		publishedByOwner: "Une IA de son propriétaire les publie.",
		publishedByMember: "Une IA de l'un de ses membres les publie.",
		people: "Personnes",
		makeMember: "Rendre membre",
		makeOwner: "Rendre propriétaire",
		makeQuestion: (name, owner) => `Faire ${de(name)} ${owner ? "un propriétaire" : "un membre"} ?`,
		yes: "Oui",
		remove: "Retirer",
		removeQuestion: (name) => `Retirer ${name} de cette Room ?`,
		leave: "Quitter cette Room",
		leaveQuestion: "Quitter cette Room ? Vous n'ouvrirez plus ses Windows.",
		leaveConfirm: "Quitter",
		addPerson: "Ajouter une personne",
		as: "En tant que",
		willOpen: (name) => `${name} pourra ouvrir et publier les Windows de cette Room.`,
		add: "Ajouter",
		cancel: "Annuler",
		addressStays: "L'adresse de la Room ne change pas.",
		rename: "Renommer",
		pageOfRoom: "Page de la Room",
		homeText: (link) => <>Ce que {link} montre à ses lecteurs.</>,
		show: "Afficher",
		roomWindows: "Les Windows de la Room",
		homeHelp: "Les lecteurs qui ne peuvent pas ouvrir cette Window voient la liste des Windows qu'ils ouvrent.",
		listHelp: "Chaque lecteur voit les Windows qu'il peut ouvrir, avec un aperçu de chacune.",
		save: "Enregistrer",
		deleteRoom: "Supprimer cette Room",
		deleteText: (title, windows) =>
			`Cela supprime définitivement la Room ${title} : ses personnes, ${windows}, celles de sa corbeille, et leurs fichiers. Cela ne s'annule pas.`,
		typeToConfirm: (name) => (
			<>
				Tapez <span class="c-mono">{name}</span> pour confirmer
			</>
		),
		deleteForGood: "Supprimer définitivement",
	},
});

/** The icon of a Room: a person for a personal Room, a group for a team Room, a folder for a Room of one's own in an individual Cahut. */
export function roomIcon(r: Pick<Room, "kind">, individual: boolean): typeof User {
	if (r.kind === "personal") return User;
	return individual ? Folder : Users;
}

/** Rooms as cards, each a link to its page. */
export function RoomCards({ rooms }: { rooms: Room[] }) {
	const { individual } = useConsole();
	return (
		<div class="cards">
			{rooms.map((r) => {
				const Icon = roomIcon(r, individual);
				return (
					<Link key={r.name} class="c-card card card--link" href={`/rooms/${r.name}`}>
						<span class="c-window__title">
							<Icon class="i" size={16} aria-hidden="true" />
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
				);
			})}
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
	if (room.error) return <Failure message={room.error} />;
	if (!room.data) return null;
	const r = room.data;
	const manage = r.kind === "team" && (r.role === "owner" || api.isAdmin);
	// In an individual Cahut, a Room keeps one member: nobody joins or leaves it.
	const members = !api.individual;
	return (
		<>
			<PageHead label={r.kind === "personal" ? t.personalRoom : t.room} title={r.title}>
				{plural(r.windows, "Window", "Windows")}, {plural(r.members, t.member, t.members)} ·{" "}
				<span class="c-mono">
					{location.hostname}/{r.name}/
				</span>
			</PageHead>
			<div class="layout">
				<section>
					<h2 class="c-h4">Windows</h2>
					{windows.error && <Failure message={windows.error} />}
					{windows.data === null && <Notice>{t.notInRoom}</Notice>}
					{windows.data && windows.data.windows.length === 0 && (
						<Notice>
							{t.noWindow} {api.individual ? t.publishedByOwner : t.publishedByMember}
						</Notice>
					)}
					{windows.data && windows.data.windows.length > 0 && (
						<WindowCards windows={windows.data.windows} />
					)}
					{windows.data && <Trash room={r.name} />}
				</section>
				<div class="aside-stack">
					<People room={r} manage={manage} members={members} onChange={room.reload} />
					{(r.role === "owner" || api.isAdmin) && windows.data && (
						<HomePage room={r} windows={windows.data.windows} onChange={room.reload} />
					)}
				</div>
			</div>
		</>
	);
}

/** The people of a Room, and their management for its owners, when its Cahut lets people share Rooms. */
function People({ room, manage, members, onChange }: { room: RoomDetail; manage: boolean; members: boolean; onChange: () => void }) {
	const api = useConsole();
	const [error, setError] = useState("");
	const act = async (run: () => Promise<unknown>) => {
		setError("");
		try {
			await run();
			api.reloadRooms();
			onChange();
		} catch (err) {
			setError(explain(err));
		}
	};
	const owners = room.people.filter((p) => p.role === "owner").length;
	const me = api.session.handle;
	return (
		<aside class="c-card panel aside">
			<h2 class="c-h4">{t.people}</h2>
			<ul class="people">
				{room.people.map((p) => (
					<li key={p.handle} class="people__row">
						<span class="c-avatar" aria-hidden="true">
							{initials(p.name)}
						</span>
						<span class="people__who">
							<span class="c-small">{p.name}</span>
							<span class="c-mono c-muted people__handle">{p.handle}</span>
						</span>
						<span class="c-badge c-badge--neutral">{p.role === "owner" ? t.owner : t.memberRole}</span>
						{manage && members && p.handle !== me && (
							<span class="people__actions">
								{(p.role === "member" || owners > 1) && (
									<ConfirmButton
										label={p.role === "owner" ? t.makeMember : t.makeOwner}
										question={t.makeQuestion(p.name, p.role !== "owner")}
										confirm={t.yes}
										onConfirm={() =>
											act(() =>
												api.cahut("set_room_member_role", { room: room.name, user: p.handle, role: p.role === "owner" ? "member" : "owner" }),
											)
										}
									/>
								)}
								{(p.role === "member" || owners > 1) && (
									<ConfirmButton
										label={t.remove}
										question={t.removeQuestion(p.name)}
										confirm={t.remove}
										danger
										onConfirm={() => act(() => api.cahut("remove_room_member", { room: room.name, user: p.handle }))}
									/>
								)}
							</span>
						)}
					</li>
				))}
			</ul>
			{error && <Failure message={error} />}
			{manage && members && <AddPerson room={room} onAdd={(handle, role) => act(() => api.cahut("add_room_member", { room: room.name, user: handle, role }))} />}
			{manage && <Retitle room={room} onRetitle={(title) => act(() => api.cahut("update_room", { room: room.name, title }))} />}
			{manage && <DeleteRoom room={room} />}
			{members && room.kind === "team" && room.role && (room.role === "member" || owners > 1) && (
				<ConfirmButton
					label={t.leave}
					question={t.leaveQuestion}
					confirm={t.leaveConfirm}
					danger
					onConfirm={async () => {
						await act(() => api.cahut("remove_room_member", { room: room.name, user: me }));
						navigate("/rooms");
					}}
				/>
			)}
		</aside>
	);
}

function AddPerson({ room, onAdd }: { room: RoomDetail; onAdd: (handle: string, role: string) => Promise<void> }) {
	const api = useConsole();
	const people = useLoad(() => api.cahut("list_users", {}));
	const [handle, setHandle] = useState("");
	const [role, setRole] = useState("member");
	const [asking, setAsking] = useState(false);
	const inRoom = new Set(room.people.map((p) => p.handle));
	const candidates = (people.data?.users ?? []).filter((u) => !inRoom.has(u.handle) && u.status !== "disabled");
	if (candidates.length === 0) return null;
	const chosen = candidates.find((u) => u.handle === handle);
	return (
		<form
			class="form"
			onSubmit={(e) => {
				e.preventDefault();
				setAsking(true);
			}}
		>
			<label class="c-field">
				<span class="c-field__label">{t.addPerson}</span>
				<select class="c-input" required value={handle} onChange={(e) => setHandle(e.currentTarget.value)}>
					<option value="">{t.choosePerson}</option>
					{candidates.map((u) => (
						<option key={u.handle} value={u.handle}>
							{u.name} ({u.handle})
						</option>
					))}
				</select>
			</label>
			<label class="c-field">
				<span class="c-field__label">{t.as}</span>
				<select class="c-input" value={role} onChange={(e) => setRole(e.currentTarget.value)}>
					<option value="member">{t.memberRole}</option>
					<option value="owner">{t.owner}</option>
				</select>
			</label>
			{asking && chosen ? (
				<span class="confirm" role="group">
					<span class="c-small">{t.willOpen(chosen.name)}</span>
					<button
						class="c-btn c-btn--primary c-btn--sm"
						type="button"
						onClick={async () => {
							await onAdd(handle, role);
							setAsking(false);
							setHandle("");
						}}
					>
						{t.add}
					</button>
					<button class="c-btn c-btn--ghost c-btn--sm" type="button" onClick={() => setAsking(false)}>
						{t.cancel}
					</button>
				</span>
			) : (
				<div>
					<button class="c-btn c-btn--secondary c-btn--sm" type="submit">
						<Plus class="i" size={14} aria-hidden="true" />
						{t.add}
					</button>
				</div>
			)}
		</form>
	);
}

function Retitle({ room, onRetitle }: { room: RoomDetail; onRetitle: (title: string) => Promise<void> }) {
	const [title, setTitle] = useState(room.title);
	return (
		<form
			class="form"
			onSubmit={(e) => {
				e.preventDefault();
				onRetitle(title);
			}}
		>
			<label class="c-field">
				<span class="c-field__label">{t.title}</span>
				<input class="c-input" required value={title} onInput={(e) => setTitle(e.currentTarget.value)} />
				<span class="c-field__help">{t.addressStays}</span>
			</label>
			<div>
				<button class="c-btn c-btn--secondary c-btn--sm" type="submit" disabled={title.trim() === room.title}>
					{t.rename}
				</button>
			</div>
		</form>
	);
}

/**
 * What the address of a Room shows: the list of its Windows, or one of
 * them, for its readers who open it. See docs/adr/0032-the-page-of-a-room.md.
 */
function HomePage({ room, windows, onChange }: { room: RoomDetail; windows: Window[]; onChange: () => void }) {
	const api = useConsole();
	const current = room.home ?? "";
	const [home, setHome] = useState(current);
	const [busy, setBusy] = useState(false);
	const [error, setError] = useState("");
	const save = async (e: Event) => {
		e.preventDefault();
		setBusy(true);
		setError("");
		try {
			await api.cahut("set_room_home", { room: room.name, window: home || undefined });
			onChange();
		} catch (err) {
			setError(explain(err));
		}
		setBusy(false);
	};
	return (
		<section class="c-card panel aside">
			<h2 class="c-h4">{t.pageOfRoom}</h2>
			<p class="c-small c-muted">
				{t.homeText(
					<a class="c-link c-mono" href={`/${room.name}/`} target="_blank" rel="noopener">
						{location.host}/{room.name}/
					</a>,
				)}
			</p>
			<form class="form" onSubmit={save}>
				<label class="c-field">
					<span class="c-field__label">{t.show}</span>
					<select class="c-input" value={home} onChange={(e) => setHome(e.currentTarget.value)}>
						<option value="">{t.roomWindows}</option>
						{windows.map((w) => (
							<option key={w.name} value={w.name}>
								{w.title}
							</option>
						))}
					</select>
					<span class="c-field__help">
						{home ? t.homeHelp : t.listHelp}
					</span>
				</label>
				{error && <Failure message={error} />}
				<div>
					<button class="c-btn c-btn--secondary c-btn--sm" type="submit" disabled={busy || home === current}>
						{t.save}
					</button>
				</div>
			</form>
		</section>
	);
}

/** Deletes a team Room for good, once its name is typed. */
function DeleteRoom({ room }: { room: RoomDetail }) {
	const api = useConsole();
	const [open, setOpen] = useState(false);
	const [typed, setTyped] = useState("");
	const [busy, setBusy] = useState(false);
	const [error, setError] = useState("");
	if (!open) {
		return (
			<div>
				<button class="c-btn c-btn--secondary c-btn--sm" type="button" onClick={() => setOpen(true)}>
					{t.deleteRoom}
				</button>
			</div>
		);
	}
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
		<form class="delete-room" onSubmit={remove}>
			<p class="c-small">{t.deleteText(room.title, plural(room.windows, "Window", "Windows"))}</p>
			<label class="c-field">
				<span class="c-field__label">{t.typeToConfirm(room.name)}</span>
				<input class="c-input c-mono" autoComplete="off" spellcheck={false} value={typed} onInput={(e) => setTyped(e.currentTarget.value)} />
			</label>
			{error && <Failure message={error} />}
			<div class="actions">
				<button class="c-btn c-btn--danger c-btn--sm" type="submit" disabled={busy || typed.trim() !== room.name}>
					{t.deleteForGood}
				</button>
				<button class="c-btn c-btn--ghost c-btn--sm" type="button" onClick={() => setOpen(false)}>
					{t.cancel}
				</button>
			</div>
		</form>
	);
}
