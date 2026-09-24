import {
  Card,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

export default function DashboardHomePage() {
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-semibold">Panel</h1>
        <p className="text-muted-foreground">
          Perchly operasyon ve mühendislik paneline hoş geldin.
        </p>
      </div>
      <Card>
        <CardHeader>
          <CardTitle>Faz 0 — İskelet</CardTitle>
          <CardDescription>
            Giriş ve gezinme hazır. Kullanıcılar, personalar, konuşmalar ve
            diğer bölümler sıradaki fazlarda dolduruluyor.
          </CardDescription>
        </CardHeader>
      </Card>
    </div>
  );
}
