package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"perchly-backend/internal/model"
)

type RequestLogRepository struct {
	pool *pgxpool.Pool
}

func NewRequestLogRepository(pool *pgxpool.Pool) *RequestLogRepository {
	return &RequestLogRepository{pool: pool}
}

// Create persists one request log row. Query/route params and body are
// JSON-encoded here (rather than relying on pgx to marshal arbitrary Go
// values into jsonb) so nil maps/values cleanly become SQL NULL instead
// of the literal string "null".
func (r *RequestLogRepository) Create(ctx context.Context, entry model.RequestLog) error {
	queryParams, err := marshalNullableJSON(entry.QueryParams)
	if err != nil {
		return fmt.Errorf("marshal query_params: %w", err)
	}
	routeParams, err := marshalNullableJSON(entry.RouteParams)
	if err != nil {
		return fmt.Errorf("marshal route_params: %w", err)
	}
	body, err := marshalNullableJSON(entry.Body)
	if err != nil {
		return fmt.Errorf("marshal body: %w", err)
	}
	requestHeaders, err := marshalNullableJSON(entry.RequestHeaders)
	if err != nil {
		return fmt.Errorf("marshal request_headers: %w", err)
	}
	responseHeaders, err := marshalNullableJSON(entry.ResponseHeaders)
	if err != nil {
		return fmt.Errorf("marshal response_headers: %w", err)
	}

	_, err = r.pool.Exec(ctx, `
		INSERT INTO request_logs (
			user_id, method, route_pattern, path, query_params, route_params, body,
			status_code, duration_ms, request_headers, response_headers, response_body,
			client_version, platform, os_version, device_model,
			user_agent, ip_address
		) VALUES (
			$1::uuid, $2, $3, $4, $5::jsonb, $6::jsonb, $7::jsonb,
			$8, $9, $10::jsonb, $11::jsonb, $12,
			$13, $14, $15, $16,
			$17, $18
		)
	`,
		entry.UserID, entry.Method, entry.RoutePattern, entry.Path, queryParams, routeParams, body,
		entry.StatusCode, entry.DurationMS, requestHeaders, responseHeaders, entry.ResponseBody,
		entry.ClientVersion, entry.Platform, entry.OSVersion, entry.DeviceModel,
		entry.UserAgent, entry.IPAddress,
	)
	return err
}

// ListForAdmin returns a page of request logs, newest first, each
// joined with its user's email/display_name (nil for an anonymous user
// or an unauthenticated request) so the admin dashboard can show "who"
// without a lookup per row. Optionally filtered by a case-insensitive
// substring match against the route pattern or path, a minimum status
// code (e.g. 400 to surface only failed requests), and/or one specific
// user. Zero/nil values mean "don't filter on that". Also returns the
// total row count matching the filter, for pagination.
func (r *RequestLogRepository) ListForAdmin(ctx context.Context, search string, statusMin int, userID *string, limit, offset int) ([]model.RequestLogEntry, int, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	const where = `
		($1 = '' OR l.route_pattern ILIKE '%' || $1 || '%' OR l.path ILIKE '%' || $1 || '%')
		AND ($2 = 0 OR l.status_code >= $2)
		AND ($3::uuid IS NULL OR l.user_id = $3::uuid)
	`

	var total int
	if err := r.pool.QueryRow(ctx, `
		SELECT count(*) FROM request_logs l WHERE `+where,
		search, statusMin, userID,
	).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.pool.Query(ctx, `
		SELECT
			l.id, l.user_id::text, l.method, l.route_pattern, l.path, l.query_params, l.route_params, l.body,
			l.status_code, l.duration_ms, l.request_headers, l.response_headers, l.response_body,
			l.client_version, l.platform, l.os_version, l.device_model,
			l.user_agent, l.ip_address, l.created_at,
			u.email, u.display_name
		FROM request_logs l
		LEFT JOIN users u ON u.id = l.user_id
		WHERE `+where+`
		ORDER BY l.created_at DESC
		LIMIT $4 OFFSET $5
	`, search, statusMin, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	logs := make([]model.RequestLogEntry, 0)
	for rows.Next() {
		l, err := scanRequestLogEntry(rows)
		if err != nil {
			return nil, 0, err
		}
		logs = append(logs, l)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return logs, total, nil
}

func scanRequestLogEntry(row rowScanner) (model.RequestLogEntry, error) {
	var l model.RequestLogEntry
	err := row.Scan(
		&l.ID, &l.UserID, &l.Method, &l.RoutePattern, &l.Path, &l.QueryParams, &l.RouteParams, &l.Body,
		&l.StatusCode, &l.DurationMS, &l.RequestHeaders, &l.ResponseHeaders, &l.ResponseBody,
		&l.ClientVersion, &l.Platform, &l.OSVersion, &l.DeviceModel,
		&l.UserAgent, &l.IPAddress, &l.CreatedAt,
		&l.UserEmail, &l.UserDisplayName,
	)
	return l, err
}

// marshalNullableJSON returns nil (SQL NULL) for a nil/empty value
// instead of the JSON literal "null", so an absent body or empty query
// string reads as NULL in the database rather than a JSON null.
func marshalNullableJSON(v any) ([]byte, error) {
	if isEmptyJSONValue(v) {
		return nil, nil
	}
	return json.Marshal(v)
}

func isEmptyJSONValue(v any) bool {
	switch value := v.(type) {
	case nil:
		return true
	case map[string]any:
		return len(value) == 0
	default:
		return false
	}
}
