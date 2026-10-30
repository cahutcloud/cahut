# Cahut · Plan du lot 5 : Knock first

23 octobre 2026 · Olivier Durand · Validé, achevé

## En bref

Le lot 5 fait de Knock first une demande plutôt qu'un refus. Quand une IA appelle une opération qui ne s'annule pas, ou qui change ce qu'une personne peut faire, Cahut vérifie l'appel sans l'exécuter et crée un Knock : une demande en attente, qui expire au bout de 24 heures. L'IA reçoit le lien du Knock et le transmet à sa personne, qui approuve ou refuse dans la console ; l'approbation exécute l'opération, et le Logbook montre la demande puis l'exécution. Le lot ajoute trois opérations qui en avaient besoin : purger une Window de la corbeille, supprimer une Room d'équipe, régler la portée d'une IA connectée. Il compte cinq pull requests, pour environ six jours et demi à temps plein, et s'ouvre sur une décision écrite (ADR 21).

**Fin de lot**, reprise du [plan de la V0](plan-v0.md) : Claude demande à rendre une Window publique, la demande apparaît dans la console, l'approbation l'exécute, le Logbook montre les deux étapes.

## Point de départ

Ce que les lots 1 à 4 ont laissé :

- Le registre connaît le drapeau `KnockFirst`, posé au lot 2 sur sept opérations : `create_user`, `set_user_role`, `disable_user`, `enable_user`, `add_room_member`, `set_room_member_role`, `remove_room_member`. Une IA qui en appelle une reçoit `ErrApprovalRequired`, inscrit au Logbook comme un refus, avec un message qui envoie sa personne à la console. MCP marque ces outils `destructiveHint`.
- `set_window_access` refuse de la même façon l'accès public à une IA ([ADR 19](adr/0019-window-versions-access-and-declaration.md)).
- Une Window à la corbeille y reste : rien ne la purge. Une Room ne se supprime pas.
- Une IA connectée garde la portée et les Rooms de son consentement ; pour les élargir, sa personne la coupe et la reconnecte ([ADR 16](adr/0016-oauth-for-ai-clients.md)).
- La console n'a pas de notification. Elle appelle l'API par jeton de session, et ses lectures ne vont pas au Logbook.

## Décision à écrire d'abord

### ADR 21 · An AI knocks first, and the person it acts for answers

L'ADR 21 fixe :

- **Ce qui frappe.** Les sept opérations du lot 2, et deux nouvelles, `purge_window` et `delete_room`, frappent chaque fois qu'une IA les appelle. `set_window_access` frappe quand il rend une Window publique, `set_ai_reach` quand il élargit la portée d'une IA. Une personne ne frappe jamais : dans la console, elle agit tout de suite, après une confirmation. La ligne de commande non plus.
- **Vérifier sans exécuter.** Le registre vérifie l'appel comme tout autre, exécute l'opération dans sa transaction, puis l'annule : l'IA apprend tout de suite qu'un handle est pris ou qu'une Room n'a qu'un propriétaire. Une opération qui frappe garde ses effets dans sa transaction et ne touche au disque qu'après la validation. Une opération qui décide de frapper en cours de route peut s'arrêter là ; `set_ai_reach` s'arrête avant de vérifier les Rooms nommées, pour qu'une IA limitée n'apprenne pas, en les nommant, celles de sa personne.
- **Le Knock.** L'opération, son entrée, ce qu'elle touche, la Room, l'IA, la personne, la création, l'expiration. Son entrée ne porte aucun secret. Il expire au bout de 24 heures, sans réglage en V0. Le même appel de la même IA rend le Knock qui attend déjà ; une IA a au plus 20 Knocks en attente.
- **Ce que reçoit l'IA.** Une erreur, dont le message nomme la personne, donne l'adresse du Knock dans la console, `https://<domaine>/console/knocks/<id>`, et son expiration. `get_knock` et `list_knocks` lui donnent l'issue des siens. L'adresse n'est pas un secret : seule la personne, connectée, y répond.
- **Qui répond.** La personne pour qui l'IA agit, et elle seule : un administrateur ne répond pas pour une autre personne. `approve_knock` et `deny_knock` sont réservées aux personnes.
- **L'approbation.** Elle exécute l'opération comme l'IA, avec l'entrée du Knock, dans la transaction qui marque le Knock : le registre revérifie tout à cet instant. Une opération qui échoue marque le Knock échoué, sans autre effet. Le résultat va à la personne qui approuve, jamais à l'IA : approuvé, `create_user` lui rend le lien d'activation.
- **La fin.** Un Knock finit exécuté, échoué, refusé, expiré, ou annulé quand son IA perd l'accès avant la réponse : jeton révoqué, connexion coupée, personne désactivée.
- **Notification.** La console montre le nombre de Knocks en attente dans sa navigation et dans le titre de l'onglet, et le redemande toutes les 30 secondes. L'e-mail viendra avec l'envoi d'e-mails.
- **Logbook.** L'appel qui frappe s'inscrit sous son opération avec l'issue `knocked` et le numéro du Knock ; l'exécution approuvée, sous la même opération, au nom de l'IA, par la console, avec le même numéro ; le refus, sous `deny_knock`. L'expiration et l'annulation n'y vont pas.

L'ADR 21 répond à deux questions ouvertes du plan de la V0 : la liste des actions irréversibles, et l'expiration par défaut d'un Knock.

## Modèle de données

Une migration, `cahut/0009_knocks` :

| Table | Contenu |
| --- | --- |
| `knocks` | Personne, jeton ou connexion de l'IA, nom de l'IA, opération, entrée en JSON, cible et détail, Room, création, expiration, état, date et auteur de la réponse, message d'erreur |
| `logbook` | Colonne `knock_id` ; l'issue `knocked` |

Les Windows ont une clé étrangère vers leur Room sans cascade : `delete_room` efface d'abord ses Windows. Le Logbook n'en a pas : ses entrées survivent à la Room.

## Le registre

Comment naît un Knock :

1. Une IA appelle une opération. Le registre vérifie l'acteur, le niveau, la portée, le rôle et l'état de la personne, comme pour tout appel.
2. Il exécute l'opération dans sa transaction. Elle frappe si elle est marquée `KnockFirst`, ou si elle appelle `c.Knock()` en cours de route ; `c.Knock()` lui dit aussi si l'appel ne fait que frapper, pour qu'elle puisse s'arrêter.
3. Si l'appel frappe, le registre annule la transaction et exécute les fonctions d'échec de l'opération, puis, dans une nouvelle transaction, crée le Knock ou retrouve celui qui attend, et écrit l'entrée `knocked`.
4. Il rend une erreur qui enveloppe `ErrApprovalRequired` et porte le Knock ; l'adaptateur MCP en tire le message.

Comment s'exécute un Knock approuvé : `approve_knock` charge le Knock, vérifie qu'il attend, qu'il n'a pas expiré et qu'il revient à la personne, relit l'IA (jeton, connexion, portée), puis exécute l'opération comme l'IA, marquée approuvée, dans un point de sauvegarde (`SAVEPOINT`) de sa propre transaction. Il marque le Knock exécuté, ou échoué en revenant au point de sauvegarde, et écrit l'entrée du Logbook, le tout dans une seule transaction.

## Opérations

| Opération | Qui l'appelle | Ce qu'elle fait |
| --- | --- | --- |
| `approve_knock` | personne, pour ses IA | Exécute l'opération du Knock comme l'IA ; rend son résultat |
| `deny_knock` | personne, pour ses IA | Refuse le Knock |
| `list_knocks` | personne : ceux de ses IA ; IA : les siens | Les Knocks en attente, puis les derniers terminés |
| `get_knock` | personne : ceux de ses IA ; IA : les siens | Un Knock, son état, son expiration, son erreur s'il a échoué |
| `purge_window` | membre de la Room ; frappe pour une IA | Efface pour de bon une Window de la corbeille, ses versions et ses fichiers ; son nom redevient libre |
| `delete_room` | propriétaire de la Room, administrateur ; frappe pour une IA | Efface pour de bon une Room d'équipe, ses membres et ses Windows, corbeille comprise ; son nom redevient libre |
| `set_ai_reach` | personne pour ses IA ; une IA pour elle-même, qui frappe pour élargir | Règle la portée et les Rooms d'une IA connectée par OAuth ; offerte aussi à une IA en lecture seule |
| `set_window_access` | inchangée | Frappe pour l'accès public, au lieu de refuser |
| Les sept opérations du lot 2 | inchangées | Frappent au lieu de refuser |

`approve_knock` et `deny_knock` sont réservées aux personnes (`PeopleOnly`). `list_knocks` et `get_knock` sont des lectures, offertes aux IA. `create_user`, exécutée sur approbation, crée le lien d'activation pour la personne qui approuve.

## MCP

- Le message d'un appel qui frappe, par exemple : « Waiting for Olivier Durand to approve: making a Window public cannot be undone. Give them this link: https://staging.cahut.cloud/console/knocks/12. It expires on 24 October 2026 at 14:05 UTC. Call get_knock with id 12 to learn the outcome. »
- Les descriptions des outils qui frappent le disent ; celle de `set_window_access` ne parle plus de refus.
- Les instructions du serveur MCP ajoutent une phrase : certaines actions attendent une personne, et l'IA lui transmet le lien.
- `destructiveHint` reste sur les opérations marquées `KnockFirst`.

## Console

| Écran | Contenu |
| --- | --- |
| Navigation | « Knocks », avec le nombre en attente ; le titre de l'onglet le porte aussi ; redemandé toutes les 30 secondes et au retour sur l'onglet |
| Knocks, `/knocks` | Les Knocks en attente des IA de la personne : la demande en une phrase, l'IA, l'expiration, Approve et Deny ; les Knocks terminés en dessous, avec leur issue |
| Un Knock, `/knocks/<id>` | Connexion d'abord si besoin ; la demande en une phrase et le détail de son entrée ; l'IA et son client ; l'expiration ; Approve et Deny ; après l'approbation, le résultat, dont le lien d'activation d'une personne créée |
| Fiche d'une Window | Dans la corbeille, « Purge », après une confirmation |
| Room | « Delete Room » pour une Room d'équipe, après une confirmation qui dit combien de Windows partent avec elle et demande de taper le nom de la Room |
| AIs and access | « Change » sur une IA connectée : sa portée et ses Rooms, comme sur l'écran de consentement |

La console écrit chaque opération qui frappe en une phrase (« Make the Window Q3 review public », « Add Jane Doe to the Cahut as a member ») et montre toute autre par son nom et son entrée.

## Tests

- **Naissance d'un Knock** : chaque opération qui frappe, appelée par une IA, ne change rien, ni en base ni sur le disque, crée un Knock et une entrée `knocked` ; une entrée fausse est refusée tout de suite ; le même appel rend le même Knock ; la limite de 20 ; une personne et la ligne de commande agissent tout de suite.
- **Réponse** : la personne approuve et l'opération s'exécute comme l'IA, avec une entrée au Logbook qui porte le Knock ; une autre personne, un administrateur et une IA ne peuvent ni approuver ni refuser ; refus ; expiration ; annulation après révocation du jeton, coupure de la connexion, désactivation de la personne ; une opération qui échoue à l'approbation laisse tout en l'état et marque le Knock échoué ; un Knock ne s'approuve qu'une fois.
- **Résultat** : `create_user` approuvée rend le lien d'activation à la personne, jamais à l'IA ; `get_knock` ne le montre pas.
- **Opérations conditionnelles** : `set_window_access` frappe pour `public` seulement ; `set_ai_reach` frappe pour élargir et agit pour restreindre ; une IA limitée qui nomme une Room inconnue ne l'apprend pas avant l'approbation ; une IA en lecture seule voit `set_ai_reach` et rien d'autre qui écrive.
- **Purge et suppression** : `purge_window` refuse une Window hors de la corbeille, efface ses versions, ses fichiers et ses laissez-passer, libère son nom ; `delete_room` refuse une Room personnelle, efface les Windows de la Room et leurs fichiers, garde le Logbook, coupe la Room des connexions limitées.
- **Cloisonnement** : le Knock d'une Cahut n'existe pas dans une autre.
- **Console** : dans le navigateur, le badge, l'approbation depuis le lien d'une IA, la purge et la suppression d'une Room avec leurs confirmations, le changement de portée.

## Découpage en pull requests

| # | Pull request | Contenu | Durée |
| --- | --- | --- | --- |
| 1 | Décision | Ce plan et l'ADR 21 | 0,5 j |
| 2 | Knocks | Migration 0009, Knocks dans le registre, `approve_knock`, `deny_knock`, `list_knocks`, `get_knock`, Logbook, message MCP ; les opérations du lot 2 et l'accès public frappent | 2 j |
| 3 | Opérations irréversibles | `purge_window`, `delete_room`, `set_ai_reach` | 1,5 j |
| 4 | Console | Knocks et leur page, badge, purge, suppression d'une Room, portée d'une IA, types TypeScript | 1,5 j |
| 5 | Documentation et démonstration | Pages de la documentation, référence, consignes MCP, CLAUDE.md, staging, bilan | 1 j |

Soit environ 6,5 jours. Les pull requests s'enchaînent dans l'ordre.

## Démonstration de fin de lot

**Prérequis** : la version du lot installée sur `staging.cahut.cloud`, Claude Code branché par OAuth avec la Room `marketing`.

1. Demander à Claude de rendre publique une Window de `marketing`. Claude reçoit le Knock et donne son lien.
2. La console montre « Knocks 1 ». Ouvrir le lien : la demande, l'IA, l'expiration.
3. Approuver : la Window s'ouvre en navigation privée, sans connexion.
4. Claude appelle `get_knock` et confirme que la Window est publique.
5. Le Logbook montre l'appel `knocked`, puis l'exécution approuvée, avec le même numéro.
6. Demander à Claude de purger une Window de la corbeille, puis refuser : `get_knock` le dit, et la Window reste dans la corbeille.
7. Brancher une IA en lecture seule, lui demander de publier : elle demande l'écriture par `set_ai_reach`, l'approbation la lui donne, elle publie.

## Hors du lot 5

- Les notifications par e-mail : avec l'envoi d'e-mails, après la V0.
- Une expiration réglable par Cahut, et des Knocks auxquels un administrateur ou plusieurs personnes répondent : après la V0.
- Effacer des sauvegardes ou en changer la destination : avec les sauvegardes, après la V0.
- Vider la corbeille d'elle-même au bout d'un temps : après la V0.
- Supprimer une personne : pas en V0, où on la désactive ([ADR 13](adr/0013-people-sessions-and-rights.md)).

## Bilan

Le lot 5 s'est achevé le 23 octobre 2026, en cinq pull requests :

| # | Pull request | Contenu |
| --- | --- | --- |
| 1 | #57 | Le plan et l'ADR 21 |
| 2 | #58 | Les Knocks dans le registre, le paquet `knock`, ses quatre opérations ; les opérations du lot 2 et l'accès public frappent |
| 3 | #60 | `purge_window`, `delete_room`, `set_ai_reach`, et des identifiants de Rooms et de Windows jamais réutilisés |
| 4 | #61 | La console : Knocks et leur page, compteur, purge, suppression d'une Room, portée d'une IA |
| 5 | #62 | Documentation, site, staging, bilan |

**Écarts par rapport au plan.**

- La migration `cahut/0009` refait la table du Logbook : SQLite ne change pas la contrainte de sa colonne `outcome`, qui doit admettre `knocked`.
- Une migration de plus, `cahut/0010` : SQLite donne à une nouvelle ligne le plus grand identifiant plus un, si bien qu'une Room supprimée aurait cédé son identifiant à la suivante, dont les membres auraient lu le Logbook de l'ancienne. Les tables `rooms` et `windows` sont refaites avec `AUTOINCREMENT`. Une migration qui commence par `-- foreign keys off` s'exécute désormais clés étrangères coupées, comme SQLite le demande, et les vérifie avant de valider.
- `set_ai_reach` est offerte aux IA en lecture seule par un nouveau drapeau des opérations, `ReadOnlyAIs`.
- `delete_room` vit dans le paquet `window`, qui garde les fichiers des Windows.
- Les états expiré et annulé ne s'écrivent pas en base : ils se lisent de l'heure, et du jeton ou de la connexion de l'IA. L'état annulé s'écrit `canceled`, en anglais américain comme le veut le linter : l'ADR 21 est corrigée en ce sens.
- La console dit « Decline », comme la carte Knock du kit de la charte ; l'opération reste `deny_knock`.
- Ce que rend une approbation, comme le lien d'activation d'une personne créée, s'affiche aussi sur la liste des Knocks, pas seulement sur la page d'un Knock.
- Au scénario du lot 2, l'IA de Bruno, qui n'est pas propriétaire de `marketing`, est refusée tout de suite au lieu de frapper : l'appel est vérifié avant de devenir un Knock.
- Pas de nouvelle dépendance.

**Essais.**

- Les tests couvrent la vérification à blanc (rien ne change, en base ni sur le disque), l'appel faux refusé tout de suite, les Knocks conditionnels, l'approbation exécutée comme l'IA puis revérifiée, l'échec à l'approbation, le refus, l'expiration, l'annulation par jeton révoqué ou connexion terminée, le même Knock pour le même appel, la limite de 20, le lien d'activation remis à la personne et jamais à l'IA, seule la personne qui répond ; la purge, la suppression d'une Room, `set_ai_reach` ; la migration 0010 sur une base de la migration 9 ; de bout en bout, un Knock fait par MCP et approuvé par l'API de la console.
- En local, dans le panneau de navigateur, contre le serveur de développement et ses données, mises à jour par les migrations 9 et 10 : une IA de test frappe par MCP pour rendre une Window publique et ajouter une personne ; le lien du Knock connecte sur place et y reste ; l'approbation rend la Window publique et `get_knock` le dit ; le lien d'activation s'affiche après l'approbation depuis la liste ; un nouveau Knock apparaît dans le compteur et dans le titre de l'onglet en moins de 30 secondes ; refus d'une purge ; le Logbook et ses liens vers les Knocks ; purge d'une Window et suppression d'une Room, fichiers effacés, identifiant non réutilisé ; une IA en lecture seule frappe pour écrire, puis « Change » lui donne toutes les Rooms ; la page d'un Knock à 375 px.
- Sur le staging, la version du lot est installée (0.2.1-0.20261023091304-c51a955d85e1), avec une sauvegarde dans `/var/lib/cahut.bak-2026-10-23-batch5`. Les migrations sont passées, et les Windows, leurs versions, leurs Keys et leurs déclarations sont intactes. Le connecteur `cahut-staging` de Claude Code s'est reconnecté avec les nouveaux outils et les nouvelles consignes. Claude a publié [Batch 5 demo](https://staging.cahut.cloud/marketing/batch-5-demo/) dans `marketing`, puis demandé à la rendre publique : le Knock 1 a attendu, la Window fermée aux inconnus. Olivier l'a approuvé dans la console : la Window s'ouvre au public, `get_knock` dit `done`, et le Logbook montre l'appel `knocked` par MCP puis son exécution par la console, sous le même numéro.
- La release [Cahut 0.3.0](https://github.com/cahutcloud/cahut/releases/tag/v0.3.0) est publiée, et installée sur le staging depuis GitHub, avec une sauvegarde dans `/var/lib/cahut.bak-2026-10-23-v0.3.0`.

**Reste à faire.**

- La fin de la démonstration, avec le compte d'Olivier : le refus d'une purge, et une IA en lecture seule qui demande l'écriture.
- Firefox et Safari, toujours, depuis le lot 4.

**Suites.**

- Les notifications par e-mail, avec l'envoi d'e-mails, après la V0.
- `cahut logbook`, en ligne de commande, ne montre pas encore le numéro du Knock d'une entrée.
