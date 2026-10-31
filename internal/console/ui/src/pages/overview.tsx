// The first screen once signed in: for a person, their latest Windows, their
// Rooms and what the Cahut holds. An administrator of the server finds the
// Cahuts of the server in the console of the server. See
// docs/adr/0058-a-console-for-the-server.md.
import { Domain, Failure, Notice, PageHead, WindowCards } from "../components";
import { ConnectSteps } from "../connect";
import { useConsole } from "../context";
import { que, words } from "../i18n";
import { LimitBanner } from "../limit-banner";
import { useLoad } from "../load";
import { Link } from "../router";
import { RoomCards } from "./rooms";
import { Usage } from "./usage";

const t = words({
	en: {
		overview: "Overview",
		signedInAs: (name: string, admin: boolean) =>
			`Signed in as ${name}, ${admin ? "an administrator of this Cahut" : "a member of this Cahut"}.`,
		connectFirst: "Connect your first AI",
		connectText: "Your AI publishes on Cahut over MCP. Connect it, then ask it to make a document, a deck or a dashboard.",
		latestWindows: "Latest Windows",
		noWindow: () => (
			<>
				No Window yet. Ask your AI to publish one: <Link class="c-link" href="/access">connect an AI</Link> first.
			</>
		),
		yourRooms: "Your Rooms",
	},
	fr: {
		overview: "Vue d'ensemble",
		signedInAs: (name, admin) =>
			`Connecté en tant ${que(name)}, ${admin ? "administrateur" : "membre"} de cette Cahut.`,
		connectFirst: "Connectez votre première IA",
		connectText:
			"Votre IA publie sur Cahut par MCP. Connectez-la, puis demandez-lui un document, une présentation ou un tableau de bord.",
		latestWindows: "Dernières Windows",
		noWindow: () => (
			<>
				Pas encore de Window. Demandez à votre IA d'en publier une : <Link class="c-link" href="/access">connectez une IA</Link>{" "}
				d'abord.
			</>
		),
		yourRooms: "Vos Rooms",
	},
});

export function Overview() {
	const api = useConsole();
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
	// Until the person connects an AI, the overview shows how.
	const noAI = !connections.some((c) => c.status !== "ended") && !tokens.some((t) => !t.revoked_at);
	return (
		<>
			<PageHead label={t.overview} title={<Domain name={location.hostname} />}>
				{t.signedInAs(account.name, account.role === "admin")}
			</PageHead>
			<LimitBanner settings={about} />
			{noAI && (
				<section class="section c-card panel">
					<h2 class="c-h4">{t.connectFirst}</h2>
					<p class="c-small">{t.connectText}</p>
					<ConnectSteps mcp={about.mcp ?? `${location.origin}/mcp`} />
				</section>
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
