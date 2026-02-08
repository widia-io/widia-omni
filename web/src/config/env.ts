export const env = {
  apiUrl: import.meta.env.VITE_API_URL as string,
  supabaseUrl: import.meta.env.VITE_SUPABASE_URL as string,
} as const;
