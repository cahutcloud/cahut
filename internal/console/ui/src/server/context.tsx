// What every screen of the console of the server shares once signed in:
// the session of its administrator, and the API bound to it, for the server
// and for each Cahut. See docs/adr/0058-a-console-for-the-server.md.
import { createContext, type ComponentChildren } from "preact";
import { useContext } from "preact/hooks";
import { ApiError, type CahutCall, server as callServer, serverInCahut } from "../api";
import type { ServerOperations } from "../api.gen";
import type { Session } from "../session";

export interface ServerConsole {
	/** The session of the administrator of the server. */
	session: Session;
	/** Runs an operation of the server. */
	server<K extends keyof ServerOperations>(op: K, input: ServerOperations[K]["input"]): Promise<ServerOperations[K]["output"]>;
	/** Runs the operations that serve the server in the Cahut at domain. */
	cahutOf(domain: string): CahutCall;
	/** Keeps the new name of the administrator in the session of this browser. */
	renamed(admin: { name: string }): void;
	/** Signs out of the server, and keeps the session of a person, if the browser holds one. */
	signOut(): void;
}

const context = createContext<ServerConsole | null>(null);

export function useServer(): ServerConsole {
	const c = useContext(context);
	if (!c) {
		throw new Error("useServer outside a signed-in console of the server");
	}
	return c;
}

/**
 * Binds the API to the session of the administrator. A session that has
 * ended calls expired.
 */
export function bindServer(session: Session, expired: () => void): Pick<ServerConsole, "session" | "server" | "cahutOf"> {
	const watch = <T,>(p: Promise<T>): Promise<T> =>
		p.catch((err: unknown) => {
			if (err instanceof ApiError && err.status === 401) {
				expired();
			}
			throw err;
		});
	return {
		session,
		server: (op, input) => watch(callServer(op, input, session.token)),
		cahutOf: (domain) => (op, input) => watch(serverInCahut(domain, op, input, session.token)),
	};
}

export function ServerProvider({ value, children }: { value: ServerConsole; children: ComponentChildren }) {
	return <context.Provider value={value}>{children}</context.Provider>;
}
