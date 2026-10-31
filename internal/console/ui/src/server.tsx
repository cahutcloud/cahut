// The console of the server: its brand, then the app. See
// docs/adr/0058-a-console-for-the-server.md.
import "@charte/tokens/cahut-fonts.css";
import "@charte/tokens/cahut-tokens.css";
import "@charte/ui/cahut-ui.css";
import "./console.css";
import { render } from "preact";
import { ServerApp } from "./server/app";

render(<ServerApp />, document.getElementById("app") as HTMLElement);
