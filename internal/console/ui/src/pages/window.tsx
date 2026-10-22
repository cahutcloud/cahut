// The page of a Window: where it opens, who opens it and until when, its
// two versions, what its AI declared it does, and the trash. See
// docs/adr/0019-window-versions-access-and-declaration.md.
import { ArrowUpRight, RotateCcw, Trash2, Undo2 } from "lucide-preact";
import { useState } from "preact/hooks";
import { explain } from "../api";
import type { AccessSet, Declaration, WindowDetail } from "../api.gen";
import { AccessBadge, accesses, ago, bytes, ConfirmButton, CopyField, Failure, kindName, Notice, PageHead, plural } from "../components";
import { useConsole } from "../context";
import { useLoad } from "../load";
import { Link } from "../router";

export function WindowPage({ room, name }: { room: string; name: string }) {
	const api = useConsole();
	const { data, error, reload } = useLoad(() => api.cahut("get_window", { window: `${room}/${name}` }), [room, name]);
	const [failure, setFailure] = useState("");
	// The Key that the last change drew or set, shown this once.
	const [shared, setShared] = useState<AccessSet | null>(null);
	if (error) return <Failure message={error} />;
	if (!data) return null;
	const w = data;
	const act = async (run: () => Promise<unknown>) => {
		setFailure("");
		try {
			await run();
			api.reloadRooms();
			reload();
		} catch (err) {
			setFailure(explain(err));
		}
	};
	return (
		<>
			<PageHead
				label={kindName(w.kind)}
				title={w.title}
				actions={
					!w.trashed_at && (
						<a class="c-btn c-btn--secondary" href={w.url} target="_blank" rel="noopener noreferrer">
							<ArrowUpRight class="i" size={16} aria-hidden="true" />
							Open
						</a>
					)
				}
			>
				In <Link class="c-link" href={`/rooms/${w.room}`}>{w.room}</Link> · version {w.version} · updated {ago(w.updated_at).toLowerCase()} ·{" "}
				<span class="c-mono">{w.url}</span>
			</PageHead>
			{failure && <Failure message={failure} />}
			{w.trashed_at ? (
				<section class="section">
					<Notice>This Window is in the trash: it does not open, and it is in no list. Restore it to open it again, as it was.</Notice>
					<div>
						<button class="c-btn c-btn--primary" type="button" onClick={() => act(() => api.cahut("restore_window", { window: `${w.room}/${w.name}` }))}>
							<Undo2 class="i" size={16} aria-hidden="true" />
							Restore
						</button>
					</div>
				</section>
			) : (
				<div class="layout">
					<section>
						<AccessPanel
							window={w}
							shared={shared}
							onSet={async (input) => {
								setShared(null);
								const out = await api.cahut("set_window_access", input);
								if (out.key) setShared(out);
								reload();
							}}
						/>
						<Versions window={w} onRollback={() => act(() => api.cahut("rollback_window", { window: `${w.room}/${w.name}` }))} />
					</section>
					<aside class="aside">
						<DeclarationPanel declaration={w.declaration} />
						<div class="c-card panel aside">
							<h2 class="c-h4">Trash</h2>
							<p class="c-small c-muted">A Window in the trash no longer opens, and keeps its versions and its access until you restore it.</p>
							<div>
								<ConfirmButton
									label="Put in the trash"
									question="It will no longer open."
									confirm="Put in the trash"
									danger
									onConfirm={() => act(() => api.cahut("trash_window", { window: `${w.room}/${w.name}` }))}
								/>
							</div>
						</div>
					</aside>
				</div>
			)}
		</>
	);
}

type AccessInput = { window: string; access: string; ends_at?: string; key?: string; new_key?: boolean };

/** The access of a Window, and the form that changes it. */
function AccessPanel(props: { window: WindowDetail; shared: AccessSet | null; onSet: (input: AccessInput) => Promise<void> }) {
	const w = props.window;
	const [access, setAccess] = useState(w.access);
	const [ends, setEnds] = useState(w.access_ends_at ? w.access_ends_at.slice(0, 10) : "");
	const [keyChoice, setKeyChoice] = useState<"keep" | "draw" | "choose">(w.access === "key" ? "keep" : "draw");
	const [key, setKey] = useState("");
	const [asking, setAsking] = useState(false);
	const [error, setError] = useState("");
	const [busy, setBusy] = useState(false);

	const save = async () => {
		setError("");
		setBusy(true);
		const input: AccessInput = { window: `${w.room}/${w.name}`, access };
		if (access !== "room" && ends) {
			// The access ends at the end of the chosen day, where the person is.
			input.ends_at = new Date(`${ends}T23:59:59`).toISOString();
		}
		if (access === "key" && keyChoice === "draw" && w.access === "key") input.new_key = true;
		if (access === "key" && keyChoice === "choose") input.key = key;
		try {
			await props.onSet(input);
			setAsking(false);
			setKey("");
			setKeyChoice("keep");
		} catch (err) {
			setError(explain(err));
		} finally {
			setBusy(false);
		}
	};
	const submit = (e: Event) => {
		e.preventDefault();
		// Opening a Window to the public asks first: once read, it cannot be taken back.
		if (access === "public" && w.access !== "public") {
			setAsking(true);
			return;
		}
		save();
	};
	const today = new Date();
	const tomorrow = new Date(today.getFullYear(), today.getMonth(), today.getDate() + 1).toLocaleDateString("en-CA");

	return (
		<div class="c-card panel aside">
			<div class="access-head">
				<h2 class="c-h4">Access</h2>
				<AccessBadge access={w.access} />
			</div>
			<p class="c-small c-muted">
				{capitalize((accesses[w.access] ?? accesses.room).who)} can open it
				{w.access_ends_at ? `, until ${new Date(w.access_ends_at).toLocaleDateString("en-US", { month: "short", day: "numeric", year: "numeric" })}. Then only its Room can.` : "."}
			</p>
			{props.shared?.key && (
				<div class="c-alert c-alert--success share" role="status">
					<div>
						<strong>Share the address and the Key</strong>
						Copy the Key now: Cahut keeps only a trace of it, and shows it this once.
						<CopyField label="the address" value={props.shared.url} />
						<CopyField label="the Key" value={props.shared.key} />
					</div>
				</div>
			)}
			<form class="form" onSubmit={submit}>
				<fieldset class="consent__group">
					<legend class="c-field__label">Who opens it</legend>
					{Object.entries(accesses).map(([value, a]) => (
						<label class="c-check" key={value}>
							<input type="radio" name="access" checked={access === value} onChange={() => setAccess(value)} />
							<span>
								{a.name}: {a.who}
							</span>
						</label>
					))}
				</fieldset>
				{access === "key" && (
					<fieldset class="consent__group">
						<legend class="c-field__label">Key</legend>
						{w.access === "key" && (
							<label class="c-check">
								<input type="radio" name="key" checked={keyChoice === "keep"} onChange={() => setKeyChoice("keep")} />
								<span>Keep its Key</span>
							</label>
						)}
						<label class="c-check">
							<input type="radio" name="key" checked={keyChoice === "draw"} onChange={() => setKeyChoice("draw")} />
							<span>{w.access === "key" ? "A new Key, drawn by Cahut: the old one stops working" : "A Key drawn by Cahut"}</span>
						</label>
						<label class="c-check">
							<input type="radio" name="key" checked={keyChoice === "choose"} onChange={() => setKeyChoice("choose")} />
							<span>A Key I choose</span>
						</label>
						{keyChoice === "choose" && (
							<label class="c-field consent__rooms">
								<span class="c-field__label">The Key</span>
								<input class="c-input c-mono" required minLength={10} autocomplete="off" value={key} onInput={(e) => setKey(e.currentTarget.value)} />
								<span class="c-field__help">10 characters at least. Readers type it to open the Window.</span>
							</label>
						)}
					</fieldset>
				)}
				{access !== "room" && (
					<label class="c-field">
						<span class="c-field__label">Until</span>
						<input class="c-input" type="date" min={tomorrow} value={ends} onInput={(e) => setEnds(e.currentTarget.value)} />
						<span class="c-field__help">Leave empty for no end. After this day, only its Room opens the Window.</span>
					</label>
				)}
				{error && <Failure message={error} />}
				{asking ? (
					<span class="confirm" role="group">
						<span class="c-small">Anyone with its address will open it, and may keep a copy. Make it public?</span>
						<button class="c-btn c-btn--primary c-btn--sm" type="button" disabled={busy} onClick={save}>
							Make it public
						</button>
						<button class="c-btn c-btn--ghost c-btn--sm" type="button" onClick={() => setAsking(false)}>
							Cancel
						</button>
					</span>
				) : (
					<div>
						<button class="c-btn c-btn--primary" type="submit" disabled={busy}>
							Save the access
						</button>
					</div>
				)}
			</form>
		</div>
	);
}

function capitalize(s: string): string {
	return s.charAt(0).toUpperCase() + s.slice(1);
}

/** The current and the previous version of a Window. */
function Versions({ window: w, onRollback }: { window: WindowDetail; onRollback: () => void }) {
	return (
		<div class="c-card panel aside">
			<h2 class="c-h4">Versions</h2>
			<p class="c-small c-muted">Cahut keeps the current version and the one before. Going back can be undone: the version you leave becomes the previous one.</p>
			<ul class="versions">
				{w.versions.map((v) => (
					<li key={v.number} class="versions__row">
						<span>
							<b>Version {v.number}</b> {v.current && <span class="c-badge c-badge--success">Current</span>}
							<span class="c-small c-muted versions__meta">
								{ago(v.published_at)} · for {v.published_by} · {plural(v.file_count, "file", "files")}, {bytes(v.size)}
							</span>
						</span>
						{!v.current && (
							<button class="c-btn c-btn--secondary c-btn--sm" type="button" onClick={onRollback}>
								<RotateCcw class="i" size={14} aria-hidden="true" />
								Go back to it
							</button>
						)}
					</li>
				))}
			</ul>
		</div>
	);
}

/** What the AI declared that the pages of the current version do. */
function DeclarationPanel({ declaration: d }: { declaration?: Declaration }) {
	return (
		<div class="c-card panel aside">
			<h2 class="c-h4">What it does</h2>
			{!d ? (
				<p class="c-small c-muted">
					Not declared: this version was published before declarations, and may load and call any address. Its next version will declare
					what it does.
				</p>
			) : (
				<>
					<p class="c-small c-muted">As its AI declared it. The pages can load and call nothing else.</p>
					<dl class="declaration">
						<dt>Scripts</dt>
						<dd>{d.scripts ? "Yes" : "No"}</dd>
						<dt>Loads and calls</dt>
						<dd>{d.origins?.length ? d.origins.map((o) => <div class="c-mono" key={o}>{o}</div>) : "Nothing outside this Cahut"}</dd>
						<dt>Forms send to</dt>
						<dd>{d.forms?.length ? d.forms.map((o) => <div class="c-mono" key={o}>{o}</div>) : "Nowhere"}</dd>
					</dl>
				</>
			)}
		</div>
	);
}

/** The Windows of a Room in the trash, when it has some. */
export function Trash({ room }: { room: string }) {
	const api = useConsole();
	const { data } = useLoad(() => api.cahut("list_windows", { room, trashed: true }).catch(() => null), [room]);
	if (!data || data.windows.length === 0) return null;
	return (
		<section class="section">
			<h2 class="c-h4">
				<Trash2 class="i" size={16} aria-hidden="true" /> Trash
			</h2>
			<ul class="versions">
				{data.windows.map((w) => (
					<li key={w.name} class="versions__row">
						<Link class="c-link" href={`/rooms/${w.room}/${w.name}`}>
							{w.title}
						</Link>
						<span class="c-small c-muted">{w.trashed_at ? `In the trash since ${ago(w.trashed_at).toLowerCase()}` : ""}</span>
					</li>
				))}
			</ul>
		</section>
	);
}
