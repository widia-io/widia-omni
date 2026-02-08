import { Link } from "react-router";
import { CheckCircle2 } from "lucide-react";
import { Button } from "@/components/ui/button";

export function Component() {
  return (
    <div className="flex flex-col items-center justify-center py-20 text-center">
      <CheckCircle2 className="mb-4 h-16 w-16 text-accent-green" />
      <h1 className="text-2xl font-bold mb-2">Pagamento confirmado!</h1>
      <p className="mb-8 text-text-secondary">Seu plano foi atualizado com sucesso.</p>
      <Button asChild>
        <Link to="/dashboard">Voltar ao Dashboard</Link>
      </Button>
    </div>
  );
}
