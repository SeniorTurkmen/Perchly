-- Full request/response headers, for the admin log viewer's "how was
-- this request actually made" detail view — beyond the specific device
-- headers already broken out into their own columns. Redacted the same
-- way body/query_params already are (see internal/requestlog/redact.go,
-- now covering "authorization" and "cookie" too).
ALTER TABLE request_logs
    ADD COLUMN request_headers  JSONB,
    ADD COLUMN response_headers JSONB;
