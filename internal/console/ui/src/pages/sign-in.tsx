// Sign-in, for the people of this Cahut and for the administrators of the
// server, who sign in at the console of any Cahut.
import { useEffect, useState } from "preact/hooks";
import { cahut, explain, server } from "../api";
import { Logo } from "../logo";
import { Link } from "../router";
import type { Session } from "../session";

export function SignIn({ kind, purpose, onSignIn }: { kind: Session["kind"]; purpose?: string; onSignIn: (s: Session) => void }) {
	const [forgot, setForgot] = useState(false);
	// A person who forgot their password gets a link by e-mail only where
	// this Cahut sends e-mail. See
	// docs/adr/0034-e-mail-through-the-relay-of-each-cahut-only.md.
	const [sendsMail, setSendsMail] = useState(false);
	useEffect(() => {
		if (kind === "person") {
			cahut("describe_sign_in", {}).then(
				(p) => setSendsMail(p.sends_mail),
				() => undefined,
			);
		}
	}, [kind]);
	const [handle, setHandle] = useState("");
	const [password, setPassword] = useState("");
	const [error, setError] = useState("");
	const [busy, setBusy] = useState(false);

	const submit = async (e: Event) => {
		e.preventDefault();
		setBusy(true);
		setError("");
		try {
			if (kind === "admin") {
				const s = await server("sign_in", { handle, password });
				onSignIn({ kind, token: s.token, handle: s.admin.handle, name: s.admin.name, expiresAt: s.expires_at });
			} else {
				const s = await cahut("sign_in", { handle, password });
				onSignIn({ kind, token: s.token, handle: s.user.handle, name: s.user.name, expiresAt: s.expires_at });
			}
		} catch (err) {
			setError(explain(err));
			setBusy(false);
		}
	};

	if (forgot) return <Forgot who={handle} onBack={() => setForgot(false)} />;
	return (
		<main class="auth">
			<div class="c-card auth__card">
				<Logo height={32} />
				<div class="auth__head">
					<h1 class="c-h3">{kind === "admin" ? "Administer this server" : "Sign in"}</h1>
					<p class="c-small c-muted">
						{kind === "admin"
							? `Sign in as an administrator of the server, at ${location.hostname}.`
							: `Sign in to ${location.hostname}${purpose ? ` ${purpose}` : ""}.`}
					</p>
				</div>
				<form onSubmit={submit}>
					<label class="c-field">
						<span class="c-field__label">Handle</span>
						<input
							class="c-input"
							name="username"
							autocomplete="username"
							autocapitalize="none"
							spellcheck={false}
							required
							value={handle}
							onInput={(e) => setHandle(e.currentTarget.value)}
						/>
					</label>
					<label class="c-field">
						<span class="c-field__label">Password</span>
						<input
							class="c-input"
							type="password"
							name="password"
							autocomplete="current-password"
							required
							value={password}
							onInput={(e) => setPassword(e.currentTarget.value)}
						/>
					</label>
					{error && (
						<div class="c-alert c-alert--danger" role="alert">
							{error}
						</div>
					)}
					<button class="c-btn c-btn--primary" type="submit" disabled={busy}>
						Sign in
					</button>
				</form>
				{kind === "admin" ? (
					<div class="auth__foot">
						<p class="c-small c-muted">
							Forgot your password? Another administrator of the server, or <span class="c-mono">cahut admin link</span> on the server,
							gives you a new link.
						</p>
						<p class="c-small">
							<Link class="c-link" href="/sign-in">
								Sign in as a person of this Cahut
							</Link>
						</p>
					</div>
				) : (
					<div class="auth__foot">
						{sendsMail ? (
							<p class="c-small">
								<button class="c-link link-button" type="button" onClick={() => setForgot(true)}>
									Forgot your password?
								</button>
							</p>
						) : (
							<p class="c-small c-muted">Forgot your password? An administrator of the Cahut can give you a new link.</p>
						)}
						<p class="c-small">
							<Link class="c-link" href="/server/sign-in">
								Administer the server
							</Link>
						</p>
					</div>
				)}
			</div>
		</main>
	);
}

/**
 * Asks for a link where a person chooses a new password. It goes to the
 * address of their account; the answer is the same whether the account
 * exists or not. See docs/adr/0031-e-mail-through-a-relay.md.
 */
function Forgot({ who: given, onBack }: { who: string; onBack: () => void }) {
	const [who, setWho] = useState(given);
	const [sent, setSent] = useState("");
	const [error, setError] = useState("");
	const [busy, setBusy] = useState(false);
	const submit = async (e: Event) => {
		e.preventDefault();
		setBusy(true);
		setError("");
		try {
			const r = await cahut("request_password_link", { who });
			setSent(r.message);
		} catch (err) {
			setError(explain(err));
		}
		setBusy(false);
	};
	return (
		<main class="auth">
			<div class="c-card auth__card">
				<Logo height={32} />
				<div class="auth__head">
					<h1 class="c-h3">Choose a new password</h1>
					<p class="c-small c-muted">
						Give your handle, or the e-mail address of your account. A link to choose a new password goes to that address. If your
						account has none, an administrator of the Cahut can give you a link.
					</p>
				</div>
				{sent ? (
					<div class="c-alert c-alert--success" role="status">
						{sent}
					</div>
				) : (
					<form onSubmit={submit}>
						<label class="c-field">
							<span class="c-field__label">Handle or e-mail address</span>
							<input
								class="c-input"
								name="username"
								autocomplete="username"
								autocapitalize="none"
								spellcheck={false}
								required
								value={who}
								onInput={(e) => setWho(e.currentTarget.value)}
							/>
						</label>
						{error && (
							<div class="c-alert c-alert--danger" role="alert">
								{error}
							</div>
						)}
						<button class="c-btn c-btn--primary" type="submit" disabled={busy}>
							Send me a link
						</button>
					</form>
				)}
				<p class="c-small">
					<button class="c-link link-button" type="button" onClick={onBack}>
						Back to sign in
					</button>
				</p>
			</div>
		</main>
	);
}
