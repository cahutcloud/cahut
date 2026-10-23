// The house style of the Cahut, for its administrators: the name of the
// organization, its logo, its colors, its fonts and what it asks of AIs,
// with a preview before saving, the previous version, and the neutral
// house style. See docs/adr/0022-house-style-of-each-cahut.md.
import { ExternalLink, Trash2, Upload } from "lucide-preact";
import { useEffect, useMemo, useState } from "preact/hooks";
import { explain } from "../api";
import type { Font, FontInput, HouseStyle, SetInput } from "../api.gen";
import { ConfirmButton, Failure, Notice, PageHead } from "../components";
import { useConsole } from "../context";
import { useLoad } from "../load";

type ColorName = "text" | "background" | "primary" | "secondary";
const colorNames: { key: ColorName; label: string; help: string }[] = [
	{ key: "text", label: "Text", help: "Body text." },
	{ key: "background", label: "Background", help: "The background of pages and slides." },
	{ key: "primary", label: "Primary", help: "Headings of slides, links, buttons, slides that open a part." },
	{ key: "secondary", label: "Secondary", help: "Accents, quotes, the progress of a deck." },
];

const fallbacks: Record<string, { label: string; stack: string }> = {
	"sans-serif": { label: "Sans serif", stack: 'system-ui, -apple-system, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif' },
	serif: { label: "Serif", stack: 'Georgia, "Times New Roman", Times, serif' },
	monospace: { label: "Monospace", stack: 'ui-monospace, SFMono-Regular, Menlo, Consolas, "Liberation Mono", monospace' },
};

/** A file of a font: one the house style keeps, or one to send. */
interface FileDraft {
	key: string;
	name: string;
	size: number;
	weight: number;
	maxWeight?: number;
	style: string;
	keep?: string; // the name of a file the house style has
	url?: string;
	content?: string; // base64
	buffer?: ArrayBuffer;
}

interface FontDraft {
	family: string;
	fallback: string;
	files: FileDraft[];
}

/** A logo to send, or to remove when content is empty. */
interface LogoDraft {
	content: string; // base64
	preview: string; // a data: URL
}

interface Draft {
	name: string;
	colors: Record<ColorName, string>;
	chart: string;
	heading: FontDraft;
	body: FontDraft;
	logo?: LogoDraft;
	logoInverse?: LogoDraft;
	guidance: string;
}

function fontDraft(f: Font): FontDraft {
	return {
		family: f.family ?? "",
		fallback: f.fallback,
		files: (f.files ?? []).map((file) => ({
			key: file.name,
			name: file.name,
			size: file.size,
			weight: file.weight,
			maxWeight: file.max_weight,
			style: file.style,
			keep: file.name,
			url: file.url,
		})),
	};
}

function draftOf(h: HouseStyle): Draft {
	return {
		name: h.name === location.hostname ? "" : h.name,
		colors: { text: h.colors.text, background: h.colors.background, primary: h.colors.primary, secondary: h.colors.secondary },
		chart: h.own_chart_colors ? h.chart_colors.join(", ") : "",
		heading: fontDraft(h.fonts.heading),
		body: fontDraft(h.fonts.body),
		guidance: h.guidance ?? "",
	};
}

export function HouseStylePage() {
	const api = useConsole();
	const { data, error, reload } = useLoad(() => api.cahut("get_house_style", {}));
	if (!api.isAdmin) return <Notice>Only the administrators of this Cahut set its house style.</Notice>;
	if (error) return <Failure message={error} />;
	if (!data) return null;
	return <Editor key={`${data.set_at ?? "neutral"}-${data.name}-${data.active}`} style={data} onSaved={reload} />;
}

function Editor({ style, onSaved }: { style: HouseStyle; onSaved: () => void }) {
	const api = useConsole();
	const [draft, setDraft] = useState(() => draftOf(style));
	const [touched, setTouched] = useState<Set<string>>(new Set());
	const [failure, setFailure] = useState("");
	const [warnings, setWarnings] = useState<string[]>(style.warnings ?? []);
	const [busy, setBusy] = useState(false);

	const change = <K extends keyof Draft>(field: K, value: Draft[K], what: string = field) => {
		setDraft((d) => ({ ...d, [field]: value }));
		setTouched((t) => new Set(t).add(what));
	};
	const run = async (call: () => Promise<HouseStyle>) => {
		setBusy(true);
		setFailure("");
		try {
			const h = await call();
			setWarnings(h.warnings ?? []);
			onSaved();
		} catch (err) {
			setFailure(explain(err));
		} finally {
			setBusy(false);
		}
	};
	const save = (e: Event) => {
		e.preventDefault();
		run(() => api.cahut("set_house_style", inputOf(draft, touched)));
	};
	const version = style.set_at ? `?v=${encodeURIComponent(style.set_at)}` : "";

	return (
		<>
			<PageHead
				label="This Cahut"
				title="House style"
				actions={
					<span class="house__templates">
						<span class="c-small c-muted">Templates</span>
						{["document", "deck", "dashboard"].map((kind) => (
							<a class="c-btn c-btn--secondary c-btn--sm" href={`/-/templates/${kind}`} target="_blank" rel="noopener" key={kind}>
								{kind.charAt(0).toUpperCase() + kind.slice(1)}
								<ExternalLink class="i" size={14} aria-hidden="true" />
							</a>
						))}
					</span>
				}
			>
				The look of your organization. AIs read it before they make a Window, and every Window that links it takes a change at once, as do the
				pages that readers see, such as the page of a Key.
			</PageHead>
			{style.set_at && (
				<p class="c-small c-muted section">
					Set on {new Date(style.set_at).toLocaleString("en-GB", { dateStyle: "long", timeStyle: "short" })}
					{style.set_by ? ` for ${style.set_by}` : ""}.
				</p>
			)}
			<section class="section c-card panel house__use">
				<div class="house__use-text">
					<h2 class="c-h4" id="house-use">AIs use this house style</h2>
					<p class="c-small c-muted">
						{style.active
							? "AIs read it before they make a Window, and start from its templates."
							: "Inactive: AIs get nothing of it, and make Windows in the style the user asks for, such as guidelines of your own given to them as a skill. The pages that readers see, and the Windows that link the style sheet, keep it."}
					</p>
				</div>
				<button
					class="c-toggle"
					type="button"
					role="switch"
					aria-checked={style.active}
					aria-labelledby="house-use"
					disabled={busy}
					onClick={() => run(() => api.cahut("set_house_style", { active: !style.active }))}
				/>
			</section>
			<form class="house section" onSubmit={save}>
				<div class="house__form">
					<section class="c-card panel form">
						<h2 class="c-h4">Organization</h2>
						<label class="c-field">
							<span class="c-field__label">Name</span>
							<input
								class="c-input"
								placeholder={location.hostname}
								maxLength={100}
								value={draft.name}
								onInput={(e) => change("name", e.currentTarget.value)}
							/>
							<span class="c-field__help">At the top of the pages that readers see, and in the templates. The domain until you give one.</span>
						</label>
						<div class="form__row">
							<LogoField
								label="Logo"
								help="For light backgrounds. SVG, PNG or WebP, 512 KB at most."
								current={style.has_logo ? `/-/house-style/logo${version}` : ""}
								draft={draft.logo}
								onChange={(l) => change("logo", l)}
							/>
							<LogoField
								label="Logo for dark backgrounds"
								help="On slides in the primary color. Without it, they show the name."
								current={style.has_logo_inverse ? `/-/house-style/logo-inverse${version}` : ""}
								draft={draft.logoInverse}
								dark={draft.colors.primary}
								onChange={(l) => change("logoInverse", l)}
							/>
						</div>
					</section>

					<section class="c-card panel form">
						<h2 class="c-h4">Colors</h2>
						<div class="house__colors">
							{colorNames.map((c) => (
								<ColorField
									key={c.key}
									label={c.label}
									help={c.help}
									value={draft.colors[c.key]}
									onChange={(v) => change("colors", { ...draft.colors, [c.key]: v }, `colors.${c.key}`)}
								/>
							))}
						</div>
						<Contrasts colors={draft.colors} />
						<label class="c-field">
							<span class="c-field__label">Colors of charts</span>
							<input
								class="c-input c-mono"
								placeholder="Derived from the primary and secondary colors"
								value={draft.chart}
								onInput={(e) => change("chart", e.currentTarget.value)}
							/>
							<span class="c-field__help">
								Up to 8 colors, in order, separated by commas. Empty, Cahut derives them from the primary and secondary colors when you save.
							</span>
						</label>
						<Swatches colors={chartOf(draft, style)} />
					</section>

					<section class="c-card panel form">
						<h2 class="c-h4">Fonts</h2>
						<FontField label="Headings" font={draft.heading} onChange={(f) => change("heading", f)} />
						<FontField label="Body text" font={draft.body} onChange={(f) => change("body", f)} />
						<p class="c-small c-muted">
							The browsers of readers download these fonts from this Cahut. Upload only fonts whose license allows it on the web.
						</p>
					</section>

					<section class="c-card panel form">
						<h2 class="c-h4">Guidance for AIs</h2>
						<label class="c-field">
							<span class="c-field__label">What the organization asks of AIs</span>
							<textarea
								class="c-input house__guidance"
								maxLength={4000}
								rows={6}
								placeholder="Sentence case for titles. One idea per slide. Figures in thousands."
								value={draft.guidance}
								onInput={(e) => change("guidance", e.currentTarget.value)}
							/>
							<span class="c-field__help">
								Its tone of voice, its rules for titles or charts. AIs read it with the house style; readers never see it. {draft.guidance.length} / 4000
							</span>
						</label>
					</section>

					{failure && <Failure message={failure} />}
					{warnings.length > 0 && (
						<div class="c-alert c-alert--warning" role="status">
							<ul class="house__warnings">
								{warnings.map((w) => (
									<li key={w}>{w}</li>
								))}
							</ul>
						</div>
					)}
					<div class="house__actions">
						<button class="c-btn c-btn--primary" type="submit" disabled={busy || touched.size === 0}>
							Save the house style
						</button>
						{touched.size > 0 && (
							<button
								class="c-btn c-btn--ghost"
								type="button"
								onClick={() => {
									setDraft(draftOf(style));
									setTouched(new Set());
								}}
							>
								Discard the changes
							</button>
						)}
					</div>
				</div>
				<aside class="house__aside">
					<Preview draft={draft} style={style} version={version} />
				</aside>
			</form>

			<section class="section c-card panel">
				<h2 class="c-h4">Versions</h2>
				{style.previous ? (
					<div class="house__version">
						<p class="c-small">
							The previous version:{" "}
							{style.previous.neutral
								? "the neutral house style"
								: `set on ${new Date(style.previous.set_at).toLocaleDateString("en-GB", { dateStyle: "long" })}${style.previous.set_by ? ` for ${style.previous.set_by}` : ""}`}
							. Going back keeps this one as the previous version.
						</p>
						<ConfirmButton
							label="Go back to it"
							question="Go back to the previous house style?"
							confirm="Go back"
							onConfirm={() => run(() => api.cahut("rollback_house_style", {}))}
						/>
					</div>
				) : (
					<p class="c-small c-muted">No previous version yet.</p>
				)}
				{!style.neutral && (
					<div class="house__version">
						<p class="c-small">The neutral house style uses the fonts of the system, dark text on white, and a blue.</p>
						<ConfirmButton
							label="Back to neutral"
							question="Start again from the neutral house style? This one becomes the previous version."
							confirm="Back to neutral"
							onConfirm={() => run(() => api.cahut("set_house_style", { neutral: true }))}
						/>
					</div>
				)}
			</section>
		</>
	);
}

/** What the draft changes, as set_house_style takes it. */
function inputOf(d: Draft, touched: Set<string>): SetInput {
	const input: SetInput = {};
	if (touched.has("name")) input.name = d.name.trim();
	const colors: Partial<Record<ColorName, string>> = {};
	for (const c of colorNames) {
		if (touched.has(`colors.${c.key}`)) colors[c.key] = d.colors[c.key];
	}
	if (Object.keys(colors).length > 0) input.colors = colors;
	if (touched.has("chart")) input.chart_colors = d.chart.split(/[\s,;]+/).filter(Boolean);
	if (touched.has("heading")) input.heading_font = fontInput(d.heading);
	if (touched.has("body")) input.body_font = fontInput(d.body);
	if (touched.has("logo") && d.logo) input.logo = { content: d.logo.content, encoding: "base64" };
	if (touched.has("logoInverse") && d.logoInverse) input.logo_inverse = { content: d.logoInverse.content, encoding: "base64" };
	if (touched.has("guidance")) input.guidance = d.guidance;
	return input;
}

function fontInput(f: FontDraft): FontInput {
	return {
		family: f.family.trim(),
		fallback: f.fallback,
		files: f.files.map((file) => ({
			...(file.keep ? { keep: file.keep } : { content: file.content }),
			weight: file.weight,
			max_weight: file.maxWeight,
			style: file.style,
		})),
	};
}

// Colors, as the house style computes them. See internal/housestyle/color.go.

function parse(hex: string): [number, number, number] | null {
	let h = hex.trim().toLowerCase();
	if (!/^#([0-9a-f]{3}|[0-9a-f]{6})$/.test(h)) return null;
	if (h.length === 4) h = `#${h[1]}${h[1]}${h[2]}${h[2]}${h[3]}${h[3]}`;
	return [parseInt(h.slice(1, 3), 16), parseInt(h.slice(3, 5), 16), parseInt(h.slice(5, 7), 16)];
}

function luminance(c: [number, number, number]): number {
	const [r, g, b] = c.map((v) => {
		const s = v / 255;
		return s <= 0.04045 ? s / 12.92 : ((s + 0.055) / 1.055) ** 2.4;
	});
	return 0.2126 * r + 0.7152 * g + 0.0722 * b;
}

function contrast(a: string, b: string): number {
	const ca = parse(a);
	const cb = parse(b);
	if (!ca || !cb) return 0;
	const [hi, lo] = [luminance(ca), luminance(cb)].sort((x, y) => y - x);
	return (hi + 0.05) / (lo + 0.05);
}

function mix(a: string, b: string, t: number): string {
	const ca = parse(a);
	const cb = parse(b);
	if (!ca || !cb) return a;
	return `#${ca.map((v, i) => Math.round(t * v + (1 - t) * cb[i]).toString(16).padStart(2, "0")).join("")}`;
}

/** The text on a color: the text color or the background, whichever reads better. */
function on(color: string, text: string, background: string): string {
	return contrast(color, text) >= contrast(color, background) ? text : background;
}

function muted(text: string, background: string): string {
	for (let t = 0.55; t < 1; t += 0.05) {
		const m = mix(text, background, t);
		if (contrast(m, background) >= 4.5) return m;
	}
	return text;
}

function chartOf(d: Draft, style: HouseStyle): string[] {
	const own = d.chart.split(/[\s,;]+/).filter((c) => parse(c));
	return own.length > 0 ? own.slice(0, 8) : style.chart_colors;
}

function ColorField(props: { label: string; help: string; value: string; onChange: (v: string) => void }) {
	const valid = parse(props.value) !== null;
	return (
		<label class="c-field">
			<span class="c-field__label">{props.label}</span>
			<span class="house__color">
				<input
					class="house__picker"
					type="color"
					aria-label={`${props.label}, picker`}
					value={valid ? mix(props.value, props.value, 1) : "#000000"}
					onInput={(e) => props.onChange(e.currentTarget.value)}
				/>
				<input
					class="c-input c-mono"
					aria-label={props.label}
					value={props.value}
					maxLength={7}
					spellcheck={false}
					aria-invalid={!valid}
					onInput={(e) => props.onChange(e.currentTarget.value)}
				/>
			</span>
			<span class="c-field__help">{props.help}</span>
		</label>
	);
}

function Contrasts({ colors }: { colors: Record<ColorName, string> }) {
	// Text on the background below its need is refused; the others warn.
	const rows: { what: string; ratio: number; need: number; refused?: boolean }[] = [
		{ what: "Text on the background", ratio: contrast(colors.text, colors.background), need: 4.5, refused: true },
		{ what: "Primary on the background", ratio: contrast(colors.primary, colors.background), need: 3 },
		{ what: "Text on the primary color", ratio: contrast(on(colors.primary, colors.text, colors.background), colors.primary), need: 4.5 },
		{ what: "Secondary on the background", ratio: contrast(colors.secondary, colors.background), need: 3 },
	];
	const badge = (r: (typeof rows)[number]) =>
		r.ratio >= r.need ? ["c-badge--success", "Reads well"] : r.refused ? ["c-badge--danger", "Too low to save"] : ["c-badge--warning", "Hard to read"];
	return (
		<ul class="house__contrasts">
			{rows.map((r) => {
				const [variant, word] = badge(r);
				return (
					<li key={r.what}>
						<span>{r.what}</span>
						<span class="c-mono">{r.ratio.toFixed(1)}:1</span>
						<span class={`c-badge ${variant}`}>{word}</span>
					</li>
				);
			})}
		</ul>
	);
}

function Swatches({ colors }: { colors: string[] }) {
	return (
		<div class="house__swatches" aria-label="Colors of charts, in order">
			{colors.map((c, i) => (
				<span class="house__swatch" key={`${i}-${c}`} title={`${i + 1}: ${c}`} style={{ "--swatch": c }} />
			))}
		</div>
	);
}

async function base64(buffer: ArrayBuffer): Promise<string> {
	const bytes = new Uint8Array(buffer);
	let s = "";
	for (let i = 0; i < bytes.length; i += 0x8000) {
		s += String.fromCharCode(...bytes.subarray(i, i + 0x8000));
	}
	return btoa(s);
}

function LogoField(props: {
	label: string;
	help: string;
	current: string;
	draft?: LogoDraft;
	dark?: string;
	onChange: (l: LogoDraft) => void;
}) {
	const shown = props.draft ? props.draft.preview : props.current;
	const pick = async (e: Event) => {
		const file = (e.currentTarget as HTMLInputElement).files?.[0];
		if (!file) return;
		const content = await base64(await file.arrayBuffer());
		props.onChange({ content, preview: `data:${file.type || "image/svg+xml"};base64,${content}` });
	};
	return (
		<div class="c-field">
			<span class="c-field__label">{props.label}</span>
			<div class="house__logo" style={props.dark ? { "--logo-bg": props.dark } : undefined} data-dark={props.dark ? "" : undefined}>
				{shown ? <img src={shown} alt="" /> : <span class="c-small c-muted">None</span>}
			</div>
			<span class="house__logo-actions">
				<label class="c-btn c-btn--secondary c-btn--sm">
					<Upload class="i" size={14} aria-hidden="true" />
					{shown ? "Replace" : "Upload"}
					<input class="house__file" type="file" accept=".svg,.png,.webp,image/svg+xml,image/png,image/webp" onChange={pick} />
				</label>
				{shown && (
					<button class="c-btn c-btn--ghost c-btn--sm" type="button" onClick={() => props.onChange({ content: "", preview: "" })}>
						<Trash2 class="i" size={14} aria-hidden="true" />
						Remove
					</button>
				)}
			</span>
			<span class="c-field__help">{props.help}</span>
		</div>
	);
}

const weightNames: [RegExp, number][] = [
	[/thin|hairline/i, 100],
	[/extra-?light|ultra-?light/i, 200],
	[/light/i, 300],
	[/medium/i, 500],
	[/semi-?bold|demi-?bold/i, 600],
	[/extra-?bold|ultra-?bold/i, 800],
	[/black|heavy/i, 900],
	[/bold/i, 700],
];

/** Guesses the weight and the style of a font file from its name, such as Inter-SemiBoldItalic.woff2. */
function guess(name: string): { weight: number; maxWeight?: number; style: string } {
	const style = /italic|oblique/i.test(name) ? "italic" : "normal";
	if (/variable|\[wght|-vf\b/i.test(name)) return { weight: 100, maxWeight: 900, style };
	const found = weightNames.find(([re]) => re.test(name));
	return { weight: found ? found[1] : 400, style };
}

function FontField({ label, font, onChange }: { label: string; font: FontDraft; onChange: (f: FontDraft) => void }) {
	const add = async (e: Event) => {
		const input = e.currentTarget as HTMLInputElement;
		const files = [...(input.files ?? [])];
		const added: FileDraft[] = [];
		for (const file of files) {
			const buffer = await file.arrayBuffer();
			added.push({ key: `${file.name}-${Date.now()}-${added.length}`, name: file.name, size: file.size, ...guess(file.name), content: await base64(buffer), buffer });
		}
		input.value = "";
		const family = font.family || (files[0] ? files[0].name.replace(/[-_ ]?(thin|light|regular|medium|bold|black|italic|variable|semi|extra|\[.*\]).*$/i, "").replace(/\.[a-z0-9]+$/i, "") : "");
		onChange({ ...font, family, files: [...font.files, ...added] });
	};
	const set = (i: number, part: Partial<FileDraft>) => onChange({ ...font, files: font.files.map((f, j) => (j === i ? { ...f, ...part } : f)) });
	return (
		<fieldset class="house__font">
			<legend class="c-field__label">{label}</legend>
			<div class="form__row">
				<label class="c-field">
					<span class="c-field__label">Family</span>
					<input
						class="c-input"
						placeholder="The font of the system"
						value={font.family}
						onInput={(e) => onChange({ ...font, family: e.currentTarget.value })}
					/>
				</label>
				<label class="c-field">
					<span class="c-field__label">Falls back on</span>
					<select class="c-input" value={font.fallback} onChange={(e) => onChange({ ...font, fallback: e.currentTarget.value })}>
						{Object.entries(fallbacks).map(([value, f]) => (
							<option value={value} key={value}>
								{f.label}
							</option>
						))}
					</select>
				</label>
			</div>
			{font.files.length > 0 && (
				<div class="table-wrap">
					<table class="table">
						<thead>
							<tr>
								<th>File</th>
								<th>Weight</th>
								<th>Style</th>
								<th />
							</tr>
						</thead>
						<tbody>
							{font.files.map((f, i) => (
								<tr key={f.key}>
									<td class="c-mono c-small">
										{f.name} <span class="c-muted">{Math.ceil(f.size / 1024)} KB</span>
									</td>
									<td>
										<select
											class="c-input"
											aria-label={`Weight of ${f.name}`}
											value={f.maxWeight ? `${f.weight}-${f.maxWeight}` : String(f.weight)}
											onChange={(e) => {
												const [w, max] = e.currentTarget.value.split("-").map(Number);
												set(i, { weight: w, maxWeight: max });
											}}
										>
											{[100, 200, 300, 400, 500, 600, 700, 800, 900].map((w) => (
												<option value={String(w)} key={w}>
													{w}
												</option>
											))}
											<option value="100-900">100 to 900, variable</option>
										</select>
									</td>
									<td>
										<select class="c-input" aria-label={`Style of ${f.name}`} value={f.style} onChange={(e) => set(i, { style: e.currentTarget.value })}>
											<option value="normal">Normal</option>
											<option value="italic">Italic</option>
										</select>
									</td>
									<td class="table__actions">
										<button
											class="c-btn c-btn--ghost c-btn--sm"
											type="button"
											aria-label={`Remove ${f.name}`}
											onClick={() => onChange({ ...font, files: font.files.filter((_, j) => j !== i) })}
										>
											<Trash2 class="i" size={14} aria-hidden="true" />
										</button>
									</td>
								</tr>
							))}
						</tbody>
					</table>
				</div>
			)}
			<span class="house__logo-actions">
				<label class="c-btn c-btn--secondary c-btn--sm">
					<Upload class="i" size={14} aria-hidden="true" />
					Add font files
					<input class="house__file" type="file" multiple accept=".woff2,.woff,.ttf,.otf,font/woff2,font/woff,font/ttf,font/otf" onChange={add} />
				</label>
				<span class="c-small c-muted">WOFF2, WOFF, TrueType or OpenType, 1 MB each. Without files, the family must be on the readers' systems.</span>
			</span>
		</fieldset>
	);
}

/** The font-family of a font of the draft, as the preview loads it. */
function stackOf(f: FontDraft, preview: string): string {
	const stack = fallbacks[f.fallback]?.stack ?? fallbacks["sans-serif"].stack;
	if (f.files.length > 0) return `"${preview}", ${stack}`;
	return f.family.trim() ? `"${f.family.trim()}", ${stack}` : stack;
}

/** Loads the files of a font of the draft under the name preview. */
function usePreviewFont(f: FontDraft, preview: string) {
	useEffect(() => {
		const faces = f.files.map((file) => {
			const source = file.buffer ?? `url(${file.url})`;
			const weight = file.maxWeight ? `${file.weight} ${file.maxWeight}` : String(file.weight);
			return new FontFace(preview, source, { weight, style: file.style });
		});
		for (const face of faces) {
			document.fonts.add(face);
			face.load().catch(() => undefined);
		}
		return () => {
			for (const face of faces) document.fonts.delete(face);
		};
	}, [f.files, preview]);
}

function Preview({ draft, style, version }: { draft: Draft; style: HouseStyle; version: string }) {
	usePreviewFont(draft.heading, "hs-preview-heading");
	usePreviewFont(draft.body, "hs-preview-body");
	const c = draft.colors;
	const valid = colorNames.every((n) => parse(c[n.key]));
	// Headings without a family of their own take the body font.
	const ownHeading = draft.heading.family.trim() !== "" || draft.heading.files.length > 0;
	const vars = useMemo(() => {
		if (!valid) return {};
		return {
			"--p-text": c.text,
			"--p-background": c.background,
			"--p-primary": c.primary,
			"--p-on-primary": on(c.primary, c.text, c.background),
			"--p-secondary": c.secondary,
			"--p-muted": muted(c.text, c.background),
			"--p-line": mix(c.text, c.background, 0.16),
			"--p-surface": mix(c.text, c.background, 0.04),
			"--p-heading": ownHeading ? stackOf(draft.heading, "hs-preview-heading") : stackOf(draft.body, "hs-preview-body"),
			"--p-body": stackOf(draft.body, "hs-preview-body"),
		};
	}, [draft, valid]);
	const logo = draft.logo ? draft.logo.preview : style.has_logo ? `/-/house-style/logo${version}` : "";
	const inverse = draft.logoInverse ? draft.logoInverse.preview : style.has_logo_inverse ? `/-/house-style/logo-inverse${version}` : "";
	const name = draft.name.trim() || location.hostname;
	const chart = chartOf(draft, style);
	return (
		<div class="house__preview" aria-label="Preview" style={vars}>
			<span class="c-label">Preview, before saving</span>
			<div class="preview">
				<div class="preview__top">{logo ? <img src={logo} alt={name} /> : <strong>{name}</strong>}</div>
				<span class="preview__kicker">Report</span>
				<h3 class="preview__title">Quarterly review</h3>
				<p class="preview__text">
					Revenue grew by 12% over the quarter, led by the North. <a href="#preview">See the figures</a>.
				</p>
				<div class="preview__kpis">
					<span class="preview__kpi">
						<span>Revenue</span>
						<strong>2,980</strong>
					</span>
					<span class="preview__kpi">
						<span>Orders</span>
						<strong>1,412</strong>
					</span>
				</div>
				<div class="preview__bars">
					{chart.slice(0, 4).map((color, i) => (
						<span key={`${i}-${color}`} style={{ "--bar": color, "--w": `${90 - i * 18}%` }} />
					))}
				</div>
				<span class="preview__button">Open the report</span>
			</div>
			<div class="preview preview--slide">
				<span class="preview__kicker">Part 1</span>
				<h3 class="preview__title">Results</h3>
				<div class="preview__foot">{inverse ? <img src={inverse} alt={name} /> : <strong>{name}</strong>}</div>
			</div>
		</div>
	);
}
