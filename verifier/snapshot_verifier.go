package verifier

import (
	"fmt"
	"math/big"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/vitelabs/go-vite/chain"
	"github.com/vitelabs/go-vite/common/types"
	"github.com/vitelabs/go-vite/consensus"
	"github.com/vitelabs/go-vite/crypto"
	"github.com/vitelabs/go-vite/ledger"
	"github.com/vitelabs/go-vite/monitor"
)

type SnapshotVerifier struct {
	reader chain.Chain
	cs     consensus.Verifier
}

func NewSnapshotVerifier(ch chain.Chain, cs consensus.Verifier) *SnapshotVerifier {
	verifier := &SnapshotVerifier{reader: ch, cs: cs}
	return verifier
}

func (self *SnapshotVerifier) VerifyNetSb(block *ledger.SnapshotBlock) error {
	if err := self.verifyTimestamp(block); err != nil {
		return err
	}
	if err := self.verifyDataValidity(block); err != nil {
		return err
	}
	return nil
}

func (self *SnapshotVerifier) verifyTimestamp(block *ledger.SnapshotBlock) error {
	if block.Timestamp == nil {
		return errors.New("timestamp is nil")
	}

	if block.Timestamp.After(time.Now().Add(time.Hour)) {
		return errors.New("snapshot Timestamp not arrive yet")
	}
	return nil
}

func (self *SnapshotVerifier) verifyDataValidity(block *ledger.SnapshotBlock) error {
	computedHash := block.ComputeHash()
	if block.Hash.IsZero() || computedHash != block.Hash {
		return ErrVerifyHashFailed
	}

	if self.reader.IsGenesisSnapshotBlock(block) {
		return nil
	}

	if len(block.Signature) == 0 || len(block.PublicKey) == 0 {
		return errors.New("signature or publicKey is nil")
	}
	isVerified, _ := crypto.VerifySig(block.PublicKey, block.Hash.Bytes(), block.Signature)
	if !isVerified {
		return ErrVerifySignatureFailed
	}
	return nil
}

func (self *SnapshotVerifier) verifySelf(block *ledger.SnapshotBlock, stat *SnapshotBlockVerifyStat) error {
	defer monitor.LogTime("verify", "snapshotSelf", time.Now())

	if block.Height == types.GenesisHeight {
		snapshotBlock := self.reader.GetGenesisSnapshotBlock()
		if block.Hash != snapshotBlock.Hash {
			stat.result = FAIL
			return errors.New("genesis block error.")
		}
	}
	return nil
}

func (self *SnapshotVerifier) verifyAccounts(block *ledger.SnapshotBlock, prev *ledger.SnapshotBlock, stat *SnapshotBlockVerifyStat) error {
	defer monitor.LogTime("verify", "snapshotAccounts", time.Now())

	for addr, b := range block.SnapshotContent {
		hash, e := self.reader.GetAccountBlockHashByHeight(&addr, b.Height)
		if e != nil {
			return e
		}
		if hash == nil {
			stat.results[addr] = PENDING
		} else if *hash == b.Hash {
			stat.results[addr] = SUCCESS
		} else {
			stat.results[addr] = FAIL
			stat.result = FAIL
			return errors.New(fmt.Sprintf("account[%s] fork, height:[%d], hash:[%s]",
				addr.String(), b.Height, b.Hash))
		}
	}
	for _, v := range stat.results {
		if v != SUCCESS {
			return nil
		}
	}

	trie, err := self.reader.GenStateTrie(prev.StateHash, block.SnapshotContent)
	if err != nil {
		return err
	}
	if *trie.Hash() != block.StateHash {
		return errors.New("state hash is not equals.")
	}
	block.StateTrie = trie
	return nil
}

func (self *SnapshotVerifier) verifyAccountsTimeout(block *ledger.SnapshotBlock, stat *SnapshotBlockVerifyStat) error {
	defer monitor.LogTime("verify", "snapshotAccountsTimeout", time.Now())
	head := self.reader.GetLatestSnapshotBlock()
	if head.Height != block.Height-1 {
		return errors.New("snapshot pending for height:" + strconv.FormatUint(head.Height, 10))
	}
	if head.Hash != block.PrevHash {
		return errors.New(fmt.Sprintf("block is not next. prevHash:%s, headHash:%s", block.PrevHash, head.Hash))
	}

	for addr, hashH := range block.SnapshotContent {
		_, err := self.VerifyAccountTimeout(addr, hashH, block.Height)
		if err != nil {
			stat.result = FAIL
			return err
		}
	}
	return nil
}

func (self *SnapshotVerifier) VerifyAccountTimeout(addr types.Address, hashH *ledger.HashHeight, snapshotHeight uint64) (*ledger.HashHeight, error) {

	defer monitor.LogTime("verify", "accountTimeout", time.Now())

	first, e := self.reader.GetFirstConfirmedAccountBlockBySbHeight(snapshotHeight, &addr)
	if e != nil {
		return nil, e
	}

	if first == nil {
		if hashH != nil {
			return nil, errors.Errorf("account block[%s:%d:%s] is nil.", addr, hashH.Height, hashH.Hash)
		}
		return nil, errors.New("account block is nil.")
	}
	refer, e := self.reader.GetSnapshotBlockHeadByHash(&first.SnapshotHash)

	if e != nil {
		return nil, e
	}
	if refer == nil {
		return nil, errors.New("snapshot block is nil.")
	}

	ok := self.VerifyTimeout(snapshotHeight, refer.Height)
	if !ok {
		return &ledger.HashHeight{Height: first.Height, Hash: first.Hash}, errors.New("snapshot account block timeout.")
	}
	return nil, nil
}

func (self *SnapshotVerifier) VerifyTimeout(nowHeight uint64, referHeight uint64) bool {
	if nowHeight-referHeight > types.AccountLimitSnapshotHeight {
		return false
	}
	return true
}

func (self *SnapshotVerifier) VerifyReferred(block *ledger.SnapshotBlock) *SnapshotBlockVerifyStat {
	defer monitor.LogTime("verify", "snapshotBlock", time.Now())
	stat := self.newVerifyStat(block)
	// todo add state check
	err := self.verifySelf(block, stat)
	if err != nil {
		stat.errMsg = err.Error()
		return stat
	}

	head := self.reader.GetLatestSnapshotBlock()
	if !block.Timestamp.After(*head.Timestamp) {
		stat.result = FAIL
		stat.errMsg = "timestamp must be greater."
		return stat
	}

	// verify accounts exist
	err = self.verifyAccounts(block, head, stat)
	if err != nil {
		stat.errMsg = err.Error()
		return stat
	}
	for _, v := range stat.results {
		if v == FAIL || v == PENDING {
			return stat
		}
	}

	// verify accounts timeout
	err = self.verifyAccountsTimeout(block, stat)
	if err != nil {
		stat.errMsg = err.Error()
		return stat
	}

	if block.Height != types.GenesisHeight {
		// verify producer
		result, e := self.cs.VerifySnapshotProducer(block)
		if e != nil {
			stat.result = FAIL
			stat.errMsg = e.Error()
			return stat
		}
		if !result {
			stat.result = FAIL
			stat.errMsg = "verify snapshot producer fail."
			return stat
		}
	}
	stat.result = SUCCESS
	return stat
}

//func (self *SnapshotVerifier) VerifyProducer(block *ledger.SnapshotBlock) *SnapshotBlockVerifyStat {
//	defer monitor.LogTime("verify", "snapshotProducer", time.Now())
//	vStat := self.newVerifyStat(block)
//	return vStat
//}

type AccountHashH struct {
	Addr   *types.Address
	Hash   *types.Hash
	Height *big.Int
}

type SnapshotBlockVerifyStat struct {
	result       VerifyResult
	results      map[types.Address]VerifyResult
	errMsg       string
	accountTasks []*AccountPendingTask
	snapshotTask *SnapshotPendingTask
}

func (self *SnapshotBlockVerifyStat) ErrMsg() string {
	return self.errMsg
}

func (self *SnapshotBlockVerifyStat) VerifyResult() VerifyResult {
	return self.result
}

func (self *SnapshotBlockVerifyStat) Results() map[types.Address]VerifyResult {
	return self.results
}

func (self *SnapshotVerifier) newVerifyStat(b *ledger.SnapshotBlock) *SnapshotBlockVerifyStat {
	stat := &SnapshotBlockVerifyStat{result: PENDING}
	stat.results = make(map[types.Address]VerifyResult)
	for k := range b.SnapshotContent {
		stat.results[k] = PENDING
	}
	return stat
}
