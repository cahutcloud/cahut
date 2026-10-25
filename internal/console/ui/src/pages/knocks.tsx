// Knocks: the calls of your AIs that cannot be undone, or that change what a
// person may do. They wait for you to approve them, which runs them, or to
// decline them. An AI gives you the link of its Knock, /knocks/<id>. See
// docs/adr/0021-ais-knock-first.md.
import { Check } from "lucide-preact";
import { useState } from "preact/hooks";
import { explain } from "../api";
import type { Activation, Answer, Knock } from "../api.gen";
import { CopyField, Failure, Notice, PageHead, useArmed } from "../components";
import { useConsole } from "../context";
import { de, locale, words } from "../i18n";
import { useLoad } from "../load";
import { href, Link } from "../router";

const t = words({
	en: {
		waiting: "Knock first",
		done: "Approved",
		failed: "Failed",
		denied: "Declined",
		expired: "Expired",
		canceled: "Canceled",
		asks: {
			setWindowAccess: (window: string) => `make ${window} public`,
			purgeWindow: (window: string) => `delete ${window} for good, with its versions and its files`,
			deleteRoom: (room: string) => `delete the Room ${room} for good, with its Windows and their files`,
			createUser: (name: string, handle: string, admin: boolean) =>
				`add ${name} (${handle}) to this Cahut${admin ? ", as an administrator" : ""}`,
			setUserRole: (user: string, admin: boolean) => `make ${user} ${admin ? "an administrator" : "a member"} of this Cahut`,
			disableUser: (user: string) => `disable ${user}: their sessions end, and their AIs stop`,
			enableUser: (user: string) => `enable ${user} again`,
			addRoomMember: (user: string, room: string, owner: boolean) => `add ${user} to the Room ${room}${owner ? ", as an owner" : ""}`,
			setRoomMemberRole: (user: string, room: string, owner: boolean) =>
				`make ${user} ${owner ? "an owner" : "a member"} of the Room ${room}`,
			removeRoomMember: (user: string, room: string) => `remove ${user} from the Room ${room}`,
			setAIReach: (write: boolean, all: boolean, rooms: string[]) =>
				`${write ? "read and publish" : "read only"}, in ${all ? "all your Rooms" : `${rooms.length === 1 ? "the Room" : "the Rooms"} ${rooms.join(", ")}`}`,
			run: (operation: string, target: string | undefined) => `run ${operation}${target ? ` on ${target}` : ""}`,
		},
		expiresNow: "expires now",
		expiresInMinutes: (n: number) => `expires in ${n} min`,
		expiresInHours: (n: number) => `expires in ${n} h`,
		approvedRan: (when: string) => `Approved ${when}: the call ran.`,
		approvedFailed: (when: string, error: string | undefined) => `Approved ${when}, but the call failed: ${error ?? "no reason given"}`,
		declinedNeverRan: (when: string) => `Declined ${when}: the call never ran.`,
		expiredNeverRan: (when: string) => `Expired ${when}, with no answer: the call never ran.`,
		canceledNeverRan: "Canceled: the AI lost its access before an answer, and the call never ran.",
		wantsTo: (what: string) => `Wants to ${what}.`,
		onBehalfOf: (user: string) => `on behalf of ${user}`,
		approve: "Approve",
		decline: "Decline",
		knocksText:
			"What your AIs cannot undo, or what changes what a person may do, waits here for you. Approving runs the call, as the AI asked it; a Knock expires after 24 hours.",
		waitingForYou: "Waiting for you",
		nothingWaits: "Nothing waits for you.",
		past: "Answered, expired or canceled",
		onlyThePerson: "The person an AI acts for answers its Knocks. Sign in to this Cahut as that person.",
		asked: (client: string, when: string, user: string) => `${client} asked on ${when}, on behalf of ${user}.`,
		allKnocks: "All Knocks",
		theCall: "The call",
		inTheRoom: "in the Room",
		hasAccount: (name: string, until: string) =>
			`${name} has an account. Send them this link, where they choose their password. It works once, until ${until}, and is shown only now.`,
		activationLink: "the activation link",
		publicNow: "The Window is public now: anyone with its address opens it.",
		address: "the address",
	},
	fr: {
		waiting: "Knock first",
		done: "Approuvé",
		failed: "Échec",
		denied: "Refusé",
		expired: "Expiré",
		canceled: "Annulé",
		asks: {
			setWindowAccess: (window) => `rendre ${window} publique`,
			purgeWindow: (window) => `supprimer définitivement ${window}, avec ses versions et ses fichiers`,
			deleteRoom: (room) => `supprimer définitivement la Room ${room}, avec ses Windows et leurs fichiers`,
			createUser: (name, handle, admin) => `ajouter ${name} (${handle}) à cette Cahut${admin ? ", comme administrateur" : ""}`,
			setUserRole: (user, admin) => `faire ${de(user)} ${admin ? "un administrateur" : "un membre"} de cette Cahut`,
			disableUser: (user) => `désactiver ${user} : ses sessions se terminent, et ses IA s'arrêtent`,
			enableUser: (user) => `réactiver ${user}`,
			addRoomMember: (user, room, owner) => `ajouter ${user} à la Room ${room}${owner ? ", comme propriétaire" : ""}`,
			setRoomMemberRole: (user, room, owner) => `faire ${de(user)} ${owner ? "un propriétaire" : "un membre"} de la Room ${room}`,
			removeRoomMember: (user, room) => `retirer ${user} de la Room ${room}`,
			setAIReach: (write, all, rooms) =>
				`${write ? "lire et publier" : "lire seulement"}, dans ${all ? "toutes vos Rooms" : `${rooms.length === 1 ? "la Room" : "les Rooms"} ${rooms.join(", ")}`}`,
			run: (operation, target) => `exécuter ${operation}${target ? ` sur ${target}` : ""}`,
		},
		expiresNow: "expire maintenant",
		expiresInMinutes: (n) => `expire dans ${n} min`,
		expiresInHours: (n) => `expire dans ${n} h`,
		approvedRan: (when) => `Approuvé le ${when} : l'appel s'est exécuté.`,
		approvedFailed: (when, error) => `Approuvé le ${when}, mais l'appel a échoué : ${error ?? "aucune raison donnée"}`,
		declinedNeverRan: (when) => `Refusé le ${when} : l'appel ne s'est jamais exécuté.`,
		expiredNeverRan: (when) => `Expiré le ${when}, sans réponse : l'appel ne s'est jamais exécuté.`,
		canceledNeverRan: "Annulé : l'IA a perdu son accès avant une réponse, et l'appel ne s'est jamais exécuté.",
		wantsTo: (what) => `Veut ${what}.`,
		onBehalfOf: (user) => `pour ${user}`,
		approve: "Approuver",
		decline: "Refuser",
		knocksText:
			"Ce que vos IA ne peuvent pas défaire, ou ce qui change ce qu'une personne peut faire, vous attend ici. Approuver exécute l'appel, tel que l'IA l'a demandé ; un Knock expire au bout de 24 heures.",
		waitingForYou: "En attente de votre réponse",
		nothingWaits: "Rien ne vous attend.",
		past: "Répondus, expirés ou annulés",
		onlyThePerson:
			"C'est la personne pour qui une IA agit qui répond à ses Knocks. Connectez-vous à cette Cahut en tant que cette personne.",
		asked: (client, when, user) => `${client} a demandé le ${when}, pour ${user}.`,
		allKnocks: "Tous les Knocks",
		theCall: "L'appel",
		inTheRoom: "dans la Room",
		hasAccount: (name, until) =>
			`${name} a un compte. Envoyez-lui ce lien, pour y choisir son mot de passe. Il ne sert qu'une fois, jusqu'au ${until}, et ne s'affiche que maintenant.`,
		activationLink: "le lien d'activation",
		publicNow: "La Window est maintenant publique : quiconque a son adresse l'ouvre.",
		address: "l'adresse",
	},
});

/** The states of a Knock, as a word and a badge. */
const states: Record<string, { word: string; badge: string }> = {
	waiting: { word: t.waiting, badge: "c-badge--warning" },
	done: { word: t.done, badge: "c-badge--success" },
	failed: { word: t.failed, badge: "c-badge--danger" },
	denied: { word: t.denied, badge: "c-badge--neutral" },
	expired: { word: t.expired, badge: "c-badge--neutral" },
	canceled: { word: t.canceled, badge: "c-badge--neutral" },
};

function text(v: unknown): string {
	return typeof v === "string" ? v : "";
}

/** What a Knock asks, after "Wants to", as in "make olivier/q3 public". */
export function asks(k: Knock): string {
	const i = k.input;
	const user = text(i.user);
	const room = text(i.room);
	switch (k.operation) {
		case "set_window_access":
			return t.asks.setWindowAccess(k.target ?? text(i.window));
		case "purge_window":
			return t.asks.purgeWindow(k.target ?? text(i.window));
		case "delete_room":
			return t.asks.deleteRoom(room);
		case "create_user":
			return t.asks.createUser(text(i.name), text(i.handle), Boolean(i.admin));
		case "set_user_role":
			return t.asks.setUserRole(user, i.role === "admin");
		case "disable_user":
			return t.asks.disableUser(user);
		case "enable_user":
			return t.asks.enableUser(user);
		case "add_room_member":
			return t.asks.addRoomMember(user, room, i.role === "owner");
		case "set_room_member_role":
			return t.asks.setRoomMemberRole(user, room, i.role === "owner");
		case "remove_room_member":
			return t.asks.removeRoomMember(user, room);
		case "set_ai_reach": {
			const rooms = Array.isArray(i.rooms) ? i.rooms.map(String) : [];
			return t.asks.setAIReach(i.scope === "write", Boolean(i.all_rooms), rooms);
		}
	}
	return t.asks.run(k.operation, k.target);
}

/** Says when a Knock expires, as in "expires in 23 h". */
function expires(k: Knock): string {
	const minutes = Math.round((new Date(k.expires_at).getTime() - Date.now()) / 60000);
	if (minutes < 1) return t.expiresNow;
	if (minutes < 60) return t.expiresInMinutes(minutes);
	return t.expiresInHours(Math.round(minutes / 60));
}

function when(at: string): string {
	return new Date(at).toLocaleString(locale, { dateStyle: "medium", timeStyle: "short" });
}

/** What happened to a Knock, once it no longer waits. */
function outcome(k: Knock): string {
	switch (k.state) {
		case "done":
			return t.approvedRan(when(k.answered_at ?? k.created_at));
		case "failed":
			return t.approvedFailed(when(k.answered_at ?? k.created_at), k.error);
		case "denied":
			return t.declinedNeverRan(when(k.answered_at ?? k.created_at));
		case "expired":
			return t.expiredNeverRan(when(k.expires_at));
		case "canceled":
			return t.canceledNeverRan;
	}
	return "";
}

/** A Knock, as a card, with Approve and Decline while it waits. */
function KnockCard(props: { knock: Knock; onAnswer?: (a: Answer | Knock) => void; link?: boolean }) {
	const api = useConsole();
	const k = props.knock;
	const [busy, setBusy] = useState(false);
	const [error, setError] = useState("");
	const armed = useArmed();
	const state = states[k.state] ?? states.expired;
	const answer = async (approve: boolean) => {
		setBusy(true);
		setError("");
		try {
			const out = approve ? await api.cahut("approve_knock", { knock: k.id }) : await api.cahut("deny_knock", { knock: k.id });
			api.reloadKnocks();
			props.onAnswer?.(out);
		} catch (err) {
			setError(explain(err));
		} finally {
			setBusy(false);
		}
	};
	return (
		<article class={`c-card ${k.state === "waiting" ? "c-knock" : "panel knock-past"}`}>
			<div class="c-knock__head">
				<span class="c-ai">
					<span class="c-ai__mark" aria-hidden="true">
						{k.client.slice(0, 2)}
					</span>
					{k.client}
				</span>
				<span class={`c-badge ${state.badge} knock__state`}>
					{k.state === "waiting" && <span class="c-dot" aria-hidden="true" />}
					{state.word}
				</span>
			</div>
			<p class="c-small">
				{props.link ? (
					<Link class="c-link" href={`/knocks/${k.id}`}>
						{t.wantsTo(asks(k))}
					</Link>
				) : (
					t.wantsTo(asks(k))
				)}
			</p>
			<p class="c-mono c-muted knock__meta">
				{k.state === "waiting" ? expires(k) : when(k.created_at)} · {t.onBehalfOf(k.user)} · Knock {k.id}
			</p>
			{k.state !== "waiting" && <p class="c-small c-muted">{outcome(k)}</p>}
			{error && <Failure message={error} />}
			{k.state === "waiting" && (
				<div class="c-knock__actions">
					<button class="c-btn c-btn--primary c-btn--sm" type="button" disabled={busy || !armed} onClick={() => answer(true)}>
						<Check class="i" size={14} aria-hidden="true" />
						{t.approve}
					</button>
					<button class="c-btn c-btn--secondary c-btn--sm" type="button" disabled={busy} onClick={() => answer(false)}>
						{t.decline}
					</button>
				</div>
			)}
		</article>
	);
}

/** The Knocks of your AIs: those that wait, then the latest answered. */
export function Knocks() {
	const api = useConsole();
	const { data, error, reload } = useLoad(() => api.cahut("list_knocks", {}), [api.knocks]);
	// What the last approval gave back, such as an activation link, shown once.
	const [last, setLast] = useState<Answer | null>(null);
	if (error) return <Failure message={error} />;
	if (!data) return null;
	const waiting = data.knocks.filter((k) => k.state === "waiting");
	const past = data.knocks.filter((k) => k.state !== "waiting");
	return (
		<>
			<PageHead label={t.waiting} title="Knocks">
				{t.knocksText}
			</PageHead>
			<section class="section">
				<h2 class="c-h4">{t.waitingForYou}</h2>
				{last && <Result knock={last.knock} output={last.output} />}
				{waiting.length === 0 && <p class="c-muted">{t.nothingWaits}</p>}
				<div class="knocks">
					{waiting.map((k) => (
						<KnockCard
							key={k.id}
							knock={k}
							link
							onAnswer={(a) => {
								setLast("output" in a ? a : null);
								reload();
							}}
						/>
					))}
				</div>
			</section>
			{past.length > 0 && (
				<section class="section">
					<h2 class="c-h4">{t.past}</h2>
					<div class="knocks">
						{past.map((k) => (
							<KnockCard key={k.id} knock={k} link />
						))}
					</div>
				</section>
			)}
		</>
	);
}

/** One Knock, where the link of an AI leads. */
export function KnockPage({ id }: { id: number }) {
	const api = useConsole();
	const { data, error, reload } = useLoad(() => api.cahut("get_knock", { knock: id }), [id]);
	const [output, setOutput] = useState<unknown>(undefined);
	if (api.session.kind !== "person") {
		return <Notice>{t.onlyThePerson}</Notice>;
	}
	if (error) return <Failure message={error} />;
	if (!data) return null;
	const k = data;
	const fields = Object.entries(k.input);
	return (
		<>
			<PageHead label={t.waiting} title={`Knock ${k.id}`}>
				{t.asked(k.client, when(k.created_at), k.user)}{" "}
				<Link class="c-link" href="/knocks">
					{t.allKnocks}
				</Link>
			</PageHead>
			<div class="layout">
				<section>
					<KnockCard
						knock={k}
						onAnswer={(a) => {
							if ("output" in a) setOutput(a.output);
							reload();
						}}
					/>
					<Result knock={k} output={output} />
				</section>
				<aside class="c-card panel aside">
					<h2 class="c-h4">{t.theCall}</h2>
					<p class="c-small">
						<span class="c-mono">{k.operation}</span>
						{k.room && (
							<>
								{" "}
								{t.inTheRoom} <Link class="c-link" href={`/rooms/${k.room}`}>{k.room}</Link>
							</>
						)}
					</p>
					{fields.length > 0 && (
						<dl class="knock__input">
							{fields.map(([name, value]) => (
								<div key={name}>
									<dt class="c-small c-muted">{name}</dt>
									<dd class="c-mono c-small">
										{typeof value === "string" ? value : Array.isArray(value) ? value.join(", ") : JSON.stringify(value)}
									</dd>
								</div>
							))}
						</dl>
					)}
				</aside>
			</div>
		</>
	);
}

/** What an approved call gave back, when the person needs it. */
function Result({ knock: k, output }: { knock: Knock; output: unknown }) {
	if (k.state !== "done" || !output || typeof output !== "object") return null;
	const o = output as { activation?: Activation; name?: string; url?: string; access?: string };
	if (k.operation === "create_user" && o.activation) {
		return (
			<div class="result section" role="status">
				<p class="c-small">
					{t.hasAccount(o.name ?? "", new Date(o.activation.expires_at).toLocaleDateString(locale, { dateStyle: "long" }))}
				</p>
				<CopyField label={t.activationLink} value={`${location.origin}${href("/activate")}#${o.activation.code}`} />
			</div>
		);
	}
	if (k.operation === "set_window_access" && o.url) {
		return (
			<div class="result section" role="status">
				<p class="c-small">{t.publicNow}</p>
				<CopyField label={t.address} value={o.url} />
			</div>
		);
	}
	return null;
}
