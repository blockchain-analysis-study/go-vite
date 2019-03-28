package chain

import (
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/vitelabs/go-vite/common/types"
	"github.com/vitelabs/go-vite/crypto/ed25519"
	"github.com/vitelabs/go-vite/ledger"
	"github.com/vitelabs/go-vite/monitor"
	"time"
)

// 0 means error, 1 means not exist, 2 means general account, 3 means contract account.
func (c *chain) AccountType(address *types.Address) (uint64, error) {
	monitorTags := []string{"chain", "AccountType"}
	defer monitor.LogTimerConsuming(monitorTags, time.Now())

	if types.IsPrecompiledContractAddress(*address) {
		return ledger.AccountTypeContract, nil
	}

	account, err := c.GetAccount(address)
	if err != nil {
		return ledger.AccountTypeError, err
	}

	if account == nil {
		return ledger.AccountTypeNotExist, nil
	}

	genesisBlock, err := c.chainDb.Ac.GetBlockByHeight(account.AccountId, 1)

	if err != nil {
		return ledger.AccountTypeError, err
	}

	if genesisBlock == nil {
		return ledger.AccountTypeNotExist, nil
	}

	fromBlock, getBlockErr := c.chainDb.Ac.GetBlock(&genesisBlock.FromBlockHash)
	if getBlockErr != nil {
		return ledger.AccountTypeError, getBlockErr
	}

	gid, getBlockErr := c.ChainDb().Ac.GetContractGidFromSendCreateBlock(fromBlock)
	if getBlockErr != nil {
		return ledger.AccountTypeError, getBlockErr
	}

	if gid != nil {
		return ledger.AccountTypeContract, nil
	}

	return ledger.AccountTypeGeneral, nil
}

// TODO cache
func (c *chain) GetAccount(address *types.Address) (*ledger.Account, error) {
	monitorTags := []string{"chain", "GetAccount"}
	defer monitor.LogTimerConsuming(monitorTags, time.Now())

	account, err := c.chainDb.Account.GetAccountByAddress(address)
	if err != nil {
		c.log.Error("Query account failed, error is "+err.Error(), "method", "GetAccount")
		return nil, err
	}
	return account, nil
}

func (c *chain) newAccountId() (uint64, error) {

	lastAccountId, err := c.chainDb.Account.GetLastAccountId()

	if err != nil {
		return 0, err
	}
	return lastAccountId + 1, nil
}

func (c *chain) createAccount(batch *leveldb.Batch, accountId uint64, address *types.Address, publicKey ed25519.PublicKey) (*ledger.Account, error) {
	account := &ledger.Account{
		AccountAddress: *address,
		AccountId:      accountId,
		PublicKey:      publicKey,
	}

	c.chainDb.Account.WriteAccountIndex(batch, accountId, address)
	if err := c.chainDb.Account.WriteAccount(batch, account); err != nil {
		return nil, err
	}
	return account, nil
}
