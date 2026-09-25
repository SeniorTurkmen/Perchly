import { getTranslations } from "next-intl/server";
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
  const t = await getTranslations("providers");
  const tc = await getTranslations("common");

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
            {t("backToList")}
          </Link>
        </div>
        <form action={boundDeleteCredential}>
          <ConfirmSubmitButton
            confirmMessage={t("confirmDeleteCredential")}
            variant="outline"
            className="text-destructive hover:text-destructive"
          >
            {t("deleteCredential")}
          </ConfirmSubmitButton>
        </form>
      </div>

      <CredentialForm
        credential={credential}
        action={boundUpdateCredential}
        submitLabel={t("save")}
      />

      <Separator />

      <div className="space-y-4">
        <div>
          <h2 className="text-lg font-semibold">{t("modelsTitle")}</h2>
          <p className="text-muted-foreground">
            {t("modelsSubtitle")}
          </p>
        </div>

        {models.length === 0 && (
          <p className="text-sm text-muted-foreground">{t("noModelsYet")}</p>
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
                      <Label htmlFor={`display_name-${m.id}`}>{t("displayName")}</Label>
                      <Input
                        id={`display_name-${m.id}`}
                        name="display_name"
                        defaultValue={m.display_name}
                        required
                      />
                      <p className="text-xs text-muted-foreground">
                        {t("modelNameLabel")} <span className="font-mono">{m.model_name}</span>{" "}
                        {t("immutable")}
                      </p>
                    </div>
                    <div className="flex items-center gap-2">
                      {m.is_default && <Badge>{t("default")}</Badge>}
                      <Badge variant={m.is_active ? "default" : "secondary"}>
                        {m.is_active ? tc("active") : tc("inactive")}
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
                      <Label htmlFor={`is_default-${m.id}`}>{t("defaultForCredential")}</Label>
                    </div>
                    <div className="flex items-center gap-2">
                      <Switch
                        id={`is_active-${m.id}`}
                        name="is_active"
                        defaultChecked={m.is_active}
                      />
                      <Label htmlFor={`is_active-${m.id}`}>{tc("active")}</Label>
                    </div>
                  </div>

                  <div className="flex items-center gap-2">
                    <Button type="submit" size="sm">
                      {t("save")}
                    </Button>
                  </div>
                </form>
                <form action={boundDeleteModel} className="mt-2">
                  <ConfirmSubmitButton
                    confirmMessage={t("confirmDeleteModel", { name: m.display_name })}
                    className="text-destructive hover:text-destructive"
                  >
                    {t("deleteModel")}
                  </ConfirmSubmitButton>
                </form>
              </div>
            );
          })}
        </div>

        <Separator />

        <div className="max-w-md space-y-4">
          <h3 className="font-medium">{t("addModel")}</h3>
          <form action={boundCreateModel} className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="model_name">{t("modelName")}</Label>
              <Input
                id="model_name"
                name="model_name"
                placeholder={t("modelNamePlaceholder")}
                required
              />
              <p className="text-xs text-muted-foreground">
                {t("modelNameHint")}
              </p>
            </div>
            <div className="space-y-2">
              <Label htmlFor="display_name">{t("displayName")}</Label>
              <Input id="display_name" name="display_name" placeholder={t("displayNamePlaceholder")} required />
            </div>
            <div className="flex items-center gap-6">
              <div className="flex items-center gap-2">
                <Switch id="is_default" name="is_default" />
                <Label htmlFor="is_default">{t("makeDefault")}</Label>
              </div>
              <div className="flex items-center gap-2">
                <Switch id="is_active" name="is_active" defaultChecked />
                <Label htmlFor="is_active">{tc("active")}</Label>
              </div>
            </div>
            <Button type="submit">{t("addModel")}</Button>
          </form>
        </div>
      </div>
    </div>
  );
}
