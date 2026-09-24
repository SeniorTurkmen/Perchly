import { createCredentialAction } from "../actions";
import { CredentialForm } from "../credential-form";

export default async function NewCredentialPage({
  searchParams,
}: {
  searchParams: Promise<{ error?: string }>;
}) {
  const { error } = await searchParams;

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold">Kimlik bilgisi ekle</h1>
        <p className="text-muted-foreground">
          Bir LLM sağlayıcısı için API anahtarı ekle — anahtar şifrelenmiş saklanır.
        </p>
      </div>
      <CredentialForm action={createCredentialAction} submitLabel="Oluştur" error={error} />
    </div>
  );
}
