// The account: who is signed in, their e-mail address and notices, the
// language of the console, a new password, and signing out.
import { useState } from "preact/hooks";
import { explain } from "../api";
import type { EmailChange } from "../api.gen";
import { Failure, PageHead } from "../components";
import { useConsole } from "../context";
import { lang, langs, locale, setLang, words } from "../i18n";
import { useLoad } from "../load";

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
			"C'est votre mot de passe d'administrateur du serveur. Votre compte de personne d'un Cahut, si vous en avez un, garde le sien.",
		personPassword:
			"C'est votre mot de passe dans ce Cahut. Votre compte d'administrateur du serveur, si vous en avez un, garde le sien.",
		current: "Mot de passe actuel",
		newPassword: "Nouveau mot de passe",
		atLeast: (n) => `Au moins ${n} caractères. Vos autres sessions prennent fin.`,
		again: "Le même nouveau mot de passe, une seconde fois",
		changed: "Mot de passe changé. Vos autres sessions ont pris fin.",
		change: "Changer le mot de passe",
		signOut: "Se déconnecter",
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
	},
});

export function Account({ onSignOut }: { onSignOut: () => void }) {
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
			await (api.session.kind === "admin" ? api.server("change_password", input) : api.cahut("change_password", input));
			setPassword("");
			setNext("");
			setAgain("");
			setDone(true);
		} catch (err) {
			setError(explain(err));
		}
	};
	return (
		<>
			<PageHead label={t.account} title={api.session.name}>
				<span class="c-mono">{api.session.handle}</span>
				{api.session.kind === "admin" ? t.admin : t.at(location.hostname)}
			</PageHead>
			{api.session.kind === "person" && api.sendsMail && <EmailSection />}
			<Language />
			<section class="section c-card panel narrow">
				<h2 class="c-h4">{t.changePassword}</h2>
				<p class="c-small c-muted">{api.session.kind === "admin" ? t.adminPassword : t.personPassword}</p>
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
			</section>
			<button class="c-btn c-btn--secondary" type="button" onClick={onSignOut}>
				{t.signOut}
			</button>
		</>
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
