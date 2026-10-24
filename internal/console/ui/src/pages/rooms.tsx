// The Rooms: those of the person, every Room for an administrator, a new
// team Room, and the page of one Room with its Windows, its people and
// what its address shows.
import { Folder, Plus, User, Users } from "lucide-preact";
import { useState } from "preact/hooks";
import { ApiError, explain } from "../api";
import type { Room, RoomDetail, Window } from "../api.gen";
import { ConfirmButton, Failure, initials, Notice, PageHead, plural, WindowCards } from "../components";
import { useConsole } from "../context";
import { useLoad } from "../load";
import { Link, navigate } from "../router";
import { Trash } from "./window";

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
								{plural(r.windows, "Window", "Windows")} · {plural(r.members, "member", "members")}
							</span>
							{r.role && <span class="c-badge c-badge--neutral">{r.role === "owner" ? "Owner" : "Member"}</span>}
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
				{api.individual
					? "A Room holds the Windows of its owner, at its own address. In this Cahut, each person works alone: a Room keeps one member."
					: "A Room gathers its people and their Windows, at its own address. Windows published in a team Room are for its members."}
			</PageHead>
			{mine.rooms.length > 0 && (
				<section class="section">
					<h2 class="c-h4">Your Rooms</h2>
					<RoomCards rooms={mine.rooms} />
				</section>
			)}
			{all && (
				<section class="section">
					<h2 class="c-h4">Every Room of this Cahut</h2>
					<p class="c-small c-muted">As an administrator, you manage the people of every Room, and open the Windows of yours only.</p>
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
			<h2 class="c-h4">{api.individual ? "New Room" : "New team Room"}</h2>
			<form class="form" onSubmit={submit}>
				<div class="form__row">
					<label class="c-field">
						<span class="c-field__label">Title</span>
						<input class="c-input" required placeholder="Marketing" value={title} onInput={(e) => setTitle(e.currentTarget.value)} />
					</label>
					<label class="c-field">
						<span class="c-field__label">Name, in its address</span>
						<input class="c-input c-mono" placeholder="marketing" autocapitalize="none" value={name} onInput={(e) => setName(e.currentTarget.value)} />
						<span class="c-field__help">
							{api.individual
								? `From the title when empty, after ${api.session.kind === "admin" ? "the handle of its owner" : `${api.session.handle}-`}. It cannot change later.`
								: "From the title when empty. It cannot change later."}
						</span>
					</label>
				</div>
				{people.data && (
					<label class="c-field">
						<span class="c-field__label">Owner</span>
						<select class="c-input" required value={owner} onChange={(e) => setOwner(e.currentTarget.value)}>
							<option value="">Choose a person</option>
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
						Create the Room
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
			<PageHead label={r.kind === "personal" ? "Personal Room" : "Room"} title={r.title}>
				{plural(r.windows, "Window", "Windows")}, {plural(r.members, "member", "members")} ·{" "}
				<span class="c-mono">
					{location.hostname}/{r.name}/
				</span>
			</PageHead>
			<div class="layout">
				<section>
					<h2 class="c-h4">Windows</h2>
					{windows.error && <Failure message={windows.error} />}
					{windows.data === null && <Notice>You are not in this Room. As an administrator, you manage its people, not its Windows.</Notice>}
					{windows.data && windows.data.windows.length === 0 && (
						<Notice>No Window in this Room yet. {api.individual ? "An AI of its owner publishes them." : "An AI of one of its members publishes them."}</Notice>
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
			<h2 class="c-h4">People</h2>
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
						<span class="c-badge c-badge--neutral">{p.role === "owner" ? "Owner" : "Member"}</span>
						{manage && members && p.handle !== me && (
							<span class="people__actions">
								{(p.role === "member" || owners > 1) && (
									<ConfirmButton
										label={p.role === "owner" ? "Make member" : "Make owner"}
										question={`Make ${p.name} ${p.role === "owner" ? "a member" : "an owner"}?`}
										confirm="Yes"
										onConfirm={() =>
											act(() =>
												api.cahut("set_room_member_role", { room: room.name, user: p.handle, role: p.role === "owner" ? "member" : "owner" }),
											)
										}
									/>
								)}
								{(p.role === "member" || owners > 1) && (
									<ConfirmButton
										label="Remove"
										question={`Remove ${p.name} from this Room?`}
										confirm="Remove"
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
					label="Leave this Room"
					question="Leave this Room? You will no longer open its Windows."
					confirm="Leave"
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
				<span class="c-field__label">Add a person</span>
				<select class="c-input" required value={handle} onChange={(e) => setHandle(e.currentTarget.value)}>
					<option value="">Choose a person</option>
					{candidates.map((u) => (
						<option key={u.handle} value={u.handle}>
							{u.name} ({u.handle})
						</option>
					))}
				</select>
			</label>
			<label class="c-field">
				<span class="c-field__label">As</span>
				<select class="c-input" value={role} onChange={(e) => setRole(e.currentTarget.value)}>
					<option value="member">Member</option>
					<option value="owner">Owner</option>
				</select>
			</label>
			{asking && chosen ? (
				<span class="confirm" role="group">
					<span class="c-small">
						{chosen.name} will open and publish the Windows of this Room.
					</span>
					<button
						class="c-btn c-btn--primary c-btn--sm"
						type="button"
						onClick={async () => {
							await onAdd(handle, role);
							setAsking(false);
							setHandle("");
						}}
					>
						Add
					</button>
					<button class="c-btn c-btn--ghost c-btn--sm" type="button" onClick={() => setAsking(false)}>
						Cancel
					</button>
				</span>
			) : (
				<div>
					<button class="c-btn c-btn--secondary c-btn--sm" type="submit">
						<Plus class="i" size={14} aria-hidden="true" />
						Add
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
				<span class="c-field__label">Title</span>
				<input class="c-input" required value={title} onInput={(e) => setTitle(e.currentTarget.value)} />
				<span class="c-field__help">The address of the Room stays.</span>
			</label>
			<div>
				<button class="c-btn c-btn--secondary c-btn--sm" type="submit" disabled={title.trim() === room.title}>
					Rename
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
			<h2 class="c-h4">Page of the Room</h2>
			<p class="c-small c-muted">
				What{" "}
				<a class="c-link c-mono" href={`/${room.name}/`} target="_blank" rel="noopener">
					{location.host}/{room.name}/
				</a>{" "}
				shows to its readers.
			</p>
			<form class="form" onSubmit={save}>
				<label class="c-field">
					<span class="c-field__label">Show</span>
					<select class="c-input" value={home} onChange={(e) => setHome(e.currentTarget.value)}>
						<option value="">The Windows of the Room</option>
						{windows.map((w) => (
							<option key={w.name} value={w.name}>
								{w.title}
							</option>
						))}
					</select>
					<span class="c-field__help">
						{home
							? "Readers who may not open this Window see the list of the Windows they open."
							: "Each reader sees the Windows they may open, with a preview of each."}
					</span>
				</label>
				{error && <Failure message={error} />}
				<div>
					<button class="c-btn c-btn--secondary c-btn--sm" type="submit" disabled={busy || home === current}>
						Save
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
					Delete this Room
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
			<p class="c-small">
				This deletes the Room {room.title} for good: its people, {plural(room.windows, "Window", "Windows")}, those in its trash, and
				their files. It cannot be undone.
			</p>
			<label class="c-field">
				<span class="c-field__label">
					Type <span class="c-mono">{room.name}</span> to confirm
				</span>
				<input class="c-input c-mono" autoComplete="off" spellcheck={false} value={typed} onInput={(e) => setTyped(e.currentTarget.value)} />
			</label>
			{error && <Failure message={error} />}
			<div class="actions">
				<button class="c-btn c-btn--danger c-btn--sm" type="submit" disabled={busy || typed.trim() !== room.name}>
					Delete for good
				</button>
				<button class="c-btn c-btn--ghost c-btn--sm" type="button" onClick={() => setOpen(false)}>
					Cancel
				</button>
			</div>
		</form>
	);
}
