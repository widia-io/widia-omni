#!/usr/bin/env bash
set -euo pipefail

OWNER="${WIDIA_PM_OWNER:-widia-io}"
HOMEBREW_TAP_REPO="${WIDIA_HOMEBREW_TAP_REPO:-${OWNER}/homebrew-tap}"
SCOOP_BUCKET_REPO="${WIDIA_SCOOP_BUCKET_REPO:-${OWNER}/scoop-bucket}"
SOURCE_REPO="${WIDIA_SOURCE_REPO:-}"
CREATE_MISSING=false
CHECK_TOKEN_ACCESS=false
STRICT=false
DEFAULT_BRANCH="${WIDIA_PM_DEFAULT_BRANCH:-main}"

usage() {
  cat <<'USAGE'
Usage: scripts/bootstrap-package-managers.sh [options]

Options:
  --create                 Cria repositórios de package manager se não existirem.
  --check-token-access     Valida os tokens dos secrets contra os repositórios de destino.
  --strict                 Falha se algum token não estiver disponível no ambiente.
  --owner <owner>          Sobrescrever dono dos repositórios (default: widia-io).
  --homebrew-repo <repo>   Sobrescrever repositório Homebrew (owner/name).
  --scoop-repo <repo>      Sobrescrever repositório Scoop (owner/name).
  --source-repo <repo>     Sobrescrever repositório fonte (owner/name).
  --help                   Exibe este help.
USAGE
}

failures=0

log_section() {
  echo
  echo "===> $1"
}

command_exists() {
  command -v "$1" >/dev/null 2>&1
}

assert_prereqs() {
  if ! command_exists gh; then
    echo "gh (GitHub CLI) é obrigatório" >&2
    exit 1
  fi

  if ! gh auth status >/dev/null 2>&1; then
    echo "gh não está autenticado. Rode: gh auth login" >&2
    exit 1
  fi
}

normalize_repo() {
  local repo="$1"
  printf '%s' "${repo#https://github.com/}" | sed -E 's#^git@github.com:##; s#\.git$##'
}

repo_exists() {
  local repo="$1"
  gh api "repos/$repo" >/dev/null 2>&1
}

create_repo() {
  local repo="$1"
  local description="$2"

  echo "Criando $repo..."
  if ! gh repo create "$repo" --public --description "$description" >/dev/null; then
    echo "Não foi possível criar $repo" >&2
    failures=$((failures + 1))
    return 1
  fi

  # garante branch principal conhecida
  gh api --method PATCH "repos/$repo" -f default_branch="$DEFAULT_BRANCH" >/dev/null 2>&1 || true
}

repo_file_exists() {
  local repo="$1"
  local path="$2"

  if gh api "repos/$repo/contents/$path" >/dev/null 2>&1; then
    return 0
  fi
  return 1
}

write_repo_file() {
  local repo="$1"
  local path="$2"
  local content="$3"

  if repo_file_exists "$repo" "$path"; then
    return 0
  fi

  local encoded
  encoded="$(printf '%s' "$content" | base64 | tr -d '\n')"
  gh api --method PUT "repos/$repo/contents/$path" \
    --field message="chore: bootstrap package-manager repo" \
    --field branch="$DEFAULT_BRANCH" \
    --field content="$encoded" \
    >/dev/null
}

ensure_repo() {
  local repo="$1"
  local description="$2"
  local bootstrap="$3"

  if repo_exists "$repo"; then
    echo "OK $repo já existe"
    return 0
  fi

  if [[ "$CREATE_MISSING" != true ]]; then
    echo "FALTA: $repo não existe. Use --create para criar."
    failures=$((failures + 1))
    return 1
  fi

  if ! create_repo "$repo" "$description"; then
    return 1
  fi

  if [[ "$bootstrap" == homebrew ]]; then
    write_repo_file "$repo" "Casks/.gitkeep" "# placeholder\n"
    write_repo_file "$repo" "README.md" "# homebrew-tap\n\nRepositório do tap Homebrew para o CLI widia.\n\n\n"
  else
    write_repo_file "$repo" "README.md" "# scoop-bucket\n\nRepositório do bucket Scoop para o CLI widia.\n\n\n"
  fi

  echo "Criado: $repo"
}

secret_exists() {
  local repo="$1"
  local secret_name="$2"

  if gh secret list --repo "$repo" | awk '{print $1}' | grep -qx "$secret_name"; then
    return 0
  fi

  return 1
}

check_token_access() {
  local token_name="$1"
  local target_repo="$2"
  local token="${!token_name-}"

  if [[ -z "$token" ]]; then
    if [[ "$STRICT" == true ]]; then
      echo "FALTA: export $token_name para validação de acesso real."
      failures=$((failures + 1))
    else
      echo "AVISO: $token_name não definida no ambiente; validação de acesso pulada."
    fi
    return 1
  fi

  local can_push
  if ! can_push="$(gh api \
    -H "Authorization: token $token" \
    "repos/$target_repo" --jq '.permissions.push' 2>/dev/null)"; then
    echo "FALHA: token $token_name não possui acesso a $target_repo"
    failures=$((failures + 1))
    return 1
  fi

  if [[ "$can_push" != "true" ]]; then
    echo "FALHA: token $token_name sem permissão de gravação em $target_repo"
    failures=$((failures + 1))
    return 1
  fi

  echo "OK $token_name com acesso para publicar em $target_repo"
  return 0
}

main() {
  while [[ $# -gt 0 ]]; do
    case "$1" in
      --create)
        CREATE_MISSING=true
        ;;
      --check-token-access)
        CHECK_TOKEN_ACCESS=true
        ;;
      --strict)
        STRICT=true
        ;;
      --owner)
        OWNER="$2"
        HOMEBREW_TAP_REPO="${OWNER}/homebrew-tap"
        SCOOP_BUCKET_REPO="${OWNER}/scoop-bucket"
        shift
        ;;
      --homebrew-repo)
        HOMEBREW_TAP_REPO="$2"
        shift
        ;;
      --scoop-repo)
        SCOOP_BUCKET_REPO="$2"
        shift
        ;;
      --source-repo)
        SOURCE_REPO="$2"
        shift
        ;;
      --help)
        usage
        exit 0
        ;;
      *)
        echo "Opção desconhecida: $1" >&2
        usage
        exit 1
        ;;
    esac
    shift
  done

  assert_prereqs

  if [[ -z "$SOURCE_REPO" ]]; then
    if ! git remote get-url origin >/dev/null 2>&1; then
      echo "Não foi possível detectar repositório fonte. Use --source-repo owner/name"
      exit 1
    fi
    SOURCE_REPO="$(normalize_repo "$(git remote get-url origin)")"
  fi

  log_section "Validação de repositórios"
  ensure_repo "$HOMEBREW_TAP_REPO" "Homebrew tap for widia CLI" homebrew
  ensure_repo "$SCOOP_BUCKET_REPO" "Scoop bucket for widia CLI" scoop

  log_section "Validação de secrets no repositório fonte ($SOURCE_REPO)"
  if secret_exists "$SOURCE_REPO" "HOMEBREW_TAP_GITHUB_TOKEN"; then
    echo "OK secret HOMEBREW_TAP_GITHUB_TOKEN existe"
  else
    echo "FALHA: secret HOMEBREW_TAP_GITHUB_TOKEN não encontrado"
    failures=$((failures + 1))
  fi

  if secret_exists "$SOURCE_REPO" "SCOOP_BUCKET_GITHUB_TOKEN"; then
    echo "OK secret SCOOP_BUCKET_GITHUB_TOKEN existe"
  else
    echo "FALHA: secret SCOOP_BUCKET_GITHUB_TOKEN não encontrado"
    failures=$((failures + 1))
  fi

  if [[ "$CHECK_TOKEN_ACCESS" == true ]]; then
    log_section "Validação de acesso dos tokens"
    check_token_access HOMEBREW_TAP_GITHUB_TOKEN "$HOMEBREW_TAP_REPO"
    check_token_access SCOOP_BUCKET_GITHUB_TOKEN "$SCOOP_BUCKET_REPO"
  fi

  if [[ "$failures" -gt 0 ]]; then
    echo
    echo "Resultado: com falhas ($failures)."
    exit 1
  fi

  echo
  echo "Resultado: tudo OK."
}

main "$@"
