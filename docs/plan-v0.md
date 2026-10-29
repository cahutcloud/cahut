# Cahut · Plan de la V0

21 octobre 2026 · Olivier Durand · Proposition à valider

## En bref

La V0 est le MVP de Cahut : un jalon privé, pas une version publique. Le lancement garde son périmètre complet ; la V0 prouve la boucle centrale (installer, brancher une IA, publier, partager, tracer) et pose une architecture sur laquelle le reste du périmètre viendra se greffer sans reprise.

Elle se construit en neuf lots, du lot 0 au lot 8, pour environ 12 à 15 semaines à temps plein. Deux jalons la ponctuent :

- **Jalon PoC**, à la fin du lot 3, vers la sixième semaine : Cahut sur un VPS, une IA qui publie une présentation par MCP, la console. C'est l'étape 4 de la feuille de route du README.
- **Jalon V0**, à la fin du lot 8 : les cinq critères de réussite du README passent sur le périmètre de la V0.

## Périmètre

| Dans la V0 | Après la V0 |
| --- | --- |
| Installation en une commande (AlmaLinux, Ubuntu), premier démarrage guidé | Applis en conteneurs : runtimes, Dockerfile, bases, secrets, journaux, limites |
| Serveur MCP (HTTP, OAuth) et jeton personnel pour les IA de terminal | Accès SSH à commande forcée |
| Windows statiques : documents, présentations, tableaux de bord, sites (HTML et ressources) | Commentaires, recherche |
| Deux versions par Window, retour à la précédente | Export PDF et PowerPoint |
| Accès de chaque Window : Room seulement, interne, Key, public, avec date de fin | SSO par OpenID Connect pour chaque Cahut |
| Plusieurs Cahuts par serveur, comptes propres à chaque Cahut, Rooms personnelles et d'équipe | L'envoi d'e-mails |
| Skill pour Claude, empaquetée en plugin, et consignes du serveur MCP | Sauvegardes S3 ou SFTP et restauration |
| Knock first et Logbook complets | Autres clients IA (Mistral, Copilot) |
| Console web sur son propre port | Tâches planifiées, quotas par Cahut |
| Mises à jour depuis les releases GitHub, avec retour arrière | |
| Claude et un deuxième client IA (ChatGPT ou Le Chat) | |

**Hypothèses.**

- Toute modification d'une Window passe par l'IA : pas d'éditeur en V0.
- Les premiers testeurs sont Olivier, puis quelques profils « geek sur VPS », sans exclure une petite équipe.
- Sans sauvegardes en V0, les testeurs gardent des instantanés de leur VPS.

## Principes de découpage

- **Un registre d'opérations, seule porte d'entrée.** Chaque opération est typée, classée réversible ou irréversible, et écrit au Logbook. MCP, l'API HTTP de la console et plus tard SSH n'en sont que des adaptateurs minces. Les règles du produit (opérations bornées, Knock first, tout est tracé, une seule API) tiennent ainsi par construction, dès la première ligne.
- **Des tranches verticales.** Chaque lot livre ses opérations, leurs outils MCP, leurs écrans de console et leurs entrées au Logbook, et se termine par une démonstration sur un vrai VPS.
- **Les risques d'abord.** MCP avec un vrai client, TLS automatique et le service de HTML produit par une IA arrivent au lot 1 ; OAuth avec deux clients au lot 3.
- **Un lot, une ou quelques pull requests** vers `main`, CI verte, CLAUDE.md tenu à jour.

## Les lots

| Lot | Objet | Durée indicative |
| --- | --- | --- |
| 0 | Socle et décisions | 2 à 3 jours |
| 1 | Tranche verticale : publier par MCP, servir en HTTPS | 2 semaines |
| 2 | Comptes, Rooms et console | 2 semaines |
| 3 | Brancher les IA : OAuth | 1,5 à 2 semaines |
| 4 | Windows complètes : versions et accès | 1,5 semaine |
| 5 | Knock first | 1 semaine |
| 6 | Skill Claude | 0,5 semaine |
| 7 | Installation, premier démarrage, mises à jour | 2 semaines |
| 8 | Deuxième client IA et recette | 1,5 à 2 semaines |

Les durées sont comptées à temps plein ; elles s'étirent selon la disponibilité.

### Lot 0 · Socle et décisions

- Décisions d'architecture écrites sous forme d'ADR, en anglais puisqu'elles seront publiques au lancement (voir « Décisions à prendre au lot 0 »).
- Squelette Go : `cmd/cahut`, paquets sous `internal/`, Makefile ou Taskfile.
- CI GitHub Actions : lint, tests, binaire statique pour linux/amd64 et linux/arm64 ; goreleaser pour les releases.
- CLAUDE.md complété avec les commandes de build et de test.

**Fin de lot** : la CI publie une release brouillon avec le binaire.

### Lot 1 · Tranche verticale : publier par MCP, servir en HTTPS

C'est la première tranche verticale de la feuille de route.

- Registre d'opérations, Logbook, SQLite avec migrations, stockage des fichiers des Windows sur disque.
- Plusieurs Cahuts par serveur dès le modèle de données : une base par Cahut et une base du serveur, une Cahut déclarée par son domaine en ligne de commande (ADR 8).
- Opérations `publish_window`, `get_window` et `list_windows`, exposées en outils MCP (Streamable HTTP) avec le SDK Go officiel.
- Jeton personnel créé en ligne de commande, pour Claude Code ; il servira ensuite aux IA de terminal.
- Service des Windows en HTTPS à `<domaine>/<room>/<window>/` (ADR 9), chacune dans un bac à sable CSP (ADR 10), un certificat ACME automatique par Cahut.
- Script `install.sh` rudimentaire : télécharge la release, crée l'utilisateur système et le service systemd.
- Mesure de la taille maximale d'une publication passée en argument d'outil.

**Fin de lot** : depuis Claude Code, « publie cette présentation sur Cahut » renvoie une adresse HTTPS qui s'ouvre sur le VPS, et l'action figure au Logbook.

### Lot 2 · Comptes, Rooms et console

Le détail du lot est dans le [plan du lot 2](plan-lot-2.md).

- Administrateurs du serveur, qui déclarent les Cahuts, et administrateurs de chaque Cahut, qui en gèrent les personnes ; mot de passe haché en argon2id, sessions. L'administrateur crée un compte et transmet un lien d'activation à usage unique : pas d'e-mail en V0.
- Room personnelle créée d'office pour chaque personne, au nom de son handle ; Rooms d'équipe avec propriétaires et membres ; les droits sont appliqués par le registre d'opérations.
- Coquille de la console : connexion, navigation, liste des Rooms et des Windows, lecture du Logbook, déclaration des Cahuts. Application légère embarquée dans le binaire, construite avec `cahut-ui.css`, qui n'appelle que l'API publique. Servie sur le port 8443 de chaque domaine, cloisonnée par Cahut, authentifiée par jeton et non par cookie (ADR 11), puis sous `/console/` sur le port 443 (ADR 20).

**Fin de lot** : deux Cahuts sur le même serveur ; dans l'une, deux comptes et une Room d'équipe ; chacun ne voit que ses Windows ; le Logbook se lit dans la console.

### Lot 3 · Brancher les IA : OAuth

- Serveur d'autorisation OAuth 2.1, un par Cahut, sur le port de la console : métadonnées de découverte, enregistrement dynamique des clients ou Client ID Metadata Documents selon ce qu'exigent les clients, code d'autorisation avec PKCE, jetons de rafraîchissement, révocation.
- Écran de consentement : l'utilisateur choisit les Rooms et la portée qu'il donne à l'IA.
- Écran des IA connectées : chaque IA, au nom de qui elle agit, sa portée, révocation en un clic.
- Test de fumée avec ChatGPT dès ce lot, pour lever tôt le risque du deuxième client.

**Fin de lot, jalon PoC** : Cahut ajoutée comme connecteur dans claude.ai, une présentation publiée dans une Room d'équipe, visible dans la console.

### Lot 4 · Windows complètes : versions et accès

Le détail du lot est dans le [plan du lot 4](plan-lot-4.md).

- `update_window` garde la version précédente ; `rollback_window` y revient.
- Accès de chaque Window : Room seulement (par défaut), interne (les personnes connectées de la Cahut), Key (un mot de passe), public ; date de fin réglable et prolongeable.
- La Key d'une Window est hachée et s'ouvre par une page d'accès servie par la Cahut. Les lecteurs d'une Window interne se connectent sur le port 443 en passant par la connexion de la console (ADR 11).
- Les Windows protégées sont servies dans un cadre (iframe) en bac à sable, par une adresse à capacité (ADR 10) ; tests sous Chrome, Firefox et Safari.
- La page d'une Room liste les Windows que son lecteur peut ouvrir ; une Window fermée répond comme une Window absente.
- La déclaration de l'IA (scripts, appels vers l'extérieur, formulaires) est enregistrée, affichée dans la console et appliquée par une politique CSP : une Window qui déclare ne rien appeler ne peut rien appeler.
- Mise à la corbeille, réversible.
- Console : fiche d'une Window avec ses versions, son accès et sa déclaration.
- En attendant le lot 5, l'IA ne peut pas rendre une Window publique.

**Fin de lot** : une personne non technique partage une présentation avec une Key depuis la console, le lien s'ouvre en navigation privée, puis on revient à la version précédente.

### Lot 5 · Knock first

Le détail du lot est dans le [plan du lot 5](plan-lot-5.md).

- Une opération irréversible demandée par une IA crée un Knock : une demande en attente, avec une expiration. L'IA reçoit le statut et le lien d'approbation, qu'elle transmet à l'utilisateur.
- Boîte des Knocks dans la console : approuver ou refuser. L'exécution est tracée avec l'IA, l'humain qui a approuvé et le résultat.
- Liste proposée pour la V0 : purger une Window, rendre une Window publique, élargir la portée d'une IA, ajouter une personne, changer les droits d'une personne (son rôle, son état, ses Rooms), supprimer une Room. Le lot 2 marque déjà celles qui existent ([ADR 13](adr/0013-people-sessions-and-rights.md)).
- Notification dans la console ; les notifications par e-mail viendront avec l'envoi d'e-mails.

**Fin de lot** : Claude demande à rendre une Window publique, la demande apparaît dans la console, l'approbation l'exécute, le Logbook montre les deux étapes.

### Lot 6 · Skill Claude

Le détail du lot est dans le [plan du lot 6](plan-lot-6.md).

- Skill Claude empaquetée en plugin : formats, accès, liens relatifs, bac à sable (pas de `localStorage`), Knock first. Instructions du serveur MCP pour les clients sans skill.

**Fin de lot** : avec la skill, un brief donne une présentation publiée en une seule demande.

### Lot 7 · Installation, premier démarrage, mises à jour

Le détail du lot est dans le [plan du lot 7](plan-lot-7.md).

- Une commande sur AlmaLinux et Ubuntu vierges : utilisateur système, service systemd avec `CAP_NET_BIND_SERVICE`, pare-feu (firewalld ou ufw) pour les ports 80 et 443, contextes SELinux sur AlmaLinux, commande `cahut doctor` pour diagnostiquer, rappel d'ouvrir ces ports chez l'hébergeur.
- Premier démarrage guidé dans le navigateur, ouvert par un lien à usage unique affiché en fin d'installation : administrateur du serveur, première Cahut et son domaine, première IA (adresse MCP à copier, pas à pas pour chaque client). Adresse de repli sans domaine à trancher : un certificat d'adresse IP est une piste.
- Mises à jour : vérification des releases GitHub signées, remplacement du binaire, contrôle de santé, retour arrière automatique.
- Test d'installation automatisé sur machines virtuelles vierges (Lima en local).

**Fin de lot** : installation chronométrée en moins de 10 minutes sur un VPS vierge de chaque système.

### Lot 8 · Deuxième client IA et recette

Le détail du lot est dans le [plan du lot 8](plan-lot-8.md).

- ChatGPT (mode développeur) ou Le Chat de bout en bout, avec les ajustements que chaque client demande : schémas d'outils, taille des contenus, consignes.
- Revue de sécurité : OAuth, bac à sable des Windows, cloisonnement entre Cahuts, sessions de la console et des lecteurs, limites de taille et de débit.
- Recette sur les cinq critères de réussite du README, ramenés au périmètre de la V0.
- README : section Installation réelle.

**Fin de lot, jalon V0** : les cinq critères passent sur deux VPS vierges, AlmaLinux et Ubuntu, avec deux clients IA.

## Critères de réussite de la V0

| Critère | Seuil |
| --- | --- |
| Installation sur un VPS vierge | Moins de 10 minutes, une commande, sur AlmaLinux et sur Ubuntu |
| Première publication par une IA | Moins de 5 minutes après le premier démarrage, sans toucher au serveur |
| Partage par une personne non technique | Une présentation partagée par lien, avec une Key, en quelques minutes |
| IA prises en charge | La même Cahut fonctionne avec Claude et un deuxième client |
| Gouvernance | Toute action d'une IA figure au Logbook, et s'annule ou a été approuvée |

## Décisions à prendre au lot 0

| Sujet | Recommandation |
| --- | --- |
| Web et TLS | Intégré au binaire avec certmagic plutôt que Caddy à côté : un seul binaire à installer et à mettre à jour |
| Origine des Windows | Une origine par Window (`<window>.home.example.com`) dès qu'un wildcard existe, certificats à la demande. Sinon une CSP `sandbox`, qui protège la console mais casse `localStorage` dans les pages. Le premier démarrage demande donc un wildcard, avec une adresse de repli de type sslip.io, à vérifier. Revu ensuite : un domaine par Cahut sans wildcard, les Windows en chemins dans un bac à sable CSP, la console sur son propre port (ADR 8 à 11), puis sous `/console/` du port 443 (ADR 20) |
| Base de données | SQLite en pur Go (`modernc.org/sqlite`), pour garder un binaire statique |
| Console | Petite application TypeScript (Preact ou Svelte, construite avec Vite) embarquée par `go:embed`, cliente de l'API publique : la règle d'une seule API se vérifie d'elle-même |
| OAuth | Bibliothèque éprouvée (fosite) ou implémentation minimale bien testée : à comparer |
| Compte GitHub | Organisation dédiée, à créer avant le lot 1 : son adresse est écrite dans le script d'installation et dans le module de mise à jour |

## Risques à surveiller

- **Taille des publications.** Une présentation avec des images passe mal en argument d'outil. À mesurer au lot 1 ; repli : un envoi en plusieurs opérations (ouvrir, ajouter des fichiers, valider).
- **Différences entre clients MCP.** OAuth, schémas d'outils et limites varient d'un client à l'autre, d'où le test avec ChatGPT dès le lot 3.
- **HTML non maîtrisé.** Une Window peut contenir n'importe quel script : le bac à sable CSP est la protection principale des autres Windows, de la console et des sessions. Il casse `localStorage` dans les pages : les consignes aux IA doivent le dire.
- **Cloisonnement entre Cahuts.** Une fuite d'une Cahut vers une autre serait grave pour un hébergeur ; une base par Cahut la rend structurellement difficile.
- **Port 8443.** Certains réseaux d'entreprise ne laissent sortir que le port 443 : la console y était inaccessible, les Windows non. Levé par l'ADR 20 : la console est servie sous `/console/` sur le port 443.
- **SELinux et pare-feu sur AlmaLinux.** C'est le principal danger pour le critère des 10 minutes.

## Après la V0

Proposition d'ordre pour atteindre le périmètre complet du lancement, en réponse à la question ouverte du brief :

1. Compléter les pages : commentaires, recherche, SSO par Cahut, envoi d'e-mails, sauvegardes et restauration, quotas par Cahut.
2. Les applis, le plus gros morceau : conteneurs sans privilèges, runtimes, mode Dockerfile, bases, secrets, journaux, limites. Sans sous-domaine, une appli n'a pas d'origine propre : une décision d'architecture à prendre avant ce lot.
3. Avant l'ouverture : export PDF et PowerPoint, accès SSH à commande forcée, autres clients IA.

## Questions ouvertes

- [x] Une Window ouverte par l'IA en interne ou avec une Key doit-elle frapper (Knock first) ? Non : seul l'accès public frappe ([ADR 19](adr/0019-window-versions-access-and-declaration.md)).
- [x] Quelle adresse de repli pour essayer Cahut sans domaine ? L'adresse IP du serveur, avec un certificat de Let's Encrypt pour cette adresse ([ADR 26](adr/0026-cahut-at-an-ip-address.md)).
- [x] La liste des actions irréversibles de la V0 est-elle la bonne ? Oui, telle que proposée ([ADR 21](adr/0021-ais-knock-first.md)) ; effacer des sauvegardes frappera quand elles viendront.
- [x] Quelle expiration par défaut pour un Knock ? 24 heures, sans réglage en V0 ([ADR 21](adr/0021-ais-knock-first.md)).
- [x] Le deuxième client de la recette : ChatGPT ou Le Chat ? ChatGPT, sur le web ; Le Chat est essayé aussi, sans être un critère ([ADR 29](adr/0029-chatgpt-second-ai-client.md)).
- [ ] Les hypothèses du périmètre tiennent-elles (pas d'éditeur, premiers testeurs « geek sur VPS ») ?
