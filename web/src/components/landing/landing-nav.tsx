import { Link } from "react-router";
import { Button } from "@/components/ui/button";
import { Logo } from "@/components/logo";

export function LandingNav() {
  return (
    <nav className="fixed top-0 left-0 right-0 z-50 border-b border-border/50 bg-bg-primary/80 backdrop-blur-sm">
      <div className="mx-auto flex h-16 max-w-6xl items-center justify-between px-6">
        <Link to="/" className="flex items-center gap-2">
          <Logo size={32} />
          <span className="text-lg font-bold text-text-primary">Widia Omni</span>
        </Link>
        <div className="hidden items-center gap-6 md:flex">
          <a href="#features" className="text-sm text-text-secondary transition-colors hover:text-text-primary">Funcionalidades</a>
          <a href="#pricing" className="text-sm text-text-secondary transition-colors hover:text-text-primary">Planos</a>
        </div>
        <div className="flex items-center gap-3">
          <Button variant="ghost" size="sm" asChild>
            <Link to="/login">Entrar</Link>
          </Button>
          <Button size="sm" asChild>
            <Link to="/register">Comecar gratis</Link>
          </Button>
        </div>
      </div>
    </nav>
  );
}
