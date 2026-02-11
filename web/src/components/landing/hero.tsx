import { Link } from "react-router";
import { Button } from "@/components/ui/button";
import { ArrowRight } from "lucide-react";

export function Hero() {
  return (
    <section className="relative flex min-h-[70vh] flex-col items-center justify-center px-6 pt-16 text-center">
      <div className="absolute inset-0 bg-[radial-gradient(ellipse_at_center,_var(--color-accent-orange)/8%_0%,transparent_70%)]" />
      <div className="relative z-10">
        <div className="mb-4 inline-flex items-center gap-2 rounded-full border border-border bg-bg-card px-4 py-1.5 text-xs text-text-secondary">
          <span className="h-1.5 w-1.5 rounded-full bg-accent-green animate-pulse" />
          Novo: módulo financeiro disponível
        </div>
        <h1 className="mx-auto max-w-3xl text-4xl font-bold leading-tight tracking-tight md:text-6xl">
          Gerencie sua vida{" "}
          <span className="bg-gradient-to-r from-accent-orange to-accent-sand bg-clip-text text-transparent">
            com propósito
          </span>
        </h1>
        <p className="mx-auto mt-5 max-w-xl font-serif text-lg text-text-secondary">
          Metas, hábitos, tarefas, finanças e journaling — tudo em uma única plataforma
          desenhada para sua evolução pessoal.
        </p>
        <div className="mt-8 flex items-center justify-center gap-4">
          <Button size="lg" asChild>
            <Link to="/register">
              Começar grátis <ArrowRight className="h-4 w-4" />
            </Link>
          </Button>
          <Button variant="outline" size="lg" asChild>
            <a href="#pricing">Ver planos</a>
          </Button>
        </div>
      </div>
    </section>
  );
}
