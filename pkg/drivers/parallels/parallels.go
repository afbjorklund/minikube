// +build !darwin

package parallels

import "k8s.io/minikube/pkg/libmachine/drivers"

func NewDriver(hostName, storePath string) drivers.Driver {
	return drivers.NewDriverNotSupported("parallels", hostName, storePath)
}
