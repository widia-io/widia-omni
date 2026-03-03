# PR: [Titulo curto]

## Objetivo
- Descreva em uma frase o que esta mudança entrega.

## Mudanças realizadas
- [ ] Lista as principais alterações:
  - [ ] .goreleaser.yml
  - [ ] scripts/
  - [ ] docs/
  - [ ] Makefile
  - [ ] Outros:

## Contexto de negócio
- Problema que está resolvendo
- Impacto esperado para usuários (CLI)

## Validação
Marque o que foi executado:
- [ ] `go run github.com/goreleaser/goreleaser/v2@latest check`
- [ ] `scripts/bootstrap-package-managers.sh --check-token-access --strict`
- [ ] `make release-cli-snapshot`
- [ ] `go test ./cmd/cli ./internal/cli`

### Resultado
- Comandos acima passaram: ✅ / ❌
- Observações de ambiente / limitações:

## Mudanças de publicação
- [ ] Tokens de package manager adicionados em `widia-io/widia-omni`:
  - [ ] `HOMEBREW_TAP_GITHUB_TOKEN`
  - [ ] `SCOOP_BUCKET_GITHUB_TOKEN`
- [ ] Repositórios de distribuição provisionados:
  - [ ] `widia-io/homebrew-tap`
  - [ ] `widia-io/scoop-bucket`

## Checklist de merge
- [ ] Não introduz regressão em build/instalação local do CLI
- [ ] Documentação atualizada
- [ ] Comportamento em português mantido no TUI/documentação conforme padrão
- [ ] Branch sem conflitos com `main`

## Evidências
Cole prints ou logs importantes (se aplicável).
