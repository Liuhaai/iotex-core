package staking

import (
	"fmt"
	"github.com/iotexproject/iotex-core/ioctl/util"
	"math/big"
	"os"
	"sort"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/iotexproject/iotex-core/action/protocol"
	"github.com/iotexproject/iotex-core/blockchain/genesis"
	"github.com/iotexproject/iotex-core/pkg/log"
	"github.com/iotexproject/iotex-core/state"
)

// VoteReviser is used to recalculate candidate votes.
type VoteReviser struct {
	reviseHeights []uint64
	cache         map[uint64]CandidateList
	c             genesis.VoteWeightCalConsts
}

// NewVoteReviser creates a VoteReviser.
func NewVoteReviser(c genesis.VoteWeightCalConsts, reviseHeights ...uint64) *VoteReviser {
	return &VoteReviser{
		reviseHeights: reviseHeights,
		cache:         make(map[uint64]CandidateList),
		c:             c,
	}
}

// Revise recalculate candidate votes on preset revising height.
func (vr *VoteReviser) Revise(csm CandidateStateManager, height uint64) error {
	if !vr.isCacheExist(height) {
		cands, err := vr.calculateVoteWeight(csm)
		if err != nil {
			return err
		}
		vr.storeToCache(height, cands)
	}
	return vr.flush(height, csm)
}

func (vr *VoteReviser) Check(csm CandidateStateManager, height uint64) error {
	file, err := os.OpenFile("votes.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	s := fmt.Sprintf("======= height = %d\n", height)
	file.WriteString(s)
	revised, err := vr.calculateVoteWeight(csm)
	if err != nil {
		file.WriteString(err.Error() + "\n")
		return err
	}
	cands, _, err := getAllCandidates(csm)
	if err != nil {
		file.WriteString(err.Error() + "\n")
		return err
	}
	s = fmt.Sprintf("|| cand size: revise = %d, fromdb = %d\n", len(revised), len(cands))
	file.WriteString(s)
	if len(revised) != len(cands) {
		return ErrTypeAssertion
	}

	sort.Sort(cands)
	revise := make(map[string]*Candidate)
	for i := range revised {
		revise[revised[i].Owner.String()] = revised[i]
	}
	for i := range cands {
		owner := cands[i].Owner.String()
		s := fmt.Sprintf("|| ==> %s\n", owner)
		file.WriteString(s)
		candm, ok := revise[owner]
		if !ok {
			file.WriteString("owner does not exist in revised\n")
			return ErrTypeAssertion
		}

		s = fmt.Sprintf("|| votes: revise = %d, fromdb = %d\n",
			util.RauToString(candm.Votes, util.IotxDecimalNum), util.RauToString(cands[i].Votes, util.IotxDecimalNum))
		file.WriteString(s)
		s = fmt.Sprintf("|| stake: revise = %d, fromdb = %d\n",
			util.RauToString(candm.SelfStake, util.IotxDecimalNum), util.RauToString(cands[i].SelfStake, util.IotxDecimalNum))
		file.WriteString(s)
		if cands[i].Votes.Cmp(candm.Votes) != 0 || cands[i].SelfStake.Cmp(candm.SelfStake) != 0 {
			return ErrTypeAssertion
		}
	}
	return nil
}

func (vr *VoteReviser) storeToCache(height uint64, cands CandidateList) {
	vr.cache[height] = cands
}

func (vr *VoteReviser) isCacheExist(height uint64) bool {
	_, ok := vr.cache[height]
	return ok
}

// NeedRevise returns true if height needs revise
func (vr *VoteReviser) NeedRevise(height uint64) bool {
	for _, h := range vr.reviseHeights {
		if height == h {
			return true
		}
	}
	return false
}

func (vr *VoteReviser) calculateVoteWeight(sm protocol.StateManager) (CandidateList, error) {
	cands, _, err := getAllCandidates(sm)
	switch {
	case errors.Cause(err) == state.ErrStateNotExist:
	case err != nil:
		return nil, err
	}
	candm := make(map[string]*Candidate)
	for _, cand := range cands {
		candm[cand.Owner.String()] = cand.Clone()
		candm[cand.Owner.String()].Votes = new(big.Int)
		candm[cand.Owner.String()].SelfStake = new(big.Int)
	}
	buckets, _, err := getAllBuckets(sm)
	switch {
	case errors.Cause(err) == state.ErrStateNotExist:
	case err != nil:
		return nil, err
	}

	for _, bucket := range buckets {
		if bucket.isUnstaked() {
			continue
		}
		cand, ok := candm[bucket.Candidate.String()]
		if !ok {
			log.L().Error("invalid bucket candidate", zap.Uint64("bucket index", bucket.Index), zap.String("candidate", bucket.Candidate.String()))
			continue
		}

		if cand.SelfStakeBucketIdx == bucket.Index {
			cand.AddVote(calculateVoteWeight(vr.c, bucket, true))
			cand.SelfStake = bucket.StakedAmount
		} else {
			cand.AddVote(calculateVoteWeight(vr.c, bucket, false))
		}
	}

	cands = make(CandidateList, 0, len(candm))
	for _, cand := range candm {
		cands = append(cands, cand)
	}
	return cands, nil
}

func (vr *VoteReviser) flush(height uint64, csm CandidateStateManager) error {
	cands, ok := vr.cache[height]
	if !ok {
		return nil
	}
	sort.Sort(cands)
	for _, cand := range cands {
		if err := csm.Upsert(cand); err != nil {
			return err
		}
	}
	return nil
}
