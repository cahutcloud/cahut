// Calls the API of the consoles: that of the Cahut, under /console/api/,
// and that of the server, under /server/api/. Every operation and its types
// come from api.gen.ts, which make console-types writes from the Go code: a
// call to an operation that does not exist, or with the wrong input, fails
// the build. See docs/adr/0014-console-http-api.md and
// docs/adr/0058-a-console-for-the-server.md.
import type { CahutOperations, ServerCahutOperations, ServerOperations } from "./api.gen";
import { words } from "./i18n";
import { say } from "./refusals";

const t = words({
	en: { noAnswer: "Cahut does not answer. Check your connection, then try again." },
	fr: { noAnswer: "Cahut ne répond pas. Vérifiez votre connexion, puis réessayez." },
});

/**
 * A call that the API refused, with its status and its message, and the
 * code and values of a refusal that the console says in its language.
 */
export class ApiError extends Error {
	status: number;
	code?: string;
	values?: unknown;
	constructor(status: number, message: string, code?: string, values?: unknown) {
		super(message);
		this.status = status;
		this.code = code;
		this.values = values;
	}
}

async function post(path: string, input: unknown, token?: string): Promise<unknown> {
	const headers: Record<string, string> = { "Content-Type": "application/json" };
	if (token) {
		headers.Authorization = `Bearer ${token}`;
	}
	const res = await fetch(path, { method: "POST", headers, body: JSON.stringify(input) });
	const body = (await res.json().catch(() => ({}))) as { error?: string; code?: string; values?: unknown };
	if (!res.ok) {
		throw new ApiError(res.status, body.error ?? res.statusText, body.code, body.values);
	}
	return body;
}

/** Runs an operation of the Cahut of this domain. */
export function cahut<K extends keyof CahutOperations>(
	op: K,
	input: CahutOperations[K]["input"],
	token?: string,
): Promise<CahutOperations[K]["output"]> {
	return post(`/console/api/${op}`, input, token) as Promise<CahutOperations[K]["output"]>;
}

/** Runs an operation of the server, for its administrators. */
export function server<K extends keyof ServerOperations>(
	op: K,
	input: ServerOperations[K]["input"],
	token?: string,
): Promise<ServerOperations[K]["output"]> {
	return post(`/server/api/${op}`, input, token) as Promise<ServerOperations[K]["output"]>;
}

/**
 * Runs, for an administrator of the server, an operation that serves the
 * server in the Cahut at domain, whichever domain they signed in at.
 */
export function serverInCahut<K extends keyof ServerCahutOperations>(
	domain: string,
	op: K,
	input: ServerCahutOperations[K]["input"],
	token?: string,
): Promise<ServerCahutOperations[K]["output"]> {
	return post(`/server/api/cahuts/${encodeURIComponent(domain)}/${op}`, input, token) as Promise<ServerCahutOperations[K]["output"]>;
}

/** Runs an operation that serves the server in one Cahut, bound to it and to a session. */
export type CahutCall = <K extends keyof ServerCahutOperations>(
	op: K,
	input: ServerCahutOperations[K]["input"],
) => Promise<ServerCahutOperations[K]["output"]>;

// The prefixes that the registry puts before the message of an error, which
// the console does not show.
const prefixes = ["invalid input: ", "not allowed: ", "not found: ", "too many attempts: ", "needs the approval of a person: "];

/**
 * Says what went wrong, in a sentence for the person who reads it: that of
 * the refusal's code, in the language of the console, or else the message
 * of the API, which the server writes in English, for AIs and people alike.
 * See docs/adr/0037-refusals-with-a-code-for-the-console.md.
 */
export function explain(err: unknown): string {
	if (!(err instanceof ApiError)) {
		return t.noAnswer;
	}
	const said = err.code === undefined ? undefined : say(err.code, err.values);
	if (said !== undefined) {
		return said;
	}
	let msg = err.message;
	for (const p of prefixes) {
		msg = msg.startsWith(p) ? msg.slice(p.length) : msg;
	}
	msg = msg.charAt(0).toUpperCase() + msg.slice(1);
	return msg.endsWith(".") ? msg : `${msg}.`;
}
