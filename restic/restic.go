package restic

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"syscall"

	"go.uber.org/zap"
)

var dryRun = false

func init() {
	if v, ok := os.LookupEnv("DRY_RUN"); ok {
		if len(v) > 0 {
			dryRun = true
		}
	}
}

type ResticRepo interface {
	Name() string

	Check(context.Context) error
	Forget(context.Context) ([]ForgetGroup, error)
	GetSnapshots(context.Context) ([]GroupedSnapshots, error)
	GetRawStats(context.Context) (RawDataStats, error)
	GetRestoreStats(context.Context, string) (RestoreSizeStats, error)
}

type resticRepo struct {
	name            string
	repository      string
	environment     map[string]string
	retentionPolicy RetentionConfig

	mu *sync.Mutex

	logger *zap.Logger
}

func New(name string, config RepoConfig) (ResticRepo, error) {
	if config.Environment == nil {
		config.Environment = make(map[string]string)
	}

	if len(config.EnvironmentFile) > 0 {
		var env map[string]string
		envFile, err := os.Open(config.EnvironmentFile)
		if err != nil {
			return nil, err
		}
		defer envFile.Close()
		data, err := io.ReadAll(envFile)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(data, &env)
		if err != nil {
			return nil, err
		}

		for k, v := range env {
			config.Environment[k] = v
		}
	}

	for _, cmd := range []string{"check", "forget", "snapshots", "stats"} {
		commandErrors.WithLabelValues(cmd, name)
	}

	return &resticRepo{
		name:            name,
		repository:      config.Repository,
		environment:     config.Environment,
		retentionPolicy: config.Retention,
		mu:              &sync.Mutex{},
		logger:          zap.L().Named("restic").With(zap.String("repo", name)),
	}, nil
}

func (r *resticRepo) Name() string {
	return r.name
}

func (r *resticRepo) cmd(ctx context.Context, c string, args ...string) ([]byte, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	args = append([]string{c}, args...)
	args = append(args, "--json")
	cmd := exec.CommandContext(ctx, "restic", args...)
	cmd.Cancel = func() error {
		return cmd.Process.Signal(syscall.SIGTERM)
	}
	cmd.Env = append(cmd.Environ(), fmt.Sprintf("RESTIC_REPOSITORY=%s", r.repository))
	for k, v := range r.environment {
		cmd.Env = append(cmd.Environ(), fmt.Sprintf("%s=%s", k, v))
	}
	r.logger.Debug("running restic command", zap.String("cmd", cmd.String()))
	output, err := cmd.Output()
	if err != nil {
		commandErrors.WithLabelValues(c, r.name).Inc()
		if exitError, ok := err.(*exec.ExitError); ok {
			r.logger.Error("restic command exited with an error", zap.String("cmd", cmd.String()), zap.ByteString("stderr", exitError.Stderr), zap.Int("exitCode", exitError.ExitCode()))
		}
	}
	return output, err
}

func (r *resticRepo) Check(ctx context.Context) error {
	_, err := r.cmd(ctx, "check")
	return err
}

func (r *resticRepo) Forget(ctx context.Context) ([]ForgetGroup, error) {
	args := r.retentionPolicy.ToArgs()
	if dryRun {
		args = append([]string{"-n"}, args...)
	}
	out, err := r.cmd(ctx, "forget", args...)
	if err != nil {
		return nil, err
	}

	var res []ForgetGroup
	err = json.Unmarshal(out, &res)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *resticRepo) GetSnapshots(ctx context.Context) ([]GroupedSnapshots, error) {
	out, err := r.cmd(ctx, "snapshots", "--group-by", "host,path")
	if err != nil {
		return nil, err
	}

	var res []GroupedSnapshots
	err = json.Unmarshal(out, &res)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func (r *resticRepo) GetRawStats(ctx context.Context) (RawDataStats, error) {
	out, err := r.cmd(ctx, "stats", "--mode", "raw-data")
	if err != nil {
		return RawDataStats{}, err
	}

	var res RawDataStats
	err = json.Unmarshal(out, &res)
	if err != nil {
		return RawDataStats{}, err
	}

	return res, nil
}

func (r *resticRepo) GetRestoreStats(ctx context.Context, snapshot string) (RestoreSizeStats, error) {
	out, err := r.cmd(ctx, "stats", "--mode", "restore-size", snapshot)
	if err != nil {
		return RestoreSizeStats{}, err
	}

	var res RestoreSizeStats
	err = json.Unmarshal(out, &res)
	if err != nil {
		return RestoreSizeStats{}, err
	}

	return res, nil
}
