import { useState } from "react";
import { Link, Navigate, useSearchParams } from "react-router";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useRegister } from "@/hooks/use-register";
import { useAuthStore } from "@/stores/auth-store";

export function Component() {
  const accessToken = useAuthStore((s) => s.accessToken);
  const [searchParams] = useSearchParams();
  const [name, setName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const register = useRegister();
  const referralCode = searchParams.get("ref") ?? undefined;

  if (accessToken) return <Navigate to="/dashboard" replace />;

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    register.mutate({ email, password, data: { display_name: name, referral_code: referralCode } });
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-4 rounded-[14px] border border-border bg-bg-card p-6">
      <div className="mb-6 text-center">
        <h1 className="text-xl font-bold text-text-primary">Criar conta</h1>
        <p className="mt-1 text-sm text-text-secondary font-serif">Comece a gerenciar sua vida</p>
      </div>

      <div className="space-y-2">
        <Label htmlFor="name">Nome</Label>
        <Input id="name" placeholder="Seu nome" value={name} onChange={(e) => setName(e.target.value)} required />
      </div>

      <div className="space-y-2">
        <Label htmlFor="email">Email</Label>
        <Input id="email" type="email" placeholder="seu@email.com" value={email} onChange={(e) => setEmail(e.target.value)} required />
      </div>

      <div className="space-y-2">
        <Label htmlFor="password">Senha</Label>
        <Input id="password" type="password" placeholder="Min. 8 caracteres" value={password} onChange={(e) => setPassword(e.target.value)} required minLength={8} />
      </div>

      {register.error && (
        <p className="text-sm text-accent-rose">{register.error.message}</p>
      )}

      <Button type="submit" className="w-full" disabled={register.isPending}>
        {register.isPending ? "Criando conta..." : "Criar conta"}
      </Button>

      <p className="text-center text-sm text-text-muted">
        Ja tem conta?{" "}
        <Link to="/login" className="text-accent-orange hover:text-accent-orange/80 transition-colors">
          Voltar ao login
        </Link>
      </p>
    </form>
  );
}
