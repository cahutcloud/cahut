# Cahut · Plan du lot 4 : Windows complètes, versions et accès

22 octobre 2026 · Olivier Durand · Validé

## En bref

Le lot 4 fait d'une Window une page qu'on gère : l'IA la met à jour sans perdre la version précédente, on y revient d'un geste, on la met à la corbeille, et chaque Window s'ouvre à qui on choisit : sa Room, tout le Cahut, les détenteurs d'une Key ou le public, jusqu'à une date si l'on veut. L'IA déclare ce que fait la Window (scripts, adresses qu'elle appelle, formulaires), et Cahut en tire la politique CSP qui l'enferme. Le lot compte six pull requests, pour environ sept jours et demi à temps plein. Il s'ouvre sur une décision écrite (ADR 19).

**Fin de lot**, reprise du [plan de la V0](plan-v0.md) : une personne non technique partage une présentation avec une Key depuis la console, le lien s'ouvre en navigation privée, puis on revient à la version précédente.

## Point de départ

Ce que les lots 1 à 3 ont laissé :

- `publish_window` crée une Window, à jamais dans sa première version. `get_window` et `list_windows` la décrivent. Les fichiers vivent sur le disque sous `windows/<id>/<version>/`, et la table `windows` porte déjà un numéro de version.
- N'importe qui ayant l'adresse d'une Window la lit : chaque fichier est servi à son adresse, dans le bac à sable CSP de l'[ADR 10](adr/0010-windows-in-a-csp-sandbox.md), sans contrôle d'accès. L'adresse d'une Room répond 404 et celle du Cahut, une ligne de texte.
- Les personnes se connectent à la console, sur le port 8443, avec un jeton de session gardé dans le `localStorage` de son origine. Rien ne connaît encore le lecteur sur le port 443.
- L'ADR 10 a fixé le principe des Windows protégées (une page de Cahut qui encadre la Window dans une iframe en bac à sable, par une adresse à capacité sous `/-/`) et l'[ADR 11](adr/0011-console-on-its-own-port.md) celui de la connexion du lecteur (la console remet un code à usage unique à `/-/session`, qui pose le cookie du lecteur). Le lot 4 les construit.

## Décision à écrire d'abord

### ADR 19 · Windows keep two versions, open to an access, and declare what they call

L'ADR 19 fixe :

- **Versions.** `update_window` remplace tous les fichiers d'une Window : la nouvelle version devient la version courante, celle qu'elle remplace devient la précédente, et la plus ancienne s'efface. `rollback_window` échange la courante et la précédente : un retour en arrière s'annule par un autre. Les numéros ne reviennent jamais.
- **Corbeille.** `trash_window` retire une Window : elle ne s'ouvre plus et quitte les listes, son nom reste pris. `restore_window` la rend telle qu'elle était. La purge viendra avec Knock first, au lot 5.
- **Accès.** Room seulement (par défaut), interne, Key, public, avec une date de fin facultative, après laquelle la Window ne s'ouvre plus qu'à sa Room. Les membres de la Room le règlent. Une IA règle tout sauf l'accès public, qui attend Knock first.
- **Keys.** Choisie (10 à 256 caractères) ou tirée par Cahut, hachée en Argon2id, montrée une fois. Le lecteur la tape sur la page de la Window ; Cahut lui pose un laissez-passer, un cookie propre à la Window, valable 30 jours au plus.
- **Lecteurs.** Un lecteur se connecte sur le port 443 en passant par la console : un code de 60 secondes, puis un cookie de session de lecteur, aux mêmes durées que les sessions de la console.
- **Service.** Une Window publique est servie à son adresse. Une Window protégée l'est par une page de Cahut qui l'encadre ; ses fichiers viennent de `/-/w/<capacité>/`, une capacité signée par HMAC avec un secret du Cahut, liée à la Window et à la session ou au laissez-passer du lecteur, valable quatre heures. Chaque fichier revérifie l'accès.
- **Déclaration.** `publish_window` et `update_window` exigent une déclaration : scripts ou non, origines que les pages chargent ou appellent, origines où leurs formulaires envoient. Cahut en tire la CSP de chaque fichier. Les Windows publiées avant le lot 4 restent « non déclarées », servies avec le seul bac à sable.

## Modèle de données

Trois migrations, une par pull request :

| Migration | Contenu |
| --- | --- |
| `cahut/0006_versions` | Table `window_versions` (Window, numéro, taille, nombre de fichiers, qui l'a publiée, quand), remplie depuis `windows` ; colonnes `previous` et `trashed_at` de `windows` |
| `cahut/0007_access` | Colonnes `access`, `access_ends_at` et `key_hash` de `windows` ; table `window_passes` (laissez-passer des Keys, hachés) ; colonne `reader` de `sessions` ; table `reader_codes` ; table `secrets` (le secret des capacités) |
| `cahut/0008_declarations` | Colonne `declaration` de `window_versions`, en JSON, vide pour les versions publiées avant |

Les Windows publiées avant le lot 4 s'ouvrent à leur Room seulement après la migration : une personne les rouvre dans la console.

## Opérations

| Opération | Qui l'appelle | Ce qu'elle fait |
| --- | --- | --- |
| `publish_window` | membre de la Room | Exige désormais une déclaration |
| `update_window` | membre de la Room | Publie une nouvelle version, garde la précédente ; peut changer le titre |
| `rollback_window` | membre de la Room | Revient à la version précédente ; un second appel y retourne |
| `set_window_access` | membre de la Room ; une IA, sauf l'accès public | Règle l'accès, la date de fin et la Key ; rend la Key une seule fois |
| `trash_window`, `restore_window` | membre de la Room | Met à la corbeille, en sort |
| `get_window` | membre de la Room | Ajoute l'accès, les deux versions et la déclaration |
| `list_windows` | personne | Ajoute l'accès ; `trashed` liste la corbeille |
| `sign_in_to_windows` | personne, dans la console | Rend le code qui ouvre une session de lecteur sur le port 443 |

## Serveur web

Sur le port 443, en plus de MCP et d'OAuth :

| Adresse | Ce qu'elle sert |
| --- | --- |
| `/` | La page du Cahut : ses Rooms pour un lecteur connecté, un lien pour se connecter sinon |
| `/<room>/` | Les Windows de la Room que le lecteur peut ouvrir ; 404 s'il n'en ouvre aucune et n'est pas membre |
| `/<room>/<window>/…` | Une Window publique ; la page qui encadre une Window protégée ; la page de sa Key ; ou 404 |
| `/-/w/<capacité>/…` | Les fichiers d'une Window protégée, dans le bac à sable |
| `/-/key` | Reçoit une Key (POST), pose le laissez-passer |
| `/-/sign-in` | Envoie le lecteur à la console, qui le connecte |
| `/-/session` | Échange le code de la console contre le cookie du lecteur |
| `/-/sign-out` | Ferme la session du lecteur (POST) |
| `/-/cahut.css`, `/-/frame.js` | Le style des pages de Cahut, et le script qui passe le fragment de l'adresse à la Window encadrée |

Les pages de Cahut sur le port 443 sont rendues côté serveur, sans script en ligne, sous une CSP stricte avec `frame-ancestors 'none'`. Elles portent la house style du Cahut, neutre jusqu'au lot 6.

## Console

| Écran | Contenu |
| --- | --- |
| Fiche d'une Window, `/rooms/<room>/<window>` | Titre, adresse, sorte ; accès (niveau, date de fin, Key à copier une fois créée) ; versions (courante et précédente, retour en arrière) ; déclaration ; corbeille |
| Room | Chaque Window mène à sa fiche, avec son accès en badge ; la corbeille de la Room |
| Connexion du lecteur, `/read` | Se connecter si besoin, puis renvoyer le navigateur vers la Window demandée |

## Tests

- **Versions** : mise à jour, retour en arrière et retour au retour, numéros, fichiers effacés sur le disque, échec qui ne laisse rien.
- **Accès** : la matrice des appelants (membre, propriétaire, personne hors de la Room, administrateur, IA limitée, IA en lecture seule) ; l'IA refusée pour l'accès public ; la date de fin ; la Key jamais au Logbook.
- **Service** : chaque niveau d'accès pour un inconnu, une personne du Cahut, un membre, un détenteur de la Key ; capacité falsifiée, expirée, d'une autre Window, d'un autre Cahut ; laissez-passer après changement de Key ; session de lecteur après désactivation ; Window à la corbeille ; Room et racine ; limite de débit des Keys ; adresse de retour étrangère refusée.
- **CSP** : la politique tirée de chaque déclaration, et celle des Windows non déclarées.
- **Navigateurs** : une présentation encadrée, en plein écran, sous Chrome, Firefox et Safari.

## Découpage en pull requests

| # | Pull request | Contenu | Durée |
| --- | --- | --- | --- |
| 1 | Décision | Ce plan et l'ADR 19 | 0,5 j |
| 2 | Versions et corbeille | Migration 0006, `update_window`, `rollback_window`, `trash_window`, `restore_window`, `get_window` et `list_windows` enrichies | 1 j |
| 3 | Accès | Migration 0007, `set_window_access`, Keys, sessions des lecteurs, `sign_in_to_windows`, pages de Cahut, capacités, page `/read` de la console | 2,5 j |
| 4 | Déclaration | Migration 0008, déclaration exigée, CSP tirée de la déclaration, consignes MCP | 1 j |
| 5 | Console | Fiche d'une Window, accès, versions, déclaration, corbeille | 1,5 j |
| 6 | Documentation et démonstration | Pages de la documentation, CLAUDE.md, essais sous Chrome, Firefox et Safari, staging, bilan | 1 j |

Soit environ 7,5 jours. Les pull requests s'enchaînent dans l'ordre.

## Démonstration de fin de lot

**Prérequis** : la version du lot installée sur `staging.cahut.cloud`.

1. Une IA publie une présentation dans la Room `marketing`, puis la met à jour : la version 2 est servie.
2. Dans la console, sur la fiche de la Window, choisir « Anyone with its Key », une date de fin, et copier le lien et la Key.
3. Ouvrir le lien en navigation privée : la page de la Key s'affiche ; la Key ouvre la présentation, en plein écran.
4. Sur la fiche, revenir à la version précédente : la version 1 est servie au même lien.
5. Le Logbook montre la publication, la mise à jour, l'accès (sans la Key) et le retour en arrière.

## Hors du lot 4

- Knock first : lot 5. D'ici là, une IA ne rend pas une Window publique, et la purge d'une Window n'existe pas.
- La house style des pages de Cahut : lot 6. Le lot 4 les sert dans un style neutre.
- Renommer ou déplacer une Window : après la V0.
- Plus de deux versions, l'historique complet : après la V0.
