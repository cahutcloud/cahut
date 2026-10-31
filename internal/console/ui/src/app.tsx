// Routes the pages of the console, and keeps the session.
import { useEffect, useState } from "preact/hooks";
import { cahut, server } from "./api";
import type { LegalLinks, Room } from "./api.gen";
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

/** Ends a session on the server, whatever it says. */
function end(s: Session): Promise<unknown> {
	const call = s.kind === "admin" ? server("sign_out", {}, s.token) : cahut("sign_out", {}, s.token);
	return call.catch(() => undefined);
}

export function App() {
	const path = usePath();
	// The console of the Cahut acts with the session of a person only; that
	// of an administrator of the server, which the first start may keep in
	// this browser too, is for the console of the server. See
	// docs/adr/0058-a-console-for-the-server.md.
	const [held, setHeld] = useState(load);
	// Whether this Cahut refused a call of the person because it is
	// suspended: the console then shows the notice of a suspended Cahut. See
	// docs/adr/0048-suspending-a-cahut.md.
	const [suspended, setSuspended] = useState(false);
	const session = suspended ? null : held.person;

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
	const drop = () => {
		clear("person");
		setHeld((h) => ({ ...h, person: null }));
	};
	const signIn = (s: Session) => {
		hold(s);
		navigate("/", true);
	};
	// The session has ended: forget it, and sign in again. The page of a
	// Knock keeps its address, and signs in on the spot.
	const expired = () => {
		drop();
		if (!path.startsWith("/knocks/")) {
			navigate("/sign-in", true);
		}
	};
	// Signing out ends the session of the person, and leaves that of the
	// server, if the browser holds one, to the console of the server.
	const signOut = async () => {
		if (held.person) {
			await end(held.person); // signed out here, whatever the server says
		}
		drop();
		navigate("/sign-in", true);
	};

	// The pages of the server live in the console of the server, which an
	// older link or a bookmark reaches with its code. See
	// docs/adr/0058-a-console-for-the-server.md.
	if (path === "/server" || path.startsWith("/server/")) {
		return <Away to={`${path}${location.search}${fragment ? `#${fragment}` : ""}`} />;
	}
	switch (path) {
		case "/sign-in":
			return <SignIn kind="person" onSignIn={signIn} />;
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
	// need the person: they sign in on the spot, and the address keeps the
	// request.
	const stay = (s: Session | null) => (s ? hold(s) : drop());
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
	if (suspended && held.person) {
		return <Suspended onSignOut={signOut} />;
	}
	if (!session) {
		return <Redirect to="/sign-in" />;
	}
	return (
		<SignedIn
			key={session.token}
			session={session}
			path={path}
			onSession={hold}
			onExpired={expired}
			onSuspended={() => setSuspended(true)}
			onSignOut={signOut}
		/>
	);
}

function SignedIn(props: {
	session: Session;
	path: string;
	onSession: (s: Session) => void;
	onExpired: () => void;
	onSuspended: () => void;
	onSignOut: () => void;
}) {
	const { session, path } = props;
	const api = bind(session, props.onExpired, props.onSuspended);
	const [rooms, setRooms] = useState<Room[]>([]);
	const [roomsVersion, setRoomsVersion] = useState(0);
	const [isAdmin, setAdmin] = useState(false);
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
		api.cahut("list_rooms", {}).then((r) => setRooms(r.rooms), () => undefined);
	}, [roomsVersion]);
	// The person may have changed their name or handle in another browser.
	const renamed = (p: { handle: string; name: string }) => {
		if (p.handle !== session.handle || p.name !== session.name) {
			props.onSession({ ...session, handle: p.handle, name: p.name });
		}
	};
	useEffect(() => {
		api.cahut("get_account", {}).then((a) => (setAdmin(a.role === "admin"), renamed(a)), () => undefined);
	}, []);
	const knocks = useWaitingKnocks(api);
	const reports = useWaitingReports(api, isAdmin);

	const value = {
		...api,
		isAdmin,
		individual,
		sendsMail,
		legal,
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
function useWaitingKnocks(api: ReturnType<typeof bind>): { count: number; reload: () => void } {
	const [count, setCount] = useState(0);
	const [round, setRound] = useState(0);
	useEffect(() => {
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
 * Knocks are counted, when the person administers it.
 */
function useWaitingReports(api: ReturnType<typeof bind>, moderates: boolean): { count: number; reload: () => void } {
	const [count, setCount] = useState(0);
	const [round, setRound] = useState(0);
	useEffect(() => {
		if (!moderates) return;
		const ask = () => {
			if (document.visibilityState === "visible") {
				api.cahut("list_reports", { waiting: true, limit: 1 }).then((l) => setCount(l.waiting), () => undefined);
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
