import Link from "next/link";
import { getTranslations } from "next-intl/server";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { requireSessionToken } from "@/lib/auth";
import { adminListLLMCredentials } from "@/lib/backend";

// Brand/product names — never translated.
const PROVIDER_LABELS: Record<string, string> = {
  openai: "OpenAI",
  anthropic: "Anthropic",
  gemini: "Google Gemini",
  huggingface: "Hugging Face",
  deepseek: "DeepSeek",
};

export default async function ProvidersPage() {
  const token = await requireSessionToken();
  const credentials = await adminListLLMCredentials(token);
  const t = await getTranslations("providers");
  const tc = await getTranslations("common");

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold">{t("title")}</h1>
          <p className="text-muted-foreground">
            {t("subtitle", { count: credentials.length })}
          </p>
        </div>
        <Button render={<Link href="/providers/new" />}>{t("addCredential")}</Button>
      </div>

      <div className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>{t("provider")}</TableHead>
              <TableHead>{t("label")}</TableHead>
              <TableHead>{t("token")}</TableHead>
              <TableHead>{tc("status")}</TableHead>
              <TableHead className="text-right">{t("manage")}</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {credentials.length === 0 && (
              <TableRow>
                <TableCell colSpan={5} className="text-center text-muted-foreground">
                  {t("noCredentialsYet")}
                </TableCell>
              </TableRow>
            )}
            {credentials.map((credential) => (
              <TableRow key={credential.id}>
                <TableCell>
                  {PROVIDER_LABELS[credential.provider] ?? credential.provider}
                </TableCell>
                <TableCell>{credential.label}</TableCell>
                <TableCell className="font-mono text-sm text-muted-foreground">
                  {credential.api_key_preview}
                </TableCell>
                <TableCell>
                  <Badge variant={credential.is_active ? "default" : "secondary"}>
                    {credential.is_active ? tc("active") : tc("inactive")}
                  </Badge>
                </TableCell>
                <TableCell className="text-right">
                  <Button
                    render={<Link href={`/providers/${credential.id}`} />}
                    variant="ghost"
                    size="sm"
                  >
                    {t("manage")}
                  </Button>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>
    </div>
  );
}
