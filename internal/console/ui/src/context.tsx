// What every screen of a signed-in console of the Cahut shares: the session
// of the person, the API bound to it, and the Rooms of the sidebar. The
// console of the server has its own: see server/context.tsx and
// docs/adr/0058-a-console-for-the-server.md.
import { createContext, type ComponentChildren } from "preact";
import { useContext } from "preact/hooks";
import { ApiError, cahut as callCahut } from "./api";
import type { CahutOperations, LegalLinks } from "./api.gen";
import type { Session } from "./session";

export interface Console {
	/** The session of the person. */
	session: Session;
	/** Runs an operation of the Cahut of this domain, with the session. */
	cahut<K extends keyof CahutOperations>(op: K, input: CahutOperations[K]["input"]): Promise<CahutOperations[K]["output"]>;
	/** Whether the person administers this Cahut. */
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
 * Binds the API to the session of the person. A session that has ended
 * calls expired. A call that a suspended Cahut refuses calls suspended: the
 * console then shows the notice of a suspended Cahut.
 */
export function bind(session: Session, expired: () => void, suspended: () => void = () => undefined): Pick<Console, "session" | "cahut"> {
	const watch = <T,>(p: Promise<T>): Promise<T> =>
		p.catch((err: unknown) => {
			if (err instanceof ApiError && err.status === 401) {
				expired();
			}
			if (err instanceof ApiError && err.code === "cahut_suspended") {
				suspended();
			}
			throw err;
		});
	return {
		session,
		cahut: (op, input) => watch(callCahut(op, input, session.token)),
	};
}

export function ConsoleProvider({ value, children }: { value: Console; children: ComponentChildren }) {
	return <context.Provider value={value}>{children}</context.Provider>;
}
