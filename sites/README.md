# Sites

The website and the documentation of Cahut, as decided in [ADR 12](../docs/adr/0012-website-and-docs-as-astro-sites.md).

| Folder | Address | What it is |
| --- | --- | --- |
| `cahut.cloud/` | https://cahut.cloud | The website, an [Astro](https://docs.astro.build) site, with the brand at `/brand` |
| `docs.cahut.cloud/` | https://docs.cahut.cloud | The documentation, an Astro site with [Starlight](https://starlight.astro.build) |
| `brand/` | | `@cahut/brand`, the Astro integration of the brand that both sites use |

The folders form one npm workspace, with one lockfile, `package-lock.json`.

## Working on the sites

The sites need Node.js at the version of [`.nvmrc`](../.nvmrc). With nvm, run `nvm use` in the repository. Then, from the repository's root:

| Task | What it does |
| --- | --- |
| `make site-dev` | Serves cahut.cloud on http://localhost:4321 |
| `make docs-dev` | Serves docs.cahut.cloud on http://localhost:4322 |
| `make sites` | Builds both sites into their `dist/` folders, as CI does |
| `make docs-reference` | Writes the reference of the docs from the code, with Go only |

The first task installs the packages with `npm ci`. To add a package, run `npm install <package> --workspace <site>` in `sites/`, and state why in the pull request. Packages run no install scripts (`.npmrc`).

## The reference of the docs

The pages of `docs.cahut.cloud/src/content/docs/reference/` come from the code: the MCP tools as a Cahut serves them to an AI, the operations of the registries, and the commands as `cahut` describes them. `make docs-reference` writes them with [`internal/reference`](../internal/reference/main.go), which owns that folder: never edit its pages by hand. Run it after changing an operation, a schema or a command, and commit the result with the change. CI fails when the committed reference is out of date.

## The brand

The brand has one source: [`docs/charte-graphique`](../docs/charte-graphique/README.md). The sites never copy its files: they import them through the `@charte` alias, which the `@cahut/brand` integration maps to that folder.

```astro
---
import '@charte/tokens/cahut-fonts.css';
import '@charte/tokens/cahut-tokens.css';
import logo from '@charte/logos/cahut-logo-horizontal.svg?url';
---
```

The integration also serves the favicons and the web manifest of `docs/charte-graphique/logos/favicon` at the root of each site. Colors, type, spacing, radii and shadows come from the tokens: never retype a value.

## Deployment

The [Sites workflow](../.github/workflows/sites.yml) builds both sites for every pull request that touches them or the brand. On `main`, it uploads them to the web server with rsync over SSH, into the document root of each domain, and deletes the files that are no longer built. It never deletes a hidden file or folder, such as `.well-known/` or the PHP settings that Plesk keeps there. To upload `main` again without a change, run the workflow from the Actions tab.

The upload stays off until the repository's variables name the server. To turn it on, on the server:

1. Serve `cahut.cloud` and `docs.cahut.cloud` as static sites over HTTPS, and redirect `www.cahut.cloud` to `cahut.cloud`.
2. Give the system user of these sites an SSH shell that can run rsync.
3. Copy [`deploy-command.sh`](deploy-command.sh) to that user's `~/.ssh/`, executable. It is the key's forced command: the key can then only upload into `httpdocs/` or `docs.cahut.cloud/` with rsync, with no shell.
4. Add the public half of a new key, used for nothing else, to that user's `~/.ssh/authorized_keys`, behind that command:

   ```
   command="/var/www/vhosts/cahut.cloud/.ssh/deploy-command.sh",restrict ssh-ed25519 AAAA... sites-deploy
   ```

Then, in the repository's settings on GitHub, under Secrets and variables, Actions:

| Name | Kind | Value |
| --- | --- | --- |
| `SITES_SSH_KEY` | Secret | The private half of that key |
| `SITES_SSH_HOST` | Variable | The server's host name |
| `SITES_SSH_USER` | Variable | The system user of the sites |
| `SITES_SSH_KNOWN_HOSTS` | Variable | The server's host keys, from `ssh-keyscan <host>`, checked against the server |
| `SITES_SSH_PORT` | Variable | The SSH port, if not 22 |
| `SITES_WWW_DIR` | Variable | The document root of cahut.cloud, relative to the user's home, if not `httpdocs` |
| `SITES_DOCS_DIR` | Variable | The document root of docs.cahut.cloud, if not `docs.cahut.cloud` |

The default folders are those of Plesk for a domain and its subdomain. If you change them, change `deploy-command.sh` too.
