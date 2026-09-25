import { getTranslations } from "next-intl/server";

import { requireSessionToken } from "@/lib/auth";
import { adminListLLMCredentials, adminListLLMModels } from "@/lib/backend";

import { createPersonaAction } from "../actions";
import { buildModelOptions, PersonaForm } from "../persona-form";

export default async function NewPersonaPage({
  searchParams,
}: {
  searchParams: Promise<{ error?: string }>;
}) {
  const { error } = await searchParams;
  const token = await requireSessionToken();
  const t = await getTranslations("personas");
  const [models, credentials] = await Promise.all([
    adminListLLMModels(token),
    adminListLLMCredentials(token),
  ]);

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold">{t("newTitle")}</h1>
        <p className="text-muted-foreground">
          {t("newSubtitle")}
        </p>
      </div>
      <PersonaForm
        modelOptions={await buildModelOptions(models, credentials)}
        action={createPersonaAction}
        submitLabel={t("create")}
        error={error}
      />
    </div>
  );
}
