// A small router over the History API: the console has a few pages, and
// the server answers every one of their addresses with the app.
import type { JSX } from "preact";
import { useEffect, useState } from "preact/hooks";

const changed = "cahut:navigate";

/** Goes to the page at to, as a link would. */
export function navigate(to: string, replace = false): void {
	if (replace) {
		history.replaceState(null, "", to);
	} else {
		history.pushState(null, "", to);
	}
	dispatchEvent(new Event(changed));
}

/** Returns the path of the current page, and follows it. */
export function usePath(): string {
	const [path, setPath] = useState(location.pathname);
	useEffect(() => {
		const update = () => setPath(location.pathname);
		addEventListener("popstate", update);
		addEventListener(changed, update);
		// A child may have navigated already, as Redirect does: the effects
		// of children run before those of their parents.
		update();
		return () => {
			removeEventListener("popstate", update);
			removeEventListener(changed, update);
		};
	}, []);
	return path;
}

/** A link to a page of the console, which opens without reloading it. */
export function Link(props: JSX.HTMLAttributes<HTMLAnchorElement> & { href: string }) {
	const onClick = (e: JSX.TargetedMouseEvent<HTMLAnchorElement>) => {
		if (e.button !== 0 || e.metaKey || e.ctrlKey || e.shiftKey || e.altKey) {
			return;
		}
		e.preventDefault();
		navigate(props.href);
	};
	return <a {...props} onClick={onClick} />;
}

/** Goes to the page at to once it renders. */
export function Redirect({ to }: { to: string }) {
	useEffect(() => navigate(to, true), [to]);
	return null;
}
