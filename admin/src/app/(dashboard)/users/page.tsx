import Link from "next/link";

import { LocalDateTime } from "@/components/local-date-time";
import { Badge } from "@/components/ui/badge";
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
import { adminListUsers } from "@/lib/backend";

const PAGE_SIZE = 20;

export default async function UsersPage({
  searchParams,
}: {
  searchParams: Promise<{ search?: string; offset?: string }>;
}) {
  const params = await searchParams;
  const token = await requireSessionToken();

  const search = params.search ?? "";
  const offset = Math.max(0, Number(params.offset) || 0);

  const { users, total } = await adminListUsers(token, {
    search,
    limit: PAGE_SIZE,
    offset,
  });

  const pageHref = (nextOffset: number) => {
    const qs = new URLSearchParams();
    if (search) qs.set("search", search);
    qs.set("offset", String(nextOffset));
    return `/users?${qs}`;
  };

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold">Kullanıcılar</h1>
        <p className="text-muted-foreground">
          {total} kullanıcı — e-posta veya görünen ada göre ara.
        </p>
      </div>

      <form className="flex max-w-sm gap-2">
        <Input
          name="search"
          defaultValue={search}
          placeholder="E-posta veya ad ara..."
        />
        <Button type="submit">Ara</Button>
      </form>

      <div className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Kullanıcı</TableHead>
              <TableHead>Tür</TableHead>
              <TableHead>Oluşturulma</TableHead>
              <TableHead className="text-right">Detay</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {users.length === 0 && (
              <TableRow>
                <TableCell colSpan={4} className="text-center text-muted-foreground">
                  Sonuç bulunamadı.
                </TableCell>
              </TableRow>
            )}
            {users.map((user) => (
              <TableRow key={user.id}>
                <TableCell>
                  <div className="font-medium">
                    {user.display_name ?? user.email ?? "İsimsiz kullanıcı"}
                  </div>
                  {user.email && (
                    <div className="text-sm text-muted-foreground">{user.email}</div>
                  )}
                </TableCell>
                <TableCell>
                  <Badge variant={user.is_anonymous ? "secondary" : "default"}>
                    {user.is_anonymous ? "Anonim" : "Kayıtlı"}
                  </Badge>
                </TableCell>
                <TableCell>
                  <LocalDateTime value={user.created_at} />
                </TableCell>
                <TableCell className="text-right">
                  <Button
                    render={<Link href={`/users/${user.id}`} />}
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
