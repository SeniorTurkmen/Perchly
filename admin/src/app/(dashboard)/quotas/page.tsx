import Link from "next/link";
import { getTranslations } from "next-intl/server";

import { Button } from "@/components/ui/button";
import {
  Card,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";

export default async function QuotasPage() {
  const t = await getTranslations("quotas");
  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-semibold">{t("title")}</h1>
      <Card>
        <CardHeader>
          <CardTitle>{t("cardTitle")}</CardTitle>
          <CardDescription>{t("cardDescription")}</CardDescription>
        </CardHeader>
      </Card>
      <Button render={<Link href="/users" />} variant="outline">
        {t("goToUsers")}
      </Button>
    </div>
  );
}
