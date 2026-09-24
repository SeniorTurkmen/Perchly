import Link from "next/link";

import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { LogsTabs } from "@/components/logs-tabs";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { requireSessionToken } from "@/lib/auth";
import { adminListActivity } from "@/lib/backend";

const PAGE_SIZE = 50;

const dateFormatter = new Intl.DateTimeFormat("tr-TR", {
  dateStyle: "medium",
  timeStyle: "medium",
});

function actionVariant(action: string): "default" | "secondary" | "destructive" {
  if (action.includes("delete")) return "destructive";
  if (action.includes("login")) return "secondary";
  return "default";
}

export default async function AdminActivityPage({
  searchParams,
}: {
  searchParams: Promise<{ admin_user_id?: string; offset?: string }>;
}) {
  const params = await searchParams;
  const token = await requireSessionToken();

  const adminUserId = params.admin_user_id;
  const offset = Math.max(0, Number(params.offset) || 0);

  const { entries, total } = await adminListActivity(token, {
    adminUserId,
    limit: PAGE_SIZE,
    offset,
  });

  const pageHref = (nextOffset: number) => {
    const qs = new URLSearchParams();
    if (adminUserId) qs.set("admin_user_id", adminUserId);
    qs.set("offset", String(nextOffset));
    return `/logs/admin?${qs}`;
  };

  const filterByAdminHref = (id: string) => `/logs/admin?admin_user_id=${id}`;

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold">Loglar</h1>
        <p className="text-muted-foreground">
          {total} admin işlemi{adminUserId ? " — bu admine göre filtrelendi" : ""}.
        </p>
      </div>

      <LogsTabs active="admin" />

      {adminUserId && (
        <Button render={<Link href="/logs/admin" />} variant="ghost" size="sm">
          Admin filtresini kaldır
        </Button>
      )}

      <div className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Admin</TableHead>
              <TableHead>İşlem</TableHead>
              <TableHead>Hedef</TableHead>
              <TableHead>Zaman</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {entries.length === 0 && (
              <TableRow>
                <TableCell colSpan={4} className="text-center text-muted-foreground">
                  Sonuç bulunamadı.
                </TableCell>
              </TableRow>
            )}
            {entries.map((entry) => (
              <TableRow key={entry.id}>
                <TableCell>
                  <Link href={filterByAdminHref(entry.admin_user_id)} className="hover:underline">
                    {entry.admin_email}
                  </Link>
                </TableCell>
                <TableCell>
                  <Badge variant={actionVariant(entry.action)}>{entry.action}</Badge>
                </TableCell>
                <TableCell className="text-sm">
                  <div className="text-muted-foreground">
                    {entry.target_type}
                    {entry.target_id ? ` · ${entry.target_id.slice(0, 8)}…` : ""}
                  </div>
                  {entry.detail && (
                    <details className="mt-1">
                      <summary className="cursor-pointer text-xs text-muted-foreground">
                        Detay
                      </summary>
                      <pre className="mt-1 max-h-72 w-full max-w-xl overflow-y-auto rounded border bg-muted/30 p-2 text-xs whitespace-pre-wrap break-all">
                        {JSON.stringify(entry.detail, null, 2)}
                      </pre>
                    </details>
                  )}
                </TableCell>
                <TableCell className="text-sm text-muted-foreground">
                  {dateFormatter.format(new Date(entry.created_at))}
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
