// The page of a Cahut in the console of the server: what its administrators
// set for it, its moderation and its service tokens, from whichever domain
// they signed in at. See docs/adr/0058-a-console-for-the-server.md.
import { ArrowRight, ArrowUpRight } from "lucide-preact";
import { useMemo, useState } from "preact/hooks";
import { explain } from "../api";
import { CopyField, Domain, Failure, originOf, PageHead } from "../components";
import { ipAddress } from "../connect";
import { words } from "../i18n";
import { useLoad } from "../load";
import { useQuery } from "../router";
import { ModerationOf } from "../pages/moderation";
import { ServerSettings } from "../pages/settings";
import { useServer } from "./context";
import { ServiceTokens } from "./tokens";

const t = words({
	en: {
		label: "Cahut",
		lead: "What the administrators of the server set for this Cahut, its moderation and its service tokens. Its people and its Rooms are for its own administrators.",
		address: "Its address",
		firstStart: "First start",
		firstStartText: "This Cahut has nobody yet. Its first person chooses their name, their handle and their password at a link of first start, which works once, for 24 hours.",
		giveLink: "Give a link of first start",
		setupLink: "the link of first start",
		ready: "Your Cahut is ready",
		readyText: (origin: string) => (
			<>
				It answers at <strong>{origin}/</strong>, where each Room and each Window has its own address. In its console, you connect your first
				AI and find what it publishes. Here, you set what the administrators of the server set for each Cahut.
			</>
		),
		noDomain: (
			<>
				This Cahut has no domain yet. When you have one, point it to this server, then move the Cahut there from the list of the Cahuts, or
				with <code class="c-mono">cahut domain rename</code>. The links shared before then stop working.
			</>
		),
		openConsole: "Open its console",
	},
	fr: {
		label: "Cahut",
		lead: "Ce que les administrateurs du serveur règlent pour cette Cahut, sa modération et ses jetons de service. Ses personnes et ses Rooms relèvent de ses propres administrateurs.",
		address: "Son adresse",
		firstStart: "Premier démarrage",
		firstStartText:
			"Cette Cahut n'a encore personne. Sa première personne choisit son nom, son identifiant et son mot de passe sur un lien de premier démarrage, qui ne sert qu'une fois, pendant 24 heures.",
		giveLink: "Donner un lien de premier démarrage",
		setupLink: "le lien de premier démarrage",
		ready: "Votre Cahut est prête",
		readyText: (origin) => (
			<>
				Elle répond à <strong>{origin}/</strong>, où chaque Room et chaque Window a sa propre adresse. Dans sa console, vous connectez votre
				première IA et retrouvez ce qu'elle publie. Ici, vous réglez ce que les administrateurs du serveur règlent pour chaque Cahut.
			</>
		),
		noDomain: (
			<>
				Cette Cahut n'a pas encore de domaine. Quand vous en aurez un, pointez-le vers ce serveur, puis déplacez-y la Cahut depuis la liste
				des Cahuts, ou avec <code class="c-mono">cahut domain rename</code>. Les liens partagés d'ici là cessent alors de fonctionner.
			</>
		),
		openConsole: "Ouvrir sa console",
	},
});

export function CahutPage({ domain }: { domain: string }) {
	const api = useServer();
	const call = useMemo(() => api.cahutOf(domain), [api, domain]);
	const { data, error, reload } = useLoad(() => call("get_cahut", {}), [domain]);
	const [decisions, setDecisions] = useState(0);
	// The first start ends here, with ?ready in the address, when it made the
	// administrator of the server; the query follows each navigation. See
	// docs/adr/0058-a-console-for-the-server.md.
	const ready = useQuery().has("ready");
	if (error) return <Failure message={error} />;
	if (!data) return null;
	const origin = originOf(data.domain);
	return (
		<>
			<PageHead label={t.label} title={<Domain name={data.domain} />}>
				{t.lead}{" "}
				<a class="c-link" href={`${origin}/`}>
					{t.address}
					<ArrowUpRight class="i" size={14} aria-hidden="true" />
				</a>
			</PageHead>
			{ready && (
				<section class="section c-card panel" role="status">
					<h2 class="c-h4">{t.ready}</h2>
					<p class="c-small">{t.readyText(origin)}</p>
					{ipAddress.test(data.domain) && <div class="c-alert c-alert--info">{t.noDomain}</div>}
					<div>
						<a class="c-btn c-btn--primary" href={`${origin}/console/`}>
							{t.openConsole}
							<ArrowRight class="i" size={16} aria-hidden="true" />
						</a>
					</div>
				</section>
			)}
			{data.usage?.people === 0 && <FirstStart domain={data.domain} origin={origin} />}
			<ServerSettings settings={data} call={call} onChanged={reload} />
			<ModerationOf call={call} origin={origin} host onDecided={() => setDecisions((n) => n + 1)} version={decisions} />
			<ServiceTokens call={call} />
		</>
	);
}

/** Gives the link where the first person of a Cahut that has nobody starts it. */
function FirstStart({ domain, origin }: { domain: string; origin: string }) {
	const api = useServer();
	const [link, setLink] = useState("");
	const [error, setError] = useState("");
	const give = async () => {
		setError("");
		try {
			const l = await api.server("create_setup_link", { domain });
			setLink(`${origin}/console/setup#${l.code}`);
		} catch (err) {
			setError(explain(err));
		}
	};
	return (
		<section class="section c-card panel">
			<h2 class="c-h4">{t.firstStart}</h2>
			<p class="c-small c-muted">{t.firstStartText}</p>
			{error && <Failure message={error} />}
			{link ? (
				<CopyField label={t.setupLink} value={link} />
			) : (
				<div>
					<button class="c-btn c-btn--secondary" type="button" onClick={give}>
						{t.giveLink}
					</button>
				</div>
			)}
		</section>
	);
}
