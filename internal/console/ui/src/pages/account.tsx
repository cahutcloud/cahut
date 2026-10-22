// The account: who is signed in, a new password, and signing out.
import { useState } from "preact/hooks";
import { explain } from "../api";
import { Failure, PageHead } from "../components";
import { useConsole } from "../context";

const minPassword = 15;

export function Account({ onSignOut }: { onSignOut: () => void }) {
	const api = useConsole();
	const [password, setPassword] = useState("");
	const [next, setNext] = useState("");
	const [again, setAgain] = useState("");
	const [error, setError] = useState("");
	const [done, setDone] = useState(false);
	const submit = async (e: Event) => {
		e.preventDefault();
		setError("");
		setDone(false);
		if (next !== again) {
			setError("The two new passwords differ.");
			return;
		}
		try {
			const input = { password, new_password: next };
			await (api.session.kind === "admin" ? api.server("change_password", input) : api.cahut("change_password", input));
			setPassword("");
			setNext("");
			setAgain("");
			setDone(true);
		} catch (err) {
			setError(explain(err));
		}
	};
	return (
		<>
			<PageHead label="Account" title={api.session.name}>
				<span class="c-mono">{api.session.handle}</span>
				{api.session.kind === "admin" ? ", administrator of the server." : `, at ${location.hostname}.`}
			</PageHead>
			<section class="section c-card panel narrow">
				<h2 class="c-h4">Change your password</h2>
				<form class="form" onSubmit={submit}>
					<label class="c-field">
						<span class="c-field__label">Current password</span>
						<input class="c-input" type="password" autocomplete="current-password" required value={password} onInput={(e) => setPassword(e.currentTarget.value)} />
					</label>
					<label class="c-field">
						<span class="c-field__label">New password</span>
						<input class="c-input" type="password" autocomplete="new-password" minLength={minPassword} required value={next} onInput={(e) => setNext(e.currentTarget.value)} />
						<span class="c-field__help">At least {minPassword} characters. Your other sessions end.</span>
					</label>
					<label class="c-field">
						<span class="c-field__label">The same new password again</span>
						<input class="c-input" type="password" autocomplete="new-password" minLength={minPassword} required value={again} onInput={(e) => setAgain(e.currentTarget.value)} />
					</label>
					{error && <Failure message={error} />}
					{done && (
						<div class="c-alert c-alert--success" role="status">
							Password changed. Your other sessions have ended.
						</div>
					)}
					<div>
						<button class="c-btn c-btn--primary" type="submit">
							Change the password
						</button>
					</div>
				</form>
			</section>
			<button class="c-btn c-btn--secondary" type="button" onClick={onSignOut}>
				Sign out
			</button>
		</>
	);
}
