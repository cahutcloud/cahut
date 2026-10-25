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
import { de, locale, words } from "../i18n";
import { useLoad } from "../load";
import { Link } from "../router";

const t = words({
	en: {
		administrator: "administrator",
		administrators: "administrators",
		withoutPassword: (n: number) => `${n} without a password yet`,
		disabledCount: (n: number) => `${n} disabled`,
		thisServer: "This server",
		title: "Server administrators",
		newAdmin: "New administrator",
		intro:
			"They hold every Cahut of this server: they declare Cahuts, set their kind and limits, read the Logbook of the server, and manage the people and Rooms of every Cahut, at its console.",
		cahutOnly: () => (
			<>
				To make someone an administrator of this Cahut only, add them in{" "}
				<Link class="c-link" href="/people">
					People
				</Link>{" "}
				as an administrator.
			</>
		),
		adminColumn: "Administrator",
		added: "Added",
		state: "State",
		actions: "Actions",
		you: "You",
		yourAccount: "Your account",
		manageLabel: (name: string) => `Manage ${name}`,
		manage: "Manage",
		created: (name: string) => (
			<>
				<strong>{name} is an administrator of this server.</strong>
				They hold every Cahut of it once they have chosen their password.
			</>
		),
		addAnother: "Add another administrator",
		done: "Done",
		dialogIntro: "They will hold every Cahut of this server. An administrator of one Cahut only is added in its People, as an administrator.",
		name: "Name",
		handle: "Handle",
		handleHelp: "Their name to sign in as an administrator of the server. It cannot change later.",
		cancel: "Cancel",
		add: "Add the administrator",
		addedOn: (day: string) => `Added ${day}`,
		choosePassword: "Link to choose a password",
		newPassword: "New password",
		pendingHelp: (name: string) => `${name} has not chosen a password yet. A new link replaces the one they had.`,
		forgotHelp: (name: string) =>
			`If ${name} forgot their password, give them a link where they choose a new one. Their password works until they use it.`,
		newLink: "Give a new link",
		enable: "Enable",
		enableHelp: (name: string) =>
			`${name} cannot sign in until an administrator of the server enables them. They sign in again with their password, or a new link.`,
		enabledPending: (name: string) => `${name} is enabled, and has no password yet: give them a new link.`,
		enabled: (name: string) => `${name} can sign in again.`,
		disable: "Disable",
		disableHelp:
			"Their sessions and their link end, and they cannot sign in until an administrator of the server enables them again. The server keeps at least one administrator who is not disabled.",
		askDisable: (name: string) => `Disable ${name}? Their sessions end now.`,
		disabled: (name: string) => `${name} is disabled.`,
	},
	fr: {
		administrator: "administrateur",
		administrators: "administrateurs",
		withoutPassword: (n) => `${n} encore sans mot de passe`,
		disabledCount: (n) => `${n} ${n > 1 ? "désactivés" : "désactivé"}`,
		thisServer: "Ce serveur",
		title: "Administrateurs du serveur",
		newAdmin: "Nouvel administrateur",
		intro:
			"Ils ont la main sur tous les Cahuts de ce serveur : ils déclarent des Cahuts, fixent leur type et leurs limites, lisent le Logbook du serveur, et gèrent les personnes et les Rooms de chaque Cahut, depuis sa console.",
		cahutOnly: () => (
			<>
				Pour faire de quelqu'un un administrateur de ce Cahut seulement, ajoutez cette personne dans{" "}
				<Link class="c-link" href="/people">
					Personnes
				</Link>
				, comme administrateur.
			</>
		),
		adminColumn: "Administrateur",
		added: "Ajouté le",
		state: "État",
		actions: "Actions",
		you: "Vous",
		yourAccount: "Votre compte",
		manageLabel: (name) => `Gérer ${name}`,
		manage: "Gérer",
		created: (name) => (
			<>
				<strong>{name} est administrateur de ce serveur.</strong>
				Cette personne a la main sur tous ses Cahuts dès qu'elle a choisi son mot de passe.
			</>
		),
		addAnother: "Ajouter un autre administrateur",
		done: "Terminé",
		dialogIntro:
			"Cette personne aura la main sur tous les Cahuts de ce serveur. Un administrateur d'un seul Cahut s'ajoute dans ses Personnes, comme administrateur.",
		name: "Nom",
		handle: "Identifiant",
		handleHelp: "Son nom pour se connecter comme administrateur du serveur. Il ne pourra plus changer.",
		cancel: "Annuler",
		add: "Ajouter l'administrateur",
		addedOn: (day) => `Ajouté le ${day}`,
		choosePassword: "Lien pour choisir un mot de passe",
		newPassword: "Nouveau mot de passe",
		pendingHelp: (name) => `${name} n'a pas encore choisi de mot de passe. Un nouveau lien remplace l'ancien.`,
		forgotHelp: (name) =>
			`Si ${name} a oublié son mot de passe, donnez-lui un lien pour en choisir un nouveau. Son mot de passe actuel reste valable jusqu'à l'usage du lien.`,
		newLink: "Donner un nouveau lien",
		enable: "Réactiver",
		enableHelp: (name) =>
			`${name} ne peut pas se connecter tant qu'un administrateur du serveur ne réactive pas son compte. Cette personne se reconnecte ensuite avec son mot de passe, ou un nouveau lien.`,
		enabledPending: (name) => `Le compte ${de(name)} est réactivé, encore sans mot de passe : donnez-lui un nouveau lien.`,
		enabled: (name) => `${name} peut de nouveau se connecter.`,
		disable: "Désactiver",
		disableHelp:
			"Ses sessions et son lien prennent fin, et cette personne ne peut plus se connecter tant qu'un administrateur du serveur ne réactive pas son compte. Le serveur garde toujours au moins un administrateur actif.",
		askDisable: (name) => `Désactiver ${name} ? Ses sessions prennent fin tout de suite.`,
		disabled: (name) => `Le compte ${de(name)} est désactivé.`,
	},
});

/** The day an administrator was added, as in "24 Oct 2026". */
function added(a: Admin): string {
	return new Date(a.created_at).toLocaleDateString(locale, { day: "numeric", month: "short", year: "numeric" });
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
	const summary = [plural(admins.length, t.administrator, t.administrators)];
	const waiting = admins.filter((a) => a.status === "pending").length;
	const disabled = admins.filter((a) => a.status === "disabled").length;
	if (waiting) summary.push(t.withoutPassword(waiting));
	if (disabled) summary.push(t.disabledCount(disabled));
	return (
		<>
			<PageHead
				label={t.thisServer}
				title={t.title}
				actions={
					<button class="c-btn c-btn--primary" type="button" onClick={() => setAdding(true)}>
						<Plus class="i" size={16} aria-hidden="true" />
						{t.newAdmin}
					</button>
				}
			>
				{t.intro}
			</PageHead>
			<div class="section">
				<Notice>{t.cahutOnly()}</Notice>
			</div>
			<section class="section">
				<p class="c-small c-muted people-summary">{summary.join(" · ")}</p>
				<div class="table-wrap">
					<table class="table people-table">
						<thead>
							<tr>
								<th>{t.adminColumn}</th>
								<th class="people-table__wide">{t.added}</th>
								<th class="people-table__wide">{t.state}</th>
								<th>
									<span class="visually-hidden">{t.actions}</span>
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
													{isSelf(a) && <span class="c-badge c-badge--neutral">{t.you}</span>}
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
												{t.yourAccount}
											</Link>
										) : (
											<button class="c-btn c-btn--secondary c-btn--sm" type="button" aria-label={t.manageLabel(a.name)} onClick={() => setManaging(a.handle)}>
												{t.manage}
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
			<Dialog title={t.newAdmin} onClose={onClose}>
				<div class="c-alert c-alert--success" role="status">
					<Check class="i" size={16} aria-hidden="true" />
					<div>{t.created(created.name)}</div>
				</div>
				<ActivationResult name={created.name} activation={created.activation} page="/server/activate" />
				<div class="dialog__foot">
					<button class="c-btn c-btn--secondary" type="button" onClick={again}>
						{t.addAnother}
					</button>
					<button class="c-btn c-btn--primary" type="button" onClick={onClose}>
						{t.done}
					</button>
				</div>
			</Dialog>
		);
	}
	return (
		<Dialog title={t.newAdmin} onClose={onClose}>
			<p class="c-small c-muted dialog__intro">{t.dialogIntro}</p>
			<form class="form" onSubmit={submit}>
				<label class="c-field">
					<span class="c-field__label">{t.name}</span>
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
					<span class="c-field__label">{t.handle}</span>
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
					<span class="c-field__help">{t.handleHelp}</span>
				</label>
				{error && <Failure message={error} />}
				<div class="dialog__foot">
					<button class="c-btn c-btn--secondary" type="button" onClick={onClose}>
						{t.cancel}
					</button>
					<button class="c-btn c-btn--primary" type="submit" disabled={busy}>
						{t.add}
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
					<span>{t.addedOn(added(a))}</span>
					<StatusBadge status={a.status} />
				</>
			}
			onClose={props.onClose}
		>
			<div class="settings">
				{a.status !== "disabled" && (
					<Setting
						key="link"
						title={a.status === "pending" ? t.choosePassword : t.newPassword}
						help={a.status === "pending" ? t.pendingHelp(a.name) : t.forgotHelp(a.name)}
						action={t.newLink}
						onAct={async () => {
							const link = await run(() => api.server("create_activation_link", { admin: a.handle }));
							return <ActivationResult name={a.name} activation={link} page="/server/activate" />;
						}}
					/>
				)}
				{a.status === "disabled" ? (
					<Setting
						key="state"
						title={t.enable}
						help={t.enableHelp(a.name)}
						action={t.enable}
						onAct={async () => {
							const b = await run(() => api.server("enable_admin", { admin: a.handle }));
							return b.status === "pending" ? t.enabledPending(a.name) : t.enabled(a.name);
						}}
					/>
				) : (
					<Setting
						key="state"
						title={t.disable}
						help={t.disableHelp}
						action={t.disable}
						danger
						ask={t.askDisable(a.name)}
						onAct={async () => {
							await run(() => api.server("disable_admin", { admin: a.handle }));
							return t.disabled(a.name);
						}}
					/>
				)}
			</div>
		</Dialog>
	);
}
