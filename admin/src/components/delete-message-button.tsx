"use client";

import { useTranslations } from "next-intl";
import type { MouseEvent } from "react";

import { Button } from "@/components/ui/button";

export function DeleteMessageButton() {
  const t = useTranslations("conversations");
  const tc = useTranslations("common");
  return (
    <Button
      type="submit"
      variant="ghost"
      size="sm"
      className="text-destructive hover:text-destructive"
      onClick={(e: MouseEvent<HTMLButtonElement>) => {
        if (!confirm(t("confirmDeleteMessage"))) {
          e.preventDefault();
        }
      }}
    >
      {tc("delete")}
    </Button>
  );
}
