import { Link } from "react-router";
import { XCircle } from "lucide-react";
import { Button } from "@/components/ui/button";

export function Component() {
  return (
    <div className="flex flex-col items-center justify-center py-20 text-center">
      <XCircle className="mb-4 h-16 w-16 text-accent-rose" />
      <h1 className="text-2xl font-bold mb-2">Pagamento cancelado</h1>
      <p className="mb-8 text-text-secondary">Nenhuma cobranca foi realizada.</p>
      <Button asChild>
        <Link to="/billing">Ver planos</Link>
      </Button>
    </div>
  );
}
