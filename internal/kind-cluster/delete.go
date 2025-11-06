package kind_cluster

import (
	"log/slog"
	"os"
	"os/exec"

	"go.yaml.in/yaml/v3"
)

func DeleteCluster(configPath string) error {
	file, err := os.Open(configPath)
	if err != nil {
		slog.Error("Could not open config file", "err", err)
	}
	defer file.Close()

	var config KindConfig
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		slog.Error("Could not decode config", "err", err)
	}

	if config.Name == "" {
		slog.Error("Cluster name is empty")
	}

	cmd := exec.Command("kind", "delete", "cluster", "--name", config.Name)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		slog.Error("Delete cluster failed", "name", config.Name, "err", err)
	}

	slog.Info("Delete cluster successfully")
	return nil
}
