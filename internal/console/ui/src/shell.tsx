// The frame of every page once signed in: the top bar, with the account, and
// the navigation. The screens of the console fill its main area.
import { House, LogOut } from "lucide-preact";
import type { ComponentChildren } from "preact";
import { Logo } from "./logo";
import { Link } from "./router";
import type { Session } from "./session";

/** The initials of a name, for its avatar. */
function initials(name: string): string {
	const words = name.trim().split(/\s+/);
	return (words[0].charAt(0) + (words.length > 1 ? words[words.length - 1].charAt(0) : "")).toUpperCase();
}

export function Shell(props: { session: Session; path: string; onSignOut: () => void; children: ComponentChildren }) {
	const { session, path } = props;
	return (
		<div class="c-app">
			<header class="c-topbar">
				<Link class="c-topbar__logo" href="/">
					<Logo height={26} />
				</Link>
				<div class="console-account">
					<span class="c-avatar" aria-hidden="true">
						{initials(session.name)}
					</span>
					<span class="c-small console-account__name">
						{session.name}
						{session.kind === "admin" && <span class="c-muted"> · server</span>}
					</span>
					<button class="c-btn c-btn--ghost c-btn--sm" type="button" onClick={props.onSignOut}>
						<LogOut class="i" size={16} aria-hidden="true" />
						Sign out
					</button>
				</div>
			</header>
			<nav class="c-sidebar" aria-label="Main">
				<Link class="c-nav" href="/" aria-current={path === "/" ? "page" : undefined}>
					<House class="i" size={18} aria-hidden="true" />
					Overview
				</Link>
			</nav>
			<main class="c-main">{props.children}</main>
		</div>
	);
}
