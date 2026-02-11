import { PricingCard } from "./pricing-card";

const plans = [
  {
    name: "Grátis",
    price: "R$0",
    features: ["3 áreas de vida", "5 metas ativas", "Hábitos e tarefas básicos", "Journal diário"],
    cta: "Começar grátis",
  },
  {
    name: "Pro",
    price: "R$19",
    period: "/mês",
    features: ["Áreas ilimitadas", "Metas ilimitadas", "Módulo financeiro", "Exportação de dados", "Histórico de score"],
    cta: "Assinar Pro",
    highlighted: true,
  },
  {
    name: "Premium",
    price: "R$39",
    period: "/mês",
    features: ["Tudo do Pro", "AI Insights", "Acesso via API", "Armazenamento estendido", "Suporte prioritário"],
    cta: "Assinar Premium",
  },
];

export function PricingTable() {
  return (
    <section id="pricing" className="mx-auto max-w-5xl px-6 py-24">
      <h2 className="mb-3 text-center text-3xl font-bold">Planos simples e transparentes</h2>
      <p className="mx-auto mb-12 max-w-md text-center font-serif text-text-secondary">
        Comece grátis, evolua quando quiser.
      </p>
      <div className="grid gap-6 md:grid-cols-3">
        {plans.map((p) => (
          <PricingCard key={p.name} {...p} />
        ))}
      </div>
    </section>
  );
}
