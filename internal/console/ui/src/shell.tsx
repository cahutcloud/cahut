// The frame of every page once signed in: the bar of the server, when the
// browser holds its session, above the top bar of the Cahut, with who acts
// in it, and the navigation, which a button opens on a narrow screen, with
// the account and the languages of the console at its foot. The screens of
// the console fill its main area, at whose foot the legal pages of the Cahut
// are linked.
import { ArrowUpRight, Check, ChevronsUpDown, Languages as LanguagesIcon, LayoutGrid, LogIn, LogOut, Menu, X } from "lucide-preact";
import { type ComponentChildren, Fragment } from "preact";
import { useEffect, useRef, useState } from "preact/hooks";
import type { Room } from "./api.gen";
import { initials, Languages, Legal } from "./components";
import { useConsole } from "./context";
import { words } from "./i18n";
import { Logo } from "./logo";
import { originOf } from "./pages/overview";
import { roomPictogram } from "./pages/rooms";
import { Pictogram } from "./pictograms";
import { href, Link } from "./router";
import type { Session } from "./session";

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
		thisServer: "This server",
		cahuts: "Cahuts",
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
		administration: "Server administration",
		switchCahut: "Switch Cahut",
		cahutsOfServer: "Cahuts of this server",
		youAreHere: "You are here",
		openConsole: "Open its console",
		manageCahuts: "Manage the Cahuts of the server",
		eachConsole: "Each Cahut has its own console: you sign in there as an administrator of the server.",
		serverAccount: "Your account as an administrator of the server",
		signOutServer: "Sign out of the server",
		actingAsServer: () => (
			<>
				In this Cahut, you act as an <strong>administrator of the server</strong>.
			</>
		),
		signInAsPerson: "Sign in as a person",
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
		thisServer: "Ce serveur",
		cahuts: "Cahuts",
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
		administration: "Administration du serveur",
		switchCahut: "Changer de Cahut",
		cahutsOfServer: "Cahuts de ce serveur",
		youAreHere: "Vous êtes ici",
		openConsole: "Ouvrir sa console",
		manageCahuts: "Gérer les Cahuts du serveur",
		eachConsole: "Chaque Cahut a sa propre console : vous vous y connectez en administrateur du serveur.",
		serverAccount: "Votre compte d'administrateur du serveur",
		signOutServer: "Se déconnecter du serveur",
		actingAsServer: () => (
			<>
				Dans cette Cahut, vous agissez en <strong>administrateur du serveur</strong>.
			</>
		),
		signInAsPerson: "Se connecter en tant que personne",
	},
});

export function Shell(props: { path: string; rooms: Room[]; onSignOut: () => void; children: ComponentChildren }) {
	const { session, serverSession, isAdmin, individual, knocks, legal } = useConsole();
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
	// The pages of the server follow those of the person when the browser
	// holds both sessions. See docs/adr/0043-the-console-holds-two-sessions.md.
	const serverPages = (cahuts: boolean) => (
		<>
			<div class="c-label">{t.thisServer}</div>
			{cahuts && nav("/server/cahuts", <Pictogram name="cahut" />, t.cahuts)}
			{nav("/server/admins", <Pictogram name="serverAdmins" />, t.serverAdmins)}
			{nav("/server/logbook", <Pictogram name="logbook" />, t.serverLogbook)}
			{nav("/server/updates", <Pictogram name="updates" />, t.updates)}
		</>
	);
	return (
		<div class="c-app" data-server={serverSession ? "" : undefined}>
			<header class="console-head">
				{serverSession && <ServerBar admin={serverSession} path={props.path} />}
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
					{/* The top bar of the Cahut says who acts in it: the person, or the
					    administrator of the server alone. See
					    docs/adr/0044-two-bars-for-the-server-and-the-cahut.md. */}
					{person ? (
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
					) : (
						<div class="console-acting">
							<p class="c-small c-muted">{t.actingAsServer()}</p>
							<Link class="c-btn c-btn--secondary c-btn--sm" href="/sign-in" aria-label={t.signInAsPerson}>
								<LogIn class="i" size={16} aria-hidden="true" />
								<span class="console-account__name">{t.signInAsPerson}</span>
							</Link>
						</div>
					)}
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
				{!person && serverPages(false)}
				<div class="c-label">{t.thisCahut}</div>
				{isAdmin && nav("/people", <Pictogram name="people" />, t.people)}
				{isAdmin && nav("/house-style", <Pictogram name="houseStyle" />, t.houseStyle)}
				{(isAdmin || serverSession) && nav("/settings", <Pictogram name="limits" />, t.kindAndLimits)}
				{isAdmin && nav("/mail", <Pictogram name="email" />, t.email)}
				{!person && nav("/rooms", <Pictogram name="room" />, t.rooms, false)}
				{!person && nav("/logbook", <Pictogram name="logbook" />, t.logbook)}
				{person && nav("/access", <Pictogram name="aiAccess" />, t.access)}
				{person && serverSession && serverPages(true)}
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

/**
 * The bar of the server, above the top bar while the browser holds the
 * session of an administrator of the server: the Cahut of this domain, which
 * opens the list of the Cahuts of the server, and the administrator, who
 * signs out of the server here. See
 * docs/adr/0044-two-bars-for-the-server-and-the-cahut.md.
 */
function ServerBar({ admin, path }: { admin: Session; path: string }) {
	const api = useConsole();
	const [open, setOpen] = useState(false);
	const [cahuts, setCahuts] = useState<string[] | null>(null);
	const box = useRef<HTMLDivElement>(null);
	const button = useRef<HTMLButtonElement>(null);
	useEffect(() => setOpen(false), [path]);
	useEffect(() => {
		if (!open) return;
		if (!cahuts) {
			api.server("list_domains", {}).then(
				(d) => setCahuts(d.cahuts.map((c) => c.domain)),
				() => setCahuts([location.hostname]),
			);
		}
		// A click outside the list, or Escape, closes it.
		const click = (e: MouseEvent) => {
			if (!box.current?.contains(e.target as Node)) setOpen(false);
		};
		const key = (e: KeyboardEvent) => {
			if (e.key === "Escape") {
				setOpen(false);
				button.current?.focus();
			}
		};
		document.addEventListener("click", click);
		document.addEventListener("keydown", key);
		return () => {
			document.removeEventListener("click", click);
			document.removeEventListener("keydown", key);
		};
	}, [open]);
	return (
		<div class="server-bar">
			<span class="server-bar__mode">
				<Pictogram name="serverAdmins" />
				<span class="server-bar__words">{t.administration}</span>
			</span>
			<span class="server-bar__sep" aria-hidden="true">
				/
			</span>
			<div class="server-bar__cahut" ref={box}>
				<button
					ref={button}
					class="server-bar__switch"
					type="button"
					aria-expanded={open}
					aria-controls="server-cahuts"
					title={t.switchCahut}
					onClick={() => setOpen(!open)}
				>
					<Pictogram name="cahut" />
					<span class="server-bar__domain">{location.hostname}</span>
					<span class="visually-hidden">, {t.switchCahut}</span>
					<ChevronsUpDown class="i" size={16} aria-hidden="true" />
				</button>
				{open && (
					<div id="server-cahuts" class="c-card cahut-menu">
						<span class="c-label">{t.cahutsOfServer}</span>
						{cahuts?.map((domain) =>
							domain === location.hostname ? (
								<span key={domain} class="cahut-menu__item" aria-current="true">
									<Pictogram name="cahut" />
									<span class="cahut-menu__domain">{domain}</span>
									<span class="cahut-menu__hint">
										<Check class="i" size={14} aria-hidden="true" />
										{t.youAreHere}
									</span>
								</span>
							) : (
								<a key={domain} class="cahut-menu__item" href={`${originOf(domain)}${href("/server/sign-in")}`}>
									<Pictogram name="cahut" />
									<span class="cahut-menu__domain">{domain}</span>
									<span class="cahut-menu__hint">
										{t.openConsole}
										<ArrowUpRight class="i" size={14} aria-hidden="true" />
									</span>
								</a>
							),
						)}
						<hr />
						<Link class="cahut-menu__item" href="/server/cahuts">
							<LayoutGrid class="i" size={18} aria-hidden="true" />
							<span>{t.manageCahuts}</span>
						</Link>
						<p class="cahut-menu__note">{t.eachConsole}</p>
					</div>
				)}
			</div>
			<div class="server-bar__account">
				<Link class="server-bar__who" href="/account" title={t.serverAccount}>
					<span class="c-avatar" aria-hidden="true">
						{initials(admin.name)}
					</span>
					<span class="server-bar__words">{admin.name}</span>
				</Link>
				<button class="c-btn c-btn--ghost c-btn--sm" type="button" title={t.signOutServer} onClick={api.leaveServer}>
					<LogOut class="i" size={16} aria-hidden="true" />
					<span class="server-bar__words">{t.signOutServer}</span>
				</button>
			</div>
		</div>
	);
}
