// The updates of the server, for its administrators: the version running,
// the newest release, the last check and the last update. The updater runs
// as root: the console asks it to check or to update, and it does within a
// minute. See docs/adr/0024-signed-releases-and-updates.md.
import { ArrowUpRight, RefreshCw } from "lucide-preact";
import { useEffect, useState } from "preact/hooks";
import { explain } from "../api";
import type { Updates as UpdatesInfo } from "../api.gen";
import { agoInText, ConfirmButton, Failure, Notice, PageHead } from "../components";
import { useServer } from "../server/context";
import { locale, words } from "../i18n";
import { useLoad } from "../load";

/** Who ran an update, as the updater says it: command line, timer or console. */
const runners: Record<string, string> = { "command line": "la ligne de commande", timer: "le timer de systemd", console: "la console" };

const t = words({
	en: {
		patch: "Every day, the updater installs the patch releases of the minor version running, and reports the others.",
		all: "Every day, the updater installs any newer release.",
		check: "Every day, the updater reports newer releases, and installs none.",
		off: "The updater asks GitHub nothing until someone asks it here or on the command line.",
		updated: "Updated",
		rolledBack: "Rolled back",
		refused: "Nothing installed",
		failed: "Failed",
		label: "This server",
		title: "Updates",
		lead: (running: string) =>
			`Cahut ${running} runs on this server. New releases come from GitHub, signed: the updater checks the signature, installs the release, and goes back to the version running if the new one does not start.`,
		unmanaged: (
			<>
				This server does not update itself: install.sh sets up the updater. Until then, update it on the server with{" "}
				<code class="c-mono">sudo /usr/local/bin/cahut update</code>.
			</>
		),
		everyDay: "Every day",
		changePolicy: (
			<>
				Change it with <code class="c-mono">CAHUT_UPDATES</code> in <code class="c-mono">/etc/cahut/cahut.env</code>, or with{" "}
				<code class="c-mono">install.sh --updates</code>: patch, all, check or off.
			</>
		),
		releases: "Releases",
		checking: "The updater reads the releases, within a minute.",
		installing: (version: string) =>
			`The updater installs Cahut ${version}, within a minute. The console may stop answering for a moment.`,
		notRead: "The updater has not read the releases yet.",
		checkFailed: (ago: string, error: string) => `The last check, ${ago}, could not read the releases: ${error}`,
		available: (latest: string) => (
			<>
				Cahut <strong>{latest}</strong> is available.
			</>
		),
		whatIsNew: "What is new",
		upToDate: (running: string, ago: string) => `Cahut ${running} is up to date. The updater last checked ${ago}.`,
		updateTo: (latest: string) => `Update to ${latest}`,
		updateAsk: (running: string, latest: string) =>
			`The server stops for a few seconds, and goes back to ${running} if ${latest} fails.`,
		install: (latest: string) => `Install ${latest}`,
		checkNow: "Check now",
		lastUpdate: "Last update",
		lastRun: (from: string, to: string, on: string, by: string) => `${from} to ${to}, on ${on}, by the ${by}.`,
	},
	fr: {
		patch: "Chaque jour, le programme de mise à jour installe les versions correctives de la version mineure en service, et signale les autres.",
		all: "Chaque jour, le programme de mise à jour installe toute version plus récente.",
		check: "Chaque jour, le programme de mise à jour signale les versions plus récentes, et n'en installe aucune.",
		off: "Le programme de mise à jour ne demande rien à GitHub tant que personne ne le lui demande ici ou en ligne de commande.",
		updated: "Mis à jour",
		rolledBack: "Retour en arrière",
		refused: "Rien d'installé",
		failed: "Échec",
		label: "Ce serveur",
		title: "Mises à jour",
		lead: (running) =>
			`Cahut ${running} tourne sur ce serveur. Les nouvelles versions viennent de GitHub, signées : le programme de mise à jour vérifie la signature, installe la version, et revient à la version en service si la nouvelle ne démarre pas.`,
		unmanaged: (
			<>
				Ce serveur ne se met pas à jour seul : install.sh installe le programme de mise à jour. D'ici là, mettez-le à jour sur le serveur
				avec <code class="c-mono">sudo /usr/local/bin/cahut update</code>.
			</>
		),
		everyDay: "Chaque jour",
		changePolicy: (
			<>
				Changez ce réglage avec <code class="c-mono">CAHUT_UPDATES</code> dans <code class="c-mono">/etc/cahut/cahut.env</code>, ou avec{" "}
				<code class="c-mono">install.sh --updates</code> : patch, all, check ou off.
			</>
		),
		releases: "Versions",
		checking: "Le programme de mise à jour lit les versions, d'ici une minute.",
		installing: (version) =>
			`Le programme de mise à jour installe Cahut ${version}, d'ici une minute. La console peut cesser de répondre un moment.`,
		notRead: "Le programme de mise à jour n'a pas encore lu les versions.",
		checkFailed: (ago, error) => `La dernière vérification, ${ago}, n'a pas pu lire les versions : ${error}`,
		available: (latest) => (
			<>
				Cahut <strong>{latest}</strong> est disponible.
			</>
		),
		whatIsNew: "Les nouveautés",
		upToDate: (running, ago) => `Cahut ${running} est à jour. Le programme de mise à jour a vérifié pour la dernière fois ${ago}.`,
		updateTo: (latest) => `Mettre à jour vers ${latest}`,
		updateAsk: (running, latest) => `Le serveur s'arrête quelques secondes, et revient à ${running} si ${latest} échoue.`,
		install: (latest) => `Installer ${latest}`,
		checkNow: "Vérifier maintenant",
		lastUpdate: "Dernière mise à jour",
		lastRun: (from, to, on, by) => `${from} vers ${to}, le ${on}, par ${runners[by] ?? by}.`,
	},
});

const policies: Record<string, string> = {
	patch: t.patch,
	all: t.all,
	check: t.check,
	off: t.off,
};

const results: Record<string, { badge: string; text: string }> = {
	updated: { badge: "c-badge--success", text: t.updated },
	rolled_back: { badge: "c-badge--warning", text: t.rolledBack },
	refused: { badge: "c-badge--neutral", text: t.refused },
	failed: { badge: "c-badge--danger", text: t.failed },
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

const day = (at: string) => new Date(at).toLocaleString(locale, { dateStyle: "medium", timeStyle: "short" });

export function Updates() {
	const api = useServer();
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
			<PageHead label={t.label} title={t.title}>
				{t.lead(data.running)}
			</PageHead>
			{!data.managed ? (
				<Notice>{t.unmanaged}</Notice>
			) : (
				<>
					<Release data={data} onAsk={ask} />
					{failure && <Failure message={failure} />}
					<LastUpdate data={data} />
					<section class="section c-card panel">
						<h2 class="c-h4">{t.everyDay}</h2>
						<p class="c-small">{policies[data.policy || "patch"]}</p>
						<p class="c-small c-muted">{t.changePolicy}</p>
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
			<h2 class="c-h4">{t.releases}</h2>
			{data.requested ? (
				<p class="c-small" role="status">
					<RefreshCw class="i updates__spin" size={14} aria-hidden="true" />{" "}
					{data.requested === "check" ? t.checking : t.installing(data.requested)}
				</p>
			) : !check ? (
				<p class="c-small">{t.notRead}</p>
			) : check.error ? (
				<div class="c-alert c-alert--warning">{t.checkFailed(agoInText(check.at), check.error)}</div>
			) : available ? (
				<p class="c-small">
					{t.available(latest)}{" "}
					{check.url && (
						<a class="c-link" href={check.url} target="_blank" rel="noreferrer">
							{t.whatIsNew}
							<ArrowUpRight class="i" size={14} aria-hidden="true" />
						</a>
					)}
				</p>
			) : (
				<p class="c-small">{t.upToDate(data.running, agoInText(check.at))}</p>
			)}
			{!data.requested && (
				<div class="updates__actions">
					{available && (
						<ConfirmButton
							label={t.updateTo(latest)}
							question={t.updateAsk(data.running, latest)}
							confirm={t.install(latest)}
							onConfirm={() => onAsk(latest)}
						/>
					)}
					<button class="c-btn c-btn--secondary c-btn--sm" type="button" onClick={() => onAsk("")}>
						<RefreshCw class="i" size={14} aria-hidden="true" />
						{t.checkNow}
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
			<h2 class="c-h4">{t.lastUpdate}</h2>
			<p class="c-small">
				<span class={`c-badge ${r.badge}`}>{r.text}</span> {t.lastRun(run.from, run.to, day(run.ended), run.by)}
			</p>
			{run.reason && <p class="c-small c-muted">{run.reason}</p>}
		</section>
	);
}
