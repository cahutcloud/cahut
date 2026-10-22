// The community files at the root of the repository, which the project pages
// and the policy on the name are built from (ADR 18): where each one is
// served, and the plugin of the Markdown processor that turns a file into its
// page. A page imports its file as @repo/<FILE>.md; the French pages import
// the translations of ./fr/, which have the same names.
import type { SatteriAstroData, SatteriProcessorOptions } from '@astrojs/markdown-satteri';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { frenchSpaces, localize, type Lang } from '../i18n';

export const github = 'https://github.com/getcahut/cahut';

/** The root of the repository, where the community files are. */
export const repository = path.resolve(fileURLToPath(new URL('../../../..', import.meta.url)));

/** The translations of the community files, by language. */
const translations = { fr: path.resolve(fileURLToPath(new URL('./fr', import.meta.url))) };

/** The pages of the project section, in the order of its navigation. */
export const pages = ['governance', 'roadmap', 'contribute', 'support', 'security', 'conduct'] as const;
export type ProjectPage = (typeof pages)[number];

/** The address of each page, before its language. */
export const paths: Record<ProjectPage, string> = {
	governance: '/project/',
	roadmap: '/project/roadmap/',
	contribute: '/project/contribute/',
	support: '/project/support/',
	security: '/project/security/',
	conduct: '/project/code-of-conduct/',
};

/**
 * The files of the repository that have a place on the site: the community
 * files, each the source of a page or of a section, and the files that a
 * page sums up.
 */
const served: Record<string, { page?: ProjectPage; href?: string }> = {
	'GOVERNANCE.md': { page: 'governance' },
	'CONTRIBUTING.md': { page: 'contribute' },
	'SECURITY.md': { page: 'security' },
	'CODE_OF_CONDUCT.md': { page: 'conduct' },
	'SUPPORT.md': { page: 'support' },
	// A section of the brand page, which is in English only.
	'TRADEMARK.md': { href: '/brand/#name-and-logo' },
	'docs/charte-graphique/README.md': { href: '/brand/' },
};

/** The files that become a section of a page: their headings move down one level. */
const sections = new Set(['TRADEMARK.md']);

/**
 * Where a link of a community file leads on the site. The files link to one
 * another and to the rest of the repository as GitHub reads them, by a path
 * from the root: a community file leads to its page, any other file opens on
 * GitHub. Addresses with a scheme, anchors and paths of the site stay as
 * they are.
 */
export function resolveLink(url: string, lang: Lang): string {
	if (/^(?:[a-z][a-z0-9+.-]*:|[#/])/i.test(url)) return url;
	const [file = '', hash] = url.replace(/^\.\//, '').split('#');
	const suffix = hash ? `#${hash}` : '';
	const place = served[file];
	if (place?.page) return localize(lang, paths[place.page]) + suffix;
	if (place?.href) return place.href + suffix;
	return `${github}/blob/main/${file}${suffix}`;
}

type Plugins = NonNullable<SatteriProcessorOptions['mdastPlugins']>;

/**
 * The plugin of the Markdown processor for the community files and their
 * translations; it leaves every other file alone. The title of the file goes
 * to the frontmatter, as `title`, for the page to show, and leaves the
 * content. The links lead to the site. The headings of a file that becomes a
 * section of a page move down one level. In French, the spaces before the
 * punctuation become no-break spaces, as in the dictionaries of the site.
 */
export const communityPages: Plugins[number] = ({ fileURL }) => {
	if (!fileURL) return false;
	const file = fileURLToPath(fileURL);
	const folder = path.dirname(file);
	const lang: Lang | undefined = folder === repository ? 'en' : folder === translations.fr ? 'fr' : undefined;
	const name = path.basename(file);
	if (!lang || !(name in served)) return false;
	const section = sections.has(name);
	let titled = false;
	return {
		name: 'community-pages',
		heading(node, ctx) {
			if (node.depth === 1 && !titled) {
				titled = true;
				const title = ctx.textContent(node);
				const astro = ctx.data.astro as SatteriAstroData | undefined;
				if (astro) astro.frontmatter.title = lang === 'fr' ? frenchSpaces(title) : title;
				ctx.removeNode(node);
			} else if (section && node.depth < 6) {
				ctx.setProperty(node, 'depth', (node.depth + 1) as typeof node.depth);
			}
		},
		link(node, ctx) {
			const url = resolveLink(node.url, lang);
			if (url !== node.url) ctx.setProperty(node, 'url', url);
		},
		text:
			lang === 'fr'
				? (node, ctx) => {
						const value = frenchSpaces(node.value);
						if (value !== node.value) ctx.setProperty(node, 'value', value);
					}
				: undefined,
	};
};
