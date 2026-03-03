# Publicação em Package Managers

Este projeto publica o CLI `meufoco` automaticamente para:

- Homebrew Cask (`homebrew_casks`) -> `widia-io/homebrew-tap`
- Scoop (`scoops`) -> `widia-io/scoop-bucket`

A publicação é feita no workflow `Release CLI` via GoReleaser.

## Pré-requisitos nos repositórios de distribuição

1. Criar os repositórios públicos:
   - `github.com/widia-io/homebrew-tap`
   - `github.com/widia-io/scoop-bucket`

Observacao:

- O Homebrew/Scoop consumem `url` dos artefatos de `releases` sem autenticação.
  Se o repositório `widia-io/widia-omni` for privado, a instalação por package manager não
  funcionará para usuários sem token do GitHub. Nessa situação, mantenha um repositório de releases público ou use o instalador com token.

2. Garantir permissões:
   - Token PAT com `contents: write` para cada repositório.

3. Adicionar segredos no repositório principal (`widia-io/widia-omni`):
   - `HOMEBREW_TAP_GITHUB_TOKEN`
   - `SCOOP_BUCKET_GITHUB_TOKEN`

## Estrutura esperada dos repositórios

A criação pode ser manual; o GoReleaser cria/atualiza os manifests quando a release roda.

- `homebrew-tap`:
  - Branch padrão: `main`
  - Pasta gerada: `Casks/`
  - Arquivo gerado por release: `Casks/meufoco.rb`

- `scoop-bucket`:
  - Branch padrão: `main`
  - Arquivo gerado por release: `meufoco.json`

## Como a publicação funciona

1. No push de tag (ou `workflow_dispatch`), workflow `release-cli` roda o GoReleaser.
2. O GoReleaser gera:
   - archives multiplataforma no `dist/`
  - `Casks/meufoco.rb`
  - `scoop/meufoco.json`
3. Os artefatos são publicados no GitHub Releases.
4. O GoReleaser faz commit dos manifests nos repositórios configurados.

## Verificação pós-release

- Testar instalação Homebrew:

```bash
brew tap widia-io/tap https://github.com/widia-io/homebrew-tap
brew install --cask meufoco
```

- Testar instalação Scoop:

```bash
scoop bucket add meufoco https://github.com/widia-io/scoop-bucket.git
scoop install meufoco
scoop update meufoco
```

- Verificar que o binário bate versão:

```bash
meufoco version
```

## Script de bootstrap/validação

Existe um script para automatizar checagem de repositórios e secrets:

```bash
scripts/bootstrap-package-managers.sh
```

Comportamento padrão:
- valida se os repositórios `homebrew-tap` e `scoop-bucket` existem,
- valida se os secrets `HOMEBREW_TAP_GITHUB_TOKEN` e `SCOOP_BUCKET_GITHUB_TOKEN` existem no repositório fonte.

Opções úteis:

```bash
scripts/bootstrap-package-managers.sh --create
```

- `--create`: cria os repositórios de package managers faltantes e adiciona arquivos mínimos de bootstrap.

```bash
scripts/bootstrap-package-managers.sh --create --check-token-access --strict
```

- `--check-token-access`: valida acesso real dos tokens (via variável de ambiente) aos repositórios.
- `--strict`: falha se `HOMEBREW_TAP_GITHUB_TOKEN` ou `SCOOP_BUCKET_GITHUB_TOKEN` não estiverem exportados.
- `--owner`: altera o dono dos repositórios (padrão `widia-io`).
- `--homebrew-repo` / `--scoop-repo`: sobrescreve cada repositório manualmente.

Exemplo completo local (sem criar repositórios):

```bash
export HOMEBREW_TAP_GITHUB_TOKEN=***
export SCOOP_BUCKET_GITHUB_TOKEN=***
scripts/bootstrap-package-managers.sh --check-token-access --strict
```
