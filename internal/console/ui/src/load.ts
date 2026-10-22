// Loads what a screen shows from the API, once it renders.
import { useEffect, useState } from "preact/hooks";
import { ApiError, explain } from "./api";

export interface Loaded<T> {
	data?: T;
	error?: string;
}

/**
 * Runs load when the screen renders, and again when deps change. A session
 * that has ended calls onExpired, which signs the person out.
 */
export function useLoad<T>(load: () => Promise<T>, onExpired: () => void, deps: unknown[] = []): Loaded<T> {
	const [state, setState] = useState<Loaded<T>>({});
	useEffect(() => {
		let live = true;
		load().then(
			(data) => live && setState({ data }),
			(err) => {
				if (err instanceof ApiError && err.status === 401) {
					onExpired();
				} else if (live) {
					setState({ error: explain(err) });
				}
			},
		);
		return () => {
			live = false;
		};
	}, deps);
	return state;
}
