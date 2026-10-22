// The horizontal logo of Cahut, from the files of the brand as they are:
// the inverse one when the system shows dark colors.
import logo from "@charte/logos/cahut-logo-horizontal.svg";
import inverse from "@charte/logos/cahut-logo-horizontal-inverse.svg";

export function Logo({ height }: { height: number }) {
	return (
		<picture>
			<source srcset={inverse} media="(prefers-color-scheme: dark)" />
			<img src={logo} alt="Cahut" height={height} />
		</picture>
	);
}
