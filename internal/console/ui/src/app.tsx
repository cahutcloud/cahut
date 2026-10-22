// Routes the pages of the console, and keeps the session.
import { useState } from "preact/hooks";
import { cahut, server } from "./api";
import { Activate } from "./pages/activate";
import { Overview } from "./pages/overview";
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
	return (
		<Shell session={session} path={path} onSignOut={signOut}>
			{path === "/" ? (
				<Overview session={session} onExpired={expired} />
			) : (
				<div class="page-head">
					<h1 class="c-h2">Nothing here</h1>
					<p class="c-muted">This page does not exist.</p>
				</div>
			)}
		</Shell>
	);
}
