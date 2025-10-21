package biz

import (
	"context"
	"sync"

	"github.com/robfig/cron/v3"

	reflectutil "github.com/pachirode/distributed-task-system-demo/pkg/util/reflect"
)

var (
	registryLock = new(sync.Mutex)
	registry     = make(map[string]Watcher)
)

type Watcher interface {
	Init(ctx context.Context, config *Config) error
	cron.Job
}

type ISpec interface {
	Spec() string
}

func Register(watcher Watcher) {
	registryLock.Lock()
	defer registryLock.Unlock()

	name := reflectutil.StructName(watcher)
	if _, ok := registry[name]; ok {
		panic("duplicate watcher entry: " + name)
	}

	registry[name] = watcher
}

func ListWatchers() map[string]Watcher {
	registryLock.Lock()
	defer registryLock.Unlock()

	return registry
}
