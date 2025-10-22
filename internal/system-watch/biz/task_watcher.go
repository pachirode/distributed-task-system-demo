package biz

import (
	"context"
	"log/slog"
	"sync"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/pachirode/distributed-task-system-demo/internal/system-watch/model"
	"github.com/pachirode/distributed-task-system-demo/internal/system-watch/store"
	"github.com/pachirode/distributed-task-system-demo/pkg/meta"
)

type taskWatcher struct {
	store     store.IStore
	clientset kubernetes.Interface

	wg sync.WaitGroup
}

var _ Watcher = (*taskWatcher)(nil)

func (w *taskWatcher) Init(ctx context.Context, config *Config) error {
	w.store = config.Store
	w.clientset = config.Clientset
	return nil
}

func (w *taskWatcher) Spec() string {
	return "@every 30s"
}

func (w *taskWatcher) Run() {
	w.wg.Add(2)
	slog.Info("Task watcher start run")

	go func() {
		ctx := context.Background()

		_, tasks, err := w.store.Tasks().List(ctx, meta.WithFilter(map[string]any{
			"status": model.TaskStatusNormal,
		}))
		if err != nil {
			slog.Error("Error to list tasks", "err", err)
			return
		}

		var wg sync.WaitGroup
		wg.Add(len(tasks))

		for _, task := range tasks {
			slog.Debug("Current task infos", "namespace", task.Namespace, "name", task.Name)
			go func(task *model.TaskM) {
				defer wg.Done()
				job, err := w.clientset.BatchV1().Jobs(task.Namespace).Create(ctx, toJob(task), metav1.CreateOptions{})
				if err != nil {
					slog.Error("Error to create job", "namespace", task.Namespace, "name", task.Name, "err", err)
					return
				}

				task.Status = model.TaskStatusPending
				if err := w.store.Tasks().Update(ctx, task); err != nil {
					slog.Error("Error to update task status", "err", err)
					return
				}
				slog.Info("Successfully created job", "namespace", job.Namespace, "name", job.Name)
			}(task)
		}
		wg.Wait()
	}()

	// 同步状态
	go func() {
		defer w.wg.Done()
		ctx := context.Background()

		_, tasks, err := w.store.Tasks().List(ctx, meta.WithFilterNot(map[string]any{
			"status": []string{model.TaskStatusNormal, model.TaskStatusSucceeded, model.TaskStatusFailed},
		}))

		if err != nil {
			slog.Error("Error to list tasks", "err", err)
			return
		}

		var wg sync.WaitGroup
		wg.Add(len(tasks))
		for _, task := range tasks {
			go func(task *model.TaskM) {
				defer wg.Done()
				job, err := w.clientset.BatchV1().Jobs(task.Namespace).Get(ctx, task.Name, metav1.GetOptions{})
				if err != nil {
					slog.Error("Error to get task", "err", err)
					return
				}

				task.Status = toTaskStatus(job)
				if err := w.store.Tasks().Update(ctx, task); err != nil {
					slog.Error("Error to update task status")
					return
				}

				slog.Info("Successfully update job status", "namespace", job.Namespace, "name", job.Name)
			}(task)
		}
		wg.Wait()
	}()

	w.wg.Wait()
	slog.Info("Task watch is complete")
}

func init() {
	Register(&taskWatcher{})
}
