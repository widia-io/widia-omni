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
