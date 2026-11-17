// The page of a skill: its SKILL.md as AIs read it, its files, its two
// versions, and what the person may do with it: write a new version, roll
// back, export it as a zip, or delete it. See
// docs/adr/0064-the-skills-of-a-cahut.md.
import { Download, FileArchive, PenLine } from "lucide-preact";
import { useId, useState } from "preact/hooks";
import { explain } from "../api";
import type { Skill, SkillFileInfo } from "../api.gen";
import { agoInText, bytes, ConfirmButton, Failure, PageHead } from "../components";
import { useConsole } from "../context";
import { words } from "../i18n";
import { useLoad } from "../load";
import { Link, navigate } from "../router";

const t = words({
	en: {
		scopes: { cahut: "Skill of this Cahut", personal: "My skill", builtin: "Comes with Cahut" } as Record<string, string>,
		version: (n: number) => `Version ${n}`,
		written: (when: string) => `written ${when}`,
		by: (handle: string) => `by ${handle}`,
		byGone: "by a person no longer in this Cahut",
		via: { console: "in the console", import: "from a zip" } as Record<string, string>,
		text: "SKILL.md",
		textHelp: "What AIs read when they follow this skill.",
		files: "Files",
		filesHelp: "AIs read a file when SKILL.md tells them to.",
		noFiles: "No file besides SKILL.md.",
		download: "Download",
		downloadFile: (path: string) => `Download ${path}`,
		write: "Write a new version",
		versions: "Versions",
		versionsHelp: "Cahut keeps the current version and the one before. Going back can be undone: the version you leave becomes the previous one.",
		onlyVersion: "This is its first version.",
		previous: (n: number) => `Version ${n} is kept as the previous one.`,
		rollBack: (n: number) => `Roll back to version ${n}`,
		rollBackQuestion: (n: number) => `Bring back version ${n}? AIs follow it from their next conversation.`,
		rollBackConfirm: "Roll back",
		export: "Export",
		exportHelp: "A zip of the skill, in a folder of its name, as Claude and other AI clients install it.",
		delete: "Delete",
		deleteHelp: "Deleting a skill takes both its versions and its files, for good.",
		deleteQuestion: "Delete this skill and both its versions for good? This cannot be undone.",
		deleteConfirm: "Delete for good",
		adminsChange: "The administrators of this Cahut change its skills.",
		comesWith: "It comes with this version of Cahut, and changes with its updates.",
		allSkills: "All skills",
		everyWindow: "For every Window",
		everyWindowHelp: "AIs must read this skill before they publish or update a Window.",
		everyWindowLimits: "Three skills of this Cahut at most, each with a SKILL.md of 16 KiB at most.",
	},
	fr: {
		scopes: { cahut: "Skill de cette Cahut", personal: "Mon skill", builtin: "Fourni avec Cahut" },
		version: (n) => `Version ${n}`,
		written: (when) => `écrite ${when}`,
		by: (handle) => `par ${handle}`,
		byGone: "par une personne qui n'est plus dans cette Cahut",
		via: { console: "dans la console", import: "depuis un zip" },
		text: "SKILL.md",
		textHelp: "Ce que lisent les IA qui suivent ce skill.",
		files: "Fichiers",
		filesHelp: "Une IA lit un fichier quand le SKILL.md le lui dit.",
		noFiles: "Aucun fichier en dehors du SKILL.md.",
		download: "Télécharger",
		downloadFile: (path) => `Télécharger ${path}`,
		write: "Écrire une nouvelle version",
		versions: "Versions",
		versionsHelp: "Cahut garde la version actuelle et la précédente. Revenir en arrière s'annule : la version que vous quittez devient la précédente.",
		onlyVersion: "C'est sa première version.",
		previous: (n) => `La version ${n} est gardée comme précédente.`,
		rollBack: (n) => `Revenir à la version ${n}`,
		rollBackQuestion: (n) => `Revenir à la version ${n} ? Les IA la suivent dès leur prochaine conversation.`,
		rollBackConfirm: "Revenir",
		export: "Exporter",
		exportHelp: "Un zip du skill, dans un dossier à son nom, tel que Claude et les autres clients d'IA l'installent.",
		delete: "Supprimer",
		deleteHelp: "Supprimer un skill emporte ses deux versions et ses fichiers, pour de bon.",
		deleteQuestion: "Supprimer définitivement ce skill et ses deux versions ? Cela ne s'annule pas.",
		deleteConfirm: "Supprimer définitivement",
		adminsChange: "Les administrateurs de cette Cahut changent ses skills.",
		comesWith: "Il vient avec cette version de Cahut, et change avec ses mises à jour.",
		allSkills: "Tous les skills",
		everyWindow: "Pour chaque Window",
		everyWindowHelp: "Les IA doivent lire ce skill avant de publier ou de modifier une Window.",
		everyWindowLimits: "Trois skills de cette Cahut au plus, chacun avec un SKILL.md de 16 Kio au plus.",
	},
});

/** Gives the browser a file to save, from its content as the API gives it. */
function save(name: string, content: string, encoding: string | undefined, type: string) {
	const data = encoding === "base64" ? Uint8Array.from(atob(content), (c) => c.charCodeAt(0)) : content;
	const url = URL.createObjectURL(new Blob([data], { type }));
	const a = document.createElement("a");
	a.href = url;
	a.download = name;
	a.click();
	setTimeout(() => URL.revokeObjectURL(url), 1000);
}

/** Says who wrote the current version of a skill, when, and from where. */
function written(s: Skill): string {
	const parts = [t.version(s.number ?? 1)];
	if (s.written_at) {
		const who = s.written_by ? t.by(s.written_by) : t.byGone;
		parts.push([t.written(agoInText(s.written_at)), who, s.via && t.via[s.via]].filter(Boolean).join(" "));
	}
	return parts.join(" · ");
}

/** The skill named name in scope, as the person reads it. */
export function SkillPage({ scope, name }: { scope: string; name: string }) {
	const api = useConsole();
	const ref = `${scope}:${name}`;
	const { data, error, reload } = useLoad(() => api.cahut("get_skill", { skill: ref }), [ref]);
	const [failure, setFailure] = useState("");
	const back = (
		<p class="c-small">
			<Link class="c-link" href="/skills">
				{t.allSkills}
			</Link>
		</p>
	);
	if (error) {
		return (
			<>
				{back}
				<Failure message={error} />
			</>
		);
	}
	if (!data) return null;
	const s = data;
	const writes = scope === "personal" || (scope === "cahut" && api.isAdmin);
	const act = async (run: () => Promise<unknown>) => {
		setFailure("");
		try {
			await run();
		} catch (err) {
			setFailure(explain(err));
		}
	};
	return (
		<>
			{back}
			<PageHead label={t.scopes[scope] ?? scope} title={<span class="skill-title">{name}</span>} lead={s.description}>
				{written(s)}
			</PageHead>
			{failure && <Failure message={failure} />}
			<div class="layout">
				<section>
					<div class="c-card panel aside">
						<h2 class="c-h4 c-mono">{t.text}</h2>
						<p class="c-small c-muted">{t.textHelp}</p>
						<pre class="skill-text c-mono">{s.text}</pre>
					</div>
					<Files skill={ref} files={s.files ?? []} onFailure={setFailure} />
				</section>
				<aside class="aside">
					{scope === "cahut" && api.isAdmin && (
						<EveryWindow
							on={!!s.every_window}
							onChange={(on) => act(async () => (await api.cahut("write_skill", { scope: "cahut", skill: name, every_window: on }), reload()))}
						/>
					)}
					{writes && (
						<div class="c-card panel aside">
							<h2 class="c-h4">{t.versions}</h2>
							<p class="c-small c-muted">{t.versionsHelp}</p>
							<p class="c-small">{s.previous ? t.previous(s.previous) : t.onlyVersion}</p>
							<div class="actions">
								<Link class="c-btn c-btn--secondary c-btn--sm" href={`/skills/${scope}/${name}/write`}>
									<PenLine class="i" size={14} aria-hidden="true" />
									{t.write}
								</Link>
								{s.previous && (
									<ConfirmButton
										label={t.rollBack(s.previous)}
										question={t.rollBackQuestion(s.previous)}
										confirm={t.rollBackConfirm}
										onConfirm={() => act(async () => (await api.cahut("rollback_skill", { skill: ref }), reload()))}
									/>
								)}
							</div>
						</div>
					)}
					<div class="c-card panel aside">
						<h2 class="c-h4">{t.export}</h2>
						<p class="c-small c-muted">{t.exportHelp}</p>
						{!writes && <p class="c-small c-muted">{scope === "builtin" ? t.comesWith : t.adminsChange}</p>}
						<div>
							<button
								class="c-btn c-btn--secondary c-btn--sm"
								type="button"
								onClick={() =>
									act(async () => {
										const out = await api.cahut("export_skill", { skill: ref });
										save(`${out.name}.zip`, out.zip, "base64", "application/zip");
									})
								}
							>
								<FileArchive class="i" size={14} aria-hidden="true" />
								{t.export}
							</button>
						</div>
					</div>
					{writes && (
						<div class="c-card panel aside">
							<h2 class="c-h4">{t.delete}</h2>
							<p class="c-small c-muted">{t.deleteHelp}</p>
							<div>
								<ConfirmButton
									label={t.delete}
									question={t.deleteQuestion}
									confirm={t.deleteConfirm}
									danger
									onConfirm={() =>
										act(async () => {
											await api.cahut("delete_skill", { skill: ref });
											navigate("/skills");
										})
									}
								/>
							</div>
						</div>
					)}
				</aside>
			</div>
		</>
	);
}

/** The mark of a skill of the Cahut for every Window, which its administrators give or take off. */
function EveryWindow(props: { on: boolean; onChange: (on: boolean) => Promise<void> }) {
	const help = useId();
	const [busy, setBusy] = useState(false);
	return (
		<div class="c-card panel aside">
			<label class="c-check">
				<input
					type="checkbox"
					checked={props.on}
					disabled={busy}
					aria-describedby={help}
					onChange={async (e) => {
						setBusy(true);
						await props.onChange(e.currentTarget.checked);
						setBusy(false);
					}}
				/>
				{t.everyWindow}
			</label>
			<div id={help}>
				<p class="c-small">{t.everyWindowHelp}</p>
				<p class="c-small c-muted">{t.everyWindowLimits}</p>
			</div>
		</div>
	);
}

/** The files of a skill, each with its size and a button to download it. */
function Files(props: { skill: string; files: SkillFileInfo[]; onFailure: (message: string) => void }) {
	const api = useConsole();
	const download = async (f: SkillFileInfo) => {
		props.onFailure("");
		try {
			const out = await api.cahut("get_skill", { skill: props.skill, file: f.path });
			save(f.path.split("/").pop() ?? f.path, out.content ?? "", out.encoding, f.type);
		} catch (err) {
			props.onFailure(explain(err));
		}
	};
	return (
		<div class="c-card panel aside">
			<h2 class="c-h4">{t.files}</h2>
			<p class="c-small c-muted">{t.filesHelp}</p>
			{props.files.length === 0 ? (
				<p class="c-small c-muted">{t.noFiles}</p>
			) : (
				<ul class="skill-list">
					{props.files.map((f) => (
						<li key={f.path} class="skill-row">
							<span class="c-mono c-small skill-file">{f.path}</span>
							<span class="skill-row__meta c-small c-muted">
								{bytes(f.size)}
								<button
									class="c-btn c-btn--ghost c-btn--sm"
									type="button"
									aria-label={t.downloadFile(f.path)}
									onClick={() => download(f)}
								>
									<Download class="i" size={14} aria-hidden="true" />
									{t.download}
								</button>
							</span>
						</li>
					))}
				</ul>
			)}
		</div>
	);
}
