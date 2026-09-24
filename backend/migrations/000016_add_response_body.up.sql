-- The response body too, not just its headers — truncated at write time
-- to a reasonable size (see internal/requestlog/middleware.go's
-- maxCapturedResponseBody) so a long SSE chat stream isn't buffered
-- whole into a log row. TEXT, not JSONB: unlike the request body
-- (always JSON, this API's own convention), a response can be an SSE
-- stream or plain text, not just JSON.
ALTER TABLE request_logs
    ADD COLUMN response_body TEXT;
