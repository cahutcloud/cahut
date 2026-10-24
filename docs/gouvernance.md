# Cahut · Gouvernance et vie du projet

Ce document propose comment Cahut est dirigé et comment il vit au quotidien : qui décide, comment on contribue, où l'on se parle, quel soutien attendre, et ce que deviennent le nom, la licence et l'argent. Il part d'une réalité, une seule personne porte le projet, et prépare la suite sans la présumer. Il s'appuie sur ce que des projets open source comparables ont appris, souvent à leurs dépens. Les faits cités sont datés et sourcés à la fin.

## En bref

- **Un modèle : ouvert au regard, encadré pour la contribution, honnête sur le temps.** Olivier décide et écrit ses décisions, comme les ADR le font déjà. Tout le monde peut lire, installer, signaler, proposer. Les corrections sont bienvenues directement ; les fonctionnalités se discutent avant d'être codées. Le projet dit ce qu'il peut tenir : une réponse dans la semaine, pas un support.
- **Trois promesses écrites avant l'ouverture du dépôt**, parce qu'elles coûtent cher à prendre plus tard : la licence ne changera pas, il n'y aura pas de version « pro » fermée, le nom est protégé mais le code est libre.
- **Des phases déclenchées par des faits, pas par des dates.** Seul jusqu'à l'ouverture du dépôt, puis les premiers utilisateurs, puis les premiers contributeurs réguliers, puis une équipe de mainteneurs avec des règles de décision collectives et, si l'argent arrive, une structure pour porter le nom et le domaine.
- **Un seul endroit par usage.** GitHub pour le code, les tickets, les discussions et les releases ; cahut.cloud pour expliquer le projet et donner des nouvelles ; docs.cahut.cloud pour la documentation. Pas de chat tant qu'une seule personne devrait le tenir.
- **Les IA sont bienvenues, les humains répondent.** Cahut est construit avec Claude Code. Une contribution assistée par une IA est acceptée si la personne qui la signe la comprend, l'a testée et le dit.
- **Ce qu'il reste à décider** est à la fin. La licence est tranchée depuis le 22 octobre 2026 : AGPL-3.0 ([ADR 15](adr/0015-license.md)). Restent le moment d'ouvrir le dépôt, les canaux, la marque, l'argent et la personne de confiance qui garde une clé de secours.

## Le point de départ

Ce que le dépôt a déjà, et qui est déjà de la gouvernance :

- Les décisions d'architecture sont écrites, numérotées, publiques et jamais réécrites ([docs/adr](adr/README.md)). C'est le registre des décisions du projet ; il suffit de l'étendre aux décisions de produit et de périmètre.
- Les releases sont construites par CI, relues et publiées à la main ([ADR 7](adr/0007-module-path-repository-and-releases.md)) ; les dépendances et les actions sont épinglées et suivies par Dependabot.
- [CLAUDE.md](../CLAUDE.md) fixe les conventions de code, de langue et de style : c'est un guide de contribution, écrit pour les IA, qui vaut aussi pour les humains.
- Le site et la documentation vivent dans le dépôt, avec une page « Where Cahut stands » qui dit ce qui marche et ce qui vient.
- L'organisation GitHub `getcahut` et le domaine cahut.cloud existent, et le chemin de module `cahut.cloud` ne dépend d'aucune forge.
- La licence est choisie : AGPL-3.0 pour tout le dépôt, le nom et le logo à part ([ADR 15](adr/0015-license.md)), avec le texte dans LICENSE et son sens en clair dans le README.
- Les fichiers communautaires : GOVERNANCE, CONTRIBUTING, CODE_OF_CONDUCT, SECURITY, SUPPORT et TRADEMARK, les gabarits de tickets et de pull requests et FUNDING, écrits à partir de ce document le 22 octobre 2026.

Ce qui manque :

- **Les Discussions** sont désactivées, le signalement privé des failles n'est pas activé, les boîtes `security@`, `conduct@` et `contact@cahut.cloud` n'existent pas encore, aucune marque n'est déposée.
- **Le site renvoie vers un dépôt privé** : le bouton GitHub de cahut.cloud n'ouvre rien pour un visiteur.

La contrainte qui commande tout : une personne. Nadia Eghbal (*Working in Public*, 2020) classe les projets selon leurs utilisateurs et leurs contributeurs. Les projets à succès sont le plus souvent des « stades » : beaucoup d'utilisateurs, très peu de contributeurs, un mainteneur. Cahut en sera un longtemps. La gouvernance se conçoit donc pour ce cas : écrire une fois ce qui sera demandé cent fois, automatiser ce qui peut l'être, protéger le temps du mainteneur. Mike McQuaid (Homebrew) le dit sans détour : la licence dit « as is », le mainteneur ne doit rien ; l'épuisement vient de se comporter en service client.

## Ce qu'on apprend des autres

| Projet | Comment il vit | Ce qu'on en retient pour Cahut |
| --- | --- | --- |
| **PocketBase** (Go, SQLite, un binaire, une personne) | Projet personnel, sans entreprise ni association ; le README prévient que la compatibilité n'est pas garantie avant la 1.0 et que les fonctionnalités se discutent d'abord ; depuis avril 2026, les pull requests extérieures sont fermées à cause du spam généré par des IA ; une bourse promise en 2025 a échoué sur la paperasse, et l'auteur vit sur ses économies | Dire ce que le projet est et ce que 0.x veut dire ; sans structure, l'argent est fragile |
| **Miniflux** (Go, une personne) | Logiciel « opiniâtre », avec une page qui le dit : périmètre volontairement petit, demandes non garanties, une seule base de données prise en charge faute de temps pour en tester d'autres ; l'auteur vend une version hébergée à 15 $ par an depuis 2013, avec le même code ; dons par Liberapay | Le non est une pratique normale, écrite une fois pour toutes ; une offre hébergée par l'auteur est compatible avec l'open source |
| **Litestream** (Go, Ben Johnson) | « Open source mais fermé aux contributions » en janvier 2021, après l'épuisement de l'auteur sur BoltDB ; ouvert aux correctifs neuf mois plus tard ; depuis 2025, un CONTRIBUTING écrit : petites corrections bienvenues, grosses fonctionnalités « généralement pas acceptées », aide d'une IA admise ; l'auteur est payé par Fly.io pour y travailler depuis 2022 | On peut ouvrir le code sans ouvrir le développement, et le dire ; la règle tient quatre ans plus tard |
| **SQLite** | « Open source, pas open contribution » : pas de patch sans déclaration de domaine public | Le modèle le plus fermé qui marche ; Cahut n'ira pas jusque-là, mais la formule libère |
| **Caddy** (Go, Matt Holt) | Vit de sponsors (paliers GitHub Sponsors jusqu'à plusieurs milliers de dollars par mois, sans fonctionnalité payante) ; les contributeurs réguliers reçoivent le droit de pousser, seuls les mainteneurs autorisés publient ; en septembre 2017, un en-tête publicitaire ajouté aux binaires gratuits a provoqué une levée de boucliers, annulée en deux jours avec des excuses | Tout ce qui touche l'argent se fait à ciel ouvert, jamais dans le binaire |
| **Uptime Kuma** (une personne, puis une petite équipe) | Louis Lam garde le dernier mot ; des mainteneurs « juniors » fusionnent les petites corrections et gèrent les tickets, les fonctionnalités passent par un ticket d'abord ; en 2025, un mainteneur écrit que le temps de relecture est à sec depuis des mois et que les grosses pull requests d'interface ne seront pas accompagnées | Des rôles gradués soulagent ; le goulot reste le temps de relecture, et il faut le dire |
| **Vaultwarden** (Rust, AGPL) | Serveur compatible Bitwarden, renommé en avril 2021 (ex-bitwarden_rs) pour couper court à la confusion et aux problèmes de marque ; un mainteneur salarié de Bitwarden contribue sur son temps libre, relu par les autres | Une marque se respecte dans les deux sens ; un conflit d'intérêts se déclare et se relit |
| **Immich** (AGPL) | Né d'une personne, rejoint FUTO en mai 2024 : une dizaine de personnes payées, des promesses écrites (AGPL maintenue, pas de CLA, pas de fonctionnalité payante, pas de publicité, autonomie de l'équipe) ; le prix : la marque et le logo sont passés à FUTO, les dons ont cessé ; deux ans après, l'équipe dit que les promesses ont tenu ; les pull requests générées par des IA sont refusées | Si un financeur arrive, les promesses s'écrivent avant, et le nom a un prix |
| **Home Assistant** | Nabu Casa, l'entreprise fondée en 2018 par les développeurs, vend le cloud et le matériel et reverse l'essentiel de ses profits ; l'Open Home Foundation, fondation suisse créée en avril 2024, détient les projets pour les protéger d'un rachat ou d'un abandon ; les IA sont admises si l'auteur comprend tout, les agents autonomes refusés | Séparer celui qui vend de celui qui possède le commun |
| **Mastodon** (AGPL) | Des structures à but non lucratif portent le projet, financé par des dons ; politique de marque publique, mise à jour en août 2026 ; le fondateur a quitté la direction en novembre 2025, contre une somme unique pour la marque et son travail passé ; l'entité européenne promise en janvier 2025 n'était toujours pas créée en février 2026 | Une structure et une politique de marque permettent au fondateur de s'effacer ; une structure prend plus de temps qu'annoncé |
| **Plausible** (AGPL) | Une entreprise de dix personnes, autofinancée, dont le seul revenu est le cloud hébergé en Europe ; la Community Edition sort deux fois par an, sans support, et quelques fonctionnalités (entonnoirs, SSO) restent réservées au cloud ; les fonctionnalités se discutent d'abord | Un modèle qui marche, mais pas celui de Cahut : Cahut promet le même binaire pour tous |
| **Ghost** (MIT) | Une fondation à but non lucratif, sans actionnaire, dont les statuts interdisent la revente ; financée à 100 % par l'hébergement Ghost(Pro), comptes publics en direct ; les fonctionnalités se discutent sur le forum avant ; un CLA | Le modèle le plus lisible pour « indépendant et durable » |
| **Syncthing** (MPL) | Une fondation suédoise minuscule paie le site et les serveurs, reçoit les dons et peut financer une fonctionnalité ; DCO ; pas de pull request non sollicitée d'un nouveau venu, pas de code généré ; en octobre 2024, l'application Android officielle a été arrêtée faute de motivation, et dit comme tel | Une structure peut rester minuscule ; arrêter une pièce du projet est un droit, qui se dit |
| **Gitea et Forgejo** | En octobre 2022, le mainteneur principal de Gitea a transféré les domaines et la marque à une entreprise créée pour vendre du support et de l'hébergement, sans que les contributeurs le sachent ; une lettre ouverte de 49 contributeurs a demandé une association à but non lucratif ; Forgejo est né sept semaines plus tard, porté par Codeberg e.V., une association berlinoise dont les membres élisent le bureau | Qui tient le nom et le domaine tient le projet : le dire, et prévoir où ils iront |
| **Paperless-ngx** | Paperless abandonné, Paperless-ng abandonné à son tour avec 180 tickets et 40 pull requests en attente et un mainteneur injoignable ; en février 2022, la communauté a créé une organisation GitHub avec des équipes « pour éviter le point de défaillance unique », et publié cinq semaines plus tard | Prévoir sa propre absence évite un fork douloureux |
| **listmonk** (Go, AGPL) | Écrit par le directeur technique de Zerodha, en partie pour les besoins de l'entreprise ; toute pull request reste à la discrétion des mainteneurs, proposition en ticket d'abord, pas de Discussions ni de chat ; un guide qui dit que les mainteneurs ne doivent rien ; gel des pull requests sur l'interface pendant sa réécriture | Un projet personnel dure si le périmètre tient et si les règles sont écrites |
| **OpenClaw** (MIT) | Projet de week-end de novembre 2025, viral en janvier 2026 : près de 2 800 pull requests et autant de tickets en janvier, plus de 5 000 de chaque dans la première quinzaine de février ; l'auteur a écrit qu'il ne pouvait plus maintenir seul, ajouté des mainteneurs, plafonné les pull requests ouvertes par auteur (vingt), renvoyé les fonctionnalités vers des plugins, fait trier par des agents IA en lecture seule, puis rejoint OpenAI ; une fondation financée par des donateurs porte le projet depuis juillet 2026 | Le succès soudain arrive avant la gouvernance : les fichiers, les règles et les soupapes doivent exister avant |

Les leçons qui reviennent :

1. **Dire tôt ce que le projet est** : personnel, opiniâtre, open source sans être ouvert à tout.
2. **L'argent ne trahit pas si le même code sert tout le monde** et si rien n'est caché dans le binaire.
3. **Qui tient le nom, le domaine et l'organisation tient le projet.** Ce point se règle par écrit, avant que l'argent ou le succès n'arrivent.
4. **Un changement de licence après coup finit en fork** : HashiCorp (2023, OpenTofu), Redis (2024, Valkey), Elastic (2021, OpenSearch). Ce qui l'a rendu possible, c'est le CLA signé par les contributeurs. Redis et Elastic sont revenus à l'AGPL, en reconnaissant le mal fait.
5. **La confiance se gagne, elle ne se cède pas sous pression.** La porte dérobée de xz (2024) a été posée par un « contributeur » devenu mainteneur après deux ans de pression sur un mainteneur seul et fatigué.
6. **Les règles sur l'IA se décident avant le déluge** : curl a fermé son programme de primes en janvier 2026 sous les faux rapports générés ; PocketBase a fermé ses pull requests extérieures en avril 2026 pour la même raison ; Ghostty, Gitea, Forgejo, Zig, Immich, Syncthing et le noyau Linux ont chacun écrit leur politique en 2025 ou 2026.

## Les principes

Ils sont peu nombreux, et ils s'écrivent dans GOVERNANCE.md pour que chacun puisse s'y référer.

1. **Le mainteneur décide, et écrit pourquoi.** Olivier a le dernier mot (le modèle « BDFL » que l'[Open Source Guide](https://opensource.guide/leadership-and-governance/) donne par défaut aux petits projets). Chaque décision qui compte devient un ADR, lisible par tous, avec les options écartées.
2. **Un non est provisoire, un oui est pour toujours.** Une fonctionnalité acceptée se maintient des années ; une fonctionnalité refusée peut revenir avec de meilleurs arguments. Le périmètre du README et les ADR disent où est la limite.
3. **Ouvert au regard, encadré pour la contribution.** Tout se lit, tout se teste, tout se signale. Les corrections entrent directement ; les fonctionnalités se discutent d'abord ; certaines choses ne se feront pas.
4. **Le temps est la ressource rare, et c'est dit.** Le projet publie ce qu'il peut tenir : un créneau par semaine, un accusé de réception dans la semaine, pas d'engagement de délai au-delà.
5. **Honnête sur l'état du logiciel.** Le numéro de version, la page « Where Cahut stands » et les notes de release disent ce qui marche, ce qui casse et ce qui change.
6. **La licence ne change pas, et il n'y a pas de version fermée.** Le code sert tout le monde de la même façon, auto-hébergé ou hébergé par quelqu'un d'autre. Aucun fichier ne demande de céder ses droits.
7. **Le nom est protégé, le code est libre.** « Cahut » et le logo sont une marque ; le code et la documentation sont libres. Un fork prend un autre nom.
8. **La sécurité passe avant tout le reste.** Un rapport de faille interrompt le reste ; il a un canal privé, un délai de réponse et une publication coordonnée.
9. **Rien ne sort du serveur en cachette.** La seule requête sortante du logiciel est la vérification des releases sur GitHub, documentée et désactivable. Pas de télémétrie.
10. **Les IA sont bienvenues, les humains répondent.** Une contribution assistée par une IA se déclare, se comprend et se teste par la personne qui la signe.
11. **Le socle ne se vote pas.** Les règles de produit du README, opérations bornées, *knock first*, une seule API, pas d'IA sur le serveur, tout est tracé, isolation, et les points 6 à 9 ci-dessus ne changent qu'avec l'accord du mainteneur principal et de tous les mainteneurs, par un ADR.

## Les rôles

| Rôle | Qui | Ce qu'il peut faire | Comment on y entre | Comment on en sort |
| --- | --- | --- | --- | --- |
| Utilisateur | Toute personne qui installe Cahut ou lit une Window | Lire, signaler, proposer, répondre aux autres | En arrivant | |
| Contributeur | Toute personne dont une contribution a été retenue : code, documentation, traduction, aide répétée au tri | Tout ce qui précède ; cité dans les notes de release | Une pull request fusionnée ou une aide suivie | |
| Trieur | Un contributeur régulier qui aide à classer | Étiqueter, fermer les doublons, demander des précisions, répondre dans les Discussions (droit *triage* de GitHub) | Invité par le mainteneur principal après quelques mois d'aide | Inactif six mois : retiré, avec merci |
| Mainteneur | Une personne qui relit et fusionne | Relire, fusionner, publier une release, voix dans les décisions ; nommé dans MAINTAINERS.md | Invité après six mois de contributions suivies et la confiance du mainteneur principal (droit *write*) | Peut se retirer et devenir émérite ; inactif un an : émérite |
| Mainteneur principal | Olivier | Tout ce qui précède ; tranche ; détient le nom, le domaine et l'organisation GitHub | Fondateur | Prévoit sa succession, ci-dessous |

Le rôle de trieur est la première marche : il coûte peu de confiance, soulage vite, et révèle qui tient dans la durée. Un trieur ne modifie pas le code ; un mainteneur, si. Le passage de l'un à l'autre se fait sur invitation, jamais sur demande insistante. Uptime Kuma fait de même avec des mainteneurs « juniors », qui fusionnent les petites corrections et gèrent les tickets, et des « seniors », qui décident des fonctionnalités.

**La personne de confiance.** Un projet d'une seule personne meurt avec elle, ou s'arrête quand elle décroche. La parade tient en une mesure : une personne de confiance, désignée dès l'ouverture du dépôt, qui n'a pas besoin de savoir coder. Elle est second propriétaire de l'organisation GitHub, a un accès d'urgence au registrar du domaine et à la clé de signature des releases (par exemple l'accès d'urgence d'un gestionnaire de mots de passe), et détient une note scellée qui dit quoi faire. Si Olivier reste injoignable six mois, ou sur sa demande, elle publie un avis dans le dépôt, nomme un mainteneur parmi les contributeurs connus, ou transfère le projet à une structure, selon ce qu'Olivier aura écrit d'avance. Paperless-ngx a dû se reconstituer de l'extérieur ; mieux vaut une clé sous le paillasson que de forcer la porte.

**Ce que la confiance exige.** Le précédent de xz impose quelques règles mécaniques, recommandées par l'OpenSSF après l'incident : double authentification obligatoire sur l'organisation, protection de la branche `main` (CI verte, pas de poussée directe), aucun droit accordé sous pression ni à une personne inconnue, une période d'observation avant chaque promotion, et des releases signées (prévues au lot 7). Dès qu'il y a deux mainteneurs, un changement qui touche la sécurité ou les releases se relit à deux.

## Les décisions

- **Au quotidien**, le mainteneur décide dans les tickets et les pull requests, et dit pourquoi en une phrase.
- **Ce qui engage l'architecture, le produit ou le vocabulaire** passe par un ADR. Le processus existe ; il s'étend aux décisions de produit. Un ADR peut être proposé par n'importe qui, en pull request, au statut *Proposed*. Il reste ouvert aux commentaires une à deux semaines s'il change quelque chose d'important, puis le mainteneur principal l'accepte ou le refuse, avec les options écartées. C'est un processus de *RFC* sans le cérémonial.
- **La feuille de route est publique et sans dates** : les lots du [plan de la V0](plan-v0.md), repris en jalons GitHub (« Batch 2 · Accounts, Rooms and console »), et la page « Where Cahut stands » de la documentation. Les idées se votent dans les Discussions ; un vote renseigne, il ne décide pas.
- **Quand il y a plusieurs mainteneurs** (phase 3), les décisions ordinaires se prennent par consensus tacite : une proposition reste ouverte sept jours ; sans objection, elle passe. Une objection dit pourquoi et propose une issue. Le mainteneur principal tranche les désaccords. Chaque mainteneur déclare ses conflits d'intérêts, par exemple s'il travaille pour un hébergeur.
- **Le socle** (principe 11) ne change qu'à l'unanimité des mainteneurs, par un ADR.

## Les contributions

### Ce qui est bienvenu tout de suite

Une correction de bug avec son test, une amélioration de la documentation, une faute dans le site, une traduction du site, un message d'erreur plus clair, une dépendance à jour. Pas besoin de demander : ouvrez la pull request.

### Ce qui se discute d'abord

Une fonctionnalité, une nouvelle opération ou un nouvel outil MCP, un changement d'interface ou de vocabulaire, une nouvelle dépendance, tout ce qui touche la sécurité, le bac à sable ou *knock first*. Ouvrez une Discussion (catégorie *Ideas*) ou un ticket avant de coder : une pull request de mille lignes qui arrive par surprise sera fermée gentiment, quelle que soit sa qualité. C'est la règle de PocketBase, de Litestream et de Miniflux, et elle protège les deux côtés.

### Ce que Cahut refusera

Le README le dit déjà : créer du contenu, faire tourner un modèle sur le serveur, donner un shell à une IA, dépendre d'un fournisseur, s'étendre sur plusieurs machines. S'y ajoutent : une version réduite ou une version fermée, de la télémétrie, une fonctionnalité réservée à une offre payante. Une page « Why Cahut doesn't… » dans la documentation rassemblera ces refus avec leurs raisons, comme Miniflux tient sa page de logiciel opiniâtre : chaque refus renvoie à un ADR ou au README, et la question ne se repose plus.

### Le parcours d'une pull request

1. Une branche par sujet, une pull request par sujet, petite de préférence.
2. `make check` vert ; les conventions de [CLAUDE.md](../CLAUDE.md) (anglais, messages à l'impératif, pas de logique métier dans les adaptateurs, pas de dépendance sans raison).
3. La description dit ce qui change et pourquoi, et comment ça a été testé.
4. Le mainteneur relit dans son créneau hebdomadaire ; il demande des changements, fusionne, ou explique pourquoi non.
5. Une pull request sans réponse de son auteur pendant trente jours est fermée, réouvrable à tout moment.

### La signature des commits

Cahut demande le [Developer Certificate of Origin](https://developercertificate.org/) (DCO) : chaque commit porte `Signed-off-by: Nom <adresse>` (`git commit -s`), par lequel l'auteur certifie avoir le droit de contribuer ce code sous la licence du projet. Ce n'est pas un contrat : rien n'est cédé, aucun document à signer, pas d'avocat. C'est ce que font le noyau Linux, GitLab (depuis 2017, à la demande de Debian), Gitea, OpenTofu et Valkey. L'application GitHub [DCO](https://github.com/apps/dco) vérifie les pull requests ; le réglage « Require contributors to sign off on web-based commits » couvre les commits faits depuis l'interface web.

Pas de CLA. Un CLA (*Contributor License Agreement*) est ce qui a permis à HashiCorp, Redis et Elastic de changer de licence contre leurs contributeurs. Caddy, Ghost et Home Assistant en demandent un et restent dignes de confiance, mais une fondation ou une entreprise connue porte leur promesse ; un projet d'une personne n'a que ses actes. Sans CLA, chaque contributeur garde ses droits, et la promesse « la licence ne change pas » devient structurelle : dès que du code extérieur est entré, personne, Olivier compris, ne peut plus changer la licence sans l'accord de tous. Immich l'a écrit noir sur blanc en rejoignant FUTO : AGPL maintenue, pas de CLA.

### Les contributions assistées par une IA

Cahut est écrit avec Claude Code, et [CLAUDE.md](../CLAUDE.md) est public : une interdiction serait hypocrite. Forgejo, Zig, Immich et Syncthing interdisent tout code généré, et PocketBase a fermé ses pull requests extérieures ; curl, Gitea, Ghostty, Caddy, Home Assistant, LLVM, Fedora, OpenClaw et le noyau Linux l'acceptent sous conditions. Cahut suit les seconds, avec leurs règles :

1. **Dites-le.** L'outil et son étendue, dans la pull request ou en pied de commit (`Assisted-by: Claude Code`, le trailer qu'adoptent Fedora et LLVM).
2. **Comprenez tout.** La personne qui signe a lu, compris, exécuté et testé ce qu'elle propose, et répond elle-même en relecture. Une réponse générée à une relecture ferme la pull request.
3. **La signature est la vôtre.** Le `Signed-off-by` certifie ce que vous avez fait, pas ce que l'outil a fait : un agent ne l'ajoute que sur votre instruction et en votre nom, jamais de lui-même (règle du noyau Linux), et aucun agent n'ouvre de ticket ni de pull request sans qu'une personne l'ait relu.
4. **Reproduisez avant de signaler.** Un ticket écrit par une IA sans reproduction est fermé. Un rapport de sécurité inventé vaut un bannissement, comme chez curl.
5. **Les visuels viennent de la charte.** Pas d'image ni d'illustration générée dans le site ou la documentation.

Le mainteneur peut utiliser une IA pour résumer, dédoublonner ou préparer une réponse. Il lit et décide lui-même, et le dit de la même façon. Si le volume déborde (le cas d'OpenClaw), GitHub permet depuis 2026 de limiter les pull requests et les tickets aux collaborateurs ou de plafonner les pull requests ouvertes par personne : ce sont des soupapes, pas un mode normal.

Conséquence à connaître : la fondation NLnet, qui finance beaucoup de logiciels libres européens, interdit depuis octobre 2026 le code et la documentation produits par une IA générative dans les travaux qu'elle finance, et demande de déclarer un dépôt qui en contient déjà. Cahut n'est donc pas un candidat pour elle, et l'assume.

### La reconnaissance

Les notes de chaque release nomment les contributeurs. GitHub tient la liste. Les étiquettes `good first issue` et `help wanted` disent par où commencer. Pas de programme de primes : il attire les rapports générés, comme curl l'a mesuré (moins de 5 % de failles réelles en 2025, programme fermé en janvier 2026).

## Les tickets et le tri

### Tickets ou Discussions

| Où | Pour quoi |
| --- | --- |
| Issues | Un bug reproductible, ou un travail accepté. Un ticket est quelque chose que quelqu'un va faire |
| Discussions · Q&A | Une question : installation, configuration, « est-ce normal ? » |
| Discussions · Ideas | Une fonctionnalité à proposer, avec les votes pour mesurer l'intérêt ; acceptée, elle devient un ticket |
| Discussions · Show and tell | Ce que les gens publient avec Cahut |
| Discussions · Announcements | Les releases et les nouvelles, écrites par le mainteneur |
| Signalement privé | Une faille de sécurité, jamais dans un ticket |

Les Discussions sont la base de connaissances du projet : elles se cherchent, elles se lient, elles ne demandent aucun logiciel propriétaire. C'est l'une des raisons de ne pas ouvrir de chat au début.

### Les gabarits

Des formulaires (`.github/ISSUE_TEMPLATE/*.yml`), courts : un pour les bugs (version de `cahut version`, système, méthode d'installation, ce qui était attendu, ce qui s'est passé, extrait du Logbook sans secret), un pour les fonctionnalités acceptées, et un `config.yml` qui renvoie les questions vers les Discussions et les failles vers le signalement privé. Un gabarit de pull request rappelle le DCO, les tests et la déclaration d'IA.

### Les étiquettes

Un jeu court, en anglais, posé dès l'ouverture : le type (`bug`, `feature`, `docs`, `site`, `chore`), l'état (`needs-triage`, `needs-info`, `confirmed`, `accepted`, `blocked`), la zone (`area: mcp`, `area: console`, `area: windows`, `area: install`, `area: security`), et les marqueurs (`good first issue`, `help wanted`, `breaking`, `knock-first`). Les jalons reprennent les lots.

### Le rythme

Un créneau fixe par semaine, d'une à deux heures. Chaque nouveau ticket reçoit une réponse dans la semaine : confirmé, besoin de précisions, doublon, hors périmètre, ou accepté et placé dans un jalon. Ce n'est pas un engagement de correction. Un ticket `needs-info` sans réponse pendant trente jours se ferme, et se rouvre d'un commentaire. Pas de robot qui ferme les tickets confirmés pour inactivité : c'est la pratique la plus détestée des communautés, et elle cache le travail au lieu de le faire.

### Les soupapes

Pour un pic : épingler une Discussion de réponses fréquentes, fermer vite ce qui n'est pas reproductible, demander aux trieurs et à la communauté de répondre, verrouiller les fils anciens, plafonner les pull requests ouvertes par auteur (OpenClaw en tolère vingt), et si cela déborde encore, limiter temporairement la création de tickets et de pull requests aux collaborateurs avec un mot d'explication dans le README, comme PocketBase l'a fait. Pour une pause du mainteneur : une ligne dans le README et dans les Discussions, avec la date de retour. Personne ne le reproche à un projet qui le dit.

## Le support

Trois niveaux, et pas de quatrième :

1. **La documentation**, première ligne, avec une page de dépannage tenue à partir des questions qui reviennent.
2. **Les Discussions (Q&A)**, où la communauté répond et où le mainteneur passe dans son créneau.
3. **Les tickets**, pour les bugs reproductibles.

Pas de support par e-mail, pas de message privé, pas de délai garanti. Les seules adresses privées sont `security@cahut.cloud` pour les failles, `conduct@cahut.cloud` pour le code de conduite et `contact@cahut.cloud` pour la marque et les partenariats ; aucune ne fait de support. Une page « Getting help » (SUPPORT.md, que GitHub affiche à l'ouverture d'un ticket) le dit en dix lignes, et dit comment bien demander : la version, le système, la méthode d'installation, le Logbook. Une commande `cahut doctor`, qui rassemblerait ces informations sans secret, ferait gagner du temps à tout le monde ; c'est une idée de produit, à placer dans un lot.

Le support payant pour les PME, évoqué dans le brief, est une affaire de phase 3 : après la 1.0, séparé du projet, et jamais au détriment des réponses publiques.

## Les canaux

### Au quotidien

| Canal | Usage | Rythme |
| --- | --- | --- |
| GitHub (code, Issues, Discussions, Releases, jalons) | La source de vérité, le seul endroit où le travail se fait | Chaque semaine |
| cahut.cloud | Expliquer le projet, ses règles et ses nouvelles ; un flux RSS | À chaque release |
| docs.cahut.cloud | Installer, brancher une IA, comprendre, administrer, contribuer | Avec le code |
| Lettre d'information | Les releases et les nouvelles, rien d'autre ; au début, un flux RSS suffit, la lettre vient au lancement | Une fois par mois au plus |
| Mastodon et LinkedIn | Les mêmes nouvelles, en deux phrases ; Mastodon pour le monde du logiciel libre, LinkedIn pour les PME et le réseau d'Olivier ; Bluesky ou X si le temps reste | À chaque release |
| Chat | Aucun au début | Voir ci-dessous |

Un chat (Discord, Matrix ou Zulip) ne se tient pas seul : il demande une présence, il perd les réponses, il attire les questions qui devraient être des tickets. Il viendra quand deux personnes au moins pourront le modérer, en phase 2, et la question se posera alors entre Discord, où sont les auto-hébergeurs, et Zulip ou Matrix, où sont les projets libres. Le même identifiant partout, `getcahut`, à réserver dès maintenant sur Mastodon, Bluesky, LinkedIn et X, puisque `cahut` est pris sur GitHub.

### Le lancement

Les lieux où un projet auto-hébergé se fait connaître ont chacun leurs règles, et elles ont durci en 2026 :

- **Show HN** (Hacker News) : quelque chose qu'on peut installer et essayer, un projet « non trivial », l'auteur présent pour répondre, aucun texte généré dans le message, pas de sollicitation de votes.
- **r/selfhosted** : depuis avril 2026, un projet de moins de trois mois ne se présente que dans le fil hebdomadaire des nouveaux projets, et le gabarit demande d'être transparent sur la part de l'IA.
- **awesome-selfhosted** : première release datant de plus de quatre mois, maintenance active, instructions d'installation qui marchent, licence libre ; la pull request doit être écrite par un humain. La liste exclut les plateformes pour déployer des applications arbitraires : Cahut s'y présente comme une maison pour les documents, les présentations et les tableaux de bord, pas comme un PaaS.
- **selfh.st** et sa lettre hebdomadaire, le répertoire d'applications auto-hébergées ; **LinuxFr.org** et le **Journal du hacker** pour le public francophone.

Ces délais se comptent à partir de la première release taguée : ouvrir le dépôt tôt fait courir les compteurs.

### Du côté des IA

Cahut se branche sur les IA : il doit être trouvable là où leurs utilisateurs cherchent des connecteurs. Le registre officiel MCP (registry.modelcontextprotocol.io), les listes de serveurs MCP et, pour chaque client, l'endroit où se déclarent les plugins et les skills, sont des canaux de distribution à part entière, à prendre lot par lot avec chaque client IA.

## Versions et releases

- **Semantic Versioning.** Tant que Cahut est en 0.x, tout peut changer, et le README le dit comme PocketBase le dit. La 1.0 marque le lancement public : à partir de là, une version majeure est la seule qui casse quelque chose.
- **Une release à la fin de chaque lot**, puis toutes les deux à quatre semaines ; une release de correctif dès qu'une faille est corrigée. Pas de calendrier fixe : Forgejo publie chaque trimestre avec une LTS annuelle, c'est un modèle pour plus tard, pas pour maintenant.
- **Les notes de release sont écrites pour des humains**, selon [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) : les points marquants, ce qui casse et comment migrer, puis *Added*, *Changed*, *Fixed*, *Security*, et les contributeurs remerciés. GoReleaser peut regrouper les entrées par étiquette des pull requests (`changelog.use: github`) ; le brouillon se relit à la main, comme l'ADR 7 le veut. GitHub Releases est le journal des changements ; pas de CHANGELOG.md à tenir en double.
- **Une seule version prise en charge**, la dernière : les mises à jour automatiques le rendent tenable. Un outil MCP ou une opération qui disparaît est annoncé une version avant.
- **Signatures et provenance.** Les sommes de contrôle existent ; la signature des releases arrive au lot 7, et les attestations de provenance de GitHub peuvent suivre.

## Sécurité

Un serveur auto-hébergé qui reçoit du HTML d'une IA et le sert au public se juge sur sa sécurité. Le dispositif, à poser avant l'ouverture :

- **SECURITY.md** : où signaler, quoi attendre, ce qui est dans le périmètre. Le signalement privé de GitHub (*private vulnerability reporting*, gratuit, activé par le propriétaire du dépôt) ou `security@cahut.cloud`. Accusé de réception sous sept jours (le badge OpenSSF demande quatorze), correction et publication coordonnées, quatre-vingt-dix jours au plus, bien moins pour une faille grave ; crédit au rapporteur. Dans le périmètre : le binaire, `install.sh`, les sites, et toute évasion du bac à sable d'une Window. Hors périmètre : le contenu des Windows publiées, les clients IA, les services tiers.
- **Un avis de sécurité GitHub** pour chaque faille corrigée, avec un CVE demandé par GitHub ; la release de correctif le cite.
- **security.txt** ([RFC 9116](https://www.rfc-editor.org/rfc/rfc9116)) sur cahut.cloud : `Contact`, `Expires`, `Policy`, `Preferred-Languages: en, fr`.
- **La règle anti-rapports générés** de la section précédente, écrite dans SECURITY.md.
- **Dans la CI** : `govulncheck` sur les dépendances Go, en plus de Dependabot ; une analyse statique est déjà là avec golangci-lint.
- **Le badge OpenSSF Best Practices**, niveau *passing*, visé avant le lancement : c'est une liste de contrôle gratuite (licence, site en HTTPS, processus de contribution, tests automatiques, analyse statique, réponse aux failles sous quatorze jours, pas de secret dans le dépôt) et un signal de sérieux lisible de loin. Le Scorecard de l'OpenSSF, automatique, complète.

## Nom, marque et licence

### La licence

C'est la première décision, et elle conditionne les autres. Les deux options du brief :

| | Apache-2.0 | AGPL-3.0 |
| --- | --- | --- |
| Esprit | Permissive : chacun fait ce qu'il veut, y compris une version fermée | Copyleft réseau : qui modifie Cahut et le fait tourner pour d'autres publie ses modifications |
| Qui l'utilise | Caddy, Home Assistant, Kubernetes, Go | Nextcloud, Mastodon, Plausible, Immich, Vaultwarden, listmonk |
| Hébergeurs partenaires (OVH, Hostinger) | Libres | Libres aussi : installer et distribuer Cahut sans le modifier n'oblige à rien, et Nextcloud, AGPL, est chez tous les hébergeurs |
| Offre hébergée par Olivier | Possible | Possible : le projet publie déjà tout |
| Un concurrent qui fermerait une version améliorée | Possible | Pas sans publier |
| Brevets | Clause explicite | Clause équivalente |
| Entreprises frileuses | Aucune réserve | Certaines refusent l'AGPL pour des bibliothèques ; Cahut est un serveur installé, pas une bibliothèque |
| Contributeurs | Indifférent | Indifférent |

Recommandation : **AGPL-3.0 pour le serveur**, c'est la licence des projets auto-hébergés dont Cahut se réclame, et la seule qui protège la promesse « le même code pour tous ». À côté, **Apache-2.0 ou MIT pour ce que les gens copient chez eux** : les skills et les gabarits destinés aux IA, les exemples ; et **CC BY-SA 4.0 pour la documentation**. Le nom et le logo ne sont couverts par aucune de ces licences, et le fichier LICENSE le dit.

Le choix se tient dans les deux sens : avec l'Apache-2.0, le projet garderait la liberté d'une version fermée ou d'un partenariat exclusif ; avec l'AGPL et sans CLA, il y renonce pour toujours, et c'est ce qui rend la promesse crédible. La question à Olivier n'est donc pas seulement « quelle licence », c'est « est-ce que je ferme cette porte ».

**Décision.** Le 22 octobre 2026, Olivier a retenu l'AGPL-3.0, version 3 seule, pour tout le dépôt, sans CLA ([ADR 15](adr/0015-license.md)) : la porte d'une version fermée est fermée. Les skills et les gabarits porteront l'Apache-2.0 quand ils arriveront ; une licence Creative Commons pour la documentation se décidera à l'ouverture du dépôt.

### Le nom

Les licences libres couvrent le code, pas le nom : l'Apache-2.0 le dit noir sur blanc, et c'est pour cela que Mastodon, Nextcloud et Rust ont une politique de marque. Pour Cahut :

- **Déposer la marque verbale « Cahut »** à l'INPI avant l'ouverture du dépôt (environ 190 € pour une classe, en ligne), puis à l'EUIPO quand le projet a une traction (environ 850 € pour une classe). Les classes à retenir, logiciels et services en ligne, sont à vérifier avec l'INPI.
- **Publier une politique de marque courte**, sur le modèle de Mastodon (mise à jour en août 2026) et de Nextcloud : libre de dire « compatible avec Cahut », « tourne sur Cahut », de distribuer Cahut sans le modifier, y compris en image de conteneur, de l'auto-héberger, d'écrire et d'enseigner à son sujet ; interdit d'utiliser le nom ou le logo dans le nom d'un produit, d'une entreprise, d'un domaine ou d'un service, de laisser croire à un lien officiel, ou de garder le nom sur une version modifiée. Un fork prend un autre nom. Un hébergement commercial sous le nom « Cahut » demande un accord écrit, à demander à `contact@cahut.cloud` : c'est la base du programme de partenaires du brief. La politique de Rust de 2023 montre la limite : trop restrictive, elle a soulevé la communauté et a été réécrite en 2025.
- **Le logo reste tel quel**, en toutes versions, comme la charte le dit déjà ; la page `/brand` du site accueille la politique de marque.
- **Le domaine** cahut.cloud se renouvelle longtemps à l'avance, avec un second contact.

## Argent et pérennité

### Ce que l'argent peut faire, et ce qu'il ne fera pas

Les dons à un projet open source d'une personne paient un domaine, un serveur de test, un dépôt de marque et quelques outils. Ils ne paient pas un salaire : c'est vrai pour presque tous les projets de cette taille, et le dire d'emblée évite la déception des deux côtés. Cahut vivra d'abord du temps libre d'Olivier ; l'argent vient ensuite, par des services, pas par des dons.

### Les dons

- **GitHub Sponsors** dès l'ouverture, dans `.github/FUNDING.yml` : zéro frais sur les dons des comptes personnels, la France est couverte, et c'est là que sont les développeurs.
- **Liberapay** en second : une association française, aucun frais de plateforme, juste ceux du paiement ; ses dons sont récurrents et plafonnés, elle convient aux petits soutiens réguliers.
- **Une phrase publique sur l'usage des fonds** : le domaine, le serveur de test, la marque, les outils ; tout le reste est rendu visible quand il existe.
- Pas d'Open Collective au début : son hôte européen, devenu Open Source Europe, prend 8 % et sert quand une équipe dépense collectivement, en phase 3.

### Les subventions

Peu de pistes réalistes, et il vaut mieux le savoir que d'y passer des semaines :

| Guichet | Verdict |
| --- | --- |
| NLnet (ex-NGI Zero, désormais Open Internet Stack : Restack, CodeSupply) | Non : la politique d'octobre 2026 interdit le code produit par une IA générative dans les travaux financés, et les projets « liés à l'IA » sont hors périmètre sous un million d'utilisateurs |
| Sovereign Tech Fund (Allemagne) | Non : pas d'applications destinées aux utilisateurs, projets de plus de 50 000 € |
| Prix BlueHats (DINUM et NLnet, 4 × 10 000 €) | Peut-être plus tard : il faut qu'une administration française utilise le logiciel ; pas d'édition 2026 trouvée |
| Instrument européen de maintenance de l'open source (stratégie de juin 2026) | À suivre : annoncé, sans périmètre ni budget |

### Une offre hébergée, du service

Le brief le prévoit : une offre hébergée en Europe, et du support et des services pour les PME. C'est le modèle de Miniflux, de Ghost et de Home Assistant ; Plausible va plus loin en réservant quelques fonctionnalités à son cloud, ce que Cahut ne fera pas. Trois règles :

1. **Le même binaire pour tous.** Aucune fonctionnalité réservée. La règle « une seule API » s'étend au modèle économique.
2. **Une frontière nette.** L'offre est une activité d'Olivier (micro-entreprise ou société, quand elle existe), avec son nom commercial, sa facturation et ses conditions ; le projet reste le projet. Pour vendre un abonnement, Polar fait office de marchand officiel et gère la TVA, ce que ni GitHub Sponsors ni Liberapay ne font.
3. **Le projet n'en dépend pas.** Si l'offre s'arrête, Cahut continue ; si elle grandit, le nom et le domaine passent à une structure neutre, ci-dessous.

Ce sont des affaires de phase 3, après la 1.0. Avant, en parler sur le site suffirait à créer une attente.

### Une structure, un jour

Pas de structure aujourd'hui : l'Open Source Guide le dit, une entité juridique ne sert que quand il y a de l'argent à tenir. Quand il y en aura, ou quand il y aura une équipe, trois voies :

- **Une association loi 1901** : deux dirigeants, une déclaration en ligne, une publication gratuite au Journal officiel ; elle détient le nom, le domaine et les comptes, et reçoit des dons manuels. Simple, française, bien connue des hébergeurs et des administrations.
- **Un hôte fiscal**, Open Source Europe (Bruxelles) : aucune entité à créer, le projet reste autonome, les dépenses sont publiques ; 8 % des fonds.
- **The Commons Conservancy** (Amsterdam) : une structure partagée qui porte des « programmes » avec leurs propres statuts, sans gérer d'argent elle-même. C'est l'option « Open Home Foundation en miniature ».

Le critère : qui doit détenir le nom et le domaine pour que la promesse tienne même si Olivier change d'avis, de métier ou de vie. Le brief demandait « compte personnel ou organisation dédiée » : l'organisation existe ; la question suivante est à qui elle appartient. Et cela se prépare tôt : Mastodon a annoncé son entité européenne en janvier 2025 et ne l'avait pas encore en février 2026.

## Code de conduite

Le [Contributor Covenant](https://www.contributor-covenant.org/) en version 3.0 (juillet 2025), tel quel, avec un canal de signalement : `conduct@cahut.cloud`. Son échelle de réponses se suit sans en sauter les marches sauf gravité : avertissement privé, activité limitée pour un temps, suspension temporaire, exclusion définitive. Tant qu'Olivier est seul, c'est lui qui reçoit et qui décide, et le document le dit honnêtement, y compris pour un signalement qui le concernerait : dans ce cas, la personne de confiance reçoit. En phase 2, un second mainteneur partage ce rôle.

La modération des Discussions suit le même texte. Un principe simple pour le quotidien, celui de Mike McQuaid : bloquer sans débat qui est grossier, et ne pas s'excuser d'un bug.

## Les fichiers du dépôt et les pages du site

### Dans le dépôt

GitHub lit ces fichiers et les affiche aux bons endroits : à la racine, en anglais, courts.

| Fichier | Contenu | Quand |
| --- | --- | --- |
| LICENSE | Le texte de l'AGPL-3.0 ; la mention sur le nom et le logo est dans le README | Fait |
| README.md | Déjà là ; une section *Contributing and community* renvoie aux fichiers ci-dessous | Fait |
| GOVERNANCE.md | Les principes, les rôles, comment on décide, les phases : le résumé public, en anglais, de ce document | Fait |
| CONTRIBUTING.md | Ce qui est bienvenu, ce qui se discute, le parcours d'une pull request, le DCO, la règle sur l'IA, le renvoi à CLAUDE.md et au plan | Fait |
| CODE_OF_CONDUCT.md | Le Contributor Covenant 3.0 et le canal de signalement | Fait |
| SECURITY.md | Où signaler, délais, périmètre, versions prises en charge | Fait |
| SUPPORT.md | Où demander de l'aide, et comment bien demander | Fait |
| TRADEMARK.md | La politique de marque | Fait |
| MAINTAINERS.md | Les mainteneurs et les émérites | Phase 2 |
| .github/ISSUE_TEMPLATE/, PULL_REQUEST_TEMPLATE.md, FUNDING.yml | Les gabarits et les dons | Fait |
| .well-known/security.txt (dans le site) | Le fichier de la RFC 9116 | Avant l'ouverture |

CLAUDE.md reste la convention de travail, lue par les IA ; CONTRIBUTING.md est écrit pour les humains et renvoie à CLAUDE.md pour les conventions, afin de ne rien tenir en double. Ce document-ci, en français, reste dans `docs/` comme le brief et les plans : il dit le pourquoi ; les fichiers de la racine disent le quoi.

### Sur cahut.cloud

Une section du site plutôt qu'un sous-domaine : un seul site à tenir, le pied de page a déjà sa colonne « Project », et la version française existe. Un sous-domaine (`community.cahut.cloud`) n'aurait de sens que pour un forum auto-hébergé, un jour.

| Page | Contenu | Source |
| --- | --- | --- |
| `/project/` | Comment Cahut est mené : qui décide, les principes, les phases, comment aider, l'argent | GOVERNANCE.md |
| `/project/contribute/` | Comment contribuer, avec la règle sur l'IA | CONTRIBUTING.md |
| `/project/security/` | Signaler une faille | SECURITY.md |
| `/project/code-of-conduct/` | Le code de conduite | CODE_OF_CONDUCT.md |
| `/project/support/` | Où demander de l'aide, et comment bien demander | SUPPORT.md |
| `/project/roadmap/` | Les lots, sans dates, et ce qui marche aujourd'hui | Le plan de la V0 et la page « Where Cahut stands » |
| `/project/news/` | Les releases et les nouvelles, avec un flux RSS | GitHub Releases, lues à la construction du site |
| `/brand/` (existe) | Plus « Using the name and logo » | TRADEMARK.md |

Les pages se construisent à partir des fichiers du dépôt, comme le site importe déjà la charte par l'alias `@charte` : une seule source, pas de copie qui vieillit. Les versions françaises sous `/fr/project/` sont des traductions tenues dans le site. La construction du site se déclenche à la publication d'une release, pour que la page des nouvelles suive. Tout cela tient dans un ADR, « Les pages du projet sont construites à partir des fichiers communautaires du dépôt » ([ADR 18](adr/0018-project-pages-from-community-files.md)), et une pull request par page.

### Dans la documentation

Un chapitre « Contribute » dans docs.cahut.cloud : préparer sa machine (ce que CLAUDE.md et le README disent aujourd'hui), le tour du code (`internal/`), écrire une opération, écrire un ADR, publier une release. Et un chapitre « Decisions » qui rend les ADR lisibles sur le site, à partir de `docs/adr/`, si le chargeur de contenu d'Astro s'y prête ; sinon un lien vers GitHub suffit.

## Comment le projet vit : les phases

| Phase | Déclencheur | Ce qui change | Ce qui ne change pas |
| --- | --- | --- | --- |
| **0 · Avant l'ouverture** | Maintenant | Les décisions de ce document, les fichiers, la marque, les réglages du dépôt, les pages du site ; la relecture complète du dépôt et de son historique | Le développement en lots |
| **1 · Le dépôt est ouvert** | L'ouverture du dépôt, recommandée au jalon V0 plutôt qu'au lancement public | Le créneau hebdomadaire, les releases par lot, les nouvelles ; les premiers tickets et les premières `good first issue` ; les compteurs des listes et des forums commencent à courir | Olivier décide seul ; pas de chat ; pas de support |
| **2 · Les premiers contributeurs réguliers** | Deux personnes avec trois pull requests fusionnées en trois mois, ou un tri qui déborde du créneau | Le rôle de trieur, MAINTAINERS.md, la personne de confiance confirmée, un second mainteneur pour les relectures sensibles et le code de conduite ; peut-être un chat, à deux pour le tenir | Le socle ; le même créneau par personne |
| **3 · Une équipe** | Trois mainteneurs actifs depuis six mois, ou de l'argent (revenus ou dons au-dessus des coûts), ou l'offre hébergée | Le consensus tacite, GOVERNANCE.md en version 2, une structure pour le nom et le domaine, la frontière avec l'activité commerciale, le support payant | Le socle ; le mainteneur principal tranche |

Pourquoi ouvrir au jalon V0 : le site renvoie déjà vers GitHub, le brief fonde la crédibilité sur un code qui se lit, les premiers testeurs visés sont des « geeks sur VPS » qui lisent le code avant d'installer, l'avertissement du README cadre les attentes, et les délais d'awesome-selfhosted et de r/selfhosted se comptent à partir de la première release. Le prix : des tickets plus tôt. Les soupapes sont là pour ça. Le lancement public avec le périmètre complet reste un événement à part, plus tard, comme le brief le veut.

### La semaine du mainteneur

Ce que le projet annonce, et ce qu'Olivier se tient :

| Quoi | Quand | Combien |
| --- | --- | --- |
| Tri des tickets, Discussions, relecture des pull requests | Un créneau fixe par semaine | 1 à 2 h |
| Release | À la fin de chaque lot, puis toutes les deux à quatre semaines | 2 h |
| Nouvelles : site, lettre, Mastodon, LinkedIn | À chaque release | 1 h |
| Sécurité | Dès réception, avant tout le reste | Ce qu'il faut |

Environ quatre heures par semaine en dehors du développement, et pas plus pendant les semaines de lot : c'est écrit dans SUPPORT.md, et c'est ce qui évite que le projet mange la personne.

## À faire avant l'ouverture du dépôt

- [x] Choisir la licence et écrire LICENSE, avec la mention sur le nom et le logo : fait le 22 octobre 2026 ([ADR 15](adr/0015-license.md)).
- [ ] Déposer la marque « Cahut » à l'INPI ; réserver `getcahut` sur Mastodon, Bluesky, LinkedIn et X.
- [x] Écrire GOVERNANCE.md, CONTRIBUTING.md, CODE_OF_CONDUCT.md, SECURITY.md, SUPPORT.md, TRADEMARK.md, les gabarits et FUNDING.yml ; compléter le README : fait le 22 octobre 2026.
- [x] Régler le dépôt autant que GitHub le permet à un dépôt privé du plan gratuit : Discussions et leurs catégories, étiquettes, jalons par lot, signature des commits faits sur le web : fait le 22 octobre 2026 ; fusion par squash seule, branches supprimées après fusion, alertes et correctifs de sécurité de Dependabot, actions épinglées par empreinte obligatoires, environnement `sites` réservé à `main`, CODEOWNERS et réglages de l'application DCO : fait le 23 octobre 2026.
- [x] Dans l'interface, par Olivier : la double authentification obligatoire sur l'organisation et l'application DCO (<https://github.com/apps/dco>) installée sur le dépôt : fait, constaté le 24 octobre 2026.
- [ ] Le profil GitHub Sponsors, par Olivier.
- [ ] Le jour de l'ouverture, ce que GitHub refuse à un dépôt privé : appliquer les règles de `.github/rulesets/` (protection de `main` et des étiquettes de release), activer le signalement privé des failles, l'analyse des secrets avec la protection des envois, l'analyse CodeQL par défaut, et demander l'approbation des workflows pour toute contribution extérieure ; ajouter `DCO` aux vérifications requises de `main` une fois l'application installée. Le même jour, remplacer le badge de version du README, écrit à la main à chaque release tant que shields.io ne peut pas lire les releases d'un dépôt privé, par le badge dynamique `https://img.shields.io/github/v/release/getcahut/cahut?label=release&color=B9E1FB&labelColor=1F2A33`.
- [ ] Créer `security@`, `conduct@` et `contact@cahut.cloud`, par Olivier.
- [x] Publier `security.txt` sur cahut.cloud : fait le 24 octobre 2026. Il nomme les deux canaux de SECURITY.md : `security@cahut.cloud`, dont la boîte reste à créer, et le signalement privé de GitHub, actif le jour de l'ouverture. Il expire 180 jours après chaque construction du site.
- [ ] Désigner la personne de confiance, lui donner les accès de secours et la note scellée.
- [x] Ajouter `govulncheck` à la CI : fait le 24 octobre 2026 (`make vuln`).
- [ ] Remplir le questionnaire du badge OpenSSF, une fois le dépôt ouvert : il lit le dépôt public.
- [x] Écrire l'ADR des pages du projet et construire les pages `/project/` du site, en anglais et en français : fait ([ADR 18](adr/0018-project-pages-from-community-files.md)).
- [ ] Relire tout le dépôt et son historique avant l'ouverture, comme le brief le demande : l'historique s'ouvre tel quel et ne se réécrit pas. Un premier passage, le 24 octobre 2026, sur les 143 commits : aucun secret (clés privées, jetons de GitHub, d'OpenAI, d'AWS ou de Slack, jetons de Cahut), `tmp/` jamais commité, aucun nom d'autre projet ni de client. EKLOS, cité dans le brief, était un ancien nom de Cahut. Reste la relecture d'Olivier.

## Décisions à prendre

| Sujet | Options | Recommandation |
| --- | --- | --- |
| Licence du serveur | Apache-2.0 ou AGPL-3.0 | **Tranché le 22 octobre 2026 : AGPL-3.0** ([ADR 15](adr/0015-license.md)) ; Apache-2.0 pour les skills et les gabarits quand ils arriveront ; CC BY-SA 4.0 pour la documentation, à confirmer à l'ouverture |
| Fermer la porte d'une version fermée | Garder l'option ou y renoncer par écrit | **Tranché avec la licence** : la porte est fermée, sans CLA ; la promesse publique s'écrira dans GOVERNANCE.md |
| Certification des contributions | DCO, CLA ou rien | DCO, avec l'application GitHub |
| Moment de l'ouverture du dépôt | Au jalon V0 ou au lancement public | Au jalon V0 |
| Chat communautaire | Aucun, Discord, Zulip, Matrix | Aucun en phase 1 ; à choisir en phase 2 |
| Lettre d'information | Flux RSS seul, Buttondown, listmonk auto-hébergé | Flux RSS dès l'ouverture, une lettre au lancement |
| Réseaux | Mastodon, LinkedIn, Bluesky, X | Mastodon et LinkedIn ; identifiant `getcahut` réservé partout |
| Dons | GitHub Sponsors, Liberapay, Open Collective | GitHub Sponsors et Liberapay dès l'ouverture ; Open Source Europe en phase 3 si besoin |
| Marque | INPI, EUIPO, rien | INPI avant l'ouverture, EUIPO avec la traction |
| Pages du site | Section `/project/` ou sous-domaine | Section `/project/`, construite à partir des fichiers du dépôt |
| Structure juridique | Aucune, association, hôte fiscal, Commons Conservancy | Aucune avant la phase 3 ; association loi 1901 en première intention |
| Télémétrie | Aucune, vérification de version seule, mesure d'installations en option | Vérification de version seule, documentée et désactivable |
| Code de conduite | Contributor Covenant 3.0 ou 2.1 | 3.0 |
| Langue des échanges | Anglais, français, les deux | Anglais pour tout ce qui est public ; une réponse en français n'est jamais refusée |

## Questions ouvertes

- [ ] Qui est la personne de confiance ?
- [ ] L'offre hébergée fait-elle partie du plan, et à quel horizon ? Elle pèse sur la marque et la structure.
- [ ] L'adresse de signalement : `security@cahut.cloud` suppose une boîte sur le domaine ; qui la relève pendant une absence ?
- [ ] Quelle place pour le français dans la documentation et les Discussions, au-delà du site ?
- [ ] Un hébergeur partenaire voudra-t-il un accord de marque avant l'ouverture ?

## Sources

Guides et modèles :

- Open Source Guides, *Leadership and Governance* : https://opensource.guide/leadership-and-governance/
- GitHub, *Minimum Viable Governance* : https://github.com/github/MVG
- Contributor Covenant 3.0 (juillet 2025) : https://www.contributor-covenant.org/version/3/0/code_of_conduct/
- Developer Certificate of Origin : https://developercertificate.org/ ; l'application GitHub DCO : https://github.com/apps/dco ; GitLab passe du CLA au DCO (2017) : https://about.gitlab.com/blog/gitlab-switches-to-dco-license/
- OpenSSF Best Practices, critères *passing* : https://www.bestpractices.dev/en/criteria/0 ; Scorecard : https://scorecard.dev/
- Signalement privé des failles sur GitHub : https://docs.github.com/en/code-security/security-advisories/working-with-repository-security-advisories/configuring-private-vulnerability-reporting-for-a-repository ; RFC 9116, security.txt : https://www.rfc-editor.org/rfc/rfc9116
- Semantic Versioning : https://semver.org/ ; Keep a Changelog : https://keepachangelog.com/en/1.1.0/ ; calendrier de Forgejo : https://forgejo.org/docs/latest/admin/release-schedule/
- Nadia Eghbal, *Working in Public*, résumé des types de projets : https://project-types.github.io/ ; Mike McQuaid, *Open Source Maintainers Owe You Nothing* : https://mikemcquaid.com/open-source-maintainers-owe-you-nothing/

Politiques sur l'IA :

- curl, *On AI use in curl* : https://github.com/curl/curl/blob/master/docs/CONTRIBUTE.md ; la fin des primes (janvier 2026) : https://daniel.haxx.se/blog/2026/01/26/the-end-of-the-curl-bug-bounty/
- Ghostty, AI policy : https://github.com/ghostty-org/ghostty/blob/main/AI_POLICY.md ; Gitea : https://github.com/go-gitea/gitea/blob/main/CONTRIBUTING.md ; Forgejo : https://codeberg.org/forgejo/governance/raw/branch/main/AIAgreement.md ; Zig : https://ziglang.org/code-of-conduct/
- Noyau Linux, *coding-assistants* : https://github.com/torvalds/linux/blob/master/Documentation/process/coding-assistants.rst ; LLVM : https://llvm.org/docs/AIToolPolicy.html
- GitHub, réglages d'accès aux pull requests (février 2026) : https://github.blog/changelog/2026-02-13-new-repository-settings-for-configuring-pull-request-access/
- NLnet, politique sur l'IA générative 2.0 (octobre 2026) : https://nlnet.nl/foundation/policies/generativeAI/2.0 ; Restack : https://nlnet.nl/restack/

Leçons :

- Chronologie de la porte dérobée de xz, par Russ Cox : https://research.swtch.com/xz-timeline ; alerte OpenSSF et OpenJS : https://openssf.org/blog/2024/04/15/open-source-security-openssf-and-openjs-foundations-issue-alert-for-social-engineering-takeovers-of-open-source-projects/
- HashiCorp passe à la BSL (2023) : https://www.hashicorp.com/en/blog/hashicorp-adopts-business-source-license ; LWN sur le rôle du CLA : https://lwn.net/Articles/942346 ; OpenTofu : https://www.linuxfoundation.org/press/announcing-opentofu
- Redis change de licence (2024) : https://redis.io/blog/redis-adopts-dual-source-available-licensing/ ; Valkey : https://www.linuxfoundation.org/press/linux-foundation-launches-open-source-valkey-community ; retour à l'AGPL (2025) : https://redis.io/blog/agplv3/
- Elastic et OpenSearch (2021) : https://aws.amazon.com/blogs/opensource/stepping-up-for-a-truly-open-source-elasticsearch/ ; Elastic revient à l'AGPL (2024) : https://www.elastic.co/blog/elasticsearch-is-open-source-again
- SQLite, *open source, not open contribution* : https://sqlite.org/copyright.html

Projets :

- PocketBase, contributions : https://github.com/pocketbase/pocketbase/blob/master/CONTRIBUTING.md ; point de février 2026 : https://github.com/pocketbase/pocketbase/discussions/7287
- Miniflux, logiciel opiniâtre : https://miniflux.app/opinionated.html ; version hébergée : https://miniflux.app/hosting.html
- Litestream, contributions : https://github.com/benbjohnson/litestream/blob/main/CONTRIBUTING.md ; l'auteur chez Fly.io : https://fly.io/blog/all-in-on-sqlite-litestream/
- Caddy, sponsors : https://caddyserver.com/sponsor ; l'en-tête de 2017 : https://news.ycombinator.com/item?id=15237923 et https://github.com/caddyserver/caddy/pull/1866
- Uptime Kuma, contributions et rôles : https://github.com/louislam/uptime-kuma/blob/master/CONTRIBUTING.md ; le temps de relecture (2025) : https://github.com/louislam/uptime-kuma/issues/5665
- Vaultwarden, le renommage (2021) : https://github.com/dani-garcia/vaultwarden/discussions/1642
- Immich rejoint FUTO (2024) : https://immich.app/blog/immich-joins-futo ; deux ans après : https://immich.app/blog/futo-two-years-later
- Home Assistant, Open Home Foundation (2024) : https://www.home-assistant.io/blog/2024/04/24/state-of-the-open-home-2024/ ; structure : https://www.openhomefoundation.org/structure ; Nabu Casa : https://www.nabucasa.com/about/
- Mastodon, vers une entité européenne (2025) : https://blog.joinmastodon.org/2025/01/the-people-should-own-the-town-square/ ; le fondateur passe la main : https://blog.joinmastodon.org/2025/11/my-next-chapter-with-mastodon/ ; point de février 2026 : https://blog.joinmastodon.org/2026/02/administrative-updates/
- Plausible, à propos : https://plausible.io/about ; contributions : https://github.com/plausible/analytics/blob/master/CONTRIBUTING.md
- Ghost, à propos : https://ghost.org/about/ ; contributions : https://github.com/TryGhost/Ghost/blob/main/.github/CONTRIBUTING.md
- Syncthing, fondation : https://syncthing.net/foundation/ ; contributions : https://github.com/syncthing/syncthing/blob/main/CONTRIBUTING.md
- Gitea Ltd (2022) : https://blog.gitea.com/2022/10/open-source-sustainment-and-the-future-of-gitea/ ; la lettre ouverte : https://gitea-open-letter.coding.social/ ; Forgejo, gouvernance : https://forgejo.org/docs/latest/contributor/governance/ ; le fork dur (2024) : https://forgejo.org/2024-02-forking-forward/
- Paperless-ng, le constat d'abandon (2022) : https://github.com/jonaswinkler/paperless-ng/issues/1599 ; la reprise : https://github.com/jonaswinkler/paperless-ng/issues/1632
- listmonk, contributions : https://github.com/knadh/listmonk/blob/master/CONTRIBUTING.md
- OpenClaw, le récit de l'auteur (janvier 2026) : https://steipete.me/posts/2026/openclaw ; la fondation (juillet 2026) : https://openclaw.ai/blog/introducing-openclaw-foundation ; les mainteneurs vus par GitHub (août 2026) : https://github.blog/open-source/maintainers/openclaw-went-viral-meet-the-maintainers-building-and-securing-it/

Lancement, marque, argent et structures :

- awesome-selfhosted, critères : https://github.com/awesome-selfhosted/awesome-selfhosted-data/blob/master/CONTRIBUTING.md ; Show HN : https://news.ycombinator.com/showhn.html ; règles de Hacker News : https://news.ycombinator.com/newsguidelines.html ; selfh.st : https://selfh.st/about/
- Politique de marque de Mastodon : https://joinmastodon.org/trademark ; de Nextcloud : https://nextcloud.com/trademarks/ ; de Rust : https://rustfoundation.org/policy/rust-trademark-policy/ ; Apache-2.0, article 6 : https://www.apache.org/licenses/LICENSE-2.0
- INPI, coût d'une marque : https://www.inpi.fr/realiser-demarches/propriete-intellectuelle/deposant-et-cout-dune-marque ; EUIPO : https://www.euipo.europa.eu/en/trade-marks/before-applying/fees-payments
- GitHub Sponsors : https://docs.github.com/en/sponsors/getting-started-with-github-sponsors/about-github-sponsors ; Liberapay : https://liberapay.com/about/faq ; Open Source Europe : https://opensourceeurope.org/ ; Polar : https://polar.sh/resources/pricing
- Sovereign Tech Fund : https://www.sovereign.tech/programs/fund ; prix BlueHats 2024 : https://nlnet.nl/bluehatsprize/2024/ ; stratégie open source de la Commission (juin 2026) : https://digital-strategy.ec.europa.eu/en/policies/open-source-strategy
- Association loi 1901 : https://www.service-public.gouv.fr/associations/vosdroits/F1119 ; The Commons Conservancy : https://commonsconservancy.org/
