// AIs and access: the AIs connected over OAuth, and the personal tokens with
// which AIs act for a person, over MCP. A token shows once, with the command
// that connects Claude Code. See docs/adr/0016-oauth-for-ai-clients.md.
import { Plug } from "lucide-preact";
import { Fragment } from "preact";
import { useState } from "preact/hooks";
import { explain } from "../api";
import type { Connection, NewToken, TokenInfo } from "../api.gen";
import { ConfirmButton, CopyField, Failure, PageHead } from "../components";
import { ConnectSteps } from "../connect";
import { useConsole } from "../context";
import { locale, que, words } from "../i18n";
import { useLoad } from "../load";

const t = words({
	en: {
		thisCahut: "This Cahut",
		title: "AIs and access",
		intro: "An AI acts for you over MCP, within what you grant it. Every call it makes goes to the Logbook, and you disconnect it in one click.",
		connect: "Connect an AI",
		connectHelp:
			"Give your AI client the address of this Cahut. It sends you here to sign in, and you choose the Rooms it reaches and whether it may publish. For Claude, the skill teaches it the formats, the sandbox of Windows, sharing them, and Knock first.",
		yourAIs: "Your connected AIs",
		theirAIs: "The connected AIs of others",
		theirAIsHelp: "As an administrator, you can disconnect them.",
		yourTokens: "Your tokens",
		theirTokens: "The tokens of others",
		theirTokensHelp: "As an administrator, you can revoke them.",
		createTitle: "Create a personal token",
		createHelp: "For an AI that cannot sign in, such as a script or a client without OAuth. A token has all your rights, in every Room.",
		tokenFor: "What the token is for",
		tokenForPlaceholder: "Claude Code on my laptop",
		tokenForHelp: "The Logbook shows this name with every call the AI makes.",
		create: "Create a token",
		created: (name: string) => (
			<>
				Token for <b>{name}</b>. It is shown once: copy it now. Cahut keeps only a hash of it.
			</>
		),
		theToken: "the token",
		toConnect: "To connect Claude Code, run this on the machine where it runs:",
		theCommand: "the command",
		never: "Never",
		activeAI: "Active",
		expiredAI: "Expired",
		disconnected: "Disconnected",
		noAI: "No AI connected yet.",
		person: "Person",
		ai: "AI",
		may: "May",
		rooms: "Rooms",
		lastUsed: "Last used",
		state: "State",
		readAndPublish: "Read and publish",
		readOnly: "Read only",
		allRooms: "All Rooms",
		noneLeft: "None left",
		change: "Change",
		disconnect: "Disconnect",
		disconnectQuestion: (client: string) => `Disconnect ${client}? It loses access at once.`,
		whatMayDo: (client: string) => `What ${client} may do`,
		readAndPublishHelp: "Read and publish: everything you can do in its Rooms",
		where: "Where",
		allYourRooms: "All your Rooms, and those you join later",
		onlyThese: "Only these Rooms",
		personalRoom: "Your personal Room",
		save: "Save",
		cancel: "Cancel",
		noToken: "No token yet.",
		name: "Name",
		createdAt: "Created",
		revoked: "Revoked",
		activeToken: "Active",
		revoke: "Revoke",
		revokeQuestion: (name: string) => `Revoke ${name}? The AI that uses it loses access at once.`,
	},
	fr: {
		thisCahut: "Ce Cahut",
		title: "IA et accès",
		intro:
			"Une IA agit pour vous par MCP, dans les limites de ce que vous lui accordez. Chacun de ses appels va au Logbook, et vous la déconnectez en un clic.",
		connect: "Connecter une IA",
		connectHelp:
			"Donnez à votre client d'IA l'adresse de ce Cahut. Il vous envoie ici pour vous connecter, et vous choisissez les Rooms qu'il atteint et s'il peut publier. Pour Claude, la skill lui apprend les formats, le bac à sable des Windows, leur partage, et Knock first.",
		yourAIs: "Vos IA connectées",
		theirAIs: "Les IA connectées des autres",
		theirAIsHelp: "En tant qu'administrateur, vous pouvez les déconnecter.",
		yourTokens: "Vos jetons",
		theirTokens: "Les jetons des autres",
		theirTokensHelp: "En tant qu'administrateur, vous pouvez les révoquer.",
		createTitle: "Créer un jeton personnel",
		createHelp:
			"Pour une IA qui ne peut pas se connecter, comme un script ou un client sans OAuth. Un jeton a tous vos droits, dans chaque Room.",
		tokenFor: "À quoi sert le jeton",
		tokenForPlaceholder: "Claude Code sur mon portable",
		tokenForHelp: "Le Logbook montre ce nom avec chaque appel de l'IA.",
		create: "Créer un jeton",
		created: (name) => (
			<>
				Jeton pour <b>{name}</b>. Il ne s'affiche qu'une fois : copiez-le maintenant. Cahut n'en garde qu'une empreinte.
			</>
		),
		theToken: "le jeton",
		toConnect: "Pour connecter Claude Code, lancez cette commande sur la machine où il tourne :",
		theCommand: "la commande",
		never: "Jamais",
		activeAI: "Active",
		expiredAI: "Expirée",
		disconnected: "Déconnectée",
		noAI: "Pas encore d'IA connectée.",
		person: "Personne",
		ai: "IA",
		may: "Peut",
		rooms: "Rooms",
		lastUsed: "Dernier usage",
		state: "État",
		readAndPublish: "Lecture et publication",
		readOnly: "Lecture seule",
		allRooms: "Toutes les Rooms",
		noneLeft: "Plus aucune",
		change: "Modifier",
		disconnect: "Déconnecter",
		disconnectQuestion: (client) => `Déconnecter ${client} ? Son accès cesse aussitôt.`,
		whatMayDo: (client) => `Ce ${que(client)} peut faire`,
		readAndPublishHelp: "Lecture et publication : tout ce que vous pouvez faire dans ses Rooms",
		where: "Où",
		allYourRooms: "Toutes vos Rooms, et celles que vous rejoindrez plus tard",
		onlyThese: "Seulement ces Rooms",
		personalRoom: "Votre Room personnelle",
		save: "Enregistrer",
		cancel: "Annuler",
		noToken: "Pas encore de jeton.",
		name: "Nom",
		createdAt: "Créé",
		revoked: "Révoqué",
		activeToken: "Actif",
		revoke: "Révoquer",
		revokeQuestion: (name) => `Révoquer ${name} ? L'IA qui s'en sert perd son accès aussitôt.`,
	},
});

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
	const theirTokens = (others.data?.[0].tokens ?? []).filter((tk) => tk.user !== api.session.handle);
	const theirAIs = (others.data?.[1].connections ?? []).filter((c) => c.user !== api.session.handle && c.status !== "ended");
	return (
		<>
			<PageHead label={t.thisCahut} title={t.title}>
				{t.intro}
			</PageHead>
			<section class="section c-card panel">
				<h2 class="c-h4">{t.connect}</h2>
				<p class="c-small">{t.connectHelp}</p>
				<ConnectSteps mcp={about.mcp ?? ""} />
			</section>
			<section class="section">
				<h2 class="c-h4">{t.yourAIs}</h2>
				<ConnectedAIs connections={connections.filter((c) => c.status !== "ended")} mine onChange={reload} />
			</section>
			{theirAIs.length > 0 && (
				<section class="section">
					<h2 class="c-h4">{t.theirAIs}</h2>
					<p class="c-small c-muted">{t.theirAIsHelp}</p>
					<ConnectedAIs connections={theirAIs} showUser onChange={others.reload} />
				</section>
			)}
			<NewTokenForm mcp={about.mcp ?? ""} onCreated={reload} />
			<section class="section">
				<h2 class="c-h4">{t.yourTokens}</h2>
				<Tokens tokens={tokens} onRevoked={reload} />
			</section>
			{theirTokens.length > 0 && (
				<section class="section">
					<h2 class="c-h4">{t.theirTokens}</h2>
					<p class="c-small c-muted">{t.theirTokensHelp}</p>
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
			<h2 class="c-h4">{t.createTitle}</h2>
			<p class="c-small c-muted">{t.createHelp}</p>
			<form class="form" onSubmit={submit}>
				<label class="c-field">
					<span class="c-field__label">{t.tokenFor}</span>
					<input class="c-input" required placeholder={t.tokenForPlaceholder} value={name} onInput={(e) => setName(e.currentTarget.value)} />
					<span class="c-field__help">{t.tokenForHelp}</span>
				</label>
				{error && <Failure message={error} />}
				<div>
					<button class="c-btn c-btn--accent" type="submit">
						<Plug class="i" size={16} aria-hidden="true" />
						{t.create}
					</button>
				</div>
			</form>
			{token && (
				<div class="result" role="status">
					<p class="c-small">{t.created(token.name)}</p>
					<CopyField label={t.theToken} value={token.token} />
					<p class="c-small">{t.toConnect}</p>
					<CopyField
						label={t.theCommand}
						value={`claude mcp add --transport http cahut ${mcp} --header "Authorization: Bearer ${token.token}"`}
					/>
				</div>
			)}
		</section>
	);
}

const when = (at?: string) => (at ? new Date(at).toLocaleString(locale, { dateStyle: "medium", timeStyle: "short" }) : t.never);

const states: Record<string, { label: string; badge: string }> = {
	active: { label: t.activeAI, badge: "c-badge--success" },
	expired: { label: t.expiredAI, badge: "c-badge--neutral" },
	ended: { label: t.disconnected, badge: "c-badge--neutral" },
};

function ConnectedAIs(props: { connections: Connection[]; showUser?: boolean; mine?: boolean; onChange: () => void }) {
	const api = useConsole();
	const [error, setError] = useState("");
	// The connection whose reach the person changes.
	const [changing, setChanging] = useState(0);
	if (props.connections.length === 0) {
		return <p class="c-muted">{t.noAI}</p>;
	}
	return (
		<div class="table-wrap">
			{error && <Failure message={error} />}
			<table class="table">
				<thead>
					<tr>
						{props.showUser && <th>{t.person}</th>}
						<th>{t.ai}</th>
						<th>{t.may}</th>
						<th>{t.rooms}</th>
						<th>{t.lastUsed}</th>
						<th>{t.state}</th>
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
									<td>{c.scope === "write" ? t.readAndPublish : t.readOnly}</td>
									<td class="c-small">{c.all_rooms ? t.allRooms : c.rooms.join(", ") || t.noneLeft}</td>
									<td class="c-mono c-small">{when(c.last_used_at)}</td>
									<td>
										<span class={`c-badge ${state.badge}`}>{state.label}</span>
									</td>
									<td class="table__actions">
										{props.mine && c.status !== "ended" && changing !== c.id && (
											<button class="c-btn c-btn--secondary c-btn--sm" type="button" onClick={() => setChanging(c.id)}>
												{t.change}
											</button>
										)}
										{c.status !== "ended" && (
											<ConfirmButton
												label={t.disconnect}
												question={t.disconnectQuestion(c.client)}
												confirm={t.disconnect}
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
				<legend class="c-field__label">{t.whatMayDo(c.client)}</legend>
				<label class="c-check">
					<input type="radio" name={`scope-${c.id}`} checked={scope === "write"} onChange={() => setScope("write")} />
					<span>{t.readAndPublishHelp}</span>
				</label>
				<label class="c-check">
					<input type="radio" name={`scope-${c.id}`} checked={scope === "read"} onChange={() => setScope("read")} />
					<span>{t.readOnly}</span>
				</label>
			</fieldset>
			<fieldset class="consent__group">
				<legend class="c-field__label">{t.where}</legend>
				<label class="c-check">
					<input type="radio" name={`rooms-${c.id}`} checked={allRooms} onChange={() => setAllRooms(true)} />
					<span>{t.allYourRooms}</span>
				</label>
				<label class="c-check">
					<input type="radio" name={`rooms-${c.id}`} checked={!allRooms} onChange={() => setAllRooms(false)} />
					<span>{t.onlyThese}</span>
				</label>
				{!allRooms && (
					<div class="consent__rooms">
						{mine.map((r) => (
							<label class="c-check" key={r.name}>
								<input type="checkbox" checked={chosen.includes(r.name)} onChange={(e) => toggle(r.name, e.currentTarget.checked)} />
								<span>
									{r.kind === "personal" ? t.personalRoom : r.title} <span class="c-mono c-muted">{r.name}</span>
								</span>
							</label>
						))}
					</div>
				)}
			</fieldset>
			{(error || rooms.error) && <Failure message={error || (rooms.error ?? "")} />}
			<div class="actions">
				<button class="c-btn c-btn--primary c-btn--sm" type="submit" disabled={!allRooms && chosen.length === 0}>
					{t.save}
				</button>
				<button class="c-btn c-btn--ghost c-btn--sm" type="button" onClick={onDone}>
					{t.cancel}
				</button>
			</div>
		</form>
	);
}

function Tokens({ tokens, showUser, onRevoked }: { tokens: TokenInfo[]; showUser?: boolean; onRevoked: () => void }) {
	const api = useConsole();
	const [error, setError] = useState("");
	if (tokens.length === 0) {
		return <p class="c-muted">{t.noToken}</p>;
	}
	return (
		<div class="table-wrap">
			{error && <Failure message={error} />}
			<table class="table">
				<thead>
					<tr>
						{showUser && <th>{t.person}</th>}
						<th>{t.name}</th>
						<th>{t.createdAt}</th>
						<th>{t.lastUsed}</th>
						<th>{t.state}</th>
						<th />
					</tr>
				</thead>
				<tbody>
					{tokens.map((tk) => (
						<tr key={tk.id}>
							{showUser && <td class="c-mono">{tk.user}</td>}
							<td>{tk.name}</td>
							<td class="c-mono c-small">{when(tk.created_at)}</td>
							<td class="c-mono c-small">{when(tk.last_used_at)}</td>
							<td>
								{tk.revoked_at ? (
									<span class="c-badge c-badge--neutral">{t.revoked}</span>
								) : (
									<span class="c-badge c-badge--success">{t.activeToken}</span>
								)}
							</td>
							<td class="table__actions">
								{!tk.revoked_at && (
									<ConfirmButton
										label={t.revoke}
										question={t.revokeQuestion(tk.name)}
										confirm={t.revoke}
										danger
										onConfirm={async () => {
											setError("");
											try {
												await api.cahut("revoke_token", { id: tk.id });
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
