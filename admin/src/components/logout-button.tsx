import { getTranslations } from "next-intl/server";

import { Button } from "@/components/ui/button";

// Posts to the /logout Route Handler (see src/app/logout/route.ts) —
// a plain URL-based form, not a Server Function, on purpose.
export async function LogoutButton() {
  const t = await getTranslations("common");

  return (
    <form method="post" action="/logout">
      <Button type="submit" variant="ghost" size="sm">
        {t("logout")}
      </Button>
    </form>
  );
}
