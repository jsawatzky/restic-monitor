package monitor

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	repoStatus = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "restic",
			Name:      "repo_status",
			Help:      "status of integrity checks on the repo",
		},
		[]string{"repo"},
	)
	snapshotCount = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "restic",
			Name:      "snapshot_count",
			Help:      "number of snapshots stored in the repo",
		},
		[]string{"repo", "host", "path"},
	)
	lastSnapshot = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "restic",
			Name:      "last_snapshot",
			Help:      "unix timestamp of last snapshot",
		},
		[]string{"repo", "host", "path"},
	)
	repoSize = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "restic",
			Name:      "repo_size_bytes",
			Help:      "raw size of the repo",
		},
		[]string{"repo"},
	)
	repoUncompressedSize = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "restic",
			Name:      "repo_uncompressed_size_bytes",
			Help:      "raw uncompressed size of the repo",
		},
		[]string{"repo"},
	)
	repoCompressionRatio = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "restic",
			Name:      "repo_compression_ratio",
			Help:      "compression ratio of the repo",
		},
		[]string{"repo"},
	)
	repoBlobCount = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "restic",
			Name:      "repo_blob_count",
			Help:      "number of blobs in the repo",
		},
		[]string{"repo"},
	)
	snapshotSize = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "restic",
			Name:      "snapshot_size",
			Help:      "restored size of latest snapshot",
		},
		[]string{"repo", "host", "path"},
	)
	snapshotFileCount = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "restic",
			Name:      "snapshot_file_count",
			Help:      "number of files in latest snapshot",
		},
		[]string{"repo", "host", "path"},
	)
	snapshotsForgotten = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "restic",
			Name:      "snapshots_forgotten",
			Help:      "number of snapshots forgotten during maintenance",
		},
		[]string{"repo", "host", "path"},
	)
)
