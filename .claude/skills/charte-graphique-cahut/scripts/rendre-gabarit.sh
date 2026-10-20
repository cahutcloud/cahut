#!/bin/sh
# Rend une page HTML de la charte (gabarit de diapositive, visuel de réseau, image de partage) en PNG
# avec Chrome sans interface. Les polices de la charte sont chargées depuis les fichiers locaux.
#
# Usage : sh rendre-gabarit.sh fichier.html [sortie.png] [LARGEURxHAUTEUR]
#   sortie.png       par défaut, le même nom que le fichier HTML, en .png, dans le même dossier
#   LARGEURxHAUTEUR  par défaut, lu dans la règle body{width:…px;height:…px} du gabarit
# Variables : ECHELLE=2 pour un rendu en double densité, CHROME=chemin pour un autre navigateur.
set -eu

[ $# -ge 1 ] || { echo "Usage : sh $0 fichier.html [sortie.png] [LARGEURxHAUTEUR]" >&2; exit 1; }
SOURCE="$1"
[ -f "$SOURCE" ] || { echo "Fichier introuvable : $SOURCE" >&2; exit 1; }
SOURCE_ABS="$(cd "$(dirname "$SOURCE")" && pwd)/$(basename "$SOURCE")"
SORTIE="${2:-${SOURCE%.html}.png}"
ECHELLE="${ECHELLE:-1}"

if [ $# -ge 3 ]; then
  TAILLE="$3"
else
  TAILLE="$(grep -oE 'body\{width:[0-9]+px;height:[0-9]+px' "$SOURCE" | head -1 | sed -E 's/.*width:([0-9]+)px;height:([0-9]+)px/\1x\2/')"
  [ -n "$TAILLE" ] || { echo "Taille introuvable dans le gabarit : passez LARGEURxHAUTEUR en troisième argument." >&2; exit 1; }
fi
LARGEUR="${TAILLE%x*}"
HAUTEUR="${TAILLE#*x}"

if [ -z "${CHROME:-}" ]; then
  for c in "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome" \
           "/Applications/Chromium.app/Contents/MacOS/Chromium" \
           "$(command -v google-chrome || true)" "$(command -v chromium || true)"; do
    if [ -n "$c" ] && [ -x "$c" ]; then CHROME="$c"; break; fi
  done
fi
[ -n "${CHROME:-}" ] || { echo "Chrome introuvable : définissez CHROME=chemin." >&2; exit 1; }

# Profil jetable, pour ne pas toucher au Chrome ouvert de l'utilisateur.
PROFIL="$(mktemp -d)"
trap 'rm -rf "$PROFIL"' EXIT

"$CHROME" --headless=new --disable-gpu --hide-scrollbars --allow-file-access-from-files \
  --user-data-dir="$PROFIL" --virtual-time-budget=3000 \
  --force-device-scale-factor="$ECHELLE" --window-size="$LARGEUR,$HAUTEUR" \
  --screenshot="$SORTIE" "file://$SOURCE_ABS" >/dev/null 2>&1

[ -s "$SORTIE" ] || { echo "Le rendu a échoué : $SORTIE est absent ou vide." >&2; exit 1; }
echo "$SORTIE ($LARGEUR × $HAUTEUR px, échelle $ECHELLE)"
