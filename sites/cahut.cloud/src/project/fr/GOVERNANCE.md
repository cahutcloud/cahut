# Comment Cahut est menée

Cahut est un projet open source personnel d'Olivier Durand, indépendant de toute entreprise et de tout fournisseur d'IA. Une seule personne la maintient aujourd'hui. Cette page dit comment les décisions se prennent, qui fait quoi, ce que le projet promet, et comment cela change à mesure que le projet grandit. Les raisons sont dans [docs/gouvernance.md](docs/gouvernance.md), en français.

## Les principes

1. **Le mainteneur décide, et écrit pourquoi.** Olivier a le dernier mot. Chaque décision qui compte devient un ADR dans [docs/adr](docs/adr/README.md), avec les options écartées.
2. **Un non est provisoire, un oui est pour toujours.** Une fonctionnalité acceptée se maintient des années. Une fonctionnalité refusée peut revenir avec de meilleurs arguments.
3. **Ouvert au regard, encadré pour la contribution.** Tout se lit, se lance, se teste et se signale. Les corrections entrent directement, les fonctionnalités se discutent d'abord, et certaines choses ne se feront pas. [CONTRIBUTING.md](CONTRIBUTING.md) dit lesquelles.
4. **Le temps est la ressource rare, et c'est dit.** Un créneau fixe par semaine, une réponse dans la semaine, aucune promesse au-delà. [SUPPORT.md](SUPPORT.md) dit à quoi s'attendre.
5. **Honnête sur l'état du logiciel.** Le numéro de version, la page d'état de la documentation et les notes de release disent ce qui marche, ce qui casse et ce qui change. Cahut est en 0.x : tout peut changer avant la 1.0.
6. **La licence ne change pas, et il n'y a pas de version fermée.** Cahut est sous AGPL-3.0 ([ADR 15](docs/adr/0015-license.md)). Le même code sert tout le monde, auto-hébergé ou hébergé par quelqu'un d'autre. Il n'y a pas de contrat de licence de contribution : les contributeurs gardent leurs droits, si bien que personne, l'auteur compris, ne peut changer la licence de Cahut une fois des contributions extérieures entrées.
7. **Le nom est protégé, le code est libre.** « Cahut » et son logo identifient ce projet ([TRADEMARK.md](TRADEMARK.md)). Un fork prend un autre nom.
8. **La sécurité passe avant tout.** Un rapport de faille interrompt tout le reste ([SECURITY.md](SECURITY.md)).
9. **Rien ne quitte votre serveur en cachette.** La seule requête sortante du logiciel est la vérification des nouvelles releases sur GitHub, documentée et facultative. Il n'y a pas de télémétrie.
10. **Les IA sont bienvenues, les humains répondent.** Une contribution faite avec une IA est déclarée, comprise et testée par la personne qui la signe.
11. **Le socle ne se vote pas.** Les règles de produit du README (opérations bornées, knock first, une seule API, pas d'IA sur le serveur, tout est consigné, isolation) et les principes 6 à 9 ne changent qu'avec l'accord du mainteneur principal et de chaque mainteneur, par un ADR.

## Les rôles

| Rôle | Qui | Ce qu'il fait | Comment on y entre | Comment on en sort |
| --- | --- | --- | --- | --- |
| Utilisateur | Toute personne qui installe Cahut ou lit une Window | Lire, signaler, proposer, aider les autres | En arrivant | |
| Contributeur | Toute personne dont le travail a été retenu : code, documentation, traduction, aide régulière au tri | Tout ce qui précède ; cité dans les notes de release | Une pull request fusionnée, ou une aide suivie | |
| Trieur | Un contributeur régulier qui aide à classer les tickets | Étiqueter, fermer les doublons, demander des précisions, répondre dans les Discussions, avec le droit de tri de GitHub | Invité par le mainteneur principal après quelques mois d'aide | Six mois sans activité : retiré, avec des remerciements |
| Mainteneur | Une personne qui relit et fusionne | Relire, fusionner, publier les releases, une voix dans les décisions, avec le droit d'écriture ; listé dans MAINTAINERS.md dès qu'il y en a deux | Invité après six mois de contributions suivies et la confiance du mainteneur principal | Peut se retirer et devenir émérite ; un an sans activité : émérite |
| Mainteneur principal | Olivier Durand ([@odura](https://github.com/odura)) | Tout ce qui précède ; tranche ; détient le nom, le domaine et l'organisation GitHub | Fondateur | Prépare sa succession, ci-dessous |

Les rôles se proposent, ils ne se réclament pas. En demander un avec insistance est le plus sûr moyen de ne pas l'obtenir.

## Les décisions

- **Au quotidien**, le mainteneur décide dans les tickets et les pull requests, et dit pourquoi en une phrase.
- **Tout ce qui engage l'architecture, le produit ou le vocabulaire** passe par un ADR. N'importe qui peut en proposer un, en pull request, au statut *Proposed*. Il reste ouvert aux commentaires une à deux semaines s'il change quelque chose d'important. Puis le mainteneur principal l'accepte ou le refuse, avec les options écartées.
- **La feuille de route est publique et sans dates** : les jalons de ce dépôt, un par lot du plan de la V0, et la page d'état de la documentation. Les votes sur les idées renseignent ; ils ne décident pas.
- **Avec plusieurs mainteneurs**, les décisions ordinaires passent par consensus tacite : une proposition reste ouverte sept jours, et passe sans objection. Une objection dit pourquoi et propose une issue. Le mainteneur principal tranche. Chaque mainteneur déclare ses conflits d'intérêts, par exemple s'il travaille pour un hébergeur.
- **Le socle** (principe 11) ne change qu'avec l'accord de tous, par un ADR.

## La succession

Un projet d'une seule personne doit survivre à cette personne. Une personne de confiance, qui n'a pas besoin de savoir coder, est second propriétaire de l'organisation GitHub et détient un accès d'urgence au domaine et à la clé de signature des releases, avec des instructions écrites. Si le mainteneur principal reste injoignable six mois, ou s'il le demande, cette personne publie un avis dans le dépôt, nomme un mainteneur parmi les contributeurs connus, ou transmet le projet à une structure, selon ce que disent les instructions. La personne de confiance sera nommée ici une fois choisie.

La confiance se gagne ; elle ne s'accorde jamais sous la pression. La double authentification est obligatoire sur l'organisation, `main` est protégée, personne ne reçoit de droits pour avoir insisté, et chaque promotion suit une période d'observation. Dès qu'il y a deux mainteneurs, un changement qui touche la sécurité ou les releases est relu par deux personnes.

## Comment le projet grandit

| Phase | Commence quand | Ce qui change |
| --- | --- | --- |
| 1 · Le dépôt est ouvert | Le dépôt s'ouvre | Un créneau hebdomadaire pour le tri et les relectures, une release par lot, des nouvelles sur le site. Le mainteneur principal décide seul. Pas de chat, et pas de support au-delà des Discussions |
| 2 · Des contributeurs réguliers | Deux personnes avec trois pull requests fusionnées en trois mois, ou un tri qui déborde du créneau | Des trieurs, MAINTAINERS.md, la personne de confiance confirmée, un second mainteneur pour les relectures sensibles et pour le code de conduite ; peut-être un chat, avec deux personnes pour le tenir |
| 3 · Une équipe | Trois mainteneurs actifs depuis six mois, ou de l'argent au-delà des coûts, ou une offre hébergée | Le consensus tacite, une seconde version de cette page, une structure pour porter le nom et le domaine, une frontière nette entre le projet et toute activité commerciale |

Ce qui ne change pas : le socle, la licence, et le temps que chacun donne.

## L'argent

Les dons paient le domaine, un serveur de test, la marque et quelques outils. Ils ne paient pas de salaire, et nous le disons. Le projet les reçoit par GitHub Sponsors, et par Liberapay une fois son compte ouvert. Une offre hébergée ou du support payant, s'ils viennent, font tourner le même binaire que tout le monde, sans fonctionnalité réservée, et restent à l'écart du projet.

## Changer cette page

Cette page change par une pull request, et par un ADR pour tout ce qui dépasse la formulation. Le document de fond, avec les raisons et les sources, est [docs/gouvernance.md](docs/gouvernance.md).
