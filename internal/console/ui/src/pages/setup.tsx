// The first start of a Cahut, from the single-use link that the
// installation prints: the first person chooses their account, the house
// style of the organization, and connects their first AI. See
// docs/adr/0025-first-start-in-the-browser.md.
import { ArrowRight, Upload } from "lucide-preact";
import { useEffect, useState } from "preact/hooks";
import { cahut, explain, server } from "../api";
import type { Setup as SetupInfo } from "../api.gen";
import { Languages } from "../components";
import { ConnectedStatus, ConnectSteps, ipAddress, useConnectedAI } from "../connect";
import { fragment as code } from "../fragment";
import { de, words } from "../i18n";
import { Logo } from "../logo";
import { Link, navigate } from "../router";
import type { Session } from "../session";
import { minPassword } from "./activate";
import { base64 } from "./house-style";

const t = words({
	en: {
		steps: ["You", "Your Cahut", "Your first AI"],
		noCode: "This link has no code. Open the whole link that the installation printed.",
		gone: "This link does not open the first start",
		newLink: "On the server, this command gives a new link, valid for 24 hours, as long as the Cahut has nobody:",
		signInInstead: (
			<>
				If the first start is done,{" "}
				<Link class="c-link" href="/sign-in">
					sign in
				</Link>{" "}
				instead.
			</>
		),
		differ: "The two passwords differ.",
		welcome: "Welcome to your Cahut",
		youServer: (domain: string) =>
			`You will administer the Cahut of ${domain}, and this server, with the handle and the password you choose here.`,
		you: (domain: string) => `You will administer the Cahut of ${domain}, with the handle and the password you choose here.`,
		name: "Your name",
		handle: "Handle",
		handleHelp: "You sign in with it, and it names your personal Room.",
		password: "Password",
		passwordHelp: (min: number) => `At least ${min} characters. A short sentence works well.`,
		again: "The same password again",
		create: "Create my account",
		yourCahut: "Your Cahut",
		answers: (origin: string) => (
			<>
				It answers at <strong>{origin}/</strong>, where each Room and each Window has its own address.
			</>
		),
		noDomain: (
			<>
				This Cahut has no domain yet. When you have one, point it to this server, then move the Cahut there, in the console of the
				administrator of the server or with <code class="c-mono">cahut domain rename</code>. The links shared before then stop working.
			</>
		),
		houseStyle:
			"The house style gives every Window that AIs make the look of your organization. Change it later, fonts included, in House style.",
		organization: "Name of the organization",
		color: "Main color",
		logo: "Logo",
		replace: "Replace",
		upload: "Upload",
		logoHelp: "SVG, PNG or WebP, 512 KB at most.",
		save: "Save and continue",
		neutral: "Keep the neutral style",
		firstAI: "Your first AI",
		connects: "Your AI connects to this Cahut over MCP. You choose the Rooms it reaches, and everything it does goes to the Logbook.",
		open: "Open the console",
		skip: "Skip, open the console",
	},
	fr: {
		steps: ["Vous", "Votre Cahut", "Votre première IA"],
		noCode: "Ce lien n'a pas de code. Ouvrez le lien entier qu'a affiché l'installation.",
		gone: "Ce lien n'ouvre pas le premier démarrage",
		newLink: "Sur le serveur, cette commande donne un nouveau lien, valable 24 heures, tant que la Cahut n'a personne :",
		signInInstead: (
			<>
				Si le premier démarrage est fait,{" "}
				<Link class="c-link" href="/sign-in">
					connectez-vous
				</Link>{" "}
				plutôt.
			</>
		),
		differ: "Les deux mots de passe diffèrent.",
		welcome: "Bienvenue dans votre Cahut",
		youServer: (domain) =>
			`Vous administrerez la Cahut ${de(domain)}, et ce serveur, avec l'identifiant et le mot de passe que vous choisissez ici.`,
		you: (domain) => `Vous administrerez la Cahut ${de(domain)}, avec l'identifiant et le mot de passe que vous choisissez ici.`,
		name: "Votre nom",
		handle: "Identifiant",
		handleHelp: "Vous vous connectez avec, et il donne son nom à votre Room personnelle.",
		password: "Mot de passe",
		passwordHelp: (min) => `Au moins ${min} caractères. Une courte phrase convient bien.`,
		again: "Le même mot de passe, une seconde fois",
		create: "Créer mon compte",
		yourCahut: "Votre Cahut",
		answers: (origin) => (
			<>
				Elle répond à <strong>{origin}/</strong>, où chaque Room et chaque Window a sa propre adresse.
			</>
		),
		noDomain: (
			<>
				Cette Cahut n'a pas encore de domaine. Quand vous en aurez un, pointez-le vers ce serveur, puis déplacez-y la Cahut, dans la console
				de l'administrateur du serveur ou avec <code class="c-mono">cahut domain rename</code>. Les liens partagés d'ici là cessent
				alors de fonctionner.
			</>
		),
		houseStyle:
			"La house style donne à chaque Window que font les IA l'apparence de votre organisation. Vous la modifierez plus tard, polices comprises, dans House style.",
		organization: "Nom de l'organisation",
		color: "Couleur principale",
		logo: "Logo",
		replace: "Remplacer",
		upload: "Importer",
		logoHelp: "SVG, PNG ou WebP, 512 ko au plus.",
		save: "Enregistrer et continuer",
		neutral: "Garder le style neutre",
		firstAI: "Votre première IA",
		connects: "Votre IA se connecte à cette Cahut par MCP. Vous choisissez les Rooms qu'elle atteint, et tout ce qu'elle fait va au Logbook.",
		open: "Ouvrir la console",
		skip: "Passer, ouvrir la console",
	},
});

const steps = t.steps;

export function Setup({ onSession }: { onSession: (s: Session) => void }) {
	const [info, setInfo] = useState<SetupInfo | null>(null);
	const [error, setError] = useState(code ? "" : t.noCode);
	const [step, setStep] = useState(0);
	const [session, setSession] = useState<Session | null>(null);

	useEffect(() => {
		if (code) server("get_setup", { code }).then(setInfo, (err) => setError(explain(err)));
	}, []);

	const signedIn = (s: Session) => {
		setSession(s);
		onSession(s);
		setStep(1);
	};
	return (
		<main class="setup">
			<div class="setup__frame">
				<header class="setup__head">
					<Logo height={30} />
					<ol class="setup__steps">
						{steps.map((name, i) => (
							<li key={name} aria-current={i === step ? "step" : undefined} data-done={i < step ? "" : undefined}>
								<span class="setup__number">{i + 1}</span>
								{name}
							</li>
						))}
					</ol>
				</header>
				<div class="c-card setup__card">
					{error ? (
						<Gone message={error} />
					) : step === 0 ? (
						info && <You info={info} onDone={signedIn} />
					) : step === 1 && session ? (
						<YourCahut session={session} onDone={() => setStep(2)} />
					) : (
						session && <FirstAI session={session} />
					)}
				</div>
				{step === 0 && <Languages />}
			</div>
		</main>
	);
}

function Gone({ message }: { message: string }) {
	return (
		<div class="setup__body">
			<h1 class="c-h3">{t.gone}</h1>
			<div class="c-alert c-alert--warning" role="alert">
				{message}
			</div>
			<p class="c-small">{t.newLink}</p>
			<pre class="c-mono setup__command">sudo -u cahut /usr/local/bin/cahut setup --domain {location.hostname}</pre>
			<p class="c-small">{t.signInInstead}</p>
		</div>
	);
}

/** A handle made from a name: its first word, in the letters a handle takes. */
function handleOf(name: string): string {
	const first = (name.trim().split(/\s+/)[0] ?? "").normalize("NFD").replace(/[̀-ͯ]/g, "");
	return first
		.toLowerCase()
		.replace(/[^a-z0-9-]/g, "")
		.replace(/^[^a-z]+/, "")
		.slice(0, 32);
}

function You({ info, onDone }: { info: SetupInfo; onDone: (s: Session) => void }) {
	const [name, setName] = useState("");
	const [handle, setHandle] = useState("");
	const [edited, setEdited] = useState(false);
	const [password, setPassword] = useState("");
	const [again, setAgain] = useState("");
	const [error, setError] = useState("");
	const [busy, setBusy] = useState(false);

	const submit = async (e: Event) => {
		e.preventDefault();
		if (password !== again) {
			setError(t.differ);
			return;
		}
		setBusy(true);
		setError("");
		try {
			const r = await server("set_up", { code, handle, name, password });
			const s = r.session;
			onDone({ kind: "person", token: s.token, handle: s.user.handle, name: s.user.name, expiresAt: s.expires_at });
		} catch (err) {
			setError(explain(err));
			setBusy(false);
		}
	};
	return (
		<form class="setup__body" onSubmit={submit}>
			<div class="setup__intro">
				<h1 class="c-h3">{t.welcome}</h1>
				<p class="c-small c-muted">{info.server_admin ? t.youServer(info.domain) : t.you(info.domain)}</p>
			</div>
			<div class="form__row">
				<label class="c-field">
					<span class="c-field__label">{t.name}</span>
					<input
						class="c-input"
						required
						autocomplete="name"
						value={name}
						onInput={(e) => {
							setName(e.currentTarget.value);
							if (!edited) setHandle(handleOf(e.currentTarget.value));
						}}
					/>
				</label>
				<label class="c-field">
					<span class="c-field__label">{t.handle}</span>
					<input
						class="c-input c-mono"
						required
						autocapitalize="none"
						autocomplete="username"
						value={handle}
						onInput={(e) => {
							setEdited(true);
							setHandle(e.currentTarget.value);
						}}
					/>
					<span class="c-field__help">{t.handleHelp}</span>
				</label>
			</div>
			<label class="c-field">
				<span class="c-field__label">{t.password}</span>
				<input
					class="c-input"
					type="password"
					autocomplete="new-password"
					minLength={minPassword}
					required
					value={password}
					onInput={(e) => setPassword(e.currentTarget.value)}
				/>
				<span class="c-field__help">{t.passwordHelp(minPassword)}</span>
			</label>
			<label class="c-field">
				<span class="c-field__label">{t.again}</span>
				<input
					class="c-input"
					type="password"
					autocomplete="new-password"
					minLength={minPassword}
					required
					value={again}
					onInput={(e) => setAgain(e.currentTarget.value)}
				/>
			</label>
			{error && (
				<div class="c-alert c-alert--danger" role="alert">
					{error}
				</div>
			)}
			<div class="setup__actions">
				<button class="c-btn c-btn--primary" type="submit" disabled={busy}>
					{t.create}
					<ArrowRight class="i" size={16} aria-hidden="true" />
				</button>
			</div>
		</form>
	);
}

function YourCahut({ session, onDone }: { session: Session; onDone: () => void }) {
	const [name, setName] = useState("");
	const [primary, setPrimary] = useState("#2563eb");
	const [logo, setLogo] = useState<{ content: string; preview: string } | null>(null);
	const [error, setError] = useState("");
	const [busy, setBusy] = useState(false);
	useEffect(() => {
		cahut("get_house_style", {}, session.token).then(
			(h) => setPrimary(h.colors.primary),
			() => undefined,
		);
	}, []);

	const pick = async (e: Event) => {
		const file = (e.currentTarget as HTMLInputElement).files?.[0];
		if (!file) return;
		const content = await base64(await file.arrayBuffer());
		setLogo({ content, preview: `data:${file.type || "image/svg+xml"};base64,${content}` });
	};
	const save = async (e: Event) => {
		e.preventDefault();
		setBusy(true);
		setError("");
		try {
			await cahut(
				"set_house_style",
				{
					name: name.trim(),
					colors: { primary },
					...(logo ? { logo: { content: logo.content, encoding: "base64" } } : {}),
				},
				session.token,
			);
			onDone();
		} catch (err) {
			setError(explain(err));
			setBusy(false);
		}
	};
	const ip = ipAddress.test(location.hostname);
	return (
		<form class="setup__body" onSubmit={save}>
			<div class="setup__intro">
				<h1 class="c-h3">{t.yourCahut}</h1>
				<p class="c-small c-muted">{t.answers(location.origin)}</p>
			</div>
			{ip && (
				<div class="c-alert c-alert--info">{t.noDomain}</div>
			)}
			<p class="c-small">{t.houseStyle}</p>
			<label class="c-field">
				<span class="c-field__label">{t.organization}</span>
				<input class="c-input" placeholder={location.hostname} value={name} onInput={(e) => setName(e.currentTarget.value)} />
			</label>
			<div class="form__row">
				<label class="c-field">
					<span class="c-field__label">{t.color}</span>
					<span class="setup__color">
						<input type="color" value={primary} onInput={(e) => setPrimary(e.currentTarget.value)} />
						<code class="c-mono">{primary}</code>
					</span>
				</label>
				<div class="c-field">
					<span class="c-field__label">{t.logo}</span>
					<span class="setup__logo">
						{logo && <img src={logo.preview} alt="" />}
						<label class="c-btn c-btn--secondary c-btn--sm">
							<Upload class="i" size={14} aria-hidden="true" />
							{logo ? t.replace : t.upload}
							<input class="house__file" type="file" accept=".svg,.png,.webp,image/svg+xml,image/png,image/webp" onChange={pick} />
						</label>
					</span>
					<span class="c-field__help">{t.logoHelp}</span>
				</div>
			</div>
			{error && (
				<div class="c-alert c-alert--danger" role="alert">
					{error}
				</div>
			)}
			<div class="setup__actions">
				<button class="c-btn c-btn--primary" type="submit" disabled={busy}>
					{t.save}
					<ArrowRight class="i" size={16} aria-hidden="true" />
				</button>
				<button class="c-btn c-btn--ghost" type="button" onClick={onDone}>
					{t.neutral}
				</button>
			</div>
		</form>
	);
}

function FirstAI({ session }: { session: Session }) {
	const mcp = `${location.origin}/mcp`;
	const connected = useConnectedAI(async () => {
		const [ais, tokens] = await Promise.all([
			cahut("list_connected_ais", { user: session.handle }, session.token),
			cahut("list_tokens", { user: session.handle }, session.token),
		]);
		return [ais.connections, tokens.tokens];
	});
	return (
		<div class="setup__body">
			<div class="setup__intro">
				<h1 class="c-h3">{t.firstAI}</h1>
				<p class="c-small c-muted">{t.connects}</p>
			</div>
			<ConnectSteps mcp={mcp} />
			<ConnectedStatus name={connected} />
			<div class="setup__actions">
				<button class="c-btn c-btn--primary" type="button" onClick={() => navigate("/", true)}>
					{connected ? t.open : t.skip}
					<ArrowRight class="i" size={16} aria-hidden="true" />
				</button>
			</div>
		</div>
	);
}
