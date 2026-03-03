# CLI do Meufoco

## Instalação rápida (Linux/macOS)

```bash
curl -fsSL https://raw.githubusercontent.com/widia-io/widia-omni/main/scripts/install-cli.sh | bash
```

Esse comando:
- detecta seu sistema (Linux/macOS, amd64/arm64),
- consulta a release mais recente do GitHub,
- baixa o binário correto e instala em `/usr/local/bin/widia`,
- valida exibindo a versão do CLI instalado.

### Variáveis de ambiente

- `WIDIA_CLI_VERSION`: define a versão (ex.: `v1.0.0`).
- `WIDIA_CLI_INSTALL_DIR`: altera o diretório de instalação (padrão `/usr/local/bin`).

## Instalação via package manager

### Homebrew (macOS/Linux)

Após configurar o tap publicado pelo release, os usuários podem instalar com:

```bash
brew tap widia-io/tap
brew install widia
```

### Scoop (Windows)

Após publicar no bucket do Scoop, os usuários podem instalar com:

```bash
scoop bucket add widia https://github.com/widia-io/scoop-bucket.git
scoop install widia
```

## Configuração para publicar em package managers

No `goreleaser`, a publicação para Homebrew e Scoop usa estes tokens/segredos:

- `HOMEBREW_TAP_GITHUB_TOKEN`: PAT com permissão para gravar no repositório `widia-io/homebrew-tap`.
- `SCOOP_BUCKET_GITHUB_TOKEN`: PAT com permissão para gravar no repositório `widia-io/scoop-bucket`.

Se não configurar os dois segredos, a etapa de release ainda publica no GitHub, mas pode pular/ falhar a publicação desses fórmulas.

## Instalacao alternativa via Go

```bash
go install github.com/widia-io/widia-omni/cmd/cli@latest
```

> Esse caminho é útil para desenvolvedores Go e não depende de GitHub Releases.

## Estrutura de distribuição

O CLI é empacotado no formato:
- `widia-cli_<versao>_<os>_<arch>.tar.gz` para Linux e macOS,
- `widia-cli_<versao>_windows_<arch>.zip` para Windows.

Exemplos:
- `widia-cli_v1.0.0_linux_amd64.tar.gz`
- `widia-cli_v1.0.0_darwin_arm64.tar.gz`
- `widia-cli_v1.0.0_windows_amd64.zip`

## Publicação de release (desenvolvedor)

1. Criar tag semântica: `v0.1.0`
2. Enviar a tag para GitHub (`git push --tags`)
3. Workflow `release-cli` compila automaticamente o CLI para todas as plataformas e publica no GitHub Releases.

Também existe atalho local:

```bash
make release-cli-snapshot   # gera artifacts locais sem publicar
```
