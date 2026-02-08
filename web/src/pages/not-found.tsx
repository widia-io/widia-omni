import { Link } from "react-router";

export function NotFoundPage() {
  return (
    <div className="flex min-h-screen flex-col items-center justify-center gap-4 bg-bg-primary">
      <h1 className="font-mono text-6xl font-bold text-accent-orange">404</h1>
      <p className="text-text-secondary">Pagina nao encontrada</p>
      <Link
        to="/"
        className="mt-4 rounded-lg bg-accent-orange px-6 py-2 text-sm font-medium text-bg-primary transition-colors hover:bg-accent-orange/90"
      >
        Voltar ao inicio
      </Link>
    </div>
  );
}
