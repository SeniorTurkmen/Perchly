import { Button } from "@/components/ui/button";

// Posts to the /logout Route Handler (see src/app/logout/route.ts) —
// a plain URL-based form, not a Server Function, on purpose.
export function LogoutButton() {
  return (
    <form method="post" action="/logout">
      <Button type="submit" variant="ghost" size="sm">
        Çıkış yap
      </Button>
    </form>
  );
}
