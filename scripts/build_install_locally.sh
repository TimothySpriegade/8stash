#!/usr/bin/env bash
set -euo pipefail

# Configurable via env:
# - BIN_NAME: force binary name
# - MAIN_PKG: force main package to build (e.g. ./cmd/app)
# - INSTALL_DIR: install destination (default: ~/bin or ~/.local/bin)

MAIN_PKG=../cmd/8stash
command -v go >/dev/null 2>&1 || { echo "go not found in PATH"; exit 1; }

# Choose install dir
if [ -n "${INSTALL_DIR:-}" ]; then
  DEST="$INSTALL_DIR"
else
  if [ -d "$HOME/bin" ]; then
    DEST="$HOME/bin"
  else
    DEST="$HOME/.local/bin"
  fi
fi
mkdir -p "$DEST"

# Detect main package
if [ -n "${MAIN_PKG:-}" ]; then
  PKG="$MAIN_PKG"
else
  mapfile -t mains < <(go list -f '{{if eq .Name "main"}}{{.ImportPath}}{{end}}' ./... | sed '/^$/d') || true
  if [ "${#mains[@]}" -ge 1 ]; then
    PKG="${mains[0]}"
  else
    PKG="."
  fi
fi

# Decide binary name
if [ -n "${BIN_NAME:-}" ]; then
  NAME="$BIN_NAME"
else
  mod="$(go list -m -f '{{.Path}}' 2>/dev/null || true)"
  if [ -n "$mod" ]; then
    NAME="${mod##*/}"
  else
    NAME="$(basename "$(pwd)")"
  fi
fi

# Build to a temp file, then install atomically
tmpdir="$(mktemp -d)"
trap 'rm -rf "$tmpdir"' EXIT

out="$tmpdir/$NAME"
echo "Building $PKG -> $out"
GO111MODULE=on go build -trimpath -ldflags "-s -w" -o "$out" "$PKG"

# Install (replaces old version if exists)
echo "Installing to $DEST/$NAME"
install -m 0755 "$out" "$DEST/$NAME"

echo "Installed: $DEST/$NAME"
