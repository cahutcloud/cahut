// The page of a search: the Windows of the person whose titles,
// descriptions, tags or pages hold some words, the best first, each with a
// few words of its pages around those found, and the tags of the Windows
// found, which narrow them. The field of the top bar searches; on a narrow
// screen, where it moves to the menu, the page has its own. See
// docs/adr/0055-tags-and-search-of-windows.md.
import { useState } from "preact/hooks";
import { Failure, Notice, PageHead, plural, SearchForm, TagFilter, tagsOf, WindowCards } from "../components";
import { useConsole } from "../context";
import { words } from "../i18n";
import { useLoad } from "../load";
import { useQuery } from "../router";

const t = words({
	en: {
		label: "Search",
		title: "Search your Windows",
		quoted: (q: string) => `“${q}”`,
		lead: "Words of their titles, descriptions, tags or pages, whatever their accents and case. The start of a word is enough.",
		field: "Search your Windows",
		found: (n: number, q: string) => `${plural(n, "Window holds", "Windows hold")} “${q}”, the best first.`,
		none: (q: string) => `No Window of your Rooms holds “${q}”. Try fewer words, or the start of a word.`,
		byTag: "Show the Windows found with a tag",
		allTags: "All",
	},
	fr: {
		label: "Recherche",
		title: "Chercher dans vos Windows",
		quoted: (q) => `« ${q} »`,
		lead: "Des mots de leurs titres, descriptions, tags ou pages, quels que soient leurs accents et leur casse. Le début d'un mot suffit.",
		field: "Chercher dans vos Windows",
		found: (n, q) => `${plural(n, "Window contient", "Windows contiennent")} « ${q} », les meilleures d'abord.`,
		none: (q) => `Aucune Window de vos Rooms ne contient « ${q} ». Essayez moins de mots, ou le début d'un mot.`,
		byTag: "Montrer les Windows trouvées d'un tag",
		allTags: "Toutes",
	},
});

export function SearchPage() {
	const api = useConsole();
	const query = (useQuery().get("q") ?? "").trim();
	const [tag, setTag] = useState("");
	const { data, error } = useLoad(() => (query ? api.cahut("list_windows", { query }) : Promise.resolve(null)), [query]);
	const found = data?.windows;
	const tags = tagsOf(found ?? []);
	const shown = tags.some(([name]) => name === tag) ? found?.filter((w) => w.tags.includes(tag)) : found;
	const toggle = (name: string) => setTag(name === tag ? "" : name);
	return (
		<>
			<PageHead label={t.label} title={query ? t.quoted(query) : t.title}>
				{query && found ? (found.length > 0 ? t.found(found.length, query) : null) : t.lead}
			</PageHead>
			<SearchForm class="section search-page__field" label={t.field} query={query} />
			{error && <Failure message={error} />}
			{found && found.length === 0 && <Notice>{t.none(query)}</Notice>}
			{tags.length > 0 && <TagFilter tags={tags} selected={shown === found ? "" : tag} label={t.byTag} all={t.allTags} onTag={toggle} />}
			{shown && shown.length > 0 && <WindowCards windows={shown} showRoom query={query} tag={shown === found ? "" : tag} onTag={toggle} />}
		</>
	);
}
