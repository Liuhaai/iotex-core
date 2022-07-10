package main

import (
	"context"
	"fmt"

	"github.com/iotexproject/iotex-core/action/protocol"
	"github.com/iotexproject/iotex-core/blockchain/block"
	"github.com/iotexproject/iotex-core/blockchain/blockdao"
	"github.com/iotexproject/iotex-core/blockchain/genesis"
	"github.com/iotexproject/iotex-core/blocksync"
	"github.com/iotexproject/iotex-core/chainservice"
	"github.com/iotexproject/iotex-core/config"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var cmdStatus = &cobra.Command{
	Use:   "status [genesis.yaml] [config.yaml] [trieDBPath]",
	Short: "display the current height of trie.db",
	Args:  cobra.RangeArgs(2, 3),
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		gs, err := genesis.New(args[0])
		if err != nil {
			return errors.Wrap(err, "failed to load genesis config")
		}
		cfg, err := config.New([]string{args[1]}, nil)
		if err != nil {
			return errors.Wrap(err, "failed to load config")
		}
		cfg.Genesis = gs
		if len(args) > 2 {
			cfg.Chain.TrieDBPath = args[2]
		}
		config.SetEVMNetworkID(cfg.Chain.EVMNetworkID)
		daoCfg := cfg.DB
		daoCfg.DbPath = cfg.Chain.ChainDBPath
		daoCfg.ReadOnly = true
		dao := blockdao.NewBlockDAO(nil, daoCfg, block.NewDeserializer(cfg.Chain.EVMNetworkID))
		builder := chainservice.NewBuilder(cfg).
			SetBlockDAO(dao).
			SetBlockSync(blocksync.NewDummyBlockSyncer())
		cs, err := builder.Build()
		if err != nil {
			return errors.Wrap(err, "failed to build chain service")
		}
		ctx := protocol.WithFeatureWithHeightCtx(
			genesis.WithGenesisContext(
				protocol.WithBlockchainCtx(
					context.Background(),
					protocol.BlockchainCtx{
						ChainID: cs.ChainID(),
					},
				),
				cfg.Genesis,
			),
		)
		if err := dao.Start(ctx); err != nil {
			return errors.Wrap(err, "failed to start block dao")
		}
		defer dao.Stop(ctx)
		blockDao := cs.BlockDAO()
		factory := cs.StateFactory()
		if err := factory.Start(ctx); err != nil {
			return errors.Wrap(err, "failed to start state factory")
		}
		defer factory.Stop(ctx)

		daoHeight, err := blockDao.Height()
		if err != nil {
			return err
		}
		indexerHeight, err := factory.Height()
		if err != nil {
			return err
		}
		fmt.Println("the height of trie.db is ", indexerHeight)
		fmt.Println("the height of chain.db is ", daoHeight)

		return nil
	},
}
