# Cahut · Plan du lot 11 : la console du serveur

31 octobre 2026 · Olivier Durand · Proposé

> Pour les agents : ce plan s'exécute tâche par tâche, avec `superpowers:subagent-driven-development` ou `superpowers:executing-plans`. Les étapes se cochent (`- [ ]`).

## En bref

Le lot 11 sépare les deux rôles d'administration, comme le décide l'[ADR 58](adr/0058-a-console-for-the-server.md). La console de la Cahut, sous `/console/`, sert ses personnes. La console du serveur, sous `/server/` de chaque domaine, sert les administrateurs du serveur. Depuis le domaine où ils se connectent, ils règlent toutes les Cahuts et modèrent chacune, sans rien voir de leurs personnes ni de leurs Rooms.

**Architecture.**
- L'app Preact de `internal/console/ui` reçoit une seconde entrée, `server.html`, avec sa coque et ses pages. Les deux entrées partagent les composants, les mots, le client de l'API et la charte, et un seul build les produit.
- Côté Go :
  - `/server/` sert cette entrée ;
  - `/server/api/<op>` reçoit les opérations du serveur ;
  - `/server/api/cahuts/<domaine>/<op>` exécute dans la Cahut nommée les opérations qui servent le serveur ;
  - le registre n'ouvre plus aux administrateurs du serveur que ces opérations-là.

**Outils.** Go et la bibliothèque standard, Preact et Vite tels qu'ils sont. Aucune dépendance nouvelle, ni en Go ni en npm.

**Spec.** [ADR 58](adr/0058-a-console-for-the-server.md). Ce plan s'appuie sur lui, et qui l'exécute lit les deux.

Le lot compte six pull requests. Chacune laisse `main` utilisable, et la release qui le clôt, 1.1.0, passe en staging puis en production **avant le gel du 15 novembre**.

**Fin de lot**, sur le staging :

- Sur `my.staging.cahut.cloud/server/`, l'administrateur du serveur se connecte une fois. Il règle le type et les limites d'`account.staging.cahut.cloud`, suspend cette Cahut puis la reprend, sans se reconnecter.
- Une Window d'`account.staging` est signalée. Il l'ouvre en aperçu depuis `my.staging…/server/`, la retire, puis la rétablit.
- Sur `my.staging.cahut.cloud/console/`, Olivier, en tant que personne, ne voit aucune trace du serveur. Avec la seule session du serveur dans le navigateur, `/console/` demande de se connecter.
- Avec curl, une session du serveur reçoit 401 à `/console/api/list_users` et 404 à `/server/api/cahuts/<domaine>/list_users`.
- Le premier démarrage d'une machine vierge (`make install-test`) imprime l'adresse de `/server/`, et dans le navigateur, il s'achève dans la console du serveur.

## Point de départ

- **Une seule app**, `base: "/console/"`. Ses pages du serveur sont des chemins de la console :
  - `/console/server/sign-in` et `/console/server/activate` ;
  - `/console/server/cahuts`, `/admins`, `/logbook` et `/updates`.

  `AdminActivationURL` imprime `/console/server/activate#` ([web.go:89](../internal/web/web.go), [cli/admin.go:517](../internal/cli/admin.go)).
- **Deux sessions** dans le `localStorage` de l'origine : `cahut.session` (la personne) et `cahut.server-session` (le serveur), dans [session.ts](../internal/console/ui/src/session.ts).
- **L'API.**
  - `POST /console/api/<op>` exécute une opération dans la Cahut du `Host`, avec une session de personne ou du serveur.
  - `POST /console/api/server/<op>` exécute une opération du registre du serveur.
  - `sessionActor` ([api.go](../internal/web/api.go)) vérifie une session du serveur au domaine de la requête.
- **Le registre** ([operation.go](../internal/operation/operation.go)).
  - `Admin` désigne à la fois la ligne de commande (`Handle == ""`, `CommandLine()`) et un administrateur du serveur.
  - `admit` met `c.Admin = true` pour tout `Admin`. Un administrateur du serveur atteint donc toute opération `CahutAdmin`.
  - Dans une Cahut, les opérations au niveau `ServerAdmin` sont `set_cahut_settings`, `set_legal_links`, `set_moderation_address`, `suspend_cahut`, `resume_cahut`, `create_service_token`, `list_service_tokens` et `revoke_service_token`.
  - Celles de la modération sont au niveau `CahutAdmin` ([moderation.go](../internal/moderation/moderation.go)).
  - Le registre du serveur ne contient que des opérations `ServerAdmin` ou `Visitors`.
- **Les aperçus des modérateurs.** Une frame `FrameByHost` (`'h'`) nomme la session du serveur. `moderatorOpens` l'accepte par `SessionLasts(ctx, id, c.Domain, now)` ([admins.go:527](../internal/server/admins.go)), qui exige que la session ait été ouverte au domaine de la Cahut. `preview_reported_window` donne un chemin relatif, `/-/frame/<capacité>/`.
- **La CSP de la console** ([console.go](../internal/console/console.go)) : `script-src <origine>/console/assets/` et `frame-src 'self'`.
- **Le premier démarrage** ([setup.tsx](../internal/console/ui/src/pages/setup.tsx)) a deux étapes, « You » et « Your first AI ». `set_up` renvoie `server_admin`, et la console ouvre alors la session du serveur par `sign_in`.
- **Les noms réservés** ([address.go:22](../internal/address/address.go)) : `mcp`, `api`, `admin`, `console`, `help` et `oauth`. `server` n'en fait pas partie.
- **Les tests.**
  - Côté Go, les fixtures de `internal/web` (`setup`, `call`, `activate`) servent `localhost` et `beta.localhost`.
  - La console n'a pas de tests d'interface : on la vérifie dans le navigateur, sur une Cahut de développement.

## Contraintes communes

- Pas de nouvelle dépendance Go ni npm, et `CGO_ENABLED=0`.
- Toute action passe par le registre. Les adaptateurs décodent, appellent et encodent ; le routage `/server/api/cahuts/<domaine>/` ne choisit que la Cahut et l'opération.
- Chaque écran garde ses mots en anglais et en français dans `words({ en, fr })`. Le build échoue si un mot français manque.
- Un refus qu'une personne rencontre dans la console a un code (`operation.NewRefusal`), dit dans les deux langues dans [refusals.ts](../internal/console/ui/src/refusals.ts). Les 401 restent des `account.ErrInvalidSession`.
- La barre de la console du serveur reprend les styles de la barre du serveur actuelle (`console.css`, `data-server`) : Ink, et Paper en thème sombre, avec les jetons de `@charte`. Jamais de valeur hexadécimale retapée.
- Vocabulaire : « Server » en anglais, « Serveur » en français. Cahut, Room, Window et Key prennent une capitale.
- `make console-types` après tout changement d'opération, `make docs-reference` après tout changement d'opération ou de commande, puis `make check` avant de pousser.
- Commits en anglais, signés (`git commit -s`), fusion par squash une fois la CI verte.
- Revérifier les numéros libres d'ADR avant chaque PR.
- En développement, le port 8080 est pris par Herd. On sert donc par `go run -tags console ./cmd/cahut serve --dev --data tmp/dev --dev-addr localhost:8090`, avec `localhost` et `beta.localhost` déclarées. Les liens imprimés gardent le port 8080 : corriger le port à la main.
- Aucune release ni installation en production à partir du 15 novembre.

## Points de vigilance

Ce sont des cas que l'ADR implique sans que son texte les teste. Chacun a son test dans la tâche qui possède le code.

1. **Un domaine écrit autrement que `list_domains`** dans `/server/api/cahuts/<domaine>/` : en majuscules ou avec un point final, il désigne la même Cahut. S'il est inconnu, vide, `..` ou encodé (`%2F`), la réponse est un 404 en JSON, jamais un 500. Tâche 3.
2. **Une Cahut suspendue**, celle du domaine de la requête ou celle que nomme le chemin : `/server/` et sa route `cahuts` répondent toujours, et `resume_cahut` passe. Tâches 3 et 5.
3. **Une session d'une autre sorte** :
   - une session de personne, un jeton d'IA ou un jeton de service sur `/server/api/*` reçoit 401 ;
   - une session du serveur sur `/console/api/*` reçoit 401 à partir de la PR 4 ;
   - une session du serveur présentée à un autre domaine que le sien reçoit 401.

   Tâches 3 et 11.
4. **Se déconnecter d'une console garde l'autre** : les deux vivent dans la même origine, sous deux clés. Tâches 7 et 12.
5. **L'aperçu d'une Window signalée d'une autre Cahut**, demandé avec une session ouverte ailleurs : la frame s'ouvre, puis se ferme quand la session prend fin ou que l'administrateur est désactivé. Tâches 3 et 9.

## Découpage en pull requests

| PR | Branche | Contenu | Tâches |
| --- | --- | --- | --- |
| 1 | `claude/server-console` | L'ADR 58 et ce plan | — |
| 2 | `claude/server-api` | Le serveur atteint chaque Cahut : champ `ForServer`, nom `server` réservé, routes `/server/api/…`, types. Additive | 1 à 4 |
| 3 | `claude/server-console-app` | La console du serveur, et le départ des pages du serveur de la console de la Cahut | 5 à 10 |
| 4 | `claude/one-console-each` | La bascule : la règle du registre, `/console/api` aux personnes seules, le nettoyage de la console de la Cahut | 11 et 12 |
| 5 | `claude/first-start-server` | Le premier démarrage s'achève dans `/server/`, et `install.sh` l'imprime | 13 |
| 6 | `claude/server-console-docs` | La doc, la feuille de route, la vérification, puis la release 1.1.0 | 14 et 15 |

**Calendrier.**

| Date | Étape |
| --- | --- |
| 31 octobre | PR 1 |
| 2 novembre | PR 2 |
| 5 novembre | PR 3 |
| 6 novembre | PR 4 |
| 7 novembre | PR 5 |
| 9 novembre | PR 6 |
| 10 novembre | Release 1.1.0 et staging |
| 12 novembre au plus tard | Production |

Il reste trois jours de marge avant le gel.

---

## PR 2 · Le serveur atteint chaque Cahut

### Tâche 1 : les opérations qui servent le serveur

**Fichiers.**
- Modifier [operation.go](../internal/operation/operation.go) : `Op`, `Info` et `info()`.
- Modifier [moderation.go](../internal/moderation/moderation.go) : ses six opérations.
- Modifier [cahut.go](../internal/cahut/cahut.go) : `get_cahut`.
- Tests : `internal/cahut/cahut_test.go` et `internal/server/server_test.go`.

**Interfaces.**
- Produit :
  - `Op.ForServer bool` et `Info.ForServer bool` ;
  - `func (i Info) ServesServer() bool`, vrai quand `i.Level == ServerAdmin || i.ForServer`.
- Le commentaire de `ForServer` dit : *the administrators of the server call it in each Cahut, from the console of the server, beside those at their level: its moderation, and get_cahut, which reads the settings they set. See docs/adr/0058-a-console-for-the-server.md.*

- [ ] **Écrire le test qui échoue.** `TestOperationsThatServeTheServer` liste, dans le registre d'une Cahut, les noms dont `ServesServer()` est vrai, triés :

```go
want := []string{"create_service_token", "dismiss_report", "get_cahut", "list_public_windows", "list_reports",
	"list_service_tokens", "preview_reported_window", "reinstate_window", "resume_cahut", "revoke_service_token",
	"set_cahut_settings", "set_legal_links", "set_moderation_address", "suspend_cahut", "withdraw_window"}
```

  À côté, `TestServerOperationsAreForItsAdministrators` vérifie que chaque opération du registre du serveur est `ServerAdmin` ou `Visitors`.
- [ ] **Le lancer.** `go test ./internal/cahut/ -run TestOperationsThatServeTheServer` échoue : `ServesServer` n'existe pas.
- [ ] **Implémenter.** Ajouter le champ, le recopier dans `info()`, puis mettre `ForServer: true` sur `list_reports`, `list_public_windows`, `preview_reported_window`, `dismiss_report`, `withdraw_window`, `reinstate_window` et `get_cahut`.
- [ ] **Relancer** les deux tests : ils passent. `go test -race ./internal/...` reste vert, puisque rien ne lit encore le champ.
- [ ] **Commiter** : *Mark the operations that serve the server in each Cahut.*

### Tâche 2 : `server` est un nom réservé

**Fichiers.**
- Modifier [address.go:22](../internal/address/address.go).
- Tests : `internal/address/address_test.go`, plus la table des handles refusés de `internal/account`.

- [ ] **Écrire les tests qui échouent.**
  - `Reserved("server")` est vrai.
  - Dans le test existant des handles réservés, `create_user` avec le handle `server` est refusé avec le code `handle_reserved`, comme `console`.
- [ ] **Les lancer** : ils échouent.
- [ ] **Implémenter** : ajouter `"server": true` à `reserved`.
- [ ] **Les relancer** : ils passent.
- [ ] **Commiter** : *Reserve the name server for the console of the server.*

### Tâche 3 : les routes `/server/api/`

**Fichiers.**
- Modifier [web.go](../internal/web/web.go) : `ServerPath`, le mux `serverConsole` et `ServeHTTP`.
- Modifier [api.go](../internal/web/api.go) : `serveCahutAsServer`.
- Modifier [admins.go](../internal/server/admins.go) : `SessionLasts`.
- Modifier [windows.go](../internal/web/windows.go) : `moderatorOpens`.
- Tests : `internal/web/api_test.go` et `internal/web/report_test.go`.

**Interfaces.**
- Consomme : `Info.ServesServer()` (tâche 1).
- Produit :
  - `const ServerPath = "/server/"` ;
  - `func ServerConsoleURL(domain string) string`, construite comme `ConsoleURL`, avec `ServerPath` ;
  - `func (s *Server) serveCahutAsServer(w http.ResponseWriter, r *http.Request)` ;
  - `func (s *Server) SessionLasts(ctx context.Context, id int64, now time.Time) (bool, error)`, sans domaine.

**Les routes**, sur un mux à part, `s.serverConsole` :
- `POST /server/api/cahuts/{domain}/{operation}` → `serveCahutAsServer` ;
- `POST /server/api/{operation}` → `serveServerAPI` ;
- `/server/api/` → `refuseAPI`.

`ServeHTTP` envoie au mux `serverConsole` tout chemin qui commence par `ServerPath`, **avant** `serveSuspended`, comme il le fait pour la console. `/console/api/server/` reste en place jusqu'à la PR 4.

**`serveCahutAsServer`** procède comme `callAPI`, avec `setAPIHeaders` et `readCall(w, r, MaxAPIRequestSize)`, dans cet ordre :
1. Il vérifie la session du serveur au domaine de la requête (`sessionActor(r, siteOf(r), true)`).
2. Il normalise le domaine du chemin (minuscules, sans point final).
3. Il ouvre la Cahut par `s.site`. Un `ErrNotFound` donne un 404 : `no Cahut at %q on this server: list_domains gives them`.
4. Il cherche l'opération dans le registre de cette Cahut. Une opération inconnue, ou dont `ServesServer()` est faux, donne un 404 : `no operation %q for the server in a Cahut`.
5. Il met `actor.Domain` au domaine de la Cahut, puis appelle `s.invoke`.

- [ ] **Écrire le test qui échoue**, `TestServerAPI`. `root` est activé par `POST /server/api/activate_account` à `localhost`.

```go
r := call(t, f, f.ts, http.MethodPost, "localhost", "/server/api/list_domains", root, ``)            // 200, 2 Cahuts
r = call(t, f, f.ts, http.MethodPost, "localhost", "/server/api/cahuts/beta.localhost/get_cahut", root, `{}`) // 200
r = call(t, f, f.ts, http.MethodPost, "localhost", "/server/api/cahuts/BETA.localhost./get_cahut", root, `{}`) // 200
r = call(t, f, f.ts, http.MethodPost, "localhost", "/server/api/cahuts/nope.localhost/get_cahut", root, `{}`)  // 404, JSON with "error"
r = call(t, f, f.ts, http.MethodPost, "localhost", "/server/api/cahuts/..%2Fetc/get_cahut", root, `{}`)      // 404
r = call(t, f, f.ts, http.MethodPost, "localhost", "/server/api/cahuts/beta.localhost/list_users", root, `{}`) // 404
r = call(t, f, f.ts, http.MethodPost, "localhost", "/server/api/cahuts/beta.localhost/suspend_cahut", root, `{"reason": "test"}`) // 200
r = call(t, f, f.ts, http.MethodPost, "localhost", "/server/api/cahuts/beta.localhost/get_cahut", root, `{}`)  // 200, with its suspension
r = call(t, f, f.ts, http.MethodPost, "localhost", "/server/api/cahuts/beta.localhost/resume_cahut", root, `{}`) // 200
```

  Le même test vérifie aussi :
  - une session de personne (`activate(…, "camille", false)`) et le jeton d'IA `f.token` reçoivent 401, sur `/server/api/list_domains` comme sur la route `cahuts` ;
  - `root` présenté à `beta.localhost` reçoit 401 ;
  - le Logbook de `beta.localhost`, lu par la ligne de commande, contient `suspend_cahut` sous le handle `root`.
- [ ] **Écrire le test qui échoue**, `TestHostPreviewsAReportOfAnotherCahut`, avec les aides de `report_test.go` :
  1. une Window est publiée et signalée à `beta.localhost` ;
  2. `root` se connecte à `localhost` et appelle `/server/api/cahuts/beta.localhost/preview_reported_window` ;
  3. la frame, demandée à l'hôte `beta.localhost`, répond 200 ;
  4. après `POST /server/api/sign_out`, la même frame répond comme une frame disparue (le statut de `gone()`).
- [ ] **Les lancer** : `go test ./internal/web/ -run 'TestServerAPI|TestHostPreviewsAReportOfAnotherCahut'` échoue.
- [ ] **Implémenter** les routes, `serveCahutAsServer` et `SessionLasts` sans domaine (la requête SQL perd `s.domain = ?`), et mettre `moderatorOpens` à jour.
- [ ] **Les relancer** : ils passent. Puis `make test`.
- [ ] **Commiter** : *Serve the API of the server under /server/api, for every Cahut.*

### Tâche 4 : les types et la référence

**Fichiers.**
- Modifier [apitypes/main.go](../internal/apitypes/main.go).
- Régénérer `internal/console/ui/src/api.gen.ts` et la référence des docs.
- Modifier [api.ts](../internal/console/ui/src/api.ts).

**Interfaces.**
- Produit, dans `api.gen.ts` :
  - `export type ServerCahutOperations = Pick<CahutOperations, …>`, avec les noms dont `ServesServer()` est vrai.
  - Les commentaires des tables nomment leurs routes : `/console/api/<operation>`, `/server/api/<operation>` et `/server/api/cahuts/<domain>/<operation>`.
- Produit, dans `api.ts` :
  - `cahut(op, input, token)` poste toujours à `/console/api/<op>` et ne dépend plus de `base` ;
  - `server(op, input, token)` poste à `/server/api/<op>` ;
  - `serverInCahut<K extends keyof ServerCahutOperations>(domain: string, op: K, input, token?)` poste à `/server/api/cahuts/${encodeURIComponent(domain)}/${op}`. La console du serveur l'enveloppe dans `cahutOf(domain)` (tâche 7) ;
  - `export type CahutCall = <K extends keyof ServerCahutOperations>(op: K, input: ServerCahutOperations[K]["input"]) => Promise<ServerCahutOperations[K]["output"]>`.

- [ ] **Implémenter** : le générateur écrit la troisième table, puis `make console-types` et `make docs-reference`.
- [ ] **Vérifier.** `make console` compile : la console actuelle appelle désormais `/server/api/` pour ses opérations du serveur, que la tâche 3 sert. Puis `make check`.
- [ ] **Vérifier dans le navigateur** (`make console`, puis le serveur de développement sur 8090) : la page Cahuts de la console actuelle se charge et `list_domains` répond par `/server/api/`.
- [ ] **Commiter** : *Type the operations that the server calls in a Cahut.*

---

## PR 3 · La console du serveur

### Tâche 5 : servir les pages de `/server/`

**Fichiers.**
- Modifier [console.go](../internal/console/console.go) et [console_test.go](../internal/console/console_test.go).
- Modifier [web.go](../internal/web/web.go) : mux `serverConsole`, `AdminActivationURL`.
- Modifier [cli/admin.go:517](../internal/cli/admin.go).
- Tests : `internal/web/web_test.go` et `internal/cli/admin_test.go`.

**Interfaces.**
- Produit :
  - `func ServerHandler(frames func(*http.Request) []string) http.Handler`, qui sert `server.html` pour toute adresse de page, avec la même CSP que la console, sauf `frame-src 'self' <origines>` ;
  - `func (s *Server) cahutOrigins(r *http.Request) []string` : pour chaque domaine de `s.srv.Domains`, `<schéma>://<domaine>[:<port>]`, avec le schéma et le port de `r` ;
  - `AdminActivationURL(domain)` devient `ServerConsoleURL(domain) + "activate#"`.
- `newHandler(files fs.FS)` devient `newHandler(files fs.FS, page string, frames func(*http.Request) []string)` : `Handler()` passe `"index.html", nil`.

- [ ] **Écrire les tests qui échouent.**
  - `TestServerHandler` (console) : `/`, `/cahuts/beta.localhost` et `/sign-in` servent `server.html`, et la CSP contient `frame-src 'self' http://beta.localhost;`.
  - `TestServerConsoleRoutes` (web) :
    - `GET /server` répond 301 vers `/server/` ;
    - `GET /server/cahuts` porte la politique de la console, et non le bac à sable ;
    - quand la Cahut `localhost` est suspendue, `GET /server/cahuts` porte toujours cette politique, et non l'avis d'une Cahut suspendue.
  - `admin_test.go` : `cahut admin link` imprime `/server/activate#`.
- [ ] **Les lancer** : ils échouent.
- [ ] **Implémenter** : `GET /server/` → `http.StripPrefix("/server", console.ServerHandler(s.cahutOrigins))`, la redirection `/server` → `/server/`, l'URL d'activation, et l'impression de la ligne de commande.
- [ ] **Les relancer** : ils passent. Puis `make test`.
- [ ] **Commiter** : *Serve the console of the server under /server/.*

### Tâche 6 : la seconde entrée de Vite

**Fichiers.**
- Créer `internal/console/ui/server.html` (une copie d'`index.html`, avec `/src/server.tsx`) et `internal/console/ui/src/server.tsx`.
- Modifier [vite.config.ts](../internal/console/ui/vite.config.ts) et [router.tsx](../internal/console/ui/src/router.tsx).

**Interfaces.**
- Produit :
  - `router.tsx` : `export const base`, qui vaut `"/server"` quand `location.pathname` est `/server` ou commence par `/server/`, et `"/console"` sinon ;
  - `vite.config.ts` : `build.rollupOptions.input = { console: "index.html", server: "server.html" }`, le proxy `"/server/api"` vers le serveur de développement, et un petit plugin qui, en développement, sert `/console/server.html` pour les adresses `/server/*` hors `/server/api/` ;
  - `src/server.tsx` importe la charte et `console.css`, puis rend `<ServerApp />` (tâche 7).

- [ ] **Implémenter.**
- [ ] **Vérifier.** `make console` produit `dist/index.html` et `dist/server.html`, qui chargent leurs scripts depuis `/console/assets/`. `make test` reste vert.
- [ ] **Commiter**, avec la tâche 7.

### Tâche 7 : l'app du serveur, sa session et sa coque

**Fichiers.**
- Créer, dans `internal/console/ui/src/server/` : `app.tsx`, `context.tsx`, `shell.tsx` et `account.tsx`.
- Modifier [admins.tsx](../internal/console/ui/src/pages/admins.tsx), [updates.tsx](../internal/console/ui/src/pages/updates.tsx) et [logbook.tsx](../internal/console/ui/src/pages/logbook.tsx).
- Déplacer `ServerOverview` d'[overview.tsx](../internal/console/ui/src/pages/overview.tsx) vers `server/cahuts.tsx`.
- Modifier [account.tsx](../internal/console/ui/src/pages/account.tsx) : exporter `PasswordForm`.

**Interfaces.**
- Produit, dans `server/context.tsx` :
  - `interface ServerConsole { session: Session; server: …; cahutOf(domain: string): CahutCall; renamed(a: { name: string }): void; signOut(): void }`.
  - `useServer()` et `ServerProvider`.
- Produit, dans `server/app.tsx` : `ServerApp` lit `load().admin` et suit ces routes :

| Adresse | Écran |
| --- | --- |
| `/sign-in` | `SignIn kind="admin"` |
| `/activate` | `Activate kind="admin"` |
| `/` | redirige vers `/cahuts` |
| `/cahuts` | la liste des Cahuts |
| `/cahuts/<domaine>` | la page d'une Cahut (tâche 9) |
| `/admins` | les administrateurs |
| `/logbook` | le Logbook du serveur |
| `/updates` | les mises à jour |
| `/account` | le compte |

  Sans session, toute autre adresse mène à `/sign-in`. Un 401 efface la session et y renvoie aussi.
- Produit, dans `server/shell.tsx` : `ServerShell`, avec la barre du haut en Ink (logo, « Server » ou « Serveur », le nom de l'administrateur et « Sign out »), et un menu Cahuts, Administrators, Logbook, Updates, avec Account en pied, plus les langues.
- `Logbook` prend `load: (limit: number) => Promise<…>` au lieu de `of`. La console de la Cahut lui passe `list_logbook` de la Cahut, et celle du serveur `list_logbook` du serveur.
- `Admins`, `Updates` et `Cahuts` passent de `useConsole()` à `useServer()`. Pour les liens vers une Cahut, ils utilisent `originOf(domain)` d'`overview.tsx`, à déplacer dans `components.tsx`.

- [ ] **Implémenter.** Se déconnecter du serveur appelle `sign_out` puis `clear("admin")` et ne touche pas à `cahut.session`.
- [ ] **Vérifier dans le navigateur**, sur 8090 :
  - `/server/` mène à la connexion ;
  - un administrateur créé par `go run ./cmd/cahut admin create --data tmp/dev …` s'active à `/server/activate#…` ;
  - Cahuts, Administrators, Logbook et Updates s'ouvrent ;
  - en anglais comme en français, la barre Ink dit « Server » ou « Serveur », en thème clair et en thème sombre ;
  - à 375 px de large, le menu s'ouvre par un bouton.
- [ ] **Vérifier le point de vigilance 4.** Avec une personne connectée dans `/console/` et l'administrateur dans `/server/`, se déconnecter de `/server/` laisse `/console/` ouverte, et inversement.
- [ ] **Commiter** : *Give the server a console of its own, under /server/.*

### Tâche 8 : les jetons de service d'une Cahut

**Fichiers.**
- Créer `internal/console/ui/src/server/tokens.tsx`.

**Interfaces.**
- Consomme : `cahutOf(domain)` avec `list_service_tokens`, `create_service_token` et `revoke_service_token`.
- Produit : `ServiceTokens({ call }: { call: CahutCall })`. La liste des jetons, la création avec un nom, le jeton montré une fois dans un `CopyField`, puis la révocation après un `ConfirmButton`.

- [ ] **Implémenter**, mots en anglais et en français.
- [ ] **Vérifier dans le navigateur** : créer un jeton, le voir une fois, le révoquer. `cahut service-token list --data tmp/dev --domain beta.localhost` suit.
- [ ] **Commiter**, avec la tâche 9.

### Tâche 9 : la page d'une Cahut

**Fichiers.**
- Créer `internal/console/ui/src/server/cahut.tsx`.
- Modifier [settings.tsx](../internal/console/ui/src/pages/settings.tsx) et [moderation.tsx](../internal/console/ui/src/pages/moderation.tsx).

**Interfaces.**
- `settings.tsx` exporte `SettingsForm`, `LegalForm`, `ModerationAddress`, `SuspensionForm` et `LimitsTable`. Chacun prend `call: CahutCall` à la place de `api.cahutAsServer`.
- `moderation.tsx` : `Moderation({ call, frameOrigin, host }: { call: CahutCall; frameOrigin: string; host: boolean })`.
  - Les cartes reçoivent `call`, `frameOrigin` et `host` par leurs props, et `useModerator` disparaît.
  - `frameOrigin` précède le chemin de la frame. Il est vide dans la console de la Cahut, et vaut `originOf(domain)` dans celle du serveur.
  - `host` dit que la console agit pour l'hébergeur, seul à rétablir ce qu'il a retiré.
- `CahutPage({ domain }: { domain: string })` lit `get_cahut`, puis montre :
  1. son adresse, en lien simple ;
  2. le type et les limites ;
  3. les pages légales ;
  4. l'adresse de modération ;
  5. la suspension ;
  6. la modération ;
  7. les jetons de service (tâche 8).

  Quand la Cahut n'a encore personne, un bouton donne un lien de premier démarrage (`create_setup_link`), affiché comme `<origine>/console/setup#<code>`.
- `server/cahuts.tsx` : chaque Cahut de la liste mène à `/cahuts/<domaine>`, avec le nombre de signalements en attente et sa suspension.

- [ ] **Implémenter.** La console de la Cahut passe `call={api.cahutAsServer}` ou `call={api.cahut}`, comme elle le fait aujourd'hui, jusqu'à la PR 4.
- [ ] **Vérifier dans le navigateur.** Depuis `http://localhost:8090/server/`, régler `beta.localhost` :
  - changer ses limites ;
  - écrire une page légale ;
  - suspendre la Cahut, puis la reprendre ;
  - signaler une Window de `beta.localhost` par `/-/report`, l'ouvrir en aperçu, la retirer, puis la rétablir.

  La console du navigateur ne montre aucune erreur de CSP, ce qui vérifie le point de vigilance 5.
- [ ] **Commiter** : *Set and moderate each Cahut from the console of the server.*

### Tâche 10 : les pages du serveur quittent la console de la Cahut

**Fichiers.**
- Modifier [app.tsx](../internal/console/ui/src/app.tsx) et [shell.tsx](../internal/console/ui/src/shell.tsx).

- [ ] **Implémenter.**
  - La barre latérale perd la section « This server » (`serverPages`).
  - `Page` perd les cas `/server/*`.
  - Une adresse `/console/server/<reste>`, celle d'un marque-page ou d'un ancien lien, part vers `/server/<reste>` par `location.replace`, fragment compris : un ancien lien d'activation garde ainsi son code.
  - La barre du serveur (`ServerBar`) reste jusqu'à la PR 4, et son lien « Manage the Cahuts of the server » mène à `/server/cahuts`.
- [ ] **Vérifier dans le navigateur** : `/console/server/admins` mène à `/server/admins`, et la console de la Cahut n'a plus de section du serveur.
- [ ] **Commiter** : *Move the pages of the server out of the console of the Cahut.*

---

## PR 4 · Chacun sa console

### Tâche 11 : le registre et l'API

**Fichiers.**
- Modifier [operation.go](../internal/operation/operation.go) (`authorize`) et [api.go](../internal/web/api.go) (`sessionActor`).
- Modifier [web.go](../internal/web/web.go) : retirer la route `/console/api/server/`.
- Tests : `internal/operation/operation_test.go` et `internal/web/api_test.go`. Corriger ensuite `legal_test.go`, `report_test.go`, `scenario_test.go` et `suspended_test.go`, ainsi que tout test qui appelle une opération des personnes ou des Rooms en administrateur du serveur.

**La règle**, dans `authorize`, après le `switch` :

```go
if a.Kind == Admin && !a.CommandLine() && !info.ServesServer() && !info.Visitors {
	return Forbiddenf("%s is for the people of the Cahut: the administrators of the server call in a Cahut what serves the server only", info.Name)
}
```

`sessionActor` refuse une session du serveur quand `!ofServer` : `the console of the Cahut takes the sessions of its people: the administrators of the server use /server/`, enveloppé dans `account.ErrInvalidSession`.

- [ ] **Écrire le test qui échoue**, `TestAServerAdministratorCallsWhatServesTheServer`. Il prend un registre de test avec une opération `AnyPerson`, une `CahutAdmin`, une `ServerAdmin` et une `CahutAdmin` avec `ForServer`.
  - `Actor{Kind: Admin, Handle: "root", Session: 1}` : les deux premières renvoient `ErrForbidden`, les deux autres passent.
  - La ligne de commande, `Actor{Kind: Admin}`, passe partout.
- [ ] **Écrire le test qui échoue** : `TestAPIForServerAdmins` est réécrit.
  - `root` reçoit 401 sur `/console/api/list_users`, avec `/server/` dans le message ;
  - `/console/api/server/list_domains` répond 404 ;
  - `/server/api/cahuts/localhost/list_users` répond 404.
- [ ] **Les lancer** : ils échouent.
- [ ] **Implémenter**, puis `make test`, et faire passer chaque test qui casse par l'une des deux voies :
  - la route `/server/api/cahuts/<domaine>/` pour ce qui sert le serveur ;
  - la ligne de commande ou une personne administratrice pour le reste.
- [ ] **Relancer** : `make check` est vert, puis `make docs-reference`.
- [ ] **Commiter** : *Keep the administrators of the server to what serves the server in a Cahut.*

### Tâche 12 : la console de la Cahut, pour ses personnes

**Fichiers.**
- Modifier [app.tsx](../internal/console/ui/src/app.tsx), [context.tsx](../internal/console/ui/src/context.tsx), [shell.tsx](../internal/console/ui/src/shell.tsx) et [session.ts](../internal/console/ui/src/session.ts).
- Modifier, dans `pages/` : [account.tsx](../internal/console/ui/src/pages/account.tsx), [settings.tsx](../internal/console/ui/src/pages/settings.tsx), [moderation.tsx](../internal/console/ui/src/pages/moderation.tsx), [overview.tsx](../internal/console/ui/src/pages/overview.tsx) et [sign-in.tsx](../internal/console/ui/src/pages/sign-in.tsx).
- Modifier `console.css` : les styles de la barre du serveur passent sous la coque du serveur.

**Interfaces.**
- `Console` perd `serverSession`, `cahutAsServer`, `server`, `leaveServer` et `serverRenamed`.
- `App` n'agit qu'avec `held.person`. Le mode « administrateur du serveur seul » disparaît, et sans personne la console mène à `/sign-in`.
- `session.ts` garde `load()`, et donc la migration d'avant l'ADR 43. La console de la Cahut ne lit plus que `person`.

- [ ] **Implémenter.**
  - Retirer de la coque `ServerBar`, `actingAsServer` et `signInAsPerson`, ainsi que les branches `!person`.
  - « Kind and limits » ne montre plus que la lecture : `LimitsTable`, `LegalTable` et l'adresse de modération.
  - Le bandeau d'une Cahut suspendue dit, en anglais : *The administrators of the server resume it.* En français : *Les administrateurs du serveur la réactivent.*
  - La modération appelle `api.cahut` avec `host={false}`. Une Window retirée par le serveur s'y montre comme retirée par le serveur, sans bouton pour la rétablir.
  - La page du compte perd la section du serveur et l'ajout de l'autre compte.
  - Retirer les mots devenus inutiles, dans les deux langues.
- [ ] **Vérifier dans le navigateur, avec les deux sessions dans le stockage** :
  - `/console/` ne montre aucune barre ni aucune page du serveur ;
  - se déconnecter de `/console/` laisse `/server/` connectée ;
  - avec la seule session du serveur, `/console/` mène à la connexion ;
  - une personne qui administre `localhost` voit « Kind and limits » en lecture seule.
- [ ] **Vérifier** : `make console` et `make check`.
- [ ] **Commiter** : *Keep the console of the Cahut for its people.*

---

## PR 5 · Le premier démarrage

### Tâche 13 : finir dans la console du serveur

**Fichiers.**
- Modifier [setup.tsx](../internal/console/ui/src/pages/setup.tsx) et `src/server/cahut.tsx`.
- Modifier [install.sh](../install.sh) (vers les lignes 420 à 430) et [test/install/run.sh](../test/install/run.sh).

**Interfaces.**
- `Setup` n'a plus qu'une étape, « You » ou « Vous ». `FirstAI` quitte le premier démarrage ; la vue d'ensemble garde son invitation à connecter une IA.
- À la fin de « You » :
  - si `server_admin` est vrai, il garde les deux sessions (`save`), puis `location.assign("/server/cahuts/" + info.domain + "?ready")` ;
  - sinon, il garde la personne et va à `/`.
- `CahutPage` montre, quand l'adresse porte `?ready`, un encadré :
  - en anglais : *Your Cahut is ready*, avec le lien *Open its console* vers `<origine>/console/` ;
  - en français : *Votre Cahut est prête*, avec *Ouvrir sa console*.
- `install.sh` imprime `The console of the server: https://$domain/server/`, après le lien du premier démarrage comme dans la branche « has people already ».

- [ ] **Écrire le test qui échoue** : `test/install/run.sh` vérifie que la sortie contient `/server/`, comme il vérifie `/console/setup#`.
- [ ] **Implémenter.**
- [ ] **Vérifier dans le navigateur**, avec un dossier de données neuf (`tmp/dev-first`) et `go run ./cmd/cahut setup --data tmp/dev-first --domain localhost` :
  - le premier démarrage crée les deux comptes et s'achève dans `/server/cahuts/localhost?ready` ;
  - le lien de l'encadré ouvre `/console/`, déjà connecté.

  Une seconde Cahut du même serveur s'achève dans `/console/`.
- [ ] **Vérifier** : `make install-test SYSTEM=ubuntu-24.04`.
- [ ] **Commiter** : *End the first start in the console of the server.*

---

## PR 6 · La doc et la release

### Tâche 14 : la doc

**Fichiers.**
- docs.cahut.cloud : `admin/console.md`, `install/first-start.md`, `install/server.md`, `concepts/cahuts.md`, `admin/commands.md` et `admin/mail.md`, plus la référence régénérée.
- [CLAUDE.md](../CLAUDE.md) : le paragraphe « Windows open at… », l'arborescence de `internal/console` et la phrase des ADR.
- La feuille de route de cahut.cloud et la page d'état des docs, comme à chaque fin de lot (chercher `currentStep`), en anglais et en français (`en.ts`, `fr.ts`).

- [ ] **Écrire la doc** : deux consoles, qui fait quoi, l'adresse de `/server/`, et ce qu'un hébergeur règle pour chaque Cahut.
- [ ] **Vérifier** : `make sites` et `make check`.
- [ ] **Commiter** : *Document the console of the server.*

### Tâche 15 : la vérification et la release

- [ ] **Avant la release**, lister les Rooms de chaque Cahut de staging (`my.staging`, `account.staging`), de production (`my`, `account`) et de `demo`, en administrateur de chaque Cahut (« Every Room ») ou par `list_rooms` sur MCP. Aucune ne doit s'appeler `server`. S'il y en a une, s'arrêter et en parler avec Olivier avant de continuer.
- [ ] **Préparer la release 1.1.0** : le badge du README, puis le tag `v1.1.0`. Olivier relit le brouillon et le publie.
- [ ] **Staging** : `sudo /usr/local/bin/cahut update`, puis rejouer la « Fin de lot » du début de ce plan.
- [ ] **Production** avant le 13 novembre, avec la même vérification, sans rien changer à ses Cahuts.

## Hors du lot 11

- Un Logbook de chaque Cahut en lecture pour le serveur, et toute vue de ses personnes ou de ses Rooms.
- Un domaine propre à la console du serveur.
- Un accès des administrateurs du serveur par MCP, et l'e-mail pour eux.
- Un contrôle de `cahut doctor` pour les Rooms nommées comme une page de Cahut.
