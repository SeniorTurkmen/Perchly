"use client";

import {
  ClipboardList,
  Gauge,
  GitBranch,
  KeyRound,
  LayoutDashboard,
  MessagesSquare,
  ScrollText,
  Sparkles,
  Users,
} from "lucide-react";
import { useTranslations } from "next-intl";
import Link from "next/link";
import { usePathname } from "next/navigation";

import { cn } from "@/lib/utils";

const NAV_ITEMS = [
  { href: "/", key: "dashboard", icon: LayoutDashboard },
  { href: "/users", key: "users", icon: Users },
  { href: "/personas", key: "personas", icon: Sparkles },
  { href: "/conversations", key: "conversations", icon: MessagesSquare },
  { href: "/quotas", key: "quotas", icon: Gauge },
  { href: "/providers", key: "providers", icon: KeyRound },
  { href: "/onboarding", key: "onboarding", icon: ClipboardList },
  { href: "/logs", key: "logs", icon: ScrollText },
  { href: "/engineering", key: "engineering", icon: GitBranch },
] as const;

export function NavSidebar() {
  const pathname = usePathname();
  const t = useTranslations("nav");

  return (
    <nav className="hidden w-56 shrink-0 border-r bg-muted/20 p-4 sm:block">
      <div className="mb-6 px-2 text-lg font-semibold">Perchly</div>
      <ul className="space-y-1">
        {NAV_ITEMS.map(({ href, key, icon: Icon }) => {
          const isActive =
            href === "/" ? pathname === "/" : pathname.startsWith(href);
          return (
            <li key={href}>
              <Link
                href={href}
                className={cn(
                  "flex items-center gap-2 rounded-md px-2 py-1.5 text-sm transition-colors",
                  isActive
                    ? "bg-primary text-primary-foreground"
                    : "text-muted-foreground hover:bg-muted hover:text-foreground",
                )}
              >
                <Icon className="size-4" />
                {t(key)}
              </Link>
            </li>
          );
        })}
      </ul>
    </nav>
  );
}
