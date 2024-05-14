package restic

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	commandRepoLocked = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "restic",
			Subsystem: "command",
			Name:      "repo_locked",
			Help:      "1 if the repo is locked by another process, 0 otherwise",
		},
	)
	commandUnknownErrors = promauto.NewCounter(
		prometheus.CounterOpts{
			Namespace: "restic",
			Subsystem: "command",
			Name:      "unknown_errors_count",
			Help:      "number of unknown errors encountered when running restic commands",
		},
	)
	commandConnectionErrors = promauto.NewCounter(
		prometheus.CounterOpts{
			Namespace: "restic",
			Subsystem: "command",
			Name:      "connection_errors_count",
			Help:      "number of connection errors encountered when running restic commands",
		},
	)
	commandCount = promauto.NewCounter(
		prometheus.CounterOpts{
			Namespace: "restic",
			Subsystem: "command",
			Name:      "count",
			Help:      "number of restic commands run",
		},
	)
)
