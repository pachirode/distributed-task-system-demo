package biz

import (
	"context"
	"log"
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
	log.Print("Task watcher start run")

	go func() {
		ctx := context.Background()

		_, tasks, err := w.store.Tasks().List(ctx, meta.WithFilter(map[string]any{
			"status": model.TaskStatusNormal,
		}))
		if err != nil {
			log.Fatalf("Error to list tasks: %v", err)
			return
		}

		var wg sync.WaitGroup
		wg.Add(len(tasks))

		for _, task := range tasks {
			go func(task *model.TaskM) {
				defer wg.Done()
				job, err := w.clientset.BatchV1().Jobs(task.Namespace).Create(ctx, toJob(task), metav1.CreateOptions{})
				if err != nil {
					log.Fatalf("Error to create job, namespace %v, name %v : %v", task.Namespace, task.Name, job)
					return
				}

				task.Status = model.TaskStatusPending
				if err := w.store.Tasks().Update(ctx, task); err != nil {
					log.Fatalf("Error to update task status: %v", err)
					return
				}
				log.Printf("Successfully created job, namespace: %v, name: %v", job.Namespace, job.Name)
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
			log.Fatalf("Error to list tasks: %v", err)
			return
		}

		var wg sync.WaitGroup
		wg.Add(len(tasks))
		for _, task := range tasks {
			go func(task *model.TaskM) {
				defer wg.Done()
				job, err := w.clientset.BatchV1().Jobs(task.Namespace).Get(ctx, task.Name, metav1.GetOptions{})
				if err != nil {
					log.Fatalf("Error to get task: %v", err)
					return
				}

				task.Status = toTaskStatus(job)
				if err := w.store.Tasks().Update(ctx, task); err != nil {
					log.Fatalf("Error to update task status")
					return
				}

				log.Printf("Successfully update job status, namespace: %v, name: %v", job.Namespace, job.Name)
			}(task)
		}
		wg.Wait()
	}()

	w.wg.Wait()
	log.Print("Task watch is complete")
}

func init() {
	Register(&taskWatcher{})
}
