// People, for the administrators of the Cahut: who is in it, and two
// dialogs. One adds a person and gives the link where they choose their
// password; the other manages a person: their e-mail address, their role, a
// new link, disabling.
import { Check, Plus } from "lucide-preact";
import { useId, useState } from "preact/hooks";
import { explain } from "../api";
import type { Activation, User } from "../api.gen";
import { ActivationResult, Dialog, Failure, handleFrom, initials, Notice, Outcome, PageHead, plural, Setting, StatusBadge } from "../components";
import { useConsole } from "../context";
import { useLoad } from "../load";
import { Link } from "../router";

const roles: Record<string, { word: string; help: string }> = {
	member: { word: "Member", help: "Works in their Rooms and connects their AIs." },
	admin: { word: "Administrator", help: "Also manages the people, the Rooms and the settings of this Cahut." },
};

function role(u: User) {
	return roles[u.role] ?? roles.member;
}

export function People() {
	const api = useConsole();
	const { data, error, reload } = useLoad(() => api.cahut("list_users", {}));
	const [adding, setAdding] = useState(false);
	const [managing, setManaging] = useState("");
	if (!api.isAdmin) return <Notice>Only the administrators of this Cahut manage its people.</Notice>;
	if (error) return <Failure message={error} />;
	if (!data) return null;

	const users = data.users;
	const isSelf = (u: User) => api.session.kind === "person" && u.handle === api.session.handle;
	const managed = users.find((u) => u.handle === managing);
	const count = (test: (u: User) => boolean) => users.filter(test).length;
	const summary = [plural(users.length, "person", "people"), plural(count((u) => u.role === "admin"), "administrator", "administrators")];
	const waiting = count((u) => u.status === "pending");
	const disabled = count((u) => u.status === "disabled");
	if (waiting) summary.push(`${waiting} without a password yet`);
	if (disabled) summary.push(`${disabled} disabled`);
	return (
		<>
			<PageHead
				label="This Cahut"
				title="People"
				actions={
					<button class="c-btn c-btn--primary" type="button" onClick={() => setAdding(true)}>
						<Plus class="i" size={16} aria-hidden="true" />
						New person
					</button>
				}
			>
				Each person signs in with their own password, has a personal Room, and connects their AIs.
			</PageHead>
			<section class="section">
				<p class="c-small c-muted people-summary">{summary.join(" · ")}</p>
				<div class="table-wrap">
					<table class="table people-table">
						<thead>
							<tr>
								<th>Person</th>
								{api.sendsMail && <th class="people-table__wide">E-mail address</th>}
								<th class="people-table__wide">Role</th>
								<th class="people-table__wide">State</th>
								<th>
									<span class="visually-hidden">Actions</span>
								</th>
							</tr>
						</thead>
						<tbody>
							{users.map((u) => (
								<tr key={u.handle}>
									<td>
										<span class="person">
											<span class="c-avatar" aria-hidden="true">
												{initials(u.name)}
											</span>
											<span class="people__who">
												<span class="people__name">
													{u.name}
													{isSelf(u) && <span class="c-badge c-badge--neutral">You</span>}
												</span>
												<span class="c-mono c-muted people__handle">{u.handle}</span>
												<span class="people-table__narrow">
													<span class="c-small c-muted">{role(u).word}</span>
													<StatusBadge status={u.status} />
												</span>
											</span>
										</span>
									</td>
									{api.sendsMail && (
										<td class="people-table__wide">
											{u.email ? <span class="c-mono">{u.email}</span> : <span class="c-muted">No address</span>}
										</td>
									)}
									<td class="people-table__wide">{role(u).word}</td>
									<td class="people-table__wide">
										<StatusBadge status={u.status} />
									</td>
									<td class="table__actions">
										<button class="c-btn c-btn--secondary c-btn--sm" type="button" aria-label={`Manage ${u.name}`} onClick={() => setManaging(u.handle)}>
											Manage
										</button>
									</td>
								</tr>
							))}
						</tbody>
					</table>
				</div>
			</section>
			{adding && <NewPerson onCreated={reload} onClose={() => setAdding(false)} />}
			{managed && <ManagePerson user={managed} self={isSelf(managed)} onChanged={reload} onClose={() => setManaging("")} />}
		</>
	);
}

/** The dialog that adds a person, then gives the link where they choose their password. */
function NewPerson({ onCreated, onClose }: { onCreated: () => void; onClose: () => void }) {
	const api = useConsole();
	const [name, setName] = useState("");
	const [handle, setHandle] = useState("");
	const [handleTouched, setHandleTouched] = useState(false);
	const [email, setEmail] = useState("");
	const [admin, setAdmin] = useState(false);
	const [error, setError] = useState("");
	const [busy, setBusy] = useState(false);
	const [created, setCreated] = useState<{ user: User; activation: Activation } | null>(null);

	const submit = async (e: Event) => {
		e.preventDefault();
		setError("");
		setBusy(true);
		try {
			const u = await api.cahut("create_user", { handle, name, admin, email: api.sendsMail ? email.trim() : "" });
			onCreated();
			if (u.activation) setCreated({ user: u, activation: u.activation });
			else onClose();
		} catch (err) {
			setError(explain(err));
		} finally {
			setBusy(false);
		}
	};
	const again = () => {
		setCreated(null);
		setName("");
		setHandle("");
		setHandleTouched(false);
		setEmail("");
		setAdmin(false);
	};

	if (created) {
		return (
			<Dialog title="New person" onClose={onClose}>
				<div class="c-alert c-alert--success" role="status">
					<Check class="i" size={16} aria-hidden="true" />
					<div>
						<strong>{created.user.name} is in this Cahut.</strong>
						Their personal Room is at <span class="c-mono">{`${location.hostname}/${created.user.handle}`}</span>.
					</div>
				</div>
				<ActivationResult name={created.user.name} email={created.user.email} activation={created.activation} page="/activate" />
				<div class="dialog__foot">
					<button class="c-btn c-btn--secondary" type="button" onClick={again}>
						Add another person
					</button>
					<button class="c-btn c-btn--primary" type="button" onClick={onClose}>
						Done
					</button>
				</div>
			</Dialog>
		);
	}
	return (
		<Dialog title="New person" onClose={onClose}>
			<form class="form" onSubmit={submit}>
				<label class="c-field">
					<span class="c-field__label">Name</span>
					<input
						class="c-input"
						required
						autofocus
						placeholder="Ana Lima"
						value={name}
						onInput={(e) => {
							setName(e.currentTarget.value);
							if (!handleTouched) setHandle(handleFrom(e.currentTarget.value));
						}}
					/>
				</label>
				<label class="c-field">
					<span class="c-field__label">Handle</span>
					<input
						class="c-input c-mono"
						required
						placeholder="ana"
						autocapitalize="none"
						spellcheck={false}
						value={handle}
						onInput={(e) => {
							setHandle(e.currentTarget.value);
							setHandleTouched(true);
						}}
					/>
					<span class="c-field__help">
						Their name to sign in, and the address of their personal Room:{" "}
						<span class="c-mono">{`${location.hostname}/${handle || "handle"}`}</span>. It cannot change later.
					</span>
				</label>
				{api.sendsMail && (
					<label class="c-field">
						<span class="c-field__label">E-mail address</span>
						<input class="c-input" type="email" placeholder="ana@example.com" value={email} onInput={(e) => setEmail(e.currentTarget.value)} />
						<span class="c-field__help">Optional. The link where they choose their password goes there, and so do their notices.</span>
					</label>
				)}
				<RoleChoice admin={admin} onChange={setAdmin} />
				{error && <Failure message={error} />}
				<div class="dialog__foot">
					<button class="c-btn c-btn--secondary" type="button" onClick={onClose}>
						Cancel
					</button>
					<button class="c-btn c-btn--primary" type="submit" disabled={busy}>
						Add the person
					</button>
				</div>
			</form>
		</Dialog>
	);
}

/** The role of a new person, each with what it lets them do. */
function RoleChoice({ admin, onChange }: { admin: boolean; onChange: (admin: boolean) => void }) {
	const name = useId();
	return (
		<fieldset class="choices">
			<legend class="c-field__label">Role</legend>
			{(["member", "admin"] as const).map((r) => (
				<label key={r} class="choice">
					<input type="radio" name={name} checked={admin === (r === "admin")} onChange={() => onChange(r === "admin")} />
					<span class="choice__name">{roles[r].word}</span>
					<span class="choice__help">{roles[r].help}</span>
				</label>
			))}
		</fieldset>
	);
}

/** The dialog that manages a person: each change with what it does. */
function ManagePerson(props: { user: User; self: boolean; onChanged: () => void; onClose: () => void }) {
	const api = useConsole();
	const u = props.user;
	const run = async <T,>(call: () => Promise<T>): Promise<T> => {
		const out = await call();
		props.onChanged();
		return out;
	};
	const toAdmin = u.role !== "admin";
	return (
		<Dialog
			title={u.name}
			meta={
				<>
					<span class="c-mono">{u.handle}</span>
					<span aria-hidden="true">·</span>
					<span>{role(u).word}</span>
					<StatusBadge status={u.status} />
				</>
			}
			onClose={props.onClose}
		>
			<div class="settings">
				{api.sendsMail && (
					<EmailSetting key="email" user={u} onSave={(email) => run(() => api.cahut("set_user_email", { user: u.handle, email }))} />
				)}
				<Setting
					key="role"
					title="Role"
					help={`${role(u).word}. ${role(u).help}`}
					action={toAdmin ? "Make administrator" : "Make member"}
					ask={
						toAdmin
							? `Make ${u.name} an administrator? They will manage the people, the Rooms and the settings of this Cahut.`
							: props.self
								? "Make yourself a member? You will no longer manage this Cahut."
								: `Make ${u.name} a member? They will no longer manage this Cahut.`
					}
					onAct={async () => {
						await run(() => api.cahut("set_user_role", { user: u.handle, role: toAdmin ? "admin" : "member" }));
						return `${props.self ? "You are" : `${u.name} is`} now ${toAdmin ? "an administrator" : "a member"}.`;
					}}
				/>
				{!props.self && u.status !== "disabled" && (
					<Setting
						key="link"
						title={u.status === "pending" ? "Link to choose a password" : "New password"}
						help={
							u.status === "pending"
								? `${u.name} has not chosen a password yet. A new link replaces the one they had.`
								: `If ${u.name} forgot their password, give them a link where they choose a new one. Their password works until they use it.`
						}
						action="Give a new link"
						onAct={async () => {
							const a = await run(() => api.cahut("create_activation_link", { user: u.handle }));
							return <ActivationResult name={u.name} email={u.email} activation={a} page="/activate" />;
						}}
					/>
				)}
				{!props.self &&
					(u.status === "disabled" ? (
						<Setting
							key="state"
							title="Enable"
							help={`${u.name} cannot sign in, and their AIs cannot act, until an administrator enables them.`}
							action="Enable"
							onAct={async () => {
								const v = await run(() => api.cahut("enable_user", { user: u.handle }));
								return v.status === "pending" ? `${u.name} is enabled, and has no password yet: give them a new link.` : `${u.name} can sign in again.`;
							}}
						/>
					) : (
						<Setting
							key="state"
							title="Disable"
							help="Their sessions end, and neither they nor their AIs can act until an administrator enables them again. Their Rooms and Windows stay."
							action="Disable"
							danger
							ask={`Disable ${u.name}? Their sessions end now.`}
							onAct={async () => {
								await run(() => api.cahut("disable_user", { user: u.handle }));
								return `${u.name} is disabled.`;
							}}
						/>
					))}
			</div>
			{props.self && (
				<p class="c-small c-muted dialog__note">
					Your password changes in{" "}
					<Link class="c-link" href="/account">
						your account
					</Link>
					.
				</p>
			)}
		</Dialog>
	);
}

/** The e-mail address of a person, where their links and notices go. */
function EmailSetting({ user, onSave }: { user: User; onSave: (email: string) => Promise<unknown> }) {
	const id = useId();
	const [value, setValue] = useState(user.email ?? "");
	const [busy, setBusy] = useState(false);
	const [said, setSaid] = useState("");
	const [error, setError] = useState("");
	const email = value.trim();
	const save = async (e: Event) => {
		e.preventDefault();
		setBusy(true);
		setError("");
		setSaid("");
		try {
			await onSave(email);
			setSaid(email ? "E-mail address saved." : "E-mail address removed.");
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
					E-mail address
				</label>
				<p class="c-small c-muted">Their links and notices go there. Leave it empty to remove it.</p>
			</div>
			<div class="setting__more setting__inline">
				<input
					id={id}
					class="c-input"
					type="email"
					value={value}
					onInput={(e) => (setValue(e.currentTarget.value), setSaid(""))}
				/>
				<button class="c-btn c-btn--secondary" type="submit" disabled={busy || email === (user.email ?? "")}>
					Save
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
