// The files of the brand guidelines that the brand page shows or offers to
// download. Vite copies each one into the build; `no-inline` keeps it a file
// even when it is small enough to become a data URL.
import { statSync } from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { root } from 'astro:config/server';

// Glob keys are paths from the root of the site, such as
// ../../docs/charte-graphique/logos/cahut-logo.svg.
const urls = byFile(
	import.meta.glob(
		[
			'@charte/charte-graphique-cahut.pdf',
			'@charte/couleurs/*.{ase,gpl}',
			'@charte/tokens/cahut-tokens.json',
			'@charte/typographie/polices/*.txt',
			'@charte/logos/*.svg',
			'@charte/logos/png/*.png',
			'@charte/logos/favicon/favicon{.ico,-16.png,-32.png}',
			'@charte/icones/*.svg',
			'@charte/illustrations/*.{svg,png}',
		],
		{ query: '?url&no-inline', import: 'default', eager: true },
	),
);

const sources = byFile(
	import.meta.glob(['@charte/icones/*.svg', '@charte/logos/cahut-logo.svg'], {
		query: '?raw',
		import: 'default',
		eager: true,
	}),
);

interface File {
	/** The glob key: the file's path from the root of the site. */
	key: string;
	/** What the import gave: an address in the build, or the file's text. */
	value: string;
}

/** Keys a glob of the guidelines by the path of each file in docs/charte-graphique. */
function byFile(modules: Record<string, unknown>): Map<string, File> {
	return new Map(
		Object.entries(modules).map(([key, value]) => [
			key.replace(/^.*?\/charte-graphique\//, ''),
			{ key, value: String(value) },
		]),
	);
}

function find(files: Map<string, File>, file: string): File {
	const found = files.get(file);
	if (!found) {
		throw new Error(`brand page: docs/charte-graphique/${file} is not among the files of the page`);
	}
	return found;
}

/** The address in the build of a file of the guidelines, given by its path in docs/charte-graphique. */
export function url(file: string): string {
	return find(urls, file).value;
}

/** The source of an SVG file of the guidelines, to put in the page itself. */
export function source(file: string): string {
	return find(sources, file).value;
}

/** The size of a file of the guidelines, in words, such as "3.4 MB". */
export function size(file: string): string {
	const bytes = statSync(path.resolve(fileURLToPath(root), find(urls, file).key)).size;
	if (bytes < 1000) {
		return `${bytes} bytes`;
	}
	if (bytes < 1_000_000) {
		return `${Math.round(bytes / 1000)} kB`;
	}
	return `${(bytes / 1_000_000).toFixed(1)} MB`;
}

/** The file name of a file of the guidelines, for the download attribute of a link. */
export function name(file: string): string {
	return path.basename(file);
}
