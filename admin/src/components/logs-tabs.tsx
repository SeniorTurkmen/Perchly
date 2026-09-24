import Link from "next/link";

import { cn } from "@/lib/utils";

export function LogsTabs({ active }: { active: "requests" | "admin" }) {
  const tabs = [
    { key: "requests", label: "Kullanıcı İstekleri", href: "/logs" },
    { key: "admin", label: "Admin İşlemleri", href: "/logs/admin" },
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
