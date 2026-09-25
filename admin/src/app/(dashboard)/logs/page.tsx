import { getTranslations } from "next-intl/server";
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
import { LogsTabs } from "@/components/logs-tabs";
import { requireSessionToken } from "@/lib/auth";
import { adminListLogs } from "@/lib/backend";
import { cn } from "@/lib/utils";

const PAGE_SIZE = 50;

function statusVariant(status: number): "default" | "secondary" | "destructive" {
  if (status >= 500) return "destructive";
  if (status >= 400) return "secondary";
  return "default";
}

// The detail panels below exist to answer "how was this request
// actually made" without needing a separate tool — but a single huge
// value (a long token, a big payload) can make the whole block
// unreadable, so every string is capped before display and the
// rendered block itself is capped too, as a second backstop.
const MAX_STRING_LENGTH = 500;
const MAX_BLOCK_LENGTH = 6000;

function truncateStrings(value: unknown, seen: WeakSet<object> = new WeakSet()): unknown {
  if (typeof value === "string") {
    return value.length > MAX_STRING_LENGTH
      ? `${value.slice(0, MAX_STRING_LENGTH)}… (+${value.length - MAX_STRING_LENGTH} karakter)`
      : value;
  }
  if (Array.isArray(value)) {
    return value.map((item) => truncateStrings(item, seen));
  }
  if (value && typeof value === "object") {
    if (seen.has(value)) return "[circular]";
    seen.add(value);
    return Object.fromEntries(
      Object.entries(value).map(([key, val]) => [key, truncateStrings(val, seen)]),
    );
  }
  return value;
}

function capBlock(text: string): string {
  return text.length > MAX_BLOCK_LENGTH
    ? `${text.slice(0, MAX_BLOCK_LENGTH)}\n… (kesildi)`
    : text;
}

function formatJson(value: unknown): string {
  return capBlock(JSON.stringify(truncateStrings(value), null, 2));
}

// response_body is stored as raw text (it might be an SSE stream, not
// just JSON) — pretty-print it when it happens to parse as JSON, and
// fall back to showing it verbatim otherwise (an SSE chunk, a plain
// error page, ...).
function formatResponseBody(text: string | null): string | null {
  if (!text) return null;
  try {
    return formatJson(JSON.parse(text));
  } catch {
    return capBlock(text);
  }
}

const codeBlockClass =
  "mt-1 max-h-72 w-full max-w-xl overflow-y-auto rounded border bg-muted/30 p-2 text-xs whitespace-pre-wrap break-all";

export default async function LogsPage({
  searchParams,
}: {
  searchParams: Promise<{
    search?: string;
    status_min?: string;
    user_id?: string;
    offset?: string;
  }>;
}) {
  const params = await searchParams;
  const token = await requireSessionToken();
  const t = await getTranslations("logs");
  const tc = await getTranslations("common");

  const STATUS_FILTERS = [
    { label: tc("all"), value: "0" },
    { label: "4xx+", value: "400" },
    { label: "5xx", value: "500" },
  ] as const;

  const search = params.search ?? "";
  const statusMin = Number(params.status_min) || 0;
  const userId = params.user_id;
  const offset = Math.max(0, Number(params.offset) || 0);

  const { logs, total } = await adminListLogs(token, {
    search,
    statusMin,
    userId,
    limit: PAGE_SIZE,
    offset,
  });

  const pageHref = (overrides: { offset?: number; statusMin?: number }) => {
    const qs = new URLSearchParams();
    if (search) qs.set("search", search);
    const nextStatusMin = overrides.statusMin ?? statusMin;
    if (nextStatusMin) qs.set("status_min", String(nextStatusMin));
    if (userId) qs.set("user_id", userId);
    qs.set("offset", String(overrides.offset ?? offset));
    return `/logs?${qs}`;
  };

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold">{t("title")}</h1>
        <p className="text-muted-foreground">
          {t("subtitleRequests", { count: total })}{userId ? ` — ${t("filteredByUser")}` : ""}
        </p>
      </div>

      <LogsTabs active="requests" />

      <div className="flex flex-wrap items-center gap-4">
        <form className="flex max-w-sm gap-2">
          <Input name="search" defaultValue={search} placeholder={t("searchRoutePlaceholder")} />
          {statusMin ? <input type="hidden" name="status_min" value={statusMin} /> : null}
          {userId ? <input type="hidden" name="user_id" value={userId} /> : null}
          <Button type="submit">{tc("search")}</Button>
        </form>
        <div className="flex gap-1">
          {STATUS_FILTERS.map((filter) => (
            <Button
              key={filter.value}
              render={<Link href={pageHref({ statusMin: Number(filter.value), offset: 0 })} />}
              variant={statusMin === Number(filter.value) ? "default" : "outline"}
              size="sm"
            >
              {filter.label}
            </Button>
          ))}
        </div>
        {userId && (
          <Button render={<Link href="/logs" />} variant="ghost" size="sm">
            {t("clearUserFilter")}
          </Button>
        )}
      </div>

      <div className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>{t("request")}</TableHead>
              <TableHead>{t("colUser")}</TableHead>
              <TableHead>{tc("status")}</TableHead>
              <TableHead>{t("duration")}</TableHead>
              <TableHead>{t("platform")}</TableHead>
              <TableHead>{t("time")}</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {logs.length === 0 && (
              <TableRow>
                <TableCell colSpan={6} className="text-center text-muted-foreground">
                  {tc("noResults")}
                </TableCell>
              </TableRow>
            )}
            {logs.map((log) => (
              <TableRow key={log.id}>
                <TableCell>
                  <span className="font-mono text-xs text-muted-foreground">{log.method}</span>{" "}
                  <span className="font-medium">{log.route_pattern || log.path}</span>
                  {(log.body || log.query_params || log.request_headers) && (
                    <details className="mt-1">
                      <summary className="cursor-pointer text-xs text-muted-foreground">
                        {t("howWasItSent")}
                      </summary>
                      <pre className={codeBlockClass}>
                        {formatJson({
                          query: log.query_params,
                          body: log.body,
                          headers: log.request_headers,
                        })}
                      </pre>
                    </details>
                  )}
                  {(log.response_headers || log.response_body) && (
                    <details className="mt-1">
                      <summary className="cursor-pointer text-xs text-muted-foreground">
                        {t("whatCameBack")}
                      </summary>
                      <pre className={codeBlockClass}>
                        {formatJson({ headers: log.response_headers })}
                      </pre>
                      {log.response_body && (
                        <pre className={codeBlockClass}>{formatResponseBody(log.response_body)}</pre>
                      )}
                    </details>
                  )}
                </TableCell>
                <TableCell className="text-sm">
                  {log.user_id ? (
                    <Link href={`/users/${log.user_id}`} className="hover:underline">
                      {log.user_display_name ?? log.user_email ?? t("anonymousUser")}
                    </Link>
                  ) : (
                    <span className="text-muted-foreground">—</span>
                  )}
                </TableCell>
                <TableCell>
                  <Badge variant={statusVariant(log.status_code)}>{log.status_code}</Badge>
                </TableCell>
                <TableCell className={cn(log.duration_ms > 1000 && "text-destructive")}>
                  {log.duration_ms} ms
                </TableCell>
                <TableCell className="text-sm text-muted-foreground">
                  {[log.platform, log.client_version].filter(Boolean).join(" · ") || "—"}
                </TableCell>
                <TableCell className="text-sm text-muted-foreground">
                  <LocalDateTime value={log.created_at} style="medium" />
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
            render={<Link href={pageHref({ offset: Math.max(0, offset - PAGE_SIZE) })} />}
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
            render={<Link href={pageHref({ offset: offset + PAGE_SIZE })} />}
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
