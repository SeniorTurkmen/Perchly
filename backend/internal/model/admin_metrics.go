package model

// AdminMetrics is a snapshot of today's activity for the admin
// dashboard home page. "Today" is UTC calendar-day, not per-user local
// time — good enough for an at-a-glance operational number, unlike the
// quota system's reset logic which has to be exact per user.
type AdminMetrics struct {
	TotalUsers               int `json:"total_users"`
	NewUsersToday            int `json:"new_users_today"`
	MessagesSentToday        int `json:"messages_sent_today"`
	ActiveConversationsToday int `json:"active_conversations_today"`
	// ErrorRateToday is the fraction (0-1) of today's requests that
	// returned a 5xx status. 0 when there were no requests today yet.
	ErrorRateToday float64 `json:"error_rate_today"`
}
