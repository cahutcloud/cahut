# Rulesets

The rulesets of the repository, kept here so that they are reviewed like
code. GitHub does not read this folder: apply a file with the API, or import
it from Settings, Rules, Rulesets.

```bash
gh api -X POST repos/cahutcloud/cahut/rulesets --input .github/rulesets/main.json
```

- `main.json` protects `main`: no deletion and no force push, a pull request
  merged by squash, the Lint, Test, Build and Install checks of CI and the
  DCO check of the sign-offs green, and every
  review thread resolved. Administrators may merge a pull request without
  its checks in an emergency, never push to `main` directly. No approval is
  required while one person maintains the project.
- `release-tags.json` keeps the `v*` tags to administrators: nobody else
  creates, moves or deletes a release tag (`release.yml` drafts a release
  from it).

To change a ruleset, change its file in a pull request, then apply it again
with `PUT repos/cahutcloud/cahut/rulesets/<id>`.
