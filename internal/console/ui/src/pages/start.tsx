// First steps: the guide that leads a person to connect a first AI, to ask
// it for a first Window, and to share it, made for a phone first. It reads
// what it shows from the API, every three seconds while it is visible and
// as soon as it is visible again, so it moves on by itself once the AI is
// connected, then once it has published, and a person who began on a phone
// finds it at the same step on a computer. Only its end, when the person
// dismisses it, stays in this browser. See
// docs/adr/0061-a-guide-to-the-first-ai.md.
import { ArrowLeft, Check, ChevronRight, CircleCheck, Copy, ExternalLink, Monitor, Smartphone } from "lucide-preact";
import type { ComponentChildren } from "preact";
import { useEffect, useState } from "preact/hooks";
import type { About, Connection, TokenInfo, Window } from "../api.gen";
import { CopyField, Notice, WindowCards } from "../components";
import { type ClientId, ClientSteps, clientNames, Film } from "../connect";
import { useConsole } from "../context";
import { words } from "../i18n";
import { Logo } from "../logo";
import { Pictogram } from "../pictograms";
import { Link, navigate, Redirect } from "../router";
import { ShareMenu } from "../share";

const t = words({
	en: {
		firstSteps: "First steps",
		step: (n: number) => `Step ${n} of 3`,
		skip: "Skip",
		back: "Back",
		home: "Overview",
		ready: (name: string) => `Your Cahut is ready, ${name}`,
		readyLead: "Your AI publishes here what it makes for you: documents, decks, dashboards. Three steps.",
		planConnect: "Connect your AI",
		planConnectHelp: "Claude, Mistral Vibe or another",
		planAsk: "Ask it for a Window",
		planAskHelp: "A page, at your address",
		planShare: "Share it",
		planShareHelp: "With a link, to whoever you choose",
		yourWindows: "Your Windows will be at",
		start: "Start",
		later: "Later",
		which: "Which AI do you want to connect?",
		whichLead: "The one you already use. You can connect others later.",
		fromPhone: "From this phone",
		onComputer: "On a computer",
		claudeHelp: "The app or claude.ai · free plan included",
		vibeHelp: "Formerly Le Chat · free plan included",
		chatgptHelp: "Business, Enterprise and Edu plans",
		claudeCodeHelp: "In a terminal",
		otherName: "Another AI",
		otherHelp: "Any client that speaks MCP",
		noAI: "I do not have an AI yet",
		connect: (client: string) => `Connect ${client}`,
		listed: "This Cahut is in the directory of connectors of Claude: nothing to copy.",
		waitingFor: (client: string) => `Waiting for ${client}`,
		comeBack: "Come back to this tab once it is connected: the next step waits for you here.",
		plan: "Which plan of ChatGPT do you have?",
		planLead: "ChatGPT publishes to a Cahut on some plans only.",
		planOrg: "Business, Enterprise or Edu",
		planOwn: "Free, Go, Plus or Pro",
		planUnsure: "I am not sure",
		planCannot: "With this plan, ChatGPT cannot publish. Claude and Mistral Vibe can, on their free plans too.",
		planWhich:
			"Business, Enterprise and Edu are plans of organizations, which an administrator of the workspace manages. A plan you pay for yourself, or the free one, is one of the others.",
		connectClaude: "Connect Claude",
		connectVibe: "Connect Mistral Vibe",
		computer: "This step is done on a computer",
		computerWhy: (client: string) =>
			client === "ChatGPT" ? "ChatGPT adds an MCP server on its website, from a computer." : `${client} runs in a terminal, on a computer.`,
		computerHow: "On your computer, open this address and sign in: the guide waits for you there, at this same step.",
		theAddress: "the address of the guide",
		showAnyway: "Show the steps here anyway",
		fromThisPhone: "Connect Claude from this phone instead",
		freeAI: "Start with a free AI",
		freeAILead: "These two AIs publish to your Cahut on their free plans. Create an account, then come back here.",
		claudeBy: "By Anthropic. On the web, and as an app for phones and computers.",
		vibeBy: "By Mistral AI, formerly Le Chat.",
		createClaude: "Create a Claude account",
		createVibe: "Create a Mistral account",
		independent: "Cahut is independent of Anthropic, Mistral AI and OpenAI.",
		haveOne: "I have an account, go on",
		isConnected: (client: string) => `${client} is connected`,
		readPublish: "It may read and publish.",
		readOnly: "It may only read: to publish, give it read and publish on AIs and access.",
		ask: "Ask for your first Window",
		askLead: "Copy an idea into your AI, or ask for what you want.",
		waitingWindow: "Waiting for your first Window",
		open: (host: string) => `Open ${host}`,
		online: "Your first Window is online",
		itsAddress: "Its address",
		youAlone: "For now, it opens to you alone.",
		itsRoom: "For now, it opens to the members of its Room.",
		shareLink: "A share link opens it to whoever holds it, without an account, until the link ends.",
		openWindow: "Open the Window",
		finish: "Finish",
		cardTitle: (n: number) => `First steps · step ${n} of 3`,
		yourAI: "your AI",
		cardConnect: "Connect your first AI",
		cardConnectHelp: "Your AI publishes on Cahut over MCP. Choose it, and follow its steps.",
		cardAsk: "Ask your AI for a first Window",
		cardAskHelp: "Your AI is connected. Ask it for a Window, then share it.",
		resume: "Resume",
		hide: "Hide",
	},
	fr: {
		firstSteps: "Premiers pas",
		step: (n) => `Étape ${n} sur 3`,
		skip: "Passer",
		back: "Retour",
		home: "Vue d'ensemble",
		ready: (name) => `Votre Cahut est prête, ${name}`,
		readyLead: "Votre IA y publie ce qu'elle fabrique pour vous : documents, présentations, tableaux de bord. Trois étapes.",
		planConnect: "Brancher votre IA",
		planConnectHelp: "Claude, Mistral Vibe ou une autre",
		planAsk: "Lui demander une Window",
		planAskHelp: "Une page, à votre adresse",
		planShare: "La partager",
		planShareHelp: "Par un lien, à qui vous voulez",
		yourWindows: "Vos Windows seront à",
		start: "Commencer",
		later: "Plus tard",
		which: "Quelle IA voulez-vous brancher ?",
		whichLead: "Celle que vous utilisez déjà. Vous en brancherez d'autres ensuite.",
		fromPhone: "Depuis ce téléphone",
		onComputer: "Sur un ordinateur",
		claudeHelp: "L'appli ou claude.ai · offre gratuite comprise",
		vibeHelp: "L'ancien Le Chat · offre gratuite comprise",
		chatgptHelp: "Offres Business, Enterprise et Edu",
		claudeCodeHelp: "Dans un terminal",
		otherName: "Une autre IA",
		otherHelp: "Tout client qui parle MCP",
		noAI: "Je n'ai pas encore d'IA",
		connect: (client) => `Branchez ${client}`,
		listed: "Cette Cahut est dans l'annuaire des connecteurs de Claude : rien à copier.",
		waitingFor: (client) => `En attente de ${client}`,
		comeBack: "Revenez sur cet onglet une fois connecté : la suite vous y attend.",
		plan: "Quelle offre ChatGPT avez-vous ?",
		planLead: "ChatGPT ne publie dans une Cahut qu'avec certaines offres.",
		planOrg: "Business, Enterprise ou Edu",
		planOwn: "Free, Go, Plus ou Pro",
		planUnsure: "Je ne sais pas",
		planCannot: "Avec cette offre, ChatGPT ne peut pas publier. Claude et Mistral Vibe le peuvent, même avec leur offre gratuite.",
		planWhich:
			"Business, Enterprise et Edu sont des offres d'organisation, qu'un administrateur de l'espace de travail gère. Une offre que vous payez vous-même, ou l'offre gratuite, est l'une des autres.",
		connectClaude: "Brancher Claude",
		connectVibe: "Brancher Mistral Vibe",
		computer: "Cette étape se fait sur un ordinateur",
		computerWhy: (client) =>
			client === "ChatGPT"
				? "ChatGPT ajoute un serveur MCP sur son site web, depuis un ordinateur."
				: `${client} tourne dans un terminal, sur un ordinateur.`,
		computerHow: "Sur votre ordinateur, ouvrez cette adresse et connectez-vous : le guide vous y attend, à cette même étape.",
		theAddress: "l'adresse du guide",
		showAnyway: "Afficher quand même les étapes ici",
		fromThisPhone: "Brancher plutôt Claude depuis ce téléphone",
		freeAI: "Commencez avec une IA gratuite",
		freeAILead: "Ces deux IA publient dans votre Cahut avec leur offre gratuite. Créez un compte, puis revenez ici.",
		claudeBy: "D'Anthropic. Sur le web, et en appli pour téléphone et ordinateur.",
		vibeBy: "De Mistral AI, l'ancien Le Chat.",
		createClaude: "Créer un compte Claude",
		createVibe: "Créer un compte Mistral",
		independent: "Cahut est indépendante d'Anthropic, de Mistral AI et d'OpenAI.",
		haveOne: "J'ai un compte, continuer",
		isConnected: (client) => `${client} est branché`,
		readPublish: "Il peut lire et publier.",
		readOnly: "Il ne peut que lire : pour publier, donnez-lui la lecture et la publication dans IA et accès.",
		ask: "Demandez votre première Window",
		askLead: "Copiez une idée dans votre IA, ou demandez ce que vous voulez.",
		waitingWindow: "En attente de votre première Window",
		open: (host) => `Ouvrir ${host}`,
		online: "Votre première Window est en ligne",
		itsAddress: "Son adresse",
		youAlone: "Pour l'instant, elle ne s'ouvre qu'à vous.",
		itsRoom: "Pour l'instant, elle s'ouvre aux membres de sa Room.",
		shareLink: "Un lien de partage l'ouvre à qui le détient, sans compte, jusqu'à la fin du lien.",
		openWindow: "Ouvrir la Window",
		finish: "Terminer",
		cardTitle: (n) => `Premiers pas · étape ${n} sur 3`,
		yourAI: "votre IA",
		cardConnect: "Branchez votre première IA",
		cardConnectHelp: "Votre IA publie sur Cahut par MCP. Choisissez-la, et suivez ses étapes.",
		cardAsk: "Demandez une première Window à votre IA",
		cardAskHelp: "Votre IA est branchée. Demandez-lui une Window, puis partagez-la.",
		resume: "Reprendre",
		hide: "Masquer",
	},
});

/** Requests that a person copies into their AI for a first Window. */
export const firstRequests = words({
	en: [
		{ label: "A page", text: "Make a page that presents my work in three parts, and publish it on my Cahut." },
		{ label: "A deck", text: "Prepare 5 slides on a subject I know well, ask me which, and publish them on my Cahut." },
		{ label: "A list to tick", text: "Make a shopping list that I can tick, and publish it on my Cahut." },
	],
	fr: [
		{ label: "Une page", text: "Fais une page qui présente mon activité en trois parties, et publie-la sur ma Cahut." },
		{ label: "Une présentation", text: "Prépare 5 diapositives sur un sujet que je connais bien, demande-moi lequel, et publie-les sur ma Cahut." },
		{ label: "Une liste à cocher", text: "Crée une liste de courses que je peux cocher, et publie-la sur ma Cahut." },
	],
});

const copyWords = words({
	en: { copy: "Copy", copied: "Copied" },
	fr: { copy: "Copier", copied: "Copiée" },
});

/** A request to copy into an AI, with its button. */
export function Request({ label, text }: { label?: string; text: string }) {
	const [copied, setCopied] = useState(false);
	const copy = async () => {
		try {
			await navigator.clipboard.writeText(text);
			setCopied(true);
			setTimeout(() => setCopied(false), 2000);
		} catch {
			// The text stays on the page, to select by hand.
		}
	};
	return (
		<div class="request">
			{label && <span class="c-label">{label}</span>}
			<q class="request__text">{text}</q>
			<div class="request__actions">
				<button class="c-btn c-btn--secondary c-btn--sm" type="button" onClick={copy}>
					{copied ? <Check class="i" size={14} aria-hidden="true" /> : <Copy class="i" size={14} aria-hidden="true" />}
					{copied ? copyWords.copied : copyWords.copy}
				</button>
			</div>
		</div>
	);
}

/** The key, in this browser, that says the person dismissed the guide. */
const hiddenKey = (handle: string) => `cahut.first-steps.${handle}`;

/** Whether the person dismissed the guide in this browser. */
export function hidden(handle: string): boolean {
	try {
		return localStorage.getItem(hiddenKey(handle)) === "hidden";
	} catch {
		return false;
	}
}

/** Dismisses the guide in this browser. */
export function hide(handle: string): void {
	try {
		localStorage.setItem(hiddenKey(handle), "hidden");
	} catch {
		// Without storage, the card comes back on the next visit.
	}
}

/** Where a person is in the guide, as the API says. */
export interface Progress {
	/** The AI that acts for the person, connected over OAuth, if any. */
	ai?: Connection;
	/** Whether an AI acts for the person, over OAuth or with a token. */
	connected: boolean;
	/** The Windows first published for the person, newest first. */
	mine: Window[];
}

/** Says where the person handle is, from their connections, tokens and Windows. */
export function progressOf(handle: string, connections: Connection[], tokens: TokenInfo[], windows: Window[]): Progress {
	const ai = connections.find((c) => c.status === "active");
	const mine = windows.filter((w) => w.created_by === handle).sort((a, b) => b.created_at.localeCompare(a.created_at));
	return { ai, connected: ai !== undefined || tokens.some((k) => !k.revoked_at), mine };
}

/** The step the person is at: 1 until an AI acts for them, 2 until it has published, then 3. */
export function stepOf(p: Progress): 1 | 2 | 3 {
	return !p.connected ? 1 : p.mine.length === 0 ? 2 : 3;
}

/** How often the guide reads where the person is. */
const every = 3_000;

/** Reads where the person is, every three seconds while the page is visible, and as soon as it is again. */
function useProgress(): Progress | undefined {
	const api = useConsole();
	const [progress, setProgress] = useState<Progress>();
	useEffect(() => {
		let live = true;
		const ask = () => {
			if (document.visibilityState !== "visible") return;
			const user = api.session.handle;
			Promise.all([api.cahut("list_connected_ais", { user }), api.cahut("list_tokens", { user }), api.cahut("list_windows", {})]).then(
				([c, k, w]) => live && setProgress(progressOf(user, c.connections, k.tokens, w.windows)),
				() => undefined,
			);
		};
		ask();
		const timer = setInterval(ask, every);
		document.addEventListener("visibilitychange", ask);
		return () => {
			live = false;
			clearInterval(timer);
			document.removeEventListener("visibilitychange", ask);
		};
	}, []);
	return progress;
}

/** Whether the console runs on a phone: a narrow screen that one touches. */
const phone = matchMedia("(pointer: coarse) and (max-width: 760px)").matches;

const clients: ClientId[] = ["claude", "vibe", "chatgpt", "claude-code", "other"];

/** The clients that a phone does not connect. */
const computerOnly: ClientId[] = ["chatgpt", "claude-code"];

/** The guide, at path, which starts with /start. */
export function Start({ path }: { path: string }) {
	const api = useConsole();
	const progress = useProgress();
	const [about, setAbout] = useState<About>();
	useEffect(() => {
		api.cahut("get_cahut", {}).then(setAbout, () => undefined);
	}, []);
	const page = path.replace(/^\/start\/?/, "").replace(/\/$/, "");
	const connected = progress?.connected ?? false;

	// Once the AI is connected, the steps of the clients give way to the
	// next step.
	useEffect(() => {
		if (connected && page !== "") navigate("/start", true);
	}, [connected, page]);

	if (!progress || !about) return <Frame step={1} />;
	const mcp = about.mcp ?? `${location.origin}/mcp`;
	const skip = () => {
		hide(api.session.handle);
		navigate("/");
	};
	if (page === "") {
		const step = stepOf(progress);
		if (step === 1) return <Welcome onLater={() => navigate("/")} />;
		if (step === 2) return <Ask progress={progress} onSkip={skip} />;
		return <Share window={progress.mine[0]} individual={about.kind === "individual"} onFinish={skip} />;
	}
	if (page === "ai") return <Choose onSkip={skip} />;
	if (page === "none") return <NoAI onSkip={skip} />;
	const client = clients.find((c) => c === page);
	if (!client) return <Redirect to="/start" />;
	return <Client key={client} client={client} mcp={mcp} listing={about.listings.claude} onSkip={skip} />;
}

/** The frame of every screen of the guide: its head, its body, and its foot under the thumb. */
function Frame(props: { step: 1 | 2 | 3; back?: string; onSkip?: () => void; dock?: ComponentChildren; children?: ComponentChildren }) {
	return (
		<div class="start">
			<header class="start__head">
				{props.back ? (
					<Link class="c-btn c-btn--secondary c-btn--sm icon-button" href={props.back} aria-label={t.back} title={t.back}>
						<ArrowLeft class="i" size={18} aria-hidden="true" />
					</Link>
				) : (
					<Link class="start__logo" href="/" aria-label={t.home}>
						<Logo height={22} />
					</Link>
				)}
				<span class="c-label">{t.step(props.step)}</span>
				{props.onSkip && (
					<button class="c-btn c-btn--ghost c-btn--sm" type="button" onClick={props.onSkip}>
						{t.skip}
					</button>
				)}
			</header>
			<main class="start__main">
				<div class="start__bar" aria-hidden="true">
					{[1, 2, 3].map((n) => (
						<span key={n} data-on={n <= props.step ? "" : undefined} />
					))}
				</div>
				{props.children}
			</main>
			{props.dock && (
				<div class="start__dock">
					<div class="start__dock-inner">{props.dock}</div>
				</div>
			)}
		</div>
	);
}

/** A line that says what the guide waits for, as it reads the API. */
function Waiting({ what, help }: { what: string; help?: string }) {
	return (
		<div class="live" role="status">
			<span class="live__dot" aria-hidden="true" />
			<span>
				{what}
				{help && <small>{help}</small>}
			</span>
		</div>
	);
}

/** A in the proposal: the welcome, once the password is chosen. */
function Welcome({ onLater }: { onLater: () => void }) {
	const { session } = useConsole();
	return (
		<Frame
			step={1}
			dock={
				<>
					<Link class="c-btn c-btn--primary" href="/start/ai">
						{t.start}
					</Link>
					<button class="c-btn c-btn--ghost" type="button" onClick={onLater}>
						{t.later}
					</button>
				</>
			}
		>
			<Pictogram name="window" large />
			<div class="start__head-text">
				<span class="c-label">{t.firstSteps}</span>
				<h1 class="c-h2">{t.ready(session.name)}</h1>
				<p class="c-muted">{t.readyLead}</p>
			</div>
			<ol class="plan">
				<li>
					<b>{t.planConnect}</b>
					<span>{t.planConnectHelp}</span>
				</li>
				<li>
					<b>{t.planAsk}</b>
					<span>{t.planAskHelp}</span>
				</li>
				<li>
					<b>{t.planShare}</b>
					<span>{t.planShareHelp}</span>
				</li>
			</ol>
			<div class="start__address">
				<span class="c-label">{t.yourWindows}</span>
				<code class="c-mono">
					{location.host}/{session.handle}/
				</code>
			</div>
		</Frame>
	);
}

const choices: Record<ClientId, { mark: string; help: string; name?: string }> = {
	claude: { mark: "Cl", help: t.claudeHelp },
	vibe: { mark: "Mi", help: t.vibeHelp },
	chatgpt: { mark: "GP", help: t.chatgptHelp },
	"claude-code": { mark: ">_", help: t.claudeCodeHelp },
	other: { mark: "··", help: t.otherHelp, name: t.otherName },
};

/** The initials of a client, in place of a mark that is not ours to show. */
function Mark({ client }: { client: ClientId }) {
	return (
		<span class="mark" aria-hidden="true">
			{choices[client].mark}
		</span>
	);
}

function Choice({ client }: { client: ClientId }) {
	const c = choices[client];
	return (
		<Link class="ai-choice" href={`/start/${client}`}>
			<Mark client={client} />
			<span class="ai-choice__text">
				<b>{c.name ?? clientNames[client]}</b>
				<small>{c.help}</small>
			</span>
			<ChevronRight class="i" size={18} aria-hidden="true" />
		</Link>
	);
}

/** B: which AI. On a phone, those that need a computer come after the others. */
function Choose({ onSkip }: { onSkip: () => void }) {
	const here = phone ? clients.filter((c) => !computerOnly.includes(c) && c !== "other") : clients;
	const there = phone ? clients.filter((c) => computerOnly.includes(c) || c === "other") : [];
	return (
		<Frame step={1} back="/start" onSkip={onSkip}>
			<div class="start__head-text">
				<h1 class="c-h2">{t.which}</h1>
				<p class="c-muted">{t.whichLead}</p>
			</div>
			{phone && (
				<span class="start__group c-label">
					<Smartphone class="i" size={14} aria-hidden="true" />
					{t.fromPhone}
				</span>
			)}
			<div class="ai-choices">
				{here.map((c) => (
					<Choice key={c} client={c} />
				))}
			</div>
			{there.length > 0 && (
				<>
					<span class="start__group c-label">
						<Monitor class="i" size={14} aria-hidden="true" />
						{t.onComputer}
					</span>
					<div class="ai-choices">
						{there.map((c) => (
							<Choice key={c} client={c} />
						))}
					</div>
				</>
			)}
			<Link class="c-link c-small" href="/start/none">
				{t.noAI}
			</Link>
		</Frame>
	);
}

/** C, H and I: the steps of one client, with the plan of ChatGPT first, and on a phone the clients that need a computer. */
function Client({ client, mcp, listing, onSkip }: { client: ClientId; mcp: string; listing?: string; onSkip: () => void }) {
	const name = clientNames[client];
	// Another client is the person's own AI, which the guide cannot name.
	const ai = client === "other" ? t.yourAI : name;
	const [plan, setPlan] = useState<"org" | "own" | "unsure">();
	const [anyway, setAnyway] = useState(false);
	const dock = <Waiting what={t.waitingFor(ai)} help={t.comeBack} />;
	const head = (
		<div class="start__head-text">
			<span class="chip">
				<Mark client={client} />
				{client === "other" ? t.otherName : name}
			</span>
			<h1 class="c-h2">{t.connect(ai)}</h1>
			{client === "claude" && listing && <p class="c-muted">{t.listed}</p>}
		</div>
	);

	if (client === "chatgpt" && plan !== "org") {
		const option = (value: "org" | "own" | "unsure", label: string) => (
			<label class="option">
				<input type="radio" name="plan" checked={plan === value} onChange={() => setPlan(value)} />
				<span>{label}</span>
			</label>
		);
		return (
			<Frame
				step={1}
				back="/start/ai"
				onSkip={onSkip}
				dock={
					plan === "own" && (
						<>
							<Link class="c-btn c-btn--primary" href="/start/claude">
								{t.connectClaude}
							</Link>
							<Link class="c-btn c-btn--secondary" href="/start/vibe">
								{t.connectVibe}
							</Link>
						</>
					)
				}
			>
				<div class="start__head-text">
					<span class="chip">
						<Mark client="chatgpt" />
						ChatGPT
					</span>
					<h1 class="c-h2">{t.plan}</h1>
					<p class="c-muted">{t.planLead}</p>
				</div>
				<fieldset class="options">
					<legend class="visually-hidden">{t.plan}</legend>
					{option("org", t.planOrg)}
					{option("own", t.planOwn)}
					{option("unsure", t.planUnsure)}
				</fieldset>
				{plan === "own" && <Notice>{t.planCannot}</Notice>}
				{plan === "unsure" && <Notice>{t.planWhich}</Notice>}
			</Frame>
		);
	}

	if (phone && computerOnly.includes(client) && !anyway) {
		return (
			<Frame
				step={1}
				back="/start/ai"
				onSkip={onSkip}
				dock={
					<Link class="c-btn c-btn--secondary" href="/start/claude">
						{t.fromThisPhone}
					</Link>
				}
			>
				{head}
				<h2 class="c-h4">{t.computer}</h2>
				<p class="c-small c-muted">{t.computerWhy(name)}</p>
				<p class="c-small">{t.computerHow}</p>
				<CopyField label={t.theAddress} value={`${location.origin}/console/start/${client}`} />
				<button class="c-btn c-btn--ghost c-btn--sm start__anyway" type="button" onClick={() => setAnyway(true)}>
					{t.showAnyway}
				</button>
			</Frame>
		);
	}

	return (
		<Frame step={1} back="/start/ai" onSkip={onSkip} dock={dock}>
			{head}
			<ClientSteps client={client} mcp={mcp} listing={listing} />
			<Film client={client} listing={listing} />
		</Frame>
	);
}

/** J: no AI yet. */
function NoAI({ onSkip }: { onSkip: () => void }) {
	return (
		<Frame
			step={1}
			back="/start/ai"
			onSkip={onSkip}
			dock={
				<Link class="c-btn c-btn--primary" href="/start/ai">
					{t.haveOne}
				</Link>
			}
		>
			<div class="start__head-text">
				<h1 class="c-h2">{t.freeAI}</h1>
				<p class="c-muted">{t.freeAILead}</p>
			</div>
			<div class="c-card panel start__ai">
				<span class="chip">
					<Mark client="claude" />
					Claude
				</span>
				<p class="c-small c-muted">{t.claudeBy}</p>
				<Out href="https://claude.ai/">{t.createClaude}</Out>
			</div>
			<div class="c-card panel start__ai">
				<span class="chip">
					<Mark client="vibe" />
					Mistral Vibe
				</span>
				<p class="c-small c-muted">{t.vibeBy}</p>
				<Out href="https://mistral.ai/">{t.createVibe}</Out>
			</div>
			<p class="c-small c-muted">{t.independent}</p>
		</Frame>
	);
}

function Out({ href, children }: { href: string; children: ComponentChildren }) {
	return (
		<div>
			<a class="c-btn c-btn--secondary" href={href} target="_blank" rel="noopener noreferrer">
				{children}
				<ExternalLink class="i" size={16} aria-hidden="true" />
			</a>
		</div>
	);
}

/** The address of the site of a client, from the host its browser went back to, when it has one. */
function siteOf(host: string): string | undefined {
	return /^[a-z0-9-]+(\.[a-z0-9-]+)+$/i.test(host) && !/^(localhost|127\.)/.test(host) ? `https://${host}/` : undefined;
}

/** F: the AI is connected; it waits for the first Window. */
function Ask({ progress, onSkip }: { progress: Progress; onSkip: () => void }) {
	const { ai } = progress;
	const site = ai ? siteOf(ai.redirect_host) : undefined;
	return (
		<Frame
			step={2}
			onSkip={onSkip}
			dock={
				<>
					<Waiting what={t.waitingWindow} />
					{site && ai && (
						<a class="c-btn c-btn--accent" href={site} target="_blank" rel="noopener noreferrer">
							{t.open(ai.redirect_host)}
							<ExternalLink class="i" size={16} aria-hidden="true" />
						</a>
					)}
				</>
			}
		>
			{ai && (
				<p class="start__done">
					<CircleCheck class="i" size={18} aria-hidden="true" />
					<span>
						{t.isConnected(ai.client)}. {ai.scope === "read" ? t.readOnly : t.readPublish}
					</span>
				</p>
			)}
			<div class="start__head-text">
				<h1 class="c-h2">{t.ask}</h1>
				<p class="c-muted">{t.askLead}</p>
			</div>
			<div class="requests">
				{firstRequests.map((r) => (
					<Request key={r.label} label={r.label} text={r.text} />
				))}
			</div>
		</Frame>
	);
}

/** G: the first Window, its address, and the menu that shares it. */
function Share({ window: w, individual, onFinish }: { window: Window; individual: boolean; onFinish: () => void }) {
	return (
		<Frame
			step={3}
			dock={
				<button class="c-btn c-btn--primary" type="button" onClick={onFinish}>
					{t.finish}
				</button>
			}
		>
			<h1 class="c-h2">{t.online}</h1>
			<WindowCards windows={[w]} />
			<div class="start__address">
				<span class="c-label">{t.itsAddress}</span>
				<code class="c-mono">{w.url}</code>
			</div>
			<p class="c-small c-muted">
				{w.access === "room" && `${individual ? t.youAlone : t.itsRoom} `}
				{t.shareLink}
			</p>
			<div class="start__share">
				<ShareMenu window={w} />
				<a class="c-btn c-btn--secondary" href={w.url} target="_blank" rel="noopener noreferrer">
					{t.openWindow}
					<ExternalLink class="i" size={16} aria-hidden="true" />
				</a>
			</div>
		</Frame>
	);
}

/** K: on the overview, the step the person is at, until their AI has published or they dismiss the guide. */
export function FirstStepsCard({ progress, onHide }: { progress: Progress; onHide: () => void }) {
	const step = stepOf(progress);
	return (
		<section class="section c-card panel first-steps">
			<span class="c-label">{t.cardTitle(step)}</span>
			<div class="start__bar" aria-hidden="true">
				{[1, 2, 3].map((n) => (
					<span key={n} data-on={n <= step ? "" : undefined} />
				))}
			</div>
			<h2 class="c-h4">{step === 1 ? t.cardConnect : t.cardAsk}</h2>
			<p class="c-small c-muted">{step === 1 ? t.cardConnectHelp : t.cardAskHelp}</p>
			<div class="first-steps__actions">
				<Link class="c-btn c-btn--primary" href="/start">
					{t.resume}
				</Link>
				<button class="c-btn c-btn--ghost" type="button" onClick={onHide}>
					{t.hide}
				</button>
			</div>
		</section>
	);
}
