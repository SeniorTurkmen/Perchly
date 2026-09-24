import Link from "next/link";
import { notFound } from "next/navigation";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { ConfirmSubmitButton } from "@/components/confirm-submit-button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Separator } from "@/components/ui/separator";
import { Switch } from "@/components/ui/switch";
import { requireSessionToken } from "@/lib/auth";
import {
  AdminApiError,
  adminGetLLMCredential,
  adminListLLMModels,
} from "@/lib/backend";

import {
  createModelAction,
  deleteCredentialAction,
  deleteModelAction,
  updateCredentialAction,
  updateModelAction,
} from "../actions";
import { CredentialForm } from "../credential-form";

export default async function CredentialDetailPage({
  params,
  searchParams,
}: {
  params: Promise<{ id: string }>;
  searchParams: Promise<{ error?: string }>;
}) {
  const { id } = await params;
  const { error } = await searchParams;
  const token = await requireSessionToken();

  let credential;
  try {
    credential = await adminGetLLMCredential(token, id);
  } catch (err) {
    if (err instanceof AdminApiError && err.status === 404) notFound();
    throw err;
  }

  const models = await adminListLLMModels(token, id);

  const boundUpdateCredential = updateCredentialAction.bind(null, id);
  const boundDeleteCredential = deleteCredentialAction.bind(null, id);
  const boundCreateModel = createModelAction.bind(null, id);

  return (
    <div className="space-y-8">
      {error && (
        <div className="rounded-md border border-destructive/50 bg-destructive/10 px-4 py-3 text-sm text-destructive">
          {error}
        </div>
      )}

      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold">{credential.label}</h1>
          <Link href="/providers" className="text-sm text-muted-foreground hover:underline">
            AI Sağlayıcıları listesine dön
          </Link>
        </div>
        <form action={boundDeleteCredential}>
          <ConfirmSubmitButton
            confirmMessage="Bu kimlik bilgisini silmek istediğine emin misin? Önce bağlı modelleri silmen gerekir."
            variant="outline"
            className="text-destructive hover:text-destructive"
          >
            Kimlik bilgisini sil
          </ConfirmSubmitButton>
        </form>
      </div>

      <CredentialForm
        credential={credential}
        action={boundUpdateCredential}
        submitLabel="Kaydet"
      />

      <Separator />

      <div className="space-y-4">
        <div>
          <h2 className="text-lg font-semibold">Modeller</h2>
          <p className="text-muted-foreground">
            Bu kimlik bilgisi üzerinden çağrılabilecek modeller — personalar ileride
            bunlardan birini seçebilecek.
          </p>
        </div>

        {models.length === 0 && (
          <p className="text-sm text-muted-foreground">Henüz model eklenmedi.</p>
        )}

        <div className="space-y-3">
          {models.map((m) => {
            const boundUpdateModel = updateModelAction.bind(null, id, m.id);
            const boundDeleteModel = deleteModelAction.bind(null, id, m.id);
            return (
              <div key={m.id} className="rounded-lg border p-4">
                <form action={boundUpdateModel} className="space-y-4">
                  <div className="flex items-start justify-between gap-4">
                    <div className="flex-1 space-y-2">
                      <Label htmlFor={`display_name-${m.id}`}>Görünen ad</Label>
                      <Input
                        id={`display_name-${m.id}`}
                        name="display_name"
                        defaultValue={m.display_name}
                        required
                      />
                      <p className="text-xs text-muted-foreground">
                        Model adı: <span className="font-mono">{m.model_name}</span>{" "}
                        (değiştirilemez)
                      </p>
                    </div>
                    <div className="flex items-center gap-2">
                      {m.is_default && <Badge>Varsayılan</Badge>}
                      <Badge variant={m.is_active ? "default" : "secondary"}>
                        {m.is_active ? "Aktif" : "Pasif"}
                      </Badge>
                    </div>
                  </div>

                  <div className="flex items-center gap-6">
                    <div className="flex items-center gap-2">
                      <Switch
                        id={`is_default-${m.id}`}
                        name="is_default"
                        defaultChecked={m.is_default}
                      />
                      <Label htmlFor={`is_default-${m.id}`}>Bu kimlik bilgisi için varsayılan</Label>
                    </div>
                    <div className="flex items-center gap-2">
                      <Switch
                        id={`is_active-${m.id}`}
                        name="is_active"
                        defaultChecked={m.is_active}
                      />
                      <Label htmlFor={`is_active-${m.id}`}>Aktif</Label>
                    </div>
                  </div>

                  <div className="flex items-center gap-2">
                    <Button type="submit" size="sm">
                      Kaydet
                    </Button>
                  </div>
                </form>
                <form action={boundDeleteModel} className="mt-2">
                  <ConfirmSubmitButton
                    confirmMessage={`"${m.display_name}" modelini silmek istediğine emin misin?`}
                    className="text-destructive hover:text-destructive"
                  >
                    Modeli sil
                  </ConfirmSubmitButton>
                </form>
              </div>
            );
          })}
        </div>

        <Separator />

        <div className="max-w-md space-y-4">
          <h3 className="font-medium">Yeni model ekle</h3>
          <form action={boundCreateModel} className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="model_name">Model adı</Label>
              <Input
                id="model_name"
                name="model_name"
                placeholder="ör. gpt-4.1, gemini-2.5-flash"
                required
              />
              <p className="text-xs text-muted-foreground">
                Sağlayıcının API&apos;sinin beklediği tam model kimliği.
              </p>
            </div>
            <div className="space-y-2">
              <Label htmlFor="display_name">Görünen ad</Label>
              <Input id="display_name" name="display_name" placeholder="ör. GPT-4.1" required />
            </div>
            <div className="flex items-center gap-6">
              <div className="flex items-center gap-2">
                <Switch id="is_default" name="is_default" />
                <Label htmlFor="is_default">Varsayılan yap</Label>
              </div>
              <div className="flex items-center gap-2">
                <Switch id="is_active" name="is_active" defaultChecked />
                <Label htmlFor="is_active">Aktif</Label>
              </div>
            </div>
            <Button type="submit">Model ekle</Button>
          </form>
        </div>
      </div>
    </div>
  );
}
