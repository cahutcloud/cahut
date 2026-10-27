// The first screen once signed in: for a person, their latest Windows and
// their Rooms; for an administrator of the server, the Cahuts of the server.
import { ArrowUpRight, Plus } from "lucide-preact";
import { useState } from "preact/hooks";
import { explain } from "../api";
import type { NewCahut } from "../api.gen";
import { CopyField, Domain, Failure, Notice, PageHead, plural, WindowCards } from "../components";
import { ConnectSteps, ipAddress } from "../connect";
import { useConsole } from "../context";
import { de, locale, que, words } from "../i18n";
import { useLoad } from "../load";
import { href, Link } from "../router";
import { RoomCards } from "./rooms";
import { LimitsTable } from "./settings";

const t = words({
	en: {
		overview: "Overview",
		signedInAs: (name: string, admin: boolean, server: boolean) =>
			`Signed in as ${name}, ${admin ? "an administrator of this Cahut" : "a member of this Cahut"}${server ? (admin ? " and of the server" : ", and an administrator of the server") : ""}.`,
		connectFirst: "Connect your first AI",
		connectText: "Your AI publishes on Cahut over MCP. Connect it, then ask it to make a document, a deck or a dashboard.",
		latestWindows: "Latest Windows",
		noWindow: () => (
			<>
				No Window yet. Ask your AI to publish one: <Link class="c-link" href="/access">connect an AI</Link> first.
			</>
		),
		yourRooms: "Your Rooms",
		yourLimits: "Your limits",
		cahutLimits: "Limits of this Cahut",
		thisServer: "This server",
		administer: (host: string) =>
			`You administer this server, and here the Cahut of ${host}. For another Cahut, sign in at its own console.`,
		declaredOn: (date: string) => `Declared on ${date}`,
		atConsole: "You are at its console.",
		openConsole: "Open its console",
		moveToDomain: "Move to a domain",
		changeDomain: "Change its domain",
		moveText: (domain: string) => (
			<>
				Point the DNS record of the new domain to this server first. Everything in the Cahut stays, but the links that name{" "}
				<span class="c-mono">{domain}</span> stop working, and AIs connect again at the new address.
			</>
		),
		newDomain: "New domain",
		moveCahut: "Move the Cahut",
		cancel: "Cancel",
		declareCahut: "Declare a Cahut",
		declareText: "Point the DNS record of the domain to this server first. The Cahut gets its certificate at the first visit.",
		domain: "Domain",
		firstAdmin: "Handle of its first administrator",
		theirName: "Their name",
		declare: "Declare",
		declared: (domain: string) => (
			<>
				Declared <span class="c-mono">{domain}</span>.
			</>
		),
		sendLink: (name: string) => ` Send ${name} this link, where they choose their password. It works once, for 7 days.`,
		activationLink: "the activation link",
	},
	fr: {
		overview: "Vue d'ensemble",
		signedInAs: (name, admin, server) =>
			`Connecté en tant ${que(name)}, ${admin ? "administrateur" : "membre"} de cette Cahut${server ? (admin ? " et du serveur" : " et administrateur du serveur") : ""}.`,
		connectFirst: "Connectez votre première IA",
		connectText:
			"Votre IA publie sur Cahut par MCP. Connectez-la, puis demandez-lui un document, une présentation ou un tableau de bord.",
		latestWindows: "Dernières Windows",
		noWindow: () => (
			<>
				Pas encore de Window. Demandez à votre IA d'en publier une : <Link class="c-link" href="/access">connectez une IA</Link>{" "}
				d'abord.
			</>
		),
		yourRooms: "Vos Rooms",
		yourLimits: "Vos limites",
		cahutLimits: "Limites de cette Cahut",
		thisServer: "Ce serveur",
		administer: (host) =>
			`Vous administrez ce serveur, et ici la Cahut ${de(host)}. Pour une autre Cahut, connectez-vous à sa propre console.`,
		declaredOn: (date) => `Déclarée le ${date}`,
		atConsole: "Vous êtes sur sa console.",
		openConsole: "Ouvrir sa console",
		moveToDomain: "Passer à un domaine",
		changeDomain: "Changer de domaine",
		moveText: (domain) => (
			<>
				Faites d'abord pointer l'enregistrement DNS du nouveau domaine vers ce serveur. Tout ce que contient la Cahut reste, mais
				les liens qui nomment <span class="c-mono">{domain}</span> cessent de fonctionner, et les IA se reconnectent à la
				nouvelle adresse.
			</>
		),
		newDomain: "Nouveau domaine",
		moveCahut: "Déplacer la Cahut",
		cancel: "Annuler",
		declareCahut: "Déclarer une Cahut",
		declareText:
			"Faites d'abord pointer l'enregistrement DNS du domaine vers ce serveur. La Cahut obtient son certificat à la première visite.",
		domain: "Domaine",
		firstAdmin: "Identifiant de son premier administrateur",
		theirName: "Son nom",
		declare: "Déclarer",
		declared: (domain) => (
			<>
				La Cahut <span class="c-mono">{domain}</span> est déclarée.
			</>
		),
		sendLink: (name) => ` Envoyez ce lien à ${name}, qui y choisira son mot de passe. Il ne sert qu'une fois, pendant 7 jours.`,
		activationLink: "le lien d'activation",
	},
});

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
			<PageHead label={t.overview} title={<Domain name={location.hostname} />}>
				{t.signedInAs(account.name, account.role === "admin", api.serverSession !== null)}
			</PageHead>
			{noAI && (
				<section class="section c-card panel">
					<h2 class="c-h4">{t.connectFirst}</h2>
					<p class="c-small">{t.connectText}</p>
					<ConnectSteps mcp={about.mcp ?? `${location.origin}/mcp`} />
				</section>
			)}
			<section class="section">
				<h2 class="c-h4">{t.latestWindows}</h2>
				{windows.length === 0 ? (
					<Notice>{t.noWindow()}</Notice>
				) : (
					<WindowCards windows={windows.slice(0, 6)} showRoom />
				)}
			</section>
			<section class="section">
				<h2 class="c-h4">{t.yourRooms}</h2>
				<RoomCards rooms={rooms} />
			</section>
			{account.role !== "admin" && Object.keys(about.limits).length > 0 && (
				<section class="section">
					<h2 class="c-h4">{about.yours ? t.yourLimits : t.cahutLimits}</h2>
					<LimitsTable settings={about} />
				</section>
			)}
		</>
	);
}

/** The origin of the console of a Cahut of this server: each has its own, on the same port of its own domain. */
export function originOf(domain: string): string {
	return `${location.protocol}//${domain}${location.port ? `:${location.port}` : ""}`;
}

/** The Cahuts of the server, for its administrators. */
export function ServerOverview() {
	const api = useConsole();
	const { data, error, reload } = useLoad(() => api.server("list_domains", {}));
	if (error) return <Failure message={error} />;
	if (!data) return null;
	const origin = originOf;
	return (
		<>
			<PageHead label={t.thisServer} title={plural(data.cahuts.length, "Cahut", "Cahuts")}>
				{t.administer(location.hostname)}
			</PageHead>
			<div class="cards">
				{data.cahuts.map((c) => (
					<article key={c.domain} class="c-card card">
						<div class="c-window__title c-mono">{c.domain}</div>
						<span class="c-small c-muted">{t.declaredOn(new Date(c.created_at).toLocaleDateString(locale, { dateStyle: "medium" }))}</span>
						{c.domain === location.hostname ? (
							<span class="c-small">{t.atConsole}</span>
						) : (
							<a class="c-link c-small" href={`${origin(c.domain)}${href("/server/sign-in")}`}>
								{t.openConsole}
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
					{ip ? t.moveToDomain : t.changeDomain}
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
			<p class="c-small">{t.moveText(props.domain)}</p>
			<label class="c-field">
				<span class="c-field__label">{t.newDomain}</span>
				<input class="c-input c-mono" required placeholder="cahut.example.com" value={to} onInput={(e) => setTo(e.currentTarget.value)} />
			</label>
			{error && <Failure message={error} />}
			<div class="updates__actions">
				<button class="c-btn c-btn--primary c-btn--sm" type="submit" disabled={busy}>
					{t.moveCahut}
				</button>
				<button class="c-btn c-btn--ghost c-btn--sm" type="button" onClick={() => setOpen(false)}>
					{t.cancel}
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
			<h2 class="c-h4">{t.declareCahut}</h2>
			<p class="c-small c-muted">{t.declareText}</p>
			<form class="form" onSubmit={submit}>
				<label class="c-field">
					<span class="c-field__label">{t.domain}</span>
					<input class="c-input c-mono" required placeholder="cahut.example.com" value={domain} onInput={(e) => setDomain(e.currentTarget.value)} />
				</label>
				<div class="form__row">
					<label class="c-field">
						<span class="c-field__label">{t.firstAdmin}</span>
						<input class="c-input" autocapitalize="none" value={handle} onInput={(e) => setHandle(e.currentTarget.value)} />
					</label>
					<label class="c-field">
						<span class="c-field__label">{t.theirName}</span>
						<input class="c-input" value={name} onInput={(e) => setName(e.currentTarget.value)} />
					</label>
				</div>
				{error && <Failure message={error} />}
				<div>
					<button class="c-btn c-btn--primary" type="submit">
						<Plus class="i" size={16} aria-hidden="true" />
						{t.declare}
					</button>
				</div>
			</form>
			{done && (
				<div class="result" role="status">
					<p class="c-small">
						{t.declared(done.domain)}
						{done.admin?.activation && t.sendLink(done.admin.name)}
					</p>
					{done.admin?.activation && (
						<CopyField label={t.activationLink} value={`${origin(done.domain)}${href("/activate")}#${done.admin.activation.code}`} />
					)}
				</div>
			)}
		</section>
	);
}
