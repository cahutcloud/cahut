// Calls the API of the console, under the console's path. Every
// operation and its types come from api.gen.ts, which make console-types
// writes from the Go code: a call to an operation that does not exist, or
// with the wrong input, fails the build. See docs/adr/0014-console-http-api.md.
import type { CahutOperations, ServerOperations } from "./api.gen";
import { words } from "./i18n";
import { base } from "./router";

const t = words({
	en: { noAnswer: "Cahut does not answer. Check your connection, then try again." },
	fr: { noAnswer: "Cahut ne répond pas. Vérifiez votre connexion, puis réessayez." },
});

/** A call that the API refused, with its status and its message. */
export class ApiError extends Error {
	status: number;
	constructor(status: number, message: string) {
		super(message);
		this.status = status;
	}
}

async function post(path: string, input: unknown, token?: string): Promise<unknown> {
	const headers: Record<string, string> = { "Content-Type": "application/json" };
	if (token) {
		headers.Authorization = `Bearer ${token}`;
	}
	const res = await fetch(path, { method: "POST", headers, body: JSON.stringify(input) });
	const body = (await res.json().catch(() => ({}))) as { error?: string };
	if (!res.ok) {
		throw new ApiError(res.status, body.error ?? res.statusText);
	}
	return body;
}

/** Runs an operation of the Cahut of this domain. */
export function cahut<K extends keyof CahutOperations>(
	op: K,
	input: CahutOperations[K]["input"],
	token?: string,
): Promise<CahutOperations[K]["output"]> {
	return post(`${base}/api/${op}`, input, token) as Promise<CahutOperations[K]["output"]>;
}

/** Runs an operation of the server, for its administrators. */
export function server<K extends keyof ServerOperations>(
	op: K,
	input: ServerOperations[K]["input"],
	token?: string,
): Promise<ServerOperations[K]["output"]> {
	return post(`${base}/api/server/${op}`, input, token) as Promise<ServerOperations[K]["output"]>;
}

// The prefixes that the registry puts before the message of an error, which
// the console does not show.
const prefixes = ["invalid input: ", "not allowed: ", "not found: ", "too many attempts: ", "needs the approval of a person: "];

/**
 * Says what went wrong, in a sentence for the person who reads it. The API
 * writes its messages in English, for AIs and people alike.
 */
export function explain(err: unknown): string {
	if (!(err instanceof ApiError)) {
		return t.noAnswer;
	}
	let msg = err.message;
	for (const p of prefixes) {
		msg = msg.startsWith(p) ? msg.slice(p.length) : msg;
	}
	msg = msg.charAt(0).toUpperCase() + msg.slice(1);
	return msg.endsWith(".") ? msg : `${msg}.`;
}
