"use client";

import type { MouseEvent } from "react";

import { Button } from "@/components/ui/button";

export function DeleteMessageButton() {
  return (
    <Button
      type="submit"
      variant="ghost"
      size="sm"
      className="text-destructive hover:text-destructive"
      onClick={(e: MouseEvent<HTMLButtonElement>) => {
        if (!confirm("Bu mesajı kalıcı olarak silmek istediğine emin misin?")) {
          e.preventDefault();
        }
      }}
    >
      Sil
    </Button>
  );
}
