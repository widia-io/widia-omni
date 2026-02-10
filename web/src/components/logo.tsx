import { useUIStore } from "@/stores/ui-store";

interface LogoProps {
  size?: number;
  className?: string;
}

export function Logo({ size = 32, className }: LogoProps) {
  const theme = useUIStore((s) => s.theme);
  const isDark =
    theme === "dark" ||
    (theme === "system" && window.matchMedia("(prefers-color-scheme: dark)").matches);

  const bg = isDark ? "#F0874A" : "#E8580C";
  const fg = isDark ? "#0F0F17" : "#FFFFFF";
  const sw = size <= 20 ? 4 : size <= 32 ? 3.5 : 3;
  const cr = size <= 20 ? 5 : size <= 32 ? 4.5 : 4;

  return (
    <svg width={size} height={size} viewBox="0 0 42 42" fill="none" className={className}>
      <rect width="42" height="42" rx="10" fill={bg} />
      <path d="M13 29A11.5 11.5 0 0 1 21 9.5" stroke={fg} strokeOpacity={0.35} strokeWidth={sw} strokeLinecap="round" />
      <path d="M21 9.5A11.5 11.5 0 0 1 32.5 21" stroke={fg} strokeOpacity={0.6} strokeWidth={sw} strokeLinecap="round" />
      <path d="M32.5 21A11.5 11.5 0 0 1 29 29" stroke={fg} strokeWidth={sw} strokeLinecap="round" />
      <circle cx="21" cy="22" r={cr} fill={fg} />
    </svg>
  );
}
