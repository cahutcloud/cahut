// The page of the link that confirms a new e-mail address, for a person of
// this Cahut or an administrator of the server. The link is the proof: no
// sign-in is needed. Its code comes in the fragment of the address. See
// docs/adr/0031-e-mail-through-a-relay.md.
import { useEffect, useState } from "preact/hooks";
import { cahut, explain, server } from "../api";
import { fragment as code } from "../fragment";
import { Logo } from "../logo";
import { Link } from "../router";
import type { Session } from "../session";

export function ConfirmEmail({ kind }: { kind: Session["kind"] }) {
	const [state, setState] = useState<{ email?: string; error?: string }>({});
	useEffect(() => {
		if (!code) return;
		const confirm = kind === "admin" ? server("confirm_email", { code }) : cahut("confirm_email", { code });
		confirm.then(
			(r) => setState({ email: r.email }),
			(err) => setState({ error: explain(err) }),
		);
	}, []);
	return (
		<main class="auth">
			<div class="c-card auth__card">
				<Logo height={32} />
				<div class="auth__head">
					<h1 class="c-h3">Your e-mail address</h1>
				</div>
				{!code && (
					<div class="c-alert c-alert--warning" role="alert">
						This link has no code. Open the whole link of the e-mail.
					</div>
				)}
				{state.email && (
					<div class="c-alert c-alert--success" role="status">
						<div>
							Your e-mail address is now <span class="c-mono">{state.email}</span>. Your old address got a notice.
						</div>
					</div>
				)}
				{state.error && (
					<div class="c-alert c-alert--danger" role="alert">
						{state.error}
					</div>
				)}
				<p class="c-small">
					<Link class="c-link" href={kind === "admin" ? "/server/sign-in" : "/"}>
						Go to the console
					</Link>
				</p>
			</div>
		</main>
	);
}
