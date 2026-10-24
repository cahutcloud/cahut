// The first start of a Cahut, from the single-use link that the
// installation prints: the first person chooses their account, the house
// style of the organization, and connects their first AI. See
// docs/adr/0025-first-start-in-the-browser.md.
import { ArrowRight, Upload } from "lucide-preact";
import { useEffect, useState } from "preact/hooks";
import { cahut, explain, server } from "../api";
import type { Setup as SetupInfo } from "../api.gen";
import { ConnectedStatus, ConnectSteps, useConnectedAI } from "../connect";
import { fragment as code } from "../fragment";
import { Logo } from "../logo";
import { Link, navigate } from "../router";
import type { Session } from "../session";
import { minPassword } from "./activate";
import { base64 } from "./house-style";

const steps = ["You", "Your Cahut", "Your first AI"];

export function Setup({ onSession }: { onSession: (s: Session) => void }) {
	const [info, setInfo] = useState<SetupInfo | null>(null);
	const [error, setError] = useState(code ? "" : "This link has no code. Open the whole link that the installation printed.");
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
			</div>
		</main>
	);
}

function Gone({ message }: { message: string }) {
	return (
		<div class="setup__body">
			<h1 class="c-h3">This link does not open the first start</h1>
			<div class="c-alert c-alert--warning" role="alert">
				{message}
			</div>
			<p class="c-small">
				On the server, this command gives a new link, valid for 24 hours, as long as the Cahut has nobody:
			</p>
			<pre class="c-mono setup__command">sudo -u cahut /usr/local/bin/cahut setup --domain {location.hostname}</pre>
			<p class="c-small">
				If the first start is done,{" "}
				<Link class="c-link" href="/sign-in">
					sign in
				</Link>{" "}
				instead.
			</p>
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
			setError("The two passwords differ.");
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
				<h1 class="c-h3">Welcome to your Cahut</h1>
				<p class="c-small c-muted">
					{info.server_admin
						? `You will administer the Cahut of ${info.domain}, and this server, with the handle and the password you choose here.`
						: `You will administer the Cahut of ${info.domain}, with the handle and the password you choose here.`}
				</p>
			</div>
			<div class="form__row">
				<label class="c-field">
					<span class="c-field__label">Your name</span>
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
					<span class="c-field__label">Handle</span>
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
					<span class="c-field__help">You sign in with it, and it names your personal Room.</span>
				</label>
			</div>
			<label class="c-field">
				<span class="c-field__label">Password</span>
				<input
					class="c-input"
					type="password"
					autocomplete="new-password"
					minLength={minPassword}
					required
					value={password}
					onInput={(e) => setPassword(e.currentTarget.value)}
				/>
				<span class="c-field__help">At least {minPassword} characters. A short sentence works well.</span>
			</label>
			<label class="c-field">
				<span class="c-field__label">The same password again</span>
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
					Create my account
					<ArrowRight class="i" size={16} aria-hidden="true" />
				</button>
			</div>
		</form>
	);
}

/** An address that is an IPv4 address, not a domain. */
const ipAddress = /^\d{1,3}(\.\d{1,3}){3}$/;

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
				<h1 class="c-h3">Your Cahut</h1>
				<p class="c-small c-muted">
					It answers at <strong>{location.origin}/</strong>, where each Room and each Window has its own address.
				</p>
			</div>
			{ip && (
				<div class="c-alert c-alert--info">
					This Cahut has no domain yet. When you have one, point it to this server, then move the Cahut there, in the console of
					the administrator of the server or with <code class="c-mono">cahut domain rename</code>. The links shared before then stop
					working.
				</div>
			)}
			<p class="c-small">
				The house style gives every Window that AIs make the look of your organization. Change it later, fonts included, in House
				style.
			</p>
			<label class="c-field">
				<span class="c-field__label">Name of the organization</span>
				<input class="c-input" placeholder={location.hostname} value={name} onInput={(e) => setName(e.currentTarget.value)} />
			</label>
			<div class="form__row">
				<label class="c-field">
					<span class="c-field__label">Main color</span>
					<span class="setup__color">
						<input type="color" value={primary} onInput={(e) => setPrimary(e.currentTarget.value)} />
						<code class="c-mono">{primary}</code>
					</span>
				</label>
				<div class="c-field">
					<span class="c-field__label">Logo</span>
					<span class="setup__logo">
						{logo && <img src={logo.preview} alt="" />}
						<label class="c-btn c-btn--secondary c-btn--sm">
							<Upload class="i" size={14} aria-hidden="true" />
							{logo ? "Replace" : "Upload"}
							<input class="house__file" type="file" accept=".svg,.png,.webp,image/svg+xml,image/png,image/webp" onChange={pick} />
						</label>
					</span>
					<span class="c-field__help">SVG, PNG or WebP, 512 KB at most.</span>
				</div>
			</div>
			{error && (
				<div class="c-alert c-alert--danger" role="alert">
					{error}
				</div>
			)}
			<div class="setup__actions">
				<button class="c-btn c-btn--primary" type="submit" disabled={busy}>
					Save and continue
					<ArrowRight class="i" size={16} aria-hidden="true" />
				</button>
				<button class="c-btn c-btn--ghost" type="button" onClick={onDone}>
					Keep the neutral style
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
				<h1 class="c-h3">Your first AI</h1>
				<p class="c-small c-muted">
					Your AI connects to this Cahut over MCP. You choose the Rooms it reaches, and everything it does goes to the Logbook.
				</p>
			</div>
			<ConnectSteps mcp={mcp} />
			<ConnectedStatus name={connected} />
			<div class="setup__actions">
				<button class="c-btn c-btn--primary" type="button" onClick={() => navigate("/", true)}>
					{connected ? "Open the console" : "Skip, open the console"}
					<ArrowRight class="i" size={16} aria-hidden="true" />
				</button>
			</div>
		</div>
	);
}
