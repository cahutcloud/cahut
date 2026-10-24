// The administrators of the server, for themselves: who they are, and two
// dialogs. One adds an administrator and gives the link where they choose
// their password; the other manages one: a new link, disabling. They hold
// every Cahut of the server, unlike the administrators of a Cahut, whom the
// screen of its people makes.
import { Check, Plus } from "lucide-preact";
import { useState } from "preact/hooks";
import { explain } from "../api";
import type { Activation, Admin } from "../api.gen";
import { ActivationResult, Dialog, Failure, handleFrom, initials, Notice, PageHead, plural, Setting, StatusBadge } from "../components";
import { useConsole } from "../context";
import { useLoad } from "../load";
import { Link } from "../router";

/** The day an administrator was added, as in "24 Oct 2026". */
function added(a: Admin): string {
	return new Date(a.created_at).toLocaleDateString("en-GB", { day: "numeric", month: "short", year: "numeric" });
}

export function Admins() {
	const api = useConsole();
	const { data, error, reload } = useLoad(() => api.server("list_admins", {}));
	const [adding, setAdding] = useState(false);
	const [managing, setManaging] = useState("");
	if (error) return <Failure message={error} />;
	if (!data) return null;

	const admins = data.admins;
	const isSelf = (a: Admin) => api.session.kind === "admin" && a.handle === api.session.handle;
	const managed = admins.find((a) => a.handle === managing);
	const summary = [plural(admins.length, "administrator", "administrators")];
	const waiting = admins.filter((a) => a.status === "pending").length;
	const disabled = admins.filter((a) => a.status === "disabled").length;
	if (waiting) summary.push(`${waiting} without a password yet`);
	if (disabled) summary.push(`${disabled} disabled`);
	return (
		<>
			<PageHead
				label="This server"
				title="Server administrators"
				actions={
					<button class="c-btn c-btn--primary" type="button" onClick={() => setAdding(true)}>
						<Plus class="i" size={16} aria-hidden="true" />
						New administrator
					</button>
				}
			>
				They hold every Cahut of this server: they declare Cahuts, set their kind and limits, read the Logbook of the server, and manage
				the people and Rooms of every Cahut, at its console.
			</PageHead>
			<div class="section">
				<Notice>
					To make someone an administrator of this Cahut only, add them in <Link class="c-link" href="/people">People</Link> as an
					administrator.
				</Notice>
			</div>
			<section class="section">
				<p class="c-small c-muted people-summary">{summary.join(" · ")}</p>
				<div class="table-wrap">
					<table class="table people-table">
						<thead>
							<tr>
								<th>Administrator</th>
								<th class="people-table__wide">Added</th>
								<th class="people-table__wide">State</th>
								<th>
									<span class="visually-hidden">Actions</span>
								</th>
							</tr>
						</thead>
						<tbody>
							{admins.map((a) => (
								<tr key={a.handle}>
									<td>
										<span class="person">
											<span class="c-avatar" aria-hidden="true">
												{initials(a.name)}
											</span>
											<span class="people__who">
												<span class="people__name">
													{a.name}
													{isSelf(a) && <span class="c-badge c-badge--neutral">You</span>}
												</span>
												<span class="c-mono c-muted people__handle">{a.handle}</span>
												<span class="people-table__narrow">
													<StatusBadge status={a.status} />
												</span>
											</span>
										</span>
									</td>
									<td class="people-table__wide c-muted">{added(a)}</td>
									<td class="people-table__wide">
										<StatusBadge status={a.status} />
									</td>
									<td class="table__actions">
										{isSelf(a) ? (
											<Link class="c-btn c-btn--secondary c-btn--sm" href="/account">
												Your account
											</Link>
										) : (
											<button class="c-btn c-btn--secondary c-btn--sm" type="button" aria-label={`Manage ${a.name}`} onClick={() => setManaging(a.handle)}>
												Manage
											</button>
										)}
									</td>
								</tr>
							))}
						</tbody>
					</table>
				</div>
			</section>
			{adding && <NewAdmin onCreated={reload} onClose={() => setAdding(false)} />}
			{managed && <ManageAdmin admin={managed} onChanged={reload} onClose={() => setManaging("")} />}
		</>
	);
}

/** The dialog that adds an administrator of the server, then gives the link where they choose their password. */
function NewAdmin({ onCreated, onClose }: { onCreated: () => void; onClose: () => void }) {
	const api = useConsole();
	const [name, setName] = useState("");
	const [handle, setHandle] = useState("");
	const [handleTouched, setHandleTouched] = useState(false);
	const [error, setError] = useState("");
	const [busy, setBusy] = useState(false);
	const [created, setCreated] = useState<{ name: string; activation: Activation } | null>(null);

	const submit = async (e: Event) => {
		e.preventDefault();
		setError("");
		setBusy(true);
		try {
			const a = await api.server("create_admin", { handle, name });
			onCreated();
			setCreated({ name: a.name, activation: a.activation });
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
	};

	if (created) {
		return (
			<Dialog title="New administrator" onClose={onClose}>
				<div class="c-alert c-alert--success" role="status">
					<Check class="i" size={16} aria-hidden="true" />
					<div>
						<strong>{created.name} is an administrator of this server.</strong>
						They hold every Cahut of it once they have chosen their password.
					</div>
				</div>
				<ActivationResult name={created.name} activation={created.activation} page="/server/activate" />
				<div class="dialog__foot">
					<button class="c-btn c-btn--secondary" type="button" onClick={again}>
						Add another administrator
					</button>
					<button class="c-btn c-btn--primary" type="button" onClick={onClose}>
						Done
					</button>
				</div>
			</Dialog>
		);
	}
	return (
		<Dialog title="New administrator" onClose={onClose}>
			<p class="c-small c-muted dialog__intro">
				They will hold every Cahut of this server. An administrator of one Cahut only is added in its People, as an administrator.
			</p>
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
					<span class="c-field__help">Their name to sign in as an administrator of the server. It cannot change later.</span>
				</label>
				{error && <Failure message={error} />}
				<div class="dialog__foot">
					<button class="c-btn c-btn--secondary" type="button" onClick={onClose}>
						Cancel
					</button>
					<button class="c-btn c-btn--primary" type="submit" disabled={busy}>
						Add the administrator
					</button>
				</div>
			</form>
		</Dialog>
	);
}

/** The dialog that manages another administrator of the server: each change with what it does. */
function ManageAdmin(props: { admin: Admin; onChanged: () => void; onClose: () => void }) {
	const api = useConsole();
	const a = props.admin;
	const run = async <T,>(call: () => Promise<T>): Promise<T> => {
		const out = await call();
		props.onChanged();
		return out;
	};
	return (
		<Dialog
			title={a.name}
			meta={
				<>
					<span class="c-mono">{a.handle}</span>
					<span aria-hidden="true">·</span>
					<span>Added {added(a)}</span>
					<StatusBadge status={a.status} />
				</>
			}
			onClose={props.onClose}
		>
			<div class="settings">
				{a.status !== "disabled" && (
					<Setting
						key="link"
						title={a.status === "pending" ? "Link to choose a password" : "New password"}
						help={
							a.status === "pending"
								? `${a.name} has not chosen a password yet. A new link replaces the one they had.`
								: `If ${a.name} forgot their password, give them a link where they choose a new one. Their password works until they use it.`
						}
						action="Give a new link"
						onAct={async () => {
							const link = await run(() => api.server("create_activation_link", { admin: a.handle }));
							return <ActivationResult name={a.name} activation={link} page="/server/activate" />;
						}}
					/>
				)}
				{a.status === "disabled" ? (
					<Setting
						key="state"
						title="Enable"
						help={`${a.name} cannot sign in until an administrator of the server enables them. They sign in again with their password, or a new link.`}
						action="Enable"
						onAct={async () => {
							const b = await run(() => api.server("enable_admin", { admin: a.handle }));
							return b.status === "pending" ? `${a.name} is enabled, and has no password yet: give them a new link.` : `${a.name} can sign in again.`;
						}}
					/>
				) : (
					<Setting
						key="state"
						title="Disable"
						help="Their sessions and their link end, and they cannot sign in until an administrator of the server enables them again. The server keeps at least one administrator who is not disabled."
						action="Disable"
						danger
						ask={`Disable ${a.name}? Their sessions end now.`}
						onAct={async () => {
							await run(() => api.server("disable_admin", { admin: a.handle }));
							return `${a.name} is disabled.`;
						}}
					/>
				)}
			</div>
		</Dialog>
	);
}
