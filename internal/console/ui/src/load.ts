// Loads what a screen shows from the API, once it renders.
import { useCallback, useEffect, useState } from "preact/hooks";
import { explain } from "./api";

export interface Loaded<T> {
	data?: T;
	error?: string;
	/** Loads again, after a change. */
	reload: () => void;
}

/** Runs load when the screen renders, and again when deps change or on reload. */
export function useLoad<T>(load: () => Promise<T>, deps: unknown[] = []): Loaded<T> {
	const [state, setState] = useState<{ data?: T; error?: string }>({});
	const [round, setRound] = useState(0);
	const reload = useCallback(() => setRound((n) => n + 1), []);
	useEffect(() => {
		let live = true;
		load().then(
			(data) => live && setState({ data }),
			(err) => live && setState({ error: explain(err) }),
		);
		return () => {
			live = false;
		};
	}, [...deps, round]);
	return { ...state, reload };
}
