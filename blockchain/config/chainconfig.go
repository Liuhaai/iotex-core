package chainconfig

import (
	"time"

	"github.com/iotexproject/go-pkgs/crypto"

	"github.com/iotexproject/iotex-core/db"
	"github.com/iotexproject/iotex-election/committee"
)

// Dardanelles consensus config
const (
	SigP256k1  = "secp256k1"
	SigP256sm2 = "p256sm2"
)

type (
	// Chain is the config struct for blockchain package
	Chain struct {
		ChainDBPath            string           `yaml:"chainDBPath"`
		TrieDBPath             string           `yaml:"trieDBPath"`
		IndexDBPath            string           `yaml:"indexDBPath"`
		BloomfilterIndexDBPath string           `yaml:"bloomfilterIndexDBPath"`
		CandidateIndexDBPath   string           `yaml:"candidateIndexDBPath"`
		StakingIndexDBPath     string           `yaml:"stakingIndexDBPath"`
		ID                     uint32           `yaml:"id"`
		EVMNetworkID           uint32           `yaml:"evmNetworkID"`
		Address                string           `yaml:"address"`
		ProducerPrivKey        string           `yaml:"producerPrivKey"`
		SignatureScheme        []string         `yaml:"signatureScheme"`
		EmptyGenesis           bool             `yaml:"emptyGenesis"`
		GravityChainDB         db.Config        `yaml:"gravityChainDB"`
		Committee              committee.Config `yaml:"committee"`

		EnableTrielessStateDB bool `yaml:"enableTrielessStateDB"`
		// EnableStateDBCaching enables cachedStateDBOption
		EnableStateDBCaching bool `yaml:"enableStateDBCaching"`
		// EnableArchiveMode is only meaningful when EnableTrielessStateDB is false
		EnableArchiveMode bool `yaml:"enableArchiveMode"`
		// EnableAsyncIndexWrite enables writing the block actions' and receipts' index asynchronously
		EnableAsyncIndexWrite bool `yaml:"enableAsyncIndexWrite"`
		// deprecated
		EnableSystemLogIndexer bool `yaml:"enableSystemLog"`
		// EnableStakingProtocol enables staking protocol
		EnableStakingProtocol bool `yaml:"enableStakingProtocol"`
		// EnableStakingIndexer enables staking indexer
		EnableStakingIndexer bool `yaml:"enableStakingIndexer"`
		// deprecated by DB.CompressBlock
		CompressBlock bool `yaml:"compressBlock"`
		// AllowedBlockGasResidue is the amount of gas remained when block producer could stop processing more actions
		AllowedBlockGasResidue uint64 `yaml:"allowedBlockGasResidue"`
		// MaxCacheSize is the max number of blocks that will be put into an LRU cache. 0 means disabled
		MaxCacheSize int `yaml:"maxCacheSize"`
		// PollInitialCandidatesInterval is the config for committee init db
		PollInitialCandidatesInterval time.Duration `yaml:"pollInitialCandidatesInterval"`
		// StateDBCacheSize is the max size of statedb LRU cache
		StateDBCacheSize int `yaml:"stateDBCacheSize"`
		// WorkingSetCacheSize is the max size of workingset cache in state factory
		WorkingSetCacheSize uint64 `yaml:"workingSetCacheSize"`
	}
)

var (
	Default = Chain{
		ChainDBPath:            "/var/data/chain.db",
		TrieDBPath:             "/var/data/trie.db",
		IndexDBPath:            "/var/data/index.db",
		BloomfilterIndexDBPath: "/var/data/bloomfilter.index.db",
		CandidateIndexDBPath:   "/var/data/candidate.index.db",
		StakingIndexDBPath:     "/var/data/staking.index.db",
		ID:                     1,
		EVMNetworkID:           4689,
		Address:                "",
		ProducerPrivKey:        generateRandomKey(SigP256k1),
		SignatureScheme:        []string{SigP256k1},
		EmptyGenesis:           false,
		GravityChainDB:         db.Config{DbPath: "/var/data/poll.db", NumRetries: 10},
		Committee: committee.Config{
			GravityChainAPIs: []string{},
		},
		EnableTrielessStateDB:         true,
		EnableStateDBCaching:          false,
		EnableArchiveMode:             false,
		EnableAsyncIndexWrite:         true,
		EnableSystemLogIndexer:        false,
		EnableStakingProtocol:         true,
		EnableStakingIndexer:          false,
		CompressBlock:                 false,
		AllowedBlockGasResidue:        10000,
		MaxCacheSize:                  0,
		PollInitialCandidatesInterval: 10 * time.Second,
		StateDBCacheSize:              1000,
		WorkingSetCacheSize:           20,
	}
)

func generateRandomKey(scheme string) string {
	// generate a random key
	switch scheme {
	case SigP256k1:
		sk, _ := crypto.GenerateKey()
		return sk.HexString()
	case SigP256sm2:
		sk, _ := crypto.GenerateKeySm2()
		return sk.HexString()
	}
	return ""
}
