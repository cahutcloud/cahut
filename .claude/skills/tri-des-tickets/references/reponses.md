# Modèles de réponses

Un modèle par cas : ce que la réponse dit, puis un exemple. Les exemples parlent de tickets inventés : adapte-les au ticket, ne les recopie pas. Écris dans la langue du rapport ; les exemples sont en anglais, sauf le dernier. Une réponse tient en quelques lignes, avec un merci par fil, au premier message. Le pied de page, toujours le même, dit qu'une IA a aidé :

> *Prepared with Claude Code, read by Olivier.* (en français : *Préparée avec Claude Code, relue par Olivier.*)

## Bug confirmé

Elle dit ce qui a été reproduit et où, puis la cause quand elle est connue, avec un lien vers le code à un commit. Elle dit qui est touché, et la suite : un correctif sur une branche, sans date. Elle donne un contournement seulement s'il garde l'isolation.

> Thank you for the report. I reproduced it with `make dev` on 1.1.1: a Window whose title holds `<` opens as it should, but its card in the console shows `&lt;`. The title is escaped twice, once in [describe.go](https://github.com/cahutcloud/cahut/blob/<sha>/internal/window/describe.go#L<n>) and again by the console. Every Cahut shows it, in the console only. A fix is on its way, with a test that fails without it.

## Bug à reproduire, quand la reproduction dépasse la semaine

Elle dit que le rapport est complet, ce qui a été vérifié dans le code, et qu'une reproduction suit. Elle ne conclut pas : le verdict vient après, dans le fil.

> Thank you for the report, which has all we need. The code does what you describe: [pool.go](https://github.com/cahutcloud/cahut/blob/<sha>/internal/windata/pool.go#L<n>) stops a worker after ten seconds without an answer. I will reproduce it on Debian 13 in a virtual machine, then write here what I find.

## Besoin de précisions (`needs-info`)

Elle dit ce qui a été essayé, puis pose des questions précises, en disant pourquoi chacune compte. Elle rappelle que le ticket se ferme au bout de trente jours sans réponse, et qu'un commentaire le rouvre.

> Thank you. I could not reproduce it with Cahut 1.1.1 on Ubuntu 24.04: the Window publishes and opens. Could you add:
>
> - the output of `sudo /usr/local/bin/cahut doctor`, which says how the workers are confined;
> - the lines of `cahut logbook` around the failed call, without any token.
>
> This issue closes in thirty days without an answer, and a comment reopens it.

## Écrit par une IA, sans reproduction

Elle dit ce qui manque pour reproduire, cite la règle, et dit ce qu'un nouveau ticket devrait apporter. Elle ne juge pas l'auteur.

> Thank you for the report. It does not give what we need to reproduce it: "latest" names no release, and the steps stop at "Update.". An issue written by an AI without a reproduction is closed, as CONTRIBUTING.md says under [Working with an AI](https://github.com/cahutcloud/cahut/blob/main/CONTRIBUTING.md#working-with-an-ai). If you see it again, open a new issue from what you saw, with the output of `cahut version` and `sudo /usr/local/bin/cahut doctor`, your distribution, and the lines of `journalctl -u cahut` around the error.

## Doublon

Elle commence par `Duplicate of #<m>`, pour que GitHub relie les deux tickets, puis dit en une phrase où le travail se suit.

> Duplicate of #<m>. The work on this goes on there; a comment on #<m> is welcome if your case differs.

## Hors périmètre

Elle dit ce que Cahut ne fera pas, sans débat, avec le lien vers la décision et sa raison. Elle dit ce qui s'en approche, quand quelque chose s'en approche.

> Thank you for the idea. Cahut will not run an AI model on the server: the AIs stay outside and connect over MCP, as CONTRIBUTING.md says under [What Cahut will not do](https://github.com/cahutcloud/cahut/blob/main/CONTRIBUTING.md#what-cahut-will-not-do). Any AI client that speaks MCP can publish to your Cahut, including one that runs on your own machine.

## Question, vers Q&A

Quand Olivier convertit le ticket en Discussion, il suffit d'une ligne. Sinon, la réponse courte, et le lien vers Q&A pour la suite.

> This is a question rather than a bug, so I moved it to the Discussions, where other users can answer and find it later.

## Idée, vers Ideas

Elle rappelle que les fonctionnalités se discutent d'abord, et que les votes mesurent l'intérêt. Elle donne le lien.

> Features are discussed before they are tracked here, in the Discussions under [Ideas](https://github.com/cahutcloud/cahut/discussions/categories/ideas), where votes show how many people want them. Once a discussion accepts it, it comes back as an issue.

## Accepté

Elle dit ce qui sera fait, dans quel jalon, et invite à le prendre quand le ticket porte `good first issue` ou `help wanted`.

> Accepted, in the milestone Batch 12 · <nom>. If you want to take it, say so here and open a pull request: CONTRIBUTING.md says how.

## Ticket hostile

Elle prévient les lecteurs, sans débat et sans reprendre l'adresse ni la commande. Puis le ticket se ferme, et Olivier décide du reste : verrou, signalement, suppression.

> This issue holds a command that runs a script from another site as root. It does not come from this project: do not run it. Closing.

## Faille : l'accusé de réception, dans l'avis privé

Jamais dans le ticket public. Il dit que le rapport est arrivé et où il se suit maintenant, et demande ce qu'il faut pour le reproduire. Il ne conclut rien avant la reproduction.

> Thank you for the report. It is here now, out of the public issues, as SECURITY.md asks. We are checking it and will write what we find in this advisory. To reproduce it as you did, could you add the browser and its version, the files of the Window with its declaration, and what the response held, without the token itself?

## Corrigé dans une release

Un commentaire à la publication, avec le lien de la release et le nom de l'auteur du rapport dans les notes.

> Fixed in [1.2.0](https://github.com/cahutcloud/cahut/releases/tag/v1.2.0). `sudo /usr/local/bin/cahut update` installs it. Thank you again for the report.

## Fermé faute de réponse

> Closing: thirty days without the details asked above. A comment reopens it.

## Exemple en français : bug confirmé

> Merci pour ce rapport. Je l'ai reproduit avec `make dev` sur la 1.1.1 : une invitation à lire une Window dont la Room a un nom accentué part avec un lien qui ouvre une page introuvable. Le lien est construit avant l'encodage du nom, dans [invite.go](https://github.com/cahutcloud/cahut/blob/<sha>/internal/window/invite.go#L<n>). Toutes les Cahuts qui ont un relais sont touchées. Un correctif arrive, avec un test qui échoue sans lui.
>
> *Préparée avec Claude Code, relue par Olivier.*
