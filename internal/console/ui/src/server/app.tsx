// Routes the pages of the console of the server, and keeps the session of
// its administrator, beside that of a person of the Cahut, which it leaves
// as it is. See docs/adr/0058-a-console-for-the-server.md.
import { useState } from "preact/hooks";
import { server } from "../api";
import { PageHead } from "../components";
import { words } from "../i18n";
import { Activate } from "../pages/activate";
import { Admins } from "../pages/admins";
import { ServerLogbook } from "../pages/logbook";
import { SignIn } from "../pages/sign-in";
import { Updates } from "../pages/updates";
import { navigate, Redirect, usePath } from "../router";
import { clear, load, save, type Session } from "../session";
import { ServerAccount } from "./account";
import { CahutPage } from "./cahut";
import { Cahuts } from "./cahuts";
import { bindServer, ServerProvider } from "./context";
import { ServerShell } from "./shell";

const t = words({
	en: { notFound: "Not found", nothingHere: "Nothing here", noPage: "This page does not exist." },
	fr: { notFound: "Introuvable", nothingHere: "Rien ici", noPage: "Cette page n'existe pas." },
});

/** Ends a session of the server on the server, whatever it says. */
function end(s: Session): Promise<unknown> {
	return server("sign_out", {}, s.token).catch(() => undefined);
}

export function ServerApp() {
	const path = usePath();
	const [session, setSession] = useState(() => load().admin);
	// Keeps the session, and ends the one it replaces.
	const hold = (s: Session) => {
		if (session && session.token !== s.token) {
			end(session);
		}
		save(s);
		setSession(s);
	};
	const drop = () => {
		clear("admin");
		setSession(null);
	};
	const signIn = (s: Session) => {
		hold(s);
		navigate("/cahuts", true);
	};
	const expired = () => {
		drop();
		navigate("/sign-in", true);
	};
	const signOut = async () => {
		if (session) {
			await end(session); // signed out here, whatever the server says
		}
		drop();
		navigate("/sign-in", true);
	};

	switch (path) {
		case "/sign-in":
			return <SignIn kind="admin" onSignIn={signIn} />;
		case "/activate":
			return <Activate kind="admin" onSignIn={signIn} />;
	}
	if (!session) {
		return <Redirect to="/sign-in" />;
	}
	const value = {
		...bindServer(session, expired),
		renamed: (a: { name: string }) => {
			if (a.name !== session.name) hold({ ...session, name: a.name });
		},
		signOut,
	};
	return (
		<ServerProvider key={session.token} value={value}>
			<ServerShell path={path}>
				<Page path={path} />
			</ServerShell>
		</ServerProvider>
	);
}

function Page({ path }: { path: string }) {
	const cahut = path.match(/^\/cahuts\/([^/]+)\/?$/);
	if (cahut) return <CahutPage key={cahut[1]} domain={decodeURIComponent(cahut[1])} />;
	switch (path) {
		case "/":
			return <Redirect to="/cahuts" />;
		case "/cahuts":
			return <Cahuts />;
		case "/admins":
			return <Admins />;
		case "/logbook":
			return <ServerLogbook />;
		case "/updates":
			return <Updates />;
		case "/account":
			return <ServerAccount />;
	}
	return (
		<PageHead label={t.notFound} title={t.nothingHere}>
			{t.noPage}
		</PageHead>
	);
}
