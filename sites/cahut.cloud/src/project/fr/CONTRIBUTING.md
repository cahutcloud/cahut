# Contribuer à Cahut

Merci de votre intérêt. Cahut est maintenue par une seule personne, en public. Cette page dit ce qui aide, ce qu'il faut discuter d'abord, et comment un changement entre. [GOVERNANCE.md](GOVERNANCE.md) dit comment le projet est mené ; le [code de conduite](CODE_OF_CONDUCT.md) s'applique partout.

## Bienvenu tout de suite

Ouvrez une pull request sans demander d'abord pour :

- une correction de bug, avec un test qui montre le bug ;
- une meilleure page de la documentation, ou une faute corrigée dans le site ou dans un message ;
- une traduction du site : les mots anglais sont dans `sites/cahut.cloud/src/i18n/en.ts` et les mots français dans `fr.ts`, changés ensemble, et les pages françaises de la section Projet sont dans `sites/cahut.cloud/src/project/fr/` ;
- un message d'erreur plus clair, qui dit à une IA comment corriger son appel ;
- une mise à jour de dépendance.

## À discuter d'abord

Ouvrez une Discussion dans *Ideas*, ou un ticket, avant d'écrire du code pour :

- une fonctionnalité ;
- une nouvelle opération, et donc un nouvel outil MCP ;
- un changement de l'interface ou du vocabulaire ;
- une nouvelle dépendance, avec sa raison ; une dépendance ne doit pas avoir besoin de cgo ;
- tout ce qui touche la sécurité, le bac à sable ou knock first.

Une pull request de mille lignes qui arrive sans prévenir est fermée gentiment, quelle que soit sa qualité. Cela protège votre temps autant que le nôtre.

## Ce que Cahut ne fera pas

Le README le dit : créer du contenu, faire tourner un modèle d'IA sur le serveur, donner un shell à une IA, dépendre d'un fournisseur, s'étendre sur plusieurs machines. Pas non plus de version réduite ou fermée, de télémétrie, ni de fonctionnalité réservée à une offre payante. La documentation garde ces décisions et leurs raisons, pour que la question ne se pose pas deux fois.

## Comment une pull request entre

1. Une branche par sujet, une pull request par sujet, petite quand c'est possible.
2. Une fonctionnalité nouvelle, ou un changement de comportement, vient avec ses tests : ils montrent ce qu'elle fait et l'empêchent de casser plus tard. Une correction de bug vient avec un test qui échoue sans elle.
3. `make check` passe : le lint, govulncheck et les tests, tels que la CI les lance.
4. Le code suit [CLAUDE.md](CLAUDE.md) : l'anglais partout, des sujets de commit à l'impératif, chaque action par le registre d'opérations, pas de logique métier dans les adaptateurs, la bibliothèque standard d'abord. Ce fichier est écrit pour les IA qui travaillent dans ce dépôt, et il s'applique à tout le monde.
5. La description dit ce qui change, pourquoi, et comment cela a été testé. Pour une fonctionnalité, elle renvoie à la discussion.
6. Le mainteneur relit dans un créneau hebdomadaire. Il demande des changements, fusionne, ou explique pourquoi non.
7. Une pull request sans réponse de son auteur pendant trente jours est fermée. Rouvrez-la quand vous voulez.

### Signez vos commits

Cahut utilise le [Developer Certificate of Origin](https://developercertificate.org/). Chaque commit porte la ligne `Signed-off-by: Votre Nom <vous@example.com>`, ajoutée par `git commit -s`, par laquelle vous certifiez avoir le droit de contribuer ce code sous la licence du projet. Rien n'est cédé, il n'y a rien d'autre à signer, et aucun avocat n'intervient. Il n'y a pas de contrat de licence de contribution, et il n'y en aura pas : c'est ainsi que la licence reste ce qu'elle est ([ADR 15](docs/adr/0015-license.md)).

### Travailler avec une IA

Cahut est écrite avec Claude Code, et [CLAUDE.md](CLAUDE.md) est public. Les contributions faites avec une IA sont bienvenues, à cinq règles :

1. **Dites-le.** Nommez l'outil et sa part du travail, dans la pull request ou dans un pied de commit tel que `Assisted-by: Claude Code`.
2. **Comprenez tout.** La personne qui signe a lu, compris, lancé et testé le changement, et répond elle-même aux relectures. Une réponse générée à une relecture ferme la pull request.
3. **La signature est la vôtre.** `Signed-off-by` certifie ce que vous avez fait, pas ce que l'outil a fait : un agent ne l'ajoute que sur votre instruction et en votre nom, et aucun agent n'ouvre de ticket ni de pull request sans qu'une personne l'ait lu.
4. **Reproduisez avant de signaler.** Un ticket écrit par une IA sans reproduction est fermé. Un rapport de sécurité inventé vaut un bannissement à son auteur.
5. **Les visuels viennent de la charte.** Aucune image ni illustration générée dans le site ou dans la documentation.

## Signaler un bug

Utilisez le formulaire de rapport de bug. Donnez la version de `cahut version`, le système, comment Cahut a été installée, ce que vous attendiez, ce qui s'est passé, et les lignes du Logbook qui vont avec, sans aucun jeton. Les questions vont dans les Discussions ([SUPPORT.md](SUPPORT.md)). Les failles passent par [SECURITY.md](SECURITY.md), jamais par un ticket.

## La reconnaissance

Les notes de release nomment les contributeurs de chaque release. Les étiquettes `good first issue` et `help wanted` disent par où commencer. Il n'y a pas de programme de primes.
