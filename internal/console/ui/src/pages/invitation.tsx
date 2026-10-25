// The page of a personal link of an invitation to read a Window: who invites,
// to which Window, with what note. A click opens the Window: the console asks
// for a code valid a minute, which /-/invitation turns into a pass, as the
// right Key does. Opening the link does nothing until the click, so that the
// mail filters that open links use nothing up. See
// docs/adr/0031-e-mail-through-a-relay.md.
import { ArrowUpRight } from "lucide-preact";
import { useState } from "preact/hooks";
import { cahut, explain } from "../api";
import { Languages, LegalBeforeSignIn } from "../components";
import { fragment as code } from "../fragment";
import { locale, words } from "../i18n";
import { useLoad } from "../load";
import { Logo } from "../logo";

const t = words({
	en: {
		noCode: "This link has no code. Open the whole link of the e-mail.",
		invitation: "Invitation",
		invites: (by: string, ai: string | undefined, host: string) => `${by} invites you to read it${ai ? `, through ${ai}` : ""}, at ${host}.`,
		wrote: (by: string, ai: string | undefined) => `${ai ? `${ai}, for ${by}` : by} wrote:`,
		open: (title: string) => `Open ${title}`,
		yours: (until: string) =>
			`This link is yours: it opens the Window on this device, and on others, until ${until}, unless the invitation is taken back.`,
	},
	fr: {
		noCode: "Ce lien n'a pas de code. Ouvrez en entier le lien de l'e-mail.",
		invitation: "Invitation",
		invites: (by, ai, host) => `${by} vous invite à la lire${ai ? `, via ${ai}` : ""}, sur ${host}.`,
		wrote: (by, ai) => `${ai ? `${ai}, pour ${by}` : by} a écrit :`,
		open: (title) => `Ouvrir ${title}`,
		yours: (until) =>
			`Ce lien est à vous : il ouvre la Window sur cet appareil, et sur d'autres, jusqu'au ${until}, sauf si l'invitation est retirée.`,
	},
});

export function Invitation() {
	const { data, error } = useLoad(() => cahut("get_invitation", { code }));
	const [failure, setFailure] = useState("");
	const [busy, setBusy] = useState(false);
	const open = async () => {
		setBusy(true);
		setFailure("");
		try {
			const o = await cahut("open_invitation", { code });
			location.assign(`/-/invitation?code=${encodeURIComponent(o.code)}`);
		} catch (err) {
			setFailure(explain(err));
			setBusy(false);
		}
	};
	return (
		<main class="auth">
			<div class="c-card auth__card">
				<Logo height={32} />
				{!code ? (
					<div class="c-alert c-alert--warning" role="alert">
						{t.noCode}
					</div>
				) : error ? (
					<div class="c-alert c-alert--danger" role="alert">
						{error}
					</div>
				) : (
					data && (
						<>
							<div class="auth__head">
								<p class="c-label">{t.invitation}</p>
								<h1 class="c-h3">{data.title}</h1>
								<p class="c-small c-muted">{t.invites(data.by, data.ai, location.hostname)}</p>
							</div>
							{data.note && (
								<blockquote class="quote">
									<p class="c-small c-muted">{t.wrote(data.by, data.ai)}</p>
									<p class="note">{data.note}</p>
								</blockquote>
							)}
							{failure && (
								<div class="c-alert c-alert--danger" role="alert">
									{failure}
								</div>
							)}
							<button class="c-btn c-btn--primary" type="button" disabled={busy} onClick={open}>
								<ArrowUpRight class="i" size={16} aria-hidden="true" />
								{t.open(data.title)}
							</button>
							<p class="c-small c-muted">{t.yours(new Date(data.expires_at).toLocaleDateString(locale, { dateStyle: "long" }))}</p>
						</>
					)
				)}
			</div>
			<Languages />
			<LegalBeforeSignIn />
		</main>
	);
}
