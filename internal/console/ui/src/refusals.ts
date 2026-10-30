// The refusals that people meet, in the words of the console. The server
// gives each a code, which never changes, and the values of its message,
// next to its message in English: the console says it in its own language.
// Every code that make console-types writes into api.gen.ts needs its
// sentence in both languages here, or the build fails. See
// docs/adr/0037-refusals-with-a-code-for-the-console.md.
import type { Refusals } from "./api.gen";
import { de, locale, plural, words } from "./i18n";

/** A sentence for each code: a string, or a function of its values. */
type Sentences = { [C in keyof Refusals]: Refusals[C] extends null ? string : (v: Refusals[C]) => string };

// The fields of the legal pages, as set_legal_links names them: terms, or
// fr.terms for its address in French. See
// docs/adr/0042-the-legal-pages-of-a-cahut.md.
const fields = words({
	en: {
		legal_notice: "Legal notice",
		terms: "Terms",
		privacy: "Privacy",
		support: "Support",
		inFrench: (field: string) => `${field}, in French`,
	},
	fr: {
		legal_notice: "Mentions légales",
		terms: "Conditions d'utilisation",
		privacy: "Confidentialité",
		support: "Assistance",
		inFrench: (field) => `${field}, en français`,
	},
});

/** The field of the form that a link of set_legal_links names, such as "Terms, in French" for fr.terms. */
function field(link: string): string {
	const french = link.startsWith("fr.");
	const page = french ? link.slice(3) : link;
	const names: Record<string, string> = { legal_notice: fields.legal_notice, terms: fields.terms, privacy: fields.privacy, support: fields.support };
	const name = names[page] ?? page;
	return french ? fields.inFrench(name) : name;
}

/** The day and the time of an instant that the server gives, as people read it. */
function when(at: string): string {
	return new Date(at).toLocaleString(locale, { dateStyle: "long", timeStyle: "short" });
}

const said = words<Sentences>({
	en: {
		account_disabled: "This account is disabled: an administrator of the Cahut can enable it.",
		activation_link_gone: "This activation link is unknown, used or expired: ask an administrator of the Cahut for a new one.",
		cahut_needs_an_admin: "The Cahut needs an administrator: make another person one first.",
		cahut_suspended:
			"This Cahut is suspended by the administrators of its server. It keeps everything it holds, and opens again once they resume it: ask its administrators.",
		description_control: "The description holds a control character: write it in plain text.",
		description_too_long: (v) => `The description has ${v.length} characters: use at most ${v.limit}, a sentence or two.`,
		email_already_yours: (v) => `${v.email} is already your e-mail address.`,
		email_link_gone: "This link is unknown, used or expired: change your address again from your account.",
		email_taken: (v) => `${v.email} is already the e-mail address of another person of this Cahut.`,
		handle_invalid: (v) =>
			`"${v.handle}" cannot be a handle: use 2 to 32 lowercase letters, digits and hyphens, starting with a letter and ending with a letter or a digit.`,
		handle_knocks_waiting:
			"Knocks wait for you: answer them before you change your handle, since they name the addresses of your personal Room.",
		handle_reserved: (v) => `The handle "${v.handle}" is reserved: it names a page of every Cahut.`,
		handle_room_cannot_move: (v) =>
			`Your Room ${v.room} would be named ${v.name}, which another Room holds or which is too long: choose another handle.`,
		handle_taken: (v) => `The handle "${v.handle}" is taken, by a person or a Room.`,
		invitation_author_only: () =>
			"This Window opens to you alone: nobody else signs in to read it, and it has no Key. Make a share link instead, which opens once for a single reader.",
		invitation_gone:
			"This invitation does not open the Window any more: it was taken back, it expired, or the Window changed its Key or its access. Ask the person who invited you for a new one.",
		invite_address_internal: () =>
			"This Window opens to the people of the Cahut only: invite them by handle, or set its access to key first, so that each address gets a personal link.",
		invite_address_room_only: () =>
			"This Window opens to its Room only: invite its members by handle, or set its access to key first, so that each address gets a personal link.",
		invite_bad_address: (v) => `"${v.to}" is not an e-mail address: give one address alone, such as jane@example.com.`,
		invite_by_address: (v) =>
			`"${v.to}" is not an e-mail address: in this Cahut, which is individual, invite people by address, to a Window with a Key or public.`,
		invite_disabled: (v) => `${v.handle} is disabled, and cannot read the Window.`,
		invite_no_person: (v) => `No person of this Cahut has the handle "${v.to}": invite someone else by e-mail address.`,
		invite_not_in_room: (v) =>
			`${v.handle} is not in the Room ${v.room}, and this Window opens to its Room only: add them to the Room, or set its access to internal or key first.`,
		key_has_spaces: "The Key starts or ends with a space.",
		legal_link_in_french_only: (v) => `${field(v.link)}: give the address of the page for the other languages too.`,
		legal_link_invalid: (v) => `${field(v.link)}: give the address of a page, which starts with https://, or leave the field empty.`,
		key_too_long: (v) => `The Key has more than ${plural(v.limit, "character", "characters")}.`,
		key_too_short: (v) => `The Key has ${plural(v.length, "character", "characters")}: give it ${v.limit} at least, or let Cahut draw one.`,
		name_empty: "Give a name.",
		name_too_long: "This name is too long.",
		no_mail: "This Cahut sends no e-mail: an administrator of the Cahut sets its relay first, on the page E-mail of the console.",
		no_mail_to_invite: "This Cahut sends no e-mail: give them the address of the Window yourself, with its Key if it has one.",
		no_password_link: "This Cahut sends no e-mail: an administrator of the Cahut can give you a new link.",
		password_too_long: (v) => `The password has ${plural(v.length, "character", "characters")}: use at most ${v.limit}.`,
		password_too_short: (v) => `The password has ${plural(v.length, "character", "characters")}: use at least ${v.limit}.`,
		room_name_invalid: (v) =>
			`"${v.name}" cannot be the name of a Room: use 1 to 63 lowercase letters, digits and hyphens, starting and ending with a letter or a digit.`,
		room_name_reserved: (v) => `The name "${v.name}" is reserved: it names a page of every Cahut.`,
		room_name_taken: (v) => `The name "${v.name}" is taken, by a Room or a person: choose another.`,
		room_name_unavailable: (v) => `The name "${v.name}" is not available: choose another.`,
		room_name_yours: (v) => `You already have a Room "${v.name}": choose another name.`,
		room_needs_an_owner: (v) => `The Room ${v.room} needs an owner: make another person one first.`,
		server_account_disabled: "This account is disabled: another administrator of the server can enable it.",
		server_activation_link_gone:
			"This activation link is unknown, used or expired: ask another administrator of the server for a new one, or run cahut admin link on the server.",
		session_ended: "Your session has ended: sign in again.",
		setup_done: (v) => `The Cahut of ${v.domain} has people already: sign in at its console.`,
		setup_link_gone: "This link to the first start does not work any more: it was used, it expired, or it is for another address.",
		share_link_end_past: (v) => `This end is past: a share link ends between now and ${when(v.latest)}.`,
		share_link_ended: () => "This share link has ended already: it opens nothing more.",
		share_link_label: (v) => `Write the label on one line, in ${plural(v.limit, "character", "characters")} at most.`,
		share_link_openings: (v) => `Give a number of readers from 1 to ${v.limit}, or none for any number.`,
		share_link_public: () => "This Window is public: its permanent address opens it to anyone. Copy that link instead.",
		share_link_too_long: (v) => `A share link lasts 30 days at most: give an end before ${when(v.latest)}.`,
		title_empty: "The title is empty: give one.",
		title_too_long: "The title is too long: shorten it.",
		too_many_attempts_from_address: "Too many attempts from this address: wait a minute.",
		too_many_invitations: (v) =>
			`You and your AIs sent ${plural(v.sent, "invitation", "invitations")} in the last 24 hours, and ${v.limit} is the most: send the rest later.`,
		too_many_refused_sign_ins: (v) =>
			`Too many refused sign-ins for this handle: try again in ${plural(v.minutes, "minute", "minutes")}.`,
		too_many_share_links: (v) =>
			`This Window has ${v.limit} share links that stand, the most: take one back before you make another.`,
		too_many_share_links_today: (v) =>
			`You and your AIs made ${plural(v.made, "share link", "share links")} in the last 24 hours, and ${v.limit} is the most: make the rest later.`,
		too_many_sign_ins_at_once: "Too many sign-ins at once: try again in a few seconds.",
		window_name_invalid: (v) =>
			`"${v.name}" cannot be the name of a Window: use 1 to 63 lowercase letters, digits and hyphens, starting and ending with a letter or a digit.`,
		window_author_only: (v) => `Only the author of this Window, ${v.author}, can keep it for themself.`,
		window_name_taken: (v) => `The name "${v.name}" is taken in the Room ${v.room}, by a Window or one in its trash: choose another.`,
		window_without_author:
			"This Window has no author since the person it was published for was deleted: it opens to its Room, its Cahut, a Key or the public.",
		wrong_current_password: "The current password is wrong.",
		wrong_handle_or_password: "Wrong handle or password.",
	},
	fr: {
		account_disabled: "Ce compte est désactivé : un administrateur de la Cahut peut le réactiver.",
		activation_link_gone:
			"Ce lien d'activation est inconnu, déjà utilisé ou expiré : demandez-en un nouveau à un administrateur de la Cahut.",
		cahut_needs_an_admin: "La Cahut a besoin d'un administrateur : faites d'abord d'une autre personne un administrateur.",
		cahut_suspended:
			"Cette Cahut est suspendue par les administrateurs de son serveur. Elle garde tout ce qu'elle contient, et rouvre dès qu'ils la réactivent : demandez à ses administrateurs.",
		description_control: "La description contient un caractère de contrôle : écrivez-la en texte simple.",
		description_too_long: (v) => `La description fait ${v.length} caractères : ${v.limit} au plus, une phrase ou deux.`,
		email_already_yours: (v) => `${v.email} est déjà votre adresse e-mail.`,
		email_link_gone: "Ce lien est inconnu, déjà utilisé ou expiré : changez à nouveau votre adresse depuis votre compte.",
		email_taken: (v) => `${v.email} est déjà l'adresse e-mail d'une autre personne de cette Cahut.`,
		handle_invalid: (v) =>
			`« ${v.handle} » ne peut pas être un identifiant : utilisez de 2 à 32 lettres minuscules, chiffres et tirets, en commençant par une lettre et en finissant par une lettre ou un chiffre.`,
		handle_knocks_waiting:
			"Des Knocks vous attendent : répondez-y avant de changer d'identifiant, car ils nomment les adresses de votre Room personnelle.",
		handle_reserved: (v) => `L'identifiant « ${v.handle} » est réservé : il nomme une page de chaque Cahut.`,
		handle_room_cannot_move: (v) =>
			`Votre Room ${v.room} s'appellerait ${v.name}, nom qu'une autre Room porte déjà ou qui est trop long : choisissez un autre identifiant.`,
		handle_taken: (v) => `L'identifiant « ${v.handle} » est déjà pris, par une personne ou une Room.`,
		invitation_author_only: () =>
			"Cette Window ne s'ouvre qu'à vous : personne d'autre ne s'y connecte pour la lire, et elle n'a pas de Key. Faites plutôt un lien de partage, qui s'ouvre une fois pour une seule personne.",
		invitation_gone:
			"Cette invitation n'ouvre plus la Window : elle a été retirée, elle a expiré, ou la Window a changé de Key ou d'accès. Demandez-en une nouvelle à la personne qui vous l'a envoyée.",
		invite_address_internal: () =>
			"Cette Window s'ouvre aux personnes de la Cahut seulement : invitez-les par identifiant, ou passez d'abord son accès à Key, pour que chaque adresse reçoive un lien personnel.",
		invite_address_room_only: () =>
			"Cette Window s'ouvre à sa Room seulement : invitez ses membres par identifiant, ou passez d'abord son accès à Key, pour que chaque adresse reçoive un lien personnel.",
		invite_bad_address: (v) => `« ${v.to} » n'est pas une adresse e-mail : donnez une seule adresse, comme jeanne@example.com.`,
		invite_by_address: (v) =>
			`« ${v.to} » n'est pas une adresse e-mail : dans cette Cahut, qui est individuelle, invitez par adresse, à une Window avec une Key ou publique.`,
		invite_disabled: (v) => `Le compte de ${v.handle} est désactivé : cette personne ne peut pas lire la Window.`,
		invite_no_person: (v) => `Aucune personne de cette Cahut n'a l'identifiant « ${v.to} » : invitez quelqu'un d'autre par adresse e-mail.`,
		invite_not_in_room: (v) =>
			`${v.handle} n'est pas dans la Room ${v.room}, et cette Window s'ouvre à sa Room seulement : ajoutez cette personne à la Room, ou passez d'abord son accès à Interne ou à Key.`,
		key_has_spaces: "La Key commence ou finit par un espace.",
		legal_link_in_french_only: (v) => `${field(v.link)} : donnez aussi l'adresse de la page pour les autres langues.`,
		legal_link_invalid: (v) => `${field(v.link)} : donnez l'adresse d'une page, qui commence par https://, ou laissez le champ vide.`,
		key_too_long: (v) => `La Key a plus de ${plural(v.limit, "caractère", "caractères")}.`,
		key_too_short: (v) =>
			`La Key a ${plural(v.length, "caractère", "caractères")} : il lui en faut ${v.limit} au moins, ou laissez Cahut en tirer une au sort.`,
		name_empty: "Donnez un nom.",
		name_too_long: "Ce nom est trop long.",
		no_mail: "Cette Cahut n'envoie pas d'e-mail : un administrateur de la Cahut doit d'abord régler son relais, sur la page E-mail de la console.",
		no_mail_to_invite: "Cette Cahut n'envoie pas d'e-mail : donnez vous-même l'adresse de la Window, avec sa Key si elle en a une.",
		no_password_link: "Cette Cahut n'envoie pas d'e-mail : un administrateur de la Cahut peut vous donner un nouveau lien.",
		password_too_long: (v) => `Le mot de passe a ${plural(v.length, "caractère", "caractères")} : utilisez-en ${v.limit} au plus.`,
		password_too_short: (v) => `Le mot de passe a ${plural(v.length, "caractère", "caractères")} : utilisez-en au moins ${v.limit}.`,
		room_name_invalid: (v) =>
			`« ${v.name} » ne peut pas être le nom d'une Room : utilisez de 1 à 63 lettres minuscules, chiffres et tirets, en commençant et en finissant par une lettre ou un chiffre.`,
		room_name_reserved: (v) => `Le nom « ${v.name} » est réservé : il nomme une page de chaque Cahut.`,
		room_name_taken: (v) => `Le nom « ${v.name} » est déjà pris, par une Room ou une personne : choisissez-en un autre.`,
		room_name_unavailable: (v) => `Le nom « ${v.name} » n'est pas disponible : choisissez-en un autre.`,
		room_name_yours: (v) => `Vous avez déjà une Room « ${v.name} » : choisissez un autre nom.`,
		room_needs_an_owner: (v) => `La Room ${v.room} a besoin d'un propriétaire : rendez d'abord une autre personne propriétaire.`,
		server_account_disabled: "Ce compte est désactivé : un autre administrateur du serveur peut le réactiver.",
		server_activation_link_gone:
			"Ce lien d'activation est inconnu, déjà utilisé ou expiré : demandez-en un nouveau à un autre administrateur du serveur, ou lancez cahut admin link sur le serveur.",
		session_ended: "Votre session a pris fin : reconnectez-vous.",
		setup_done: (v) => `La Cahut ${de(v.domain)} a déjà des personnes : connectez-vous à sa console.`,
		setup_link_gone: "Ce lien de premier démarrage ne fonctionne plus : il a servi, il a expiré, ou il est pour une autre adresse.",
		share_link_end_past: (v) => `Cette fin est passée : un lien de partage finit entre maintenant et le ${when(v.latest)}.`,
		share_link_ended: () => "Ce lien de partage a déjà pris fin : il n'ouvre plus rien.",
		share_link_label: (v) => `Écrivez l'étiquette sur une ligne, en ${plural(v.limit, "caractère", "caractères")} au plus.`,
		share_link_openings: (v) => `Donnez un nombre de lecteurs de 1 à ${v.limit}, ou aucun pour un nombre sans limite.`,
		share_link_public: () => "Cette Window est publique : son adresse permanente l'ouvre à tous. Copiez plutôt ce lien.",
		share_link_too_long: (v) => `Un lien de partage dure 30 jours au plus : donnez une fin avant le ${when(v.latest)}.`,
		title_empty: "Le titre est vide : donnez-en un.",
		title_too_long: "Le titre est trop long : raccourcissez-le.",
		too_many_attempts_from_address: "Trop de tentatives depuis cette adresse : patientez une minute.",
		too_many_invitations: (v) =>
			`Vous et vos IA avez envoyé ${plural(v.sent, "invitation", "invitations")} ces dernières 24 heures, et ${v.limit} est le maximum : envoyez les autres plus tard.`,
		too_many_refused_sign_ins: (v) =>
			`Trop de connexions refusées pour cet identifiant : réessayez dans ${plural(v.minutes, "minute", "minutes")}.`,
		too_many_share_links: (v) =>
			`Cette Window a ${v.limit} liens de partage actifs, le maximum : retirez-en un avant d'en créer un autre.`,
		too_many_share_links_today: (v) =>
			`Vous et vos IA avez créé ${plural(v.made, "lien de partage", "liens de partage")} ces dernières 24 heures, et ${v.limit} est le maximum : créez les autres plus tard.`,
		too_many_sign_ins_at_once: "Trop de connexions en même temps : réessayez dans quelques secondes.",
		window_name_invalid: (v) =>
			`« ${v.name} » ne peut pas être le nom d'une Window : utilisez de 1 à 63 lettres minuscules, chiffres et tirets, en commençant et en finissant par une lettre ou un chiffre.`,
		window_author_only: (v) => `Seule la personne qui a publié cette Window, ${v.author}, peut la garder pour elle.`,
		window_name_taken: (v) => `Le nom « ${v.name} » est déjà pris dans la Room ${v.room}, par une Window ou une Window à la corbeille : choisissez-en un autre.`,
		window_without_author:
			"Cette Window n'a plus d'auteur depuis la suppression de la personne pour qui elle a été publiée : elle s'ouvre à sa Room, à sa Cahut, à une Key ou au public.",
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
