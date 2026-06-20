#!/usr/bin/env bash
set -euo pipefail

REPO="https://github.com/kabirnarang39/claude-team"
INSTALL_DIR="$HOME/.local/bin"

# ── Fetch latest release ──────────────────────────────────────────────────────
VERSION="$(curl -fsSL "https://api.github.com/repos/kabirnarang39/claude-team/releases/latest" \
  | grep -o '"tag_name": "[^"]*"' | grep -o 'v[^"]*')"
if [ -z "$VERSION" ]; then
  echo "ERROR: Could not fetch latest release from GitHub. Check your internet connection."
  exit 1
fi
SKILL_DIR="$HOME/.claude/skills"
MCP_DIR="$HOME/.claude/anton-mcp"

# ── Detect platform ──────────────────────────────────────────────────────────
OS="$(uname -s)"
ARCH="$(uname -m)"
case "$OS-$ARCH" in
  Darwin-arm64)  PLATFORM="darwin-arm64"  ;;
  Darwin-x86_64) PLATFORM="darwin-amd64"  ;;
  Linux-x86_64)  PLATFORM="linux-amd64"   ;;
  *)
    echo "ERROR: Unsupported platform $OS-$ARCH"
    echo "Supported: macOS arm64/amd64, Linux amd64"
    exit 1
    ;;
esac

# ── Check prerequisites ──────────────────────────────────────────────────────
check() {
  command -v "$1" &>/dev/null || {
    echo "ERROR: $1 is required but not installed."
    echo "Install: $2"
    exit 1
  }
}
check claude  "https://claude.ai/download"
check node    "https://nodejs.org"
check npm     "https://nodejs.org"

# ── Enforce Node >= 20 ────────────────────────────────────────────────────────
if ! node -e "if(+process.version.slice(1).split('.')[0]<20)process.exit(1)" 2>/dev/null; then
  echo "ERROR: Node.js 20+ required. Current version: $(node --version)"
  echo "Install the latest LTS: https://nodejs.org"
  exit 1
fi

# ── Download binary ──────────────────────────────────────────────────────────
echo "Installing Anton $VERSION for $PLATFORM..."
mkdir -p "$INSTALL_DIR"
BINARY_URL="$REPO/releases/download/$VERSION/anton-$PLATFORM"
curl -fsSL "$BINARY_URL" -o "$INSTALL_DIR/anton"
chmod +x "$INSTALL_DIR/anton"
# Remove macOS provenance/quarantine attributes that block execution on macOS 26+
# Must clear both the binary and its parent directory
xattr -c "$INSTALL_DIR/anton" 2>/dev/null || true
xattr -c "$INSTALL_DIR" 2>/dev/null || true

# ── Extract skills and MCP from source archive ───────────────────────────────
echo "Installing Anton skills and MCP server..."
TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

SOURCE_URL="$REPO/archive/refs/tags/$VERSION.tar.gz"
curl -fsSL "$SOURCE_URL" | tar -xz -C "$TMP_DIR" --strip-components=1

# Install skills
mkdir -p "$SKILL_DIR"
cp "$TMP_DIR/skills/"*.md "$SKILL_DIR/"

# Install MCP coordinator
rm -rf "$MCP_DIR"
cp -r "$TMP_DIR/mcp" "$MCP_DIR"
echo "Installing MCP dependencies..."
(cd "$MCP_DIR" && npm install --silent)

# Install coordinator logic, workflows, and roles to global location
ANTON_DIR="$HOME/.claude/anton"
rm -rf "$ANTON_DIR"
mkdir -p "$ANTON_DIR"
cp -r "$TMP_DIR/coordinators" "$ANTON_DIR/"
cp -r "$TMP_DIR/workflows" "$ANTON_DIR/"
cp -r "$TMP_DIR/roles" "$ANTON_DIR/"

# ── PATH check ───────────────────────────────────────────────────────────────
if ! command -v anton &>/dev/null; then
  echo ""
  echo "NOTE: Add $INSTALL_DIR to your PATH:"
  case "$SHELL" in
    */zsh)  echo "  echo 'export PATH=\"\$HOME/.local/bin:\$PATH\"' >> ~/.zshrc && source ~/.zshrc" ;;
    */bash) echo "  echo 'export PATH=\"\$HOME/.local/bin:\$PATH\"' >> ~/.bashrc && source ~/.bashrc" ;;
    *)      echo "  export PATH=\"\$HOME/.local/bin:\$PATH\"" ;;
  esac
fi

echo ""
echo "────────────────────────────────────────────────────────"
echo "  Anton $VERSION installed successfully."
echo "────────────────────────────────────────────────────────"
echo ""
echo "  1. Go to ANY project directory:"
echo "       cd ~/my-project"
echo ""
echo "  2. Start the Anton dashboard:"
echo "       anton"
echo "     (First run auto-registers the MCP coordinator"
echo "      in .claude/settings.json for this project.)"
echo ""
echo "  3. Open Claude Code in the same directory:"
echo "       claude"
echo ""
echo "  4. Dispatch a task:"
echo "       /team-dispatch build user auth with JWT tokens"
echo ""
echo "  Run 'anton --check' at any time to verify setup."
echo "────────────────────────────────────────────────────────"
