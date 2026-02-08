import { useState } from "react";
import { Link, useSearchParams } from "react-router";
import { useMutation } from "@tanstack/react-query";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { api } from "@/lib/api-client";

export function Component() {
  const [params] = useSearchParams();
  const [password, setPassword] = useState("");
  const [done, setDone] = useState(false);

  const mutation = useMutation({
    mutationFn: (password: string) =>
      api("/auth/reset-password", {
        method: "POST",
        body: JSON.stringify({ access_token: params.get("token"), password }),
      }),
    onSuccess: () => setDone(true),
  });

  if (done) {
    return (
      <div className="rounded-[14px] border border-border bg-bg-card p-6 text-center">
        <h1 className="text-xl font-bold text-text-primary mb-2">Senha redefinida</h1>
        <p className="text-sm text-text-secondary font-serif mb-6">Sua senha foi alterada com sucesso.</p>
        <Link to="/login" className="text-sm text-accent-orange hover:text-accent-orange/80 transition-colors">
          Ir para login
        </Link>
      </div>
    );
  }

  return (
    <form onSubmit={(e) => { e.preventDefault(); mutation.mutate(password); }} className="space-y-4 rounded-[14px] border border-border bg-bg-card p-6">
      <div className="mb-6 text-center">
        <h1 className="text-xl font-bold text-text-primary">Nova senha</h1>
        <p className="mt-1 text-sm text-text-secondary font-serif">Defina sua nova senha</p>
      </div>

      <div className="space-y-2">
        <Label htmlFor="password">Nova senha</Label>
        <Input id="password" type="password" placeholder="Min. 8 caracteres" value={password} onChange={(e) => setPassword(e.target.value)} required minLength={8} />
      </div>

      {mutation.error && <p className="text-sm text-accent-rose">{mutation.error.message}</p>}

      <Button type="submit" className="w-full" disabled={mutation.isPending}>
        {mutation.isPending ? "Salvando..." : "Redefinir senha"}
      </Button>

      <p className="text-center text-sm">
        <Link to="/login" className="text-text-muted hover:text-accent-orange transition-colors">
          Voltar ao login
        </Link>
      </p>
    </form>
  );
}
