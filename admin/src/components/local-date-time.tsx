"use client";

import { useLocale } from "next-intl";
import { useSyncExternalStore } from "react";

// Every timestamp from the backend is UTC (Postgres TIMESTAMPTZ, server
// runs in Etc/UTC — see backend's migrations) serialized as an ISO 8601
// string with an explicit offset. Formatting it with Intl.DateTimeFormat
// on the SERVER (as a plain server component would) uses the Next.js
// server's own timezone, not the admin's — wrong the moment server and
// viewer aren't in the same timezone. This must run client-side.
const STYLES = {
  short: { dateStyle: "medium", timeStyle: "short" },
  medium: { dateStyle: "medium", timeStyle: "medium" },
} as const;

// This value never changes after mount (for a given value/style), so
// there's nothing to subscribe to — useSyncExternalStore is used here
// purely for its server/client snapshot split, not for reacting to
// updates.
function subscribe() {
  return () => {};
}

// Renders `value` (an ISO 8601 UTC timestamp) in the viewer's own local
// timezone. getServerSnapshot returns a stable placeholder so SSR output
// and the client's first hydration pass match exactly (React requires
// this); once hydrated, React re-checks getSnapshot, which runs in the
// browser and formats with the browser's actual timezone, and swaps in
// the real value. Deliberately not a useEffect+setState — that formats
// eagerly on the server too (using the server's timezone during SSR),
// which is the exact bug this component exists to avoid.
export function LocalDateTime({
  value,
  style = "short",
}: {
  value: string;
  style?: keyof typeof STYLES;
}) {
  const locale = useLocale();
  const formatted = useSyncExternalStore(
    subscribe,
    () => new Intl.DateTimeFormat(locale, STYLES[style]).format(new Date(value)),
    () => null,
  );

  return <span suppressHydrationWarning>{formatted ?? "…"}</span>;
}
