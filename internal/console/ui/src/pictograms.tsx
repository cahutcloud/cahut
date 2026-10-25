// The pictograms of the brand that the console draws, as they are: those of
// the navigation, and the Window of the vocabulary for a Room without one.
// The console reads their files when it builds and draws them inline, so
// that their stroke follows the text, light or dark, and their accent stays.
// See § 5 of docs/charte-graphique/README.md.
import account from "@charte/icones/navigation/account.svg?raw";
import aiAccess from "@charte/icones/navigation/ai-access.svg?raw";
import cahut from "@charte/icones/navigation/cahut.svg?raw";
import email from "@charte/icones/navigation/email.svg?raw";
import houseStyle from "@charte/icones/navigation/house-style.svg?raw";
import knockFirst from "@charte/icones/navigation/knock-first.svg?raw";
import limits from "@charte/icones/navigation/limits.svg?raw";
import logbook from "@charte/icones/navigation/logbook.svg?raw";
import people from "@charte/icones/navigation/people.svg?raw";
import room from "@charte/icones/navigation/room.svg?raw";
import roomNew from "@charte/icones/navigation/room-new.svg?raw";
import roomPersonal from "@charte/icones/navigation/room-personal.svg?raw";
import roomTeam from "@charte/icones/navigation/room-team.svg?raw";
import serverAdmins from "@charte/icones/navigation/server-admins.svg?raw";
import updates from "@charte/icones/navigation/updates.svg?raw";
import windowOfVocabulary from "@charte/icones/window.svg?raw";

const files = {
	account,
	aiAccess,
	cahut,
	email,
	houseStyle,
	knockFirst,
	limits,
	logbook,
	people,
	room,
	roomNew,
	roomPersonal,
	roomTeam,
	serverAdmins,
	updates,
	window: windowOfVocabulary,
};

export type PictogramName = keyof typeof files;

// Each file names itself, with a title and a label. Beside the words of a
// link, a pictogram is decoration, and its title would show on hover.
const drawn = Object.fromEntries(
	Object.entries(files).map(([name, svg]) => [name, svg.replace(/<title>[^<]*<\/title>/, "").replace(/ role="img" aria-label="[^"]*"/, "")]),
) as Record<PictogramName, string>;

/** A pictogram, 20 px wide beside the words it goes with, or 64 px on an empty screen. */
export function Pictogram({ name, large }: { name: PictogramName; large?: boolean }) {
	return <span class={large ? "pictogram pictogram--large" : "pictogram"} aria-hidden="true" dangerouslySetInnerHTML={{ __html: drawn[name] }} />;
}
