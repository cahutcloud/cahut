// Builds the console into dist/, which the binary embeds with the console
// build tag. The brand comes from docs/charte-graphique through the @charte
// alias, as in the sites: nothing of it is copied here.
import { fileURLToPath } from "node:url";
import preact from "@preact/preset-vite";
import { defineConfig } from "vite";

const charte = fileURLToPath(new URL("../../../docs/charte-graphique", import.meta.url));

export default defineConfig({
	plugins: [preact()],
	resolve: { alias: { "@charte": charte } },
	build: {
		outDir: "dist",
		emptyOutDir: true,
		// Every asset is a file of its own: the console's policy loads no
		// font or script from a data: address.
		assetsInlineLimit: 0,
	},
	server: {
		port: 5173,
		strictPort: true,
		fs: { allow: [".", charte] },
		// make dev serves the API of the console on port 8081. The proxy keeps
		// the host, so that the API answers as the Cahut at localhost.
		proxy: { "/api": { target: "http://localhost:8081" } },
	},
});
