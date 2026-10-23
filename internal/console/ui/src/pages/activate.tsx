// The page of an activation link, where a person, or an administrator of the
// server, chooses their password. The code comes in the fragment of the
// address, which the server never sees.
import { useState } from "preact/hooks";
import { cahut, explain, server } from "../api";
import { fragment as code } from "../fragment";
import { Logo } from "../logo";
import type { Session } from "../session";

export const minPassword = 15;

export function Activate({ kind, onSignIn }: { kind: Session["kind"]; onSignIn: (s: Session) => void }) {
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
			if (kind === "admin") {
				const s = await server("activate_account", { code, password });
				onSignIn({ kind, token: s.token, handle: s.admin.handle, name: s.admin.name, expiresAt: s.expires_at });
			} else {
				const s = await cahut("activate_account", { code, password });
				onSignIn({ kind, token: s.token, handle: s.user.handle, name: s.user.name, expiresAt: s.expires_at });
			}
		} catch (err) {
			setError(explain(err));
			setBusy(false);
		}
	};

	return (
		<main class="auth">
			<div class="c-card auth__card">
				<Logo height={32} />
				<div class="auth__head">
					<h1 class="c-h3">Choose your password</h1>
					<p class="c-small c-muted">
						{kind === "admin"
							? "You administer the server. Choose the password you will sign in with."
							: `Choose the password you will sign in to ${location.hostname} with.`}
					</p>
				</div>
				{code ? (
					<form onSubmit={submit}>
						<label class="c-field">
							<span class="c-field__label">New password</span>
							<input
								class="c-input"
								type="password"
								name="new-password"
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
								name="new-password-again"
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
						<button class="c-btn c-btn--primary" type="submit" disabled={busy}>
							Save and sign in
						</button>
					</form>
				) : (
					<div class="c-alert c-alert--warning" role="alert">
						This link has no code. Open the whole link you received, or ask an administrator for a new one.
					</div>
				)}
			</div>
		</main>
	);
}
