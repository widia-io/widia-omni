#!/usr/bin/env bash
set -euo pipefail

REPO="${WIDIA_CLI_REPO:-widia-io/widia-omni}"
VERSION="${WIDIA_CLI_VERSION:-latest}"
INSTALL_DIR="${WIDIA_CLI_INSTALL_DIR:-/usr/local/bin}"
CLI_NAME="widia"
GITHUB_TOKEN="${WIDIA_CLI_GITHUB_TOKEN:-${GITHUB_TOKEN:-${GH_TOKEN:-${GITHUB_ACCESS_TOKEN:-}}}}"

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
  ARCHIVE_EXT="zip"
else
  ARCHIVE_EXT="tar.gz"
fi

tmpdir="$(mktemp -d)"
trap 'rm -rf "$tmpdir"' EXIT

http_download() {
  local url="$1"
  local output="$2"
  local accept_header="${3:-}"
  local token_header=()

  if [ "$TRANSFER" = "curl" ]; then
    if [ -n "$GITHUB_TOKEN" ]; then
      token_header=(-H "Authorization: Bearer ${GITHUB_TOKEN}")
    fi

    if [ -n "$accept_header" ]; then
      curl -fsSL ${token_header[@]+"${token_header[@]}"} -H "Accept: ${accept_header}" -o "$output" "$url"
    else
      curl -fsSL ${token_header[@]+"${token_header[@]}"} -o "$output" "$url"
    fi
  else
    if [ -n "$GITHUB_TOKEN" ]; then
      token_header=(--header="Authorization: Bearer ${GITHUB_TOKEN}")
    fi

    if [ -n "$accept_header" ]; then
      wget -q ${token_header[@]+"${token_header[@]}"} --header="Accept: ${accept_header}" -O "$output" "$url"
    else
      wget -q ${token_header[@]+"${token_header[@]}"} -O "$output" "$url"
    fi
  fi
}

find_asset_api_url() {
  local asset_name="$1"
  local asset_line
  local start_line
  local api_url=""

  asset_line="$(grep -nF "\"name\": \"${asset_name}\"" "$tmpdir/release.json" | head -n1 | cut -d: -f1 || true)"
  if [ -z "$asset_line" ]; then
    return 1
  fi

  if [ "$asset_line" -gt 30 ]; then
    start_line=$((asset_line - 30))
  else
    start_line=1
  fi

  api_url="$(sed -n "${start_line},${asset_line}p" "$tmpdir/release.json" \
    | grep -oE '\"url\"[[:space:]]*:[[:space:]]*\"https://api\.github\.com/repos/[^\"]*/releases/assets/[0-9]+\"' \
    | sed -E 's/^\"url\"[[:space:]]*:[[:space:]]*\"(.*)\"$/\1/' \
    | tail -n 1)"

  if [ -z "$api_url" ]; then
    return 1
  fi

  echo "$api_url"
}

if ! http_download "$API_URL" "$tmpdir/release.json"; then
  echo "Nao foi possivel consultar a release em ${API_URL}" >&2
  echo "Verifique se o repositório ${REPO} e visivel publicamente ou se o token foi informado." >&2
  exit 1
fi

RELEASE_TAG="$(sed -n 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^\"]*\)".*/\1/p' "$tmpdir/release.json" | head -n1)"
ASSET_NAME="$(sed -n 's/.*"name"[[:space:]]*:[[:space:]]*"\([^\"]*\)".*/\1/p' "$tmpdir/release.json" | grep -E "${CLI_NAME}-cli_.*_${OS}_${ARCH}\\.${ARCHIVE_EXT}$" | head -n1)"
ASSET_URL="$(sed -n 's/.*"browser_download_url"[[:space:]]*:[[:space:]]*"\([^\"]*\)".*/\1/p' "$tmpdir/release.json" | grep -F "$ASSET_NAME" | head -n1)"
ASSET_API_URL="$(find_asset_api_url "$ASSET_NAME")"

if [ -z "$RELEASE_TAG" ]; then
  echo "Nao foi possivel determinar a versao da release" >&2
  exit 1
fi

if [ "$VERSION" = "latest" ]; then
  VERSION="$RELEASE_TAG"
fi

if [ -z "$ASSET_NAME" ]; then
  echo "Nao foi possivel localizar um binario para ${OS}/${ARCH} na release ${RELEASE_TAG}" >&2
  echo "Arquivos disponiveis:" >&2
  sed -n 's/.*"name"[[:space:]]*:[[:space:]]*"\([^\"]*\)".*/\1/p' "$tmpdir/release.json" >&2
  exit 1
fi

if [ -z "$ASSET_URL" ]; then
  echo "Asset para ${ASSET_NAME} nao encontrado no JSON da release." >&2
  echo "Asset nao encontrado para ${OS}/${ARCH} na versao ${VERSION}" >&2
  exit 1
fi

if [ -n "$GITHUB_TOKEN" ] && [ -n "$ASSET_API_URL" ]; then
  FILENAME="${ASSET_NAME}"
  echo "Baixando ${CLI_NAME} ${VERSION} para ${OS}/${ARCH} (autenticado)"
  ARCHIVE_PATH="${tmpdir}/${FILENAME}"
  http_download "$ASSET_API_URL" "$ARCHIVE_PATH" "application/octet-stream"
else
  FILENAME="${ASSET_NAME}"
  echo "Baixando ${CLI_NAME} ${VERSION} para ${OS}/${ARCH}"
  ARCHIVE_PATH="${tmpdir}/${FILENAME}"
  http_download "$ASSET_URL" "$ARCHIVE_PATH"
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
