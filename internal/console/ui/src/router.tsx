// A small router over the History API: the console has a few pages, and
// the server answers every one of their addresses with the app.
//
// The console is served under /console/ of its Cahut's domain (see
// docs/adr/0020-console-under-a-path-of-port-443.md), and the console of
// the server under /server/ (see docs/adr/0058-a-console-for-the-server.md).
// Their pages name their paths without that prefix, such as /logbook: the
// router adds it to the address and takes it off again.
import type { AccessibleAnchorHTMLAttributes, TargetedMouseEvent } from "preact";
import { useEffect, useState } from "preact/hooks";

const changed = "cahut:navigate";

/** The path of this console, without its final slash: /console, or /server for the console of the server. */
export const base = /^\/server(\/|$)/.test(location.pathname) ? "/server" : "/console";

/** Returns the address of the page at to, a path of the console. */
export function href(to: string): string {
	return to.startsWith("/") ? base + to : to;
}

/** Returns the path of the current page, without the console's prefix. */
function current(): string {
	const p = location.pathname;
	return p === base ? "/" : p.startsWith(`${base}/`) ? p.slice(base.length) : p;
}

/** Goes to the page at to, as a link would. */
export function navigate(to: string, replace = false): void {
	if (replace) {
		history.replaceState(null, "", href(to));
	} else {
		history.pushState(null, "", href(to));
	}
	dispatchEvent(new Event(changed));
}

/** Returns the path of the current page, and follows it. */
export function usePath(): string {
	const [path, setPath] = useState(current);
	useEffect(() => {
		const update = () => setPath(current());
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

/**
 * Returns the query of the current address, such as ?q=q3, and follows it:
 * a page whose path stays, such as a search, changes with it.
 */
export function useQuery(): URLSearchParams {
	const [search, setSearch] = useState(location.search);
	useEffect(() => {
		const update = () => setSearch(location.search);
		addEventListener("popstate", update);
		addEventListener(changed, update);
		update();
		return () => {
			removeEventListener("popstate", update);
			removeEventListener(changed, update);
		};
	}, []);
	return new URLSearchParams(search);
}

/** A link to a page of the console, which opens without reloading it. */
export function Link(props: AccessibleAnchorHTMLAttributes<HTMLAnchorElement> & { href: string }) {
	const onClick = (e: TargetedMouseEvent<HTMLAnchorElement>) => {
		if (e.button !== 0 || e.metaKey || e.ctrlKey || e.shiftKey || e.altKey) {
			return;
		}
		e.preventDefault();
		navigate(props.href);
	};
	return <a {...props} href={href(props.href)} onClick={onClick} />;
}

/** Goes to the page at to once it renders. */
export function Redirect({ to }: { to: string }) {
	useEffect(() => navigate(to, true), [to]);
	return null;
}

/**
 * Goes to the address to, outside this console, in place of the current
 * page: from the console of the Cahut to that of the server, for one.
 */
export function Away({ to }: { to: string }) {
	useEffect(() => location.replace(to), [to]);
	return null;
}
