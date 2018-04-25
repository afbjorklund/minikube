package profile

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/spf13/viper"
	cfg "k8s.io/minikube/pkg/minikube/config"
	"k8s.io/minikube/pkg/minikube/constants"
	pkgutil "k8s.io/minikube/pkg/util"
)

// SaveConfig saves profile cluster configuration in
// $MINIKUBE_HOME/profiles/<profilename>/config.json
func SaveConfig(profile string, clusterConfig cfg.Config) error {
	data, err := json.MarshalIndent(clusterConfig, "", "    ")
	if err != nil {
		return err
	}

	profileConfigFile := constants.GetProfileFile(viper.GetString(cfg.MachineProfile))

	if err := os.MkdirAll(filepath.Dir(profileConfigFile), 0700); err != nil {
		return err
	}

	if err := saveConfigToFile(data, profileConfigFile); err != nil {
		return err
	}

	return nil
}

func saveConfigToFile(data []byte, file string) error {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return ioutil.WriteFile(file, data, 0600)
	}

	tmpfi, err := ioutil.TempFile(filepath.Dir(file), "config.json.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(tmpfi.Name())

	if err = ioutil.WriteFile(tmpfi.Name(), data, 0600); err != nil {
		return err
	}

	if err = tmpfi.Close(); err != nil {
		return err
	}

	if err = os.Remove(file); err != nil {
		return err
	}

	if err = os.Rename(tmpfi.Name(), file); err != nil {
		return err
	}
	return nil
}

func LoadConfigFromFile(profile string) (cfg.Config, error) {
	var cc cfg.Config

	if profile == "" {
		return cc, fmt.Errorf("Profile name cannot be empty.")
	}

	profileConfigFile := constants.GetProfileFile(profile)

	if _, err := os.Stat(profileConfigFile); os.IsNotExist(err) {
		return cc, err
	}

	data, err := ioutil.ReadFile(profileConfigFile)
	if err != nil {
		return cc, err
	}

	if err := json.Unmarshal(data, &cc); err != nil {
		return cc, err
	}

	cc.MachineConfig.Downloader = pkgutil.DefaultDownloader{}

	return cc, nil
}

func LoadClusterConfigs() ([]cfg.Config, error) {
	files := constants.GetProfileFiles()

	configs := make([]cfg.Config, len(files))
	for i, f := range files {
		c, err := loadConfigFromFile(f)
		if err != nil {
			return []cfg.Config{}, errors.Wrapf(err, "Error loading config from file: %s", f)
		}
		configs[i] = c
	}

	return configs, nil
}

func loadConfigFromFile(file string) (cfg.Config, error) {
	var c cfg.Config

	reader, err := os.Open(file)
	defer reader.Close()
	if err != nil {
		return c, err
	}

	err = json.NewDecoder(reader).Decode(&c)
	return c, err
}
