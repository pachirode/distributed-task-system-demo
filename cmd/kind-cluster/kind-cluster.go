package main

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/util/sets"

	kind_cluster "github.com/pachirode/distributed-task-system-demo/internal/kind-cluster"
)

const helpText = `Usage: main [flags] arg [arg...]

This is a pflag example.

Flags:
`

var (
	roleWork     = pflag.IntP("work-num", "w", 0, "role work number")
	roleControl  = pflag.IntP("control-num", "c", 1, "role control-plane number")
	roleExternal = pflag.IntP("external-load-number", "e", 0, "role external-load-balancer number")
	name         = pflag.StringP("name", "n", "test-k8s", "cluster name")
	operation    = pflag.StringP("operation", "o", "create", "kind options")
	subOperation = pflag.StringP("sub-operation", "s", "nodes", "kind cluster sub operation")
	image        = pflag.StringP("image", "i", "", "kind load docker-image name")
	loadNodes    = pflag.StringSlice("load-node", []string{"all"}, "kind load docker-image node name")
	configFile   = pflag.StringP("config-file", "f", "./kind-config.yaml", "kind config file path")
	help         = pflag.BoolP("help", "h", false, "Show this help")

	usage = func() {
		fmt.Printf("%s", helpText)
		pflag.PrintDefaults()
	}
)

func main() {
	pflag.Usage = usage
	pflag.Parse()

	if *help {
		pflag.Usage()
		return
	}

	if *roleControl < 1 {
		*roleControl = 1
	}

	if *roleWork < 0 {
		*roleWork = 0
	}

	if *roleWork < 2 || *roleExternal < 0 {
		*roleExternal = 0
	}

	if *image != "" {
		imageInfo := strings.Split(strings.TrimSpace(*image), ":")
		if len(imageInfo) == 1 {
			*image = fmt.Sprintf("%s:latest", *image)
		}
	}

	clusterNodes := sets.NewString(*loadNodes...)

	switch *operation {
	case "create":
		err := kind_cluster.CreateConfigFile(*configFile, *name, *roleControl, *roleWork, *roleExternal)
		if err != nil {
			slog.Error("Create config file failed", "err", err)
			os.Exit(-1)
		}
		err = kind_cluster.CreateKindClusterWithConfig(*configFile)
		if err != nil {
			slog.Error("Create kind cluster failed")
		}

	case "delete":
		err := kind_cluster.DeleteCluster(*configFile)
		if err != nil {
			slog.Error("Delete cluster failed", "err", err)
		}
	case "info":
		switch *subOperation {
		case "config":
			err := kind_cluster.GetClusterInfoSetting(*name)
			if err != nil {
				slog.Error("Get cluster info failed", "name", *name, "err", err)
			}
		case "nodes":
			nodes, err := kind_cluster.GetKindNodes(*name)
			if err != nil {
				slog.Error("Get cluster nodes failed")
			}

			slog.Info("Get cluster nodes successfully", "nodes", nodes)
		case "images":
			if clusterNodes.Has("all") {
				nodes, err := kind_cluster.GetKindNodes(*name)
				if err != nil {
					slog.Error("Get cluster nodes failed")
				}
				kind_cluster.GetNodeImage(nodes)
			} else {
				kind_cluster.GetNodeImage(*loadNodes)
			}
		case "container":
			if clusterNodes.Has("all") {
				nodes, err := kind_cluster.GetKindNodes(*name)
				if err != nil {
					slog.Error("Get cluster nodes failed")
				}
				kind_cluster.GetNodeContainer(nodes)
			} else {
				kind_cluster.GetNodeContainer(*loadNodes)
			}
		}
	case "load":
		if *image == "" {
			slog.Error("Image name is empty")
			return
		}
		switch *subOperation {
		case "pull":
			if clusterNodes.Has("all") {
				nodes, err := kind_cluster.GetKindNodes(*name)
				if err != nil {
					slog.Error("Get cluster nodes failed")
				}
				kind_cluster.LoadDockerImageToNode(nodes, *image)
			} else {
				kind_cluster.LoadDockerImageToNode(*loadNodes, *image)
			}
		default:
			if clusterNodes.Has("all") {
				kind_cluster.LoadAllDockerImage(*image, *name)
			} else {
				kind_cluster.LoadDockerImage(*image, *name, *loadNodes)
			}
		}
	}
}
