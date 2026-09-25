import { getLocale, getTranslations } from "next-intl/server";

import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { DistributionBar } from "@/components/distribution-bar";
import { requireSessionToken } from "@/lib/auth";
import { adminGetOnboardingInsights } from "@/lib/backend";

const AGE_RANGE_ORDER = ["under18", "age18to24", "age25to34", "age35plus"];
const MOOD_PREFERENCE_ORDER = ["motivation", "dailyChat", "hobbyTalk", "skipped", "none"];

function rate(count: number, total: number, formatter: Intl.NumberFormat): string {
  return total > 0 ? formatter.format(count / total) : "—";
}

function sortByKnownOrder(counts: Record<string, number>, order: string[]) {
  const known = order
    .filter((key) => key in counts)
    .map((key) => ({ key, count: counts[key] }));
  const unknown = Object.entries(counts)
    .filter(([key]) => !order.includes(key))
    .map(([key, count]) => ({ key, count }));
  return [...known, ...unknown];
}

export default async function OnboardingPage() {
  const token = await requireSessionToken();
  const insights = await adminGetOnboardingInsights(token);
  const t = await getTranslations("onboarding");
  const locale = await getLocale();
  const percentFormatter = new Intl.NumberFormat(locale, {
    style: "percent",
    maximumFractionDigits: 1,
  });

  const AGE_RANGE_LABELS: Record<string, string> = {
    under18: t("ageRange.under18"),
    age18to24: t("ageRange.age18to24"),
    age25to34: t("ageRange.age25to34"),
    age35plus: t("ageRange.age35plus"),
  };
  const MOOD_PREFERENCE_LABELS: Record<string, string> = {
    motivation: t("moodPreference.motivation"),
    dailyChat: t("moodPreference.dailyChat"),
    hobbyTalk: t("moodPreference.hobbyTalk"),
    skipped: t("moodPreference.skipped"),
    none: t("moodPreference.none"),
  };

  const {
    total_users: totalUsers,
    completed_onboarding: completed,
    minor_count: minorCount,
    notifications_granted_count: notificationsCount,
    preferred_name_set_count: preferredNameCount,
    skip_hitap_count: skipHitapCount,
    age_range_counts: ageRangeCounts,
    mood_preference_counts: moodPreferenceCounts,
    top_selected_personas: topPersonas,
  } = insights;

  const stats = [
    {
      label: t("stats.completed"),
      value: `${completed} / ${totalUsers}`,
      hint: rate(completed, totalUsers, percentFormatter),
    },
    { label: t("stats.minors"), value: String(minorCount), hint: rate(minorCount, completed, percentFormatter) },
    {
      label: t("stats.notifications"),
      value: String(notificationsCount),
      hint: rate(notificationsCount, completed, percentFormatter),
    },
    {
      label: t("stats.preferredName"),
      value: String(preferredNameCount),
      hint: rate(preferredNameCount, completed, percentFormatter),
    },
    {
      label: t("stats.skipHitap"),
      value: String(skipHitapCount),
      hint: rate(skipHitapCount, completed, percentFormatter),
    },
  ];

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold">{t("title")}</h1>
        <p className="text-muted-foreground">
          {t("subtitle")}
        </p>
      </div>

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {stats.map((stat) => (
          <Card key={stat.label}>
            <CardHeader>
              <CardDescription>{stat.label}</CardDescription>
              <CardTitle className="text-2xl">{stat.value}</CardTitle>
              <CardDescription>{stat.hint}</CardDescription>
            </CardHeader>
          </Card>
        ))}
      </div>

      <div className="grid gap-4 md:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>{t("ageRangeTitle")}</CardTitle>
            <CardDescription>{t("ageRangeDescription")}</CardDescription>
          </CardHeader>
          <CardContent className="space-y-3">
            {completed === 0 && (
              <p className="text-sm text-muted-foreground">{t("noDataYet")}</p>
            )}
            {sortByKnownOrder(ageRangeCounts, AGE_RANGE_ORDER).map(({ key, count }) => (
              <DistributionBar
                key={key}
                label={AGE_RANGE_LABELS[key] ?? key}
                count={count}
                total={completed}
              />
            ))}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>{t("moodPreferenceTitle")}</CardTitle>
            <CardDescription>{t("moodPreferenceDescription")}</CardDescription>
          </CardHeader>
          <CardContent className="space-y-3">
            {completed === 0 && (
              <p className="text-sm text-muted-foreground">{t("noDataYet")}</p>
            )}
            {sortByKnownOrder(moodPreferenceCounts, MOOD_PREFERENCE_ORDER).map(
              ({ key, count }) => (
                <DistributionBar
                  key={key}
                  label={MOOD_PREFERENCE_LABELS[key] ?? key}
                  count={count}
                  total={completed}
                />
              ),
            )}
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>{t("topPersonasTitle")}</CardTitle>
          <CardDescription>{t("topPersonasDescription")}</CardDescription>
        </CardHeader>
        <CardContent className="space-y-3">
          {topPersonas.length === 0 && (
            <p className="text-sm text-muted-foreground">{t("noDataYet")}</p>
          )}
          {topPersonas.map((persona) => (
            <DistributionBar
              key={persona.persona_id}
              label={persona.persona_name}
              count={persona.count}
              total={completed}
            />
          ))}
        </CardContent>
      </Card>
    </div>
  );
}
