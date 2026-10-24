# Cahut · Plan du lot 9 : l'envoi d'e-mails

24 octobre 2026 · Olivier Durand · Validé

## En bref

Le lot 9 permet à chaque Cahut d'envoyer des e-mails, comme l'a décidé l'ADR 31. Les administrateurs du serveur règlent un relais SMTP pour tout le serveur, et ceux d'un Cahut peuvent régler le leur. Chaque personne peut avoir une adresse e-mail :

- son lien d'activation y arrive ;
- elle y reçoit, sans passer par un administrateur, un lien pour choisir un nouveau mot de passe.

Les membres d'une Room et leurs IA invitent des gens à lire une Window :

- dans le Cahut, par leur handle ;
- au-dehors, par leur adresse, avec un lien personnel qui tient lieu de Key.

Les Knocks, la sécurité des comptes et la santé du serveur envoient des notifications.

Le lot compte sept pull requests, pour environ onze jours et demi à temps plein. Sa décision, l'ADR 31, est déjà acceptée (#104).

**Fin de lot**, sur le staging :

- une personne qui a oublié son mot de passe en choisit un nouveau depuis la page de connexion, sans administrateur ;
- Claude invite une adresse extérieure à lire une Window à Key, qui s'ouvre d'un clic depuis l'e-mail ;
- un Knock arrive par e-mail à sa personne, qui l'approuve depuis le lien.

## Point de départ

Ce que la V0 et l'ADR 30 ont laissé :

- **Comptes.**
  - Une personne a un handle et un nom, sans adresse e-mail. Un administrateur du serveur aussi, dans `server.db`.
  - `create_user` et `create_activation_link` rendent un lien d'activation, `https://<domaine>/console/activate#<code>`, valable 7 jours. L'administrateur le transmet à la main.
  - La table `activations` garde un lien par personne, haché. `activate_account` pose le mot de passe et met fin aux autres sessions de la personne.
- **Lecteurs et Keys.**
  - Un lecteur se connecte aux Windows par la console, qui remet un code à usage unique à `/-/session`.
  - La bonne Key donne un laissez-passer, gardé dans `window_passes`. Une nouvelle Key efface les laissez-passer de l'ancienne.
- **Knocks.** Un Knock naît dans le paquet `knock`, attend 24 heures et ne se voit que dans la console.
- **Limites et santé du serveur.**
  - Le paquet `settings` compte les limites de l'ADR 30 (#103, #105, #106).
  - `cahut doctor` mesure l'espace libre du dossier des données.
  - À son démarrage, le service inscrit au Logbook du serveur les mises à jour que rapporte `status.json`.
- **Code réutilisable.**
  - La garde qui refuse les adresses non publiques vit dans `internal/oauth/clients.go`, pour les documents des clients OAuth.
  - `Call.AfterCommit` exécute une fonction une fois la transaction validée.
- **Ce qui manque.**
  - Le service n'a aucune tâche périodique.
  - Aucun outil MCP n'est marqué `openWorldHint`.

## Décision

L'[ADR 31](adr/0031-e-mail-through-a-relay.md), « Each Cahut sends e-mail through its own relay or that of its server », est fusionné (#104) et fixe ce que fait ce lot. Olivier y a tranché quatre points :

- **Deux relais.** Celui du serveur sert au serveur et aux Cahuts qui n'ont pas de relais à eux. Celui d'un Cahut passe en priorité pour ce Cahut, et doit être à une adresse publique.
- **La charte de Cahut** dans les e-mails, pour l'instant. La house style viendra avec une autre décision.
- **Une IA invite sans frapper**, dans des limites : 20 destinataires par appel, 100 invitations par jour et par personne.
- **L'anglais**, comme la console.

L'ADR fixe aussi le reste :

- le mot de passe d'un relais est chiffré sous une clé du serveur, et le réglage d'un relais n'est jamais proposé aux IA ;
- seuls la personne et les administrateurs de son Cahut lisent une adresse ;
- le lien de mot de passe oublié vaut une heure ;
- la Key n'est jamais envoyée par e-mail ;
- les messages partent après la validation de la transaction, depuis une file en mémoire, avec la bibliothèque standard de Go.

## Modèle de données

Chaque pull request apporte sa propre migration.

| Migration | Contenu |
| --- | --- |
| `cahut/0016_mail_relay`, `server/0009_mail_relay` | `mail_relay`, une seule ligne : hôte, port, mode TLS, utilisateur, mot de passe scellé, adresse et nom de l'expéditeur, date du réglage |
| `cahut/0017_addresses`, `server/0010_addresses` | Une colonne `email` dans `users` et dans `admins`, unique sans tenir compte de la casse. Une table `email_changes` : la personne, la nouvelle adresse, le code haché, l'expiration |
| `cahut/0018_invitations` | `invitations` : la Window ; qui invite, et l'IA quand c'en est une ; la personne invitée ou l'adresse ; le code haché, pour une Window à Key ; la création, l'expiration, la révocation. Une colonne `invitation_id` dans `window_passes` |
| `cahut/0019_notices`, `server/0011_notices` | `notices_sent` : le genre, la clé et la date d'envoi, pour qu'une même notification ne parte qu'une fois par jour. Une colonne `notices_off` dans `users`. Une colonne `access_set_by` dans `windows` |

Le mot de passe d'un relais est scellé en AES-256-GCM, sous une clé de 32 octets :

- la clé est tirée au premier usage dans `<données>/mail.key`, en mode 0600, à côté de `server.db` ;
- un Cahut copié sans ce fichier perd le mot de passe de son relais, qu'il faut retaper.

## Le paquet `mail`

`internal/mail` porte tout ce qui touche à SMTP, sans connaître les opérations.

- **Relais.**
  - Il lit et écrit le relais d'un Cahut ou du serveur, et scelle ou descelle son mot de passe.
  - Il choisit le relais de chaque message : celui du Cahut, sinon celui du serveur, sinon aucun.
- **Connexion.**
  - `net/smtp`, en TLS implicite ou en STARTTLS, avec un certificat vérifié. Authentification PLAIN quand le relais la propose.
  - La connexion au relais d'un Cahut passe par la garde des adresses publiques. Cette garde quitte `internal/oauth` pour `internal/address`, qui sert aux deux usages, et elle vérifie l'adresse IP à laquelle elle se connecte réellement.
  - Seul le relais du serveur, sur son adresse de bouclage, peut se passer de TLS.
- **Messages.**
  - Chaque message a un gabarit, avec une partie texte et une partie HTML (`multipart/alternative`, `quoted-printable`).
  - La charte de Cahut s'applique :
    - les couleurs et les polices viennent de `cahut-tokens.json`, recopiées dans un fichier Go qu'un test compare aux jetons ;
    - le logo horizontal, en PNG de la charte, voyage avec le message (`multipart/related`).
  - Les en-têtes portent `Message-ID`, `Date` et `Auto-Submitted: auto-generated`.
  - Quand le relais ne donne pas de nom d'expéditeur, le message prend le nom du Cahut, celui de sa house style.
- **File.**
  - `Call.AfterCommit` remplit une file en mémoire du serveur.
  - Elle envoie un message à la fois par relais, réessaie au bout d'une, cinq et dix minutes, puis abandonne.
  - Elle garde pour chaque relais la date du dernier envoi et la dernière erreur, que `get_mail_relay` rend.
- **Développement.** Avec `serve --dev` et sans relais, chaque message s'écrit en `.eml` dans `<données>/mail/`, et le journal donne ses liens. Les parcours s'essaient ainsi en local, sans serveur SMTP.
- **Tests.** `internal/mail/mailtest` est un relais SMTP en mémoire, avec STARTTLS sur un certificat de test. Il garde les messages reçus, que les tests lisent.

## Opérations

| Opération | Où | Qui l'appelle | Ce qu'elle fait |
| --- | --- | --- | --- |
| `set_mail_relay` | Cahut, serveur | Les administrateurs du Cahut, ou ceux du serveur ; jamais une IA | Se connecte au relais dans `Prepare`, puis l'enregistre. Un relais vide est retiré |
| `get_mail_relay` | Cahut, serveur | Les mêmes | Le relais sans son mot de passe, le dernier envoi et la dernière erreur. Pour un Cahut, dit s'il passe par le relais du serveur |
| `send_test_mail` | Cahut, serveur | Les mêmes | Envoie un message à l'adresse de l'appelant, et attend la réponse du relais |
| `create_user`, `create_admin` | Cahut, serveur | Inchangées, avec une adresse en plus | Envoient le lien d'activation à cette adresse |
| `create_activation_link` | Cahut, serveur | Inchangée | Envoie aussi le lien à l'adresse de la personne |
| `set_user_email` | Cahut | Les administrateurs ; frappe pour une IA | Donne une adresse à une personne, ou la lui retire |
| `change_email` | Cahut, serveur | La personne elle-même ; jamais une IA | Envoie à la nouvelle adresse un lien de confirmation, valable 24 heures. Une adresse vide retire la sienne |
| `confirm_email` | Cahut, serveur | Les visiteurs | Prend le code, remplace l'adresse et prévient l'ancienne |
| `request_password_link` | Cahut, serveur | Les visiteurs | Envoie un lien valable une heure à l'adresse du compte. Répond pareil dans tous les cas |
| `invite_to_window` | Cahut | Les membres de la Room et leurs IA, sans frapper | Envoie l'adresse de la Window, ou un lien personnel pour une Window à Key |
| `list_invitations` | Cahut | Les membres de la Room et leurs IA | Les invitations d'une Window |
| `revoke_invitation` | Cahut | Les membres de la Room et leurs IA | Met fin à une invitation et à ses laissez-passer. Défait `invite_to_window` |
| `get_invitation` | Cahut | Les visiteurs | Ce que montre la page d'une invitation : la Window, qui invite, la note |
| `open_invitation` | Cahut | Les visiteurs | Rend un code à usage unique, que la console remet à `/-/invitation`, qui pose le laissez-passer |
| `set_mail_notices` | Cahut | La personne elle-même ; jamais une IA | Coupe ou rétablit les notifications qui ne concernent ni son compte ni un Knock |
| `get_cahut` | Cahut | Inchangée | Dit en plus si le Cahut envoie des e-mails |

Précisions sur ces opérations :

- **Opérations réservées aux personnes** (`PeopleOnly`) : `set_mail_relay`, `get_mail_relay`, `send_test_mail`, `change_email`, `confirm_email`, `request_password_link`, `get_invitation`, `open_invitation` et `set_mail_notices`.
- **Lecture des adresses.** `get_account` et `list_users` donnent les adresses aux personnes qui ont le droit de les lire, et jamais à une IA. Approuvée, `create_user` rend l'adresse à la personne qui approuve, comme le lien.
- **Limites de `request_password_link`.** Elles se comptent par adresse IP, comme la connexion, et par compte : au plus un e-mail toutes les 5 minutes.
- **Règles d'`invite_to_window`.**
  - Par handle, seulement vers une personne qui ouvre déjà la Window.
  - Par adresse, seulement pour une Window à Key ou publique.
  - Dans un Cahut individuel, seulement par adresse, sans dire si cette adresse appartient à une personne du Cahut.
  - Le Logbook inscrit les destinataires, coupés comme tout champ ; `list_invitations` garde la liste entière.

## MCP

- `invite_to_window`, `list_invitations` et `revoke_invitation` deviennent des outils. `set_user_email` en devient un aussi pour les IA des administrateurs, et il frappe.
- `invite_to_window` est le premier outil marqué `openWorldHint`, parce qu'il écrit à des gens hors du Cahut. Un nouveau champ des opérations, `OpenWorld`, le déclare.
- Les consignes du serveur MCP gagnent une demi-phrase : `invite_to_window` envoie par e-mail l'adresse d'une Window, et un lien personnel pour une Window à Key.
- Dans `get_cahut`, une IA voit si le Cahut envoie des e-mails. Sans relais, elle donne elle-même l'adresse de la Window ou sa Key.

## Ligne de commande

- `cahut mail set`, `cahut mail show` et `cahut mail test` règlent, montrent et essaient un relais.
  - Sans `--domain`, la commande vise le relais du serveur ; avec `--domain`, celui d'un Cahut.
  - Le mot de passe se lit sur l'entrée standard (`--password-stdin`), jamais en argument, où `ps` le montrerait.
- `cahut user create` et `cahut admin create` prennent `--email`.
- `cahut user email --user <handle> <adresse>` donne une adresse à une personne, ou la lui retire.

## Les e-mails

| Message | Destinataire | Objet |
| --- | --- | --- |
| Activation | La nouvelle personne | Choose your password for {Cahut} |
| Mot de passe oublié | La personne | Choose a new password for {Cahut} |
| Confirmation d'adresse | La nouvelle adresse | Confirm your e-mail address |
| Adresse changée | L'ancienne adresse | Your e-mail address changed |
| Invitation | Le destinataire | {Nom} invites you to read {Window} |
| Knock | La personne de l'IA | {IA} is waiting for your answer |
| Mot de passe changé, jeton créé, IA connectée, rôle changé | La personne | Your password changed, A personal token was created, {Client} is connected to your account, You are now an administrator of {Cahut} |
| Ajout à une Room | La personne | You were added to {Room} |
| Fin d'un accès | Qui l'a réglé en dernier | {Window} closes to its readers on {date} |
| Limite | Les administrateurs du Cahut et du serveur, ou la personne dans un Cahut individuel | {Cahut} is at 90% of its {limit}, ou {Cahut} reached its {limit} |
| Mise à jour, certificat, disque | Les administrateurs du serveur | An update failed on {machine}, Cahut {version} is ready to install, The certificate of {domain} cannot be renewed, The disk of {machine} is almost full |
| Essai | L'appelant | A test from {Cahut} |

Ce que disent ces messages :

- **Invitation.** Le message nomme la personne qui invite, et l'IA quand c'est une IA qui l'a envoyé. La note s'affiche à part des mots de Cahut.
- **Knock.** Le message donne le titre de l'opération, ce qu'elle touche, l'expiration et le lien. Il ne contient jamais de texte écrit par l'IA.
- **Nom de la machine.** Les messages du serveur nomment la machine par son nom d'hôte.

## La tâche horaire

Le service lance une tâche chaque heure. Elle prévient :

- **la personne qui a réglé l'accès d'une Window** (`access_set_by`) quand cet accès finit dans les trois jours, une fois par fin d'accès ;
- **les administrateurs du Cahut et du serveur** quand une limite d'un Cahut atteint 90 % ou est atteinte, une fois par jour et par limite. Dans un Cahut individuel, la personne est prévenue pour ses propres limites ;
- **les administrateurs du serveur** quand le dossier des données est plein à 90 %, mesuré comme le fait `cahut doctor`, une fois par jour ;
- **les administrateurs du serveur** quand `status.json` annonce une version qui attend une personne, une fois par version.

Deux autres notifications viennent d'ailleurs :

- **Certificat.** CertMagic signale l'échec du renouvellement d'un certificat (événement `cert_failed`). Le service prévient une fois par jour et par domaine.
- **Mise à jour échouée.** Au démarrage, quand le service inscrit au Logbook une mise à jour échouée ou annulée, il prévient aussi.

`notices_sent` retient ce qui est parti, pour que rien ne reparte deux fois après un redémarrage.

## Console

| Écran | Contenu |
| --- | --- |
| Connexion | « Forgot your password? » : un handle ou une adresse, puis le même message dans tous les cas. Pareil pour les administrateurs du serveur |
| Activation, `/activate` | Inchangée. Dit « Choose a new password » quand le lien vient d'un oubli |
| Confirmation d'adresse, `/confirm-email` | Prend le code du fragment, confirme, et montre l'adresse retenue |
| Compte | L'adresse, son état et le moyen de la changer ; les notifications à couper |
| People | Pour les administrateurs : l'adresse de chaque personne, à donner à la création ou à changer. « Send by e-mail » à côté du lien d'activation |
| Réglages du Cahut | Le relais : celui du serveur ou celui du Cahut. Le formulaire, l'essai, le dernier envoi, la dernière erreur |
| Administration du serveur | Le relais du serveur, sous la même forme |
| Fiche d'une Window | « Invite » : des handles ou des adresses, et une note. La liste des invitations, avec « Revoke » |
| Invitation, `/invitation` | Sans connexion : qui invite, à quelle Window, avec quelle note. « Open » ouvre la Window |
| Premier démarrage | L'adresse de la première personne, facultative |

## Tests

- **Relais.**
  - Refus d'un relais de Cahut à une adresse privée : 127.0.0.1, 10.0.0.0/8, ::1, ou un nom qui résout vers une adresse privée au moment de la connexion. Le relais du serveur est accepté sur 127.0.0.1.
  - Refus d'un relais sans TLS.
  - Le mot de passe n'est jamais rendu ni inscrit au Logbook, et il reste illisible sans `mail.key`.
  - Le relais du Cahut passe en priorité. Sans lui, le message part par celui du serveur, et sans relais du tout, il ne part pas.
- **Envoi.**
  - Rien ne part avant la validation de la transaction. Un appel qui échoue, ou qui frappe, n'envoie rien.
  - Après un échec, la file réessaie puis abandonne, et garde la dernière erreur.
- **Adresses.**
  - Deux personnes ne peuvent pas avoir la même adresse, quelle que soit la casse.
  - `set_user_email` frappe quand une IA l'appelle.
  - `change_email` ne remplace l'adresse qu'après confirmation, et prévient l'ancienne.
  - Aucune IA ne lit d'adresse : ni par `get_account`, ni par `list_users`, ni dans un Knock, ni dans le résultat de `create_user`.
- **Mot de passe oublié.**
  - Même réponse pour un compte inconnu, sans adresse ou désactivé, et aucun envoi dans ces cas.
  - Limites par adresse IP et par compte, avec un e-mail au plus toutes les 5 minutes.
  - Le lien vaut une heure et remplace le précédent. L'ancien mot de passe fonctionne jusqu'à l'usage du lien, qui ferme les sessions.
  - Les administrateurs du serveur ont le même parcours.
- **Invitations.**
  - Par handle, seulement vers une personne qui ouvre déjà la Window. Par adresse, seulement pour une Window à Key ou publique ; refus pour l'accès `room` ou `internal`.
  - Dans un Cahut individuel, par adresse seulement, avec la même réponse que l'adresse soit celle d'une personne du Cahut ou non.
  - Le lien personnel donne un laissez-passer. Ouvrir la page sans cliquer n'accorde rien.
  - L'invitation prend fin à sa révocation, avec une nouvelle Key, un autre accès ou la fin de l'accès.
  - Au plus 20 destinataires par appel, et 100 invitations par jour, IA comprises.
  - Le Logbook inscrit les destinataires. Aucun message ne contient de Key.
- **Notifications.**
  - Chaque événement de l'ADR 31 envoie son message, une fois par jour au plus, même après un redémarrage.
  - `set_mail_notices` les coupe, sauf celles qui concernent le compte ou un Knock.
- **Cahut individuel.** Le test de l'ADR 30, qui parcourt un Cahut individuel de deux personnes, lit aussi les e-mails : aucun message à l'une ne nomme l'autre.
- **Gouvernance.** `revoke_invitation` entre dans `undo` pour `invite_to_window`. Les nouvelles opérations ont un titre.
- **Cloisonnement.** Le relais, les adresses et les invitations d'un Cahut n'existent pas dans un autre, et le relais d'un Cahut ne sert jamais à un autre.
- **Console.** Les parcours s'essaient dans le navigateur, avec les `.eml` du mode `--dev`.

## Découpage en pull requests

| # | Pull request | Contenu | Durée |
| --- | --- | --- | --- |
| 1 | Plan | Ce plan | 0,5 j |
| 2 | Relais | Le paquet `mail`, la garde déplacée dans `internal/address`, le scellement, la file, `mailtest`, le mode `--dev` ; les migrations `cahut/0016` et `server/0009` ; `set_mail_relay`, `get_mail_relay`, `send_test_mail` ; `cahut mail` | 2 j |
| 3 | Adresses et liens | Les migrations `cahut/0017` et `server/0010` ; les adresses dans `create_user` et `create_admin` ; `set_user_email`, `change_email`, `confirm_email`, `request_password_link` ; les e-mails d'activation ; `get_cahut` ; la ligne de commande | 2 j |
| 4 | Invitations | La migration `cahut/0018` ; `invite_to_window`, `list_invitations`, `revoke_invitation`, `get_invitation`, `open_invitation` ; `/-/invitation` ; les limites ; `OpenWorld` | 2 j |
| 5 | Notifications | Les migrations `cahut/0019` et `server/0011` ; les Knocks, la sécurité des comptes, les Rooms ; la tâche horaire, les certificats, les mises à jour ; `set_mail_notices` | 2 j |
| 6 | Console | Les écrans ci-dessus, les types TypeScript | 2 j |
| 7 | Documentation, staging et bilan | Une page sur l'envoi d'e-mails dans le guide d'administration ; les pages des concepts ; la référence ; les consignes MCP ; CLAUDE.md ; une relecture de sécurité des parcours de mot de passe et d'invitation ; le relais du staging ; la démonstration ; le bilan | 1 j |

Soit environ 11,5 jours. Les pull requests s'enchaînent dans l'ordre. Chacune régénère la référence de la documentation (`make docs-reference`), que la CI vérifie.

## Démonstration de fin de lot

**Prérequis** :

- la version du lot installée sur `staging.cahut.cloud` ;
- un relais SMTP pour le serveur du staging (voir les questions ouvertes) ;
- une adresse e-mail sur le compte d'Olivier ;
- Claude Code branché par OAuth.

1. Dans les réglages du Cahut, la console montre le relais du serveur. « Send a test » arrive dans la boîte d'Olivier.
2. Se déconnecter, puis cliquer sur « Forgot your password? » et donner le handle. L'e-mail arrive, et son lien permet de choisir un nouveau mot de passe. Les autres sessions se ferment, puis « Your password changed » arrive.
3. Demander à Claude de mettre une Window de `marketing` en accès Key et d'inviter une seconde adresse d'Olivier. L'e-mail nomme Olivier et Claude. En navigation privée, son lien ouvre la page d'invitation, et « Open » ouvre la Window.
4. Révoquer l'invitation dans la console. Le lien n'ouvre plus rien, et le laissez-passer non plus.
5. Demander à Claude de rendre la Window publique. L'e-mail du Knock arrive, et son lien permet de l'approuver.
6. Le Logbook montre l'invitation et son destinataire, la révocation et le Knock, sans aucun code.
7. Donner à un autre Cahut du staging un relais à 127.0.0.1 : la console le refuse et dit pourquoi.

## Questions ouvertes

- [ ] Quel relais pour le staging et la démonstration ? Une boîte du serveur Plesk de cahut.cloud, par exemple `no-reply@cahut.cloud` avec le SPF et le DKIM du domaine, ou un fournisseur d'e-mails ? Il faut un compte et ses identifiants avant la pull request 7.
- [ ] Si c'est Plesk : accepte-t-il qu'une autre machine, le staging, envoie par lui en SMTP authentifié, sur le port 587 ou 465 ?

## Hors du lot 9

- La house style dans les e-mails : une décision à venir, après l'ADR 31.
- Les e-mails dans d'autres langues, avec une console traduite.
- Les retours des messages non distribués, le désabonnement par lien, et les API HTTP des fournisseurs d'e-mails.
- Une limite d'e-mails par Cahut parmi celles de l'ADR 30, si les hébergeurs la demandent.
- Les lettres d'information et les résumés.
- Les notifications des sauvegardes : avec les sauvegardes.

## Bilan

Le code du lot 9 est fusionné le 24 octobre 2026, en huit pull requests. Il reste le relais du staging et la démonstration.

**Après le lot : l'ADR 34.** Le relais du serveur, le repli sur lui et l'e-mail des administrateurs du serveur sont retirés une fois le lot fusionné, sur décision d'Olivier. Seul un Cahut a un relais, réglé par ses administrateurs ; sans relais, il n'envoie rien et sa console ne montre aucune fonction d'e-mail : ni « Forgot your password? », ni adresse, ni invitation. Les administrateurs du serveur n'ont pas d'adresse, et le disque, les certificats et les mises à jour se lisent dans la console du serveur et dans `cahut doctor`. La migration `server/0013_no_mail` retire de `server.db` le relais, les adresses et les avis du serveur. Ce plan décrit le lot tel qu'il a été fait ; ce qui touche au relais du serveur ne vaut plus.

| # | Pull request | Contenu |
| --- | --- | --- |
| — | #104 | L'ADR 31 |
| 1 | #108 | Ce plan |
| 2 | #110 | Le paquet `mail` : les relais du serveur et de chaque Cahut, le scellement des mots de passe, la file, la mise en page à la charte de Cahut, `mailtest`, `cahut mail` |
| 3 | #113 | Les adresses des personnes et des administrateurs du serveur, leur confirmation, le mot de passe oublié, les liens d'activation par e-mail |
| 4 | #117 | Les invitations à lire une Window, le lien personnel et `/-/invitation` |
| 5 | #118 | Les notifications : Knocks, comptes, Rooms, la tâche horaire, les certificats, les mises à jour |
| 6 | #119 | Les écrans de la console |
| 7 | #120 | La documentation, CLAUDE.md, la relecture de sécurité, ce bilan |

**Écarts par rapport au plan.**

- **Les numéros des migrations** ont suivi le travail des autres sessions :
  - `server/0010_mail_relay`, `server/0011_addresses` et `server/0012_notices`, car la désactivation des administrateurs (#107) avait pris `server/0009` ;
  - `cahut/0019_invitations` et `cahut/0020_notices`, car la page d'accueil d'une Room (#115) avait pris `cahut/0018`.
- **Les opérations du relais** vivent dans le paquet `mail`, qui connaît donc les opérations, contrairement à ce que disait le plan.
- **`get_cahut` dit si le Cahut envoie des e-mails** dès la pull request 2, et non la 3.
- **Le registre fournit des services d'un Cahut à ses opérations** (`Registry.Provide`, `Call.Value`, `Registry.Value`), sans changer leur forme. Un Knock enregistré exécute `Knocked.Then`, qui envoie son e-mail.
- **`send_test_mail`** va à l'adresse de la personne qui l'appelle ; seule la ligne de commande nomme une adresse, comme le dit l'ADR 31.
- **Le type de sortie des opérations du relais** s'appelle `mail.Delivery` : `MailRelay` bégayait, et `Setup` entrait en collision avec la sortie de `get_setup` dans les types de la console.
- **La ligne de commande** prend `cahut user email --handle --email` au lieu de `--user <handle> <adresse>`.
- **En développement**, `Service.Sends` répond oui sans relais, pour que les parcours écrivent leurs messages sur le disque. Sans cela, les opérations, qui vérifient d'abord qu'un relais existe, n'envoyaient rien.
- **La mesure du disque** passe par `doctor.DiskUsage`, que les tests remplacent. Le test des notifications échouait sur une machine dont le disque était plein à 96 %.
- **La console** range les pages « E-mail » et « E-mail of the server », montre l'adresse d'une personne sous son nom, et ne propose de reprendre que les invitations qui portent un lien personnel.
- Pas de nouvelle dépendance.

**Relecture de sécurité** des parcours du mot de passe et des invitations.

- **Vérifié.**
  - `request_password_link` répond pareil pour un compte inconnu, sans adresse ou désactivé. Ses appels passent par le limiteur des visiteurs, par adresse IP, et un compte reçoit un lien toutes les 5 minutes au plus.
  - Le lien vaut une heure, sert une fois, porte son code dans le fragment, et n'est gardé que haché. L'utiliser ferme les autres sessions et prévient l'adresse.
  - Une IA ne change ni le relais, ni l'adresse de sa personne (`PeopleOnly`). Elle frappe pour l'adresse d'une autre personne.
  - Un administrateur qui change l'adresse d'une personne n'a pas plus de pouvoir qu'avant : il lui donnait déjà un lien d'activation.
  - Le relais d'un Cahut ne joint que des adresses publiques, vérifiées au moment de la connexion. Le mot de passe d'un relais ne part pas vers un nouvel hôte : il ne reste que si l'hôte et l'utilisateur ne changent pas, et son scellement est lié à eux.
  - Les en-têtes n'admettent ni saut de ligne ni seconde adresse. Le HTML échappe la note, et refuse les liens `javascript:`.
  - Ouvrir la page d'une invitation n'accorde rien. Le code qui donne le laissez-passer sert une fois, pendant une minute, et une invitation révoquée ou dépassée par une nouvelle Key n'en donne plus.
  - Le Logbook ne contient aucun code, ni l'adresse d'une personne. Il nomme seulement les destinataires d'une invitation, tapés par les membres de la Room.
- **Risques acceptés.**
  - Le nom d'un client OAuth, écrit par un tiers, figure dans l'e-mail d'un Knock comme dans la console, coupé à 60 caractères sur une ligne.
  - Un Cahut hébergé qui n'a pas de relais à lui envoie par celui du serveur sous son propre nom. Le pied de page dit de quel Cahut vient le message.
  - La dernière erreur du relais du serveur, montrée aux administrateurs d'un tel Cahut, nomme l'hôte de ce relais.
  - La note d'une invitation peut contenir une adresse web, que les clients de messagerie rendent cliquable. Le message la montre comme les mots de qui l'a écrite.

**Essais.**

- Les tests tournent contre un relais SMTP en mémoire et couvrent chaque opération, chaque refus et chaque message du plan, ainsi que la mise en page et la charte. Le test de gouvernance compte `invite_to_window` et `revoke_invitation`, et celui des outils MCP vérifie `openWorldHint`.
- Dans le navigateur, contre `serve --dev` qui écrit ses messages sur le disque :
  - le mot de passe oublié jusqu'à la page « Choose a new password » ;
  - le changement d'adresse et sa confirmation ;
  - les pages « E-mail » et « People » ;
  - l'invitation d'une personne et d'une adresse, puis la page du lien personnel, qui ouvre la Window dans son cadre.

**Reste à faire.**

- Choisir le relais du staging : une boîte du serveur Plesk de cahut.cloud ou un fournisseur, avec le SPF et le DKIM de l'expéditeur.
- Publier une release du lot, l'installer sur le staging, régler son relais, puis dérouler la démonstration.

**Suites.**

- La house style dans les e-mails, avec une décision à venir.
- Les retours des messages non distribués, si les relais des hébergeurs en ont besoin.
