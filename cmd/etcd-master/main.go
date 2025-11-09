package main

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"sync"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

const (
	etcdEndpoint = "127.0.0.1:2379"
	electionKey  = "/service/master"
)

func main() {
	nodeCount := 3
	var wg sync.WaitGroup

	for i := 0; i < nodeCount; i++ {
		wg.Add(1)
		nodeID := fmt.Sprintf("node-%d", i)
		go func(id string) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					slog.Error("panic occurred", "err", r, "id", nodeID)
				}
			}()
			runNode(id)
		}(nodeID)
	}

	wg.Wait()
}

func runNode(nodeID string) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{etcdEndpoint},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		slog.Error("Error to connect etcd", "err", err)
	}
	defer cli.Close()

	for {
		var (
			err        error
			sess       *concurrency.Session
			retryCount = 0
		)

		for {
			slog.Info("creating session", "attemp", retryCount+1, "id", nodeID)
			sess, err = concurrency.NewSession(cli, concurrency.WithTTL(5))
			if err == nil {
				slog.Info("creating session successfully", "id", nodeID)
				break
			}
			retryCount++
			if retryCount >= 3 {
				slog.Info("Error to create session", "err", err, "id", nodeID)
				return
			}

			backoff := time.Duration(1<<retryCount) * time.Second
			slog.Warn("Create session failed", "err", err, "retryTime", backoff, "id", nodeID)
			time.Sleep(backoff)
		}

		slog.Info("starting create election", "id", nodeID)
		election := concurrency.NewElection(sess, electionKey)
		ctx := context.Background()

		slog.Info("starting election campain", "id", nodeID)
		if err := election.Campaign(ctx, nodeID); err != nil {
			slog.Error("Campain failed", "err", err, "id", nodeID)
			sess.Close()
			time.Sleep(2 * time.Second)
			continue
		}

		slog.Info("node get master", "id", nodeID)
		workTime := time.Duration(rand.IntN(10)+5) * time.Second
		time.Sleep(workTime)

		slog.Info("resign later", "resignTime", workTime, "id", nodeID)
		election.Resign(ctx)
		sess.Close()

		time.Sleep(time.Duration(rand.IntN(5) + 3))
	}
}
