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

const AGE_RANGE_LABELS: Record<string, string> = {
  under18: "18 yaş altı",
  age18to24: "18-24",
  age25to34: "25-34",
  age35plus: "35+",
};

const AGE_RANGE_ORDER = ["under18", "age18to24", "age25to34", "age35plus"];

const MOOD_PREFERENCE_LABELS: Record<string, string> = {
  motivation: "Motivasyon",
  dailyChat: "Günlük sohbet",
  hobbyTalk: "Hobi / kitap",
  skipped: "Atlandı",
  none: "Cevaplanmadı",
};

const MOOD_PREFERENCE_ORDER = ["motivation", "dailyChat", "hobbyTalk", "skipped", "none"];

const percentFormatter = new Intl.NumberFormat("tr-TR", {
  style: "percent",
  maximumFractionDigits: 1,
});

function rate(count: number, total: number): string {
  return total > 0 ? percentFormatter.format(count / total) : "—";
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
      label: "Onboarding tamamlayan",
      value: `${completed} / ${totalUsers}`,
      hint: rate(completed, totalUsers),
    },
    { label: "Reşit olmayan", value: String(minorCount), hint: rate(minorCount, completed) },
    {
      label: "Bildirim izni veren",
      value: String(notificationsCount),
      hint: rate(notificationsCount, completed),
    },
    {
      label: "Hitap belirleyen",
      value: String(preferredNameCount),
      hint: rate(preferredNameCount, completed),
    },
    {
      label: "Hitabı atlayan",
      value: String(skipHitapCount),
      hint: rate(skipHitapCount, completed),
    },
  ];

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold">Onboarding</h1>
        <p className="text-muted-foreground">
          Kullanıcıların onboarding akışında verdiği cevapların özeti.
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
            <CardTitle>Yaş aralığı</CardTitle>
            <CardDescription>Onboarding&apos;i tamamlayanlar arasında.</CardDescription>
          </CardHeader>
          <CardContent className="space-y-3">
            {completed === 0 && (
              <p className="text-sm text-muted-foreground">Henüz veri yok.</p>
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
            <CardTitle>Ruh hali tercihi</CardTitle>
            <CardDescription>Onboarding&apos;de seçilen ilgi alanı.</CardDescription>
          </CardHeader>
          <CardContent className="space-y-3">
            {completed === 0 && (
              <p className="text-sm text-muted-foreground">Henüz veri yok.</p>
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
          <CardTitle>En çok seçilen personalar</CardTitle>
          <CardDescription>Onboarding&apos;de ilk seçim olarak.</CardDescription>
        </CardHeader>
        <CardContent className="space-y-3">
          {topPersonas.length === 0 && (
            <p className="text-sm text-muted-foreground">Henüz veri yok.</p>
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
