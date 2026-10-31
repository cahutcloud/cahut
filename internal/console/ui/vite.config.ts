// Builds the console into dist/, which the binary embeds with the console
// build tag. The brand comes from docs/charte-graphique through the @charte
// alias, as in the sites: nothing of it is copied here.
import { fileURLToPath } from "node:url";
import preact from "@preact/preset-vite";
import { defineConfig, type Plugin } from "vite";

const charte = fileURLToPath(new URL("../../../docs/charte-graphique", import.meta.url));

// In development, the addresses of the console of the server, but its API,
// open its page, as the server does once the console is built.
function serverConsole(): Plugin {
	return {
		name: "server-console",
		configureServer(server) {
			server.middlewares.use((req, _res, next) => {
				if (req.url && /^\/server(\/|$|\?)/.test(req.url) && !req.url.startsWith("/server/api/")) {
					req.url = "/console/server.html";
				}
				next();
			});
		},
	};
}

export default defineConfig({
	plugins: [preact(), serverConsole()],
	resolve: { alias: { "@charte": charte } },
	// The console is served under /console/ of its Cahut's domain, and its
	// assets with it; the console of the server, under /server/, loads them
	// from there. See docs/adr/0020-console-under-a-path-of-port-443.md and
	// docs/adr/0058-a-console-for-the-server.md.
	base: "/console/",
	build: {
		outDir: "dist",
		emptyOutDir: true,
		// Every asset is a file of its own: the console's policy loads no
		// font or script from a data: address.
		assetsInlineLimit: 0,
		rollupOptions: {
			input: {
				console: fileURLToPath(new URL("index.html", import.meta.url)),
				server: fileURLToPath(new URL("server.html", import.meta.url)),
			},
		},
	},
	server: {
		port: 5173,
		strictPort: true,
		fs: { allow: [".", charte] },
		// make dev serves the API of the console on port 8080, and the icons
		// that its page names under /-/brand/. The proxy keeps the host, so
		// that the API answers as the Cahut at localhost.
		proxy: {
			"/console/api": { target: "http://localhost:8080" },
			"/server/api": { target: "http://localhost:8080" },
			"/-/brand": { target: "http://localhost:8080" },
		},
	},
});
