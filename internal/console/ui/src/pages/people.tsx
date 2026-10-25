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
import { de, words } from "../i18n";
import { useLoad } from "../load";
import { Link } from "../router";

const t = words({
	en: {
		member: "Member",
		memberHelp: "Works in their Rooms and connects their AIs.",
		admin: "Administrator",
		adminHelp: "Also manages the people, the Rooms and the settings of this Cahut.",
		onlyAdmins: "Only the administrators of this Cahut manage its people.",
		person: "person",
		people: "people",
		administrator: "administrator",
		administrators: "administrators",
		withoutPassword: (n: number) => `${n} without a password yet`,
		disabledCount: (n: number) => `${n} disabled`,
		thisCahut: "This Cahut",
		title: "People",
		newPerson: "New person",
		intro: "Each person signs in with their own password, has a personal Room, and connects their AIs.",
		personColumn: "Person",
		email: "E-mail address",
		role: "Role",
		state: "State",
		actions: "Actions",
		you: "You",
		noAddress: "No address",
		manageLabel: (name: string) => `Manage ${name}`,
		manage: "Manage",
		created: (name: string, room: string) => (
			<>
				<strong>{name} is in this Cahut.</strong>
				Their personal Room is at <span class="c-mono">{room}</span>.
			</>
		),
		addAnother: "Add another person",
		done: "Done",
		name: "Name",
		handle: "Handle",
		handleWord: "handle",
		handleHelp: (address: string) => (
			<>
				Their name to sign in, and the address of their personal Room: <span class="c-mono">{address}</span>. It cannot change later.
			</>
		),
		emailHelp: "Optional. The link where they choose their password goes there, and so do their notices.",
		cancel: "Cancel",
		add: "Add the person",
		makeAdmin: "Make administrator",
		makeMember: "Make member",
		askAdmin: (name: string) => `Make ${name} an administrator? They will manage the people, the Rooms and the settings of this Cahut.`,
		askSelfMember: "Make yourself a member? You will no longer manage this Cahut.",
		askMember: (name: string) => `Make ${name} a member? They will no longer manage this Cahut.`,
		roleDone: (self: boolean, name: string, admin: boolean) =>
			`${self ? "You are" : `${name} is`} now ${admin ? "an administrator" : "a member"}.`,
		choosePassword: "Link to choose a password",
		newPassword: "New password",
		pendingHelp: (name: string) => `${name} has not chosen a password yet. A new link replaces the one they had.`,
		forgotHelp: (name: string) =>
			`If ${name} forgot their password, give them a link where they choose a new one. Their password works until they use it.`,
		newLink: "Give a new link",
		enable: "Enable",
		enableHelp: (name: string) => `${name} cannot sign in, and their AIs cannot act, until an administrator enables them.`,
		enabledPending: (name: string) => `${name} is enabled, and has no password yet: give them a new link.`,
		enabled: (name: string) => `${name} can sign in again.`,
		disable: "Disable",
		disableHelp:
			"Their sessions end, and neither they nor their AIs can act until an administrator enables them again. Their Rooms and Windows stay.",
		askDisable: (name: string) => `Disable ${name}? Their sessions end now.`,
		disabled: (name: string) => `${name} is disabled.`,
		selfNote: () => (
			<>
				Your password changes in{" "}
				<Link class="c-link" href="/account">
					your account
				</Link>
				.
			</>
		),
		emailSaved: "E-mail address saved.",
		emailRemoved: "E-mail address removed.",
		emailSettingHelp: "Their links and notices go there. Leave it empty to remove it.",
		save: "Save",
	},
	fr: {
		member: "Membre",
		memberHelp: "Travaille dans ses Rooms et connecte ses IA.",
		admin: "Administrateur",
		adminHelp: "Gère aussi les personnes, les Rooms et les réglages de cette Cahut.",
		onlyAdmins: "Seuls les administrateurs de cette Cahut en gèrent les personnes.",
		person: "personne",
		people: "personnes",
		administrator: "administrateur",
		administrators: "administrateurs",
		withoutPassword: (n) => `${n} encore sans mot de passe`,
		disabledCount: (n) => `${n} ${n > 1 ? "désactivées" : "désactivée"}`,
		thisCahut: "Cette Cahut",
		title: "Personnes",
		newPerson: "Nouvelle personne",
		intro: "Chaque personne se connecte avec son propre mot de passe, a une Room personnelle et connecte ses IA.",
		personColumn: "Personne",
		email: "Adresse e-mail",
		role: "Rôle",
		state: "État",
		actions: "Actions",
		you: "Vous",
		noAddress: "Pas d'adresse",
		manageLabel: (name) => `Gérer ${name}`,
		manage: "Gérer",
		created: (name, room) => (
			<>
				<strong>{name} est dans cette Cahut.</strong>
				Sa Room personnelle est à l'adresse <span class="c-mono">{room}</span>.
			</>
		),
		addAnother: "Ajouter une autre personne",
		done: "Terminé",
		name: "Nom",
		handle: "Identifiant",
		handleWord: "identifiant",
		handleHelp: (address) => (
			<>
				Son nom pour se connecter, et l'adresse de sa Room personnelle : <span class="c-mono">{address}</span>. Il ne pourra plus changer.
			</>
		),
		emailHelp: "Facultative. Le lien pour choisir son mot de passe y est envoyé, de même que ses notifications.",
		cancel: "Annuler",
		add: "Ajouter la personne",
		makeAdmin: "Rendre administrateur",
		makeMember: "Rendre membre",
		askAdmin: (name) => `Faire ${de(name)} un administrateur ? ${name} gérera les personnes, les Rooms et les réglages de cette Cahut.`,
		askSelfMember: "Devenir membre ? Vous ne gérerez plus cette Cahut.",
		askMember: (name) => `Faire ${de(name)} un membre ? ${name} ne gérera plus cette Cahut.`,
		roleDone: (self, name, admin) => `${self ? "Vous êtes" : `${name} est`} maintenant ${admin ? "administrateur" : "membre"}.`,
		choosePassword: "Lien pour choisir un mot de passe",
		newPassword: "Nouveau mot de passe",
		pendingHelp: (name) => `${name} n'a pas encore choisi de mot de passe. Un nouveau lien remplace l'ancien.`,
		forgotHelp: (name) =>
			`Si ${name} a oublié son mot de passe, donnez-lui un lien pour en choisir un nouveau. Son mot de passe actuel reste valable jusqu'à l'usage du lien.`,
		newLink: "Donner un nouveau lien",
		enable: "Réactiver",
		enableHelp: (name) => `${name} ne peut ni se connecter ni faire agir ses IA, tant qu'un administrateur ne réactive pas son compte.`,
		enabledPending: (name) => `Le compte ${de(name)} est réactivé, encore sans mot de passe : donnez-lui un nouveau lien.`,
		enabled: (name) => `${name} peut de nouveau se connecter.`,
		disable: "Désactiver",
		disableHelp:
			"Ses sessions prennent fin, et ni cette personne ni ses IA ne peuvent agir tant qu'un administrateur ne réactive pas son compte. Ses Rooms et ses Windows restent.",
		askDisable: (name) => `Désactiver ${name} ? Ses sessions prennent fin tout de suite.`,
		disabled: (name) => `Le compte ${de(name)} est désactivé.`,
		selfNote: () => (
			<>
				Votre mot de passe se change dans{" "}
				<Link class="c-link" href="/account">
					votre compte
				</Link>
				.
			</>
		),
		emailSaved: "Adresse e-mail enregistrée.",
		emailRemoved: "Adresse e-mail retirée.",
		emailSettingHelp: "Ses liens et ses notifications y sont envoyés. Laissez le champ vide pour la retirer.",
		save: "Enregistrer",
	},
});

const roles: Record<string, { word: string; help: string }> = {
	member: { word: t.member, help: t.memberHelp },
	admin: { word: t.admin, help: t.adminHelp },
};

function role(u: User) {
	return roles[u.role] ?? roles.member;
}

export function People() {
	const api = useConsole();
	const { data, error, reload } = useLoad(() => api.cahut("list_users", {}));
	const [adding, setAdding] = useState(false);
	const [managing, setManaging] = useState("");
	if (!api.isAdmin) return <Notice>{t.onlyAdmins}</Notice>;
	if (error) return <Failure message={error} />;
	if (!data) return null;

	const users = data.users;
	const isSelf = (u: User) => api.session.kind === "person" && u.handle === api.session.handle;
	const managed = users.find((u) => u.handle === managing);
	const count = (test: (u: User) => boolean) => users.filter(test).length;
	const summary = [plural(users.length, t.person, t.people), plural(count((u) => u.role === "admin"), t.administrator, t.administrators)];
	const waiting = count((u) => u.status === "pending");
	const disabled = count((u) => u.status === "disabled");
	if (waiting) summary.push(t.withoutPassword(waiting));
	if (disabled) summary.push(t.disabledCount(disabled));
	return (
		<>
			<PageHead
				label={t.thisCahut}
				title={t.title}
				actions={
					<button class="c-btn c-btn--primary" type="button" onClick={() => setAdding(true)}>
						<Plus class="i" size={16} aria-hidden="true" />
						{t.newPerson}
					</button>
				}
			>
				{t.intro}
			</PageHead>
			<section class="section">
				<p class="c-small c-muted people-summary">{summary.join(" · ")}</p>
				<div class="table-wrap">
					<table class="table people-table">
						<thead>
							<tr>
								<th>{t.personColumn}</th>
								{api.sendsMail && <th class="people-table__wide">{t.email}</th>}
								<th class="people-table__wide">{t.role}</th>
								<th class="people-table__wide">{t.state}</th>
								<th>
									<span class="visually-hidden">{t.actions}</span>
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
													{isSelf(u) && <span class="c-badge c-badge--neutral">{t.you}</span>}
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
											{u.email ? <span class="c-mono">{u.email}</span> : <span class="c-muted">{t.noAddress}</span>}
										</td>
									)}
									<td class="people-table__wide">{role(u).word}</td>
									<td class="people-table__wide">
										<StatusBadge status={u.status} />
									</td>
									<td class="table__actions">
										<button class="c-btn c-btn--secondary c-btn--sm" type="button" aria-label={t.manageLabel(u.name)} onClick={() => setManaging(u.handle)}>
											{t.manage}
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
			<Dialog title={t.newPerson} onClose={onClose}>
				<div class="c-alert c-alert--success" role="status">
					<Check class="i" size={16} aria-hidden="true" />
					<div>{t.created(created.user.name, `${location.hostname}/${created.user.handle}`)}</div>
				</div>
				<ActivationResult name={created.user.name} email={created.user.email} activation={created.activation} page="/activate" />
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
		<Dialog title={t.newPerson} onClose={onClose}>
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
					<span class="c-field__help">{t.handleHelp(`${location.hostname}/${handle || t.handleWord}`)}</span>
				</label>
				{api.sendsMail && (
					<label class="c-field">
						<span class="c-field__label">{t.email}</span>
						<input class="c-input" type="email" placeholder="ana@example.com" value={email} onInput={(e) => setEmail(e.currentTarget.value)} />
						<span class="c-field__help">{t.emailHelp}</span>
					</label>
				)}
				<RoleChoice admin={admin} onChange={setAdmin} />
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

/** The role of a new person, each with what it lets them do. */
function RoleChoice({ admin, onChange }: { admin: boolean; onChange: (admin: boolean) => void }) {
	const name = useId();
	return (
		<fieldset class="choices">
			<legend class="c-field__label">{t.role}</legend>
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
					title={t.role}
					help={`${role(u).word}. ${role(u).help}`}
					action={toAdmin ? t.makeAdmin : t.makeMember}
					ask={toAdmin ? t.askAdmin(u.name) : props.self ? t.askSelfMember : t.askMember(u.name)}
					onAct={async () => {
						await run(() => api.cahut("set_user_role", { user: u.handle, role: toAdmin ? "admin" : "member" }));
						return t.roleDone(props.self, u.name, toAdmin);
					}}
				/>
				{!props.self && u.status !== "disabled" && (
					<Setting
						key="link"
						title={u.status === "pending" ? t.choosePassword : t.newPassword}
						help={u.status === "pending" ? t.pendingHelp(u.name) : t.forgotHelp(u.name)}
						action={t.newLink}
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
							title={t.enable}
							help={t.enableHelp(u.name)}
							action={t.enable}
							onAct={async () => {
								const v = await run(() => api.cahut("enable_user", { user: u.handle }));
								return v.status === "pending" ? t.enabledPending(u.name) : t.enabled(u.name);
							}}
						/>
					) : (
						<Setting
							key="state"
							title={t.disable}
							help={t.disableHelp}
							action={t.disable}
							danger
							ask={t.askDisable(u.name)}
							onAct={async () => {
								await run(() => api.cahut("disable_user", { user: u.handle }));
								return t.disabled(u.name);
							}}
						/>
					))}
			</div>
			{props.self && (
				<p class="c-small c-muted dialog__note">{t.selfNote()}</p>
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
			setSaid(email ? t.emailSaved : t.emailRemoved);
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
					{t.email}
				</label>
				<p class="c-small c-muted">{t.emailSettingHelp}</p>
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
					{t.save}
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
