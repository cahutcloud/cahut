// The brand of Cahut for the sites. Its single source is docs/charte-graphique:
// the sites import its files through the @charte alias and never copy them.

import { createReadStream, readdirSync } from 'node:fs';
import { cp } from 'node:fs/promises';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

/** The brand guidelines, with every source file of the brand. */
export const charte = fileURLToPath(new URL('../../docs/charte-graphique', import.meta.url));

const repository = fileURLToPath(new URL('../..', import.meta.url));
const favicons = path.join(charte, 'logos', 'favicon');

const contentTypes = {
	'.ico': 'image/x-icon',
	'.png': 'image/png',
	'.svg': 'image/svg+xml',
	'.webmanifest': 'application/manifest+json',
};

/**
 * Serves the favicons and the web manifest at the root of the development
 * server, before Astro's own routes.
 *
 * @type {import('vite').Plugin}
 */
const faviconsInDevelopment = {
	name: 'cahut-brand-favicons',
	configureServer(server) {
		const names = new Set(readdirSync(favicons));
		server.middlewares.use((req, res, next) => {
			const name = (req.url ?? '').split('?')[0].slice(1);
			if (!names.has(name)) {
				return next();
			}
			res.setHeader('Content-Type', contentTypes[path.extname(name)] ?? 'application/octet-stream');
			createReadStream(path.join(favicons, name)).pipe(res);
		});
	},
};

/**
 * The Astro integration of the brand. It maps `@charte/...` imports to the
 * brand guidelines, lets the development server read them, and puts the
 * favicons and the web manifest at the root of the site.
 *
 * @returns {import('astro').AstroIntegration}
 */
export default function brand() {
	return {
		name: '@cahut/brand',
		hooks: {
			'astro:config:setup': ({ updateConfig }) => {
				updateConfig({
					vite: {
						plugins: [faviconsInDevelopment],
						resolve: { alias: [{ find: '@charte', replacement: charte }] },
						server: { fs: { allow: [repository] } },
					},
				});
			},
			'astro:build:done': async ({ dir }) => {
				await cp(favicons, fileURLToPath(dir), { recursive: true });
			},
		},
	};
}
