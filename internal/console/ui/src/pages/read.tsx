// Signs a reader in to open Windows at the address of this Cahut, on the
// port of its Windows: the console gives a single-use code, and sends the
// browser to /-/session with it, which sets the reader's cookie there. See
// docs/adr/0019-window-versions-access-and-declaration.md.
import { useEffect, useState } from "preact/hooks";
import { ApiError, cahut, explain } from "../api";
import { Failure } from "../components";
import { de, words } from "../i18n";
import { Logo } from "../logo";
import type { Session } from "../session";

const t = words({
	en: {
		cannot: "Cannot open the Windows",
		opening: (host: string) => `Opening the Windows of ${host}…`,
	},
	fr: {
		cannot: "Impossible d'ouvrir les Windows",
		opening: (host) => `Ouverture des Windows ${de(host)}…`,
	},
});

/** The path of the Cahut to go back to, from the page's address. */
function destination(): string {
	const to = new URLSearchParams(location.search).get("to") ?? "/";
	return to.startsWith("/") && !to.startsWith("//") ? to : "/";
}

export function Read({ session, onExpired }: { session: Session; onExpired: () => void }) {
	const [error, setError] = useState("");
	useEffect(() => {
		Promise.all([cahut("get_cahut", {}, session.token), cahut("sign_in_to_windows", {}, session.token)]).then(
			([about, code]) => {
				const to = destination();
				location.replace(`${about.url}-/session?code=${encodeURIComponent(code.code)}&to=${encodeURIComponent(to)}`);
			},
			(err: unknown) => {
				if (err instanceof ApiError && err.status === 401) {
					onExpired();
					return;
				}
				setError(explain(err));
			},
		);
	}, []);
	return (
		<main class="auth">
			<div class="c-card auth__card">
				<Logo height={32} />
				{error ? (
					<div class="auth__head">
						<h1 class="c-h3">{t.cannot}</h1>
						<Failure message={error} />
					</div>
				) : (
					<p class="c-muted">{t.opening(location.hostname)}</p>
				)}
			</div>
		</main>
	);
}
