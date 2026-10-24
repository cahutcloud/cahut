// The first screen once signed in: for a person, their latest Windows and
// their Rooms; for an administrator of the server, the Cahuts of the server.
import { ArrowUpRight, Plus } from "lucide-preact";
import { useState } from "preact/hooks";
import { explain } from "../api";
import type { NewCahut } from "../api.gen";
import { CopyField, Failure, Notice, PageHead, plural, WindowCards } from "../components";
import { ConnectSteps, ipAddress } from "../connect";
import { useConsole } from "../context";
import { useLoad } from "../load";
import { href, Link } from "../router";
import { RoomCards } from "./rooms";
import { LimitsTable } from "./settings";

export function Overview() {
	return useConsole().session.kind === "admin" ? <ServerOverview /> : <PersonOverview />;
}

function PersonOverview() {
	const api = useConsole();
	const { data, error } = useLoad(() =>
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
			<PageHead label="Overview" title={location.hostname}>
				Signed in as {account.name}, {account.role === "admin" ? "an administrator" : "a member"} of this Cahut.
			</PageHead>
			{noAI && (
				<section class="section c-card panel">
					<h2 class="c-h4">Connect your first AI</h2>
					<p class="c-small">
						Your AI publishes on Cahut over MCP. Connect it, then ask it to make a document, a deck or a dashboard.
					</p>
					<ConnectSteps mcp={about.mcp ?? `${location.origin}/mcp`} />
				</section>
			)}
			<section class="section">
				<h2 class="c-h4">Latest Windows</h2>
				{windows.length === 0 ? (
					<Notice>
						No Window yet. Ask your AI to publish one: <Link class="c-link" href="/access">connect an AI</Link> first.
					</Notice>
				) : (
					<WindowCards windows={windows.slice(0, 6)} showRoom />
				)}
			</section>
			<section class="section">
				<h2 class="c-h4">Your Rooms</h2>
				<RoomCards rooms={rooms} />
			</section>
			{account.role !== "admin" && Object.keys(about.limits).length > 0 && (
				<section class="section">
					<h2 class="c-h4">{about.yours ? "Your limits" : "Limits of this Cahut"}</h2>
					<LimitsTable settings={about} />
				</section>
			)}
		</>
	);
}

function ServerOverview() {
	const api = useConsole();
	const { data, error, reload } = useLoad(() => api.server("list_domains", {}));
	if (error) return <Failure message={error} />;
	if (!data) return null;
	// Each Cahut has its own console, on the same port of its own domain.
	const origin = (domain: string) => `${location.protocol}//${domain}${location.port ? `:${location.port}` : ""}`;
	return (
		<>
			<PageHead label="This server" title={plural(data.cahuts.length, "Cahut", "Cahuts")}>
				You administer this server, and here the Cahut of {location.hostname}. For another Cahut, sign in at its own console.
			</PageHead>
			<div class="cards">
				{data.cahuts.map((c) => (
					<article key={c.domain} class="c-card card">
						<div class="c-window__title c-mono">{c.domain}</div>
						<span class="c-small c-muted">Declared on {new Date(c.created_at).toLocaleDateString("en-GB", { dateStyle: "medium" })}</span>
						{c.domain === location.hostname ? (
							<span class="c-small">You are at its console.</span>
						) : (
							<a class="c-link c-small" href={`${origin(c.domain)}${href("/server/sign-in")}`}>
								Open its console
								<ArrowUpRight class="i" size={14} aria-hidden="true" />
							</a>
						)}
						<MoveCahut domain={c.domain} origin={origin} onMoved={reload} />
					</article>
				))}
			</div>
			<DeclareCahut origin={origin} onDeclared={reload} />
		</>
	);
}

/**
 * Moves a Cahut to another domain, or from the address of the server to a
 * domain. See docs/adr/0026-cahut-at-an-ip-address.md.
 */
function MoveCahut(props: { domain: string; origin: (domain: string) => string; onMoved: () => void }) {
	const api = useConsole();
	const [open, setOpen] = useState(false);
	const [to, setTo] = useState("");
	const [error, setError] = useState("");
	const [busy, setBusy] = useState(false);
	const ip = ipAddress.test(props.domain);
	if (!open) {
		return (
			<div>
				<button class={`c-btn c-btn--sm ${ip ? "c-btn--primary" : "c-btn--ghost"}`} type="button" onClick={() => setOpen(true)}>
					{ip ? "Move to a domain" : "Change its domain"}
				</button>
			</div>
		);
	}
	const move = async (e: Event) => {
		e.preventDefault();
		setBusy(true);
		setError("");
		try {
			const moved = await api.server("rename_domain", { domain: props.domain, to });
			// The sessions at the old address end with it: sign in again at the
			// new one.
			if (props.domain === location.hostname) {
				location.assign(`${props.origin(moved.domain)}${href("/server/sign-in")}`);
				return;
			}
			setOpen(false);
			props.onMoved();
		} catch (err) {
			setError(explain(err));
		}
		setBusy(false);
	};
	return (
		<form class="form" onSubmit={move}>
			<p class="c-small">
				Point the DNS record of the new domain to this server first. Everything in the Cahut stays, but the links that name{" "}
				<span class="c-mono">{props.domain}</span> stop working, and AIs connect again at the new address.
			</p>
			<label class="c-field">
				<span class="c-field__label">New domain</span>
				<input class="c-input c-mono" required placeholder="cahut.example.com" value={to} onInput={(e) => setTo(e.currentTarget.value)} />
			</label>
			{error && <Failure message={error} />}
			<div class="updates__actions">
				<button class="c-btn c-btn--primary c-btn--sm" type="submit" disabled={busy}>
					Move the Cahut
				</button>
				<button class="c-btn c-btn--ghost c-btn--sm" type="button" onClick={() => setOpen(false)}>
					Cancel
				</button>
			</div>
		</form>
	);
}

/** The form that declares a Cahut, with its first administrator. */
function DeclareCahut({ origin, onDeclared }: { origin: (domain: string) => string; onDeclared: () => void }) {
	const api = useConsole();
	const [domain, setDomain] = useState("");
	const [handle, setHandle] = useState("");
	const [name, setName] = useState("");
	const [error, setError] = useState("");
	const [done, setDone] = useState<NewCahut | null>(null);
	const submit = async (e: Event) => {
		e.preventDefault();
		setError("");
		try {
			const admin = handle || name ? { handle, name } : undefined;
			setDone(await api.server("add_domain", { domain, admin }));
			setDomain("");
			setHandle("");
			setName("");
			onDeclared();
		} catch (err) {
			setError(explain(err));
		}
	};
	return (
		<section class="section c-card panel">
			<h2 class="c-h4">Declare a Cahut</h2>
			<p class="c-small c-muted">
				Point the DNS record of the domain to this server first. The Cahut gets its certificate at the first visit.
			</p>
			<form class="form" onSubmit={submit}>
				<label class="c-field">
					<span class="c-field__label">Domain</span>
					<input class="c-input c-mono" required placeholder="cahut.example.com" value={domain} onInput={(e) => setDomain(e.currentTarget.value)} />
				</label>
				<div class="form__row">
					<label class="c-field">
						<span class="c-field__label">Handle of its first administrator</span>
						<input class="c-input" autocapitalize="none" value={handle} onInput={(e) => setHandle(e.currentTarget.value)} />
					</label>
					<label class="c-field">
						<span class="c-field__label">Their name</span>
						<input class="c-input" value={name} onInput={(e) => setName(e.currentTarget.value)} />
					</label>
				</div>
				{error && <Failure message={error} />}
				<div>
					<button class="c-btn c-btn--primary" type="submit">
						<Plus class="i" size={16} aria-hidden="true" />
						Declare
					</button>
				</div>
			</form>
			{done && (
				<div class="result" role="status">
					<p class="c-small">
						Declared <span class="c-mono">{done.domain}</span>.
						{done.admin?.activation && ` Send ${done.admin.name} this link, where they choose their password. It works once, for 7 days.`}
					</p>
					{done.admin?.activation && (
						<CopyField label="the activation link" value={`${origin(done.domain)}${href("/activate")}#${done.admin.activation.code}`} />
					)}
				</div>
			)}
		</section>
	);
}
