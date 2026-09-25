import Link from "next/link";
import { getTranslations } from "next-intl/server";
import { notFound } from "next/navigation";

import { LocalDateTime } from "@/components/local-date-time";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { requireSessionToken } from "@/lib/auth";
import {
  AdminApiError,
  adminGetUser,
  adminListPersonas,
} from "@/lib/backend";

import { setCreditsAction, setQuotaAction } from "./actions";

export default async function UserDetailPage({
  params,
  searchParams,
}: {
  params: Promise<{ id: string }>;
  searchParams: Promise<{ error?: string }>;
}) {
  const { id } = await params;
  const { error } = await searchParams;
  const token = await requireSessionToken();
  const t = await getTranslations("users");

  let detail;
  try {
    detail = await adminGetUser(token, id);
  } catch (err) {
    if (err instanceof AdminApiError && err.status === 404) notFound();
    throw err;
  }

  const personas = await adminListPersonas(token);
  const personaMap = new Map(personas.map((p) => [p.id, p]));
  const personasWithoutQuota = personas.filter(
    (p) => !detail.quotas.some((q) => q.persona_id === p.id),
  );

  const boundSetQuota = setQuotaAction.bind(null, id);
  const boundSetCredits = setCreditsAction.bind(null, id);

  return (
    <div className="space-y-6">
      {error && (
        <div className="rounded-md border border-destructive/50 bg-destructive/10 px-4 py-3 text-sm text-destructive">
          {error}
        </div>
      )}

      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold">
            {detail.user.display_name ?? detail.user.email ?? t("unnamedUser")}
          </h1>
          <p className="text-muted-foreground">{detail.user.id}</p>
        </div>
        <div className="flex gap-2">
          <Button render={<Link href={`/conversations?user_id=${detail.user.id}`} />} variant="outline">
            {t("viewConversations")}
          </Button>
          <Button render={<Link href={`/logs?user_id=${detail.user.id}`} />} variant="outline">
            {t("viewLogs")}
          </Button>
        </div>
      </div>

      <div className="grid gap-4 md:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>{t("accountTitle")}</CardTitle>
          </CardHeader>
          <CardContent className="space-y-2 text-sm">
            <div className="flex justify-between">
              <span className="text-muted-foreground">{t("emailLabel")}</span>
              <span>{detail.user.email ?? "—"}</span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-muted-foreground">{t("colType")}</span>
              <Badge variant={detail.user.is_anonymous ? "secondary" : "default"}>
                {detail.user.is_anonymous ? t("anonymous") : t("registered")}
              </Badge>
            </div>
            <div className="flex justify-between">
              <span className="text-muted-foreground">{t("colCreated")}</span>
              <span>
                <LocalDateTime value={detail.user.created_at} />
              </span>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>{t("creditTitle")}</CardTitle>
            <CardDescription>{t("currentBalance", { amount: detail.credits })}</CardDescription>
          </CardHeader>
          <CardContent>
            <form action={boundSetCredits} className="flex items-end gap-2">
              <div className="space-y-2">
                <Label htmlFor="amount">{t("newBalance")}</Label>
                <Input
                  id="amount"
                  name="amount"
                  type="number"
                  min={0}
                  defaultValue={detail.credits}
                  className="w-32"
                />
              </div>
              <Button type="submit">{t("update")}</Button>
            </form>
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>{t("quotaTitle")}</CardTitle>
          <CardDescription>{t("quotaDescription")}</CardDescription>
        </CardHeader>
        <CardContent className="space-y-6">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>{t("persona")}</TableHead>
                <TableHead>{t("usedToday")}</TableHead>
                <TableHead>{t("dailyLimit")}</TableHead>
                <TableHead>{t("lastReset")}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {detail.quotas.length === 0 && (
                <TableRow>
                  <TableCell colSpan={4} className="text-center text-muted-foreground">
                    {t("noQuotaYet")}
                  </TableCell>
                </TableRow>
              )}
              {detail.quotas.map((quota) => (
                <TableRow key={quota.persona_id}>
                  <TableCell>
                    {personaMap.get(quota.persona_id)?.name ?? quota.persona_id}
                  </TableCell>
                  <TableCell>{quota.message_count_today}</TableCell>
                  <TableCell>
                    <form action={boundSetQuota} className="flex items-center gap-2">
                      <input type="hidden" name="persona_id" value={quota.persona_id} />
                      <Input
                        name="daily_limit"
                        type="number"
                        min={0}
                        defaultValue={quota.daily_limit}
                        className="w-20"
                      />
                      <Button type="submit" size="sm" variant="outline">
                        {t("save")}
                      </Button>
                    </form>
                  </TableCell>
                  <TableCell>
                    <LocalDateTime value={quota.last_reset_at} />
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>

          {personasWithoutQuota.length > 0 && (
            <div className="border-t pt-4">
              <p className="mb-2 text-sm font-medium">
                {t("addCustomLimit")}
              </p>
              <form action={boundSetQuota} className="flex items-end gap-2">
                <div className="space-y-2">
                  <Label htmlFor="new-persona">{t("persona")}</Label>
                  <select
                    id="new-persona"
                    name="persona_id"
                    className="h-9 rounded-md border bg-background px-3 text-sm"
                  >
                    {personasWithoutQuota.map((p) => (
                      <option key={p.id} value={p.id}>
                        {p.name}
                      </option>
                    ))}
                  </select>
                </div>
                <div className="space-y-2">
                  <Label htmlFor="new-limit">{t("dailyLimit")}</Label>
                  <Input
                    id="new-limit"
                    name="daily_limit"
                    type="number"
                    min={0}
                    defaultValue={20}
                    className="w-24"
                  />
                </div>
                <Button type="submit" variant="outline">
                  {t("add")}
                </Button>
              </form>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
