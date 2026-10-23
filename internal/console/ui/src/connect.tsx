// How to connect an AI to this Cahut, client by client, and whether one has
// connected: the first start, the overview and AIs and access show it. See
// docs/adr/0025-first-start-in-the-browser.md.
import { CircleCheck, Download, LoaderCircle } from "lucide-preact";
import { useEffect, useState } from "preact/hooks";
import type { Connection, TokenInfo } from "./api.gen";
import { CopyField } from "./components";
import { Link } from "./router";

const clients = [
	{ id: "claude-code", name: "Claude Code" },
	{ id: "claude", name: "Claude apps" },
	{ id: "chatgpt", name: "ChatGPT" },
	{ id: "le-chat", name: "Le Chat" },
	{ id: "other", name: "Other clients" },
] as const;

type Client = (typeof clients)[number]["id"];

/** The steps to connect each kind of AI client, in tabs. */
export function ConnectSteps({ mcp }: { mcp: string }) {
	const [client, setClient] = useState<Client>("claude-code");
	return (
		<div class="connect">
			<div class="c-tabs" role="tablist" aria-label="AI clients">
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
				{client === "chatgpt" && <ChatGPT mcp={mcp} />}
				{client === "le-chat" && <LeChat mcp={mcp} />}
				{client === "other" && <Other mcp={mcp} />}
			</div>
		</div>
	);
}

function ClaudeCode({ mcp }: { mcp: string }) {
	return (
		<>
			<p class="c-small">
				The Cahut plugin connects this Cahut and teaches Claude its house style. It asks for the domain of this Cahut,{" "}
				<code class="c-mono">{location.host}</code>.
			</p>
			<CopyField label="the command that adds the marketplace" value="/plugin marketplace add getcahut/cahut" />
			<CopyField label="the command that installs the plugin" value="/plugin install cahut@cahut" />
			<p class="c-small">Or connect the Cahut only, from a terminal, then sign in here when Claude Code sends you:</p>
			<CopyField label="the command that adds the Cahut" value={`claude mcp add --transport http cahut ${mcp}`} />
			<CopyField label="the command that signs Claude Code in" value="claude mcp login cahut" />
		</>
	);
}

function ClaudeApps({ mcp }: { mcp: string }) {
	return (
		<>
			<p class="c-small">
				In claude.ai or the Claude app, open Settings, then Connectors, and add a custom connector with this address. Connect it:
				Claude sends you here to sign in, and you choose the Rooms it reaches.
			</p>
			<CopyField label="the address" value={mcp} />
			<p class="c-small">Then upload the skill for Claude in the settings of Claude, beside the connector.</p>
			<div>
				<a class="c-btn c-btn--secondary c-btn--sm" href="/-/cahut-skill.zip" download>
					<Download class="i" size={14} aria-hidden="true" />
					Download the skill
				</a>
			</div>
		</>
	);
}

function ChatGPT({ mcp }: { mcp: string }) {
	return (
		<>
			<p class="c-small">
				In ChatGPT on the web, open Plugins, select the plus button, then Add custom MCP server. Give it a name, this address,
				and OAuth as its authentication, then create it. ChatGPT sends you here to sign in, and you choose the Rooms it reaches.
				It asks you to confirm the calls that change something.
			</p>
			<CopyField label="the address" value={mcp} />
			<p class="c-small">Custom MCP servers need a Business, Enterprise or Edu plan of ChatGPT.</p>
		</>
	);
}

function LeChat({ mcp }: { mcp: string }) {
	return (
		<>
			<p class="c-small">
				In Le Chat, open Connectors and add a custom MCP connector with this address. Le Chat finds that the Cahut signs in with
				OAuth and sends you here to sign in.
			</p>
			<CopyField label="the address" value={mcp} />
		</>
	);
}

function Other({ mcp }: { mcp: string }) {
	return (
		<>
			<p class="c-small">
				Give any MCP client this address, over HTTP. A client that signs in with OAuth sends you here; one that does not takes a
				personal token, which <Link class="c-link" href="/access">AIs and access</Link> creates.
			</p>
			<CopyField label="the address" value={mcp} />
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
			{name} is connected. Ask it to publish something on Cahut.
		</p>
	) : (
		<p class="connect__status" role="status">
			<LoaderCircle class="i connect__spin" size={18} aria-hidden="true" />
			Waiting for your AI. This page sees it as soon as it connects.
		</p>
	);
}

