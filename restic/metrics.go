package restic

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	commandErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "restic",
			Subsystem: "command",
			Name:      "errors",
			Help:      "number of errors when running restic commands",
		},
		[]string{"cmd", "repo"},
	)
)
