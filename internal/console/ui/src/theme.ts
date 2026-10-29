// The pages where an administrator of the server signs in take the color of
// the server, as its bar does once they are in: Ink, or Paper when the
// system shows dark colors. Each is then the inverse of the same page for a
// person of the Cahut, and reads at once as that of the server. See
// docs/adr/0044-two-bars-for-the-server-and-the-cahut.md.
import { useEffect, useLayoutEffect, useState } from "preact/hooks";

export type Theme = "light" | "dark";

const dark = matchMedia("(prefers-color-scheme: dark)");

/**
 * Gives the page the theme of the server while the screen is shown, when
 * `on`, and returns it; the page follows the system again once it is gone.
 */
export function useServerTheme(on: boolean): Theme | undefined {
	const [systemDark, setSystemDark] = useState(dark.matches);
	useEffect(() => {
		const change = () => setSystemDark(dark.matches);
		dark.addEventListener("change", change);
		return () => dark.removeEventListener("change", change);
	}, []);
	const theme: Theme | undefined = on ? (systemDark ? "light" : "dark") : undefined;
	// Before the first paint, so that the page never shows in the colors of
	// a person first.
	useLayoutEffect(() => {
		if (!theme) return;
		const root = document.documentElement;
		root.dataset.theme = theme;
		return () => {
			delete root.dataset.theme;
		};
	}, [theme]);
	return theme;
}
