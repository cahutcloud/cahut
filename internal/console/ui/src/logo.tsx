// The horizontal logo of Cahut, from the files of the brand as they are:
// the inverse one when the system shows dark colors, or when the page has a
// theme of its own and it is dark.
import logo from "@charte/logos/cahut-logo-horizontal.svg";
import inverse from "@charte/logos/cahut-logo-horizontal-inverse.svg";
import type { Theme } from "./theme";

export function Logo({ height, theme }: { height: number; theme?: Theme }) {
	if (theme) return <img src={theme === "dark" ? inverse : logo} alt="Cahut" height={height} />;
	return (
		<picture>
			<source srcset={inverse} media="(prefers-color-scheme: dark)" />
			<img src={logo} alt="Cahut" height={height} />
		</picture>
	);
}
