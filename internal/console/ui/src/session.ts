// The sessions of whoever signed in to this console: that of a person of
// this Cahut, that of an administrator of the server, or both, each in its
// own place. They live in the storage of the Cahut's origin, which the
// sandboxed Windows cannot read, and go to the API in the Authorization
// header: the API reads no cookie. See
// docs/adr/0011-console-on-its-own-port.md,
// docs/adr/0020-console-under-a-path-of-port-443.md and
// docs/adr/0043-the-console-holds-two-sessions.md.

export interface Session {
	/** A person of this Cahut, or an administrator of the server. */
	kind: "person" | "admin";
	token: string;
	handle: string;
	name: string;
	/** When the session ends at the latest, in RFC 3339. */
	expiresAt: string;
}

/** The sessions that this browser holds, one in each place. */
export interface Held {
	person: Session | null;
	admin: Session | null;
}

// Until ADR 43, the one session of the console, of either kind, lived under
// the key of the person.
const keys = { person: "cahut.session", admin: "cahut.server-session" };

function read(key: string): Session | null {
	try {
		const s = JSON.parse(localStorage.getItem(key) ?? "null") as Session | null;
		if (s && Date.parse(s.expiresAt) > Date.now()) {
			return s;
		}
	} catch {
		// A session that cannot be read is no session.
	}
	localStorage.removeItem(key);
	return null;
}

/** Returns the sessions kept in this browser that have not ended. */
export function load(): Held {
	const first = read(keys.person);
	if (first?.kind === "admin") {
		localStorage.removeItem(keys.person);
		save(first);
		return { person: null, admin: first };
	}
	return { person: first, admin: read(keys.admin) };
}

/** Keeps a session in the place of its kind. */
export function save(s: Session): void {
	localStorage.setItem(keys[s.kind], JSON.stringify(s));
}

/** Forgets the session of a kind. */
export function clear(kind: Session["kind"]): void {
	localStorage.removeItem(keys[kind]);
}
