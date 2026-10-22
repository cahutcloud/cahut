// The first screen once signed in: for a person, their Rooms; for an
// administrator of the server, the Cahuts of the server.
import { ArrowUpRight, User, Users } from "lucide-preact";
import { cahut, server } from "../api";
import type { Room } from "../api.gen";
import { useLoad } from "../load";
import type { Session } from "../session";

interface Props {
	session: Session;
	onExpired: () => void;
}

function plural(n: number, one: string, many: string): string {
	return `${n} ${n === 1 ? one : many}`;
}

export function Overview({ session, onExpired }: Props) {
	return session.kind === "admin" ? <ServerOverview session={session} onExpired={onExpired} /> : <RoomsOverview session={session} onExpired={onExpired} />;
}

function RoomsOverview({ session, onExpired }: Props) {
	const { data, error } = useLoad(
		() => Promise.all([cahut("get_account", {}, session.token), cahut("list_rooms", {}, session.token)]),
		onExpired,
	);
	if (error) {
		return <Failure message={error} />;
	}
	if (!data) {
		return null;
	}
	const [account, { rooms }] = data;
	return (
		<>
			<div class="page-head">
				<span class="c-label">Overview</span>
				<h1 class="c-h1">{location.hostname}</h1>
				<p class="c-muted">
					Signed in as {account.name}, {account.role === "admin" ? "an administrator" : "a member"} of this Cahut.
				</p>
			</div>
			<h2 class="c-h4">Your Rooms</h2>
			<div class="cards">
				{rooms.map((r) => (
					<RoomCard key={r.name} room={r} />
				))}
			</div>
		</>
	);
}

function RoomCard({ room }: { room: Room }) {
	const Icon = room.kind === "personal" ? User : Users;
	return (
		<article class="c-card card">
			<div class="c-window__title">
				<Icon class="i" size={16} aria-hidden="true" />
				{room.title}
			</div>
			<span class="c-mono c-small c-muted">/{room.name}/</span>
			<div class="c-window__meta">
				<span>
					{plural(room.windows, "Window", "Windows")} · {plural(room.members, "member", "members")}
				</span>
				<span class="c-badge c-badge--neutral">{room.role === "owner" ? "Owner" : "Member"}</span>
			</div>
		</article>
	);
}

function ServerOverview({ session, onExpired }: Props) {
	const { data, error } = useLoad(() => server("list_domains", {}, session.token), onExpired);
	if (error) {
		return <Failure message={error} />;
	}
	if (!data) {
		return null;
	}
	// Each Cahut has its own console, on the same port of its own domain.
	const consoleOf = (domain: string) => `${location.protocol}//${domain}${location.port ? `:${location.port}` : ""}/`;
	return (
		<>
			<div class="page-head">
				<span class="c-label">This server</span>
				<h1 class="c-h1">{plural(data.cahuts.length, "Cahut", "Cahuts")}</h1>
				<p class="c-muted">
					You administer this server, and here the Cahut of {location.hostname}. For another Cahut, sign in at its own
					console.
				</p>
			</div>
			<div class="cards">
				{data.cahuts.map((c) => (
					<article key={c.domain} class="c-card card">
						<div class="c-window__title c-mono">{c.domain}</div>
						<span class="c-small c-muted">Declared on {new Date(c.created_at).toLocaleDateString("en-GB", { dateStyle: "medium" })}</span>
						{c.domain === location.hostname ? (
							<span class="c-small">You are at its console.</span>
						) : (
							<a class="c-link c-small" href={consoleOf(c.domain)}>
								Open its console
								<ArrowUpRight class="i" size={14} aria-hidden="true" />
							</a>
						)}
					</article>
				))}
			</div>
		</>
	);
}

function Failure({ message }: { message: string }) {
	return (
		<div class="c-alert c-alert--danger" role="alert">
			{message}
		</div>
	);
}
