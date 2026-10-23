// Passes the fragment of the address, such as the slide of a deck, on to
// the Window in the frame, which the browser does not do by itself.
const frame = document.querySelector("iframe");
const src = frame.getAttribute("src");
const follow = () => {
	if (location.hash) {
		frame.src = src + location.hash;
	}
};
follow();
addEventListener("hashchange", follow);

// Gives the keyboard to the Window, so that the arrows move through a deck.
frame.addEventListener("load", () => frame.focus());
