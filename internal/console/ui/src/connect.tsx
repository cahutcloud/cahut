// How to connect an AI to this Cahut, client by client, one action per
// step: the first steps of the console show those of one client, and AIs
// and access those of every client, in tabs. See
// docs/adr/0025-first-start-in-the-browser.md and
// docs/adr/0061-a-guide-to-the-first-ai.md.
import { Download, ExternalLink, Play } from "lucide-preact";
import type { ComponentChildren } from "preact";
import { useState } from "preact/hooks";
import { CopyField, ipAddress, Notice } from "./components";
import { lang, words } from "./i18n";
import { Link } from "./router";

/** The AI clients that the console has steps for. */
export type ClientId = "claude" | "vibe" | "chatgpt" | "claude-code" | "other";

const t = words({
	en: {
		others: "Other clients",
		clients: "AI clients",
		address: "the address of this Cahut",
		openListing: "Open the listing of this Cahut in Claude",
		openInClaude: "Open in Claude",
		openListingHelp: "In the Claude app if it is installed, or else in your browser.",
		selectConnect: "Select Connect",
		claudeSendsYou: "Claude sends you here, where you choose the Rooms it reaches and whether it may publish.",
		approve: (client: string) => `Approve ${client}`,
		approveHelp: "Read and publish, in all your Rooms: the choice it proposes suits a start.",
		yourself: "Add the connector yourself, with the address",
		copyAddress: "Copy the address of this Cahut",
		openConnectors: "Open the connectors of Claude",
		openClaude: "Open claude.ai",
		claudeConnector:
			"Select Add, then Add custom connector. Name it Cahut, paste the address, then select Add. Claude finds by itself that the Cahut signs in with OAuth.",
		connectAndApprove: (client: string) => `Select Connect, then approve ${client}`,
		claudeFree: "Every plan of Claude adds custom connectors; the Free plan holds one. The connector then serves in the app and on claude.ai.",
		skill: "On a computer, the skill for Claude teaches it the formats of Windows and how to share them. Upload it under Customize, beside the connector.",
		downloadSkill: "Download the skill",
		vibeOpen: "Open the connectors of Mistral Vibe",
		vibeHow:
			"Under Context, then Connectors, select New connector, then New custom connector. Give it a name without spaces, such as cahut, and paste the address as its server.",
		vibeAdd: "Select Add connector, then approve Mistral Vibe",
		vibePlans:
			"Every plan of Mistral Vibe adds custom connectors, Free included. Only an administrator of the account adds one: on the Free and Pro plans, its owner.",
		chatgptIP: (
			<>
				ChatGPT does not connect to a Cahut at an IP address. Ask the administrators of the server to move it to a domain, then add
				it to ChatGPT at its new address.
			</>
		),
		chatgptPlans:
			"ChatGPT publishes to a Cahut on its Business, Enterprise and Edu plans only. With Free, Go, Plus or Pro, it cannot: connect Claude or Mistral Vibe, which work on their free plans too.",
		chatgptOpen: "In ChatGPT on the web, open Plugins",
		chatgptHow:
			"Select Add, then Add custom MCP server. Give it a name and paste the address as its server URL: OAuth is its authentication already. Tick the box that says you understand, then select Create as a plugin.",
		chatgptApproveHelp: "It asks you to confirm the calls that change something.",
		pluginTitle: "In Claude Code, add the Cahut plugin",
		plugin: (host: string) => (
			<>
				The plugin connects this Cahut and teaches Claude to make and publish Windows. It asks for the domain of this Cahut,{" "}
				<code class="c-mono">{host}</code>.
			</>
		),
		addMarketplace: "the command that adds the marketplace",
		installPlugin: "the command that installs the plugin",
		approveWhenSent: "Approve Claude Code when it sends you here",
		orTerminal: "Or connect the Cahut only, from a terminal:",
		addCahut: "the command that adds the Cahut",
		login: "the command that signs Claude Code in",
		otherGive: "Give it to your MCP client, over HTTP",
		other: (
			<>
				A client that signs in with OAuth sends you here. One that does not takes a personal token, which{" "}
				<Link class="c-link" href="/access">
					AIs and access
				</Link>{" "}
				creates.
			</>
		),
		film: "Watch the film · 45 s",
	},
	fr: {
		others: "Autres clients",
		clients: "Clients d'IA",
		address: "l'adresse de cette Cahut",
		openListing: "Ouvrez la fiche de cette Cahut dans Claude",
		openInClaude: "Ouvrir dans Claude",
		openListingHelp: "Dans l'appli Claude si elle est installée, sinon dans votre navigateur.",
		selectConnect: "Choisissez Connecter",
		claudeSendsYou: "Claude vous envoie ici, où vous choisissez les Rooms qu'il atteint et s'il peut publier.",
		approve: (client) => `Autorisez ${client}`,
		approveHelp: "Lecture et publication, dans toutes vos Rooms : le choix qu'il propose convient pour commencer.",
		yourself: "Ajouter le connecteur vous-même, avec l'adresse",
		copyAddress: "Copiez l'adresse de cette Cahut",
		openConnectors: "Ouvrez les connecteurs de Claude",
		openClaude: "Ouvrir claude.ai",
		claudeConnector:
			"Choisissez Ajouter, puis Ajouter un connecteur personnalisé. Nommez-le Cahut, collez l'adresse, puis choisissez Ajouter. Claude voit seul que la Cahut se connecte par OAuth.",
		connectAndApprove: (client) => `Choisissez Connecter, puis autorisez ${client}`,
		claudeFree:
			"Toutes les offres de Claude ajoutent des connecteurs personnalisés ; l'offre Free en accepte un. Le connecteur sert ensuite dans l'appli et sur claude.ai.",
		skill: "Sur un ordinateur, la skill pour Claude lui apprend les formats des Windows et leur partage. Importez-la dans Personnaliser, à côté du connecteur.",
		downloadSkill: "Télécharger la skill",
		vibeOpen: "Ouvrez les connecteurs de Mistral Vibe",
		vibeHow:
			"Dans Contexte, puis Connecteurs, choisissez Nouveau connecteur, puis Nouveau connecteur personnalisé. Donnez-lui un nom sans espace, comme cahut, et collez l'adresse comme serveur.",
		vibeAdd: "Choisissez Ajouter un connecteur, puis autorisez Mistral Vibe",
		vibePlans:
			"Toutes les offres de Mistral Vibe ajoutent des connecteurs personnalisés, Free comprise. Seul un administrateur du compte en ajoute un : avec les offres Free et Pro, son titulaire.",
		chatgptIP: (
			<>
				ChatGPT ne se connecte pas à une Cahut à une adresse IP. Demandez aux administrateurs du serveur de la déplacer vers un
				domaine, puis ajoutez-la à ChatGPT à sa nouvelle adresse.
			</>
		),
		chatgptPlans:
			"ChatGPT ne publie dans une Cahut qu'avec ses offres Business, Enterprise et Edu. Avec Free, Go, Plus ou Pro, il ne le peut pas : connectez Claude ou Mistral Vibe, qui marchent aussi avec leur offre gratuite.",
		chatgptOpen: "Dans ChatGPT sur le web, ouvrez Plugins",
		chatgptHow:
			"Choisissez Ajouter, puis Ajouter un serveur MCP personnalisé. Donnez-lui un nom et collez l'adresse comme URL du serveur : son authentification est déjà OAuth. Cochez la case qui dit que vous comprenez, puis choisissez Créer en tant que plugin.",
		chatgptApproveHelp: "Il vous demande de confirmer les appels qui changent quelque chose.",
		pluginTitle: "Dans Claude Code, ajoutez le plugin Cahut",
		plugin: (host) => (
			<>
				Le plugin connecte cette Cahut et apprend à Claude à faire et publier des Windows. Il demande le domaine de cette Cahut,{" "}
				<code class="c-mono">{host}</code>.
			</>
		),
		addMarketplace: "la commande qui ajoute la marketplace",
		installPlugin: "la commande qui installe le plugin",
		approveWhenSent: "Autorisez Claude Code quand il vous envoie ici",
		orTerminal: "Ou connectez seulement la Cahut, depuis un terminal :",
		addCahut: "la commande qui ajoute la Cahut",
		login: "la commande qui connecte Claude Code",
		otherGive: "Donnez-la à votre client MCP, en HTTP",
		other: (
			<>
				Un client qui se connecte par OAuth vous envoie ici. Un autre prend un jeton personnel, que crée la page{" "}
				<Link class="c-link" href="/access">
					IA et accès
				</Link>
				.
			</>
		),
		film: "Voir le film · 45 s",
	},
});

/** The names of the clients, as people know them. */
export const clientNames: Record<ClientId, string> = {
	claude: "Claude",
	vibe: "Mistral Vibe",
	chatgpt: "ChatGPT",
	"claude-code": "Claude Code",
	other: t.others,
};

/**
 * The films of 45 seconds that show the steps of a client, in the language
 * of the console: the listing of My Cahut in the directory of Claude, or the
 * address added as a connector.
 */
const films: Partial<Record<ClientId | "listing", Record<"en" | "fr", string>>> = {
	listing: { en: "JMhOVCx-L74", fr: "XYfDC_bodZo" },
	claude: { en: "am2iBk5AX0U", fr: "QtWHG6qjjKw" },
	vibe: { en: "I_2rw66sR5s", fr: "vWm_Wy6qS2Q" },
};

/** The link to the film of the steps of client, if there is one. */
export function Film({ client, listing }: { client: ClientId; listing?: string }) {
	const film = films[client === "claude" && listing ? "listing" : client];
	if (!film) return null;
	return (
		<a class="film" href={`https://www.youtube.com/watch?v=${film[lang]}`} target="_blank" rel="noopener noreferrer">
			<Play class="i" size={16} aria-hidden="true" />
			{t.film}
		</a>
	);
}

/** A step: what to do, then what helps do it. */
function Step({ title, children }: { title: ComponentChildren; children?: ComponentChildren }) {
	return (
		<li>
			<div class="steps__body">
				<b>{title}</b>
				{children}
			</div>
		</li>
	);
}

/** A link out of the console, as a button. */
function Out({ href, children, accent }: { href: string; children: ComponentChildren; accent?: boolean }) {
	return (
		<a class={accent ? "c-btn c-btn--accent" : "c-btn c-btn--secondary"} href={href} target="_blank" rel="noopener noreferrer">
			{children}
			<ExternalLink class="i" size={16} aria-hidden="true" />
		</a>
	);
}

/**
 * The steps that connect client to this Cahut, at the address mcp. A Cahut
 * that the directory of Claude lists leads Claude to its listing.
 */
export function ClientSteps({ client, mcp, listing }: { client: ClientId; mcp: string; listing?: string }) {
	switch (client) {
		case "claude":
			return listing ? <ClaudeListed mcp={mcp} listing={listing} /> : <ClaudeByAddress mcp={mcp} />;
		case "vibe":
			return <Vibe mcp={mcp} />;
		case "chatgpt":
			return <ChatGPT mcp={mcp} />;
		case "claude-code":
			return <ClaudeCode mcp={mcp} />;
		case "other":
			return <Other mcp={mcp} />;
	}
}

function ClaudeListed({ mcp, listing }: { mcp: string; listing: string }) {
	return (
		<>
			<ol class="steps">
				<Step title={t.openListing}>
					<div>
						<Out href={listing} accent>
							{t.openInClaude}
						</Out>
					</div>
					<p class="c-small c-muted">{t.openListingHelp}</p>
				</Step>
				<Step title={t.selectConnect}>
					<p class="c-small c-muted">{t.claudeSendsYou}</p>
				</Step>
				<Step title={t.approve("Claude")}>
					<p class="c-small c-muted">{t.approveHelp}</p>
				</Step>
			</ol>
			<details class="steps__more">
				<summary>{t.yourself}</summary>
				<CopyField label={t.address} value={mcp} />
				<p class="c-small c-muted">{t.claudeConnector}</p>
			</details>
			<Skill />
		</>
	);
}

function ClaudeByAddress({ mcp }: { mcp: string }) {
	return (
		<>
			<ol class="steps">
				<Step title={t.copyAddress}>
					<CopyField label={t.address} value={mcp} />
				</Step>
				<Step title={t.openConnectors}>
					<div>
						<Out href="https://claude.ai/customize/connectors" accent>
							{t.openClaude}
						</Out>
					</div>
					<p class="c-small c-muted">{t.claudeConnector}</p>
				</Step>
				<Step title={t.connectAndApprove("Claude")}>
					<p class="c-small c-muted">{t.claudeFree}</p>
				</Step>
			</ol>
			<Skill />
		</>
	);
}

function Skill() {
	return (
		<div class="steps__extra">
			<p class="c-small c-muted">{t.skill}</p>
			<div>
				<a class="c-btn c-btn--secondary c-btn--sm" href="/-/cahut-skill.zip" download>
					<Download class="i" size={14} aria-hidden="true" />
					{t.downloadSkill}
				</a>
			</div>
		</div>
	);
}

function Vibe({ mcp }: { mcp: string }) {
	return (
		<ol class="steps">
			<Step title={t.copyAddress}>
				<CopyField label={t.address} value={mcp} />
			</Step>
			<Step title={t.vibeOpen}>
				<p class="c-small c-muted">{t.vibeHow}</p>
			</Step>
			<Step title={t.vibeAdd}>
				<p class="c-small c-muted">{t.vibePlans}</p>
			</Step>
		</ol>
	);
}

function ChatGPT({ mcp }: { mcp: string }) {
	// The servers of ChatGPT never call an MCP server at an IP address, and
	// it says it finds no OAuth settings there.
	if (ipAddress.test(location.hostname)) {
		return <Notice>{t.chatgptIP}</Notice>;
	}
	return (
		<>
			<Notice>{t.chatgptPlans}</Notice>
			<ol class="steps">
				<Step title={t.copyAddress}>
					<CopyField label={t.address} value={mcp} />
				</Step>
				<Step title={t.chatgptOpen}>
					<p class="c-small c-muted">{t.chatgptHow}</p>
				</Step>
				<Step title={t.approve("ChatGPT")}>
					<p class="c-small c-muted">{t.chatgptApproveHelp}</p>
				</Step>
			</ol>
		</>
	);
}

function ClaudeCode({ mcp }: { mcp: string }) {
	return (
		<>
			<ol class="steps">
				<Step title={t.pluginTitle}>
					<p class="c-small c-muted">{t.plugin(location.host)}</p>
					<CopyField label={t.addMarketplace} value="/plugin marketplace add cahutcloud/cahut" />
					<CopyField label={t.installPlugin} value="/plugin install cahut@cahut" />
				</Step>
				<Step title={t.approveWhenSent}>
					<p class="c-small c-muted">{t.approveHelp}</p>
				</Step>
			</ol>
			<div class="steps__extra">
				<p class="c-small c-muted">{t.orTerminal}</p>
				<CopyField label={t.addCahut} value={`claude mcp add --transport http cahut ${mcp}`} />
				<CopyField label={t.login} value="claude mcp login cahut" />
			</div>
		</>
	);
}

function Other({ mcp }: { mcp: string }) {
	return (
		<ol class="steps">
			<Step title={t.copyAddress}>
				<CopyField label={t.address} value={mcp} />
			</Step>
			<Step title={t.otherGive}>
				<p class="c-small c-muted">{t.other}</p>
			</Step>
		</ol>
	);
}

const tabs: ClientId[] = ["claude", "vibe", "chatgpt", "claude-code", "other"];

/** The steps to connect each kind of AI client, in tabs. */
export function ConnectSteps({ mcp, listing }: { mcp: string; listing?: string }) {
	const [client, setClient] = useState<ClientId>("claude");
	return (
		<div class="connect">
			<div class="c-tabs" role="tablist" aria-label={t.clients}>
				{tabs.map((c) => (
					<button
						key={c}
						class="c-tab"
						type="button"
						role="tab"
						id={`connect-${c}`}
						aria-selected={client === c}
						aria-controls="connect-panel"
						onClick={() => setClient(c)}
					>
						{clientNames[c]}
					</button>
				))}
			</div>
			<div class="connect__panel" role="tabpanel" id="connect-panel" aria-labelledby={`connect-${client}`}>
				<ClientSteps client={client} mcp={mcp} listing={listing} />
				<Film client={client} listing={listing} />
			</div>
		</div>
	);
}
