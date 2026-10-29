// What a person meets in the console of a suspended Cahut, in place of its
// sign-in and of its pages: the Cahut keeps everything, and nobody but the
// administrators of its server acts in it until they resume it. The reason
// is for its administrators, who read it by e-mail: the notice does not say
// it. See docs/adr/0048-suspending-a-cahut.md.
import { LegalBeforeSignIn, Languages } from "../components";
import { words } from "../i18n";
import { Logo } from "../logo";
import { Link } from "../router";

const t = words({
	en: {
		title: "This Cahut is suspended",
		text: "The administrators of its server suspended it. It keeps everything it holds: your Rooms, your Windows and your AIs find their place again once they resume it.",
		ask: "If you work in this Cahut, ask its administrators.",
		signOut: "Sign out",
		backToConsole: "Back to the console",
	},
	fr: {
		title: "Cette Cahut est suspendue",
		text: "Les administrateurs de son serveur l'ont suspendue. Elle garde tout ce qu'elle contient : vos Rooms, vos Windows et vos IA retrouvent leur place dès qu'ils la réactivent.",
		ask: "Si vous travaillez dans cette Cahut, demandez à ses administrateurs.",
		signOut: "Se déconnecter",
		backToConsole: "Retour à la console",
	},
});

/**
 * The page of a suspended Cahut. A person who holds a session signs out of
 * it here; a browser that holds the session of the server goes back to the
 * console, where its administrator acts.
 */
export function Suspended(props: { onSignOut?: () => void; beside?: boolean }) {
	return (
		<main class="auth">
			<div class="c-card auth__card">
				<Logo height={32} />
				<div class="auth__head">
					<h1 class="c-h3">{t.title}</h1>
					<p class="c-small c-mono">{location.hostname}</p>
				</div>
				<p class="c-small">{t.text}</p>
				<p class="c-small c-muted">{t.ask}</p>
				{(props.onSignOut || props.beside) && (
					<div class="auth__foot">
						{props.onSignOut && (
							<p class="c-small">
								<button class="c-link link-button" type="button" onClick={props.onSignOut}>
									{t.signOut}
								</button>
							</p>
						)}
						{props.beside && (
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
			<LegalBeforeSignIn />
		</main>
	);
}
