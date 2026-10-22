# Cahut · Plan du lot 3 : brancher les IA par OAuth

22 octobre 2026 · Olivier Durand · Validé

## En bref

Le lot 3 permet de brancher une IA sur un Cahut sans copier de jeton : l'IA ouvre la console, la personne se connecte, choisit les Rooms et la portée qu'elle lui donne, et l'IA reçoit ses jetons. Chaque Cahut devient son propre serveur d'autorisation OAuth 2.1, sur le port de la console ([ADR 6](adr/0006-in-house-oauth-server.md), [ADR 11](adr/0011-console-on-its-own-port.md)). Le lot compte cinq pull requests, pour environ huit jours à temps plein. Il s'ouvre sur une décision écrite (ADR 16) et se termine par le jalon PoC.

**Fin de lot, jalon PoC**, reprise du [plan de la V0](plan-v0.md) : Cahut ajouté comme connecteur dans claude.ai, une présentation publiée dans une Room d'équipe, visible dans la console.

## Point de départ

Ce que le lot 2 a laissé :

- Une IA se branche avec un jeton personnel, créé dans la console ou en ligne de commande, qu'on colle dans la commande `claude mcp add`. Ce jeton agit avec tous les droits de sa personne.
- Le registre connaît l'acteur IA, avec la personne pour qui il agit et le jeton qu'il présente. Il relit le rôle et l'état de la personne à chaque appel.
- La console tourne sur le port 8443, avec une session gardée dans le `localStorage` de son origine, sans cookie. Son écran « AIs and access » liste les jetons personnels.
- L'endpoint MCP répond 401 sans jeton, mais sans en-tête `WWW-Authenticate` qui dise où s'autoriser : aucun client ne peut découvrir de serveur d'autorisation.

## Décision à écrire d'abord

### ADR 16 · OAuth for AI clients, in practice

L'ADR 6 a choisi d'écrire le serveur d'autorisation. L'ADR 16 en fixe la forme :

- **Adresses.** L'émetteur est `https://<domaine>:8443`. Ses métadonnées (RFC 8414) sont à `/.well-known/oauth-authorization-server` ; l'autorisation à `/oauth/authorize`, une page de la console ; le jeton à `/oauth/token`, l'enregistrement à `/oauth/register`, la révocation à `/oauth/revoke`. Les métadonnées de la ressource protégée (RFC 9728) sont sur le port 443, à `/.well-known/oauth-protected-resource/mcp`, et la ressource est `https://<domaine>/mcp`.
- **Portée.** Deux valeurs : `read`, qui ne permet que les opérations de lecture, et `write`, qui permet tout ce que la personne peut faire. La personne choisit, sans dépasser ce que le client a demandé.
- **Rooms.** La personne donne toutes ses Rooms, y compris celles qu'elle rejoindra, ou une liste. Une IA limitée ne voit pas les autres Rooms, n'y publie pas et ne crée pas de Room.
- **Connexion.** Chaque consentement crée une connexion : une IA, la personne pour qui elle agit, sa portée et ses Rooms. C'est ce que la console appelle une IA connectée. Une connexion prend fin quand la personne ou un administrateur la coupe, quand le client révoque son jeton de rafraîchissement, après 30 jours sans usage, et elle est suspendue avec sa personne.
- **Jetons.** Opaques, tirés de `crypto/rand` et gardés hachés, comme les jetons personnels : `cahut_oat_` pour l'accès, valable une heure, `cahut_ort_` pour le rafraîchissement, valable 30 jours et renouvelé à chaque usage. Un jeton de rafraîchissement déjà utilisé coupe la connexion, sauf dans les 30 secondes qui suivent son renouvellement, pour ne pas punir deux rafraîchissements simultanés.
- **Codes d'autorisation.** Valables 5 minutes, à usage unique, liés au client, à l'adresse de retour, au PKCE S256 et à la ressource. Un code présenté deux fois coupe la connexion.
- **Clients.** Par Client ID Metadata Document, récupéré avec une garde contre le SSRF (adresses publiques seulement, adresse vérifiée à la connexion, pas de redirection, 5 secondes, 16 Kio) et gardé en mémoire une heure ; ou par enregistrement dynamique, ouvert, limité en débit par adresse, qui ignore les champs inconnus. Un client enregistré qui n'a reçu aucune connexion en 7 jours est effacé.
- **Adresses de retour.** Égalité exacte, sauf pour les adresses de bouclage en `http` (`127.0.0.1`, `[::1]`, `localhost`), dont le port est libre (RFC 8252). Les schémas `https`, `http` en bouclage et les schémas privés des applications sont admis ; `javascript`, `data`, `file` et leurs semblables sont refusés.
- **Consentement.** L'écran est une page de la console, qui appelle des opérations du registre, réservées aux personnes : `describe_authorization`, `approve_authorization`, `deny_authorization`. Il montre le nom du client et l'hôte de l'adresse de retour. Sans cookie, il n'y a pas de CSRF ; `frame-ancestors 'none'` empêche le détournement de clic.
- **IA connectées.** `list_connected_ais` et `disconnect_ai`, réservées aux personnes : chacun pour ses IA, un administrateur pour toutes, comme les jetons personnels.
- **Logbook.** Le consentement, le refus d'une connexion et sa coupure s'y inscrivent. Chaque appel d'une IA connectée porte la connexion et le nom de son client. Les échanges de jetons ne s'y inscrivent pas, sauf la coupure d'une connexion pour un jeton rejoué.
- **Knock first.** Au lot 3, la personne consent elle-même dans la console : rien à approuver. Élargir la portée d'une IA déjà connectée passera par Knock first au lot 5.

## Modèle de données

Une migration, `cahut/0004` :

| Table | Contenu |
| --- | --- |
| `oauth_clients` | Clients enregistrés dynamiquement : identifiant, nom, adresses de retour, adresse du site, création, dernière connexion |
| `connections` | Les IA connectées : personne, identifiant du client, nom du client, hôte de retour, portée, toutes les Rooms ou non, création, dernier usage, fin et sa raison |
| `connection_rooms` | Les Rooms d'une connexion limitée |
| `oauth_codes` | Code haché, connexion, adresse de retour, défi PKCE, ressource, expiration, usage |
| `oauth_tokens` | Jeton haché, connexion, sorte (accès ou rafraîchissement), création, expiration, renouvellement |
| `logbook` | Colonne `connection_id`, à côté de `token_id` |

## Opérations

Toutes réservées aux personnes : elles prouvent qui consent, ou touchent aux accès des IA, comme les jetons personnels.

| Opération | Qui l'appelle | Ce qu'elle fait |
| --- | --- | --- |
| `describe_authorization` | personne | Vérifie une demande d'autorisation, rend le client, l'hôte de retour, la portée demandée et les Rooms à choisir ; ou l'adresse où renvoyer l'erreur |
| `approve_authorization` | personne | Crée la connexion et le code, rend l'adresse de retour |
| `deny_authorization` | personne | Rend l'adresse de retour avec `access_denied` |
| `list_connected_ais` | personne ; un admin pour tous | Les IA connectées : client, personne, portée, Rooms, dernier usage |
| `disconnect_ai` | personne ; un admin pour tous | Coupe une connexion : ses jetons cessent à l'instant |

Les points du protocole (métadonnées, enregistrement, jeton, révocation) ne sont pas des opérations : ce sont des échanges de jetons, servis par le paquet `internal/oauth`. La révocation d'une connexion par son client passe par le registre et s'inscrit au Logbook.

## Portée d'une IA dans le registre

- L'acteur IA porte sa portée : lecture seule ou non, et la liste de ses Rooms quand elle est limitée. Un jeton personnel garde tous les droits de sa personne.
- Le registre refuse une opération qui écrit à une IA en lecture seule, et l'endpoint MCP ne lui offre que les outils de lecture.
- L'assistant commun des Rooms (`room.Find`) répond qu'une Room hors de portée n'existe pas. `list_rooms`, `list_windows`, `get_account` et `list_logbook` ne montrent que les Rooms de la connexion ; `create_room` est refusée à une IA limitée.

## Console

| Écran | Contenu |
| --- | --- |
| Consentement, `/oauth/authorize` | Connexion d'abord si besoin, puis : le nom de l'IA et l'hôte où elle sera renvoyée, la portée (lecture, ou lecture et écriture), les Rooms (toutes, ou une liste à cocher), Autoriser et Refuser |
| AIs and access | Une section « Connected AIs » au-dessus des jetons personnels : l'IA, pour qui elle agit, sa portée, ses Rooms, son dernier usage, Disconnect ; l'adresse MCP à donner à un client, sans jeton |

## Serveur web

- Sur le port 8443 : les métadonnées du serveur d'autorisation et les points `/oauth/token`, `/oauth/register`, `/oauth/revoke`, avec CORS ouvert puisqu'ils ne lisent aucun cookie ; `/oauth/authorize` sert la console.
- Sur le port 443 : les métadonnées de la ressource, et un 401 de `/mcp` qui porte `WWW-Authenticate: Bearer resource_metadata="…"`.
- `/mcp` accepte les jetons personnels et les jetons d'accès OAuth.
- En développement, l'émetteur est `http://localhost:8081` et la ressource `http://localhost:8080/mcp`.

## Tests

- **Protocole** : un parcours complet avec le client OAuth du SDK Go de MCP, par enregistrement dynamique et par Client ID Metadata Document.
- **Pièges de l'ADR 6** : code à usage unique et lié, PKCE manquant ou faux, ressource étrangère, adresse de retour inexacte et bouclage à port libre, jeton de rafraîchissement rejoué hors et dans le délai de grâce, `iss` dans la réponse, garde SSRF (adresses privées, redirection, réponse trop grosse), champs inconnus et limite de débit de l'enregistrement, purge des clients.
- **Portée** : une IA en lecture seule ne voit ni n'appelle les outils d'écriture ; une IA limitée ne voit ni ne touche les autres Rooms ; une personne désactivée suspend ses connexions ; une connexion coupée ferme ses jetons.
- **Cloisonnement** : un jeton d'un Cahut est refusé par un autre ; la ressource d'un autre Cahut est refusée.
- **Clients réels** : Claude Code en local, de bout en bout dans le navigateur ; claude.ai et ChatGPT sur le staging.

## Découpage en pull requests

| # | Pull request | Contenu | Durée |
| --- | --- | --- | --- |
| 1 | Décision | Ce plan et l'ADR 16 | 0,5 j |
| 2 | Portée d'une IA | Portée et Rooms de l'acteur, appliquées par le registre et les opérations, outils filtrés | 1 j |
| 3 | Serveur d'autorisation | Migration, paquet `internal/oauth`, métadonnées, enregistrement, CIMD, autorisation, jeton, révocation, opérations, MCP | 3,5 j |
| 4 | Console | Écran de consentement, IA connectées, types TypeScript | 1,5 j |
| 5 | Documentation et jalon | Pages de la documentation, CLAUDE.md, staging, essais avec Claude Code, claude.ai et ChatGPT, bilan | 1,5 j |

Soit environ 8 jours. Les pull requests s'enchaînent dans l'ordre.

## Démonstration de fin de lot

**Prérequis** : la release du lot installée sur `staging.cahut.cloud`, le port 8443 ouvert.

1. Dans claude.ai, ajouter un connecteur personnalisé à l'adresse `https://staging.cahut.cloud/mcp`.
2. claude.ai ouvre la console : se connecter, choisir la Room d'équipe `marketing` et la portée « lecture et écriture », autoriser.
3. Demander à Claude une présentation, publiée dans `marketing`.
4. Dans la console, la Window apparaît dans la Room, et le Logbook montre le consentement puis la publication, au nom de l'IA connectée.
5. L'IA essaie de publier dans la Room personnelle : la Room n'existe pas pour elle.
6. Dans « AIs and access », couper la connexion : l'appel suivant de claude.ai est refusé.
7. Test de fumée avec ChatGPT, en mode développeur : brancher le même Cahut et lister les Windows.

Les étapes 1 à 7 demandent les comptes claude.ai et ChatGPT d'Olivier : il les rejoue lui-même. Le lot les prépare par un parcours scripté sur le staging et par Claude Code en local.

## Hors du lot 3

- Knock first, et donc l'élargissement de la portée d'une IA connectée : lot 5. D'ici là, on coupe la connexion et on en crée une autre.
- Les accès des Windows : lot 4. Une Window reste lisible par quiconque a son adresse.
- OpenID Connect, et le SSO des personnes : après la V0.
- Le deuxième client de bout en bout : lot 8. Le lot 3 n'en fait qu'un test de fumée.
