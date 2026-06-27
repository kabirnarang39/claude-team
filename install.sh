#!/usr/bin/env bash
set -euo pipefail

REPO="https://github.com/kabirnarang39/claude-team"
API="https://api.github.com/repos/kabirnarang39/claude-team"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"
SKILL_DIR="$HOME/.claude/skills"
MCP_DIR="$HOME/.claude/anton-mcp"
ANTON_DIR="$HOME/.claude/anton"

die() {
  echo "ERROR: $*" >&2
  exit 1
}

need() {
  command -v "$1" >/dev/null 2>&1 || die "$1 is required. Install: $2"
}

uninstall() {
  echo "Removing Anton files..."
  rm -f "$INSTALL_DIR/anton"
  rm -rf "$MCP_DIR" "$ANTON_DIR"
  rm -rf \
    "$SKILL_DIR/team-dispatch" \
    "$SKILL_DIR/team-resume" \
    "$SKILL_DIR/team-status" \
    "$SKILL_DIR/team-stop"
  echo "Anton removed. Project-local .claude/settings.json and .claude-team/ run data are left untouched."
}

if [ "${1:-}" = "--uninstall" ]; then
  uninstall
  exit 0
fi

need curl "https://curl.se"
need tar "system package manager"
need claude "https://claude.ai/download"
need node "https://nodejs.org"
need npm "https://nodejs.org"

if ! node -e "if (+process.version.slice(1).split('.')[0] < 20) process.exit(1)" >/dev/null 2>&1; then
  die "Node.js 20+ required. Current version: $(node --version)"
fi

VERSION="$(
  curl -fsSL "$API/releases/latest" |
    sed -n 's/.*"tag_name": "\(v[^"]*\)".*/\1/p' |
    head -n 1
)"
[ -n "$VERSION" ] || die "Could not fetch latest release from GitHub."

OS="$(uname -s)"
ARCH="$(uname -m)"
case "$OS-$ARCH" in
  Darwin-arm64)  PLATFORM="darwin-arm64" ;;
  Darwin-x86_64) PLATFORM="darwin-amd64" ;;
  Linux-x86_64)  PLATFORM="linux-amd64" ;;
  *)
    die "Unsupported platform $OS-$ARCH. Supported: macOS arm64/amd64, Linux amd64."
    ;;
esac

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

echo "Installing Anton $VERSION for $PLATFORM..."
mkdir -p "$INSTALL_DIR"

BINARY_URL="$REPO/releases/download/$VERSION/anton-$PLATFORM"
curl -fsSL "$BINARY_URL" -o "$INSTALL_DIR/anton"
chmod +x "$INSTALL_DIR/anton"

CHECKSUMS="$TMP_DIR/checksums.txt"
if curl -fsSL "$REPO/releases/download/$VERSION/checksums.txt" -o "$CHECKSUMS"; then
  expected_line="$(grep "anton-$PLATFORM$" "$CHECKSUMS" || true)"
  if [ -n "$expected_line" ] && command -v shasum >/dev/null 2>&1; then
    expected="$(printf '%s\n' "$expected_line" | awk '{print $1}')"
    actual="$(shasum -a 256 "$INSTALL_DIR/anton" | awk '{print $1}')"
    [ "$expected" = "$actual" ] || die "Checksum verification failed for anton-$PLATFORM."
    echo "Checksum verified."
  else
    echo "Checksum file downloaded, but no matching checksum verifier was available. Continuing."
  fi
else
  echo "Warning: checksums.txt unavailable for $VERSION. Continuing without checksum verification."
fi

if command -v xattr >/dev/null 2>&1; then
  xattr -c "$INSTALL_DIR/anton" 2>/dev/null || true
  xattr -c "$INSTALL_DIR" 2>/dev/null || true
fi
if command -v codesign >/dev/null 2>&1; then
  codesign --sign - "$INSTALL_DIR/anton" 2>/dev/null || true
fi

echo "Installing Anton skills, workflows, roles, and MCP coordinator..."
SOURCE_URL="$REPO/archive/refs/tags/$VERSION.tar.gz"
curl -fsSL "$SOURCE_URL" | tar -xz -C "$TMP_DIR" --strip-components=1

mkdir -p "$SKILL_DIR"
for skill_file in "$TMP_DIR/skills/"*.md; do
  skill_name="$(basename "$skill_file" .md)"
  mkdir -p "$SKILL_DIR/$skill_name"
  cp "$skill_file" "$SKILL_DIR/$skill_name/SKILL.md"
done

rm -rf "$MCP_DIR"
cp -R "$TMP_DIR/mcp" "$MCP_DIR"
echo "Installing MCP coordinator dependencies..."
if [ -f "$MCP_DIR/package-lock.json" ]; then
  (cd "$MCP_DIR" && npm ci --omit=dev --silent)
else
  (cd "$MCP_DIR" && npm install --omit=dev --silent)
fi

rm -rf "$ANTON_DIR"
mkdir -p "$ANTON_DIR"
cp -R "$TMP_DIR/coordinators" "$ANTON_DIR/"
cp -R "$TMP_DIR/workflows" "$ANTON_DIR/"
cp -R "$TMP_DIR/roles" "$ANTON_DIR/"
cp "$TMP_DIR/mcp-registry.yaml" "$ANTON_DIR/mcp-registry.yaml"

if ! command -v anton >/dev/null 2>&1; then
  echo
  echo "Add $INSTALL_DIR to your PATH:"
  case "${SHELL:-}" in
    */zsh)  echo "  echo 'export PATH=\"\$HOME/.local/bin:\$PATH\"' >> ~/.zshrc && source ~/.zshrc" ;;
    */bash) echo "  echo 'export PATH=\"\$HOME/.local/bin:\$PATH\"' >> ~/.bashrc && source ~/.bashrc" ;;
    *)      echo "  export PATH=\"\$HOME/.local/bin:\$PATH\"" ;;
  esac
fi

cat <<EOF

Anton $VERSION installed successfully.

Files written:
  $INSTALL_DIR/anton
  $SKILL_DIR/team-dispatch
  $SKILL_DIR/team-resume
  $SKILL_DIR/team-status
  $SKILL_DIR/team-stop
  $MCP_DIR
  $ANTON_DIR

Next:
  1. cd ~/my-project
  2. anton --check
  3. anton
  4. claude
  5. /team-dispatch build user auth with JWT tokens

Uninstall:
  curl -fsSL https://raw.githubusercontent.com/kabirnarang39/claude-team/main/install.sh | bash -s -- --uninstall
EOF
