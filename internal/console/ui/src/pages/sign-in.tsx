// Sign-in, for the people of this Cahut and for the administrators of the
// server, who sign in at the console of any Cahut. Each sign-in keeps the
// other session that the browser holds. See
// docs/adr/0043-the-console-holds-two-sessions.md. The sign-in of a person
// does not show the way to that of the server, which its administrators
// reach at /server/sign-in, or from the page of their account; that of the
// server takes the theme of the server, the inverse of that of a person.
// Behind the card, the Windows of the open house come out, and go back in
// once signed in.
import { useEffect, useRef, useState } from "preact/hooks";
import { cahut, explain, server } from "../api";
import type { LegalLinks } from "../api.gen";
import { AuthLogo, describeSignIn, Languages, Legal, LegalBeforeSignIn } from "../components";
import { words } from "../i18n";
import { Logo } from "../logo";
import { Link } from "../router";
import { type SceneMotion, SignInScene } from "../sign-in-scene";
import type { Session } from "../session";
import { useServerTheme } from "../theme";
import { Suspended } from "./suspended";

const t = words({
	en: {
		administer: "Administer this server",
		signIn: "Sign in",
		asAdmin: (host: string) => `Sign in as an administrator of the server, at ${host}.`,
		to: (host: string, purpose?: string) => `Sign in to ${host}${purpose ? ` ${purpose}` : ""}.`,
		handle: "Handle",
		password: "Password",
		adminForgot: (
			<>
				Forgot your password? Another administrator of the server, or <span class="c-mono">cahut admin link</span> on the server,
				gives you a new link.
			</>
		),
		asPerson: "Sign in as a person of this Cahut",
		forgot: "Forgot your password?",
		askAdmin: "Forgot your password? An administrator of the Cahut can give you a new link.",
		newPassword: "Choose a new password",
		giveHandle:
			"Give your handle, or the e-mail address of your account. A link to choose a new password goes to that address. If your account has none, an administrator of the Cahut can give you a link.",
		sent: (message: string) => message,
		handleOrEmail: "Handle or e-mail address",
		sendLink: "Send me a link",
		back: "Back to sign in",
		keepsPerson: " Your session as a person stays open.",
		keepsAdmin: " Your session as an administrator of the server stays open.",
		backToConsole: "Back to the console",
	},
	fr: {
		administer: "Administrer ce serveur",
		signIn: "Se connecter",
		asAdmin: (host) => `Connectez-vous en tant qu'administrateur du serveur, à ${host}.`,
		to: (host, purpose) => `Connectez-vous à ${host}${purpose ? ` ${purpose}` : ""}.`,
		handle: "Identifiant",
		password: "Mot de passe",
		adminForgot: (
			<>
				Mot de passe oublié ? Un autre administrateur du serveur, ou <span class="c-mono">cahut admin link</span> sur le serveur,
				vous donne un nouveau lien.
			</>
		),
		asPerson: "Se connecter en tant que personne de cette Cahut",
		forgot: "Mot de passe oublié ?",
		askAdmin: "Mot de passe oublié ? Un administrateur de la Cahut peut vous donner un nouveau lien.",
		newPassword: "Choisissez un nouveau mot de passe",
		giveHandle:
			"Donnez votre identifiant, ou l'adresse e-mail de votre compte. Un lien pour choisir un nouveau mot de passe part à cette adresse. Si votre compte n'en a pas, un administrateur de la Cahut peut vous donner un lien.",
		// The answer is the same for every account: the console says it in French.
		sent: () =>
			"Si un compte de cette Cahut a cet identifiant ou cette adresse, un lien pour choisir un nouveau mot de passe est parti à son adresse. Il est valable une heure.",
		handleOrEmail: "Identifiant ou adresse e-mail",
		sendLink: "M'envoyer un lien",
		back: "Retour à la connexion",
		keepsPerson: " Votre session de personne reste ouverte.",
		keepsAdmin: " Votre session d'administrateur du serveur reste ouverte.",
		backToConsole: "Retour à la console",
	},
});

export function SignIn(props: {
	kind: Session["kind"];
	purpose?: string;
	/** Whether the browser holds the session of the other kind, which stays open. */
	beside?: boolean;
	onSignIn: (s: Session) => void;
}) {
	const { kind, purpose, beside, onSignIn } = props;
	const theme = useServerTheme(kind === "admin");
	const [forgot, setForgot] = useState(false);
	// A person who forgot their password gets a link by e-mail only where
	// this Cahut sends e-mail. See
	// docs/adr/0034-e-mail-through-the-relay-of-each-cahut-only.md. The
	// foot of the page links the legal pages of the Cahut.
	const [sendsMail, setSendsMail] = useState(false);
	const [legal, setLegal] = useState<LegalLinks>();
	// Nobody but the administrators of the server signs in to a suspended
	// Cahut. See docs/adr/0048-suspending-a-cahut.md.
	const [suspended, setSuspended] = useState(false);
	useEffect(() => {
		describeSignIn().then(
			(p) => (setSendsMail(p.sends_mail), setLegal(p.legal), setSuspended(p.suspended === true)),
			() => undefined,
		);
	}, []);
	const [handle, setHandle] = useState("");
	const [password, setPassword] = useState("");
	const [error, setError] = useState("");
	const [busy, setBusy] = useState(false);
	const card = useRef<HTMLDivElement>(null);
	const scene = useRef<SceneMotion | undefined>(undefined);

	const submit = async (e: Event) => {
		e.preventDefault();
		setBusy(true);
		setError("");
		try {
			if (kind === "admin") {
				const s = await server("sign_in", { handle, password });
				await scene.current?.home();
				onSignIn({ kind, token: s.token, handle: s.admin.handle, name: s.admin.name, expiresAt: s.expires_at });
			} else {
				const s = await cahut("sign_in", { handle, password });
				await scene.current?.home();
				onSignIn({ kind, token: s.token, handle: s.user.handle, name: s.user.name, expiresAt: s.expires_at });
			}
		} catch (err) {
			setError(explain(err));
			setBusy(false);
			scene.current?.startle();
		}
	};

	if (suspended && kind === "person") return <Suspended beside={beside} />;
	if (forgot) return <Forgot who={handle} onBack={() => setForgot(false)} />;
	return (
		<main class="auth auth--scene">
			<SignInScene card={card} motion={scene} />
			<div class="c-card auth__card" ref={card}>
				<AuthLogo server={theme} />
				<div class="auth__head">
					<h1 class="c-h3">{kind === "admin" ? t.administer : t.signIn}</h1>
					<p class="c-small c-muted">
						{kind === "admin" ? t.asAdmin(location.hostname) : t.to(location.hostname, purpose)}
						{beside && (kind === "admin" ? t.keepsPerson : t.keepsAdmin)}
					</p>
				</div>
				<form onSubmit={submit}>
					<label class="c-field">
						<span class="c-field__label">{t.handle}</span>
						<input
							class="c-input"
							name="username"
							autocomplete="username"
							autocapitalize="none"
							spellcheck={false}
							required
							value={handle}
							onInput={(e) => setHandle(e.currentTarget.value)}
						/>
					</label>
					<label class="c-field">
						<span class="c-field__label">{t.password}</span>
						<input
							class="c-input"
							type="password"
							name="password"
							autocomplete="current-password"
							required
							value={password}
							onInput={(e) => setPassword(e.currentTarget.value)}
						/>
					</label>
					{error && (
						<div class="c-alert c-alert--danger" role="alert">
							{error}
						</div>
					)}
					<button class="c-btn c-btn--primary" type="submit" disabled={busy}>
						{t.signIn}
					</button>
				</form>
				{kind === "admin" ? (
					<div class="auth__foot">
						<p class="c-small c-muted">{t.adminForgot}</p>
						<p class="c-small">
							<Link class="c-link" href={beside ? "/" : "/sign-in"}>
								{beside ? t.backToConsole : t.asPerson}
							</Link>
						</p>
					</div>
				) : (
					<div class="auth__foot">
						{sendsMail ? (
							<p class="c-small">
								<button class="c-link link-button" type="button" onClick={() => setForgot(true)}>
									{t.forgot}
								</button>
							</p>
						) : (
							<p class="c-small c-muted">{t.askAdmin}</p>
						)}
						{beside && (
							<p class="c-small">
								<Link class="c-link" href="/">
									{t.backToConsole}
								</Link>
							</p>
						)}
					</div>
				)}
			</div>
			<Languages />
			<Legal links={legal} />
		</main>
	);
}

/**
 * Asks for a link where a person chooses a new password. It goes to the
 * address of their account; the answer is the same whether the account
 * exists or not. See docs/adr/0031-e-mail-through-a-relay.md.
 */
function Forgot({ who: given, onBack }: { who: string; onBack: () => void }) {
	const [who, setWho] = useState(given);
	const [sent, setSent] = useState("");
	const [error, setError] = useState("");
	const [busy, setBusy] = useState(false);
	const submit = async (e: Event) => {
		e.preventDefault();
		setBusy(true);
		setError("");
		try {
			const r = await cahut("request_password_link", { who });
			setSent(t.sent(r.message));
		} catch (err) {
			setError(explain(err));
		}
		setBusy(false);
	};
	return (
		<main class="auth">
			<div class="c-card auth__card">
				<Logo height={32} />
				<div class="auth__head">
					<h1 class="c-h3">{t.newPassword}</h1>
					<p class="c-small c-muted">{t.giveHandle}</p>
				</div>
				{sent ? (
					<div class="c-alert c-alert--success" role="status">
						{sent}
					</div>
				) : (
					<form onSubmit={submit}>
						<label class="c-field">
							<span class="c-field__label">{t.handleOrEmail}</span>
							<input
								class="c-input"
								name="username"
								autocomplete="username"
								autocapitalize="none"
								spellcheck={false}
								required
								value={who}
								onInput={(e) => setWho(e.currentTarget.value)}
							/>
						</label>
						{error && (
							<div class="c-alert c-alert--danger" role="alert">
								{error}
							</div>
						)}
						<button class="c-btn c-btn--primary" type="submit" disabled={busy}>
							{t.sendLink}
						</button>
					</form>
				)}
				<p class="c-small">
					<button class="c-link link-button" type="button" onClick={onBack}>
						{t.back}
					</button>
				</p>
			</div>
			<Languages />
			<LegalBeforeSignIn />
		</main>
	);
}
