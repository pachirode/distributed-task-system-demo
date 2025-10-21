package biz

import (
	"k8s.io/client-go/kubernetes"

	"github.com/pachirode/distributed-task-system-demo/internal/system-watch/store"
)

type Config struct {
	Store     store.IStore
	Clientset kubernetes.Interface
}
