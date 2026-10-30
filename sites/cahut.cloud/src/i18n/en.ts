// The words of cahut.cloud in English, the language of the launch. Every
// other language follows the shape of this file (see ./index.ts). Short HTML,
// such as <b>, is allowed where a component renders the text with set:html.

export default {
	meta: {
		title: 'Cahut: a home to publish and share what your AIs make',
		description:
			'Publish the documents, slide decks, dashboards and small apps your AIs make as pages at your address, share them by link and work on them as a team, on your own server, whichever AI made them. Cahut is in early development.',
		imageAlt:
			'Your home to publish and share what your AIs make. Documents, decks, dashboards and apps, on your own server.',
		locale: 'en_US',
	},

	base: {
		skip: 'Skip to content',
		home: 'Cahut, home',
		mainNav: 'Main',
		nav: {
			quickstart: 'Get started',
			how: 'How it works',
			examples: 'Examples',
			knock: 'Knock first',
			features: 'Features',
			docs: 'Docs',
			brand: 'Brand',
			project: 'Project',
			github: 'GitHub',
		},
		footer: { cahut: 'Cahut', project: 'Project', follow: 'Follow' },
		signature: 'Your home to publish and share what your AIs make.',
		about:
			'Cahut is a personal open source project by Olivier Durand, independent of any company and of any AI vendor. It is in early development.',
		// Around the link to Tetra7.
		hosted: ['My Cahut, at my.cahut.cloud, is offered by ', ' in support of the project.'],
		theme: { dark: 'Switch to the dark theme', light: 'Switch to the light theme', short: 'Theme' },
		language: { label: 'Version française', short: 'FR' },
		github: 'Cahut on GitHub',
		menu: 'Menu',
		// My Cahut, the Cahut at my.cahut.cloud that Tetra7 offers (src/links.ts).
		account: {
			signIn: 'Sign in',
			signInTitle: 'Open your Cahut on my.cahut.cloud',
			signup: 'Create my Cahut',
			free: 'Free to try, no credit card. Limited seats during the discovery phase.',
			member: 'Already a member? Open my.cahut.cloud',
		},
	},

	hero: {
		status: 'In early development: the first version works',
		// One line each: what you do, what your AI does, what you do next.
		title: ['You ask.', 'Your AI publishes.', 'You share.'],
		lead: 'Cahut is where your AIs publish what they make for you: a CV, a deck, a dashboard, a small game. Each page gets its own address, in your Cahut, and you decide who can open it.',
		install: 'Install on my server',
		facts: ['On your server or on my.cahut.cloud', 'Any AI, over MCP', 'A link, not a file'],
		// The scene beside the title: an AI publishes an example, then shares it.
		stage: {
			label: 'An AI publishes a page in a Cahut, then makes it public once a person approves',
			ai: 'Your AI',
			clients: 'Claude, Mistral Vibe, ChatGPT…',
			publishing: 'Publishing the Window in your Cahut…',
			online: "It's online, in the Room",
			access: 'Access',
			knock: 'Your AI asks to make this Window public.',
			approve: 'Approve',
			approved: 'Approved',
			example: 'Example',
			more: 'See the examples',
		},
	},

	statement: {
		label: 'The idea',
		title: 'Your AIs make it and publish it. You share it and work on it together.',
		lead: 'Everything they produce lives in one place, as pages at your address, on your server, whatever the AI.',
		text: 'Cahut does not create content, it hosts it. It does not compete with AIs or creation tools: it is the neutral, self-hosted home where their work lives, and where your team works on it.',
		layers: [
			{
				layer: 'Create',
				who: 'The AIs',
				what: 'Claude, ChatGPT, Mistral, Copilot and others make the content: documents, decks, dashboards, apps.',
				tags: ['Claude', 'ChatGPT', 'Mistral', 'Copilot'],
			},
			{
				layer: 'Publish, share, work together',
				who: 'Cahut, on your server',
				what: 'A page at your address, a link opened to the right people, Rooms where the team and its AIs work, versions, comments, search, export and the Logbook.',
				tags: ['Publish', 'Share', 'Rooms', 'Versions', 'Comments', 'Search', 'Export', 'Logbook'],
			},
			{
				layer: 'Read',
				who: 'The people you share with',
				what: 'They open a link in their browser, on a phone or a computer. No file to send, nothing to install.',
				tags: ['Your team', 'Your clients', 'Family and friends', 'The public'],
			},
		],
	},

	how: {
		label: 'How it works',
		title: 'Three steps, with the AI you already use',
		steps: [
			{
				title: 'You ask',
				text: 'You write to your AI, as usual: Claude or Mistral Vibe, formerly Le Chat, even on their free plans. ChatGPT publishes on its Business, Enterprise and Edu plans only, not on Free, Plus or Pro.',
			},
			{ title: 'Your AI publishes', text: 'It makes a Window: a living page at its own address, in one of your Rooms.' },
			{
				title: 'You share',
				text: 'A Window opens first to the members of its Room. Open it to your whole Cahut, to anyone with its Key, or to the public. For the public, the AI knocks first and you approve.',
			},
		],
		// The parts of the address under the steps, in its order.
		address: [
			{ word: 'Your Cahut', meaning: 'at your domain' },
			{ word: 'A Room', meaning: 'the space of a person or a team' },
			{ word: 'A Window', meaning: 'the page your AI published' },
		],
	},

	// The demos of demo.cahut.cloud, in the order of src/components/home/examples.ts.
	examples: {
		label: 'Examples',
		title: 'See for yourself: pages published by an AI',
		lead: 'Each one started from a single request. They are public: open them.',
		all: 'All the demos, with their prompts',
		items: [
			{
				title: 'An online CV',
				prompt: 'Can you create a Window in my Room on Cahut to publish my CV? Here is my background.',
			},
			{
				title: 'A team report',
				prompt: 'Write the monthly report of the Platform team from this data, and publish it for the executive committee.',
			},
			{
				title: 'A travel journal',
				prompt: 'Here are our photos and notes from Iceland. Make a travel journal with a map of our route, and publish it for the family.',
			},
			{
				title: 'A 3D configurator',
				prompt: 'Create a 3D configurator for our armchairs, with fabrics, price and a PDF quote. Publish it in the PME Room.',
			},
			{ title: 'An investor deck', prompt: 'Prepare the deck for our €2.5M fundraise from these figures.' },
			{
				title: 'A live dashboard',
				prompt: 'Make a dashboard of our 14 sites around the world, for the screen at head office.',
			},
			{ title: 'A made-to-measure game', prompt: 'Noé turns 9 on Saturday. Invent a little game for him, with a surprise at the end.' },
			{ title: 'A mini poll', prompt: 'Create a short poll to choose where the 2027 offsite goes.' },
		],
	},

	knock: {
		label: 'Governance',
		title: 'Knock first',
		lead: 'A home has a door. The AI is the intended operator of Cahut, but never without guardrails: what can be undone, it does alone. What cannot waits for a human.',
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
		lead: 'The full scope ships at launch, with no reduced trial version. Today, the first version works: it installs on a server, connects Claude and ChatGPT, publishes and shares documents, decks, dashboards and sites, finds them by tag or by word, and gathers a team in Rooms. Apps, comments, export, single sign-on and backups are the target.',
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
					{ area: 'Search', text: 'Find any page, deck or dashboard of your Rooms by a word of it or a tag.' },
					{ area: 'Export', text: 'PDF and PowerPoint, for recipients still on Office.' },
				],
			},
			{
				title: 'Connect',
				text: 'Every AI, one door.',
				features: [
					{
						area: 'AI access',
						text: 'A native MCP server (HTTP, OAuth) and one skill per AI client, packaged as a plugin where the client allows it: Claude first, then ChatGPT, Mistral, Copilot and others. SSH access limited to the same operations, for terminal AIs.',
					},
				],
			},
			{
				title: 'Share and work together',
				text: 'A link opened to the right people, and Rooms for the team.',
				features: [
					{
						area: 'Sharing',
						text: 'By link, Window by Window: its Room only, internal (everyone in the Cahut), a Key (a password) or public, with an optional end date.',
					},
					{
						area: 'Invitations',
						text: 'A Window sent by e-mail, to people of the Cahut or outside it, with a personal link in place of its Key.',
					},
					{ area: 'Comments', text: 'Reader feedback directly on the pages.' },
					{
						area: 'Users and Rooms',
						text: 'The people of each Cahut and its admins, a personal Room for each person, and team Rooms with owners and members, where people and their AIs work on the same Windows.',
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
						text: 'See what is published, share, comment, approve irreversible actions, administer the server and its Cahuts, under /console/ of each Cahut.',
					},
					{
						area: 'Domains and TLS',
						text: 'Several Cahuts per server, each at its own domain or subdomain. No wildcard needed, automatic certificates.',
					},
					{
						area: 'Limits',
						text: 'People, Rooms, Windows and disk for each Cahut, set by the administrators of the server, and Cahuts where each person works alone.',
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
			{ word: 'Knock first', meaning: 'The governance rule: the AI knocks before any irreversible action.' },
			{ word: 'Logbook', meaning: 'The record of every action.' },
		],
		accessLabels: { room: 'Room only', internal: 'Internal', key: 'Key', public: 'Public' },
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
			roadmap: 'The batches of Cahut, what works today, and what makes it ready for launch.',
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
			batchesTitle: 'The batches',
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
					title: 'A skill for Claude',
					text: 'A skill for the Claude apps, and a plugin for Claude Code, that tell Claude how to make a Window and publish it on Cahut.',
				},
				{
					title: 'Installation, first start and updates',
					text: 'One command on a fresh AlmaLinux or Ubuntu server, a guided first start in the browser, and updates from the releases, with a rollback.',
				},
				{
					title: 'A second AI client and acceptance',
					text: 'A second AI client tested from end to end, a security review, and the success criteria checked on a fresh server.',
				},
				{
					title: 'E-mail from each Cahut',
					text: 'Each Cahut sends e-mail through its own SMTP relay: links to choose a password, invitations to read a Window, and notices of Knocks.',
				},
			],
			milestones: 'The milestones on GitHub',
			todayTitle: 'What works today',
			todayLead: 'Batches 0 to 9 are done, and the first version met the success criteria on a fresh server. In a Cahut, today:',
			today: [
				'One command that installs a Linux server, a first start in the browser, and daily updates from signed releases, with a rollback.',
				'One service on a Linux server, holding one or more Cahuts, each at its own domain with its certificate, or at the IPv4 address of the server until it has one.',
				'Collaborative Cahuts, or individual ones where each person works alone, and limits on people, Rooms, Windows and disk, set by the administrators of the server.',
				'People who sign in, administrators of the Cahut and of the server, and team Rooms with owners and members. The page of each Room lists its Windows, or shows the one its owners choose.',
				'A web console, in English and French: Knocks, Rooms, Windows, the Logbook, people, connected AIs, personal tokens, e-mail and the server.',
				'AI clients that connect over OAuth, such as claude.ai, Claude Code, Mistral Vibe and ChatGPT, to the Rooms and the scope a person grants, or with a personal token.',
				'Windows published over MCP: documents, decks, dashboards and static sites, served in a sandbox that holds them to what their AI declares they load and call.',
				'Two versions of each Window with a rollback, a trash, and an access: Room only, internal, with a Key or public, until a date.',
				'Tags on Windows, and a search of their titles, descriptions, tags and pages, in the console and for AIs.',
				'Knock first: what an AI cannot undo, or what changes what a person may do, waits for its person, who approves or declines it in the console.',
				'E-mail through the SMTP relay of each Cahut: links to choose a password, invitations to read a Window, and notices of Knocks, of the account and of limits.',
				'A plugin for Claude Code, and a skill for the Claude apps, that teach Claude to make Windows and publish them.',
				'The Logbook of every call, every change and every sign-in.',
			],
			caution: 'No backups yet: keep snapshots of your server, and Cahut is not ready for work you cannot lose.',
			statusText: 'The documentation keeps the detail, area by area:',
			status: 'Where Cahut stands',
			afterTitle: 'After the first version',
			afterLead: 'Then, towards the full scope of the launch, in about this order:',
			after: [
				'Comments, single sign-on, and backups and restore.',
				'Apps in unprivileged containers: runtimes, a Dockerfile mode, databases, secrets, logs and resource limits.',
				'Export to PDF and PowerPoint, SSH access limited to the same operations, and other AI clients.',
			],
		},
	},

	closing: {
		title: 'Your first Window in a few minutes.',
		lead: 'Cahut is free and open source. Try it without a server, or install it on yours.',
		hosted: {
			title: 'Try it without a server',
			text: 'Create your Cahut on My Cahut and connect your AI in two minutes: Claude or Mistral Vibe, even on their free plans, or ChatGPT on a Business, Enterprise or Edu plan.',
			// Around the link to Tetra7.
			fine: ['Free, limited seats. My Cahut is offered by ', ', which hosts it in France to support the project.'],
			more: 'All about My Cahut',
		},
		server: {
			title: 'Install it on your server',
			text: 'One command on an Ubuntu or AlmaLinux server.',
			copy: 'Copy',
			copied: 'Copied',
			fine: 'Every feature, free and unlimited, under the AGPL-3.0.',
			guide: 'The install guide',
		},
		docs: 'Read the docs',
		github: 'Follow the work on GitHub',
	},

	legal: {
		label: 'Legal',
		nav: { notice: 'Legal notice', privacy: 'Privacy' },
		updated: 'Last updated on',
	},
};
