package api

import (
	"github.com/pkg/errors"
	"github.com/vitelabs/go-vite/chain"
	"github.com/vitelabs/go-vite/chain/trie_gc"
	"github.com/vitelabs/go-vite/common/types"
	"github.com/vitelabs/go-vite/generator"
	"github.com/vitelabs/go-vite/ledger"
	"github.com/vitelabs/go-vite/log15"
	"github.com/vitelabs/go-vite/vite"
	"strconv"
)

// !!! Block = Transaction = TX

func NewLedgerApi(vite *vite.Vite) *LedgerApi {
	api := &LedgerApi{
		chain: vite.Chain(),
		//signer:        vite.Signer(),
		log: log15.New("module", "rpc_api/ledger_api"),
	}

	return api
}

type GcStatus struct {
	Code        uint8  `json:"code"`
	Description string `json:"description"`

	ClearedHeight uint64 `json:"clearedHeight"`
	MarkedHeight  uint64 `json:"markedHeight"`
}

type LedgerApi struct {
	chain chain.Chain
	log   log15.Logger
}

func (l LedgerApi) String() string {
	return "LedgerApi"
}

func (l *LedgerApi) ledgerBlockToRpcBlock(block *ledger.AccountBlock) (*AccountBlock, error) {
	return ledgerToRpcBlock(block, l.chain)
}

func (l *LedgerApi) ledgerBlocksToRpcBlocks(list []*ledger.AccountBlock) ([]*AccountBlock, error) {
	var blocks []*AccountBlock
	for _, item := range list {
		rpcBlock, err := l.ledgerBlockToRpcBlock(item)
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, rpcBlock)
	}
	return blocks, nil
}

func (l *LedgerApi) GetBlockByHash(blockHash *types.Hash) (*AccountBlock, error) {
	block, getError := l.chain.GetAccountBlockByHash(blockHash)

	if getError != nil {
		l.log.Error("GetAccountBlockByHash failed, error is "+getError.Error(), "method", "GetBlockByHash")

		return nil, getError
	}
	if block == nil {
		return nil, nil
	}

	return l.ledgerBlockToRpcBlock(block)
}

func (l *LedgerApi) GetBlocksByHash(addr types.Address, originBlockHash *types.Hash, count uint64) ([]*AccountBlock, error) {
	l.log.Info("GetBlocksByHash")

	list, getError := l.chain.GetAccountBlocksByHash(addr, originBlockHash, count, false)
	if getError != nil {
		return nil, getError
	}

	if blocks, err := l.ledgerBlocksToRpcBlocks(list); err != nil {
		l.log.Error("GetConfirmTimes failed, error is "+err.Error(), "method", "GetBlocksByHash")
		return nil, err
	} else {
		return blocks, nil
	}

}

func (l *LedgerApi) GetBlocksByHashInToken(addr types.Address, originBlockHash *types.Hash, tokenTypeId types.TokenTypeId, count uint64) ([]*AccountBlock, error) {
	l.log.Info("GetBlocksByHashInToken")
	fti := l.chain.Fti()
	if fti == nil {
		err := errors.New("config.OpenFilterTokenIndex is false, api can't work")
		return nil, err
	}

	account, err := l.chain.GetAccount(&addr)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, nil
	}

	hashList, err := fti.GetBlockHashList(account, originBlockHash, tokenTypeId, count)
	if err != nil {
		return nil, err
	}

	blockList := make([]*ledger.AccountBlock, len(hashList))
	for index, blockHash := range hashList {
		block, err := l.chain.GetAccountBlockByHash(&blockHash)
		if err != nil {
			return nil, err
		}

		blockList[index] = block
	}
	return l.ledgerBlocksToRpcBlocks(blockList)
}

type Statistics struct {
	SnapshotBlockCount uint64 `json:"snapshotBlockCount"`
	AccountBlockCount  uint64 `json:"accountBlockCount"`
}

func (l *LedgerApi) GetStatistics() (*Statistics, error) {
	latestSnapshotBlock := l.chain.GetLatestSnapshotBlock()
	allLatestAccountBlock, err := l.chain.GetAllLatestAccountBlock()

	if err != nil {
		return nil, err
	}
	var accountBlockCount uint64
	for _, block := range allLatestAccountBlock {
		accountBlockCount += block.Height
	}

	return &Statistics{
		SnapshotBlockCount: latestSnapshotBlock.Height,
		AccountBlockCount:  accountBlockCount,
	}, nil
}

func (l *LedgerApi) GetVmLogListByHash(logHash types.Hash) (ledger.VmLogList, error) {
	logList, err := l.chain.GetVmLogList(&logHash)
	if err != nil {
		l.log.Error("GetVmLogList failed, error is "+err.Error(), "method", "GetVmLogListByHash")
		return nil, err
	}
	return logList, err
}

func (l *LedgerApi) GetBlocksByHeight(addr types.Address, height uint64, count uint64, forward bool) ([]*AccountBlock, error) {
	accountBlocks, err := l.chain.GetAccountBlocksByHeight(addr, height, count, forward)
	if err != nil {
		l.log.Error("GetAccountBlocksByHeight failed, error is "+err.Error(), "method", "GetBlocksByHeight")
		return nil, err
	}
	if len(accountBlocks) <= 0 {
		return nil, nil
	}
	return l.ledgerBlocksToRpcBlocks(accountBlocks)
}

func (l *LedgerApi) GetBlockByHeight(addr types.Address, heightStr string) (*AccountBlock, error) {
	height, err := strconv.ParseUint(heightStr, 10, 64)
	if err != nil {
		return nil, err
	}

	accountBlock, err := l.chain.GetAccountBlockByHeight(&addr, height)
	if err != nil {
		l.log.Error("GetAccountBlockByHeight failed, error is "+err.Error(), "method", "GetBlockByHeight")
		return nil, err
	}

	if accountBlock == nil {
		return nil, nil
	}
	return l.ledgerBlockToRpcBlock(accountBlock)
}

func (l *LedgerApi) GetBlocksByAccAddr(addr types.Address, index int, count int) ([]*AccountBlock, error) {
	l.log.Info("GetBlocksByAccAddr")

	list, getErr := l.chain.GetAccountBlocksByAddress(&addr, index, 1, count)

	if getErr != nil {
		l.log.Info("GetBlocksByAccAddr", "err", getErr)
		return nil, getErr
	}

	if blocks, err := l.ledgerBlocksToRpcBlocks(list); err != nil {
		l.log.Error("GetConfirmTimes failed, error is "+err.Error(), "method", "GetBlocksByAccAddr")
		return nil, err
	} else {
		return blocks, nil
	}
}

func (l *LedgerApi) GetAccountByAccAddr(addr types.Address) (*RpcAccountInfo, error) {
	l.log.Info("GetAccountByAccAddr")

	account, err := l.chain.GetAccount(&addr)
	if err != nil {
		l.log.Error("GetAccount failed, error is "+err.Error(), "method", "GetAccountByAccAddr")
		return nil, err
	}

	if account == nil {
		return nil, nil
	}

	latestAccountBlock, err := l.chain.GetLatestAccountBlock(&addr)
	if err != nil {
		l.log.Error("GetLatestAccountBlock failed, error is "+err.Error(), "method", "GetAccountByAccAddr")
		return nil, err
	}

	totalNum := uint64(0)
	if latestAccountBlock != nil {
		totalNum = latestAccountBlock.Height
	}

	balanceMap, err := l.chain.GetAccountBalance(&addr)
	if err != nil {
		l.log.Error("GetAccountBalance failed, error is "+err.Error(), "method", "GetAccountByAccAddr")
		return nil, err
	}

	tokenBalanceInfoMap := make(map[types.TokenTypeId]*RpcTokenBalanceInfo)
	for tokenId, amount := range balanceMap {
		token, _ := l.chain.GetTokenInfoById(&tokenId)
		tokenBalanceInfoMap[tokenId] = &RpcTokenBalanceInfo{
			TokenInfo:   RawTokenInfoToRpc(token, tokenId),
			TotalAmount: amount.String(),
			Number:      nil,
		}
	}

	rpcAccount := &RpcAccountInfo{
		AccountAddress:      account.AccountAddress,
		TotalNumber:         strconv.FormatUint(totalNum, 10),
		TokenBalanceInfoMap: tokenBalanceInfoMap,
	}

	return rpcAccount, nil
}

func (l *LedgerApi) GetSnapshotBlockByHash(hash types.Hash) (*ledger.SnapshotBlock, error) {
	block, err := l.chain.GetSnapshotBlockByHash(&hash)
	if err != nil {
		l.log.Error("GetSnapshotBlockByHash failed, error is "+err.Error(), "method", "GetSnapshotBlockByHash")
	}
	return block, err
}

func (l *LedgerApi) GetSnapshotBlockByHeight(height uint64) (*ledger.SnapshotBlock, error) {
	block, err := l.chain.GetSnapshotBlockByHeight(height)
	if err != nil {
		l.log.Error("GetSnapshotBlockByHash failed, error is "+err.Error(), "method", "GetSnapshotBlockByHeight")
	}
	return block, err
}

func (l *LedgerApi) GetSnapshotChainHeight() string {
	l.log.Info("GetLatestSnapshotChainHeight")
	return strconv.FormatUint(l.chain.GetLatestSnapshotBlock().Height, 10)
}

func (l *LedgerApi) GetLatestSnapshotChainHash() *types.Hash {
	l.log.Info("GetLatestSnapshotChainHash")
	return &l.chain.GetLatestSnapshotBlock().Hash
}

func (l *LedgerApi) GetLatestBlock(addr types.Address) (*AccountBlock, error) {
	l.log.Info("GetLatestBlock")
	block, getError := l.chain.GetLatestAccountBlock(&addr)
	if getError != nil {
		l.log.Error("GetLatestAccountBlock failed, error is "+getError.Error(), "method", "GetLatestBlock")
		return nil, getError
	}

	if block == nil {
		return nil, nil
	}

	return l.ledgerBlockToRpcBlock(block)
}

func (l *LedgerApi) GetTokenMintage(tti types.TokenTypeId) (*RpcTokenInfo, error) {
	l.log.Info("GetTokenMintage")
	if t, err := l.chain.GetTokenInfoById(&tti); err != nil {
		return nil, err
	} else {
		return RawTokenInfoToRpc(t, tti), nil
	}
}

func (l *LedgerApi) GetSenderInfo() (*KafkaSendInfo, error) {
	l.log.Info("GetSenderInfo")
	if l.chain.KafkaSender() == nil {
		return nil, nil
	}
	senderInfo := &KafkaSendInfo{}

	var totalErr error
	senderInfo.TotalEvent, totalErr = l.chain.GetLatestBlockEventId()
	if totalErr != nil {
		l.log.Error("GetLatestBlockEventId failed, error is "+totalErr.Error(), "method", "GetKafkaSenderInfo")

		return nil, totalErr
	}

	for _, producer := range l.chain.KafkaSender().Producers() {
		senderInfo.Producers = append(senderInfo.Producers, createKafkaProducerInfo(producer))
	}

	for _, producer := range l.chain.KafkaSender().RunProducers() {
		senderInfo.RunProducers = append(senderInfo.RunProducers, createKafkaProducerInfo(producer))
	}

	return senderInfo, nil
}

func (l *LedgerApi) GetBlockMeta(hash *types.Hash) (*ledger.AccountBlockMeta, error) {
	return l.chain.GetAccountBlockMetaByHash(hash)
}

func (l *LedgerApi) GetFittestSnapshotHash(accAddr *types.Address, sendBlockHash *types.Hash) (*types.Hash, error) {
	if accAddr == nil && sendBlockHash == nil {
		latestBlock := l.chain.GetLatestSnapshotBlock()
		if latestBlock != nil {
			return &latestBlock.Hash, nil
		}
		return nil, generator.ErrGetFittestSnapshotBlockFailed
	}
	var referredList []types.Hash
	if sendBlockHash != nil {
		sendBlock, _ := l.chain.GetAccountBlockByHash(sendBlockHash)
		if sendBlock == nil {
			return nil, generator.ErrGetSnapshotOfReferredBlockFailed
		}
		referredList = append(referredList, sendBlock.SnapshotHash)
	}

	prevHash, fittestHash, err := generator.GetFittestGeneratorSnapshotHash(l.chain, accAddr, referredList, false)
	if err != nil {
		return nil, err
	}
	if prevHash == nil {
		return fittestHash, nil
	}
	prevQuota, err := l.chain.GetPledgeQuota(*prevHash, *accAddr)
	if err != nil {
		return nil, err
	}
	fittestQuota, err := l.chain.GetPledgeQuota(*fittestHash, *accAddr)
	if err != nil {
		return nil, err
	}
	if prevQuota <= fittestQuota {
		return fittestHash, nil
	} else {
		ok, err := calculatedPoW(l.chain, accAddr, *prevHash)
		if err != nil {
			return nil, err
		}
		if ok {
			return fittestHash, nil
		}
		return prevHash, nil
	}

	//gap := uint64(0)
	//targetHeight := latestBlock.Height
	//
	//if targetHeight > gap {
	//	targetHeight = latestBlock.Height - gap
	//} else {
	//	targetHeight = 1
	//}
	//
	//targetSnapshotBlock, err := l.chain.GetSnapshotBlockByHeight(targetHeight)
	//if err != nil {
	//	return nil, err
	//}
	//return &targetSnapshotBlock.Hash, nil

}

func calculatedPoW(chain chain.Chain, addr *types.Address, snapshotHash types.Hash) (bool, error) {
	prevBlock, err := chain.GetLatestAccountBlock(addr)
	if err != nil {
		return false, err
	}
	for {
		if prevBlock != nil && prevBlock.SnapshotHash == snapshotHash {
			if isPoW(prevBlock.Nonce) {
				return true, nil
			}
			prevBlock, err = chain.GetAccountBlockByHash(&prevBlock.PrevHash)
			if err != nil {
				return false, err
			}
		} else {
			return false, nil
		}
	}
}

func (l *LedgerApi) GetNeedSnapshotContent() map[types.Address]*ledger.HashHeight {
	return l.chain.GetNeedSnapshotContent()
}

func (l *LedgerApi) SetSenderHasSend(producerId uint8, hasSend uint64) {
	l.log.Info("SetSenderHasSend")

	if l.chain.KafkaSender() == nil {
		return
	}
	l.chain.KafkaSender().SetHasSend(producerId, hasSend)
}

func (l *LedgerApi) StopSender(producerId uint8) {
	l.log.Info("StopSender")

	if l.chain.KafkaSender() == nil {
		return
	}
	l.chain.KafkaSender().StopById(producerId)
}

func (l *LedgerApi) AccountType(addr types.Address) (uint64, error) {
	return l.chain.AccountType(&addr)
}

func (l *LedgerApi) GetVmLogList(blockHash types.Hash) (ledger.VmLogList, error) {
	block, err := l.chain.GetAccountBlockByHash(&blockHash)
	if block == nil {
		if err != nil {
			return nil, err
		}
		return nil, errors.New("get block failed")
	}
	if block.LogHash == nil {
		code, err2 := l.chain.AccountType(&block.AccountAddress)
		if err2 != nil {
			return nil, err
		}
		if code == ledger.AccountTypeContract {
			return nil, errors.New("log hash can't be error")
		}
		return nil, nil
	}
	return l.chain.GetVmLogList(block.LogHash)
}

func (l *LedgerApi) GetGcStatus() *GcStatus {
	statusCode := l.chain.TrieGc().Status()

	gStatus := &GcStatus{
		Code: statusCode,
	}
	switch statusCode {
	case trie_gc.STATUS_STOPPED:
		gStatus.Description = "STATUS_STOPPED"
	case trie_gc.STATUS_STARTED:
		gStatus.Description = "STATUS_STARTED"
	case trie_gc.STATUS_MARKING_AND_CLEANING:
		gStatus.Description = "STATUS_MARKING_AND_CLEANING"
	}
	return gStatus
}
