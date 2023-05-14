package monitor

import (
	"context"
	"math/rand"
	"strings"
	"time"

	"github.com/jsawatzky/restic-monitor/restic"
	"go.uber.org/zap"
)

type Poller interface {
	Run(ctx context.Context)
	Poll(ctx context.Context)
}

type poller struct {
	repo     restic.ResticRepo
	interval time.Duration

	logger *zap.Logger
}

func NewPoller(repo restic.ResticRepo, interval time.Duration) (Poller, error) {
	return &poller{
		repo:     repo,
		interval: interval,
		logger:   zap.L().Named("poller").With(zap.String("repo", repo.Name())),
	}, nil
}

func (p *poller) Poll(ctx context.Context) {
	p.logger.Info("polling repo")

	err := p.repo.Check(ctx)
	if err != nil {
		p.logger.Error("repo check failed", zap.Error(err))
		repoStatus.WithLabelValues(p.repo.Name()).Set(0)
	} else {
		repoStatus.WithLabelValues(p.repo.Name()).Set(1)
	}

	snapshots, err := p.repo.GetSnapshots(ctx)
	if err != nil {
		p.logger.Error("failed to get snapshots", zap.Error(err))
	}

	for _, group := range snapshots {
		path := strings.Join(group.Key.Paths, ",")
		var latestSnapshot restic.Snapshot
		for _, snapshot := range group.Snapshots {
			if latestSnapshot.Time.Before(snapshot.Time) {
				latestSnapshot = snapshot
			}
		}

		snapshotCount.WithLabelValues(p.repo.Name(), group.Key.Hostname, path).Set(float64(len(group.Snapshots)))
		lastSnapshot.WithLabelValues(p.repo.Name(), group.Key.Hostname, path).Set(float64(latestSnapshot.Time.Unix()))

		restoreStats, err := p.repo.GetRestoreStats(ctx, latestSnapshot.Id)
		if err != nil {
			p.logger.Error("failed to get restore stats", zap.Error(err), zap.String("snapshot", latestSnapshot.ShortID))
			continue
		}

		snapshotSize.WithLabelValues(p.repo.Name(), group.Key.Hostname, path).Set(float64(restoreStats.TotalSize))
		snapshotFileCount.WithLabelValues(p.repo.Name(), group.Key.Hostname, path).Set(float64(restoreStats.TotalFileCount))
	}

	rawStats, err := p.repo.GetRawStats(ctx)
	if err != nil {
		p.logger.Error("failed to get raw stats", zap.Error(err))
		return
	}

	repoSize.WithLabelValues(p.repo.Name()).Set(float64(rawStats.TotalSize))
	repoUncompressedSize.WithLabelValues(p.repo.Name()).Set(float64(rawStats.TotalUncompressedSize))
	repoCompressionRatio.WithLabelValues(p.repo.Name()).Set(float64(rawStats.CompressionRatio))
	repoBlobCount.WithLabelValues(p.repo.Name()).Set(float64(rawStats.TotalBlobCount))

	p.logger.Info("poll complete")
}

func (p *poller) Run(ctx context.Context) {

	randWait := time.Duration(rand.Intn(int(p.interval.Seconds()))) * time.Second
	timer := time.NewTimer(randWait)

	select {
	case <-ctx.Done():
		if !timer.Stop() {
			<-timer.C
		}
		return
	case <-timer.C:
	}

	p.Poll(ctx)

	ticker := time.NewTicker(p.interval)

	for {
		select {
		case <-ctx.Done():
			ticker.Stop()
			return
		case <-ticker.C:
			p.Poll(ctx)
		}
	}

}
