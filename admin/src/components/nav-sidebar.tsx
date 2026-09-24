"use client";

import {
  Gauge,
  GitBranch,
  LayoutDashboard,
  MessagesSquare,
  ScrollText,
  Sparkles,
  Users,
} from "lucide-react";
import Link from "next/link";
import { usePathname } from "next/navigation";

import { cn } from "@/lib/utils";

const NAV_ITEMS = [
  { href: "/", label: "Panel", icon: LayoutDashboard },
  { href: "/users", label: "Kullanıcılar", icon: Users },
  { href: "/personas", label: "Personalar", icon: Sparkles },
  { href: "/conversations", label: "Konuşmalar", icon: MessagesSquare },
  { href: "/quotas", label: "Kota & Kredi", icon: Gauge },
  { href: "/logs", label: "Loglar", icon: ScrollText },
  { href: "/engineering", label: "Mühendislik", icon: GitBranch },
] as const;

export function NavSidebar() {
  const pathname = usePathname();

  return (
    <nav className="hidden w-56 shrink-0 border-r bg-muted/20 p-4 sm:block">
      <div className="mb-6 px-2 text-lg font-semibold">Perchly</div>
      <ul className="space-y-1">
        {NAV_ITEMS.map(({ href, label, icon: Icon }) => {
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
                {label}
              </Link>
            </li>
          );
        })}
      </ul>
    </nav>
  );
}
