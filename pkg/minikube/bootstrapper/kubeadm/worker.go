package kubeadm

import (
	"fmt"
	"io"

	"github.com/pkg/errors"
	"k8s.io/minikube/pkg/minikube"
	cfg "k8s.io/minikube/pkg/minikube/config"
)

type WorkerBootstrapper struct {
	config cfg.KubernetesConfig
	ui     io.Writer
}

func NewWorkerBootstrapper(c cfg.KubernetesConfig, ui io.Writer) minikube.Bootstrapper {
	return &WorkerBootstrapper{config: c, ui: ui}
}

func (nb *WorkerBootstrapper) Bootstrap(n minikube.Node) error {
	ip, err := n.IP()
	if err != nil {
		return errors.Wrap(err, "Error getting node's IP")
	}

	runner, err := n.Runner()
	if err != nil {
		return errors.Wrap(err, "Error getting node's runner")
	}

	b := NewKubeadmBootstrapperForRunner(n.MachineName(), ip, runner)

	fmt.Fprintln(nb.ui, "Moving assets into node...")
	if err := b.UpdateNode(nb.config); err != nil {
		return errors.Wrap(err, "Error updating node")
	}
	fmt.Fprintln(nb.ui, "Setting up certs...")
	if err := b.SetupCerts(nb.config); err != nil {
		return errors.Wrap(err, "Error configuring authentication")
	}

	fmt.Fprintln(nb.ui, "Joining node to cluster...")
	if err := b.JoinNode(nb.config); err != nil {
		return errors.Wrap(err, "Error joining node to cluster")
	}
	return nil
}
