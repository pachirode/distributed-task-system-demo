package kind_cluster

import (
	"bytes"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
)

func GetClusterInfoSetting(name string) error {
	cmd := exec.Command("kind", "get", "clusters")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		slog.Error("Failed to get clusters", "err", err)
		return err
	}

	cmd = exec.Command("kind", "get", "kubeconfig", "--name", name)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		slog.Error("Failed to get kubeconfig for the cluster", "err", err)
		return err
	}

	slog.Info("Cluster information retrieved successfully")
	return nil
}

func GetKindNodes(clusterName string) ([]string, error) {
	cmd := exec.Command("kind", "get", "nodes", "--name", clusterName)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	nodes := strings.Fields(string(out))
	return nodes, nil
}

func ExecInNode(node, command string, args ...string) (string, error) {
	allArgs := append([]string{"exec", node, command}, args...)
	cmd := exec.Command("docker", allArgs...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()
	return out.String(), err
}

func GetNodeContainer(nodes []string) {

	for _, node := range nodes {
		fmt.Println("=== Node: ", node, " ===")

		containers, err := ExecInNode(node, "crictl", "ps", "-a")
		if err != nil {
			fmt.Println("Error getting containers: ", err)
		} else {
			fmt.Println("Containers:\n", containers)
		}
	}
}

func GetNodeImage(nodes []string) {
	for _, node := range nodes {
		fmt.Println("=== Node: ", node, " ===")

		containers, err := ExecInNode(node, "crictl", "images")
		if err != nil {
			fmt.Println("Error getting containers: ", err)
		} else {
			fmt.Println("Containers:\n", containers)
		}
	}

}
