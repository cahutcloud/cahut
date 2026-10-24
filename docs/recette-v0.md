# Cahut · Recette de la V0

23 octobre 2026 · Lot 8 · Olivier Durand

## En bref

La V0 est acquise quand les cinq critères du README passent sur deux VPS vierges, un sous AlmaLinux et un sous Ubuntu, avec Claude et ChatGPT. Ce document dit comment le vérifier, pas à pas et chronomètre en main, et garde les résultats de chaque passage. Une partie de la recette tourne aussi à chaque changement, en CI : le test d'installation sur un Ubuntu vierge, et un test qui rejoue les quatre autres critères sur un Cahut de test.

## Les critères

| # | Critère | Seuil de la V0 |
| --- | --- | --- |
| 1 | Installation sur un VPS vierge | Moins de 10 minutes, une commande, sur AlmaLinux et sur Ubuntu |
| 2 | Première publication par une IA | Moins de 5 minutes après l'ouverture du lien du premier démarrage, sans toucher au serveur |
| 3 | Partage par une personne non technique | Une présentation partagée par lien, avec une Key, en quelques minutes, depuis la console seule |
| 4 | IA prises en charge | Le même Cahut fonctionne avec Claude et ChatGPT : chacun publie, met à jour et partage une Window |
| 5 | Gouvernance | Toute action d'une IA figure au Logbook, et s'annule ou a été approuvée |

## Ce que la CI vérifie à chaque changement

- **Critère 1.** Le job *Install* lance [test/install/run.sh](../test/install/run.sh) sur l'Ubuntu vierge de GitHub : `install.sh`, `cahut doctor`, une mise à jour et un retour arrière. `make install-test` fait de même sous Lima, sur AlmaLinux 10 et Ubuntu 24.04. Ni l'un ni l'autre n'obtient de certificat : ils n'ont pas de domaine public.
- **Critères 2 à 5.** `TestV0Acceptance` ([internal/web/acceptance_test.go](../internal/web/acceptance_test.go)) rejoue sur un Cahut de test :
  - le premier démarrage par son lien et la house style ;
  - un client qui s'identifie par son document, comme ChatGPT, lit la house style et publie une présentation ;
  - la personne lui donne une Key par l'API de la console, et un lecteur l'ouvre avec la Key ;
  - un second client, qui s'enregistre comme Le Chat, met la présentation à jour puis revient à la version précédente ;
  - la demande de la rendre publique frappe, la personne approuve, la présentation s'ouvre à tous ;
  - le Logbook nomme chaque appel, son client et sa personne, et chacun ne change rien, s'annule ou a été approuvé.
- **Critère 5, sur le registre.** `TestEveryAIOperationIsReversibleOrApproved` vérifie que chaque opération offerte aux IA ne change rien, frappe d'abord, ou a son retour en arrière nommé. Une nouvelle opération qui n'est dans aucun des trois cas fait échouer le test.

Ces tests ne remplacent pas la recette sur de vrais serveurs : ils ne voient ni le pare-feu de l'hébergeur, ni Let's Encrypt, ni les serveurs d'Anthropic et d'OpenAI.

## Matériel

- Deux VPS vierges chez un hébergeur, à l'image standard : AlmaLinux 10 et Ubuntu 24.04, 2 vCPU, 4 Go de mémoire, 40 Go de disque, une adresse IPv4 publique, les ports 80 et 443 ouverts chez l'hébergeur.
- Un domaine ou sous-domaine pour chacun, dont l'enregistrement A pointe sur son adresse avant de commencer.
- La release à recetter, publiée et signée.
- Un compte claude.ai, un compte ChatGPT qui ajoute un serveur MCP personnalisé (selon l'aide d'OpenAI d'octobre 2026 : offre Business, Enterprise ou Edu, dans Plugins), et en option un compte Le Chat.
- Un navigateur avec une fenêtre de navigation privée, et un chronomètre.

## Déroulé

À faire sur chaque VPS. Noter chaque temps et chaque écart dans le tableau des résultats.

### 1. Installation

Se connecter en SSH au VPS, démarrer le chronomètre, puis :

```bash
curl -fsSL https://cahut.cloud/install.sh | sudo sh -s -- --domain <domaine>
```

**On doit voir** la signature de la release vérifiée, le service actif, le pare-feu ouvert (firewalld sous AlmaLinux, ufw sous Ubuntu), `cahut doctor` sans ligne `fail`, puis le lien du premier démarrage. Arrêter le chronomètre à l'affichage du lien : **moins de 10 minutes**.

### 2. Premier démarrage et première publication

Démarrer un second chronomètre et ouvrir le lien.

1. **Vous** : nom, handle, mot de passe de 15 caractères au moins.
2. **Votre Cahut** : le nom de l'organisation et une couleur.
3. **Votre première IA** : dans claude.ai, Settings, Connectors, ajouter un connecteur personnalisé à l'adresse MCP affichée ; se connecter quand claude.ai renvoie vers la console, garder « Read and publish » et toutes les Rooms. L'écran du premier démarrage dit que Claude est branché.
4. Demander à Claude : « Fais une présentation de trois diapositives sur nos résultats du trimestre, à la house style de Cahut, et publie-la. »

**On doit voir** Claude appeler `get_house_style` puis `publish_window`, et rendre une adresse `https://<domaine>/<handle>/<window>/`. Arrêter le chronomètre quand la présentation s'ouvre dans la console : **moins de 5 minutes**, sans commande sur le serveur.

### 3. Partage par une personne non technique

Sans terminal, dans la console :

1. Ouvrir la présentation, choisir l'accès « Key », et laisser Cahut tirer une Key (« A Key drawn by Cahut »).
2. Copier l'adresse et la Key.
3. Dans une fenêtre de navigation privée, ouvrir l'adresse, taper la Key.

**On doit voir** la présentation, à la house style, et pouvoir passer les diapositives. Le tout prend **quelques minutes**.

### 4. Deuxième client : ChatGPT

1. Dans ChatGPT sur le web, Plugins, le bouton plus, puis « Add custom MCP server » : un nom, l'adresse MCP du Cahut, OAuth ; accepter l'avertissement, créer le plugin.
2. Se connecter quand ChatGPT renvoie vers la console, garder « Read and publish » et toutes les Rooms.
3. Dans une conversation, choisir le plugin et demander : « Mets à jour la présentation des résultats du trimestre : ajoute une diapositive de conclusion, à la house style. »
4. Puis : « Reviens à la version précédente de cette présentation. »

**On doit voir** ChatGPT demander une confirmation avant chaque appel qui change quelque chose, lire la house style, publier la nouvelle version, puis revenir à la précédente. Le lien avec la Key montre chaque fois la version en place.

### 5. Gouvernance

1. Demander à ChatGPT : « Rends cette présentation publique. »
2. ChatGPT rend le lien d'un Knock ; la console le montre dans Knocks. L'approuver.
3. Ouvrir l'adresse de la présentation en navigation privée, sans Key.
4. Dans la console, ouvrir le Logbook.

**On doit voir** le Knock approuvé puis exécuté, la présentation publique, et au Logbook chaque appel de Claude et de ChatGPT, au nom de la personne, avec le client qui l'a fait. Chaque appel ne change rien, s'annule (publication, mise à jour, retour arrière, accès par Key) ou a été approuvé (l'accès public, avec son Knock).

### 6. En option

- **Le Chat** : dans Le Chat, Connectors, ajouter un connecteur MCP personnalisé à l'adresse du Cahut ; Le Chat trouve OAuth seul. Lister les Windows, puis en publier une.
- **Sans domaine** : sur un troisième VPS, ou à la place du domaine sur l'Ubuntu, `install.sh` sans `--domain` installe le Cahut à l'adresse IPv4 du serveur, avec un certificat de Let's Encrypt pour cette adresse. Brancher claude.ai et ChatGPT à `https://<adresse>/mcp`, puis, dans la console, « Move to a domain ».

## Résultats

| Date | Système | Version | Installation | Première publication | Partage | Clients | Gouvernance | Remarques |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| 23/10/2026 | AlmaLinux 10.2 arm64, Lima | 0.4.0 | 59 s | | | | | Lot 7 : `install.sh` sans certificat, test d'installation entier en 152 s |
| 23/10/2026 | Ubuntu 24.04 arm64, Lima | 0.4.0 | 59 s | | | | | Lot 7, idem |
| 23/10/2026 | Ubuntu 24.04 amd64, CI | lot 8 | 49 s | | | | | Job *Install*, à chaque pull request |
| 23/10/2026 | Cahut de test, CI | lot 8 | | passe | passe | passe (CIMD et DCR) | passe | `TestV0Acceptance` |
| 23/10/2026 | AlmaLinux 10.2 amd64, staging, Cahut à `51.83.68.4` | lot 8 | | | passe, avec la Key | Codex et Vibe, par Internet, jetons personnels | passe : accès public en Knock | Certificat d'adresse IP de Let's Encrypt, `shortlived` ; pas un VPS vierge |
