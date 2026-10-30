# Politique de sécurité

Cahut est un serveur qui reçoit du HTML d'une IA et le sert à des lecteurs. Sa sécurité compte plus que n'importe quelle fonctionnalité. Merci de signaler ce que vous trouvez.

## Signaler une faille

Signalez en privé, jamais dans un ticket ou une discussion publics :

- par le signalement privé de GitHub : [signaler une faille](https://github.com/cahutcloud/cahut/security/advisories/new) ;
- ou par e-mail à `security@cahut.cloud`.

Dites ce que vous avez trouvé, comment le reproduire, et ce que cela permet selon vous. En anglais ou en français.

## À quoi s'attendre

- Un accusé de réception sous sept jours.
- Une correction et une publication coordonnée sous quatre-vingt-dix jours, bien plus tôt pour une faille grave. Vous connaissez la date avant que nous publiions.
- Un avis de sécurité GitHub avec un CVE, et une release de correctif qui le cite.
- Votre nom dans l'avis et dans les notes de release, sauf si vous préférez ne pas être nommé.

Il n'y a pas de programme de primes.

## Périmètre

Dans le périmètre : le binaire `cahut` et ses commandes, le point d'accès MCP, la console et son API, `install.sh`, la chaîne de release, et les sites cahut.cloud et docs.cahut.cloud. Tout moyen pour une Window de lire une autre Window, la console ou la session du lecteur est une faille : le bac à sable est la protection principale ([ADR 10](docs/adr/0010-windows-in-a-csp-sandbox.md)).

Hors périmètre : le contenu que les gens publient dans leurs Windows, les clients d'IA, et les services tiers.

## Versions prises en charge

Seule la dernière release reçoit des corrections. Cahut est en 0.x : mettez à jour dès qu'une release sort. Un serveur se met à jour chaque jour vers les versions correctives de sa version mineure, et `sudo /usr/local/bin/cahut update` installe toute release plus récente.

## Les rapports écrits par une IA

Reproduisez avant de signaler. Un rapport généré par une IA que personne n'a vérifié est fermé sans réponse. Un rapport inventé vaut un bannissement à son auteur.
