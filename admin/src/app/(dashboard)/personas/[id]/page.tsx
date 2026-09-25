import { getTranslations } from "next-intl/server";
import { notFound } from "next/navigation";

import { requireSessionToken } from "@/lib/auth";
import {
  AdminApiError,
  adminGetPersona,
  adminListLLMCredentials,
  adminListLLMModels,
} from "@/lib/backend";

import { updatePersonaAction } from "../actions";
import { buildModelOptions, PersonaForm } from "../persona-form";

export default async function EditPersonaPage({
  params,
  searchParams,
}: {
  params: Promise<{ id: string }>;
  searchParams: Promise<{ error?: string }>;
}) {
  const { id } = await params;
  const { error } = await searchParams;
  const token = await requireSessionToken();
  const t = await getTranslations("personas");

  let persona;
  try {
    persona = await adminGetPersona(token, id);
  } catch (err) {
    if (err instanceof AdminApiError && err.status === 404) notFound();
    throw err;
  }

  const [models, credentials] = await Promise.all([
    adminListLLMModels(token),
    adminListLLMCredentials(token),
  ]);

  const boundUpdate = updatePersonaAction.bind(null, id);

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold">{persona.name}</h1>
        <p className="text-muted-foreground">{persona.slug}</p>
      </div>
      <PersonaForm
        persona={persona}
        modelOptions={await buildModelOptions(models, credentials)}
        action={boundUpdate}
        submitLabel={t("save")}
        error={error}
      />
    </div>
  );
}
