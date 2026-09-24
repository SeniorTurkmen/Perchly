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
import { adminListLLMCredentials } from "@/lib/backend";

const PROVIDER_LABELS: Record<string, string> = {
  openai: "OpenAI",
  anthropic: "Anthropic",
  gemini: "Google Gemini",
  huggingface: "Hugging Face",
};

export default async function ProvidersPage() {
  const token = await requireSessionToken();
  const credentials = await adminListLLMCredentials(token);

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold">AI Sağlayıcıları</h1>
          <p className="text-muted-foreground">
            {credentials.length} kayıtlı kimlik bilgisi — token&apos;lar veritabanında
            şifrelenmiş saklanır.
          </p>
        </div>
        <Button render={<Link href="/providers/new" />}>Kimlik bilgisi ekle</Button>
      </div>

      <div className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Sağlayıcı</TableHead>
              <TableHead>Etiket</TableHead>
              <TableHead>Token</TableHead>
              <TableHead>Durum</TableHead>
              <TableHead className="text-right">Yönet</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {credentials.length === 0 && (
              <TableRow>
                <TableCell colSpan={5} className="text-center text-muted-foreground">
                  Henüz bir kimlik bilgisi eklenmedi.
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
                    {credential.is_active ? "Aktif" : "Pasif"}
                  </Badge>
                </TableCell>
                <TableCell className="text-right">
                  <Button
                    render={<Link href={`/providers/${credential.id}`} />}
                    variant="ghost"
                    size="sm"
                  >
                    Yönet
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
