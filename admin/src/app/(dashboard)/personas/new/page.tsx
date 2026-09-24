import { createPersonaAction } from "../actions";
import { PersonaForm } from "../persona-form";

export default async function NewPersonaPage({
  searchParams,
}: {
  searchParams: Promise<{ error?: string }>;
}) {
  const { error } = await searchParams;

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold">Yeni persona</h1>
        <p className="text-muted-foreground">
          Yeni bir persona oluştur — tüm alanlar zorunlu, dial değerleri 0-100.
        </p>
      </div>
      <PersonaForm action={createPersonaAction} submitLabel="Oluştur" error={error} />
    </div>
  );
}
