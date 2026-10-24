// The kind and the limits of the Cahut: the administrators of the server set
// them, its administrators read them, and its people read what concerns
// them. See docs/adr/0030-kind-and-limits-of-each-cahut.md.
import { useState } from "preact/hooks";
import { explain } from "../api";
import type { Limits, Settings as CahutSettings } from "../api.gen";
import { bytes, Failure, Notice, PageHead } from "../components";
import { useConsole } from "../context";
import { useLoad } from "../load";

/** The limits of a Cahut, as the screens name them, with what each counts. */
const limits: { key: keyof Limits; name: string; help: string; disk?: boolean; perPerson?: boolean }[] = [
	{ key: "people", name: "People", help: "Active or waiting for their activation, administrators included. Disabled people do not count." },
	{ key: "admins", name: "Administrators", help: "Administrators of the Cahut, counted the same way." },
	{ key: "rooms", name: "Rooms", help: "Team Rooms, or in an individual Cahut the Rooms beside the personal ones." },
	{ key: "windows", name: "Windows", help: "Windows out of the trash, in the whole Cahut." },
	{ key: "windows_per_room", name: "Windows per Room", help: "Windows out of the trash in one Room, personal Rooms included." },
	{ key: "disk", name: "Disk", help: "Every version of every Window, the trash included, and the files of the house style.", disk: true },
	{ key: "rooms_per_person", name: "Rooms per person", help: "Rooms of each person beside their personal Room.", perPerson: true },
	{ key: "windows_per_person", name: "Windows per person", help: "Windows of each person out of the trash, over their Rooms.", perPerson: true },
	{ key: "disk_per_person", name: "Disk per person", help: "Every version of the Windows of each person, the trash included.", disk: true, perPerson: true },
];

/** What the Cahut, or the person in an individual Cahut, holds against the limit key. */
function held(s: CahutSettings, key: keyof Limits): number | undefined {
	switch (key) {
		case "people":
		case "admins":
		case "rooms":
		case "windows":
		case "disk":
			return s.usage?.[key];
		case "rooms_per_person":
			return s.yours?.rooms;
		case "windows_per_person":
			return s.yours?.windows;
		case "disk_per_person":
			return s.yours?.disk;
	}
	return undefined;
}

const kinds: Record<string, { name: string; about: string }> = {
	collaborative: { name: "Collaborative", about: "People create team Rooms, add one another to them, and open Windows to the whole Cahut." },
	individual: {
		name: "Individual",
		about: "People share the Cahut, not their work: each Room keeps one member, and nobody sees the others.",
	},
};

export function Settings() {
	const api = useConsole();
	const { data, error, reload } = useLoad(() => api.cahut("get_cahut", {}));
	if (error) return <Failure message={error} />;
	if (!data) return null;
	const server = api.session.kind === "admin";
	return (
		<>
			<PageHead label="This Cahut" title="Kind and limits">
				{server
					? "As an administrator of the server, you choose how people work in this Cahut, and bound what it holds."
					: "The administrators of the server choose how people work in this Cahut, and bound what it holds."}
			</PageHead>
			<section class="section c-card panel">
				<h2 class="c-h4">{(kinds[data.kind] ?? kinds.collaborative).name} Cahut</h2>
				<p class="c-small">{(kinds[data.kind] ?? kinds.collaborative).about}</p>
			</section>
			<section class="section">
				<h2 class="c-h4">Limits</h2>
				<LimitsTable settings={data} />
			</section>
			{server && <SettingsForm settings={data} onSaved={() => (reload(), api.reloadKind(), api.reloadRooms())} />}
		</>
	);
}

/** The limits of a Cahut with what it holds, as its reader may read them. */
export function LimitsTable({ settings: s }: { settings: CahutSettings }) {
	const rows = limits.filter((l) => (l.perPerson ? s.kind === "individual" : s.usage !== undefined || s.limits[l.key] !== undefined));
	if (rows.length === 0) return <Notice>This Cahut has no limit.</Notice>;
	const show = (n: number | undefined, disk?: boolean) => (n === undefined ? "" : disk ? bytes(n) : String(n));
	return (
		<div class="table-wrap">
			<table class="table">
				<thead>
					<tr>
						<th>What</th>
						<th>{s.kind === "individual" && s.usage === undefined ? "Yours" : "Holds"}</th>
						<th>Limit</th>
						<th />
					</tr>
				</thead>
				<tbody>
					{rows.map((l) => {
						const limit = s.limits[l.key];
						const now = held(s, l.key);
						const full = limit !== undefined && now !== undefined && now >= limit;
						return (
							<tr key={l.key}>
								<td>
									<span class="limit">
										<span class="c-small">{l.name}</span>
										<span class="c-field__help">{l.help}</span>
									</span>
								</td>
								<td class="c-small limit__value">{show(now, l.disk)}</td>
								<td class="c-small limit__value">{limit === undefined ? <span class="c-muted">No limit</span> : show(limit, l.disk)}</td>
								<td>{full && <span class="c-badge c-badge--warning">{now > limit ? "Over the limit" : "At the limit"}</span>}</td>
							</tr>
						);
					})}
				</tbody>
			</table>
		</div>
	);
}

/** How a limit is typed in the form: a count, or gigabytes for the disk. */
function typed(n: number | undefined, disk?: boolean): string {
	if (n === undefined) return "";
	return disk ? String(n / 1e9) : String(n);
}

/** The form where an administrator of the server sets the kind and the limits. */
function SettingsForm({ settings, onSaved }: { settings: CahutSettings; onSaved: () => void }) {
	const api = useConsole();
	const [kind, setKind] = useState(settings.kind);
	const [values, setValues] = useState<Record<string, string>>(() =>
		Object.fromEntries(limits.map((l) => [l.key, typed(settings.limits[l.key], l.disk)])),
	);
	const [error, setError] = useState("");
	const [saved, setSaved] = useState(false);
	const [busy, setBusy] = useState(false);
	const submit = async (e: Event) => {
		e.preventDefault();
		setBusy(true);
		setError("");
		setSaved(false);
		const set: Limits = {};
		for (const l of limits) {
			const v = values[l.key].trim();
			if (v === "" || (l.perPerson && kind !== "individual")) continue;
			set[l.key] = l.disk ? Math.round(Number(v) * 1e9) : Number(v);
		}
		try {
			await api.cahut("set_cahut_settings", { kind, limits: set });
			setSaved(true);
			onSaved();
		} catch (err) {
			setError(explain(err));
		}
		setBusy(false);
	};
	return (
		<section class="section c-card panel">
			<h2 class="c-h4">Change them</h2>
			<form class="form" onSubmit={submit}>
				<fieldset class="consent__group">
					<legend class="c-field__label">Kind</legend>
					{Object.entries(kinds).map(([value, k]) => (
						<label class="c-check" key={value}>
							<input type="radio" name="kind" checked={kind === value} onChange={() => setKind(value)} />
							<span>
								{k.name}: {k.about}
							</span>
						</label>
					))}
				</fieldset>
				{kind === "individual" && settings.kind !== "individual" && (
					<p class="c-small c-muted">
						A Cahut becomes individual once no Room has more than one member and no Window opens to the whole Cahut. The names of new
						Rooms then start with the handle of their owner.
					</p>
				)}
				<div class="form__row">
					{limits
						.filter((l) => !l.perPerson || kind === "individual")
						.map((l) => (
							<label class="c-field" key={l.key}>
								<span class="c-field__label">
									{l.name}
									{l.disk && " (GB)"}
								</span>
								<input
									class="c-input"
									type="number"
									inputMode={l.disk ? "decimal" : "numeric"}
									min={l.key === "people" || l.key === "admins" ? 1 : 0}
									step={l.disk ? "any" : 1}
									placeholder="No limit"
									value={values[l.key]}
									onInput={(e) => setValues({ ...values, [l.key]: e.currentTarget.value })}
								/>
							</label>
						))}
				</div>
				<p class="c-small c-muted">
					Leave a limit empty to lift it. Lowering one deletes nothing: the Cahut keeps what it holds and refuses what would add to it. The
					disk counts in gigabytes of a billion bytes.
				</p>
				{error && <Failure message={error} />}
				{saved && (
					<p class="c-small" role="status">
						Saved. The change is in the Logbook of this Cahut.
					</p>
				)}
				<div>
					<button class="c-btn c-btn--primary" type="submit" disabled={busy}>
						Save
					</button>
				</div>
			</form>
		</section>
	);
}
