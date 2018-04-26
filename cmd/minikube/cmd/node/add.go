package node

import (
	"fmt"
	"os"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"k8s.io/minikube/cmd/minikube/profile"
	cmdutil "k8s.io/minikube/cmd/util"
	"k8s.io/minikube/pkg/minikube/config"
)

func NewCmdAdd() *cobra.Command {
	return &cobra.Command{
		Use:   "add <node_name>",
		Short: "Adds a node to the cluster",
		Long:  "Adds a node tot the cluster",
		Run:   add,
	}
}

func add(cmd *cobra.Command, args []string) {
	// TODO Make clusterName into `--cluster=` flag
	clusterName := viper.GetString(config.MachineProfile)

	nodeName := ""
	if len(args) > 0 {
		nodeName = args[0]
	}

	cfg, err := profile.LoadConfigFromFile(clusterName)
	if err != nil {
		glog.Errorln("Error loading profile config: ", err)
		cmdutil.MaybeReportErrorAndExit(err)
	}

	if nodeName == "" {
		nodeName = fmt.Sprintf("node-%d", len(cfg.Nodes)+1)
	}

	node := config.NodeConfig{
		Name: nodeName,
	}

	cfg.Nodes = append(cfg.Nodes, node)

	if err := profile.SaveConfig(clusterName, cfg); err != nil {
		glog.Errorln("Error saving profile cluster configuration: ", err)
		os.Exit(1)
	}

	fmt.Println("Added node: ", node.Name)
}
