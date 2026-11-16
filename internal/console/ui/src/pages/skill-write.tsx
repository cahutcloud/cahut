// Writing a skill in the console: its SKILL.md in a field, and the files
// to add, replace or remove; a new skill starts from a header to fill in.
// And importing one: the zip of a skill, as AI clients export them, read in
// the browser and sent whole to import_skill. See
// docs/adr/0064-the-skills-of-a-cahut.md.
import { FileArchive, FilePlus, FileUp, X } from "lucide-preact";
import { useId, useState } from "preact/hooks";
import { explain } from "../api";
import type { Skill, SkillFileInfo } from "../api.gen";
import { bytes, Failure, Notice, PageHead } from "../components";
import { useConsole } from "../context";
import { words } from "../i18n";
import { useLoad } from "../load";
import { say } from "../refusals";
import { Link, navigate } from "../router";

const t = words({
	en: {
		scopes: { cahut: "Skills of this Cahut", personal: "My skills" } as Record<string, string>,
		newTitle: "Write a skill",
		editTitle: (name: string) => `A new version of ${name}`,
		editLead: "Cahut keeps the version before, and you can roll back to it.",
		text: "SKILL.md",
		textHelp: "A header between two lines ---, with the name of the skill and its description, then what the AI should do.",
		descriptionHelp: "AIs see the first 160 characters of the description. Say what the skill does and when to follow it.",
		descriptionCount: (n: number) => `This one holds ${n}.`,
		files: "Files",
		filesHelp:
			"A file at the same path replaces the one the skill holds. Put what the AI reads in references/, and what it copies, such as fonts or a logo, in assets/.",
		noFiles: "No file besides SKILL.md.",
		add: "Add files",
		remove: "Remove",
		keep: "Keep",
		removeFile: (path: string) => `Remove ${path}`,
		keepFile: (path: string) => `Keep ${path}`,
		removed: "Removed",
		replaced: "Replaced",
		dropFile: (path: string) => `Leave out ${path}`,
		path: (name: string) => `Path of ${name}`,
		save: "Save a new version",
		create: "Write the skill",
		cancel: "Cancel",
		nameTaken: (name: string) => `There is a skill named ${name} here already: open it to write a new version.`,
		tooLargeForPage: "This page sends 1 MiB at most at once: put the skill in a zip, and import it.",
		template: "---\nname: my-skill\ndescription: What this skill does, and when to follow it.\n---\n\n# My skill\n\nWrite here what the AI should do, step by step.\n",
		importTitle: "Import a zip",
		importLead:
			"The zip of a skill, as Claude and other AI clients export it: a SKILL.md at its root or in a single folder, with its files. Its header gives the name of the skill.",
		importReplace: "A skill of the same name gets a new version, with the files of the zip only. The version before stays, and you can roll back to it.",
		zip: "Zip",
		chooseZip: "Choose a zip",
		noZip: "No zip chosen yet.",
		importButton: "Import",
		allSkills: "All skills",
	},
	fr: {
		scopes: { cahut: "Skills de cette Cahut", personal: "Mes skills" },
		newTitle: "Écrire un skill",
		editTitle: (name) => `Une nouvelle version de ${name}`,
		editLead: "Cahut garde la version d'avant, et vous pouvez y revenir.",
		text: "SKILL.md",
		textHelp: "Un en-tête entre deux lignes ---, avec le nom du skill et sa description, puis ce que l'IA doit faire.",
		descriptionHelp: "Les IA voient les 160 premiers caractères de la description. Dites ce que fait le skill et quand le suivre.",
		descriptionCount: (n) => `Celle-ci en compte ${n}.`,
		files: "Fichiers",
		filesHelp:
			"Un fichier au même chemin remplace celui du skill. Mettez ce que l'IA lit dans references/, et ce qu'elle copie, comme des polices ou un logo, dans assets/.",
		noFiles: "Aucun fichier en dehors du SKILL.md.",
		add: "Ajouter des fichiers",
		remove: "Retirer",
		keep: "Garder",
		removeFile: (path) => `Retirer ${path}`,
		keepFile: (path) => `Garder ${path}`,
		removed: "Retiré",
		replaced: "Remplacé",
		dropFile: (path) => `Laisser ${path} de côté`,
		path: (name) => `Chemin de ${name}`,
		save: "Enregistrer une nouvelle version",
		create: "Écrire le skill",
		cancel: "Annuler",
		nameTaken: (name) => `Il y a déjà un skill nommé ${name} ici : ouvrez-le pour en écrire une nouvelle version.`,
		tooLargeForPage: "Cette page envoie 1 Mio au plus à la fois : mettez le skill dans un zip, et importez-le.",
		template:
			"---\nname: mon-skill\ndescription: Ce que fait ce skill, et quand le suivre.\n---\n\n# Mon skill\n\nÉcrivez ici ce que l'IA doit faire, étape par étape.\n",
		importTitle: "Importer un zip",
		importLead:
			"Le zip d'un skill, tel que Claude et les autres clients d'IA l'exportent : un SKILL.md à sa racine ou dans un seul dossier, avec ses fichiers. Son en-tête donne le nom du skill.",
		importReplace:
			"Un skill du même nom reçoit une nouvelle version, avec les fichiers du zip seulement. La version d'avant reste, et vous pouvez y revenir.",
		zip: "Zip",
		chooseZip: "Choisir un zip",
		noZip: "Aucun zip choisi pour l'instant.",
		importButton: "Importer",
		allSkills: "Tous les skills",
	},
});

/** The largest call that the API of the console reads, and the one it reads for import_skill. */
const maxCall = 1 << 20;
const maxImport = 6 << 20;

/**
 * Reads the name and the description of a SKILL.md from its header, for
 * the help under its field and the name of a new skill: a value on its line,
 * in quotes or not. The server reads the header in full, and is the judge.
 */
export function headerOf(text: string): { name?: string; description?: string } {
	const m = text.replace(/^﻿/, "").match(/^---\r?\n([\s\S]*?)\r?\n---[ \t]*(\r?\n|$)/);
	if (!m) return {};
	const value = (key: string): string | undefined => {
		const line = m[1].match(new RegExp(`^${key}:[ \\t]*(.*?)[ \\t]*\\r?$`, "m"));
		if (!line) return undefined;
		const v = line[1];
		if (/^".*"$/.test(v)) {
			try {
				return JSON.parse(v) as string;
			} catch {
				return v.slice(1, -1);
			}
		}
		if (/^'.*'$/.test(v)) return v.slice(1, -1).replace(/''/g, "'");
		return v.replace(/[ \t]+#.*$/, "");
	};
	return { name: value("name"), description: value("description") };
}

/** Reads a file of the browser as base64. */
function base64Of(file: Blob): Promise<string> {
	return new Promise((resolve, reject) => {
		const reader = new FileReader();
		reader.onload = () => resolve(String(reader.result).replace(/^data:[^,]*,/, ""));
		reader.onerror = () => reject(reader.error);
		reader.readAsDataURL(file);
	});
}

/** The size of a call to the API, as it travels. */
function sizeOf(input: unknown): number {
	return new TextEncoder().encode(JSON.stringify(input)).length;
}

interface Added {
	id: number;
	file: File;
	path: string;
}

/** The last id given to a file added on the page. */
let lastId = 0;

/**
 * Says why the person may not write in scope, or nothing: an individual
 * Cahut has no skills of the Cahut, and only its administrators write them.
 * The API judges all the same.
 */
function barred(api: { individual: boolean; isAdmin: boolean }, scope: string): string | undefined {
	if (scope !== "cahut") return undefined;
	if (api.individual) return say("skill_scope_individual", null);
	if (!api.isAdmin) return say("skill_admins_only", null);
	return undefined;
}

/** The head of a page that writes in scope, with why the person may not, in place of its form. */
function Barred(props: { scope: string; title: string; why: string }) {
	return (
		<>
			<PageHead label={t.scopes[props.scope] ?? props.scope} title={props.title} />
			<Notice>{props.why}</Notice>
			<p class="c-small">
				<Link class="c-link" href="/skills">
					{t.allSkills}
				</Link>
			</p>
		</>
	);
}

/** Writes a new skill in scope, or a new version of the skill name. */
export function SkillWritePage({ scope, name }: { scope: string; name?: string }) {
	const api = useConsole();
	const why = barred(api, scope);
	if (why) return <Barred scope={scope} title={name ? t.editTitle(name) : t.newTitle} why={why} />;
	return <Write scope={scope} name={name} />;
}

function Write({ scope, name }: { scope: string; name?: string }) {
	const api = useConsole();
	const ref = name && `${scope}:${name}`;
	const { data, error } = useLoad<Skill | string[]>(
		() =>
			ref
				? api.cahut("get_skill", { skill: ref })
				: api.cahut("find_skills", {}).then((f) => f.skills.filter((s) => s.scope === scope).map((s) => s.skill.replace(/^[a-z]+:/, ""))),
		[ref],
	);
	if (error) return <Failure message={error} />;
	if (!data) return null;
	return Array.isArray(data) ? <Editor scope={scope} taken={data} /> : <Editor scope={scope} skill={data} name={name} />;
}

function Editor(props: { scope: string; name?: string; skill?: Skill; taken?: string[] }) {
	const api = useConsole();
	const id = useId();
	const held = props.skill?.files ?? [];
	const [text, setText] = useState(props.skill?.text ?? t.template);
	const [removed, setRemoved] = useState<string[]>([]);
	const [added, setAdded] = useState<Added[]>([]);
	const [busy, setBusy] = useState(false);
	const [failure, setFailure] = useState("");
	const header = headerOf(text);
	const description = header.description !== undefined && !/^[|>]/.test(header.description) ? header.description : undefined;
	const replacing = (f: SkillFileInfo) => added.some((a) => a.path.toLowerCase() === f.path.toLowerCase());
	const back = props.name ? `/skills/${props.scope}/${props.name}` : "/skills";

	const add = (files: FileList | null) => {
		const picked = Array.from(files ?? []);
		setAdded((list) => [...list, ...picked.map((file) => ({ id: ++lastId, file, path: file.name }))]);
	};
	const submit = async (e: Event) => {
		e.preventDefault();
		setFailure("");
		// A new skill takes its name from its header: without one that the
		// page reads, the header is what to fix.
		if (!props.name && header.name === undefined) {
			setFailure(say("skill_header_invalid", { problem: "" }) ?? "");
			return;
		}
		const skill = props.name ?? header.name ?? "";
		if (!props.name && props.taken?.includes(skill)) {
			setFailure(t.nameTaken(skill));
			return;
		}
		// Files go as base64, a third larger: a call that cannot pass is
		// stopped before the browser reads them.
		const estimate = new TextEncoder().encode(text).length + added.reduce((n, a) => n + Math.ceil(a.file.size / 3) * 4 + a.path.length + 64, 0);
		if (estimate > maxCall) {
			setFailure(t.tooLargeForPage);
			return;
		}
		setBusy(true);
		try {
			const files = await Promise.all(added.map(async (a) => ({ path: a.path.trim(), content: await base64Of(a.file), encoding: "base64" })));
			const remove = removed.filter((p) => !added.some((a) => a.path.trim().toLowerCase() === p.toLowerCase()));
			const input = { scope: props.scope, skill, text, ...(files.length > 0 && { files }), ...(remove.length > 0 && { remove }) };
			if (sizeOf(input) > maxCall) {
				setFailure(t.tooLargeForPage);
				return;
			}
			const out = await api.cahut("write_skill", input);
			navigate(`/skills/${props.scope}/${out.skill}`);
		} catch (err) {
			setFailure(explain(err));
		} finally {
			setBusy(false);
		}
	};

	return (
		<>
			<PageHead
				label={t.scopes[props.scope] ?? props.scope}
				title={props.name ? t.editTitle(props.name) : t.newTitle}
				lead={props.name ? t.editLead : undefined}
			/>
			<form class="section skill-form" onSubmit={submit}>
				<div class="c-card panel aside">
					<label class="c-h4 c-mono" for={id}>
						{t.text}
					</label>
					<p class="c-small c-muted">{t.textHelp}</p>
					<textarea
						id={id}
						class="c-input c-mono skill-editor"
						rows={18}
						spellcheck={false}
						value={text}
						aria-describedby={`${id}-help`}
						onInput={(e) => setText(e.currentTarget.value)}
					/>
					<span class="c-field__help" id={`${id}-help`} aria-live="polite">
						{t.descriptionHelp} {description !== undefined && t.descriptionCount([...description].length)}
					</span>
				</div>
				<div class="c-card panel aside">
					<h2 class="c-h4">{t.files}</h2>
					<p class="c-small c-muted">{t.filesHelp}</p>
					{held.length === 0 && added.length === 0 && <p class="c-small c-muted">{t.noFiles}</p>}
					{(held.length > 0 || added.length > 0) && (
						<ul class="skill-list">
							{held.map((f) => {
								const gone = removed.includes(f.path);
								const replaced = replacing(f);
								return (
									<li key={f.path} class="skill-row">
										<span class={`c-mono c-small skill-file${gone ? " skill-file--gone" : ""}`}>{f.path}</span>
										<span class="skill-row__meta c-small c-muted">
											{replaced ? (
												<span class="c-badge c-badge--neutral">{t.replaced}</span>
											) : gone ? (
												<span class="c-badge c-badge--neutral">{t.removed}</span>
											) : (
												bytes(f.size)
											)}
											{!replaced && (
												<button
													class="c-btn c-btn--ghost c-btn--sm"
													type="button"
													aria-label={gone ? t.keepFile(f.path) : t.removeFile(f.path)}
													onClick={() => setRemoved((list) => (gone ? list.filter((p) => p !== f.path) : [...list, f.path]))}
												>
													{gone ? t.keep : t.remove}
												</button>
											)}
										</span>
									</li>
								);
							})}
							{added.map((a) => (
								<li key={a.id} class="skill-row">
									<input
										class="c-input c-mono skill-path"
										value={a.path}
										aria-label={t.path(a.file.name)}
										onInput={(e) => {
										const path = e.currentTarget.value;
										setAdded((list) => list.map((x) => (x.id === a.id ? { ...x, path } : x)));
									}}
									/>
									<span class="skill-row__meta c-small c-muted">
										{bytes(a.file.size)}
										<button
											class="c-btn c-btn--ghost c-btn--sm"
											type="button"
											aria-label={t.dropFile(a.path)}
											onClick={() => setAdded((list) => list.filter((x) => x.id !== a.id))}
										>
											<X class="i" size={14} aria-hidden="true" />
										</button>
									</span>
								</li>
							))}
						</ul>
					)}
					<div>
						<label class="c-btn c-btn--secondary c-btn--sm skill-add">
							<FilePlus class="i" size={14} aria-hidden="true" />
							{t.add}
							<input
								class="visually-hidden"
								type="file"
								multiple
								onChange={(e) => {
									add(e.currentTarget.files);
									e.currentTarget.value = "";
								}}
							/>
						</label>
					</div>
				</div>
				{failure && <Failure message={failure} />}
				<div class="actions">
					<button class="c-btn c-btn--primary" type="submit" disabled={busy}>
						{props.name ? t.save : t.create}
					</button>
					<Link class="c-btn c-btn--ghost" href={back}>
						{t.cancel}
					</Link>
				</div>
			</form>
		</>
	);
}

/** Imports the zip of a skill into scope. */
export function SkillImportPage({ scope }: { scope: string }) {
	const api = useConsole();
	const why = barred(api, scope);
	if (why) return <Barred scope={scope} title={t.importTitle} why={why} />;
	return <Import scope={scope} />;
}

function Import({ scope }: { scope: string }) {
	const api = useConsole();
	const id = useId();
	const [zip, setZip] = useState<File | null>(null);
	const [busy, setBusy] = useState(false);
	const [failure, setFailure] = useState("");
	const submit = async (e: Event) => {
		e.preventDefault();
		if (!zip) return;
		setFailure("");
		if (Math.ceil(zip.size / 3) * 4 > maxImport) {
			setFailure(say("skill_too_large", null) ?? "");
			return;
		}
		setBusy(true);
		try {
			const input = { scope, zip: await base64Of(zip) };
			if (sizeOf(input) > maxImport) {
				setFailure(say("skill_too_large", null) ?? "");
				return;
			}
			const out = await api.cahut("import_skill", input);
			navigate(`/skills/${scope}/${out.skill}`);
		} catch (err) {
			setFailure(explain(err));
		} finally {
			setBusy(false);
		}
	};
	return (
		<>
			<PageHead label={t.scopes[scope] ?? scope} title={t.importTitle} lead={t.importLead} />
			<form class="section c-card panel narrow skill-form" onSubmit={submit}>
				<p class="c-small c-muted">{t.importReplace}</p>
				<div class="c-field">
					<span class="c-field__label" id={id}>
						{t.zip}
					</span>
					<div class="actions">
						<label class="c-btn c-btn--secondary c-btn--sm skill-add">
							<FileArchive class="i" size={14} aria-hidden="true" />
							{t.chooseZip}
							<input
								class="visually-hidden"
								type="file"
								accept=".zip,application/zip"
								aria-describedby={id}
								onChange={(e) => (setZip(e.currentTarget.files?.[0] ?? null), setFailure(""))}
							/>
						</label>
						<span class="c-small c-muted skill-file">{zip ? `${zip.name} · ${bytes(zip.size)}` : t.noZip}</span>
					</div>
				</div>
				{failure && <Failure message={failure} />}
				<div class="actions">
					<button class="c-btn c-btn--primary" type="submit" disabled={busy || !zip}>
						<FileUp class="i" size={14} aria-hidden="true" />
						{t.importButton}
					</button>
					<Link class="c-btn c-btn--ghost" href="/skills">
						{t.cancel}
					</Link>
				</div>
			</form>
		</>
	);
}
