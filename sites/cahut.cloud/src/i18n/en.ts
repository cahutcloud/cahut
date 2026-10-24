// The words of cahut.cloud in English, the language of the launch. Every
// other language follows the shape of this file (see ./index.ts). Short HTML,
// such as <b>, is allowed where a component renders the text with set:html.

export default {
	meta: {
		title: 'Cahut: the home for everything your AIs make',
		description:
			'Documents, slide decks, dashboards and small apps live online as living pages, in your house style, on your own server, whichever AI made them. Cahut is in early development.',
		imageAlt:
			'Your own cloud for everything your AI makes. Documents, decks, dashboards and apps, in your house style, on your own server.',
		locale: 'en_US',
	},

	base: {
		skip: 'Skip to content',
		home: 'Cahut, home',
		mainNav: 'Main',
		nav: {
			quickstart: 'Get started',
			how: 'How it works',
			knock: 'Knock first',
			features: 'Features',
			docs: 'Docs',
			brand: 'Brand',
			project: 'Project',
			github: 'GitHub',
		},
		footer: { cahut: 'Cahut', project: 'Project' },
		signature: 'Your own cloud for everything your AI makes.',
		about:
			'Cahut is a personal open source project by Olivier Durand, independent of any company and of any AI vendor. It is in early development.',
		theme: { dark: 'Switch to the dark theme', light: 'Switch to the light theme' },
		language: { label: 'Version française', short: 'FR' },
	},

	hero: {
		status: 'In early development: the first version works',
		title: 'The home for everything your AIs make.',
		lead: 'Documents, slide decks, dashboards and small apps live online as living pages, in your house style, on your own server, whichever AI made them.',
		start: 'Get started',
		github: 'Follow the work on GitHub',
		facts: ['One command on your server', 'Any AI, over MCP', 'A link, not a file'],
		caption: 'A document, a dashboard, a site and a picture join a cloud with an open door, lit from inside.',
		log: { action: 'published', done: 'Done' },
	},

	quickstart: {
		label: 'Quick start',
		title: 'One command, then ask your AI.',
		lead: 'Cahut installs on a Linux server as a service, with its HTTPS certificate. Your AI connects over MCP and publishes its first Window in a few minutes.',
		allOptions: 'All install options',
		install: 'Install',
		tabs: {
			server: { label: 'On a server', meta: 'AlmaLinux · Ubuntu · amd64 · arm64' },
			local: { label: 'On your machine', meta: 'Git · make · Go' },
		},
		comments: {
			install: 'Install Cahut as a service, with your first Cahut',
			account: 'Open the link it prints: your account, your house style, your AI',
			connect: 'Or add your Cahut to Claude Code yourself, and sign in',
			build: 'Build Cahut from source',
			declare: 'Declare a Cahut at localhost, with your account and a token',
			serve: 'Serve it on http://localhost:8080',
			connectLocal: 'Add it to Claude Code, with the token',
			ask: 'Then ask your AI',
		},
		prompt: 'Make a one-page summary of this project as a document, and publish it on Cahut.',
		reply: 'Published at',
		copy: 'Copy',
		copied: 'Copied',
		needs: {
			label: 'You need',
			items: [
				'A Linux server with systemd',
				'A domain that points to it, or its IPv4 address',
				'Ports 80 and 443 open',
				'An AI client that speaks MCP, such as Claude Code',
			],
		},
		note: 'The repository opens at the public launch: until then, these commands work for those who have access to it.',
		status: 'Where Cahut stands',
	},

	why: {
		label: 'Why Cahut',
		title: 'The file is going away.',
		lead: 'Word, PowerPoint and Excel were designed for humans who write, lay out and calculate by hand. When an AI produces the content, it naturally produces living pages, each with its own address.',
		text: 'You no longer send an attachment, you share a link: the page updates, takes comments, keeps its versions, and exports to PDF or PowerPoint for whoever still needs a file.',
		problemsTitle: 'Today, what AIs make has no home.',
		problems: [
			{
				title: 'Scattered',
				text: 'A Claude link here, a ChatGPT link there, a Gamma deck somewhere else. Each AI keeps what it makes on its own service.',
			},
			{
				title: 'Converted back',
				text: 'For lack of anything better, the result is turned back into a PowerPoint or a Word file so it can be shared and filed.',
			},
			{
				title: 'Uncontrolled',
				text: 'Who has access, until when, which version is the right one, where the data lives. Nobody really knows.',
			},
			{ title: 'Inconsistent', text: "Every AI has its own style, and the organization's brand gets lost." },
			{
				title: 'Unguarded',
				text: 'Handing a server to an AI means giving it a root shell, or going through a platform built for humans that it can barely use.',
			},
		],
	},

	files: {
		compare: 'Compare',
		files: 'Today: files',
		link: 'With Cahut: a link',
		today: 'Today',
		withCahut: 'With Cahut',
		items: [
			{ name: 'Q3 review v4 FINAL.pptx', meta: 'Attached, 8.2 MB' },
			{ name: 'Q3 review v5 (2).pptx', meta: 'Forwarded twice' },
			{ name: 'q3-notes.docx', meta: 'Converted back' },
			{ name: 'A chat link', meta: 'On one AI service' },
			{ name: 'Q3 numbers.xlsx', meta: 'Which version?' },
			{ name: 'A deck elsewhere', meta: 'On another service' },
		],
		facts: ['Updated 2 minutes ago', 'Version 2, the previous one kept', '3 comments'],
		captionFiles: 'Six files and links for one deck. Which one is right, and who can see it?',
		captionLink: 'One address. The page updates, takes comments, keeps its versions, and exports to PDF or PowerPoint.',
	},

	statement: {
		label: 'The idea',
		title: 'Your AIs make it, Cahut keeps it alive.',
		lead: 'Everything they produce is published, shared, organized and protected in one place, in your colors, on your server, whatever the AI.',
		text: 'Cahut does not create content, it hosts it. It does not compete with AIs or creation tools: it is the neutral, self-hosted home where their work lives.',
		layers: [
			{
				layer: 'Create',
				who: 'The AIs',
				what: 'Claude, ChatGPT, Mistral, Copilot and others make the content: documents, decks, dashboards, apps.',
				tags: ['Claude', 'ChatGPT', 'Mistral', 'Copilot'],
			},
			{
				layer: 'Keep and share',
				who: 'Cahut, on your server',
				what: 'Publishing, links, permissions, versions, comments, search, export and the Logbook.',
				tags: ['Publish', 'Share', 'Versions', 'Comments', 'Search', 'Export', 'Logbook'],
			},
			{
				layer: 'Look and feel',
				who: 'The house style, carried by Cahut',
				what: 'Your colors and formats, given to every AI as templates and skills.',
				tags: ['Colors', 'Fonts', 'Logos', 'Templates'],
			},
		],
	},

	how: {
		label: 'How it works',
		title: 'From a request to a link, in five steps.',
		lead: 'The AIs stay outside. They connect to Cahut over MCP, drop their work through bounded operations, then disconnect. Your readers open it in a browser.',
	},

	flow: {
		group: 'How Cahut works, in five steps',
		step: 'Step',
		of: 'of 5',
		pause: 'Pause',
		play: 'Play',
		steps: [
			{
				title: 'Connect',
				heading: 'Add Cahut to your AI, once.',
				text: 'Your AI connects to your Cahut over MCP, the open protocol that links AIs to tools. You authorize it with OAuth, with the scope you choose. No AI model runs on your server.',
				call: 'MCP · OAuth · on behalf of Olivier',
			},
			{
				title: 'Publish',
				heading: 'Ask for a deck. The AI publishes a Window.',
				text: 'The AI calls one typed operation, never a shell. Cahut applies your house style, files the Window in its Room, and hands back its address.',
				call: 'publish_window(room: "marketing", window: "q3-review")',
			},
			{
				title: 'Share',
				heading: 'Share a link, not a file.',
				text: 'Each Window opens to its Room only, to everyone in your Cahut, to anyone with its Key, or to the public. Readers open it in a browser and comment on it.',
				call: 'share_window(window: "q3-review", access: "key")',
			},
			{
				title: 'Knock first',
				heading: 'What cannot be undone waits for you.',
				text: 'Making a Window public cannot be undone. So the AI knocks: the request waits in the console, with a notification and an expiry, until a person approves it.',
				call: 'share_window(access: "public") → waiting for Olivier',
			},
			{
				title: 'Logbook',
				heading: 'Every action is in the Logbook.',
				text: 'Which AI, on whose behalf, what it did and with what result. The Logbook keeps every action, readable in the console.',
				call: 'logbook: 4 entries, 1 approved by a person',
			},
		],
		ais: 'Your AIs',
		aisNote: 'They stay outside. No AI runs on your server.',
		server: 'Your server',
		cahut: 'Your Cahut',
		houseStyle: 'House style',
		room: 'Room · marketing',
		noWindows: 'No Windows yet',
		knock: 'Knock first',
		nothingWaits: 'Nothing waits for you',
		waiting: 'Waiting',
		approved: 'Approved',
		knockText: 'Wants to make <b>Q3 review</b> public.',
		approve: 'Approve',
		decline: 'Decline',
		knockDone: 'Approved by Olivier, 09:33',
		logbook: 'Logbook',
		log: [
			'connected for Olivier',
			'published <b>Q3 review</b>',
			'gave <b>Q3 review</b> a Key',
			'asked to make <b>Q3 review</b> public',
		],
		readers: 'Your readers',
		readerNames: [
			{ who: 'Marketing team', short: 'Team' },
			{ who: 'A client, with the Key', short: 'Key' },
			{ who: 'Anyone', short: 'Public' },
		],
		closed: 'Closed',
		comment: 'Looks good',
		access: { room: 'Room only', internal: 'Internal', key: 'Key', public: 'Public' },
		packets: { linkKey: 'Link and Key', link: 'Link' },
	},

	knock: {
		label: 'Governance',
		title: 'Knock first',
		lead: 'The AI is the intended operator of Cahut, but never without guardrails. What can be undone, it does alone. What cannot waits for a human.',
		reversible: 'Reversible',
		alone: 'The AI does it alone',
		aloneItems: [
			'Publish or update a Window (the previous version is kept)',
			'Restrict access, extend a lifetime',
			'Deploy an app within the set limits',
			'Roll back to the previous version',
			'Run a backup',
		],
		irreversible: 'Irreversible',
		approves: 'A human approves',
		approvedItems: [
			'Permanently delete a Window, an app or a database',
			'Make a Window public',
			'Add a person to a Cahut',
			'Change the permissions of a user or an AI',
			'Erase backups or change their destination',
		],
		try: 'Try it: answer the request.',
		principlesTitle: 'Seven principles hold it together.',
		principles: [
			{
				title: 'Bounded operations',
				text: 'The AI acts through typed operations (publish a page, create a database, share a link), never through a free shell, over MCP as over SSH.',
			},
			{
				title: 'Reversible or approved',
				text: 'What can be undone, the AI does alone. What cannot waits for a human to approve it in the console, with a notification and an expiry.',
			},
			{
				title: 'The AI declares what it publishes',
				text: 'Each publication states what it contains and what it does: an app that collects data, calls out, opens a form.',
			},
			{
				title: 'Per-agent permissions',
				text: 'Each AI acts on behalf of a person, with the scope that person grants, revocable at any time.',
			},
			{
				title: 'Everything is logged',
				text: 'Which AI, on whose behalf, which action, which result: the Logbook is readable in the console.',
			},
			{
				title: 'Data stays home',
				text: 'Everything lives on your server. Backups only go to the external storage you choose.',
			},
			{
				title: 'Isolation',
				text: "Each Window runs in a sandbox: its scripts cannot read the other Windows, the console or the reader's session. Each app runs in an unprivileged container with its own resource limits. The Cahuts of one server share nothing but the machine.",
			},
		],
	},

	demo: {
		console: 'the console of your Cahut',
		waitingCount: '1 waiting',
		nothingWaits: 'Nothing waits',
		waiting: 'Waiting',
		request: 'Wants to make <b>Q3 review</b> public. Today it opens to anyone with its Key.',
		meta: 'expires in 23 h · on behalf of Olivier',
		approve: 'Approve',
		decline: 'Decline',
		logbook: 'Logbook · today',
		log: ['asked to make <b>Q3 review</b> public', 'gave <b>Q3 review</b> a Key', 'published <b>Q3 review</b>'],
		done: 'Done',
		again: 'Ask again',
		answers: {
			approve: {
				badge: 'Approved',
				result: 'Approved. Q3 review is public, and Claude knows it.',
				log: 'approved making <b>Q3 review</b> public',
				done: 'Done',
			},
			decline: {
				badge: 'Declined',
				result: 'Declined. Q3 review keeps its Key, and Claude knows it.',
				log: 'declined making <b>Q3 review</b> public',
				done: 'Declined',
			},
		},
	},

	features: {
		label: 'Features',
		title: 'What ships at launch',
		lead: 'The full scope ships at launch, with no reduced trial version. Today, the first slice works: an AI publishes a Window over MCP, and Cahut serves it over HTTPS. The rest is the target.',
		groups: [
			{
				title: 'Publish',
				text: 'What the AIs make, online at its own address.',
				features: [
					{
						area: 'Pages',
						text: 'Documents, decks, dashboards and static sites, each at its own address in its Room.',
					},
					{
						area: 'Apps',
						text: 'Predefined runtimes, such as Node and Python, and an advanced Dockerfile mode. Variables and secrets, PostgreSQL or SQLite databases, runtime logs, resource limits.',
					},
					{ area: 'Versions', text: 'Two versions kept per publication. Roll back in one gesture.' },
					{ area: 'Search', text: 'Find any page, deck or dashboard across everything published.' },
					{ area: 'Export', text: 'PDF and PowerPoint, for recipients still on Office.' },
				],
			},
			{
				title: 'Connect',
				text: 'Every AI, one door, your style.',
				features: [
					{
						area: 'AI access',
						text: 'A native MCP server (HTTP, OAuth) and one skill per AI client, packaged as a plugin where the client allows it: Claude first, then ChatGPT, Mistral, Copilot and others. SSH access limited to the same operations, for terminal AIs.',
					},
					{
						area: 'House style',
						text: "The organization's colors, fonts, logos and templates, given to every AI so everything comes out on brand.",
					},
				],
			},
			{
				title: 'Share',
				text: 'A link, opened to the right people.',
				features: [
					{
						area: 'Sharing',
						text: 'By link, Window by Window: its Room only, internal (everyone in the Cahut), a Key (a password) or public, with an optional end date.',
					},
					{ area: 'Comments', text: 'Reader feedback directly on the pages.' },
					{
						area: 'Users and Rooms',
						text: 'The people of each Cahut and its admins, a personal Room for each person, team Rooms with owners and members.',
					},
					{
						area: 'Sign-in',
						text: 'Accounts of each Cahut, and single sign-on with OpenID Connect: Google, Microsoft Entra, Okta and others.',
					},
				],
			},
			{
				title: 'Run',
				text: 'One server of your own, kept in order.',
				features: [
					{
						area: 'Installation',
						text: 'One command on a fresh VPS or dedicated server (AlmaLinux or Ubuntu), a guided first start, and the linking of the admin and their first AI.',
					},
					{
						area: 'Web console',
						text: 'See what is published, share, comment, approve irreversible actions, administer the server and its Cahuts, on its own HTTPS port.',
					},
					{
						area: 'Domains and TLS',
						text: 'Several Cahuts per server, each at its own domain or subdomain. No wildcard needed, automatic certificates.',
					},
					{ area: 'Logbook', text: 'The history of every action: who, which AI, on whose behalf, with what result.' },
					{ area: 'Backups', text: 'Optional, to external S3 or SFTP storage. Restore from the console.' },
					{ area: 'Updates', text: 'Automatic, checked against the releases published on GitHub, with rollback.' },
				],
			},
		],
		notDoneTitle: 'What Cahut does not do',
		notDone: [
			"Create content: that is the AIs' job.",
			'Run an AI model on the server: the AIs stay outside and connect to it.',
			'Depend on one AI vendor, or lock work into a proprietary format.',
			'Give an AI free access to the server.',
			'Run across several machines: one server holds one or more Cahuts.',
		],
	},

	vocabulary: {
		label: 'Vocabulary',
		title: 'The language of the house',
		lead: 'Cahut speaks the language of the house, in plain English. Start with an address: it reads like a house.',
		parts: [
			{
				word: 'Cahut',
				meaning: 'Your instance, at its own domain. One server can hold several, each with its own people and database.',
			},
			{ word: 'Room', meaning: 'A personal or team space, with its members. Each person has their own Room.' },
			{
				word: 'Window',
				meaning: 'A living page: document, deck, dashboard or app. It keeps its previous version.',
			},
		],
		words: [
			{ word: 'Key', meaning: 'The password that opens a protected Window.' },
			{ word: 'House style', meaning: "The organization's brand, applied by every AI." },
			{ word: 'Knock first', meaning: 'The governance rule: the AI knocks before any irreversible action.' },
			{ word: 'Logbook', meaning: 'The record of every action.' },
		],
		accessTitle: 'Access',
		accessText: 'Each Window opens to one of four circles. You choose, Window by Window.',
		access: [
			'The members of its Room. The default.',
			'Everyone signed in to your Cahut.',
			'Anyone who has its Key.',
			'Anyone with the link. Knocks first.',
		],
	},

	roadmap: {
		label: 'Roadmap',
		title: 'Where Cahut stands',
		lead: 'Cahut is a personal project by Olivier Durand, independent of any company and of any AI vendor. It is written in Go, in the open, one step at a time.',
		steps: [
			'Protect the name: GitHub organization, package names, trademark search.',
			'Write the target architecture: compare the options and settle them, with their reasons.',
			'Build the first vertical slice in Go: install on a server, publish a page over MCP, serve it over HTTPS.',
			'Proof of concept: Cahut installed on a fresh VPS, an AI publishing a deck over MCP, the console.',
			'Roadmap to the public launch.',
		],
		underWay: 'Under way',
		more: 'The full roadmap',
		criteriaTitle: 'Ready for launch when',
		criterion: 'Criterion',
		threshold: 'Threshold',
		criteria: [
			{ criterion: 'Install on a fresh VPS', threshold: 'Under 10 minutes, one command' },
			{ criterion: 'First publication by an AI', threshold: 'Under 5 minutes, without touching the server' },
			{ criterion: 'Sharing by a non-technical person', threshold: 'A deck shared by link in a few minutes' },
			{ criterion: 'Supported AIs', threshold: 'The same Cahut works with at least two different AI clients' },
			{
				criterion: 'Governance',
				threshold: 'Every AI action is in the Logbook, and is either reversible or was approved',
			},
		],
	},

	project: {
		label: 'Project',
		nav: {
			governance: 'Governance',
			roadmap: 'Roadmap',
			contribute: 'Contribute',
			support: 'Get help',
			security: 'Security',
			conduct: 'Code of conduct',
		},
		descriptions: {
			governance: 'Who decides in the Cahut project, its principles, its roles, how it grows, and what the money does.',
			roadmap: 'The batches of the first version of Cahut, what works today, and what makes it ready for launch.',
			contribute:
				'What helps Cahut right away, what to discuss first, how a pull request gets in, and the rules for contributions made with an AI.',
			support: 'Where to ask for help with Cahut, how to ask well, and what to expect.',
			security: 'How to report a vulnerability in Cahut, privately, what to expect, and what is in scope.',
			conduct: 'The code of conduct of the Cahut community, the Contributor Covenant 3.0, and where to report an incident.',
		},
		// Under the content of a page: the file of the repository it comes from.
		source: {
			file: 'This page is {file} at the root of the repository, rendered at each build of the site.',
			translation: 'This page translates {file} at the root of the repository, which is the reference, in English.',
			conduct:
				'This page is the official translation of the Contributor Covenant 3.0, with the reporting channel of Cahut. {file} at the root of the repository is the reference, in English.',
			change: 'To change it, open a pull request.',
		},
		roadmap: {
			title: 'Roadmap',
			lead: 'Cahut is built in batches, each one a working slice from the AI to the reader. Their order is set, their dates are not: the milestones of the repository say where each batch stands.',
			batchesTitle: 'The batches of the first version',
			batch: 'Batch',
			states: { done: 'Done', next: 'Next' },
			batches: [
				{
					title: 'Foundations and decisions',
					text: 'The first architecture records, the Go skeleton, the continuous integration and the release build.',
				},
				{
					title: 'Publish over MCP, serve over HTTPS',
					text: 'An AI publishes a Window with a personal token. Cahut serves it at its own address, in a sandbox, and the call goes to the Logbook.',
				},
				{
					title: 'People, Rooms and the console',
					text: 'People sign in; administrators of each Cahut and of the server; team Rooms with their rights; the web console and its API.',
				},
				{
					title: 'Connect the AIs: OAuth',
					text: 'Each Cahut is the authorization server of its MCP endpoint. On a consent screen, a person grants an AI some Rooms and a scope, and sees the connected AIs in the console.',
				},
				{
					title: 'Complete Windows: versions and access',
					text: 'Updating a Window keeps the previous version, and one step rolls back. Each Window opens to its Room only, to everyone signed in to its Cahut, to anyone with its Key, or to the public.',
				},
				{
					title: 'Knock first',
					text: 'An AI asks before anything it cannot undo. The request waits in the console, with a notification and an expiry, until a person approves or declines it.',
				},
				{
					title: 'House style and a skill for Claude',
					text: "The organization's colors, fonts, logo and templates, given to every AI, and a skill that tells Claude how to publish on Cahut.",
				},
				{
					title: 'Installation, first start and updates',
					text: 'One command on a fresh AlmaLinux or Ubuntu server, a guided first start in the browser, and updates from the releases, with a rollback.',
				},
				{
					title: 'A second AI client and acceptance',
					text: 'A second AI client tested from end to end, a security review, and the success criteria checked on fresh servers.',
				},
			],
			milestones: 'The milestones on GitHub',
			todayTitle: 'What works today',
			todayLead: 'Batches 0 to 8 are done. In a Cahut, today:',
			today: [
				'One command that installs a Linux server, a first start in the browser, and daily updates from signed releases, with a rollback.',
				'One service on a Linux server, holding one or more Cahuts, each at its own domain with its certificate, or at the IPv4 address of the server until it has one.',
				'People who sign in, administrators of the Cahut and of the server, and team Rooms with owners and members.',
				'A web console: Knocks, Rooms, Windows, the Logbook, people, connected AIs and personal tokens.',
				'AI clients that connect over OAuth, such as claude.ai, Claude Code, ChatGPT and Le Chat, to the Rooms and the scope a person grants, or with a personal token.',
				'Windows published over MCP: documents, decks, dashboards and static sites, served in a sandbox that holds them to what their AI declares they load and call.',
				'Two versions of each Window with a rollback, a trash, and an access: Room only, internal, with a Key or public, until a date.',
				'Knock first: what an AI cannot undo, or what changes what a person may do, waits for its person, who approves or declines it in the console.',
				'A house style for each Cahut: its colors, fonts, logo and guidance, templates for documents, decks and dashboards, and a deck engine, which AIs read before they publish.',
				'A plugin for Claude Code, and a skill for the Claude apps, that teach Claude to make Windows in the house style.',
				'The Logbook of every call, every change and every sign-in.',
			],
			caution: 'No backups yet: keep snapshots of your server, and Cahut is not ready for work you cannot lose.',
			statusText: 'The documentation keeps the detail, area by area:',
			status: 'Where Cahut stands',
			afterTitle: 'After the first version',
			afterLead: 'Then, towards the full scope of the launch, in about this order:',
			after: [
				'Comments, search, single sign-on, backups and restore, and quotas per Cahut.',
				'Apps in unprivileged containers: runtimes, a Dockerfile mode, databases, secrets, logs and resource limits.',
				'Export to PDF and PowerPoint, SSH access limited to the same operations, and other AI clients.',
			],
		},
	},

	closing: {
		title: 'Share a link, not a file.',
		lead: 'Cahut is built in the open, one step at a time. Follow the work on GitHub, or read in the docs how it fits together.',
		github: 'Follow the work on GitHub',
		docs: 'Read the docs',
	},

	legal: {
		label: 'Legal',
		nav: { notice: 'Legal notice', privacy: 'Privacy' },
		updated: 'Last updated on',
	},
};
