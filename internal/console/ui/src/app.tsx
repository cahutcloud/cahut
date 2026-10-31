// Routes the pages of the console, and keeps the session.
import { useEffect, useState } from "preact/hooks";
import { cahut, server } from "./api";
import type { LegalLinks, Room, Suspension } from "./api.gen";
import { PageHead } from "./components";
import { bind, ConsoleProvider } from "./context";
import { fragment } from "./fragment";
import { words } from "./i18n";
import { Access } from "./pages/access";
import { Account } from "./pages/account";
import { Activate } from "./pages/activate";
import { ConfirmEmail } from "./pages/confirm-email";
import { Consent } from "./pages/consent";
import { Invitation } from "./pages/invitation";
import { KnockPage, Knocks } from "./pages/knocks";
import { Moderation } from "./pages/moderation";
import { Logbook } from "./pages/logbook";
import { MailPage } from "./pages/mail";
import { Overview } from "./pages/overview";
import { People } from "./pages/people";
import { Read } from "./pages/read";
import { RoomPage, RoomsPage } from "./pages/rooms";
import { SearchPage } from "./pages/search";
import { Setup } from "./pages/setup";
import { Settings } from "./pages/settings";
import { SignIn } from "./pages/sign-in";
import { Suspended } from "./pages/suspended";
import { WindowPage } from "./pages/window";
import { Away, navigate, Redirect, usePath } from "./router";
import { clear, load, save, type Session } from "./session";
import { Shell } from "./shell";

const t = words({
	en: {
		toConnect: "to connect an AI",
		toRead: "to open its Windows",
		toKnock: "to answer a Knock",
		notFound: "Not found",
		nothingHere: "Nothing here",
		noPage: "This page does not exist.",
	},
	fr: {
		toConnect: "pour connecter une IA",
		toRead: "pour ouvrir ses Windows",
		toKnock: "pour répondre à un Knock",
		notFound: "Introuvable",
		nothingHere: "Rien ici",
		noPage: "Cette page n'existe pas.",
	},
});

/** Where someone signs in again with an account: as a person, or as an administrator of the server. */
function signInPage(kind: Session["kind"]): string {
	return kind === "admin" ? "/server/sign-in" : "/sign-in";
}

/** Ends a session on the server, whatever it says. */
function end(s: Session): Promise<unknown> {
	const call = s.kind === "admin" ? server("sign_out", {}, s.token) : cahut("sign_out", {}, s.token);
	return call.catch(() => undefined);
}

export function App() {
	const path = usePath();
	const [held, setHeld] = useState(load);
	// Whether this Cahut refused a call of the person because it is
	// suspended: the console then acts with the session of the server, or
	// shows the notice of a suspended Cahut. See
	// docs/adr/0048-suspending-a-cahut.md.
	const [suspended, setSuspended] = useState(false);
	// The console acts as the person when the browser holds their session,
	// and as the administrator of the server when it holds theirs alone. See
	// docs/adr/0043-the-console-holds-two-sessions.md.
	const session = (suspended ? null : held.person) ?? held.admin;

	// Keeps a session in its own place, beside the other, and ends the one
	// it replaces.
	const hold = (s: Session) => {
		const old = held[s.kind];
		if (old && old.token !== s.token) {
			end(old);
		}
		save(s);
		setHeld((h) => ({ ...h, [s.kind]: s }));
	};
	const drop = (kind: Session["kind"]) => {
		clear(kind);
		setHeld((h) => ({ ...h, [kind]: null }));
	};
	// An administrator of the server who signs in beside a person goes to
	// the pages of the server.
	const signIn = (s: Session) => {
		hold(s);
		navigate(s.kind === "admin" && held.person ? "/server/cahuts" : "/", true);
	};
	// A session has ended: forget it, and sign in again with its account.
	// The page of a Knock keeps its address, and signs in on the spot.
	const expired = (kind: Session["kind"]) => {
		drop(kind);
		if (!path.startsWith("/knocks/")) {
			navigate(signInPage(kind), true);
		}
	};
	// Each bar signs out of its own session, and the console stays open with
	// the other one, if the browser holds it. See
	// docs/adr/0044-two-bars-for-the-server-and-the-cahut.md.
	const signOutOf = async (kind: Session["kind"]) => {
		const s = held[kind];
		if (s) {
			await end(s); // signed out here, whatever the server says
		}
		drop(kind);
		navigate(held[kind === "person" ? "admin" : "person"] ? "/" : signInPage(kind), true);
	};
	// Signing out ends the session the console acts with: that of the
	// person, or that of the server when the browser holds it alone.
	const signOut = () => signOutOf(session?.kind ?? "person");
	const leaveServer = () => signOutOf("admin");

	// The pages of the server live in the console of the server, which an
	// older link or a bookmark reaches with its code. See
	// docs/adr/0058-a-console-for-the-server.md.
	if (path === "/server" || path.startsWith("/server/")) {
		return <Away to={`${path}${location.search}${fragment ? `#${fragment}` : ""}`} />;
	}
	switch (path) {
		case "/sign-in":
			return <SignIn kind="person" beside={held.admin !== null} onSignIn={signIn} />;
		case "/activate":
			return <Activate kind="person" onSignIn={signIn} />;
		case "/confirm-email":
			return <ConfirmEmail />;
		case "/invitation":
			return <Invitation />;
		case "/setup":
			return <Setup onSession={hold} />;
	}
	// The consent screen, the sign-in of readers and the page of a Knock
	// need the person: they sign in on the spot, beside the session of the
	// server if the browser holds it, and the address keeps the request.
	const stay = (s: Session | null) => (s ? hold(s) : drop("person"));
	// An AI client sends the person to the consent screen.
	if (path === "/oauth/authorize") {
		return held.person ? (
			<Consent key={held.person.token} session={held.person} onExpired={() => stay(null)} />
		) : (
			<SignIn kind="person" purpose={t.toConnect} onSignIn={stay} />
		);
	}
	// A reader of the Windows of this Cahut signs in here, then goes back to
	// the address of the Windows. The administrators of the server read no
	// Windows.
	if (path === "/read") {
		return held.person ? (
			<Read key={held.person.token} session={held.person} onExpired={() => stay(null)} />
		) : (
			<SignIn kind="person" purpose={t.toRead} onSignIn={stay} />
		);
	}
	// An AI gives its person the link of its Knock.
	if (!held.person && path.startsWith("/knocks/")) {
		return <SignIn kind="person" purpose={t.toKnock} onSignIn={stay} />;
	}
	if (!session && suspended && held.person) {
		return <Suspended onSignOut={() => signOutOf("person")} />;
	}
	if (!session) {
		return <Redirect to="/sign-in" />;
	}
	return (
		<SignedIn
			key={`${session.token} ${held.admin?.token ?? ""}`}
			session={session}
			serverSession={held.admin}
			path={path}
			onSession={hold}
			onExpired={expired}
			onSuspended={setSuspended}
			onSignOut={signOut}
			onLeaveServer={leaveServer}
		/>
	);
}

function SignedIn(props: {
	session: Session;
	serverSession: Session | null;
	path: string;
	onSession: (s: Session) => void;
	onExpired: (kind: Session["kind"]) => void;
	onSuspended: (suspended: boolean) => void;
	onSignOut: () => void;
	onLeaveServer: () => void;
}) {
	const { session, path } = props;
	const api = bind(session, props.serverSession, props.onExpired, () => props.onSuspended(true));
	const [rooms, setRooms] = useState<Room[]>([]);
	const [roomsVersion, setRoomsVersion] = useState(0);
	const [isAdmin, setAdmin] = useState(session.kind === "admin");
	const [individual, setIndividual] = useState(false);
	const [sendsMail, setSendsMail] = useState(false);
	const [legal, setLegal] = useState<LegalLinks>();
	const [suspension, setSuspension] = useState<Suspension>();
	const [cahutVersion, setCahutVersion] = useState(0);
	useEffect(() => {
		api.cahut("get_cahut", {}).then(
			(a) => (setIndividual(a.kind === "individual"), setSendsMail(a.sends_mail), setLegal(a.legal), setSuspension(a.suspension)),
			() => undefined,
		);
	}, [cahutVersion]);
	useEffect(() => {
		if (session.kind === "person") {
			api.cahut("list_rooms", {}).then((r) => setRooms(r.rooms), () => undefined);
		}
	}, [roomsVersion]);
	// The person may have changed their name or handle in another browser.
	const renamed = (p: { handle: string; name: string }) => {
		if (p.handle !== session.handle || p.name !== session.name) {
			props.onSession({ ...session, handle: p.handle, name: p.name });
		}
	};
	useEffect(() => {
		if (session.kind === "person") {
			api.cahut("get_account", {}).then((a) => (setAdmin(a.role === "admin"), renamed(a)), () => undefined);
		}
	}, []);
	const knocks = useWaitingKnocks(api, session);
	const reports = useWaitingReports(api, session.kind !== "person" || isAdmin || Boolean(props.serverSession));

	const value = {
		...api,
		leaveServer: props.onLeaveServer,
		isAdmin,
		individual,
		sendsMail,
		legal,
		suspension,
		setSuspended: (on: boolean) => {
			props.onSuspended(on);
			setCahutVersion((n) => n + 1);
		},
		reloadCahut: () => setCahutVersion((n) => n + 1),
		roomsVersion,
		reloadRooms: () => setRoomsVersion((n) => n + 1),
		knocks: knocks.count,
		reloadKnocks: knocks.reload,
		reports: reports.count,
		reloadReports: reports.reload,
		renamed: (p: { handle: string; name: string }) => {
			renamed(p);
			setRoomsVersion((n) => n + 1);
		},
		serverRenamed: (a: { name: string }) => {
			const s = props.serverSession;
			if (s && a.name !== s.name) {
				props.onSession({ ...s, name: a.name });
			}
		},
	};
	return (
		<ConsoleProvider value={value}>
			<Shell path={path} rooms={rooms} onSignOut={props.onSignOut}>
				<Page path={path} onSignOut={props.onSignOut} />
			</Shell>
		</ConsoleProvider>
	);
}

/** How often the console asks again how many Knocks wait. */
const knockEvery = 30_000;

/**
 * Counts the Knocks that wait for the person, every 30 seconds and when the
 * tab comes back to the front, and says it in the title of the tab.
 */
function useWaitingKnocks(api: ReturnType<typeof bind>, session: Session): { count: number; reload: () => void } {
	const [count, setCount] = useState(0);
	const [round, setRound] = useState(0);
	useEffect(() => {
		if (session.kind !== "person") return;
		const ask = () => {
			if (document.visibilityState === "visible") {
				api.cahut("list_knocks", { waiting: true, limit: 1 }).then((l) => setCount(l.waiting), () => undefined);
			}
		};
		ask();
		const timer = setInterval(ask, knockEvery);
		document.addEventListener("visibilitychange", ask);
		return () => {
			clearInterval(timer);
			document.removeEventListener("visibilitychange", ask);
		};
	}, [round]);
	useEffect(() => {
		document.title = count > 0 ? `(${count}) Cahut` : "Cahut";
	}, [count]);
	return { count, reload: () => setRound((n) => n + 1) };
}

/**
 * Counts the reports that wait for the moderators of this Cahut, as the
 * Knocks are counted, when the console acts as one of them: as an
 * administrator of the Cahut, or with the session of the server.
 */
function useWaitingReports(api: ReturnType<typeof bind>, moderates: boolean): { count: number; reload: () => void } {
	const [count, setCount] = useState(0);
	const [round, setRound] = useState(0);
	useEffect(() => {
		if (!moderates) return;
		const call = api.session.kind === "person" && api.serverSession ? api.cahutAsServer : api.cahut;
		const ask = () => {
			if (document.visibilityState === "visible") {
				call("list_reports", { waiting: true, limit: 1 }).then((l) => setCount(l.waiting), () => undefined);
			}
		};
		ask();
		const timer = setInterval(ask, knockEvery);
		document.addEventListener("visibilitychange", ask);
		return () => {
			clearInterval(timer);
			document.removeEventListener("visibilitychange", ask);
		};
	}, [round, moderates]);
	return { count, reload: () => setRound((n) => n + 1) };
}

function Page({ path, onSignOut }: { path: string; onSignOut: () => void }) {
	const knock = path.match(/^\/knocks\/([0-9]+)\/?$/);
	if (knock) return <KnockPage key={path} id={Number(knock[1])} />;
	const room = path.match(/^\/rooms\/([a-z0-9-]+)\/?$/);
	if (room) return <RoomPage name={room[1]} />;
	const win = path.match(/^\/rooms\/([a-z0-9-]+)\/([a-z0-9-]+)\/?$/);
	if (win) return <WindowPage key={path} room={win[1]} name={win[2]} />;
	switch (path) {
		case "/":
			return <Overview />;
		case "/rooms":
			return <RoomsPage />;
		case "/search":
			return <SearchPage />;
		case "/logbook":
			return <Logbook />;
		case "/knocks":
			return <Knocks />;
		case "/people":
			return <People />;
		case "/settings":
			return <Settings />;
		case "/mail":
			return <MailPage />;
		case "/moderation":
			return <Moderation />;
		case "/access":
			return <Access />;
		case "/account":
			return <Account onSignOut={onSignOut} />;
	}
	return (
		<PageHead label={t.notFound} title={t.nothingHere}>
			{t.noPage}
		</PageHead>
	);
}
