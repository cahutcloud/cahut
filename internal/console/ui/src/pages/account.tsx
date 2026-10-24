// The account: who is signed in, their e-mail address and notices, a new
// password, and signing out.
import { useState } from "preact/hooks";
import { explain } from "../api";
import type { EmailChange } from "../api.gen";
import { Failure, PageHead } from "../components";
import { useConsole } from "../context";
import { useLoad } from "../load";

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
			{api.session.kind === "person" && api.sendsMail && <EmailSection />}
			<section class="section c-card panel narrow">
				<h2 class="c-h4">Change your password</h2>
				<p class="c-small c-muted">
					{api.session.kind === "admin"
						? "This is your password as an administrator of the server. Your account as a person of a Cahut, if you have one, keeps its own."
						: "This is your password in this Cahut. Your account as an administrator of the server, if you have one, keeps its own."}
				</p>
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

/** What the account says of its e-mail address and notices. */
interface Mailbox {
	email?: string;
	change?: EmailChange;
	noticesOff?: boolean;
}

/**
 * The e-mail address of a person, where links and notices go, in a Cahut
 * that sends e-mail: a new one takes effect once the link sent to it is
 * used. The person also turns off the notices that are neither about their
 * account nor about a Knock. See docs/adr/0031-e-mail-through-a-relay.md and
 * docs/adr/0034-e-mail-through-the-relay-of-each-cahut-only.md.
 */
function EmailSection() {
	const api = useConsole();
	const { data, reload } = useLoad(async (): Promise<Mailbox> => {
		const a = await api.cahut("get_account", {});
		return { email: a.email, change: a.email_change, noticesOff: a.notices_off };
	});
	const [email, setEmail] = useState("");
	const [said, setSaid] = useState("");
	const [error, setError] = useState("");
	if (!data) return null;
	const act = async (run: () => Promise<string>) => {
		setError("");
		setSaid("");
		try {
			setSaid(await run());
			reload();
		} catch (err) {
			setError(explain(err));
		}
	};
	const change = (e: Event) => {
		e.preventDefault();
		act(async () => {
			const input = { email };
			const r = await api.cahut("change_email", input);
			setEmail("");
			return r.waiting ? `A link went to ${r.waiting.email}. Your address changes once you use it.` : "Your address is removed.";
		});
	};
	return (
		<section class="section c-card panel narrow">
			<h2 class="c-h4">E-mail</h2>
			<p class="c-small">
				{data.email ? (
					<>
						Links and notices go to <span class="c-mono">{data.email}</span>.
					</>
				) : (
					"Your account has no e-mail address: links come from an administrator, and no notice reaches you."
				)}
			</p>
			{data.change && (
				<p class="c-small c-muted">
					<span class="c-mono">{data.change.email}</span> waits for you to use the link sent to it, until{" "}
					{new Date(data.change.expires_at).toLocaleString("en-GB", { dateStyle: "long", timeStyle: "short" })}.
				</p>
			)}
			<form class="form" onSubmit={change}>
				<label class="c-field">
					<span class="c-field__label">{data.email ? "New address" : "Address"}</span>
					<input class="c-input" type="email" autocomplete="email" required value={email} onInput={(e) => setEmail(e.currentTarget.value)} />
					<span class="c-field__help">A link goes to it: your address changes once you use it, within 24 hours.</span>
				</label>
				<div class="actions">
					<button class="c-btn c-btn--primary" type="submit">
						Send the link
					</button>
					{data.email && (
						<button
							class="c-btn c-btn--secondary"
							type="button"
							onClick={() =>
								act(async () => {
									await api.cahut("change_email", { email: "" });
									return "Your address is removed.";
								})
							}
						>
							Remove your address
						</button>
					)}
				</div>
			</form>
			<label class="c-check">
				<input
					type="checkbox"
					checked={!data.noticesOff}
					onChange={(e) => {
						const off = !e.currentTarget.checked;
						act(async () => {
							await api.cahut("set_mail_notices", { off });
							return off
								? "Notices are off, but for those about your account and your Knocks."
								: "Notices are on.";
						});
					}}
				/>
				Tell me by e-mail when I join a Room, or when the access of a Window that I set ends
			</label>
			{said && (
				<div class="c-alert c-alert--success" role="status">
					{said}
				</div>
			)}
			{error && <Failure message={error} />}
		</section>
	);
}
