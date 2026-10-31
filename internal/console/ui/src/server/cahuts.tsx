// The Cahuts of the server, for its administrators: each with its domain,
// its suspension and the reports that wait, and the page where they set and
// moderate it; then the form that declares a Cahut. See
// docs/adr/0058-a-console-for-the-server.md.
import { ArrowUpRight, Plus } from "lucide-preact";
import { useState } from "preact/hooks";
import { explain } from "../api";
import type { NewCahut } from "../api.gen";
import { CopyField, Failure, originOf, PageHead, plural } from "../components";
import { ipAddress } from "../connect";
import { locale, words } from "../i18n";
import { useLoad } from "../load";
import { Link } from "../router";
import { useServer } from "./context";

const t = words({
	en: {
		thisServer: "This server",
		lead: "Each Cahut has its own domain, its people and its database. Open one to set it and to moderate it.",
		declaredOn: (date: string) => `Declared on ${date}`,
		kinds: { collaborative: "Collaborative Cahut", individual: "Individual Cahut" } as Record<string, string>,
		suspended: "Suspended",
		suspendedOn: (date: string, reason: string) => `Since ${date}: ${reason}`,
		reportsWait: (n: number) => (n === 1 ? "1 report waits" : `${n} reports wait`),
		address: "Its address",
		addressOf: (domain: string) => `Its address: ${domain}`,
		moveToDomain: "Move to a domain",
		moveToDomainOf: (domain: string) => `Move to a domain: ${domain}`,
		changeDomain: "Change its domain",
		changeDomainOf: (domain: string) => `Change its domain: ${domain}`,
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
		thisServer: "Ce serveur",
		lead: "Chaque Cahut a son domaine, ses personnes et sa base de données. Ouvrez-en une pour la régler et la modérer.",
		declaredOn: (date) => `Déclarée le ${date}`,
		kinds: { collaborative: "Cahut collaborative", individual: "Cahut individuelle" },
		suspended: "Suspendue",
		suspendedOn: (date, reason) => `Depuis le ${date} : ${reason}`,
		reportsWait: (n) => (n === 1 ? "1 signalement attend" : `${n} signalements attendent`),
		address: "Son adresse",
		addressOf: (domain) => `Son adresse : ${domain}`,
		moveToDomain: "Passer à un domaine",
		moveToDomainOf: (domain) => `Passer à un domaine : ${domain}`,
		changeDomain: "Changer de domaine",
		changeDomainOf: (domain) => `Changer de domaine : ${domain}`,
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

/** The Cahuts of the server. */
export function Cahuts() {
	const api = useServer();
	const { data, error, reload } = useLoad(() => api.server("list_domains", {}));
	if (error) return <Failure message={error} />;
	if (!data) return null;
	return (
		<>
			<PageHead label={t.thisServer} title={plural(data.cahuts.length, "Cahut", "Cahuts")}>
				{t.lead}
			</PageHead>
			<div class="cards">
				{data.cahuts.map((c) => (
					<article key={c.domain} class="c-card card">
						<Link class="c-window__title c-mono" href={`/cahuts/${encodeURIComponent(c.domain)}`}>
							{c.domain}
						</Link>
						{c.kind && <span class="c-small">{t.kinds[c.kind] ?? c.kind}</span>}
						<span class="c-small c-muted">{t.declaredOn(new Date(c.created_at).toLocaleDateString(locale, { dateStyle: "medium" }))}</span>
						{/* A suspended Cahut serves nothing but to the administrators of the
						    server. See docs/adr/0048-suspending-a-cahut.md. */}
						{c.suspension && (
							<span class="c-small">
								<span class="c-badge c-badge--warning">{t.suspended}</span>{" "}
								{t.suspendedOn(new Date(c.suspension.since).toLocaleDateString(locale, { dateStyle: "medium" }), c.suspension.reason)}
							</span>
						)}
						{/* The reports that wait for its moderators. See
						    docs/adr/0054-reporting-and-withdrawing-a-window.md. */}
						{(c.reports ?? 0) > 0 && (
							<span class="c-small">
								<span class="c-badge c-badge--warning">{t.reportsWait(c.reports ?? 0)}</span>
							</span>
						)}
						<a class="c-link c-small" href={`${originOf(c.domain)}/`} aria-label={t.addressOf(c.domain)}>
							{t.address}
							<ArrowUpRight class="i" size={14} aria-hidden="true" />
						</a>
						<MoveCahut domain={c.domain} onMoved={reload} />
					</article>
				))}
			</div>
			<DeclareCahut onDeclared={reload} />
		</>
	);
}

/**
 * Moves a Cahut to another domain, or from the address of the server to a
 * domain. See docs/adr/0026-cahut-at-an-ip-address.md.
 */
function MoveCahut(props: { domain: string; onMoved: () => void }) {
	const api = useServer();
	const [open, setOpen] = useState(false);
	const [to, setTo] = useState("");
	const [error, setError] = useState("");
	const [busy, setBusy] = useState(false);
	const ip = ipAddress.test(props.domain);
	if (!open) {
		return (
			<div>
				<button
					class={`c-btn c-btn--sm ${ip ? "c-btn--primary" : "c-btn--ghost"}`}
					type="button"
					aria-label={ip ? t.moveToDomainOf(props.domain) : t.changeDomainOf(props.domain)}
					onClick={() => setOpen(true)}
				>
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
				location.assign(`${originOf(moved.domain)}/server/sign-in`);
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
function DeclareCahut({ onDeclared }: { onDeclared: () => void }) {
	const api = useServer();
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
						<CopyField label={t.activationLink} value={`${originOf(done.domain)}/console/activate#${done.admin.activation.code}`} />
					)}
				</div>
			)}
		</section>
	);
}
