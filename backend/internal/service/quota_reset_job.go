package service

import (
	"context"
	"log"
	"time"
)

type QuotaResetRepo interface {
	ResetDueForNewLocalDay(ctx context.Context) (int64, error)
}

// QuotaResetJob periodically zeroes out message_count_today for users
// whose local day has rolled over, per their own timezone (see
// QuotaRepository.ResetDueForNewLocalDay). It complements — doesn't
// replace — QuotaService.Check, which also lazily sees a fresh
// message_count_today at read time via the same underlying comparison;
// this job just keeps the numbers fresh in the database (e.g. for a
// dashboard) even for users who don't send a message right at midnight.
type QuotaResetJob struct {
	repo     QuotaResetRepo
	interval time.Duration
}

func NewQuotaResetJob(repo QuotaResetRepo, interval time.Duration) *QuotaResetJob {
	if interval <= 0 {
		interval = time.Minute
	}
	return &QuotaResetJob{repo: repo, interval: interval}
}

// Run blocks, ticking until ctx is canceled. Call it with `go job.Run(ctx)`.
func (j *QuotaResetJob) Run(ctx context.Context) {
	j.tick(ctx) // catch up immediately on startup rather than waiting a full interval

	ticker := time.NewTicker(j.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			j.tick(ctx)
		}
	}
}

func (j *QuotaResetJob) tick(ctx context.Context) {
	n, err := j.repo.ResetDueForNewLocalDay(ctx)
	if err != nil {
		log.Printf("quota reset job failed: %v", err)
		return
	}
	if n > 0 {
		log.Printf("quota reset job: reset %d quota row(s) for a new local day", n)
	}
}
