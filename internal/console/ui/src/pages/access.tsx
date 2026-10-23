// AIs and access: the AIs connected over OAuth, and the personal tokens with
// which AIs act for a person, over MCP. A token shows once, with the command
// that connects Claude Code. See docs/adr/0016-oauth-for-ai-clients.md.
import { Plug } from "lucide-preact";
import { Fragment } from "preact";
import { useState } from "preact/hooks";
import { explain } from "../api";
import type { Connection, NewToken, TokenInfo } from "../api.gen";
import { ConfirmButton, CopyField, Failure, PageHead } from "../components";
import { useConsole } from "../context";
import { useLoad } from "../load";

export function Access() {
	const api = useConsole();
	const { data, error, reload } = useLoad(() =>
		Promise.all([
			api.cahut("get_cahut", {}),
			api.cahut("list_tokens", { user: api.session.handle }),
			api.cahut("list_connected_ais", { user: api.session.handle }),
		]),
	);
	const others = useLoad(() =>
		api.isAdmin ? Promise.all([api.cahut("list_tokens", {}), api.cahut("list_connected_ais", {})]) : Promise.resolve(null),
	);
	if (error) return <Failure message={error} />;
	if (!data) return null;
	const [about, { tokens }, { connections }] = data;
	const theirTokens = (others.data?.[0].tokens ?? []).filter((t) => t.user !== api.session.handle);
	const theirAIs = (others.data?.[1].connections ?? []).filter((c) => c.user !== api.session.handle && c.status !== "ended");
	return (
		<>
			<PageHead label="This Cahut" title="AIs and access">
				An AI acts for you over MCP, within what you grant it. Every call it makes goes to the Logbook, and you disconnect it in
				one click.
			</PageHead>
			<section class="section c-card panel">
				<h2 class="c-h4">Connect an AI</h2>
				<p class="c-small">
					Give your AI client the address of this Cahut. It sends you here to sign in, and you choose the Rooms it reaches and
					whether it may publish.
				</p>
				<CopyField label="the address" value={about.mcp ?? ""} />
				<p class="c-small c-muted">
					In claude.ai or Claude Desktop, add it as a custom connector. In Claude Code, run{" "}
					<code class="c-mono">claude mcp add --transport http cahut {about.mcp}</code>, then <code class="c-mono">/mcp</code>.
				</p>
			</section>
			<section class="section">
				<h2 class="c-h4">Your connected AIs</h2>
				<ConnectedAIs connections={connections.filter((c) => c.status !== "ended")} mine onChange={reload} />
			</section>
			{theirAIs.length > 0 && (
				<section class="section">
					<h2 class="c-h4">The connected AIs of others</h2>
					<p class="c-small c-muted">As an administrator, you can disconnect them.</p>
					<ConnectedAIs connections={theirAIs} showUser onChange={others.reload} />
				</section>
			)}
			<NewTokenForm mcp={about.mcp ?? ""} onCreated={reload} />
			<section class="section">
				<h2 class="c-h4">Your tokens</h2>
				<Tokens tokens={tokens} onRevoked={reload} />
			</section>
			{theirTokens.length > 0 && (
				<section class="section">
					<h2 class="c-h4">The tokens of others</h2>
					<p class="c-small c-muted">As an administrator, you can revoke them.</p>
					<Tokens tokens={theirTokens} showUser onRevoked={others.reload} />
				</section>
			)}
		</>
	);
}

function NewTokenForm({ mcp, onCreated }: { mcp: string; onCreated: () => void }) {
	const api = useConsole();
	const [name, setName] = useState("");
	const [error, setError] = useState("");
	const [token, setToken] = useState<NewToken | null>(null);
	const submit = async (e: Event) => {
		e.preventDefault();
		setError("");
		try {
			setToken(await api.cahut("create_token", { name }));
			setName("");
			onCreated();
		} catch (err) {
			setError(explain(err));
		}
	};
	return (
		<section class="section c-card panel">
			<h2 class="c-h4">Create a personal token</h2>
			<p class="c-small c-muted">
				For an AI that cannot sign in, such as a script or a client without OAuth. A token has all your rights, in every Room.
			</p>
			<form class="form" onSubmit={submit}>
				<label class="c-field">
					<span class="c-field__label">What the token is for</span>
					<input class="c-input" required placeholder="Claude Code on my laptop" value={name} onInput={(e) => setName(e.currentTarget.value)} />
					<span class="c-field__help">The Logbook shows this name with every call the AI makes.</span>
				</label>
				{error && <Failure message={error} />}
				<div>
					<button class="c-btn c-btn--accent" type="submit">
						<Plug class="i" size={16} aria-hidden="true" />
						Create a token
					</button>
				</div>
			</form>
			{token && (
				<div class="result" role="status">
					<p class="c-small">
						Token for <b>{token.name}</b>. It is shown once: copy it now. Cahut keeps only a hash of it.
					</p>
					<CopyField label="the token" value={token.token} />
					<p class="c-small">To connect Claude Code, run this on the machine where it runs:</p>
					<CopyField
						label="the command"
						value={`claude mcp add --transport http cahut ${mcp} --header "Authorization: Bearer ${token.token}"`}
					/>
				</div>
			)}
		</section>
	);
}

const when = (at?: string) => (at ? new Date(at).toLocaleString("en-GB", { dateStyle: "medium", timeStyle: "short" }) : "Never");

const states: Record<string, { label: string; badge: string }> = {
	active: { label: "Active", badge: "c-badge--success" },
	expired: { label: "Expired", badge: "c-badge--neutral" },
	ended: { label: "Disconnected", badge: "c-badge--neutral" },
};

function ConnectedAIs(props: { connections: Connection[]; showUser?: boolean; mine?: boolean; onChange: () => void }) {
	const api = useConsole();
	const [error, setError] = useState("");
	// The connection whose reach the person changes.
	const [changing, setChanging] = useState(0);
	if (props.connections.length === 0) {
		return <p class="c-muted">No AI connected yet.</p>;
	}
	return (
		<div class="table-wrap">
			{error && <Failure message={error} />}
			<table class="table">
				<thead>
					<tr>
						{props.showUser && <th>Person</th>}
						<th>AI</th>
						<th>May</th>
						<th>Rooms</th>
						<th>Last used</th>
						<th>State</th>
						<th />
					</tr>
				</thead>
				<tbody>
					{props.connections.map((c) => {
						const state = states[c.status] ?? states.ended;
						return (
							<Fragment key={c.id}>
								<tr>
									{props.showUser && <td class="c-mono">{c.user}</td>}
									<td>
										{c.client}
										<div class="c-mono c-small c-muted">{c.redirect_host}</div>
									</td>
									<td>{c.scope === "write" ? "Read and publish" : "Read only"}</td>
									<td class="c-small">{c.all_rooms ? "All Rooms" : c.rooms.join(", ") || "None left"}</td>
									<td class="c-mono c-small">{when(c.last_used_at)}</td>
									<td>
										<span class={`c-badge ${state.badge}`}>{state.label}</span>
									</td>
									<td class="table__actions">
										{props.mine && c.status !== "ended" && changing !== c.id && (
											<button class="c-btn c-btn--secondary c-btn--sm" type="button" onClick={() => setChanging(c.id)}>
												Change
											</button>
										)}
										{c.status !== "ended" && (
											<ConfirmButton
												label="Disconnect"
												question={`Disconnect ${c.client}? It loses access at once.`}
												confirm="Disconnect"
												danger
												onConfirm={async () => {
													setError("");
													try {
														await api.cahut("disconnect_ai", { id: c.id });
														props.onChange();
													} catch (err) {
														setError(explain(err));
													}
												}}
											/>
										)}
									</td>
								</tr>
								{changing === c.id && (
									<tr>
										<td colSpan={props.showUser ? 7 : 6}>
											<ReachForm
												connection={c}
												onDone={() => {
													setChanging(0);
													props.onChange();
												}}
											/>
										</td>
									</tr>
								)}
							</Fragment>
						);
					})}
				</tbody>
			</table>
		</div>
	);
}

/** Changes what a connected AI may do, and where, as on the consent screen. */
function ReachForm({ connection: c, onDone }: { connection: Connection; onDone: () => void }) {
	const api = useConsole();
	const rooms = useLoad(() => api.cahut("list_rooms", {}));
	const [scope, setScope] = useState(c.scope);
	const [allRooms, setAllRooms] = useState(c.all_rooms);
	const [chosen, setChosen] = useState<string[]>(c.rooms);
	const [error, setError] = useState("");
	const mine = (rooms.data?.rooms ?? []).filter((r) => r.role);
	const toggle = (name: string, on: boolean) => setChosen((list) => (on ? [...list, name] : list.filter((r) => r !== name)));
	const save = async (e: Event) => {
		e.preventDefault();
		setError("");
		try {
			await api.cahut("set_ai_reach", { id: c.id, scope, all_rooms: allRooms, rooms: allRooms ? [] : chosen });
			onDone();
		} catch (err) {
			setError(explain(err));
		}
	};
	return (
		<form class="form" onSubmit={save}>
			<fieldset class="consent__group">
				<legend class="c-field__label">What {c.client} may do</legend>
				<label class="c-check">
					<input type="radio" name={`scope-${c.id}`} checked={scope === "write"} onChange={() => setScope("write")} />
					<span>Read and publish: everything you can do in its Rooms</span>
				</label>
				<label class="c-check">
					<input type="radio" name={`scope-${c.id}`} checked={scope === "read"} onChange={() => setScope("read")} />
					<span>Read only</span>
				</label>
			</fieldset>
			<fieldset class="consent__group">
				<legend class="c-field__label">Where</legend>
				<label class="c-check">
					<input type="radio" name={`rooms-${c.id}`} checked={allRooms} onChange={() => setAllRooms(true)} />
					<span>All your Rooms, and those you join later</span>
				</label>
				<label class="c-check">
					<input type="radio" name={`rooms-${c.id}`} checked={!allRooms} onChange={() => setAllRooms(false)} />
					<span>Only these Rooms</span>
				</label>
				{!allRooms && (
					<div class="consent__rooms">
						{mine.map((r) => (
							<label class="c-check" key={r.name}>
								<input type="checkbox" checked={chosen.includes(r.name)} onChange={(e) => toggle(r.name, e.currentTarget.checked)} />
								<span>
									{r.kind === "personal" ? "Your personal Room" : r.title} <span class="c-mono c-muted">{r.name}</span>
								</span>
							</label>
						))}
					</div>
				)}
			</fieldset>
			{(error || rooms.error) && <Failure message={error || (rooms.error ?? "")} />}
			<div class="actions">
				<button class="c-btn c-btn--primary c-btn--sm" type="submit" disabled={!allRooms && chosen.length === 0}>
					Save
				</button>
				<button class="c-btn c-btn--ghost c-btn--sm" type="button" onClick={onDone}>
					Cancel
				</button>
			</div>
		</form>
	);
}

function Tokens({ tokens, showUser, onRevoked }: { tokens: TokenInfo[]; showUser?: boolean; onRevoked: () => void }) {
	const api = useConsole();
	const [error, setError] = useState("");
	if (tokens.length === 0) {
		return <p class="c-muted">No token yet.</p>;
	}
	return (
		<div class="table-wrap">
			{error && <Failure message={error} />}
			<table class="table">
				<thead>
					<tr>
						{showUser && <th>Person</th>}
						<th>Name</th>
						<th>Created</th>
						<th>Last used</th>
						<th>State</th>
						<th />
					</tr>
				</thead>
				<tbody>
					{tokens.map((t) => (
						<tr key={t.id}>
							{showUser && <td class="c-mono">{t.user}</td>}
							<td>{t.name}</td>
							<td class="c-mono c-small">{when(t.created_at)}</td>
							<td class="c-mono c-small">{when(t.last_used_at)}</td>
							<td>
								{t.revoked_at ? <span class="c-badge c-badge--neutral">Revoked</span> : <span class="c-badge c-badge--success">Active</span>}
							</td>
							<td class="table__actions">
								{!t.revoked_at && (
									<ConfirmButton
										label="Revoke"
										question={`Revoke ${t.name}? The AI that uses it loses access at once.`}
										confirm="Revoke"
										danger
										onConfirm={async () => {
											setError("");
											try {
												await api.cahut("revoke_token", { id: t.id });
												onRevoked();
											} catch (err) {
												setError(explain(err));
											}
										}}
									/>
								)}
							</td>
						</tr>
					))}
				</tbody>
			</table>
		</div>
	);
}
