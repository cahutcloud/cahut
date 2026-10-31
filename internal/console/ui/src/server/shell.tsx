// The frame of every page of the console of the server once signed in: its
// top bar, Ink, or Paper in the dark theme, which names the server and its
// administrator, and its navigation, which a button opens on a narrow
// screen, with the account and the languages at its foot. See
// docs/adr/0058-a-console-for-the-server.md.
import { Languages as LanguagesIcon, LogOut, Menu, X } from "lucide-preact";
import type { ComponentChildren } from "preact";
import { useEffect, useState } from "preact/hooks";
import { initials, Languages } from "../components";
import { words } from "../i18n";
import { BarLogo } from "../logo";
import { Pictogram } from "../pictograms";
import { Link } from "../router";
import { useServer } from "./context";

const t = words({
	en: {
		menu: "Menu",
		main: "Main",
		server: "Server",
		cahuts: "Cahuts",
		admins: "Administrators",
		logbook: "Logbook",
		updates: "Updates",
		account: "Account",
		yourAccount: "Your account as an administrator of the server",
		signOut: "Sign out",
	},
	fr: {
		menu: "Menu",
		main: "Principale",
		server: "Serveur",
		cahuts: "Cahuts",
		admins: "Administrateurs",
		logbook: "Logbook",
		updates: "Mises à jour",
		account: "Compte",
		yourAccount: "Votre compte d'administrateur du serveur",
		signOut: "Se déconnecter",
	},
});

export function ServerShell(props: { path: string; children: ComponentChildren }) {
	const api = useServer();
	const [open, setOpen] = useState(false);
	useEffect(() => setOpen(false), [props.path]);
	const nav = (href: string, icon: ComponentChildren, label: string, exact = true) => (
		<Link
			class="c-nav"
			href={href}
			aria-current={(exact ? props.path === href : props.path === href || props.path.startsWith(`${href}/`)) ? "page" : undefined}
		>
			{icon}
			{label}
		</Link>
	);
	return (
		<div class="c-app">
			<header class="console-head">
				<div class="c-topbar server-topbar">
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
					<Link class="c-topbar__logo" href="/cahuts">
						<BarLogo height={26} />
					</Link>
					<span class="server-topbar__mark">
						<Pictogram name="serverAdmins" />
						{t.server}
					</span>
					<div class="console-account">
						<Link class="console-account__link" href="/account" title={t.yourAccount}>
							<span class="c-avatar" aria-hidden="true">
								{initials(api.session.name)}
							</span>
							<span class="c-small console-account__name">{api.session.name}</span>
						</Link>
						<button class="c-btn c-btn--ghost c-btn--sm" type="button" aria-label={t.signOut} onClick={api.signOut}>
							<LogOut class="i" size={16} aria-hidden="true" />
							<span class="console-account__name">{t.signOut}</span>
						</button>
					</div>
				</div>
			</header>
			<nav id="sidebar" class="c-sidebar" aria-label={t.main} data-open={open}>
				{nav("/cahuts", <Pictogram name="cahut" />, t.cahuts, false)}
				{nav("/admins", <Pictogram name="serverAdmins" />, t.admins)}
				{nav("/logbook", <Pictogram name="logbook" />, t.logbook)}
				{nav("/updates", <Pictogram name="updates" />, t.updates)}
				<div class="sidebar-foot">
					{nav("/account", <Pictogram name="account" />, t.account)}
					<div class="sidebar-languages">
						<LanguagesIcon class="i" size={16} aria-hidden="true" />
						<Languages />
					</div>
				</div>
			</nav>
			<main class="c-main">{props.children}</main>
		</div>
	);
}
