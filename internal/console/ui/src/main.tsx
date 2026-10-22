// The console of a Cahut: its brand, then the app.
import "@charte/tokens/cahut-fonts.css";
import "@charte/tokens/cahut-tokens.css";
import "@charte/ui/cahut-ui.css";
import "./console.css";
import favicon from "@charte/logos/cahut-favicon.svg";
import { render } from "preact";
import { App } from "./app";

const icon = document.createElement("link");
icon.rel = "icon";
icon.type = "image/svg+xml";
icon.href = favicon;
document.head.append(icon);

render(<App />, document.getElementById("app") as HTMLElement);
