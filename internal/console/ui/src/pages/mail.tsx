// The e-mail relay of this Cahut, for its administrators: the SMTP relay
// that sends activation links, invitations and notices, a test, and how the
// last message went. Without a relay, the Cahut sends no e-mail. See
// docs/adr/0034-e-mail-through-the-relay-of-each-cahut-only.md.
import { Send } from "lucide-preact";
import { useState } from "preact/hooks";
import { explain } from "../api";
import type { Delivery, SetMailRelayInput } from "../api.gen";
import { agoInText, ConfirmButton, Failure, Notice, PageHead } from "../components";
import { useConsole } from "../context";
import { words } from "../i18n";
import { useLoad } from "../load";

const t = words({
	en: {
		onlyAdmins: "Only the administrators of this Cahut set its e-mail.",
		testSent: (to: string) => `A test went to ${to}. Look in that mailbox.`,
		label: "This Cahut",
		title: "E-mail",
		lead: "The SMTP relay that sends the e-mail of this Cahut, from an address of its own: activation links, invitations and notices. Without one, this Cahut sends no e-mail, and the console offers nothing that would.",
		howTitle: "How e-mail goes out",
		lastSent: (ago: string) => `A message last went out ${ago}.`,
		lastError: (ago: string | undefined, error: string) => `The last message that failed, ${ago ?? "lately"}: ${error}`,
		testTo: "Send a test to",
		testYours: "Send a test to your address",
		test: "Send a test",
		through: (host: string, port: number, security: string, from: string) => `Through ${host}:${port}, with ${security}, from ${from}.`,
		development:
			"This Cahut has no relay. The server runs for development: it writes the messages of this Cahut to disk, in the mail folder of its data.",
		none: "This Cahut has no relay: it sends no e-mail. People get their links from an administrator, and nobody can invite by e-mail or ask for a new password.",
		changeRelay: "Change the relay",
		setRelay: "Set a relay for this Cahut",
		provider:
			"Your provider of e-mail gives these settings. The relay is on the internet, and speaks TLS. Cahut connects and signs in before it keeps it.",
		host: "Host",
		port: "Port",
		security: "Security",
		starttls: "STARTTLS, usually on port 587",
		tls: "TLS from the start, usually on port 465",
		username: "User name",
		usernameHelp: "Leave it empty when the relay takes no authentication.",
		password: "Password",
		kept: "Kept",
		keepHelp: "Leave it empty to keep the password.",
		sealedHelp: "Cahut keeps it sealed, and never shows it again.",
		from: "From",
		fromHelp: "An address that the relay accepts.",
		name: "Name",
		namePlaceholder: "The name of the Cahut",
		connecting: "Connecting…",
		save: "Connect and save",
		remove: "Remove the relay",
		removeAsk: "Remove it? This Cahut stops sending e-mail, and the console no longer offers what would.",
		removeConfirm: "Remove",
	},
	fr: {
		onlyAdmins: "Seuls les administrateurs de ce Cahut règlent son e-mail.",
		testSent: (to) => `Un test est parti à ${to}. Regardez dans cette boîte.`,
		label: "Ce Cahut",
		title: "E-mail",
		lead: "Le relais SMTP qui envoie les e-mails de ce Cahut, depuis sa propre adresse : liens d'activation, invitations et notifications. Sans relais, ce Cahut n'envoie aucun e-mail, et la console n'offre rien qui en enverrait.",
		howTitle: "Comment partent les e-mails",
		lastSent: (ago) => `Le dernier message est parti ${ago}.`,
		lastError: (ago, error) => `Le dernier message en échec, ${ago ?? "récemment"} : ${error}`,
		testTo: "Envoyer un test à",
		testYours: "Envoyer un test à votre adresse",
		test: "Envoyer un test",
		through: (host, port, security, from) => `Par ${host}:${port}, en ${security}, depuis ${from}.`,
		development:
			"Ce Cahut n'a pas de relais. Le serveur tourne pour le développement : il écrit les messages de ce Cahut sur le disque, dans le dossier mail de ses données.",
		none: "Ce Cahut n'a pas de relais : il n'envoie aucun e-mail. Les personnes reçoivent leurs liens d'un administrateur, et personne ne peut inviter par e-mail ni demander un nouveau mot de passe.",
		changeRelay: "Modifier le relais",
		setRelay: "Régler un relais pour ce Cahut",
		provider:
			"Votre fournisseur d'e-mail donne ces réglages. Le relais est sur internet, et parle TLS. Cahut s'y connecte et s'y identifie avant de le garder.",
		host: "Hôte",
		port: "Port",
		security: "Sécurité",
		starttls: "STARTTLS, en général sur le port 587",
		tls: "TLS dès le départ, en général sur le port 465",
		username: "Nom d'utilisateur",
		usernameHelp: "Laissez-le vide quand le relais ne demande pas d'authentification.",
		password: "Mot de passe",
		kept: "Gardé",
		keepHelp: "Laissez-le vide pour garder le mot de passe.",
		sealedHelp: "Cahut le garde scellé, et ne l'affiche plus jamais.",
		from: "Expéditeur",
		fromHelp: "Une adresse que le relais accepte.",
		name: "Nom",
		namePlaceholder: "Le nom du Cahut",
		connecting: "Connexion…",
		save: "Se connecter et enregistrer",
		remove: "Retirer le relais",
		removeAsk: "Le retirer ? Ce Cahut cesse d'envoyer des e-mails, et la console n'offre plus ce qui en enverrait.",
		removeConfirm: "Retirer",
	},
});

export function MailPage() {
	const api = useConsole();
	const { data, error, reload } = useLoad(() => api.cahut("get_mail_relay", {}));
	const [failure, setFailure] = useState("");
	const [said, setSaid] = useState("");
	// An administrator of the server has no address: their test goes where
	// they say.
	const [to, setTo] = useState("");
	const own = api.session.kind === "person";
	if (!api.isAdmin) return <Notice>{t.onlyAdmins}</Notice>;
	if (error) return <Failure message={error} />;
	if (!data) return null;

	const act = async (run: () => Promise<string>) => {
		setFailure("");
		setSaid("");
		try {
			setSaid(await run());
			reload();
		} catch (err) {
			setFailure(explain(err));
		}
	};
	const test = (e: Event) => {
		e.preventDefault();
		act(async () => t.testSent((await api.cahut("send_test_mail", own ? {} : { to })).to));
	};
	const changed = () => (reload(), api.reloadCahut());
	return (
		<>
			<PageHead label={t.label} title={t.title}>
				{t.lead}
			</PageHead>
			<section class="section c-card panel">
				<h2 class="c-h4">{t.howTitle}</h2>
				<p class="c-small">{describe(data)}</p>
				{data.last_sent && <p class="c-small c-muted">{t.lastSent(agoInText(data.last_sent))}</p>}
				{data.last_error && (
					<div class="c-alert c-alert--warning" role="status">
						{t.lastError(data.last_error_at ? agoInText(data.last_error_at) : undefined, data.last_error)}
					</div>
				)}
				{data.relay && (
					<form class="form" onSubmit={test}>
						{!own && (
							<label class="c-field">
								<span class="c-field__label">{t.testTo}</span>
								<input class="c-input" type="email" required autocomplete="email" value={to} onInput={(e) => setTo(e.currentTarget.value)} />
							</label>
						)}
						<div>
							<button class="c-btn c-btn--secondary" type="submit">
								<Send class="i" size={16} aria-hidden="true" />
								{own ? t.testYours : t.test}
							</button>
						</div>
					</form>
				)}
				{said && (
					<div class="c-alert c-alert--success" role="status">
						{said}
					</div>
				)}
				{failure && <Failure message={failure} />}
			</section>
			<RelayForm
				key={data.relay?.set_at ?? "none"}
				delivery={data}
				onSave={async (input) => {
					await api.cahut("set_mail_relay", input);
					changed();
				}}
				onRemove={async () => {
					await api.cahut("set_mail_relay", { host: "" });
					changed();
				}}
			/>
		</>
	);
}

/** Says through which relay e-mail goes out, if any. */
function describe(d: Delivery): string {
	if (d.relay) {
		const from = d.relay.name ? `${d.relay.name} <${d.relay.from}>` : d.relay.from;
		return t.through(d.relay.host, d.relay.port, d.relay.security === "tls" ? "TLS" : "STARTTLS", from);
	}
	if (d.sends) return t.development;
	return t.none;
}

function RelayForm(props: { delivery: Delivery; onSave: (input: SetMailRelayInput) => Promise<void>; onRemove: () => Promise<void> }) {
	const r = props.delivery.relay;
	const [host, setHost] = useState(r?.host ?? "");
	const [port, setPort] = useState(r ? String(r.port) : "");
	const [security, setSecurity] = useState(r?.security ?? "starttls");
	const [username, setUsername] = useState(r?.username ?? "");
	const [password, setPassword] = useState("");
	const [from, setFrom] = useState(r?.from ?? "");
	const [name, setName] = useState(r?.name ?? "");
	const [busy, setBusy] = useState(false);
	const [error, setError] = useState("");
	const kept = r?.password && host === r.host && username === r.username;
	const submit = async (e: Event) => {
		e.preventDefault();
		setBusy(true);
		setError("");
		try {
			await props.onSave({ host, port: port ? Number(port) : undefined, security, username, password, from, name });
			setPassword("");
		} catch (err) {
			setError(explain(err));
		}
		setBusy(false);
	};
	const remove = async () => {
		setError("");
		try {
			await props.onRemove();
		} catch (err) {
			setError(explain(err));
		}
	};
	return (
		<section class="section c-card panel">
			<h2 class="c-h4">{r ? t.changeRelay : t.setRelay}</h2>
			<p class="c-small c-muted">{t.provider}</p>
			<form class="form" onSubmit={submit}>
				<div class="form__row">
					<label class="c-field">
						<span class="c-field__label">{t.host}</span>
						<input class="c-input c-mono" required placeholder="smtp.example.com" autocapitalize="none" value={host} onInput={(e) => setHost(e.currentTarget.value)} />
					</label>
					<label class="c-field">
						<span class="c-field__label">{t.port}</span>
						<input class="c-input c-mono" inputMode="numeric" placeholder={security === "tls" ? "465" : "587"} value={port} onInput={(e) => setPort(e.currentTarget.value)} />
					</label>
					<label class="c-field">
						<span class="c-field__label">{t.security}</span>
						<select class="c-input" value={security} onChange={(e) => setSecurity(e.currentTarget.value)}>
							<option value="starttls">{t.starttls}</option>
							<option value="tls">{t.tls}</option>
						</select>
					</label>
				</div>
				<div class="form__row">
					<label class="c-field">
						<span class="c-field__label">{t.username}</span>
						<input class="c-input" autocomplete="off" autocapitalize="none" value={username} onInput={(e) => setUsername(e.currentTarget.value)} />
						<span class="c-field__help">{t.usernameHelp}</span>
					</label>
					<label class="c-field">
						<span class="c-field__label">{t.password}</span>
						<input
							class="c-input"
							type="password"
							autocomplete="new-password"
							placeholder={kept ? t.kept : ""}
							value={password}
							onInput={(e) => setPassword(e.currentTarget.value)}
						/>
						<span class="c-field__help">
							{kept ? t.keepHelp : t.sealedHelp}
						</span>
					</label>
				</div>
				<div class="form__row">
					<label class="c-field">
						<span class="c-field__label">{t.from}</span>
						<input class="c-input c-mono" type="email" required placeholder="cahut@example.com" value={from} onInput={(e) => setFrom(e.currentTarget.value)} />
						<span class="c-field__help">{t.fromHelp}</span>
					</label>
					<label class="c-field">
						<span class="c-field__label">{t.name}</span>
						<input class="c-input" placeholder={t.namePlaceholder} value={name} onInput={(e) => setName(e.currentTarget.value)} />
					</label>
				</div>
				{error && <Failure message={error} />}
				<div class="actions">
					<button class="c-btn c-btn--primary" type="submit" disabled={busy}>
						{busy ? t.connecting : t.save}
					</button>
					{r && (
						<ConfirmButton
							label={t.remove}
							question={t.removeAsk}
							confirm={t.removeConfirm}
							danger
							onConfirm={remove}
						/>
					)}
				</div>
			</form>
		</section>
	);
}
