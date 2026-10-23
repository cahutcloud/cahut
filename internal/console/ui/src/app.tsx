// Routes the pages of the console, and keeps the session.
import { useEffect, useState } from "preact/hooks";
import { cahut, server } from "./api";
import type { Room } from "./api.gen";
import { PageHead } from "./components";
import { bind, ConsoleProvider } from "./context";
import { Access } from "./pages/access";
import { Account } from "./pages/account";
import { Activate } from "./pages/activate";
import { Admins } from "./pages/admins";
import { Consent } from "./pages/consent";
import { HouseStylePage } from "./pages/house-style";
import { KnockPage, Knocks } from "./pages/knocks";
import { Logbook } from "./pages/logbook";
import { Overview } from "./pages/overview";
import { People } from "./pages/people";
import { Read } from "./pages/read";
import { RoomPage, RoomsPage } from "./pages/rooms";
import { SignIn } from "./pages/sign-in";
import { WindowPage } from "./pages/window";
import { navigate, Redirect, usePath } from "./router";
import { clear, load, save, type Session } from "./session";
import { Shell } from "./shell";

/** Where someone signs in again: as a person, or as an administrator. */
function signInPage(s: Session | null): string {
	return s?.kind === "admin" ? "/server/sign-in" : "/sign-in";
}

export function App() {
	const path = usePath();
	const [session, setSession] = useState(load);

	const signIn = (s: Session) => {
		save(s);
		setSession(s);
		navigate("/", true);
	};
	// The session has ended: forget it, and sign in again. The page of a
	// Knock keeps its address, and signs in on the spot.
	const expired = () => {
		clear();
		setSession(null);
		if (!path.startsWith("/knocks/")) {
			navigate(signInPage(session), true);
		}
	};
	const signOut = async () => {
		if (session) {
			const end = session.kind === "admin" ? server("sign_out", {}, session.token) : cahut("sign_out", {}, session.token);
			await end.catch(() => undefined); // signed out here, whatever the server says
		}
		expired();
	};

	switch (path) {
		case "/sign-in":
			return <SignIn kind="person" onSignIn={signIn} />;
		case "/server/sign-in":
			return <SignIn kind="admin" onSignIn={signIn} />;
		case "/activate":
			return <Activate kind="person" onSignIn={signIn} />;
		case "/server/activate":
			return <Activate kind="admin" onSignIn={signIn} />;
	}
	// An AI client sends the person to the consent screen: they sign in on
	// the spot, and the address keeps the request.
	if (path === "/oauth/authorize") {
		const stay = (s: Session | null) => {
			if (s) save(s);
			else clear();
			setSession(s);
		};
		return session?.kind === "person" ? (
			<Consent key={session.token} session={session} onExpired={() => stay(null)} />
		) : (
			<SignIn kind="person" purpose="to connect an AI" onSignIn={stay} />
		);
	}
	// A reader of the Windows of this Cahut signs in here, then goes back to
	// the address of the Windows.
	if (path === "/read") {
		const stay = (s: Session | null) => {
			if (s) save(s);
			else clear();
			setSession(s);
		};
		return session ? (
			<Read key={session.token} session={session} onExpired={() => stay(null)} />
		) : (
			<SignIn kind="person" purpose="to open its Windows" onSignIn={stay} />
		);
	}
	// An AI gives its person the link of its Knock: they sign in on the
	// spot, and the address keeps the Knock.
	if (!session && path.startsWith("/knocks/")) {
		const stay = (s: Session) => {
			save(s);
			setSession(s);
		};
		return <SignIn kind="person" purpose="to answer a Knock" onSignIn={stay} />;
	}
	if (!session) {
		return <Redirect to="/sign-in" />;
	}
	return <SignedIn key={session.token} session={session} path={path} onExpired={expired} onSignOut={signOut} />;
}

function SignedIn(props: { session: Session; path: string; onExpired: () => void; onSignOut: () => void }) {
	const { session, path } = props;
	const api = bind(session, props.onExpired);
	const [rooms, setRooms] = useState<Room[]>([]);
	const [roomsVersion, setRoomsVersion] = useState(0);
	const [isAdmin, setAdmin] = useState(session.kind === "admin");
	useEffect(() => {
		if (session.kind === "person") {
			api.cahut("list_rooms", {}).then((r) => setRooms(r.rooms), () => undefined);
		}
	}, [roomsVersion]);
	useEffect(() => {
		if (session.kind === "person") {
			api.cahut("get_account", {}).then((a) => setAdmin(a.role === "admin"), () => undefined);
		}
	}, []);
	const knocks = useWaitingKnocks(api, session);

	const value = {
		...api,
		isAdmin,
		roomsVersion,
		reloadRooms: () => setRoomsVersion((n) => n + 1),
		knocks: knocks.count,
		reloadKnocks: knocks.reload,
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
		case "/access":
			return <Access />;
		case "/account":
			return <Account onSignOut={onSignOut} />;
		case "/server/admins":
			return <Admins />;
		case "/server/logbook":
			return <Logbook of="server" />;
	}
	return <PageHead label="Not found" title="Nothing here">This page does not exist.</PageHead>;
}
