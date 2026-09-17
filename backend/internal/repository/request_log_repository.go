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

	_, err = r.pool.Exec(ctx, `
		INSERT INTO request_logs (
			user_id, method, route_pattern, path, query_params, route_params, body,
			status_code, duration_ms, client_version, platform, os_version, device_model,
			user_agent, ip_address
		) VALUES (
			$1::uuid, $2, $3, $4, $5::jsonb, $6::jsonb, $7::jsonb,
			$8, $9, $10, $11, $12, $13,
			$14, $15
		)
	`,
		entry.UserID, entry.Method, entry.RoutePattern, entry.Path, queryParams, routeParams, body,
		entry.StatusCode, entry.DurationMS, entry.ClientVersion, entry.Platform, entry.OSVersion, entry.DeviceModel,
		entry.UserAgent, entry.IPAddress,
	)
	return err
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
