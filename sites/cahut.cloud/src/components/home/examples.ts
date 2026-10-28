// The demos of demo.cahut.cloud that the home shows, in the order of the
// dictionaries (examples.items). The first four play in the hero. Each was
// published by an AI; the images are screenshots of their first screen.
import cv from '../../assets/examples/cv.jpg';
import deck from '../../assets/examples/deck.jpg';
import globe from '../../assets/examples/globe.jpg';
import islande from '../../assets/examples/islande.jpg';
import jeu from '../../assets/examples/jeu.jpg';
import revue from '../../assets/examples/revue.jpg';
import sondage from '../../assets/examples/sondage.jpg';
import vauclair from '../../assets/examples/vauclair.jpg';

/** The page of demo.cahut.cloud that shows every demo with its prompt. */
export const examplesURL = 'https://demo.cahut.cloud/examples/';

/** The demos, with the title of the Room each one lives in. */
export const examples = [
	{ url: 'https://demo.cahut.cloud/perso/cv-de-camille-roussel/', room: 'Perso', image: cv },
	{ url: 'https://demo.cahut.cloud/entreprise/revue-d-equipe/', room: 'Grande entreprise', image: revue },
	{ url: 'https://demo.cahut.cloud/perso/islande/', room: 'Perso', image: islande },
	{ url: 'https://demo.cahut.cloud/pme/maison-vauclair/', room: 'PME', image: vauclair },
	{ url: 'https://demo.cahut.cloud/pme/levee-de-fonds-2027/', room: 'PME', image: deck },
	{ url: 'https://demo.cahut.cloud/entreprise/operations-mondiales/', room: 'Grande entreprise', image: globe },
	{ url: 'https://demo.cahut.cloud/perso/le-jeu-de-noe/', room: 'Perso', image: jeu },
	{ url: 'https://demo.cahut.cloud/entreprise/seminaire-2027/', room: 'Grande entreprise', image: sondage },
];

/** The address of a demo without its scheme, as a browser shows it. */
export function bare(url: string): string {
	return url.replace(/^https:\/\//, '');
}
