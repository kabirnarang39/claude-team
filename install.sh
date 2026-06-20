#!/usr/bin/env bash
set -euo pipefail

REPO="https://github.com/<org>/claude-team"
VERSION="v1.0.0"
INSTALL_DIR="$HOME/.local/bin"
SKILL_DIR="$HOME/.claude/skills"

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

if [[ -z "${ANTHROPIC_API_KEY:-}" ]]; then
  echo "WARN: ANTHROPIC_API_KEY is not set — agents will fail to run."
  echo "      Set it with: export ANTHROPIC_API_KEY=sk-ant-..."
fi

# ── Download binary ──────────────────────────────────────────────────────────
echo "Installing Anton $VERSION for $PLATFORM..."
mkdir -p "$INSTALL_DIR"
BINARY_URL="$REPO/releases/download/$VERSION/anton-$PLATFORM"
curl -fsSL "$BINARY_URL" -o "$INSTALL_DIR/anton"
chmod +x "$INSTALL_DIR/anton"

# ── Install skills ───────────────────────────────────────────────────────────
echo "Installing Anton skills..."
mkdir -p "$SKILL_DIR"
SKILLS_URL="$REPO/releases/download/$VERSION/skills.tar.gz"
curl -fsSL "$SKILLS_URL" | tar -xz -C "$SKILL_DIR"

# ── Install MCP server ───────────────────────────────────────────────────────
echo "Installing MCP server dependencies..."
MCP_DIR="$HOME/.claude/anton-mcp"
mkdir -p "$MCP_DIR"
MCP_URL="$REPO/releases/download/$VERSION/mcp.tar.gz"
curl -fsSL "$MCP_URL" | tar -xz -C "$MCP_DIR"
(cd "$MCP_DIR" && npm install --silent)

# ── PATH check ───────────────────────────────────────────────────────────────
if ! command -v anton &>/dev/null; then
  echo ""
  echo "NOTE: Add $INSTALL_DIR to your PATH:"
  echo "  echo 'export PATH=\"\$HOME/.local/bin:\$PATH\"' >> ~/.zshrc && source ~/.zshrc"
fi

echo ""
echo "Anton $VERSION installed successfully."
echo "Start the dashboard: anton"
echo "Then in Claude Code: /team-dispatch <your task>"
