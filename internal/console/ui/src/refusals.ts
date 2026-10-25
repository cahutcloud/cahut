// The refusals that people meet, in the words of the console. The server
// gives each a code, which never changes, and the values of its message,
// next to its message in English: the console says it in its own language.
// Every code that make console-types writes into api.gen.ts needs its
// sentence in both languages here, or the build fails. See
// docs/adr/0037-refusals-with-a-code-for-the-console.md.
import type { Refusals } from "./api.gen";
import { plural, words } from "./i18n";

/** A sentence for each code: a string, or a function of its values. */
type Sentences = { [C in keyof Refusals]: Refusals[C] extends null ? string : (v: Refusals[C]) => string };

const said = words<Sentences>({
	en: {
		account_disabled: "This account is disabled: an administrator of the Cahut can enable it.",
		activation_link_gone: "This activation link is unknown, used or expired: ask an administrator of the Cahut for a new one.",
		email_link_gone: "This link is unknown, used or expired: change your address again from your account.",
		email_taken: (v) => `${v.email} is already the e-mail address of another person of this Cahut.`,
		invitation_gone:
			"This invitation does not open the Window any more: it was taken back, it expired, or the Window changed its Key or its access. Ask the person who invited you for a new one.",
		no_password_link: "This Cahut sends no e-mail: an administrator of the Cahut can give you a new link.",
		password_too_long: (v) => `The password has ${plural(v.length, "character", "characters")}: use at most ${v.limit}.`,
		password_too_short: (v) => `The password has ${plural(v.length, "character", "characters")}: use at least ${v.limit}.`,
		server_account_disabled: "This account is disabled: another administrator of the server can enable it.",
		server_activation_link_gone:
			"This activation link is unknown, used or expired: ask another administrator of the server for a new one, or run cahut admin link on the server.",
		session_ended: "Your session has ended: sign in again.",
		too_many_attempts_from_address: "Too many attempts from this address: wait a minute.",
		too_many_refused_sign_ins: (v) =>
			`Too many refused sign-ins for this handle: try again in ${plural(v.minutes, "minute", "minutes")}.`,
		too_many_sign_ins_at_once: "Too many sign-ins at once: try again in a few seconds.",
		wrong_current_password: "The current password is wrong.",
		wrong_handle_or_password: "Wrong handle or password.",
	},
	fr: {
		account_disabled: "Ce compte est désactivé : un administrateur du Cahut peut le réactiver.",
		activation_link_gone:
			"Ce lien d'activation est inconnu, déjà utilisé ou expiré : demandez-en un nouveau à un administrateur du Cahut.",
		email_link_gone: "Ce lien est inconnu, déjà utilisé ou expiré : changez à nouveau votre adresse depuis votre compte.",
		email_taken: (v) => `${v.email} est déjà l'adresse e-mail d'une autre personne de ce Cahut.`,
		invitation_gone:
			"Cette invitation n'ouvre plus la Window : elle a été retirée, elle a expiré, ou la Window a changé de Key ou d'accès. Demandez-en une nouvelle à la personne qui vous l'a envoyée.",
		no_password_link: "Ce Cahut n'envoie pas d'e-mail : un administrateur du Cahut peut vous donner un nouveau lien.",
		password_too_long: (v) => `Le mot de passe a ${plural(v.length, "caractère", "caractères")} : utilisez-en ${v.limit} au plus.`,
		password_too_short: (v) => `Le mot de passe a ${plural(v.length, "caractère", "caractères")} : utilisez-en au moins ${v.limit}.`,
		server_account_disabled: "Ce compte est désactivé : un autre administrateur du serveur peut le réactiver.",
		server_activation_link_gone:
			"Ce lien d'activation est inconnu, déjà utilisé ou expiré : demandez-en un nouveau à un autre administrateur du serveur, ou lancez cahut admin link sur le serveur.",
		session_ended: "Votre session a pris fin : reconnectez-vous.",
		too_many_attempts_from_address: "Trop de tentatives depuis cette adresse : patientez une minute.",
		too_many_refused_sign_ins: (v) =>
			`Trop de connexions refusées pour cet identifiant : réessayez dans ${plural(v.minutes, "minute", "minutes")}.`,
		too_many_sign_ins_at_once: "Trop de connexions en même temps : réessayez dans quelques secondes.",
		wrong_current_password: "Le mot de passe actuel est incorrect.",
		wrong_handle_or_password: "Identifiant ou mot de passe incorrect.",
	},
});

/**
 * Says the refusal of code with its values, in the language of the console,
 * or returns undefined for a code that the console does not know.
 */
export function say(code: string, values: unknown): string | undefined {
	if (!Object.hasOwn(said, code)) {
		return undefined;
	}
	const s = said[code as keyof Sentences];
	return typeof s === "function" ? (s as (v: unknown) => string)(values) : s;
}
