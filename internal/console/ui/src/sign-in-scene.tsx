// The scene behind the card of the sign-in page: the open house of the
// brand, cut out of docs/charte-graphique/illustrations/maison-ouverte.svg
// as it is. Its four Windows come out from behind the card, float around it
// and wake now and then, as an AI works on them; once the person is signed
// in, they go back in. It is decoration only: hidden from assistive
// technologies, never under the pointer, and at rest when the system asks
// for reduced motion.
import house from "@charte/illustrations/maison-ouverte.svg?raw";
import clouds from "@charte/illustrations/motif-nuages.svg";
import { useEffect, useRef } from "preact/hooks";

/** What the sign-in page asks of the scene. */
export type SceneMotion = {
	/** Sends the Windows back behind the card, and resolves once they are in. */
	home(): Promise<void>;
	/** Shakes the card and startles the Windows, as a sign-in is refused. */
	startle(): void;
	stop(): void;
};

export function SignInScene({ card, motion }: { card: { current: HTMLElement | null }; motion: { current?: SceneMotion } }) {
	const scene = useRef<HTMLDivElement>(null);
	const { viewBox, hills, parts } = drawing();
	useEffect(() => {
		if (!scene.current || !card.current) return;
		const m = play(scene.current, card.current);
		motion.current = m;
		return () => {
			m.stop();
			motion.current = undefined;
		};
	}, []);
	return (
		<div class="scene" aria-hidden="true" ref={scene}>
			<div class="scene__clouds" style={{ backgroundImage: `url("${clouds}")` }} />
			<div class="scene__sun">
				<div class="scene__idle">
					<div class="scene__disc" />
				</div>
			</div>
			{hills.map((d, i) => (
				<div key={d} class={i ? "scene__hills scene__hills--front" : "scene__hills scene__hills--back"}>
					<div class="scene__idle">
						<svg viewBox={viewBox} preserveAspectRatio="none">
							<path d={d} />
						</svg>
					</div>
				</div>
			))}
			<svg class="scene__lines" />
			{parts.map((p) => (
				<div key={p.name} class="scene__window" data-window={p.name}>
					<div class="scene__idle">
						<div class="scene__fly" dangerouslySetInnerHTML={{ __html: p.svg }} />
					</div>
				</div>
			))}
		</div>
	);
}

type Name = "doc" | "image" | "chart" | "code";
type Part = { name: Name; svg: string; rot: number; size: [number, number] };

// The Windows of the illustration, in the order of its file, after the house.
const names: Name[] = ["doc", "image", "chart", "code"];
const ns = "http://www.w3.org/2000/svg";
let drawn: { viewBox: string; hills: string[]; parts: Part[] } | undefined;

/** Takes the hills and the Windows out of the illustration, once. */
function drawing() {
	if (drawn) return drawn;
	const art = new DOMParser().parseFromString(house, "image/svg+xml").documentElement;
	const children = [...art.children];
	drawn = {
		viewBox: art.getAttribute("viewBox") ?? "0 0 1200 800",
		hills: children.filter((e) => e.tagName === "path" && !e.hasAttribute("stroke-dasharray")).map((e) => e.getAttribute("d") ?? ""),
		parts: children
			.filter((e) => e.tagName === "g")
			.slice(1)
			.map((g, i) => cutOut(g, names[i])),
	};
	return drawn;
}

/**
 * Draws one Window on its own, upright, its tilt kept apart for the motion,
 * and names its parts, whose colours then follow the theme.
 */
function cutOut(g: Element, name: Name): Part {
	const rot = Number(/rotate\(\s*(-?[\d.]+)/.exec(g.getAttribute("transform") ?? "")?.[1] ?? 0);
	const frame = g.querySelector("rect") as Element;
	const [x, y, w, h] = ["x", "y", "width", "height"].map((a) => Number(frame.getAttribute(a)));
	const pad = Number(frame.getAttribute("stroke-width") ?? 0) / 2;
	const svg = g.ownerDocument.createElementNS(ns, "svg");
	svg.setAttribute("viewBox", `${x - pad} ${y - pad} ${w + 2 * pad} ${h + 2 * pad}`);
	svg.append(...g.children);
	const kids = [...svg.children];
	const mark = (el: Element | undefined, cls: string, draw = false) => {
		el?.setAttribute("class", cls);
		if (draw) el?.setAttribute("pathLength", "1");
	};
	const paths = kids.filter((e) => e.tagName === "path");
	mark(frame, "scene__frame");
	if (name === "doc") paths.forEach((p, i) => mark(p, i ? "scene__grey scene__draw" : "scene__ink scene__draw", true));
	if (name === "chart") {
		kids.filter((e) => e.tagName === "rect" && e !== frame).forEach((r) => mark(r, "scene__bar"));
		paths.forEach((p) => mark(p, "scene__accent scene__draw scene__trend", true));
	}
	if (name === "code") paths.forEach((p) => mark(p, "scene__accent scene__draw", true));
	if (name === "image") {
		const photo = [...(kids.find((e) => e.tagName === "svg")?.children ?? [])];
		mark(photo.find((e) => e.tagName === "rect"), "scene__photo");
		mark(photo.find((e) => e.tagName === "circle"), "scene__photo-sun");
		photo.filter((e) => e.tagName === "path").forEach((p, i) => mark(p, i ? "scene__hill-2" : "scene__hill-1"));
	}
	return { name, svg: new XMLSerializer().serializeToString(svg), rot, size: [w + 2 * pad, h + 2 * pad] };
}

// Where each Window rests: beside the card on a wide screen (its side, its
// distance from the edge of the card and its height from its middle, at full
// size), above the card on a narrow one; how far it follows the pointer, and
// how it floats.
const spec: Record<Name, { wide: { side: number; dx: number; dy: number }; narrow: { x: number; y: number; rot: number }; depth: number; float: { ax: number; ay: number; T: number; ph: number; rot: number } }> = {
	doc: { wide: { side: -1, dx: 150, dy: -115 }, narrow: { x: -114, y: -2, rot: -12 }, depth: 22, float: { ax: 3, ay: 9, T: 6.2, ph: 0, rot: 0.8 } },
	image: { wide: { side: 1, dx: 170, dy: 95 }, narrow: { x: 44, y: -16, rot: 5 }, depth: 25, float: { ax: 4, ay: 9, T: 7.4, ph: 4.4, rot: 0.9 } },
	chart: { wide: { side: -1, dx: 165, dy: 165 }, narrow: { x: -40, y: -18, rot: -5 }, depth: 18, float: { ax: 3, ay: 10, T: 6.8, ph: 3.1, rot: 0.7 } },
	code: { wide: { side: 1, dx: 120, dy: -185 }, narrow: { x: 124, y: -8, rot: 12 }, depth: 30, float: { ax: 4, ay: 8, T: 5.4, ph: 1.7, rot: 1 } },
};

// The Windows wake one after the other: a dashboard is updated, a document
// knocks first, a deck and an app are published, a document is rewritten.
const wakes: { name: Name; knock?: boolean }[] = [{ name: "chart" }, { name: "doc", knock: true }, { name: "image" }, { name: "code" }, { name: "doc" }];

type Point = { x: number; y: number };
type Box = { cx: number; cy: number; hw: number; hh: number; top: number; w: number; h: number };
type Tween = { at: number; dur: number; ease: (t: number) => number; update: (e: number, p: number) => void; start?: () => void; done?: () => void; begun?: boolean };
type Win = {
	name: Name;
	el: HTMLElement;
	idle: HTMLElement;
	fly: HTMLElement;
	art: SVGSVGElement;
	line: SVGPathElement;
	mask: SVGPathElement;
	size: [number, number];
	tilt: number;
	w: number;
	h: number;
	O: Point;
	S: Point;
	anchor: Point;
	pts: Point[];
	lut: number[];
	rot: number;
	side: number;
	amp: number;
	off: Point;
	react: { x: number; y: number; s: number };
	frozen: boolean;
	reveal: number;
};

const clamp = (v: number, a: number, b: number) => Math.min(b, Math.max(a, v));
const TAU = Math.PI * 2;
const easeOut = (t: number) => 1 - Math.pow(1 - t, 3);
const easeIn = (t: number) => t * t * t;
const easeInOut = (t: number) => (t < 0.5 ? 4 * t * t * t : 1 - Math.pow(-2 * t + 2, 3) / 2);
// Springs that settle within their time: a soft one for a path, a livelier one for a size.
const springSoft = (t: number) => 1 - Math.exp(-7 * t) * (Math.cos(8 * t) + (7 / 8) * Math.sin(8 * t));
const springPop = (t: number) => 1 - Math.exp(-6 * t) * (Math.cos(10 * t) + 0.6 * Math.sin(10 * t));
const easeInBack = (t: number, s = 1.5) => t * t * ((s + 1) * t - s);
// How long each Window takes to go back in, and how far apart they leave.
const homeTime = 620,
	homeStep = 50;

// The dotted lines of the illustration, from the middle of the card out to each Window.
const bez = (a: number, b: number, c: number, d: number, t: number) => {
	const u = 1 - t;
	return u * u * u * a + 3 * u * u * t * b + 3 * u * t * t * c + t * t * t * d;
};
const pointAt = (P: Point[], t: number): Point => ({ x: bez(P[0].x, P[1].x, P[2].x, P[3].x, t), y: bez(P[0].y, P[1].y, P[2].y, P[3].y, t) });
function curve(S: Point, E: Point, upward: boolean): Point[] {
	const dx = E.x - S.x,
		dy = E.y - S.y;
	return upward
		? [S, { x: S.x, y: S.y + dy * 0.6 }, { x: E.x, y: E.y - dy * 0.1 }, E]
		: [S, { x: S.x + dx * 0.6, y: S.y + dy * 0.1 }, { x: E.x - dx * 0.3, y: E.y }, E];
}
const pathOf = (P: Point[]) => `M${P[0].x.toFixed(1)} ${P[0].y.toFixed(1)}C${P.slice(1)
	.map((p) => `${p.x.toFixed(1)} ${p.y.toFixed(1)}`)
	.join(" ")}`;
// How far along its length a point of the curve is, so that a line keeps up with its Window.
function lengths(P: Point[]) {
	const n = 40,
		acc = [0];
	let prev = P[0],
		len = 0;
	for (let i = 1; i <= n; i++) {
		const q = pointAt(P, i / n);
		len += Math.hypot(q.x - prev.x, q.y - prev.y);
		acc.push(len);
		prev = q;
	}
	return acc.map((v) => v / (len || 1));
}
function along(L: number[], t: number) {
	t = clamp(t, 0, 1);
	const f = t * (L.length - 1),
		i = Math.floor(f);
	return i >= L.length - 1 ? 1 : L[i] + (L[i + 1] - L[i]) * (f - i);
}
const dash = (el: SVGElement, v: number) => (el.style.strokeDashoffset = String(v));

function play(scene: HTMLElement, card: HTMLElement): SceneMotion {
	const reduced = matchMedia("(prefers-reduced-motion: reduce)").matches;
	const narrowQuery = matchMedia("(max-width: 859px)");
	const parts = Object.fromEntries(drawing().parts.map((p) => [p.name, p])) as Record<Name, Part>;
	const one = <T extends Element>(s: string) => scene.querySelector(s) as T;
	const all = <T extends Element>(s: string) => [...scene.querySelectorAll<T>(s)];
	const sun = { el: one<HTMLElement>(".scene__sun"), idle: one<HTMLElement>(".scene__sun .scene__idle"), disc: one<HTMLElement>(".scene__disc"), d: 0 };
	const hills = all<HTMLElement>(".scene__hills").map((el) => ({ el, idle: el.querySelector(".scene__idle") as HTMLElement, svg: el.querySelector("svg") as SVGSVGElement }));
	const lines = one<SVGSVGElement>(".scene__lines");

	// The hills fill the foot of the page, from their top to the foot of the illustration.
	for (const h of hills) {
		const box = (h.svg.querySelector("path") as SVGPathElement).getBBox();
		const [, , vw, vh] = (h.svg.getAttribute("viewBox") ?? "").split(" ").map(Number);
		h.svg.setAttribute("viewBox", `0 ${box.y} ${vw} ${vh - box.y}`);
	}

	const defs = document.createElementNS(ns, "defs");
	lines.appendChild(defs);
	const wins: Win[] = all<HTMLElement>(".scene__window").map((el) => {
		const name = el.dataset.window as Name;
		const mask = document.createElementNS(ns, "mask");
		mask.id = `scene-${name}`;
		for (const [k, v] of Object.entries({ maskUnits: "userSpaceOnUse", x: "0", y: "0", width: "100%", height: "100%" })) mask.setAttribute(k, v);
		const reveal = document.createElementNS(ns, "path");
		reveal.setAttribute("class", "scene__reveal");
		reveal.setAttribute("pathLength", "1");
		mask.appendChild(reveal);
		defs.appendChild(mask);
		const line = document.createElementNS(ns, "path");
		line.setAttribute("class", "scene__line");
		line.setAttribute("mask", `url(#scene-${name})`);
		lines.appendChild(line);
		const origin = { x: 0, y: 0 };
		return {
			name,
			el,
			idle: el.querySelector(".scene__idle") as HTMLElement,
			fly: el.querySelector(".scene__fly") as HTMLElement,
			art: el.querySelector("svg") as SVGSVGElement,
			line,
			mask: reveal,
			size: parts[name].size,
			tilt: parts[name].rot,
			w: 0,
			h: 0,
			O: origin,
			S: origin,
			anchor: origin,
			pts: [origin, origin, origin, origin],
			lut: [0],
			rot: 0,
			side: 1,
			amp: 0,
			off: { x: 0, y: 0 },
			react: { x: 0, y: 0, s: 1 },
			frozen: false,
			reveal: 0,
		};
	});
	const byName = Object.fromEntries(wins.map((o) => [o.name, o])) as Record<Name, Win>;

	const state = { narrow: false, k: 1, C: { x: 0, y: 0 }, box: undefined as Box | undefined, px: 0, py: 0, tx: 0, ty: 0, t0: performance.now(), busy: false, started: false, wake: -1 };

	// Timers and tweens, all cleared when the page goes.
	const timers = new Set<number>();
	const later = (fn: () => void, ms: number) => {
		const id = window.setTimeout(() => {
			timers.delete(id);
			fn();
		}, ms);
		timers.add(id);
	};
	const tweens = new Set<Tween>();
	const tween = (o: { delay?: number; dur: number; ease?: (t: number) => number; update: Tween["update"]; start?: () => void; done?: () => void }) =>
		tweens.add({ ease: (t) => t, ...o, at: performance.now() + (o.delay ?? 0) });
	function stepTweens(now: number) {
		for (const tw of [...tweens]) {
			if (now < tw.at) continue;
			if (!tw.begun) {
				tw.begun = true;
				tw.start?.();
			}
			const p = Math.min(1, (now - tw.at) / tw.dur);
			tw.update(p >= 1 ? 1 : tw.ease(p), p);
			if (p >= 1) {
				tweens.delete(tw);
				tw.done?.();
			}
		}
	}

	const routeFor = (o: Win, E: Point) => curve(o.S, E, state.narrow);
	function placeAt(o: Win, pos: number, scale: number, rot: number, rotY: number) {
		const p = pointAt(o.pts, pos);
		o.fly.style.transform = `translate(${(p.x - o.anchor.x).toFixed(2)}px, ${(p.y - o.anchor.y).toFixed(2)}px) perspective(700px) rotateY(${rotY.toFixed(2)}deg) rotate(${rot.toFixed(2)}deg) scale(${scale.toFixed(3)})`;
	}

	// The inner life of each Window.
	const doc = byName.doc.art,
		chart = byName.chart.art,
		image = byName.image.art,
		code = byName.code.art;
	const docLines = [...doc.querySelectorAll<SVGPathElement>(".scene__draw")];
	const bars = [...chart.querySelectorAll<SVGRectElement>(".scene__bar")];
	const trend = chart.querySelector(".scene__trend") as SVGPathElement;
	const brackets = code.querySelector(".scene__draw") as SVGPathElement;
	const photoSun = image.querySelector(".scene__photo-sun") as SVGCircleElement;
	const chartFrame = chart.querySelector(".scene__frame") as SVGRectElement;
	const num = (el: Element, a: string) => Number(el.getAttribute(a));
	const base = num(bars[0], "y") + num(bars[0], "height");
	const xs = bars.map((b) => num(b, "x") + num(b, "width") / 2);
	const frameX = num(chartFrame, "x"),
		frameY = num(chartFrame, "y"),
		frameW = num(chartFrame, "width");
	const highest = base - (frameY + 36);
	let data = bars.map((b) => num(b, "height")),
		cur = data.slice();
	function drawChart(hs: number[]) {
		bars.forEach((b, i) => {
			b.setAttribute("y", String(base - hs[i]));
			b.setAttribute("height", String(Math.max(0, hs[i])));
		});
		// The line runs over the bars, as in the illustration.
		const pts = hs.map((h, i) => [xs[i], base - h - 8]);
		const ends = [[frameX + 22, pts[0][1] + 6], ...pts, [frameX + frameW - 14, Math.max(frameY + 16, pts[pts.length - 1][1] - 22)]];
		trend.setAttribute("d", "M" + ends.map((p) => `${p[0].toFixed(1)} ${p[1].toFixed(1)}`).join("L"));
	}
	const newData = () => xs.map((_, i) => clamp(26 + i * 11 + Math.random() * 26 - 12, 22, highest));
	const sunAt = (y: number) => photoSun.setAttribute("transform", `translate(0 ${y.toFixed(1)})`);

	function resetInner() {
		docLines.forEach((l) => dash(l, 1.05));
		cur = cur.map(() => 0);
		drawChart(cur);
		dash(trend, 1.05);
		sunAt(60);
		dash(brackets, 1.05);
	}
	function settleInner() {
		docLines.forEach((l) => dash(l, 0));
		cur = data.slice();
		drawChart(cur);
		dash(trend, 0);
		sunAt(0);
		dash(brackets, 0);
	}
	const inner: Record<Name, (again: boolean) => void> = {
		doc(again) {
			if (!again) {
				docLines.forEach((l, i) => tween({ delay: i * 85, dur: 420, ease: easeOut, update: (e) => dash(l, 1.05 * (1 - e)) }));
				return;
			}
			docLines.slice(-3).forEach((l, i) => {
				tween({ delay: i * 50, dur: 220, ease: easeIn, update: (e) => dash(l, 1.05 * e) });
				tween({ delay: 320 + i * 120, dur: 420, ease: easeOut, update: (e) => dash(l, 1.05 * (1 - e)) });
			});
		},
		chart(again) {
			if (!again) {
				data.forEach((h, i) =>
					tween({
						delay: i * 70,
						dur: 900,
						update: (_, p) => {
							cur[i] = h * (p >= 1 ? 1 : springPop(p));
							drawChart(cur);
						},
					}),
				);
				tween({ delay: 380, dur: 650, ease: easeOut, update: (e) => dash(trend, 1.05 * (1 - e)) });
				return;
			}
			const from = cur.slice(),
				to = newData();
			data = to;
			tween({
				dur: 1000,
				ease: easeInOut,
				update: (e) => {
					cur = from.map((v, i) => v + (to[i] - v) * e);
					drawChart(cur);
				},
			});
		},
		image(again) {
			const rise = () => tween({ dur: 1100, update: (_, p) => sunAt(60 * (1 - (p >= 1 ? 1 : springSoft(p)))) });
			if (again) tween({ dur: 280, ease: easeIn, update: (e) => sunAt(60 * e), done: rise });
			else rise();
		},
		code(again) {
			const draw = () => tween({ dur: 750, ease: easeInOut, update: (e) => dash(brackets, 1.05 * (1 - e)) });
			if (again) tween({ dur: 240, ease: easeIn, update: (e) => dash(brackets, 1.05 * e), done: draw });
			else draw();
		},
	};

	// From the layout of the card, so that its own motion does not move the scene.
	const cardBox = (): Box => {
		const hw = card.offsetWidth / 2,
			hh = card.offsetHeight / 2;
		return { cx: card.offsetLeft + hw, cy: card.offsetTop + hh, hw, hh, top: card.offsetTop, w: scene.offsetWidth, h: scene.offsetHeight };
	};

	// Sizes, at the start and when the window changes.
	function layout() {
		state.narrow = narrowQuery.matches;
		const k = (state.k = state.narrow ? 0.5 : clamp(Math.min(innerWidth / 1500, innerHeight / 950), 0.6, 0.95));
		for (const o of wins) {
			o.w = o.size[0] * k;
			o.h = o.size[1] * k;
			o.art.setAttribute("width", String(o.w));
			o.art.setAttribute("height", String(o.h));
		}
		sun.d = state.narrow ? 72 : 116 * k;
		sun.el.style.width = sun.el.style.height = `${sun.d}px`;
		lines.style.display = state.narrow ? "none" : "";
		place(cardBox());
	}

	// Positions, from where the card is.
	function place(b: Box) {
		const k = state.k;
		state.box = b;
		state.C = { x: b.cx, y: b.cy };
		for (const o of wins) {
			const s = spec[o.name];
			o.O = state.narrow ? { x: b.cx + s.narrow.x, y: b.top + s.narrow.y } : { x: b.cx + s.wide.side * (b.hw + s.wide.dx * k), y: b.cy + s.wide.dy * k };
			o.rot = state.narrow ? s.narrow.rot : o.tilt;
			o.side = state.narrow ? Math.sign(s.narrow.x) || 1 : s.wide.side;
			o.S = state.narrow ? { x: o.O.x, y: b.top + 120 } : { ...state.C };
			o.el.style.transform = `translate(${(o.O.x - o.w / 2).toFixed(1)}px, ${(o.O.y - o.h / 2).toFixed(1)}px)`;
		}
		const sx = state.narrow ? b.cx + 70 : b.cx + b.hw * 0.62,
			sy = state.narrow ? b.top - 50 : b.top - 2;
		sun.el.style.transform = `translate(${(sx - sun.d / 2).toFixed(1)}px, ${(sy - sun.d / 2).toFixed(1)}px)`;
	}

	// When the card grows, as an alert shows, the scene follows it.
	const follow = new ResizeObserver(() => {
		if (!state.box) return;
		const from = state.box,
			to = cardBox();
		if (Math.abs(from.cy - to.cy) < 0.5 && Math.abs(from.hw - to.hw) < 0.5) return;
		if (reduced) return place(to);
		const keys = Object.keys(to) as (keyof Box)[];
		tween({ dur: 450, ease: easeOut, update: (e) => place(Object.fromEntries(keys.map((key) => [key, from[key] + (to[key] - from[key]) * e])) as Box) });
	});
	follow.observe(card);

	// Each frame: the pointer, the floating, the lines.
	let raf = 0;
	function frame(now: number) {
		stepTweens(now);
		const t = (now - state.t0) / 1000,
			m = !reduced;
		state.px += (state.tx - state.px) * 0.05;
		state.py += (state.ty - state.py) * 0.05;
		const px = m ? state.px : 0,
			py = m ? state.py : 0;
		for (const o of wins) {
			if (o.frozen) continue;
			const f = spec[o.name].float,
				a = m ? o.amp : 0,
				d = spec[o.name].depth * (state.narrow ? 0.3 : 1);
			const x = px * d + a * f.ax * Math.sin((TAU * t) / f.T + f.ph) + o.react.x;
			const y = py * d * 0.7 + a * f.ay * Math.sin((TAU * t) / (f.T * 0.83) + f.ph * 1.7) + o.react.y;
			const r = a * f.rot * Math.sin((TAU * t) / (f.T * 1.3) + f.ph);
			o.off = { x, y };
			o.idle.style.transform = `translate3d(${x.toFixed(2)}px, ${y.toFixed(2)}px, 0) rotate(${r.toFixed(3)}deg) scale(${o.react.s.toFixed(3)})`;
		}
		sun.idle.style.transform = `translate3d(${(px * 10).toFixed(2)}px, ${(py * 8 + (m ? 3 * Math.sin((TAU * t) / 11) : 0)).toFixed(2)}px, 0)`;
		hills[0].idle.style.transform = `translate3d(${(px * 7).toFixed(2)}px, ${(py * 3).toFixed(2)}px, 0)`;
		hills[1].idle.style.transform = `translate3d(${(px * 16).toFixed(2)}px, ${(py * 5).toFixed(2)}px, 0)`;
		if (!state.narrow) {
			for (const o of wins) {
				const d = pathOf(curve(state.C, { x: o.O.x + o.off.x, y: o.O.y + o.off.y }, false));
				o.line.setAttribute("d", d);
				o.mask.setAttribute("d", d);
				dash(o.mask, Number((1.05 * (1 - o.reveal)).toFixed(4)));
			}
		}
		raf = requestAnimationFrame(frame);
	}

	// A Window comes out from behind the card.
	function fly(o: Win, delay: number) {
		o.amp = 0;
		o.frozen = false;
		o.anchor = o.O;
		o.reveal = 0;
		o.pts = routeFor(o, o.O);
		o.lut = lengths(o.pts);
		o.el.style.opacity = "0";
		placeAt(o, 0, 0.3, 0, o.side * 55);
		tween({
			delay,
			dur: 1400,
			start: () => (o.el.style.opacity = ""),
			update: (_, p) => {
				const pos = p >= 1 ? 1 : springSoft(p),
					pop = p >= 1 ? 1 : springPop(p);
				placeAt(o, pos, 0.3 + 0.7 * pop, o.rot * pop, o.side * 55 * (1 - Math.min(1, pos)));
				o.reveal = along(o.lut, pos);
			},
		});
		later(() => inner[o.name](false), delay + 520);
		tween({ delay: delay + 900, dur: 1500, ease: easeInOut, update: (e) => (o.amp = e) });
	}

	// And goes back in, with a step back first.
	function goHome(o: Win, delay: number) {
		tween({
			delay,
			dur: homeTime,
			start: () => {
				o.frozen = true;
				o.anchor = { x: o.O.x + o.off.x, y: o.O.y + o.off.y };
				o.pts = routeFor(o, o.anchor);
				o.lut = lengths(o.pts);
			},
			update: (_, p) => {
				const u = easeInBack(p, 1.2),
					q = clamp(u, 0, 1);
				placeAt(o, 1 - u, 1 - 0.7 * q, o.rot * (1 - q), o.side * 55 * q);
				o.reveal = along(o.lut, 1 - u);
			},
			done: () => (o.el.style.opacity = "0"),
		});
	}

	// A Window wakes and shows what changed in it; for a Knock, it knocks twice towards the card.
	function wake(w: (typeof wakes)[number]) {
		const o = byName[w.name];
		if (w.knock) {
			tween({ dur: 760, update: (_, p) => (o.react.x = -o.side * 9 * state.k * Math.max(0, Math.sin(p * Math.PI * 4)) * (1 - p * 0.4)), done: () => (o.react.x = 0) });
			return;
		}
		tween({
			dur: 900,
			update: (_, p) => {
				const b = Math.sin(Math.PI * Math.min(1, p * 1.6));
				o.react.y = -10 * state.k * b;
				o.react.s = 1 + 0.05 * b;
			},
			done: () => {
				o.react.y = 0;
				o.react.s = 1;
			},
		});
		inner[w.name](true);
	}
	function cycle() {
		later(() => {
			if (!state.busy) {
				state.wake = (state.wake + 1) % wakes.length;
				wake(wakes[state.wake]);
			}
			cycle();
		}, 3800);
	}

	// Everything at rest: with reduced motion, or after the window changed.
	function settle() {
		for (const o of wins) {
			o.el.style.opacity = "";
			o.frozen = false;
			o.anchor = o.O;
			o.reveal = 1;
			o.pts = routeFor(o, o.O);
			o.amp = reduced ? 0 : 1;
			o.react = { x: 0, y: 0, s: 1 };
			placeAt(o, 1, 1, o.rot, 0);
		}
		settleInner();
		sun.disc.style.transform = "";
		sun.disc.style.opacity = "";
		hills.forEach((h) => (h.svg.style.transform = ""));
		scene.style.opacity = "1";
	}

	function intro() {
		layout();
		state.started = true;
		if (reduced) return settle();
		tween({ dur: 900, update: (e) => (scene.style.opacity = String(e)) });
		hills.forEach((h, i) => {
			const from = 60 + i * 40;
			h.svg.style.transform = `translateY(${from}px)`;
			tween({ delay: i * 120, dur: 1300, ease: easeOut, update: (e) => (h.svg.style.transform = `translateY(${((1 - e) * from).toFixed(1)}px)`) });
		});
		// The sun rises from behind the card.
		const rise = state.narrow ? 110 : 150 * state.k;
		sun.disc.style.transform = `translateY(${rise}px)`;
		sun.disc.style.opacity = "0";
		tween({ delay: 300, dur: 1900, ease: easeOut, start: () => (sun.disc.style.opacity = ""), update: (e) => (sun.disc.style.transform = `translateY(${((1 - e) * rise).toFixed(1)}px)`) });
		resetInner();
		(["doc", "code", "chart", "image"] as Name[]).forEach((n, i) => fly(byName[n], 550 + i * 115));
		later(cycle, 2200);
	}

	const onPointer = (e: PointerEvent) => {
		state.tx = (e.clientX / innerWidth - 0.5) * 2;
		state.ty = (e.clientY / innerHeight - 0.5) * 2;
	};
	const onLeave = () => {
		state.tx = 0;
		state.ty = 0;
	};
	const fine = matchMedia("(pointer: fine)").matches;
	if (fine) {
		addEventListener("pointermove", onPointer);
		document.addEventListener("pointerleave", onLeave);
	}
	let resizing = 0;
	const onResize = () => {
		clearTimeout(resizing);
		resizing = window.setTimeout(() => {
			if (state.busy) return;
			tweens.clear();
			layout();
			settle();
		}, 150);
	};
	addEventListener("resize", onResize);

	intro();
	raf = requestAnimationFrame(frame);

	return {
		home() {
			if (!state.started || reduced) return Promise.resolve();
			state.busy = true;
			return new Promise((resolve) => {
				// The console opens once the last of them is in.
				(["image", "chart", "code", "doc"] as Name[]).forEach((n, i) => goHome(byName[n], i * homeStep));
				later(resolve, 3 * homeStep + homeTime + 60);
			});
		},
		startle() {
			if (reduced) return;
			tween({ dur: 480, update: (_, p) => (card.style.transform = p >= 1 ? "" : `translateX(${(10 * Math.sin(p * Math.PI * 5) * (1 - p)).toFixed(2)}px)`) });
			wins.forEach((o, i) =>
				tween({ delay: i * 45, dur: 620, update: (_, p) => (o.react.y = -12 * state.k * Math.sin(Math.PI * Math.min(1, p * 2)) * (1 - p)), done: () => (o.react.y = 0) }),
			);
		},
		stop() {
			cancelAnimationFrame(raf);
			timers.forEach(clearTimeout);
			clearTimeout(resizing);
			tweens.clear();
			follow.disconnect();
			removeEventListener("resize", onResize);
			if (fine) {
				removeEventListener("pointermove", onPointer);
				document.removeEventListener("pointerleave", onLeave);
			}
			card.style.transform = "";
		},
	};
}
