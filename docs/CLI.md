# CLI do Meufoco

## Instalação rápida (Linux/macOS)

```bash
curl -fsSL https://raw.githubusercontent.com/widia-io/widia-omni/main/scripts/install-cli.sh | bash
```

Para repositório privado (exige autenticação), use um token com leitura do repositório:

```bash
export WIDIA_CLI_GITHUB_TOKEN=<token>
curl -H "Authorization: token ${WIDIA_CLI_GITHUB_TOKEN}" -fsSL https://raw.githubusercontent.com/widia-io/widia-omni/main/scripts/install-cli.sh | bash
```

Esse comando:
- detecta seu sistema (Linux/macOS, amd64/arm64),
- consulta a release mais recente do GitHub,
- baixa o binário correto e instala em `/usr/local/bin/widia`,
- valida exibindo a versão do CLI instalado.

### Variáveis de ambiente

- `WIDIA_CLI_VERSION`: define a versão (ex.: `v1.0.0`).
- `WIDIA_CLI_INSTALL_DIR`: altera o diretório de instalação (padrão `/usr/local/bin`).
- `WIDIA_CLI_GITHUB_TOKEN`: token para repositórios privados.

## Instalação via package manager

### Homebrew (macOS)

Com o tap publicado, instale via Homebrew:

```bash
brew tap widia-io/tap https://github.com/widia-io/homebrew-tap
brew install --cask widia
```

Se já tiver o tap configurado localmente:

```bash
brew install --cask widia
```

### Scoop (Windows)

Após o bucket estar disponível, instale com:

```bash
scoop bucket add widia https://github.com/widia-io/scoop-bucket.git
scoop install widia
```

Se o bucket já existir, atualize antes do install:

```bash
scoop update widia
scoop install widia
```

## Configuração para publicar em package managers

A publicação no GoReleaser já está configurada em `/.goreleaser.yml` nos blocos `homebrew_casks` e `scoops`.

Para funcionamento automático no GitHub Actions, adicione os segredos do repositório:

- `HOMEBREW_TAP_GITHUB_TOKEN`: PAT com permissão para gravar no repositório `widia-io/homebrew-tap`.
- `SCOOP_BUCKET_GITHUB_TOKEN`: PAT com permissão para gravar no repositório `widia-io/scoop-bucket`.

Mais detalhes de bootstrap e fluxo de publicação dos repositórios ficam em: `docs/package-managers/README.md`.

Se os segredos não estiverem configurados, a release ainda publica no GitHub, mas a etapa de publicação dos package managers pode pular/falhar.

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
- `widia-cli_1.0.0_linux_amd64.tar.gz`
- `widia-cli_1.0.0_darwin_arm64.tar.gz`
- `widia-cli_1.0.0_windows_amd64.zip`

## Publicação de release (desenvolvedor)

1. Criar tag semântica: `v0.1.0`
2. Enviar a tag para GitHub (`git push --tags`)
3. Workflow `release-cli` compila automaticamente o CLI para todas as plataformas e publica no GitHub Releases.

Também existe atalho local:

```bash
make release-cli-snapshot   # gera artifacts locais sem publicar
```
