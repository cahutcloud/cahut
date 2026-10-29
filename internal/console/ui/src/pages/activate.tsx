// The page of an activation link, where a person, or an administrator of the
// server, chooses their password. The code comes in the fragment of the
// address, which the server never sees. That of the server takes the theme
// of the server, as its sign-in does.
import { useState } from "preact/hooks";
import { cahut, explain, server } from "../api";
import { AuthLogo, Languages, LegalBeforeSignIn } from "../components";
import { fragment as code } from "../fragment";
import { words } from "../i18n";
import type { Session } from "../session";
import { useServerTheme } from "../theme";

export const minPassword = 15;

const t = words({
	en: {
		differ: "The two passwords differ.",
		chooseNew: "Choose a new password",
		choose: "Choose your password",
		reset: "Your current password works until you save this one. Your other sessions then end.",
		admin: "You administer the server. Choose the password you will sign in with.",
		person: (host: string) => `Choose the password you will sign in to ${host} with.`,
		newPassword: "New password",
		atLeast: (n: number) => `At least ${n} characters. A short sentence works well.`,
		again: "The same password again",
		save: "Save and sign in",
		noCode: "This link has no code. Open the whole link you received, or ask an administrator for a new one.",
	},
	fr: {
		differ: "Les deux mots de passe diffèrent.",
		chooseNew: "Choisissez un nouveau mot de passe",
		choose: "Choisissez votre mot de passe",
		reset: "Votre mot de passe actuel fonctionne jusqu'à ce que vous enregistriez celui-ci. Vos autres sessions prennent alors fin.",
		admin: "Vous administrez le serveur. Choisissez le mot de passe avec lequel vous vous connecterez.",
		person: (host) => `Choisissez le mot de passe avec lequel vous vous connecterez à ${host}.`,
		newPassword: "Nouveau mot de passe",
		atLeast: (n) => `Au moins ${n} caractères. Une courte phrase convient bien.`,
		again: "Le même mot de passe, une seconde fois",
		save: "Enregistrer et se connecter",
		noCode: "Ce lien n'a pas de code. Ouvrez le lien reçu en entier, ou demandez-en un nouveau à un administrateur.",
	},
});

// A link that a forgotten password sent says so in its address, after a
// question mark: the page then speaks of a new password.
const reset = new URLSearchParams(location.search).has("reset");

export function Activate({ kind, onSignIn }: { kind: Session["kind"]; onSignIn: (s: Session) => void }) {
	const theme = useServerTheme(kind === "admin");
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
			if (kind === "admin") {
				const s = await server("activate_account", { code, password });
				onSignIn({ kind, token: s.token, handle: s.admin.handle, name: s.admin.name, expiresAt: s.expires_at });
			} else {
				const s = await cahut("activate_account", { code, password });
				onSignIn({ kind, token: s.token, handle: s.user.handle, name: s.user.name, expiresAt: s.expires_at });
			}
		} catch (err) {
			setError(explain(err));
			setBusy(false);
		}
	};

	return (
		<main class="auth">
			<div class="c-card auth__card">
				<AuthLogo server={theme} />
				<div class="auth__head">
					<h1 class="c-h3">{reset ? t.chooseNew : t.choose}</h1>
					<p class="c-small c-muted">
						{reset ? t.reset : kind === "admin" ? t.admin : t.person(location.hostname)}
					</p>
				</div>
				{code ? (
					<form onSubmit={submit}>
						<label class="c-field">
							<span class="c-field__label">{t.newPassword}</span>
							<input
								class="c-input"
								type="password"
								name="new-password"
								autocomplete="new-password"
								minLength={minPassword}
								required
								value={password}
								onInput={(e) => setPassword(e.currentTarget.value)}
							/>
							<span class="c-field__help">{t.atLeast(minPassword)}</span>
						</label>
						<label class="c-field">
							<span class="c-field__label">{t.again}</span>
							<input
								class="c-input"
								type="password"
								name="new-password-again"
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
						<button class="c-btn c-btn--primary" type="submit" disabled={busy}>
							{t.save}
						</button>
					</form>
				) : (
					<div class="c-alert c-alert--warning" role="alert">
						{t.noCode}
					</div>
				)}
			</div>
			<Languages />
			<LegalBeforeSignIn />
		</main>
	);
}
