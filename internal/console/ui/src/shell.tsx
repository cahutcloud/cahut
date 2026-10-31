// The frame of every page once signed in: the top bar of the Cahut, with the
// person, and the navigation, which a button opens on a narrow screen, with
// the account and the languages of the console at its foot. The screens of
// the console fill its main area, at whose foot the legal pages of the Cahut
// are linked. The console of the server has a frame of its own: see
// server/shell.tsx and docs/adr/0058-a-console-for-the-server.md.
import { Languages as LanguagesIcon, LogOut, Menu, X } from "lucide-preact";
import { type ComponentChildren, Fragment } from "preact";
import { useEffect, useState } from "preact/hooks";
import type { Room } from "./api.gen";
import { initials, Languages, Legal, SearchForm } from "./components";
import { useConsole } from "./context";
import { words } from "./i18n";
import { Logo } from "./logo";
import { roomPictogram } from "./pages/rooms";
import { Pictogram } from "./pictograms";
import { Link, useQuery } from "./router";

const t = words({
	en: {
		waiting: " waiting",
		menu: "Menu",
		signOut: "Sign out",
		main: "Main",
		overview: "Overview",
		knocks: "Knocks",
		logbook: "Logbook",
		rooms: "Rooms",
		everyRoom: "Every Room",
		newRoom: "New Room",
		thisCahut: "This Cahut",
		people: "People",
		kindAndLimits: "Kind and limits",
		email: "E-mail",
		moderation: "Moderation",
		access: "AIs and access",
		account: "Account",
		yourAccount: "Your account",
		search: "Search your Windows",
	},
	fr: {
		waiting: " en attente",
		menu: "Menu",
		signOut: "Se déconnecter",
		main: "Principale",
		overview: "Vue d'ensemble",
		knocks: "Knocks",
		logbook: "Logbook",
		rooms: "Rooms",
		everyRoom: "Toutes les Rooms",
		newRoom: "Nouvelle Room",
		thisCahut: "Cette Cahut",
		people: "Personnes",
		kindAndLimits: "Type et limites",
		email: "E-mail",
		moderation: "Modération",
		access: "IA et accès",
		account: "Compte",
		yourAccount: "Votre compte",
		search: "Chercher dans vos Windows",
	},
});

export function Shell(props: { path: string; rooms: Room[]; onSignOut: () => void; children: ComponentChildren }) {
	const { session, isAdmin, individual, knocks, reports, legal } = useConsole();
	const [open, setOpen] = useState(false);
	useEffect(() => setOpen(false), [props.path]);
	// The field of the search keeps the words of the page of results.
	const params = useQuery();
	const query = props.path === "/search" ? (params.get("q") ?? "") : "";

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
	return (
		<div class="c-app">
			<header class="console-head">
				<div class="c-topbar">
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
					{/* A person searches their Windows from every page: in the top bar,
					    or at the head of the menu on a narrow screen. See
					    docs/adr/0055-tags-and-search-of-windows.md. */}
					<SearchForm class="console-search" label={t.search} query={query} />
					<div class="console-account">
						<Link class="console-account__link" href="/account" title={t.yourAccount}>
							<span class="c-avatar" aria-hidden="true">
								{initials(session.name)}
							</span>
							<span class="c-small console-account__name">{session.name}</span>
						</Link>
						<button class="c-btn c-btn--ghost c-btn--sm" type="button" aria-label={t.signOut} onClick={props.onSignOut}>
							<LogOut class="i" size={16} aria-hidden="true" />
							<span class="console-account__name">{t.signOut}</span>
						</button>
					</div>
				</div>
			</header>
			<nav id="sidebar" class="c-sidebar" aria-label={t.main} data-open={open}>
				<SearchForm class="sidebar-search" label={t.search} query={query} />
				{nav("/", <Pictogram name="cahut" />, t.overview)}
				{nav("/knocks", <Pictogram name="knockFirst" />, t.knocks, false, knocks)}
				{nav("/logbook", <Pictogram name="logbook" />, t.logbook)}
				<div class="c-label">{t.rooms}</div>
				{props.rooms.map((r) => (
					<Fragment key={r.name}>{nav(`/rooms/${r.name}`, <Pictogram name={roomPictogram(r, individual)} />, r.title)}</Fragment>
				))}
				{nav("/rooms", <Pictogram name="roomNew" />, isAdmin ? t.everyRoom : t.newRoom)}
				<div class="c-label">{t.thisCahut}</div>
				{isAdmin && nav("/people", <Pictogram name="people" />, t.people)}
				{isAdmin && nav("/settings", <Pictogram name="limits" />, t.kindAndLimits)}
				{isAdmin && nav("/mail", <Pictogram name="email" />, t.email)}
				{isAdmin && nav("/moderation", <Pictogram name="moderation" />, t.moderation, true, reports)}
				{nav("/access", <Pictogram name="aiAccess" />, t.access)}
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
