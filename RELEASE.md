# Fork release strategy

- `install.sh` belongs to `benyamins78/vpn-ui` and delegates all deployment
  logic to `deploy.sh`.
- `deploy.sh` downloads the latest release from `benyamins78/vpn-ui`, verifies
  the matching SHA256 entry before installation, and supports amd64 and arm64.
- The same installer performs both fresh installations and upgrades.
- `codex/unified-control-plane` is the active development/testing branch.
- A `v*` tag runs CI, builds amd64 and arm64 artifacts, creates `SHA256SUMS`,
  and publishes a GitHub Release. The installer consumes those release assets.

The release workflow does not publish incomplete artifacts: tests and shell
lint must pass first, and both architecture builds must succeed. The expected
release assets are `vpn-ui-amd64`, `vpn-ui-arm64`, and `SHA256SUMS`.
