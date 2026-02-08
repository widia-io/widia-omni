import { useState } from "react";
import { Link } from "react-router";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useLogin } from "@/hooks/use-login";

export function Component() {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const login = useLogin();

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    login.mutate({ email, password });
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-4 rounded-[14px] border border-border bg-bg-card p-6">
      <div className="mb-6 text-center">
        <h1 className="text-xl font-bold text-text-primary">Entrar</h1>
        <p className="mt-1 text-sm text-text-secondary font-serif">Acesse sua conta</p>
      </div>

      <div className="space-y-2">
        <Label htmlFor="email">Email</Label>
        <Input id="email" type="email" placeholder="seu@email.com" value={email} onChange={(e) => setEmail(e.target.value)} required />
      </div>

      <div className="space-y-2">
        <Label htmlFor="password">Senha</Label>
        <Input id="password" type="password" placeholder="********" value={password} onChange={(e) => setPassword(e.target.value)} required />
      </div>

      {login.error && (
        <p className="text-sm text-accent-rose">{login.error.message}</p>
      )}

      <Button type="submit" className="w-full" disabled={login.isPending}>
        {login.isPending ? "Entrando..." : "Entrar"}
      </Button>

      <div className="flex items-center justify-between text-sm">
        <Link to="/forgot-password" className="text-text-muted hover:text-accent-orange transition-colors">
          Esqueceu a senha?
        </Link>
        <Link to="/register" className="text-accent-orange hover:text-accent-orange/80 transition-colors">
          Criar conta
        </Link>
      </div>
    </form>
  );
}
