// People, for the administrators of the Cahut: who is in it, a new person
// with the link where they choose their password, roles, and disabling.
import { Plus } from "lucide-preact";
import { useState } from "preact/hooks";
import { explain } from "../api";
import type { Activation, User } from "../api.gen";
import { ConfirmButton, CopyField, Failure, initials, Notice, PageHead } from "../components";
import { useConsole } from "../context";
import { useLoad } from "../load";
import { href } from "../router";

const statuses: Record<string, { word: string; badge: string }> = {
	active: { word: "Active", badge: "c-badge--success" },
	pending: { word: "No password yet", badge: "c-badge--warning" },
	disabled: { word: "Disabled", badge: "c-badge--neutral" },
};

/** The address where a person uses an activation code, at this console. */
function activationLink(a: Activation): string {
	return `${location.origin}${href("/activate")}#${a.code}`;
}

export function People() {
	const api = useConsole();
	const { data, error, reload } = useLoad(() => api.cahut("list_users", {}));
	const [link, setLink] = useState<{ name: string; email?: string; activation: Activation } | null>(null);
	const [failure, setFailure] = useState("");
	if (!api.isAdmin) return <Notice>Only the administrators of this Cahut manage its people.</Notice>;
	if (error) return <Failure message={error} />;
	if (!data) return null;

	const act = async (run: () => Promise<unknown>) => {
		setFailure("");
		try {
			await run();
			reload();
		} catch (err) {
			setFailure(explain(err));
		}
	};
	return (
		<>
			<PageHead label="This Cahut" title="People">
				Each person signs in with their own password, has a personal Room, and connects their AIs.
			</PageHead>
			<NewPerson onCreated={(u, a) => (reload(), setLink({ name: u.name, email: u.email, activation: a }))} />
			{link && <LinkPanel name={link.name} email={link.email} activation={link.activation} />}
			{failure && <Failure message={failure} />}
			<div class="table-wrap section">
				<table class="table">
					<thead>
						<tr>
							<th>Person</th>
							<th>Role</th>
							<th>State</th>
							<th />
						</tr>
					</thead>
					<tbody>
						{data.users.map((u) => (
							<Row
								key={u.handle}
								user={u}
								self={api.session.kind === "person" && u.handle === api.session.handle}
								mail={api.sendsMail}
								onRole={(role) => act(() => api.cahut("set_user_role", { user: u.handle, role }))}
								onDisable={() => act(() => api.cahut("disable_user", { user: u.handle }))}
								onEnable={() => act(() => api.cahut("enable_user", { user: u.handle }))}
								onEmail={(email) => act(() => api.cahut("set_user_email", { user: u.handle, email }))}
								onLink={() =>
									act(async () =>
										setLink({ name: u.name, email: u.email, activation: await api.cahut("create_activation_link", { user: u.handle }) }),
									)
								}
							/>
						))}
					</tbody>
				</table>
			</div>
		</>
	);
}

function Row(props: {
	user: User;
	self: boolean;
	/** Whether this Cahut sends e-mail, and so shows the address of each person. */
	mail: boolean;
	onRole: (role: string) => Promise<void>;
	onDisable: () => Promise<void>;
	onEnable: () => Promise<void>;
	onEmail: (email: string) => Promise<void>;
	onLink: () => Promise<void>;
}) {
	const u = props.user;
	const s = statuses[u.status] ?? statuses.active;
	return (
		<tr>
			<td>
				<span class="person">
					<span class="c-avatar" aria-hidden="true">
						{initials(u.name)}
					</span>
					<span class="people__who">
						<span>{u.name}</span>
						<span class="c-mono c-muted people__handle">{u.handle}</span>
					</span>
				</span>
				{props.mail && <EmailCell email={u.email} onSave={props.onEmail} />}
			</td>
			<td>{u.role === "admin" ? "Administrator" : "Member"}</td>
			<td>
				<span class={`c-badge ${s.badge}`}>{s.word}</span>
			</td>
			<td class="table__actions">
				{u.status !== "disabled" && (
					<ConfirmButton
						label="New link"
						question={`Give ${u.name} a new link? Their password works until they use it.`}
						confirm="Give a link"
						onConfirm={props.onLink}
					/>
				)}
				<ConfirmButton
					label={u.role === "admin" ? "Make member" : "Make administrator"}
					question={u.role === "admin" ? `Make ${u.name} a member?` : `Make ${u.name} an administrator of this Cahut?`}
					confirm="Yes"
					onConfirm={() => props.onRole(u.role === "admin" ? "member" : "admin")}
				/>
				{!props.self &&
					(u.status === "disabled" ? (
						<ConfirmButton label="Enable" question={`Enable ${u.name} again?`} confirm="Enable" onConfirm={props.onEnable} />
					) : (
						<ConfirmButton
							label="Disable"
							question={`Disable ${u.name}? Their sessions end and their AIs stop.`}
							confirm="Disable"
							danger
							onConfirm={props.onDisable}
						/>
					))}
			</td>
		</tr>
	);
}

function NewPerson({ onCreated }: { onCreated: (u: User, a: Activation) => void }) {
	const api = useConsole();
	const [handle, setHandle] = useState("");
	const [name, setName] = useState("");
	const [admin, setAdmin] = useState(false);
	const [email, setEmail] = useState("");
	const [error, setError] = useState("");
	const submit = async (e: Event) => {
		e.preventDefault();
		setError("");
		try {
			const u = await api.cahut("create_user", { handle, name, admin, email: api.sendsMail ? email : "" });
			if (u.activation) onCreated(u, u.activation);
			setHandle("");
			setName("");
			setEmail("");
			setAdmin(false);
		} catch (err) {
			setError(explain(err));
		}
	};
	return (
		<section class="section c-card panel">
			<h2 class="c-h4">New person</h2>
			<form class="form" onSubmit={submit}>
				<div class="form__row">
					<label class="c-field">
						<span class="c-field__label">Name</span>
						<input class="c-input" required placeholder="Ana Lima" value={name} onInput={(e) => setName(e.currentTarget.value)} />
					</label>
					<label class="c-field">
						<span class="c-field__label">Handle</span>
						<input class="c-input c-mono" required placeholder="ana" autocapitalize="none" value={handle} onInput={(e) => setHandle(e.currentTarget.value)} />
						<span class="c-field__help">It names their personal Room, at {location.hostname}/handle. It cannot change.</span>
					</label>
				</div>
				{api.sendsMail && (
					<label class="c-field">
						<span class="c-field__label">E-mail address</span>
						<input class="c-input" type="email" placeholder="ana@example.com" value={email} onInput={(e) => setEmail(e.currentTarget.value)} />
						<span class="c-field__help">Optional. Their link goes there, and so do their notices.</span>
					</label>
				)}
				<label class="c-check">
					<input type="checkbox" checked={admin} onChange={(e) => setAdmin(e.currentTarget.checked)} />
					Administrator of this Cahut
				</label>
				{error && <Failure message={error} />}
				<div>
					<button class="c-btn c-btn--primary" type="submit">
						<Plus class="i" size={16} aria-hidden="true" />
						Create the person
					</button>
				</div>
			</form>
		</section>
	);
}

function LinkPanel({ name, email, activation }: { name: string; email?: string; activation: Activation }) {
	const until = new Date(activation.expires_at).toLocaleDateString("en-GB", { dateStyle: "long" });
	return (
		<div class="result section" role="status">
			<p class="c-small">
				{activation.emailed
					? `The link where ${name} chooses their password went to ${email ?? "their address"}. It works once, until ${until}; here it is too, shown only now.`
					: `Send ${name} this link, where they choose their password. It works once, until ${until}, and is shown only now.`}
			</p>
			<CopyField label="the activation link" value={activationLink(activation)} />
		</div>
	);
}

/** The e-mail address of a person, which an administrator changes on the spot. */
function EmailCell({ email, onSave }: { email?: string; onSave: (email: string) => Promise<void> }) {
	const [editing, setEditing] = useState(false);
	const [value, setValue] = useState(email ?? "");
	if (!editing) {
		return (
			<span class="email-cell">
				<span class={email ? "c-mono c-small" : "c-small c-muted"}>{email ?? "None"}</span>
				<button class="c-btn c-btn--ghost c-btn--sm" type="button" onClick={() => (setValue(email ?? ""), setEditing(true))}>
					{email ? "Change" : "Add"}
				</button>
			</span>
		);
	}
	const save = async (e: Event) => {
		e.preventDefault();
		await onSave(value.trim());
		setEditing(false);
	};
	return (
		<form class="email-cell" onSubmit={save}>
			<input class="c-input" type="email" aria-label="E-mail address" value={value} onInput={(e) => setValue(e.currentTarget.value)} />
			<button class="c-btn c-btn--primary c-btn--sm" type="submit">
				Save
			</button>
			<button class="c-btn c-btn--ghost c-btn--sm" type="button" onClick={() => setEditing(false)}>
				Cancel
			</button>
		</form>
	);
}
