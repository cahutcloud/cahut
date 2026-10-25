---
name: charte-graphique-cahut
description: Charte graphique et ton de Cahut, la maison auto-hébergée de ce que fabriquent les IA. Logo SVG et ses versions, palette Ink, Paper, Sky, Light Grey et Soft Yellow avec échelles et règles de contraste, polices Plus Jakarta Sans et IBM Plex Mono, jetons CSS (thèmes clair et sombre), kit d'interface de la console (classes c-), pictogrammes du vocabulaire (Cahut, Room, Window, Key, House style, Knock first, Logbook), icônes Lucide, illustrations, photos Unsplash, gabarits HTML de diapositives et de visuels de réseaux avec un script de rendu en PNG ; voix calme et précise en anglais simple pour les textes du produit. À utiliser dès qu'il faut créer, habiller ou relire un support de Cahut (site cahut.cloud, landing page, écran ou composant de la console, maquette, présentation, diapositive, image de partage, bannière, e-mail, signature, illustration, documentation) ou écrire un texte du produit (interface, message d'erreur, site, annonce), même si le mot « charte » n'est pas prononcé.
---

# Charte graphique Cahut

Cette skill sert à produire des supports **fidèles à Cahut** et à les relire. Elle résume la charte et dit par quels fichiers commencer ; la référence complète reste [`docs/charte-graphique/README.md`](../../../docs/charte-graphique/README.md) (version 1.4 du 25 octobre 2026), dont tous les fichiers sources sont rangés dans `docs/charte-graphique/`.

Dans ce document, les chemins partent de la racine du dépôt et **`charte/`** désigne `docs/charte-graphique/`.

## L'identité en une phrase

Un abri simple et à soi, posé dans le nuage : **papier chaud, encre bleu-gris, une touche de ciel et un rayon de soleil rare**, des formes arrondies et beaucoup d'air, et une voix **calme, précise et accueillante** qui parle la langue de la maison en anglais simple. Cahut accueille le travail des IA, elle ne le crée pas : rien dans l'image ne doit faire « IA magique ».

## Périmètre : la charte Cahut, ou celle du client

- **Charte Cahut** : tout ce qui est à Cahut elle-même (site, documentation, console, écran de connexion, e-mails envoyés par une Cahut, présentations, réseaux sociaux).
- **Charte de l'organisation (*house style*)** : les pages que Cahut publie pour ses utilisateurs. La marque de l'organisation passe en premier ; Cahut s'efface. Mention facultative en pied de page avec `charte/logos/cahut-badge-hosted.svg` (ou `-inverse`), 32 px de haut au plus. Jamais le logo Cahut au-dessus ou à côté de celui d'un client.

## Comment travailler

1. **Identifier le support** et partir du bon fichier :

   | Support | Point de départ |
   |---|---|
   | Écran ou composant de la console, page servie par Cahut | `charte/ui/cahut-ui.css` (classes `c-`), modèles `charte/ui/console.html` et `charte/ui/composants.html` |
   | Page web, landing, site cahut.cloud | `charte/tokens/cahut-tokens.css` et `charte/tokens/cahut-fonts.css`, puis les composants de `cahut-ui.css` |
   | Site cahut.cloud ou documentation docs.cahut.cloud (dossier `sites/`, Astro) | Les mêmes fichiers, importés par l'alias `@charte` (`@charte/tokens/cahut-tokens.css`), jamais copiés : voir `sites/README.md` |
   | Page autonome hors du dépôt (artifact, page publiée ailleurs) | Recette « Page autonome » de `references/recettes.md` |
   | Diapositive, image de partage, bannière, avatar | Le gabarit voisin dans `charte/applications/gabarits/`, rendu avec `scripts/rendre-gabarit.sh` |
   | E-mail, signature | `charte/applications/gabarits/signature-email.html` et la recette « E-mail » |
   | Texte du produit, du site ou d'une annonce | Section « Voix et ton » ci-dessous, puis § 11 de la charte |

2. **Ne jamais retaper une valeur** : couleurs, tailles, espacements, rayons et ombres viennent des jetons (`var(--color-*)` pour les rôles, `var(--cahut-*)` pour les valeurs brutes). Le logo, les pictogrammes et les illustrations s'utilisent depuis leurs fichiers SVG.
3. **Lire la section de la charte** qui concerne le support quand un détail manque (zone de protection, hiérarchie complète, contrastes, formats des visuels). Les recettes pratiques sont dans [`references/recettes.md`](references/recettes.md).
4. **Rendre et regarder** le résultat : capture du rendu, thèmes clair et sombre pour une interface, largeur de 375 px pour une page.
5. **Relire avec la liste de contrôle** en fin de fichier.

## Les règles visuelles essentielles

**Couleurs.** Cinq couleurs, dans ces proportions indicatives :

| Couleur | HEX | Rôle | Part |
|---|---|---|---|
| Paper | `#F7F5F0` | Fond principal | 60 % |
| Ink | `#1F2A33` | Texte, logo, traits, fonds sombres | 22 % |
| Sky | `#5FA8D3` | Accent, états actifs | 10 % |
| Light Grey | `#D9DCDF` | Bordures, séparateurs, reliefs (= Ink 200) | 5 % |
| Soft Yellow | `#F4D27A` | Lumière, présence, ce qui attend une décision (*knock first*) | 3 % |

- Chaque couleur a une échelle de 50 à 950 dans les jetons. Repères : Ink = Ink 900, Sky = Sky 400, Soft Yellow = Yellow 300.
- **Contraste (WCAG 2.2 AA)** : sur Sky, on écrit en Ink, jamais en blanc. Sky et Soft Yellow ne portent jamais de texte sur Paper. Liens en Sky 700 (`--color-link`), texte secondaire en Ink 600. Une information ne passe jamais par la couleur seule : icône ou mot à côté.
- **Un seul accent Sky par écran, un seul jaune.** Le jaune reste rare.
- **Mode sombre** : fond Ink 950, surfaces Ink 900 et 800, texte Paper, texte secondaire Ink 300, accent Sky 400, liens Sky 300. Les jetons basculent seuls (`data-theme="dark"` ou réglage du système) ; le logo passe en version inversée.

**Typographie.**

- **Plus Jakarta Sans** (400 à 800) pour les titres, le texte et l'interface. **IBM Plex Mono** (400 à 600) pour tout ce qui se tape ou se copie : adresses (`cahut.cloud`), commandes, codes couleur, horodatages du *logbook*, étiquettes.
- Hiérarchie : Display 56/64 px 800 (−2 %), Titre 1 40/48 800 (−1 %), Titre 2 32/40 700 (−1 %), Titre 3 24/32 700, Titre 4 20/28 600, texte 16/24 400, petit texte 14/20, étiquette Mono 12/16 500 en capitales (+12 %). Classes `c-display`, `c-h1` à `c-h4`, `c-label`, `c-mono`.
- Titres en casse de phrase, sans point final sauf pour une phrase d'accroche. Texte aligné à gauche, jamais justifié, lignes de 60 à 80 caractères. Deux graisses au plus par bloc.
- Polices de secours : `system-ui` et `ui-monospace` sur le web, Arial et Consolas dans un PowerPoint destiné à un poste sans les polices.

**Mise en page.**

- Grille de 12 colonnes, gouttières de 24 px, marges de 32 px (16 px sur mobile), largeur de lecture de 720 px au plus.
- Espacements en multiples de 4 px (`--space-*`). Rayons : 6 px petits contrôles, 10 px boutons et champs, 16 px cartes et photos, 24 px grands panneaux.
- Bordures de 1 px en Light Grey ; ombres très légères (`--shadow-*`), jamais de lueur colorée ni de dégradé.
- Fond Paper, cartes en blanc, Ink pour les moments forts (section, citation, pied de page). Au moins un tiers de vide dans une composition.

**Logo.** Un nuage d'un seul trait dont la base s'ouvre sur une porte, et le logotype « cahut » en minuscules.

| Fond | Fichier |
|---|---|
| Clair (Paper, blanc) | `cahut-logo.svg`, `cahut-logo-horizontal.svg`, `cahut-symbole.svg` |
| Ink ou sombre | les mêmes en `-inverse` |
| Impression une couleur | `-noir`, `-blanc` |
| Symbole sous 64 px | `cahut-symbole-petites-tailles.svg` (trait renforcé) |
| Icône d'application | `cahut-icone-app.svg` (Paper, symbole Ink, porte éclairée), `cahut-icone-app-ink.svg` |
| Avatar sur GitHub, LinkedIn, X | `charte/applications/avatar-reseaux.png` : carré Paper plein, symbole Ink et porte éclairée, sans coins arrondis, le réseau pose son masque |

- Tous dans `charte/logos/` (exports PNG dans `logos/png/`, favicons dans `logos/favicon/`). Logo horizontal pour les en-têtes et barres de navigation.
- Zone de protection : la largeur de la porte (environ un quart de la largeur du symbole). Tailles minimales : logo 96 px, horizontal 100 px, symbole 24 px, favicon 16 px.
- Ne pas déformer, recolorer, incliner, ombrer, recomposer ni remplir le logo (seule exception : la porte éclairée en Soft Yellow de l'icône d'application et de l'avatar) ; ne pas redessiner le logotype dans une police ; pas de fond à moins de 4,5:1 ni de photo chargée derrière.

**Pictogrammes et icônes.**

- Les **pictogrammes** (`charte/icones/`, grille de 64 px, trait de 3 px) illustrent le vocabulaire de la maison sur le site, la documentation, les écrans vides et l'accueil. Trait en `currentColor`, un seul accent. Taille minimale 32 px.
- Les **icônes d'interface** sont des [Lucide](https://lucide.dev) (grille de 24 px, trait de 2 px, couleur du texte), en 16, 20 ou 24 px. Une action importante a toujours une étiquette texte à côté de son icône.

**Illustrations et photos.**

- Illustrations en aplats, sans dégradé ni ombre, traits Ink arrondis, collines en Ink 100 et 200, soleil Soft Yellow, accents Sky (`charte/illustrations/`). La porte éclairée en jaune veut dire « quelqu'un est là ».
- Photos à la lumière naturelle, tons clairs et doux, lieux réels et calmes, en quatre familles : ciel, maison, lumière, travail. Sélection et crédits dans `charte/photos/`. Coins arrondis de 16 px dans l'interface ; texte sur photo seulement sur un bandeau Paper ou un voile Ink à 60 %.
- **Jamais** : robots, cerveaux, circuits lumineux, étincelles « magiques », personnages 3D, ambiances sombres et néons, poignées de main de banque d'images, filtres appuyés.

**Composants de la console** (`cahut-ui.css`) : bouton principal Ink (`c-btn--primary`, un seul par zone) ; bouton accent Sky avec texte Ink (`c-btn--accent`) pour publier ou connecter ; badges d'accès d'une Window toujours avec icône et mot (`c-badge--public` Sky, `--key` jaune, `--internal` neutre, `--room` neutre atténué avec cadenas) ; carte *knock first* à liseré Soft Yellow (`c-knock`) qui dit quelle IA, quoi, au nom de qui, l'échéance, avec Approve et Decline ; *logbook* (`c-log`) avec l'heure en mono, l'IA, l'action, le résultat ; anneau de focus Sky de 3 px toujours visible au clavier.

## Voix et ton

Les textes du produit, du site et des visuels sont **en anglais**, langue du lancement. La charte et le brief restent en français.

| Mot | Ce qu'il désigne |
|---|---|
| Cahut | Une instance sur le serveur de l'utilisateur, à son propre domaine ; un serveur peut en porter plusieurs |
| Room | Un espace personnel ou d'équipe, à l'adresse `<domaine>/<room>` |
| Window | Une page vivante : document, présentation, tableau de bord, appli, à l'adresse `<domaine>/<room>/<window>` |
| Key | Le mot de passe qui ouvre une Window protégée |
| House style | La charte de l'organisation, appliquée par toutes les IA |
| Knock first | L'IA frappe avant toute action irréversible |
| Logbook | Le registre de toutes les actions |

- Une Window s'ouvre à sa Room seulement (*Room only*, par défaut), à toute la Cahut (*Internal*), à qui a sa Key, ou au public (*Public*). Le mot Door n'existe plus : on parle de l'accès d'une Window.
- Cahut, Room, Window et Key prennent une capitale dans le texte courant (« 12 Windows, 4 members », « 3 Rooms in your Cahut »).
- Phrases courtes, verbes concrets : *publish*, *share*, *approve*, *roll back*. Dire qui agit : quelle IA, au nom de qui, avec quel résultat.
- Rassurer sans alarmer : ni capitales d'insistance, ni point d'exclamation, ni emoji d'alerte. Pas de promesse magique sur l'IA.
- Casse de phrase pour les titres, boutons et menus. Chiffres en chiffres.
- **Cahut** avec une capitale dans le texte, jamais « CAHUT » ; « cahut » en minuscules seulement dans le logo.
- En français, Cahut est féminin, le produit comme chaque instance : « une Cahut », « votre Cahut est prête », « Cahut accueille le travail, elle ne le crée pas ».
- Signature : *Your own cloud for everything your AI makes.*

| Écrire | Éviter |
|---|---|
| Claude wants to make *Q3 review* public. | AI ACTION REQUIRED!! |
| Share a link, not a file. | Revolutionize your AI workflow. |
| Published. The previous version is kept. | Success! Your content has been deployed. |
| This Window closed yesterday. | Error 403: access forbidden. |

## Liste de contrôle avant de livrer

**Visuel**
- [ ] Couleurs, typographie, espacements et rayons tirés des jetons, aucune valeur retapée ni teinte hors palette.
- [ ] Paper dominant, un seul accent Sky, jaune rare et justifié (lumière ou attente).
- [ ] Aucun texte blanc sur Sky, aucun texte Sky ou jaune sur Paper ; contraste de 4,5:1 pour le texte courant.
- [ ] Plus Jakarta Sans pour le texte, IBM Plex Mono pour adresses, commandes, codes et horodatages.
- [ ] Logo issu du bon fichier pour son fond, intact, avec sa zone de protection et au-dessus de sa taille minimale.
- [ ] Aucune ombre lourde, lueur, dégradé décoratif, robot ou étincelle.
- [ ] Interface lisible au clavier (focus visible), correcte en clair et en sombre, lisible à 375 px de large.
- [ ] Sur une page publiée pour un client : sa marque d'abord, Cahut au plus en badge de pied de page.

**Texte**
- [ ] En anglais simple, phrases courtes, casse de phrase, chiffres en chiffres.
- [ ] Vocabulaire de la maison employé à bon escient et avec ses capitales.
- [ ] Qui agit est dit ; aucun point d'exclamation, aucune capitale d'insistance, aucune promesse magique.
- [ ] « Cahut » bien écrit.

## Fichiers

| Fichier | Contenu | Quand le lire |
|---|---|---|
| `docs/charte-graphique/README.md` | La charte complète : marque, logo, couleurs, typographie, icônes, illustrations, photos, mise en page, interface, applications, voix, fichiers | Tout détail absent d'ici |
| `docs/charte-graphique/index.html` | Version visuelle de la charte (et `charte-graphique-cahut.pdf`) | Pour voir le rendu attendu |
| `charte/tokens/cahut-tokens.css`, `cahut-tokens.json` | Jetons : couleurs brutes et rôles, typographie, espacements, rayons, ombres, mouvement, thèmes clair et sombre | Tout support |
| `charte/tokens/cahut-fonts.css`, `charte/typographie/polices/` | Polices web WOFF2 et TTF (OFL) | Toute page, tout document |
| `charte/ui/cahut-ui.css`, `composants.html`, `console.html` | Kit d'interface et maquettes de la console (`#dark` pour le mode sombre) | Écran, composant, page servie par Cahut |
| `charte/logos/` | SVG maîtres, PNG, favicons, badge *hosted by Cahut* | Tout support |
| `charte/icones/`, `charte/illustrations/` | Pictogrammes du vocabulaire, illustrations et motif de nuages | Site, documentation, écrans vides, visuels |
| `charte/photos/` | Planche, aperçus, `selection.json` (crédits), `telecharger-photos.sh` (HD) | Tout support avec photo |
| `charte/applications/gabarits/` | Gabarits HTML : diapositives titre, section et contenu, image de partage, aperçu GitHub, couverture de page et bannière de profil LinkedIn, bannière X, signature e-mail | Présentation, réseaux, e-mail |
| [`references/recettes.md`](references/recettes.md) | Recettes : page dans le dépôt, page autonome, e-mail, icônes, gabarits et rendu PNG, photos, présentation, modification de la charte | Avant de produire |
| `scripts/rendre-gabarit.sh` | Rend une page HTML de taille fixe en PNG avec Chrome sans interface | Après chaque gabarit modifié ou créé |
| `scripts/imprimer-charte.mjs` | Réimprime `charte-graphique-cahut.pdf` depuis `index.html` avec Chrome sans interface (Node 22 ou plus) | Après chaque changement de la charte |
