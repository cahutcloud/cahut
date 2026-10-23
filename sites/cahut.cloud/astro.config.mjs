// @ts-check
import { satteri } from '@astrojs/markdown-satteri';
import brand from '@cahut/brand';
import { defineConfig } from 'astro/config';
import { legalPagesInFrench } from './src/legal/legal';
import { communityPages, repository } from './src/project/community';

export default defineConfig({
	site: 'https://cahut.cloud',
	integrations: [brand()],
	// The project pages are built from the community files at the root of the
	// repository, imported as @repo/<FILE>.md (ADR 18). The legal pages are
	// Markdown files of the site, in English and in French.
	markdown: { processor: satteri({ mdastPlugins: [communityPages, legalPagesInFrench] }) },
	vite: { resolve: { alias: [{ find: '@repo', replacement: repository }] } },
	server: { port: 4321 },
});
