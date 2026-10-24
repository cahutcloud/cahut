// What every screen of a signed-in console shares: the session, the API
// bound to it, and the Rooms of the sidebar.
import { createContext, type ComponentChildren } from "preact";
import { useContext } from "preact/hooks";
import { ApiError, cahut as callCahut, server as callServer } from "./api";
import type { CahutOperations, ServerOperations } from "./api.gen";
import type { Session } from "./session";

export interface Console {
	session: Session;
	/** Runs an operation of the Cahut of this domain, with the session. */
	cahut<K extends keyof CahutOperations>(op: K, input: CahutOperations[K]["input"]): Promise<CahutOperations[K]["output"]>;
	/** Runs an operation of the server, with the session of its administrator. */
	server<K extends keyof ServerOperations>(op: K, input: ServerOperations[K]["input"]): Promise<ServerOperations[K]["output"]>;
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
	/** Reads this Cahut again, after a change of its kind or of its relay. */
	reloadCahut(): void;
	/** Loads the Rooms of the sidebar again, after a change. */
	reloadRooms(): void;
	/** Bumps when the Rooms change, for the screens that show them. */
	roomsVersion: number;
	/** How many Knocks of the person's AIs wait for them. */
	knocks: number;
	/** Counts the Knocks that wait again, after an answer. */
	reloadKnocks(): void;
}

const context = createContext<Console | null>(null);

export function useConsole(): Console {
	const c = useContext(context);
	if (!c) {
		throw new Error("useConsole outside a signed-in console");
	}
	return c;
}

/** Binds the API to session: a session that has ended calls expired. */
export function bind(session: Session, expired: () => void): Pick<Console, "session" | "cahut" | "server"> {
	const watch = <T,>(p: Promise<T>): Promise<T> =>
		p.catch((err: unknown) => {
			if (err instanceof ApiError && err.status === 401) {
				expired();
			}
			throw err;
		});
	return {
		session,
		cahut: (op, input) => watch(callCahut(op, input, session.token)),
		server: (op, input) => watch(callServer(op, input, session.token)),
	};
}

export function ConsoleProvider({ value, children }: { value: Console; children: ComponentChildren }) {
	return <context.Provider value={value}>{children}</context.Provider>;
}
