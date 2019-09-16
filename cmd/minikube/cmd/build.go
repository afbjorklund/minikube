/*
Copyright 2019 The Kubernetes Authors All rights reserved.

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
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"

	"github.com/golang/glog"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"k8s.io/minikube/pkg/minikube/cluster"
	"k8s.io/minikube/pkg/minikube/constants"
	"k8s.io/minikube/pkg/minikube/exit"
	"k8s.io/minikube/pkg/minikube/machine"
)

const (
	podmanFlag = "podman"
)

func getDockerBinary() (string, error) {
	version := constants.DefaultBuildDockerVersion
	if runtime.GOOS == "windows" {
		version = constants.FallbackBuildDockerVersion
	}
	archive, err := machine.CacheDockerArchive("docker", version, runtime.GOOS, runtime.GOARCH)
	if err != nil {
		return "", errors.Wrap(err, "Failed to download docker")
	}

	binary := "docker"
	if runtime.GOOS == "windows" {
		binary = "docker.exe"
	}
	path := filepath.Join(filepath.Dir(archive), binary)
	_, err = os.Stat(path)
	if err == nil {
		return path, nil
	}
	if !os.IsNotExist(err) {
		return "", errors.Wrapf(err, "stat %s", path)
	}
	err = machine.ExtractBinary(archive, path, fmt.Sprintf("docker/%s", binary))
	if err != nil {
		return "", errors.Wrap(err, "Failed to extract docker")
	}

	return path, nil
}

// buildCmd represents the build command
var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build a container image",
	Long: `Run the docker client, download it if necessary.
Examples:
minikube build .`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			exit.UsageT("Usage: minikube build -- [OPTIONS] PATH | URL | -")
		}
		api, err := machine.NewAPIClient()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting client: %v\n", err)
			os.Exit(1)
		}
		defer api.Close()

		usePodman := viper.GetBool(podmanFlag)
		useDocker := !usePodman

		var path string
		options := []string{}
		if useDocker {
			path, err = getDockerBinary()

			envMap, err := cluster.GetHostDockerEnv(api)
			if err != nil {
				exit.WithError("Failed to get docker env", err)
			}

			tlsVerify := envMap["DOCKER_TLS_VERIFY"]
			certPath := envMap["DOCKER_CERT_PATH"]
			dockerHost := envMap["DOCKER_HOST"]

			if tlsVerify != "" {
				options = append(options, "--tlsverify")
			}
			if certPath != "" {
				options = append(options, "--tlscacert", filepath.Join(certPath, "ca.pem"))
				options = append(options, "--tlscert", filepath.Join(certPath, "cert.pem"))
				options = append(options, "--tlskey", filepath.Join(certPath, "key.pem"))
			}
			options = append(options, "-H", dockerHost)
		} else {
			path = "podman"
		}

		options = append(options, "build")
		args = append(options, args...)

		glog.Infof("Running %s %v", path, args)
		if useDocker {
			c := exec.Command(path, args...)
			c.Stdin = os.Stdin
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			if err := c.Run(); err != nil {
				var rc int
				if exitError, ok := err.(*exec.ExitError); ok {
					waitStatus := exitError.Sys().(syscall.WaitStatus)
					rc = waitStatus.ExitStatus()
				} else {
					fmt.Fprintf(os.Stderr, "Error running %s: %v\n", path, err)
					rc = 1
				}
				os.Exit(rc)
			}
		} else {
			cmd := []string{"sudo", path}
			args = append(cmd, args...)
			err := cluster.CreateSSHShell(api, args)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error running ssh sudo %s: %v\n", path, err)
				os.Exit(1)
			}
		}
	},
}

func init() {
	buildCmd.Flags().Bool(podmanFlag, false, "Use Podman to build, instead of Docker")
	if err := viper.BindPFlags(buildCmd.Flags()); err != nil {
		exit.WithError("unable to bind flags", err)
	}
}
