import { useState } from "react";
import { useProfile, useUpdateProfile, usePreferences, useUpdatePreferences, useWorkspace, useUpdateWorkspace, useWorkspaceUsage, useExportData, useDeleteAccount } from "@/hooks/use-settings";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";
import { Select, SelectTrigger, SelectValue, SelectContent, SelectItem } from "@/components/ui/select";
import { Skeleton } from "@/components/ui/skeleton";
import { Progress } from "@/components/ui/progress";

function ProfileTab() {
  const { data: profile, isLoading } = useProfile();
  const update = useUpdateProfile();
  const [name, setName] = useState("");
  const [tz, setTz] = useState("");

  if (isLoading) return <Skeleton className="h-40 rounded-[14px]" />;

  const displayName = name || profile?.display_name || "";
  const timezone = tz || profile?.timezone || "";

  return (
    <div className="rounded-[14px] border border-border bg-bg-card p-6 space-y-4">
      <div className="space-y-2">
        <Label>Nome</Label>
        <Input value={displayName} onChange={(e) => setName(e.target.value)} />
      </div>
      <div className="space-y-2">
        <Label>Email</Label>
        <Input value={profile?.email ?? ""} disabled className="opacity-60" />
      </div>
      <div className="space-y-2">
        <Label>Fuso horário</Label>
        <Input value={timezone} onChange={(e) => setTz(e.target.value)} placeholder="America/Sao_Paulo" />
      </div>
      <Button onClick={() => update.mutate({ display_name: displayName, timezone })} disabled={update.isPending}>
        Salvar perfil
      </Button>
    </div>
  );
}

function PreferencesTab() {
  const { data: prefs, isLoading } = usePreferences();
  const update = useUpdatePreferences();
  const [currency, setCurrency] = useState("");
  const [focusLimit, setFocusLimit] = useState("");

  if (isLoading) return <Skeleton className="h-40 rounded-[14px]" />;

  return (
    <div className="rounded-[14px] border border-border bg-bg-card p-6 space-y-4">
      <div className="space-y-2">
        <Label>Moeda</Label>
        <Select value={currency || prefs?.currency || "BRL"} onValueChange={setCurrency}>
          <SelectTrigger><SelectValue /></SelectTrigger>
          <SelectContent>
            <SelectItem value="BRL">BRL (R$)</SelectItem>
            <SelectItem value="USD">USD ($)</SelectItem>
            <SelectItem value="EUR">EUR (E)</SelectItem>
          </SelectContent>
        </Select>
      </div>
      <div className="space-y-2">
        <Label>Limite de foco diário</Label>
        <Input type="number" value={focusLimit || String(prefs?.daily_focus_limit ?? 3)} onChange={(e) => setFocusLimit(e.target.value)} />
      </div>
      <div className="space-y-2">
        <Label>Início da semana</Label>
        <Select value={String(prefs?.week_starts_on ?? 1)} onValueChange={(v) => update.mutate({ week_starts_on: Number(v) })}>
          <SelectTrigger><SelectValue /></SelectTrigger>
          <SelectContent>
            <SelectItem value="0">Domingo</SelectItem>
            <SelectItem value="1">Segunda</SelectItem>
          </SelectContent>
        </Select>
      </div>
      <Button onClick={() => update.mutate({ currency: currency || prefs?.currency, daily_focus_limit: Number(focusLimit) || prefs?.daily_focus_limit })} disabled={update.isPending}>
        Salvar preferências
      </Button>
    </div>
  );
}

function WorkspaceTab() {
  const { data: ws, isLoading } = useWorkspace();
  const { data: usage } = useWorkspaceUsage();
  const update = useUpdateWorkspace();
  const [wsName, setWsName] = useState("");

  if (isLoading) return <Skeleton className="h-40 rounded-[14px]" />;

  return (
    <div className="space-y-6">
      <div className="rounded-[14px] border border-border bg-bg-card p-6 space-y-4">
        <div className="space-y-2">
          <Label>Nome do workspace</Label>
          <Input value={wsName || ws?.name || ""} onChange={(e) => setWsName(e.target.value)} />
        </div>
        <Button onClick={() => update.mutate({ name: wsName })} disabled={!wsName || update.isPending}>
          Salvar
        </Button>
      </div>
      {usage && (
        <div className="rounded-[14px] border border-border bg-bg-card p-6">
          <h3 className="font-semibold mb-4">Uso</h3>
          <div className="space-y-3">
            {[
              { label: "Areas", used: usage.counters.areas_count, max: usage.limits.max_areas },
              { label: "Metas", used: usage.counters.goals_count, max: usage.limits.max_goals },
              { label: "Hábitos", used: usage.counters.habits_count, max: usage.limits.max_habits },
            ].map((u) => (
              <div key={u.label}>
                <div className="flex justify-between text-sm mb-1">
                  <span>{u.label}</span>
                  <span className="text-text-muted">{u.used}/{u.max}</span>
                </div>
                <Progress value={u.max > 0 ? (u.used / u.max) * 100 : 0} />
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}

function AccountTab() {
  const exportData = useExportData();
  const deleteAccount = useDeleteAccount();
  const [confirmDelete, setConfirmDelete] = useState(false);

  return (
    <div className="space-y-6">
      <div className="rounded-[14px] border border-border bg-bg-card p-6">
        <h3 className="font-semibold mb-2">Exportar dados</h3>
        <p className="text-sm text-text-secondary mb-4">Receba um arquivo com todos os seus dados.</p>
        <Button variant="outline" onClick={() => exportData.mutate()} disabled={exportData.isPending}>
          {exportData.isPending ? "Exportando..." : "Solicitar exportação"}
        </Button>
        {exportData.isSuccess && <p className="mt-2 text-sm text-accent-green">Exportação solicitada! Você receberá por email.</p>}
      </div>
      <div className="rounded-[14px] border border-accent-rose/30 bg-accent-rose/5 p-6">
        <h3 className="font-semibold text-accent-rose mb-2">Zona de perigo</h3>
        <p className="text-sm text-text-secondary mb-4">Excluir sua conta permanentemente. Esta ação não pode ser desfeita.</p>
        {!confirmDelete ? (
          <Button variant="destructive" onClick={() => setConfirmDelete(true)}>
            Excluir minha conta
          </Button>
        ) : (
          <div className="flex items-center gap-3">
            <Button variant="destructive" onClick={() => deleteAccount.mutate()} disabled={deleteAccount.isPending}>
              Confirmar exclusão
            </Button>
            <Button variant="ghost" onClick={() => setConfirmDelete(false)}>Cancelar</Button>
          </div>
        )}
      </div>
    </div>
  );
}

export function Component() {
  return (
    <div>
      <h1 className="text-2xl font-bold mb-6">Configurações</h1>
      <Tabs defaultValue="profile">
        <TabsList>
          <TabsTrigger value="profile">Perfil</TabsTrigger>
          <TabsTrigger value="preferences">Preferências</TabsTrigger>
          <TabsTrigger value="workspace">Workspace</TabsTrigger>
          <TabsTrigger value="account">Conta</TabsTrigger>
        </TabsList>
        <TabsContent value="profile"><ProfileTab /></TabsContent>
        <TabsContent value="preferences"><PreferencesTab /></TabsContent>
        <TabsContent value="workspace"><WorkspaceTab /></TabsContent>
        <TabsContent value="account"><AccountTab /></TabsContent>
      </Tabs>
    </div>
  );
}
