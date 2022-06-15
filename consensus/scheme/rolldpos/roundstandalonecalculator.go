package rolldpos

import (
	"encoding/hex"
	"time"

	"github.com/iotexproject/iotex-core/action/protocol/rolldpos"
	"github.com/iotexproject/iotex-core/blockchain/block"
	"github.com/iotexproject/iotex-core/consensus/consensusfsm"
	"github.com/iotexproject/iotex-core/endorsement"
	"github.com/pkg/errors"
)

type (
	FooterValidator struct {
		manager *MiniEpochManager
	}

	MiniEpochManager struct {
		cfg consensusfsm.ConsensusConfig

		// proposer
		delegates           []string
		isTimeBasedRotation bool
		rp                  *rolldpos.Protocol
		chain               ChainManager
	}

	miniRoundCalculator struct {
		collections map[string]*blockEndorsementCollection
	}
)

func newMiniRoundCalculator() *miniRoundCalculator {
	return &miniRoundCalculator{
		collections: make(map[string]*blockEndorsementCollection),
	}
}

func (rc *miniRoundCalculator) addBlock(blk *block.Block) {
	blkHash := blk.HashBlock()
	encodedBlockHash := hex.EncodeToString(blkHash[:])
	if c, exists := rc.collections[encodedBlockHash]; exists {
		c.SetBlock(blk)
		return
	}
	rc.collections[encodedBlockHash] = newBlockEndorsementCollection(blk)
}

func (rc *miniRoundCalculator) addVoteEndorsement(vote *ConsensusVote, en *endorsement.Endorsement) error {
	if !endorsement.VerifyEndorsement(vote, en) {
		return errors.New("invalid endorsement for the vote")
	}
	blockHash := hex.EncodeToString(vote.BlockHash())
	c, exist := rc.collections[blockHash]
	if !exist {
		return errors.New("the corresponding block not received")
	}
	if err := c.AddEndorsement(vote.Topic(), en); err != nil {
		return err
	}
	rc.collections[blockHash] = c
	return nil
}

func (rc *miniRoundCalculator) countTopicEndorsement(blkHash []byte, topics []ConsensusVoteTopic) int {
	c, exist := rc.collections[hex.EncodeToString(blkHash)]
	if !exist {
		return 0
	}
	return len(c.Endorsements(topics))
}

func (manager *MiniEpochManager) IsDelegate(addr string) bool {
	for _, v := range manager.delegates {
		if addr == v {
			return true
		}
	}
	return false
}

/************************ Calc Proposer ************************/
// Proposer returns the block producer of the round
func (manager *MiniEpochManager) Proposer(height uint64, roundStartTime time.Time) (string, error) {
	idx := height
	if manager.isTimeBasedRotation {
		roundNum, err := manager.getRoundNum(height, manager.cfg.BlockInterval(height), roundStartTime, 0)
		if err != nil {
			return "", err
		}
		idx += uint64(roundNum)
	}
	return manager.delegates[idx%uint64(len(manager.delegates))], nil
}

func (rc *MiniEpochManager) getRoundNum(
	height uint64,
	blockInterval time.Duration,
	now time.Time,
	toleratedOvertime time.Duration,
) (roundNum uint32, err error) {
	lastBlockTime := time.Unix(rc.chain.Genesis().Timestamp, 0)
	if height > 1 {
		bc := rc.chain.Genesis().Blockchain
		if bc.IsBering(height) {
			var lastBlock *block.Header
			if lastBlock, err = rc.chain.BlockHeaderByHeight(height - 1); err != nil {
				return
			}
			lastBlockTime = lastBlockTime.Add(lastBlock.Timestamp().Sub(lastBlockTime) / blockInterval * blockInterval)
		} else {
			err = errors.New("light mode isn't supported before Bering hardfork")
			return
		}
	}
	if !lastBlockTime.Before(now) {
		// TODO: if this is the case, it is possible that the system time is far behind the time of other nodes.
		// better error handling may be needed on the caller side
		err = errors.Wrapf(
			errInvalidCurrentTime,
			"last block time %s is after than current time %s",
			lastBlockTime,
			now,
		)
		return
	}
	duration := now.Sub(lastBlockTime)
	if duration > blockInterval {
		roundNum = uint32(duration / blockInterval)
		if toleratedOvertime == 0 || duration%blockInterval < toleratedOvertime {
			roundNum--
		}
	}
	return
}

/************************ LastBlockInEpoch ************************/
func (manager *MiniEpochManager) UpdateWith(blk *block.Block) error {
	if !manager.isLastBlockInEpoch(blk.Height()) {
		return nil
	}
	delegates := blk.Header.DelegatesAddr()
	// if uint64(len(delegates)) != manager.rp.NumDelegates() {
	// 	return errors.New("invalid delegate list")
	// }
	manager.delegates = delegates
	return nil
}

func (manager *MiniEpochManager) isLastBlockInEpoch(height uint64) bool {
	return height == manager.rp.GetEpochLastBlockHeight(manager.rp.GetEpochNum(height))
}

func (fv *FooterValidator) ValidateBlockVote(blk *block.Block) error {
	round := newMiniRoundCalculator()

	round.addBlock(blk)

	blkHash := blk.HashBlock()
	for _, en := range blk.Endorsements() {
		if err := round.addVoteEndorsement(
			NewConsensusVote(blkHash[:], COMMIT),
			en,
		); err != nil {
			return err
		}
	}

	if 3*round.countTopicEndorsement(blkHash[:], []ConsensusVoteTopic{COMMIT}) <= 2*len(fv.manager.delegates) {
		return ErrInsufficientEndorsements
	}
	return nil
}

func (fv *FooterValidator) ValidateBlockProducer(blk *block.Block) error {
	blkProducer := blk.ProducerAddress()

	if !fv.manager.IsDelegate(blkProducer) {
		return errors.Errorf(
			"block proposer %s is not a valid delegate",
			blk.ProducerAddress(),
		)
	}

	proposer, err := fv.manager.Proposer(blk.Height(), blk.Timestamp())
	if err != nil {
		return err
	}

	if proposer != blkProducer {
		return errors.Errorf(
			"block proposer %s is invalid",
			blk.ProducerAddress(),
		)
	}

	return nil
}
