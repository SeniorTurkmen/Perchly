import { getTranslations } from "next-intl/server";

import { createCredentialAction } from "../actions";
import { CredentialForm } from "../credential-form";

export default async function NewCredentialPage({
  searchParams,
}: {
  searchParams: Promise<{ error?: string }>;
}) {
  const { error } = await searchParams;
  const t = await getTranslations("providers");

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold">{t("addCredential")}</h1>
        <p className="text-muted-foreground">
          {t("addCredentialSubtitle")}
        </p>
      </div>
      <CredentialForm action={createCredentialAction} submitLabel={t("create")} error={error} />
    </div>
  );
}
