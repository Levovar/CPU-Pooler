package types

import (
	"errors"
	"strings"
	"github.com/go-yaml/yaml"
	"github.com/golang/glog"
	"github.com/Levovar/CPU-Pooler/pkg/k8sclient"
	"io/ioutil"
	"path/filepath"
)

// Pool defines cpupool
type Pool struct {
	CPUs string `yaml:"cpus"`
}

// PoolConfig defines pool configuration for a node
type PoolConfig struct {
	Pools        map[string]Pool   `yaml:"pools"`
	NodeSelector map[string]string `yaml:"nodeSelector"`
}

// PoolConfigDir defines the pool configuration file location
var PoolConfigDir = "/etc/cpu-pooler"

func DeterminePoolConfig() (PoolConfig,string,error) {
	nodeLabels,err := k8sclient.GetNodeLabels()
	if err != nil {
		return PoolConfig{}, "", errors.New("Following error happend when trying to read K8s API server Node object:" + err.Error())
	}
	return readPoolConfig(nodeLabels)
}

// ReadPoolConfig implements pool configuration file reading
func readPoolConfig(labelMap map[string]string) (PoolConfig, string, error) {
	files, err := filepath.Glob(filepath.Join(PoolConfigDir, "poolconfig-*"))
	if err != nil {
		return PoolConfig{}, "", err
	}
	for _, f := range files {
		pools, err := ReadPoolConfigFile(f)
		if err != nil {
			return PoolConfig{}, "", err
		}
		if labelMap == nil {
			glog.Infof("Using first configuration file %s as pool config in lieu of missing Node information", f)
			return pools, f, nil
		}
		for label, labelValue := range labelMap {
			if value, ok := pools.NodeSelector[label]; ok {
				if value == labelValue {
					glog.Infof("Using configuration file %s for pool config", f)
					return pools, f, nil
				}
			}
		}
	}
	return PoolConfig{}, "", errors.New("No matching pool configuration file found for provided nodeSelector labels")
}

// ReadPoolConfigFile reads a pool configuration file
func ReadPoolConfigFile(name string) (PoolConfig, error) {
	var pools PoolConfig
	file, err := ioutil.ReadFile(name)
	if err != nil {
		return PoolConfig{}, errors.New("Could not read poolconfig file named: " + name + " because:" + err.Error())
	} else {
		err = yaml.Unmarshal([]byte(file), &pools)
		if err != nil {
			return PoolConfig{}, errors.New("Poolconfig file could not be parsed because:" + err.Error())
		}
	}
	return pools, err
}

func (poolConf PoolConfig) SelectPool(prefix string) Pool {
	for poolName, pool := range poolConf.Pools {
		if strings.HasPrefix(poolName, prefix) {
			return pool
		}
	}
	return Pool{}
}