// The first screen once signed in: for a person, their latest Windows and
// their Rooms; for an administrator of the server, the Cahuts of the server.
import { ArrowUpRight, Plus } from "lucide-preact";
import { useState } from "preact/hooks";
import { explain } from "../api";
import type { NewCahut } from "../api.gen";
import { CopyField, Failure, Notice, PageHead, plural, WindowCard } from "../components";
import { useConsole } from "../context";
import { useLoad } from "../load";
import { Link } from "../router";
import { RoomCards } from "./rooms";

export function Overview() {
	return useConsole().session.kind === "admin" ? <ServerOverview /> : <PersonOverview />;
}

function PersonOverview() {
	const api = useConsole();
	const { data, error } = useLoad(() =>
		Promise.all([api.cahut("get_account", {}), api.cahut("list_windows", {}), api.cahut("list_rooms", {})]),
	);
	if (error) return <Failure message={error} />;
	if (!data) return null;
	const [account, { windows }, { rooms }] = data;
	return (
		<>
			<PageHead label="Overview" title={location.hostname}>
				Signed in as {account.name}, {account.role === "admin" ? "an administrator" : "a member"} of this Cahut.
			</PageHead>
			<section class="section">
				<h2 class="c-h4">Latest Windows</h2>
				{windows.length === 0 ? (
					<Notice>
						No Window yet. Ask your AI to publish one: <Link class="c-link" href="/access">connect an AI</Link> first.
					</Notice>
				) : (
					<div class="cards cards--windows">
						{windows.slice(0, 6).map((w) => (
							<WindowCard key={w.url} window={w} showRoom />
						))}
					</div>
				)}
			</section>
			<section class="section">
				<h2 class="c-h4">Your Rooms</h2>
				<RoomCards rooms={rooms} />
			</section>
		</>
	);
}

function ServerOverview() {
	const api = useConsole();
	const { data, error, reload } = useLoad(() => api.server("list_domains", {}));
	if (error) return <Failure message={error} />;
	if (!data) return null;
	// Each Cahut has its own console, on the same port of its own domain.
	const origin = (domain: string) => `${location.protocol}//${domain}${location.port ? `:${location.port}` : ""}`;
	return (
		<>
			<PageHead label="This server" title={plural(data.cahuts.length, "Cahut", "Cahuts")}>
				You administer this server, and here the Cahut of {location.hostname}. For another Cahut, sign in at its own console.
			</PageHead>
			<div class="cards">
				{data.cahuts.map((c) => (
					<article key={c.domain} class="c-card card">
						<div class="c-window__title c-mono">{c.domain}</div>
						<span class="c-small c-muted">Declared on {new Date(c.created_at).toLocaleDateString("en-GB", { dateStyle: "medium" })}</span>
						{c.domain === location.hostname ? (
							<span class="c-small">You are at its console.</span>
						) : (
							<a class="c-link c-small" href={`${origin(c.domain)}/server/sign-in`}>
								Open its console
								<ArrowUpRight class="i" size={14} aria-hidden="true" />
							</a>
						)}
					</article>
				))}
			</div>
			<DeclareCahut origin={origin} onDeclared={reload} />
		</>
	);
}

/** The form that declares a Cahut, with its first administrator. */
function DeclareCahut({ origin, onDeclared }: { origin: (domain: string) => string; onDeclared: () => void }) {
	const api = useConsole();
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
			<h2 class="c-h4">Declare a Cahut</h2>
			<p class="c-small c-muted">
				Point the DNS record of the domain to this server first. The Cahut gets its certificate at the first visit.
			</p>
			<form class="form" onSubmit={submit}>
				<label class="c-field">
					<span class="c-field__label">Domain</span>
					<input class="c-input c-mono" required placeholder="cahut.example.com" value={domain} onInput={(e) => setDomain(e.currentTarget.value)} />
				</label>
				<div class="form__row">
					<label class="c-field">
						<span class="c-field__label">Handle of its first administrator</span>
						<input class="c-input" autocapitalize="none" value={handle} onInput={(e) => setHandle(e.currentTarget.value)} />
					</label>
					<label class="c-field">
						<span class="c-field__label">Their name</span>
						<input class="c-input" value={name} onInput={(e) => setName(e.currentTarget.value)} />
					</label>
				</div>
				{error && <Failure message={error} />}
				<div>
					<button class="c-btn c-btn--primary" type="submit">
						<Plus class="i" size={16} aria-hidden="true" />
						Declare
					</button>
				</div>
			</form>
			{done && (
				<div class="result" role="status">
					<p class="c-small">
						Declared <span class="c-mono">{done.domain}</span>.
						{done.admin?.activation && ` Send ${done.admin.name} this link, where they choose their password. It works once, for 7 days.`}
					</p>
					{done.admin?.activation && (
						<CopyField label="the activation link" value={`${origin(done.domain)}/activate#${done.admin.activation.code}`} />
					)}
				</div>
			)}
		</section>
	);
}
