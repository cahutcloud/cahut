// The addresses of my.cahut.cloud, the Cahut hosted by Olivier Durand, apart
// from the project (GOVERNANCE.md): account.cahut.cloud signs people up, in
// the language of the page, and my.cahut.cloud signs its members in.
import type { Lang } from './i18n';

/** The sign-up form of my.cahut.cloud, in a language of the site. */
export function signupURL(lang: Lang): string {
	return `https://account.cahut.cloud/${lang}/signup`;
}

/** Where the members of my.cahut.cloud sign in. */
export const signInURL = 'https://my.cahut.cloud/console/';
