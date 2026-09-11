# Fork release strategy

- `main` tracks the upstream-compatible stable line and is not changed by this
  development task.
- `codex/unified-control-plane` is the active development/testing branch.
- A version tag such as `v1.9.4-fork.1` on a reviewed commit runs CI, builds
  amd64 and arm64 artifacts, creates `SHA256SUMS`, and publishes a GitHub
  Release. The installer consumes only those versioned-release assets.

The release workflow does not publish incomplete artifacts: tests and shell
lint must pass first, and both architecture builds must succeed. Until the fork
has a real GitHub owner/repository and a first release tag, the exact public
installer URL cannot be generated safely.
