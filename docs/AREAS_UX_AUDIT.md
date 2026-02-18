# Areas UX/UI Audit — Áreas vs Tarefas

## Resumo Executivo

A experiência de Áreas evoluiu de CRUD básico para uma tela operacional com cards ricos, filtros, detail view, confirmação destrutiva, reorder drag-and-drop e feedback por toast. Ainda assim, Tarefas continua mais madura em densidade de interação. Esta rodada fecha os gaps críticos de UX restantes e alinha o modal de Área ao padrão híbrido `Todoist + Tasks`.

---

## Changelog

### P0 Implementado — PR #13 (`feat/areas-ux-upgrade`)

| Item | Status | Notes |
|---|---|---|
| Area detail page | Done | `/areas/:id` com stats, metas, tarefas (preview) e projetos |
| Enriquecimento de cards | Done | Score, barra, counters (tasks/goals/projects), ícone e hover accent |
| Delete com confirmação | Done | `ConfirmDialog` com contexto de entidades vinculadas |

### P1 Implementado — `feat/areas-ux-upgrade`

| Item | Status | Notes |
|---|---|---|
| Icon picker | Done | Componente reutilizável com busca |
| Empty state | Done | Mensagem + CTA para criar área |
| Filtro e busca | Done | Chips Ativas/Inativas + busca textual + limpar |

### P2 (parcialmente já implementado no código)

| Item | Status | Notes |
|---|---|---|
| Shortcut `N` para nova área | Done | Atalho global na página de Áreas |
| Drag to reorder | Done | DnD com persistência via `/areas/:id/reorder` |
| Toasts em create/update/delete | Done | Feedback de sucesso/erro nas mutações de Área |
| Color picker com labels | Done | Swatches com labels de texto |
| Skeleton responsivo | Done | Grid 1/2/3 colunas por breakpoint |

### Correções críticas desta entrega (atual)

| Item | Status | Notes |
|---|---|---|
| Refatoração completa do modal de Área | Done | Componente único reutilizado em `/areas` e `/areas/:id` |
| Deeplink para Tasks com filtro aplicado | Done | `tasks.tsx` lê `?area_id=` na inicialização |
| Estado de erro/not-found em `/areas/:id` | Done | Empty/error state com CTA “Voltar para Áreas” |
| ConfirmDialog assíncrono correto | Done | Não fecha automaticamente; fechamento controlado por sucesso |
| Tokens inválidos no IconPicker | Done | Classes substituídas por tokens válidos do design system |
| Padronização PT-BR no escopo alterado | Done | “Áreas”, “área”, “Ícone”, “Criar área”, “Editar área” |
| Consumo de limite na home de Áreas | Done | Badge de uso (`usado/limite`) no header, no mesmo padrão de Tarefas |
| Sync bidirecional URL↔filtros em Tasks | Done | Filtros refletem em querystring e restauram estado via back/forward |
| URL sync com filtros avançados em Tasks | Done | Inclui `goal_id`, `due_from` e `due_to` no parse/build da querystring |
| Ações inline no detail da Área | Done | Criação rápida de tarefa/meta/projeto em `/areas/:id`, pré-vinculada à área |
| Header mobile otimizado no detail da Área | Done | Ações compactas e hierarquia em duas camadas para melhor legibilidade no mobile |
| Cobertura de testes críticos (web) | Done | Unit/component tests para URL de filtros, modal de área (Enter/Esc/limite) e ConfirmDialog |

---

## Estado Atual: Áreas vs Tarefas

| Capacidade | Tarefas | Áreas |
|---|---|---|
| Criação rápida | Inline quick-add + parsing inteligente | Modal rápido com foco no nome + metadata lateral |
| Criação expandida | Dialog rico com sidebar e metadados | Dialog híbrido (principal + sidebar + avançado) |
| Filtros | Status, áreas, projetos, seções, labels | Ativas/inativas + busca |
| Hierarquia | Subtarefas + seções + colapso | Lista de áreas (flat) |
| Navegação contextual | Alta | Média (lista -> detail -> links para tarefas/projetos) |
| Feedback de ação | Alto (toasts + animações) | Alto (toasts + confirmações + estados de erro) |
| Reordenação | Reorder de tarefas/seções | Reorder por drag-and-drop |

---

## Gaps Críticos Fechados nesta Rodada

1. **Deep link quebrado para Tasks**
   - Antes: `/tasks?area_id=...` não aplicava filtro automaticamente.
   - Agora: filtro inicial por `area_id` é aplicado na montagem da página.

2. **Resiliência insuficiente no detail de Área**
   - Antes: `isLoading || !summary` podia mascarar erro como loading infinito.
   - Agora: estados explícitos de loading vs erro/not-found.

3. **Confirmação destrutiva prematura**
   - Antes: diálogo fechava no clique de confirmar, antes do sucesso da mutation.
   - Agora: diálogo só fecha no sucesso, mantendo contexto no erro.

4. **Inconsistência visual no IconPicker**
   - Antes: classes inexistentes (`bg-bg-base`, `hover:bg-bg-hover`).
   - Agora: tokens válidos e consistentes com superfícies do app.

5. **Inconsistência de linguagem PT-BR**
   - Antes: mix de textos sem acentuação.
   - Agora: padronização no escopo alterado.

---

## Arquivos-chave da rodada atual

- `web/src/components/areas/area-form-dialog.tsx`
- `web/src/pages/areas.tsx`
- `web/src/pages/area-detail.tsx`
- `web/src/pages/tasks.tsx`
- `web/src/components/ui/confirm-dialog.tsx`
- `web/src/components/ui/icon-picker.tsx`

---

## Próximos passos recomendados (não incluídos)

1. Adicionar cenário E2E para criação inline no detail de Área (tarefa/meta/projeto) validando vínculo automático por `area_id`.
2. Expandir URL sync para filtros adicionais de Tasks, se forem promovidos na UI (ex.: `has_parent`, janelas de data preset, multi-select de etiquetas).
3. Ajustar tracking/telemetria das novas ações inline de Área para medir adoção (quando houver camada de analytics).
