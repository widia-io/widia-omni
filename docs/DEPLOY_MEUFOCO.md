# Deploy de Produção — meufoco.app

## Visão Geral

O deploy de produção do `meufoco.app` é automatizado via GitHub Actions em todo merge/push para `main`.
Fluxo:

1. Executa testes/build (`go` + `web`).
2. Publica imagens Docker privadas no GHCR.
3. Sincroniza manifests versionados para `/opt/stacks/meufoco`.
4. Roda migrations do banco.
5. Aplica `docker compose up -d`.
6. Valida health-checks.
7. Em caso de falha, faz rollback automático para `.last_successful_tag`.

## Arquivos Versionados

Arquivos de deploy agora fazem parte do repositório:

1. `/Users/bruno/Developer/widia/widia-omni/docker-compose.prod.yml`
2. `/Users/bruno/Developer/widia/widia-omni/deploy/web.Dockerfile`
3. `/Users/bruno/Developer/widia/widia-omni/deploy/nginx.conf`
4. `/Users/bruno/Developer/widia/widia-omni/.github/workflows/deploy-meufoco-prod.yml`
5. `/Users/bruno/Developer/widia/widia-omni/.github/workflows/rollback-meufoco-prod.yml`

## Contrato de Imagens

Imagens utilizadas no ambiente de produção:

1. `ghcr.io/widia-io/meufoco-app:${IMAGE_TAG}` (API + Worker)
2. `ghcr.io/widia-io/meufoco-web:${IMAGE_TAG}` (Web)

Fallback local no Compose: `${IMAGE_TAG:-main}`.

## GitHub Environment `production`

Criar `Environment` chamado `production` no repositório `widia-io/widia-omni`.

### Secrets obrigatórios

1. `MEUFOCO_PROD_HOST` (ex.: `82.112.245.152`)
2. `MEUFOCO_PROD_PORT` (ex.: `22`)
3. `MEUFOCO_PROD_USER` (ex.: `deploy`)
4. `MEUFOCO_PROD_SSH_KEY` (chave privada do usuário `deploy`)
5. `MEUFOCO_PROD_KNOWN_HOSTS` (saída de `ssh-keyscan -H 82.112.245.152`)

### Vars obrigatórias

1. `MEUFOCO_STACK_DIR` (ex.: `/opt/stacks/meufoco`)
2. `MEUFOCO_API_URL` (`https://api.meufoco.app`)
3. `MEUFOCO_SUPABASE_URL` (`https://supabase.meufoco.app`)

## Preparação One-Time do Servidor

Executar uma vez no host (como root):

```bash
adduser --disabled-password --gecos "" deploy
usermod -aG docker deploy

install -d -m 700 -o deploy -g deploy /home/deploy/.ssh
install -d -m 755 -o deploy -g deploy /opt/stacks/meufoco

# Ajuste permissões dos arquivos versionados do stack
chown -R deploy:deploy /opt/stacks/meufoco/deploy /opt/stacks/meufoco/sql
chown deploy:deploy /opt/stacks/meufoco/docker-compose.prod.yml

# .env.production fica fora do Git e somente leitura para deploy
chown root:deploy /opt/stacks/meufoco/.env.production
chmod 640 /opt/stacks/meufoco/.env.production
```

Adicionar chave pública do CI:

```bash
cat >> /home/deploy/.ssh/authorized_keys
chmod 600 /home/deploy/.ssh/authorized_keys
chown deploy:deploy /home/deploy/.ssh/authorized_keys
```

## Como o Deploy Funciona

Workflow: `/Users/bruno/Developer/widia/widia-omni/.github/workflows/deploy-meufoco-prod.yml`

1. Trigger automático em `push` para `main`.
2. `concurrency` evita corrida entre deploys.
3. Publica imagens com tags:
   - `${github.sha}`
   - `main`
4. Sincroniza para o host:
   - `docker-compose.prod.yml`
   - `deploy/`
   - `sql/migrations/`
5. Cria `.image-tag.env` com `IMAGE_TAG=<sha>`.
6. Executa migrations com `migrate/migrate`.
7. Sobe serviços `web`, `api`, `worker`, `redis`.
8. Valida:
   - `https://api.meufoco.app/health`
   - `https://meufoco.app`
   - health dos containers `meufoco-api` e `meufoco-web`
9. Sucesso: atualiza `/opt/stacks/meufoco/.last_successful_tag`.
10. Falha: rollback automático para última tag saudável.

## Rollback Manual

Workflow: `/Users/bruno/Developer/widia/widia-omni/.github/workflows/rollback-meufoco-prod.yml`

1. Executar `workflow_dispatch`.
2. Informar `image_tag` (SHA já publicada no GHCR).
3. Workflow faz:
   - pull das imagens da tag
   - `up -d` em `web/api/worker`
   - valida health
   - atualiza `.last_successful_tag`

## Runbook de Diagnóstico

No servidor:

```bash
cd /opt/stacks/meufoco
cat .last_successful_tag
cat .image-tag.env
docker compose --env-file .env.production --env-file .image-tag.env -f docker-compose.prod.yml ps
docker logs meufoco-api --tail=200
docker logs meufoco-web --tail=200
curl -fsS https://api.meufoco.app/health
curl -fsS https://meufoco.app
```

Inspeção de imagem em execução:

```bash
docker inspect --format='{{.Config.Image}}' meufoco-api
docker inspect --format='{{.Config.Image}}' meufoco-web
```

## Limites e Escopo

1. Este pipeline não deploya o stack `supabase-meufoco`.
2. `.env.production` não é alterado pelo workflow.
3. Rollback automático depende de `.last_successful_tag` existente.
