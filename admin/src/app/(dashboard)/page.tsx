import {
  Card,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { requireSessionToken } from "@/lib/auth";
import { adminMetrics } from "@/lib/backend";

const numberFormatter = new Intl.NumberFormat("tr-TR");
const percentFormatter = new Intl.NumberFormat("tr-TR", {
  style: "percent",
  maximumFractionDigits: 1,
});

export default async function DashboardHomePage() {
  const token = await requireSessionToken();
  const metrics = await adminMetrics(token);

  const stats = [
    { label: "Toplam kullanıcı", value: numberFormatter.format(metrics.total_users) },
    { label: "Bugün yeni kullanıcı", value: numberFormatter.format(metrics.new_users_today) },
    { label: "Bugün gönderilen mesaj", value: numberFormatter.format(metrics.messages_sent_today) },
    { label: "Bugün aktif konuşma", value: numberFormatter.format(metrics.active_conversations_today) },
    { label: "Bugünkü hata oranı", value: percentFormatter.format(metrics.error_rate_today) },
  ];

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold">Panel</h1>
        <p className="text-muted-foreground">
          Perchly operasyon ve mühendislik paneline hoş geldin.
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
