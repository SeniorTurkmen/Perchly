import type { ReactNode } from "react";

import { LanguageSwitcher } from "@/components/language-switcher";
import { LogoutButton } from "@/components/logout-button";
import { NavSidebar } from "@/components/nav-sidebar";
import { requireAdmin } from "@/lib/auth";

export default async function DashboardLayout({
  children,
}: {
  children: ReactNode;
}) {
  const admin = await requireAdmin();

  return (
    <div className="flex min-h-screen">
      <NavSidebar />
      <div className="flex flex-1 flex-col">
        <header className="flex items-center justify-between border-b px-6 py-3">
          <span className="text-sm text-muted-foreground">{admin.email}</span>
          <div className="flex items-center gap-2">
            <LanguageSwitcher />
            <LogoutButton />
          </div>
        </header>
        <main className="flex-1 p-6">{children}</main>
      </div>
    </div>
  );
}
