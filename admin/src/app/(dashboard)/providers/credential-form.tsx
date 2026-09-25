import { getTranslations } from "next-intl/server";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Switch } from "@/components/ui/switch";
import type { LLMCredential } from "@/lib/backend";

// Brand/product names — never translated.
const PROVIDERS = [
  { value: "openai", label: "OpenAI" },
  { value: "anthropic", label: "Anthropic" },
  { value: "gemini", label: "Google Gemini" },
  { value: "huggingface", label: "Hugging Face" },
  { value: "deepseek", label: "DeepSeek" },
] as const;

export async function CredentialForm({
  credential,
  action,
  submitLabel,
  error,
}: {
  credential?: LLMCredential;
  action: (formData: FormData) => Promise<void>;
  submitLabel: string;
  error?: string;
}) {
  const t = await getTranslations("providers");
  const tc = await getTranslations("common");
  const isEdit = Boolean(credential);

  return (
    <form action={action} className="max-w-xl space-y-6">
      {error && <p className="text-sm text-destructive">{error}</p>}

      {isEdit ? (
        <div className="space-y-2">
          <Label>{t("provider")}</Label>
          <p className="text-sm text-muted-foreground">
            {PROVIDERS.find((p) => p.value === credential!.provider)?.label ??
              credential!.provider}{" "}
            — {t("providerImmutable")}
          </p>
        </div>
      ) : (
        <div className="space-y-2">
          <Label htmlFor="provider">{t("provider")}</Label>
          <Select name="provider" defaultValue="openai" required>
            <SelectTrigger id="provider" className="w-full">
              <SelectValue placeholder={t("selectProvider")} />
            </SelectTrigger>
            <SelectContent>
              {PROVIDERS.map((p) => (
                <SelectItem key={p.value} value={p.value}>
                  {p.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
      )}

      <div className="space-y-2">
        <Label htmlFor="label">{t("label")}</Label>
        <Input
          id="label"
          name="label"
          defaultValue={credential?.label}
          placeholder={t("labelPlaceholder")}
          required
        />
      </div>

      <div className="space-y-2">
        <Label htmlFor="api_key">
          {isEdit ? t("newApiKey") : t("apiKey")}
        </Label>
        <Input
          id="api_key"
          name="api_key"
          type="password"
          autoComplete="off"
          placeholder={isEdit ? t("currentKey", { preview: credential?.api_key_preview ?? "" }) : "sk-..."}
          required={!isEdit}
        />
        <p className="text-xs text-muted-foreground">
          {isEdit ? t("apiKeyHintEdit") : t("apiKeyHintNew")}
        </p>
      </div>

      <div className="space-y-2">
        <Label htmlFor="base_url">{t("baseUrl")}</Label>
        <Input
          id="base_url"
          name="base_url"
          defaultValue={credential?.base_url ?? ""}
          placeholder={t("baseUrlPlaceholder")}
        />
      </div>

      <div className="flex items-center gap-2">
        <Switch id="is_active" name="is_active" defaultChecked={credential?.is_active ?? true} />
        <Label htmlFor="is_active">{tc("active")}</Label>
      </div>

      <Button type="submit">{submitLabel}</Button>
    </form>
  );
}
