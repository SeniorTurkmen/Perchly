import Link from "next/link";

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

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold">Personalar</h1>
          <p className="text-muted-foreground">
            {personas.length} persona — sistem promptu dahil tam düzenleme.
          </p>
        </div>
        <Button render={<Link href="/personas/new" />}>Yeni persona</Button>
      </div>

      <div className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Persona</TableHead>
              <TableHead>Kategori</TableHead>
              <TableHead>Model</TableHead>
              <TableHead>Durum</TableHead>
              <TableHead>Sıra</TableHead>
              <TableHead className="text-right">Düzenle</TableHead>
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
                    ? (modelNameById.get(persona.llm_model_id) ?? "bilinmeyen model")
                    : "Varsayılan"}
                </TableCell>
                <TableCell>
                  <Badge variant={persona.is_active ? "default" : "secondary"}>
                    {persona.is_active ? "Aktif" : "Pasif"}
                  </Badge>
                </TableCell>
                <TableCell>{persona.sort_order}</TableCell>
                <TableCell className="text-right">
                  <Button
                    render={<Link href={`/personas/${persona.id}`} />}
                    variant="ghost"
                    size="sm"
                  >
                    Düzenle
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
