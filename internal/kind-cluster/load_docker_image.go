package kind_cluster

import (
	"log/slog"
	"os"
	"os/exec"
	"strings"
)

func ImageExists(image string) (bool, error) {
	cmd := exec.Command("docker", "images", "--format", "{{.Repository}}:{{.Tag}}")
	out, err := cmd.Output()
	if err != nil {
		return false, err
	}

	images := strings.Split(strings.TrimSpace(string(out)), "\n")
	for _, img := range images {
		slog.Debug("Get host images", "image")
		if img == image {
			return true, nil
		}
	}
	return false, nil
}

func LoadDockerImageToNode(nodes []string, image string) {
	for _, node := range nodes {
		cmd := exec.Command("docker", "exec", node, "crictl", "pull", image)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err := cmd.Run()
		if err != nil {
			slog.Error("Failed to load docker image to node", "image", image, "node", node, "err", err)
			return
		}

		slog.Info("Successfully loaded docker image to node", "image", image, "node", node)
	}
}

func LoadDockerImage(image string, clusterName string, nodes []string) {
	exists, err := ImageExists(image)
	if err != nil {
		slog.Error("Get host image message failed", "err", err)
		return
	}

	if !exists {
		slog.Error("Image not found on host")
		return
	}

	for _, node := range nodes {
		cmd := exec.Command("kind", "load", "docker-image", image, "--name", clusterName, "--nodes", node)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err = cmd.Run()
		if err != nil {
			slog.Error("Failed to load docker image to kind cluster", "image", image, "err", err)
			return
		}
		slog.Info("Image loaded to kind cluster successfully", "cluster", clusterName, "node", node)
	}
}

func LoadAllDockerImage(image string, clusterName string) {
	exists, err := ImageExists(image)
	if err != nil {
		slog.Error("Get host image message failed", "err", err)
		return
	}

	if !exists {
		slog.Error("Image not found on host")
		return
	}

	cmd := exec.Command("kind", "load", "docker-image", image, "--name", clusterName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		slog.Error("Failed to load docker image to kind cluster", "image", image, "err", err)
		return
	}
	slog.Info("Image loaded to kind cluster successfully", "cluster", clusterName)
}
