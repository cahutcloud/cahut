// The account: who is signed in, their name and handle, their e-mail
// address and notices, the language of the console, a new password, the
// other account that the console holds or adds, and signing out.
import { LogOut } from "lucide-preact";
import { useState } from "preact/hooks";
import { explain } from "../api";
import type { EmailChange } from "../api.gen";
import { ConfirmButton, Failure, PageHead } from "../components";
import { useConsole } from "../context";
import { lang, langs, locale, setLang, words } from "../i18n";
import { useLoad } from "../load";
import { Link } from "../router";

const minPassword = 15;

const t = words({
	en: {
		differ: "The two new passwords differ.",
		account: "Account",
		admin: ", administrator of the server.",
		at: (host: string) => `, at ${host}.`,
		language: "Language",
		languageHelp: "The console speaks this language in this browser. What the server writes, such as its messages, stays in English.",
		changePassword: "Change your password",
		adminPassword:
			"This is your password as an administrator of the server. Your account as a person of a Cahut, if you have one, keeps its own.",
		personPassword:
			"This is your password in this Cahut. Your account as an administrator of the server, if you have one, keeps its own.",
		current: "Current password",
		newPassword: "New password",
		atLeast: (n: number) => `At least ${n} characters. Your other sessions end.`,
		again: "The same new password again",
		changed: "Password changed. Your other sessions have ended.",
		change: "Change the password",
		signOut: "Sign out",
		signOutTitle: "Sign out",
		sessionEnds: (when: string) => `Your session in this browser lasts until ${when} at the latest. Sign out to end it now.`,
		bothEnd: " Signing out also ends your session as an administrator of the server. To end that one only, stop administering the server here.",
		serverTitle: "Administrator of the server",
		serverHeld: (handle: string) => (
			<>
				This browser also holds your session as an administrator of the server, with the handle <span class="c-mono">{handle}</span>: the
				pages of the server join the console. This account keeps its own password.
			</>
		),
		leaveServer: "Stop administering the server here",
		addServerTitle: "Administer the server",
		addServer: "If you also administer this server, sign in with that account: the pages of the server join the console, and you stay signed in.",
		addServerLink: "Sign in to the server",
		addPersonTitle: "Your account in this Cahut",
		addPerson:
			"If you also have an account as a person of this Cahut, sign in with it: its Rooms, its Knocks and its AIs join the console, and you stay signed in to the server.",
		addPersonLink: "Sign in as a person",
		linkWent: (email: string) => `A link went to ${email}. Your address changes once you use it.`,
		removed: "Your address is removed.",
		email: "E-mail",
		goTo: (email: string) => (
			<>
				Links and notices go to <span class="c-mono">{email}</span>.
			</>
		),
		noEmail: "Your account has no e-mail address: links come from an administrator, and no notice reaches you.",
		waits: (email: string, until: string) => (
			<>
				<span class="c-mono">{email}</span> waits for you to use the link sent to it, until {until}.
			</>
		),
		newAddress: "New address",
		address: "Address",
		addressHelp: "A link goes to it: your address changes once you use it, within 24 hours.",
		send: "Send the link",
		remove: "Remove your address",
		off: "Notices are off, but for those about your account and your Knocks.",
		on: "Notices are on.",
		notices: "Tell me by e-mail when I join a Room, or when the access of a Window that I set ends",
		profile: "Profile",
		name: "Name",
		nameHelp: "As the people of this Cahut see it. Your personal Room takes it too, unless you gave it another title.",
		saveName: "Save the name",
		nameSaved: "Name saved.",
		handle: "Handle",
		handleHelp: (host: string, handle: string) => (
			<>
				You sign in with it. It names your personal Room too, at <span class="c-mono">{`${host}/${handle}/`}</span>.
			</>
		),
		handleWarning: (host: string, old: string, next: string) => (
			<>
				Your personal Room and its Windows move to <span class="c-mono">{`${host}/${next}/`}</span>. The links to them under{" "}
				<span class="c-mono">{`${host}/${old}/`}</span> stop working, for you and for everyone you gave them to, and{" "}
				<span class="c-mono">{old}</span> becomes free for another person or Room.
			</>
		),
		handleWarningIndividual: (host: string, old: string, next: string) => (
			<>
				Your Rooms and their Windows move: <span class="c-mono">{`${host}/${old}/`}</span> to{" "}
				<span class="c-mono">{`${host}/${next}/`}</span>, and <span class="c-mono">{`${host}/${old}-…/`}</span> to{" "}
				<span class="c-mono">{`${host}/${next}-…/`}</span>. The links to them under the old addresses stop working, for you and for
				everyone you gave them to, and <span class="c-mono">{old}</span> becomes free for another person.
			</>
		),
		changeHandle: "Change the handle",
		handleQuestion: "The old links stop working.",
		changeIt: "Change it",
		handleChanged: (handle: string) => `Your handle is now ${handle}: sign in with it from now on.`,
	},
	fr: {
		differ: "Les deux nouveaux mots de passe diffèrent.",
		account: "Compte",
		admin: ", administrateur du serveur.",
		at: (host) => `, sur ${host}.`,
		language: "Langue",
		languageHelp:
			"La console parle cette langue dans ce navigateur. Ce qu'écrit le serveur, comme ses messages d'erreur, reste en anglais.",
		changePassword: "Changer de mot de passe",
		adminPassword:
			"C'est votre mot de passe d'administrateur du serveur. Votre compte de personne d'une Cahut, si vous en avez un, garde le sien.",
		personPassword:
			"C'est votre mot de passe dans cette Cahut. Votre compte d'administrateur du serveur, si vous en avez un, garde le sien.",
		current: "Mot de passe actuel",
		newPassword: "Nouveau mot de passe",
		atLeast: (n) => `Au moins ${n} caractères. Vos autres sessions prennent fin.`,
		again: "Le même nouveau mot de passe, une seconde fois",
		changed: "Mot de passe changé. Vos autres sessions ont pris fin.",
		change: "Changer le mot de passe",
		signOut: "Se déconnecter",
		signOutTitle: "Déconnexion",
		sessionEnds: (when) => `Votre session dans ce navigateur dure jusqu'au ${when} au plus tard. Déconnectez-vous pour la terminer maintenant.`,
		bothEnd: " La déconnexion met aussi fin à votre session d'administrateur du serveur. Pour ne fermer que celle-ci, cessez d'administrer le serveur ici.",
		serverTitle: "Administrateur du serveur",
		serverHeld: (handle) => (
			<>
				Ce navigateur tient aussi votre session d'administrateur du serveur, avec l'identifiant <span class="c-mono">{handle}</span> : les
				pages du serveur s'ajoutent à la console. Ce compte garde son propre mot de passe.
			</>
		),
		leaveServer: "Ne plus administrer le serveur ici",
		addServerTitle: "Administrer le serveur",
		addServer:
			"Si vous administrez aussi ce serveur, connectez-vous avec ce compte : les pages du serveur s'ajoutent à la console, sans vous déconnecter.",
		addServerLink: "Se connecter au serveur",
		addPersonTitle: "Votre compte dans cette Cahut",
		addPerson:
			"Si vous avez aussi un compte de personne dans cette Cahut, connectez-vous avec lui : ses Rooms, ses Knocks et ses IA s'ajoutent à la console, sans vous déconnecter du serveur.",
		addPersonLink: "Se connecter en tant que personne",
		linkWent: (email) => `Un lien est parti à ${email}. Votre adresse change dès que vous l'utilisez.`,
		removed: "Votre adresse est retirée.",
		email: "E-mail",
		goTo: (email) => (
			<>
				Les liens et les notifications vont à <span class="c-mono">{email}</span>.
			</>
		),
		noEmail:
			"Votre compte n'a pas d'adresse e-mail : les liens viennent d'un administrateur, et aucune notification ne vous parvient.",
		waits: (email, until) => (
			<>
				<span class="c-mono">{email}</span> attend que vous utilisiez le lien qui lui a été envoyé, jusqu'au {until}.
			</>
		),
		newAddress: "Nouvelle adresse",
		address: "Adresse",
		addressHelp: "Un lien y part : votre adresse change dès que vous l'utilisez, dans les 24 heures.",
		send: "Envoyer le lien",
		remove: "Retirer votre adresse",
		off: "Les notifications sont désactivées, sauf celles sur votre compte et vos Knocks.",
		on: "Les notifications sont activées.",
		notices: "Me prévenir par e-mail quand je rejoins une Room, ou quand prend fin l'accès que j'ai donné à une Window",
		profile: "Profil",
		name: "Nom",
		nameHelp: "Tel que le voient les personnes de cette Cahut. Votre Room personnelle le prend aussi, sauf si vous lui avez donné un autre titre.",
		saveName: "Enregistrer le nom",
		nameSaved: "Nom enregistré.",
		handle: "Identifiant",
		handleHelp: (host, handle) => (
			<>
				Vous vous connectez avec lui. Il nomme aussi votre Room personnelle, à <span class="c-mono">{`${host}/${handle}/`}</span>.
			</>
		),
		handleWarning: (host, old, next) => (
			<>
				Votre Room personnelle et ses Windows passent à <span class="c-mono">{`${host}/${next}/`}</span>. Les liens vers elles sous{" "}
				<span class="c-mono">{`${host}/${old}/`}</span> ne marcheront plus, pour vous comme pour toutes les personnes à qui vous les avez
				donnés, et <span class="c-mono">{old}</span> devient libre pour une autre personne ou une Room.
			</>
		),
		handleWarningIndividual: (host, old, next) => (
			<>
				Vos Rooms et leurs Windows changent d'adresse : <span class="c-mono">{`${host}/${old}/`}</span> devient{" "}
				<span class="c-mono">{`${host}/${next}/`}</span>, et <span class="c-mono">{`${host}/${old}-…/`}</span> devient{" "}
				<span class="c-mono">{`${host}/${next}-…/`}</span>. Les liens vers elles sous les anciennes adresses ne marcheront plus, pour
				vous comme pour toutes les personnes à qui vous les avez donnés, et <span class="c-mono">{old}</span> devient libre pour une autre
				personne.
			</>
		),
		changeHandle: "Changer d'identifiant",
		handleQuestion: "Les anciens liens ne marcheront plus.",
		changeIt: "Changer",
		handleChanged: (handle) => `Votre identifiant est désormais ${handle} : connectez-vous avec lui à partir de maintenant.`,
	},
});

export function Account({ onSignOut }: { onSignOut: () => void }) {
	const api = useConsole();
	const person = api.session.kind === "person";
	// The session of the server, beside that of the person. See
	// docs/adr/0043-the-console-holds-two-sessions.md.
	const server = person ? api.serverSession : null;
	return (
		<>
			<PageHead label={t.account} title={api.session.name}>
				<span class="c-mono">{api.session.handle}</span>
				{person ? t.at(location.hostname) : t.admin}
			</PageHead>
			{person && <Profile />}
			{person && api.sendsMail && <EmailSection />}
			<Language />
			<section class="section c-card panel narrow">
				<h2 class="c-h4">{t.changePassword}</h2>
				<p class="c-small c-muted">{person ? t.personPassword : t.adminPassword}</p>
				<PasswordForm kind={api.session.kind} />
			</section>
			{server && (
				<section class="section c-card panel narrow">
					<h2 class="c-h4">{t.serverTitle}</h2>
					<p class="c-small c-muted">{t.serverHeld(server.handle)}</p>
					<PasswordForm kind="admin" />
				</section>
			)}
			{person && !server && api.isAdmin && (
				<section class="section c-card panel narrow">
					<h2 class="c-h4">{t.addServerTitle}</h2>
					<p class="c-small c-muted">{t.addServer}</p>
					<div>
						<Link class="c-btn c-btn--secondary" href="/server/sign-in">
							{t.addServerLink}
						</Link>
					</div>
				</section>
			)}
			{!person && (
				<section class="section c-card panel narrow">
					<h2 class="c-h4">{t.addPersonTitle}</h2>
					<p class="c-small c-muted">{t.addPerson}</p>
					<div>
						<Link class="c-btn c-btn--secondary" href="/sign-in">
							{t.addPersonLink}
						</Link>
					</div>
				</section>
			)}
			<section class="section c-card panel narrow">
				<h2 class="c-h4">{t.signOutTitle}</h2>
				<p class="c-small c-muted">
					{t.sessionEnds(new Date(api.session.expiresAt).toLocaleString(locale, { dateStyle: "long", timeStyle: "short" }))}
					{server && t.bothEnd}
				</p>
				<div class="actions">
					<button class="c-btn c-btn--secondary" type="button" onClick={onSignOut}>
						<LogOut class="i" size={16} aria-hidden="true" />
						{t.signOut}
					</button>
					{server && (
						<button class="c-btn c-btn--ghost" type="button" onClick={api.leaveServer}>
							{t.leaveServer}
						</button>
					)}
				</div>
			</section>
		</>
	);
}

/** Changes the password of an account: that of the person, or that of the server. */
function PasswordForm({ kind }: { kind: "person" | "admin" }) {
	const api = useConsole();
	const [password, setPassword] = useState("");
	const [next, setNext] = useState("");
	const [again, setAgain] = useState("");
	const [error, setError] = useState("");
	const [done, setDone] = useState(false);
	const submit = async (e: Event) => {
		e.preventDefault();
		setError("");
		setDone(false);
		if (next !== again) {
			setError(t.differ);
			return;
		}
		try {
			const input = { password, new_password: next };
			await (kind === "admin" ? api.server("change_password", input) : api.cahut("change_password", input));
			setPassword("");
			setNext("");
			setAgain("");
			setDone(true);
		} catch (err) {
			setError(explain(err));
		}
	};
	return (
		<form class="form" onSubmit={submit}>
			<label class="c-field">
				<span class="c-field__label">{t.current}</span>
				<input class="c-input" type="password" autocomplete="current-password" required value={password} onInput={(e) => setPassword(e.currentTarget.value)} />
			</label>
			<label class="c-field">
				<span class="c-field__label">{t.newPassword}</span>
				<input class="c-input" type="password" autocomplete="new-password" minLength={minPassword} required value={next} onInput={(e) => setNext(e.currentTarget.value)} />
				<span class="c-field__help">{t.atLeast(minPassword)}</span>
			</label>
			<label class="c-field">
				<span class="c-field__label">{t.again}</span>
				<input class="c-input" type="password" autocomplete="new-password" minLength={minPassword} required value={again} onInput={(e) => setAgain(e.currentTarget.value)} />
			</label>
			{error && <Failure message={error} />}
			{done && (
				<div class="c-alert c-alert--success" role="status">
					{t.changed}
				</div>
			)}
			<div>
				<button class="c-btn c-btn--primary" type="submit">
					{t.change}
				</button>
			</div>
		</form>
	);
}

/**
 * The name and the handle of the person, which they change themselves. A
 * new handle renames their personal Room: the console says first that the
 * links to its Windows stop working, and asks again on the spot.
 */
function Profile() {
	const api = useConsole();
	const { handle: current, name: currentName } = api.session;
	const host = location.host;
	const [name, setName] = useState(currentName);
	const [handle, setHandle] = useState(current);
	const [said, setSaid] = useState("");
	const [error, setError] = useState("");
	const act = async (run: () => Promise<string>) => {
		setError("");
		setSaid("");
		try {
			setSaid(await run());
		} catch (err) {
			setError(explain(err));
		}
	};
	const saveName = (e: Event) => {
		e.preventDefault();
		act(async () => {
			const u = await api.cahut("change_name", { name });
			setName(u.name);
			api.renamed(u);
			return t.nameSaved;
		});
	};
	const changeHandle = () =>
		act(async () => {
			const u = await api.cahut("change_handle", { handle });
			api.renamed(u);
			return t.handleChanged(u.handle);
		});
	const moving = handle !== "" && handle !== current;
	return (
		<section class="section c-card panel narrow">
			<h2 class="c-h4">{t.profile}</h2>
			<form class="form" onSubmit={saveName}>
				<label class="c-field">
					<span class="c-field__label">{t.name}</span>
					<input class="c-input" type="text" autocomplete="name" maxLength={200} required value={name} onInput={(e) => setName(e.currentTarget.value)} />
					<span class="c-field__help">{t.nameHelp}</span>
				</label>
				<div>
					<button class="c-btn c-btn--primary" type="submit" disabled={name.trim() === "" || name.trim() === currentName}>
						{t.saveName}
					</button>
				</div>
			</form>
			<div class="form">
				<label class="c-field">
					<span class="c-field__label">{t.handle}</span>
					<input
						class="c-input c-mono"
						type="text"
						autocomplete="username"
						autocapitalize="none"
						spellcheck={false}
						maxLength={32}
						value={handle}
						onInput={(e) => setHandle(e.currentTarget.value.trim().toLowerCase())}
					/>
					<span class="c-field__help">{t.handleHelp(host, current)}</span>
				</label>
				{moving && (
					<>
						<div class="c-alert c-alert--warning" role="status">
							<div>{(api.individual ? t.handleWarningIndividual : t.handleWarning)(host, current, handle)}</div>
						</div>
						<div>
							<ConfirmButton label={t.changeHandle} question={t.handleQuestion} confirm={t.changeIt} danger onConfirm={changeHandle} />
						</div>
					</>
				)}
			</div>
			{said && (
				<div class="c-alert c-alert--success" role="status">
					{said}
				</div>
			)}
			{error && <Failure message={error} />}
		</section>
	);
}

/** The language of the console in this browser, which a choice changes at once. */
function Language() {
	return (
		<section class="section c-card panel narrow">
			<h2 class="c-h4">{t.language}</h2>
			<p class="c-small c-muted">{t.languageHelp}</p>
			<fieldset class="consent__group">
				<legend class="visually-hidden">{t.language}</legend>
				{langs.map((l) => (
					<label key={l.lang} class="c-check" lang={l.lang}>
						<input type="radio" name="language" checked={l.lang === lang} onChange={() => setLang(l.lang)} />
						<span>{l.name}</span>
					</label>
				))}
			</fieldset>
		</section>
	);
}

/** What the account says of its e-mail address and notices. */
interface Mailbox {
	email?: string;
	change?: EmailChange;
	noticesOff?: boolean;
}

/**
 * The e-mail address of a person, where links and notices go, in a Cahut
 * that sends e-mail: a new one takes effect once the link sent to it is
 * used. The person also turns off the notices that are neither about their
 * account nor about a Knock. See docs/adr/0031-e-mail-through-a-relay.md and
 * docs/adr/0034-e-mail-through-the-relay-of-each-cahut-only.md.
 */
function EmailSection() {
	const api = useConsole();
	const { data, reload } = useLoad(async (): Promise<Mailbox> => {
		const a = await api.cahut("get_account", {});
		return { email: a.email, change: a.email_change, noticesOff: a.notices_off };
	});
	const [email, setEmail] = useState("");
	const [said, setSaid] = useState("");
	const [error, setError] = useState("");
	if (!data) return null;
	const act = async (run: () => Promise<string>) => {
		setError("");
		setSaid("");
		try {
			setSaid(await run());
			reload();
		} catch (err) {
			setError(explain(err));
		}
	};
	const change = (e: Event) => {
		e.preventDefault();
		act(async () => {
			const input = { email };
			const r = await api.cahut("change_email", input);
			setEmail("");
			return r.waiting ? t.linkWent(r.waiting.email) : t.removed;
		});
	};
	return (
		<section class="section c-card panel narrow">
			<h2 class="c-h4">{t.email}</h2>
			<p class="c-small">{data.email ? t.goTo(data.email) : t.noEmail}</p>
			{data.change && (
				<p class="c-small c-muted">
					{t.waits(data.change.email, new Date(data.change.expires_at).toLocaleString(locale, { dateStyle: "long", timeStyle: "short" }))}
				</p>
			)}
			<form class="form" onSubmit={change}>
				<label class="c-field">
					<span class="c-field__label">{data.email ? t.newAddress : t.address}</span>
					<input class="c-input" type="email" autocomplete="email" required value={email} onInput={(e) => setEmail(e.currentTarget.value)} />
					<span class="c-field__help">{t.addressHelp}</span>
				</label>
				<div class="actions">
					<button class="c-btn c-btn--primary" type="submit">
						{t.send}
					</button>
					{data.email && (
						<button
							class="c-btn c-btn--secondary"
							type="button"
							onClick={() =>
								act(async () => {
									await api.cahut("change_email", { email: "" });
									return t.removed;
								})
							}
						>
							{t.remove}
						</button>
					)}
				</div>
			</form>
			<label class="c-check">
				<input
					type="checkbox"
					checked={!data.noticesOff}
					onChange={(e) => {
						const off = !e.currentTarget.checked;
						act(async () => {
							await api.cahut("set_mail_notices", { off });
							return off ? t.off : t.on;
						});
					}}
				/>
				{t.notices}
			</label>
			{said && (
				<div class="c-alert c-alert--success" role="status">
					{said}
				</div>
			)}
			{error && <Failure message={error} />}
		</section>
	);
}
