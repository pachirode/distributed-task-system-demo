package system_watch

import (
	"log/slog"
	"time"

	"k8s.io/client-go/kubernetes"

	"github.com/pachirode/distributed-task-system-demo/internal/system-watch/biz"
	"github.com/pachirode/distributed-task-system-demo/internal/system-watch/store"
	"github.com/pachirode/distributed-task-system-demo/pkg/db"
)

const (
	Every3Seconds = "@every 3s"
)

var (
	lockName          = "system-watch-lock"
	jobStopTimeout    = 3 * time.Minute
	extendExpiration  = 5 * time.Second
	defaultExpiration = 2 * extendExpiration
)

type SystemWatchConfig struct {
	MySQLOptions *db.MySQLOptions
	RedisOptions *db.RedisOptions
	Clientset    kubernetes.Interface
}

func (c *SystemWatchConfig) NewSystemWatchConfig() (*biz.Config, error) {
	gormDB, err := db.NewMySQL(c.MySQLOptions)
	if err != nil {
		slog.Error("Error to create mysql client", "err", err)
		return nil, err
	}

	datastore := store.NewStore(gormDB)

	return &biz.Config{Store: datastore, Clientset: c.Clientset}, nil
}
