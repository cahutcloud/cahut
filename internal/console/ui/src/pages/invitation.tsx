// The page of a personal link of an invitation to read a Window: who invites,
// to which Window, with what note. A click opens the Window: the console asks
// for a code valid a minute, which /-/invitation turns into a pass, as the
// right Key does. Opening the link does nothing until the click, so that the
// mail filters that open links use nothing up. See
// docs/adr/0031-e-mail-through-a-relay.md.
import { ArrowUpRight } from "lucide-preact";
import { useState } from "preact/hooks";
import { cahut, explain } from "../api";
import { fragment as code } from "../fragment";
import { useLoad } from "../load";
import { Logo } from "../logo";

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
						This link has no code. Open the whole link of the e-mail.
					</div>
				) : error ? (
					<div class="c-alert c-alert--danger" role="alert">
						{error}
					</div>
				) : (
					data && (
						<>
							<div class="auth__head">
								<p class="c-label">Invitation</p>
								<h1 class="c-h3">{data.title}</h1>
								<p class="c-small c-muted">
									{data.by} invites you to read it{data.ai ? `, through ${data.ai}` : ""}, at {location.hostname}.
								</p>
							</div>
							{data.note && (
								<blockquote class="quote">
									<p class="c-small c-muted">{data.ai ? `${data.ai}, for ${data.by}` : data.by} wrote:</p>
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
								Open {data.title}
							</button>
							<p class="c-small c-muted">
								This link is yours: it opens the Window on this device, and on others, until{" "}
								{new Date(data.expires_at).toLocaleDateString("en-GB", { dateStyle: "long" })}, unless the invitation is taken back.
							</p>
						</>
					)
				)}
			</div>
		</main>
	);
}
