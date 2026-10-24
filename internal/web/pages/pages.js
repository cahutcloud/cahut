// Shows the preview of each Window on the page of its Room: the Window
// drawn 1280 pixels wide, shrunk to its card, over the icon of its kind
// once it loads. Without this script, the cards keep their icon. See
// docs/adr/0032-the-page-of-a-room.md.
const width = 1280;
for (const frame of document.querySelectorAll("iframe.card__preview")) {
	new ResizeObserver(([entry]) => {
		frame.style.setProperty("--preview-scale", entry.contentRect.width / width);
	}).observe(frame.parentElement);
	// The frame loads once its listener is there, so that none is missed.
	frame.src = frame.dataset.src;
	frame.addEventListener("load", () => frame.classList.add("card__preview--shown"));
}
