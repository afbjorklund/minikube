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

package util

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/blang/semver"
	units "github.com/docker/go-units"
	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
)

const (
	downloadURL = "https://storage.googleapis.com/minikube/releases/%s/minikube-%s-amd64%s"
)

// CalculateSizeInMB returns the number of MB in the human readable string
func CalculateSizeInMB(humanReadableSize string) (int, error) {
	_, err := strconv.ParseInt(humanReadableSize, 10, 64)
	if err == nil {
		humanReadableSize += "mb"
	}
	// parse the size suffix binary instead of decimal so that 1G -> 1024MB instead of 1000MB
	size, err := units.RAMInBytes(humanReadableSize)
	if err != nil {
		return 0, fmt.Errorf("FromHumanSize: %v", err)
	}

	return int(size / units.MiB), nil
}

// ConvertMBToBytes converts MB to bytes
func ConvertMBToBytes(mbSize int) int64 {
	return int64(mbSize) * units.MiB
}

// ConvertBytesToMB converts bytes to MB
func ConvertBytesToMB(byteSize int64) int {
	return int(ConvertUnsignedBytesToMB(uint64(byteSize)))
}

// ConvertUnsignedBytesToMB converts bytes to MB
func ConvertUnsignedBytesToMB(byteSize uint64) int64 {
	return int64(byteSize / units.MiB)
}

// LocalCPU returns the cpu usage
// returns: busy, idle (%)
func LocalCPU() (int, int, error) {
	p, err := cpu.Percent(time.Second, false)
	if err != nil {
		return 0, 0, err
	}
	return int(p[0]), int(100.0 - p[0]), nil
}

func mb(bytes uint64) uint64 {
	return bytes / 1024 / 1024
}

// LocalMem returns the memory free
// returns: total, available (in mb)
func LocalMem() (uint64, uint64, error) {
	v, err := mem.VirtualMemory()
	if err != nil {
		return 0, 0, err
	}
	return mb(v.Total), mb(v.Free), nil
}

// LocalDisk returns the disk free
// returns: total, available (in mb)
func LocalDisk(mountpoint string) (uint64, uint64, error) {
	d, err := disk.Usage(mountpoint)
	if err != nil {
		return 0, 0, err
	}
	return mb(d.Total), mb(d.Free), nil
}

// ParseVMStat parses the output of the `vmstat` command
// returns: busy, idle (%)
func ParseVMStat(out string) (int, int, error) {
	//procs -----------memory---------- ---swap-- -----io---- -system-- ------cpu-----
	//r  b   swpd   free   buff  cache   si   so    bi    bo   in   cs us sy id wa st
	//0  0      0 24562996 667340 3463592    0    0    53    40  279  222  6  4 90  0  0
	//0  0      0 24562680 667348 3463632    0    0     0    68 1192 7571  2  1 97  0  0
	outlines := strings.Split(out, "\n")
	l := len(outlines)
	for _, line := range outlines[l-2 : l-1] {
		parsedLine := strings.Fields(line)
		if len(parsedLine) < 17 {
			continue
		}
		us, err := strconv.Atoi(parsedLine[12])
		if err != nil {
			return 0, 0, err
		}
		sy, err := strconv.Atoi(parsedLine[13])
		if err != nil {
			return 0, 0, err
		}
		id, err := strconv.Atoi(parsedLine[14])
		if err != nil {
			return 0, 0, err
		}
		wa, err := strconv.Atoi(parsedLine[15])
		if err != nil {
			return 0, 0, err
		}
		st, err := strconv.Atoi(parsedLine[16])
		if err != nil {
			return 0, 0, err
		}
		return us + sy + wa + st, id, nil
	}
	return 0, 0, errors.New("No matching data found")
}

// ParseMemFree parses the output of the `free -m` command
// returns: total, available
func ParseMemFree(out string) (uint64, uint64, error) {
	//             total        used        free      shared  buff/cache   available
	//Mem:           1987         706         194           1        1086        1173
	//Swap:             0           0           0
	outlines := strings.Split(out, "\n")
	l := len(outlines)
	for _, line := range outlines[1 : l-1] {
		parsedLine := strings.Fields(line)
		if len(parsedLine) < 7 {
			continue
		}
		t, err := strconv.ParseUint(parsedLine[1], 10, 64)
		if err != nil {
			return 0, 0, err
		}
		a, err := strconv.ParseUint(parsedLine[6], 10, 64)
		if err != nil {
			return 0, 0, err
		}
		m := strings.Trim(parsedLine[0], ":")
		if m == "Mem" {
			return t, a, nil
		}
	}
	return 0, 0, errors.New("No matching data found")
}

// ParseDiskFree parses the output of the `df -m` command
// returns: total, available
func ParseDiskFree(out string, mountpoint string) (uint64, uint64, error) {
	// Filesystem     1M-blocks  Used Available Use% Mounted on
	// /dev/sda1          39643  3705     35922  10% /
	outlines := strings.Split(out, "\n")
	l := len(outlines)
	for _, line := range outlines[1 : l-1] {
		parsedLine := strings.Fields(line)
		if len(parsedLine) < 6 {
			continue
		}
		t, err := strconv.ParseUint(parsedLine[1], 10, 64)
		if err != nil {
			return 0, 0, err
		}
		a, err := strconv.ParseUint(parsedLine[3], 10, 64)
		if err != nil {
			return 0, 0, err
		}
		m := parsedLine[5]
		if m == mountpoint {
			return t, a, nil
		}
	}
	return 0, 0, errors.New("No matching data found")
}

// GetBinaryDownloadURL returns a suitable URL for the platform
func GetBinaryDownloadURL(version, platform string) string {
	switch platform {
	case "windows":
		return fmt.Sprintf(downloadURL, version, platform, ".exe")
	default:
		return fmt.Sprintf(downloadURL, version, platform, "")
	}
}

// ChownR does a recursive os.Chown
func ChownR(path string, uid, gid int) error {
	return filepath.Walk(path, func(name string, info os.FileInfo, err error) error {
		if err == nil {
			err = os.Chown(name, uid, gid)
		}
		return err
	})
}

// MaybeChownDirRecursiveToMinikubeUser changes ownership of a dir, if requested
func MaybeChownDirRecursiveToMinikubeUser(dir string) error {
	if os.Getenv("CHANGE_MINIKUBE_NONE_USER") != "" && os.Getenv("SUDO_USER") != "" {
		username := os.Getenv("SUDO_USER")
		usr, err := user.Lookup(username)
		if err != nil {
			return errors.Wrap(err, "Error looking up user")
		}
		uid, err := strconv.Atoi(usr.Uid)
		if err != nil {
			return errors.Wrapf(err, "Error parsing uid for user: %s", username)
		}
		gid, err := strconv.Atoi(usr.Gid)
		if err != nil {
			return errors.Wrapf(err, "Error parsing gid for user: %s", username)
		}
		if err := ChownR(dir, uid, gid); err != nil {
			return errors.Wrapf(err, "Error changing ownership for: %s", dir)
		}
	}
	return nil
}

// ParseKubernetesVersion parses the Kubernetes version
func ParseKubernetesVersion(version string) (semver.Version, error) {
	return semver.Make(version[1:])
}
