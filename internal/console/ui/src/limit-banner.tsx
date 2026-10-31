// The banner of the overview when a limit is close or reached: what is
// close, and what to do about it. It warns from the same share of a limit as
// the e-mails that the Cahut sends its administrators, nearLimit, and names
// only what get_cahut gives the person: the limits of the whole Cahut to its
// administrators and to the members of a collaborative Cahut, and in an
// individual Cahut each person's own limits, never what the others hold.
// See docs/adr/0030-kind-and-limits-of-each-cahut.md.
import { Gauge } from "lucide-preact";
import { type Limits, nearLimit, type Settings } from "./api.gen";
import { bytes, plural } from "./components";
import { useConsole } from "./context";
import { de, locale, words } from "./i18n";

const count = (n: number) => n.toLocaleString(locale);

const t = words({
	en: {
		near: (what: string) => `This Cahut is close to its limit of ${what}`,
		reached: (what: string) => `This Cahut reached its limit of ${what}`,
		nearYours: (what: string) => `You are close to your limit of ${what}`,
		reachedYours: (what: string) => `You reached your limit of ${what}`,
		people: "people",
		admins: "administrators",
		rooms: "Rooms",
		windows: "Windows",
		disk: "disk",
		hasPeople: (n: number, limit: number) => `It has ${plural(n, "person", "people")}, for a limit of ${count(limit)}.`,
		hasAdmins: (n: number, limit: number) =>
			`It has ${plural(n, "administrator", "administrators")}, for a limit of ${count(limit)}.`,
		hasTeamRooms: (n: number, limit: number) => `It has ${plural(n, "team Room", "team Rooms")}, for a limit of ${count(limit)}.`,
		hasRooms: (n: number, limit: number) =>
			`It has ${plural(n, "Room", "Rooms")} beside the personal ones, for a limit of ${count(limit)}.`,
		holdsWindows: (n: number, limit: number) =>
			`It holds ${plural(n, "Window", "Windows")} out of the trash, for a limit of ${count(limit)}.`,
		keeps: (n: number, limit: number) => `It keeps ${bytes(n)}, the trash included, for a limit of ${bytes(limit)}.`,
		youHaveRooms: (n: number, limit: number) =>
			`You have ${plural(n, "Room", "Rooms")} beside your personal Room, for a limit of ${count(limit)}.`,
		youHoldWindows: (n: number, limit: number) =>
			`You hold ${plural(n, "Window", "Windows")} out of the trash, for a limit of ${count(limit)}.`,
		youKeep: (n: number, limit: number) => `Your Windows keep ${bytes(n)}, the trash included, for a limit of ${bytes(limit)}.`,
		refuses: "The Cahut now refuses what would add to it.",
		disablePeople: "Disable a person who no longer needs an account, since disabled people do not count",
		demoteAdmin: "Make one of the administrators a member",
		deleteRoom: "Delete a Room that is no longer used",
		trashWindows: "Put the Windows that are no longer needed in the trash",
		purgeWindows: "Delete Windows in the trash for good, since they still count",
		askServer: ", or ask the administrators of the server for a higher limit.",
		askCahut: ", or ask an administrator of the Cahut about a higher limit.",
	},
	fr: {
		near: (what) => `Cette Cahut approche de sa limite ${de(what)}`,
		reached: (what) => `Cette Cahut a atteint sa limite ${de(what)}`,
		nearYours: (what) => `Vous approchez de votre limite ${de(what)}`,
		reachedYours: (what) => `Vous avez atteint votre limite ${de(what)}`,
		people: "personnes",
		admins: "administrateurs",
		rooms: "Rooms",
		windows: "Windows",
		disk: "disque",
		hasPeople: (n, limit) => `Elle compte ${plural(n, "personne", "personnes")}, pour une limite de ${count(limit)}.`,
		hasAdmins: (n, limit) => `Elle compte ${plural(n, "administrateur", "administrateurs")}, pour une limite de ${count(limit)}.`,
		hasTeamRooms: (n, limit) => `Elle compte ${plural(n, "Room d'équipe", "Rooms d'équipe")}, pour une limite de ${count(limit)}.`,
		hasRooms: (n, limit) =>
			`Elle compte ${plural(n, "Room", "Rooms")} en plus des Rooms personnelles, pour une limite de ${count(limit)}.`,
		holdsWindows: (n, limit) =>
			`Elle contient ${plural(n, "Window", "Windows")} hors de la corbeille, pour une limite de ${count(limit)}.`,
		keeps: (n, limit) => `Elle occupe ${bytes(n)}, corbeille comprise, pour une limite de ${bytes(limit)}.`,
		youHaveRooms: (n, limit) =>
			`Vous avez ${plural(n, "Room", "Rooms")} en plus de votre Room personnelle, pour une limite de ${count(limit)}.`,
		youHoldWindows: (n, limit) =>
			`Vous avez ${plural(n, "Window", "Windows")} hors de la corbeille, pour une limite de ${count(limit)}.`,
		youKeep: (n, limit) => `Vos Windows occupent ${bytes(n)}, corbeille comprise, pour une limite de ${bytes(limit)}.`,
		refuses: "La Cahut refuse désormais ce qui s'y ajouterait.",
		disablePeople: "Désactivez une personne qui n'a plus besoin de son compte, puisque les personnes désactivées ne comptent pas",
		demoteAdmin: "Faites d'un des administrateurs un membre",
		deleteRoom: "Supprimez une Room qui ne sert plus",
		trashWindows: "Mettez à la corbeille les Windows qui ne servent plus",
		purgeWindows: "Supprimez définitivement des Windows de la corbeille, puisqu'elles comptent encore",
		askServer: ", ou demandez une limite plus haute aux administrateurs du serveur.",
		askCahut: ", ou demandez une limite plus haute à un administrateur de la Cahut.",
	},
});

/** A limit that the banner watches: what it counts, and what makes room under it. */
interface Watched {
	key: keyof Limits;
	what: string;
	held: (s: Settings) => number | undefined;
	says: (n: number, limit: number) => string;
	remedy: string;
	/** It bounds the person, in an individual Cahut, rather than the Cahut. */
	yours?: boolean;
	/** Only the administrators add to it: the people and the administrators of the Cahut. */
	admins?: boolean;
}

/**
 * The limits that the banner watches, those of the e-mails: a Room holds
 * Windows up to its own limit, which the console counts nowhere.
 */
const watched: Watched[] = [
	{ key: "people", what: t.people, held: (s) => s.usage?.people, says: t.hasPeople, remedy: t.disablePeople, admins: true },
	{ key: "admins", what: t.admins, held: (s) => s.usage?.admins, says: t.hasAdmins, remedy: t.demoteAdmin, admins: true },
	{ key: "rooms", what: t.rooms, held: (s) => s.usage?.rooms, says: t.hasTeamRooms, remedy: t.deleteRoom },
	{ key: "windows", what: t.windows, held: (s) => s.usage?.windows, says: t.holdsWindows, remedy: t.trashWindows },
	{ key: "disk", what: t.disk, held: (s) => s.usage?.disk, says: t.keeps, remedy: t.purgeWindows },
	{ key: "rooms_per_person", what: t.rooms, held: (s) => s.yours?.rooms, says: t.youHaveRooms, remedy: t.deleteRoom, yours: true },
	{ key: "windows_per_person", what: t.windows, held: (s) => s.yours?.windows, says: t.youHoldWindows, remedy: t.trashWindows, yours: true },
	{ key: "disk_per_person", what: t.disk, held: (s) => s.yours?.disk, says: t.youKeep, remedy: t.purgeWindows, yours: true },
];

/**
 * Whether n is at nearLimit percent of limit, or beyond, as the server
 * reckons it: a limit of zero allows none, and is not one to come near.
 */
function near(n: number, limit: number): boolean {
	return limit > 0 && n * 100 >= limit * nearLimit;
}

/** Warns of each limit that is close or reached, with what to do; nothing when none is. */
export function LimitBanner({ settings: s }: { settings: Settings }) {
	const api = useConsole();
	const close = watched.flatMap((w) => {
		const limit = s.limits[w.key];
		const n = w.held(s);
		if (limit === undefined || n === undefined || (w.admins && !api.isAdmin) || !near(n, limit)) return [];
		return [{ ...w, n, limit, reached: n >= limit }];
	});
	if (close.length === 0) return null;
	// Who raises a limit, as the refusals say: the administrators of the
	// server, whom an administrator of the Cahut asks, and whom a member asks
	// through an administrator of the Cahut.
	const ask = api.isAdmin ? t.askServer : t.askCahut;
	return (
		<section class="section">
			{close.map((c) => {
				const title = c.yours ? (c.reached ? t.reachedYours : t.nearYours) : c.reached ? t.reached : t.near;
				// In an individual Cahut, the limit of Rooms counts those beside the personal ones.
				const says = c.key === "rooms" && s.kind === "individual" ? t.hasRooms : c.says;
				return (
					<div key={c.key} class="c-alert c-alert--warning" role="status">
						<Gauge class="i" size={18} aria-hidden="true" />
						<div>
							<strong>{title(c.what)}</strong>
							{says(c.n, c.limit)} {c.reached && `${t.refuses} `}
							{c.remedy}
							{ask}
						</div>
					</div>
				);
			})}
		</section>
	);
}
