# Cahut · Revue de sécurité de la V0

23 octobre 2026 · Lot 8 · Olivier Durand

## En bref

Avant le jalon de la V0, cinq relecteurs ont lu le code de Cahut, chacun sur un domaine, à la recherche d'attaques concrètes : qui attaque, ce qu'il envoie, ce qu'il gagne. Ils n'ont trouvé aucun moyen de passer d'un Cahut à un autre, ni pour une IA d'approuver ses propres Knocks ou d'appeler une opération réservée aux personnes, ni pour le script d'une Window de lire la console ou la session d'un lecteur. Ils ont trouvé deux faiblesses graves, toutes deux d'épuisement du serveur, sept moyennes et dix faibles. Toutes sont corrigées dans le lot, avec un test, sauf une, qui tient à la conception et attend une décision après la V0 : une Window peut imiter la page de connexion de la console, sur la même adresse. Les corrections sont dans les pull requests 3 et 4 du lot, [getcahut/cahut#88](https://github.com/getcahut/cahut/pull/88) et la suivante.

Cette revue est faite par le projet lui-même, avec des agents de Claude Code. Elle ne remplace pas un audit par un tiers, à décider avant l'ouverture du dépôt.

## Méthode

Le code relu est celui de `main` au commit `52ad20e`, à la fin du lot 7. Chaque domaine a eu son relecteur, qui ne connaissait que le code, les ADR et le domaine à lire, sans hypothèse à confirmer :

| Domaine | Ce qui a été lu |
| --- | --- |
| OAuth et jetons des IA | `internal/oauth`, l'écran de consentement, la vérification des jetons dans l'adaptateur MCP, les jetons personnels |
| Bac à sable des Windows | Le service des Windows et de `/-/`, la déclaration et sa politique CSP, les adresses à capacité, la page de la Key, la house style, les aperçus de la console |
| Cloisonnement entre Cahuts | Le choix du Cahut par l'hôte et par SNI, les certificats, l'API de la console, les administrateurs du serveur, `rename_domain`, le lien du premier démarrage, le programme de mise à jour |
| Sessions et Knocks | La connexion, les liens d'activation, les sessions de la console et des lecteurs, l'approbation des Knocks, le texte des IA affiché dans la console |
| Limites de taille et de débit | Les délais du serveur HTTP, les tailles des requêtes, les limiteurs, le coût d'Argon2id, l'enregistrement des clients, la mémoire et le disque à la portée d'un visiteur ou d'une IA |

Chaque relecteur a prouvé ce qu'il pouvait par un test, dans une copie du dépôt, et a dit ce qu'il avait lu et trouvé sain. Chaque constat a ensuite été vérifié, corrigé et couvert par un test de non-régression.

## Ce qui a été trouvé sain

- **Le cloisonnement entre Cahuts.** Le Cahut d'une requête se choisit par son hôte, mis en minuscules et sans port, et une requête à un domaine inconnu reçoit 421. Chaque Cahut lit sa propre base ; un jeton, une session, un code OAuth ou un lien de premier démarrage d'un Cahut est inconnu des autres. Les fichiers des Windows s'ouvrent par `os.Root`, sous le dossier de leur Cahut. Après `rename_domain`, l'ancien domaine ne sert plus rien.
- **Les droits.** Une IA ne peut ni appeler une opération `PeopleOnly` ni approuver un Knock : l'API de la console ne prend que les sessions des personnes. Un Knock s'approuve une fois, par la personne pour qui l'IA agit, avec l'entrée enregistrée, que la console montre en entier ; l'appel est vérifié de nouveau à l'approbation.
- **Le bac à sable.** Chaque réponse part avec une politique `sandbox`, que seuls les gestionnaires de Cahut remplacent ; les fichiers d'une Window, de tout type, et leurs erreurs gardent la leur, sans `allow-same-origin` ni `allow-top-navigation`. La déclaration ne laisse rien injecter dans la politique. Les adresses à capacité sont des HMAC propres à chaque Cahut, et chaque fichier revérifie l'accès.
- **OAuth.** PKCE `S256` obligatoire, adresses de retour comparées exactement, codes à usage unique et liés à leur client, jetons de rafraîchissement renouvelés et rejeu détecté, `iss` dans chaque réponse d'autorisation, document de client récupéré derrière une garde qui refuse les adresses privées, sans redirection ni mandataire.
- **Les secrets.** Jetons de 256 bits dont seul le SHA-256 est gardé, mots de passe en Argon2id aux paramètres de l'OWASP, aucun jeton ni mot de passe au Logbook ni dans les journaux, cookies des lecteurs en `__Host-`, `HttpOnly`, `Secure` et `SameSite=Lax`.
- **La console.** Aucun HTML venu d'une IA n'y est interprété ; elle refuse d'être encadrée (`frame-ancestors 'none'`) et isole sa fenêtre (`Cross-Origin-Opener-Policy`).
- **Le programme de mise à jour.** La demande de la console est lue sans suivre de lien, bornée à 64 octets ; une version plus ancienne est refusée deux fois ; les signatures se vérifient avec les clés du binaire.

## Constats

| # | Gravité | Domaine | Constat | Ce qui en est fait |
| --- | --- | --- | --- | --- |
| 1 | Grave | Limites | Une IA, même en lecture seule, écrivait au Logbook des lignes aussi longues que sa requête (40 Mio) ; trois lignes faisaient tomber le serveur chaque fois qu'on lisait le Logbook | Chaque champ d'une entrée est coupé à 512 ou 1 024 octets (PR 3) |
| 2 | Grave | Limites, sessions | La connexion hachait le mot de passe en tenant la seule connexion d'écriture du Cahut ; une rafale de connexions ralentissait toutes les écritures, et une adresse IPv6 contournait les limites | Le mot de passe se vérifie avant la transaction, par une étape `Prepare` du registre ; huit connexions au plus se vérifient à la fois ; un mot de passe trop long est refusé sans hachage ; le corps des appels sans session tient en 16 Kio ; une adresse IPv6 compte avec tout son /64 (PR 3) |
| 3 | Moyenne | Limites, sessions | Dix mauvais mots de passe, de n'importe où, bloquaient le bon pour 15 minutes, et la réponse disait si le handle existait | Les refus se comptent par handle et par adresse (10), et par handle de partout (100) ; un handle inconnu se comporte comme un connu ; le Logbook garde l'adresse de chaque connexion (PR 3) |
| 4 | Moyenne | Limites, bac à sable | Quelques adresses fermaient une Window à Key à tous ses lecteurs | Une Window attaquée s'ouvre encore à qui vient avec la bonne Key ; chaque adresse qui s'est trompée attend une minute (PR 3) |
| 5 | Moyenne | Limites, OAuth | L'enregistrement des clients, ouvert à tous, pouvait remplir le disque depuis un seul /64 | Dix adresses de retour de 512 octets au plus, mille clients en attente au plus par Cahut, effacés après 24 heures, et le /64 (PR 3) |
| 6 | Moyenne | OAuth | Un document de client lent, récupéré dans la transaction, bloquait tout le Cahut | Le document se récupère avant la transaction, huit à la fois au plus sur le serveur (PR 3) |
| 7 | Moyenne | Bac à sable | Une session ou un passe effacé rendait son numéro au suivant, et une ancienne adresse de cadre se rouvrait | Les tables `sessions` et `window_passes` sont refaites avec `AUTOINCREMENT` (migration 13, PR 3) |
| 8 | Moyenne | Sessions | Un double clic piégé pouvait tomber sur « Approve » ou « Connect » | Ces boutons s'activent après un moment au premier plan et un mouvement du pointeur ou une touche (PR 3) |
| 9 | Moyenne | Bac à sable | Une Window publiée par une IA peut imiter la connexion de la console, sur la même adresse, et envoyer le mot de passe ailleurs | Risque accepté pour la V0 : voir plus bas |
| 10 | Faible | Sessions | Se déconnecter de la console laissait la session de lecture ouverte dans le navigateur | La déconnexion ferme aussi la session de lecture qu'elle a ouverte (PR 3) |
| 11 | Faible | OAuth, sessions | Une demande invalide renvoyait seule vers le site du client, ce qui faisait de la console une redirection ouverte | La console montre l'erreur et l'hôte, et n'y va que sur un clic (PR 3) |
| 12 | Faible | OAuth | Un hôte de retour en lettres d'un autre alphabet passait pour un autre site sur l'écran de consentement | Les hôtes doivent être en ASCII, et l'écran les montre en minuscules, comme le navigateur les lit (PR 3) |
| 13 | Faible | Bac à sable | Un autre site pouvait compter les cadres d'une page de Cahut et savoir si son visiteur ouvre une Window | Les pages de Cahut envoient `Cross-Origin-Opener-Policy: same-origin` (PR 3) |
| 14 | Faible | Limites | Les sessions MCP n'avaient pas de plafond par jeton | Un jeton ouvre vingt sessions d'un coup, puis une toutes les trois secondes (PR 4) |
| 15 | Faible | Limites | Rien ne bornait les grosses requêtes MCP simultanées ni la mémoire du service | Deux requêtes de plus de 4 Mio à la fois, les autres attendent ; l'unité systemd pose `MemoryHigh=75%` et `MemoryMax=85%` (PR 4) |
| 16 | Faible | Limites | Les points `/oauth/token` et `/oauth/revoke` n'avaient pas de limite | Soixante appels d'un coup par adresse, puis un par seconde (PR 4) |
| 17 | Faible | Limites | Une IA pouvait remplir le disque, partagé par tous les Cahuts, en publiant | Une publication qui laisserait moins de 1 Gio libre est refusée ; les quotas par Cahut restent après la V0 (PR 4) |
| 18 | Faible | Cloisonnement | Le programme de mise à jour, en root, suivait les liens du dossier des données | Il refuse tout ce qui n'est pas un fichier ordinaire dans un vrai dossier, et ouvre les fichiers sans suivre de lien (PR 4) |
| 19 | Faible | Cloisonnement | Tant que le serveur n'a pas d'administrateur, le premier démarrage de n'importe quel Cahut en fait un | `cahut setup` le dit, pour que le lien n'aille qu'à la bonne personne ; la règle de l'ADR 25 reste (PR 4) |

Deux durcissements, sans attaque connue, sont faits aussi (PR 3) : la console n'exécute que les scripts de `/console/assets/`, et les pages de Cahut que `/-/frame.js`, alors que les Windows publient des scripts à la même origine ; la garde des documents de client refuse aussi les plages IPv6 qui portent une adresse IPv4 (`::/96`, `::ffff:0:0:0/96`).

## Risque accepté pour la V0

**Une Window peut imiter la console.** Les Windows et la console partagent l'origine de leur Cahut ([ADR 10](adr/0010-windows-in-a-csp-sandbox.md)). Le bac à sable empêche le script d'une Window de lire la console, mais pas d'en copier l'apparence : une IA trompée par une injection de prompt peut publier, sans Knock, une Window de sa Room qui dit « votre session a expiré » et envoie le mot de passe tapé à une origine que sa déclaration nomme. Le lecteur voit la vraie adresse, en HTTPS.

Ce qui limite le risque en V0 : la Window doit déclarer l'origine où son formulaire envoie, et la console montre cette déclaration ; une Window publique passe par un Knock ; les personnes d'un Cahut sont peu nombreuses et se connaissent ; le Logbook garde qui a publié quoi.

Les parades, à trancher après la V0 dans une décision écrite :

- servir les Windows depuis un autre domaine que la console, ce que l'ADR 10 avait écarté pour éviter d'en demander deux ;
- entourer chaque Window, publique comprise, d'une barre de Cahut qui dit qu'une IA l'a faite et dans quelle Room ;
- refuser par défaut les champs de mot de passe dans les Windows, sauf déclaration.

## Après la V0

- La décision sur le risque accepté ci-dessus.
- Un audit par un tiers avant l'ouverture du dépôt, si les moyens le permettent.
- Les quotas par Cahut, prévus après la V0 : la PR 4 ne pose qu'un plancher d'espace libre pour tout le serveur.
