// The legal notice and the privacy policy of the sites. Each page is a
// Markdown file of ./en/ with its translation of the same name in ./fr/; its
// frontmatter gives its title, its description and the date of its last
// change. The docs link to the English pages.
import type { SatteriProcessorOptions } from '@astrojs/markdown-satteri';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { frenchSpaces } from '../i18n';

/** The legal pages, in the order of the footer. */
export const legalPages = ['notice', 'privacy'] as const;
export type LegalPage = (typeof legalPages)[number];

/** The address of each page, before its language. */
export const legalPaths: Record<LegalPage, string> = {
	notice: '/legal/',
	privacy: '/privacy/',
};

/** What the frontmatter of a legal page gives. */
export interface LegalFrontmatter {
	title: string;
	description: string;
	/** The date of the last change, as YYYY-MM-DD. */
	updated: string;
}

const french = path.resolve(fileURLToPath(new URL('./fr', import.meta.url)));

type Plugins = NonNullable<SatteriProcessorOptions['mdastPlugins']>;

/**
 * The plugin of the Markdown processor for the French legal pages; it leaves
 * every other file alone. The spaces before the punctuation become no-break
 * spaces, as in the dictionaries of the site.
 */
export const legalPagesInFrench: Plugins[number] = ({ fileURL }) => {
	if (!fileURL || path.dirname(fileURLToPath(fileURL)) !== french) return false;
	return {
		name: 'legal-pages-in-french',
		text(node, ctx) {
			const value = frenchSpaces(node.value);
			if (value !== node.value) ctx.setProperty(node, 'value', value);
		},
	};
};
