# Cahut · Plan du lot 12 : les skills de la Cahut

15 novembre 2026 · Olivier Durand · Proposé

> Pour les agents : ce plan s'exécute tâche par tâche, avec `superpowers:subagent-driven-development` ou `superpowers:executing-plans`. Les étapes se cochent (`- [ ]`).

## En bref

Le lot 12 donne à chaque Cahut un référentiel de skills que ses membres partagent, quelles que soient leurs IA, comme le décide l'ADR 64.

- Les administrateurs écrivent les skills de la Cahut dans la console, et chaque personne écrit les siens.
- Chaque IA branchée sur la Cahut voit le catalogue de ses skills dans la description de `find_skills`, et lit le bon avec `get_skill`. Personne n'installe ni n'importe rien, et une modification arrive à la conversation suivante.
- Une porte dans `publish_window` et `update_window` rend certaine la lecture des skills marqués « pour chaque Window ».

**Architecture.**
- Un paquet nouveau, `internal/skill` : le format d'un skill (nom, en-tête, limites, empreinte), son zip, son stockage, ses opérations, le catalogue et la porte.
- `internal/window` : le champ `follows`, qui passe la porte, et `from_skill`, qui copie un fichier d'un skill dans une Window.
- `internal/web` : la clé des serveurs MCP prend la révision du catalogue, `mcpServer` ajoute le catalogue à la description de `find_skills`, les instructions reçoivent un pointeur, et l'API de la console lit un appel plus gros pour `import_skill`.
- La console : un écran « Skills ».

**Outils.** Go et la bibliothèque standard, avec `go.yaml.in/yaml/v3` pour lire l'en-tête d'un `SKILL.md`. Ce module est déjà dans le graphe du binaire, par zap : il devient une dépendance directe, et la PR 2 dit pourquoi. Preact et Vite tels qu'ils sont, sans paquet npm nouveau.

**Spec.** [ADR 64](adr/0064-the-skills-of-a-cahut.md), que la PR 1 écrit. Qui exécute ce plan la lit d'abord.

Le lot compte sept pull requests, dont la cinquième seulement si l'essai d'une liste de 48 Ko passe. Chacune laisse `main` utilisable. La release qui le clôt, 1.4.0, passe en préproduction, puis en production après la recette.

**Fin de lot**, sur une Cahut collaborative de la préproduction :

1. Un administrateur importe `relance-facture`. La demande « Rédige l'e-mail de relance de la facture F-2026-0417 du 2 octobre, de 1 840 €, que la Menuiserie Bertin n'a pas encore réglée », faite à claude.ai, Claude Code, ChatGPT, Mistral Vibe et Codex sans nommer ni la Cahut ni le skill, donne un e-mail qui suit le skill.
2. L'administrateur change une phrase du skill dans la console. La conversation suivante de chaque IA suit la nouvelle version.
3. La charte d'une organisation, marquée « pour chaque Window » : une IA qui demande un deck sans l'avoir lue est arrêtée par la porte, puis publie un deck qui suit la charte, polices et logo copiés avec `from_skill`.
4. Un skill exporté s'installe tel quel dans Claude Code et dans Mistral Vibe, et un skill de Claude s'importe tel quel.
5. Un skill personnel se lit quand la personne le nomme, et `find_skills` le trouve.

---

## Point de départ

Sur origin/main du 15 novembre 2026 (`117b55a`), release 1.3.2 en préproduction et en production :

- **La liste des outils** d'un administrateur qui peut écrire, dans une Cahut collaborative, pèse 37 336 octets pour 41 outils : 3 624 octets de marge sous 40 Ko ([ADR 57](adr/0057-the-list-of-tools-under-40-kb.md)). Celle d'un membre pesait 33 112 octets pour 35 outils sur `f5b6e32`. Le test est `TestMCPToolsFollowTheRole`, dans [web_test.go](../internal/web/web_test.go), `len(list) > 40<<10`.
- **Les instructions** du serveur MCP font 3 396 caractères, et leur premier paragraphe 436. `TestInstructionsStartWithWhatToPublish` vérifie ses 512 premiers caractères.
- **`mcpHandler`** ([mcp.go](../internal/web/mcp.go)) garde un serveur MCP par jeu d'outils, `map[[3]bool]*mcp.Server`, et lit le type de la Cahut à chaque requête avec `settings.IsIndividual`.
- **Un fichier d'une Window** est `window.File{Path, Content, Encoding}` ([window.go](../internal/window/window.go)), `content` obligatoire.
- **L'API de la console** lit un appel de 1 Mio au plus (`MaxAPIRequestSize`), 16 Kio pour les opérations `Visitors` ([api.go](../internal/web/api.go)).
- **La dernière migration** d'une Cahut est `0038_listings.sql`. **La dernière ADR** est la 63.
- **Le skill `cahut`** est embarqué par [plugins.go](../plugins/plugins.go) et servi à `/-/cahut-skill.zip`. Le plugin est en version 0.10.0.
- **Les essais du 14 et du 15 novembre**, avec Mistral Vibe 2.1.0 et Claude Code 2.1.322, que donne l'ADR 64 :
  - pour une tâche qui ne publie rien, le catalogue dans la description de `find_skills` donne 50 sur 50 chez Vibe et 30 sur 30 chez Claude Code, autant que des skills installés en natif ;
  - sans lui, Vibe n'y pense jamais, puisqu'il ne passe pas les instructions du serveur à son modèle ;
  - une porte dans `publish_window` a arrêté Vibe 10 fois sur 10, puis chaque deck a suivi la charte.

## Contraintes communes

- Les conventions de [CLAUDE.md](../CLAUDE.md) : tout passe par le registre ; chaque opération a un `Title` ; les erreurs pour l'appelant enveloppent `ErrInvalid`, `ErrNotFound` ou `ErrForbidden` avec un message qui dit quoi corriger ; un refus que les personnes rencontrent a un code (ADR 37) ; le Logbook ne garde jamais le contenu d'un fichier ni le texte d'un skill.
- Le code, ses commentaires et les textes du produit en anglais ; `docs/` en français, sauf `docs/adr/`.
- Après chaque changement d'une opération : `make console-types` et `make docs-reference`.
- **Chaque PR qui touche la liste des outils** donne son poids, mesuré par le test, dans sa description.
- `make check` est vert avant chaque push. Chaque commit est signé (`git commit -s`).
- Les numéros (ADR 64 et 65, migration `0039`) se revérifient sur origin/main avant chaque PR.

## Points de vigilance

1. **La place dans la liste des outils.** Sous 40 Ko, le catalogue n'a qu'environ 1 200 octets, soit 6 skills. La PR 4 réserve 700 octets à `follows` et `from_skill` (PR 6). L'essai de la tâche 17 décide si la borne monte à 48 Ko.
2. **Les 512 premiers caractères des instructions** ne changent pas : le pointeur va au début du second paragraphe.
3. **Les serveurs MCP des révisions passées** sont oubliés, sans quoi chaque écriture en laisserait un en mémoire.
4. **La porte ne vise que les IA.** Une personne qui publie n'est jamais arrêtée.
5. **Une Cahut individuelle** n'a pas de skills de la Cahut. Elle n'a donc ni catalogue ni porte.
6. **Les skills personnels des autres** restent fermés aux administrateurs, dans l'API comme dans le Logbook.
7. **Le zip** se lit avec ses bornes, à mesure : une bombe zip ne doit jamais tenir en mémoire.
8. **L'en-tête rendu** met `name` et `description` entre guillemets doubles : Mistral Vibe ignore sans rien dire un skill dont la description contient « : » sans guillemets.
9. **ChatGPT garde une copie figée de la liste des outils.** La recette dit s'il faut rafraîchir le connecteur pour voir un skill nouveau, et la doc le dit.

## Découpage en pull requests

| PR | Branche | Contenu | Tâches |
| --- | --- | --- | --- |
| 1 | `claude/skills-adr` | L'ADR 64 et ce plan | 1 |
| 2 | `claude/skills-store` | Le référentiel : le format, le zip, la migration, les opérations des personnes, la borne d'`import_skill`, les types. Rien ne change pour les IA. | 2 à 7 |
| 3 | `claude/skills-console` | L'écran « Skills » de la console | 8 à 10 |
| 4 | `claude/skills-for-ais` | `find_skills` et `get_skill` offerts aux IA, le catalogue, la révision, le pointeur, « What AIs see », puis la mesure en local | 11 à 16 |
| — | — | L'essai d'une liste de 48 Ko, sur la préproduction | 17 |
| 5 | `claude/tools-under-48-kb` | L'ADR 65, la borne et le budget du catalogue. Seulement si l'essai passe. | 18 |
| 6 | `claude/skills-gate` | La marque « pour chaque Window », `follows` et la porte, `from_skill`, puis la mesure en local | 19 à 23 |
| 7 | `claude/skills-docs` | Le skill `cahut` du plugin, la doc, CLAUDE.md, la feuille de route, la release et la recette | 24 à 27 |

**Durées estimées** : une demi-journée pour la PR 1, deux jours pour la PR 2, un jour et demi pour la PR 3, deux jours pour la PR 4, une demi-journée pour l'essai et la PR 5, deux jours pour la PR 6, un jour pour la PR 7, puis la recette.

---

## PR 1 · L'ADR et le plan

### Tâche 1 : l'ADR 64

**Fichiers.**
- Créer `docs/adr/0064-the-skills-of-a-cahut.md`.
- Modifier [docs/adr/README.md](adr/README.md) : sa table, avec les lignes des ADR 30 et 49 que l'ADR 64 étend, et leur statut.
- Modifier [CLAUDE.md](../CLAUDE.md) : la phrase qui résume les ADR, et ce plan dans la liste à lire et dans la carte du dépôt.
- Créer `docs/plan-lot-12.md` : ce plan.

**Ce que dit l'ADR**, dans la forme des autres (Status, Date, Context, Decision, Consequences, Alternatives considered) :
- **Context.** Les skills vivent dans chaque IA, et une équipe aux IA différentes ne garde pas de référentiel commun. L'ADR 49 a promis la marque d'une organisation par ses skills. Ce qui arrive au modèle de chaque client : la liste des outils partout ; les instructions coupées, ou absentes chez Mistral Vibe. Les chiffres des trois essais.
- **Decision.**
  - Les portées `cahut`, `personal` et `builtin`. Pas de skills de la Cahut dans une Cahut individuelle.
  - Le format des Agent Skills, deux versions, une empreinte par version, pas de `scripts/`.
  - Le catalogue des skills de la Cahut dans la description de `find_skills`, borné. Un pointeur dans les instructions.
  - La porte : `follows`, avec l'empreinte, pour les skills marqués « pour chaque Window ».
  - `from_skill`.
  - Les IA lisent les skills et n'en écrivent pas : les opérations qui écrivent sont `PeopleOnly`, et la raison est écrite.
  - La règle de priorité. Le Logbook sans le texte des recherches.
- **Consequences.** Le poids de la liste ; l'ADR 49 et l'ADR 30 étendues (le disque compte les skills) ; ce qui reste aléatoire et ce qui est certain.
- **Alternatives considered.** Celles de la partie 21 de la spec.

- [ ] Vérifier sur origin/main que 64, 65 et `0039` sont libres.
- [ ] Écrire l'ADR et le plan, ajouter la ligne de la table.
- [ ] **Commiter** : *Decide the skills of a Cahut.* Ouvrir la PR. Olivier la relit avant toute autre PR du lot.

---

## PR 2 · Le référentiel

### Tâche 2 : le format d'un skill

**Fichiers.**
- Créer `internal/skill/format.go` et `internal/skill/format_test.go`.
- Modifier `go.mod` et `go.sum` : `go.yaml.in/yaml/v3` devient direct.

**Interfaces.**

```go
// Package skill keeps the skills of a Cahut: instructions that its people
// write for some tasks, in the format of Agent Skills, which AIs read over
// MCP. See docs/adr/0064-the-skills-of-a-cahut.md.
package skill

const (
	MaxText            = 64 << 10 // a SKILL.md
	MaxEveryWindowText = 16 << 10 // the SKILL.md of a skill for every Window, which a refusal of the gate carries
	MaxFile            = 2 << 20
	MaxSize            = 4 << 20 // all the files of a version, SKILL.md included
	MaxFiles           = 100
	MaxPerScope        = 50
	MaxEveryWindow     = 3
	BuiltinName        = "cahut"
)

type Scope string

const (
	ScopeCahut    Scope = "cahut"
	ScopePersonal Scope = "personal"
	ScopeBuiltin  Scope = "builtin"
)

// File is a file of a skill, other than its SKILL.md.
type File struct {
	Path    string
	Type    string // its media type, from its extension
	Content []byte
}

// Header is what Cahut reads of the front matter of a SKILL.md.
type Header struct{ Name, Description string }

func CheckName(name string) error                 // 1 to 64 of a-z, 0-9 and -, no - at either end, no --
func ParseText(text string) (Header, error)         // the front matter of a SKILL.md, read with go.yaml.in/yaml/v3
func RenderText(text string) (string, error)        // the same SKILL.md, with name and description in double quotes
func CheckPath(p string) error                      // relative, clean, not SKILL.md, not under scripts/
func Digest(text string, files []File) string       // 8 hex digits of the SHA-256 of the text and the files, by path
```

- [ ] **Écrire les tests qui échouent.**
  - `TestCheckName` : une table, dont `charte-acme`, `a`, 64 caractères (admis), et 65 caractères, `-a`, `a-`, `a--b`, `A`, `é` (refusés).
  - `TestParseText` : une description qui contient « : » sans guillemets, comme Claude l'accepte ; `|` et `>` ; `metadata` sur un niveau ; sans en-tête, sans `name`, sans `description`, une description de 1 025 caractères (refusés).
  - `TestRenderTextQuotes` : le texte rendu se relit avec `yaml.v3` en mode strict, `name` et `description` sont entre guillemets doubles, les autres clés et le corps sont inchangés octet pour octet.
  - `TestCheckPath` : `references/deck.md` et `assets/fonts/inter.woff2` admis ; `../x`, `/x`, `a//b`, `SKILL.md`, `scripts/run.sh` refusés.
  - `TestDigest` : le même pour les mêmes fichiers dans un autre ordre ; un autre si un octet change.
- [ ] **Les lancer** : `go test ./internal/skill/` échoue, le paquet n'existe pas.
- [ ] **Implémenter.** `go get go.yaml.in/yaml/v3@v3.0.4`, puis `go mod tidy`.
- [ ] **Relancer** : les tests passent.
- [ ] **Commiter** : *Read and check the format of a skill.*

### Tâche 3 : le zip d'un skill

**Fichiers.** Créer `internal/skill/zip.go` et `internal/skill/zip_test.go`.

**Interfaces.**

```go
// ReadZip reads a skill from a zip, with its SKILL.md at the root or in a
// single folder, as AI clients export them. It checks every path and every
// limit as it reads, and never holds more than MaxSize.
func ReadZip(data []byte) (text string, files []File, err error)

// WriteZip writes a skill as a zip, in a folder of its name, its SKILL.md
// rendered by RenderText.
func WriteZip(name, text string, files []File) ([]byte, error)
```

- [ ] **Écrire les tests qui échouent** : `TestReadZipAtRoot`, `TestReadZipInAFolder`, `TestReadZipRefuses` (deux dossiers, `..`, un chemin absolu, un lien symbolique, un doublon, `scripts/`, sans `SKILL.md`, 101 fichiers, un fichier de 2 Mio et un octet, une bombe zip de 64 Mio de zéros), `TestWriteZipReadsBack` (les fichiers reviennent octet pour octet, l'en-tête rendu).
- [ ] **Les lancer**, puis **implémenter** avec `archive/zip` et `io.LimitReader` : chaque fichier est borné à `MaxFile`, le tout à `MaxSize`, à mesure de la lecture.
- [ ] **Relancer** : les tests passent.
- [ ] **Commiter** : *Read and write the zip of a skill.*

### Tâche 4 : la migration

**Fichiers.** Créer `internal/store/migrations/cahut/0039_skills.sql`.

```sql
-- The skills of a Cahut and of its people. See
-- docs/adr/0064-the-skills-of-a-cahut.md.
CREATE TABLE skills (
  id INTEGER PRIMARY KEY,
  scope TEXT NOT NULL CHECK (scope IN ('cahut', 'personal')),
  owner_id INTEGER REFERENCES users (id) ON DELETE CASCADE, -- NULL for a skill of the Cahut
  name TEXT NOT NULL,
  every_window INTEGER NOT NULL DEFAULT 0,
  current_version INTEGER,
  previous_version INTEGER,
  created_at TEXT NOT NULL,
  CHECK ((scope = 'cahut') = (owner_id IS NULL)),
  CHECK (every_window = 0 OR scope = 'cahut')
);
CREATE UNIQUE INDEX skills_of_the_cahut ON skills (name) WHERE owner_id IS NULL;
CREATE UNIQUE INDEX skills_of_a_person ON skills (owner_id, name) WHERE owner_id IS NOT NULL;

CREATE TABLE skill_versions (
  id INTEGER PRIMARY KEY,
  skill_id INTEGER NOT NULL REFERENCES skills (id) ON DELETE CASCADE,
  number INTEGER NOT NULL,
  digest TEXT NOT NULL,
  text TEXT NOT NULL,          -- the SKILL.md as received
  description TEXT NOT NULL,
  size INTEGER NOT NULL,       -- the SKILL.md and the files
  written_at TEXT NOT NULL,
  written_by INTEGER REFERENCES users (id) ON DELETE SET NULL,
  via TEXT NOT NULL CHECK (via IN ('console', 'import'))
);

CREATE TABLE skill_files (
  version_id INTEGER NOT NULL REFERENCES skill_versions (id) ON DELETE CASCADE,
  path TEXT NOT NULL,
  type TEXT NOT NULL,
  content BLOB NOT NULL,
  PRIMARY KEY (version_id, path)
);

-- Words match whatever their case and accents, as for Windows (0034).
CREATE VIRTUAL TABLE skill_search USING fts5(name, description, text, tokenize = 'unicode61 remove_diacritics 2');
CREATE TRIGGER skills_leave_search AFTER DELETE ON skills BEGIN
  DELETE FROM skill_search WHERE rowid = old.id;
END;

-- The revision of the catalog that AIs see, which every change to a skill
-- of the Cahut raises in its transaction.
CREATE TABLE skill_catalog (id INTEGER PRIMARY KEY CHECK (id = 1), revision INTEGER NOT NULL);
INSERT INTO skill_catalog (id, revision) VALUES (1, 0);
```

- [ ] La table des personnes est `users` (`0001_init.sql`). Vérifier que la suppression d'une personne (ADR 47 et 59) emporte ses skills personnels, et ne laisse aucun skill de la Cahut sans auteur lisible.
- [ ] `go test ./internal/store/` passe : les migrations s'appliquent sur une base vierge et sur une copie de `tmp/dev`.
- [ ] **Commiter** : *Keep the skills of a Cahut in its database.*

### Tâche 5 : les opérations des personnes

**Fichiers.**
- Créer `internal/skill/skill.go` (le service et `Register`), `internal/skill/store.go` (les lectures et écritures), `internal/skill/refusals.go`, et leurs tests.
- Modifier [cahut.go](../internal/cahut/cahut.go) : `skill.Register(r, …)` dans `Open`.
- Modifier [plugins.go](../plugins/plugins.go) : `func SkillFS() fs.FS`, qui rend les fichiers du skill intégré.

**Interfaces.** Sept opérations, toutes `PeopleOnly` dans cette PR. La PR 4 offre `find_skills` et `get_skill` aux IA.

| Opération | Titre | Entrée | Sortie |
| --- | --- | --- | --- |
| `find_skills` | Find skills | `{query?}` | `{skills: [{skill, scope, description, version, number, every_window}]}` |
| `get_skill` | Read a skill | `{skill, file?}` | `{skill, scope, version, number, written_by, written_at, via, every_window, text, files: [{path, size, type}], note}`, ou avec `file` : `{skill, file, content, encoding}` |
| `write_skill` | Write a skill | `{scope, skill, text?, files?: [{path, content, encoding}], remove?: [path]}` | le skill, comme `get_skill` sans `text` |
| `import_skill` | Import a skill | `{scope, zip}` (base64) | le skill |
| `export_skill` | Export a skill | `{skill}` | `{name, zip}` (base64) |
| `rollback_skill` | Roll back a skill | `{skill}` | le skill |
| `delete_skill` | Delete a skill | `{skill}` | `{skill, deleted: true}` |

- `skill` se donne seul, ou avec la portée en préfixe (`cahut:`, `personal:`). Un nom porté par deux portées sans préfixe est refusé, et le message dit d'ajouter le préfixe.
- `get_skill` rend le texte par `RenderText`, et `note` dit en une ou deux phrases la règle de priorité et `from_skill`. Le skill intégré se lit depuis `plugins.SkillFS()`.
- `write_skill` garde les fichiers qu'il ne nomme pas. Un skill nouveau demande `text`, dont le `name` est celui du skill.
- `import_skill` lit le zip dans son `Prepare`, avec `ReadZip`, et garde le résultat avec `Call.Keep`. Son `Run` écrit comme `write_skill`.
- Les droits : la portée `cahut` demande un administrateur et une Cahut collaborative ; la portée `personal` vaut pour la personne seule ; un administrateur ne lit ni n'écrit les skills personnels d'un autre.
- Chaque écriture, retour en arrière ou suppression d'un skill de la Cahut augmente `skill_catalog.revision` dans sa transaction.
- Les refus avec un code : `skill_name_invalid`, `skill_name_reserved`, `skill_header_invalid`, `skill_too_large`, `skill_has_scripts`, `skill_zip_invalid`, `skill_too_many`, `skill_scope_individual`.
- Le Logbook : la portée, le nom, la version, le nombre et le poids des fichiers. Pour `find_skills`, le nombre de résultats seulement, jamais `query`.

- [ ] **Écrire les tests qui échouent**, dans `internal/skill/skill_test.go`, avec une Cahut de test comme les autres paquets :
  - `TestWriteAndGetASkill`, `TestWriteKeepsTheFilesItDoesNotName`, `TestRemoveAFile`, `TestRollBackASkill`, `TestDeleteASkill` ;
  - `TestOnlyAdministratorsWriteTheSkillsOfTheCahut`, `TestNoSkillsOfTheCahutInAnIndividualCahut`, `TestAdministratorsDoNotReadThePersonalSkillsOfOthers` ;
  - `TestTheSameNameInTwoScopes`, `TestAnAmbiguousNameIsRefused`, `TestTheBuiltinSkill` ;
  - `TestFindSkillsWhateverTheCaseAndAccents`, `TestTheLogbookKeepsNoQuery` ;
  - `TestTheRevisionFollowsTheSkillsOfTheCahut` (une écriture personnelle ne la change pas) ;
  - `TestImportAndExport` (le zip exporté se réimporte à l'identique), `TestTheLimits`.
- [ ] **Les lancer** : ils échouent.
- [ ] **Implémenter** le stockage, puis les opérations, une à une, jusqu'à ce que les tests passent.
- [ ] Vérifier que `TestEveryOperationHasATitle` passe, et que la liste des outils du test de [web_test.go](../internal/web/web_test.go) n'a pas changé.
- [ ] **Commiter** : *Let people write, read, import and export skills.*

### Tâche 6 : un appel plus gros pour `import_skill`

**Fichiers.** Modifier [operation.go](../internal/operation/operation.go) (`Op.LargeRequest` et `Info.LargeRequest`) et [api.go](../internal/web/api.go) (`MaxLargeAPIRequestSize = 6 << 20`).

- [ ] **Écrire le test qui échoue** : `TestImportSkillTakesALargeCall`, dans `internal/web` : un zip de 3 Mio passe par `/console/api/import_skill`, un appel de 2 Mio à une autre opération reçoit 413.
- [ ] **Implémenter** : `LargeRequest: true` sur `import_skill`, et la borne dans `invoke`, à côté de `Visitors`.
- [ ] **Commiter** : *Let the console import a skill of 4 MiB.*

### Tâche 7 : les types, la référence et les refus

- [ ] `make console-types`, puis les phrases des huit codes dans [refusals.ts](../internal/console/ui/src/refusals.ts), en anglais et en français.
- [ ] `make docs-reference`.
- [ ] `make check` est vert.
- [ ] **Commiter** : *Write the types and the reference of the operations of skills.*
- [ ] **Ouvrir la PR.** Sa description dit pourquoi `go.yaml.in/yaml/v3` devient direct : l'en-tête des Agent Skills est du YAML, un lecteur maison refuserait des en-têtes valides venus d'ailleurs, et le module est déjà dans le graphe par zap.

---

## PR 3 · L'écran « Skills »

### Tâche 8 : la liste

**Fichiers.**
- Créer `internal/console/ui/src/pages/skills.tsx`.
- Modifier [router.tsx](../internal/console/ui/src/router.tsx) et [shell.tsx](../internal/console/ui/src/shell.tsx) : la route `/console/skills` et son entrée dans la navigation.

- [ ] Deux parties, « Skills of this Cahut » et « My skills », d'après `find_skills`. Une Cahut individuelle n'a que la seconde. Le skill intégré paraît en lecture seule, avec le lien de `/-/cahut-skill.zip`.
- [ ] Chaque ligne : le nom, la description, la version et sa date.
- [ ] Les boutons « Write a skill » et « Import a zip » : pour chacun dans « My skills », pour les administrateurs dans « Skills of this Cahut ».
- [ ] L'encadré de la ligne à coller, en recours : « If your AI does not think of the skills of this Cahut, paste this line in its instructions », puis la ligne, avec un bouton pour la copier.
- [ ] Les mots dans `const t = words({ en: { … }, fr: { … } })`.
- [ ] **Commiter** : *List the skills of a Cahut in its console.*

### Tâche 9 : la page d'un skill

**Fichiers.** Créer `internal/console/ui/src/pages/skill.tsx`.

- [ ] Le nom, la portée, la description ; la version, sa date, qui l'a écrite et par où.
- [ ] Le texte du `SKILL.md`, en police à chasse fixe, sans mise en forme.
- [ ] Les fichiers, chacun téléchargeable par `get_skill` avec `file`.
- [ ] « Roll back to the previous version », « Export » (le zip d'`export_skill`), « Delete », avec une confirmation qui dit que les deux versions partent pour de bon.
- [ ] **Commiter** : *Show a skill in the console.*

### Tâche 10 : écrire et importer

**Fichiers.** Créer `internal/console/ui/src/pages/skill-write.tsx`.

- [ ] Écrire : le texte du `SKILL.md` dans un champ, les fichiers à ajouter, remplacer ou retirer. Un skill nouveau part d'un modèle d'en-tête.
- [ ] Sous le champ, l'aide sur la description : « AIs see the first 160 characters. Say what the skill does and when to follow it. »
- [ ] Importer : un zip, lu en base64 dans le navigateur, envoyé à `import_skill`.
- [ ] Chaque refus dit sa phrase de `refusals.ts`.
- [ ] Essayer chaque page avec `make dev` et `make console-dev`, en anglais et en français, en clair et en sombre, puis à 375 px de large. Les captures vont dans la PR.
- [ ] **Commiter** : *Write and import skills in the console.* Ouvrir la PR.

---

## PR 4 · Les IA trouvent les skills

### Tâche 11 : le catalogue

**Fichiers.** Créer `internal/skill/catalog.go` et `internal/skill/catalog_test.go`.

**Interfaces.**

```go
// CatalogBytes is the budget of the lines of the catalog, in bytes of
// JSON, within the bound of the list of tools (ADR 57).
const CatalogBytes = 1200

// Catalog returns the lines that the description of find_skills carries:
// a line for each skill of the Cahut, its name then its description cut at
// 160 characters on a word, those for every Window first, then by name,
// within budget, and a last line for those that do not fit. It returns ""
// for an individual Cahut. Revision is the revision of the catalog.
func Catalog(ctx context.Context, q Querier, budget int) (lines string, revision int64, err error)

// Revision returns the revision of the catalog of a Cahut.
func Revision(ctx context.Context, q Querier) (int64, error)
```

- [ ] **Écrire les tests qui échouent** : `TestCatalogOrder`, `TestCatalogCutsDescriptionsOnAWord`, `TestCatalogWithinItsBudget` (la ligne « and N more skills: find_skills lists them »), `TestNoCatalogInAnIndividualCahut`, `TestCatalogLeavesOutPersonalSkills`.
- [ ] **Implémenter**, puis **relancer**.
- [ ] **Commiter** : *Write the catalog of the skills of a Cahut.*

### Tâche 12 : `find_skills` et `get_skill` offerts aux IA

**Fichiers.** Modifier `internal/skill/skill.go` et [web_test.go](../internal/web/web_test.go).

- [ ] Retirer `PeopleOnly` de `find_skills` et de `get_skill`, et les marquer `ReadOnly`.
- [ ] Leurs descriptions, courtes. Pour `find_skills` : « The skills of this Cahut: how the user's team wants some tasks done, whatever the task, such as an e-mail, a report or a deck. Before a task, read with get_skill the skill below whose description fits it, and follow it. With query, find the others and the user's own skills. »
- [ ] Ajouter les deux noms à `want` dans le test de la liste des outils.
- [ ] **Écrire le test qui échoue** : `TestToolListWithAFullCatalog`. Une Cahut collaborative avec 50 skills de la Cahut aux descriptions de 1 024 caractères : la liste d'un administrateur qui peut écrire pèse moins de `40<<10 - 700` octets, les 700 réservés à la PR 6.
- [ ] **Commiter** : *Offer find_skills and get_skill to AIs.*

### Tâche 13 : un serveur MCP par révision du catalogue

**Fichiers.** Modifier [mcp.go](../internal/web/mcp.go) et ses tests.

- [ ] La clé de `servers` devient `struct{ individual, admin, readOnly bool; revision int64 }`. `server(r)` lit la révision avec `skill.Revision`, comme il lit `settings.IsIndividual`. Quand une révision nouvelle arrive, les serveurs des autres révisions sont oubliés.
- [ ] `mcpServer` ajoute les lignes de `skill.Catalog(…, skill.CatalogBytes)` à la description de `find_skills`.
- [ ] **Écrire les tests qui échouent** :
  - `TestTheCatalogFollowsTheSkills` : une requête sans session voit un skill de la Cahut juste écrit, et ne voit plus un skill supprimé ;
  - `TestAPersonalSkillKeepsTheServer` : une écriture personnelle ne refait pas de serveur ;
  - `TestOldServersAreForgotten` : après dix écritures, la Cahut ne garde que les serveurs de la dernière révision.
- [ ] **Implémenter**, puis **relancer**.
- [ ] **Commiter** : *Show AIs the catalog of the skills of a Cahut.*

### Tâche 14 : le pointeur

**Fichiers.** Modifier [mcp.go](../internal/web/mcp.go) (`instructions`) et [web_test.go](../internal/web/web_test.go).

- [ ] **Écrire le test qui échoue** : `TestInstructionsPointToSkills` : « find_skills » dans les 2 048 premiers caractères, que garde claude.ai.
- [ ] Mettre au début du second paragraphe : « Cahut also keeps the skills of the user's team, for any task, even one that publishes nothing here: before a task, call find_skills and follow the skill that fits it. »
- [ ] `TestInstructionsStartWithWhatToPublish` passe sans changement.
- [ ] **Commiter** : *Point AIs to the skills of a Cahut.*

### Tâche 15 : « What AIs see »

**Fichiers.** Modifier `internal/skill/skill.go` (une opération `preview_skill_catalog`, `PeopleOnly`, pour les administrateurs) et `internal/console/ui/src/pages/skills.tsx`.

- [ ] L'opération rend les lignes du catalogue, les octets qu'elles prennent, le budget, et les noms des skills qui n'y tiennent pas.
- [ ] La console les montre aux administrateurs, sous « Skills of this Cahut ».
- [ ] `make console-types`, `make docs-reference`.
- [ ] **Commiter** : *Show administrators what AIs see of their skills.*

### Tâche 16 : la mesure en local

Avant de fusionner, avec le binaire de la branche, dans une Cahut de dev collaborative :

- [ ] Poser `relance-facture` (objet « [Lumen · Relance 1] », délai de neuf jours ouvrés, référence « LUM- » suivie du numéro, phrase « Si votre paiement a croisé ce message… », signature « Le pôle Adhérents et facturation ») et trois leurres : une charte, un modèle de rapport, des mentions de formulaire.
- [ ] Demander dix fois à Mistral Vibe et dix fois à Claude Code, chacun dans un dossier vierge, avec une personne neuve : « Rédige l'e-mail de relance de la facture F-2026-0417 du 2 octobre, de 1 840 €, que la Menuiserie Bertin n'a pas encore réglée. »
- [ ] **Le critère** : 10 sur 10 pour chacun, le skill lu avant d'écrire (d'après le Logbook) et l'e-mail avec les cinq marques. Dix fois une traduction, en contrôle, sans aucun skill lu.
- [ ] Les chiffres et le poids de la liste vont dans la description de la PR. **Ouvrir la PR.**

---

## L'essai d'une liste de 48 Ko

### Tâche 17 : essayer la liste à 48 Ko

L'ADR 57 demande d'essayer la liste avec Claude, ChatGPT, Codex et Mistral Vibe avant de monter la borne. Sans cela, le catalogue reste à environ 6 skills.

- [ ] Construire un binaire d'essai depuis `main`, la PR 4 fusionnée, avec `CatalogBytes = 9000` : avec 50 skills de la Cahut, la liste d'un administrateur pèse alors 48 Ko environ. Ce binaire ne va jamais dans une release.
- [ ] L'installer sur la préproduction avec `sudo /usr/local/bin/cahut update --binary`, dans une Cahut collaborative (vérifier son type avec `cahut settings show`), avec 50 skills dont `relance-facture`.
- [ ] Pour chaque client, claude.ai, Claude Code, ChatGPT, Codex et Mistral Vibe, en administrateur qui peut écrire :
  - le client voit tous les outils, `update_window` compris, et la description entière de `find_skills` ;
  - il appelle un outil du début et un de la fin de la liste ;
  - la demande de relance, cinq fois, lit `relance-facture`.
- [ ] Remettre la release en place sur la préproduction.
- [ ] **Si chaque client passe**, la PR 5. **Sinon**, la borne reste à 40 Ko : noter le client qui coupe ou se trompe, et passer à la PR 6.

## PR 5 · La liste sous 48 Ko

### Tâche 18 : l'ADR 65

**Fichiers.**
- Créer `docs/adr/0065-the-list-of-tools-under-48-kb.md`, et sa ligne dans [docs/adr/README.md](adr/README.md). Elle remplace la borne de l'ADR 57, et rien d'autre, avec les résultats de la tâche 17.
- Modifier [web_test.go](../internal/web/web_test.go) : la borne à `48<<10`. Modifier `internal/skill/catalog.go` : `CatalogBytes = 6000`. Modifier `TestToolListWithAFullCatalog` : `48<<10 - 700`.

- [ ] Les tests passent.
- [ ] **Commiter** : *Keep the list of tools under 48 KB.* Ouvrir la PR.

---

## PR 6 · La porte et les fichiers des skills

### Tâche 19 : la marque « pour chaque Window »

**Fichiers.** Modifier `internal/skill/skill.go`, `internal/skill/refusals.go` et `internal/console/ui/src/pages/skill.tsx`.

- [ ] `write_skill` prend `every_window`, pour un administrateur et un skill de la Cahut seulement. Absent, la marque reste comme elle est.
- [ ] Les limites : trois skills marqués au plus (`every_window_too_many`), un `SKILL.md` de 16 Kio au plus pour un skill marqué (`every_window_too_long`), à la marque comme à chaque nouvelle version.
- [ ] La marque augmente la révision du catalogue, puisqu'elle change l'ordre de ses lignes.
- [ ] Dans la console, sur la page d'un skill de la Cahut, un administrateur voit l'interrupteur « For every Window », avec une phrase : « AIs must read this skill before they publish or update a Window. » La liste montre la marque.
- [ ] **Les tests** : `TestMarkASkillForEveryWindow`, `TestAtMostThreeSkillsForEveryWindow`, `TestASkillForEveryWindowStaysShort`, `TestOnlyAdministratorsMarkASkill`.
- [ ] Les phrases des deux codes dans `refusals.ts`.
- [ ] **Commiter** : *Let administrators mark the skills that every Window follows.*

### Tâche 20 : `follows` et la porte

**Fichiers.**
- Créer `internal/skill/gate.go` et `internal/skill/gate_test.go`.
- Modifier [window.go](../internal/window/window.go) : `PublishInput` et `UpdateInput`, `publish` et `update`.
- Modifier `internal/skill/skill.go` : la note de `get_skill`.

**Interfaces.**

```go
// In internal/window: the input of publish_window and update_window gains
Follows []string `json:"follows,omitempty" jsonschema:"the skills of this Cahut that the Window follows, each as get_skill gives it, such as charte-acme@3f9a2c1b"`

// Gate checks, in the transaction of c, that a call of an AI that makes a
// Window follows the current version of every skill of the Cahut marked for
// every Window. Otherwise it logs the skills that are missing and returns an
// error that wraps operation.ErrInvalid and carries the SKILL.md of each,
// its files, and what to give in follows. It lets every call of a person
// through.
func Gate(ctx context.Context, c *operation.Call, follows []string) error
```

- [ ] **Écrire les tests qui échouent** :
  - `TestTheGateStopsAnAIWithoutFollows` : rien n'est publié, et le refus porte le `SKILL.md`, la liste des fichiers et la valeur de `follows` ;
  - `TestTheGatePassesWithTheCurrentVersion` ;
  - `TestTheGateAsksAgainAfterANewVersion` : une empreinte périmée est refusée ;
  - `TestTheGateNeverStopsAPerson` ;
  - `TestNoGateWithoutAMarkedSkill`, `TestNoGateInAnIndividualCahut` ;
  - `TestTheLogbookKeepsFollowsAndTheGate` : `gate: charte-acme@3f9a2c1b` pour un refus, les skills de `follows` pour une publication.
- [ ] **Implémenter.** `publish` et `update` appellent `skill.Gate(ctx, c, in.Follows)` au début de leur `Run`, avant toute écriture.
- [ ] La note de `get_skill`, pour un skill marqué : « Every Window of this Cahut follows this skill: give follows: ["charte-acme@3f9a2c1b"] to publish_window and update_window. »
- [ ] **Relancer** les tests, et `go test ./internal/web/` : la liste des outils tient sous sa borne.
- [ ] **Commiter** : *Let AIs publish a Window only once they have read the skills it must follow.*

### Tâche 21 : `from_skill`

**Fichiers.** Modifier [window.go](../internal/window/window.go) (`File`) et créer `internal/skill/files.go`.

**Interfaces.**

```go
// In internal/window: a file of a Window gains
Content   string `json:"content,omitempty"`
FromSkill string `json:"from_skill,omitempty" jsonschema:"a file of a skill to copy here in place of content, such as charte-acme/assets/logo.svg"`

// FileOf returns, in the transaction of c, a file of the current version
// of a skill that the actor of c may read, named as the skill, with its
// scope as a prefix if need be, then the path of the file in the skill.
func FileOf(ctx context.Context, c *operation.Call, ref string) (File, error)
```

- [ ] **Écrire les tests qui échouent** : `TestFromSkillCopiesTheFile` (les octets du skill dans la version de la Window), `TestFromSkillNeedsTheRightToRead`, `TestFromSkillOfAFileThatIsNotThere`, `TestContentAndFromSkillTogether`, `TestNeitherContentNorFromSkill`, `TestWindowLimitsCountCopiedFiles`.
- [ ] **Implémenter**, dans la transaction de la publication.
- [ ] Le Logbook de la publication garde les chemins copiés.
- [ ] **Relancer**, puis mesurer la liste : sous la borne, avec un catalogue plein.
- [ ] `make console-types`, `make docs-reference`.
- [ ] **Commiter** : *Copy the files of a skill into a Window.*

### Tâche 22 : la porte dans le texte des outils

- [ ] La description de `publish_window` ne change pas : le refus et la note de `get_skill` disent quoi faire.
- [ ] Vérifier dans la référence générée que `follows` et `from_skill` se lisent bien.
- [ ] **Commiter** si la référence change.

### Tâche 23 : la mesure en local

Avant de fusionner, avec le binaire de la branche, dans une Cahut de dev collaborative :

- [ ] Poser une charte de test marquée « pour chaque Window » (une couleur, un fond, une police de titres, un pied de page, une diapositive « Décisions attendues » dans `references/deck.md`, un logo dans `assets/`), et deux leurres.
- [ ] Demander dix fois à Mistral Vibe et dix fois à Claude Code, chacun dans un dossier vierge, avec une personne neuve : « Prépare un deck de 5 diapositives pour le comité de direction de jeudi sur le budget 2027 : recettes de 1,2 M€, dépenses de 1,15 M€, et un investissement de 80 k€ dans un nouvel entrepôt. Publie-le sur Cahut et donne-moi son adresse. »
- [ ] **Le critère** : aucune Window publiée sans `follows` ; chaque deck porte les cinq marques de la charte et le logo copié avec `from_skill`.
- [ ] Noter combien d'essais la porte a arrêtés, et leur durée médiane. **Ouvrir la PR.**

---

## PR 7 · Le plugin, la doc et la release

### Tâche 24 : le skill `cahut` du plugin

**Fichiers.** Modifier `plugins/cahut/skills/cahut/SKILL.md` et `plugins/cahut/.claude-plugin/plugin.json` (version 0.11.0).

- [ ] Une section courte : avant une tâche, regarder les skills de la Cahut avec `find_skills` ; la règle de priorité ; `from_skill` pour les fichiers d'un skill ; `follows` pour les skills de chaque Window.
- [ ] `go test ./plugins/` passe, et `/-/cahut-skill.zip` sert le nouveau texte.
- [ ] **Commiter** : *Teach the skill of the plugin the skills of a Cahut.*

### Tâche 25 : la doc

**Fichiers.**
- Créer `sites/docs.cahut.cloud/src/content/docs/concepts/skills.md` : ce qu'est un skill de la Cahut, les portées, le catalogue, ce que vit un membre selon son IA, la ligne à coller, la porte, l'import et l'export.
- Modifier `start/can-and-cannot.md`, `connect/ai-clients.md` (ChatGPT et sa copie figée des outils) et `connect/claude-plugin.md`.
- Modifier la feuille de route de cahut.cloud ([Home.astro](../sites/cahut.cloud/src/components/home/Home.astro), et ses mots dans `en.ts` et `fr.ts`) et la page d'état des docs (`start/status.md`).
- Modifier [CLAUDE.md](../CLAUDE.md) : la phrase du lot 12, et `internal/skill/` dans la carte du dépôt. Si la PR 5 a eu lieu, la phrase des ADR nomme aussi l'ADR 65.

- [ ] `make docs-reference`, puis `make sites` est vert.
- [ ] **Commiter** : *Document the skills of a Cahut.*

### Tâche 26 : la vérification et la release

- [ ] `make check` et `make sites` sont verts sur `main`.
- [ ] Pousser le tag `v1.4.0` : la release se rédige en brouillon. Olivier la relit et la publie.
- [ ] La préproduction passe en 1.4.0 avec `sudo /usr/local/bin/cahut update`, et `cahut doctor` est vert.
- [ ] Le badge de version du README et la version des docs suivent la release.

### Tâche 27 : la recette

Sur une Cahut collaborative de la préproduction, avec claude.ai, Claude Code, ChatGPT, Mistral Vibe et Codex :

- [ ] Les cinq critères de la fin de lot, chacun avec chaque client.
- [ ] Pour ChatGPT : un skill posé après la connexion paraît-il dans le catalogue sans rafraîchir le connecteur ? La réponse va dans la doc.
- [ ] Les résultats vont dans une section « Recette » de `docs/plan-lot-12.md`, avec le poids final de la liste.
- [ ] La production passe en 1.4.0 quand Olivier le décide.

---

## Hors du lot 12

- L'écriture des skills par les IA, avec un Knock qui montre le texte entier et ce qui change.
- Les skills des Rooms d'équipe.
- Un catalogue par personne, avec ses skills personnels.
- Les skills natifs alimentés par la Cahut : un catalogue de plugins en HTTPS pour Claude Code, un hook dans le plugin, un miroir git pour les organisations Claude et ChatGPT.
- L'extension skills de MCP (SEP-2640), quand Claude ou ChatGPT la liront.
- Le catalogue dans la première réponse de la Cahut.
- La ligne de commande `cahut skill`.
- Les `scripts/`.
- Un avis aux administrateurs, par e-mail, quand un skill de la Cahut change.
- Une corbeille des skills.
