# Cahut · Plan du lot 8 : deuxième client IA et recette

23 octobre 2026 · Olivier Durand · Validé, achevé

## En bref

Le lot 8 clôt la V0. ChatGPT devient le deuxième client IA, de bout en bout : il se branche par OAuth, lit la house style, publie une présentation, la met à jour et la partage, comme Claude, par les mêmes outils. Le Chat de Mistral est essayé aussi, sans être un critère. Une revue de sécurité passe sur OAuth, le bac à sable des Windows, le cloisonnement entre Cahuts, les sessions de la console et des lecteurs, et les limites de taille et de débit ; ce qu'elle trouve se corrige dans le lot. Un test de recette rejoue les cinq critères du README sur un Cahut de test, en CI, et la recette sur deux VPS vierges est écrite pas à pas, chronomètre en main. Le README reçoit sa vraie section Installation. Le lot compte sept pull requests, pour environ huit jours à temps plein, et s'ouvre sur une décision écrite (ADR 29).

**Fin de lot**, reprise du [plan de la V0](plan-v0.md), **jalon V0** : les cinq critères passent sur deux VPS vierges, AlmaLinux et Ubuntu, avec deux clients IA.

## Point de départ

Ce que les lots 1 à 7 ont laissé :

- **Claude** se branche de trois façons : claude.ai et Claude Desktop par un connecteur personnalisé, Claude Code par `claude mcp add` ou par le plugin `cahut`, avec sa skill. Les trois s'identifient par leur Client ID Metadata Document (CIMD). Le jalon PoC est atteint depuis le lot 3 avec claude.ai.
- **OpenAI.** Le serveur OAuth de chaque Cahut suit déjà ce que ChatGPT demande ([ADR 16](adr/0016-oauth-for-ai-clients.md), [ADR 17](adr/0017-oauth-endpoints-on-port-443.md)) : le CIMD `https://chatgpt.com/oauth/client.json`, qui préfère `private_key_jwt` et offre aussi `none`, la seule méthode que Cahut prend ; `S256` ; le paramètre `iss` dans la réponse d'autorisation ; `resource` repris dans le jeton ; `offline_access` parmi les portées, pour les jetons de rafraîchissement. Codex a passé le test de fumée du lot 3. ChatGPT sur le web, dont le connecteur passe par les serveurs d'OpenAI, n'a pas encore été essayé.
- **Le Chat** n'a jamais été essayé. L'[ADR 6](adr/0006-in-house-oauth-server.md) note qu'il s'enregistre par enregistrement dynamique (DCR), et que les champs inconnus doivent être ignorés : Cahut les ignore.
- **Les outils.** Un administrateur qui peut écrire voit 31 outils ; `tools/list` pèse 87 Ko. Les schémas viennent des types Go ; chaque outil porte ses quatre indications (`readOnlyHint`, `destructiveHint` pour ce qui frappe d'abord, `idempotentHint`, `openWorldHint`), sans titre. Les instructions du serveur font 1 900 caractères ; Codex ne garantit que les 512 premiers.
- **Les limites.** Une requête de publication tient en `MaxRequestSize`, une Window en 500 fichiers et en `window.MaxSize` ; des limiteurs par adresse comptent les connexions, les enregistrements de clients et les Keys tapées, et un autre les Keys fausses de chaque Window.
- **L'installation** se fait d'une commande, avec une release signée, `cahut doctor` et un premier démarrage dans le navigateur ; un test d'installation tourne en CI et sous Lima. Le staging tourne en 0.4.0.
- **Restes du lot 7** : la démonstration chronométrée sur deux VPS vierges ; un certificat d'adresse IP jamais obtenu pour de vrai ; vérifier que claude.ai et ChatGPT acceptent un serveur MCP à une adresse IP ; Firefox et Safari.
- **Essais à l'ouverture du lot.** Contre un Cahut local et un jeton personnel, Codex 0.159 (modèle `gpt-6-astra` d'OpenAI) et Mistral Vibe 2.1 (Devstral 2) ont reçu la même demande : une présentation de trois diapositives à la house style, publiée dans la Room personnelle, puis ouverte par une Key. Tous deux ont lu la house style, publié et donné la Key, du premier coup et sans erreur. Codex a tronqué la liste des outils en la découvrant : 87 Ko, dont 61 de schémas de sortie.

## Décision à écrire d'abord

### ADR 29 · ChatGPT is the second AI client of the V0, through the same tools as Claude

L'[ADR 29](adr/0029-chatgpt-second-ai-client.md) tranche la question ouverte du plan de la V0 et fixe :

- **Le deuxième client** est ChatGPT sur le web, par un serveur MCP personnalisé : le plus répandu après Claude, et celui dont Cahut suit déjà les exigences d'OAuth. Depuis octobre 2026, l'aide d'OpenAI le place dans la page Plugins (« Add custom MCP server »), sans mode développeur, pour les offres Business, Enterprise et Edu ; les offres individuelles passaient jusque-là par le mode développeur. Le Chat est essayé dans le lot, et documenté s'il marche, sans être un critère de la V0.
- **Les mêmes outils pour tous les clients.** Aucun outil, schéma ou consigne n'est réservé à un client : ce qu'un client demande, tous le reçoivent.
- **Les instructions du serveur** disent d'abord, en moins de 512 caractères, ce qu'une IA doit savoir pour publier ; le reste suit.
- **Chaque outil** porte un titre lisible et ses quatre indications, et un schéma d'entrée, sans schéma de sortie : la liste des outils reste sous 32 Ko.
- **Ce qu'un client ne sait pas faire**, comme lire une skill ou des ressources MCP, les instructions et les descriptions des outils le portent.

## Les clients IA

| Client | Branchement | Qui l'essaie |
| --- | --- | --- |
| Claude Code | `claude mcp add`, puis `claude mcp login` ; CIMD | Claude, en local et sur le staging |
| claude.ai, Claude Desktop | Connecteur personnalisé ; CIMD | Olivier, sur le staging |
| ChatGPT | Plugins, « Add custom MCP server », à l'adresse MCP avec OAuth ; CIMD, méthode `none` | Olivier, sur le staging |
| Codex | `codex mcp add --url`, puis `codex mcp login` ; CIMD en bouclage | Claude, en local, avec les modèles d'OpenAI |
| Le Chat | Connecteur MCP personnalisé ; DCR | Olivier, sur le staging |
| Mistral Vibe | Serveur MCP en HTTP avec un jeton personnel | Claude, en local, avec les modèles de Mistral |

Codex et Vibe tournent sur l'ordinateur d'Olivier, avec ses comptes : ils éprouvent les schémas et les consignes avec les modèles d'OpenAI et de Mistral, sans passer par les serveurs de leurs éditeurs. ChatGPT, claude.ai et Le Chat sur le web passent par ces serveurs, qui demandent une adresse publique : ils s'essaient sur le staging.

Les ajustements, tirés des essais à l'ouverture du lot et de ce que publient OpenAI et Mistral :

- **Les instructions du serveur** réordonnées : l'essentiel pour publier dans les 512 premiers caractères, le reste après.
- **Un titre pour chaque outil**, en anglais simple (« Publish a Window ») : ChatGPT l'affiche quand il demande à la personne de confirmer un appel. Les opérations reçoivent un champ `Title`, qu'un test exige.
- **Une liste d'outils plus légère.** Les outils ne portent plus de schéma de sortie : le résultat reste structuré, et la référence de la documentation décrit toujours chaque sortie. Un test borne la taille de `tools/list` à 32 Ko pour un administrateur qui peut écrire.
- **Le pas à pas** de la console et la documentation pour ChatGPT (Plugins, puis « Add custom MCP server ») et pour Le Chat (un connecteur MCP personnalisé), avec ce que chacun demande et ce qu'il affiche.
- **Un Cahut à l'adresse IP du staging**, pour obtenir un vrai certificat d'adresse IP et savoir quels clients acceptent un serveur MCP à une adresse IP.

Les schémas d'entrée gardent leurs types multiples (`["null", "array"]`) : les modèles d'OpenAI et de Mistral les ont pris sans erreur.

## Revue de sécurité

La revue couvre cinq domaines, chacun lu par un relecteur indépendant qui cherche des attaques concrètes : qui attaque, ce qu'il envoie, ce qu'il gagne.

| Domaine | Ce qui est lu |
| --- | --- |
| OAuth et jetons des IA | PKCE, adresses de retour, codes et jetons de rafraîchissement, `private_key_jwt`, garde des CIMD et des JWKS, portée et Rooms d'une connexion, révocation, écran de consentement |
| Bac à sable des Windows | La politique CSP de chaque réponse d'une Window, les types, les chemins, les adresses à capacité, la page de la Key, la house style sous `/-/`, les aperçus de la console |
| Cloisonnement entre Cahuts | Le choix du Cahut par l'hôte et par SNI, les jetons et les sessions d'un Cahut sur un autre, les administrateurs, `rename_domain`, le lien du premier démarrage, la demande de mise à jour |
| Sessions et Knocks | Les sessions de la console et des lecteurs, la connexion, les liens d'activation, l'approbation des Knocks, le clickjacking, le texte des IA affiché dans la console |
| Limites de taille et de débit | Mémoire, processeur et disque à la portée d'un visiteur ou d'une IA, l'adresse du client, les limiteurs, les délais du serveur HTTP, les documents distants |

Chaque constat est vérifié, puis corrigé avec un test dans le lot, ou consigné avec sa raison quand il relève d'après la V0. La revue est écrite dans [revue-securite-v0.md](revue-securite-v0.md) : sa méthode, ce qui a été lu et trouvé sain, chaque constat et ce qui en a été fait. Les quotas par Cahut restent après la V0, comme le prévoit le plan.

## Recette

Les cinq critères du README, ramenés au périmètre de la V0 :

| Critère | Seuil | Comment on le vérifie |
| --- | --- | --- |
| Installation sur un VPS vierge | Moins de 10 minutes, une commande, sur AlmaLinux et sur Ubuntu | Chronomètre, de `curl … \| sudo sh` au lien du premier démarrage ; test d'installation en CI et sous Lima |
| Première publication par une IA | Moins de 5 minutes après le premier démarrage, sans toucher au serveur | Chronomètre, du lien du premier démarrage à l'adresse que rend l'IA |
| Partage par une personne non technique | Une présentation partagée par lien, avec une Key, en quelques minutes | Dans la console, sans terminal ; le lien s'ouvre en navigation privée avec la Key |
| IA prises en charge | Le même Cahut fonctionne avec Claude et ChatGPT | Chacun publie, met à jour et partage une Window du même Cahut |
| Gouvernance | Toute action d'une IA figure au Logbook, et s'annule ou a été approuvée | Le Logbook de la recette ; une règle vérifiée sur le registre |

**Le test de recette**, en Go, rejoue les critères 2 à 5 sur un Cahut de test, à chaque CI : le premier démarrage par son lien, deux clients OAuth qui se branchent comme ChatGPT (CIMD et `private_key_jwt`) et comme Le Chat (DCR), chacun publie une présentation à la house style par MCP, la personne lui donne une Key par l'API de la console, un lecteur l'ouvre avec la Key, l'IA demande à la rendre publique et frappe, la personne approuve. Le Logbook doit montrer chaque appel, au nom de la bonne IA et de la bonne personne.

**La règle de gouvernance** se vérifie sur le registre : chaque opération offerte aux IA ne change rien, frappe d'abord, ou a son retour en arrière, nommé dans le test (`publish_window` par `trash_window`, `update_window` par `rollback_window`, `trash_window` par `restore_window`…). Une opération nouvelle qui n'est dans aucun des trois cas fait échouer le test.

**La procédure de recette** est écrite dans [recette-v0.md](recette-v0.md) : le matériel, chaque étape et son chronomètre, ce qu'on doit voir, et un tableau des résultats que chaque passage remplit, avec la date, le système, la version et les temps.

## README

- La section **Installation** donne la commande, ce qu'elle fait, le premier démarrage, le branchement d'une IA et les mises à jour, avec les liens vers la documentation.
- L'avertissement du haut dit où en est Cahut : une V0 privée, ce qu'elle fait et ce qui viendra au lancement.
- Les lignes qui décrivent encore la console sur son propre port suivent l'[ADR 20](adr/0020-console-under-a-path-of-port-443.md).

## Découpage en pull requests

| # | Pull request | Contenu | Durée |
| --- | --- | --- | --- |
| 1 | Décisions | Ce plan et l'ADR 29 | 0,5 j |
| 2 | Clients IA | Les instructions, les titres des outils, la liste sans schémas de sortie, le pas à pas de ChatGPT et de Le Chat dans la console et la documentation, les essais avec Codex et Vibe | 1,5 j |
| 3 | Revue de sécurité : OAuth et sessions | Les corrections d'OAuth, des sessions de la console et des lecteurs, et des Knocks | 1,5 j |
| 4 | Revue de sécurité : bac à sable, cloisonnement et limites | Les corrections des Windows, du cloisonnement entre Cahuts et des limites ; `revue-securite-v0.md` | 1,5 j |
| 5 | Recette | Le test de recette, la règle de gouvernance, `recette-v0.md` | 1 j |
| 6 | README et documentation | La section Installation, l'état de la V0, les pages de la documentation, CLAUDE.md | 0,5 j |
| 7 | Staging et bilan | La version du lot sur le staging, un Cahut à son adresse IP, les essais avec les vrais clients, le bilan du lot et de la V0 | 1 j |

Soit environ 7,5 jours. Les pull requests s'enchaînent dans l'ordre ; les corrections de la revue se regroupent selon ce qu'elle trouve.

## Démonstration de fin de lot

**Prérequis** : une release qui porte le lot ; un VPS vierge sous AlmaLinux 10 et un sous Ubuntu 24.04, chacun avec un domaine qui pointe sur lui ; un compte claude.ai et un compte ChatGPT qui ajoute un serveur MCP personnalisé (offre Business, Enterprise ou Edu, ou mode développeur d'une offre individuelle s'il existe encore) ; une fenêtre de navigation privée.

Sur chaque VPS, en suivant [recette-v0.md](recette-v0.md) :

1. Chronomètre en main, `curl -fsSL https://cahut.cloud/install.sh | sudo sh -s -- --domain <domaine>` : la signature vérifiée, le pare-feu ouvert, `cahut doctor` sans échec, le lien du premier démarrage. Moins de 10 minutes.
2. Ouvrir le lien : créer son compte, nommer l'organisation, choisir une couleur, brancher claude.ai. Demander à Claude une présentation : elle s'ouvre à l'adresse qu'il rend, moins de 5 minutes après l'ouverture du lien.
3. Sans terminal, dans la console : donner une Key à la présentation et copier son lien. Il s'ouvre en navigation privée avec la Key.
4. Brancher ChatGPT au même Cahut : il lit la house style, met à jour la présentation, puis revient à la version précédente.
5. Demander à ChatGPT de rendre la présentation publique : il frappe, la console montre le Knock, l'approbation l'exécute.
6. Lire le Logbook : chaque appel des deux IA y figure, et chacun s'annule ou a été approuvé.

## Hors du lot 8

- Le Chat, Copilot, Gemini et les autres clients comme critères : après la V0. Le Chat est essayé, sans engagement.
- Un audit de sécurité par un tiers : la revue du lot est faite par le projet lui-même ; un audit extérieur reste à décider avant l'ouverture du dépôt.
- Les quotas par Cahut, les sauvegardes, l'envoi d'e-mails : après la V0, comme le prévoit son plan.
- La release qui porte le lot : Olivier décide quand l'étiqueter et la publier.

## Bilan

Le lot 8 s'est achevé le 23 octobre 2026, en sept pull requests :

| # | Pull request | Contenu |
| --- | --- | --- |
| 1 | #86 | Le plan et l'ADR 29 |
| 2 | #87 | Un titre pour chaque outil, une liste d'outils sans schémas de sortie, les instructions réordonnées, ChatGPT et Le Chat dans la console et la documentation |
| 3 | #88 | La revue de sécurité, et ses corrections de la connexion, des sessions, d'OAuth et du Logbook |
| 4 | #89 | Les corrections des limites, du programme de mise à jour et du premier démarrage |
| 5 | #90 | Le test de recette, la règle de gouvernance sur le registre, la procédure de recette |
| 6 | #91 | La section Installation du README, l'état de Cahut dans la documentation |
| 7 | #92 | Le staging, le site, le bilan |

**Écarts par rapport au plan.**

- **ChatGPT.** L'aide d'OpenAI a changé en octobre 2026 : un serveur MCP personnalisé s'ajoute depuis la page Plugins, « Add custom MCP server », sans mode développeur, et l'aide ne cite plus que les offres Business, Enterprise et Edu. Le plan et l'ADR 29 ont été corrigés avant leur fusion. Le document de ChatGPT préfère `private_key_jwt` et offre `none`, la seule méthode que Cahut prend.
- **Les schémas d'entrée** n'ont pas changé : les modèles d'OpenAI et de Mistral ont pris les types multiples sans erreur. La liste des outils passe de 87 à 27 Ko sans les schémas de sortie ; Codex la tronque encore à la découverte, puis lit les définitions dont il a besoin, sans échec.
- **La revue** a trouvé deux constats graves, sept moyens et dix faibles, et deux durcissements. Tous sont corrigés avec un test, sauf un risque de conception accepté pour la V0 : une Window peut imiter la connexion de la console, sur la même adresse ([revue-securite-v0.md](revue-securite-v0.md)). Les corrections ont pris deux pull requests, comme prévu.
- **Le registre** gagne une étape `Prepare`, avant la transaction, que le plan ne prévoyait pas : la connexion et la lecture des documents des clients OAuth y passent, pour ne plus tenir la seule connexion d'écriture d'un Cahut.
- **Trois migrations** : `cahut/0012` et `server/0008` gardent l'adresse des connexions au Logbook ; `cahut/0013` refait les sessions et les passes avec `AUTOINCREMENT`.
- **`oauth.Server.SetDocumentTransport`** sert aux tests d'autres paquets, qui servent un document de client sur une adresse de bouclage.
- **Le site** comptait encore les lots 0 à 6 comme faits : sa feuille de route compte maintenant les lots 0 à 8.
- Pas de nouvelle dépendance.

**Essais.**

- **Clients locaux, à l'ouverture du lot.** Codex 0.159 (`gpt-6-astra`) et Mistral Vibe 2.1 (Devstral 2), contre un Cahut local et un jeton personnel : chacun a lu la house style, publié une présentation et l'a ouverte par une Key, du premier coup.
- **Tests.** La revue a laissé un test par constat corrigé ; `TestV0Acceptance` rejoue les critères 2 à 5 avec un client par document (comme ChatGPT) et un client par enregistrement dynamique (comme Le Chat) ; `TestEveryAIOperationIsReversibleOrApproved` tient la règle de gouvernance sur le registre. Le lint passe pour macOS et pour Linux.
- **Navigateur.** Le pas à pas de ChatGPT et de Le Chat à 1 280 px et 375 px, où les cinq onglets défilent dans leur barre ; le bouton « Approve » d'un Knock, inactif à l'ouverture de la page, actif après un moment au premier plan et un mouvement du pointeur.
- **Staging.** La version du lot est installée (0.4.1-0.20261023153447-190b2860fe47) par le nouvel `install.sh`, avec une sauvegarde dans `/var/lib/cahut.bak-2026-10-23-batch8` ; l'unité porte `MemoryHigh` et `MemoryMax`, et `cahut doctor` ne trouve aucun problème. Le connecteur de Claude Code de la session se rebranche seul.
- **Un Cahut à l'adresse IP, pour de vrai.** Un troisième Cahut est déclaré à l'adresse du staging, `51.83.68.4` : Let's Encrypt lui a délivré un certificat d'adresse IP, au profil `shortlived`, valable du 23 au 30 octobre, que curl valide. Par Internet, avec des jetons personnels révoqués ensuite, Codex y a publié une présentation à la house style, l'a mise à jour, l'a ouverte par une Key, puis a demandé à la rendre publique : la demande a frappé (Knock 1). Vibe y a listé les Windows, publié un document et ramené la présentation à sa version précédente. Le Logbook nomme chaque appel, son client et sa personne. La présentation s'ouvre avec sa Key, sous une politique CSP qui nomme `https://51.83.68.4`.

**Release 0.5.0.** Étiquetée le 23 octobre au commit 6c25197, après une CI verte. Le workflow des releases a préparé le brouillon : les archives pour amd64 et arm64, leurs sommes et la signature. Leur signature est vérifiée avec la clé d'`install.sh`, et leurs sommes avec `shasum`. Les binaires sont statiques, avec la console, et `cahut version` y dit `cahut 0.5.0 (commit 6c25197, 2026-10-23T15:48:39Z, go1.27.3 linux/amd64)`. Le brouillon a son titre et ses notes ; Olivier l'a publié le 24 octobre. Le même jour, sur le staging, après une sauvegarde dans `/var/lib/cahut.bak-2026-10-24-v0.5.0`, `sudo cahut update` a lu la release, vérifié sa signature et sa somme, et l'a installée. Le binaire installé est celui de l'archive, octet pour octet ; `cahut doctor` ne trouve aucun problème, avec trois Cahuts, et le Logbook du serveur note la mise à jour.

**Reste à faire.**

- **Le jalon V0** : la recette sur deux VPS vierges, AlmaLinux et Ubuntu, avec claude.ai et ChatGPT, en suivant [recette-v0.md](recette-v0.md). Elle demande les VPS, leurs domaines, et les comptes d'Olivier, dont un ChatGPT qui ajoute un serveur MCP personnalisé.
- Brancher claude.ai et ChatGPT au staging, et à `https://51.83.68.4/mcp`, pour savoir s'ils acceptent un serveur MCP à une adresse IP ; essayer Le Chat. Le Knock 1 du Cahut à l'adresse IP attend une réponse jusqu'au 24 octobre ; ce Cahut a une personne, `olivier`, dont le lien d'activation n'a pas servi, et rien ne supprime encore un Cahut.
- La décision sur le risque accepté par la revue, après la V0.
- Toujours : Firefox et Safari ; ranger la clé privée des releases dans le gestionnaire de mots de passe d'Olivier.

