# Changelog

## Unreleased

### Added
- GitHub Actions **Release** workflow (`workflow_dispatch`) — cross-compile `nanvil`, `ncast`, and `nsmith` and publish GitHub Releases
- Website download section with latest release from GitHub API
- `ncast` CLI for transactions and contract calls against nanvil
- `nsmith` multi-language contract compiler (Go, Python, C#, Java)
- Example contracts for all nsmith languages under `integration/testcontracts/examples/`
- `./scripts/test-nsmith-examples.sh` to compile all language examples
- `--data-dir` for persistent chain snapshots across restarts

### Removed
- neo-go CLI, examples, and documentation not required for nanvil
