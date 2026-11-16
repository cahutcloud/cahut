// Skills: what tells the AIs of a Cahut how to do a task, such as the brand
// of a team or the reminder of an unpaid invoice. The skills of the Cahut,
// which its administrators write and everyone's AIs read; the skills of the
// person, which no one else reads; and the skill of the plugin, which comes
// with Cahut. An individual Cahut has no skills of the Cahut. See
// docs/adr/0064-the-skills-of-a-cahut.md.
import { ChevronRight, Download, Eye, FileUp, PenLine } from "lucide-preact";
import type { ComponentChildren } from "preact";
import type { Summary } from "../api.gen";
import { agoInText, CopyField, Failure, PageHead } from "../components";
import { useConsole } from "../context";
import { locale, plural, words } from "../i18n";
import { useLoad } from "../load";
import { Link } from "../router";

const t = words({
	en: {
		label: "This Cahut",
		title: "Skills",
		lead: "A skill tells an AI how to do a task: the brand of a team, a reminder to send, a way to check a contract. The AIs connected to this Cahut read its skills, whichever AI each person uses.",
		leadIndividual: "A skill tells an AI how to do a task: your brand, a reminder to send, a way to check a contract. The AIs connected to this Cahut read your skills, whichever AI you use.",
		ofTheCahut: "Skills of this Cahut",
		ofTheCahutHelp: "Its administrators write them, and the AIs of everyone in this Cahut read them.",
		noneOfTheCahut: "This Cahut has no skill yet.",
		noneOfTheCahutMember: "This Cahut has no skill yet. Its administrators write them.",
		mine: "My skills",
		mineHelp: "No one but you and your AIs reads them, not even the administrators.",
		noneOfMine: "You have no skill yet.",
		builtin: "Comes with Cahut",
		builtinHelp: "The skill of the plugin for Claude, which every Cahut gives its AIs. It tells them how to publish Windows.",
		readOnly: "Read only",
		downloadZip: "Download its zip",
		write: "Write a skill",
		import: "Import a zip",
		version: (n: number) => `Version ${n}`,
		written: (when: string) => `written ${when}`,
		everyWindow: "For every Window",
		nameIt: (name: string) => `Name a skill to your AI, as in “Follow the skill ${name}”, and it reads it.`,
		pasteHelp: "If your AI does not think of the skills of this Cahut, paste this line in its instructions:",
		pasteLine: "Before each task, call the find_skills tool of Cahut to see whether my team has a skill that applies to it, then follow it.",
		pasteLineIndividual: "Before each task, call the find_skills tool of Cahut to see whether one of my skills applies to it, then follow it.",
		theLine: "the line",
		aisSee: "What AIs see",
		shown: (n: number, of: number) => `${n} of ${plural(of, "skill", "skills")}`,
		aisSeeHelp: (bytes: string, budget: string) =>
			`Every AI connected to this Cahut receives these lines in the description of find_skills, from the start of each conversation. They take ${bytes} of ${budget} bytes.`,
		left: (names: string) => `Left out, for lack of room: ${names}. AIs still find them with find_skills.`,
	},
	fr: {
		label: "Cette Cahut",
		title: "Skills",
		lead: "Un skill dit à une IA comment faire une tâche : la charte d'une équipe, une relance à envoyer, une façon de vérifier un contrat. Les IA branchées sur cette Cahut lisent ses skills, quelle que soit l'IA de chaque personne.",
		leadIndividual:
			"Un skill dit à une IA comment faire une tâche : votre charte, une relance à envoyer, une façon de vérifier un contrat. Les IA branchées sur cette Cahut lisent vos skills, quelle que soit votre IA.",
		ofTheCahut: "Skills de cette Cahut",
		ofTheCahutHelp: "Ses administrateurs les écrivent, et les IA de chaque personne de cette Cahut les lisent.",
		noneOfTheCahut: "Cette Cahut n'a pas encore de skill.",
		noneOfTheCahutMember: "Cette Cahut n'a pas encore de skill. Ses administrateurs les écrivent.",
		mine: "Mes skills",
		mineHelp: "Personne d'autre que vous et vos IA ne les lit, pas même les administrateurs.",
		noneOfMine: "Vous n'avez pas encore de skill.",
		builtin: "Fourni avec Cahut",
		builtinHelp: "Le skill du plugin pour Claude, que chaque Cahut donne à ses IA. Il leur dit comment publier des Windows.",
		readOnly: "Lecture seule",
		downloadZip: "Télécharger son zip",
		write: "Écrire un skill",
		import: "Importer un zip",
		version: (n) => `Version ${n}`,
		written: (when) => `écrite ${when}`,
		everyWindow: "Pour chaque Window",
		nameIt: (name) => `Nommez un skill à votre IA, comme dans « Suis le skill ${name} », et elle le lit.`,
		pasteHelp: "Si votre IA ne pense pas aux skills de cette Cahut, collez cette ligne dans ses instructions :",
		pasteLine: "Avant chaque tâche, appelle l'outil find_skills de Cahut pour voir si mon équipe a un skill qui s'y applique, puis suis-le.",
		pasteLineIndividual: "Avant chaque tâche, appelle l'outil find_skills de Cahut pour voir si l'un de mes skills s'y applique, puis suis-le.",
		theLine: "la ligne",
		aisSee: "Ce que voient les IA",
		shown: (n, of) => `${n} sur ${plural(of, "skill", "skills")}`,
		aisSeeHelp: (bytes, budget) =>
			`Chaque IA branchée sur cette Cahut reçoit ces lignes dans la description de find_skills, dès le début de chaque conversation. Elles prennent ${bytes} octets sur ${budget}.`,
		left: (names) => `Laissés de côté, faute de place : ${names}. Les IA les trouvent quand même avec find_skills.`,
	},
});

/** The name of a skill without the prefix of its scope, which find_skills adds when the person reads it in two scopes. */
export function bareName(s: Pick<Summary, "skill" | "scope">): string {
	return s.skill.startsWith(`${s.scope}:`) ? s.skill.slice(s.scope.length + 1) : s.skill;
}

/** The page of a skill in the console, which names its scope: get_skill then reads it without doubt. */
export function skillPage(s: Pick<Summary, "skill" | "scope">): string {
	return `/skills/${s.scope}/${bareName(s)}`;
}

/** The skills that the person reads: those of the Cahut, their own, and the skill of the plugin. */
export function SkillsPage() {
	const api = useConsole();
	const { data, error } = useLoad(() => api.cahut("find_skills", {}));
	if (error) return <Failure message={error} />;
	if (!data) return null;
	const of = (scope: string) => data.skills.filter((s) => s.scope === scope);
	// The example names a skill that the person reads, if they read one.
	const example = bareName(data.skills.find((s) => s.scope !== "builtin") ?? { skill: "relance-facture", scope: "cahut" });
	return (
		<>
			<PageHead label={t.label} title={t.title} lead={api.individual ? t.leadIndividual : t.lead} />
			{!api.individual && (
				<Part
					title={t.ofTheCahut}
					help={t.ofTheCahutHelp}
					scope="cahut"
					writes={api.isAdmin}
					skills={of("cahut")}
					none={api.isAdmin ? t.noneOfTheCahut : t.noneOfTheCahutMember}
				>
					{api.isAdmin && of("cahut").length > 0 && <WhatAIsSee />}
				</Part>
			)}
			<Part title={t.mine} help={t.mineHelp} scope="personal" writes skills={of("personal")} none={t.noneOfMine} />
			{of("builtin").length > 0 && (
				<Part
					title={t.builtin}
					help={t.builtinHelp}
					scope="builtin"
					skills={of("builtin")}
					actions={
						<a class="c-btn c-btn--secondary c-btn--sm" href="/-/cahut-skill.zip" download>
							<Download class="i" size={14} aria-hidden="true" />
							{t.downloadZip}
						</a>
					}
				/>
			)}
			<section class="section c-card panel">
				<p class="c-small">{t.nameIt(example)}</p>
				<p class="c-small c-muted">{t.pasteHelp}</p>
				<CopyField wrap label={t.theLine} value={api.individual ? t.pasteLineIndividual : t.pasteLine} />
			</section>
		</>
	);
}

/** A part of the list: its skills, and the buttons to write or import one when the person may. */
function Part(props: {
	title: string;
	help: string;
	scope: string;
	writes?: boolean;
	skills: Summary[];
	none?: string;
	/** More buttons, beside those to write and import. */
	actions?: ComponentChildren;
	/** What comes after its skills. */
	children?: ComponentChildren;
}) {
	return (
		<section class="section c-card panel">
			<div class="skills-head">
				<div>
					<h2 class="c-h4">{props.title}</h2>
					<p class="c-small c-muted">{props.help}</p>
				</div>
				<div class="actions">
					{props.writes && (
						<>
							<Link class="c-btn c-btn--secondary c-btn--sm" href={`/skills/new/${props.scope}`}>
								<PenLine class="i" size={14} aria-hidden="true" />
								{t.write}
							</Link>
							<Link class="c-btn c-btn--secondary c-btn--sm" href={`/skills/import/${props.scope}`}>
								<FileUp class="i" size={14} aria-hidden="true" />
								{t.import}
							</Link>
						</>
					)}
					{props.actions}
				</div>
			</div>
			{props.skills.length === 0 ? (
				<p class="c-muted">{props.none}</p>
			) : (
				<ul class="skill-list">
					{props.skills.map((s) => (
						<li key={s.skill} class="skill-row">
							<div class="skill-row__text">
								<Link class="c-link skill-row__name" href={skillPage(s)}>
									{bareName(s)}
								</Link>
								<span class="c-small c-muted">{s.description}</span>
							</div>
							<span class="c-small c-muted skill-row__meta">
								{s.scope === "builtin" ? (
									<span class="c-badge c-badge--neutral">{t.readOnly}</span>
								) : (
									<>
										{s.every_window && <span class="c-badge c-badge--neutral">{t.everyWindow}</span>}
										{t.version(s.number)}
										{s.written_at && ` · ${t.written(agoInText(s.written_at))}`}
									</>
								)}
							</span>
						</li>
					))}
				</ul>
			)}
			{props.children}
		</section>
	);
}

/** The catalog of the skills of the Cahut, as AIs receive it in the description of find_skills, for its administrators. */
function WhatAIsSee() {
	const api = useConsole();
	const { data, error } = useLoad(() => api.cahut("preview_skill_catalog", {}));
	if (error) return <Failure message={error} />;
	if (!data) return null;
	const total = data.catalog.split("\n- ").length - 1 + data.left.length;
	return (
		<details class="ais-see">
			<summary>
				<ChevronRight class="i ais-see__chevron" size={14} aria-hidden="true" />
				<Eye class="i" size={14} aria-hidden="true" />
				{t.aisSee} · {t.shown(total - data.left.length, total)}
			</summary>
			<div class="ais-see__body">
				<p class="c-small c-muted">{t.aisSeeHelp(data.bytes.toLocaleString(locale), data.budget.toLocaleString(locale))}</p>
				<pre class="skill-text">{data.catalog}</pre>
				{data.left.length > 0 && <p class="c-small c-muted">{t.left(data.left.join(", "))}</p>}
			</div>
		</details>
	);
}
