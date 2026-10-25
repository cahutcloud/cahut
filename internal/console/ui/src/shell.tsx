// The frame of every page once signed in: the top bar, with the account, and
// the navigation, which a button opens on a narrow screen, with the account
// and the languages of the console at its foot. The screens of the console
// fill its main area, at whose foot the legal pages of the Cahut are linked.
import { Languages as LanguagesIcon, LogOut, Menu, X } from "lucide-preact";
import { type ComponentChildren, Fragment } from "preact";
import { useEffect, useState } from "preact/hooks";
import type { Room } from "./api.gen";
import { initials, Languages, Legal } from "./components";
import { useConsole } from "./context";
import { words } from "./i18n";
import { Logo } from "./logo";
import { roomPictogram } from "./pages/rooms";
import { Pictogram } from "./pictograms";
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
		account: "Account",
		yourAccount: "Your account",
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
		serverAdmins: "Administrateurs",
		serverLogbook: "Logbook du serveur",
		updates: "Mises à jour",
		thisCahut: "Cette Cahut",
		people: "Personnes",
		houseStyle: "House style",
		kindAndLimits: "Type et limites",
		email: "E-mail",
		access: "IA et accès",
		account: "Compte",
		yourAccount: "Votre compte",
	},
});

export function Shell(props: { path: string; rooms: Room[]; onSignOut: () => void; children: ComponentChildren }) {
	const { session, isAdmin, individual, knocks, legal } = useConsole();
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
					<Link class="console-account__link" href="/account" title={t.yourAccount}>
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
				{nav("/", <Pictogram name="cahut" />, t.overview)}
				{person && nav("/knocks", <Pictogram name="knockFirst" />, t.knocks, false, knocks)}
				{person && nav("/logbook", <Pictogram name="logbook" />, t.logbook)}
				{person && (
					<>
						<div class="c-label">{t.rooms}</div>
						{props.rooms.map((r) => (
							<Fragment key={r.name}>{nav(`/rooms/${r.name}`, <Pictogram name={roomPictogram(r, individual)} />, r.title)}</Fragment>
						))}
						{nav("/rooms", <Pictogram name="roomNew" />, isAdmin ? t.everyRoom : t.newRoom)}
					</>
				)}
				{!person && (
					<>
						<div class="c-label">{t.thisServer}</div>
						{nav("/server/admins", <Pictogram name="serverAdmins" />, t.serverAdmins)}
						{nav("/server/logbook", <Pictogram name="logbook" />, t.serverLogbook)}
						{nav("/server/updates", <Pictogram name="updates" />, t.updates)}
					</>
				)}
				<div class="c-label">{t.thisCahut}</div>
				{isAdmin && nav("/people", <Pictogram name="people" />, t.people)}
				{isAdmin && nav("/house-style", <Pictogram name="houseStyle" />, t.houseStyle)}
				{isAdmin && nav("/settings", <Pictogram name="limits" />, t.kindAndLimits)}
				{isAdmin && nav("/mail", <Pictogram name="email" />, t.email)}
				{!person && nav("/rooms", <Pictogram name="room" />, t.rooms, false)}
				{!person && nav("/logbook", <Pictogram name="logbook" />, t.logbook)}
				{person && nav("/access", <Pictogram name="aiAccess" />, t.access)}
				<div class="sidebar-foot">
					{nav("/account", <Pictogram name="account" />, t.account)}
					<div class="sidebar-languages">
						<LanguagesIcon class="i" size={16} aria-hidden="true" />
						<Languages />
					</div>
				</div>
			</nav>
			<main class="c-main">
				{props.children}
				<Legal links={legal} />
			</main>
		</div>
	);
}
