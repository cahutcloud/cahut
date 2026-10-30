# Cahut · Plan du lot 6 : skill Claude

23 octobre 2026 · Olivier Durand · Validé, achevé

## En bref

Le lot 6 donne à Claude une skill pour Cahut : les formats d'une Window, ce qui marche dans le bac à sable et ce qui n'y marche pas, la déclaration, les versions, l'accès et Knock first. La skill est la même pour chaque Cahut. Elle est empaquetée en plugin, `cahut`, dans une marketplace de ce dépôt, avec le serveur MCP de la Cahut dont l'utilisateur donne le domaine ; chaque Cahut la sert aussi en archive, pour les applications Claude où l'on téléverse les skills. Les consignes du serveur MCP restent pour les clients sans skill. Le lot compte trois pull requests, pour environ deux jours et demi à temps plein, et s'ouvre sur une décision écrite (ADR 22).

**Fin de lot**, reprise du [plan de la V0](plan-v0.md) : avec la skill, un brief donne une présentation publiée en une seule demande.

## Point de départ

Ce que les lots 1 à 5 ont laissé :

- Les fichiers d'une Window protégée sont servis dans un cadre, sous `/-/w/<capacité>/`, et `frame.js` passe le fragment de l'adresse au cadre.
- Les consignes du serveur MCP demandent des liens relatifs, jamais `/css/deck.css`, et une déclaration. Claude n'a pas de skill pour Cahut.
- `/-/` est réservé aux pages et aux adresses de Cahut ([ADR 9](adr/0009-rooms-and-windows-as-paths.md)) ; l'[ADR 15](adr/0015-license.md) met les skills sous Apache-2.0.
- La console a un écran AIs and access, avec l'adresse MCP à copier et la commande `claude mcp add` de Claude Code.

## Décision à écrire d'abord

L'[ADR 22](adr/0022-house-style-of-each-cahut.md) fixe :

- **La skill pour Claude**, la même pour chaque Cahut : les formats, l'accès, les liens relatifs, le bac à sable et Knock first.
- **Le plugin** `cahut`, dans une marketplace de ce dépôt : `.claude-plugin/marketplace.json` à sa racine, et le plugin dans `plugins/cahut/`, sous Apache-2.0. Il déclare le serveur MCP de la Cahut dont l'utilisateur donne le domaine à l'installation.
- **L'archive**, pour les applications Claude où l'on téléverse les skills : chaque Cahut la sert à `/-/cahut-skill.zip`, construite des mêmes fichiers, que le binaire embarque. La console y renvoie.
- **Les consignes du serveur MCP** atteignent tous les clients, et restent : la skill ajoute ce qu'elles ne peuvent pas porter, les formats en détail et quand prendre chacun.

## MCP, skill et plugin

- **Les consignes du serveur MCP**, pour les clients sans skill : les liens relatifs pour les fichiers de la Window, la déclaration, l'adresse à donner à la personne, Knock first.
- **La skill**, `plugins/cahut/skills/cahut/SKILL.md`, avec ses références : les formats (document, présentation, tableau de bord, site), ce qui marche dans le bac à sable et ce qui n'y marche pas (stockage, cookies, fenêtres), la déclaration, les images en base64 et la taille d'une publication, les versions et le retour arrière, l'accès et la Key, Knock first et le lien à transmettre.
- **Le plugin** `cahut`, dans `plugins/cahut/`, sous Apache-2.0 : son manifeste, la skill, et le serveur MCP `https://<domaine>/mcp`, dont l'utilisateur donne le domaine à l'installation (`userConfig`). Sa version suit celle de Cahut.
- **La marketplace**, `.claude-plugin/marketplace.json` à la racine du dépôt : `/plugin marketplace add cahutcloud/cahut`, puis `/plugin install cahut@cahut`. `claude plugin validate --strict` passe sur l'une et l'autre, lancé à la main : la CI n'installe pas Claude Code, et un test Go vérifie le manifeste, la marketplace et les chemins qu'ils nomment.
- **L'archive** de la skill, construite au démarrage depuis les fichiers que le binaire embarque, par un fichier Go dans `plugins/`, et servie par chaque Cahut à `/-/cahut-skill.zip`.

## Console

| Écran | Contenu |
| --- | --- |
| AIs and access | Une carte « Skill for Claude » : le plugin pour Claude Code, l'archive pour les applications Claude |

## Tests

- **Plugin** : la marketplace, le manifeste et son serveur MCP, les chemins qu'ils nomment, l'en-tête de la skill.
- **Archive** : chaque Cahut la sert, et la skill de l'archive est celle du plugin.
- **Navigateur** : dans le panneau de navigateur, la carte de la skill dans AIs and access.

## Découpage en pull requests

| # | Pull request | Contenu | Durée |
| --- | --- | --- | --- |
| 1 | Décision | Ce plan et l'ADR 22 | 0,5 j |
| 2 | Skill et plugin | La marketplace, le plugin, la skill, l'archive servie par chaque Cahut, la carte dans AIs and access | 1 j |
| 3 | Documentation et démonstration | Pages de la documentation, site, CLAUDE.md, staging, bilan | 1 j |

Soit environ 2,5 jours. Les pull requests s'enchaînent dans l'ordre.

## Démonstration de fin de lot

**Prérequis** : la version du lot installée sur `staging.cahut.cloud` ; Claude Code ; un compte claude.ai.

1. Dans Claude Code, `/plugin marketplace add cahutcloud/cahut`, puis `/plugin install cahut@cahut`, avec le domaine `staging.cahut.cloud`. Se brancher par OAuth, en écriture et sur toutes les Rooms d'Olivier.
2. Le brief : « Fais une présentation de six diapositives sur les résultats du trimestre et publie-la dans marketing. » Claude suit la skill, appelle `publish_window` et rend l'adresse : la présentation s'ouvre et se mène au clavier.
3. Dans la console, AIs and access : la carte de la skill donne les commandes de Claude Code et l'archive.
4. Dans claude.ai, avec l'archive de la skill téléversée et le connecteur de la Cahut, le même brief donne le même résultat.

## Hors du lot 6

- Le plugin dans le répertoire d'Anthropic : à l'ouverture du dépôt.

## Bilan

Le lot 6 s'est achevé le 23 octobre 2026, en trois pull requests :

| # | Pull request | Contenu |
| --- | --- | --- |
| 1 | #63 | Le plan et l'ADR 22 |
| 2 | #68 | La skill pour Claude, son plugin, la marketplace du dépôt, l'archive servie par chaque Cahut |
| 3 | #69 | Documentation, site, staging, bilan |

**Écarts par rapport au plan.**

- La page de la documentation sur le plugin s'appelle `connect/claude-plugin` : macOS ignore la casse, et Claude Code lisait un fichier `claude.md` comme des consignes. CLAUDE.md l'interdit désormais.
- Pas de nouvelle dépendance.

**Essais.**

- Les tests couvrent la marketplace, le manifeste du plugin et son serveur MCP, l'en-tête de la skill, et l'archive, qui contient les fichiers du plugin.
- `claude plugin validate --strict` passe sur la marketplace et sur le plugin.
- Dans le panneau de navigateur : la carte de la skill dans AIs and access.
- Sur le staging, la version du lot est installée (0.3.1-0.20261023105550-97b30eb61391), avec une sauvegarde dans `/var/lib/cahut.bak-2026-10-23-batch6`. Les deux Cahuts servent l'archive de la skill. Par MCP, avec un jeton personnel de démonstration d'Olivier, révoqué depuis, Claude a publié, sur le brief du plan, [Third quarter results](https://staging.cahut.cloud/marketing/third-quarter-results/) dans `marketing`, en une demande.

**Reste à faire.**

- La démonstration de fin de lot avec le compte d'Olivier : ouvrir la présentation de `marketing` et la mener au clavier.
- Installer le plugin dans Claude Code (`/plugin marketplace add cahutcloud/cahut`, puis `/plugin install cahut@cahut`, avec le domaine `staging.cahut.cloud`), retirer le serveur `cahut-staging` ajouté à la main, et refaire le brief en une demande.
- Téléverser l'archive de la skill dans claude.ai, à côté du connecteur du staging, et refaire le brief.
- Firefox et Safari, toujours, depuis le lot 4.

**Suites.**

- `TestShutdownWithAnOpenStream` a échoué une fois en CI, sur la PR #68, et passé à la relance. Réglé : le serveur termine bien le flux, mais le client laisse parfois une connexion ouverte sans y envoyer de requête, et `net/http` lui accorde 5 secondes avant de la fermer, plus que les 5 secondes qu'attendait le test. Le test attend maintenant la fin du flux, puis l'arrêt avant son échéance de 10 secondes.
- Une icône et une adresse de politique de confidentialité pour le plugin, que demande le répertoire d'Anthropic, à l'ouverture du dépôt.
- La version du plugin, 0.4.0, suit celle de Cahut : à relever à chaque release.
