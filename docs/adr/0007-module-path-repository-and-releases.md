# 7. Module path, repository and releases

- Status: Accepted
- Date: 2026-10-21

## Context

The Go module path appears in every import, in `go install` and in the build information of each binary. The project's domain is cahut.cloud. On GitHub, the account name `cahut` is taken by a user, and the repository still sits under a personal account.

Homes update themselves from GitHub releases, with rollback. They run AlmaLinux or Ubuntu on amd64 or arm64.

## Decision

- **Module path `cahut.cloud`**, the project's own domain, as tailscale.com does. One domain serves the site, the documentation and the code.
- **A dedicated GitHub organization** hosts the repository before batch 1 of the [V0 plan](../plan-v0.md), because the install script and the updater carry its address. Its name is still open.
- **Releases are built by GoReleaser** when a `v*` tag is pushed:
  - static binaries for linux/amd64 and linux/arm64, built with `CGO_ENABLED=0` and `-trimpath`;
  - tar.gz archives with a SHA-256 checksum file;
  - the version comes from the tag and the build information records the commit.
- **Releases are created as drafts.** A person reviews and publishes each one. Homes only see published releases.
- **Release signatures**, which the updater needs, come with batch 7.
- **One Go version source.** The `toolchain` line of `go.mod` sets the Go version for local builds, CI and releases. It is raised when Go ships a patch release.
- **Third-party GitHub Actions are pinned by commit hash.** Dependabot proposes updates every week.

## Consequences

- Moving the repository, renaming the organization or leaving GitHub changes no import.
- Before anyone runs `go install`, cahut.cloud must answer requests ending in `?go-get=1`, on every path, with a `go-import` meta tag that points to the repository. Until then, release binaries are the only way to install Cahut, which is the plan anyway.
- Draft releases add a manual step. This is wanted while releases are not signed.

## Alternatives considered

- **`github.com/<organization>/cahut`.** The common default, but it ties every import to the forge and to an organization name that is not settled.
- **A second domain such as getcahut.com.** It would work, but it splits the project across two domains for no gain.
- **A subdomain for imports only**, such as go.cahut.cloud, as go.uber.org does. Easier to host as a tiny static page. A fallback if the main site cannot serve the meta tag, at the cost of a longer path.
