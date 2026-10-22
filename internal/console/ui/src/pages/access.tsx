// AIs and access: the personal tokens with which AIs act for a person, over
// MCP. A token shows once, with the command that connects Claude Code.
import { Plug } from "lucide-preact";
import { useState } from "preact/hooks";
import { explain } from "../api";
import type { NewToken, TokenInfo } from "../api.gen";
import { ConfirmButton, CopyField, Failure, PageHead } from "../components";
import { useConsole } from "../context";
import { useLoad } from "../load";

export function Access() {
	const api = useConsole();
	const { data, error, reload } = useLoad(() =>
		Promise.all([api.cahut("get_cahut", {}), api.cahut("list_tokens", { user: api.session.handle })]),
	);
	const others = useLoad(() => (api.isAdmin ? api.cahut("list_tokens", {}) : Promise.resolve(null)));
	if (error) return <Failure message={error} />;
	if (!data) return null;
	const [about, { tokens }] = data;
	const theirs = (others.data?.tokens ?? []).filter((t) => t.user !== api.session.handle);
	return (
		<>
			<PageHead label="This Cahut" title="AIs and access">
				An AI acts for you with a personal token, over MCP. It has your rights, every call it makes goes to the Logbook, and
				you revoke it in one click.
			</PageHead>
			<NewTokenForm mcp={about.mcp ?? ""} onCreated={reload} />
			<section class="section">
				<h2 class="c-h4">Your tokens</h2>
				<Tokens tokens={tokens} onRevoked={reload} />
			</section>
			{theirs.length > 0 && (
				<section class="section">
					<h2 class="c-h4">The tokens of others</h2>
					<p class="c-small c-muted">As an administrator, you can revoke them.</p>
					<Tokens tokens={theirs} showUser onRevoked={others.reload} />
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
			<h2 class="c-h4">Connect an AI</h2>
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

function Tokens({ tokens, showUser, onRevoked }: { tokens: TokenInfo[]; showUser?: boolean; onRevoked: () => void }) {
	const api = useConsole();
	const [error, setError] = useState("");
	if (tokens.length === 0) {
		return <p class="c-muted">No token yet.</p>;
	}
	const when = (at?: string) => (at ? new Date(at).toLocaleString("en-GB", { dateStyle: "medium", timeStyle: "short" }) : "Never");
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
