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
import { Logbook } from "./pages/logbook";
import { Overview } from "./pages/overview";
import { People } from "./pages/people";
import { RoomPage, RoomsPage } from "./pages/rooms";
import { SignIn } from "./pages/sign-in";
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
	// The session has ended: forget it, and sign in again.
	const expired = () => {
		clear();
		setSession(null);
		navigate(signInPage(session), true);
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

	const value = { ...api, isAdmin, roomsVersion, reloadRooms: () => setRoomsVersion((n) => n + 1) };
	return (
		<ConsoleProvider value={value}>
			<Shell path={path} rooms={rooms} onSignOut={props.onSignOut}>
				<Page path={path} onSignOut={props.onSignOut} />
			</Shell>
		</ConsoleProvider>
	);
}

function Page({ path, onSignOut }: { path: string; onSignOut: () => void }) {
	const room = path.match(/^\/rooms\/([a-z0-9-]+)\/?$/);
	if (room) return <RoomPage name={room[1]} />;
	switch (path) {
		case "/":
			return <Overview />;
		case "/rooms":
			return <RoomsPage />;
		case "/logbook":
			return <Logbook of="cahut" />;
		case "/people":
			return <People />;
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
