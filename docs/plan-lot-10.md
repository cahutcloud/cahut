# Cahut · Plan du lot 10 : les données des Windows

31 octobre 2026 · Olivier Durand · Validé

## En bref

Le lot 10 donne à chaque Window une base SQLite à elle, comme le décide l'[ADR 56](adr/0056-a-database-for-each-window.md) (#228). Les pages d'une Window y lisent ce que sa déclaration rend lisible, et y écrivent par les requêtes qu'elle déclare. Les membres de sa Room et leurs IA lisent et changent tout, après une copie qu'un appel suffit à rétablir. Le SQL des Windows tourne dans des processus à part.

Sondages, formulaires de contact, pétitions, listes d'inscription et suivis d'équipe deviennent des Windows, sans service extérieur.

Le lot compte sept pull requests. Sa fin fait **Cahut 1.0**, publiée aussitôt. Pour le Show HN du 17 novembre, la version est gelée 48 heures avant : aucune release n'est publiée ni installée en production à partir du 15 novembre, quel que soit son numéro.

**Fin de lot**, sur le staging :

- Claude publie une pétition et la rend publique, après un Knock qui dit ce qu'elle garde. En navigation privée, un visiteur la signe et voit la liste des noms. Avec curl, personne ne lit les e-mails.
- Claude lit les signatures, en supprime une, puis `rollback_window_data` la rend.
- Une écriture de Claude qui supprime une colonne frappe d'abord.
- Une requête sans fin, envoyée avec curl, s'arrête au bout d'une seconde, et les autres Windows répondent pendant ce temps.

## Point de départ

- **Le bac à sable.** Les pages ont une origine opaque, et le stockage du navigateur y lève une erreur ([ADR 10](adr/0010-windows-in-a-csp-sandbox.md)). Leur CSP permet déjà `connect-src` vers l'origine de la Cahut (`Declaration.Policy`), mais rien n'y répond.
- **Les fichiers.** Ils vivent dans `windows/<id>/<version>/`. `checkPath` refuse tout chemin dont une partie commence par un point, ce qui laisse `.data/` libre.
- **Les frames.** Elles passent par `/-/frame/<capacité>/`, avec cinq sortes de capacité (`s`, `p`, `c`, `m` et `h`), et durent 4 heures. Une Window publique est servie à son adresse, sans frame. Un laissez-passer dure 30 jours (`PassTTL`).
- **Le registre.**
  - Une opération qui frappe pour certains appels (`KnocksWhen`) n'a pas de `Prepare`. `Call.Knock` peut s'arrêter tôt.
  - Une opération `ReadOnly` lit dans le pool de lecture, et le Logbook n'enregistre pas les appels d'un visiteur.
  - `publish_window` et `update_window` écrivent leurs fichiers dans `Run`, et `OnFailure` les efface si l'appel échoue.
- **Le pilote modernc.org/sqlite v1.61.0.** Il offre l'authorizer (`RegisterAuthorizer`), `_defensive`, `_query_only`, `sqlite.Limit`, la sauvegarde (`NewBackup`, `NewRestore`) et l'interruption d'une requête quand son contexte prend fin.
- **Ce qui existe déjà.**
  - Une limite par adresse : `limiter` dans `internal/web/api.go`, qui s'appuie sur `golang.org/x/time/rate`.
  - La tâche horaire du lot 9.
  - `golang.org/x/sys`, déjà dans le graphe des modules comme dépendance indirecte, pour Landlock.

## Le paquet `windata`

- **Le worker.**
  - C'est une commande cachée, `cahut window-data-worker <dossier>`, que le binaire lance pour une base. Elle reçoit ses demandes sur l'entrée standard et rend ses réponses sur la sortie standard, en JSON, une ligne par message.
  - Avant d'ouvrir la base, elle s'applique ses limites (`setrlimit`), puis Landlock, qui lui laisse le seul dossier de sa base, et le réseau quand le noyau le permet (ABI 4).
  - Elle ouvre `data.db` en WAL, en mode défensif, avec `trusted_schema` coupé, les limites de `sqlite3_limit` et une limite du tas de SQLite.
- **Le pool.**
  - Il lance un worker par base en service, l'arrête après une minute sans appel, et en garde 32 au plus.
  - Chaque appel a son délai. Au-delà, le worker est tué.
  - L'environnement du worker est vide, et il n'hérite d'aucun fichier.
  - Dans les tests des autres paquets, les workers tournent dans le processus, avec le même protocole. Les tests de `windata` lancent de vrais processus.
- **Les demandes.**
  - `select`, pour un lecteur : un `SELECT` limité à ce qui est lisible.
  - `query`, pour une IA : un `SELECT` sur tout.
  - `call` : une requête déclarée.
  - `exec` : les instructions d'une IA.
  - `check` : les migrations, les requêtes et `readable`, dans une transaction annulée.
  - `migrate`, `copy`, `restore`, `schema` et `size`.
- **L'authorizer.** Il y en a un par sorte de demande, en liste blanche d'actions et de fonctions :
  - pour un lecteur : `IGNORE` sur une colonne non lisible, et `DENY` sur une table non listée ;
  - pour une migration : rien qui supprime, renomme ou modifie des lignes.

  La comparaison du schéma, avant et après, repère ce que fait un `ALTER TABLE`. La sorte d'une requête (elle lit, ajoute, modifie ou supprime) se déduit des actions vues quand SQLite la prépare.
- **L'état d'une base.** Il vit dans la base : la table `cahut_migrations` garde le numéro et l'empreinte de chaque migration jouée. Aucune instruction d'une page ni d'une IA ne la lit ni ne l'écrit.

## Modèle de données

- **Sur le disque.** `windows/<id>/data/data.db`, son WAL, et ses copies `windows/<id>/data/copy-<horodatage>.db`, trois au plus.
- **Dans la base de la Cahut.** Une migration ajoute `window_data` : la Window, la taille de sa base et de ses copies, la date de sa dernière écriture et celle de sa dernière copie. Les limites de l'ADR 30 et la copie quotidienne s'en servent. La déclaration `data` vit dans la déclaration de chaque version, sans nouvelle colonne.
- **Les réglages.** La clé de `:reader` se dérive de la clé des frames de la Cahut, par un HMAC avec une étiquette. La limite optionnelle `data_per_window` s'ajoute aux réglages de l'ADR 30.

## Opérations

| Opération | PR | Qui | Notes |
| --- | --- | --- | --- |
| `publish_window`, `update_window` | 3 | Comme aujourd'hui | Prennent `declaration.data` et la vérifient. `update_window` frappe quand une Window publique reçoit une requête qui modifie ou supprime pour `anyone` |
| `get_window` | 3 | Comme aujourd'hui | Donne les données : tables, colonnes, nombre de lignes, taille, migrations, `readable`, requêtes, `closes`, copies |
| `select_window_data` | 4 | Lecteurs (`PeopleOnly`, `Visitors`, `ReadOnly`) | `POST .data/sql` |
| `call_window_query` | 4 | Lecteurs (`PeopleOnly`, `Visitors`, `ReadOnly` pour la base de la Cahut) | `POST .data/queries/<nom>` |
| `query_window_data` | 5 | Membres de la Room et leurs IA (`ReadOnly`) | Marque les lignes comme des données |
| `change_window_data` | 5 | Membres de la Room et leurs IA | `KnocksWhen` : supprimer ou renommer. Annulée par `rollback_window_data` |
| `rollback_window_data` | 5 | Membres de la Room et leurs IA | S'annule elle-même |
| `export_window_data` | 6 | Membres, dans la console (`PeopleOnly`) | CSV d'une table, ou fichier SQLite |

## Les appels des pages

- `POST <racine>/.data/sql` et `POST <racine>/.data/queries/<nom>`, à l'adresse d'une Window publique et sous `/-/frame/<capacité>/`. `OPTIONS` répond au contrôle préalable (preflight) de CORS.
- Les réponses sont en JSON, sans cookie, avec `Access-Control-Allow-Origin: *`. Une erreur donne un statut, un message et un code.
- `:reader` vaut `NULL` à l'adresse d'une Window publique. Pour un laissez-passer, il se calcule par un HMAC sur la Window et le laissez-passer, et pour une personne, sur la Window et la personne. Une frame de modérateur lit sans écrire.
- Les limites : une seconde de travail par appel, 64 Kio envoyés, 1 000 lignes et 1 Mio rendus. Par adresse IP (/64 en IPv6) et par Window, les lectures vont par rafales de 60 puis une par seconde, et les écritures par rafales de 10 puis une toutes les 6 secondes.

## Console

- **L'onglet Data d'une Window.**
  - Les tables et leurs lignes, affichées en texte, page par page.
  - Ce qui est lisible.
  - Les requêtes, avec ce qu'elles font et qui les appelle.
  - La taille de la base, les copies et le retour à l'une d'elles.
  - L'export d'une table en CSV, ou de la base entière.
- **Rendre une Window publique.** La confirmation dit ce que garde la Window, et le Knock aussi.
- **Les refus.** Leurs codes (`window_data_full`, `window_data_closed`, `window_data_busy`…) sont traduits en anglais et en français dans `refusals.ts`.

## Tests

- **Les colonnes privées.** Elles restent illisibles par tous les chemins : `WHERE`, `ORDER BY`, une jointure, une sous-requête, une vue, `json_object`, une CTE.
- **Le SQL hostile.** Une CTE récursive sans fin, un `printf` géant, `zeroblob` ou `randomblob`, une imbrication profonde, une jointure cartésienne : chaque requête s'arrête à temps et dans sa mémoire, et le worker suivant répond.
- **L'enfermement.** Sous Linux, un worker n'ouvre aucun autre fichier.
- **Les migrations.** Une migration qui supprime ou renomme est refusée, comme une migration déjà jouée dont le texte a changé.
- **Le Knock.** `change_window_data` frappe quand on supprime ou renomme, et la base n'a pas changé tant que personne n'a approuvé.
- **Les copies.** Un retour rend la base telle qu'elle était, et un second retour défait le premier.
- **Les appelants.** Les appels des pages sont testés pour chaque sorte de frame et pour une Window publique, et les opérations sur la matrice des appelants de l'ADR 13.
- **La gouvernance.** La table `undo` nomme `rollback_window_data`.

## Découpage en pull requests

| # | Pull request | Contenu |
| --- | --- | --- |
| 1 | Plan | Ce plan |
| 2 | Workers | Le paquet `windata` : le worker, le protocole, le pool, les authorizers, Landlock et les limites, la commande cachée, la ligne de `cahut doctor` ; ses tests, dont le SQL hostile. Rien n'est encore branché |
| 3 | Déclaration | `declaration.data`, sa vérification et ses migrations dans `publish_window` et `update_window`, la copie qui précède les migrations, le Knock de `update_window`, `get_window`, la corbeille et la purge, la migration de `window_data` |
| 4 | Pages | Les routes `.data/`, `select_window_data`, `call_window_query`, `:reader`, les limites par adresse, `closes`, la réponse 410 d'une Window retirée, les frames de modérateur |
| 5 | IA | `query_window_data`, `change_window_data`, `rollback_window_data`, la copie quotidienne, la table `undo`, les consignes MCP, l'écran de consentement |
| 6 | Console | L'onglet Data, `export_window_data`, la confirmation et le Knock d'une Window publique, les refus, les types TypeScript |
| 7 | Limites, documentation et bilan | Le disque et `data_per_window`, le skill, la documentation et sa référence, la feuille de route du site, CLAUDE.md, la démonstration sur le staging, le bilan |

Les pull requests s'enchaînent dans l'ordre et sont fusionnées dès que leur CI est verte. Chacune régénère la référence de la documentation (`make docs-reference`) quand elle change une opération. Cahut 1.0 suit la septième.

## Hors du lot 10

- Modifier des lignes à la main dans la console.
- L'effacement automatique des lignes à une date.
- Les sauvegardes, qui emporteront aussi les bases des Windows.
- La recherche en texte intégral (FTS5) pour les lecteurs, et les mises à jour en direct (SSE).
- Des requêtes déclarées qui envoient un e-mail aux membres de la Room.
