// The frame of every page once signed in: the top bar, with the account, and
// the navigation, which a button opens on a narrow screen. The screens of
// the console fill its main area.
import {
	Bot,
	DoorClosed,
	House,
	LogOut,
	Menu,
	NotebookText,
	Palette,
	Plus,
	RefreshCw,
	Server,
	ShieldCheck,
	User,
	Users,
	X,
} from "lucide-preact";
import { type ComponentChildren, Fragment } from "preact";
import { useEffect, useState } from "preact/hooks";
import type { Room } from "./api.gen";
import { initials } from "./components";
import { useConsole } from "./context";
import { Logo } from "./logo";
import { Link } from "./router";

export function Shell(props: { path: string; rooms: Room[]; onSignOut: () => void; children: ComponentChildren }) {
	const { session, isAdmin, knocks } = useConsole();
	const [open, setOpen] = useState(false);
	useEffect(() => setOpen(false), [props.path]);

	const nav = (href: string, icon: ComponentChildren, label: string, exact = true, count = 0) => (
		<Link
			class="c-nav"
			href={href}
			aria-current={(exact ? props.path === href : props.path.startsWith(href)) ? "page" : undefined}
		>
			{icon}
			{label}
			{count > 0 && (
				<span class="c-nav__count">
					{count}
					<span class="visually-hidden"> waiting</span>
				</span>
			)}
		</Link>
	);
	const person = session.kind === "person";
	return (
		<div class="c-app">
			<header class="c-topbar">
				<button
					class="c-btn c-btn--ghost c-btn--sm menu-button"
					type="button"
					aria-expanded={open}
					aria-controls="sidebar"
					onClick={() => setOpen(!open)}
				>
					{open ? <X class="i" size={18} aria-hidden="true" /> : <Menu class="i" size={18} aria-hidden="true" />}
					Menu
				</button>
				<Link class="c-topbar__logo" href="/">
					<Logo height={26} />
				</Link>
				<div class="console-account">
					<Link class="console-account__link" href="/account">
						<span class="c-avatar" aria-hidden="true">
							{initials(session.name)}
						</span>
						<span class="c-small console-account__name">
							{session.name}
							{!person && <span class="c-muted"> · server</span>}
						</span>
					</Link>
					<button class="c-btn c-btn--ghost c-btn--sm" type="button" aria-label="Sign out" onClick={props.onSignOut}>
						<LogOut class="i" size={16} aria-hidden="true" />
						<span class="console-account__name">Sign out</span>
					</button>
				</div>
			</header>
			<nav id="sidebar" class="c-sidebar" aria-label="Main" data-open={open}>
				{nav("/", <House class="i" size={18} aria-hidden="true" />, "Overview")}
				{person && nav("/knocks", <DoorClosed class="i" size={18} aria-hidden="true" />, "Knocks", false, knocks)}
				{person && nav("/logbook", <NotebookText class="i" size={18} aria-hidden="true" />, "Logbook")}
				{person && (
					<>
						<div class="c-label">Rooms</div>
						{props.rooms.map((r) => {
							const Icon = r.kind === "personal" ? User : Users;
							return <Fragment key={r.name}>{nav(`/rooms/${r.name}`, <Icon class="i" size={18} aria-hidden="true" />, r.title)}</Fragment>;
						})}
						{nav("/rooms", <Plus class="i" size={18} aria-hidden="true" />, isAdmin ? "Every Room" : "New Room")}
					</>
				)}
				{!person && (
					<>
						<div class="c-label">This server</div>
						{nav("/server/admins", <ShieldCheck class="i" size={18} aria-hidden="true" />, "Administrators")}
						{nav("/server/logbook", <NotebookText class="i" size={18} aria-hidden="true" />, "Logbook of the server")}
						{nav("/server/updates", <RefreshCw class="i" size={18} aria-hidden="true" />, "Updates")}
					</>
				)}
				<div class="c-label">This Cahut</div>
				{isAdmin && nav("/people", <Users class="i" size={18} aria-hidden="true" />, "People")}
				{isAdmin && nav("/house-style", <Palette class="i" size={18} aria-hidden="true" />, "House style")}
				{!person && nav("/rooms", <Server class="i" size={18} aria-hidden="true" />, "Rooms", false)}
				{!person && nav("/logbook", <NotebookText class="i" size={18} aria-hidden="true" />, "Logbook")}
				{person && nav("/access", <Bot class="i" size={18} aria-hidden="true" />, "AIs and access")}
			</nav>
			<main class="c-main">{props.children}</main>
		</div>
	);
}
