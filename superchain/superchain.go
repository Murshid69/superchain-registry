package superchain

import (
	"embed"
	"fmt"
	"path"
	"strings"

	"gopkg.in/yaml.v3"
)

//go:embed configs
var superchainFS embed.FS

type BlockID struct {
	Hash   Hash   `yaml:"hash"`
	Number uint64 `yaml:"number"`
}

type ChainGenesis struct {
	L1        BlockID   `yaml:"l1"`
	L2        BlockID   `yaml:"l2"`
	L2Time    uint64    `yaml:"l2_time"`
	ExtraData *HexBytes `yaml:"extra_data,omitempty"`
}

type ChainConfig struct {
	Name         string `yaml:"name"`
	ChainID      uint64 `yaml:"chain_id"`
	PublicRPC    string `yaml:"public_rpc"`
	SequencerRPC string `yaml:"sequencer_rpc"`
	Explorer     string `yaml:"explorer"`

	SystemConfigAddr Address `yaml:"system_config_addr"`
	BatchInboxAddr   Address `yaml:"batch_inbox_addr"`

	Genesis ChainGenesis `yaml:"genesis"`

	// implied by directory structure, not encoded in the file itself
	Superchain string `yaml:"-"`
}

type SuperchainL1Info struct {
	ChainID   uint64 `yaml:"chain_id"`
	PublicRPC string `yaml:"public_rpc"`
	Explorer  string `yaml:"explorer"`
}

type SuperchainConfig struct {
	Name string           `yaml:"name"`
	L1   SuperchainL1Info `yaml:"l1"`

	// TODO: not available yet
	//ProtocolVersionAddr Address `yaml:"protocol_version_addr"`
}

type Superchain struct {
	Config   SuperchainConfig
	ChainIDs []uint64
}

var Superchains = map[string]*Superchain{}

var OPChains = map[uint64]*ChainConfig{}

func init() {
	superchainTargets, err := superchainFS.ReadDir("configs")
	if err != nil {
		panic(fmt.Errorf("failed to read superchain dir: %w", err))
	}
	// iterate over superchain-target entries
	for _, s := range superchainTargets {
		if !s.IsDir() {
			continue // ignore files, e.g. a readme
		}
		// Load superchain-target config
		superchainConfigData, err := superchainFS.ReadFile(path.Join("configs", s.Name(), "superchain.yaml"))
		if err != nil {
			panic(fmt.Errorf("failed to read superchain config: %w", err))
		}
		var superchainEntry Superchain
		if err := yaml.Unmarshal(superchainConfigData, &superchainEntry.Config); err != nil {
			panic(fmt.Errorf("failed to decode superchain config: %w", err))
		}

		// iterate over the chains of this superchain-target
		chainEntries, err := superchainFS.ReadDir(path.Join("configs", s.Name()))
		if err != nil {
			panic(fmt.Errorf("failed to read superchain dir: %w", err))
		}
		for _, c := range chainEntries {
			if c.IsDir() || !strings.HasSuffix(c.Name(), ".yaml") {
				continue // ignore files. Chains must be a directory of configs.
			}
			if c.Name() == "superchain.yaml" {
				continue // already processed
			}
			// load chain config
			chainConfigData, err := superchainFS.ReadFile(path.Join("configs", s.Name(), c.Name()))
			if err != nil {
				panic(fmt.Errorf("failed to read superchain config %s/%s: %w", s.Name(), c.Name(), err))
			}
			var chainConfig ChainConfig
			if err := yaml.Unmarshal(chainConfigData, &chainConfig); err != nil {
				panic(fmt.Errorf("failed to decode chain config %s/%s: %w", s.Name(), c.Name(), err))
			}
			chainConfig.Superchain = s.Name()
			if other, ok := OPChains[chainConfig.ChainID]; ok {
				panic(fmt.Errorf("found chain config %q in superchain target %q with chain ID %d "+
					"conflicts with chain %q in superchain %q and chain ID %d",
					chainConfig.Name, chainConfig.Superchain, chainConfig.ChainID,
					other.Name, other.Superchain, other.ChainID))
			}
			superchainEntry.ChainIDs = append(superchainEntry.ChainIDs, chainConfig.ChainID)
			OPChains[chainConfig.ChainID] = &chainConfig
		}
		Superchains[superchainEntry.Config.Name] = &superchainEntry
	}
}
