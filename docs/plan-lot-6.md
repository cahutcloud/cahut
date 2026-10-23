# Cahut · Plan du lot 6 : house style et skill Claude

23 octobre 2026 · Olivier Durand · Validé, achevé

## En bref

Le lot 6 donne à chaque Cahut sa house style : le nom de l'organisation, ses couleurs, ses polices, son logo et des consignes pour les IA. Cahut la sert sous `/-/`, à une adresse stable, avec un gabarit pour chaque type de Window et un petit moteur de présentation ; une IA la lit avec `get_house_style` avant de publier, et une Window qui lie la feuille prend les couleurs de l'organisation, aujourd'hui comme après un changement de marque. Les pages de Cahut pour les lecteurs la portent aussi. Les administrateurs la règlent dans la console, ou par leur IA, et reviennent à la précédente d'un clic. Une skill pour Claude, empaquetée en plugin, apprend à Claude les formats et les règles de Cahut. Le lot compte six pull requests, pour environ huit jours à temps plein, et s'ouvre sur une décision écrite (ADR 22).

**Fin de lot**, reprise du [plan de la V0](plan-v0.md) : un même brief donne une présentation aux couleurs de l'organisation, publiée en une seule demande.

## Point de départ

Ce que les lots 1 à 5 ont laissé :

- Les pages de Cahut sur le port 443 (accueil, Room, Key, introuvable, lien expiré) sont rendues par le serveur dans un style neutre, `/-/cahut.css`, avec un thème sombre, sous une CSP qui n'admet que leurs propres fichiers ([ADR 19](adr/0019-window-versions-access-and-declaration.md)).
- La politique des fichiers d'une Window admet l'origine de son Cahut pour les styles, les polices, les images et les scripts : une Window charge ce que Cahut sert sous `/-/` sans le déclarer. Les fichiers d'une Window protégée sont servis dans un cadre, sous `/-/w/<capacité>/`, et `frame.js` passe le fragment de l'adresse au cadre.
- Les consignes du serveur MCP demandent des liens relatifs, jamais `/css/deck.css`, et ne disent rien de la house style.
- `/-/` est réservé aux pages et aux adresses de Cahut ([ADR 9](adr/0009-rooms-and-windows-as-paths.md)) ; l'[ADR 15](adr/0015-license.md) met les gabarits et les skills sous Apache-2.0.
- La console a un écran par sujet, réservé aux administrateurs quand il touche tout le Cahut (People). Elle n'appelle que l'API publique, et la marque de Cahut y reste.

## Décision à écrire d'abord

### ADR 22 · Each Cahut has a house style, served to its Windows and given to AIs

L'[ADR 22](adr/0022-house-style-of-each-cahut.md) fixe :

- **Le contenu.** Un nom ; quatre couleurs (`text`, `background`, `primary`, `secondary`) et jusqu'à huit couleurs de séries pour les graphiques, Cahut dérivant les autres ; deux polices, titres et texte, chacune une famille, une famille générique de repli et ses fichiers ; un logo et sa version pour fond sombre ; des consignes pour les IA, 4 000 caractères au plus. Un texte qui contraste à moins de 4,5:1 sur le fond est refusé ; une couleur principale peu lisible passe avec un avertissement. Limites : 512 Kio par image, 1 Mio par fichier de police, 16 fichiers de police, 8 Mio en tout.
- **La house style neutre**, tant qu'aucun administrateur n'en règle une : polices du système, texte sombre sur blanc, un bleu, pas de logo. Ce n'est pas la marque de Cahut.
- **Le changement.** `set_house_style` change ce qu'il nomme et garde le reste ; la house style d'avant devient la précédente, que `rollback_house_style` rétablit. Comme elle s'annule, elle ne frappe pas. Les administrateurs du Cahut la changent, et leurs IA qui atteignent toutes les Rooms ; une IA limitée à certaines Rooms la lit seulement. Tout le monde la lit, les IA en lecture seule comprises.
- **Ce que Cahut sert sous `/-/`.** `/-/house-style.css`, `/-/house-style/logo`, `/-/house-style/logo-inverse`, `/-/house-style/fonts/<fichier>`, `/-/deck.js`, et `/-/templates/<type>` pour voir chaque gabarit. Le tout est public, sauf les consignes, servi avec `Access-Control-Allow-Origin: *` et revalidé à chaque lecture. Sans logo, Cahut écrit le nom de l'organisation en SVG.
- **L'exception aux liens relatifs.** Une Window lie la house style par des chemins absolus de son domaine, `/-/house-style.css`, qui marchent depuis son adresse comme depuis le cadre d'une Window protégée. L'adresse est stable : changer la house style change toutes les Windows qui la lient.
- **Les gabarits et les classes.** Un gabarit par type (document, présentation, tableau de bord ; un site part de celui du document), dans le binaire, le même pour chaque Cahut. Des classes préfixées `hs-`, actives seulement sous `hs-document`, `hs-deck` et `hs-dashboard`. Classes, variables et moteur forment un contrat avec les Windows publiées : une version en ajoute, n'en renomme et n'en retire aucune avant une version majeure. Le tout sous Apache-2.0.
- **Le moteur de présentation.** Des diapositives 16:9 qui s'adaptent à l'écran, une à la fois au clavier, par boutons et par glissement ; le numéro dans l'adresse (`#5`) ; plein écran ; impression d'une diapositive par page. Sans scripts, les diapositives se suivent sur la page.
- **Pour les IA.** `get_house_style`, plutôt qu'une ressource MCP ; les consignes du serveur MCP ; une skill pour Claude, la même pour chaque Cahut, empaquetée en plugin `cahut` dans une marketplace de ce dépôt, qui déclare aussi le serveur MCP du Cahut dont l'utilisateur donne le domaine ; la même skill en archive, servie par chaque Cahut pour les applications Claude où l'on téléverse les skills.
- **Les pages de Cahut** portent la house style : le logo ou le nom en haut, ses couleurs, ses polices, sans thème sombre. La console garde la marque de Cahut.
- **Logbook.** Ce qui change, jamais le contenu d'un fichier ; pas la lecture de la feuille par les lecteurs.

## Modèle de données

Une migration, `cahut/0011_house_style` :

| Table | Contenu |
| --- | --- |
| `house_styles` | Identifiant, état (`current` ou `previous`), réglages en JSON (nom, couleurs, polices, consignes), auteur, date |
| `house_style_files` | House style, nom (`logo`, `logo-inverse`, `fonts/<fichier>`), type, taille, empreinte SHA-256, contenu |

Sans ligne, le Cahut a la house style neutre. Les fichiers sont dans la base plutôt que sur le disque : ils sont petits, suivent la transaction de l'appel et n'ont pas besoin d'`AfterCommit` ; leur empreinte sert d'`ETag`. Un changement copie la ligne courante en précédente et efface l'ancienne précédente avec ses fichiers.

## Opérations

| Opération | Qui l'appelle | Ce qu'elle fait |
| --- | --- | --- |
| `get_house_style` | toute personne, toute IA, en lecture seule comprise | Le nom, les couleurs (dérivées comprises), les polices en piles CSS, les adresses de la feuille, des logos et du moteur, les consignes, le guide des classes ; avec `kind`, le gabarit de ce type |
| `set_house_style` | administrateurs du Cahut, et leurs IA qui atteignent toutes les Rooms | Change ce qu'elle nomme, garde le reste ; `neutral` repart de la house style neutre ; rend la house style et ses avertissements |
| `rollback_house_style` | les mêmes | Rétablit la house style précédente ; celle qu'on quitte devient la précédente |

L'entrée de `set_house_style`, en abrégé :

```json
{
  "name": "Example Co",
  "colors": { "primary": "#0b5fff", "secondary": "#ff7a00" },
  "chart_colors": ["#0b5fff", "#ff7a00", "#12b886"],
  "heading_font": { "family": "Fraunces", "fallback": "serif", "files": [{ "weight": 600, "style": "normal", "content": "<base64>" }] },
  "logo": { "content": "<svg …>" },
  "guidance": "Sentence case for titles. One idea per slide."
}
```

Un logo donné vide l'efface. Une police donnée remplace la précédente, fichiers compris. Les messages d'erreur disent à l'IA quoi corriger : « colors.text #9a9a9a on colors.background #ffffff contrasts 2.8:1; text needs 4.5:1 ».

## Ce que Cahut sert sous `/-/`

| Adresse | Contenu |
| --- | --- |
| `/-/house-style.css` | Les variables `--hs-…` (couleurs, polices, logo), les `@font-face`, puis les classes des gabarits, servies par Cahut |
| `/-/house-style/logo`, `/-/house-style/logo-inverse` | Le logo et sa version pour fond sombre ; sans logo, le nom en SVG |
| `/-/house-style/fonts/<fichier>` | Les fichiers des polices |
| `/-/deck.js` | Le moteur de présentation |
| `/-/templates/document`, `deck`, `dashboard` | Chaque gabarit rempli d'un contenu d'exemple, à la house style du Cahut |
| `/-/cahut-skill.zip` | La skill pour Claude, en archive |

Les pages de Cahut chargent `/-/house-style.css` puis leur propre feuille, `/-/pages.css`, qui remplace `/-/cahut.css` et ne parle qu'en variables `--hs-…`. Leur politique ajoute `font-src 'self'`. `frame.js` donne le focus au cadre d'une Window protégée, pour que les flèches mènent la présentation.

## Gabarits et moteur

- **La feuille.** Les variables d'abord : `--hs-text`, `--hs-background`, `--hs-primary`, `--hs-on-primary`, `--hs-secondary`, `--hs-muted`, `--hs-line`, `--hs-surface`, `--hs-chart-1` à `--hs-chart-8`, `--hs-font-heading`, `--hs-font-body`, `--hs-font-mono`. Puis trois racines : `hs-document` (titres, paragraphes, listes, citations, code, tableaux, figures, encadrés, en-tête au logo, pied de page), `hs-deck` (diapositives de titre, de section et de contenu, colonnes, grand chiffre, image pleine, notes de l'orateur masquées, pied au logo et au numéro), `hs-dashboard` (en-tête, rangée d'indicateurs avec leur évolution, panneaux en grille, tableaux, barres en CSS par une variable `--value`). Des classes communes : colonnes, carte, badge, encadré.
- **Les graphiques.** Pas de moteur de graphiques : l'IA dessine en SVG avec les couleurs `--hs-chart-…`, ou déclare une bibliothèque d'un CDN et lui passe les couleurs que donne `get_house_style`.
- **Le moteur**, `/-/deck.js`, sans dépendance, quelques kilo-octets : flèches, espace, Page précédente et suivante, Début et Fin, `f` pour le plein écran, boutons au survol, glissement au doigt, numéro dans l'adresse, barre de progression, impression d'une diapositive par page. Les tailles suivent la largeur de la diapositive en unités de conteneur (`cqi`), sans calcul en JavaScript.
- **Les gabarits**, rendus par le serveur avec le nom, le logo et les couleurs du Cahut, puis remplis d'exemples que l'IA remplace : un document de deux pages, une présentation de six diapositives, un tableau de bord de quatre indicateurs et trois panneaux. Une présentation qui se sert du moteur déclare des scripts ; un document et un tableau de bord n'en ont pas besoin.
- **Le guide des classes**, un texte court embarqué dans le binaire, que `get_house_style` rend et que `make docs-reference` écrit dans la documentation. Un test vérifie que chaque classe de la feuille y figure.
- **La licence.** Le dossier des gabarits, de la feuille et du moteur porte un fichier `LICENSE` Apache-2.0.

## MCP, skill et plugin

- **Les consignes du serveur MCP** : appeler `get_house_style` avant de faire une Window, lier `/-/house-style.css`, partir du gabarit de son type, suivre les consignes de l'organisation ; les liens relatifs restent la règle pour les fichiers de la Window.
- **La skill**, `plugins/cahut/skills/cahut/SKILL.md`, avec ses références : les formats (document, présentation, tableau de bord, site), ce qui marche dans le bac à sable et ce qui n'y marche pas (stockage, cookies, fenêtres), la déclaration, les images en base64 et la taille d'une publication, les versions et le retour arrière, l'accès et la Key, Knock first et le lien à transmettre. Elle laisse la house style à `get_house_style`.
- **Le plugin** `cahut`, dans `plugins/cahut/`, sous Apache-2.0 : son manifeste, la skill, et le serveur MCP `https://<domaine>/mcp`, dont l'utilisateur donne le domaine à l'installation (`userConfig`). Sa version suit celle de Cahut.
- **La marketplace**, `.claude-plugin/marketplace.json` à la racine du dépôt : `/plugin marketplace add getcahut/cahut`, puis `/plugin install cahut@cahut`. `claude plugin validate --strict` passe sur l'une et l'autre, lancé à la main : la CI n'installe pas Claude Code, et un test Go vérifie le manifeste, la marketplace et les chemins qu'ils nomment.
- **L'archive** de la skill, construite au démarrage depuis les fichiers que le binaire embarque, par un fichier Go dans `plugins/`.

## Console

| Écran | Contenu |
| --- | --- |
| House style, `/house-style` | Pour les administrateurs, sous This Cahut. Le nom ; le logo et sa version pour fond sombre, avec leur aperçu ; les quatre couleurs et celles des graphiques, avec le contraste de chacune ; les deux polices, leur famille, leur repli et leurs fichiers, avec le poids et le style devinés du nom du fichier ; les consignes pour les IA. Un aperçu en direct, avant l'enregistrement : un titre, un paragraphe, une diapositive, une rangée d'indicateurs. Les liens vers les trois gabarits. « Previous version », avec sa date, et « Back to neutral », après une confirmation. Un rappel : les polices sont servies à tous les lecteurs, il faut une licence qui le permette |
| AIs and access | Une carte « Skill for Claude » : le plugin pour Claude Code, l'archive pour les applications Claude |

L'aperçu applique les variables du formulaire à un bloc de la console et charge les polices choisies par l'API `FontFace`, sans rien envoyer au serveur.

## Tests

- **Contenu** : couleurs au format `#rrggbb`, contraste refusé sous 4,5:1, avertissements, couleurs dérivées ; familles de polices sans caractère qui sorte du CSS ; fichiers de police et logos reconnus par leurs premiers octets, refusés sinon ; limites de taille et de nombre.
- **Changement** : un changement partiel garde le reste ; un logo vide l'efface ; `neutral` ; la précédente rétablie, puis celle qu'on a quittée ; une troisième version efface la première et ses fichiers.
- **Droits** : un membre, une IA de membre et une IA limitée à certaines Rooms ne la changent pas ; l'IA d'un administrateur la change sans frapper ; une IA en lecture seule la lit.
- **Service** : la feuille reflète la house style et change avec elle ; `ETag` et 304 ; `Access-Control-Allow-Origin`, `nosniff` et la politique par défaut sur chaque fichier ; le logo SVG d'un nom à échapper ; un gabarit servi dans le bac à sable d'une Window ; l'archive de la skill.
- **Windows** : une Window qui lie la feuille la charge sous sa politique, à son adresse comme dans le cadre d'une Window protégée, sans rien déclarer.
- **Pages de Cahut** : le logo ou le nom en haut, la page d'une Key à la house style.
- **Cloisonnement** : la house style d'un Cahut n'apparaît pas dans un autre.
- **Logbook** : ce qui change, sans contenu de fichier.
- **Plugin** : le manifeste, la marketplace et les chemins qu'ils nomment ; la skill de l'archive est celle du plugin.
- **Navigateur** : dans le panneau de navigateur, l'écran House style et son aperçu ; une présentation au clavier et à 375 px, dans une Window publique et dans le cadre d'une Window protégée ; l'impression ; un document et un tableau de bord. Firefox et Safari restent à faire, comme depuis le lot 4.

## Découpage en pull requests

| # | Pull request | Contenu | Durée |
| --- | --- | --- | --- |
| 1 | Décision | Ce plan et l'ADR 22 | 0,5 j |
| 2 | House style | Migration 0011, paquet `housestyle`, `get_house_style`, `set_house_style`, `rollback_house_style`, la feuille et les fichiers sous `/-/`, les pages de Cahut à la house style | 2 j |
| 3 | Gabarits et moteur | Les classes de la feuille, `/-/deck.js`, les trois gabarits et `/-/templates/`, le guide des classes, les consignes MCP | 2 j |
| 4 | Console | L'écran House style et son aperçu, les types TypeScript | 1,5 j |
| 5 | Skill et plugin | La marketplace, le plugin, la skill, l'archive servie par chaque Cahut, la carte dans AIs and access | 1 j |
| 6 | Documentation et démonstration | Pages de la documentation, référence, site, CLAUDE.md, staging, bilan | 1 j |

Soit environ 8 jours. Les pull requests s'enchaînent dans l'ordre.

## Démonstration de fin de lot

**Prérequis** : la version du lot installée sur `staging.cahut.cloud` ; une marque fictive (un logo SVG, deux polices sous licence OFL, quatre couleurs) ; Claude Code avec le plugin `cahut`, branché par OAuth sur `staging.cahut.cloud`, en écriture et sur toutes les Rooms d'Olivier.

1. Demander à Claude de régler la house style du Cahut à partir du dossier de la marque. Il appelle `set_house_style`, sans Knock, et la console la montre.
2. Ouvrir `/-/templates/deck` : le gabarit de présentation à la marque. Ouvrir la page de la Key d'une Window : le logo et les couleurs.
3. Le brief : « Fais une présentation de six diapositives sur les résultats du trimestre et publie-la dans marketing. » Claude appelle `get_house_style`, puis `publish_window`, et rend l'adresse : la présentation s'ouvre aux couleurs de l'organisation et se mène au clavier.
4. Le même brief sur `dev.cahut.cloud`, resté neutre : la même présentation, dans le style neutre.
5. Changer la couleur principale dans la console : la présentation change sans être republiée. « Previous version » la rétablit.
6. Dans claude.ai, avec l'archive de la skill téléversée et le connecteur du Cahut, le même brief donne le même résultat.

## Hors du lot 6

- Des gabarits propres à chaque Cahut, une house style par Room, une variante sombre : après la V0.
- La house style dans le premier démarrage guidé : lot 7.
- Un mode orateur avec les notes, l'export PDF et PowerPoint : après la V0 ; l'impression d'une diapositive par page en tient lieu.
- Un moteur de graphiques servi par Cahut : l'IA dessine en SVG ou déclare sa bibliothèque.
- La house style dans les e-mails : avec l'envoi d'e-mails, après la V0.
- Le plugin dans le répertoire d'Anthropic : à l'ouverture du dépôt.

## Bilan

Le lot 6 s'est achevé le 23 octobre 2026, en six pull requests :

| # | Pull request | Contenu |
| --- | --- | --- |
| 1 | #63 | Le plan et l'ADR 22 |
| 2 | #65 | La house style de chaque Cahut, ses trois opérations, la feuille, les logos et les polices sous `/-/`, les pages de Cahut à la house style |
| 3 | #66 | Les classes de la feuille, le moteur de présentation, les trois gabarits et `/-/templates/`, le guide des classes, les consignes MCP |
| 4 | #67 | L'écran House style de la console |
| 5 | #68 | La skill pour Claude, son plugin, la marketplace du dépôt, l'archive servie par chaque Cahut |
| 6 | #PR6 | Documentation, site, staging, bilan |

**Écarts par rapport au plan.**

- Sans version pour fond sombre, `/-/house-style/logo-inverse` donne le nom de l'organisation en lettres claires plutôt que le logo, qui se lisait mal sur la couleur principale ; la sortie de la house style en avertit.
- `get_house_style` dit si les couleurs de graphiques sont celles de la house style (`own_chart_colors`), pour que la console laisse le champ vide tant que Cahut les dérive.
- Les règles des éléments sous chaque racine n'ont aucune spécificité (`:where`) : les classes du kit et le CSS d'une Window l'emportent toujours sur elles.
- Une présentation est un conteneur pour ses diapositives, ce qui empêche le navigateur de peindre le fond du `body` sur toute la page : la feuille le peint sur la racine, le temps d'une présentation.
- La page de la documentation sur le plugin s'appelle `connect/claude-plugin` : macOS ignore la casse, et Claude Code lisait un fichier `claude.md` comme des consignes. CLAUDE.md l'interdit désormais, et rappelle que les classes du kit forment un contrat.
- La référence de la documentation gagne une page, la feuille de la house style, que `make docs-reference` écrit depuis le guide que lisent les IA.
- Le panneau de navigateur de l'application bloque de lui-même les requêtes d'une page en bac à sable (`ERR_BLOCKED_BY_CLIENT`) : les Windows et les gabarits ont été vérifiés dans Chrome.
- Pas de nouvelle dépendance.

**Essais.**

- Les tests couvrent la house style neutre, les changements partiels, `neutral`, les couleurs dérivées et celles des graphiques, le contraste refusé et les avertissements, les polices (formats, poids, polices variables, fichiers gardés par leur nom, familles qui sortiraient du CSS), les logos (SVG, PNG, WebP, retrait, nom échappé), les deux versions et le retour arrière, les droits (membres, leurs IA, IA limitée à certaines Rooms, IA en lecture seule, IA d'un administrateur sans Knock), le Logbook sans contenu de fichier, la feuille et ses en-têtes, le cloisonnement entre deux Cahuts, les pages de Cahut, la politique d'une Window qui lie la feuille, les gabarits, le moteur, les consignes MCP, le guide qui nomme chaque classe, la marketplace, le manifeste du plugin et l'archive de la skill.
- `claude plugin validate --strict` passe sur la marketplace et sur le plugin.
- En local, dans Chrome, contre un Cahut de développement et une marque fictive (logo SVG, police de titres TrueType, quatre couleurs) : les pages de Cahut, une Window qui charge la feuille, le logo et la police dans son bac à sable par le cadre d'une Key, les trois gabarits à 1 280 px et sur un écran étroit, la présentation au clavier avec l'adresse qui suit chaque diapositive, la vue d'ensemble, le fond autour des diapositives. Dans le panneau de navigateur : l'écran House style, son aperçu qui suit le formulaire, l'enregistrement avec ses avertissements, le retour à la version précédente, l'écran à 375 px, la carte de la skill dans AIs and access.
- Sur le staging, la version du lot est installée (0.3.1-0.20261023105550-97b30eb61391), avec une sauvegarde dans `/var/lib/cahut.bak-2026-10-23-batch6`. La migration 11 est passée, et les deux Cahuts servent leur house style, le moteur, les gabarits et l'archive de la skill. Par MCP, avec un jeton personnel de démonstration d'Olivier, révoqué depuis, Claude a réglé la house style d'une marque fictive, Northbay, sans Knock, puis, sur le brief du plan, a lu le gabarit de présentation et publié [Third quarter results](https://staging.cahut.cloud/marketing/third-quarter-results/) dans `marketing`, en une demande. Le même brief sur `dev.cahut.cloud`, resté neutre, a donné [la même présentation](https://dev.cahut.cloud/oauth-test/third-quarter-results/) dans le style neutre. `/-/templates/deck` montre la marque sur le staging : titres en Georgia, logo pour fond sombre, couleur principale.

**Reste à faire.**

- La démonstration de fin de lot avec le compte d'Olivier : ouvrir la présentation de `marketing` et la mener au clavier ; changer la couleur principale dans la console, voir la présentation changer sans être republiée, puis « Go back to it ».
- Installer le plugin dans Claude Code (`/plugin marketplace add getcahut/cahut`, puis `/plugin install cahut@cahut`, avec le domaine `staging.cahut.cloud`), retirer le serveur `cahut-staging` ajouté à la main, et refaire le brief en une demande.
- Téléverser l'archive de la skill dans claude.ai, à côté du connecteur du staging, et refaire le brief.
- Firefox et Safari, toujours, depuis le lot 4.

**Suites.**

- `TestShutdownWithAnOpenStream` a échoué une fois en CI, sur la PR #68, et passé à la relance : un test instable, à regarder.
- Une icône et une adresse de politique de confidentialité pour le plugin, que demande le répertoire d'Anthropic, à l'ouverture du dépôt.
- La version du plugin, 0.4.0, suit celle de Cahut : à relever à chaque release.
