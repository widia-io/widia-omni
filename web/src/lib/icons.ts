import {
  Heart, Briefcase, DollarSign, Users, Book, Sun,
  Dumbbell, Brain, Home, Gamepad2, Sparkles,
  type LucideIcon,
} from "lucide-react";

export const areaIconMap: Record<string, LucideIcon> = {
  heart: Heart,
  briefcase: Briefcase,
  "dollar-sign": DollarSign,
  users: Users,
  book: Book,
  sun: Sun,
  dumbbell: Dumbbell,
  brain: Brain,
  home: Home,
  "gamepad-2": Gamepad2,
  sparkles: Sparkles,
};

export type AreaIconContext = {
  icon?: string | null;
  slug?: string | null;
  name?: string | null;
  color?: string | null;
};

const iconAliases: Record<string, string> = {
  dollar: "dollar-sign",
  dollar_sign: "dollar-sign",
  dollarsign: "dollar-sign",
  gamepad: "gamepad-2",
  gamepad2: "gamepad-2",
  gamepad_2: "gamepad-2",
};

export function resolveAreaIconKey(icon: string | null | undefined): string | null {
  const normalized = (icon ?? "").trim().toLowerCase().replace(/[\s_]+/g, "-");
  if (!normalized) return null;
  if (normalized in areaIconMap) return normalized;

  const alias = iconAliases[normalized];
  if (alias && alias in areaIconMap) return alias;

  return null;
}

export function getAreaIcon(icon: string | null | undefined): LucideIcon | null {
  const key = resolveAreaIconKey(icon);
  return key ? (areaIconMap[key] ?? null) : null;
}

export function isRawAreaIcon(icon: string | null | undefined): boolean {
  const value = (icon ?? "").trim();
  if (!value) return false;
  return !/^[a-z0-9-]+$/i.test(value);
}

function inferIconFromText(text: string): string | null {
  if (/(saude|saúde|health|fitness|fit|wellness|gym|exerc|corpo)/i.test(text)) return "dumbbell";
  if (/(carreira|career|trabalho|work|job|business|profiss)/i.test(text)) return "briefcase";
  if (/(finan|money|dinheiro|invest|budget|orcamento|orçamento)/i.test(text)) return "dollar-sign";
  if (/(relacion|famil|family|friend|social|people|pessoa)/i.test(text)) return "users";
  if (/(estud|study|learn|book|read|educ|growth|desenvolvimento|pessoal)/i.test(text)) return "brain";
  if (/(casa|home|house|lar|ambiente)/i.test(text)) return "home";
  if (/(lazer|fun|game|hobby|leisure|divers)/i.test(text)) return "gamepad-2";
  if (/(espirit|spirit|faith|alma|mindful|gratidao|gratidão)/i.test(text)) return "sparkles";
  return null;
}

function inferIconFromColor(color: string | null | undefined): string {
  const normalized = (color ?? "").toLowerCase();
  if (normalized === "green") return "dumbbell";
  if (normalized === "orange") return "briefcase";
  if (normalized === "sand") return "dollar-sign";
  if (normalized === "rose") return "users";
  if (normalized === "blue") return "home";
  if (normalized === "sage") return "brain";
  return "heart";
}

export function getAreaIconKeyWithFallback(area?: AreaIconContext | null): string {
  const explicit = resolveAreaIconKey(area?.icon);
  if (explicit) return explicit;

  const text = `${area?.slug ?? ""} ${area?.name ?? ""}`.trim();
  const byText = inferIconFromText(text);
  if (byText) return byText;

  return inferIconFromColor(area?.color);
}

export function getAreaIconWithFallback(area?: AreaIconContext | null): LucideIcon {
  const key = getAreaIconKeyWithFallback(area);
  return areaIconMap[key] ?? Heart;
}
