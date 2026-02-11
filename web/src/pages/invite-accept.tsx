import { useEffect, useRef } from "react";
import { Link, useNavigate, useParams } from "react-router";
import { Button } from "@/components/ui/button";
import { useAuthStore } from "@/stores/auth-store";
import { useAcceptWorkspaceInvite } from "@/hooks/use-workspaces";

const PENDING_INVITE_KEY = "pending_invite_token";

export function Component() {
  const navigate = useNavigate();
  const { token } = useParams<{ token: string }>();
  const accessToken = useAuthStore((s) => s.accessToken);
  const acceptInvite = useAcceptWorkspaceInvite();
  const started = useRef(false);

  useEffect(() => {
    if (!token) return;

    sessionStorage.setItem(PENDING_INVITE_KEY, token);

    if (!accessToken) {
      navigate("/login", { replace: true });
      return;
    }

    if (started.current) return;
    started.current = true;

    acceptInvite.mutate(token, {
      onSuccess: () => {
        sessionStorage.removeItem(PENDING_INVITE_KEY);
        navigate("/dashboard", { replace: true });
      },
    });
  }, [token, accessToken, navigate, acceptInvite]);

  if (!token) {
    return (
      <div className="mx-auto mt-20 max-w-md rounded-[14px] border border-border bg-bg-card p-6 text-center">
        <h1 className="text-xl font-bold mb-2">Convite inválido</h1>
        <p className="text-sm text-text-muted mb-4">O link de convite está incompleto.</p>
        <Button asChild><Link to="/">Voltar</Link></Button>
      </div>
    );
  }

  if (acceptInvite.isError) {
    return (
      <div className="mx-auto mt-20 max-w-md rounded-[14px] border border-border bg-bg-card p-6 text-center">
        <h1 className="text-xl font-bold mb-2">Não foi possível aceitar o convite</h1>
        <p className="text-sm text-text-muted mb-4">{acceptInvite.error.message}</p>
        <div className="flex items-center justify-center gap-2">
          <Button asChild variant="outline"><Link to="/login">Entrar</Link></Button>
          <Button asChild><Link to="/register">Criar conta</Link></Button>
        </div>
      </div>
    );
  }

  return (
    <div className="mx-auto mt-20 max-w-md rounded-[14px] border border-border bg-bg-card p-6 text-center">
      <h1 className="text-xl font-bold mb-2">Aceitando convite...</h1>
      <p className="text-sm text-text-muted">Você será redirecionado automaticamente.</p>
    </div>
  );
}
