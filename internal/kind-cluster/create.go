package kind_cluster

import (
	"log/slog"
	"os"
	"os/exec"
)

func CreateKindClusterWithConfig(configPath string) error {
	cmd := exec.Command("kind", "create", "cluster", "--config", configPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		slog.Error("Error to create cluster", "err", err)
	}

	slog.Info("Cluster created successfully with custom config!")
	return nil
}
