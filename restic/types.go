package restic

import (
	"strconv"
	"time"
)

type GroupKey struct {
	Hostname string   `json:"hostname"`
	Paths    []string `json:"paths"`
}

type Snapshot struct {
	Id       string    `json:"id"`
	ShortID  string    `json:"short_id"`
	Time     time.Time `json:"time"`
	Paths    []string  `json:"paths"`
	Hostname string    `json:"hostname"`
	Username string    `json:"username"`
	Parent   string    `json:"parent"`
	Tree     string    `json:"tree"`
}

type GroupedSnapshots struct {
	Key       GroupKey   `json:"group_key"`
	Snapshots []Snapshot `json:"snapshots"`
}

type RestoreSizeStats struct {
	TotalSize      int64 `json:"total_size"`
	TotalFileCount int64 `json:"total_file_count"`
	SnapshotCount  int64 `json:"snapshots_count"`
}

type RawDataStats struct {
	TotalSize              int64   `json:"total_size"`
	TotalUncompressedSize  int64   `json:"total_uncompressed_size"`
	CompressionRatio       float64 `json:"compression_ratio"`
	CompressionProgress    float64 `json:"compression_progress"`
	CompressionSpaceSaving float64 `json:"compression_space_saving"`
	TotalBlobCount         int64   `json:"total_blob_count"`
	SnapshotCount          int64   `json:"snapshots_count"`
}

type RetentionReason struct {
	Snapshot Snapshot `json:"snapshot"`
	Matches  []string `json:"matches"`
}

type ForgetGroup struct {
	Hostname string            `json:"host"`
	Paths    []string          `json:"paths"`
	Keep     []Snapshot        `json:"keep"`
	Remove   []Snapshot        `json:"remove"`
	Reasons  []RetentionReason `json:"reasons"`
}

type RetentionConfig struct {
	LastN   int `toml:"last_n"`
	Hourly  int
	Daily   int
	Weekly  int
	Monthly int
	Yearly  int
	Tags    []string
}

func (rc RetentionConfig) ToArgs() []string {
	var ret []string
	if rc.LastN > 0 {
		ret = append(ret, "--keep-last", strconv.Itoa(rc.LastN))
	}
	if rc.Hourly > 0 {
		ret = append(ret, "--keep-hourly", strconv.Itoa(rc.Hourly))
	}
	if rc.Daily > 0 {
		ret = append(ret, "--keep-daily", strconv.Itoa(rc.Daily))
	}
	if rc.Weekly > 0 {
		ret = append(ret, "--keep-weekly", strconv.Itoa(rc.Weekly))
	}
	if rc.Monthly > 0 {
		ret = append(ret, "--keep-monthly", strconv.Itoa(rc.Monthly))
	}
	if rc.Yearly > 0 {
		ret = append(ret, "--keep-yearly", strconv.Itoa(rc.Yearly))
	}
	for _, t := range rc.Tags {
		ret = append(ret, "--keep-tag", t)
	}
	return ret
}

type RepoConfig struct {
	Repository          string
	EnvironmentFile     string `yaml:"environment_file"`
	Environment         map[string]string
	Retention           RetentionConfig
	PollingInterval     time.Duration `yaml:"polling_interval"`
	MaintenanceSchedule string        `yaml:"maintenance_schedule"`
}
