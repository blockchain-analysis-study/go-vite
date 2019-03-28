package net

import (
	"github.com/vitelabs/go-vite/common/types"
	"github.com/vitelabs/go-vite/ledger"
)

type blockFeeder interface {
	BlockSubscriber
	blockNotifier
}

type blockNotifier interface {
	notifySnapshotBlock(block *ledger.SnapshotBlock, source types.BlockSource)
	notifyAccountBlock(block *ledger.AccountBlock, source types.BlockSource)
}

type blockFeed struct {
	aSubs     map[int]AccountblockCallback
	bSubs     map[int]SnapshotBlockCallback
	currentId int
}

func newBlockFeeder() blockFeeder {
	return &blockFeed{
		aSubs: make(map[int]AccountblockCallback),
		bSubs: make(map[int]SnapshotBlockCallback),
	}
}

func (bf *blockFeed) SubscribeAccountBlock(fn AccountblockCallback) (subId int) {
	bf.currentId++
	bf.aSubs[bf.currentId] = fn
	return bf.currentId
}

func (bf *blockFeed) UnsubscribeAccountBlock(subId int) {
	delete(bf.aSubs, subId)
}

func (bf *blockFeed) SubscribeSnapshotBlock(fn SnapshotBlockCallback) (subId int) {
	bf.currentId++
	bf.bSubs[bf.currentId] = fn
	return bf.currentId
}

func (bf *blockFeed) UnsubscribeSnapshotBlock(subId int) {
	delete(bf.aSubs, subId)
}

func (bf *blockFeed) notifySnapshotBlock(block *ledger.SnapshotBlock, source types.BlockSource) {
	for _, fn := range bf.bSubs {
		if fn != nil {
			fn(block, source)
		}
	}
}

func (bf *blockFeed) notifyAccountBlock(block *ledger.AccountBlock, source types.BlockSource) {
	for _, fn := range bf.aSubs {
		if fn != nil {
			fn(block.AccountAddress, block, source)
		}
	}
}
