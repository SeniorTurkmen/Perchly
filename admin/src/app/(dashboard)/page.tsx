import { getLocale, getTranslations } from "next-intl/server";

import {
  Card,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { requireSessionToken } from "@/lib/auth";
import { adminMetrics } from "@/lib/backend";

export default async function DashboardHomePage() {
  const token = await requireSessionToken();
  const metrics = await adminMetrics(token);
  const t = await getTranslations("dashboard");
  const locale = await getLocale();

  const numberFormatter = new Intl.NumberFormat(locale);
  const percentFormatter = new Intl.NumberFormat(locale, {
    style: "percent",
    maximumFractionDigits: 1,
  });

  const stats = [
    { label: t("totalUsers"), value: numberFormatter.format(metrics.total_users) },
    { label: t("newUsersToday"), value: numberFormatter.format(metrics.new_users_today) },
    { label: t("messagesToday"), value: numberFormatter.format(metrics.messages_sent_today) },
    { label: t("activeConversationsToday"), value: numberFormatter.format(metrics.active_conversations_today) },
    { label: t("errorRateToday"), value: percentFormatter.format(metrics.error_rate_today) },
  ];

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold">{t("title")}</h1>
        <p className="text-muted-foreground">
          {t("welcome")}
        </p>
      </div>
      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {stats.map((stat) => (
          <Card key={stat.label}>
            <CardHeader>
              <CardDescription>{stat.label}</CardDescription>
              <CardTitle className="text-3xl">{stat.value}</CardTitle>
            </CardHeader>
          </Card>
        ))}
      </div>
    </div>
  );
}
