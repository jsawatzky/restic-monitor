package monitor

import (
	"context"
	"strings"

	"github.com/jsawatzky/restic-monitor/restic"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

type maintenanceJob struct {
	repo restic.ResticRepo

	logger *zap.Logger
}

func NewMaintenanceJob(repo restic.ResticRepo) cron.Job {
	return &maintenanceJob{
		repo:   repo,
		logger: zap.L().Named("maintenance").With(zap.String("repo", repo.Name())),
	}
}

func (m *maintenanceJob) Run() {
	m.logger.Info("forgetting snapshots")

	forgot, err := m.repo.Forget(context.Background())
	if err != nil {
		m.logger.Error("failed to forget snapshots", zap.Error(err))
		return
	}

	for _, f := range forgot {
		snapshotsForgotten.WithLabelValues(m.repo.Name(), f.Hostname, strings.Join(f.Paths, ",")).Add(float64(len(f.Remove)))
	}
}
