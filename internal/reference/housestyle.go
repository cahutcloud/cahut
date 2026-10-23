package main

import (
	"strings"

	"cahut.cloud/internal/housestyle"
)

// houseStylePage is the guide to the style sheet of the house style, as
// get_house_style gives it to AIs.
func houseStylePage() page {
	p := page{
		file:        "house-style-sheet.md",
		title:       "House style sheet",
		description: "The variables and the classes of /-/house-style.css, and the markup of the deck engine, as get_house_style gives them to AIs.",
		order:       5,
	}
	p.body.block("The guide that `get_house_style` gives an AI with the house style of a Cahut. " +
		"[House style](/concepts/house-style/) says what the house style holds and how it changes.")
	guide := housestyle.Guide()
	// The page has its title: the guide's own goes.
	if _, rest, ok := strings.Cut(guide, "\n"); ok && strings.HasPrefix(guide, "# ") {
		guide = rest
	}
	p.body.WriteString(strings.TrimLeft(guide, "\n"))
	return p
}
