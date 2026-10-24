// The e-mail relay of this Cahut, for its administrators: the SMTP relay
// that sends activation links, invitations and notices, a test, and how the
// last message went. Without a relay, the Cahut sends no e-mail. See
// docs/adr/0034-e-mail-through-the-relay-of-each-cahut-only.md.
import { Send } from "lucide-preact";
import { useState } from "preact/hooks";
import { explain } from "../api";
import type { Delivery, SetMailRelayInput } from "../api.gen";
import { ago, ConfirmButton, Failure, Notice, PageHead } from "../components";
import { useConsole } from "../context";
import { useLoad } from "../load";

export function MailPage() {
	const api = useConsole();
	const { data, error, reload } = useLoad(() => api.cahut("get_mail_relay", {}));
	const [failure, setFailure] = useState("");
	const [said, setSaid] = useState("");
	// An administrator of the server has no address: their test goes where
	// they say.
	const [to, setTo] = useState("");
	const own = api.session.kind === "person";
	if (!api.isAdmin) return <Notice>Only the administrators of this Cahut set its e-mail.</Notice>;
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
		act(async () => `A test went to ${(await api.cahut("send_test_mail", own ? {} : { to })).to}. Look in that mailbox.`);
	};
	const changed = () => (reload(), api.reloadCahut());
	return (
		<>
			<PageHead label="This Cahut" title="E-mail">
				The SMTP relay that sends the e-mail of this Cahut, from an address of its own: activation links, invitations and notices. Without
				one, this Cahut sends no e-mail, and the console offers nothing that would.
			</PageHead>
			<section class="section c-card panel">
				<h2 class="c-h4">How e-mail goes out</h2>
				<p class="c-small">{describe(data)}</p>
				{data.last_sent && <p class="c-small c-muted">A message last went out {ago(data.last_sent).toLowerCase()}.</p>}
				{data.last_error && (
					<div class="c-alert c-alert--warning" role="status">
						The last message that failed, {data.last_error_at ? ago(data.last_error_at).toLowerCase() : "lately"}: {data.last_error}
					</div>
				)}
				{data.relay && (
					<form class="form" onSubmit={test}>
						{!own && (
							<label class="c-field">
								<span class="c-field__label">Send a test to</span>
								<input class="c-input" type="email" required autocomplete="email" value={to} onInput={(e) => setTo(e.currentTarget.value)} />
							</label>
						)}
						<div>
							<button class="c-btn c-btn--secondary" type="submit">
								<Send class="i" size={16} aria-hidden="true" />
								{own ? "Send a test to your address" : "Send a test"}
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
		return `Through ${d.relay.host}:${d.relay.port}, with ${d.relay.security === "tls" ? "TLS" : "STARTTLS"}, from ${from}.`;
	}
	if (d.sends) return "This Cahut has no relay. The server runs for development: it writes the messages of this Cahut to disk, in the mail folder of its data.";
	return "This Cahut has no relay: it sends no e-mail. People get their links from an administrator, and nobody can invite by e-mail or ask for a new password.";
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
			<h2 class="c-h4">{r ? "Change the relay" : "Set a relay for this Cahut"}</h2>
			<p class="c-small c-muted">
				Your provider of e-mail gives these settings. The relay is on the internet, and speaks TLS. Cahut connects and signs in before it
				keeps it.
			</p>
			<form class="form" onSubmit={submit}>
				<div class="form__row">
					<label class="c-field">
						<span class="c-field__label">Host</span>
						<input class="c-input c-mono" required placeholder="smtp.example.com" autocapitalize="none" value={host} onInput={(e) => setHost(e.currentTarget.value)} />
					</label>
					<label class="c-field">
						<span class="c-field__label">Port</span>
						<input class="c-input c-mono" inputMode="numeric" placeholder={security === "tls" ? "465" : "587"} value={port} onInput={(e) => setPort(e.currentTarget.value)} />
					</label>
					<label class="c-field">
						<span class="c-field__label">Security</span>
						<select class="c-input" value={security} onChange={(e) => setSecurity(e.currentTarget.value)}>
							<option value="starttls">STARTTLS, usually on port 587</option>
							<option value="tls">TLS from the start, usually on port 465</option>
						</select>
					</label>
				</div>
				<div class="form__row">
					<label class="c-field">
						<span class="c-field__label">User name</span>
						<input class="c-input" autocomplete="off" autocapitalize="none" value={username} onInput={(e) => setUsername(e.currentTarget.value)} />
						<span class="c-field__help">Leave it empty when the relay takes no authentication.</span>
					</label>
					<label class="c-field">
						<span class="c-field__label">Password</span>
						<input
							class="c-input"
							type="password"
							autocomplete="new-password"
							placeholder={kept ? "Kept" : ""}
							value={password}
							onInput={(e) => setPassword(e.currentTarget.value)}
						/>
						<span class="c-field__help">
							{kept ? "Leave it empty to keep the password." : "Cahut keeps it sealed, and never shows it again."}
						</span>
					</label>
				</div>
				<div class="form__row">
					<label class="c-field">
						<span class="c-field__label">From</span>
						<input class="c-input c-mono" type="email" required placeholder="cahut@example.com" value={from} onInput={(e) => setFrom(e.currentTarget.value)} />
						<span class="c-field__help">An address that the relay accepts.</span>
					</label>
					<label class="c-field">
						<span class="c-field__label">Name</span>
						<input class="c-input" placeholder="The name of the Cahut" value={name} onInput={(e) => setName(e.currentTarget.value)} />
					</label>
				</div>
				{error && <Failure message={error} />}
				<div class="actions">
					<button class="c-btn c-btn--primary" type="submit" disabled={busy}>
						{busy ? "Connecting…" : "Connect and save"}
					</button>
					{r && (
						<ConfirmButton
							label="Remove the relay"
							question="Remove it? This Cahut stops sending e-mail, and the console no longer offers what would."
							confirm="Remove"
							danger
							onConfirm={remove}
						/>
					)}
				</div>
			</form>
		</section>
	);
}
