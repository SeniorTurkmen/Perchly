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
  const [models, credentials] = await Promise.all([
    adminListLLMModels(token),
    adminListLLMCredentials(token),
  ]);

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold">Yeni persona</h1>
        <p className="text-muted-foreground">
          Yeni bir persona oluştur — tüm alanlar zorunlu, dial değerleri 0-100.
        </p>
      </div>
      <PersonaForm
        modelOptions={buildModelOptions(models, credentials)}
        action={createPersonaAction}
        submitLabel="Oluştur"
        error={error}
      />
    </div>
  );
}
