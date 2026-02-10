import { useState, useMemo, useRef, useEffect } from "react";
import {
  Plus, Star, ChevronDown, ChevronRight, ListTree, Clock,
  RotateCcw, Tags, FolderPlus, Check, MoreHorizontal, CalendarDays,
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
import { cn } from "@/lib/cn";
import type { Task, TaskPriority } from "@/types/api";

// ─── Helpers ──────────────────────────────────────────

const PRIORITY_COLORS: Record<TaskPriority, string> = {
  critical: "border-accent-rose",
  high: "border-accent-orange",
  medium: "border-accent-blue",
  low: "border-border",
};

const PRIORITY_BG: Record<TaskPriority, string> = {
  critical: "bg-accent-rose/15",
  high: "bg-accent-orange/15",
  medium: "bg-accent-blue/15",
  low: "bg-border/15",
};

const LABEL_DOT_COLOR: Record<string, string> = {
  orange: "bg-accent-orange",
  blue: "bg-accent-blue",
  green: "bg-accent-green",
  rose: "bg-accent-rose",
  sand: "bg-accent-sand",
  sage: "bg-accent-sage",
};

function formatDuration(mins: number): string {
  if (mins < 60) return `${mins}min`;
  const h = Math.floor(mins / 60);
  const m = mins % 60;
  return m ? `${h}h${m}` : `${h}h`;
}

function formatDueDate(date: string): { text: string; className: string } {
  const d = new Date(date);
  if (isToday(d)) return { text: "Hoje", className: "text-accent-green" };
  if (isTomorrow(d)) return { text: "Amanha", className: "text-accent-orange" };
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
      tokens.push({ type: "date", raw: am[0], label: "Amanha" });
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
        className="group/check flex h-[18px] w-[18px] shrink-0 items-center justify-center rounded-full border-2 border-accent-green bg-accent-green transition-all duration-150"
      >
        <Check className="h-2.5 w-2.5 text-bg-primary" strokeWidth={3} />
      </button>
    );
  }

  return (
    <button
      onClick={onComplete}
      className={cn(
        "flex h-[18px] w-[18px] shrink-0 items-center justify-center rounded-full border-2 transition-all duration-150",
        PRIORITY_COLORS[priority],
        `hover:${PRIORITY_BG[priority]}`,
      )}
      style={{
        // Inline hover bg since dynamic Tailwind classes are unreliable
      }}
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

function InlineQuickAdd({
  sectionId, parentId, onExpandedCreate, trigger,
}: {
  sectionId?: string;
  parentId?: string;
  onExpandedCreate?: () => void;
  trigger?: number;
}) {
  const [isOpen, setIsOpen] = useState(false);
  const [title, setTitle] = useState("");
  const [priority, setPriority] = useState<TaskPriority>("medium");
  const inputRef = useRef<HTMLInputElement>(null);
  const wrapperRef = useRef<HTMLDivElement>(null);
  const createTask = useCreateTask();

  useEffect(() => {
    if (isOpen && inputRef.current) {
      inputRef.current.focus();
    }
  }, [isOpen]);

  useEffect(() => {
    if (!isOpen) return;
    function handleClickOutside(e: MouseEvent) {
      if (wrapperRef.current && !wrapperRef.current.contains(e.target as Node)) {
        setIsOpen(false);
        setTitle("");
        setPriority("medium");
      }
    }
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, [isOpen]);

  useEffect(() => {
    if (trigger && trigger > 0) setIsOpen(true);
  }, [trigger]);

  const parsed = parseSmartInput(title);

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!title.trim()) return;
    const payload: CreateTaskInput = {
      title: parsed.cleanTitle || title.trim(),
      priority: parsed.priority ?? priority,
      ...(parsed.dueDate && { due_date: parsed.dueDate }),
      ...(sectionId && { section_id: sectionId }),
      ...(parentId && { parent_id: parentId }),
    };
    createTask.mutate(payload, {
      onSuccess: () => {
        setTitle("");
        setPriority("medium");
        inputRef.current?.focus();
      },
    });
  }

  function handleCancel() {
    setIsOpen(false);
    setTitle("");
    setPriority("medium");
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
        <span>Adicionar tarefa</span>
      </button>
    );
  }

  const priorities: TaskPriority[] = ["critical", "high", "medium", "low"];
  const priorityLabels: Record<TaskPriority, string> = {
    critical: "P1", high: "P2", medium: "P3", low: "P4",
  };

  return (
    <div ref={wrapperRef} className={cn("my-1", parentId && "pl-8")}>
      <form onSubmit={handleSubmit} className="rounded-[14px] border border-border bg-bg-card p-3">
        <input
          ref={inputRef}
          type="text"
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          placeholder="Nome da tarefa"
          autoComplete="off"
          className="mb-2 w-full border-0 bg-transparent text-sm text-text-primary placeholder:text-text-muted focus:outline-none"
          onKeyDown={(e) => {
            if (e.key === "Escape") handleCancel();
          }}
        />
        {parsed.tokens.length > 0 && (
          <div className="mb-2 flex items-center gap-1.5">
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
        )}
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-1.5">
            {priorities.map((p) => (
              <button
                key={p}
                type="button"
                onClick={() => setPriority(p)}
                className={cn(
                  "flex h-6 w-6 items-center justify-center rounded-full border-2 text-[10px] font-semibold transition-all duration-150",
                  PRIORITY_COLORS[p],
                  priority === p ? cn(PRIORITY_BG[p], "text-text-primary") : "text-text-muted",
                )}
              >
                {priorityLabels[p]}
              </button>
            ))}
            {onExpandedCreate && (
              <button
                type="button"
                onClick={() => { handleCancel(); onExpandedCreate(); }}
                className="ml-1 flex h-6 w-6 items-center justify-center rounded-full text-text-muted transition-colors hover:text-text-primary"
                title="Mais opcoes"
              >
                <MoreHorizontal className="h-3.5 w-3.5" />
              </button>
            )}
          </div>
          <div className="flex items-center gap-2">
            <button
              type="button"
              onClick={handleCancel}
              className="px-2 py-1 text-xs text-text-muted transition-colors hover:text-text-primary"
            >
              Cancelar
            </button>
            <button
              type="submit"
              disabled={!title.trim() || createTask.isPending}
              className="rounded-lg bg-accent-orange px-3 py-1 text-xs font-medium text-bg-primary transition-opacity hover:opacity-90 disabled:opacity-50"
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
  task, childCount, isSubtask, isExpanded, onToggleExpand, onEdit, onAddSubtask,
}: {
  task: Task;
  childCount: number;
  isSubtask?: boolean;
  isExpanded?: boolean;
  onToggleExpand?: () => void;
  onEdit: () => void;
  onAddSubtask?: () => void;
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

  return (
    <div
      className={cn(
        "group/row flex items-center gap-2.5 border-b border-border/40 px-2 py-2.5 transition-colors hover:rounded-lg hover:bg-bg-card/60",
        isSubtask && "pl-8",
      )}
    >
      {/* Priority checkbox */}
      <PriorityCheckbox
        priority={task.priority}
        isCompleted={task.is_completed}
        onComplete={() => completeTask.mutate(task.id)}
        onReopen={() => reopenTask.mutate(task.id)}
      />

      {/* Title + inline indicators */}
      <div className="flex min-w-0 flex-1 cursor-pointer items-center gap-2" onClick={onEdit}>
        <span className={cn("truncate text-sm text-text-primary", task.is_completed && "text-text-muted line-through")}>
          {task.title}
        </span>

        {/* Label dots */}
        {(task.labels ?? []).length > 0 && (
          <span className="flex items-center gap-0.5">
            {(task.labels ?? []).slice(0, 4).map((l) => (
              <span
                key={l.id}
                className={cn("h-1.5 w-1.5 rounded-full", LABEL_DOT_COLOR[l.color] ?? "bg-text-muted")}
                title={l.name}
              />
            ))}
          </span>
        )}
      </div>

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

      {/* Due date */}
      {dueInfo && (
        <span className={cn("flex items-center gap-1 whitespace-nowrap text-xs", dueInfo.className)}>
          <CalendarDays className="h-3 w-3" />
          {dueInfo.text}
        </span>
      )}

      {/* Duration */}
      {task.duration_minutes && (
        <span className="flex items-center gap-1 whitespace-nowrap font-mono text-xs text-text-muted">
          <Clock className="h-3 w-3" />
          {formatDuration(task.duration_minutes)}
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

      {/* Hover-revealed actions */}
      <div className="flex items-center gap-0.5">
        {/* Focus star — always visible when active, hover-only when inactive */}
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

        {/* Schedule placeholder */}
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
        <span>Adicionar secao</span>
      </button>
    );
  }

  return (
    <form onSubmit={handleCreate} className="mt-2 flex items-center gap-2">
      <FolderPlus className="h-3.5 w-3.5 text-text-muted" />
      <input
        ref={inputRef}
        placeholder="Nome da secao"
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
          <Label>Titulo</Label>
          <Input value={title} onChange={(e) => setTitle(e.target.value)} required autoFocus />
        </div>

        <div className="space-y-2">
          <Label>Descricao</Label>
          <Textarea value={description} onChange={(e) => setDescription(e.target.value)} rows={2} placeholder="Opcional" />
        </div>

        <div className="grid grid-cols-2 gap-3">
          <div className="space-y-2">
            <Label>Prioridade</Label>
            <Select value={priority} onValueChange={(v) => setPriority(v as TaskPriority)}>
              <SelectTrigger><SelectValue /></SelectTrigger>
              <SelectContent>
                <SelectItem value="low">Baixa</SelectItem>
                <SelectItem value="medium">Media</SelectItem>
                <SelectItem value="high">Alta</SelectItem>
                <SelectItem value="critical">Critica</SelectItem>
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
            <Label>Area</Label>
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
              <Label>Secao</Label>
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
            <Label>Duracao</Label>
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
        groups.push({ id: "__none__", name: "Sem secao", tasks: unsectioned });
      }
    }

    return { topLevel: top, childrenMap: children, sectionGroups: groups };
  }, [tasks, sections, filter.section_id]);

  const hasFilters = Object.values(filter).some(Boolean);

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
        <h1 className="text-xl font-semibold text-text-primary">Tarefas</h1>
        <div className="flex items-center gap-2">
          <button
            onClick={() => setLabelMgrOpen(true)}
            className="flex items-center gap-1.5 rounded-full border border-border px-3 py-1.5 text-xs text-text-muted transition-colors hover:border-text-muted hover:text-text-primary"
          >
            <Tags className="h-3.5 w-3.5" />
            Etiquetas
          </button>
          <button
            onClick={() => openCreate()}
            className="flex items-center gap-1.5 rounded-full bg-accent-orange px-3 py-1.5 text-xs font-medium text-bg-primary transition-opacity hover:opacity-90"
          >
            <Plus className="h-3.5 w-3.5" />
            Nova tarefa
          </button>
        </div>
      </div>

      {/* ─── Filter bar ──────────────────────────── */}
      <div className="mb-4 flex flex-wrap items-center gap-2">
        <Select value={filter.is_completed ?? ""} onValueChange={(v) => setFilter((p) => ({ ...p, is_completed: v }))}>
          <SelectTrigger className="h-8 w-28 rounded-full text-xs">
            <SelectValue placeholder="Status" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="false">Pendentes</SelectItem>
            <SelectItem value="true">Concluidas</SelectItem>
          </SelectContent>
        </Select>

        <Select
          value={filter.area_id ?? ""}
          onValueChange={(v) => setFilter((p) => {
            const { section_id: _, ...rest } = p;
            return { ...rest, area_id: v };
          })}
        >
          <SelectTrigger className="h-8 w-28 rounded-full text-xs">
            <SelectValue placeholder="Area" />
          </SelectTrigger>
          <SelectContent>
            {(areas ?? []).map((a) => (
              <SelectItem key={a.id} value={a.id}>{a.name}</SelectItem>
            ))}
          </SelectContent>
        </Select>

        {filter.area_id && (
          <Select value={filter.section_id ?? ""} onValueChange={(v) => setFilter((p) => ({ ...p, section_id: v }))}>
            <SelectTrigger className="h-8 w-28 rounded-full text-xs">
              <SelectValue placeholder="Secao" />
            </SelectTrigger>
            <SelectContent>
              {(sections ?? []).map((s) => (
                <SelectItem key={s.id} value={s.id}>{s.name}</SelectItem>
              ))}
            </SelectContent>
          </Select>
        )}

        <Select value={filter.label_id ?? ""} onValueChange={(v) => setFilter((p) => ({ ...p, label_id: v }))}>
          <SelectTrigger className="h-8 w-28 rounded-full text-xs">
            <SelectValue placeholder="Etiqueta" />
          </SelectTrigger>
          <SelectContent>
            {(labels ?? []).map((l) => (
              <SelectItem key={l.id} value={l.id}>{l.name}</SelectItem>
            ))}
          </SelectContent>
        </Select>

        {hasFilters && (
          <button
            onClick={() => setFilter({})}
            className="px-2 text-xs text-text-muted transition-colors hover:text-text-primary"
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
            <p className="text-sm font-medium">Nenhuma tarefa</p>
            <p className="mt-1 text-xs">Crie sua primeira tarefa para comecar</p>
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
              <InlineQuickAdd onExpandedCreate={() => openCreate()} trigger={quickAddTrigger} />
            </div>
          )}
        </div>
      )}

      {/* Inline quick-add for empty state too */}
      {isEmpty && (
        <InlineQuickAdd onExpandedCreate={() => openCreate()} trigger={quickAddTrigger} />
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
