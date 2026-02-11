import { useState } from "react";
import { useAuthStore } from "@/stores/auth-store";
import {
  useProfile,
  useUpdateProfile,
  usePreferences,
  useUpdatePreferences,
  useWorkspace,
  useUpdateWorkspace,
  useWorkspaceUsage,
  useExportData,
  useDeleteAccount,
} from "@/hooks/use-settings";
import {
  useWorkspaces,
  useWorkspaceMembers,
  useRemoveWorkspaceMember,
  useWorkspaceInvites,
  useCreateWorkspaceInvite,
  useRevokeWorkspaceInvite,
} from "@/hooks/use-workspaces";
import {
  useReferralAttributions,
  useReferralCredits,
  useReferralMe,
  useRegenerateReferralCode,
} from "@/hooks/use-referrals";
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

async function copyText(text: string) {
  if (!text) return;
  if (navigator.clipboard?.writeText) {
    await navigator.clipboard.writeText(text);
    return;
  }
  const el = document.createElement("textarea");
  el.value = text;
  document.body.appendChild(el);
  el.select();
  document.execCommand("copy");
  document.body.removeChild(el);
}

function WorkspaceTab() {
  const authUser = useAuthStore((s) => s.user);
  const { data: ws, isLoading } = useWorkspace();
  const { data: usage } = useWorkspaceUsage();
  const referralEnabled = usage?.limits.referral_enabled === true;
  const familyEnabled = usage?.limits.family_enabled ?? false;

  const { data: allWorkspaces } = useWorkspaces();
  const { data: members } = useWorkspaceMembers();
  const { data: invites } = useWorkspaceInvites();
  const { data: referralMe } = useReferralMe(referralEnabled);
  const { data: referralAttributions } = useReferralAttributions(10, 0, referralEnabled);
  const { data: referralCredits } = useReferralCredits(10, 0, referralEnabled);

  const update = useUpdateWorkspace();
  const removeMember = useRemoveWorkspaceMember();
  const createInvite = useCreateWorkspaceInvite();
  const revokeInvite = useRevokeWorkspaceInvite();
  const regenerateCode = useRegenerateReferralCode();

  const [wsName, setWsName] = useState("");
  const [inviteEmail, setInviteEmail] = useState("");
  const [lastInviteURL, setLastInviteURL] = useState("");
  const [copyFeedback, setCopyFeedback] = useState("");

  if (isLoading) return <Skeleton className="h-40 rounded-[14px]" />;

  const currentWorkspace = allWorkspaces?.find((item) => item.is_default);
  const canManage = currentWorkspace?.role === "owner" || currentWorkspace?.role === "admin";

  const inviteLimit = usage?.limits.max_members ?? 1;
  const inviteUsed = usage?.counters.members_count ?? 1;

  return (
    <div className="space-y-6">
      <div className="rounded-[14px] border border-border bg-bg-card p-6 space-y-4">
        <h3 className="font-semibold">Workspace</h3>
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
          <h3 className="font-semibold mb-4">Uso e limites</h3>
          <div className="space-y-3">
            {[
              { label: "Áreas", used: usage.counters.areas_count, max: usage.limits.max_areas },
              { label: "Metas", used: usage.counters.goals_count, max: usage.limits.max_goals },
              { label: "Hábitos", used: usage.counters.habits_count, max: usage.limits.max_habits },
              { label: "Membros", used: usage.counters.members_count, max: usage.limits.max_members },
            ].map((u) => (
              <div key={u.label}>
                <div className="flex justify-between text-sm mb-1">
                  <span>{u.label}</span>
                  <span className="text-text-muted">{u.used}/{u.max < 0 ? "∞" : u.max}</span>
                </div>
                <Progress value={u.max > 0 ? (u.used / u.max) * 100 : 0} />
              </div>
            ))}
          </div>
        </div>
      )}

      <div className="rounded-[14px] border border-border bg-bg-card p-6 space-y-4">
        <div className="flex items-center justify-between">
          <h3 className="font-semibold">Membros</h3>
          <span className="text-xs text-text-muted">{inviteUsed}/{inviteLimit < 0 ? "∞" : inviteLimit}</span>
        </div>

        {!familyEnabled && (
          <p className="text-sm text-accent-rose">Family plan desabilitado para este workspace.</p>
        )}

        {familyEnabled && canManage && (
          <div className="space-y-2">
            <Label>Convidar por email</Label>
            <div className="flex flex-col sm:flex-row gap-2">
              <Input
                type="email"
                placeholder="email@exemplo.com"
                value={inviteEmail}
                onChange={(e) => setInviteEmail(e.target.value)}
              />
              <Button
                disabled={createInvite.isPending || !inviteEmail}
                onClick={() => {
                  createInvite.mutate(
                    { email: inviteEmail, role: "member" },
                    {
                      onSuccess: (data) => {
                        setLastInviteURL(data.invite_url);
                        setInviteEmail("");
                      },
                    },
                  );
                }}
              >
                {createInvite.isPending ? "Criando..." : "Criar convite"}
              </Button>
            </div>
            {lastInviteURL && (
              <div className="rounded-lg border border-border bg-bg-secondary/30 p-3">
                <p className="text-xs text-text-muted mb-2">Link gerado</p>
                <div className="flex flex-col sm:flex-row gap-2">
                  <Input value={lastInviteURL} readOnly />
                  <Button
                    variant="outline"
                    onClick={async () => {
                      await copyText(lastInviteURL);
                      setCopyFeedback("Link copiado!");
                      setTimeout(() => setCopyFeedback(""), 1800);
                    }}
                  >
                    Copiar
                  </Button>
                </div>
                {copyFeedback && <p className="mt-2 text-xs text-accent-green">{copyFeedback}</p>}
              </div>
            )}
          </div>
        )}

        <div className="space-y-2">
          {(members ?? []).map((member) => (
            <div key={member.user_id} className="flex items-center justify-between rounded-lg border border-border p-3">
              <div>
                <p className="text-sm font-medium">{member.display_name}</p>
                <p className="text-xs text-text-muted">{member.email} · {member.role}</p>
              </div>
              {canManage && member.role !== "owner" && member.user_id !== authUser?.id && (
                <Button
                  variant="outline"
                  className="text-xs"
                  disabled={removeMember.isPending}
                  onClick={() => removeMember.mutate(member.user_id)}
                >
                  Remover
                </Button>
              )}
            </div>
          ))}
        </div>

        {familyEnabled && canManage && (invites?.length ?? 0) > 0 && (
          <div className="space-y-2 pt-2">
            <h4 className="text-sm font-semibold">Convites</h4>
            {(invites ?? []).map((invite) => (
              <div key={invite.id} className="flex items-center justify-between rounded-lg border border-border p-3">
                <div>
                  <p className="text-sm font-medium">{invite.email}</p>
                  <p className="text-xs text-text-muted">
                    {invite.accepted_at ? "Aceito" : invite.revoked_at ? "Revogado" : "Pendente"}
                  </p>
                </div>
                {!invite.accepted_at && !invite.revoked_at && (
                  <Button
                    variant="outline"
                    className="text-xs"
                    disabled={revokeInvite.isPending}
                    onClick={() => revokeInvite.mutate(invite.id)}
                  >
                    Revogar
                  </Button>
                )}
              </div>
            ))}
          </div>
        )}
      </div>

      <div className="rounded-[14px] border border-border bg-bg-card p-6 space-y-4">
        <h3 className="font-semibold">Referral</h3>
        {!referralEnabled ? (
          <p className="text-sm text-accent-rose">Referral desabilitado para este workspace.</p>
        ) : (
          <>
            <div className="rounded-lg border border-border bg-bg-secondary/20 p-4 space-y-3">
              <div className="flex items-center justify-between gap-2">
                <div>
                  <p className="text-xs text-text-muted">Seu código</p>
                  <p className="font-mono text-lg font-semibold">{referralMe?.code ?? "-"}</p>
                </div>
                {canManage && (
                  <Button variant="outline" onClick={() => regenerateCode.mutate()} disabled={regenerateCode.isPending}>
                    Regenerar
                  </Button>
                )}
              </div>
              <div className="flex flex-col sm:flex-row gap-2">
                <Input readOnly value={referralMe?.share_url ?? ""} />
                <Button
                  variant="outline"
                  onClick={async () => {
                    await copyText(referralMe?.share_url ?? "");
                    setCopyFeedback("Link de referral copiado!");
                    setTimeout(() => setCopyFeedback(""), 1800);
                  }}
                >
                  Copiar link
                </Button>
              </div>
              <div className="grid grid-cols-3 gap-3 text-xs">
                <div className="rounded-md border border-border p-2 text-center">
                  <div className="text-text-muted">Pendentes</div>
                  <div className="text-sm font-semibold">{referralMe?.stats.pending ?? 0}</div>
                </div>
                <div className="rounded-md border border-border p-2 text-center">
                  <div className="text-text-muted">Convertidos</div>
                  <div className="text-sm font-semibold">{referralMe?.stats.converted ?? 0}</div>
                </div>
                <div className="rounded-md border border-border p-2 text-center">
                  <div className="text-text-muted">Créditos (dias)</div>
                  <div className="text-sm font-semibold">{referralMe?.credit_days ?? 0}</div>
                </div>
              </div>
            </div>

            <div className="space-y-2">
              <h4 className="text-sm font-semibold">Attributions</h4>
              {(referralAttributions ?? []).length === 0 && (
                <p className="text-sm text-text-muted">Nenhuma indicação registrada ainda.</p>
              )}
              {(referralAttributions ?? []).map((item) => (
                <div key={item.id} className="rounded-lg border border-border p-3 text-sm">
                  <p className="font-medium">{item.referral_code}</p>
                  <p className="text-xs text-text-muted">Status: {item.status}</p>
                </div>
              ))}
            </div>

            <div className="space-y-2">
              <h4 className="text-sm font-semibold">Créditos</h4>
              {(referralCredits ?? []).length === 0 && (
                <p className="text-sm text-text-muted">Nenhum crédito disponível.</p>
              )}
              {(referralCredits ?? []).map((item) => (
                <div key={item.id} className="rounded-lg border border-border p-3 text-sm flex items-center justify-between">
                  <div>
                    <p className="font-medium">{item.credit_type}</p>
                    <p className="text-xs text-text-muted">{item.status}</p>
                  </div>
                  <p className="font-semibold">{item.days} dias</p>
                </div>
              ))}
            </div>
          </>
        )}
      </div>
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
        <TabsList className="flex flex-wrap gap-1">
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
