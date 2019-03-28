package access

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/util"
	"github.com/vitelabs/go-vite/chain_db/database"
	"github.com/vitelabs/go-vite/common/helper"
	"github.com/vitelabs/go-vite/common/types"
	"github.com/vitelabs/go-vite/ledger"
	vmutil "github.com/vitelabs/go-vite/vm/util"
)

func getAccountBlockHash(dbKey []byte) *types.Hash {
	hashBytes := dbKey[17:]
	hash, _ := types.BytesToHash(hashBytes)
	return &hash
}

func getAccountBlockHeight(dbKey []byte) uint64 {
	heightBytes := dbKey[9:17]
	return binary.BigEndian.Uint64(heightBytes)
}

type AccountChain struct {
	db *leveldb.DB
}

func NewAccountChain(db *leveldb.DB) *AccountChain {
	return &AccountChain{
		db: db,
	}
}

func (ac *AccountChain) DeleteBlock(batch *leveldb.Batch, accountId uint64, height uint64, hash *types.Hash) {
	key, _ := database.EncodeKey(database.DBKP_ACCOUNTBLOCK, accountId, height, hash.Bytes())
	batch.Delete(key)
}

func (ac *AccountChain) DeleteBlockMeta(batch *leveldb.Batch, hash *types.Hash) {
	key, _ := database.EncodeKey(database.DBKP_ACCOUNTBLOCKMETA, hash.Bytes())
	batch.Delete(key)
	// Delete be snapshot
	ac.DeleteBeSnapshot(batch, hash)
}

func (ac *AccountChain) WriteBlock(batch *leveldb.Batch, accountId uint64, block *ledger.AccountBlock) error {
	buf, err := block.DbSerialize()
	if err != nil {
		return err
	}

	key, err := database.EncodeKey(database.DBKP_ACCOUNTBLOCK, accountId, block.Height, block.Hash.Bytes())

	batch.Put(key, buf)
	return nil
}

func (ac *AccountChain) WriteBlockMeta(batch *leveldb.Batch, blockHash *types.Hash, blockMeta *ledger.AccountBlockMeta) error {
	buf, err := blockMeta.Serialize()
	if err != nil {
		return err
	}

	key, _ := database.EncodeKey(database.DBKP_ACCOUNTBLOCKMETA, blockHash.Bytes())

	batch.Put(key, buf)
	return nil
}

func (ac *AccountChain) WriteBeSnapshot(batch *leveldb.Batch, blockHash *types.Hash, snapshotBlockHeight uint64) error {
	key, _ := database.EncodeKey(database.DBKP_BE_SNAPSHOT, blockHash.Bytes())

	heightBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(heightBytes, snapshotBlockHeight)

	batch.Put(key, heightBytes)
	return nil
}

func (ac *AccountChain) DeleteBeSnapshot(batch *leveldb.Batch, blockHash *types.Hash) {
	key, _ := database.EncodeKey(database.DBKP_BE_SNAPSHOT, blockHash.Bytes())
	batch.Delete(key)
}

func (ac *AccountChain) GetBeSnapshot(blockHash *types.Hash) (uint64, error) {
	key, _ := database.EncodeKey(database.DBKP_BE_SNAPSHOT, blockHash.Bytes())

	value, err := ac.db.Get(key, nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return 0, nil
		}
		return 0, err
	}

	snapshotHeight := binary.BigEndian.Uint64(value)

	return snapshotHeight, nil
}

func (ac *AccountChain) GetHashByHeight(accountId uint64, height uint64) (*types.Hash, error) {
	key, _ := database.EncodeKey(database.DBKP_ACCOUNTBLOCK, accountId, height)
	iter := ac.db.NewIterator(util.BytesPrefix(key), nil)
	defer iter.Release()

	if !iter.Last() {
		if err := iter.Error(); err != nil && err != leveldb.ErrNotFound {
			return nil, err
		}

		return nil, nil
	}

	return getAccountBlockHash(iter.Key()), nil

}

func (ac *AccountChain) IsBlockExisted(hash types.Hash) (bool, error) {
	key, _ := database.EncodeKey(database.DBKP_ACCOUNTBLOCKMETA, hash.Bytes())
	return ac.db.Has(key, nil)
}

func (ac *AccountChain) GetLatestBlock(accountId uint64) (*ledger.AccountBlock, error) {
	key, err := database.EncodeKey(database.DBKP_ACCOUNTBLOCK, accountId)
	if err != nil {
		return nil, err
	}

	iter := ac.db.NewIterator(util.BytesPrefix(key), nil)
	defer iter.Release()

	if !iter.Last() {
		if err := iter.Error(); err != nil && err != leveldb.ErrNotFound {
			return nil, err
		}
		return nil, nil
	}
	block := &ledger.AccountBlock{}
	if ddsErr := block.DbDeserialize(iter.Value()); ddsErr != nil {
		return nil, ddsErr
	}

	block.Hash = *getAccountBlockHash(iter.Key())
	return block, nil
}

func (ac *AccountChain) GetBlockListByAccountId(accountId, startHeight, endHeight uint64, forward bool) ([]*ledger.AccountBlock, error) {
	startKey, _ := database.EncodeKey(database.DBKP_ACCOUNTBLOCK, accountId, startHeight)
	limitKey, _ := database.EncodeKey(database.DBKP_ACCOUNTBLOCK, accountId, endHeight+1)

	iter := ac.db.NewIterator(&util.Range{Start: startKey, Limit: limitKey}, nil)
	defer iter.Release()

	// cap
	listLength := uint64(0)
	if endHeight >= startHeight {
		listLength = endHeight - startHeight + 1
	} else {
		return nil, errors.New("endHeight is less than startHeight")
	}

	blockList := make([]*ledger.AccountBlock, listLength)

	i := uint64(0)
	for ; iter.Next(); i++ {
		block := &ledger.AccountBlock{}
		err := block.DbDeserialize(iter.Value())

		if err != nil {
			return nil, err
		}

		block.Hash = *getAccountBlockHash(iter.Key())
		if forward {
			blockList[i] = block
		} else {
			blockList[listLength-i-1] = block
		}
	}

	if err := iter.Error(); err != nil && err != leveldb.ErrNotFound {
		return nil, err
	}

	if i <= 0 {
		return nil, nil
	}

	if forward {
		return blockList[:i], nil
	} else {
		return blockList[listLength-i:], nil
	}
}

func (ac *AccountChain) GetBlock(blockHash *types.Hash) (*ledger.AccountBlock, error) {
	blockMeta, gbmErr := ac.GetBlockMeta(blockHash)
	if gbmErr != nil {
		return nil, gbmErr
	}
	if blockMeta == nil {
		return nil, nil
	}

	key, _ := database.EncodeKey(database.DBKP_ACCOUNTBLOCK, blockMeta.AccountId, blockMeta.Height, blockHash.Bytes())

	data, err := ac.db.Get(key, nil)

	if err != nil {
		if err != leveldb.ErrNotFound {
			return nil, err
		}
		return nil, nil
	}

	accountBlock := &ledger.AccountBlock{}
	if dsErr := accountBlock.DbDeserialize(data); dsErr != nil {
		return nil, dsErr
	}
	accountBlock.Hash = *blockHash
	accountBlock.Meta = blockMeta

	return accountBlock, nil
}

func (ac *AccountChain) GetBlockMeta(blockHash *types.Hash) (*ledger.AccountBlockMeta, error) {
	key, err := database.EncodeKey(database.DBKP_ACCOUNTBLOCKMETA, blockHash.Bytes())
	if err != nil {
		return nil, err
	}
	blockMetaBytes, err := ac.db.Get(key, nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return nil, nil
		}
		return nil, err
	}

	blockMeta := &ledger.AccountBlockMeta{}
	if err := blockMeta.Deserialize(blockMetaBytes); err != nil {
		return nil, err
	}

	beSnapshot, getBeSnapshotErr := ac.GetBeSnapshot(blockHash)
	if getBeSnapshotErr != nil {
		return nil, getBeSnapshotErr
	}

	blockMeta.SnapshotHeight = beSnapshot

	return blockMeta, nil
}

func (ac *AccountChain) GetVmLogList(logListHash *types.Hash) (ledger.VmLogList, error) {
	key, _ := database.EncodeKey(database.DBKP_LOG_LIST, logListHash.Bytes())
	data, err := ac.db.Get(key, nil)
	if err != nil {
		if err != leveldb.ErrNotFound {
			return nil, err
		}
		return nil, nil
	}

	vmLogList, dErr := ledger.VmLogListDeserialize(data)
	if dErr != nil {
		return nil, err
	}

	return vmLogList, err
}

func (ac *AccountChain) getConfirmHeight(accountBlockHash *types.Hash) (uint64, *ledger.AccountBlockMeta, error) {
	accountBlockMeta, err := ac.GetBlockMeta(accountBlockHash)
	if err != nil {
		return 0, nil, err
	}
	if accountBlockMeta.SnapshotHeight > 0 {
		return accountBlockMeta.SnapshotHeight, accountBlockMeta, nil
	}
	return 0, accountBlockMeta, nil
}

func (ac *AccountChain) GetConfirmHeight(accountBlockHash *types.Hash) (uint64, error) {

	confirmHeight, accountBlockMeta, err := ac.getConfirmHeight(accountBlockHash)
	if err != nil {
		return 0, err
	}

	if confirmHeight > 0 {
		return confirmHeight, nil
	}

	if accountBlockMeta == nil {
		return 0, nil
	}

	startKey, _ := database.EncodeKey(database.DBKP_ACCOUNTBLOCK, accountBlockMeta.AccountId, accountBlockMeta.Height+1)
	endKey, _ := database.EncodeKey(database.DBKP_ACCOUNTBLOCK, accountBlockMeta.AccountId, helper.MaxUint64)

	iter := ac.db.NewIterator(&util.Range{Start: startKey, Limit: endKey}, nil)
	defer iter.Release()

	for iter.Next() {
		blockHash := getAccountBlockHash(iter.Key())
		confirmHeight, _, err := ac.getConfirmHeight(blockHash)
		if err != nil {
			return 0, err
		}

		if confirmHeight > 0 {
			return confirmHeight, nil
		}
	}

	if err := iter.Error(); err != nil && err != leveldb.ErrNotFound {
		return 0, err
	}

	return 0, nil
}

func (ac *AccountChain) WriteVmLogList(batch *leveldb.Batch, logList ledger.VmLogList) error {
	key, _ := database.EncodeKey(database.DBKP_LOG_LIST, logList.Hash().Bytes())

	buf, err := logList.Serialize()
	if err != nil {
		return err
	}

	batch.Put(key, buf)

	return nil
}

func (ac *AccountChain) DeleteVmLogList(batch *leveldb.Batch, logListHash *types.Hash) {
	key, _ := database.EncodeKey(database.DBKP_LOG_LIST, logListHash.Bytes())
	batch.Delete(key)
}

func (ac *AccountChain) GetBlockByHeight(accountId uint64, height uint64) (*ledger.AccountBlock, error) {
	key, _ := database.EncodeKey(database.DBKP_ACCOUNTBLOCK, accountId, height)

	iter := ac.db.NewIterator(util.BytesPrefix(key), nil)
	if !iter.Last() {
		if err := iter.Error(); err != nil {
			return nil, err
		}
		return nil, nil
	}

	accountBlock := &ledger.AccountBlock{}
	if dsErr := accountBlock.DbDeserialize(iter.Value()); dsErr != nil {
		return nil, dsErr
	}

	accountBlock.Hash = *getAccountBlockHash(iter.Key())

	return accountBlock, nil
}

func (ac *AccountChain) GetContractGid(accountId uint64) (*types.Gid, error) {
	genesisBlock, err := ac.GetBlockByHeight(accountId, 1)

	if err != nil {
		return nil, err
	}

	if genesisBlock == nil {
		return nil, nil

	}

	fromBlock, getBlockErr := ac.GetBlock(&genesisBlock.FromBlockHash)
	if getBlockErr != nil {
		return nil, getBlockErr
	}

	return ac.GetContractGidFromSendCreateBlock(fromBlock)
}

func (ac *AccountChain) GetContractGidFromSendCreateBlock(fromBlock *ledger.AccountBlock) (*types.Gid, error) {
	if fromBlock == nil {
		return nil, nil
	}

	if fromBlock.BlockType != ledger.BlockTypeSendCreate {
		return nil, nil
	}

	gid := vmutil.GetGidFromCreateContractData(fromBlock.Data)
	return &gid, nil
}

func (ac *AccountChain) ReopenSendBlocks(batch *leveldb.Batch, reopenList []*ledger.HashHeight, deletedMap map[uint64]uint64) error {
	for _, reopenItem := range reopenList {
		blockMeta, err := ac.GetBlockMeta(&reopenItem.Hash)
		if err != nil {
			return err
		}
		if blockMeta == nil {
			continue
		}

		// The block will be deleted, don't need be write
		if deletedHeight := deletedMap[blockMeta.AccountId]; deletedHeight != 0 && blockMeta.Height >= deletedHeight {
			continue
		}

		newReceiveBlockHeights := blockMeta.ReceiveBlockHeights

		for index, receiveBlockHeight := range blockMeta.ReceiveBlockHeights {
			if receiveBlockHeight == reopenItem.Height {
				newReceiveBlockHeights = append(blockMeta.ReceiveBlockHeights[:index], blockMeta.ReceiveBlockHeights[index+1:]...)
				break
			}
		}
		blockMeta.ReceiveBlockHeights = newReceiveBlockHeights
		writeErr := ac.WriteBlockMeta(batch, &reopenItem.Hash, blockMeta)
		if writeErr != nil {
			return err
		}
	}
	return nil
}

func (ac *AccountChain) deleteChain(batch *leveldb.Batch, accountId uint64, toHeight uint64) ([]*ledger.AccountBlock, error) {
	deletedChain := make([]*ledger.AccountBlock, 0)

	startKey, _ := database.EncodeKey(database.DBKP_ACCOUNTBLOCK, accountId, toHeight)
	endKey, _ := database.EncodeKey(database.DBKP_ACCOUNTBLOCK, accountId, helper.MaxUint64)

	iter := ac.db.NewIterator(&util.Range{Start: startKey, Limit: endKey}, nil)
	defer iter.Release()

	for iter.Next() {

		deleteBlock := &ledger.AccountBlock{}
		if dsErr := deleteBlock.DbDeserialize(iter.Value()); dsErr != nil {
			return nil, dsErr
		}

		deleteBlock.Hash = *getAccountBlockHash(iter.Key())

		// Delete vm log list
		if deleteBlock.LogHash != nil {
			ac.DeleteVmLogList(batch, deleteBlock.LogHash)
		}

		// Delete block
		ac.DeleteBlock(batch, accountId, deleteBlock.Height, &deleteBlock.Hash)

		// Delete block meta
		ac.DeleteBlockMeta(batch, &deleteBlock.Hash)

		deletedChain = append(deletedChain, deleteBlock)
	}

	if err := iter.Error(); err != nil && err != leveldb.ErrNotFound {
		return nil, err
	}

	return deletedChain, nil
}

func (ac *AccountChain) Delete(batch *leveldb.Batch, deleteMap map[uint64]uint64) (map[uint64][]*ledger.AccountBlock, error) {
	deleted := make(map[uint64][]*ledger.AccountBlock)
	for accountId, deleteHeight := range deleteMap {
		deletedChain, err := ac.deleteChain(batch, accountId, deleteHeight)
		if err != nil {
			return nil, err
		}

		if len(deletedChain) > 0 {
			deleted[accountId] = deletedChain
		}
	}

	return deleted, nil
}

func (ac *AccountChain) GetDeleteMapAndReopenList(planToDelete map[uint64]uint64, getAccountByAddress func(*types.Address) (*ledger.Account, error), needExtendDelete, needNoSnapshot bool) (map[uint64]uint64, []*ledger.HashHeight, error) {
	currentNeedDelete := planToDelete

	deleteMap := make(map[uint64]uint64)
	var reopenList []*ledger.HashHeight

	for len(currentNeedDelete) > 0 {
		nextNeedDelete := make(map[uint64]uint64)

		for accountId, needDeleteHeight := range currentNeedDelete {
			endHeight := helper.MaxUint64
			if deleteHeight := deleteMap[accountId]; deleteHeight != 0 {
				if deleteHeight <= needDeleteHeight {
					continue
				}
				endHeight = deleteHeight

				// Pre set
				deleteMap[accountId] = needDeleteHeight
			} else {
				deleteMap[accountId] = needDeleteHeight
			}

			startKey, _ := database.EncodeKey(database.DBKP_ACCOUNTBLOCK, accountId, needDeleteHeight)
			endKey, _ := database.EncodeKey(database.DBKP_ACCOUNTBLOCK, accountId, endHeight)

			iter := ac.db.NewIterator(&util.Range{Start: startKey, Limit: endKey}, nil)

			for iter.Next() {
				accountBlock := &ledger.AccountBlock{}

				if dsErr := accountBlock.DbDeserialize(iter.Value()); dsErr != nil {
					iter.Release()
					return nil, nil, dsErr
				}

				blockHash := getAccountBlockHash(iter.Key())
				accountBlockMeta, getBmErr := ac.GetBlockMeta(blockHash)
				if getBmErr != nil {
					iter.Release()
					return nil, nil, getBmErr
				}

				if needNoSnapshot && accountBlockMeta.SnapshotHeight > 0 {
					return nil, nil, errors.New("is snapshot")
				}

				if needExtendDelete && accountBlock.IsSendBlock() {

					receiveAccount, getAccountErr := getAccountByAddress(&accountBlock.ToAddress)
					if getAccountErr != nil {
						iter.Release()
						return nil, nil, getAccountErr
					}

					if receiveAccount == nil {
						continue
					}
					receiveAccountId := receiveAccount.AccountId

					for _, receiveBlockHeight := range accountBlockMeta.ReceiveBlockHeights {
						if receiveBlockHeight > 0 {
							if currentDeleteHeight, nextDeleteHeight := currentNeedDelete[receiveAccountId], nextNeedDelete[receiveAccountId]; !(currentDeleteHeight != 0 && currentDeleteHeight <= receiveBlockHeight ||
								nextDeleteHeight != 0 && nextDeleteHeight <= receiveBlockHeight) {
								nextNeedDelete[receiveAccountId] = receiveBlockHeight
							}
						}
					}
				} else if accountBlock.IsReceiveBlock() {
					reopenList = append(reopenList, &ledger.HashHeight{
						Hash:   accountBlock.FromBlockHash,
						Height: accountBlock.Height,
					})
				}
			}

			if err := iter.Error(); err != nil && err != leveldb.ErrNotFound {
				iter.Release()
				return nil, nil, err
			}
			iter.Release()
		}

		currentNeedDelete = nextNeedDelete
	}

	return deleteMap, reopenList, nil
}

// TODO: cache
func (ac *AccountChain) GetPlanToDelete(maxAccountId uint64, snapshotBlockHeight uint64) (map[uint64]uint64, error) {
	planToDelete := make(map[uint64]uint64)

	for i := uint64(1); i <= maxAccountId; i++ {
		blockKey, _ := database.EncodeKey(database.DBKP_ACCOUNTBLOCK, i)

		iter := ac.db.NewIterator(util.BytesPrefix(blockKey), nil)
		iterOk := iter.Last()

		for iterOk {
			blockHash := getAccountBlockHash(iter.Key())
			blockMeta, getBmErr := ac.GetBlockMeta(blockHash)
			if getBmErr != nil {
				iter.Release()
				return nil, getBmErr
			}

			if blockMeta == nil {
				break
			}

			if blockMeta.RefSnapshotHeight >= snapshotBlockHeight {
				planToDelete[i] = blockMeta.Height
			} else {
				break
			}
			iterOk = iter.Prev()
		}

		if err := iter.Error(); err != nil && err != leveldb.ErrNotFound {
			iter.Release()
			return nil, err
		}
		iter.Release()
	}

	return planToDelete, nil
}
func (ac *AccountChain) GetUnConfirmedSubLedgerByAccounts(accountIds []uint64) (map[uint64][]*ledger.AccountBlock, error) {
	unConfirmedAccountBlocks := make(map[uint64][]*ledger.AccountBlock)
	for _, accountId := range accountIds {
		blockList, err := ac.GetUnConfirmAccountBlocks(accountId, 0)
		if err != nil {
			return nil, err
		}

		if len(blockList) > 0 {
			unConfirmedAccountBlocks[accountId] = blockList
		}
	}
	return unConfirmedAccountBlocks, nil
}

func (ac *AccountChain) GetUnConfirmedSubLedger(maxAccountId uint64) (map[uint64][]*ledger.AccountBlock, error) {
	unConfirmedAccountBlocks := make(map[uint64][]*ledger.AccountBlock)
	for accountId := uint64(1); accountId <= maxAccountId; accountId++ {
		blockList, err := ac.GetUnConfirmAccountBlocks(accountId, 0)
		if err != nil {
			return nil, err
		}

		if len(blockList) > 0 {
			unConfirmedAccountBlocks[accountId] = blockList
		}

	}
	return unConfirmedAccountBlocks, nil
}

func (ac *AccountChain) getUnconfirmedBlocks(accountId uint64) ([]*ledger.AccountBlock, error) {
	var unconfirmedBlocks []*ledger.AccountBlock
	block, err := ac.GetLatestBlock(accountId)
	if err != nil {
		return nil, err
	}

	if block == nil {
		return nil, nil
	}
	for {
		currentHeight := block.Height
		blockMeta, getMetaErr := ac.GetBlockMeta(&block.Hash)

		if getMetaErr != nil {
			return nil, getMetaErr
		}

		if blockMeta == nil {
			err := errors.New("blockMeta is nil, but block is not nil")
			return nil, err
		}

		if blockMeta.SnapshotHeight <= 0 {
			// prepend, less garbage
			unconfirmedBlocks = append(unconfirmedBlocks, nil)
			copy(unconfirmedBlocks[1:], unconfirmedBlocks)
			unconfirmedBlocks[0] = block
		} else {
			break
		}

		if currentHeight <= 0 {
			break
		}
		block, err := ac.GetBlockByHeight(accountId, currentHeight-1)
		if err != nil {
			return nil, err
		}
		if block == nil {
			break
		}
	}

	return unconfirmedBlocks, nil
}

// TODO Add cache, call frequently.
func (ac *AccountChain) GetConfirmAccountBlock(snapshotHeight uint64, accountId uint64) (*ledger.AccountBlock, error) {
	key, _ := database.EncodeKey(database.DBKP_ACCOUNTBLOCK, accountId)

	iter := ac.db.NewIterator(util.BytesPrefix(key), nil)
	defer iter.Release()

	iterOk := iter.Last()
	for iterOk {
		accountBlockHash := getAccountBlockHash(iter.Key())
		accountBlockMeta, getMetaErr := ac.GetBlockMeta(accountBlockHash)
		if getMetaErr != nil {
			return nil, getMetaErr
		}
		if accountBlockMeta == nil {
			return nil, errors.New(fmt.Sprintf("account block meta is nil, block hash is %s", accountBlockHash))
		}
		if accountBlockMeta.SnapshotHeight > 0 && accountBlockMeta.SnapshotHeight <= snapshotHeight {
			accountBlock := &ledger.AccountBlock{}
			if dsErr := accountBlock.DbDeserialize(iter.Value()); dsErr != nil {
				return nil, dsErr
			}

			accountBlock.Hash = *accountBlockHash
			accountBlock.Meta = accountBlockMeta

			return accountBlock, nil
		}
		iterOk = iter.Prev()
	}
	if err := iter.Error(); err != nil && err != leveldb.ErrNotFound {
		return nil, err
	}

	return nil, nil
}

func (ac *AccountChain) GetFirstConfirmedBlockBeforeOrAtAbHeight(accountId, accountBlockHeight uint64) (*ledger.AccountBlock, error) {
	startKey, _ := database.EncodeKey(database.DBKP_ACCOUNTBLOCK, accountId, 1)
	endKey, _ := database.EncodeKey(database.DBKP_ACCOUNTBLOCK, accountId, accountBlockHeight+1)

	iter := ac.db.NewIterator(&util.Range{Start: startKey, Limit: endKey}, nil)
	defer iter.Release()

	iterOk := iter.Last()

	var accountBlock *ledger.AccountBlock
	for iterOk {
		tmpAccountBlockHash := getAccountBlockHash(iter.Key())
		tmpAccountBlockMeta, getMetaErr := ac.GetBlockMeta(tmpAccountBlockHash)
		if getMetaErr != nil {
			return nil, getMetaErr
		}

		tmpAccountBlock, err := ac.GetBlock(tmpAccountBlockHash)
		if err != nil {
			return nil, err
		}

		tmpAccountBlock.Hash = *tmpAccountBlockHash
		tmpAccountBlock.Meta = tmpAccountBlockMeta

		if tmpAccountBlock.Height != accountBlockHeight && tmpAccountBlock.Meta.SnapshotHeight > 0 {
			break
		}

		accountBlock = tmpAccountBlock
		iterOk = iter.Prev()
	}
	if err := iter.Error(); err != nil && err != leveldb.ErrNotFound {
		return nil, err
	}

	return accountBlock, nil
}

func (ac *AccountChain) GetUnConfirmAccountBlocks(accountId uint64, beforeHeight uint64) ([]*ledger.AccountBlock, error) {
	accountBlocks := make([]*ledger.AccountBlock, 0)

	var iter iterator.Iterator
	if beforeHeight > 0 {
		startKey, _ := database.EncodeKey(database.DBKP_ACCOUNTBLOCK, accountId, uint64(1))
		endKey, _ := database.EncodeKey(database.DBKP_ACCOUNTBLOCK, accountId, beforeHeight)
		iter = ac.db.NewIterator(&util.Range{Start: startKey, Limit: endKey}, nil)
	} else {
		key, _ := database.EncodeKey(database.DBKP_ACCOUNTBLOCK, accountId)
		iter = ac.db.NewIterator(util.BytesPrefix(key), nil)
	}

	defer iter.Release()
	iterOk := iter.Last()

	for iterOk {
		accountBlockHash := getAccountBlockHash(iter.Key())
		accountBlockMeta, getMetaErr := ac.GetBlockMeta(accountBlockHash)
		if getMetaErr != nil {
			return nil, getMetaErr
		}
		if accountBlockMeta.SnapshotHeight <= 0 {
			accountBlock := &ledger.AccountBlock{}
			if dsErr := accountBlock.DbDeserialize(iter.Value()); dsErr != nil {
				return nil, dsErr
			}

			accountBlock.Hash = *accountBlockHash
			accountBlock.Meta = accountBlockMeta

			// prepend
			accountBlocks = append(accountBlocks, nil)
			copy(accountBlocks[1:], accountBlocks)
			accountBlocks[0] = accountBlock
		} else {
			return accountBlocks, nil
		}

		iterOk = iter.Prev()
	}

	if err := iter.Error(); err != nil && err != leveldb.ErrNotFound {
		return nil, err
	}

	return accountBlocks, nil
}

func (ac *AccountChain) GetSendAndReceiveBlocks(accountId uint64, snapshotBlockHeight uint64) ([]*ledger.AccountBlock, []*ledger.AccountBlock, error) {
	sendBlocks := make([]*ledger.AccountBlock, 0)
	receiveBlocks := make([]*ledger.AccountBlock, 0)

	startKey, _ := database.EncodeKey(database.DBKP_ACCOUNTBLOCK, accountId, uint64(1))
	endKey, _ := database.EncodeKey(database.DBKP_ACCOUNTBLOCK, accountId, helper.MaxUint64)
	iter := ac.db.NewIterator(&util.Range{Start: startKey, Limit: endKey}, nil)
	defer iter.Release()

	lastSnapshotBlockHeight := uint64(0)
	sendBlockList := make([]*ledger.AccountBlock, 0)
	receiveBlockList := make([]*ledger.AccountBlock, 0)

	for iter.Next() {
		accountBlock := &ledger.AccountBlock{}
		if dsErr := accountBlock.DbDeserialize(iter.Value()); dsErr != nil {
			return nil, nil, dsErr
		}

		accountBlockHash := getAccountBlockHash(iter.Key())
		accountBlockMeta, getMetaErr := ac.GetBlockMeta(accountBlockHash)
		if getMetaErr != nil {
			return nil, nil, getMetaErr
		}

		if accountBlockMeta.SnapshotHeight > snapshotBlockHeight {
			break
		}

		accountBlock.Meta = accountBlockMeta
		accountBlock.Hash = *accountBlockHash

		if accountBlock.IsSendBlock() {
			sendBlockList = append(sendBlockList, accountBlock)
		} else {
			receiveBlockList = append(receiveBlockList, accountBlock)
		}

		if accountBlockMeta.SnapshotHeight > lastSnapshotBlockHeight {

			sendBlocks = append(sendBlocks, sendBlockList...)
			sendBlockList = make([]*ledger.AccountBlock, 0)

			receiveBlocks = append(receiveBlocks, receiveBlockList...)
			receiveBlockList = make([]*ledger.AccountBlock, 0)
			lastSnapshotBlockHeight = accountBlockMeta.SnapshotHeight
		}
	}

	if err := iter.Error(); err != nil && err != leveldb.ErrNotFound {
		return nil, nil, err
	}

	return sendBlocks, receiveBlocks, nil
}

func (ac *AccountChain) GetSendBlocks(accountId uint64, snapshotBlockHeight uint64) ([]*ledger.AccountBlock, error) {
	blocks := make([]*ledger.AccountBlock, 0)
	startKey, _ := database.EncodeKey(database.DBKP_ACCOUNTBLOCK, accountId, 1)
	endKey, _ := database.EncodeKey(database.DBKP_ACCOUNTBLOCK, accountId, helper.MaxUint64)
	iter := ac.db.NewIterator(&util.Range{Start: startKey, Limit: endKey}, nil)
	defer iter.Release()

	lastSnapshotBlockHeight := uint64(0)
	sendBlockList := make([]*ledger.AccountBlock, 0)

	for iter.Next() {
		accountBlock := &ledger.AccountBlock{}
		if dsErr := accountBlock.DbDeserialize(iter.Value()); dsErr != nil {
			return nil, dsErr
		}

		accountBlockHash := getAccountBlockHash(iter.Key())
		accountBlockMeta, getMetaErr := ac.GetBlockMeta(accountBlockHash)
		if getMetaErr != nil {
			return nil, getMetaErr
		}

		if accountBlockMeta.SnapshotHeight > snapshotBlockHeight {
			break
		}

		if accountBlock.IsSendBlock() {
			accountBlock.Meta = accountBlockMeta
			accountBlock.Hash = *getAccountBlockHash(iter.Key())

			sendBlockList = append(sendBlockList, accountBlock)
		}

		if accountBlockMeta.SnapshotHeight > lastSnapshotBlockHeight {
			blocks = append(blocks, sendBlockList...)
			sendBlockList = make([]*ledger.AccountBlock, 0)
			lastSnapshotBlockHeight = accountBlockMeta.SnapshotHeight
		}
	}
	if err := iter.Error(); err != nil && err != leveldb.ErrNotFound {
		return nil, err
	}

	return blocks, nil
}
