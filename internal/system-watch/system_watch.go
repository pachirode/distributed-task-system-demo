package system_watch

import (
	"context"
	"log"
	"time"

	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/robfig/cron/v3"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/pachirode/distributed-task-system-demo/internal/system-watch/biz"
	"github.com/pachirode/distributed-task-system-demo/pkg/db"
)

type systemWatch struct {
	runner *cron.Cron
	locker *redsync.Mutex
	config *biz.Config
}

func (c *SystemWatchConfig) New() (*systemWatch, error) {
	rdb, err := db.NewRedis(c.RedisOptions)
	if err != nil {
		log.Fatalf("Error to create redis client: %v", err)
	}

	runner := cron.New(cron.WithSeconds(), cron.WithLogger(cron.DefaultLogger), cron.WithChain(
		cron.SkipIfStillRunning(cron.DefaultLogger),
		cron.Recover(cron.DefaultLogger),
	))

	// 分布式锁
	pool := goredis.NewPool(rdb)
	lockOpts := []redsync.Option{
		redsync.WithRetryDelay(50 * time.Microsecond),
		redsync.WithTries(3),
		redsync.WithExpiry(defaultExpiration),
	}
	locker := redsync.New(pool).NewMutex(lockName, lockOpts...)

	cfg, err := c.NewSystemWatchConfig()
	if err != nil {
		return nil, err
	}

	sw := &systemWatch{runner: runner, locker: locker, config: cfg}
	if err = sw.addWatchers(); err != nil {
		return nil, err
	}

	return sw, nil
}

func (sw *systemWatch) addWatchers() error {
	for s, w := range biz.ListWatchers() {
		if err := w.Init(context.Background(), sw.config); err != nil {
			log.Fatalf("Error to construct watcher %V: %v", s, err)
			return err
		}

		spec := Every3Seconds
		if obj, ok := w.(biz.ISpec); ok {
			spec = obj.Spec()
		}

		if _, err := sw.runner.AddJob(spec, w); err != nil {
			log.Fatalf("Error to add job to the cron %v: %v", s, err)
			return err
		}
	}

	return nil
}

func (sw *systemWatch) Run(stopCh <-chan struct{}) {
	ctx := wait.ContextForChannel(stopCh)

	// 循环加锁直至加锁成功
	ticker := time.NewTicker(defaultExpiration + (5 * time.Second))
	defer ticker.Stop()

	for {
		err := sw.locker.LockContext(ctx)
		if err == nil {
			log.Printf("Successfully acquire lock: %v", lockName)
			break
		}
		<-ticker.C
	}

	// 看门狗，自动加锁
	ticker = time.NewTicker(extendExpiration)
	defer ticker.Stop()

	go func() {
		for {
			<-ticker.C
			if ok, err := sw.locker.ExtendContext(ctx); !ok || err != nil {
				log.Fatalf("Error to extend lock: %v", err)
			}
		}
	}()

	sw.runner.Start()
	log.Printf("Started system-watch")

	<-stopCh

	sw.runner.Stop()
}

func (sw *systemWatch) Stop() {
	ctx := sw.runner.Stop()
	select {
	case <-ctx.Done():
	case <-time.After(jobStopTimeout):
		log.Fatalf("Context was not done, timeout: %v", jobStopTimeout.String())
	}

	if ok, err := sw.locker.Unlock(); !ok || err != nil {
		log.Fatalf("Error to unlock")
	}
}
