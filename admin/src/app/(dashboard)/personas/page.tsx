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
import { adminListLLMModels, adminListPersonas } from "@/lib/backend";

export default async function PersonasPage() {
  const token = await requireSessionToken();
  const [personas, models] = await Promise.all([
    adminListPersonas(token),
    adminListLLMModels(token),
  ]);
  const modelNameById = new Map(models.map((m) => [m.id, m.display_name]));
  const t = await getTranslations("personas");
  const tc = await getTranslations("common");

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold">{t("title")}</h1>
          <p className="text-muted-foreground">
            {t("subtitle", { count: personas.length })}
          </p>
        </div>
        <Button render={<Link href="/personas/new" />}>{t("newPersona")}</Button>
      </div>

      <div className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>{t("fields.name")}</TableHead>
              <TableHead>{t("fields.category")}</TableHead>
              <TableHead>{t("fields.model")}</TableHead>
              <TableHead>{tc("status")}</TableHead>
              <TableHead>{t("fields.sortOrder")}</TableHead>
              <TableHead className="text-end">{tc("edit")}</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {personas.map((persona) => (
              <TableRow key={persona.id}>
                <TableCell>
                  <div className="font-medium">{persona.name}</div>
                  <div className="text-sm text-muted-foreground">{persona.slug}</div>
                </TableCell>
                <TableCell>{persona.category}</TableCell>
                <TableCell className="text-muted-foreground">
                  {persona.llm_model_id
                    ? (modelNameById.get(persona.llm_model_id) ?? t("unknownModel"))
                    : t("defaultModel")}
                </TableCell>
                <TableCell>
                  <Badge variant={persona.is_active ? "default" : "secondary"}>
                    {persona.is_active ? tc("active") : tc("inactive")}
                  </Badge>
                </TableCell>
                <TableCell>{persona.sort_order}</TableCell>
                <TableCell className="text-end">
                  <Button
                    render={<Link href={`/personas/${persona.id}`} />}
                    variant="ghost"
                    size="sm"
                  >
                    {tc("edit")}
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
