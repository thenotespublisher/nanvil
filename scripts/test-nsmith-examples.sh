#!/usr/bin/env bash
# test-nsmith-examples.sh — compile example contracts for every nsmith language.
#
# Usage:
#   ./scripts/test-nsmith-examples.sh
#   NANVIL_TOOLCHAINS=/tmp/nsmith-toolchains ./scripts/test-nsmith-examples.sh
#
# Installs missing toolchains (python, csharp) when needed. Java requires Gradle
# (system gradle or ./gradlew in the example project). C# requires dotnet.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

NSMITH="${NSMITH:-$ROOT/bin/nsmith}"
EXAMPLES="$ROOT/integration/testcontracts/examples"
OUT="${OUT:-$(mktemp -d)}"
export NANVIL_TOOLCHAINS="${NANVIL_TOOLCHAINS:-$ROOT/.cache/nsmith-toolchains}"

log() { printf '→ %s\n' "$*"; }
ok()  { printf '✓ %s\n' "$*"; }
die() { printf '✗ %s\n' "$*" >&2; exit 1; }

build_nsmith() {
  log "building nsmith"
  go build -o "$NSMITH" ./cmd/nsmith/
  ok "nsmith ready"
}

ensure_toolchains() {
  log "ensuring toolchains in $NANVIL_TOOLCHAINS"
  mkdir -p "$NANVIL_TOOLCHAINS"
  "$NSMITH" install --lang python,csharp 2>&1 || true
  ok "toolchain install attempted"
}

compile_example() {
  local lang="$1" path="$2" out="$3"
  local prefix="$OUT/$out"
  log "compiling $lang example ($path)"
  "$NSMITH" compile --lang "$lang" --out "$prefix" "$path"
  [[ -f "${prefix}.nef" && -f "${prefix}.manifest.json" ]] \
    || die "$lang compile output missing under $OUT"
  ok "$lang → ${prefix}.nef"
}

main() {
  build_nsmith
  ensure_toolchains

  compile_example go     "$EXAMPLES/go"     nsmith-go
  compile_example python "$EXAMPLES/python" nsmith-python
  compile_example csharp "$EXAMPLES/csharp"  nsmith-csharp
  compile_example java   "$EXAMPLES/java"   nsmith-java

  printf '\nAll nsmith example contracts compiled successfully.\n'
  printf 'Artifacts: %s\n' "$OUT"
}

main "$@"
