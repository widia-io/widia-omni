#!/usr/bin/env bash
set -euo pipefail

REPO="${WIDIA_CLI_REPO:-widia-io/widia-omni}"
VERSION="${WIDIA_CLI_VERSION:-latest}"
INSTALL_DIR="${WIDIA_CLI_INSTALL_DIR:-/usr/local/bin}"
CLI_NAME="widia"

if [ -z "$CLI_NAME" ]; then
  echo "Nome do CLI invalido" >&2
  exit 1
fi

if command -v curl >/dev/null 2>&1; then
  TRANSFER="curl"
elif command -v wget >/dev/null 2>&1; then
  TRANSFER="wget"
else
  echo "curl ou wget necessario para baixar o instalador" >&2
  exit 1
fi

if [ "$VERSION" = "latest" ]; then
  API_URL="https://api.github.com/repos/${REPO}/releases/latest"
else
  API_URL="https://api.github.com/repos/${REPO}/releases/tags/${VERSION}"
fi

case "$(uname -s)" in
  Linux*) OS="linux" ;;
  Darwin*) OS="darwin" ;;
  *)
    echo "Sistema operacional não suportado: $(uname -s)" >&2
    exit 1
    ;;
esac

case "$(uname -m)" in
  x86_64) ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
  *)
    echo "Arquitetura não suportada: $(uname -m)" >&2
    exit 1
    ;;
esac

if [ "$OS" = "windows" ]; then
  FILENAME="${CLI_NAME}-cli_${VERSION}_${OS}_${ARCH}.zip"
else
  FILENAME="${CLI_NAME}-cli_${VERSION}_${OS}_${ARCH}.tar.gz"
fi

tmpdir="$(mktemp -d)"
trap 'rm -rf "$tmpdir"' EXIT

if [ "$TRANSFER" = "curl" ]; then
  curl -fsSL "$API_URL" > "$tmpdir/release.json"
else
  wget -q -O "$tmpdir/release.json" "$API_URL"
fi

RELEASE_TAG="$(sed -n 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' "$tmpdir/release.json" | head -n1)"
ASSET_URL="$(sed -n 's/.*"browser_download_url"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' "$tmpdir/release.json" | grep -F "$FILENAME" | head -n1)"

if [ -z "$RELEASE_TAG" ]; then
  echo "Nao foi possivel determinar a versao da release" >&2
  exit 1
fi

if [ "$VERSION" = "latest" ]; then
  VERSION="$RELEASE_TAG"
  if [ "$OS" = "windows" ]; then
    FILENAME="${CLI_NAME}-cli_${VERSION}_${OS}_${ARCH}.zip"
  else
    FILENAME="${CLI_NAME}-cli_${VERSION}_${OS}_${ARCH}.tar.gz"
  fi
  ASSET_URL="$(sed -n 's/.*"browser_download_url"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' "$tmpdir/release.json" | grep -F "$FILENAME" | head -n1)"
fi

if [ -z "$ASSET_URL" ]; then
  echo "Asset nao encontrado para ${OS}/${ARCH} na versao ${VERSION}" >&2
  echo "Arquivos disponiveis:" >&2
  sed -n 's/.*"browser_download_url"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' "$tmpdir/release.json" >&2
  exit 1
fi

echo "Baixando ${CLI_NAME} ${VERSION} para ${OS}/${ARCH}"
ARCHIVE_PATH="${tmpdir}/${FILENAME}"
if [ "$TRANSFER" = "curl" ]; then
  curl -fsSL -o "$ARCHIVE_PATH" "$ASSET_URL"
else
  wget -q -O "$ARCHIVE_PATH" "$ASSET_URL"
fi

case "$FILENAME" in
  *.zip)
    unzip -q "$ARCHIVE_PATH" -d "$tmpdir"
    ;;
  *)
    tar -xzf "$ARCHIVE_PATH" -C "$tmpdir"
    ;;
esac

BINARY_PATH=""
if [ -f "${tmpdir}/${CLI_NAME}" ]; then
  BINARY_PATH="${tmpdir}/${CLI_NAME}"
elif [ -f "${tmpdir}/${CLI_NAME}.exe" ]; then
  BINARY_PATH="${tmpdir}/${CLI_NAME}.exe"
else
  BINARY_PATH="$(find "$tmpdir" -type f \( -name "${CLI_NAME}" -o -name "${CLI_NAME}.exe" \) | head -n1)"
fi

if [ -z "$BINARY_PATH" ]; then
  echo "Binario nao encontrado no pacote baixado" >&2
  exit 1
fi

install_path="${INSTALL_DIR%/}/${CLI_NAME}"
mkdir -p "$INSTALL_DIR"
cp "$BINARY_PATH" "$install_path"
chmod +x "$install_path"

echo "Instalado em: $install_path"
"$install_path" version
