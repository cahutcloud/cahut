# Cahut · Plan du lot 2 : comptes, Rooms et console

22 octobre 2026 · Olivier Durand · Validé

## En bref

Le lot 2 apporte les personnes à Cahut : elles se connectent, travaillent dans des Rooms d'équipe et retrouvent dans une console ce qui a été publié et ce que les IA ont fait. Il compte huit pull requests, pour environ deux semaines et demie à temps plein. Il s'ouvre sur deux décisions écrites (ADR 13 et 14) et se termine par une démonstration sur le serveur de staging.

**Fin de lot**, reprise du [plan de la V0](plan-v0.md) : deux Cahuts sur le même serveur ; dans l'une, deux comptes et une Room d'équipe ; chacun ne voit que ses Windows ; le Logbook se lit dans la console.

## Point de départ

Ce que le lot 1 a laissé :

- Les tables `users`, `rooms` et `room_members` existent, avec les rôles `admin` et `member` d'une Cahut, et `owner` et `member` d'une Room. Chaque personne a sa Room personnelle. On ne peut pas encore créer de Room d'équipe, et personne n'a de mot de passe.
- Le registre connaît trois acteurs : l'administrateur en ligne de commande, la personne (prévue pour la console, jamais utilisée) et l'IA. Le drapeau `AdminOnly` réserve une opération à la ligne de commande et la cache aux IA.
- `publish_window`, `get_window` et `list_windows` limitent déjà une IA aux Rooms de sa personne.
- Les Windows restent lisibles par quiconque a leur adresse : les accès arrivent au lot 4.
- Le serveur écoute sur les ports 80 et 443, et `install.sh` ouvre ces deux ports dans firewalld ou ufw.

## Décisions à écrire d'abord

Deux ADR ouvrent le lot, en anglais.

### ADR 13 · People, sessions and permissions

Écrit : [ADR 13](adr/0013-people-sessions-and-rights.md).

- **Deux niveaux d'administration** ([ADR 8](adr/0008-several-cahuts-per-server.md)). Les administrateurs du serveur ont un compte dans `server.db`, distinct des personnes des Cahuts : ils déclarent les Cahuts et lisent le Logbook du serveur ; dans chaque Cahut, ils gèrent les personnes et les Rooms. Les administrateurs d'une Cahut sont des personnes de cette Cahut, avec le rôle `admin`.
- **Mots de passe** hachés en argon2id (RFC 9106) avec les paramètres de l'OWASP : 19 Mio, 2 passes, 1 fil, un sel de 16 octets, stockés dans une chaîne au format PHC. Un sémaphore limite le nombre de calculs simultanés, pour qu'une rafale de connexions n'épuise pas la mémoire d'un petit VPS. Un mot de passe compte 15 caractères au moins, comme le demande le NIST quand il n'y a pas de second facteur, sans règle de composition.
- **Pas d'e-mail.** L'administrateur crée le compte et transmet un lien d'activation à usage unique, valable 7 jours, où la personne choisit son mot de passe. Le même mécanisme sert après un oubli. La ligne de commande reste la voie de secours de tous les comptes.
- **Sessions.** Des jetons opaques tirés de `crypto/rand`, préfixés `cahut_ses_` et gardés hachés en SHA-256, comme les jetons personnels. Une session expire après 7 jours sans usage, et au plus tard au bout de 30 jours. Changer son mot de passe ferme ses autres sessions ; un compte désactivé perd toutes les siennes.
- **Une session vaut pour un seul domaine.** Celle d'une personne vit dans la base de sa Cahut. Celle d'un administrateur du serveur vit dans `server.db`, liée au domaine où elle a été ouverte. Pour agir dans une autre Cahut, l'administrateur du serveur se connecte à la console de cette Cahut : chaque requête ne touche que la Cahut de son domaine.
- **Connexion protégée.** Elle est limitée en débit, par adresse et par identifiant, et un identifiant inconnu coûte le même calcul qu'un identifiant connu.
- **Droits appliqués par le registre :**
  - `AdminOnly` cède la place à un niveau requis pour chaque opération : toute personne, administrateur de la Cahut ou administrateur du serveur. Le registre relit le rôle et l'état de la personne dans la transaction de l'appel : un compte désactivé perd la main à l'instant, et ses IA aussi.
  - Les opérations vérifient elles-mêmes les droits sur une Room (propriétaire ou membre), avec un assistant commun.
  - Les opérations d'identification (se connecter, activer un compte, changer de mot de passe, créer un jeton) ne sont pas offertes aux IA : elles ne donnent aucune capacité sur le contenu, elles servent à prouver qui l'on est. Toutes les autres opérations sont offertes aux IA, dans la limite des droits de leur personne. L'IA d'un membre ne voit pas les outils réservés aux administrateurs.
  - `Irreversible` devient `KnockFirst`, parce que la liste du lot 5 compte des opérations qui s'annulent mais changent des droits. L'ADR 13 précise ainsi la classe réversible ou irréversible de l'[ADR 1](adr/0001-operation-registry.md). Une IA qui appelle une opération ainsi marquée est refusée jusqu'au lot 5, avec un message qui demande à sa personne d'agir dans la console.
- **Au Logbook**, un administrateur du serveur apparaît comme `admin` avec son identifiant ; la ligne de commande apparaît comme `admin` sans identifiant. Une connexion ou une activation s'inscrit au nom de la personne concernée.

### ADR 14 · The HTTP API of the console

Écrit : [ADR 14](adr/0014-console-http-api.md).

- `POST https://<domain>:8443/api/<operation>`, avec une entrée et une sortie en JSON, les mêmes noms et les mêmes schémas que les outils MCP. Les opérations du serveur sont sous `/api/server/<operation>`, pour ses seuls administrateurs.
- Le jeton de session passe dans `Authorization: Bearer`. Aucun cookie, aucun CORS ([ADR 11](adr/0011-console-on-its-own-port.md)).
- Les erreurs : 400 pour `ErrInvalid`, 401 sans session valide, 403 pour `ErrForbidden`, 404 pour `ErrNotFound`, 409 quand une approbation est requise, 500 sinon. Le corps de la réponse est `{"error": "..."}`. Comme sur MCP, le message d'une erreur interne n'est pas montré.
- Le corps d'une requête est limité à 1 Mio : la console ne publie pas, ce sont les IA qui publient les Windows.
- Les types TypeScript de la console sont générés à partir des types Go du registre, comme la référence de la documentation, et la CI vérifie qu'ils sont à jour. La console ne peut ainsi appeler que des opérations qui existent.

## Modèle de données

Deux migrations, `cahut/0002` et `server/0002` :

| Base | Changement |
| --- | --- |
| Cahut | `users` : `password_hash`, `status` (`pending` jusqu'à l'activation, puis `active` ou `disabled`) |
| Cahut | `activations` : lien haché, personne, auteur, expiration, date d'usage |
| Cahut | `sessions` : jeton haché, personne, création, dernier usage, expiration, révocation |
| Cahut | `logbook` : colonne `room_id`, sans clé étrangère, pour que chaque membre lise ce qui touche ses Rooms |
| Serveur | `admins`, `admin_activations`, `admin_sessions` (cette dernière avec le domaine de la session) |

Les personnes créées au lot 1 passent en `pending` : `cahut user link` leur donne un lien d'activation.

## Opérations

Colonne « IA » : *oui* si l'opération est offerte aux IA, *non* si elle ne l'est pas, *Knock* si elle est offerte mais refusée aux IA jusqu'au lot 5.

### Dans chaque Cahut

| Opération | Qui l'appelle | IA | Ce qu'elle fait |
| --- | --- | --- | --- |
| `sign_in` | personne non connectée | non | Identifiant et mot de passe, rend une session |
| `activate_account` | personne non connectée | non | Lien d'activation et nouveau mot de passe, rend une session |
| `sign_out` | personne | non | Ferme la session |
| `change_password` | personne | non | Demande l'ancien mot de passe, ferme les autres sessions |
| `get_account` | personne | oui | Qui je suis : nom, rôle, Rooms |
| `list_users` | personne | oui | Les personnes de la Cahut : identifiant, nom, rôle, état |
| `create_user` | admin de la Cahut | Knock | Existe déjà ; rend désormais un lien d'activation |
| `create_activation_link` | admin de la Cahut | non | Un lien neuf après un oubli ; l'ancien mot de passe reste valable jusqu'à son usage |
| `set_user_role` | admin de la Cahut | Knock | Administrateur ou membre ; la Cahut garde au moins un administrateur |
| `disable_user`, `enable_user` | admin de la Cahut | Knock | Désactiver ferme les sessions et suspend les jetons. On ne supprime aucune personne en V0 |
| `create_token`, `list_tokens`, `revoke_token` | une personne pour elle-même, un admin pour tous | non | Existent déjà ; s'ouvrent à la console |
| `create_room` | personne | oui | Crée une Room d'équipe, dont l'auteur devient propriétaire |
| `list_rooms` | personne | oui | Ses Rooms ; pour un admin, toutes les Rooms avec leurs membres, sans leurs Windows |
| `get_room` | membre, admin | oui | Titre, membres et leurs rôles, nombre de Windows |
| `update_room` | propriétaire, admin | oui | Change le titre |
| `add_room_member`, `set_room_member_role`, `remove_room_member` | propriétaire, admin ; chacun peut quitter une Room | Knock | Une Room garde au moins un propriétaire |
| `publish_window` | membre de la Room | oui | Existe déjà ; vérifie les droits de Room avec l'assistant commun |
| `list_windows` | personne | oui | Existe déjà ; gagne un filtre par Room |
| `list_logbook` | personne | oui | Tout pour un admin ; pour un membre, ses actions, celles de ses IA et ce qui touche ses Rooms ; pagination |

### Sur le serveur

Ces opérations ne passent pas par MCP.

| Opération | Qui l'appelle | Ce qu'elle fait |
| --- | --- | --- |
| `sign_in`, `activate_account`, `sign_out`, `change_password` | administrateur du serveur | Comme dans une Cahut |
| `create_server_admin` | administrateur du serveur, ligne de commande | Rend un lien d'activation |
| `add_domain` | administrateur du serveur, ligne de commande | Existe déjà ; peut aussi créer le premier administrateur de la Cahut et rendre son lien |
| `list_domains`, `list_logbook` | administrateur du serveur, ligne de commande | Existent déjà |

## Console

### Socle

- Le code vit dans `internal/console/` : l'application Preact en TypeScript dans `ui/`, construite par Vite dans `ui/dist` et embarquée dans le binaire par `go:embed`. Un fichier témoin versionné permet encore un `go build` sans Node.js ; le binaire sert alors une courte page qui explique comment construire la console ([ADR 5](adr/0005-preact-console-embedded.md)).
- Les paquets npm : `preact`, `vite`, `@preact/preset-vite` et `typescript`, plus `preact-iso` si les routes le demandent. Versions exactes, sans scripts d'installation, avec un `.npmrc` comme celui de `sites/`.
- La marque arrive par l'alias `@charte`, comme dans les sites : `cahut-tokens.css`, `cahut-ui.css`, les polices hébergées par Cahut, les thèmes clair et sombre.
- Une CSP stricte, sans script en ligne : `default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self' data:; connect-src 'self'; frame-ancestors 'none'; base-uri 'none'; form-action 'none'`.
- Le jeton de session est gardé dans le `localStorage` de l'origine du port 8443.

### Écrans

Ils suivent la maquette [`ui/console.html`](charte-graphique/ui/console.html) de la charte.

| Écran | Contenu |
| --- | --- |
| Connexion et activation | Identifiant et mot de passe ; choix du mot de passe depuis un lien ; un accès séparé pour les administrateurs du serveur |
| Coquille | Barre latérale : Overview, Logbook, Rooms, This Cahut (People, AIs and access), Server. Menu du compte : mot de passe, déconnexion. Knock first et House style n'apparaissent qu'avec leurs lots |
| Overview | Les dernières Windows de ses Rooms |
| Room | Les Windows en cartes, avec un pictogramme par type, qui s'ouvrent sur le port 443 ; les membres ; leur gestion pour les propriétaires ; la création d'une Room d'équipe |
| Logbook | Qui, par quel accès, sur quoi et avec quel résultat ; filtres par personne et par opération |
| People | Pour les admins : liste des personnes, création avec un lien à copier, rôle, désactivation, nouveau lien |
| AIs and access | Ses jetons personnels : création avec la commande `claude mcp add` prête à copier, liste, révocation. L'écran des IA connectées par OAuth s'y ajoute au lot 3 |
| Server | Pour les administrateurs du serveur : les Cahuts, la déclaration d'une Cahut avec son premier administrateur, le Logbook du serveur |

Il n'y a ni vignettes ni badges d'accès sur les cartes : les vignettes viendront plus tard, et les accès au lot 4.

## Serveur web, ligne de commande et installation

- Un troisième écouteur, sur `:8443`, avec la même configuration TLS que le port 443 et le même routage par domaine. Le port se règle par `--console` et `CAHUT_CONSOLE_PORT`, que lisent aussi les commandes qui affichent un lien.
- Sur ce port, `/` et ses routes servent la console (avec un repli sur `index.html`), et `/api/` sert l'API. En développement, `make dev` sert aussi la console sur `localhost:8081`, et `make console-dev` lance Vite avec un proxy vers l'API, pour le rechargement à chaud.
- Sur le port 443, la page d'accueil d'une Cahut renvoie vers sa console. Les pages de Room et la connexion des lecteurs restent au lot 4.
- Ligne de commande : `cahut user create` affiche le lien d'activation ; `cahut user link` en produit un nouveau ; `cahut admin create` crée un administrateur du serveur.
- `install.sh` ouvre aussi le port 8443 dans firewalld ou ufw, et rappelle de l'ouvrir dans le pare-feu de l'hébergeur. Le lot 7 complète l'installation.

## Construction et CI

- `make console` construit la console ; `make build` et `make snapshot` la construisent d'abord. `make test` et `make lint` se passent toujours de Node.js.
- CI : un job Console installe les paquets, vérifie les types et construit la console. Le job Build embarque la console construite. GoReleaser la construit dans un hook `before`, et le workflow de release installe Node.js.
- Dependabot suit aussi `internal/console/ui`.
- Une seule nouvelle dépendance Go, `golang.org/x/crypto` pour argon2id : elle est déjà présente comme dépendance indirecte.
- Dans CLAUDE.md : les nouvelles tâches, l'arborescence, et une version plus nuancée de la phrase « The Go tasks never need Node.js ».

## Tests

- **Matrice des droits** : chaque opération appelée par un administrateur du serveur, un administrateur de la Cahut, un propriétaire, un membre, une personne extérieure à la Room, une IA et un compte désactivé.
- **Cloisonnement** : la session d'une Cahut est refusée par une autre ; celle d'un administrateur du serveur est refusée hors de son domaine.
- **Comptes** : lien d'activation à usage unique et lien expiré, fermeture des sessions au changement de mot de passe, limitation de débit, temps de réponse identique pour un identifiant inconnu.
- **API** : codes d'erreur, en-têtes (CSP, absence de CORS, `Cache-Control: no-store`), corps trop gros.
- **Console** : vérification des types et build en CI. Le parcours de la démonstration est rejoué dans le navigateur avant chaque pull request de la console. Le lot 2 n'introduit pas de framework de test côté navigateur.

## Découpage en pull requests

| # | Pull request | Contenu | Durée |
| --- | --- | --- | --- |
| 1 | Décisions | ADR 13 et 14 | 0,5 j |
| 2 | Droits et comptes | Niveaux des opérations, `KnockFirst`, mots de passe, activation, sessions, opérations des personnes, `cahut user link` | 2,5 j |
| 3 | Administrateurs du serveur | Comptes et sessions dans `server.db`, `cahut admin create`, premier administrateur à la déclaration d'une Cahut | 1 j |
| 4 | Rooms d'équipe | Opérations des Rooms, droits de Room dans `publish_window`, `room_id` au Logbook, lecture du Logbook par les membres | 1,5 j |
| 5 | API HTTP | Écouteur 8443, adaptateur, génération des types TypeScript, port 8443 dans `install.sh` | 1,5 j |
| 6 | Socle de la console | Projet Vite, embarquement, page de repli, CSP, connexion, activation, coquille, Makefile, CI, release, Dependabot | 1,5 j |
| 7 | Écrans de la console | Overview, Rooms, Logbook, People, AIs and access, Server, compte | 3 j |
| 8 | Documentation et démonstration | CLAUDE.md, README, pages de la documentation (console, personnes, Rooms), référence, démonstration sur le staging | 1 j |

Soit environ 12,5 jours. Les pull requests 3 et 4 dépendent de la 2 et peuvent avancer en parallèle ; les suivantes s'enchaînent dans l'ordre.

## Démonstration de fin de lot

**Prérequis** : le second domaine, `beta.staging.cahut.cloud`, pointé vers le serveur de staging, et le port 8443 ouvert dans le pare-feu de l'hébergeur.

1. Installer la release du lot sur `staging.cahut.cloud`. `cahut admin create` donne un lien, avec lequel l'administrateur du serveur active son compte dans la console.
2. Dans la console, déclarer la seconde Cahut avec son premier administrateur, puis ouvrir le lien de celui-ci dans un autre navigateur.
3. Dans la première Cahut, l'administrateur crée deux personnes, Alice et Bruno, et leur transmet leurs liens.
4. Alice crée la Room d'équipe `marketing` et y ajoute Bruno.
5. Chacun crée un jeton personnel dans la console et branche Claude Code. Alice publie une Window dans `marketing` et une dans sa Room personnelle ; Bruno publie dans la sienne.
6. Dans la console comme pour son IA, Bruno voit la Window de `marketing` mais pas la Room personnelle d'Alice. La console de la seconde Cahut ne montre rien de la première.
7. Le Logbook de la console montre les publications, les créations de comptes et la Room, avec, pour chaque action, l'IA et la personne pour qui elle agit.
8. L'IA de Bruno essaie d'ajouter quelqu'un à `marketing`. Elle est refusée, avec un message qui renvoie à la console.

## Hors du lot 2

- OAuth et l'écran de consentement : lot 3.
- Les accès des Windows, la connexion des lecteurs sur le port 443 et les pages de Room : lot 4. D'ici là, une Window reste lisible par quiconque a son adresse ; « chacun ne voit que ses Windows » vaut dans la console et pour les IA.
- Knock first : lot 5. Les opérations marquées `KnockFirst` sont refusées aux IA en attendant.
- La suppression d'une personne ou d'une Room : lot 5, par Knock first. Au lot 2, on désactive une personne.
- Le second facteur, le SSO et les e-mails : après la V0.
- Le premier démarrage guidé : lot 7. Il reprendra `cahut admin create` et l'activation.

## Questions tranchées

Le 22 octobre 2026, les recommandations ont été retenues :

1. **Qui crée une Room d'équipe ?** Toute personne, qui en devient propriétaire. Les administrateurs gèrent toutes les Rooms.
2. **Un administrateur de la Cahut lit-il les Windows des Rooms dont il n'est pas membre ?** Non. Il voit les Rooms, leurs membres et tout le Logbook, mais pas le contenu. Il en va de même pour les administrateurs du serveur.
3. **Que lit un membre dans le Logbook ?** Ses actions, celles de ses IA et tout ce qui touche ses Rooms.
4. **Des comptes à part pour le serveur.** Une personne qui installe Cahut seule aura deux comptes : administrateur du serveur et personne de sa Cahut. On l'accepte en V0 ; le premier démarrage du lot 7 pourra créer les deux d'un seul geste.
5. **Durées et mot de passe.** Une session dure 7 jours sans usage et 30 jours au plus, un lien d'activation 7 jours, et un mot de passe compte 15 caractères au moins.
6. **Ajouter ou retirer un membre d'une Room passe-t-il par Knock first ?** Oui, puisque cela change les droits d'une personne. La liste du lot 5 s'allonge d'autant.
7. **Quel second domaine pour la démonstration ?** `beta.staging.cahut.cloud`, avec un enregistrement DNS qui pointe vers le serveur de staging.

Les ADR 13 et 14 reprennent ces réponses.

## Bilan

Le lot 2 s'est achevé le 22 octobre 2026, en huit pull requests :

| # | Pull request | Contenu |
| --- | --- | --- |
| 1 | #12 | Le plan et les ADR 13 et 14 |
| 2 | #13 | Droits dans le registre, mots de passe, activation, sessions |
| 3 | #14 | Administrateurs du serveur, premier administrateur d'une Cahut |
| 4 | #15 | Rooms d'équipe, droits de Room, Room de chaque entrée du Logbook |
| 5 | #17 | API HTTP sur le port 8443, types TypeScript générés |
| 6 | #19 | Socle de la console : connexion, activation, coquille |
| 7 | #26 | Écrans de la console |
| 8 | #27 | Documentation, scénario de démonstration, staging |

**Écarts par rapport au plan.**

- Une opération de plus, `get_cahut` : la console, sur un autre port, ne peut pas deviner l'adresse MCP de la Cahut qu'elle affiche dans la commande `claude mcp add`.
- Un paquet npm de plus, `lucide-preact`, pour les icônes Lucide de la charte.
- La console n'est embarquée qu'avec l'étiquette de compilation `console`. `make build`, `make dev` et GoReleaser la construisent ; un `go build` simple, les tests et le lint se passent de Node.js et servent une page d'avis.
- Le port de la console reste fixe (8443, et 8081 en développement). Le rendre réglable demande que la ligne de commande connaisse ce réglage pour écrire ses liens : à voir au lot 7.
- L'acteur d'un appel porte le domaine où il arrive, et le Logbook du serveur reçoit aussi la colonne `room_id`, pour garder la même structure que celui d'une Cahut.
- Un défaut corrigé en route : un écouteur qui ne pouvait pas démarrer, faute de port libre, laissait le serveur tourner sans lui. Le serveur s'arrête désormais et dit pourquoi.

**Démonstration.**

- Les huit étapes ont été rejouées de bout en bout en local, par l'API et MCP, puis dans un test Go, `TestBatch2Scenario`, qui les rejoue à chaque CI.
- Sur le staging, la version du lot est installée (0.1.2-dev). La console répond sur le port 8443 de `staging.cahut.cloud` et de `dev.cahut.cloud`, qui sert de second domaine à la place de `beta.staging.cahut.cloud`. Les données du lot 1 ont migré, et le jeton existant de Claude Code marche toujours, avec les outils d'un administrateur. Une sauvegarde est restée dans `/var/lib/cahut.bak-2026-10-22-batch2`.
- Sur le staging, aucun compte de démonstration (Alice, Bruno) n'a été créé : Olivier a reçu ses liens d'activation, pour son compte et pour un administrateur du serveur.

**Suites.**

- Désactiver ou retirer un administrateur du serveur : rien ne le permet encore.
- Pour le lot 5, les opérations marquées `KnockFirst` : `create_user`, `set_user_role`, `disable_user`, `enable_user`, `add_room_member`, `set_room_member_role`, `remove_room_member`.
- Un bouton de thème dans la console, qui suit pour l'instant le système.
- Aucune release ne porte encore le lot 2 : la prochaine le fera.
