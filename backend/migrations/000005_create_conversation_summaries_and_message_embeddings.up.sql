CREATE TABLE conversation_summaries (
    conversation_id        UUID PRIMARY KEY REFERENCES conversations(id) ON DELETE CASCADE,
    summary_text           TEXT NOT NULL,
    -- How many of the conversation's messages (in creation order) are
    -- already folded into summary_text, so the rolling summarizer knows
    -- exactly which messages are "new" since the last summary and never
    -- re-summarizes (or skips) the wrong range.
    covered_message_count  INTEGER NOT NULL DEFAULT 0,
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE message_embeddings (
    message_id       UUID PRIMARY KEY REFERENCES messages(id) ON DELETE CASCADE,
    -- Denormalized from messages.conversation_id so nearest-neighbor
    -- search can filter to one conversation directly on this table,
    -- without a join driving the index scan.
    conversation_id  UUID NOT NULL REFERENCES conversations(id) ON DELETE CASCADE,
    embedding        vector(1536) NOT NULL,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX message_embeddings_conversation_id_idx ON message_embeddings (conversation_id);
CREATE INDEX message_embeddings_embedding_hnsw_idx ON message_embeddings USING hnsw (embedding vector_cosine_ops);
