// The session of whoever signed in to this console. It lives in the storage
// of the console's own origin, and goes to the API in the Authorization
// header: the API reads no cookie. See docs/adr/0011-console-on-its-own-port.md.

export interface Session {
	/** A person of this Cahut, or an administrator of the server. */
	kind: "person" | "admin";
	token: string;
	handle: string;
	name: string;
	/** When the session ends at the latest, in RFC 3339. */
	expiresAt: string;
}

const key = "cahut.session";

/** Returns the session kept in this browser, if it has not ended. */
export function load(): Session | null {
	try {
		const s = JSON.parse(localStorage.getItem(key) ?? "null") as Session | null;
		if (s && Date.parse(s.expiresAt) > Date.now()) {
			return s;
		}
	} catch {
		// A session that cannot be read is no session.
	}
	clear();
	return null;
}

export function save(s: Session): void {
	localStorage.setItem(key, JSON.stringify(s));
}

export function clear(): void {
	localStorage.removeItem(key);
}
