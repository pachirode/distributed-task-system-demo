package kind_cluster

import (
	"log/slog"
	"os"

	"go.yaml.in/yaml/v3"
)

type Node struct {
	Role string `yaml:"role"`
}

type KindConfig struct {
	Kind       string `yaml:"kind"`
	ApiVersion string `yaml:"apiVersion"`
	Name       string `yaml:"name"`
	Nodes      []Node `yaml:"nodes"`
}

func CreateConfigFile(filePath string, name string, controlNum int, workNum int, externalNum int) error {
	config := KindConfig{
		Kind:       "Cluster",
		ApiVersion: "kind.x-k8s.io/v1alpha4",
		Name:       name,
		Nodes:      []Node{},
	}

	for i := 0; i < controlNum; i++ {
		config.Nodes = append(config.Nodes, Node{
			Role: "control-plane",
		})
	}

	for i := 0; i < workNum; i++ {
		config.Nodes = append(config.Nodes, Node{
			Role: "worker",
		})
	}

	for i := 0; i < externalNum; i++ {
		config.Nodes = append(config.Nodes, Node{
			Role: "external-load-balancer",
		})
	}

	file, err := os.Create(filePath)
	if err != nil {
		slog.Error("could not create config file", "err", err)
	}
	defer file.Close()

	encoder := yaml.NewEncoder(file)
	encoder.SetIndent(2)
	if err := encoder.Encode(config); err != nil {
		slog.Error("could not encode config to file", "err", err)
	}

	return nil
}
