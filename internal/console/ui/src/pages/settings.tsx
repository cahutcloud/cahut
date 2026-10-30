// The kind and the limits of the Cahut, and the links to its legal pages:
// the administrators of the server set them, its administrators read them,
// and its people read what concerns them. The administrators of the server
// suspend the Cahut here too, and resume it. See
// docs/adr/0030-kind-and-limits-of-each-cahut.md,
// docs/adr/0042-the-legal-pages-of-a-cahut.md and
// docs/adr/0048-suspending-a-cahut.md.
import { Fragment } from "preact";
import { useState } from "preact/hooks";
import { explain } from "../api";
import type { LegalLinks, LegalPages, Limits, SetLegalLinksInput, Settings as CahutSettings, Suspension } from "../api.gen";
import { bytes, Failure, legalPages, Notice, PageHead } from "../components";
import { useConsole } from "../context";
import { locale, words } from "../i18n";
import { useLoad } from "../load";

const t = words({
	en: {
		people: "People",
		peopleHelp: "Active or waiting for their activation, administrators included. Disabled people do not count.",
		admins: "Administrators",
		adminsHelp: "Administrators of the Cahut, counted the same way.",
		rooms: "Rooms",
		roomsHelp: "Team Rooms, or in an individual Cahut the Rooms beside the personal ones.",
		windows: "Windows",
		windowsHelp: "Windows out of the trash, in the whole Cahut.",
		windowsPerRoom: "Windows per Room",
		windowsPerRoomHelp: "Windows out of the trash in one Room, personal Rooms included.",
		disk: "Disk",
		diskHelp: "Every version of every Window, the trash included.",
		roomsPerPerson: "Rooms per person",
		roomsPerPersonHelp: "Rooms of each person beside their personal Room.",
		windowsPerPerson: "Windows per person",
		windowsPerPersonHelp: "Windows of each person out of the trash, over their Rooms.",
		diskPerPerson: "Disk per person",
		diskPerPersonHelp: "Every version of the Windows of each person, the trash included.",
		collaborative: "Collaborative",
		collaborativeAbout: "People create team Rooms, add one another to them, and open Windows to the whole Cahut.",
		individual: "Individual",
		individualAbout: "People share the Cahut, not their work: each Room keeps one member, and nobody sees the others.",
		label: "This Cahut",
		title: "Kind and limits",
		leadServer: "As an administrator of the server, you choose how people work in this Cahut, bound what it holds, and link its legal pages.",
		leadCahut: "The administrators of the server choose how people work in this Cahut, bound what it holds, and link its legal pages.",
		kindTitle: (kind: string) => `${kind} Cahut`,
		limits: "Limits",
		noLimit: "This Cahut has no limit.",
		what: "What",
		yours: "Yours",
		holds: "Holds",
		limit: "Limit",
		unlimited: "No limit",
		over: "Over the limit",
		at: "At the limit",
		change: "Change them",
		kind: "Kind",
		kindChoice: (name: string, about: string) => `${name}: ${about}`,
		becomesIndividual:
			"A Cahut becomes individual once no Room has more than one member and no Window opens to the whole Cahut. The names of new Rooms then start with the handle of their owner.",
		gb: " (GB)",
		lift: "Leave a limit empty to lift it. Lowering one deletes nothing: the Cahut keeps what it holds and refuses what would add to it. The disk counts in gigabytes of a billion bytes.",
		saved: "Saved. The change is in the Logbook of this Cahut.",
		save: "Save",
		legal: "Legal pages",
		legalAbout:
			"The legal notice, terms, privacy policy and support that the host of this Cahut publishes elsewhere. The console and the pages that readers see link them at their foot, in the language of each page.",
		noLegal: "This Cahut links no legal page.",
		page: "Page",
		address: "Address",
		inFrench: "In French",
		sameAddress: "Same address",
		changeLegal: "Change the links",
		inFrenchOf: (name: string) => `${name}, in French`,
		https: "An address that starts with https://",
		legalHelp:
			"Give an address in French only where the page in French differs: the console links it when it speaks French. Leave a field empty to remove its link.",
		moderation: "Address of moderation",
		moderationAbout:
			"The address that receives each report about the Windows of this Cahut, beside its administrators who have an address, such as an abuse mailbox of its host. It receives them when this Cahut sends e-mail.",
		noModeration: "This Cahut has no address of moderation.",
		moderationField: "Address",
		moderationHelp: "Leave it empty to remove it.",
		suspension: "Suspension",
		suspendTitle: "Suspend this Cahut",
		suspendHelp:
			"After a decision of moderation, or when its organization stops paying. It keeps everything, and serves nothing but to the administrators of the server until they resume it: its Windows say that it is suspended, its people cannot sign in, and their AIs cannot act in it.",
		suspend: "Suspend",
		reason: "Reason",
		reasonHelp: (n: number) =>
			`${n} of 500 characters. Its administrators read it by e-mail, when the Cahut sends e-mail, and in its Logbook once it is resumed. Readers, people and AIs never do.`,
		suspendNow: "Suspend the Cahut",
		cancel: "Cancel",
		suspendedSince: (since: string) => `Suspended since ${since}`,
		reasonIs: "Reason:",
		resumeHelp: "Resuming it opens its Windows again, and its people and their AIs find everything as it was.",
		resume: "Resume",
		changeReason: "Change the reason",
		saveReason: "Save the reason",
	},
	fr: {
		people: "Personnes",
		peopleHelp: "Actives ou en attente de leur activation, administrateurs compris. Les personnes désactivées ne comptent pas.",
		admins: "Administrateurs",
		adminsHelp: "Administrateurs de la Cahut, comptés de la même façon.",
		rooms: "Rooms",
		roomsHelp: "Les Rooms d'équipe, ou dans une Cahut individuelle les Rooms en plus des Rooms personnelles.",
		windows: "Windows",
		windowsHelp: "Les Windows hors de la corbeille, dans toute la Cahut.",
		windowsPerRoom: "Windows par Room",
		windowsPerRoomHelp: "Les Windows hors de la corbeille dans une Room, Rooms personnelles comprises.",
		disk: "Disque",
		diskHelp: "Toutes les versions de toutes les Windows, corbeille comprise.",
		roomsPerPerson: "Rooms par personne",
		roomsPerPersonHelp: "Les Rooms de chaque personne en plus de sa Room personnelle.",
		windowsPerPerson: "Windows par personne",
		windowsPerPersonHelp: "Les Windows de chaque personne hors de la corbeille, dans toutes ses Rooms.",
		diskPerPerson: "Disque par personne",
		diskPerPersonHelp: "Toutes les versions des Windows de chaque personne, corbeille comprise.",
		collaborative: "Collaborative",
		collaborativeAbout:
			"Les personnes créent des Rooms d'équipe, s'y ajoutent les unes les autres, et ouvrent des Windows à toute la Cahut.",
		individual: "Individuelle",
		individualAbout:
			"Les personnes partagent la Cahut, pas leur travail : chaque Room garde un seul membre, et personne ne voit les autres.",
		label: "Cette Cahut",
		title: "Type et limites",
		leadServer:
			"En tant qu'administrateur du serveur, vous choisissez comment les personnes travaillent dans cette Cahut, bornez ce qu'elle contient, et liez ses pages légales.",
		leadCahut:
			"Les administrateurs du serveur choisissent comment les personnes travaillent dans cette Cahut, bornent ce qu'elle contient, et lient ses pages légales.",
		kindTitle: (kind) => `Cahut ${kind.toLowerCase()}`,
		limits: "Limites",
		noLimit: "Cette Cahut n'a aucune limite.",
		what: "Ressource",
		yours: "Utilisé par vous",
		holds: "Utilisé",
		limit: "Limite",
		unlimited: "Aucune limite",
		over: "Au-delà de la limite",
		at: "À la limite",
		change: "Les modifier",
		kind: "Type",
		kindChoice: (name, about) => `${name} : ${about}`,
		becomesIndividual:
			"Une Cahut devient individuelle dès qu'aucune Room n'a plus d'un membre et qu'aucune Window ne s'ouvre à toute la Cahut. Les noms des nouvelles Rooms commencent alors par l'identifiant de leur propriétaire.",
		gb: " (Go)",
		lift: "Laissez une limite vide pour la lever. En baisser une ne supprime rien : la Cahut garde ce qu'elle contient et refuse ce qui s'y ajouterait. Le disque se compte en gigaoctets d'un milliard d'octets.",
		saved: "Enregistré. La modification figure au Logbook de cette Cahut.",
		save: "Enregistrer",
		legal: "Pages légales",
		legalAbout:
			"Les mentions légales, les conditions d'utilisation, la politique de confidentialité et l'assistance que l'hébergeur de cette Cahut publie ailleurs. La console et les pages que voient les lecteurs les lient en bas de page, dans la langue de chaque page.",
		noLegal: "Cette Cahut ne lie aucune page légale.",
		page: "Page",
		address: "Adresse",
		inFrench: "En français",
		sameAddress: "Même adresse",
		changeLegal: "Modifier les liens",
		inFrenchOf: (name) => `${name}, en français`,
		https: "Une adresse qui commence par https://",
		legalHelp:
			"Ne donnez une adresse en français que si la page en français diffère : la console la lie quand elle parle français. Laissez un champ vide pour retirer son lien.",
		moderation: "Adresse de modération",
		moderationAbout:
			"L'adresse qui reçoit chaque signalement des Windows de cette Cahut, à côté de ses administrateurs qui ont une adresse, comme une boîte d'abus de son hébergeur. Elle les reçoit quand cette Cahut envoie des e-mails.",
		noModeration: "Cette Cahut n'a pas d'adresse de modération.",
		moderationField: "Adresse",
		moderationHelp: "Laissez-la vide pour la retirer.",
		suspension: "Suspension",
		suspendTitle: "Suspendre cette Cahut",
		suspendHelp:
			"Après une décision de modération, ou quand son organisation ne paie plus. Elle garde tout, et ne sert plus rien qu'aux administrateurs du serveur jusqu'à ce qu'ils la réactivent : ses Windows disent qu'elle est suspendue, ses personnes ne peuvent plus se connecter, et leurs IA ne peuvent plus y agir.",
		suspend: "Suspendre",
		reason: "Motif",
		reasonHelp: (n) =>
			`${n} caractères sur 500. Ses administrateurs le lisent par e-mail, quand la Cahut envoie des e-mails, et dans son Logbook une fois réactivée. Les lecteurs, les personnes et les IA ne le lisent jamais.`,
		suspendNow: "Suspendre la Cahut",
		cancel: "Annuler",
		suspendedSince: (since) => `Suspendue depuis le ${since}`,
		reasonIs: "Motif :",
		resumeHelp: "La réactiver rouvre ses Windows, et ses personnes et leurs IA retrouvent tout comme avant.",
		resume: "Réactiver",
		changeReason: "Modifier le motif",
		saveReason: "Enregistrer le motif",
	},
});

/** The limits of a Cahut, as the screens name them, with what each counts. */
const limits: { key: keyof Limits; name: string; help: string; disk?: boolean; perPerson?: boolean }[] = [
	{ key: "people", name: t.people, help: t.peopleHelp },
	{ key: "admins", name: t.admins, help: t.adminsHelp },
	{ key: "rooms", name: t.rooms, help: t.roomsHelp },
	{ key: "windows", name: t.windows, help: t.windowsHelp },
	{ key: "windows_per_room", name: t.windowsPerRoom, help: t.windowsPerRoomHelp },
	{ key: "disk", name: t.disk, help: t.diskHelp, disk: true },
	{ key: "rooms_per_person", name: t.roomsPerPerson, help: t.roomsPerPersonHelp, perPerson: true },
	{ key: "windows_per_person", name: t.windowsPerPerson, help: t.windowsPerPersonHelp, perPerson: true },
	{ key: "disk_per_person", name: t.diskPerPerson, help: t.diskPerPersonHelp, disk: true, perPerson: true },
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
	collaborative: { name: t.collaborative, about: t.collaborativeAbout },
	individual: { name: t.individual, about: t.individualAbout },
};

export function Settings() {
	const api = useConsole();
	const { data, error, reload } = useLoad(() => api.cahut("get_cahut", {}));
	if (error) return <Failure message={error} />;
	if (!data) return null;
	// Only the administrators of the server set the kind, the limits and the
	// legal pages, with their session, alone or beside that of the person.
	const server = api.serverSession !== null;
	return (
		<>
			<PageHead label={t.label} title={t.title}>
				{server ? t.leadServer : t.leadCahut}
			</PageHead>
			<section class="section c-card panel">
				<h2 class="c-h4">{t.kindTitle((kinds[data.kind] ?? kinds.collaborative).name)}</h2>
				<p class="c-small">{(kinds[data.kind] ?? kinds.collaborative).about}</p>
			</section>
			<section class="section">
				<h2 class="c-h4">{t.limits}</h2>
				<LimitsTable settings={data} />
			</section>
			{server && <SettingsForm settings={data} onSaved={() => (reload(), api.reloadCahut(), api.reloadRooms())} />}
			<section class="section">
				<h2 class="c-h4">{t.legal}</h2>
				<p class="c-small c-muted">{t.legalAbout}</p>
				{!server && <LegalTable legal={data.legal} />}
			</section>
			{server && <LegalForm legal={data.legal} onSaved={() => (reload(), api.reloadCahut())} />}
			<ModerationAddress email={data.moderation_email ?? ""} server={server} onSaved={reload} />
			{server && <SuspensionForm suspension={data.suspension} onChanged={(on) => (reload(), api.setSuspended(on))} />}
		</>
	);
}

/**
 * The suspension of the Cahut, which only the administrators of the server
 * decide: they suspend it with a reason, change the reason, and resume it.
 */
function SuspensionForm({ suspension, onChanged }: { suspension?: Suspension; onChanged: (suspended: boolean) => void }) {
	const api = useConsole();
	const [asking, setAsking] = useState(false);
	const [reason, setReason] = useState(suspension?.reason ?? "");
	const [error, setError] = useState("");
	const [busy, setBusy] = useState(false);
	const call = async (run: () => Promise<unknown>, suspended: boolean) => {
		setBusy(true);
		setError("");
		try {
			await run();
			setAsking(false);
			onChanged(suspended);
		} catch (err) {
			setError(explain(err));
		}
		setBusy(false);
	};
	const suspend = (e: Event) => {
		e.preventDefault();
		call(() => api.cahutAsServer("suspend_cahut", { reason }), true);
	};
	const resume = () => call(() => api.cahutAsServer("resume_cahut", {}), false);
	const form = (submit: string) => (
		<form class="setting__more setting__confirm" onSubmit={suspend}>
			<label class="c-field">
				<span class="c-field__label">{t.reason}</span>
				<textarea class="c-input" required maxLength={500} rows={3} value={reason} onInput={(e) => setReason(e.currentTarget.value)} />
				<span class="c-field__help">{t.reasonHelp(reason.length)}</span>
			</label>
			<div class="actions">
				<button class="c-btn c-btn--danger c-btn--sm" type="submit" disabled={busy || reason.trim() === ""}>
					{submit}
				</button>
				<button
					class="c-btn c-btn--ghost c-btn--sm"
					type="button"
					onClick={() => (setAsking(false), setReason(suspension?.reason ?? ""), setError(""))}
				>
					{t.cancel}
				</button>
			</div>
		</form>
	);
	return (
		<section class="section c-card panel">
			<h2 class="c-h4">{t.suspension}</h2>
			<div class="settings">
				{suspension ? (
					<div class="setting">
						<div class="setting__text">
							<span class="setting__title">
								{t.suspendedSince(new Date(suspension.since).toLocaleDateString(locale, { dateStyle: "long" }))}
							</span>
							<p class="c-small">
								{t.reasonIs} {suspension.reason}
							</p>
							<p class="c-small c-muted">{t.resumeHelp}</p>
						</div>
						<div class="setting__action">
							<button class="c-btn c-btn--primary" type="button" disabled={busy} onClick={resume}>
								{t.resume}
							</button>
						</div>
						{asking ? (
							form(t.saveReason)
						) : (
							<p class="setting__more c-small">
								<button class="c-link link-button" type="button" onClick={() => setAsking(true)}>
									{t.changeReason}
								</button>
							</p>
						)}
					</div>
				) : (
					<div class="setting">
						<div class="setting__text">
							<span class="setting__title">{t.suspendTitle}</span>
							<p class="c-small c-muted">{t.suspendHelp}</p>
						</div>
						{!asking && (
							<div class="setting__action">
								<button class="c-btn c-btn--danger" type="button" onClick={() => setAsking(true)}>
									{t.suspend}
								</button>
							</div>
						)}
						{asking && form(t.suspendNow)}
					</div>
				)}
			</div>
			{error && <Failure message={error} />}
		</section>
	);
}

/** The links to the legal pages of a Cahut, as its administrators read them. */
function LegalTable({ legal }: { legal: LegalLinks }) {
	const rows = legalPages.filter(({ page }) => legal[page]);
	if (rows.length === 0) return <Notice>{t.noLegal}</Notice>;
	// A long address goes on to the next line after one of its slashes.
	const link = (url: string) => (
		<a class="c-link legal-address" href={url} target="_blank" rel="noopener">
			{url.split("/").map((part, i) => (
				<Fragment key={i}>
					{i > 0 && "/"}
					{i > 0 && <wbr />}
					{part}
				</Fragment>
			))}
		</a>
	);
	return (
		<div class="table-wrap">
			<table class="table">
				<thead>
					<tr>
						<th>{t.page}</th>
						<th>{t.address}</th>
						<th>{t.inFrench}</th>
					</tr>
				</thead>
				<tbody>
					{rows.map(({ page, name }) => {
						const fr = legal.fr?.[page];
						return (
							<tr key={page}>
								<td class="c-small">{name}</td>
								<td class="c-small">{link(legal[page] ?? "")}</td>
								<td class="c-small">{fr ? link(fr) : <span class="c-muted">{t.sameAddress}</span>}</td>
							</tr>
						);
					})}
				</tbody>
			</table>
		</div>
	);
}

/**
 * The address that receives the reports of this Cahut, which the
 * administrators of the server set and those of the Cahut read. See
 * docs/adr/0054-reporting-and-withdrawing-a-window.md.
 */
function ModerationAddress({ email, server, onSaved }: { email: string; server: boolean; onSaved: () => void }) {
	const api = useConsole();
	const [value, setValue] = useState(email);
	const [error, setError] = useState("");
	const [saved, setSaved] = useState(false);
	const [busy, setBusy] = useState(false);
	const submit = async (e: Event) => {
		e.preventDefault();
		setBusy(true);
		setError("");
		setSaved(false);
		try {
			await api.cahutAsServer("set_moderation_address", { email: value.trim() });
			setSaved(true);
			onSaved();
		} catch (err) {
			setError(explain(err));
		}
		setBusy(false);
	};
	return (
		<section class="section c-card panel">
			<h2 class="c-h4">{t.moderation}</h2>
			<p class="c-small c-muted">{t.moderationAbout}</p>
			{!server && (email ? <p class="c-mono c-small">{email}</p> : <Notice>{t.noModeration}</Notice>)}
			{server && (
				<form class="form" onSubmit={submit}>
					<label class="c-field">
						<span class="c-field__label">{t.moderationField}</span>
						<input
							class="c-input"
							type="email"
							autocomplete="off"
							spellcheck={false}
							placeholder="abuse@example.com"
							value={value}
							onInput={(e) => setValue(e.currentTarget.value)}
						/>
						<span class="c-field__help">{t.moderationHelp}</span>
					</label>
					{error && <Failure message={error} />}
					{saved && (
						<p class="c-small" role="status">
							{t.saved}
						</p>
					)}
					<div>
						<button class="c-btn c-btn--primary" type="submit" disabled={busy}>
							{t.save}
						</button>
					</div>
				</form>
			)}
		</section>
	);
}

/** The form where an administrator of the server sets the links to the legal pages. */
function LegalForm({ legal, onSaved }: { legal: LegalLinks; onSaved: () => void }) {
	const api = useConsole();
	const [values, setValues] = useState<Record<string, string>>(() =>
		Object.fromEntries(
			legalPages.flatMap(({ page }) => [
				[page, legal[page] ?? ""],
				[`fr.${page}`, legal.fr?.[page] ?? ""],
			]),
		),
	);
	const [error, setError] = useState("");
	const [saved, setSaved] = useState(false);
	const [busy, setBusy] = useState(false);
	const submit = async (e: Event) => {
		e.preventDefault();
		setBusy(true);
		setError("");
		setSaved(false);
		const set: SetLegalLinksInput = {};
		const fr: LegalPages = {};
		for (const { page } of legalPages) {
			const v = values[page].trim();
			const f = values[`fr.${page}`].trim();
			if (v !== "") set[page] = v;
			if (f !== "") fr[page] = f;
		}
		if (Object.keys(fr).length > 0) set.fr = fr;
		try {
			await api.cahutAsServer("set_legal_links", set);
			setSaved(true);
			onSaved();
		} catch (err) {
			setError(explain(err));
		}
		setBusy(false);
	};
	const field = (key: string, label: string, placeholder: string) => (
		<label class="c-field">
			<span class="c-field__label">{label}</span>
			<input
				class="c-input"
				type="url"
				inputMode="url"
				pattern="https://.+"
				title={t.https}
				autocomplete="off"
				spellcheck={false}
				placeholder={placeholder}
				value={values[key]}
				onInput={(e) => setValues({ ...values, [key]: e.currentTarget.value })}
			/>
		</label>
	);
	return (
		<section class="section c-card panel">
			<h2 class="c-h4">{t.changeLegal}</h2>
			<form class="form" onSubmit={submit}>
				{legalPages.map(({ page, name }) => (
					<div class="form__row" key={page}>
						{field(page, name, "https://")}
						{field(`fr.${page}`, t.inFrenchOf(name), t.sameAddress)}
					</div>
				))}
				<p class="c-small c-muted">{t.legalHelp}</p>
				{error && <Failure message={error} />}
				{saved && (
					<p class="c-small" role="status">
						{t.saved}
					</p>
				)}
				<div>
					<button class="c-btn c-btn--primary" type="submit" disabled={busy}>
						{t.save}
					</button>
				</div>
			</form>
		</section>
	);
}

/** The limits of a Cahut with what it holds, as its reader may read them. */
function LimitsTable({ settings: s }: { settings: CahutSettings }) {
	const rows = limits.filter((l) => (l.perPerson ? s.kind === "individual" : s.usage !== undefined || s.limits[l.key] !== undefined));
	if (rows.length === 0) return <Notice>{t.noLimit}</Notice>;
	const show = (n: number | undefined, disk?: boolean) => (n === undefined ? "" : disk ? bytes(n) : String(n));
	return (
		<div class="table-wrap">
			<table class="table">
				<thead>
					<tr>
						<th>{t.what}</th>
						<th>{s.kind === "individual" && s.usage === undefined ? t.yours : t.holds}</th>
						<th>{t.limit}</th>
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
								<td class="c-small limit__value">{limit === undefined ? <span class="c-muted">{t.unlimited}</span> : show(limit, l.disk)}</td>
								<td>{full && <span class="c-badge c-badge--warning">{now > limit ? t.over : t.at}</span>}</td>
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
			await api.cahutAsServer("set_cahut_settings", { kind, limits: set });
			setSaved(true);
			onSaved();
		} catch (err) {
			setError(explain(err));
		}
		setBusy(false);
	};
	return (
		<section class="section c-card panel">
			<h2 class="c-h4">{t.change}</h2>
			<form class="form" onSubmit={submit}>
				<fieldset class="consent__group">
					<legend class="c-field__label">{t.kind}</legend>
					{Object.entries(kinds).map(([value, k]) => (
						<label class="c-check" key={value}>
							<input type="radio" name="kind" checked={kind === value} onChange={() => setKind(value)} />
							<span>{t.kindChoice(k.name, k.about)}</span>
						</label>
					))}
				</fieldset>
				{kind === "individual" && settings.kind !== "individual" && (
					<p class="c-small c-muted">{t.becomesIndividual}</p>
				)}
				<div class="form__row">
					{limits
						.filter((l) => !l.perPerson || kind === "individual")
						.map((l) => (
							<label class="c-field" key={l.key}>
								<span class="c-field__label">
									{l.name}
									{l.disk && t.gb}
								</span>
								<input
									class="c-input"
									type="number"
									inputMode={l.disk ? "decimal" : "numeric"}
									min={l.key === "people" || l.key === "admins" ? 1 : 0}
									step={l.disk ? "any" : 1}
									placeholder={t.unlimited}
									value={values[l.key]}
									onInput={(e) => setValues({ ...values, [l.key]: e.currentTarget.value })}
								/>
							</label>
						))}
				</div>
				<p class="c-small c-muted">{t.lift}</p>
				{error && <Failure message={error} />}
				{saved && (
					<p class="c-small" role="status">
						{t.saved}
					</p>
				)}
				<div>
					<button class="c-btn c-btn--primary" type="submit" disabled={busy}>
						{t.save}
					</button>
				</div>
			</form>
		</section>
	);
}
