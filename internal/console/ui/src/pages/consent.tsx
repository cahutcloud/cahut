// The consent screen of OAuth: an AI client sends the person here to connect
// to this Cahut, and the person grants it a scope and some Rooms, or refuses.
// See docs/adr/0016-oauth-for-ai-clients.md.
import { useEffect, useState } from "preact/hooks";
import { ApiError, cahut, explain } from "../api";
import type { Authorization, Request as AuthRequest, LegalLinks } from "../api.gen";
import { Failure, Languages, Legal, useArmed } from "../components";
import { de, words } from "../i18n";
import { Logo } from "../logo";
import type { Session } from "../session";

const t = words({
	en: {
		sendingBack: "Sending you back…",
		reading: "Reading the request…",
		cannotConnect: "This AI cannot connect",
		invalid: "The request is invalid.",
		learnsWhy: (client: string, host: string) => (
			<>
				Nothing was granted. {client} learns why at <b class="c-mono">{host}</b>, if you go back there.
			</>
		),
		goBack: (host: string) => `Go back to ${host}`,
		tryAgain: "Nothing was granted. Go back to the AI and try again, or ask it to register again.",
		connectTitle: (client: string) => `Connect ${client}`,
		asks: (client: string, name: string, host: string, back: string) => (
			<>
				{client} asks to act for you, {name}, in the Cahut at {host}. Then it goes back to <b class="c-mono">{back}</b>.
			</>
		),
		describedBy: (host: string) => (
			<>
				Its description comes from <b class="c-mono">{host}</b>.
			</>
		),
		anyName: (host: string) =>
			`Anyone can register an AI under any name: connect it only if you started this from the AI, and ${host} is where you use it.`,
		mayDo: "What it may do",
		write: "Read and publish: everything you can do in your Rooms",
		read: "Read only: your Rooms, their Windows and the Logbook",
		where: "Where",
		allRooms: "All your Rooms, and those you join later",
		someRooms: "Only these Rooms",
		personalRoom: "Your personal Room",
		connect: "Connect",
		refuse: "Refuse",
		logbook: "Every call it makes goes to the Logbook. You can disconnect it at any time, in AIs and access.",
	},
	fr: {
		sendingBack: "Redirection…",
		reading: "Lecture de la demande…",
		cannotConnect: "Cette IA ne peut pas se connecter",
		invalid: "La demande n'est pas valide.",
		learnsWhy: (client, host) => (
			<>
				Rien n'a été accordé. {client} en apprendra la raison sur <b class="c-mono">{host}</b>, si vous y retournez.
			</>
		),
		goBack: (host) => `Retourner sur ${host}`,
		tryAgain: "Rien n'a été accordé. Retournez dans l'IA et réessayez, ou demandez-lui de s'enregistrer à nouveau.",
		connectTitle: (client) => `Connecter ${client}`,
		asks: (client, name, host, back) => (
			<>
				{client} demande à agir pour vous, {name}, dans la Cahut {de(host)}. Vous revenez ensuite sur <b class="c-mono">{back}</b>.
			</>
		),
		describedBy: (host) => (
			<>
				Sa description vient de <b class="c-mono">{host}</b>.
			</>
		),
		anyName: (host) =>
			`N'importe qui peut enregistrer une IA sous n'importe quel nom : connectez-la seulement si vous avez lancé cette connexion depuis l'IA, et si vous l'utilisez bien sur ${host}.`,
		mayDo: "Ce qu'elle peut faire",
		write: "Lecture et publication : tout ce que vous pouvez faire dans vos Rooms",
		read: "Lecture seule : vos Rooms, leurs Windows et le Logbook",
		where: "Où",
		allRooms: "Toutes vos Rooms, et celles que vous rejoindrez plus tard",
		someRooms: "Seulement ces Rooms",
		personalRoom: "Votre Room personnelle",
		connect: "Connecter",
		refuse: "Refuser",
		logbook: "Chacun de ses appels va au Logbook. Vous pouvez la déconnecter à tout moment, sur la page IA et accès.",
	},
});

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
	// An invalid request, which the client learns at this address, once the
	// person chooses to go back: the console sends nobody to a site that
	// they did not see named.
	const [back, setBack] = useState<Authorization | null>(null);
	const armed = useArmed();

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

	// The foot of the page links the legal pages of the Cahut, which the
	// directories of AI clients look for here. See
	// docs/adr/0042-the-legal-pages-of-a-cahut.md.
	const [legal, setLegal] = useState<LegalLinks>();
	useEffect(() => {
		cahut("get_cahut", {}, session.token).then(
			(a) => setLegal(a.legal),
			() => undefined,
		);
	}, []);

	useEffect(() => {
		cahut("describe_authorization", { request: req }, session.token).then((a) => {
			if (a.redirect) {
				setBack(a);
				return;
			}
			setAuth(a);
			setScope(a.scope);
		}, handle);
	}, []);

	const approve = async (e: Event) => {
		e.preventDefault();
		if (!armed) return; // Enter in the form, before the page armed
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
				{!auth && !error && !back && <p class="c-muted">{leaving ? t.sendingBack : t.reading}</p>}
				{back && (
					<div class="auth__head">
						<h1 class="c-h3">{t.cannotConnect}</h1>
						<Failure message={back.error ?? t.invalid} />
						<p class="c-small c-muted">{t.learnsWhy(back.client, back.redirect_host)}</p>
						<div>
							<button class="c-btn c-btn--secondary" type="button" disabled={leaving} onClick={() => leave(back.redirect ?? "")}>
								{t.goBack(back.redirect_host)}
							</button>
						</div>
					</div>
				)}
				{!auth && error && (
					<div class="auth__head">
						<h1 class="c-h3">{t.cannotConnect}</h1>
						<Failure message={error} />
						<p class="c-small c-muted">{t.tryAgain}</p>
					</div>
				)}
				{auth && (
					<form onSubmit={approve}>
						<div class="auth__head">
							<h1 class="c-h3">{t.connectTitle(auth.client)}</h1>
							<p class="c-small c-muted">{t.asks(auth.client, session.name, location.hostname, auth.redirect_host)}</p>
							{auth.client_host ? (
								<p class="c-small c-muted">{t.describedBy(auth.client_host)}</p>
							) : (
								<p class="c-small c-muted">{t.anyName(auth.redirect_host)}</p>
							)}
						</div>

						<fieldset class="consent__group">
							<legend class="c-field__label">{t.mayDo}</legend>
							<label class="c-check">
								<input type="radio" name="scope" checked={scope === "write"} disabled={auth.scope === "read"} onChange={() => setScope("write")} />
								<span>{t.write}</span>
							</label>
							<label class="c-check">
								<input type="radio" name="scope" checked={scope === "read"} onChange={() => setScope("read")} />
								<span>{t.read}</span>
							</label>
						</fieldset>

						<fieldset class="consent__group">
							<legend class="c-field__label">{t.where}</legend>
							<label class="c-check">
								<input type="radio" name="rooms" checked={allRooms} onChange={() => setAllRooms(true)} />
								<span>{t.allRooms}</span>
							</label>
							<label class="c-check">
								<input type="radio" name="rooms" checked={!allRooms} onChange={() => setAllRooms(false)} />
								<span>{t.someRooms}</span>
							</label>
							{!allRooms && (
								<div class="consent__rooms">
									{auth.rooms.map((r) => (
										<label class="c-check" key={r.name}>
											<input type="checkbox" checked={rooms.includes(r.name)} onChange={(e) => toggle(r.name, e.currentTarget.checked)} />
											<span>
												{r.kind === "personal" ? t.personalRoom : r.title} <span class="c-mono c-muted">{r.name}</span>
											</span>
										</label>
									))}
								</div>
							)}
						</fieldset>

						{error && <Failure message={error} />}
						<div class="consent__actions">
							<button class="c-btn c-btn--primary" type="submit" disabled={leaving || !armed || (!allRooms && rooms.length === 0)}>
								{t.connect}
							</button>
							<button class="c-btn c-btn--secondary" type="button" disabled={leaving} onClick={deny}>
								{t.refuse}
							</button>
						</div>
						<p class="c-small c-muted">{t.logbook}</p>
					</form>
				)}
			</div>
			<Languages />
			<Legal links={legal} />
		</main>
	);
}
