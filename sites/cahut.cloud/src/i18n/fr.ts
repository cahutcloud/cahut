// The words of cahut.cloud in French. The words of the house (Cahut, Room,
// Window, Key, Knock first, Logbook) stay in English, as in the brand
// guidelines. Spaces before French punctuation come from ./index.ts.
import type { Dictionary } from './index';

const fr: Dictionary = {
	meta: {
		title: 'Cahut : une maison pour publier et partager ce que fabriquent vos IA',
		description:
			"Publiez les documents, présentations, tableaux de bord et petites applis que fabriquent vos IA en pages à votre adresse, partagez-les par un lien et travaillez dessus en équipe, sur votre propre serveur, quelle que soit l'IA qui les a faits. Cahut est en début de développement.",
		imageAlt:
			'Votre maison pour publier et partager ce que fabriquent vos IA. Documents, présentations, tableaux de bord et applis, sur votre propre serveur.',
		locale: 'fr_FR',
	},

	base: {
		skip: 'Aller au contenu',
		home: 'Cahut, accueil',
		mainNav: 'Principale',
		nav: {
			quickstart: 'Démarrer',
			how: 'Comment ça marche',
			examples: 'Exemples',
			knock: 'Knock first',
			features: 'Fonctionnalités',
			docs: 'Docs',
			brand: 'Marque',
			project: 'Projet',
			github: 'GitHub',
		},
		footer: { cahut: 'Cahut', project: 'Projet', follow: 'Suivre' },
		signature: 'Votre maison pour publier et partager ce que fabriquent vos IA.',
		about:
			"Cahut est un projet open source personnel d'Olivier Durand, indépendant de toute entreprise et de tout fournisseur d'IA. Elle est en début de développement.",
		hosted: ['My Cahut, à my.cahut.cloud, est offerte par ', ' en soutien au projet.'],
		theme: { dark: 'Passer au thème sombre', light: 'Passer au thème clair', short: 'Thème' },
		language: { label: 'English version', short: 'EN' },
		github: 'Cahut sur GitHub',
		menu: 'Menu',
		account: {
			signIn: 'Se connecter',
			signInTitle: 'Ouvrir votre Cahut sur my.cahut.cloud',
			signup: 'Créer ma Cahut',
			free: 'Gratuit pour essayer, sans carte bancaire. Places limitées pendant la découverte.',
			member: 'Déjà membre ? Ouvrir my.cahut.cloud',
		},
	},

	hero: {
		status: 'En début de développement : la première version fonctionne',
		title: ['Vous demandez.', 'Votre IA publie.', 'Vous partagez.'],
		lead: "Cahut est l'endroit où vos IA publient ce qu'elles font pour vous : un CV, un deck, un tableau de bord, un petit jeu. Chaque page arrive à sa propre adresse, dans votre Cahut, et c'est vous qui décidez qui peut l'ouvrir.",
		install: 'Installer sur mon serveur',
		facts: ['Sur votre serveur ou sur my.cahut.cloud', 'Toutes les IA, par MCP', 'Un lien, pas un fichier'],
		stage: {
			label: "Une IA publie une page dans une Cahut, puis la rend publique quand une personne l'approuve",
			ai: 'Votre IA',
			clients: 'Claude, Mistral Vibe, ChatGPT…',
			publishing: 'Je publie la Window dans votre Cahut…',
			online: "C'est en ligne, dans la Room",
			access: 'Accès',
			knock: 'Votre IA demande à rendre cette Window publique.',
			approve: 'Approuver',
			approved: 'Approuvé',
			example: 'Exemple',
			more: 'Voir les exemples',
		},
	},

	statement: {
		label: "L'idée",
		title: 'Vos IA le fabriquent et le publient. Vous le partagez et y travaillez ensemble.',
		lead: "Tout ce qu'elles produisent vit au même endroit, en pages à votre adresse, sur votre serveur, quelle que soit l'IA.",
		text: "Cahut ne crée pas de contenu, elle l'héberge. Elle ne concurrence ni les IA ni les outils de création : c'est la maison neutre et auto-hébergée où leur travail vit, et où votre équipe travaille dessus.",
		layers: [
			{
				layer: 'Créer',
				who: 'Les IA',
				what: "Claude, ChatGPT, Mistral, Copilot et d'autres fabriquent le contenu : documents, présentations, tableaux de bord, applis.",
				tags: ['Claude', 'ChatGPT', 'Mistral', 'Copilot'],
			},
			{
				layer: 'Publier, partager, travailler ensemble',
				who: 'Cahut, sur votre serveur',
				what: "Une page à votre adresse, un lien ouvert aux bonnes personnes, des Rooms où l'équipe et ses IA travaillent, les versions, les commentaires, la recherche, l'export et le Logbook.",
				tags: ['Publier', 'Partager', 'Rooms', 'Versions', 'Commentaires', 'Recherche', 'Export', 'Logbook'],
			},
			{
				layer: 'Lire',
				who: 'Les personnes avec qui vous partagez',
				what: 'Elles ouvrent un lien dans leur navigateur, sur un téléphone ou un ordinateur. Aucun fichier à envoyer, rien à installer.',
				tags: ['Votre équipe', 'Vos clients', 'Famille et amis', 'Le public'],
			},
		],
	},

	how: {
		label: 'Comment ça marche',
		title: "Trois temps, avec l'IA que vous utilisez déjà",
		steps: [
			{
				title: 'Vous demandez',
				text: "Vous écrivez à votre IA, comme d'habitude : Claude ou Mistral Vibe, l'ancien Le Chat, même avec leur offre gratuite. ChatGPT ne publie qu'avec ses offres Business, Enterprise et Edu, pas avec Free, Plus ou Pro.",
			},
			{
				title: 'Votre IA publie',
				text: 'Elle en fait une Window : une page vivante, à sa propre adresse, rangée dans une de vos Rooms.',
			},
			{
				title: 'Vous partagez',
				text: "Une Window s'ouvre d'abord aux membres de sa Room. Vous l'ouvrez à toute la Cahut, à qui a sa Key, ou au public. Pour le public, l'IA frappe d'abord et vous approuvez.",
			},
		],
		address: [
			{ word: 'Votre Cahut', meaning: 'à votre domaine' },
			{ word: 'Une Room', meaning: "l'espace d'une personne ou d'une équipe" },
			{ word: 'Une Window', meaning: 'la page publiée par votre IA' },
		],
	},

	examples: {
		label: 'Exemples',
		title: 'Voyez par vous-même : des pages publiées par une IA',
		lead: "Chacune part d'une seule demande. Elles sont publiques : ouvrez-les.",
		all: 'Toutes les démos, avec leurs prompts',
		items: [
			{
				title: 'Un CV en ligne',
				prompt: 'Peux-tu créer une Window dans ma Room sur Cahut pour publier mon CV ? Voici mon parcours.',
			},
			{
				title: "Un reporting d'équipe",
				prompt: "Fais-moi le reporting mensuel de l'équipe Plateforme à partir de ces données, et publie-le pour le comité de direction.",
			},
			{
				title: 'Un carnet de voyage',
				prompt: 'Voici nos photos et nos notes de voyage en Islande. Fais-en un carnet avec la carte de notre route, et publie-le pour la famille.',
			},
			{
				title: 'Un configurateur 3D',
				prompt: 'Crée un configurateur 3D de nos fauteuils, avec les tissus, le prix et un devis en PDF. Publie-le dans la Room PME.',
			},
			{
				title: "Un deck d'investisseurs",
				prompt: 'Prépare le deck de notre levée de fonds de 2,5 M€ à partir de ces chiffres.',
			},
			{
				title: 'Un tableau de bord en direct',
				prompt: "Fais un tableau de bord de nos 14 sites dans le monde, pour l'écran du siège.",
			},
			{
				title: 'Un jeu fait sur mesure',
				prompt: 'Noé a 9 ans samedi. Invente-lui un petit jeu, avec une surprise à la fin.',
			},
			{ title: 'Un mini sondage', prompt: 'Crée un petit sondage pour choisir la destination du séminaire 2027.' },
		],
	},

	knock: {
		label: 'Gouvernance',
		title: 'Knock first',
		lead: "Une maison a une porte. L'IA est l'opératrice prévue de Cahut, mais jamais sans garde-fous : ce qui s'annule, elle le fait seule. Ce qui ne s'annule pas attend un humain.",
		reversible: 'Réversible',
		alone: "L'IA le fait seule",
		aloneItems: [
			'Publier ou mettre à jour une Window (la version précédente est gardée)',
			'Restreindre un accès, prolonger une durée de vie',
			'Déployer une appli dans les limites fixées',
			'Revenir à la version précédente',
			'Lancer une sauvegarde',
		],
		irreversible: 'Irréversible',
		approves: 'Un humain approuve',
		approvedItems: [
			'Supprimer définitivement une Window, une appli ou une base de données',
			'Rendre une Window publique',
			'Ajouter une personne à une Cahut',
			"Changer les droits d'une personne ou d'une IA",
			'Effacer des sauvegardes ou changer leur destination',
		],
		try: 'Essayez : répondez à la demande.',
		principlesTitle: "Sept principes tiennent l'ensemble.",
		principles: [
			{
				title: 'Des opérations bornées',
				text: "L'IA agit par des opérations typées (publier une page, créer une base de données, partager un lien), jamais par un shell libre, par MCP comme par SSH.",
			},
			{
				title: 'Réversible ou approuvé',
				text: "Ce qui s'annule, l'IA le fait seule. Ce qui ne s'annule pas attend qu'un humain l'approuve dans la console, avec une notification et une échéance.",
			},
			{
				title: "L'IA déclare ce qu'elle publie",
				text: "Chaque publication dit ce qu'elle contient et ce qu'elle fait : une appli qui collecte des données, appelle l'extérieur, ouvre un formulaire.",
			},
			{
				title: 'Des droits par agent',
				text: "Chaque IA agit au nom d'une personne, dans le périmètre que cette personne lui accorde, révocable à tout moment.",
			},
			{
				title: 'Tout est consigné',
				text: 'Quelle IA, au nom de qui, quelle action, quel résultat : le Logbook se lit dans la console.',
			},
			{
				title: 'Les données restent chez vous',
				text: 'Tout vit sur votre serveur. Les sauvegardes ne partent que vers le stockage externe que vous choisissez.',
			},
			{
				title: 'Isolation',
				text: "Chaque Window tourne dans un bac à sable : ses scripts ne peuvent lire ni les autres Windows, ni la console, ni la session du lecteur. Chaque appli tourne dans un conteneur sans privilèges, avec ses propres limites de ressources. Les Cahuts d'un même serveur ne partagent que la machine.",
			},
		],
	},

	demo: {
		console: 'la console de votre Cahut',
		waitingCount: '1 en attente',
		nothingWaits: "Rien n'attend",
		waiting: 'En attente',
		request: "Veut rendre <b>Q3 review</b> publique. Aujourd'hui, elle s'ouvre à qui a sa Key.",
		meta: "expire dans 23 h · au nom d'Olivier",
		approve: 'Approuver',
		decline: 'Refuser',
		logbook: "Logbook · aujourd'hui",
		log: [
			'a demandé à rendre <b>Q3 review</b> publique',
			'a donné une Key à <b>Q3 review</b>',
			'a publié <b>Q3 review</b>',
		],
		done: 'Fait',
		again: 'Redemander',
		answers: {
			approve: {
				badge: 'Approuvé',
				result: 'Approuvé. Q3 review est publique, et Claude le sait.',
				log: 'a accepté de rendre <b>Q3 review</b> publique',
				done: 'Fait',
			},
			decline: {
				badge: 'Refusé',
				result: 'Refusé. Q3 review garde sa Key, et Claude le sait.',
				log: 'a refusé de rendre <b>Q3 review</b> publique',
				done: 'Refusé',
			},
		},
	},

	features: {
		label: 'Fonctionnalités',
		title: 'Ce qui sort au lancement',
		lead: "Le périmètre complet sort au lancement, sans version d'essai réduite. Aujourd'hui, la première version fonctionne : elle s'installe sur un serveur, connecte Claude et ChatGPT, publie et partage documents, présentations, tableaux de bord et sites, les retrouve par tag ou par mot, et réunit une équipe dans des Rooms. Les applis, les commentaires, l'export, l'authentification unique et les sauvegardes sont la cible.",
		groups: [
			{
				title: 'Publier',
				text: 'Ce que fabriquent les IA, en ligne à sa propre adresse.',
				features: [
					{
						area: 'Pages',
						text: 'Documents, présentations, tableaux de bord et sites statiques, chacun à sa propre adresse dans sa Room.',
					},
					{
						area: 'Applis',
						text: "Des environnements prédéfinis, comme Node et Python, et un mode Dockerfile avancé. Variables et secrets, bases PostgreSQL ou SQLite, journaux d'exécution, limites de ressources.",
					},
					{ area: 'Versions', text: 'Deux versions gardées par publication. Retour arrière en un geste.' },
					{
						area: 'Recherche',
						text: "Retrouvez n'importe quelle page, présentation ou tableau de bord de vos Rooms par l'un de ses mots ou un tag.",
					},
					{ area: 'Export', text: 'PDF et PowerPoint, pour les destinataires encore sur Office.' },
				],
			},
			{
				title: 'Connecter',
				text: 'Toutes les IA, une seule porte.',
				features: [
					{
						area: 'Accès des IA',
						text: "Un serveur MCP natif (HTTP, OAuth) et une skill par client d'IA, en plugin quand le client le permet : Claude d'abord, puis ChatGPT, Mistral, Copilot et d'autres. Un accès SSH limité aux mêmes opérations, pour les IA en terminal.",
					},
				],
			},
			{
				title: 'Partager et travailler ensemble',
				text: "Un lien ouvert aux bonnes personnes, et des Rooms pour l'équipe.",
				features: [
					{
						area: 'Partage',
						text: 'Par lien, Window par Window : sa Room seulement, interne (toute la Cahut), une Key (un mot de passe) ou public, avec une date de fin facultative.',
					},
					{
						area: 'Invitations',
						text: 'Une Window envoyée par e-mail, aux personnes de la Cahut ou au-dehors, avec un lien personnel qui tient lieu de Key.',
					},
					{ area: 'Commentaires', text: 'Les retours des lecteurs, directement sur les pages.' },
					{
						area: 'Personnes et Rooms',
						text: "Les personnes de chaque Cahut et ses administrateurs, une Room personnelle pour chacun, et des Rooms d'équipe avec propriétaires et membres, où les personnes et leurs IA travaillent sur les mêmes Windows.",
					},
					{
						area: 'Connexion',
						text: "Des comptes propres à chaque Cahut, et l'authentification unique avec OpenID Connect : Google, Microsoft Entra, Okta et d'autres.",
					},
				],
			},
			{
				title: 'Exploiter',
				text: 'Un serveur à vous, tenu en ordre.',
				features: [
					{
						area: 'Installation',
						text: "Une commande sur un VPS vierge ou un serveur dédié (AlmaLinux ou Ubuntu), un premier démarrage guidé, et l'association de l'administrateur et de sa première IA.",
					},
					{
						area: 'Console web',
						text: 'Voir ce qui est publié, partager, commenter, approuver les actions irréversibles, administrer le serveur et ses Cahuts, sous /console/ de chaque Cahut.',
					},
					{
						area: 'Domaines et TLS',
						text: 'Plusieurs Cahuts par serveur, chacune à son domaine ou sous-domaine. Pas de wildcard, des certificats automatiques.',
					},
					{
						area: 'Limites',
						text: 'Des personnes, des Rooms, des Windows et du disque pour chaque Cahut, fixés par les administrateurs du serveur, et des Cahuts où chacun travaille seul.',
					},
					{
						area: 'Logbook',
						text: "L'historique de chaque action : qui, quelle IA, au nom de qui, avec quel résultat.",
					},
					{
						area: 'Sauvegardes',
						text: 'Facultatives, vers un stockage externe S3 ou SFTP. Restauration depuis la console.',
					},
					{
						area: 'Mises à jour',
						text: 'Automatiques, vérifiées auprès des versions publiées sur GitHub, avec retour arrière.',
					},
				],
			},
		],
		notDoneTitle: 'Ce que Cahut ne fait pas',
		notDone: [
			"Créer du contenu : c'est le travail des IA.",
			"Faire tourner un modèle d'IA sur le serveur : les IA restent dehors et s'y connectent.",
			"Dépendre d'un fournisseur d'IA, ou enfermer le travail dans un format propriétaire.",
			'Donner à une IA un accès libre au serveur.',
			'Tourner sur plusieurs machines : un serveur porte une ou plusieurs Cahuts.',
		],
	},

	vocabulary: {
		label: 'Vocabulaire',
		title: 'La langue de la maison',
		lead: 'Cahut parle la langue de la maison, en anglais simple. Commencez par une adresse : elle se lit comme une maison.',
		parts: [
			{
				word: 'Cahut',
				meaning:
					'Votre instance, à son propre domaine. Un serveur peut en porter plusieurs, chacune avec ses personnes et sa base de données.',
			},
			{
				word: 'Room',
				meaning: "Un espace personnel ou d'équipe, avec ses membres. Chaque personne a sa propre Room.",
			},
			{
				word: 'Window',
				meaning: 'Une page vivante : document, présentation, tableau de bord ou appli. Elle garde sa version précédente.',
			},
		],
		words: [
			{ word: 'Key', meaning: 'Le mot de passe qui ouvre une Window protégée.' },
			{ word: 'Knock first', meaning: "La règle de gouvernance : l'IA frappe avant toute action irréversible." },
			{ word: 'Logbook', meaning: 'Le registre de toutes les actions.' },
		],
		accessLabels: { room: 'Room seulement', internal: 'Interne', key: 'Key', public: 'Public' },
		accessTitle: 'Accès',
		accessText: "Chaque Window s'ouvre à l'un de quatre cercles. Vous choisissez, Window par Window.",
		access: [
			'Les membres de sa Room. Par défaut.',
			'Toutes les personnes connectées à votre Cahut.',
			'Quiconque a sa Key.',
			"Quiconque a le lien. L'IA frappe d'abord.",
		],
	},

	roadmap: {
		label: 'Feuille de route',
		title: 'Où en est Cahut',
		lead: "Cahut est un projet personnel d'Olivier Durand, indépendant de toute entreprise et de tout fournisseur d'IA. Elle est écrite en Go, en public, une étape après l'autre.",
		steps: [
			"Protéger le nom : organisation GitHub, noms des paquets, recherche d'antériorité de marque.",
			"Écrire l'architecture cible : comparer les options et trancher, avec leurs raisons.",
			'Construire la première tranche verticale en Go : installer sur un serveur, publier une page par MCP, la servir en HTTPS.',
			'Preuve de concept : Cahut installée sur un VPS vierge, une IA qui publie une présentation par MCP, la console.',
			"Feuille de route jusqu'au lancement public.",
		],
		underWay: 'En cours',
		more: 'La feuille de route complète',
		criteriaTitle: 'Prêt pour le lancement quand',
		criterion: 'Critère',
		threshold: 'Seuil',
		criteria: [
			{ criterion: 'Installation sur un VPS vierge', threshold: 'Moins de 10 minutes, une commande' },
			{ criterion: 'Première publication par une IA', threshold: 'Moins de 5 minutes, sans toucher au serveur' },
			{
				criterion: 'Partage par une personne non technique',
				threshold: 'Une présentation partagée par lien en quelques minutes',
			},
			{
				criterion: 'IA prises en charge',
				threshold: "La même Cahut fonctionne avec au moins deux clients d'IA différents",
			},
			{
				criterion: 'Gouvernance',
				threshold: "Chaque action d'une IA figure au Logbook, et s'annule ou a été approuvée",
			},
		],
	},

	project: {
		label: 'Projet',
		nav: {
			governance: 'Gouvernance',
			roadmap: 'Feuille de route',
			contribute: 'Contribuer',
			support: "Obtenir de l'aide",
			security: 'Sécurité',
			conduct: 'Code de conduite',
		},
		descriptions: {
			governance: "Qui décide dans le projet Cahut, ses principes, ses rôles, comment il grandit, et ce que fait l'argent.",
			roadmap: "Les lots de Cahut, ce qui marche aujourd'hui, et ce qui la rend prête pour le lancement.",
			contribute:
				"Ce qui aide Cahut tout de suite, ce qui se discute d'abord, comment une pull request entre, et les règles des contributions faites avec une IA.",
			support: "Où demander de l'aide sur Cahut, comment bien demander, et à quoi s'attendre.",
			security: "Comment signaler une faille de Cahut, en privé, à quoi s'attendre, et ce qui est dans le périmètre.",
			conduct: 'Le code de conduite de la communauté Cahut, le Contributor Covenant 3.0, et où signaler un incident.',
		},
		source: {
			file: 'Cette page est {file}, à la racine du dépôt, rendu à chaque construction du site.',
			translation: 'Cette page traduit {file}, à la racine du dépôt, qui fait foi, en anglais.',
			conduct:
				'Cette page reprend la traduction officielle du Contributor Covenant 3.0, avec le canal de signalement de Cahut. {file}, à la racine du dépôt, fait foi, en anglais.',
			change: 'Pour la changer, ouvrez une pull request.',
		},
		roadmap: {
			title: 'Feuille de route',
			lead: "Cahut se construit par lots, chacun une tranche qui fonctionne de l'IA au lecteur. Leur ordre est fixé, leurs dates ne le sont pas : les jalons du dépôt disent où en est chaque lot.",
			batchesTitle: 'Les lots',
			batch: 'Lot',
			states: { done: 'Fait', next: 'Prochain' },
			batches: [
				{
					title: 'Socle et décisions',
					text: "Les premiers ADR, le squelette Go, l'intégration continue et la construction des releases.",
				},
				{
					title: 'Publier par MCP, servir en HTTPS',
					text: "Une IA publie une Window avec un jeton personnel. Cahut la sert à sa propre adresse, dans un bac à sable, et l'appel va au Logbook.",
				},
				{
					title: 'Personnes, Rooms et console',
					text: "Des personnes qui se connectent ; des administrateurs de chaque Cahut et du serveur ; des Rooms d'équipe avec leurs droits ; la console web et son API.",
				},
				{
					title: 'Brancher les IA : OAuth',
					text: "Chaque Cahut est le serveur d'autorisation de son point d'accès MCP. Sur un écran de consentement, une personne accorde à une IA des Rooms et une portée, et voit les IA connectées dans la console.",
				},
				{
					title: 'Windows complètes : versions et accès',
					text: "Mettre à jour une Window garde la version précédente, et un geste y revient. Chaque Window s'ouvre à sa Room seulement, à toute sa Cahut, à qui a sa Key, ou au public.",
				},
				{
					title: 'Knock first',
					text: "L'IA demande avant tout ce qu'elle ne peut pas annuler. La demande attend dans la console, avec une notification et une échéance, jusqu'à ce qu'une personne l'approuve ou la refuse.",
				},
				{
					title: 'Une skill pour Claude',
					text: 'Une skill pour les applications Claude, et un plugin pour Claude Code, qui disent à Claude comment faire une Window et la publier sur Cahut.',
				},
				{
					title: 'Installation, premier démarrage et mises à jour',
					text: 'Une commande sur un serveur AlmaLinux ou Ubuntu vierge, un premier démarrage guidé dans le navigateur, et des mises à jour depuis les releases, avec retour arrière.',
				},
				{
					title: 'Deuxième client IA et recette',
					text: "Un deuxième client d'IA testé de bout en bout, une revue de sécurité, et les critères de réussite vérifiés sur un serveur vierge.",
				},
				{
					title: "L'envoi d'e-mails",
					text: 'Chaque Cahut envoie ses e-mails par son propre relais SMTP : des liens pour choisir un mot de passe, des invitations à lire une Window, et les notifications des Knocks.',
				},
			],
			milestones: 'Les jalons sur GitHub',
			todayTitle: "Ce qui marche aujourd'hui",
			todayLead: "Les lots 0 à 9 sont faits, et la première version a rempli les critères de réussite sur un serveur vierge. Aujourd'hui, dans une Cahut :",
			today: [
				"Une commande qui installe un serveur Linux, un premier démarrage dans le navigateur, et des mises à jour quotidiennes depuis les releases signées, avec retour arrière.",
				"Un service sur un serveur Linux, qui porte une ou plusieurs Cahuts, chacune à son domaine avec son certificat, ou à l'adresse IPv4 du serveur tant qu'elle n'en a pas.",
				'Des Cahuts collaboratives, ou individuelles où chacun travaille seul, et des limites sur les personnes, les Rooms, les Windows et le disque, que fixent les administrateurs du serveur.',
				"Des personnes qui se connectent, des administrateurs de la Cahut et du serveur, et des Rooms d'équipe avec propriétaires et membres. La page de chaque Room liste ses Windows, ou montre celle que ses propriétaires choisissent.",
				'Une console web, en anglais et en français : Knocks, Rooms, Windows, Logbook, personnes, IA connectées, jetons personnels, e-mail et serveur.',
				"Des clients d'IA qui se connectent par OAuth, comme claude.ai, Claude Code, Mistral Vibe et ChatGPT, aux Rooms et dans la portée qu'une personne accorde, ou avec un jeton personnel.",
				"Des Windows publiées par MCP : documents, présentations, tableaux de bord et sites statiques, servis dans un bac à sable qui les tient à ce que leur IA déclare charger et appeler.",
				"Deux versions de chaque Window avec retour en arrière, une corbeille, et un accès : Room seulement, interne, avec une Key ou public, jusqu'à une date.",
				'Des tags sur les Windows, et une recherche dans leurs titres, descriptions, tags et pages, dans la console et pour les IA.',
				"Knock first : ce qu'une IA ne peut pas défaire, ou ce qui change ce qu'une personne peut faire, attend sa personne, qui l'approuve ou le refuse dans la console.",
				'Les e-mails, par le relais SMTP de chaque Cahut : des liens pour choisir un mot de passe, des invitations à lire une Window, et les notifications des Knocks, du compte et des limites.',
				'Un plugin pour Claude Code, et une skill pour les applications Claude, qui apprennent à Claude à faire des Windows et à les publier.',
				'Le Logbook de chaque appel, de chaque changement et de chaque connexion.',
			],
			caution:
				"Pas encore de sauvegardes : gardez des instantanés de votre serveur, et ne confiez pas encore à Cahut un travail que vous ne pouvez pas perdre.",
			statusText: 'La documentation garde le détail, domaine par domaine :',
			status: 'Où en est Cahut',
			afterTitle: 'Après la première version',
			afterLead: 'Puis, vers le périmètre complet du lancement, à peu près dans cet ordre :',
			after: [
				"Les commentaires, l'authentification unique, et les sauvegardes et leur restauration.",
				'Les applis dans des conteneurs sans privilèges : environnements, mode Dockerfile, bases de données, secrets, journaux et limites de ressources.',
				"L'export en PDF et en PowerPoint, l'accès SSH limité aux mêmes opérations, et d'autres clients d'IA.",
			],
		},
	},

	closing: {
		title: 'Votre première Window en quelques minutes.',
		lead: 'Cahut est libre et open source. Essayez-la sans serveur, ou installez-la sur le vôtre.',
		hosted: {
			title: 'Essayer sans serveur',
			text: 'Créez votre Cahut sur My Cahut et connectez votre IA en deux minutes : Claude ou Mistral Vibe, même avec leur offre gratuite, ou ChatGPT avec une offre Business, Enterprise ou Edu.',
			fine: ['Gratuit, places limitées. My Cahut est offerte par ', ", qui l'héberge en France pour soutenir le projet."],
			more: 'Tout sur My Cahut',
		},
		server: {
			title: 'Installer sur votre serveur',
			text: 'Une commande sur un serveur Ubuntu ou AlmaLinux.',
			copy: 'Copier',
			copied: 'Copié',
			fine: 'Toutes les fonctions, sans limite, sous AGPL-3.0.',
			guide: "Le guide d'installation",
		},
		docs: 'Lire la documentation',
		github: 'Suivre le projet sur GitHub',
	},

	legal: {
		label: 'Informations légales',
		nav: { notice: 'Mentions légales', privacy: 'Confidentialité' },
		updated: 'Dernière mise à jour le',
	},
};

export default fr;
