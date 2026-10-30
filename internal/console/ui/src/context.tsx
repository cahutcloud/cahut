// What every screen of a signed-in console shares: the sessions, the API
// bound to them, and the Rooms of the sidebar.
import { createContext, type ComponentChildren } from "preact";
import { useContext } from "preact/hooks";
import { ApiError, cahut as callCahut, server as callServer } from "./api";
import type { CahutOperations, LegalLinks, ServerOperations, Suspension } from "./api.gen";
import type { Session } from "./session";

export interface Console {
	/** The session the console acts with: that of the person, or else that of an administrator of the server alone. */
	session: Session;
	/**
	 * The session of an administrator of the server, alone or beside that of
	 * the person. See docs/adr/0043-the-console-holds-two-sessions.md.
	 */
	serverSession: Session | null;
	/** Runs an operation of the Cahut of this domain, with the session. */
	cahut<K extends keyof CahutOperations>(op: K, input: CahutOperations[K]["input"]): Promise<CahutOperations[K]["output"]>;
	/**
	 * Runs an operation of the Cahut of this domain that only the
	 * administrators of the server call, such as its kind and limits, with
	 * the session of the server.
	 */
	cahutAsServer<K extends keyof CahutOperations>(op: K, input: CahutOperations[K]["input"]): Promise<CahutOperations[K]["output"]>;
	/** Runs an operation of the server, with the session of its administrator. */
	server<K extends keyof ServerOperations>(op: K, input: ServerOperations[K]["input"]): Promise<ServerOperations[K]["output"]>;
	/** Ends the session of the server in this browser, and keeps that of the person. */
	leaveServer(): void;
	/** Whether the person administers this Cahut, or the server. */
	isAdmin: boolean;
	/**
	 * Whether this Cahut is individual: each Room keeps one member, and
	 * people do not see one another. See
	 * docs/adr/0030-kind-and-limits-of-each-cahut.md.
	 */
	individual: boolean;
	/**
	 * Whether this Cahut sends e-mail, through the relay that its
	 * administrators set. Without one, the console offers nothing that would
	 * send e-mail. See
	 * docs/adr/0034-e-mail-through-the-relay-of-each-cahut-only.md.
	 */
	sendsMail: boolean;
	/**
	 * The links to the legal pages of this Cahut, which the foot of every
	 * screen shows. See docs/adr/0042-the-legal-pages-of-a-cahut.md.
	 */
	legal?: LegalLinks;
	/**
	 * Since when and why the administrators of the server suspended this
	 * Cahut, which only they read: the console then acts with their session.
	 * See docs/adr/0048-suspending-a-cahut.md.
	 */
	suspension?: Suspension;
	/** Says that this Cahut is now suspended, or resumed, so that the console acts with the right session. */
	setSuspended(suspended: boolean): void;
	/** Reads this Cahut again, after a change of its kind, of its relay or of its legal pages. */
	reloadCahut(): void;
	/** Loads the Rooms of the sidebar again, after a change. */
	reloadRooms(): void;
	/** Bumps when the Rooms change, for the screens that show them. */
	roomsVersion: number;
	/** How many Knocks of the person's AIs wait for them. */
	knocks: number;
	/** Counts the Knocks that wait again, after an answer. */
	reloadKnocks(): void;
	/**
	 * How many reports about the Windows of this Cahut wait for its
	 * moderators, when the console acts as one. See
	 * docs/adr/0054-reporting-and-withdrawing-a-window.md.
	 */
	reports: number;
	/** Counts the reports that wait again, after a decision. */
	reloadReports(): void;
	/**
	 * Keeps the new name or handle of the person in the session of this
	 * browser, and loads the Rooms of the sidebar again: their personal Room
	 * follows them.
	 */
	renamed(person: { handle: string; name: string }): void;
	/** Keeps the new name of the administrator of the server in the session of this browser. */
	serverRenamed(admin: { name: string }): void;
}

const context = createContext<Console | null>(null);

export function useConsole(): Console {
	const c = useContext(context);
	if (!c) {
		throw new Error("useConsole outside a signed-in console");
	}
	return c;
}

/**
 * Binds the API to the sessions: that of the console, and that of the
 * server when the browser holds it. A session that has ended calls expired
 * with its kind. A call of the person that a suspended Cahut refuses calls
 * suspended: the console then shows the notice of a suspended Cahut, or
 * acts with the session of the server.
 */
export function bind(
	session: Session,
	serverSession: Session | null,
	expired: (kind: Session["kind"]) => void,
	suspended: () => void = () => undefined,
): Pick<Console, "session" | "serverSession" | "cahut" | "cahutAsServer" | "server"> {
	const watch = <T,>(p: Promise<T>, kind: Session["kind"]): Promise<T> =>
		p.catch((err: unknown) => {
			if (err instanceof ApiError && err.status === 401) {
				expired(kind);
			}
			if (err instanceof ApiError && err.code === "cahut_suspended" && kind === "person") {
				suspended();
			}
			throw err;
		});
	return {
		session,
		serverSession,
		cahut: (op, input) => watch(callCahut(op, input, session.token), session.kind),
		cahutAsServer: (op, input) => watch(callCahut(op, input, serverSession?.token), "admin"),
		server: (op, input) => watch(callServer(op, input, serverSession?.token), "admin"),
	};
}

export function ConsoleProvider({ value, children }: { value: Console; children: ComponentChildren }) {
	return <context.Provider value={value}>{children}</context.Provider>;
}
