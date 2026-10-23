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
	const [link, setLink] = useState<{ name: string; activation: Activation } | null>(null);
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
			<NewPerson onCreated={(u, a) => (reload(), setLink({ name: u.name, activation: a }))} />
			{link && <LinkPanel name={link.name} activation={link.activation} />}
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
								onRole={(role) => act(() => api.cahut("set_user_role", { user: u.handle, role }))}
								onDisable={() => act(() => api.cahut("disable_user", { user: u.handle }))}
								onEnable={() => act(() => api.cahut("enable_user", { user: u.handle }))}
								onLink={() =>
									act(async () => setLink({ name: u.name, activation: await api.cahut("create_activation_link", { user: u.handle }) }))
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
	onRole: (role: string) => Promise<void>;
	onDisable: () => Promise<void>;
	onEnable: () => Promise<void>;
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
	const [error, setError] = useState("");
	const submit = async (e: Event) => {
		e.preventDefault();
		setError("");
		try {
			const u = await api.cahut("create_user", { handle, name, admin });
			if (u.activation) onCreated(u, u.activation);
			setHandle("");
			setName("");
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

function LinkPanel({ name, activation }: { name: string; activation: Activation }) {
	return (
		<div class="result section" role="status">
			<p class="c-small">
				Send {name} this link, where they choose their password. It works once, until{" "}
				{new Date(activation.expires_at).toLocaleDateString("en-GB", { dateStyle: "long" })}, and is shown only now.
			</p>
			<CopyField label="the activation link" value={activationLink(activation)} />
		</div>
	);
}
