// The deck engine of Cahut, served at /-/deck.js. It shows the slides of
// a deck one at a time, as large as the screen allows: the arrows, Space,
// Page Up and Page Down, Home and End move between them, F goes full
// screen, O shows them all, and the address carries the number of the
// slide, such as #5. A deck is a .hs-deck of .hs-slide; guide.md describes
// it. Apache-2.0: see LICENSE in this directory.
(() => {
	const deck = document.querySelector(".hs-deck");
	const slides = deck ? [...deck.querySelectorAll(":scope > .hs-slide")] : [];
	if (slides.length === 0) {
		return;
	}
	let at = 0;

	const make = (tag, className, text) => {
		const el = document.createElement(tag);
		el.className = className;
		if (text) {
			el.textContent = text;
		}
		return el;
	};
	const progress = make("div", "hs-progress");
	const controls = make("nav", "hs-controls");
	controls.setAttribute("aria-label", "Slides");
	const button = (label, symbol, action) => {
		const b = make("button", "", symbol);
		b.type = "button";
		b.setAttribute("aria-label", label);
		b.title = label;
		b.addEventListener("click", action);
		return b;
	};
	const count = make("span", "");
	count.setAttribute("aria-live", "polite");
	controls.append(
		button("Previous slide", "‹", () => go(at - 1)),
		count,
		button("Next slide", "›", () => go(at + 1)),
		button("Full screen", "⤢", fullscreen),
	);
	deck.append(progress, controls);

	slides.forEach((slide, i) => {
		slide.setAttribute("role", "group");
		slide.setAttribute("aria-roledescription", "slide");
		slide.setAttribute("aria-label", `${i + 1} of ${slides.length}`);
	});

	function go(i, fromAddress = false) {
		at = Math.max(0, Math.min(slides.length - 1, i));
		slides.forEach((slide, j) => {
			slide.classList.toggle("hs-current", j === at);
			slide.inert = live() && j !== at;
		});
		progress.style.setProperty("--hs-at", String((at + 1) / slides.length));
		count.textContent = `${at + 1} / ${slides.length}`;
		if (!fromAddress && location.hash !== `#${at + 1}`) {
			try {
				history.replaceState(null, "", `#${at + 1}`);
			} catch {
				location.hash = String(at + 1);
			}
		}
	}

	function live() {
		return deck.classList.contains("hs-live");
	}

	// The slide that the address names: its number, or an element in it.
	function fromAddress() {
		const hash = decodeURIComponent(location.hash.slice(1));
		if (/^\d+$/.test(hash)) {
			go(Number(hash) - 1, true);
			return;
		}
		const target = hash && document.getElementById(hash);
		const slide = target && target.closest(".hs-slide");
		go(slide ? slides.indexOf(slide) : at, true);
	}

	function fullscreen() {
		try {
			if (document.fullscreenElement) {
				document.exitFullscreen();
			} else {
				document.documentElement.requestFullscreen();
			}
		} catch {
			// A frame that does not allow full screen.
		}
	}

	// O shows every slide down the page, and back.
	function overview() {
		deck.classList.toggle("hs-live");
		go(at, true);
		if (!live()) {
			slides[at].scrollIntoView({ block: "center" });
		}
	}

	document.addEventListener("keydown", (e) => {
		if (e.altKey || e.ctrlKey || e.metaKey || e.defaultPrevented) {
			return;
		}
		// Keys that type go to fields, and Space and Enter to buttons and links.
		const t = e.target;
		if (t instanceof HTMLElement && (t.isContentEditable || /^(INPUT|TEXTAREA|SELECT)$/.test(t.tagName) ||
			/^(BUTTON|A|SUMMARY)$/.test(t.tagName) && (e.key === " " || e.key === "Enter"))) {
			return;
		}
		const moves = {
			ArrowRight: 1, ArrowDown: 1, PageDown: 1, " ": 1, Enter: 1,
			ArrowLeft: -1, ArrowUp: -1, PageUp: -1, Backspace: -1,
		};
		if (e.key === "o" || e.key === "O") {
			overview();
		} else if (!live()) {
			return;
		} else if (e.key in moves) {
			go(at + (e.shiftKey && e.key === " " ? -1 : moves[e.key]));
		} else if (e.key === "Home") {
			go(0);
		} else if (e.key === "End") {
			go(slides.length - 1);
		} else if (e.key === "f" || e.key === "F") {
			fullscreen();
		} else {
			return;
		}
		e.preventDefault();
	});

	// A swipe moves to the next slide or the previous one.
	let start = null;
	deck.addEventListener("pointerdown", (e) => {
		start = e.pointerType === "mouse" ? null : { x: e.clientX, y: e.clientY };
	});
	deck.addEventListener("pointerup", (e) => {
		if (!start || !live()) {
			return;
		}
		const dx = e.clientX - start.x;
		const dy = e.clientY - start.y;
		start = null;
		if (Math.abs(dx) > 40 && Math.abs(dx) > Math.abs(dy)) {
			go(at + (dx < 0 ? 1 : -1));
		}
	});

	// The controls show while the pointer moves, and hide after a while.
	let idle;
	const wake = () => {
		deck.classList.remove("hs-idle");
		clearTimeout(idle);
		idle = setTimeout(() => deck.classList.add("hs-idle"), 2500);
	};
	deck.addEventListener("pointermove", wake);
	deck.addEventListener("focusin", wake);

	addEventListener("hashchange", fromAddress);
	deck.classList.add("hs-live");
	fromAddress();
	wake();
})();
