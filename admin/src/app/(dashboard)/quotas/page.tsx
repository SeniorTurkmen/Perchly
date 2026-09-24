import Link from "next/link";

import { Button } from "@/components/ui/button";
import {
  Card,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

export default function QuotasPage() {
  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-semibold">Kota & Kredi</h1>
      <Card>
        <CardHeader>
          <CardTitle>Kullanıcı başına düzenleniyor</CardTitle>
          <CardDescription>
            Bir kullanıcının persona başına günlük limitini veya kredi
            bakiyesini değiştirmek için Kullanıcılar → ilgili kullanıcının
            detay sayfasına git. Buradaki tüm-kullanıcılar tablo görünümü
            (toplu düzenleme) sonraki bir fazda eklenecek.
          </CardDescription>
        </CardHeader>
      </Card>
      <Button render={<Link href="/users" />} variant="outline">
        Kullanıcılara git
      </Button>
    </div>
  );
}
