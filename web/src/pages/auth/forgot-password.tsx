import { useState } from "react";
import { Link } from "react-router";
import { useMutation } from "@tanstack/react-query";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { api } from "@/lib/api-client";

export function Component() {
  const [email, setEmail] = useState("");
  const [sent, setSent] = useState(false);

  const mutation = useMutation({
    mutationFn: (email: string) =>
      api("/auth/forgot-password", { method: "POST", body: JSON.stringify({ email }) }),
    onSuccess: () => setSent(true),
  });

  if (sent) {
    return (
      <div className="rounded-[14px] border border-border bg-bg-card p-6 text-center">
        <h1 className="text-xl font-bold text-text-primary mb-2">Email enviado</h1>
        <p className="text-sm text-text-secondary font-serif mb-6">
          Verifique sua caixa de entrada para redefinir a senha.
        </p>
        <Link to="/login" className="text-sm text-accent-orange hover:text-accent-orange/80 transition-colors">
          Voltar ao login
        </Link>
      </div>
    );
  }

  return (
    <form onSubmit={(e) => { e.preventDefault(); mutation.mutate(email); }} className="space-y-4 rounded-[14px] border border-border bg-bg-card p-6">
      <div className="mb-6 text-center">
        <h1 className="text-xl font-bold text-text-primary">Esqueceu a senha?</h1>
        <p className="mt-1 text-sm text-text-secondary font-serif">Informe seu email para recuperar</p>
      </div>

      <div className="space-y-2">
        <Label htmlFor="email">Email</Label>
        <Input id="email" type="email" placeholder="seu@email.com" value={email} onChange={(e) => setEmail(e.target.value)} required />
      </div>

      {mutation.error && <p className="text-sm text-accent-rose">{mutation.error.message}</p>}

      <Button type="submit" className="w-full" disabled={mutation.isPending}>
        {mutation.isPending ? "Enviando..." : "Enviar link"}
      </Button>

      <p className="text-center text-sm">
        <Link to="/login" className="text-text-muted hover:text-accent-orange transition-colors">
          Voltar ao login
        </Link>
      </p>
    </form>
  );
}
