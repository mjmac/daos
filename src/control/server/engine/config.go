//
// (C) Copyright 2019-2021 Intel Corporation.
//
// SPDX-License-Identifier: BSD-2-Clause-Patent
//

package engine

import (
	"strings"

	"github.com/pkg/errors"

	"github.com/mjmac/soad/src/control/server/storage"
	"github.com/mjmac/soad/src/control/system"
)

const (
	maxHelperStreamCount = 2
)

// StorageConfig encapsulates an I/O Engine's storage configuration.
type StorageConfig struct {
	SCM  storage.ScmConfig  `yaml:",inline"`
	Bdev storage.BdevConfig `yaml:",inline"`
}

// Validate ensures that the configuration meets minimum standards.
func (sc *StorageConfig) Validate() error {
	if err := sc.SCM.Validate(); err != nil {
		return errors.Wrap(err, "scm config validation failed")
	}
	if err := sc.Bdev.Validate(); err != nil {
		return errors.Wrap(err, "bdev config validation failed")
	}
	return nil
}

// FabricConfig encapsulates networking fabric configuration.
type FabricConfig struct {
	Provider        string `yaml:"provider,omitempty" cmdEnv:"CRT_PHY_ADDR_STR"`
	Interface       string `yaml:"fabric_iface,omitempty" cmdEnv:"OFI_INTERFACE"`
	InterfacePort   int    `yaml:"fabric_iface_port,omitempty" cmdEnv:"OFI_PORT,nonzero"`
	PinnedNumaNode  *uint  `yaml:"pinned_numa_node,omitempty" cmdLongFlag:"--pinned_numa_node" cmdShortFlag:"-p"`
	CrtCtxShareAddr uint32 `yaml:"crt_ctx_share_addr,omitempty" cmdEnv:"CRT_CTX_SHARE_ADDR"`
	CrtTimeout      uint32 `yaml:"crt_timeout,omitempty" cmdEnv:"CRT_TIMEOUT"`
}

// Update fills in any missing fields from the provided FabricConfig.
func (fc *FabricConfig) Update(other FabricConfig) {
	if fc.Provider == "" {
		fc.Provider = other.Provider
	}
	if fc.Interface == "" {
		fc.Interface = other.Interface
	}
	if fc.InterfacePort == 0 {
		fc.InterfacePort = other.InterfacePort
	}
	if fc.CrtCtxShareAddr == 0 {
		fc.CrtCtxShareAddr = other.CrtCtxShareAddr
	}
	if fc.CrtTimeout == 0 {
		fc.CrtTimeout = other.CrtTimeout
	}
}

// GetNumaNode retrieves the value configured by the YML if it was supplied
// returns an error if it was not configured.
func (fc *FabricConfig) GetNumaNode() (uint, error) {
	if fc.PinnedNumaNode != nil {
		return *fc.PinnedNumaNode, nil
	}
	return 0, errors.New("pinned NUMA node was not configured")
}

// Validate ensures that the configuration meets minimum standards.
func (fc *FabricConfig) Validate() error {
	if fc.Provider == "" {
		return errors.New("provider not set")
	}
	if fc.Interface == "" {
		return errors.New("fabric_iface not set")
	}
	if fc.InterfacePort == 0 {
		return errors.New("fabric_iface_port not set")
	}
	return nil
}

func mergeEnvVars(curVars []string, newVars []string) (merged []string) {
	mergeMap := make(map[string]string)
	for _, pair := range curVars {
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) != 2 || kv[0] == "" || kv[1] == "" {
			continue
		}
		// strip duplicates in curVars; shouldn't be any
		// but this will ensure it.
		if _, found := mergeMap[kv[0]]; found {
			continue
		}
		mergeMap[kv[0]] = kv[1]
	}

	mergedKeys := make(map[string]struct{})
	for _, pair := range newVars {
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) != 2 || kv[0] == "" || kv[1] == "" {
			continue
		}
		// strip duplicates in newVars
		if _, found := mergedKeys[kv[0]]; found {
			continue
		}
		mergedKeys[kv[0]] = struct{}{}
		mergeMap[kv[0]] = kv[1]
	}

	merged = make([]string, 0, len(mergeMap))
	for key, val := range mergeMap {
		merged = append(merged, strings.Join([]string{key, val}, "="))
	}

	return
}

// Config encapsulates an I/O Engine's configuration.
type Config struct {
	Rank              *system.Rank  `yaml:"rank,omitempty"`
	Modules           string        `yaml:"modules,omitempty" cmdLongFlag:"--modules" cmdShortFlag:"-m"`
	TargetCount       int           `yaml:"targets,omitempty" cmdLongFlag:"--targets,nonzero" cmdShortFlag:"-t,nonzero"`
	HelperStreamCount int           `yaml:"nr_xs_helpers" cmdLongFlag:"--xshelpernr" cmdShortFlag:"-x"`
	ServiceThreadCore int           `yaml:"first_core" cmdLongFlag:"--firstcore,nonzero" cmdShortFlag:"-f,nonzero"`
	SystemName        string        `yaml:"name,omitempty" cmdLongFlag:"--group" cmdShortFlag:"-g"`
	SocketDir         string        `yaml:"socket_dir,omitempty" cmdLongFlag:"--socket_dir" cmdShortFlag:"-d"`
	LogMask           string        `yaml:"log_mask,omitempty" cmdEnv:"D_LOG_MASK"`
	LogFile           string        `yaml:"log_file,omitempty" cmdEnv:"D_LOG_FILE"`
	Storage           StorageConfig `yaml:",inline"`
	Fabric            FabricConfig  `yaml:",inline"`
	EnvVars           []string      `yaml:"env_vars,omitempty"`
	Index             uint32        `yaml:"-" cmdLongFlag:"--instance_idx" cmdShortFlag:"-I"`
}

// NewConfig returns an I/O Engine config.
func NewConfig() *Config {
	return &Config{
		HelperStreamCount: maxHelperStreamCount,
	}
}

// Validate ensures that the configuration meets minimum standards.
func (c *Config) Validate() error {
	if err := c.Fabric.Validate(); err != nil {
		return errors.Wrap(err, "fabric config validation failed")
	}

	if err := c.Storage.Validate(); err != nil {
		return errors.Wrap(err, "storage config validation failed")
	}

	return nil
}

// CmdLineArgs returns a slice of command line arguments to be
// supplied when starting an I/O Engine instance.
func (c *Config) CmdLineArgs() ([]string, error) {
	return parseCmdTags(c, shortFlagTag, joinShortArgs, nil)
}

// CmdLineEnv returns a slice of environment variables to be
// supplied when starting an I/O Engine instance.
func (c *Config) CmdLineEnv() ([]string, error) {
	tagEnv, err := parseCmdTags(c, envTag, joinEnvVars, nil)
	if err != nil {
		return nil, err
	}

	return mergeEnvVars(c.EnvVars, tagEnv), nil
}

// HasEnvVar returns true if the configuration contains
// an environment variable with the given name.
func (c *Config) HasEnvVar(name string) bool {
	for _, keyPair := range c.EnvVars {
		if strings.HasPrefix(keyPair, name+"=") {
			return true
		}
	}
	return false
}

// WithEnvVars applies the supplied list of environment
// variables to any existing variables, with new values
// overwriting existing values.
func (c *Config) WithEnvVars(newVars ...string) *Config {
	c.EnvVars = mergeEnvVars(c.EnvVars, newVars)

	return c
}

// WithRank sets the instance rank.
func (c *Config) WithRank(r uint32) *Config {
	c.Rank = system.NewRankPtr(r)
	return c
}

// WithSystemName sets the system name to which the instance belongs.
func (c *Config) WithSystemName(name string) *Config {
	c.SystemName = name
	return c
}

// WithHostname sets the hostname to be used when generating NVMe configurations.
func (c *Config) WithHostname(name string) *Config {
	c.Storage.Bdev.Hostname = name
	return c
}

// WithSocketDir sets the path to the instance's dRPC socket directory.
func (c *Config) WithSocketDir(dir string) *Config {
	c.SocketDir = dir
	return c
}

// WithScmClass defines the type of SCM storage to be configured.
func (c *Config) WithScmClass(scmClass string) *Config {
	c.Storage.SCM.Class = storage.ScmClass(scmClass)
	return c
}

// WithScmMountPath sets the path to the device used for SCM storage.
func (c *Config) WithScmMountPoint(scmPath string) *Config {
	c.Storage.SCM.MountPoint = scmPath
	return c
}

// WithScmRamdiskSize sets the size (in GB) of the ramdisk used
// to emulate SCM (no effect if ScmClass is not RAM).
func (c *Config) WithScmRamdiskSize(size int) *Config {
	c.Storage.SCM.RamdiskSize = size
	return c
}

// WithScmDeviceList sets the list of devices to be used for SCM storage.
func (c *Config) WithScmDeviceList(devices ...string) *Config {
	c.Storage.SCM.DeviceList = devices
	return c
}

// WithBdevClass defines the type of block device storage to be used.
func (c *Config) WithBdevClass(bdevClass string) *Config {
	c.Storage.Bdev.Class = storage.BdevClass(bdevClass)
	return c
}

// WithBdevDeviceList sets the list of block devices to be used.
func (c *Config) WithBdevDeviceList(devices ...string) *Config {
	c.Storage.Bdev.DeviceList = devices
	return c
}

// WithBdevDeviceCount sets the number of devices to be created when BdevClass is malloc.
func (c *Config) WithBdevDeviceCount(count int) *Config {
	c.Storage.Bdev.DeviceCount = count
	return c
}

// WithBdevFileSize sets the backing file size (used when BdevClass is malloc or file).
func (c *Config) WithBdevFileSize(size int) *Config {
	c.Storage.Bdev.FileSize = size
	return c
}

// WithBdevConfigPath sets the path to the generated NVMe config file used by SPDK.
func (c *Config) WithBdevConfigPath(cfgPath string) *Config {
	c.Storage.Bdev.ConfigPath = cfgPath
	return c
}

// WithModules sets the list of I/O Engine modules to be loaded.
func (c *Config) WithModules(mList string) *Config {
	c.Modules = mList
	return c
}

// WithFabricProvider sets the name of the CArT fabric provider.
func (c *Config) WithFabricProvider(provider string) *Config {
	c.Fabric.Provider = provider
	return c
}

// WithFabricInterface sets the interface name to be used by this instance.
func (c *Config) WithFabricInterface(iface string) *Config {
	c.Fabric.Interface = iface
	return c
}

// WithFabricInterfacePort sets the numeric interface port to be used by this instance.
func (c *Config) WithFabricInterfacePort(ifacePort int) *Config {
	c.Fabric.InterfacePort = ifacePort
	return c
}

// WithPinnedNumaNode sets the NUMA node affinity for the I/O Engine instance
func (c *Config) WithPinnedNumaNode(numa *uint) *Config {
	c.Fabric.PinnedNumaNode = numa
	return c
}

// WithCrtCtxShareAddr defines the CRT_CTX_SHARE_ADDR for this instance
func (c *Config) WithCrtCtxShareAddr(addr uint32) *Config {
	c.Fabric.CrtCtxShareAddr = addr
	return c
}

// WithCrtTimeout defines the CRT_TIMEOUT for this instance
func (c *Config) WithCrtTimeout(timeout uint32) *Config {
	c.Fabric.CrtTimeout = timeout
	return c
}

// WithTargetCount sets the number of VOS targets to run on this instance.
func (c *Config) WithTargetCount(count int) *Config {
	c.TargetCount = count
	return c
}

// WithHelperStreamCount sets the number of XS Helper streams to run on this instance.
func (c *Config) WithHelperStreamCount(count int) *Config {
	c.HelperStreamCount = count
	return c
}

// WithServiceThreadCore sets the core index to be used for running DAOS service threads.
func (c *Config) WithServiceThreadCore(idx int) *Config {
	c.ServiceThreadCore = idx
	return c
}

// WithLogFile sets the path to the log file to be used by this instance.
func (c *Config) WithLogFile(logPath string) *Config {
	c.LogFile = logPath
	return c
}

// WithLogMask sets the DAOS logging mask to be used by this instance.
func (c *Config) WithLogMask(logMask string) *Config {
	c.LogMask = logMask
	return c
}
