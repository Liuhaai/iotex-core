package cmd

import (
	"context"
	"strconv"
	"time"

	"github.com/iotexproject/iotex-core/action/protocol"
	"github.com/iotexproject/iotex-core/blockchain"
	"github.com/iotexproject/iotex-core/blockchain/block"
	"github.com/iotexproject/iotex-core/blockchain/blockdao"
	"github.com/iotexproject/iotex-core/blockchain/genesis"
	"github.com/iotexproject/iotex-core/config"
	"github.com/iotexproject/iotex-core/db"
	"github.com/iotexproject/iotex-core/pkg/log"
	"github.com/iotexproject/iotex-core/server/itx"
	"github.com/iotexproject/iotex-core/state/factory"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var (
	// CheckHeight used to Sub command.
	CheckHeight = &cobra.Command{
		Use:   "calibrate",
		Short: "play at height x",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			arg0, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				panic(err)
			}
			bc, sf, dao := startChain()
			if err := bc.Start(context.Background()); err != nil {
				panic(err)
			}
			defer func() {
				if err := bc.Stop(context.Background()); err != nil {
					panic(err)
				}
			}()

			return calibrateToHeight(arg0)
			// if err != nil {
			// 	fmt.Printf("Check db %s height err: %v\n", args[0], err)
			// 	return err
			// }
			// fmt.Printf("Check db %s height: %d.\n", args[0], height)
			// fmt.Println("hello world")
			// return nil
		},
	}
)

func calibrateToHeight(newHeight uint64) error {
	// cfg, err := config.New([]string{}, []string{})
	// if err != nil {
	// 	return uint64(0), fmt.Errorf("failed to new config: %v", err)
	// }

	cfg := db.Config{}
	cfg.DbPath = "./foo"
	// cfg.CompressLegacy = false
	blockDao := blockdao.NewBlockDAO(nil, cfg)

	// highest height in data file
	height, err := blockDao.Height()
	if err != nil {
		return err
	}
	if newHeight > height {
		panic("asd")
	}

	// Load height value.
	// ctx := context.Background()
	if err := blockDao.SyncIndexerToTarget(123); err != nil {
		return err
	}
	// if err := blockDao.Stop(ctx); err != nil {
	// 	return uint64(0), err
	// }

	return nil
	// *************************************************************************
	ctx := context.Background()
	bcCtx, ok := protocol.GetBlockchainCtx(ctx)
	if !ok {
		return errors.New("failed to find blockchain ctx")
	}
	g, ok := genesis.ExtractGenesisContext(ctx)
	if !ok {
		return errors.New("failed to find genesis ctx")
	}

	indexer := sf
	tipHeight, err := indexer.Height()
	if err != nil {
		return err
	}
	if tipHeight > dao.tipHeight {
		// TODO: delete block
		return errors.New("indexer tip height cannot by higher than dao tip height")
	}

	tipBlk, err := dao.GetBlockByHeight(tipHeight)
	if err != nil {
		return err
	}
	for i := tipHeight + 1; i <= dao.tipHeight; i++ {
		blk, err := dao.GetBlockByHeight(i)
		if err != nil {
			return err
		}
		if blk.Receipts == nil {
			blk.Receipts, err = dao.GetReceipts(i)
			if err != nil {
				return err
			}
		}
		producer := blk.PublicKey().Address()
		if producer == nil {
			return errors.New("failed to get address")
		}
		bcCtx.Tip.Height = tipBlk.Height()
		if bcCtx.Tip.Height > 0 {
			bcCtx.Tip.Hash = tipBlk.HashHeader()
			bcCtx.Tip.Timestamp = tipBlk.Timestamp()
		} else {
			bcCtx.Tip.Hash = g.Hash()
			bcCtx.Tip.Timestamp = time.Unix(g.Timestamp, 0)
		}
		for {
			if err = indexer.PutBlock(protocol.WithBlockCtx(
				protocol.WithBlockchainCtx(ctx, bcCtx),
				protocol.BlockCtx{
					BlockHeight:    i,
					BlockTimeStamp: blk.Timestamp(),
					Producer:       producer,
					GasLimit:       g.BlockGasLimit,
				},
			), blk); err == nil {
				break
			}
			if i < g.HawaiiBlockHeight && errors.Cause(err) == block.ErrDeltaStateMismatch {
				log.L().Info("delta state mismatch", zap.Uint64("block", i))
				continue
			}
			return err
		}
		if i%5000 == 0 {
			log.L().Info(
				"indexer is catching up.",
				zap.Uint64("height", i),
			)
		}
		tipBlk = blk
	}
	log.L().Info(
		"indexer is up to date.",
		zap.Uint64("height", tipHeight),
	)
	return nil
}

func startChain() (blockchain.Blockchain, factory.Factory, blockdao.BlockDAO) {
	// disable gateway mode
	genesisCfg, err := genesis.New(genesisPath)
	if err != nil {
		panic(err)
	}
	cfg, err := config.New([]string{_overwritePath, _secretPath}, _plugins)
	if err != nil {
		panic(err)
	}
	cfg.Genesis = genesisCfg

	// create server
	svr, err := itx.NewServer(cfg)
	if err != nil {
		log.L().Fatal("Failed to create server.", zap.Error(err))
	}

	// recover chain and state
	bc := svr.ChainService(cfg.Chain.ID).Blockchain()
	sf := svr.ChainService(cfg.Chain.ID).StateFactory()
	dao := svr.ChainService(cfg.Chain.ID).BlockDAO()
	return bc, sf, dao
}
