// The films of the YouTube channel of Cahut that the home links to, in each
// language of the site. The home shows them by a link and a picture of its own:
// it calls no other service before a visitor chooses to watch, as the privacy
// policy says (src/legal).
import type { ImageMetadata } from 'astro';
import type { Lang } from '../../i18n';
import myCahutEn from '../../assets/videos/en/my-cahut.jpg';
import myCahutFr from '../../assets/videos/fr/my-cahut.jpg';

const ids = {
	claude: { en: 'am2iBk5AX0U', fr: 'QtWHG6qjjKw' },
	mistral: { en: 'I_2rw66sR5s', fr: 'vWm_Wy6qS2Q' },
	myCahut: { en: 'JMhOVCx-L74', fr: 'XYfDC_bodZo' },
} as const;

export type Film = keyof typeof ids;

/** The address of a film on YouTube, in a language of the site. */
export function filmURL(film: Film, lang: Lang): string {
	return `https://www.youtube.com/watch?v=${ids[film][lang]}`;
}

/** The picture of the film of My Cahut, the one its page on YouTube shows. */
export const myCahutPicture: Record<Lang, ImageMetadata> = { en: myCahutEn, fr: myCahutFr };
