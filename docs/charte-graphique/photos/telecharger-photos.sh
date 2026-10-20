#!/bin/sh
# Télécharge la sélection photo Cahut en pleine résolution (2400 px de large) depuis Unsplash.
# Licence Unsplash : https://unsplash.com/license (usage libre, crédit apprécié).
# Usage : sh telecharger-photos.sh [dossier]
set -eu
DEST="${1:-$(dirname "$0")/hd}"
mkdir -p "$DEST"

curl -fsSL -o "$DEST/ciel-cumulus.jpg" "https://images.unsplash.com/photo-1599768793949-2da217e9f093?w=2400&q=85&fm=jpg"  # Zbyněk Skrčený
curl -fsSL -o "$DEST/ciel-voile.jpg" "https://images.unsplash.com/photo-1566321343730-237ec35e53f3?w=2400&q=85&fm=jpg"  # Geri Mis
curl -fsSL -o "$DEST/cabane-alpage.jpg" "https://images.unsplash.com/photo-1610486870542-70d0062d150f?w=2400&q=85&fm=jpg"  # Ruben Hanssen
curl -fsSL -o "$DEST/cabane-nuages.jpg" "https://images.unsplash.com/photo-1734598241364-170c38f23d9c?w=2400&q=85&fm=jpg"  # Iglucraft
curl -fsSL -o "$DEST/interieur-lumiere.jpg" "https://images.unsplash.com/photo-1617326302376-c576dc5aa525?w=2400&q=85&fm=jpg"  # Deconovo
curl -fsSL -o "$DEST/porte-ouverte.jpg" "https://images.unsplash.com/photo-1763412050485-d7e1688f8858?w=2400&q=85&fm=jpg"  # Connie de Vries
curl -fsSL -o "$DEST/chaise-lumiere.jpg" "https://images.unsplash.com/photo-1554104683-c7063687d649?w=2400&q=85&fm=jpg"  # Ellen Qin
curl -fsSL -o "$DEST/bureau-maison.jpg" "https://images.unsplash.com/photo-1789540883496-4a5b0f9d2900?w=2400&q=85&fm=jpg"  # Hanna Lazar
curl -fsSL -o "$DEST/equipe-relecture.jpg" "https://images.unsplash.com/photo-1572021335469-31706a17aaef?w=2400&q=85&fm=jpg"  # Jud Mackrill
curl -fsSL -o "$DEST/duo-ecran.jpg" "https://images.unsplash.com/photo-1690264446153-c43074971475?w=2400&q=85&fm=jpg"  # Walls.io
echo "Photos enregistrées dans $DEST"
