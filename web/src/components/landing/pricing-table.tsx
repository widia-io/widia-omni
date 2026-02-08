import { PricingCard } from "./pricing-card";

const plans = [
  {
    name: "Gratis",
    price: "R$0",
    features: ["3 areas de vida", "5 metas ativas", "Habitos e tarefas basicos", "Journal diario"],
    cta: "Comecar gratis",
  },
  {
    name: "Pro",
    price: "R$19",
    period: "/mes",
    features: ["Areas ilimitadas", "Metas ilimitadas", "Modulo financeiro", "Exportacao de dados", "Historico de score"],
    cta: "Assinar Pro",
    highlighted: true,
  },
  {
    name: "Premium",
    price: "R$39",
    period: "/mes",
    features: ["Tudo do Pro", "AI Insights", "Acesso via API", "Armazenamento estendido", "Suporte prioritario"],
    cta: "Assinar Premium",
  },
];

export function PricingTable() {
  return (
    <section id="pricing" className="mx-auto max-w-5xl px-6 py-24">
      <h2 className="mb-3 text-center text-3xl font-bold">Planos simples e transparentes</h2>
      <p className="mx-auto mb-12 max-w-md text-center font-serif text-text-secondary">
        Comece gratis, evolua quando quiser.
      </p>
      <div className="grid gap-6 md:grid-cols-3">
        {plans.map((p) => (
          <PricingCard key={p.name} {...p} />
        ))}
      </div>
    </section>
  );
}
