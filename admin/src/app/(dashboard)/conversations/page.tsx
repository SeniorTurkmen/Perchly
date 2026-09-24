import Link from "next/link";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { requireSessionToken } from "@/lib/auth";
import { adminListConversations } from "@/lib/backend";

const PAGE_SIZE = 20;

const dateFormatter = new Intl.DateTimeFormat("tr-TR", {
  dateStyle: "medium",
  timeStyle: "short",
});

function truncate(text: string, max: number) {
  return text.length > max ? `${text.slice(0, max)}…` : text;
}

export default async function ConversationsPage({
  searchParams,
}: {
  searchParams: Promise<{ search?: string; user_id?: string; offset?: string }>;
}) {
  const params = await searchParams;
  const token = await requireSessionToken();

  const search = params.search ?? "";
  const userId = params.user_id;
  const offset = Math.max(0, Number(params.offset) || 0);

  const { conversations, total } = await adminListConversations(token, {
    search,
    userId,
    limit: PAGE_SIZE,
    offset,
  });

  const pageHref = (nextOffset: number) => {
    const qs = new URLSearchParams();
    if (search) qs.set("search", search);
    if (userId) qs.set("user_id", userId);
    qs.set("offset", String(nextOffset));
    return `/conversations?${qs}`;
  };

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold">Konuşmalar</h1>
        <p className="text-muted-foreground">
          {total} konuşma{userId ? " — bu kullanıcıya göre filtrelendi" : ""} — mesaj
          içeriğine göre ara.
        </p>
      </div>

      <form className="flex max-w-sm gap-2">
        <Input name="search" defaultValue={search} placeholder="Mesaj içeriğinde ara..." />
        {userId && <input type="hidden" name="user_id" value={userId} />}
        <Button type="submit">Ara</Button>
      </form>

      <div className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Persona</TableHead>
              <TableHead>Son mesaj</TableHead>
              <TableHead>Zaman</TableHead>
              <TableHead className="text-right">Detay</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {conversations.length === 0 && (
              <TableRow>
                <TableCell colSpan={4} className="text-center text-muted-foreground">
                  Sonuç bulunamadı.
                </TableCell>
              </TableRow>
            )}
            {conversations.map((conv) => (
              <TableRow key={conv.id}>
                <TableCell>
                  <div className="font-medium">{conv.persona.name}</div>
                  <Link
                    href={`/users/${conv.user_id}`}
                    className="text-sm text-muted-foreground hover:underline"
                  >
                    Kullanıcıyı görüntüle
                  </Link>
                </TableCell>
                <TableCell className="max-w-md">
                  {conv.last_message ? (
                    <span className="text-sm">
                      <span className="text-muted-foreground">
                        {conv.last_message.role === "user" ? "Kullanıcı: " : "Persona: "}
                      </span>
                      {truncate(conv.last_message.content, 80)}
                    </span>
                  ) : (
                    <span className="text-muted-foreground">—</span>
                  )}
                </TableCell>
                <TableCell>
                  {conv.last_message
                    ? dateFormatter.format(new Date(conv.last_message.created_at))
                    : dateFormatter.format(new Date(conv.created_at))}
                </TableCell>
                <TableCell className="text-right">
                  <Button
                    render={<Link href={`/conversations/${conv.id}`} />}
                    variant="ghost"
                    size="sm"
                  >
                    Görüntüle
                  </Button>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>

      <div className="flex items-center justify-between">
        {offset === 0 ? (
          <Button variant="outline" size="sm" disabled>
            Önceki
          </Button>
        ) : (
          <Button
            render={<Link href={pageHref(Math.max(0, offset - PAGE_SIZE))} />}
            variant="outline"
            size="sm"
          >
            Önceki
          </Button>
        )}
        <span className="text-sm text-muted-foreground">
          {total === 0 ? 0 : offset + 1}-{Math.min(offset + PAGE_SIZE, total)} / {total}
        </span>
        {offset + PAGE_SIZE >= total ? (
          <Button variant="outline" size="sm" disabled>
            Sonraki
          </Button>
        ) : (
          <Button
            render={<Link href={pageHref(offset + PAGE_SIZE)} />}
            variant="outline"
            size="sm"
          >
            Sonraki
          </Button>
        )}
      </div>
    </div>
  );
}
