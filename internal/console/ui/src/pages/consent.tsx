// The consent screen of OAuth: an AI client sends the person here to connect
// to this Cahut, and the person grants it a scope and some Rooms, or refuses.
// See docs/adr/0016-oauth-for-ai-clients.md.
import { useEffect, useState } from "preact/hooks";
import { ApiError, cahut, explain } from "../api";
import type { Authorization, Request as AuthRequest } from "../api.gen";
import { Failure } from "../components";
import { Logo } from "../logo";
import type { Session } from "../session";

/** The parameters of the authorization request, from the page's address. */
function request(): AuthRequest {
	const q = new URLSearchParams(location.search);
	const r: AuthRequest = { client_id: q.get("client_id") ?? "" };
	const optional = ["response_type", "redirect_uri", "scope", "state", "code_challenge", "code_challenge_method", "resource"] as const;
	for (const key of optional) {
		const v = q.get(key);
		if (v !== null) {
			r[key] = v;
		}
	}
	return r;
}

export function Consent({ session, onExpired }: { session: Session; onExpired: () => void }) {
	const [req] = useState(request);
	const [auth, setAuth] = useState<Authorization | null>(null);
	const [error, setError] = useState("");
	const [leaving, setLeaving] = useState(false);
	const [scope, setScope] = useState("write");
	const [allRooms, setAllRooms] = useState(true);
	const [rooms, setRooms] = useState<string[]>([]);

	// A session that has ended signs in again, on this page.
	const handle = (err: unknown) => {
		if (err instanceof ApiError && err.status === 401) {
			onExpired();
			return;
		}
		setError(explain(err));
		setLeaving(false);
	};
	const leave = (to: string) => {
		setLeaving(true);
		location.assign(to);
	};

	useEffect(() => {
		cahut("describe_authorization", { request: req }, session.token).then((a) => {
			if (a.redirect) {
				leave(a.redirect);
				return;
			}
			setAuth(a);
			setScope(a.scope);
		}, handle);
	}, []);

	const approve = async (e: Event) => {
		e.preventDefault();
		setError("");
		setLeaving(true);
		try {
			const r = await cahut("approve_authorization", { request: req, scope, all_rooms: allRooms, rooms: allRooms ? [] : rooms }, session.token);
			leave(r.redirect);
		} catch (err) {
			handle(err);
		}
	};
	const deny = async () => {
		setError("");
		setLeaving(true);
		try {
			const r = await cahut("deny_authorization", { request: req }, session.token);
			leave(r.redirect);
		} catch (err) {
			handle(err);
		}
	};
	const toggle = (name: string, on: boolean) => setRooms((list) => (on ? [...list, name] : list.filter((r) => r !== name)));

	return (
		<main class="auth">
			<div class="c-card auth__card consent">
				<Logo height={32} />
				{!auth && !error && <p class="c-muted">{leaving ? "Sending you back…" : "Reading the request…"}</p>}
				{!auth && error && (
					<div class="auth__head">
						<h1 class="c-h3">This AI cannot connect</h1>
						<Failure message={error} />
						<p class="c-small c-muted">Nothing was granted. Go back to the AI and try again, or ask it to register again.</p>
					</div>
				)}
				{auth && (
					<form onSubmit={approve}>
						<div class="auth__head">
							<h1 class="c-h3">Connect {auth.client}</h1>
							<p class="c-small c-muted">
								{auth.client} asks to act for you, {session.name}, in the Cahut at {location.hostname}. Then it goes back to{" "}
								<b class="c-mono">{auth.redirect_host}</b>.
							</p>
							{auth.client_host ? (
								<p class="c-small c-muted">
									Its description comes from <b class="c-mono">{auth.client_host}</b>.
								</p>
							) : (
								<p class="c-small c-muted">
									Anyone can register an AI under any name: connect it only if you started this from the AI, and{" "}
									{auth.redirect_host} is where you use it.
								</p>
							)}
						</div>

						<fieldset class="consent__group">
							<legend class="c-field__label">What it may do</legend>
							<label class="c-check">
								<input type="radio" name="scope" checked={scope === "write"} disabled={auth.scope === "read"} onChange={() => setScope("write")} />
								<span>
									Read and publish: everything you can do in your Rooms
								</span>
							</label>
							<label class="c-check">
								<input type="radio" name="scope" checked={scope === "read"} onChange={() => setScope("read")} />
								<span>Read only: your Rooms, their Windows and the Logbook</span>
							</label>
						</fieldset>

						<fieldset class="consent__group">
							<legend class="c-field__label">Where</legend>
							<label class="c-check">
								<input type="radio" name="rooms" checked={allRooms} onChange={() => setAllRooms(true)} />
								<span>All your Rooms, and those you join later</span>
							</label>
							<label class="c-check">
								<input type="radio" name="rooms" checked={!allRooms} onChange={() => setAllRooms(false)} />
								<span>Only these Rooms</span>
							</label>
							{!allRooms && (
								<div class="consent__rooms">
									{auth.rooms.map((r) => (
										<label class="c-check" key={r.name}>
											<input type="checkbox" checked={rooms.includes(r.name)} onChange={(e) => toggle(r.name, e.currentTarget.checked)} />
											<span>
												{r.kind === "personal" ? "Your personal Room" : r.title} <span class="c-mono c-muted">{r.name}</span>
											</span>
										</label>
									))}
								</div>
							)}
						</fieldset>

						{error && <Failure message={error} />}
						<div class="consent__actions">
							<button class="c-btn c-btn--primary" type="submit" disabled={leaving || (!allRooms && rooms.length === 0)}>
								Connect
							</button>
							<button class="c-btn c-btn--secondary" type="button" disabled={leaving} onClick={deny}>
								Refuse
							</button>
						</div>
						<p class="c-small c-muted">
							Every call it makes goes to the Logbook. You can disconnect it at any time, in AIs and access.
						</p>
					</form>
				)}
			</div>
		</main>
	);
}
