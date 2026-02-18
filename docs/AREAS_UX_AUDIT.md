# Areas UX/UI Audit — Areas vs Tasks

## Resumo Executivo

Tasks é experiência madura (~1600 linhas), múltiplos padrões de interação, feedback visual rico. Areas é CRUD básico (~114 linhas). O AreasGrid do dashboard (90 linhas) já tem mais craft que a página dedicada.

---

## Changelog

### P0 Implemented — PR #13 (`feat/areas-ux-upgrade`)

| Item | Status | Notes |
|---|---|---|
| 1. Area detail page | Done | `/areas/:id` with stats, goals, tasks preview (top 5), projects |
| 2. Enrich cards | Done | Score, progress bar, counters (tasks/goals/projects), icon bg, hover accent |
| 3. Delete with confirmation | Done | `ConfirmDialog` shows linked entity counts from AreaSummary |

**Files changed:**
- `web/src/types/api.ts` — `AreaWithStats`, `AreaStats`, `AreaSummary`
- `web/src/hooks/use-areas.ts` — `useAreas(?include=stats)`, `useAreaSummary(id)`
- `web/src/pages/areas.tsx` — rich cards, exported `AreaFormDialog`
- `web/src/pages/area-detail.tsx` — new detail page
- `web/src/components/ui/confirm-dialog.tsx` — reusable confirm dialog
- `web/src/routes/index.tsx` — `areas/:id` route

### P1 Implemented — `feat/areas-ux-upgrade`

| Item | Status | Notes |
|---|---|---|
| 4. Icon picker | Done | Popover grid of `areaIconMap` icons with search, reusable `IconPicker` component |
| 5. Empty state | Done | Centered LayoutGrid icon + message + CTA opens create dialog |
| 6. Filter/search | Done | FilterChip Ativas/Inativas toggle + inline search + "Limpar" reset |

**Files changed:**
- `web/src/components/ui/icon-picker.tsx` — new reusable icon picker popover
- `web/src/pages/areas.tsx` — IconPicker in form, FilterChip, search, empty state

---

## 1. O que Tasks faz bem (e Areas não tem)

| Capacidade | Tasks | Areas |
|---|---|---|
| **Criação rápida** | Inline quick-add com smart parsing (`p1`, `hoje`, `dd/mm`) | Só dialog |
| **Criação expandida** | Dialog 2 colunas com sidebar de metadata | Dialog single-column, 5 campos |
| **Filtros** | 5 categorias (status, áreas, projetos, seções, labels) com chips | Zero |
| **Hierarquia** | Parent/child com expand/collapse | Flat |
| **Indicadores visuais** | Priority checkbox colorido, due date badges, labels, focus star, project badge, duration | Só ícone + peso |
| **Ações inline** | 6 ações no hover (reabrir, focus, agendar, sub-tarefa, menu, expand) | Só "Excluir" |
| **Keyboard shortcuts** | `Q` para quick-add, `Escape` para cancelar | Nenhum |
| **Empty state** | Mensagem randomizada + CTA | Grid vazia (silêncio) |
| **Progressive disclosure** | Ações aparecem no hover, seções colapsáveis | Tudo visível sempre |
| **Feedback de criação** | Animação slideIn + highlight laranja por 1.5s | Nada |
| **Usage tracking** | Badge com contador diário + progress bar | Nada |
| **Seções colapsáveis** | CompletedSection com toggle | N/A |

---

## 2. Problemas específicos da página Areas

### ~~2.1. Card anêmico~~ — FIXED

Cards agora mostram score, progress bar, counters (tasks/goals/projects), ícone com fundo colorido.

### ~~2.2. Dashboard AreasGrid é MELHOR que a página Areas~~ — FIXED

Página Areas agora usa mesmo pattern do AreasGrid (score, progress bar, gradientes, hover accent, staggered animation).

### ~~2.3. Sem detail view~~ — FIXED

Click no card navega para `/areas/:id` com stats, goals, tasks preview, projects.

### ~~2.4. Delete sem confirmação~~ — FIXED

`ConfirmDialog` mostra contagem de entidades vinculadas antes de excluir.

### 2.5. Form dialog básico

- Campo "Icone" é text input livre — user precisa saber emoji ou nome Lucide
- Sem preview, busca ou picker de ícones
- Cor é 6 botões sem label — problema de acessibilidade
- Peso é número sem contexto (o que significa peso 3 vs 7?)

### 2.6. Sem busca/filtro

Com 6+ áreas, sem filtro por nome, cor, status ativo/inativo.

### 2.7. Áreas inativas invisíveis

Checkbox `is_active` existe no form, sem indicação visual no card nem filtro.

---

## 3. Recomendações priorizadas

### P0 — Impacto alto, esforço baixo — DONE (PR #13)

1. ~~**Area detail page**~~: `/areas/:id` com stats row, goals list, tasks preview (top 5 + "Ver todas"), projects, edit/delete.

2. ~~**Enriquecer os cards**~~: Score mono, progress bar gradient, counters (tasks/goals/projects), icon bg, hover accent line, staggered animation.

3. ~~**Delete com confirmação**~~: `ConfirmDialog` com variant destructive, mostra contagem de entidades vinculadas.

### P1 — Impacto alto, esforço médio — DONE

4. ~~**Icon picker**~~: Popover grid com todos ícones de `areaIconMap`, busca por nome, seleção com highlight. Componente reutilizável `icon-picker.tsx`.

5. ~~**Empty state**~~: Centered LayoutGrid icon + "Organize sua vida em áreas" + CTA "Nova área".

6. ~~**Filtro/busca**~~: FilterChip Ativas/Inativas + search input + "Limpar" button. Client-side filtering.

7. **Inline stats no card**: (already done in P0)
   ```
   ┌─────────────────────────┐
   │ 🏋️ Saúde          Score: 72 │
   │ ━━━━━━━━━━━░░░░          │
   │ 5 tasks · 2 goals · 1 projeto │
   └─────────────────────────┘
   ```

### P2 — Nice to have

8. **Keyboard shortcut** (`N` para nova área)
9. **Drag to reorder** (weight manual é confuso)
10. **Color picker melhorado**: Labels de texto nas cores + preview
11. **Toast notifications** em create/update/delete
12. **Skeleton loading** responsivo (hardcoded 3 colunas em mobile)

---

## 4. Padrões de Tasks para reutilizar

| Pattern | Onde está | Como usar em Areas |
|---|---|---|
| `FilterChip` | `tasks.tsx:817` | Filtro ativo/inativo + cor |
| `CompletedSection` | `tasks.tsx:1162` | Seção "Áreas inativas" colapsável |
| `SidebarField` | `tasks.tsx:1151` | Detail panel da área |
| Empty state pattern | `tasks.tsx:1507-1520` | Empty state Areas |
| `AREA_CHIP_ACTIVE` colors | `tasks.tsx:72-79` | Badges coloridos nos cards |
| Animação slideIn | `tasks.tsx:583` | Feedback ao criar área |

---

## 5. Visão proposta

```
┌──────────────────────────────────────────────┐
│  Areas de Vida          [Busca] [+ Nova área] │
│                                               │
│  [Ativas ✓] [Inativas] [Todas]               │
│                                               │
│  ┌─────────────┐ ┌─────────────┐ ┌────────── │
│  │ 🏋️ Saúde     │ │ 💼 Carreira  │ │ 📚 ...   │
│  │ Score: 72   │ │ Score: 45   │ │          │
│  │ ━━━━━━━░░░  │ │ ━━━░░░░░░░ │ │          │
│  │ 8 tasks     │ │ 3 tasks     │ │          │
│  │ 2 goals     │ │ 1 goal      │ │          │
│  │ 1 projeto   │ │             │ │          │
│  └─────────────┘ └─────────────┘ └────────── │
│                                               │
│  Click no card → Abre detail panel/page:      │
│  ┌──────────────────────────────────────────┐ │
│  │ Saúde — Editar                          │ │
│  │ Score: 72  Weight: 5  Cor: 🟢           │ │
│  │                                          │ │
│  │ Goals (2)                                │ │
│  │  ✅ Perder 5kg (80%)                    │ │
│  │  ⏳ Correr maratona (20%)               │ │
│  │                                          │ │
│  │ Tasks pendentes (8)  [Ver todas →]       │ │
│  │  ☐ Treinar hoje                         │ │
│  │  ☐ Comprar suplementos                  │ │
│  │  ...                                    │ │
│  │                                          │ │
│  │ Seções: Rotina · Projetos · Estudos     │ │
│  └──────────────────────────────────────────┘ │
└──────────────────────────────────────────────┘
```

---

## Conclusão

A página Areas hoje é um **form CRUD sem contexto** — user cria áreas mas não tem visibilidade sobre elas. O dashboard (AreasGrid) oferece mais valor visual. Prioridade #1: **dar profundidade à área** — mostrar o que está dentro (tasks, goals, scores, projetos) e permitir navegação. Os patterns já existem no codebase (Tasks), é questão de reutilizá-los.
