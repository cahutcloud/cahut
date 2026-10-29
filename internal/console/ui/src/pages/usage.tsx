// What the Cahut holds, on the overview: a count for each thing that its
// limits bound, with a gauge where it has a limit. A member of an individual
// Cahut sees what they hold and their own limits, never what the others
// hold. See docs/adr/0030-kind-and-limits-of-each-cahut.md.
import type { ComponentChildren } from "preact";
import type { Settings } from "../api.gen";
import { bytes } from "../components";
import { locale, words } from "../i18n";
import { Link } from "../router";

const t = words({
	en: {
		cahutHolds: "What this Cahut holds",
		youHold: "What you hold in this Cahut",
		kindAndLimits: "Kind and limits",
		people: "People",
		peopleHelp: "Active or waiting for their activation.",
		admins: "Administrators",
		adminsHelp: "Among its people.",
		teamRooms: "Team Rooms",
		teamRoomsHelp: "Personal Rooms do not count.",
		rooms: "Rooms",
		roomsHelp: "Beside the personal Rooms.",
		yourRoomsHelp: "Beside your personal Room.",
		windows: "Windows",
		windowsHelp: "Out of the trash.",
		yourWindowsHelp: "Out of the trash, in your Rooms.",
		perRoom: (n: string) => ` At most ${n} in each Room.`,
		disk: "Disk",
		diskHelp: "Every version of every Window, the trash included.",
		yourDiskHelp: "Every version of your Windows, the trash included.",
		of: (limit: string) => `of ${limit}`,
		ofLimit: (now: string, limit: string) => `${now} of ${limit}`,
		noLimit: "No limit",
		at: "At the limit",
		over: "Over the limit",
	},
	fr: {
		cahutHolds: "Ce que contient cette Cahut",
		youHold: "Ce que vous avez dans cette Cahut",
		kindAndLimits: "Type et limites",
		people: "Personnes",
		peopleHelp: "Actives ou en attente de leur activation.",
		admins: "Administrateurs",
		adminsHelp: "Parmi ses personnes.",
		teamRooms: "Rooms d'équipe",
		teamRoomsHelp: "Les Rooms personnelles ne comptent pas.",
		rooms: "Rooms",
		roomsHelp: "En plus des Rooms personnelles.",
		yourRoomsHelp: "En plus de votre Room personnelle.",
		windows: "Windows",
		windowsHelp: "Hors de la corbeille.",
		yourWindowsHelp: "Hors de la corbeille, dans vos Rooms.",
		perRoom: (n) => ` Au plus ${n} dans chaque Room.`,
		disk: "Disque",
		diskHelp: "Toutes les versions de toutes les Windows, corbeille comprise.",
		yourDiskHelp: "Toutes les versions de vos Windows, corbeille comprise.",
		of: (limit) => `sur ${limit}`,
		ofLimit: (now, limit) => `${now} sur ${limit}`,
		noLimit: "Aucune limite",
		at: "À la limite",
		over: "Au-delà de la limite",
	},
});

/** Where a gauge turns to say that its limit is near. */
const near = 0.9;

/** A thing that the limits count: what is held, and its limit if it has one. */
type Count = { key: string; name: string; help: string; now: number; limit?: number; disk?: boolean };

/**
 * What the Cahut holds, as its reader may read it: the whole Cahut for the
 * people of a collaborative Cahut and the administrators of an individual
 * one, and what the person holds in an individual Cahut, for its members
 * and for its administrators when a limit per person binds them.
 */
export function Usage({ settings: s, admin }: { settings: Settings; admin: boolean }) {
	const l = s.limits;
	const individual = s.kind === "individual";
	const cahut: Count[] = [];
	if (s.usage) {
		const u = s.usage;
		if (admin) {
			cahut.push({ key: "people", name: t.people, help: t.peopleHelp, now: u.people, limit: l.people });
			if (l.admins !== undefined) cahut.push({ key: "admins", name: t.admins, help: t.adminsHelp, now: u.admins, limit: l.admins });
		}
		cahut.push(
			individual
				? { key: "rooms", name: t.rooms, help: t.roomsHelp, now: u.rooms, limit: l.rooms }
				: { key: "rooms", name: t.teamRooms, help: t.teamRoomsHelp, now: u.rooms, limit: l.rooms },
			{
				key: "windows",
				name: t.windows,
				help: t.windowsHelp + (l.windows_per_room === undefined ? "" : t.perRoom(l.windows_per_room.toLocaleString(locale))),
				now: u.windows,
				limit: l.windows,
			},
			{ key: "disk", name: t.disk, help: t.diskHelp, now: u.disk, limit: l.disk, disk: true },
		);
	}
	const yours: Count[] = [];
	const perPerson = l.rooms_per_person !== undefined || l.windows_per_person !== undefined || l.disk_per_person !== undefined;
	if (s.yours && (!s.usage || perPerson)) {
		const y = s.yours;
		yours.push(
			{ key: "rooms", name: t.rooms, help: t.yourRoomsHelp, now: y.rooms, limit: l.rooms_per_person },
			{ key: "windows", name: t.windows, help: t.yourWindowsHelp, now: y.windows, limit: l.windows_per_person },
			{ key: "disk", name: t.disk, help: t.yourDiskHelp, now: y.disk, limit: l.disk_per_person, disk: true },
		);
	}
	// The administrators of the Cahut read its limits in full on their own
	// page.
	const more = admin && (
		<Link class="c-link c-small" href="/settings">
			{t.kindAndLimits}
		</Link>
	);
	return (
		<>
			{cahut.length > 0 && <Counts title={t.cahutHolds} counts={cahut} more={more} />}
			{yours.length > 0 && <Counts title={t.youHold} counts={yours} more={cahut.length === 0 && more} />}
		</>
	);
}

function Counts(props: { title: string; counts: Count[]; more: ComponentChildren }) {
	return (
		<section class="section">
			<div class="usage__head">
				<h2 class="c-h4">{props.title}</h2>
				{props.more}
			</div>
			<div class="usage">
				{props.counts.map((c) => (
					<Tile key={c.key} count={c} />
				))}
			</div>
		</section>
	);
}

/** One thing that the limits count: how much is held, and the gauge of its limit if it has one. */
function Tile({ count: c }: { count: Count }) {
	const show = (n: number) => (c.disk ? bytes(n) : n.toLocaleString(locale));
	const limit = c.limit;
	const state = limit === undefined || c.now < limit ? "" : c.now > limit ? t.over : t.at;
	return (
		<article class="c-card usage__tile">
			<span class="c-label">{c.name}</span>
			<p class="usage__count">
				<span class="usage__now">{show(c.now)}</span>
				{limit !== undefined && <span class="c-small c-muted"> {t.of(show(limit))}</span>}
			</p>
			{limit === undefined ? (
				<span class="c-small c-muted">{t.noLimit}</span>
			) : (
				<Gauge now={c.now} limit={limit} label={c.name} text={t.ofLimit(show(c.now), show(limit))} />
			)}
			{state && <span class="c-badge c-badge--warning usage__state">{state}</span>}
			<span class="c-field__help">{c.help}</span>
		</article>
	);
}

/** A bar that fills as what is held nears its limit. */
function Gauge(props: { now: number; limit: number; label: string; text: string }) {
	const share = props.limit === 0 ? 1 : Math.min(props.now / props.limit, 1);
	const level = props.now >= props.limit ? "full" : share >= near ? "near" : undefined;
	return (
		<div
			class="gauge"
			data-level={level}
			role="meter"
			aria-label={props.label}
			aria-valuemin={0}
			aria-valuemax={props.limit}
			aria-valuenow={Math.min(props.now, props.limit)}
			aria-valuetext={props.text}
		>
			<span class="gauge__fill" style={{ width: `${share * 100}%` }} />
		</div>
	);
}
