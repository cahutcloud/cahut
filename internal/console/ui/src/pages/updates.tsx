// The updates of the server, for its administrators: the version running,
// the newest release, the last check and the last update. The updater runs
// as root: the console asks it to check or to update, and it does within a
// minute. See docs/adr/0024-signed-releases-and-updates.md.
import { ArrowUpRight, RefreshCw } from "lucide-preact";
import { useEffect, useState } from "preact/hooks";
import { explain } from "../api";
import type { Updates as UpdatesInfo } from "../api.gen";
import { ago, ConfirmButton, Failure, Notice, PageHead } from "../components";
import { useConsole } from "../context";
import { useLoad } from "../load";

const policies: Record<string, string> = {
	patch: "Every day, the updater installs the patch releases of the minor version running, and reports the others.",
	all: "Every day, the updater installs any newer release.",
	check: "Every day, the updater reports newer releases, and installs none.",
	off: "The updater asks GitHub nothing until someone asks it here or on the command line.",
};

const results: Record<string, { badge: string; text: string }> = {
	updated: { badge: "c-badge--success", text: "Updated" },
	rolled_back: { badge: "c-badge--warning", text: "Rolled back" },
	refused: { badge: "c-badge--neutral", text: "Nothing installed" },
	failed: { badge: "c-badge--danger", text: "Failed" },
};

/** Whether version a comes after version b, as Semantic Versioning orders them. */
export function newer(a: string, b: string): boolean {
	const parse = (v: string) => {
		const [core, pre = ""] = v.replace(/^v/, "").split("+")[0].split(/-(.*)/s);
		return { n: core.split(".").map(Number), pre };
	};
	const x = parse(a);
	const y = parse(b);
	for (let i = 0; i < 3; i++) {
		if ((x.n[i] ?? 0) !== (y.n[i] ?? 0)) return (x.n[i] ?? 0) > (y.n[i] ?? 0);
	}
	if (x.pre === y.pre) return false;
	if (!x.pre) return true;
	if (!y.pre) return false;
	return x.pre > y.pre;
}

const day = (at: string) => new Date(at).toLocaleString("en-GB", { dateStyle: "medium", timeStyle: "short" });

export function Updates() {
	const api = useConsole();
	const { data, error, reload } = useLoad(() => api.server("get_updates", {}));
	const [failure, setFailure] = useState("");

	// While the updater has a request to take, ask again every few seconds.
	useEffect(() => {
		if (!data?.requested) return;
		const timer = setInterval(reload, 5000);
		return () => clearInterval(timer);
	}, [data?.requested]);

	if (error) return <Failure message={error} />;
	if (!data) return null;
	const ask = async (version: string) => {
		setFailure("");
		try {
			await api.server("request_update", { version });
			reload();
		} catch (err) {
			setFailure(explain(err));
		}
	};
	return (
		<>
			<PageHead label="This server" title="Updates">
				Cahut {data.running} runs on this server. New releases come from GitHub, signed: the updater checks the signature,
				installs the release, and goes back to the version running if the new one does not start.
			</PageHead>
			{!data.managed ? (
				<Notice>
					This server does not update itself: install.sh sets up the updater. Until then, update it on the server with{" "}
					<code class="c-mono">sudo /usr/local/bin/cahut update</code>.
				</Notice>
			) : (
				<>
					<Release data={data} onAsk={ask} />
					{failure && <Failure message={failure} />}
					<LastUpdate data={data} />
					<section class="section c-card panel">
						<h2 class="c-h4">Every day</h2>
						<p class="c-small">{policies[data.policy || "patch"]}</p>
						<p class="c-small c-muted">
							Change it with <code class="c-mono">CAHUT_UPDATES</code> in <code class="c-mono">/etc/cahut/cahut.env</code>, or with{" "}
							<code class="c-mono">install.sh --updates</code>: patch, all, check or off.
						</p>
					</section>
				</>
			)}
		</>
	);
}

function Release({ data, onAsk }: { data: UpdatesInfo; onAsk: (version: string) => void }) {
	const check = data.check;
	const latest = check?.latest ?? "";
	const available = latest !== "" && newer(latest, data.running);
	return (
		<section class="section c-card panel">
			<h2 class="c-h4">Releases</h2>
			{data.requested ? (
				<p class="c-small" role="status">
					<RefreshCw class="i updates__spin" size={14} aria-hidden="true" />{" "}
					{data.requested === "check"
						? "The updater reads the releases, within a minute."
						: `The updater installs Cahut ${data.requested}, within a minute. The console may stop answering for a moment.`}
				</p>
			) : !check ? (
				<p class="c-small">The updater has not read the releases yet.</p>
			) : check.error ? (
				<div class="c-alert c-alert--warning">
					The last check, {ago(check.at)}, could not read the releases: {check.error}
				</div>
			) : available ? (
				<p class="c-small">
					Cahut <strong>{latest}</strong> is available.{" "}
					{check.url && (
						<a class="c-link" href={check.url} target="_blank" rel="noreferrer">
							What is new
							<ArrowUpRight class="i" size={14} aria-hidden="true" />
						</a>
					)}
				</p>
			) : (
				<p class="c-small">Cahut {data.running} is up to date. The updater last checked {ago(check.at)}.</p>
			)}
			{!data.requested && (
				<div class="updates__actions">
					{available && (
						<ConfirmButton
							label={`Update to ${latest}`}
							question={`The server stops for a few seconds, and goes back to ${data.running} if ${latest} fails.`}
							confirm={`Install ${latest}`}
							onConfirm={() => onAsk(latest)}
						/>
					)}
					<button class="c-btn c-btn--secondary c-btn--sm" type="button" onClick={() => onAsk("")}>
						<RefreshCw class="i" size={14} aria-hidden="true" />
						Check now
					</button>
				</div>
			)}
		</section>
	);
}

function LastUpdate({ data }: { data: UpdatesInfo }) {
	const run = data.last;
	if (!run) return null;
	const r = results[run.result] ?? results.refused;
	return (
		<section class="section c-card panel">
			<h2 class="c-h4">Last update</h2>
			<p class="c-small">
				<span class={`c-badge ${r.badge}`}>{r.text}</span> {run.from} to {run.to}, on {day(run.ended)}, by the {run.by}.
			</p>
			{run.reason && <p class="c-small c-muted">{run.reason}</p>}
		</section>
	);
}
