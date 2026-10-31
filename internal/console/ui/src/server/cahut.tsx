// The page of a Cahut in the console of the server: what its administrators
// set for it, its moderation and its service tokens, from whichever domain
// they signed in at. See docs/adr/0058-a-console-for-the-server.md.
import { ArrowUpRight } from "lucide-preact";
import { useMemo, useState } from "preact/hooks";
import { explain } from "../api";
import { CopyField, Domain, Failure, originOf, PageHead } from "../components";
import { words } from "../i18n";
import { useLoad } from "../load";
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
	},
});

export function CahutPage({ domain }: { domain: string }) {
	const api = useServer();
	const call = useMemo(() => api.cahutOf(domain), [api, domain]);
	const { data, error, reload } = useLoad(() => call("get_cahut", {}), [domain]);
	const [decisions, setDecisions] = useState(0);
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
