import { useState, useMemo, useRef, useEffect, useCallback } from "react";
import {
  Plus, Star, ChevronDown, ChevronRight, ListTree, Clock,
  RotateCcw, Tags, FolderPlus, Check, MoreHorizontal, CalendarDays,
  Flag, Folder,
} from "lucide-react";
import { format, isToday, isPast, isTomorrow } from "date-fns";
import {
  useTasks, useCreateTask, useUpdateTask, useDeleteTask,
  useCompleteTask, useToggleFocus, useReopenTask,
} from "@/hooks/use-tasks";
import type { CreateTaskInput } from "@/hooks/use-tasks";
import { useAreas } from "@/hooks/use-areas";
import { useGoals } from "@/hooks/use-goals";
import { useLabels } from "@/hooks/use-labels";
import { useSections, useCreateSection } from "@/hooks/use-sections";
import { LabelPicker } from "@/components/tasks/label-picker";
import { LabelManagerDialog } from "@/components/tasks/label-manager-dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Skeleton } from "@/components/ui/skeleton";
import { Textarea } from "@/components/ui/textarea";
import {
  Select, SelectTrigger, SelectValue, SelectContent, SelectItem,
} from "@/components/ui/select";
import {
  Dialog, DialogContent, DialogHeader, DialogTitle,
} from "@/components/ui/dialog";
import { useWorkspaceUsage } from "@/hooks/use-settings";
import { cn } from "@/lib/cn";
import type { Task, TaskPriority } from "@/types/api";

// ─── Helpers ──────────────────────────────────────────

const PRIORITY_COLORS: Record<TaskPriority, string> = {
  critical: "border-accent-rose",
  high: "border-accent-orange",
  medium: "border-accent-blue",
  low: "border-border",
};

const PRIORITY_PILL_ACTIVE: Record<TaskPriority, string> = {
  critical: "bg-accent-rose/12 text-accent-rose border-accent-rose/25",
  high: "bg-accent-orange/12 text-accent-orange border-accent-orange/25",
  medium: "bg-accent-blue/12 text-accent-blue border-accent-blue/25",
  low: "bg-bg-secondary text-text-secondary border-border",
};

const PRIORITY_DOT: Record<TaskPriority, string> = {
  critical: "bg-accent-rose",
  high: "bg-accent-orange",
  medium: "bg-accent-blue",
  low: "bg-text-muted",
};

const PRIORITY_TOOLTIP: Record<TaskPriority, string> = {
  critical: "Urgente",
  high: "Importante",
  medium: "Normal",
  low: "Baixa",
};

const LABEL_TEXT_COLOR: Record<string, string> = {
  orange: "text-accent-orange",
  blue: "text-accent-blue",
  green: "text-accent-green",
  rose: "text-accent-rose",
  sand: "text-accent-sand",
  sage: "text-accent-sage",
};

const AREA_CHIP_ACTIVE: Record<string, string> = {
  green: "bg-accent-green/10 text-accent-green border-accent-green/20",
  orange: "bg-accent-orange/10 text-accent-orange border-accent-orange/20",
  blue: "bg-accent-blue/10 text-accent-blue border-accent-blue/20",
  rose: "bg-accent-rose/10 text-accent-rose border-accent-rose/20",
  sand: "bg-accent-sand/10 text-accent-sand border-accent-sand/20",
  sage: "bg-accent-sage/10 text-accent-sage border-accent-sage/20",
};

const EMPTY_STATES = [
  { title: "Tudo limpo!", desc: "Nada pendente. Que tal planejar algo importante?" },
  { title: "Foco total", desc: "Adicione suas prioridades do dia." },
  { title: "Momento de planejar", desc: "O que você quer conquistar hoje?" },
  { title: "Pronto para decolar?", desc: "Crie sua primeira tarefa e comece." },
];

function formatDuration(mins: number): string {
  if (mins < 60) return `${mins}min`;
  const h = Math.floor(mins / 60);
  const m = mins % 60;
  return m ? `${h}h${m}` : `${h}h`;
}

function formatDueDate(date: string): { text: string; className: string; badge?: boolean } {
  const d = new Date(date);
  if (isToday(d)) return { text: "Hoje", className: "bg-accent-orange/15 text-accent-orange border border-accent-orange/20 rounded-full px-2 py-0.5", badge: true };
  if (isTomorrow(d)) return { text: "Amanhã", className: "text-text-secondary" };
  if (isPast(d)) return { text: format(d, "dd/MM"), className: "text-accent-rose" };
  return { text: format(d, "dd/MM"), className: "text-text-muted" };
}

// ─── Smart Input Parser ──────────────────────────────

const SMART_PRIORITY: Record<string, TaskPriority> = {
  p1: "critical", p2: "high", p3: "medium", p4: "low",
};

interface SmartToken { type: "priority" | "date"; raw: string; label: string }

function parseSmartInput(text: string) {
  const tokens: SmartToken[] = [];
  let priority: TaskPriority | undefined;
  let dueDate: string | undefined;
  let clean = text;

  const pm = text.match(/\b(p[1-4])\b/i);
  if (pm) {
    priority = SMART_PRIORITY[pm[1]!.toLowerCase()];
    tokens.push({ type: "priority", raw: pm[0]!, label: pm[1]!.toUpperCase() });
    clean = clean.replace(pm[0]!, "");
  }

  const hm = text.match(/\bhoje\b/i);
  if (hm) {
    const d = new Date(); d.setHours(23, 59, 0, 0);
    dueDate = d.toISOString();
    tokens.push({ type: "date", raw: hm[0], label: "Hoje" });
    clean = clean.replace(hm[0], "");
  } else {
    const am = text.match(/\bamanh[aã]\b/i);
    if (am) {
      const d = new Date(); d.setDate(d.getDate() + 1); d.setHours(23, 59, 0, 0);
      dueDate = d.toISOString();
      tokens.push({ type: "date", raw: am[0], label: "Amanhã" });
      clean = clean.replace(am[0], "");
    } else {
      const dm = text.match(/\b(\d{1,2})\/(\d{1,2})\b/);
      if (dm) {
        const d = new Date();
        d.setMonth(parseInt(dm[2]!) - 1, parseInt(dm[1]!));
        d.setHours(23, 59, 0, 0);
        if (d < new Date()) d.setFullYear(d.getFullYear() + 1);
        dueDate = d.toISOString();
        tokens.push({ type: "date", raw: dm[0]!, label: `${dm[1]}/${dm[2]}` });
        clean = clean.replace(dm[0], "");
      }
    }
  }

  return { cleanTitle: clean.replace(/\s+/g, " ").trim(), priority, dueDate, tokens };
}

// ─── Task Usage Badge ─────────────────────────────────

function TaskUsageBadge() {
  const { data: usage } = useWorkspaceUsage();

  if (!usage) return null;

  const used = usage.counters.tasks_created_today;
  const max = usage.limits.max_tasks_per_day;
  const isUnlimited = max === -1;
  const ratio = isUnlimited ? 0 : max > 0 ? used / max : 0;
  const isFull = !isUnlimited && ratio >= 1;
  const isWarning = !isUnlimited && ratio >= 0.8 && !isFull;

  const accentColor = isFull
    ? "text-accent-rose"
    : isWarning
      ? "text-accent-orange"
      : "text-text-muted";

  const barColor = isFull
    ? "bg-accent-rose"
    : isWarning
      ? "bg-accent-orange"
      : "bg-accent-green";

  return (
    <div className="flex items-center gap-2.5 rounded-full border border-border/60 px-3 py-1">
      <div className="flex items-baseline gap-1">
        <span className={cn("font-mono text-xs font-semibold tabular-nums", accentColor)}>
          {used}
        </span>
        <span className="text-[10px] text-text-muted">/</span>
        <span className="font-mono text-[10px] text-text-muted">
          {isUnlimited ? "∞" : max}
        </span>
        <span className="text-[10px] text-text-muted">hoje</span>
      </div>
      {!isUnlimited && (
        <div className="h-1 w-14 overflow-hidden rounded-full bg-border/50">
          <div
            className={cn("h-full rounded-full transition-all duration-700 ease-out", barColor)}
            style={{ width: `${Math.min(ratio * 100, 100)}%` }}
          />
        </div>
      )}
    </div>
  );
}

// ─── Priority Checkbox ────────────────────────────────

function PriorityCheckbox({
  priority, isCompleted, onComplete, onReopen,
}: {
  priority: TaskPriority;
  isCompleted: boolean;
  onComplete: () => void;
  onReopen: () => void;
}) {
  if (isCompleted) {
    return (
      <button
        onClick={onReopen}
        title={PRIORITY_TOOLTIP[priority]}
        className="group/check flex h-[18px] w-[18px] shrink-0 items-center justify-center rounded-full border-2 border-accent-green bg-accent-green transition-all duration-150"
      >
        <Check className="h-2.5 w-2.5 text-bg-primary" strokeWidth={3} />
      </button>
    );
  }

  return (
    <button
      onClick={onComplete}
      title={PRIORITY_TOOLTIP[priority]}
      className={cn(
        "flex h-[18px] w-[18px] shrink-0 items-center justify-center rounded-full border-2 transition-all duration-150",
        PRIORITY_COLORS[priority],
      )}
      onMouseEnter={(e) => {
        const bg = priority === "critical" ? "rgba(244,63,94,0.15)"
          : priority === "high" ? "rgba(249,115,22,0.15)"
          : priority === "medium" ? "rgba(59,130,246,0.15)"
          : "rgba(46,46,40,0.15)";
        e.currentTarget.style.backgroundColor = bg;
      }}
      onMouseLeave={(e) => {
        e.currentTarget.style.backgroundColor = "transparent";
      }}
    >
      <Check className="h-2.5 w-2.5 opacity-0 transition-opacity group-hover/row:opacity-30" style={{ color: "currentColor" }} strokeWidth={3} />
    </button>
  );
}

// ─── Inline Quick Add ─────────────────────────────────

function ActionPill({
  icon: Icon, label, isActive, onClick,
}: {
  icon: React.ComponentType<{ className?: string }>;
  label: string;
  isActive?: boolean;
  onClick: () => void;
}) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={cn(
        "inline-flex items-center gap-1.5 rounded-lg border px-2.5 py-1 text-xs transition-all duration-150",
        isActive
          ? "border-accent-orange/25 bg-accent-orange/8 text-accent-orange"
          : "border-border text-text-muted hover:border-border-hover hover:text-text-secondary",
      )}
    >
      <Icon className="h-3.5 w-3.5" />
      {label}
    </button>
  );
}

function InlineQuickAdd({
  sectionId, parentId, onExpandedCreate, trigger, onCreated,
}: {
  sectionId?: string;
  parentId?: string;
  onExpandedCreate?: () => void;
  trigger?: number;
  onCreated?: (id: string) => void;
}) {
  const [isOpen, setIsOpen] = useState(false);
  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [priority, setPriority] = useState<TaskPriority>("medium");
  const [showPriority, setShowPriority] = useState(false);
  const [dueDate, setDueDate] = useState("");
  const [showDueDate, setShowDueDate] = useState(false);
  const inputRef = useRef<HTMLInputElement>(null);
  const dueDateRef = useRef<HTMLInputElement>(null);
  const wrapperRef = useRef<HTMLDivElement>(null);
  const createTask = useCreateTask();

  useEffect(() => {
    if (isOpen && inputRef.current) inputRef.current.focus();
  }, [isOpen]);

  useEffect(() => {
    if (!isOpen) return;
    function handleClickOutside(e: MouseEvent) {
      if (wrapperRef.current && !wrapperRef.current.contains(e.target as Node)) {
        resetForm();
      }
    }
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, [isOpen]);

  useEffect(() => {
    if (trigger && trigger > 0) setIsOpen(true);
  }, [trigger]);

  const parsed = parseSmartInput(title);

  function resetForm() {
    setIsOpen(false);
    setTitle("");
    setDescription("");
    setPriority("medium");
    setShowPriority(false);
    setDueDate("");
    setShowDueDate(false);
  }

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!title.trim()) return;
    const finalDue = parsed.dueDate ?? (dueDate ? new Date(dueDate).toISOString() : undefined);
    const payload: CreateTaskInput = {
      title: parsed.cleanTitle || title.trim(),
      priority: parsed.priority ?? priority,
      ...(finalDue && { due_date: finalDue }),
      ...(description.trim() && { description: description.trim() }),
      ...(sectionId && { section_id: sectionId }),
      ...(parentId && { parent_id: parentId }),
    };
    createTask.mutate(payload, {
      onSuccess: (data) => {
        setTitle("");
        setDescription("");
        setPriority("medium");
        setShowPriority(false);
        setDueDate("");
        setShowDueDate(false);
        inputRef.current?.focus();
        if (data?.id && onCreated) onCreated(data.id);
      },
    });
  }

  if (!isOpen) {
    return (
      <button
        onClick={() => setIsOpen(true)}
        className={cn(
          "flex w-full items-center gap-2 py-2 text-sm text-text-muted transition-colors hover:text-accent-orange",
          parentId && "pl-8",
        )}
      >
        <Plus className="h-4 w-4" />
        <span>{parentId ? "Adicionar sub-tarefa..." : "Adicionar tarefa..."}</span>
      </button>
    );
  }

  const priorities: TaskPriority[] = ["critical", "high", "medium", "low"];
  const priorityLabels: Record<TaskPriority, string> = {
    critical: "P1", high: "P2", medium: "P3", low: "P4",
  };

  const hasText = title.trim().length > 0;
  const hasDueDate = !!dueDate || !!parsed.dueDate;
  const hasCustomPriority = priority !== "medium" || !!parsed.priority;

  return (
    <div ref={wrapperRef} className={cn("my-1", parentId && "pl-8")}>
      <form onSubmit={handleSubmit} className="overflow-hidden rounded-[14px] border border-border bg-bg-card">
        {/* ── Input area ── */}
        <div className="px-3.5 pt-3 pb-2">
          <input
            ref={inputRef}
            type="text"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            placeholder={parentId ? "Adicionar sub-tarefa..." : "O que precisa ser feito hoje?"}
            autoComplete="off"
            className="w-full border-0 bg-transparent text-sm font-medium text-text-primary placeholder:text-text-muted focus:outline-none"
            onKeyDown={(e) => {
              if (e.key === "Escape") resetForm();
            }}
          />
          <input
            type="text"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            placeholder="Descrição"
            autoComplete="off"
            className="mt-1 w-full border-0 bg-transparent text-xs text-text-secondary placeholder:text-text-muted/60 focus:outline-none"
          />
        </div>

        {/* ── Action pills ── */}
        <div className="flex flex-wrap items-center gap-1.5 px-3.5 pb-2.5">
          <ActionPill
            icon={CalendarDays}
            label={dueDate ? formatDueDate(new Date(dueDate).toISOString()).text : "Prazo"}
            isActive={hasDueDate}
            onClick={() => {
              setShowDueDate(!showDueDate);
              setShowPriority(false);
            }}
          />
          <ActionPill
            icon={Flag}
            label={hasCustomPriority ? PRIORITY_TOOLTIP[parsed.priority ?? priority] : "Prioridade"}
            isActive={hasCustomPriority}
            onClick={() => {
              setShowPriority(!showPriority);
              setShowDueDate(false);
            }}
          />
          {!parentId && (
            <ActionPill
              icon={Folder}
              label="Area"
              onClick={() => { resetForm(); onExpandedCreate?.(); }}
            />
          )}
          <ActionPill
            icon={Tags}
            label="Etiquetas"
            onClick={() => { resetForm(); onExpandedCreate?.(); }}
          />
          {onExpandedCreate && (
            <button
              type="button"
              onClick={() => { resetForm(); onExpandedCreate(); }}
              className="inline-flex items-center justify-center rounded-lg border border-border px-1.5 py-1 text-text-muted transition-colors hover:border-border-hover hover:text-text-secondary"
              title="Mais opções"
            >
              <MoreHorizontal className="h-3.5 w-3.5" />
            </button>
          )}
        </div>

        {/* ── Inline priority picker ── */}
        {showPriority && (
          <div className="flex items-center gap-1.5 border-t border-border/50 bg-bg-secondary/30 px-3.5 py-2">
            {priorities.map((p) => (
              <button
                key={p}
                type="button"
                title={PRIORITY_TOOLTIP[p]}
                onClick={() => { setPriority(p); setShowPriority(false); }}
                className={cn(
                  "inline-flex items-center gap-1.5 rounded-full border px-2.5 py-0.5 text-[11px] font-medium transition-all duration-150",
                  priority === p
                    ? PRIORITY_PILL_ACTIVE[p]
                    : "border-transparent text-text-muted hover:bg-bg-secondary/60",
                )}
              >
                <span className={cn("h-1.5 w-1.5 rounded-full", PRIORITY_DOT[p])} />
                {priorityLabels[p]}
              </button>
            ))}
          </div>
        )}

        {/* ── Inline due date picker ── */}
        {showDueDate && (
          <div className="flex items-center gap-2 border-t border-border/50 bg-bg-secondary/30 px-3.5 py-2">
            <CalendarDays className="h-3.5 w-3.5 text-text-muted" />
            <input
              ref={dueDateRef}
              type="datetime-local"
              value={dueDate}
              onChange={(e) => setDueDate(e.target.value)}
              autoFocus
              className="h-6 border-0 bg-transparent text-xs text-text-primary focus:outline-none"
            />
            {dueDate && (
              <button
                type="button"
                onClick={() => { setDueDate(""); setShowDueDate(false); }}
                className="text-[10px] text-text-muted hover:text-text-primary"
              >
                Limpar
              </button>
            )}
          </div>
        )}

        {/* ── Bottom bar ── */}
        <div className="flex items-center justify-between border-t border-border/50 bg-bg-secondary/20 px-3.5 py-2">
          <div className="flex items-center gap-1.5">
            {parsed.tokens.map((t) => (
              <span
                key={t.raw}
                className={cn(
                  "inline-flex items-center rounded-full px-2 py-0.5 text-[10px] font-medium",
                  t.type === "priority" ? "bg-accent-orange/15 text-accent-orange" : "bg-accent-blue/15 text-accent-blue",
                )}
              >
                {t.type === "priority" ? `Prioridade: ${t.label}` : t.label}
              </span>
            ))}
          </div>
          <div className="flex items-center gap-2">
            <button
              type="button"
              onClick={resetForm}
              className="rounded-lg px-2.5 py-1 text-xs text-text-muted transition-colors hover:bg-bg-secondary hover:text-text-primary"
            >
              Cancelar
            </button>
            <button
              type="submit"
              disabled={!hasText || createTask.isPending}
              className={cn(
                "rounded-lg bg-accent-orange px-3.5 py-1 text-xs font-semibold text-bg-primary transition-all hover:opacity-90 disabled:cursor-default",
                hasText ? "opacity-100" : "opacity-30",
              )}
            >
              Criar
            </button>
          </div>
        </div>
      </form>
    </div>
  );
}

// ─── Task Row ─────────────────────────────────────────

function TaskRow({
  task, childCount, isSubtask, isExpanded, onToggleExpand, onEdit, onAddSubtask, isRecent,
}: {
  task: Task;
  childCount: number;
  isSubtask?: boolean;
  isExpanded?: boolean;
  onToggleExpand?: () => void;
  onEdit: () => void;
  onAddSubtask?: () => void;
  isRecent?: boolean;
}) {
  const deleteTask = useDeleteTask();
  const completeTask = useCompleteTask();
  const toggleFocus = useToggleFocus();
  const reopenTask = useReopenTask();
  const [showMenu, setShowMenu] = useState(false);
  const menuRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!showMenu) return;
    function handleClick(e: MouseEvent) {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) {
        setShowMenu(false);
      }
    }
    document.addEventListener("mousedown", handleClick);
    return () => document.removeEventListener("mousedown", handleClick);
  }, [showMenu]);

  const dueInfo = task.due_date ? formatDueDate(task.due_date) : null;

  const hasLabels = (task.labels ?? []).length > 0;
  const hasMeta = dueInfo || hasLabels || (task.duration_minutes && !task.is_completed);

  return (
    <div
      className={cn(
        "group/row flex items-start gap-2.5 border-b border-border/40 px-2 py-2.5 transition-colors hover:rounded-lg hover:bg-bg-card/60",
        isSubtask && "pl-8",
        isRecent && "animate-[slideIn_0.35s_ease] rounded-lg border border-accent-orange/10 bg-accent-orange/5",
      )}
    >
      {/* Priority checkbox */}
      <div className="pt-0.5">
        <PriorityCheckbox
          priority={task.priority}
          isCompleted={task.is_completed}
          onComplete={() => completeTask.mutate(task.id)}
          onReopen={() => reopenTask.mutate(task.id)}
        />
      </div>

      {/* Content column */}
      <div className="min-w-0 flex-1 cursor-pointer" onClick={onEdit}>
        {/* Title row */}
        <div className="flex items-center gap-2">
          <span className={cn("truncate text-sm text-text-primary", task.is_completed && "text-text-muted line-through")}>
            {task.title}
          </span>
        </div>

        {/* Description */}
        {task.description && (
          <p className="mt-0.5 truncate text-xs text-text-muted">{task.description}</p>
        )}

        {/* Meta row: date + duration + labels */}
        {hasMeta && (
          <div className="mt-1 flex flex-wrap items-center gap-x-2.5 gap-y-1">
            {/* Due date */}
            {dueInfo && (
              dueInfo.badge ? (
                <span className={cn("inline-flex items-center gap-1 whitespace-nowrap text-[11px] font-medium", dueInfo.className)}>
                  <CalendarDays className="h-3 w-3" />
                  {dueInfo.text}
                </span>
              ) : (
                <span className={cn("inline-flex items-center gap-1 whitespace-nowrap text-[11px]", dueInfo.className)}>
                  <CalendarDays className="h-3 w-3" />
                  {dueInfo.text}
                </span>
              )
            )}

            {/* Duration */}
            {task.duration_minutes && (
              <span className="inline-flex items-center gap-1 whitespace-nowrap font-mono text-[11px] text-text-muted">
                <Clock className="h-3 w-3" />
                {formatDuration(task.duration_minutes)}
              </span>
            )}

            {/* Label pills */}
            {hasLabels && (task.labels ?? []).map((l) => (
              <span
                key={l.id}
                className={cn("inline-flex items-center gap-1 text-[11px] font-medium", LABEL_TEXT_COLOR[l.color] ?? "text-text-muted")}
              >
                <svg className="h-3 w-3" viewBox="0 0 16 16" fill="none" stroke="currentColor" strokeWidth="1.5">
                  <path d="M2 8.5V3.5C2 2.67 2.67 2 3.5 2H8.5L14 8L8.5 14L2 8.5Z" strokeLinejoin="round" />
                  <circle cx="5.5" cy="5.5" r="1" fill="currentColor" stroke="none" />
                </svg>
                {l.name}
              </span>
            ))}
          </div>
        )}
      </div>

      {/* Right side: subtask toggle + actions */}
      <div className="flex items-center gap-1 pt-0.5">
        {/* Sub-task expand toggle */}
        {childCount > 0 && (
          <button
            onClick={onToggleExpand}
            className="flex items-center gap-1 text-xs text-text-muted transition-colors hover:text-text-primary"
          >
            <ListTree className="h-3.5 w-3.5" />
            <span className="font-mono">{childCount}</span>
            {isExpanded ? <ChevronDown className="h-3 w-3" /> : <ChevronRight className="h-3 w-3" />}
          </button>
        )}

        {/* No-date hint (only when no meta row) */}
        {!dueInfo && !task.is_completed && !hasMeta && (
          <span className="flex items-center gap-1 whitespace-nowrap text-xs text-text-muted/50">
            <CalendarDays className="h-3 w-3" />
            Sem prazo
          </span>
        )}

        {/* Reopen hint (completed tasks) */}
        {task.is_completed && (
          <button
            onClick={() => reopenTask.mutate(task.id)}
            className="flex items-center gap-1 text-xs text-text-muted opacity-0 transition-opacity group-hover/row:opacity-100 hover:text-accent-orange"
          >
            <RotateCcw className="h-3 w-3" />
            <span>Reabrir</span>
          </button>
        )}

        {/* Focus star */}
        <button
          onClick={() => toggleFocus.mutate(task.id)}
          className={cn(
            "rounded p-1 transition-all",
            task.is_focus
              ? "text-accent-orange"
              : "text-text-muted opacity-0 group-hover/row:opacity-100 hover:text-accent-orange",
          )}
        >
          <Star className="h-4 w-4" fill={task.is_focus ? "currentColor" : "none"} />
        </button>

        {/* Schedule */}
        {!task.is_completed && (
          <button
            onClick={onEdit}
            className="rounded p-1 text-text-muted opacity-0 transition-all group-hover/row:opacity-100 hover:text-text-primary"
            title="Agendar"
          >
            <CalendarDays className="h-4 w-4" />
          </button>
        )}

        {/* Add sub-task */}
        {!isSubtask && !task.is_completed && onAddSubtask && (
          <button
            onClick={onAddSubtask}
            className="rounded p-1 text-text-muted opacity-0 transition-all group-hover/row:opacity-100 hover:text-text-primary"
            title="Sub-tarefa"
          >
            <Plus className="h-4 w-4" />
          </button>
        )}

        {/* More menu */}
        <div className="relative" ref={menuRef}>
          <button
            onClick={() => setShowMenu(!showMenu)}
            className="rounded p-1 text-text-muted opacity-0 transition-all group-hover/row:opacity-100 hover:text-text-primary"
          >
            <MoreHorizontal className="h-4 w-4" />
          </button>
          {showMenu && (
            <div className="absolute right-0 top-full z-50 mt-1 w-36 rounded-lg border border-border bg-bg-elevated p-1 shadow-lg">
              <button
                onClick={() => { deleteTask.mutate(task.id); setShowMenu(false); }}
                className="flex w-full items-center gap-2 rounded-md px-2 py-1.5 text-xs text-accent-rose transition-colors hover:bg-bg-card"
              >
                Excluir tarefa
              </button>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}

// ─── Section Create Form ──────────────────────────────

function SectionCreateInline({ areaId }: { areaId: string }) {
  const createSection = useCreateSection();
  const [isOpen, setIsOpen] = useState(false);
  const [name, setName] = useState("");
  const inputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (isOpen && inputRef.current) inputRef.current.focus();
  }, [isOpen]);

  function handleCreate(e: React.FormEvent) {
    e.preventDefault();
    if (!name.trim()) return;
    createSection.mutate({ area_id: areaId, name: name.trim() }, {
      onSuccess: () => { setName(""); setIsOpen(false); },
    });
  }

  if (!isOpen) {
    return (
      <button
        onClick={() => setIsOpen(true)}
        className="mt-2 flex items-center gap-1.5 py-1 text-xs text-text-muted transition-colors hover:text-accent-orange"
      >
        <FolderPlus className="h-3 w-3" />
        <span>Adicionar seção</span>
      </button>
    );
  }

  return (
    <form onSubmit={handleCreate} className="mt-2 flex items-center gap-2">
      <FolderPlus className="h-3.5 w-3.5 text-text-muted" />
      <input
        ref={inputRef}
        placeholder="Nome da seção"
        value={name}
        onChange={(e) => setName(e.target.value)}
        className="h-7 flex-1 border-0 border-b border-border bg-transparent text-sm text-text-primary placeholder:text-text-muted focus:border-accent-orange focus:outline-none"
        onKeyDown={(e) => {
          if (e.key === "Escape") { setIsOpen(false); setName(""); }
        }}
      />
      <button
        type="submit"
        disabled={!name.trim() || createSection.isPending}
        className="rounded-lg bg-accent-orange px-2.5 py-1 text-xs font-medium text-bg-primary transition-opacity hover:opacity-90 disabled:opacity-50"
      >
        Criar
      </button>
      <button
        type="button"
        onClick={() => { setIsOpen(false); setName(""); }}
        className="px-1.5 py-1 text-xs text-text-muted hover:text-text-primary"
      >
        Cancelar
      </button>
    </form>
  );
}

// ─── Filter Chip ──────────────────────────────────────

function FilterChip({
  label, count, isActive, activeClass, onClick,
}: {
  label: string;
  count?: number;
  isActive: boolean;
  activeClass?: string;
  onClick: () => void;
}) {
  return (
    <button
      onClick={onClick}
      className={cn(
        "inline-flex items-center gap-1.5 rounded-full border px-3 py-1 text-xs font-medium transition-colors",
        isActive
          ? activeClass ?? "bg-accent-orange/10 text-accent-orange border-accent-orange/20"
          : "bg-bg-card border-border text-text-muted hover:border-text-muted",
      )}
    >
      {label}
      {count !== undefined && count > 0 && (
        <span className={cn(
          "inline-flex h-4 min-w-4 items-center justify-center rounded-full px-1 text-[10px] font-semibold",
          isActive ? "bg-current/10" : "bg-border/80 text-text-muted",
        )}>
          {count}
        </span>
      )}
    </button>
  );
}

// ─── Edit Dialog ──────────────────────────────────────

function TaskEditDialog({
  task, parentId, onClose,
}: {
  task?: Task;
  parentId?: string;
  onClose: () => void;
}) {
  const create = useCreateTask();
  const update = useUpdateTask();
  const { data: areas } = useAreas();
  const [title, setTitle] = useState(task?.title ?? "");
  const [priority, setPriority] = useState<TaskPriority>(task?.priority ?? "medium");
  const [dueDate, setDueDate] = useState(task?.due_date ? task.due_date.slice(0, 16) : "");
  const [description, setDescription] = useState(task?.description ?? "");
  const [areaId, setAreaId] = useState(task?.area_id ?? "");
  const [goalId, setGoalId] = useState(task?.goal_id ?? "");
  const [sectionId, setSectionId] = useState(task?.section_id ?? "");
  const [duration, setDuration] = useState(task?.duration_minutes ? String(task.duration_minutes) : "");
  const [labelIds, setLabelIds] = useState<string[]>(task?.labels?.map((l) => l.id) ?? []);
  const [labelMgrOpen, setLabelMgrOpen] = useState(false);

  const goalParams = areaId ? { area_id: areaId } : undefined;
  const { data: goals } = useGoals(goalParams);
  const sectionParams = areaId ? { area_id: areaId } : undefined;
  const { data: sections } = useSections(sectionParams);

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    const payload: CreateTaskInput = {
      title,
      priority,
      ...(dueDate && { due_date: new Date(dueDate).toISOString() }),
      ...(description && { description }),
      ...(areaId && { area_id: areaId }),
      ...(goalId && { goal_id: goalId }),
      ...(!parentId && sectionId && { section_id: sectionId }),
      ...(parentId && { parent_id: parentId }),
      ...(duration && { duration_minutes: Number(duration) }),
      ...(labelIds.length > 0 && { label_ids: labelIds }),
    };

    if (task) {
      update.mutate({ id: task.id, ...payload }, { onSuccess: onClose });
    } else {
      create.mutate(payload, { onSuccess: onClose });
    }
  }

  return (
    <>
      <form onSubmit={handleSubmit} className="space-y-4">
        <div className="space-y-2">
          <Label>Título</Label>
          <Input value={title} onChange={(e) => setTitle(e.target.value)} required autoFocus />
        </div>

        <div className="space-y-2">
          <Label>Descrição</Label>
          <Textarea value={description} onChange={(e) => setDescription(e.target.value)} rows={2} placeholder="Opcional" />
        </div>

        <div className="grid grid-cols-2 gap-3">
          <div className="space-y-2">
            <Label>Prioridade</Label>
            <Select value={priority} onValueChange={(v) => setPriority(v as TaskPriority)}>
              <SelectTrigger><SelectValue /></SelectTrigger>
              <SelectContent>
                <SelectItem value="low">Baixa</SelectItem>
                <SelectItem value="medium">Média</SelectItem>
                <SelectItem value="high">Alta</SelectItem>
                <SelectItem value="critical">Crítica</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div className="space-y-2">
            <Label>Prazo</Label>
            <Input type="datetime-local" value={dueDate} onChange={(e) => setDueDate(e.target.value)} />
          </div>
        </div>

        <div className="grid grid-cols-2 gap-3">
          <div className="space-y-2">
            <Label>Área</Label>
            <Select value={areaId} onValueChange={(v) => { setAreaId(v); setGoalId(""); setSectionId(""); }}>
              <SelectTrigger><SelectValue placeholder="Selecionar" /></SelectTrigger>
              <SelectContent>
                {(areas ?? []).map((a) => (
                  <SelectItem key={a.id} value={a.id}>{a.name}</SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <div className="space-y-2">
            <Label>Meta</Label>
            <Select value={goalId} onValueChange={setGoalId} disabled={!areaId}>
              <SelectTrigger><SelectValue placeholder="Selecionar" /></SelectTrigger>
              <SelectContent>
                {(goals ?? []).map((g) => (
                  <SelectItem key={g.id} value={g.id}>{g.title}</SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
        </div>

        <div className="grid grid-cols-2 gap-3">
          {!parentId && (
            <div className="space-y-2">
              <Label>Seção</Label>
              <Select value={sectionId} onValueChange={setSectionId} disabled={!areaId}>
                <SelectTrigger><SelectValue placeholder="Selecionar" /></SelectTrigger>
                <SelectContent>
                  {(sections ?? []).map((s) => (
                    <SelectItem key={s.id} value={s.id}>{s.name}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
          )}
          <div className="space-y-2">
            <Label>Duração</Label>
            <div className="flex items-center gap-2">
              <Input type="number" min="1" value={duration} onChange={(e) => setDuration(e.target.value)} className="w-24" />
              <span className="text-xs text-text-muted">min</span>
            </div>
          </div>
        </div>

        <div className="space-y-2">
          <Label>Etiquetas</Label>
          <LabelPicker selectedIds={labelIds} onChange={setLabelIds} onManageClick={() => setLabelMgrOpen(true)} />
        </div>

        <Button type="submit" className="w-full" disabled={create.isPending || update.isPending}>
          {task ? "Salvar" : "Criar tarefa"}
        </Button>
      </form>

      <LabelManagerDialog open={labelMgrOpen} onOpenChange={setLabelMgrOpen} />
    </>
  );
}

// ─── Main Page ────────────────────────────────────────

export function Component() {
  const [filter, setFilter] = useState<Record<string, string>>({});
  const { data: tasks, isLoading } = useTasks(filter);
  const { data: areas } = useAreas();
  const { data: labels } = useLabels();
  const sectionParams = filter.area_id ? { area_id: filter.area_id } : undefined;
  const { data: sections } = useSections(sectionParams);

  const [dialogOpen, setDialogOpen] = useState(false);
  const [editTask, setEditTask] = useState<Task | undefined>();
  const [parentIdForNew, setParentIdForNew] = useState<string | undefined>();
  const [labelMgrOpen, setLabelMgrOpen] = useState(false);
  const [expandedParents, setExpandedParents] = useState<Set<string>>(new Set());
  const [collapsedSections, setCollapsedSections] = useState<Set<string>>(new Set());
  const [quickAddTrigger, setQuickAddTrigger] = useState(0);
  const [recentIds, setRecentIds] = useState<Set<string>>(new Set());

  const emptyState = useMemo(
    () => EMPTY_STATES[Math.floor(Math.random() * EMPTY_STATES.length)]!,
    [],
  );

  const handleCreated = useCallback((id: string) => {
    setRecentIds((prev) => new Set(prev).add(id));
    setTimeout(() => {
      setRecentIds((prev) => {
        const next = new Set(prev);
        next.delete(id);
        return next;
      });
    }, 1500);
  }, []);

  useEffect(() => {
    function handleKeyDown(e: KeyboardEvent) {
      const tag = (e.target as HTMLElement)?.tagName;
      if (tag === "INPUT" || tag === "TEXTAREA" || tag === "SELECT") return;
      if ((e.target as HTMLElement)?.isContentEditable) return;
      if (e.key === "q" || e.key === "Q") {
        e.preventDefault();
        setQuickAddTrigger((n) => n + 1);
      }
    }
    document.addEventListener("keydown", handleKeyDown);
    return () => document.removeEventListener("keydown", handleKeyDown);
  }, []);

  function openCreate(parentId?: string) {
    setEditTask(undefined);
    setParentIdForNew(parentId);
    setDialogOpen(true);
  }

  function openEdit(task: Task) {
    setEditTask(task);
    setParentIdForNew(undefined);
    setDialogOpen(true);
  }

  function closeDialog() {
    setDialogOpen(false);
    setEditTask(undefined);
    setParentIdForNew(undefined);
  }

  function toggleExpanded(id: string) {
    setExpandedParents((prev) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id); else next.add(id);
      return next;
    });
  }

  function toggleSection(id: string) {
    setCollapsedSections((prev) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id); else next.add(id);
      return next;
    });
  }

  function toggleFilter(key: string, value: string) {
    setFilter((prev) => {
      if (prev[key] === value) {
        const { [key]: _, ...rest } = prev;
        // Clear dependent filters
        if (key === "area_id") {
          const { section_id: _s, ...cleaned } = rest;
          return cleaned;
        }
        return rest;
      }
      if (key === "area_id") {
        const { section_id: _, ...rest } = prev;
        return { ...rest, [key]: value };
      }
      return { ...prev, [key]: value };
    });
  }

  const { topLevel, childrenMap, sectionGroups } = useMemo(() => {
    const all = tasks ?? [];
    const top: Task[] = [];
    const children = new Map<string, Task[]>();

    for (const t of all) {
      if (t.parent_id) {
        const list = children.get(t.parent_id) ?? [];
        list.push(t);
        children.set(t.parent_id, list);
      } else {
        top.push(t);
      }
    }

    const hasSections = top.some((t) => t.section_id);
    const showSections = hasSections && !filter.section_id;

    let groups: { id: string; name: string; tasks: Task[] }[] | null = null;
    if (showSections && sections) {
      const sectionMap = new Map(sections.map((s) => [s.id, s.name]));
      const grouped = new Map<string, Task[]>();
      const unsectioned: Task[] = [];

      for (const t of top) {
        if (t.section_id && sectionMap.has(t.section_id)) {
          const list = grouped.get(t.section_id) ?? [];
          list.push(t);
          grouped.set(t.section_id, list);
        } else {
          unsectioned.push(t);
        }
      }

      groups = sections
        .filter((s) => grouped.has(s.id))
        .map((s) => ({ id: s.id, name: s.name, tasks: grouped.get(s.id)! }));

      if (unsectioned.length > 0) {
        groups.push({ id: "__none__", name: "Sem seção", tasks: unsectioned });
      }
    }

    return { topLevel: top, childrenMap: children, sectionGroups: groups };
  }, [tasks, sections, filter.section_id]);

  const hasFilters = Object.values(filter).some(Boolean);

  const allTasks = tasks ?? [];
  const pendingCount = allTasks.filter((t) => !t.is_completed).length;
  const completedCount = allTasks.filter((t) => t.is_completed).length;

  // ─── Loading ──────────────────────────────────────

  if (isLoading) {
    return (
      <div className="space-y-1 py-4">
        {[1, 2, 3, 4, 5].map((i) => (
          <div key={i} className="flex items-center gap-3 px-2 py-2.5">
            <Skeleton className="h-[18px] w-[18px] rounded-full" />
            <Skeleton className="h-4 flex-1 rounded" />
          </div>
        ))}
      </div>
    );
  }

  // ─── Render task + children ───────────────────────

  function renderTaskWithChildren(task: Task, isSubtask = false) {
    const children = childrenMap.get(task.id) ?? [];
    const expanded = expandedParents.has(task.id);

    return (
      <div key={task.id}>
        <TaskRow
          task={task}
          childCount={children.length}
          isSubtask={isSubtask}
          isExpanded={expanded}
          onToggleExpand={() => toggleExpanded(task.id)}
          onEdit={() => openEdit(task)}
          onAddSubtask={() => openCreate(task.id)}
          isRecent={recentIds.has(task.id)}
        />
        {expanded && children.length > 0 && (
          <div>
            {children.map((child) => renderTaskWithChildren(child, true))}
          </div>
        )}
      </div>
    );
  }

  // ─── Empty state ──────────────────────────────────

  const isEmpty = (tasks ?? []).length === 0;

  return (
    <div>
      {/* ─── Header ──────────────────────────────── */}
      <div className="mb-4 flex items-center justify-between">
        <div className="flex items-center gap-3">
          <h1 className="text-xl font-semibold text-text-primary">Tarefas</h1>
          <TaskUsageBadge />
        </div>
        <button
          onClick={() => setLabelMgrOpen(true)}
          className="flex items-center gap-1.5 rounded-full border border-border px-3 py-1.5 text-xs text-text-muted transition-colors hover:border-text-muted hover:text-text-primary"
        >
          <Tags className="h-3.5 w-3.5" />
          Etiquetas
        </button>
      </div>

      {/* ─── Filter chips ────────────────────────── */}
      <div className="mb-4 flex flex-wrap items-center gap-x-2 gap-y-2">
        {/* Status */}
        <span className="text-[10px] font-semibold uppercase tracking-wider text-text-muted">Status</span>
        <FilterChip
          label="Pendentes"
          count={pendingCount}
          isActive={filter.is_completed === "false"}
          onClick={() => toggleFilter("is_completed", "false")}
        />
        <FilterChip
          label="Concluídas"
          count={completedCount}
          isActive={filter.is_completed === "true"}
          onClick={() => toggleFilter("is_completed", "true")}
        />

        {/* Areas */}
        {(areas ?? []).length > 0 && (
          <>
            <span className="mx-0.5 h-4 w-px bg-border" />
            <span className="text-[10px] font-semibold uppercase tracking-wider text-text-muted">Áreas</span>
            {(areas ?? []).map((a) => (
              <FilterChip
                key={a.id}
                label={a.name}
                isActive={filter.area_id === a.id}
                activeClass={AREA_CHIP_ACTIVE[a.color] ?? "bg-accent-orange/10 text-accent-orange border-accent-orange/20"}
                onClick={() => toggleFilter("area_id", a.id)}
              />
            ))}
          </>
        )}

        {/* Sections (when area active) */}
        {filter.area_id && (sections ?? []).length > 0 && (
          <>
            <span className="mx-0.5 h-4 w-px bg-border" />
            <span className="text-[10px] font-semibold uppercase tracking-wider text-text-muted">Seções</span>
            {(sections ?? []).map((s) => (
              <FilterChip
                key={s.id}
                label={s.name}
                isActive={filter.section_id === s.id}
                onClick={() => toggleFilter("section_id", s.id)}
              />
            ))}
          </>
        )}

        {/* Labels */}
        {(labels ?? []).length > 0 && (
          <>
            <span className="mx-0.5 h-4 w-px bg-border" />
            <span className="text-[10px] font-semibold uppercase tracking-wider text-text-muted">Etiquetas</span>
            {(labels ?? []).map((l) => (
              <FilterChip
                key={l.id}
                label={l.name}
                isActive={filter.label_id === l.id}
                activeClass={AREA_CHIP_ACTIVE[l.color] ?? undefined}
                onClick={() => toggleFilter("label_id", l.id)}
              />
            ))}
          </>
        )}

        {/* Clear filters */}
        {hasFilters && (
          <button
            onClick={() => setFilter({})}
            className="ml-1 text-xs text-text-muted underline transition-colors hover:text-text-primary"
          >
            Limpar filtros
          </button>
        )}
      </div>

      {/* ─── Task list ───────────────────────────── */}
      {isEmpty ? (
        <div className="flex flex-col items-center gap-3 py-16 text-text-muted">
          <ListTree className="h-10 w-10" />
          <div className="text-center">
            <p className="text-sm font-medium text-text-primary">{emptyState.title}</p>
            <p className="mt-1 text-xs">{emptyState.desc}</p>
            <button
              onClick={() => setQuickAddTrigger((n) => n + 1)}
              className="mt-3 text-xs font-medium text-accent-orange transition-opacity hover:opacity-80"
            >
              + Adicionar tarefa
            </button>
          </div>
        </div>
      ) : (
        <div>
          {sectionGroups ? (
            sectionGroups.map((group, idx) => {
              const collapsed = collapsedSections.has(group.id);
              return (
                <div key={group.id} className="mb-2">
                  {/* Section header */}
                  <button
                    onClick={() => toggleSection(group.id)}
                    className="flex w-full items-center gap-2 border-b border-border/40 py-3 text-left"
                  >
                    {collapsed
                      ? <ChevronRight className="h-4 w-4 text-text-muted" />
                      : <ChevronDown className="h-4 w-4 text-text-muted" />
                    }
                    <span className="text-sm font-semibold text-text-primary">{group.name}</span>
                    <span className="font-mono text-xs text-text-muted">({group.tasks.length})</span>
                  </button>

                  {/* Section tasks */}
                  {!collapsed && (
                    <div>
                      {group.tasks.map((task) => renderTaskWithChildren(task))}
                      <InlineQuickAdd
                        sectionId={group.id !== "__none__" ? group.id : undefined}
                        onExpandedCreate={() => openCreate()}
                        onCreated={handleCreated}
                        trigger={idx === 0 ? quickAddTrigger : undefined}
                      />
                    </div>
                  )}
                </div>
              );
            })
          ) : (
            <div>
              {topLevel.map((task) => renderTaskWithChildren(task))}
              <InlineQuickAdd
                onExpandedCreate={() => openCreate()}
                onCreated={handleCreated}
                trigger={quickAddTrigger}
              />
            </div>
          )}
        </div>
      )}

      {/* Section create */}
      {filter.area_id && <SectionCreateInline areaId={filter.area_id} />}

      {/* ─── Edit Dialog ─────────────────────────── */}
      <Dialog open={dialogOpen} onOpenChange={(v) => { if (!v) closeDialog(); else setDialogOpen(true); }}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>
              {editTask ? "Editar tarefa" : parentIdForNew ? "Nova sub-tarefa" : "Nova tarefa"}
            </DialogTitle>
          </DialogHeader>
          <TaskEditDialog task={editTask} parentId={parentIdForNew} onClose={closeDialog} />
        </DialogContent>
      </Dialog>

      {/* Label Manager */}
      <LabelManagerDialog open={labelMgrOpen} onOpenChange={setLabelMgrOpen} />
    </div>
  );
}
