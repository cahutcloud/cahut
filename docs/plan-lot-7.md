# Cahut · Plan du lot 7 : installation, premier démarrage, mises à jour

23 octobre 2026 · Olivier Durand · Validé

## En bref

Le lot 7 rend Cahut installable par quelqu'un qui ne le connaît pas. Une commande sur un AlmaLinux ou un Ubuntu vierge installe le service, ouvre le pare-feu, vérifie la signature de la release et finit par un lien à usage unique : il ouvre dans le navigateur un premier démarrage guidé, où l'on choisit son compte, règle la house style de base et branche sa première IA. Sans domaine, Cahut s'installe à l'adresse IP du serveur, avec un certificat de Let's Encrypt pour cette adresse, et passe à un domaine plus tard sans rien perdre. `cahut doctor` dit ce qui ne va pas. Les mises à jour viennent des releases GitHub signées : Cahut remplace son binaire, vérifie sa santé et revient seul en arrière si la nouvelle version ne démarre pas. Un test d'installation automatisé tourne sur des machines virtuelles vierges. Le lot compte neuf pull requests, pour environ onze jours à temps plein, et s'ouvre sur trois décisions écrites (ADR 24, 25 et 26).

**Fin de lot**, reprise du [plan de la V0](plan-v0.md) : installation chronométrée en moins de 10 minutes sur un VPS vierge de chaque système.

## Point de départ

Ce que les lots 1 à 6 ont laissé :

- [install.sh](../install.sh) télécharge une release, vérifie sa somme SHA-256, ou prend un binaire donné par `--binary`. Il crée l'utilisateur système `cahut`, `/var/lib/cahut`, `/etc/cahut/cahut.env` et le service systemd, avec `CAP_NET_BIND_SERVICE` et un durcissement complet (`ProtectSystem=strict`, `NoNewPrivileges`), déclare le Cahut de `--domain`, ouvre les ports 80 et 443 dans firewalld ou ufw et rétablit les contextes SELinux du binaire et de l'unité. Il finit par les commandes à lancer pour créer son compte. Relancé, il met à jour, sans contrôle de santé ni retour arrière.
- Le staging, sous AlmaLinux 10.2, tourne ainsi depuis le lot 1 ; chaque fin de lot y installe un binaire à la main.
- Les releases sont des brouillons que GoReleaser construit au push d'une étiquette `v*`, avec un fichier de sommes, sans signature. L'[ADR 7](adr/0007-module-path-repository-and-releases.md) renvoie la signature au lot 7. Le dépôt est privé : lire les releases demande un jeton.
- Un Cahut n'existe qu'à un domaine ([ADR 8](adr/0008-several-cahuts-per-server.md)) ; `add_domain` refuse les adresses IP. CertMagic obtient un certificat par Cahut, à la demande pour un domaine déclaré en marche.
- Les administrateurs du serveur et les personnes de chaque Cahut ont des comptes distincts ([ADR 13](adr/0013-people-sessions-and-rights.md)). Une personne naît d'un lien d'activation à usage unique, où elle choisit son mot de passe.
- La console a un écran AIs and access, avec l'adresse MCP à copier et la commande de Claude Code, et la vue d'ensemble d'un administrateur du serveur, qui déclare les Cahuts. `/-/healthz` répond `ok` sur le port 443.
- Une base de données plus récente que le binaire s'ouvre sans rien dire : `store.Open` n'applique que les migrations qui manquent.

## Décisions à écrire d'abord

### ADR 24 · Releases are signed, and servers update themselves with a rollback

L'[ADR 24](adr/0024-signed-releases-and-updates.md) fixe :

- **La signature.** Une clé Ed25519 signe le fichier des sommes de chaque release : `checksums.txt.sig`, 64 octets. La clé privée est un secret du dépôt, `CAHUT_RELEASE_KEY`, que seul le workflow des releases lit ; le mainteneur principal en garde une copie hors ligne. La clé publique est dans le binaire et dans `install.sh`. Une nouvelle clé arrive dans une release signée par l'ancienne. Les releases sont signées à partir de la 0.4.0.
- **La vérification.** `install.sh` vérifie la signature avec `openssl` (3.0 ou plus, présent sur AlmaLinux 9 et 10, Ubuntu 22.04 et 24.04), puis la somme de l'archive ; `cahut update` fait de même en Go, sans dépendance.
- **`cahut update`**, lancé par root : lit les releases publiées, prend la dernière ou celle de `--version`, refuse de revenir à une version plus ancienne, télécharge l'archive, sa somme et la signature, les vérifie, essaie le binaire, arrête le service, copie les bases de données, remplace le binaire, redémarre, attend que le service réponde avec la nouvelle version et ouvre chaque Cahut. Sinon, il remet l'ancien binaire et les bases, redémarre et le dit. `--binary` installe un binaire local avec les mêmes garde-fous, sans signature, comme `install.sh --binary`.
- **L'automatique.** Un minuteur systemd lance `cahut update --unattended` chaque jour. `CAHUT_UPDATES`, dans `/etc/cahut/cahut.env`, dit ce qu'il fait : `patch` (par défaut) installe seul les versions correctives de la version mineure en place et signale les autres, `all` installe toute nouvelle release, `check` signale seulement, `off` ne demande rien à GitHub.
- **La séparation des droits.** Le service, sous l'utilisateur `cahut`, ne peut ni remplacer son binaire ni se redémarrer, et ne parle jamais à GitHub. Root met à jour ; l'état des mises à jour est dans `/var/lib/cahut-update/`, que le service lit. Depuis la console, un administrateur du serveur demande une mise à jour : le service écrit une demande, une unité `.path` de systemd la remet au programme de mise à jour, qui ne lit qu'un numéro de version, ne descend jamais et vérifie la signature. Root n'ouvre jamais les bases : il copie des fichiers, service arrêté.
- **Les bases plus récentes que le binaire** ne s'ouvrent plus : le message dit qu'une version plus récente de Cahut les a faites, et qu'il faut la réinstaller.
- **Logbook.** Le service, au démarrage, écrit au Logbook du serveur chaque mise à jour qu'il n'y a pas encore vue, réussie ou défaite, d'après l'état que root a laissé. La demande faite dans la console y figure aussi.

### ADR 25 · The first start is guided in the browser, from a single-use link

L'[ADR 25](adr/0025-first-start-in-the-browser.md) fixe :

- **Le lien.** L'installation finit par un lien `https://<adresse>/console/setup#<code>`, à usage unique, valable 24 heures, haché dans la base du serveur et lié à un Cahut. Il ne sert que tant que ce Cahut n'a personne. `cahut setup` en donne un nouveau, qui annule le précédent.
- **Les étapes.** Vous : nom, handle, mot de passe ; la première personne du Cahut, administratrice, et, tant que le serveur n'en a aucun, son administrateur, avec le même handle et le même mot de passe, en deux comptes, comme le veut l'ADR 13. Votre Cahut : son adresse, et à une adresse IP comment passer à un domaine ; le nom de l'organisation, sa couleur principale et son logo, pour la house style, ou la house style neutre. Votre première IA : l'adresse MCP à copier, et le pas à pas de Claude Code, des applications Claude, de ChatGPT et des autres clients MCP ; l'écran attend l'IA et dit quand elle est branchée. Puis la console.
- **Les opérations.** Le lien se crée en ligne de commande (`create_setup_link`), se lit (`get_setup`) et sert une fois (`set_up`), sans compte, au même débit que la connexion. `set_up` passe par les opérations qui existent, la création d'une personne et son activation, et ouvre la session de la personne. La house style et l'IA passent ensuite par les opérations de toujours.
- **Après coup.** Tant que personne n'a branché d'IA, la vue d'ensemble de la console propose de le faire, avec le même pas à pas.

### ADR 26 · A Cahut can live at the IP address of its server until it has a domain

L'[ADR 26](adr/0026-cahut-at-an-ip-address.md) fixe :

- **L'adresse de repli.** Sans domaine, un Cahut se déclare à l'adresse IP publique de son serveur, IPv4 ou IPv6 : `https://203.0.113.7/<room>/<window>/`. Seules les adresses publiques et globales passent. Une adresse ne porte qu'un Cahut.
- **Le certificat.** Let's Encrypt délivre des certificats d'adresse IP depuis janvier 2026, au profil `shortlived` seulement, valables 160 heures ; CertMagic les renouvelle. Les domaines gardent le profil classique, de 90 jours. Un client qui ouvre une adresse IP n'envoie pas de nom (SNI) : Cahut choisit le certificat par l'adresse locale de la connexion.
- **Passer à un domaine.** `rename_domain`, en ligne de commande (`cahut domain rename`) et dans la console, change l'adresse d'un Cahut : ses personnes, ses Rooms, ses Windows, sa house style et son Logbook restent. L'ancienne adresse ne sert plus rien ; les liens qui la nomment cessent de marcher, et les IA se rebranchent à la nouvelle. Le serveur relit le domaine de chaque Cahut à chaque requête, pour qu'un changement fait en ligne de commande prenne sans redémarrage.
- **L'installation.** `install.sh --ip <adresse>` déclare le Cahut à cette adresse ; sans `--domain` ni `--ip`, le script prend l'adresse source de la route par défaut si elle est publique, et demande sinon l'une des deux options.
- **Les limites.** Un certificat de six jours dépend de Let's Encrypt chaque semaine ; un client IA peut refuser une adresse IP, à vérifier avec chaque client au lot 8. L'adresse IP sert à essayer, le domaine reste la règle.

## Modèle de données

Deux migrations :

| Migration | Contenu |
| --- | --- |
| `server/0006_setup_links` | Table `setup_links` : Cahut, empreinte du code, expiration, date d'usage. Un lien par Cahut au plus |
| `server/0007_updates` | Table `updates_logged` : l'identifiant de chaque mise à jour déjà écrite au Logbook du serveur |

L'état des mises à jour vit hors des bases, dans `/var/lib/cahut-update/` (root, lisible du groupe `cahut`) :

| Fichier | Contenu |
| --- | --- |
| `status.json` | La dernière vérification (date, dernière release, ce qu'elle apporte), la dernière mise à jour (identifiant, de, vers, début, fin, résultat, raison), la règle `CAHUT_UPDATES` |
| `previous/cahut` | Le binaire d'avant la dernière mise à jour |
| `snapshot/` | Les bases de données copiées avant la dernière mise à jour, root seul |

La demande de la console est un fichier de `/var/lib/cahut`, `update-request`, qui ne contient qu'un numéro de version ou `check` ; root le lit sans suivre de lien symbolique, 64 octets au plus, puis l'efface.

## Opérations

| Opération | Qui l'appelle | Ce qu'elle fait |
| --- | --- | --- |
| `create_setup_link` | administrateurs du serveur, la ligne de commande | Le lien du premier démarrage d'un Cahut qui n'a personne ; annule le précédent |
| `get_setup` | sans compte | Dit si un code sert encore, et pour quel Cahut |
| `set_up` | sans compte, `PeopleOnly` | Consomme le code, crée la première personne, administratrice, et l'administrateur du serveur s'il n'en a aucun ; ouvre la session de la personne |
| `rename_domain` | administrateurs du serveur | Change le domaine ou l'adresse d'un Cahut |
| `get_updates` | administrateurs du serveur | La version en marche, la dernière release connue, la règle, la dernière vérification et la dernière mise à jour |
| `request_update` | administrateurs du serveur | Demande à root de vérifier, ou d'installer une version plus récente que celle en marche |

`add_domain` accepte désormais une adresse IP publique.

## Commandes et installation

| Commande | Ce qu'elle fait |
| --- | --- |
| `cahut setup [--domain D]` | Donne le lien du premier démarrage du Cahut, le premier du serveur par défaut |
| `cahut domain rename OLD NEW` | Change l'adresse d'un Cahut |
| `cahut update [--check] [--version V] [--binary PATH] [--unattended]` | Met à jour, en root, avec retour arrière |
| `cahut doctor` | Diagnostique le serveur ; en root, il vérifie aussi le pare-feu et SELinux |

**`cahut doctor`** dit, ligne à ligne, `ok`, `warn` ou `fail`, et ce qu'il faut faire : le système et l'horloge synchronisée, dont dépend ACME ; le service installé, activé, actif, et la version qui répond ; le dossier des données, son propriétaire et l'espace libre ; les ports 80 et 443 ; firewalld ou ufw ; SELinux et le contexte du binaire ; pour chaque Cahut, ce que son domaine résout et si c'est une adresse de la machine, son certificat et son échéance, et une requête HTTPS à son adresse depuis le serveur ; Let's Encrypt et GitHub joignables ; le minuteur des mises à jour et leur dernier résultat ; et le rappel d'ouvrir les ports chez l'hébergeur. Il n'ouvre pas les bases : il demande au service, qui répond sur le port 80 de la boucle locale, `/-/healthz`, avec sa version et l'état de chaque Cahut.

**`install.sh`** :

- vérifie la signature de la release avec `openssl`, puis la somme ;
- accepte `--ip`, ou trouve l'adresse publique sans `--domain` ;
- installe `cahut-update.service`, `cahut-update.timer` et `cahut-update.path`, et écrit `CAHUT_UPDATES=patch` à la première installation ; avec `GITHUB_TOKEN`, il le garde pour les mises à jour dans `/etc/cahut/update.env`, lisible de root seul, tant que le dépôt est privé ;
- lance `cahut doctor` à la fin, puis donne le lien du premier démarrage si le Cahut n'a personne, et les commandes d'avant sinon ;
- est servi par le site, à `https://cahut.cloud/install.sh`, copié du dépôt à chaque construction : `curl -fsSL https://cahut.cloud/install.sh | sudo sh -s -- --domain cahut.example.com`.

Le service de Cahut ne change pas de durcissement. Le service de mise à jour tourne en root, sans durcissement qui l'empêcherait de remplacer le binaire, de redémarrer `cahut` ou de copier les bases.

## Console

| Écran | Contenu |
| --- | --- |
| Premier démarrage, `/setup` | Plein écran, à la marque de Cahut. Trois étapes et leur progression : Vous, Votre Cahut, Votre première IA. Un lien déjà servi ou expiré dit comment en obtenir un autre (`sudo /usr/local/bin/cahut setup`) |
| Vue d'ensemble | Une carte « Connect your first AI » tant que la personne n'a ni IA branchée ni jeton, avec le pas à pas |
| AIs and access | Le pas à pas par client remplace la seule commande de Claude Code |
| This server · Updates, `/server/updates` | Pour les administrateurs du serveur : la version en marche, la dernière release et ses notes, la règle, la dernière vérification, la dernière mise à jour et son résultat ; « Check now » et « Update to 0.x.y » |
| Vue d'ensemble du serveur | « Move to a domain » sur un Cahut à une adresse IP, puis sur tout Cahut : le domaine, ce qu'il résout, la confirmation |

Le pas à pas par client est un composant partagé : Claude Code (le plugin `cahut`, ou `claude mcp add` puis `claude mcp login`), les applications Claude (un connecteur personnalisé, et l'archive de la skill), ChatGPT (mode développeur, un connecteur à l'adresse MCP), les autres clients (l'adresse MCP avec OAuth, ou un jeton personnel).

## Tests

- **Signature** : une release bien signée passe ; une signature fausse, une somme fausse, une archive changée, une clé inconnue échouent, en Go et dans `install.sh`.
- **Versions** : comparaison des numéros, versions de développement comprises ; pas de descente ; la règle `patch` ne prend que les correctifs.
- **Mise à jour** : avec un faux GitHub et un faux systemd, une mise à jour réussie, un binaire qui ne démarre pas, un service qui répond avec l'ancienne version, une migration qui échoue sur un Cahut ; chaque échec rend l'ancien binaire et les anciennes bases. La demande de la console : un lien symbolique, un contenu trop long, une version plus ancienne sont refusés.
- **Bases** : une base plus récente que le binaire est refusée, avec un message qui dit quoi faire.
- **Santé** : `/-/healthz` sur la boucle locale dit la version et ouvre chaque Cahut ; d'ailleurs, il ne dit que `ok`.
- **Adresse IP** : `add_domain` prend une adresse publique, refuse une adresse privée, de boucle ou de lien local ; un Cahut à une adresse IP sert ses Windows, sa console et son MCP ; `rename_domain` garde tout, l'ancienne adresse répond 421, sans redémarrage ; le cloisonnement tient.
- **Premier démarrage** : un code sert une fois, expire, ne sert plus quand le Cahut a quelqu'un ; l'administrateur du serveur n'est créé que s'il n'y en a aucun ; Logbook sans mot de passe ni code.
- **Diagnostic** : chaque vérification, avec des réponses simulées du système.
- **Installation** : en CI, sur la machine Ubuntu de GitHub, `install.sh --binary`, le service actif, `cahut doctor`, `cahut update --binary` avec un binaire qui ne démarre pas puis avec un bon. En local, avec Lima, la même chose sur AlmaLinux 10 et Ubuntu 24.04 vierges, chronométrée.
- **Navigateur** : le premier démarrage à 1 280 px et à 375 px, l'écran Updates, le déplacement vers un domaine.

## Découpage en pull requests

| # | Pull request | Contenu | Durée |
| --- | --- | --- | --- |
| 1 | Décisions | Ce plan et les ADR 24, 25 et 26 | 0,5 j |
| 2 | Releases signées | La clé, l'outil de signature et GoReleaser, le paquet `release` (GitHub, versions, vérification), la vérification dans `install.sh`, les bases plus récentes refusées | 1 j |
| 3 | Mises à jour | `cahut update`, les copies des bases, la santé sur la boucle locale, le retour arrière, les unités systemd, `CAHUT_UPDATES`, l'état, la demande de la console, le Logbook, `get_updates` et `request_update` | 2 j |
| 4 | Diagnostic | `cahut doctor`, lancé par `install.sh` | 1 j |
| 5 | Adresse IP | Un Cahut à une adresse IP, les certificats `shortlived`, le choix sans SNI, `rename_domain`, `install.sh --ip` | 1,5 j |
| 6 | Premier démarrage | Migration, les trois opérations, `cahut setup`, l'écran du premier démarrage, le pas à pas par client, la carte de la vue d'ensemble, le lien en fin d'installation | 2 j |
| 7 | Console du serveur | L'écran Updates, « Move to a domain » | 1 j |
| 8 | Tests d'installation | Le test en CI, le script Lima, `install.sh` servi par le site | 1 j |
| 9 | Documentation et démonstration | Pages de la documentation, référence, site, CLAUDE.md, staging, bilan | 1 j |

Soit environ 11 jours. Les pull requests s'enchaînent dans l'ordre.

## Démonstration de fin de lot

**Prérequis** : la release 0.4.0, la première signée, publiée ; un VPS vierge sous AlmaLinux 10 et un sous Ubuntu 24.04 ; un domaine qui pointe sur l'un ; Claude Code.

1. Sur le VPS AlmaLinux, chronomètre en main : `curl -fsSL https://cahut.cloud/install.sh | sudo sh -s -- --domain <domaine>`. La sortie montre la signature vérifiée, le pare-feu ouvert, `cahut doctor` sans échec, puis le lien du premier démarrage.
2. Ouvrir le lien : créer son compte, nommer l'organisation, choisir une couleur, brancher Claude Code. L'écran dit que Claude est branché. Arrêter le chronomètre : moins de 10 minutes.
3. Demander à Claude de publier une présentation : elle s'ouvre à l'adresse qu'il rend, moins de 5 minutes après le premier démarrage.
4. Sur le VPS Ubuntu, sans domaine : `curl … | sudo sh`. Cahut s'installe à l'adresse IP, avec un certificat valide ; le premier démarrage s'y fait de même. Puis, dans la console, « Move to a domain » le passe au domaine, et les Windows le suivent.
5. Sur le staging, `sudo /usr/local/bin/cahut update` passe de la version du lot à la 0.4.0 publiée ; `cahut update --binary` avec un binaire qui ne démarre pas revient seul à la 0.4.0, et le Logbook du serveur montre les deux.

## Hors du lot 7

- Les sauvegardes S3 ou SFTP et leur restauration : après la V0. Les testeurs gardent des instantanés de leur VPS ; la copie des bases avant une mise à jour n'en tient pas lieu.
- Le retour arrière à la main, des jours après une mise à jour : il perdrait ce qui a changé depuis. Une base plus récente que le binaire est refusée ; il faut réinstaller une version au moins aussi récente que celle qui l'a faite.
- Les mises à jour par un dépôt de paquets (dnf, apt) et les images de conteneur : après la V0.
- Un domaine fourni par le projet, du type `<adresse>.cahut.cloud` : écarté par l'ADR 26.
- Les autres systèmes que AlmaLinux et Ubuntu : Debian et Rocky devraient marcher, sans être testés.
- La section Installation du README : lot 8.
