# Recettes

Gestes pratiques pour appliquer la charte. Les chemins partent de la racine du dépôt ; `charte/` désigne `docs/charte-graphique/`.

## Sommaire

1. [Page dans le dépôt](#page-dans-le-dépôt)
2. [Page autonome](#page-autonome)
3. [E-mail](#e-mail)
4. [Icônes et pictogrammes](#icônes-et-pictogrammes)
5. [Gabarits et rendu en PNG](#gabarits-et-rendu-en-png)
6. [Présentation](#présentation)
7. [Photos](#photos)
8. [Modifier la charte](#modifier-la-charte)

## Page dans le dépôt

Lier les feuilles de la charte par des chemins relatifs ; `cahut-fonts.css` trouve ses polices par rapport à son propre emplacement, il fonctionne d'où qu'on le lie.

```html
<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Page title · Cahut</title>
<link rel="icon" href="../docs/charte-graphique/logos/favicon/favicon.svg" type="image/svg+xml">
<link rel="stylesheet" href="../docs/charte-graphique/tokens/cahut-fonts.css">
<link rel="stylesheet" href="../docs/charte-graphique/tokens/cahut-tokens.css">
<link rel="stylesheet" href="../docs/charte-graphique/ui/cahut-ui.css">
</head>
```

- `cahut-ui.css` pose déjà le fond, la couleur et la police du `body` à partir des rôles `--color-*`.
- Thème sombre : il suit le système ; pour le forcer, `document.documentElement.dataset.theme = "dark"` (ou `"light"`).
- Pour un site en production, servir les favicons à la racine (voir § 10 de la charte pour les balises).

## Page autonome

Pour une page qui vit hors du dépôt (artifact, page publiée ailleurs, maquette envoyée) :

1. **Polices** depuis Google Fonts :

   ```html
   <link rel="preconnect" href="https://fonts.googleapis.com">
   <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
   <link rel="stylesheet" href="https://fonts.googleapis.com/css2?family=IBM+Plex+Mono:wght@400;500;600&family=Plus+Jakarta+Sans:ital,wght@0,400..800;1,400..800&display=swap">
   ```

2. **Jetons** : lire `charte/tokens/cahut-tokens.css` et le recopier tel quel dans un `<style>`. Ne pas en retaper les valeurs de mémoire. Ajouter ensuite, si besoin, les seuls blocs de `charte/ui/cahut-ui.css` utilisés par la page.
3. **Logo** : lire le SVG voulu (par exemple `charte/logos/cahut-logo-horizontal.svg`) et l'intégrer en ligne ; il porte déjà `role="img"`, `aria-label` et `<title>`. Fixer la hauteur en CSS, garder le `viewBox`.
4. **Images** : intégrer les petites (pictogrammes, illustrations SVG) en ligne ; pour une photo, une adresse publique ou un fichier joint selon le support.

## E-mail

Les clients de messagerie ignorent les variables CSS, les SVG et souvent les polices web :

- Valeurs littérales prises dans les jetons : fond `#F7F5F0`, texte `#1F2A33`, texte secondaire `#5C6872` (Ink 600), liens `#245D83` (Sky 700), filets `#D9DCDF`, accent `#5FA8D3`.
- Mise en page en tableaux, styles en ligne, 600 px de large au plus.
- `font-family:'Plus Jakarta Sans',Arial,sans-serif` et `'IBM Plex Mono',Consolas,monospace` : la police de secours doit suffire.
- Logo et images en PNG (`charte/logos/png/`) servis depuis une adresse publique, jamais en SVG ni en chemin local.
- Signature : partir de `charte/applications/gabarits/signature-email.html`, remplacer nom, rôle et adresse de l'image.
- E-mails envoyés par un Cahut (demande *knock first* par exemple) : dire quelle IA, quelle action, au nom de qui, quelle échéance, comme la carte `c-knock` de la console.

## Icônes et pictogrammes

**Icônes d'interface (Lucide).** Récupérer le SVG sur [lucide.dev](https://lucide.dev) (ou dans le paquet `lucide-static`) et l'intégrer en ligne comme dans la console :

```html
<svg class="i" aria-hidden="true" xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><!-- tracés de l'icône --></svg>
```

- Tailles 16, 20 ou 24 px ; la couleur suit le texte. Une icône seule (bouton sans texte) prend un `aria-label` sur son bouton.

**Pictogrammes du vocabulaire** (`charte/icones/`). Les intégrer en ligne pour que le trait suive `currentColor` ; leur accent (Sky 300, Sky 400 ou Soft Yellow) est fixé dans le fichier. Taille minimale 32 px. Ils illustrent un mot de la maison, ils ne remplacent pas une icône d'interface.

## Gabarits et rendu en PNG

Les gabarits de `charte/applications/gabarits/` sont des pages HTML à taille fixe (`body{width:…px;height:…px}`) qui lient les jetons par `../../tokens/`.

1. Copier le gabarit le plus proche **dans le même dossier** (les chemins relatifs restent valables) ou corriger les chemins s'il part ailleurs.
2. Changer les textes, garder la structure : marges de 120 px sur les diapositives, étiquette en mono capitales, titre en 800 à −2 % ou −2,5 %, logo horizontal ou symbole à sa place.
3. Rendre en PNG :

   ```bash
   sh .claude/skills/charte-graphique-cahut/scripts/rendre-gabarit.sh docs/charte-graphique/applications/gabarits/og-image.html docs/charte-graphique/applications/og-image.png
   ```

   La taille est lue dans le gabarit. Pour une page sans taille fixe, la donner en troisième argument ; `ECHELLE=2` double la densité (c'est ainsi qu'est rendue la signature, 560 × 140 en double densité) :

   ```bash
   ECHELLE=2 sh .claude/skills/charte-graphique-cahut/scripts/rendre-gabarit.sh docs/charte-graphique/applications/gabarits/signature-email.html docs/charte-graphique/applications/signature-email.png 560x140
   ```

4. Ouvrir le PNG obtenu et le regarder avant de livrer : coupures de texte, débordements, polices chargées.

Formats en place : image de partage 1200 × 630, aperçu GitHub 1280 × 640, couverture de page LinkedIn 1512 × 256, bannière de profil LinkedIn 1584 × 396, bannière X 1500 × 500, avatar 400 × 400, diapositives 1920 × 1080. Les bannières sont sur Paper, avec la maison ouverte à droite ; le logo de la page ou la photo du profil se pose en bas à gauche : rien dans le quart gauche d'une couverture de page LinkedIn, rien dans le tiers gauche d'une bannière de profil LinkedIn, coin inférieur gauche libre sur X (sous y = 300, avant x = 400). LinkedIn recadre une image qui n'a pas le format attendu.

## Présentation

- **En HTML** (format préféré, cohérent avec le produit) : une page de 1920 × 1080 par diapositive, à partir des trois gabarits : titre (fond Paper, paysage à droite, titre 128 px), section (fond Ink, numéro en Sky 300, symbole inversé), contenu (étiquette, titre 68 px, points numérotés en Sky 500, carte blanche à rayon de 28 px, pied avec logo horizontal et `cahut.cloud · NN` en mono). Rendre chaque page en PNG avec le script, ou les assembler en PDF.
- **En PowerPoint**, seulement si on le demande : mêmes couleurs, Plus Jakarta Sans et IBM Plex Mono si elles sont installées sur le poste du destinataire, sinon Arial et Consolas ; logos depuis `charte/logos/png/`.
- Une idée par diapositive, au moins un tiers de vide, un seul accent Sky.

## Photos

- Les aperçus de `charte/photos/apercus/` (400 px) servent aux maquettes. Pour un rendu final, télécharger les fichiers en 2400 px :

  ```bash
  sh docs/charte-graphique/photos/telecharger-photos.sh
  ```

  Ils arrivent dans `charte/photos/hd/` ; vérifier leur poids avant de les ajouter au dépôt.
- Crédit, d'après `charte/photos/selection.json` : « Photo: Nom / Unsplash » sur un support en anglais, « Photo : Nom / Unsplash » en français.
- Nouvelle photo : même direction (lumière du jour, tons clairs, une des quatre familles), puis l'ajouter à `selection.json`, au tableau du § 7 de la charte et au script de téléchargement.

## Modifier la charte

Toute modification du logo, de la palette, des jetons ou du vocabulaire change la version de la charte. À mettre à jour ensemble :

1. `charte/README.md` : la section concernée, la version en tête et le tableau des évolutions du § 13.
2. `charte/index.html`, la version visuelle.
3. `charte/tokens/cahut-tokens.css` **et** `cahut-tokens.json` : aucun générateur ne les relie, ils se tiennent à jour à la main.
4. Si une couleur change : `charte/couleurs/cahut-palette.ase` et `.gpl`, les planches de `charte/couleurs/`, et les contrastes du § 3.
5. Les PNG qui en dépendent : planches (celle des pictogrammes a son gabarit, `charte/icones/gabarits/planche-pictogrammes.html`), gabarits rendus, captures de `ui/`, exports de logos.
6. `charte/charte-graphique-cahut.pdf` (A4 paysage, 19 pages) : le réimprimer avec `node .claude/skills/charte-graphique-cahut/scripts/imprimer-charte.mjs`, qui imprime `index.html` à l'échelle 0,75 par Chrome sans interface. Une impression directe par `--print-to-pdf` donne 28 pages. Ouvrir le PDF et contrôler la mise en page avant de livrer.
