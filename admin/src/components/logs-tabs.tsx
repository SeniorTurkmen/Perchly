import Link from "next/link";
import { getTranslations } from "next-intl/server";

import { cn } from "@/lib/utils";

export async function LogsTabs({ active }: { active: "requests" | "admin" }) {
  const t = await getTranslations("logs");
  const tabs = [
    { key: "requests", label: t("tabRequests"), href: "/logs" },
    { key: "admin", label: t("tabAdmin"), href: "/logs/admin" },
  ] as const;

  return (
    <div className="flex gap-1 border-b">
      {tabs.map((tab) => (
        <Link
          key={tab.key}
          href={tab.href}
          className={cn(
            "-mb-px border-b-2 px-3 py-2 text-sm",
            active === tab.key
              ? "border-primary font-medium text-foreground"
              : "border-transparent text-muted-foreground hover:text-foreground",
          )}
        >
          {tab.label}
        </Link>
      ))}
    </div>
  );
}
