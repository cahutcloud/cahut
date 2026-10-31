// The demos of demo.cahut.cloud that the home shows, in the order of the
// dictionaries (examples.items). Each was published by an AI in English and in
// French, but for the Northstar deck, in English only; the images are
// screenshots of their first screen. The first two lead the grid: they keep
// what their readers write in a database. Four others play in the hero.
import type { ImageMetadata } from 'astro';
import type { Lang } from '../../i18n';
import cvEn from '../../assets/examples/en/cv.jpg';
import deckEn from '../../assets/examples/en/deck.jpg';
import gameEn from '../../assets/examples/en/game.jpg';
import globeEn from '../../assets/examples/en/globe.jpg';
import icelandEn from '../../assets/examples/en/iceland.jpg';
import northstar from '../../assets/examples/en/northstar.jpg';
import pollEn from '../../assets/examples/en/poll.jpg';
import reviewEn from '../../assets/examples/en/review.jpg';
import shopEn from '../../assets/examples/en/shop.jpg';
import vauclairEn from '../../assets/examples/en/vauclair.jpg';
import cvFr from '../../assets/examples/fr/cv.jpg';
import deckFr from '../../assets/examples/fr/deck.jpg';
import gameFr from '../../assets/examples/fr/game.jpg';
import globeFr from '../../assets/examples/fr/globe.jpg';
import icelandFr from '../../assets/examples/fr/iceland.jpg';
import pollFr from '../../assets/examples/fr/poll.jpg';
import reviewFr from '../../assets/examples/fr/review.jpg';
import shopFr from '../../assets/examples/fr/shop.jpg';
import vauclairFr from '../../assets/examples/fr/vauclair.jpg';

/** The page of demo.cahut.cloud that shows every demo with its prompt. */
export const examplesURL: Record<Lang, string> = {
	en: 'https://demo.cahut.cloud/examples/',
	fr: 'https://demo.cahut.cloud/examples/?lang=fr',
};

/** The Rooms of demo.cahut.cloud, which the dictionaries name (examples.rooms). */
export type Room = 'perso' | 'pme' | 'entreprise';

interface Version {
	url: string;
	image: ImageMetadata;
}

interface Example {
	id: string;
	room: Room;
	en: Version;
	fr?: Version;
}

const demo = 'https://demo.cahut.cloud';

const examples: Example[] = [
	{
		id: 'shop',
		room: 'pme',
		en: { url: `${demo}/pme/combe-hill-apiary/`, image: shopEn },
		fr: { url: `${demo}/pme/rucher-des-combes/`, image: shopFr },
	},
	{ id: 'northstar', room: 'entreprise', en: { url: `${demo}/entreprise/northstar-steerco-6/`, image: northstar } },
	{
		id: 'cv',
		room: 'perso',
		en: { url: `${demo}/perso/camille-roussel-cv/`, image: cvEn },
		fr: { url: `${demo}/perso/cv-de-camille-roussel/`, image: cvFr },
	},
	{
		id: 'review',
		room: 'entreprise',
		en: { url: `${demo}/entreprise/team-review/`, image: reviewEn },
		fr: { url: `${demo}/entreprise/revue-d-equipe/`, image: reviewFr },
	},
	{
		id: 'iceland',
		room: 'perso',
		en: { url: `${demo}/perso/iceland/`, image: icelandEn },
		fr: { url: `${demo}/perso/islande/`, image: icelandFr },
	},
	{
		id: 'vauclair',
		room: 'pme',
		en: { url: `${demo}/pme/vauclair/`, image: vauclairEn },
		fr: { url: `${demo}/pme/maison-vauclair/`, image: vauclairFr },
	},
	{
		id: 'deck',
		room: 'pme',
		en: { url: `${demo}/pme/fundraising-2027/`, image: deckEn },
		fr: { url: `${demo}/pme/levee-de-fonds-2027/`, image: deckFr },
	},
	{
		id: 'globe',
		room: 'entreprise',
		en: { url: `${demo}/entreprise/global-operations/`, image: globeEn },
		fr: { url: `${demo}/entreprise/operations-mondiales/`, image: globeFr },
	},
	{
		id: 'game',
		room: 'perso',
		en: { url: `${demo}/perso/noahs-game/`, image: gameEn },
		fr: { url: `${demo}/perso/le-jeu-de-noe/`, image: gameFr },
	},
	{
		id: 'poll',
		room: 'entreprise',
		en: { url: `${demo}/entreprise/offsite-2027/`, image: pollEn },
		fr: { url: `${demo}/entreprise/seminaire-2027/`, image: pollFr },
	},
];

/** How many demos lead the grid, larger than the others. */
export const leading = 2;

/** The demos that play in the hero, in their order. */
export const heroScenes = ['cv', 'review', 'iceland', 'vauclair'];

/**
 * The demos in a language, each with the language it opens in: a demo that
 * exists in English only opens in English.
 */
export function examplesIn(lang: Lang) {
	return examples.map(({ id, room, ...versions }) => {
		const own = versions[lang];
		return { id, room, ...(own ?? versions.en), lang: own ? lang : ('en' as Lang) };
	});
}

/** The address of a demo without its scheme, as a browser shows it. */
export function bare(url: string): string {
	return url.replace(/^https:\/\//, '');
}
