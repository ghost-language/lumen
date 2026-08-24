#!/bin/bash
set -euo pipefail

# Only needed in Claude Code on the web sessions -- a local checkout already
# has ../ghost and system SDL2 packages from the developer's own setup.
if [ "${CLAUDE_CODE_REMOTE:-}" != "true" ]; then
  exit 0
fi

REPO_ROOT="${CLAUDE_PROJECT_DIR:-$(pwd)}"
GHOST_DIR="$(dirname "$REPO_ROOT")/ghost"

# lumen's go.mod has `replace ghostlang.org/x/ghost => ../ghost`; that sibling
# checkout has to exist before any `go build`/`go test` can resolve it.
if [ ! -d "$GHOST_DIR/.git" ]; then
  echo "Cloning ghostlang.org/x/ghost to $GHOST_DIR..."
  GIT_LFS_SKIP_SMUDGE=1 git clone --depth 1 https://github.com/ghost-language/ghost "$GHOST_DIR"
fi

# proxy.golang.org is blocked by this environment's default network policy.
# Fetch Go modules straight from their source repos over git (which is
# allowed) instead, and skip the checksum-database lookup that would
# otherwise need sum.golang.org too.
go env -w GOPROXY=direct GOSUMDB=off

# Lumen links SDL2 through cgo, so the dev packages (headers + .pc files)
# have to be present for `go build`/`go vet`/`go test` to work at all. Skip
# quietly if apt is unavailable or still blocked -- reaching
# archive.ubuntu.com/security.ubuntu.com is a network policy setting outside
# this hook's control.
if command -v apt-get >/dev/null 2>&1; then
  if ! dpkg -s libsdl2-dev >/dev/null 2>&1 || ! dpkg -s libsdl2-image-dev >/dev/null 2>&1 || \
     ! dpkg -s libsdl2-ttf-dev >/dev/null 2>&1 || ! dpkg -s libsdl2-mixer-dev >/dev/null 2>&1; then
    echo "Installing SDL2 dev packages..."
    (apt-get update -qq && apt-get install -y --no-install-recommends \
      libsdl2-dev libsdl2-image-dev libsdl2-ttf-dev libsdl2-mixer-dev pkg-config) \
      || echo "warning: could not install SDL2 dev packages -- check that archive.ubuntu.com and security.ubuntu.com are allowed by this environment's network policy" >&2
  fi
fi
