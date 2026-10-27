// Routes the pages of the console, and keeps the session.
import { useEffect, useState } from "preact/hooks";
import { cahut, server } from "./api";
import type { LegalLinks, Room } from "./api.gen";
import { PageHead } from "./components";
import { bind, ConsoleProvider, useConsole } from "./context";
import { words } from "./i18n";
import { Access } from "./pages/access";
import { Account } from "./pages/account";
import { Activate } from "./pages/activate";
import { Admins } from "./pages/admins";
import { ConfirmEmail } from "./pages/confirm-email";
import { Consent } from "./pages/consent";
import { HouseStylePage } from "./pages/house-style";
import { Invitation } from "./pages/invitation";
import { KnockPage, Knocks } from "./pages/knocks";
import { Logbook } from "./pages/logbook";
import { MailPage } from "./pages/mail";
import { Overview, ServerOverview } from "./pages/overview";
import { People } from "./pages/people";
import { Read } from "./pages/read";
import { RoomPage, RoomsPage } from "./pages/rooms";
import { Setup } from "./pages/setup";
import { Settings } from "./pages/settings";
import { SignIn } from "./pages/sign-in";
import { Updates } from "./pages/updates";
import { WindowPage } from "./pages/window";
import { navigate, Redirect, usePath } from "./router";
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
	// The console acts as the person when the browser holds their session,
	// and as the administrator of the server when it holds theirs alone. See
	// docs/adr/0043-the-console-holds-two-sessions.md.
	const session = held.person ?? held.admin;

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

	switch (path) {
		case "/sign-in":
			return <SignIn kind="person" beside={held.admin !== null} onSignIn={signIn} />;
		case "/server/sign-in":
			return <SignIn kind="admin" beside={held.person !== null} onSignIn={signIn} />;
		case "/activate":
			return <Activate kind="person" onSignIn={signIn} />;
		case "/server/activate":
			return <Activate kind="admin" onSignIn={signIn} />;
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
	onSignOut: () => void;
	onLeaveServer: () => void;
}) {
	const { session, path } = props;
	const api = bind(session, props.serverSession, props.onExpired);
	const [rooms, setRooms] = useState<Room[]>([]);
	const [roomsVersion, setRoomsVersion] = useState(0);
	const [isAdmin, setAdmin] = useState(session.kind === "admin");
	const [individual, setIndividual] = useState(false);
	const [sendsMail, setSendsMail] = useState(false);
	const [legal, setLegal] = useState<LegalLinks>();
	const [cahutVersion, setCahutVersion] = useState(0);
	useEffect(() => {
		api.cahut("get_cahut", {}).then(
			(a) => (setIndividual(a.kind === "individual"), setSendsMail(a.sends_mail), setLegal(a.legal)),
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

	const value = {
		...api,
		leaveServer: props.onLeaveServer,
		isAdmin,
		individual,
		sendsMail,
		legal,
		reloadCahut: () => setCahutVersion((n) => n + 1),
		roomsVersion,
		reloadRooms: () => setRoomsVersion((n) => n + 1),
		knocks: knocks.count,
		reloadKnocks: knocks.reload,
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

function Page({ path, onSignOut }: { path: string; onSignOut: () => void }) {
	const { serverSession } = useConsole();
	// The pages of the server need its session, which the person adds from
	// the page of their account.
	if (path.startsWith("/server/") && !serverSession) {
		return <Redirect to="/server/sign-in" />;
	}
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
		case "/logbook":
			return <Logbook of="cahut" />;
		case "/knocks":
			return <Knocks />;
		case "/people":
			return <People />;
		case "/house-style":
			return <HouseStylePage />;
		case "/settings":
			return <Settings />;
		case "/mail":
			return <MailPage />;
		case "/access":
			return <Access />;
		case "/account":
			return <Account onSignOut={onSignOut} />;
		case "/server/cahuts":
			return <ServerOverview />;
		case "/server/admins":
			return <Admins />;
		case "/server/logbook":
			return <Logbook of="server" />;
		case "/server/updates":
			return <Updates />;
	}
	return (
		<PageHead label={t.notFound} title={t.nothingHere}>
			{t.noPage}
		</PageHead>
	);
}
