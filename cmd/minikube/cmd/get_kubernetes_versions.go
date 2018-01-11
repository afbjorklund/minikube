/*
Copyright 2016 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"k8s.io/minikube/pkg/minikube/kubernetes_versions"
)

// getK8sVersionCmd represents the ip command
var getK8sVersionCmd = &cobra.Command{
	Use:   "get-k8s-version",
	Short: "Gets the default Kubernetes version used with minikube when using the kubeadm bootstrapper",
	Long:  "Gets the default Kubernetes version used with minikube when using the kubeadm bootstrapper",
	Run: func(cmd *cobra.Command, args []string) {
		kubernetes_versions.PrintKubernetesVersion(os.Stdout)
	},
}

// getK8sVersionsCmd represents the ip command
var getK8sVersionsCmd = &cobra.Command{
	Use:   "get-k8s-versions",
	Short: "Gets the list of Kubernetes versions available for minikube when using the localkube bootstrapper",
	Long:  `Gets the list of Kubernetes versions available for minikube when using the localkube bootstrapper.`,
	Run: func(cmd *cobra.Command, args []string) {
		kubernetes_versions.PrintKubernetesVersionsFromGCS(os.Stdout)
	},
}

func init() {
	RootCmd.AddCommand(getK8sVersionCmd)
	RootCmd.AddCommand(getK8sVersionsCmd)
}
