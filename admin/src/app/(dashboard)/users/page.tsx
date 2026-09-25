import Link from "next/link";
import { getTranslations } from "next-intl/server";

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
  const t = await getTranslations("users");
  const tc = await getTranslations("common");

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
        <h1 className="text-2xl font-semibold">{t("title")}</h1>
        <p className="text-muted-foreground">
          {t("subtitle", { count: total })}
        </p>
      </div>

      <form className="flex max-w-sm gap-2">
        <Input
          name="search"
          defaultValue={search}
          placeholder={t("searchPlaceholder")}
        />
        <Button type="submit">{tc("search")}</Button>
      </form>

      <div className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>{t("colUser")}</TableHead>
              <TableHead>{t("colType")}</TableHead>
              <TableHead>{t("colCreated")}</TableHead>
              <TableHead className="text-right">{t("colDetail")}</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {users.length === 0 && (
              <TableRow>
                <TableCell colSpan={4} className="text-center text-muted-foreground">
                  {tc("noResults")}
                </TableCell>
              </TableRow>
            )}
            {users.map((user) => (
              <TableRow key={user.id}>
                <TableCell>
                  <div className="font-medium">
                    {user.display_name ?? user.email ?? t("unnamedUser")}
                  </div>
                  {user.email && (
                    <div className="text-sm text-muted-foreground">{user.email}</div>
                  )}
                </TableCell>
                <TableCell>
                  <Badge variant={user.is_anonymous ? "secondary" : "default"}>
                    {user.is_anonymous ? t("anonymous") : t("registered")}
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
                    {tc("view")}
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
            {tc("previous")}
          </Button>
        ) : (
          <Button
            render={<Link href={pageHref(Math.max(0, offset - PAGE_SIZE))} />}
            variant="outline"
            size="sm"
          >
            {tc("previous")}
          </Button>
        )}
        <span className="text-sm text-muted-foreground">
          {total === 0 ? 0 : offset + 1}-{Math.min(offset + PAGE_SIZE, total)} / {total}
        </span>
        {offset + PAGE_SIZE >= total ? (
          <Button variant="outline" size="sm" disabled>
            {tc("next")}
          </Button>
        ) : (
          <Button
            render={<Link href={pageHref(offset + PAGE_SIZE)} />}
            variant="outline"
            size="sm"
          >
            {tc("next")}
          </Button>
        )}
      </div>
    </div>
  );
}
