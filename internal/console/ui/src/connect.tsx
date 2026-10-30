// How to connect an AI to this Cahut, client by client, and whether one has
// connected: the first start, the overview and AIs and access show it. See
// docs/adr/0025-first-start-in-the-browser.md.
import { CircleCheck, Download, LoaderCircle } from "lucide-preact";
import { useEffect, useState } from "preact/hooks";
import type { Connection, TokenInfo } from "./api.gen";
import { CopyField, Notice } from "./components";
import { words } from "./i18n";
import { Link } from "./router";

const t = words({
	en: {
		claudeApps: "Claude apps",
		others: "Other clients",
		clients: "AI clients",
		plugin: (host: string) => (
			<>
				The Cahut plugin connects this Cahut and teaches Claude to make and publish Windows. It asks for the domain of this Cahut,{" "}
				<code class="c-mono">{host}</code>.
			</>
		),
		addMarketplace: "the command that adds the marketplace",
		installPlugin: "the command that installs the plugin",
		orTerminal: "Or connect the Cahut only, from a terminal, then sign in here when Claude Code sends you:",
		addCahut: "the command that adds the Cahut",
		login: "the command that signs Claude Code in",
		claudeConnector:
			"In claude.ai or the Claude app, open Settings, then Connectors, and add a custom connector with this address. Connect it: Claude sends you here to sign in, and you choose the Rooms it reaches. Every plan of Claude adds custom connectors; the Free plan holds one.",
		address: "the address",
		uploadSkill: "Then upload the skill for Claude in the settings of Claude, beside the connector.",
		downloadSkill: "Download the skill",
		chatgptIP: (
			<>
				ChatGPT does not connect to a Cahut at an IP address. Point a domain to this server and move the Cahut there, in the
				console of the administrator of the server or with <code class="c-mono">cahut domain rename</code>, then add it to ChatGPT
				at its new address.
			</>
		),
		chatgpt:
			"In ChatGPT on the web, open Plugins, select the plus button, then Add custom MCP server. Give it a name, this address, and OAuth as its authentication, then create it. ChatGPT sends you here to sign in, and you choose the Rooms it reaches. It asks you to confirm the calls that change something.",
		chatgptPlans:
			"ChatGPT publishes to a Cahut on its Business, Enterprise and Edu plans only. With Free, Plus or Pro, it cannot: connect Claude or Mistral Vibe, which work on their free plans too.",
		vibe: "In Mistral Vibe, formerly Le Chat, open Connectors, select Add Connector, then the Custom MCP Connector tab. Give it a name without spaces, such as cahut, and this address. Vibe finds that the Cahut signs in with OAuth and sends you here to sign in.",
		vibePlans:
			"Every plan of Mistral Vibe adds custom MCP connectors, Free included. Only an administrator of the account adds one: on the Free and Pro plans, its owner.",
		other: (
			<>
				Give any MCP client this address, over HTTP. A client that signs in with OAuth sends you here; one that does not takes a
				personal token, which{" "}
				<Link class="c-link" href="/access">
					AIs and access
				</Link>{" "}
				creates.
			</>
		),
		connected: (name: string) => `${name} is connected. Ask it to publish something on Cahut.`,
		waiting: "Waiting for your AI. This page sees it as soon as it connects.",
	},
	fr: {
		claudeApps: "Applis Claude",
		others: "Autres clients",
		clients: "Clients d'IA",
		plugin: (host) => (
			<>
				Le plugin Cahut connecte cette Cahut et apprend à Claude à faire et publier des Windows. Il demande le domaine de cette Cahut,{" "}
				<code class="c-mono">{host}</code>.
			</>
		),
		addMarketplace: "la commande qui ajoute la marketplace",
		installPlugin: "la commande qui installe le plugin",
		orTerminal: "Ou connectez seulement la Cahut, depuis un terminal, puis connectez-vous ici quand Claude Code vous y envoie :",
		addCahut: "la commande qui ajoute la Cahut",
		login: "la commande qui connecte Claude Code",
		claudeConnector:
			"Dans claude.ai ou l'appli Claude, ouvrez les paramètres, puis les connecteurs, et ajoutez un connecteur personnalisé avec cette adresse. Connectez-le : Claude vous envoie ici pour vous connecter, et vous choisissez les Rooms qu'il atteint. Toutes les offres de Claude ajoutent des connecteurs personnalisés ; l'offre Free en accepte un.",
		address: "l'adresse",
		uploadSkill: "Puis importez la skill pour Claude dans les paramètres de Claude, à côté du connecteur.",
		downloadSkill: "Télécharger la skill",
		chatgptIP: (
			<>
				ChatGPT ne se connecte pas à une Cahut à une adresse IP. Faites pointer un domaine vers ce serveur et déplacez-y la Cahut, dans
				la console de l'administrateur du serveur ou avec <code class="c-mono">cahut domain rename</code>, puis ajoutez-le à
				ChatGPT à sa nouvelle adresse.
			</>
		),
		chatgpt:
			"Dans ChatGPT sur le web, ouvrez les plugins, choisissez le bouton plus, puis l'ajout d'un serveur MCP personnalisé. Donnez-lui un nom, cette adresse, et OAuth comme authentification, puis créez-le. ChatGPT vous envoie ici pour vous connecter, et vous choisissez les Rooms qu'il atteint. Il vous demande de confirmer les appels qui changent quelque chose.",
		chatgptPlans:
			"ChatGPT ne publie dans une Cahut qu'avec ses offres Business, Enterprise et Edu. Avec Free, Plus ou Pro, il ne le peut pas : connectez Claude ou Mistral Vibe, qui marchent aussi avec leur offre gratuite.",
		vibe: "Dans Mistral Vibe, l'ancien Le Chat, ouvrez les connecteurs, choisissez l'ajout d'un connecteur, puis l'onglet du connecteur MCP personnalisé. Donnez-lui un nom sans espace, comme cahut, et cette adresse. Vibe voit que la Cahut se connecte par OAuth et vous envoie ici pour vous connecter.",
		vibePlans:
			"Toutes les offres de Mistral Vibe ajoutent des connecteurs MCP personnalisés, Free comprise. Seul un administrateur du compte en ajoute un : avec les offres Free et Pro, son titulaire.",
		other: (
			<>
				Donnez cette adresse à tout client MCP, en HTTP. Un client qui se connecte par OAuth vous envoie ici ; un autre prend un
				jeton personnel, que crée la page{" "}
				<Link class="c-link" href="/access">
					IA et accès
				</Link>
				.
			</>
		),
		connected: (name) => `${name} est connecté. Demandez-lui de publier quelque chose sur Cahut.`,
		waiting: "En attente de votre IA. Cette page la voit dès qu'elle se connecte.",
	},
});

/** An address that is an IPv4 address, not a domain. */
export const ipAddress = /^\d{1,3}(\.\d{1,3}){3}$/;

const clients = [
	{ id: "claude-code", name: "Claude Code" },
	{ id: "claude", name: t.claudeApps },
	{ id: "vibe", name: "Mistral Vibe" },
	{ id: "chatgpt", name: "ChatGPT" },
	{ id: "other", name: t.others },
] as const;

type Client = (typeof clients)[number]["id"];

/** The steps to connect each kind of AI client, in tabs. */
export function ConnectSteps({ mcp }: { mcp: string }) {
	const [client, setClient] = useState<Client>("claude-code");
	return (
		<div class="connect">
			<div class="c-tabs" role="tablist" aria-label={t.clients}>
				{clients.map((c) => (
					<button
						key={c.id}
						class="c-tab"
						type="button"
						role="tab"
						id={`connect-${c.id}`}
						aria-selected={client === c.id}
						aria-controls="connect-panel"
						onClick={() => setClient(c.id)}
					>
						{c.name}
					</button>
				))}
			</div>
			<div class="connect__panel" role="tabpanel" id="connect-panel" aria-labelledby={`connect-${client}`}>
				{client === "claude-code" && <ClaudeCode mcp={mcp} />}
				{client === "claude" && <ClaudeApps mcp={mcp} />}
				{client === "vibe" && <Vibe mcp={mcp} />}
				{client === "chatgpt" && <ChatGPT mcp={mcp} />}
				{client === "other" && <Other mcp={mcp} />}
			</div>
		</div>
	);
}

function ClaudeCode({ mcp }: { mcp: string }) {
	return (
		<>
			<p class="c-small">{t.plugin(location.host)}</p>
			<CopyField label={t.addMarketplace} value="/plugin marketplace add cahutcloud/cahut" />
			<CopyField label={t.installPlugin} value="/plugin install cahut@cahut" />
			<p class="c-small">{t.orTerminal}</p>
			<CopyField label={t.addCahut} value={`claude mcp add --transport http cahut ${mcp}`} />
			<CopyField label={t.login} value="claude mcp login cahut" />
		</>
	);
}

function ClaudeApps({ mcp }: { mcp: string }) {
	return (
		<>
			<p class="c-small">{t.claudeConnector}</p>
			<CopyField label={t.address} value={mcp} />
			<p class="c-small">{t.uploadSkill}</p>
			<div>
				<a class="c-btn c-btn--secondary c-btn--sm" href="/-/cahut-skill.zip" download>
					<Download class="i" size={14} aria-hidden="true" />
					{t.downloadSkill}
				</a>
			</div>
		</>
	);
}

function ChatGPT({ mcp }: { mcp: string }) {
	// The servers of ChatGPT never call an MCP server at an IP address, and
	// it says it finds no OAuth settings there.
	if (ipAddress.test(location.hostname)) {
		return (
			<Notice>{t.chatgptIP}</Notice>
		);
	}
	// Its plans for people, from Free to Pro, cannot publish: say so first.
	return (
		<>
			<Notice>{t.chatgptPlans}</Notice>
			<p class="c-small">{t.chatgpt}</p>
			<CopyField label={t.address} value={mcp} />
		</>
	);
}

function Vibe({ mcp }: { mcp: string }) {
	return (
		<>
			<p class="c-small">{t.vibe}</p>
			<CopyField label={t.address} value={mcp} />
			<p class="c-small">{t.vibePlans}</p>
		</>
	);
}

function Other({ mcp }: { mcp: string }) {
	return (
		<>
			<p class="c-small">{t.other}</p>
			<CopyField label={t.address} value={mcp} />
		</>
	);
}

/** How often the console asks whether an AI has connected. */
const every = 3000;

/**
 * Asks every few seconds, while the tab is in front, whether an AI acts for
 * the person: a connection over OAuth, or a personal token it has used.
 * It returns the name of the first one, or "".
 */
export function useConnectedAI(list: () => Promise<[Connection[], TokenInfo[]]>): string {
	const [name, setName] = useState("");
	useEffect(() => {
		if (name) return;
		let stop = false;
		const ask = async () => {
			if (document.visibilityState !== "visible") return;
			try {
				const [connections, tokens] = await list();
				const c = connections.find((c) => c.status === "active");
				const t = tokens.find((t) => t.last_used_at && !t.revoked_at);
				if (!stop && (c || t)) setName(c ? c.client : t ? t.name : "");
			} catch {
				// The next round asks again.
			}
		};
		ask();
		const timer = setInterval(ask, every);
		return () => {
			stop = true;
			clearInterval(timer);
		};
	}, [name]);
	return name;
}

/** Says whether an AI has connected yet. */
export function ConnectedStatus({ name }: { name: string }) {
	return name ? (
		<p class="connect__status connect__status--done" role="status">
			<CircleCheck class="i" size={18} aria-hidden="true" />
			{t.connected(name)}
		</p>
	) : (
		<p class="connect__status" role="status">
			<LoaderCircle class="i connect__spin" size={18} aria-hidden="true" />
			{t.waiting}
		</p>
	);
}

