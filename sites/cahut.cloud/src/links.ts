// The addresses of My Cahut, the Cahut at my.cahut.cloud that Tetra7 offers
// in support of the project, apart from it (GOVERNANCE.md): account.cahut.cloud
// signs people up, in the language of the page, my.cahut.cloud signs its
// members in, and Claude lists it in its directory of connectors.
import type { Lang } from './i18n';

/** The sign-up form of my.cahut.cloud, in a language of the site. */
export function signupURL(lang: Lang): string {
	return `https://account.cahut.cloud/${lang}/signup`;
}

/** My Cahut itself, which its members open. */
export const myCahutURL = 'https://my.cahut.cloud/';

/** Where the members of my.cahut.cloud sign in. */
export const signInURL = 'https://my.cahut.cloud/console/';

/** My Cahut in the directory of Claude, where a member connects it without an address. */
export const claudeDirectoryURL = 'https://claude.ai/directory/my-cahut';

/** Tetra7, which offers My Cahut. */
export const tetra7URL = 'https://www.tetra7.io/';
