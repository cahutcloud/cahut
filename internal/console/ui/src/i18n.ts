// The languages of the console: English, and French. The console speaks the
// language that the person chose in this browser, or else the first of the
// browser's languages that it knows, or else English. Each screen keeps its
// words in both languages, next to its code, with words(): the French ones
// must have every English one, with the same arguments.
//
// The words of the house (Cahut, Room, Window, Key, House style, Knock first,
// Logbook) stay in English, as on the website. What the API answers, such as
// the message of an error, comes from the server in English. See
// docs/adr/0036-console-in-english-and-french.md.
import { cloneElement, isValidElement } from "preact";
import { fragment } from "./fragment";

export type Lang = "en" | "fr";

/** The languages that the console speaks, each in its own words. */
export const langs: { lang: Lang; name: string }[] = [
	{ lang: "en", name: "English" },
	{ lang: "fr", name: "Français" },
];

const key = "cahut.lang";

function known(l: string | null | undefined): l is Lang {
	return l === "en" || l === "fr";
}

/** The language that the person chose in this browser, if they did. */
export function chosen(): Lang | null {
	try {
		const l = localStorage.getItem(key);
		return known(l) ? l : null;
	} catch {
		return null;
	}
}

function pick(): Lang {
	const l = chosen();
	if (l) return l;
	for (const tag of navigator.languages ?? [navigator.language]) {
		const l = tag.slice(0, 2).toLowerCase();
		if (known(l)) return l;
	}
	return "en";
}

/** The language of the console, for as long as the page lives. */
export const lang: Lang = pick();
document.documentElement.lang = lang;

/** The locale of dates and numbers. */
export const locale = lang === "fr" ? "fr-FR" : "en-GB";

/**
 * Keeps the language that the person chose in this browser, and shows the
 * console in it again. The code of a link, which the console took out of the
 * address, goes back in it first, so that the page keeps it.
 */
export function setLang(l: Lang): void {
	try {
		localStorage.setItem(key, l);
	} catch {
		// Without storage, the console follows the browser's languages.
	}
	if (fragment) {
		history.replaceState(null, "", `${location.pathname}${location.search}#${fragment}`);
	}
	location.reload();
}

/**
 * French sets a space before its high punctuation and inside its quotation
 * marks, which must never begin a line: no-break before a colon and inside
 * « », narrow no-break before ; ? and !, and no-break after a number, as in
 * "15 caractères". The French words are written with plain spaces; this sets
 * them, as the website does.
 */
function frenchSpaces(s: string): string {
	return s
		.replace(/ :/g, "\u00a0:")
		.replace(/ ([;?!])/g, "\u202f$1")
		.replace(/« /g, "«\u00a0")
		.replace(/ »/g, "\u00a0»")
		.replace(/(\d) (?=\S)/g, "$1\u00a0");
}

/** Sets the French spaces in words: strings, what functions return, and the text of markup. */
function spaced<T>(value: T): T {
	if (typeof value === "string") return frenchSpaces(value) as T;
	if (typeof value === "function") return ((...args: unknown[]) => spaced(value(...args))) as T;
	if (Array.isArray(value)) return value.map(spaced) as T;
	if (isValidElement(value)) {
		const props = value.props as { children?: unknown };
		return (props.children === undefined ? value : cloneElement(value, { children: spaced(props.children) })) as T;
	}
	if (value && typeof value === "object") {
		return Object.fromEntries(Object.entries(value).map(([k, v]) => [k, spaced(v)])) as T;
	}
	return value;
}

/** Returns the words of the console's language, from the English ones and the French ones. */
export function words<T>(w: { en: T; fr: NoInfer<T> }): T {
	return lang === "fr" ? spaced(w.fr) : w.en;
}

/** Whether French elides the word before this one, as "de" before "Ana". */
function elides(word: string): boolean {
	return /^[aeiouyàâäéèêëîïôöùûüh]/i.test(word);
}

/** French "de" before a name: "de Joe", but "d'Ana". */
export function de(word: string): string {
	return elides(word) ? `d'${word}` : `de ${word}`;
}

/** French "que" before a name: "que Joe", but "qu'Ana". */
export function que(word: string): string {
	return elides(word) ? `qu'${word}` : `que ${word}`;
}

const rules = new Intl.PluralRules(locale);

/** Says a count with its noun, as in "1 Window" or "3 Windows": French says "0 Window", "1 Window". */
export function plural(n: number, one: string, many: string): string {
	return `${n.toLocaleString(locale)}${lang === "fr" ? "\u00a0" : " "}${rules.select(n) === "one" ? one : many}`;
}
