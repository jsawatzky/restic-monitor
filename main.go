package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/go-logr/zapr"
	"github.com/jsawatzky/restic-monitor/monitor"
	"github.com/jsawatzky/restic-monitor/restic"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

type ResticConfig struct {
	Repos map[string]restic.RepoConfig
}

func configureLogger() {
	logConf := zap.NewProductionConfig()
	logConf.OutputPaths = []string{"stdout"}
	logConf.ErrorOutputPaths = []string{"stdout"}
	logger, err := logConf.Build()
	if err != nil {
		log.Fatalf("error configuring zap logger: %v", err)
	}
	zap.ReplaceGlobals(logger.Named("main"))
}

func loadConfig(path string) (ResticConfig, error) {
	configFile, err := os.Open(path)
	if err != nil {
		return ResticConfig{}, err
	}
	defer configFile.Close()

	data, err := io.ReadAll(configFile)
	if err != nil {
		return ResticConfig{}, err
	}

	var ret ResticConfig
	yaml.Unmarshal(data, &ret)

	return ret, nil
}

func main() {
	defer zap.L().Sync()
	configureLogger()
	logger := zap.L()
	zaprLogger := zapr.NewLogger(logger.Named("cron"))

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(":9090", nil)
	}()

	configFilePath := "/etc/restic-monitor/config.yaml"
	if len(os.Args) > 1 {
		configFilePath = os.Args[1]
	}

	if v, ok := os.LookupEnv("DRY_RUN"); ok {
		if len(v) > 0 {
			logger.Info("running in dry run mode")
		}
	}

	config, err := loadConfig(configFilePath)
	if err != nil {
		zap.L().Error("error loading config", zap.Error(err), zap.String("filePath", configFilePath))
	}

	cronRunner := cron.New(cron.WithChain(cron.SkipIfStillRunning(zaprLogger), cron.Recover(zaprLogger)))

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)

	var wg sync.WaitGroup

	for name, repoConfig := range config.Repos {
		repo, err := restic.New(name, repoConfig)
		if err != nil {
			logger.Error("error creating restic repo", zap.Error(err), zap.String("repo", name))
			logger.Warn("skipping repo", zap.String("repo", name))
			continue
		}
		poller, err := monitor.NewPoller(repo, repoConfig.PollingInterval)
		if err != nil {
			logger.Error("error creating poller", zap.Error(err), zap.String("repo", name))
			logger.Warn("skipping repo", zap.String("repo", name))
			continue
		}
		poller.Poll(ctx)
		cronRunner.AddJob(repoConfig.MaintenanceSchedule, monitor.NewMaintenanceJob(repo))

		wg.Add(1)
		go func(p monitor.Poller, wg *sync.WaitGroup) {
			defer wg.Done()
			p.Run(ctx)
		}(poller, &wg)
		logger.Info("started monitoring repo", zap.String("repo", name), zap.Duration("polling_interval", repoConfig.PollingInterval), zap.String("maintenance_schedule", repoConfig.MaintenanceSchedule))
	}

	cronRunner.Start()

	wg.Wait()
	cancel()

	cronCtx := cronRunner.Stop()
	logger.Info("waiting for cron jobs to complete")
	<-cronCtx.Done()

	logger.Info("shutting down")
}
