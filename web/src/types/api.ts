// ─── Auth ────────────────────────────────────────────
export interface AuthResponse {
  access_token: string;
  refresh_token: string;
  user: { id: string; email: string };
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface RegisterRequest {
  email: string;
  password: string;
  data?: { display_name?: string };
}

// ─── User ────────────────────────────────────────────
export interface UserProfile {
  id: string;
  display_name: string;
  email: string;
  avatar_url: string | null;
  timezone: string;
  locale: string;
  default_workspace_id: string | null;
  onboarding_completed: boolean;
  created_at: string;
  updated_at: string;
}

export interface UserPreferences {
  user_id: string;
  week_starts_on: number;
  daily_focus_limit: number;
  notification_email: boolean;
  notification_push: boolean;
  weekly_review_day: number;
  weekly_review_time: string;
  theme: string;
  currency: string;
  score_weights: Record<string, number> | null;
  created_at: string;
  updated_at: string;
}

// ─── Workspace ───────────────────────────────────────
export interface Workspace {
  id: string;
  name: string;
  slug: string;
  owner_id: string;
  created_at: string;
  updated_at: string;
}

export interface WorkspaceUsage {
  counters: WorkspaceCounters;
  limits: EntitlementLimits;
}

export interface WorkspaceCounters {
  areas_count: number;
  goals_count: number;
  habits_count: number;
  tasks_created_today: number;
  transactions_month_count: number;
  storage_bytes_used: number;
}

export interface EntitlementLimits {
  max_areas: number;
  max_goals: number;
  max_habits: number;
  max_tasks_per_day: number;
  max_transactions_per_month: number;
  journal_enabled: boolean;
  finance_enabled: boolean;
  export_enabled: boolean;
  score_history_weeks: number;
  api_rate_limit_per_minute: number;
  storage_mb: number;
  ai_insights: boolean;
  api_access: boolean;
}

// ─── Areas ───────────────────────────────────────────
export interface LifeArea {
  id: string;
  workspace_id: string;
  name: string;
  slug: string;
  icon: string;
  color: string;
  weight: number;
  sort_order: number;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

// ─── Goals ───────────────────────────────────────────
export type GoalStatus = "not_started" | "on_track" | "at_risk" | "behind" | "completed" | "cancelled";
export type GoalPeriod = "yearly" | "quarterly" | "monthly" | "weekly";

export interface Goal {
  id: string;
  workspace_id: string;
  area_id: string | null;
  parent_id: string | null;
  title: string;
  description: string | null;
  period: GoalPeriod;
  status: GoalStatus;
  target_value: number | null;
  current_value: number;
  unit: string | null;
  start_date: string;
  end_date: string;
  completed_at: string | null;
  created_at: string;
  updated_at: string;
}

// ─── Habits ──────────────────────────────────────────
export type HabitFrequency = "daily" | "weekly" | "custom";

export interface Habit {
  id: string;
  workspace_id: string;
  area_id: string | null;
  name: string;
  color: string;
  frequency: HabitFrequency;
  target_per_week: number;
  is_active: boolean;
  sort_order: number;
  created_at: string;
  updated_at: string;
}

export interface HabitEntry {
  id: string;
  habit_id: string;
  workspace_id: string;
  date: string;
  intensity: number;
  notes: string | null;
  created_at: string;
}

export interface HabitStreak {
  habit_id: string;
  name: string;
  current_streak: number;
  longest_streak: number;
  last_check_in: string | null;
}

// ─── Labels ─────────────────────────────────────────
export interface Label {
  id: string;
  workspace_id: string;
  name: string;
  color: string;
  position: number;
  created_at: string;
  updated_at: string;
}

// ─── Sections ───────────────────────────────────────
export interface Section {
  id: string;
  workspace_id: string;
  area_id: string;
  name: string;
  position: number;
  created_at: string;
  updated_at: string;
}

// ─── Tasks ───────────────────────────────────────────
export type TaskPriority = "low" | "medium" | "high" | "critical";

export interface Task {
  id: string;
  workspace_id: string;
  area_id: string | null;
  goal_id: string | null;
  parent_id: string | null;
  section_id: string | null;
  title: string;
  description: string | null;
  priority: TaskPriority;
  position: number;
  duration_minutes: number | null;
  is_completed: boolean;
  is_focus: boolean;
  due_date: string | null;
  completed_at: string | null;
  labels: Label[];
  created_at: string;
  updated_at: string;
}

// ─── Journal ─────────────────────────────────────────
export interface JournalEntry {
  id: string;
  workspace_id: string;
  date: string;
  mood: number | null;
  energy: number | null;
  wins: string[];
  challenges: string[];
  gratitude: string[];
  notes: string | null;
  tags: string[];
  created_at: string;
  updated_at: string;
}

// ─── Scores ──────────────────────────────────────────
export interface LifeScore {
  id: string;
  workspace_id: string;
  score: number;
  week_start: string;
  area_scores: Record<string, number>;
  created_at: string;
}

export interface AreaScore {
  id: string;
  workspace_id: string;
  area_id: string;
  score: number;
  week_start: string;
  breakdown: { habits_score: number; goals_score: number; tasks_score: number } | null;
  created_at: string;
}

export interface ScoreHistory {
  life_scores: LifeScore[];
  area_scores: AreaScore[];
}

// ─── Notifications ───────────────────────────────────
export type NotificationType =
  | "weekly_review" | "streak_at_risk" | "goal_deadline"
  | "trial_ending" | "plan_changed" | "score_update"
  | "habit_reminder" | "system";

export interface Notification {
  id: string;
  workspace_id: string;
  user_id: string;
  type: NotificationType;
  channel: "in_app" | "email" | "push";
  title: string;
  body: string | null;
  data: Record<string, unknown> | null;
  is_read: boolean;
  read_at: string | null;
  created_at: string;
}

// ─── Billing ─────────────────────────────────────────
export type PlanTier = "free" | "pro" | "premium";
export type SubscriptionStatus = "trialing" | "active" | "past_due" | "canceled" | "paused" | "unpaid";

export interface Plan {
  id: string;
  tier: PlanTier;
  name: string;
  price_monthly: number;
  price_yearly: number;
  limits: EntitlementLimits;
  is_active: boolean;
}

export interface Subscription {
  id: string;
  workspace_id: string;
  plan_id: string;
  tier: PlanTier;
  status: SubscriptionStatus;
  currency: string;
  current_period_start: string | null;
  current_period_end: string | null;
  trial_end: string | null;
  created_at: string;
  updated_at: string;
}

// ─── Finance ─────────────────────────────────────────
export type TransactionType = "income" | "expense" | "investment" | "transfer";

export interface FinanceCategory {
  id: string;
  workspace_id: string;
  name: string;
  type: TransactionType;
  color: string | null;
  icon: string | null;
  parent_id: string | null;
  is_system: boolean;
  created_at: string;
}

export interface Transaction {
  id: string;
  workspace_id: string;
  category_id: string | null;
  area_id: string | null;
  type: TransactionType;
  amount: number;
  description: string | null;
  date: string;
  is_recurring: boolean;
  recurrence_rule: string | null;
  tags: string[];
  created_at: string;
}

export interface Budget {
  id: string;
  workspace_id: string;
  category_id: string | null;
  month: string;
  amount: number;
  created_at: string;
  updated_at: string;
}

export interface FinanceSummary {
  month: string;
  total_income: number;
  total_expenses: number;
  total_investments: number;
  net_balance: number;
  by_category: Array<{
    category_id: string | null;
    category_name: string | null;
    type: string;
    total: number;
    count: number;
  }>;
  budget_status: Array<{
    category_id: string | null;
    category_name: string | null;
    budget_amount: number;
    spent_amount: number;
    remaining: number;
    percentage: number;
  }>;
}

// ─── Dashboard ───────────────────────────────────────
export interface DashboardData {
  areas_count: number;
  active_goals: number;
  today_tasks: number;
  completed_today: number;
  habits_today: number;
  current_streaks: number;
  life_score: number | null;
  journal_today: boolean;
  unread_notifications: number;
}

// ─── Onboarding ──────────────────────────────────────
export interface OnboardingStatus {
  completed: boolean;
  steps: {
    areas: boolean;
    goals: boolean;
    habits: boolean;
  };
}
