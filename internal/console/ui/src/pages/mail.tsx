// The e-mail relay of this Cahut, for its administrators, or of the server,
// for its administrators: the SMTP relay that sends activation links,
// invitations and notices, a test, and how the last message went. See
// docs/adr/0031-e-mail-through-a-relay.md.
import { Send } from "lucide-preact";
import { useState } from "preact/hooks";
import { explain } from "../api";
import type { Delivery, SetMailRelayInput, TestSent } from "../api.gen";
import { ago, ConfirmButton, Failure, Notice, PageHead } from "../components";
import { useConsole } from "../context";
import { useLoad } from "../load";

export function MailPage({ of }: { of: "cahut" | "server" }) {
	const api = useConsole();
	const relay = {
		get: (): Promise<Delivery> => (of === "server" ? api.server("get_mail_relay", {}) : api.cahut("get_mail_relay", {})),
		set: (input: SetMailRelayInput): Promise<Delivery> =>
			of === "server" ? api.server("set_mail_relay", input) : api.cahut("set_mail_relay", input),
		test: (): Promise<TestSent> => (of === "server" ? api.server("send_test_mail", {}) : api.cahut("send_test_mail", {})),
	};
	const { data, error, reload } = useLoad(relay.get, [of]);
	const [failure, setFailure] = useState("");
	const [said, setSaid] = useState("");
	if (of === "cahut" && !api.isAdmin) return <Notice>Only the administrators of this Cahut set its e-mail.</Notice>;
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
	return (
		<>
			<PageHead label={of === "server" ? "This server" : "This Cahut"} title="E-mail">
				{of === "server"
					? "The SMTP relay of this server sends its e-mail, and that of every Cahut without a relay of its own."
					: "The SMTP relay that sends the e-mail of this Cahut, from an address of its own: activation links, invitations and notices. Without one, its e-mail goes through the relay of the server."}
			</PageHead>
			<section class="section c-card panel">
				<h2 class="c-h4">How e-mail goes out</h2>
				<p class="c-small">{describe(data, of)}</p>
				{data.last_sent && <p class="c-small c-muted">A message last went out {ago(data.last_sent).toLowerCase()}.</p>}
				{data.last_error && (
					<div class="c-alert c-alert--warning" role="status">
						The last message that failed, {data.last_error_at ? ago(data.last_error_at).toLowerCase() : "lately"}: {data.last_error}
					</div>
				)}
				{data.sends && (
					<div>
						<button
							class="c-btn c-btn--secondary"
							type="button"
							onClick={() => act(async () => `A test went to ${(await relay.test()).to}. Look in that mailbox.`)}
						>
							<Send class="i" size={16} aria-hidden="true" />
							Send a test to your address
						</button>
					</div>
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
				of={of}
				onSave={async (input) => {
					await relay.set(input);
					reload();
				}}
				onRemove={async () => {
					await relay.set({ host: "" });
					reload();
				}}
			/>
		</>
	);
}

/** Says through which relay e-mail goes out, if any. */
function describe(d: Delivery, of: "cahut" | "server"): string {
	if (d.relay) {
		const from = d.relay.name ? `${d.relay.name} <${d.relay.from}>` : d.relay.from;
		return `Through ${d.relay.host}:${d.relay.port}, with ${d.relay.security === "tls" ? "TLS" : d.relay.security === "starttls" ? "STARTTLS" : "no TLS"}, from ${from}.`;
	}
	if (d.server) return "This Cahut has no relay of its own: its e-mail goes through the relay of the server.";
	return of === "server"
		? "This server has no relay: the Cahuts without one of their own send no e-mail."
		: "Neither this Cahut nor the server has a relay: links go by hand, and invitations and notices do not go out.";
}

function RelayForm(props: { delivery: Delivery; of: "cahut" | "server"; onSave: (input: SetMailRelayInput) => Promise<void>; onRemove: () => Promise<void> }) {
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
			<h2 class="c-h4">{r ? "Change the relay" : props.of === "server" ? "Set the relay of the server" : "Set a relay for this Cahut"}</h2>
			<p class="c-small c-muted">
				{props.of === "server"
					? "Your provider of e-mail gives these settings. The relay may be on this machine: without TLS, only on its loopback address."
					: "Your provider of e-mail gives these settings. The relay is on the internet, and speaks TLS. Cahut connects and signs in before it keeps it."}
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
							{props.of === "server" && <option value="none">None, on this machine only</option>}
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
						<input class="c-input" placeholder={props.of === "server" ? "Cahut" : "The name of the Cahut"} value={name} onInput={(e) => setName(e.currentTarget.value)} />
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
							question={props.of === "server" ? "Remove it? The Cahuts without a relay of their own stop sending e-mail." : "Remove it? E-mail goes through the relay of the server, if it has one."}
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
