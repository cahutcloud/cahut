// The account of the administrator of the server: the language of the
// console, the password of the server, and signing out of the server. See
// docs/adr/0058-a-console-for-the-server.md.
import { LogOut } from "lucide-preact";
import { PageHead } from "../components";
import { locale, words } from "../i18n";
import { ChangePassword, Language } from "../pages/account";
import { useServer } from "./context";

const t = words({
	en: {
		account: "Account",
		admin: ", administrator of the server.",
		changePassword: "Change your password",
		adminPassword:
			"This is your password as an administrator of the server. Your account as a person of a Cahut, if you have one, keeps its own.",
		signOutTitle: "Sign out",
		sessionEnds: (when: string) =>
			`Your session as an administrator of the server lasts in this browser until ${when} at the latest. Sign out to end it now. Your session as a person of the Cahut, if this browser holds one, stays open.`,
		signOut: "Sign out of the server",
	},
	fr: {
		account: "Compte",
		admin: ", administrateur du serveur.",
		changePassword: "Changer de mot de passe",
		adminPassword:
			"C'est votre mot de passe d'administrateur du serveur. Votre compte de personne d'une Cahut, si vous en avez un, garde le sien.",
		signOutTitle: "Déconnexion",
		sessionEnds: (when) =>
			`Votre session d'administrateur du serveur dure dans ce navigateur jusqu'au ${when} au plus tard. Déconnectez-vous pour la terminer maintenant. Votre session de personne de la Cahut, si ce navigateur en tient une, reste ouverte.`,
		signOut: "Se déconnecter du serveur",
	},
});

export function ServerAccount() {
	const api = useServer();
	return (
		<>
			<PageHead label={t.account} title={api.session.name}>
				<span class="c-mono">{api.session.handle}</span>
				{t.admin}
			</PageHead>
			<Language />
			<section class="section c-card panel narrow">
				<h2 class="c-h4">{t.changePassword}</h2>
				<p class="c-small c-muted">{t.adminPassword}</p>
				<ChangePassword change={(input) => api.server("change_password", input)} />
			</section>
			<section class="section c-card panel narrow">
				<h2 class="c-h4">{t.signOutTitle}</h2>
				<p class="c-small c-muted">
					{t.sessionEnds(new Date(api.session.expiresAt).toLocaleString(locale, { dateStyle: "long", timeStyle: "short" }))}
				</p>
				<div class="actions">
					<button class="c-btn c-btn--secondary" type="button" onClick={api.signOut}>
						<LogOut class="i" size={16} aria-hidden="true" />
						{t.signOut}
					</button>
				</div>
			</section>
		</>
	);
}
