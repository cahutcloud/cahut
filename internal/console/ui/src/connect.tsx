// How to connect an AI to this Cahut, client by client: the overview and
// AIs and access show it. See docs/adr/0025-first-start-in-the-browser.md.
import { Download } from "lucide-preact";
import { useState } from "preact/hooks";
import { CopyField, ipAddress, Notice } from "./components";
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
			"In claude.ai or the Claude app, open Customize, then Connectors. Select Add, then Add custom connector. Give it a name and this address, then continue: Claude finds by itself that the Cahut signs in with OAuth. Add it, then select Connect: Claude sends you here to sign in, and you choose the Rooms it reaches. Every plan of Claude adds custom connectors; the Free plan holds one.",
		address: "the address",
		uploadSkill: "Then upload the skill for Claude under Customize, beside the connector.",
		downloadSkill: "Download the skill",
		chatgptIP: (
			<>
				ChatGPT does not connect to a Cahut at an IP address. Ask the administrators of the server to move it to a domain, then add
				it to ChatGPT at its new address.
			</>
		),
		chatgpt:
			"In ChatGPT on the web, open Plugins, select Add, then Add custom MCP server. Give it a name and this address as its server URL: OAuth is its authentication already, and ChatGPT finds the rest by itself. Tick the box that says you understand, then select Create as a plugin. ChatGPT sends you here to sign in, and you choose the Rooms it reaches. It asks you to confirm the calls that change something.",
		chatgptPlans:
			"ChatGPT publishes to a Cahut on its Business, Enterprise and Edu plans only. With Free, Plus or Pro, it cannot: connect Claude or Mistral Vibe, which work on their free plans too.",
		vibe: "In Mistral Vibe, formerly Le Chat, open Context, then Connectors. Select New connector, then New custom connector. Give it a name without spaces, such as cahut, and this address as its server: Vibe finds by itself that the Cahut signs in with OAuth. Select Add connector: Vibe sends you here to sign in, and you choose the Rooms it reaches.",
		vibePlans:
			"Every plan of Mistral Vibe adds custom connectors, Free included. Only an administrator of the account adds one: on the Free and Pro plans, its owner.",
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
			"Dans claude.ai ou l'appli Claude, ouvrez Personnaliser, puis Connecteurs. Choisissez Ajouter, puis Ajouter un connecteur personnalisé. Donnez-lui un nom et cette adresse, puis continuez : Claude voit seul que la Cahut se connecte par OAuth. Ajoutez-le, puis choisissez Connecter : Claude vous envoie ici pour vous connecter, et vous choisissez les Rooms qu'il atteint. Toutes les offres de Claude ajoutent des connecteurs personnalisés ; l'offre Free en accepte un.",
		address: "l'adresse",
		uploadSkill: "Puis importez la skill pour Claude dans Personnaliser, à côté du connecteur.",
		downloadSkill: "Télécharger la skill",
		chatgptIP: (
			<>
				ChatGPT ne se connecte pas à une Cahut à une adresse IP. Demandez aux administrateurs du serveur de la déplacer vers un
				domaine, puis ajoutez-la à ChatGPT à sa nouvelle adresse.
			</>
		),
		chatgpt:
			"Dans ChatGPT sur le web, ouvrez Plugins, choisissez Ajouter, puis Ajouter un serveur MCP personnalisé. Donnez-lui un nom et cette adresse comme URL du serveur : son authentification est déjà OAuth, et ChatGPT trouve le reste seul. Cochez la case qui dit que vous comprenez, puis choisissez Créer en tant que plugin. ChatGPT vous envoie ici pour vous connecter, et vous choisissez les Rooms qu'il atteint. Il vous demande de confirmer les appels qui changent quelque chose.",
		chatgptPlans:
			"ChatGPT ne publie dans une Cahut qu'avec ses offres Business, Enterprise et Edu. Avec Free, Plus ou Pro, il ne le peut pas : connectez Claude ou Mistral Vibe, qui marchent aussi avec leur offre gratuite.",
		vibe: "Dans Mistral Vibe, l'ancien Le Chat, ouvrez Contexte, puis Connecteurs. Choisissez Nouveau connecteur, puis Nouveau connecteur personnalisé. Donnez-lui un nom sans espace, comme cahut, et cette adresse comme serveur : Vibe voit seul que la Cahut se connecte par OAuth. Choisissez Ajouter un connecteur : Vibe vous envoie ici pour vous connecter, et vous choisissez les Rooms qu'il atteint.",
		vibePlans:
			"Toutes les offres de Mistral Vibe ajoutent des connecteurs personnalisés, Free comprise. Seul un administrateur du compte en ajoute un : avec les offres Free et Pro, son titulaire.",
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
	},
});

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

