// The first start of a Cahut, from the single-use link that the
// installation prints: the first person chooses their account, then finds
// the console of the server, when they administer it too, or that of their
// Cahut, which offers to connect their first AI. See
// docs/adr/0025-first-start-in-the-browser.md, docs/adr/0049-no-house-style.md
// and docs/adr/0058-a-console-for-the-server.md.
import { ArrowRight } from "lucide-preact";
import { useEffect, useState } from "preact/hooks";
import { explain, server } from "../api";
import type { Setup as SetupInfo } from "../api.gen";
import { Languages } from "../components";
import { fragment as code } from "../fragment";
import { de, words } from "../i18n";
import { Logo } from "../logo";
import { Link, navigate } from "../router";
import type { Session } from "../session";
import { minPassword } from "./activate";

const t = words({
	en: {
		noCode: "This link has no code. Open the whole link that the installation printed.",
		gone: "This link does not open the first start",
		newLink: "On the server, this command gives a new link, valid for 24 hours, as long as the Cahut has nobody:",
		signInInstead: (
			<>
				If the first start is done,{" "}
				<Link class="c-link" href="/sign-in">
					sign in
				</Link>{" "}
				instead.
			</>
		),
		differ: "The two passwords differ.",
		welcome: "Welcome to your Cahut",
		youServer: (domain: string) =>
			`You will administer the Cahut of ${domain}, and this server, with the handle and the password you choose here.`,
		you: (domain: string) => `You will administer the Cahut of ${domain}, with the handle and the password you choose here.`,
		name: "Your name",
		handle: "Handle",
		handleHelp: "You sign in with it, and it names your personal Room.",
		password: "Password",
		passwordHelp: (min: number) => `At least ${min} characters. A short sentence works well.`,
		again: "The same password again",
		create: "Create my account",
	},
	fr: {
		noCode: "Ce lien n'a pas de code. Ouvrez le lien entier qu'a affiché l'installation.",
		gone: "Ce lien n'ouvre pas le premier démarrage",
		newLink: "Sur le serveur, cette commande donne un nouveau lien, valable 24 heures, tant que la Cahut n'a personne :",
		signInInstead: (
			<>
				Si le premier démarrage est fait,{" "}
				<Link class="c-link" href="/sign-in">
					connectez-vous
				</Link>{" "}
				plutôt.
			</>
		),
		differ: "Les deux mots de passe diffèrent.",
		welcome: "Bienvenue dans votre Cahut",
		youServer: (domain) =>
			`Vous administrerez la Cahut ${de(domain)}, et ce serveur, avec l'identifiant et le mot de passe que vous choisissez ici.`,
		you: (domain) => `Vous administrerez la Cahut ${de(domain)}, avec l'identifiant et le mot de passe que vous choisissez ici.`,
		name: "Votre nom",
		handle: "Identifiant",
		handleHelp: "Vous vous connectez avec, et il donne son nom à votre Room personnelle.",
		password: "Mot de passe",
		passwordHelp: (min) => `Au moins ${min} caractères. Une courte phrase convient bien.`,
		again: "Le même mot de passe, une seconde fois",
		create: "Créer mon compte",
	},
});

export function Setup({ onSession }: { onSession: (s: Session) => void }) {
	const [info, setInfo] = useState<SetupInfo | null>(null);
	const [error, setError] = useState(code ? "" : t.noCode);

	useEffect(() => {
		if (code) server("get_setup", { code }).then(setInfo, (err) => setError(explain(err)));
	}, []);

	// The first start ends in the console of the server, on the page of the
	// Cahut, when it made the administrator of the server; otherwise in the
	// console of the Cahut, whose overview offers to connect an AI. See
	// docs/adr/0058-a-console-for-the-server.md.
	const done = (person: Session, admin: Session | null) => {
		onSession(person);
		if (admin && info) {
			onSession(admin);
			location.assign(`/server/cahuts/${encodeURIComponent(info.domain)}?ready`);
			return;
		}
		navigate("/", true);
	};
	return (
		<main class="setup">
			<div class="setup__frame">
				<header class="setup__head">
					<Logo height={30} />
				</header>
				<div class="c-card setup__card">{error ? <Gone message={error} /> : info && <You info={info} onDone={done} />}</div>
				<Languages />
			</div>
		</main>
	);
}

function Gone({ message }: { message: string }) {
	return (
		<div class="setup__body">
			<h1 class="c-h3">{t.gone}</h1>
			<div class="c-alert c-alert--warning" role="alert">
				{message}
			</div>
			<p class="c-small">{t.newLink}</p>
			<pre class="c-mono setup__command">sudo -u cahut /usr/local/bin/cahut setup --domain {location.hostname}</pre>
			<p class="c-small">{t.signInInstead}</p>
		</div>
	);
}

/** A handle made from a name: its first word, in the letters a handle takes. */
function handleOf(name: string): string {
	const first = (name.trim().split(/\s+/)[0] ?? "").normalize("NFD").replace(/[̀-ͯ]/g, "");
	return first
		.toLowerCase()
		.replace(/[^a-z0-9-]/g, "")
		.replace(/^[^a-z]+/, "")
		.slice(0, 32);
}

function You({ info, onDone }: { info: SetupInfo; onDone: (person: Session, admin: Session | null) => void }) {
	const [name, setName] = useState("");
	const [handle, setHandle] = useState("");
	const [edited, setEdited] = useState(false);
	const [password, setPassword] = useState("");
	const [again, setAgain] = useState("");
	const [error, setError] = useState("");
	const [busy, setBusy] = useState(false);

	const submit = async (e: Event) => {
		e.preventDefault();
		if (password !== again) {
			setError(t.differ);
			return;
		}
		setBusy(true);
		setError("");
		try {
			const r = await server("set_up", { code, handle, name, password });
			const s = r.session;
			const person: Session = { kind: "person", token: s.token, handle: s.user.handle, name: s.user.name, expiresAt: s.expires_at };
			// The administrator of the server, made with the same handle and
			// password, signs in beside the person, in the console of the
			// server. See docs/adr/0058-a-console-for-the-server.md.
			const admin = r.server_admin
				? await server("sign_in", { handle: s.user.handle, password }).then(
						(a): Session => ({ kind: "admin", token: a.token, handle: a.admin.handle, name: a.admin.name, expiresAt: a.expires_at }),
						() => null,
					)
				: null;
			onDone(person, admin);
		} catch (err) {
			setError(explain(err));
			setBusy(false);
		}
	};
	return (
		<form class="setup__body" onSubmit={submit}>
			<div class="setup__intro">
				<h1 class="c-h3">{t.welcome}</h1>
				<p class="c-small c-muted">{info.server_admin ? t.youServer(info.domain) : t.you(info.domain)}</p>
			</div>
			<div class="form__row">
				<label class="c-field">
					<span class="c-field__label">{t.name}</span>
					<input
						class="c-input"
						required
						autocomplete="name"
						value={name}
						onInput={(e) => {
							setName(e.currentTarget.value);
							if (!edited) setHandle(handleOf(e.currentTarget.value));
						}}
					/>
				</label>
				<label class="c-field">
					<span class="c-field__label">{t.handle}</span>
					<input
						class="c-input c-mono"
						required
						autocapitalize="none"
						autocomplete="username"
						value={handle}
						onInput={(e) => {
							setEdited(true);
							setHandle(e.currentTarget.value);
						}}
					/>
					<span class="c-field__help">{t.handleHelp}</span>
				</label>
			</div>
			<label class="c-field">
				<span class="c-field__label">{t.password}</span>
				<input
					class="c-input"
					type="password"
					autocomplete="new-password"
					minLength={minPassword}
					required
					value={password}
					onInput={(e) => setPassword(e.currentTarget.value)}
				/>
				<span class="c-field__help">{t.passwordHelp(minPassword)}</span>
			</label>
			<label class="c-field">
				<span class="c-field__label">{t.again}</span>
				<input
					class="c-input"
					type="password"
					autocomplete="new-password"
					minLength={minPassword}
					required
					value={again}
					onInput={(e) => setAgain(e.currentTarget.value)}
				/>
			</label>
			{error && (
				<div class="c-alert c-alert--danger" role="alert">
					{error}
				</div>
			)}
			<div class="setup__actions">
				<button class="c-btn c-btn--primary" type="submit" disabled={busy}>
					{t.create}
					<ArrowRight class="i" size={16} aria-hidden="true" />
				</button>
			</div>
		</form>
	);
}
