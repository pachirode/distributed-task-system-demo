package main

import (
	"flag"
	"log"
	"path/filepath"
	"time"

	genericapiserver "k8s.io/apiserver/pkg/server"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	system_watch "github.com/pachirode/distributed-task-system-demo/internal/system-watch"
	"github.com/pachirode/distributed-task-system-demo/pkg/db"
)

func main() {
	var kubecfg *string
	if home := homedir.HomeDir(); home != "" {
		kubecfg = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "Optional absolute path to kubeconfig")
	} else {
		kubecfg = flag.String("kubeconfig", "", "Absolute path to kubeconfig")
	}

	config, err := clientcmd.BuildConfigFromFlags("", *kubecfg)
	if err != nil {
		log.Fatalf("%v", err)
		return
	}

	config.QPS = 50
	config.Burst = 100
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("%v", err)
		return
	}

	cfg := system_watch.SystemWatchConfig{
		MySQLOptions: &db.MySQLOptions{
			Host:                  "127.0.0.1:33306",
			Username:              "root",
			Password:              "system-watch",
			Database:              "system_watch",
			MaxIdleConnections:    100,
			MaxOpenConnections:    100,
			MaxConnectionLifeTime: time.Duration(10) * time.Second,
		},
		RedisOptions: &db.RedisOptions{
			Addr:         "127.0.0.1:36379",
			Username:     "",
			Password:     "system-watch",
			Database:     0,
			MaxRetries:   3,
			MinIdleConns: 0,
			DialTimeout:  5 * time.Second,
			ReadTimeout:  3 * time.Second,
			WriteTimeout: 3 * time.Second,
			PoolSize:     10,
		},
		Clientset: clientset,
	}

	sw, err := cfg.New()
	if err != nil {
		log.Fatalf("%v", err)
		return
	}

	stopCh := genericapiserver.SetupSignalHandler()
	sw.Run(stopCh)
}
