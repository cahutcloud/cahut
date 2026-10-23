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
import { useLoad } from "../load";
import { href, Link } from "../router";

/** The states of a Knock, as a word and a badge. */
const states: Record<string, { word: string; badge: string }> = {
	waiting: { word: "Knock first", badge: "c-badge--warning" },
	done: { word: "Approved", badge: "c-badge--success" },
	failed: { word: "Failed", badge: "c-badge--danger" },
	denied: { word: "Declined", badge: "c-badge--neutral" },
	expired: { word: "Expired", badge: "c-badge--neutral" },
	canceled: { word: "Canceled", badge: "c-badge--neutral" },
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
			return `make ${k.target ?? text(i.window)} public`;
		case "purge_window":
			return `delete ${k.target ?? text(i.window)} for good, with its versions and its files`;
		case "delete_room":
			return `delete the Room ${room} for good, with its Windows and their files`;
		case "create_user":
			return `add ${text(i.name)} (${text(i.handle)}) to this Cahut${i.admin ? ", as an administrator" : ""}`;
		case "set_user_role":
			return `make ${user} ${i.role === "admin" ? "an administrator" : "a member"} of this Cahut`;
		case "disable_user":
			return `disable ${user}: their sessions end, and their AIs stop`;
		case "enable_user":
			return `enable ${user} again`;
		case "add_room_member":
			return `add ${user} to the Room ${room}${i.role === "owner" ? ", as an owner" : ""}`;
		case "set_room_member_role":
			return `make ${user} ${i.role === "owner" ? "an owner" : "a member"} of the Room ${room}`;
		case "remove_room_member":
			return `remove ${user} from the Room ${room}`;
		case "set_ai_reach": {
			const rooms = Array.isArray(i.rooms) ? i.rooms : [];
			const where = i.all_rooms ? "all your Rooms" : `${rooms.length === 1 ? "the Room" : "the Rooms"} ${rooms.join(", ")}`;
			return `${i.scope === "write" ? "read and publish" : "read only"}, in ${where}`;
		}
	}
	return `run ${k.operation}${k.target ? ` on ${k.target}` : ""}`;
}

/** Says when a Knock expires, as in "expires in 23 h". */
function expires(k: Knock): string {
	const minutes = Math.round((new Date(k.expires_at).getTime() - Date.now()) / 60000);
	if (minutes < 1) return "expires now";
	if (minutes < 60) return `expires in ${minutes} min`;
	return `expires in ${Math.round(minutes / 60)} h`;
}

function when(at: string): string {
	return new Date(at).toLocaleString("en-GB", { dateStyle: "medium", timeStyle: "short" });
}

/** What happened to a Knock, once it no longer waits. */
function outcome(k: Knock): string {
	switch (k.state) {
		case "done":
			return `Approved ${when(k.answered_at ?? k.created_at)}: the call ran.`;
		case "failed":
			return `Approved ${when(k.answered_at ?? k.created_at)}, but the call failed: ${k.error ?? "no reason given"}`;
		case "denied":
			return `Declined ${when(k.answered_at ?? k.created_at)}: the call never ran.`;
		case "expired":
			return `Expired ${when(k.expires_at)}, with no answer: the call never ran.`;
		case "canceled":
			return "Canceled: the AI lost its access before an answer, and the call never ran.";
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
						Wants to {asks(k)}.
					</Link>
				) : (
					<>Wants to {asks(k)}.</>
				)}
			</p>
			<p class="c-mono c-muted knock__meta">
				{k.state === "waiting" ? expires(k) : when(k.created_at)} · on behalf of {k.user} · Knock {k.id}
			</p>
			{k.state !== "waiting" && <p class="c-small c-muted">{outcome(k)}</p>}
			{error && <Failure message={error} />}
			{k.state === "waiting" && (
				<div class="c-knock__actions">
					<button class="c-btn c-btn--primary c-btn--sm" type="button" disabled={busy || !armed} onClick={() => answer(true)}>
						<Check class="i" size={14} aria-hidden="true" />
						Approve
					</button>
					<button class="c-btn c-btn--secondary c-btn--sm" type="button" disabled={busy} onClick={() => answer(false)}>
						Decline
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
			<PageHead label="Knock first" title="Knocks">
				What your AIs cannot undo, or what changes what a person may do, waits here for you. Approving runs the call, as the AI
				asked it; a Knock expires after 24 hours.
			</PageHead>
			<section class="section">
				<h2 class="c-h4">Waiting for you</h2>
				{last && <Result knock={last.knock} output={last.output} />}
				{waiting.length === 0 && <p class="c-muted">Nothing waits for you.</p>}
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
					<h2 class="c-h4">Answered, expired or canceled</h2>
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
		return <Notice>The person an AI acts for answers its Knocks. Sign in to this Cahut as that person.</Notice>;
	}
	if (error) return <Failure message={error} />;
	if (!data) return null;
	const k = data;
	const fields = Object.entries(k.input);
	return (
		<>
			<PageHead label="Knock first" title={`Knock ${k.id}`}>
				{k.client} asked on {when(k.created_at)}, on behalf of {k.user}.{" "}
				<Link class="c-link" href="/knocks">
					All Knocks
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
					<h2 class="c-h4">The call</h2>
					<p class="c-small">
						<span class="c-mono">{k.operation}</span>
						{k.room && (
							<>
								{" "}
								in the Room <Link class="c-link" href={`/rooms/${k.room}`}>{k.room}</Link>
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
					{o.name} has an account. Send them this link, where they choose their password. It works once, until{" "}
					{new Date(o.activation.expires_at).toLocaleDateString("en-GB", { dateStyle: "long" })}, and is shown only now.
				</p>
				<CopyField label="the activation link" value={`${location.origin}${href("/activate")}#${o.activation.code}`} />
			</div>
		);
	}
	if (k.operation === "set_window_access" && o.url) {
		return (
			<div class="result section" role="status">
				<p class="c-small">The Window is public now: anyone with its address opens it.</p>
				<CopyField label="the address" value={o.url} />
			</div>
		);
	}
	return null;
}
