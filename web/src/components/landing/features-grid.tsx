import { Target, Activity, DollarSign, Calendar, PenLine, BarChart3 } from "lucide-react";

const features = [
  { icon: Target, title: "Metas", desc: "Defina metas anuais, trimestrais e semanais com acompanhamento de progresso.", color: "text-accent-green" },
  { icon: Activity, title: "Habitos", desc: "Rastreie habitos diarios com heatmaps e sequencias de dias.", color: "text-accent-blue" },
  { icon: Calendar, title: "Tarefas", desc: "Organize tarefas por prioridade e foco, conectadas a suas metas.", color: "text-accent-orange" },
  { icon: DollarSign, title: "Financas", desc: "Categorize transacoes, defina orcamentos e acompanhe o saldo mensal.", color: "text-accent-sand" },
  { icon: PenLine, title: "Journal", desc: "Registre humor, energia, vitorias e gratidao diariamente.", color: "text-accent-rose" },
  { icon: BarChart3, title: "Life Score", desc: "Pontuacao semanal que reflete o equilibrio entre todas as areas da vida.", color: "text-accent-sage" },
];

export function FeaturesGrid() {
  return (
    <section id="features" className="mx-auto max-w-6xl px-6 py-24">
      <h2 className="mb-3 text-center text-3xl font-bold">Tudo que voce precisa</h2>
      <p className="mx-auto mb-12 max-w-lg text-center font-serif text-text-secondary">
        Uma plataforma integrada para gerenciar todas as dimensoes da sua vida.
      </p>
      <div className="grid gap-6 md:grid-cols-2 lg:grid-cols-3">
        {features.map((f) => (
          <div key={f.title} className="group rounded-[14px] border border-border bg-bg-card p-6 transition-colors hover:border-accent-orange/30">
            <f.icon className={`mb-3 h-6 w-6 ${f.color}`} />
            <h3 className="mb-1 font-semibold">{f.title}</h3>
            <p className="text-sm leading-relaxed text-text-secondary">{f.desc}</p>
          </div>
        ))}
      </div>
    </section>
  );
}
