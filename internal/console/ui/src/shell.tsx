// The frame of every page once signed in: the top bar, with the account, and
// the navigation, which a button opens on a narrow screen. The screens of
// the console fill its main area.
import {
	Bot,
	DoorClosed,
	Gauge,
	House,
	LogOut,
	Mail,
	Menu,
	NotebookText,
	Palette,
	Plus,
	RefreshCw,
	Server,
	ShieldCheck,
	Users,
	X,
} from "lucide-preact";
import { type ComponentChildren, Fragment } from "preact";
import { useEffect, useState } from "preact/hooks";
import type { Room } from "./api.gen";
import { initials } from "./components";
import { useConsole } from "./context";
import { words } from "./i18n";
import { Logo } from "./logo";
import { roomIcon } from "./pages/rooms";
import { Link } from "./router";

const t = words({
	en: {
		waiting: " waiting",
		menu: "Menu",
		server: " · server",
		signOut: "Sign out",
		main: "Main",
		overview: "Overview",
		knocks: "Knocks",
		logbook: "Logbook",
		rooms: "Rooms",
		everyRoom: "Every Room",
		newRoom: "New Room",
		thisServer: "This server",
		serverAdmins: "Server administrators",
		serverLogbook: "Logbook of the server",
		updates: "Updates",
		thisCahut: "This Cahut",
		people: "People",
		houseStyle: "House style",
		kindAndLimits: "Kind and limits",
		email: "E-mail",
		access: "AIs and access",
	},
	fr: {
		waiting: " en attente",
		menu: "Menu",
		server: " · serveur",
		signOut: "Se déconnecter",
		main: "Principale",
		overview: "Vue d'ensemble",
		knocks: "Knocks",
		logbook: "Logbook",
		rooms: "Rooms",
		everyRoom: "Toutes les Rooms",
		newRoom: "Nouvelle Room",
		thisServer: "Ce serveur",
		serverAdmins: "Administrateurs du serveur",
		serverLogbook: "Logbook du serveur",
		updates: "Mises à jour",
		thisCahut: "Ce Cahut",
		people: "Personnes",
		houseStyle: "House style",
		kindAndLimits: "Type et limites",
		email: "E-mail",
		access: "IA et accès",
	},
});

export function Shell(props: { path: string; rooms: Room[]; onSignOut: () => void; children: ComponentChildren }) {
	const { session, isAdmin, individual, knocks } = useConsole();
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
					<span class="visually-hidden">{t.waiting}</span>
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
					{t.menu}
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
							{!person && <span class="c-muted">{t.server}</span>}
						</span>
					</Link>
					<button class="c-btn c-btn--ghost c-btn--sm" type="button" aria-label={t.signOut} onClick={props.onSignOut}>
						<LogOut class="i" size={16} aria-hidden="true" />
						<span class="console-account__name">{t.signOut}</span>
					</button>
				</div>
			</header>
			<nav id="sidebar" class="c-sidebar" aria-label={t.main} data-open={open}>
				{nav("/", <House class="i" size={18} aria-hidden="true" />, t.overview)}
				{person && nav("/knocks", <DoorClosed class="i" size={18} aria-hidden="true" />, t.knocks, false, knocks)}
				{person && nav("/logbook", <NotebookText class="i" size={18} aria-hidden="true" />, t.logbook)}
				{person && (
					<>
						<div class="c-label">{t.rooms}</div>
						{props.rooms.map((r) => {
							const Icon = roomIcon(r, individual);
							return <Fragment key={r.name}>{nav(`/rooms/${r.name}`, <Icon class="i" size={18} aria-hidden="true" />, r.title)}</Fragment>;
						})}
						{nav("/rooms", <Plus class="i" size={18} aria-hidden="true" />, isAdmin ? t.everyRoom : t.newRoom)}
					</>
				)}
				{!person && (
					<>
						<div class="c-label">{t.thisServer}</div>
						{nav("/server/admins", <ShieldCheck class="i" size={18} aria-hidden="true" />, t.serverAdmins)}
						{nav("/server/logbook", <NotebookText class="i" size={18} aria-hidden="true" />, t.serverLogbook)}
						{nav("/server/updates", <RefreshCw class="i" size={18} aria-hidden="true" />, t.updates)}
					</>
				)}
				<div class="c-label">{t.thisCahut}</div>
				{isAdmin && nav("/people", <Users class="i" size={18} aria-hidden="true" />, t.people)}
				{isAdmin && nav("/house-style", <Palette class="i" size={18} aria-hidden="true" />, t.houseStyle)}
				{isAdmin && nav("/settings", <Gauge class="i" size={18} aria-hidden="true" />, t.kindAndLimits)}
				{isAdmin && nav("/mail", <Mail class="i" size={18} aria-hidden="true" />, t.email)}
				{!person && nav("/rooms", <Server class="i" size={18} aria-hidden="true" />, t.rooms, false)}
				{!person && nav("/logbook", <NotebookText class="i" size={18} aria-hidden="true" />, t.logbook)}
				{person && nav("/access", <Bot class="i" size={18} aria-hidden="true" />, t.access)}
			</nav>
			<main class="c-main">{props.children}</main>
		</div>
	);
}
