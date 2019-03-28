package generator

import (
	"fmt"
	"math/rand"
	"runtime/debug"
	"time"

	"github.com/pkg/errors"
	"github.com/vitelabs/go-vite/common/types"
	"github.com/vitelabs/go-vite/ledger"
	"github.com/vitelabs/go-vite/log15"
	"github.com/vitelabs/go-vite/vm"
	"github.com/vitelabs/go-vite/vm_context"
	"github.com/vitelabs/go-vite/vm_context/vmctxt_interface"
)

func GetFittestGeneratorSnapshotHash(chain vm_context.Chain, accAddr *types.Address,
	referredSnapshotHashList []types.Hash, isRandom bool) (prevSbHash *types.Hash, fittestSbHash *types.Hash, err error) {
	var fittestSbHeight uint64
	var referredMaxSbHeight uint64
	latestSb := chain.GetLatestSnapshotBlock()
	if latestSb == nil {
		return nil, nil, errors.New("get latest snapshotblock failed")
	}
	fittestSbHeight = latestSb.Height

	var prevSbFlag = false
	var prevSb *ledger.SnapshotBlock
	if accAddr != nil {
		prevAccountBlock, err := chain.GetLatestAccountBlock(accAddr)
		if err != nil {
			return nil, nil, err
		}
		if prevAccountBlock != nil {
			referredSnapshotHashList = append(referredSnapshotHashList, prevAccountBlock.SnapshotHash)
			prevSbFlag = true
		}
	}
	referredMaxSbHeight = uint64(1)
	if len(referredSnapshotHashList) > 0 {
		// get max referredSbHeight
		for k, v := range referredSnapshotHashList {
			vSb, _ := chain.GetSnapshotBlockByHash(&v)
			if vSb == nil {
				return nil, nil, ErrGetSnapshotOfReferredBlockFailed
			} else {
				if referredMaxSbHeight < vSb.Height {
					referredMaxSbHeight = vSb.Height
				}
				if k == len(referredSnapshotHashList)-1 && prevSbFlag {
					prevSb = vSb
				}
			}
		}
		if latestSb.Height < referredMaxSbHeight {
			return nil, nil, errors.New("the height of the snapshotblock referred can't be larger than the latest")
		}
	}
	gap2Referred := latestSb.Height - referredMaxSbHeight
	minGap, randomGap := measureGapToLatest(gap2Referred, DefaultHeightDifference)
	if isRandom {
		fittestSbHeight = latestSb.Height - randomGap
	} else {
		fittestSbHeight = latestSb.Height - minGap
	}

	// protect code
	if fittestSbHeight > latestSb.Height || fittestSbHeight < referredMaxSbHeight {
		fittestSbHeight = latestSb.Height
	}

	fittestSb, err := chain.GetSnapshotBlockByHeight(fittestSbHeight)
	if fittestSb == nil {
		if err != nil {
			return nil, nil, err
		}
		return nil, nil, ErrGetFittestSnapshotBlockFailed
	}
	fittestSbHash = &fittestSb.Hash

	if accAddr == nil || prevSb == nil || prevSb.Height < referredMaxSbHeight ||
		prevSb.Height+types.SnapshotHourHeight < fittestSbHeight || prevSb.Hash == *fittestSbHash {
		return nil, fittestSbHash, nil
	}
	return &prevSb.Hash, fittestSbHash, nil
}

func measureGapToLatest(referredGap, defaultGap uint64) (minGap, randomGap uint64) {
	var minConfirmedHeight uint64
	if referredGap <= defaultGap {
		minConfirmedHeight = 0
		return referredGap, minConfirmedHeight + getRandomHeight(referredGap)
	} else {
		minConfirmedHeight = defaultGap
		return defaultGap, minConfirmedHeight + getRandomHeight(referredGap-defaultGap)
	}
}

func getRandomHeight(gapHeight uint64) uint64 {
	randHeight := uint64(0)
	if gapHeight >= 1 {
		rand.Seed(time.Now().UnixNano())
		randHeight = uint64(rand.Intn(int(gapHeight%types.SnapshotHourHeight + 1)))
	}
	return randHeight
}

func RecoverVmContext(chain vm_context.Chain, block *ledger.AccountBlock) (vmContextList []vmctxt_interface.VmDatabase, resultErr error) {
	var tLog = log15.New("method", "RecoverVmContext")
	vmContext, err := vm_context.NewVmContext(chain, &block.SnapshotHash, &block.PrevHash, &block.AccountAddress)
	if err != nil {
		return nil, err
	}
	var sendBlock *ledger.AccountBlock = nil
	if block.IsReceiveBlock() {
		if sendBlock = vmContext.GetAccountBlockByHash(&block.FromBlockHash); sendBlock == nil {
			return nil, ErrGetVmContextValueFailed
		}
	}
	defer func() {
		if err := recover(); err != nil {
			// print stack
			debug.PrintStack()
			errDetail := fmt.Sprintf("block(addr:%v prevHash:%v sbHash:%v )", block.AccountAddress, block.PrevHash, block.SnapshotHash)
			if sendBlock != nil {
				errDetail += fmt.Sprintf("sendBlock(addr:%v hash:%v)", block.AccountAddress, block.Hash)
			}
			tLog.Error(fmt.Sprintf("generator_vm panic error %v", err), "detail", errDetail)
			resultErr = errors.New("generator_vm panic error")
		}
	}()

	newVm := *vm.NewVM()
	blockList, isRetry, err := newVm.Run(vmContext, block, sendBlock)

	tLog.Debug("vm result", fmt.Sprintf("len %v, isRetry %v, err %v", len(blockList), isRetry, err))

	for _, v := range blockList {
		vmContextList = append(vmContextList, v.VmContext)
	}

	return vmContextList, err
}
