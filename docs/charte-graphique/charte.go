// Package charte embeds the files of the brand of Cahut that the binary
// serves, as they are in the brand guidelines: the pages that a Cahut shows
// readers carry the brand of Cahut until it has a house style of its own.
// See docs/adr/0032-the-page-of-a-room.md.
package charte

import "embed"

// Files are the tokens, the fonts and their licenses, the interface kit,
// the logos, the icons and the pictures that the pages, the console and the
// MCP endpoint of a Cahut use. They keep their paths, so that the fonts
// sheet finds the fonts. See docs/adr/0040-the-icon-of-a-cahut.md.
//
//go:embed tokens/cahut-tokens.css tokens/cahut-fonts.css ui/cahut-ui.css
//go:embed typographie/polices/web/*.woff2 typographie/polices/OFL-*.txt
//go:embed logos/cahut-logo-horizontal.svg logos/cahut-logo-horizontal-inverse.svg logos/cahut-favicon.svg
//go:embed logos/cahut-icone-app.svg logos/favicon/favicon.ico logos/favicon/apple-touch-icon.png logos/favicon/icon-512.png
//go:embed icones/room.svg illustrations/etat-vide.svg illustrations/motif-nuages-sky.svg
var Files embed.FS
