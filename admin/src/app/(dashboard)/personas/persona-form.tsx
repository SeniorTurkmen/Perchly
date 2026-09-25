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
import { Textarea } from "@/components/ui/textarea";
import type { LLMCredential, LLMModel, Persona } from "@/lib/backend";

// Base UI's Select needs a non-empty value for every item, so "use the
// process-wide default" (llm_model_id: null) is represented by this
// sentinel and translated back to null in actions.ts's readPersonaInput.
export const DEFAULT_MODEL_VALUE = "__default__";

export type LLMModelOption = {
  id: string;
  label: string;
  disabled: boolean;
};

// buildModelOptions joins every stored model with its credential to
// produce a display label ("OpenAI - Prod / GPT-4.1") and disables a
// model whose credential (or the model itself) is inactive — matching
// AdminPersonaService.validateLLMModel's own check, so an admin sees why
// an option can't be picked instead of the request just failing later.
export async function buildModelOptions(
  models: LLMModel[],
  credentials: LLMCredential[],
): Promise<LLMModelOption[]> {
  const t = await getTranslations("personas");
  const credentialById = new Map(credentials.map((c) => [c.id, c]));

  return models.map((model) => {
    const credential = credentialById.get(model.credential_id);
    const credentialLabel = credential?.label ?? t("unknownCredential");
    const disabled = !model.is_active || !(credential?.is_active ?? false);
    return {
      id: model.id,
      label: `${credentialLabel} / ${model.display_name}${disabled ? ` (${t("inactive")})` : ""}`,
      disabled,
    };
  });
}

export async function PersonaForm({
  persona,
  modelOptions,
  action,
  submitLabel,
  error,
}: {
  persona?: Persona;
  modelOptions: LLMModelOption[];
  action: (formData: FormData) => Promise<void>;
  submitLabel: string;
  error?: string;
}) {
  const t = await getTranslations("personas");

  const TRAITS = [
    { key: "warmth", label: t("traits.warmth") },
    { key: "humor", label: t("traits.humor") },
    { key: "wisdom", label: t("traits.wisdom") },
    { key: "directness", label: t("traits.directness") },
    { key: "energy", label: t("traits.energy") },
  ] as const;

  // Base UI renders the raw value in the closed trigger unless `items`
  // maps each value to a label. Without this, a selected model shows as its id.
  const modelItems: Record<string, string> = {
    [DEFAULT_MODEL_VALUE]: t("defaultModel"),
  };
  for (const option of modelOptions) {
    modelItems[option.id] = option.label;
  }
  if (persona?.llm_model_id && modelItems[persona.llm_model_id] == null) {
    modelItems[persona.llm_model_id] = t("unknownModel");
  }

  return (
    <form action={action} className="max-w-2xl space-y-6">
      {error && <p className="text-sm text-destructive">{error}</p>}

      <div className="grid gap-4 sm:grid-cols-2">
        <div className="space-y-2">
          <Label htmlFor="slug">{t("fields.slug")}</Label>
          <Input id="slug" name="slug" defaultValue={persona?.slug} required />
        </div>
        <div className="space-y-2">
          <Label htmlFor="name">{t("fields.name")}</Label>
          <Input id="name" name="name" defaultValue={persona?.name} required />
        </div>
        <div className="space-y-2">
          <Label htmlFor="category">{t("fields.category")}</Label>
          <Input id="category" name="category" defaultValue={persona?.category} required />
        </div>
        <div className="space-y-2">
          <Label htmlFor="accent_color">{t("fields.accentColor")}</Label>
          <Input
            id="accent_color"
            name="accent_color"
            defaultValue={persona?.accent_color ?? "#FF6B35"}
            required
          />
        </div>
        <div className="space-y-2">
          <Label htmlFor="avatar_url">{t("fields.avatarUrl")}</Label>
          <Input id="avatar_url" name="avatar_url" defaultValue={persona?.avatar_url ?? ""} />
        </div>
        <div className="space-y-2">
          <Label htmlFor="sort_order">{t("fields.sortOrder")}</Label>
          <Input
            id="sort_order"
            name="sort_order"
            type="number"
            defaultValue={persona?.sort_order ?? 0}
          />
        </div>
      </div>

      <div className="space-y-2">
        <Label htmlFor="short_description">{t("fields.shortDescription")}</Label>
        <Textarea
          id="short_description"
          name="short_description"
          defaultValue={persona?.short_description}
          required
          rows={2}
        />
      </div>

      <div className="space-y-2">
        <Label htmlFor="tone_description">{t("fields.toneDescription")}</Label>
        <Textarea
          id="tone_description"
          name="tone_description"
          defaultValue={persona?.tone_description}
          required
          rows={2}
        />
      </div>

      <div className="space-y-2">
        <Label htmlFor="system_prompt">{t("fields.systemPrompt")}</Label>
        <Textarea
          id="system_prompt"
          name="system_prompt"
          defaultValue={persona?.system_prompt}
          required
          rows={8}
          className="font-mono text-sm"
        />
        <p className="text-xs text-muted-foreground">
          {t("systemPromptHint")}
        </p>
      </div>

      <div className="space-y-2">
        <Label htmlFor="llm_model_id">{t("fields.model")}</Label>
        <Select
          name="llm_model_id"
          defaultValue={persona?.llm_model_id ?? DEFAULT_MODEL_VALUE}
          items={modelItems}
        >
          <SelectTrigger id="llm_model_id" className="w-full">
            <SelectValue placeholder={t("selectModel")} />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value={DEFAULT_MODEL_VALUE}>
              {t("defaultModel")}
            </SelectItem>
            {modelOptions.map((option) => (
              <SelectItem key={option.id} value={option.id} disabled={option.disabled}>
                {option.label}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
        <p className="text-xs text-muted-foreground">
          {t("modelHint")}
        </p>
      </div>

      <div className="grid grid-cols-5 gap-4">
        {TRAITS.map((trait) => (
          <div key={trait.key} className="space-y-2">
            <Label htmlFor={trait.key}>{trait.label}</Label>
            <Input
              id={trait.key}
              name={trait.key}
              type="number"
              min={0}
              max={100}
              defaultValue={persona?.default_traits[trait.key] ?? 50}
            />
          </div>
        ))}
      </div>

      <div className="flex items-center gap-6">
        <div className="flex items-center gap-2">
          <Switch id="is_active" name="is_active" defaultChecked={persona?.is_active ?? true} />
          <Label htmlFor="is_active">{t("active")}</Label>
        </div>
        <div className="flex items-center gap-2">
          <Switch
            id="is_minor_appropriate"
            name="is_minor_appropriate"
            defaultChecked={persona?.is_minor_appropriate ?? true}
          />
          <Label htmlFor="is_minor_appropriate">{t("minorAppropriate")}</Label>
        </div>
      </div>

      <Button type="submit">{submitLabel}</Button>
    </form>
  );
}
