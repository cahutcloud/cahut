// The words of cahut.cloud in French. The words of the house (Cahut, Room,
// Window, Key, House style, Knock first, Logbook) stay in English, as in the
// brand guidelines. Spaces before French punctuation come from ./index.ts.
import type { Dictionary } from './index';

const fr: Dictionary = {
	meta: {
		title: 'Cahut : la maison de tout ce que fabriquent vos IA',
		description:
			"Documents, présentations, tableaux de bord et petites applis deviennent des pages vivantes en ligne, à vos couleurs, sur votre propre serveur, quelle que soit l'IA qui les a faits. Cahut est en début de développement.",
		imageAlt:
			'Votre propre nuage pour tout ce que fabriquent vos IA. Documents, présentations, tableaux de bord et applis, à vos couleurs, sur votre propre serveur.',
		locale: 'fr_FR',
	},

	base: {
		skip: 'Aller au contenu',
		home: 'Cahut, accueil',
		mainNav: 'Principale',
		nav: {
			quickstart: 'Démarrer',
			how: 'Comment ça marche',
			knock: 'Knock first',
			features: 'Fonctionnalités',
			docs: 'Docs',
			brand: 'Marque',
			project: 'Projet',
			github: 'GitHub',
		},
		footer: { cahut: 'Cahut', project: 'Projet' },
		signature: 'Votre propre nuage pour tout ce que fabriquent vos IA.',
		about:
			"Cahut est un projet open source personnel d'Olivier Durand, indépendant de toute entreprise et de tout fournisseur d'IA. Il est en début de développement.",
		theme: { dark: 'Passer au thème sombre', light: 'Passer au thème clair' },
		language: { label: 'English version', short: 'EN' },
	},

	hero: {
		status: 'En début de développement : la première tranche fonctionne',
		title: 'La maison de tout ce que fabriquent vos IA.',
		lead: "Documents, présentations, tableaux de bord et petites applis deviennent des pages vivantes en ligne, à vos couleurs, sur votre propre serveur, quelle que soit l'IA qui les a faits.",
		start: 'Démarrer',
		github: 'Suivre le projet sur GitHub',
		facts: ['Une commande sur votre serveur', 'Toutes les IA, par MCP', 'Un lien, pas un fichier'],
		caption:
			"Un document, un tableau de bord, un site et une image rejoignent un nuage à la porte ouverte, éclairée de l'intérieur.",
		log: { action: 'a publié', done: 'Fait' },
	},

	quickstart: {
		label: 'Démarrage rapide',
		title: 'Une commande, puis demandez à votre IA.',
		lead: "Cahut s'installe comme un service sur un serveur Linux, avec son certificat HTTPS. Votre IA s'y connecte par MCP et publie sa première Window en quelques minutes.",
		allOptions: "Toutes les options d'installation",
		install: 'Installer',
		tabs: {
			server: { label: 'Sur un serveur', meta: 'AlmaLinux · Ubuntu · amd64 · arm64' },
			local: { label: 'Sur votre machine', meta: 'Git · make · Go' },
		},
		comments: {
			install: 'Installer Cahut comme service, et déclarer votre premier Cahut',
			account: 'Créer votre compte, et un jeton pour votre IA',
			connect: 'Ajouter votre Cahut à Claude Code, avec le jeton',
			build: 'Compiler Cahut depuis les sources',
			declare: 'Déclarer un Cahut sur localhost, avec votre compte et un jeton',
			serve: 'Le servir sur http://localhost:8080',
			connectLocal: "L'ajouter à Claude Code, avec le jeton",
			ask: 'Puis demandez à votre IA',
		},
		prompt: 'Fais un résumé de ce projet en une page, sous forme de document, et publie-le sur Cahut.',
		reply: "Publié à l'adresse",
		copy: 'Copier',
		copied: 'Copié',
		needs: {
			label: 'Il vous faut',
			items: [
				'Un serveur Linux avec systemd',
				'Un domaine qui pointe vers lui',
				'Les ports 80 et 443 ouverts',
				"Un client d'IA qui parle MCP, comme Claude Code",
			],
		},
		note: "Le dépôt s'ouvre au lancement public : d'ici là, ces commandes fonctionnent pour qui y a accès.",
		status: 'Où en est Cahut',
	},

	why: {
		label: 'Pourquoi Cahut',
		title: "Le fichier s'efface.",
		lead: 'Word, PowerPoint et Excel ont été conçus pour des humains qui écrivent, mettent en page et calculent à la main. Quand une IA produit le contenu, elle produit naturellement des pages vivantes, chacune à sa propre adresse.',
		text: "Vous n'envoyez plus de pièce jointe, vous partagez un lien : la page se met à jour, reçoit des commentaires, garde ses versions, et s'exporte en PDF ou en PowerPoint pour qui a encore besoin d'un fichier.",
		problemsTitle: "Aujourd'hui, ce que fabriquent les IA n'a pas de maison.",
		problems: [
			{
				title: 'Éparpillé',
				text: "Un lien Claude ici, un lien ChatGPT là, une présentation Gamma ailleurs. Chaque IA garde ce qu'elle fabrique sur son propre service.",
			},
			{
				title: 'Reconverti',
				text: 'Faute de mieux, le résultat redevient un PowerPoint ou un Word pour pouvoir être partagé et rangé.',
			},
			{
				title: 'Hors de contrôle',
				text: "Qui a accès, jusqu'à quand, quelle version est la bonne, où vivent les données. Personne ne le sait vraiment.",
			},
			{ title: 'Incohérent', text: "Chaque IA a son propre style, et la marque de l'organisation se perd." },
			{
				title: 'Sans garde-fou',
				text: "Confier un serveur à une IA, c'est lui donner un shell root, ou passer par une plateforme faite pour des humains, qu'elle sait à peine utiliser.",
			},
		],
	},

	files: {
		compare: 'Comparer',
		files: "Aujourd'hui : des fichiers",
		link: 'Avec Cahut : un lien',
		today: "Aujourd'hui",
		withCahut: 'Avec Cahut',
		items: [
			{ name: 'Q3 review v4 FINAL.pptx', meta: 'En pièce jointe, 8,2 Mo' },
			{ name: 'Q3 review v5 (2).pptx', meta: 'Transféré deux fois' },
			{ name: 'q3-notes.docx', meta: 'Reconverti' },
			{ name: 'Un lien de conversation', meta: "Sur un service d'IA" },
			{ name: 'Q3 numbers.xlsx', meta: 'Quelle version ?' },
			{ name: 'Une présentation ailleurs', meta: 'Sur un autre service' },
		],
		facts: ['Mise à jour il y a 2 minutes', 'Version 2, la précédente gardée', '3 commentaires'],
		captionFiles: 'Six fichiers et liens pour une seule présentation. Lequel est le bon, et qui peut le voir ?',
		captionLink:
			"Une seule adresse. La page se met à jour, reçoit des commentaires, garde ses versions, et s'exporte en PDF ou en PowerPoint.",
	},

	statement: {
		label: "L'idée",
		title: 'Vos IA fabriquent, Cahut fait vivre.',
		lead: "Tout ce qu'elles produisent est publié, partagé, rangé et protégé au même endroit, à vos couleurs, chez vous, quelle que soit l'IA.",
		text: "Cahut ne crée pas de contenu, il l'héberge. Il ne concurrence ni les IA ni les outils de création : c'est la maison neutre et auto-hébergée où vit leur travail.",
		layers: [
			{
				layer: 'Créer',
				who: 'Les IA',
				what: "Claude, ChatGPT, Mistral, Copilot et d'autres fabriquent le contenu : documents, présentations, tableaux de bord, applis.",
				tags: ['Claude', 'ChatGPT', 'Mistral', 'Copilot'],
			},
			{
				layer: 'Garder et partager',
				who: 'Cahut, sur votre serveur',
				what: 'Publication, liens, droits, versions, commentaires, recherche, export et Logbook.',
				tags: ['Publier', 'Partager', 'Versions', 'Commentaires', 'Recherche', 'Export', 'Logbook'],
			},
			{
				layer: 'Apparence',
				who: 'La house style, portée par Cahut',
				what: 'Vos couleurs et vos formats, donnés à chaque IA sous forme de modèles et de skills.',
				tags: ['Couleurs', 'Polices', 'Logos', 'Modèles'],
			},
		],
	},

	how: {
		label: 'Comment ça marche',
		title: "D'une demande à un lien, en cinq étapes.",
		lead: "Les IA restent dehors. Elles se connectent à Cahut par MCP, déposent leur travail par des opérations bornées, puis se déconnectent. Vos lecteurs l'ouvrent dans un navigateur.",
	},

	flow: {
		group: 'Comment fonctionne Cahut, en cinq étapes',
		step: 'Étape',
		of: 'sur 5',
		pause: 'Pause',
		play: 'Lecture',
		steps: [
			{
				title: 'Connecter',
				heading: 'Ajoutez Cahut à votre IA, une fois.',
				text: "Votre IA se connecte à votre Cahut par MCP, le protocole ouvert qui relie les IA aux outils. Vous l'autorisez avec OAuth, dans le périmètre que vous choisissez. Aucun modèle d'IA ne tourne sur votre serveur.",
				call: "MCP · OAuth · au nom d'Olivier",
			},
			{
				title: 'Publier',
				heading: "Demandez une présentation. L'IA publie une Window.",
				text: "L'IA appelle une opération typée, jamais un shell. Cahut applique votre house style, range la Window dans sa Room et renvoie son adresse.",
				call: 'publish_window(room: "marketing", window: "q3-review")',
			},
			{
				title: 'Partager',
				heading: 'Partagez un lien, pas un fichier.',
				text: "Chaque Window s'ouvre à sa Room seulement, à tout votre Cahut, à qui a sa Key, ou au public. Les lecteurs l'ouvrent dans un navigateur et la commentent.",
				call: 'share_window(window: "q3-review", access: "key")',
			},
			{
				title: 'Knock first',
				heading: "Ce qui ne s'annule pas vous attend.",
				text: "Rendre une Window publique ne s'annule pas. Alors l'IA frappe : la demande attend dans la console, avec une notification et une échéance, jusqu'à ce qu'une personne l'approuve.",
				call: 'share_window(access: "public") → en attente d\'Olivier',
			},
			{
				title: 'Logbook',
				heading: 'Chaque action est dans le Logbook.',
				text: "Quelle IA, au nom de qui, ce qu'elle a fait et avec quel résultat. Le Logbook garde chaque action, lisible dans la console.",
				call: 'logbook: 4 entrées, 1 approuvée par une personne',
			},
		],
		ais: 'Vos IA',
		aisNote: 'Elles restent dehors. Aucune IA ne tourne sur votre serveur.',
		server: 'Votre serveur',
		cahut: 'Votre Cahut',
		houseStyle: 'House style',
		room: 'Room · marketing',
		noWindows: 'Pas encore de Window',
		knock: 'Knock first',
		nothingWaits: 'Rien ne vous attend',
		waiting: 'En attente',
		approved: 'Approuvé',
		knockText: 'Veut rendre <b>Q3 review</b> publique.',
		approve: 'Approuver',
		decline: 'Refuser',
		knockDone: 'Approuvé par Olivier, 09:33',
		logbook: 'Logbook',
		log: [
			"s'est connecté pour Olivier",
			'a publié <b>Q3 review</b>',
			'a donné une Key à <b>Q3 review</b>',
			'a demandé à rendre <b>Q3 review</b> publique',
		],
		readers: 'Vos lecteurs',
		readerNames: [
			{ who: 'Équipe marketing', short: 'Équipe' },
			{ who: 'Un client, avec la Key', short: 'Key' },
			{ who: 'Tout le monde', short: 'Public' },
		],
		closed: 'Fermé',
		comment: 'Parfait',
		access: { room: 'Room seulement', internal: 'Interne', key: 'Key', public: 'Public' },
		packets: { linkKey: 'Lien et Key', link: 'Lien' },
	},

	knock: {
		label: 'Gouvernance',
		title: 'Knock first',
		lead: "L'IA est l'opératrice prévue de Cahut, mais jamais sans garde-fous. Ce qui s'annule, elle le fait seule. Ce qui ne s'annule pas attend un humain.",
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
			'Ajouter une personne à un Cahut',
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
		lead: "Le périmètre complet sort au lancement, sans version d'essai réduite. Aujourd'hui, la première tranche fonctionne : une IA publie une Window par MCP, et Cahut la sert en HTTPS. Le reste est la cible.",
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
						text: "Retrouvez n'importe quelle page, présentation ou tableau de bord parmi tout ce qui est publié.",
					},
					{ area: 'Export', text: 'PDF et PowerPoint, pour les destinataires encore sur Office.' },
				],
			},
			{
				title: 'Connecter',
				text: 'Toutes les IA, une seule porte, votre style.',
				features: [
					{
						area: 'Accès des IA',
						text: "Un serveur MCP natif (HTTP, OAuth) et une skill par client d'IA, en plugin quand le client le permet : Claude d'abord, puis ChatGPT, Mistral, Copilot et d'autres. Un accès SSH limité aux mêmes opérations, pour les IA en terminal.",
					},
					{
						area: 'House style',
						text: "Les couleurs, polices, logos et modèles de l'organisation, donnés à chaque IA pour que tout sorte à la charte.",
					},
				],
			},
			{
				title: 'Partager',
				text: 'Un lien, ouvert aux bonnes personnes.',
				features: [
					{
						area: 'Partage',
						text: 'Par lien, Window par Window : sa Room seulement, interne (tout le Cahut), une Key (un mot de passe) ou public, avec une date de fin facultative.',
					},
					{ area: 'Commentaires', text: 'Les retours des lecteurs, directement sur les pages.' },
					{
						area: 'Personnes et Rooms',
						text: "Les personnes de chaque Cahut et ses administrateurs, une Room personnelle pour chacun, des Rooms d'équipe avec propriétaires et membres.",
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
						text: 'Voir ce qui est publié, partager, commenter, approuver les actions irréversibles, administrer le serveur et ses Cahuts, sur son propre port HTTPS.',
					},
					{
						area: 'Domaines et TLS',
						text: 'Plusieurs Cahuts par serveur, chacun à son domaine ou sous-domaine. Pas de wildcard, des certificats automatiques.',
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
			'Tourner sur plusieurs machines : un serveur porte un ou plusieurs Cahuts.',
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
					'Votre instance, à son propre domaine. Un serveur peut en porter plusieurs, chacun avec ses personnes et sa base de données.',
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
			{ word: 'House style', meaning: "La charte de l'organisation, appliquée par toutes les IA." },
			{ word: 'Knock first', meaning: "La règle de gouvernance : l'IA frappe avant toute action irréversible." },
			{ word: 'Logbook', meaning: 'Le registre de toutes les actions.' },
		],
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
		lead: "Cahut est un projet personnel d'Olivier Durand, indépendant de toute entreprise et de tout fournisseur d'IA. Il est écrit en Go, en public, une étape après l'autre.",
		steps: [
			"Protéger le nom : organisation GitHub, noms des paquets, recherche d'antériorité de marque.",
			"Écrire l'architecture cible : comparer les options et trancher, avec leurs raisons.",
			'Construire la première tranche verticale en Go : installer sur un serveur, publier une page par MCP, la servir en HTTPS.',
			'Preuve de concept : Cahut installé sur un VPS vierge, une IA qui publie une présentation par MCP, la console.',
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
				threshold: "Le même Cahut fonctionne avec au moins deux clients d'IA différents",
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
			roadmap: "Les lots de la première version de Cahut, ce qui marche aujourd'hui, et ce qui la rend prête pour le lancement.",
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
			batchesTitle: 'Les lots de la première version',
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
					text: "Mettre à jour une Window garde la version précédente, et un geste y revient. Chaque Window s'ouvre à sa Room seulement, à tout son Cahut, à qui a sa Key, ou au public.",
				},
				{
					title: 'Knock first',
					text: "L'IA demande avant tout ce qu'elle ne peut pas annuler. La demande attend dans la console, avec une notification et une échéance, jusqu'à ce qu'une personne l'approuve ou la refuse.",
				},
				{
					title: 'House style et skill Claude',
					text: "Les couleurs, polices, logo et modèles de l'organisation, donnés à chaque IA, et une skill qui dit à Claude comment publier sur Cahut.",
				},
				{
					title: 'Installation, premier démarrage et mises à jour',
					text: 'Une commande sur un serveur AlmaLinux ou Ubuntu vierge, un premier démarrage guidé dans le navigateur, et des mises à jour depuis les releases, avec retour arrière.',
				},
				{
					title: 'Deuxième client IA et recette',
					text: "Un deuxième client d'IA testé de bout en bout, une revue de sécurité, et les critères de réussite vérifiés sur des serveurs vierges.",
				},
			],
			milestones: 'Les jalons sur GitHub',
			todayTitle: "Ce qui marche aujourd'hui",
			todayLead: "Les lots 0 à 6 sont faits. Aujourd'hui, dans un Cahut :",
			today: [
				'Un service sur un serveur Linux, qui porte un ou plusieurs Cahuts, chacun à son domaine avec son certificat.',
				"Des personnes qui se connectent, des administrateurs du Cahut et du serveur, et des Rooms d'équipe avec propriétaires et membres.",
				'Une console web : Knocks, Rooms, Windows, Logbook, personnes, IA connectées et jetons personnels.',
				"Des clients d'IA qui se connectent par OAuth, aux Rooms et dans la portée qu'une personne accorde, ou avec un jeton personnel.",
				"Des Windows publiées par MCP : documents, présentations, tableaux de bord et sites statiques, servis dans un bac à sable qui les tient à ce que leur IA déclare charger et appeler.",
				"Deux versions de chaque Window avec retour en arrière, une corbeille, et un accès : Room seulement, interne, avec une Key ou public, jusqu'à une date.",
				"Knock first : ce qu'une IA ne peut pas défaire, ou ce qui change ce qu'une personne peut faire, attend sa personne, qui l'approuve ou le refuse dans la console.",
				"Une house style pour chaque Cahut : ses couleurs, ses polices, son logo et ses consignes, des gabarits de documents, de présentations et de tableaux de bord, et un moteur de présentation, que les IA lisent avant de publier.",
				"Un plugin pour Claude Code, et une skill pour les applications Claude, qui apprennent à Claude à faire des Windows dans la house style.",
				'Le Logbook de chaque appel, de chaque changement et de chaque connexion.',
			],
			caution:
				"Pas encore de sauvegardes : gardez des instantanés de votre serveur, et ne confiez pas encore à Cahut un travail que vous ne pouvez pas perdre.",
			statusText: 'La documentation garde le détail, domaine par domaine :',
			status: 'Où en est Cahut',
			afterTitle: 'Après la première version',
			afterLead: 'Puis, vers le périmètre complet du lancement, à peu près dans cet ordre :',
			after: [
				"Les commentaires, la recherche, l'authentification unique, les e-mails, les sauvegardes et leur restauration, et des quotas par Cahut.",
				'Les applis dans des conteneurs sans privilèges : environnements, mode Dockerfile, bases de données, secrets, journaux et limites de ressources.',
				"L'export en PDF et en PowerPoint, l'accès SSH limité aux mêmes opérations, et d'autres clients d'IA.",
			],
		},
	},

	closing: {
		title: 'Partagez un lien, pas un fichier.',
		lead: "Cahut se construit en public, une étape après l'autre. Suivez le projet sur GitHub, ou lisez dans la documentation comment il s'assemble.",
		github: 'Suivre le projet sur GitHub',
		docs: 'Lire la documentation',
	},
};

export default fr;
