# Publicação em Package Managers

Este projeto publica o CLI `widia` automaticamente para:

- Homebrew Cask (`homebrew_casks`) -> `widia-io/homebrew-tap`
- Scoop (`scoops`) -> `widia-io/scoop-bucket`

A publicação é feita no workflow `Release CLI` via GoReleaser.

## Pré-requisitos nos repositórios de distribuição

1. Criar os repositórios públicos:
   - `github.com/widia-io/homebrew-tap`
   - `github.com/widia-io/scoop-bucket`

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
  - Arquivo gerado por release: `Casks/widia.rb`

- `scoop-bucket`:
  - Branch padrão: `main`
  - Arquivo gerado por release: `widia.json`

## Como a publicação funciona

1. No push de tag (ou `workflow_dispatch`), workflow `release-cli` roda o GoReleaser.
2. O GoReleaser gera:
   - archives multiplataforma no `dist/`
   - `Casks/widia.rb`
   - `scoop/widia.json`
3. Os artefatos são publicados no GitHub Releases.
4. O GoReleaser faz commit dos manifests nos repositórios configurados.

## Verificação pós-release

- Testar instalação Homebrew:

```bash
brew tap widia-io/tap https://github.com/widia-io/homebrew-tap
brew install --cask widia
```

- Testar instalação Scoop:

```bash
scoop bucket add widia https://github.com/widia-io/scoop-bucket.git
scoop install widia
scoop update widia
```

- Verificar que o binário bate versão:

```bash
widia version
```
