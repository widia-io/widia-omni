# Plano de Evolucao: Areas, Metas e Projetos

Data: 2026-02-14

## Objetivo

Trazer mais valor para o app conectando melhor planejamento e execucao:

- manter a robustez atual de tarefas;
- destravar gestao de areas (menos "fixas");
- introduzir projetos;
- amarrar projetos, metas e areas de forma clara para onboarding, dashboard e insights.

## Diagnostico atual (baseado no codigo)

### O que ja esta forte

- Tarefas estao robustas em UX e dados (subtarefas, etiquetas, secoes, foco, prioridades, filtros, quick add):
  - `web/src/pages/tasks.tsx`
  - `internal/service/task.go`
  - `sql/migrations/000024_task_powerups.up.sql`

### Gaps que reduzem valor hoje

1. Nao existe entidade de `projetos` no dominio/API/UI.
   - Rotas atuais: areas, goals, habits, tasks, sections etc., sem `/projects`:
     - `internal/router/router.go`
   - Sem modelagem no banco:
     - `sql/migrations/`

2. Onboarding com areas "fixas" na entrada.
   - A tela usa lista predefinida de areas e nao permite criar area customizada no fluxo:
     - `web/src/pages/onboarding.tsx`

3. Onboarding cria metas e habitos sem vinculo obrigatorio com area.
   - `area_id` existe no backend, mas o frontend nao envia no onboarding:
     - `internal/service/onboarding.go`
     - `web/src/pages/onboarding.tsx`
   - Impacto: boa parte do valor analitico por area fica subutilizado.

4. Insights e score dependem de vinculo por area.
   - Queries de insights para metas/habitos usam `JOIN life_areas`, entao itens sem area podem nao entrar:
     - `internal/service/insight.go`
   - Score semanal calcula por `area_id`:
     - `internal/service/score.go`

5. Onboarding completion hoje e muito binario e com obrigatoriedade alta.
   - Fluxo atual pede metas e habitos para concluir, sem caminho leve de ativacao:
     - `web/src/pages/onboarding.tsx`

6. Gestao de areas com risco de desativacao involuntaria na edicao.
   - Backend de update espera `is_active` no payload:
     - `internal/service/area.go`
   - Form atual de areas nao envia `is_active` explicitamente:
     - `web/src/pages/areas.tsx`
   - Impacto: area pode sair de relatorios/dashboard sem o usuario entender o motivo.

## Impacto no onboarding e ativacao

### Hipoteses de impacto (alto)

- Areas fixas no inicio podem gerar baixa identificacao ("nao representa minha vida") e abandono precoce.
- Sem projetos, o usuario sai do onboarding com metas abstratas e cai em tarefas sem ponte de medio prazo.
- Sem amarracao area -> meta -> execucao, dashboard/insights tendem a perder contexto e relevancia.

### Sinais que devemos medir

- Conversao por etapa do onboarding.
- Tempo ate primeira tarefa criada.
- Percentual de metas vinculadas a area.
- Percentual de tarefas vinculadas a meta/projeto.
- Retencao D7 dos usuarios que concluem onboarding.

## Modelo alvo recomendado (MVP pragmatico)

### Hierarquia

`Area (por que) -> Meta (resultado) -> Projeto (entrega) -> Tarefa (execucao)`

### Regras de negocio (MVP)

1. Toda meta deve ter `area_id`.
2. Todo projeto deve ter `area_id` e pode ter `goal_id` (opcional no MVP, recomendado quando existir meta).
3. Toda tarefa pode ter `project_id`.
4. Se tarefa tiver `project_id`, herdar/default de `area_id` e `goal_id` do projeto para manter consistencia.

## Roadmap por fases

## Fase 0 (1 semana): Base de produto e quick wins de onboarding

Objetivo: melhorar conversao sem esperar projetos completos.

- Permitir area customizada no onboarding (alem de sugestoes).
- Fazer onboarding de metas com selecao de area.
- Tornar habitos opcionais no onboarding (com "Pular por agora").
- Corrigir update de area para preservar `is_active` (ou enviar explicitamente no form).
- Instrumentar eventos do funil (inicio, cada etapa, conclusao, primeira tarefa).
- Definir metricas baseline (antes/depois).

Valor esperado:

- Menor friccao inicial.
- Mais dados ja vinculados por area.

## Fase 1 (1-2 semanas): Backend de projetos (MVP)

Objetivo: criar fundacao tecnica para integrar planejamento e execucao.

- Migration `projects`:
  - `id`, `workspace_id`, `area_id`, `goal_id (nullable)`, `name`, `description`, `status`, `start_date`, `due_date`, `completed_at`, `created_at`, `updated_at`, `deleted_at`.
- Migration em `tasks` para `project_id` (nullable + index).
- Service/handler/router de projetos (`/api/v1/projects` + leitura no `/public/v1` se fizer sentido).
- Regras de validacao:
  - projeto e meta devem pertencer ao mesmo workspace;
  - se `project.goal_id` existir, validar consistencia com area.
- Atualizar OpenAPI e tipos frontend.

Valor esperado:

- Nova camada de organizacao sem quebrar tarefas existentes.

## Fase 2 (1-2 semanas): Frontend de projetos + integracao com tarefas

Objetivo: colocar projetos em uso real no dia a dia.

- Nova pagina `Projetos`:
  - criar/editar/status/progresso;
  - filtros por area e meta;
  - cards com progresso de tarefas.
- Em tarefas:
  - selecionar projeto no create/edit;
  - filtrar por projeto;
  - exibir chip de projeto na listagem.
- Em metas:
  - mostrar projetos relacionados por meta.
- Em areas:
  - mostrar quantidade de projetos ativos por area.

Valor esperado:

- "Ponte" clara entre metas e execucao diaria.

## Fase 3 (1 semana): Onboarding 2.0 (amarrado)

Objetivo: o usuario concluir onboarding ja com estrutura utilizavel.

Fluxo sugerido:

1. Escolher/criar areas (sugestoes + custom).
2. Criar 1-3 metas, cada uma vinculada a uma area.
3. Criar 1 projeto inicial vinculado a area/meta.
4. Criar primeira tarefa do projeto.
5. Habitos (opcional) e concluir.

Valor esperado:

- Usuario termina onboarding com cadeia completa `area -> meta -> projeto -> tarefa`.

## Fase 4 (1 semana): Dashboard e Insights com projetos

Objetivo: refletir o novo modelo nas telas de valor percebido.

- Dashboard:
  - bloco de projetos ativos (atrasados, no prazo, concluidos).
  - progresso por projeto e por meta.
- Insights:
  - incluir saude de projetos na geracao.
  - destacar risco de meta sem projeto ativo.
- Score:
  - opcional no MVP; se entrar, adicionar componente leve de "execucao de projetos".

Valor esperado:

- Mais clareza de progresso real e recomendacoes acionaveis.

## Priorizacao (valor x esforco)

### Agora (alto valor, baixo/medio esforco)

- Fase 0 completa.
- Definicao do contrato de dados de projetos.

### Proximo (alto valor, medio esforco)

- Fase 1 + Fase 2 MVP.

### Depois (alto valor, maior esforco)

- Fase 3 + Fase 4 com refinamento de score/insights.

## KPIs de sucesso

1. Onboarding completion rate: +15% em 30 dias.
2. Tempo ate primeira tarefa: reduzir para menos de 10 minutos.
3. % metas com area vinculada: >= 90%.
4. % tarefas com projeto ou meta vinculada: >= 70%.
5. Retencao D7 de novos usuarios: +8 p.p.

## Riscos e mitigacoes

1. Complexidade de modelo cedo demais.
   - Mitigacao: MVP com `project.goal_id` opcional e regras simples.
2. Quebra de fluxo de tarefas (que hoje esta forte).
   - Mitigacao: manter campos atuais, adicionar `project_id` incrementalmente.
3. Baixa adocao de projetos.
   - Mitigacao: onboarding 2.0 cria primeiro projeto obrigatorio e ja gera primeira tarefa.

## Plano pratico de execucao (proximos 14 dias)

### Semana 1

- Implementar Fase 0 inteira.
- Publicar dashboard interno simples de funil de onboarding.
- Definir schema e contratos de API de projetos.

### Semana 2

- Subir migrations + API de projetos (Fase 1).
- Entregar tela basica de projetos e vinculacao em tarefas (primeira parte da Fase 2).
- Liberar feature flag para cohort de teste.
