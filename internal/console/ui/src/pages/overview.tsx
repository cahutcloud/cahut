// The first screen once signed in: for a person, their latest Windows, their
// Rooms and what the Cahut holds, under the card of their first steps until
// their AI has published. An administrator of the server finds the Cahuts
// of the server in the console of the server. See
// docs/adr/0058-a-console-for-the-server.md and
// docs/adr/0061-a-guide-to-the-first-ai.md.
import { useState } from "preact/hooks";
import { Domain, Failure, Notice, PageHead, WindowCards } from "../components";
import { useConsole } from "../context";
import { que, words } from "../i18n";
import { LimitBanner } from "../limit-banner";
import { useLoad } from "../load";
import { Link } from "../router";
import { RoomCards } from "./rooms";
import { FirstStepsCard, hidden, hide, progressOf } from "./start";
import { Usage } from "./usage";

const t = words({
	en: {
		overview: "Overview",
		signedInAs: (name: string, admin: boolean) =>
			`Signed in as ${name}, ${admin ? "an administrator of this Cahut" : "a member of this Cahut"}.`,
		latestWindows: "Latest Windows",
		noWindow: () => (
			<>
				No Window yet. Ask your AI to publish one: <Link class="c-link" href="/start">connect an AI</Link> first.
			</>
		),
		yourRooms: "Your Rooms",
	},
	fr: {
		overview: "Vue d'ensemble",
		signedInAs: (name, admin) =>
			`Connecté en tant ${que(name)}, ${admin ? "administrateur" : "membre"} de cette Cahut.`,
		latestWindows: "Dernières Windows",
		noWindow: () => (
			<>
				Pas encore de Window. Demandez à votre IA d'en publier une : <Link class="c-link" href="/start">connectez une IA</Link>{" "}
				d'abord.
			</>
		),
		yourRooms: "Vos Rooms",
	},
});

export function Overview() {
	const api = useConsole();
	const [dismissed, setDismissed] = useState(() => hidden(api.session.handle));
	const { data, error, reload } = useLoad(() =>
		Promise.all([
			api.cahut("get_account", {}),
			api.cahut("list_windows", {}),
			api.cahut("list_rooms", {}),
			api.cahut("list_connected_ais", { user: api.session.handle }),
			api.cahut("list_tokens", { user: api.session.handle }),
			api.cahut("get_cahut", {}),
		]),
	);
	if (error) return <Failure message={error} />;
	if (!data) return null;
	const [account, { windows }, { rooms }, { connections }, { tokens }, about] = data;
	// Until an AI has published for the person, the overview leads to the
	// first steps, unless they dismissed them in this browser. A person
	// whose AI has published once is past them, even once it has expired.
	const progress = progressOf(api.session.handle, connections, tokens, windows);
	const guide = !dismissed && progress.mine.length === 0;
	return (
		<>
			<PageHead label={t.overview} title={<Domain name={location.hostname} />}>
				{t.signedInAs(account.name, account.role === "admin")}
			</PageHead>
			<LimitBanner settings={about} />
			{guide && (
				<FirstStepsCard
					progress={progress}
					onHide={() => {
						hide(api.session.handle);
						setDismissed(true);
					}}
				/>
			)}
			<section class="section">
				<h2 class="c-h4">{t.latestWindows}</h2>
				{windows.length === 0 ? (
					<Notice>{t.noWindow()}</Notice>
				) : (
					<WindowCards windows={windows.slice(0, 6)} showRoom onChanged={reload} />
				)}
			</section>
			<section class="section">
				<h2 class="c-h4">{t.yourRooms}</h2>
				<RoomCards rooms={rooms} />
			</section>
			<Usage settings={about} admin={account.role === "admin"} />
		</>
	);
}
