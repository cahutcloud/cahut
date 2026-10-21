// The design tokens of the brand guidelines, read from cahut-tokens.json, and
// what the brand page computes from them: color values, names and contrasts.
import tokens from '@charte/tokens/cahut-tokens.json';

export { tokens };

/** The scales of the palette, as named in the tokens and in the CSS variables. */
export const scales = ['ink', 'sky', 'yellow'] as const;
export type Scale = (typeof scales)[number];

const scaleNames: Record<Scale, string> = { ink: 'Ink', sky: 'Sky', yellow: 'Yellow' };

/** The name of a scale for people, such as "Ink". */
export function scaleName(scale: Scale): string {
	return scaleNames[scale];
}

interface Token {
	$value: string;
}

/** The steps of a scale, from 50 to 950, with their hex values. */
export function steps(scale: Scale): { step: string; hex: string }[] {
	return Object.entries(tokens.color[scale] as Record<string, Token>).map(([step, token]) => ({
		step,
		hex: token.$value,
	}));
}

function lookup(path: string): string {
	const found = path
		.replace(/^\{|\}$/g, '')
		.split('.')
		.reduce<unknown>((node, key) => (node as Record<string, unknown> | undefined)?.[key], tokens);
	const value = (found as Token | undefined)?.$value;
	if (typeof value !== 'string') {
		throw new Error(`brand page: no color token ${path} in cahut-tokens.json`);
	}
	return value;
}

/** The hex value of a color token, given by its path, such as "color.ink.600" or "{color.brand.paper}". */
export function hex(path: string): string {
	const value = lookup(path);
	return value.startsWith('{') ? hex(value) : value;
}

/** The color token that a role points to, such as "color.ink.950" for "role.dark.bg". */
export function target(path: string): string {
	const value = lookup(path);
	return value.startsWith('{') ? target(value.slice(1, -1)) : path.replace(/^\{|\}$/g, '');
}

/** The name of a color token for people, such as "Ink 950", "Paper" or "Light Grey". */
export function colorName(reference: string): string {
	const [, group, key] = reference.replace(/^\{|\}$/g, '').split('.');
	if (group in scaleNames) {
		return `${scaleNames[group as Scale]} ${key}`;
	}
	return key
		.split('-')
		.map((word) => word[0].toUpperCase() + word.slice(1))
		.join(' ');
}

/** The step of a scale that a brand color sits on, such as "Ink 900" for Ink. */
export function stepOf(value: string): string | undefined {
	for (const scale of scales) {
		const found = steps(scale).find((s) => s.hex.toLowerCase() === value.toLowerCase());
		if (found) {
			return `${scaleNames[scale]} ${found.step}`;
		}
	}
	return undefined;
}

/** The red, green and blue channels of a hex color, from 0 to 255. */
export function rgb(value: string): [number, number, number] {
	const n = Number.parseInt(value.slice(1), 16);
	return [(n >> 16) & 255, (n >> 8) & 255, n & 255];
}

/**
 * The naive conversion of a hex color to cyan, magenta, yellow and black, in
 * percent. The guidelines give these values as indicative only.
 */
export function cmyk(value: string): [number, number, number, number] {
	const [r, g, b] = rgb(value).map((c) => c / 255);
	const k = 1 - Math.max(r, g, b);
	if (k === 1) {
		return [0, 0, 0, 100];
	}
	const [c, m, y] = [r, g, b].map((c) => Math.round(((1 - c - k) / (1 - k)) * 100));
	return [c, m, y, Math.round(k * 100)];
}

function luminance(value: string): number {
	const [r, g, b] = rgb(value).map((c) => {
		const s = c / 255;
		return s <= 0.04045 ? s / 12.92 : ((s + 0.055) / 1.055) ** 2.4;
	});
	return 0.2126 * r + 0.7152 * g + 0.0722 * b;
}

/** The WCAG contrast ratio of two hex colors, from 1 to 21. */
export function contrast(a: string, b: string): number {
	const [light, dark] = [luminance(a), luminance(b)].sort((x, y) => y - x);
	return (light + 0.05) / (dark + 0.05);
}

/** The CSS color of text that reads best on a background: Ink or Paper. */
export function textOn(background: string): string {
	const ink = hex('color.brand.ink');
	const paper = hex('color.brand.paper');
	return contrast(background, ink) >= contrast(background, paper) ? 'var(--cahut-ink)' : 'var(--cahut-paper)';
}

/** A size or line height token in pixels, such as 56 for "5xl". */
export function pixels(group: 'size' | 'line-height', key: string): number {
	const token = (tokens.font[group] as Record<string, { $value: { value: number } }>)[key];
	if (!token) {
		throw new Error(`brand page: no font ${group} ${key} in cahut-tokens.json`);
	}
	return token.$value.value;
}
