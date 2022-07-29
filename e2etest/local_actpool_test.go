// Copyright (c) 2019 IoTeX Foundation
// This is an alpha (internal) release and is not suitable for production. This source code is provided 'as is' and no
// warranties are given as to title or non-infringement, merchantability or fitness for purpose and, to the extent
// permitted by law, all liability for your use of the code is disclaimed. This source code is governed by Apache
// License 2.0 that can be found in the LICENSE file.

package e2etest

import (
	"context"
	"math/big"
	"math/rand"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"

	"github.com/iotexproject/go-pkgs/crypto"
	"github.com/iotexproject/iotex-proto/golang/iotextypes"

	"github.com/iotexproject/iotex-core/action"
	"github.com/iotexproject/iotex-core/config"
	"github.com/iotexproject/iotex-core/p2p"
	"github.com/iotexproject/iotex-core/pkg/log"
	"github.com/iotexproject/iotex-core/server/itx"
	"github.com/iotexproject/iotex-core/test/identityset"
	"github.com/iotexproject/iotex-core/testutil"
)

func TestLocalActPool(t *testing.T) {
	require := require.New(t)
	ctx := context.Background()

	cfg := newActPoolConfig(require)
	svr, cli, err := createServerAndClient(require, cfg)
	require.NoError(err)
	chainID := cfg.Chain.ID
	defer func() {
		require.NoError(cli.Stop(ctx))
		require.NoError(svr.Stop(ctx))
	}()

	// Wait until server receives the 1st action
	tsf1, err := action.SignedTransfer(identityset.Address(0).String(), identityset.PrivateKey(1), 1, big.NewInt(1), []byte{}, uint64(100000), big.NewInt(0))
	require.NoError(err)
	require.NoError(testutil.WaitUntil(100*time.Millisecond, 10*time.Second, func() (bool, error) {
		require.NoError(cli.BroadcastOutbound(ctx, tsf1.Proto()))
		acts := svr.ChainService(chainID).ActionPool().PendingActionMap()
		return lenPendingActionMap(acts) == 1, nil
	}))

	tsf2, err := action.SignedTransfer(identityset.Address(1).String(), identityset.PrivateKey(1), 2, big.NewInt(3), []byte{}, uint64(100000), big.NewInt(0))
	require.NoError(err)
	tsf3, err := action.SignedTransfer(identityset.Address(0).String(), identityset.PrivateKey(1), 3, big.NewInt(3), []byte{}, uint64(100000), big.NewInt(0))
	require.NoError(err)
	// Create contract
	exec4, err := action.SignedExecution(action.EmptyAddress, identityset.PrivateKey(1), 4, big.NewInt(0), uint64(120000), big.NewInt(10), []byte{})
	require.NoError(err)
	// Create one invalid action
	tsf5, err := action.SignedTransfer(identityset.Address(0).String(), identityset.PrivateKey(1), 2, big.NewInt(3), []byte{}, uint64(100000), big.NewInt(0))
	require.NoError(err)

	require.NoError(cli.BroadcastOutbound(ctx, tsf2.Proto()))
	require.NoError(cli.BroadcastOutbound(ctx, tsf3.Proto()))
	require.NoError(cli.BroadcastOutbound(ctx, exec4.Proto()))
	require.NoError(cli.BroadcastOutbound(ctx, tsf5.Proto()))

	// Wait until server receives all the transfers
	require.NoError(testutil.WaitUntil(100*time.Millisecond, 10*time.Second, func() (bool, error) {
		acts := svr.ChainService(chainID).ActionPool().PendingActionMap()
		// 3 valid transfers and 1 valid execution in total
		return lenPendingActionMap(acts) == 4, nil
	}))
}

func BenchmarkActpool(b *testing.B) {
	require := require.New(b)
	ctx := context.Background()

	cfg := newActPoolConfig(require)
	svr, cli, err := createServerAndClient(require, cfg)
	require.NoError(err)
	chainID := cfg.Chain.ID
	defer func() {
		require.NoError(cli.Stop(ctx))
		require.NoError(svr.Stop(ctx))
	}()

	// Wait until server receives the 1st action
	tsf1, err := action.SignedTransfer(identityset.Address(0).String(), identityset.PrivateKey(1), 1, big.NewInt(1), []byte{}, uint64(100000), big.NewInt(0))
	require.NoError(err)
	require.NoError(testutil.WaitUntil(100*time.Millisecond, 10*time.Second, func() (bool, error) {
		require.NoError(cli.BroadcastOutbound(ctx, tsf1.Proto()))
		acts := svr.ChainService(chainID).ActionPool().PendingActionMap()
		return lenPendingActionMap(acts) == 1, nil
	}))

	// Generate msgs to be sent
	var (
		totalActs = 2
		nonceMap  = make(map[string]int)
		msgs      = make([]*iotextypes.Action, 0, totalActs)
	)
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < totalActs; i++ {
		senderIdx := rand.Intn(identityset.Size())
		receiverIdx := rand.Intn(identityset.Size())
		sender := identityset.Address(senderIdx)
		nonceMap[sender.Hex()]++
		tsf, err := action.SignedTransfer(identityset.Address(receiverIdx).String(), identityset.PrivateKey(senderIdx), uint64(nonceMap[sender.Hex()]), big.NewInt(0), []byte{}, uint64(10000), big.NewInt(0))
		require.NoError(err)
		msgs = append(msgs, tsf.Proto())
	}

	for i := range msgs {
		cli.BroadcastOutbound(ctx, msgs[i])
	}

	require.NoError(testutil.WaitUntil(100*time.Millisecond, 3*time.Second, func() (bool, error) {
		acts := svr.ChainService(chainID).ActionPool().PendingActionMap()

		res := lenPendingActionMap(acts)
		log.L().Info("asd", zap.Int("len", res))
		return res == 3, nil
	}))

}

func createServerAndClient(require *require.Assertions, cfg config.Config) (*itx.Server, p2p.Agent, error) {
	// create server
	ctx := context.Background()
	svr, err := itx.NewServer(cfg)
	require.NoError(err)
	chainID := cfg.Chain.ID
	require.NoError(svr.Start(ctx))
	require.NotNil(svr.ChainService(chainID).ActionPool())

	// create client
	cfg = newActPoolConfig(require)
	addrs, err := svr.P2PAgent().Self()
	require.NoError(err)
	cfg.Network.BootstrapNodes = []string{validNetworkAddr(addrs)}
	cli := p2p.NewAgent(
		cfg.Network,
		cfg.Chain.ID,
		cfg.Genesis.Hash(),
		func(_ context.Context, _ uint32, _ string, _ proto.Message) {},
		func(_ context.Context, _ uint32, _ peer.AddrInfo, _ proto.Message) {},
	)
	require.NoError(cli.Start(ctx))

	return svr, cli, nil
}

func newActPoolConfig(require *require.Assertions) config.Config {
	cfg := config.Default

	testTriePath, err := testutil.PathOfTempFile("trie")
	require.NoError(err)
	testDBPath, err := testutil.PathOfTempFile("db")
	require.NoError(err)
	testIndexPath, err := testutil.PathOfTempFile("index")
	require.NoError(err)

	defer func() {
		testutil.CleanupPath(testTriePath)
		testutil.CleanupPath(testDBPath)
		testutil.CleanupPath(testIndexPath)
	}()

	cfg.Chain.TrieDBPatchFile = ""
	cfg.Chain.TrieDBPath = testTriePath
	cfg.Chain.ChainDBPath = testDBPath
	cfg.Chain.IndexDBPath = testIndexPath
	cfg.ActPool.MinGasPriceStr = "0"
	cfg.Consensus.Scheme = config.NOOPScheme
	cfg.Network.Port = testutil.RandomPort()

	sk, err := crypto.GenerateKey()
	require.NoError(err)
	cfg.Chain.ProducerPrivKey = sk.HexString()
	return cfg
}
