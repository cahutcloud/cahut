// The languages of cahut.cloud. English lives at the root of the site, every
// other language under its own folder, such as /fr/.
import en from './en';
import fr from './fr';

export type Dictionary = typeof en;
export type Lang = 'en' | 'fr';

/**
 * French sets a space before its high punctuation and inside its quotation
 * marks, which must never begin a line: no-break before a colon and inside
 * « », narrow no-break before ; ? and !, and no-break after a number, as in
 * "23 h". The French words are written with plain spaces; this sets them.
 */
function frenchSpaces<T>(value: T): T {
	if (typeof value === 'string') {
		return value
			.replace(/ :/g, ' :')
			.replace(/ ([;?!])/g, ' $1')
			.replace(/« /g, '« ')
			.replace(/ »/g, ' »')
			.replace(/(\d) (?=\S)/g, '$1 ') as T;
	}
	if (Array.isArray(value)) return value.map(frenchSpaces) as T;
	if (value && typeof value === 'object') {
		return Object.fromEntries(Object.entries(value).map(([key, item]) => [key, frenchSpaces(item)])) as T;
	}
	return value;
}

const dictionaries: Record<Lang, Dictionary> = { en, fr: frenchSpaces(fr) };

export function useTranslations(lang: Lang): Dictionary {
	return dictionaries[lang];
}

/** The address of a page of the site in a language. */
export function localize(lang: Lang, path = '/'): string {
	return lang === 'en' ? path : `/${lang}${path}`;
}

/** The other language, which the header offers. */
export function otherLang(lang: Lang): Lang {
	return lang === 'en' ? 'fr' : 'en';
}
