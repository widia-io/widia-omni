# Changelog

## [Unreleased]

### Added

- **Areas UX P0**: Rich area cards (score, progress bar, counters, icon bg, hover accent), detail page `/areas/:id`, delete confirmation dialog with linked entity counts
- **Areas UX P1**: Icon picker popover (grid + search), empty state with CTA, filter bar (active/inactive toggle + name search)

- **Task power-ups backend** (Phase 1A): labels CRUD, sections CRUD, sub-tasks via parent_id, position ordering, duration_minutes, reopen/reorder endpoints, label assignment on tasks
- **Task power-ups frontend** (Phase 1B): Todoist-style UI rewrite with flat rows, priority-colored checkboxes, inline quick-add, hover-revealed actions, section grouping, sub-task nesting, label picker/manager, filter bar (status/area/section/label), full edit dialog
- **Smart keyboard shortcuts**: `Q` opens quick-add; type `p1`–`p4` to set priority; type `hoje`/`amanha`/`dd/mm` to set due date; live token preview pills
- **AI weekly insights** (M7): LLM-powered weekly review with habit/goal analysis
- **Public API** (M8): API key authentication for external integrations
- **M9 Growth (phaseado)**:
  - Family workspace management (multi-workspace listing/switch, members list/remove, invite create/list/revoke/accept)
  - Referral system (workspace referral codes, attributions, 30-day credit ledger, conversion processing on paid activation)
  - Mobile PWA foundations (vite-plugin-pwa, manifest, service worker registration, mobile app-shell/navigation updates)
- **CI/CD de produção do meufoco.app**:
  - Workflow automático em merge/push na `main` com build/test + publish em GHCR
  - Deploy remoto via SSH para `/opt/stacks/meufoco` com migrations automáticas
  - Health-checks pós deploy com rollback automático para `.last_successful_tag`
  - Workflow manual de rollback por tag (`workflow_dispatch`)
