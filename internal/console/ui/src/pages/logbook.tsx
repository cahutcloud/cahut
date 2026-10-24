// The Logbook: who acted, through which way in, on what, and with what
// result. Administrators read all of it; anyone else what they did, what
// their AIs did and what happened in their Rooms.
import { useState } from "preact/hooks";
import type { Entry } from "../api.gen";
import { Failure, PageHead } from "../components";
import { useConsole } from "../context";
import { useLoad } from "../load";
import { Link } from "../router";

const step = 200;
const most = 1000;

/** Who acted, in a few words, as in "Claude Code for olivier". */
function who(e: Entry): string {
	if (e.actor === "ai") return `${e.client ?? "An AI"} for ${e.user ?? "someone"}`;
	if (e.actor === "admin") return e.user ? `${e.user}, administrator of the server` : "The command line";
	if (e.user) return e.user;
	// A call that came without signing in, and names no account: a reader
	// who opens an invitation, a sign-in with an unknown handle, or a
	// forgotten password of an unknown account.
	return e.operation === "open_invitation" ? "Someone with an invitation" : "Someone not signed in";
}

const outcomes: Record<string, { word: string; badge: string }> = {
	ok: { word: "Done", badge: "c-badge--success" },
	knocked: { word: "Knocked", badge: "c-badge--warning" },
	refused: { word: "Refused", badge: "c-badge--warning" },
	failed: { word: "Failed", badge: "c-badge--danger" },
};

function day(at: string): string {
	const d = new Date(at);
	const today = new Date();
	const yesterday = new Date(today.getTime() - 24 * 3600 * 1000);
	if (d.toDateString() === today.toDateString()) return "Today";
	if (d.toDateString() === yesterday.toDateString()) return "Yesterday";
	return d.toLocaleDateString("en-GB", { weekday: "long", day: "numeric", month: "long" });
}

export function Logbook({ of }: { of: "cahut" | "server" }) {
	const api = useConsole();
	const [limit, setLimit] = useState(step);
	const [person, setPerson] = useState("");
	const [operation, setOperation] = useState("");
	const [outcome, setOutcome] = useState("");
	const { data, error } = useLoad(
		() => (of === "server" ? api.server("list_logbook", { limit }) : api.cahut("list_logbook", { limit })),
		[of, limit],
	);
	if (error) return <Failure message={error} />;
	if (!data) return null;

	const all = data.entries;
	const people = [...new Set(all.map(who))].sort();
	const operations = [...new Set(all.map((e) => e.operation))].sort();
	const shown = all.filter(
		(e) => (!person || who(e) === person) && (!operation || e.operation === operation) && (!outcome || e.outcome === outcome),
	);
	const days: [string, Entry[]][] = [];
	for (const e of shown) {
		const d = day(e.at);
		if (days.length === 0 || days[days.length - 1][0] !== d) days.push([d, []]);
		days[days.length - 1][1].push(e);
	}

	return (
		<>
			<PageHead label={of === "server" ? "This server" : "This Cahut"} title="Logbook">
				{of === "server"
					? "Every operation on the server: the Cahuts declared, its administrators and their sign-ins."
					: api.isAdmin
						? "Every action in this Cahut, by people and by AIs, with its result."
						: "What you did, what your AIs did, and what happened in your Rooms."}
			</PageHead>
			<div class="filters" role="search">
				<label class="c-field">
					<span class="c-field__label">Who</span>
					<select class="c-input" value={person} onChange={(e) => setPerson(e.currentTarget.value)}>
						<option value="">Everyone</option>
						{people.map((p) => (
							<option key={p}>{p}</option>
						))}
					</select>
				</label>
				<label class="c-field">
					<span class="c-field__label">Operation</span>
					<select class="c-input" value={operation} onChange={(e) => setOperation(e.currentTarget.value)}>
						<option value="">All</option>
						{operations.map((o) => (
							<option key={o}>{o}</option>
						))}
					</select>
				</label>
				<label class="c-field">
					<span class="c-field__label">Result</span>
					<select class="c-input" value={outcome} onChange={(e) => setOutcome(e.currentTarget.value)}>
						<option value="">All</option>
						<option value="ok">Done</option>
						<option value="knocked">Knocked</option>
						<option value="refused">Refused</option>
						<option value="failed">Failed</option>
					</select>
				</label>
			</div>
			{shown.length === 0 && <p class="c-muted">No entry {all.length > 0 ? "matches these filters" : "yet"}.</p>}
			{days.map(([d, entries]) => (
				<section key={d} class="c-card panel logbook-day">
					<h2 class="c-label">{d}</h2>
					<div class="c-log">
						{entries.map((e) => {
							const o = outcomes[e.outcome] ?? outcomes.failed;
							return (
								<div key={e.id} class="c-log__row">
									<span class="c-log__time">{new Date(e.at).toLocaleTimeString("en-GB", { hour: "2-digit", minute: "2-digit" })}</span>
									<div class="log-entry">
										<span>
											{e.actor === "ai" ? (
												<span class="c-ai">
													<span class="c-ai__mark" aria-hidden="true">
														{(e.client ?? "AI").slice(0, 2)}
													</span>
													{who(e)}
												</span>
											) : (
												<b>{who(e)}</b>
											)}{" "}
											<span class="c-mono log-entry__op">{e.operation}</span> {e.target && <span>{e.target}</span>}
										</span>
										{(e.detail || e.error) && <span class="c-small c-muted">{e.error ?? e.detail}</span>}
										{e.knock_id && of === "cahut" && (
											<Link class="c-link c-small" href={`/knocks/${e.knock_id}`}>
												Knock {e.knock_id}
											</Link>
										)}
										<span class="c-mono c-muted log-entry__via">via {e.via}</span>
									</div>
									<span class={`c-badge ${o.badge}`}>{o.word}</span>
								</div>
							);
						})}
					</div>
				</section>
			))}
			{all.length === limit && limit < most && (
				<button class="c-btn c-btn--secondary" type="button" onClick={() => setLimit(Math.min(limit + step, most))}>
					Show older entries
				</button>
			)}
		</>
	);
}
