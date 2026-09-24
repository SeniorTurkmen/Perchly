import Link from "next/link";
import { notFound } from "next/navigation";

import { LocalDateTime } from "@/components/local-date-time";
import { Badge } from "@/components/ui/badge";
import { DeleteMessageButton } from "@/components/delete-message-button";
import { requireSessionToken } from "@/lib/auth";
import {
  AdminApiError,
  adminGetConversation,
  adminGetPersona,
} from "@/lib/backend";
import { cn } from "@/lib/utils";

import { deleteMessageAction } from "../actions";

export default async function ConversationDetailPage({
  params,
  searchParams,
}: {
  params: Promise<{ id: string }>;
  searchParams: Promise<{ error?: string }>;
}) {
  const { id } = await params;
  const { error } = await searchParams;
  const token = await requireSessionToken();

  let detail;
  try {
    detail = await adminGetConversation(token, id);
  } catch (err) {
    if (err instanceof AdminApiError && err.status === 404) notFound();
    throw err;
  }

  const persona = await adminGetPersona(token, detail.conversation.persona_id);

  return (
    <div className="space-y-6">
      {error && (
        <div className="rounded-md border border-destructive/50 bg-destructive/10 px-4 py-3 text-sm text-destructive">
          {error}
        </div>
      )}

      <div>
        <h1 className="text-2xl font-semibold">{persona.name} ile konuşma</h1>
        <Link
          href={`/users/${detail.conversation.user_id}`}
          className="text-sm text-muted-foreground hover:underline"
        >
          Kullanıcıyı görüntüle
        </Link>
      </div>

      <div className="space-y-3">
        {detail.messages.length === 0 && (
          <p className="text-muted-foreground">Bu konuşmada mesaj kalmadı.</p>
        )}
        {detail.messages.map((message) => (
          <div
            key={message.id}
            className={cn(
              "rounded-lg border p-4",
              message.role === "user" ? "bg-muted/40" : "bg-background",
            )}
          >
            <div className="mb-2 flex items-center justify-between">
              <div className="flex items-center gap-2 text-sm">
                <Badge variant={message.role === "user" ? "secondary" : "default"}>
                  {message.role === "user" ? "Kullanıcı" : "Persona"}
                </Badge>
                {message.reaction_emoji && <span>{message.reaction_emoji}</span>}
                <span className="text-muted-foreground">
                  <LocalDateTime value={message.created_at} />
                </span>
              </div>
              <form action={deleteMessageAction.bind(null, detail.conversation.id, message.id)}>
                <DeleteMessageButton />
              </form>
            </div>
            <p className="text-sm whitespace-pre-wrap">{message.content}</p>
          </div>
        ))}
      </div>
    </div>
  );
}
