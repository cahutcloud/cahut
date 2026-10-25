// The console of a Cahut: its brand, then the app.
import "@charte/tokens/cahut-fonts.css";
import "@charte/tokens/cahut-tokens.css";
import "@charte/ui/cahut-ui.css";
import "./console.css";
import { render } from "preact";
import { App } from "./app";

render(<App />, document.getElementById("app") as HTMLElement);
