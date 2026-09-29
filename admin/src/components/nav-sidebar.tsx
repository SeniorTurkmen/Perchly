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
    <nav className="hidden w-56 shrink-0 border-e border-sidebar-border bg-sidebar p-3 sm:block">
      <div className="mb-6 flex items-center gap-2 px-2 pt-1">
        <span className="flex size-7 items-center justify-center rounded-lg bg-primary text-sm font-semibold text-primary-foreground">
          P
        </span>
        <span className="text-[15px] font-semibold text-sidebar-foreground">
          Perchly
        </span>
      </div>
      <ul className="space-y-0.5">
        {NAV_ITEMS.map(({ href, key, icon: Icon }) => {
          const isActive =
            href === "/" ? pathname === "/" : pathname.startsWith(href);
          return (
            <li key={href}>
              <Link
                href={href}
                className={cn(
                  "flex items-center gap-2.5 rounded-lg px-2.5 py-1.5 text-[13px] font-medium transition-colors",
                  isActive
                    ? "bg-sidebar-accent text-sidebar-accent-foreground"
                    : "text-sidebar-foreground/70 hover:bg-sidebar-accent/60 hover:text-sidebar-foreground",
                )}
              >
                <Icon
                  className={cn(
                    "size-4",
                    isActive ? "text-sidebar-accent-foreground" : "text-sidebar-foreground/50",
                  )}
                />
                {t(key)}
              </Link>
            </li>
          );
        })}
      </ul>
    </nav>
  );
}
