package api

import (
	"errors"
	"github.com/vitelabs/go-vite/chain"
	"github.com/vitelabs/go-vite/common/helper"
	"github.com/vitelabs/go-vite/common/types"
	"github.com/vitelabs/go-vite/crypto/ed25519"
	"github.com/vitelabs/go-vite/generator"
	"github.com/vitelabs/go-vite/ledger"
	"github.com/vitelabs/go-vite/verifier"
	"github.com/vitelabs/go-vite/vite"
	"github.com/vitelabs/go-vite/vm"
	"github.com/vitelabs/go-vite/vm/contracts/abi"
	"github.com/vitelabs/go-vite/vm/quota"
	"github.com/vitelabs/go-vite/vm/util"
	"github.com/vitelabs/go-vite/vm_context"
	"math/big"
	"time"
)

type Tx struct {
	vite *vite.Vite
}

func NewTxApi(vite *vite.Vite) *Tx {
	return &Tx{
		vite: vite,
	}
}

func (t Tx) SendRawTx(block *AccountBlock) error {
	log.Info("SendRawTx")
	if block == nil {
		return errors.New("empty block")
	}

	lb, err := block.LedgerAccountBlock()
	if err != nil {
		return err
	}
	// need to remove Later
	//if len(lb.Data) != 0 && !isPreCompiledContracts(lb.ToAddress) {
	//	return ErrorNotSupportAddNot
	//}
	//
	//if len(lb.Data) != 0 && block.BlockType == ledger.BlockTypeReceive {
	//	return ErrorNotSupportRecvAddNote
	//}

	v := verifier.NewAccountVerifier(t.vite.Chain(), t.vite.Consensus())

	blocks, err := v.VerifyforRPC(lb)
	if err != nil {
		newerr, _ := TryMakeConcernedError(err)
		return newerr
	}

	if len(blocks) > 0 && blocks[0] != nil {
		return t.vite.Pool().AddDirectAccountBlock(block.AccountAddress, blocks[0])
	} else {
		return errors.New("generator gen an empty block")
	}
	return nil
}

func (t Tx) SendTxWithPrivateKey(param SendTxWithPrivateKeyParam) (*AccountBlock, error) {

	if param.Amount == nil {
		return nil, errors.New("amount is nil")
	}

	if param.SelfAddr == nil {
		return nil, errors.New("selfAddr is nil")
	}

	if param.ToAddr == nil && param.BlockType != ledger.BlockTypeSendCreate {
		return nil, errors.New("toAddr is nil")
	}

	if param.PrivateKey == nil {
		return nil, errors.New("privateKey is nil")
	}

	var d *big.Int = nil
	if param.Difficulty != nil {
		t, ok := new(big.Int).SetString(*param.Difficulty, 10)
		if !ok {
			return nil, ErrStrToBigInt
		}
		d = t
	}

	amount, ok := new(big.Int).SetString(*param.Amount, 10)
	if !ok {
		return nil, ErrStrToBigInt
	}
	var blockType byte
	if param.BlockType > 0 {
		blockType = param.BlockType
	} else {
		blockType = ledger.BlockTypeSendCall
	}
	msg := &generator.IncomingMessage{
		BlockType:      blockType,
		AccountAddress: *param.SelfAddr,
		ToAddress:      param.ToAddr,
		TokenId:        &param.TokenTypeId,
		Amount:         amount,
		Fee:            nil,
		Data:           param.Data,
		Difficulty:     d,
	}
	_, fitestSnapshotBlockHash, err := generator.GetFittestGeneratorSnapshotHash(t.vite.Chain(), &msg.AccountAddress, nil, false)
	if err != nil {
		return nil, err
	}
	g, e := generator.NewGenerator(t.vite.Chain(), fitestSnapshotBlockHash, param.PreBlockHash, param.SelfAddr)
	if e != nil {
		return nil, e
	}
	result, e := g.GenerateWithMessage(msg, func(addr types.Address, data []byte) (signedData, pubkey []byte, err error) {
		var privkey ed25519.PrivateKey
		privkey, e := ed25519.HexToPrivateKey(*param.PrivateKey)
		if e != nil {
			return nil, nil, e
		}
		signData := ed25519.Sign(privkey, data)
		pubkey = privkey.PubByte()
		return signData, pubkey, nil
	})
	if e != nil {
		newerr, _ := TryMakeConcernedError(e)
		return nil, newerr
	}
	if result.Err != nil {
		newerr, _ := TryMakeConcernedError(result.Err)
		return nil, newerr
	}
	if len(result.BlockGenList) > 0 && result.BlockGenList[0] != nil {
		if err := t.vite.Pool().AddDirectAccountBlock(*param.SelfAddr, result.BlockGenList[0]); err != nil {
			return nil, err
		}
		return ledgerToRpcBlock(result.BlockGenList[0].AccountBlock, t.vite.Chain())

	} else {
		return nil, errors.New("generator gen an empty block")
	}

}

type SendTxWithPrivateKeyParam struct {
	SelfAddr     *types.Address    `json:"selfAddr"`
	ToAddr       *types.Address    `json:"toAddr"`
	TokenTypeId  types.TokenTypeId `json:"tokenTypeId"`
	PrivateKey   *string           `json:"privateKey"` //hex16
	Amount       *string           `json:"amount"`
	Data         []byte            `json:"data"` //base64
	Difficulty   *string           `json:"difficulty,omitempty"`
	PreBlockHash *types.Hash       `json:"preBlockHash,omitempty"`
	BlockType    byte              `json:"blockType"`
}

type CalcPoWDifficultyParam struct {
	SelfAddr     types.Address `json:"selfAddr"`
	PrevHash     types.Hash    `json:"prevHash"`
	SnapshotHash types.Hash    `json:"snapshotHash"`

	BlockType byte           `json:"blockType"`
	ToAddr    *types.Address `json:"toAddr"`
	Data      []byte         `json:"data"`

	UsePledgeQuota bool `json:"usePledgeQuota"`
}

func (t Tx) CalcPoWDifficulty(param CalcPoWDifficultyParam) (difficulty string, err error) {
	var quotaRequired uint64
	if param.BlockType == ledger.BlockTypeSendCreate {
		quotaRequired, _ = util.IntrinsicGasCost(param.Data, false)
	} else if param.BlockType == ledger.BlockTypeReceive {
		quotaRequired, _ = util.IntrinsicGasCost(nil, false)
	} else if param.BlockType == ledger.BlockTypeSendCall {
		if param.ToAddr == nil {
			return "", errors.New("toAddr is nil")
		}
		if types.IsPrecompiledContractAddress(*param.ToAddr) {
			if method, ok, err := vm.GetPrecompiledContract(*param.ToAddr, param.Data); !ok || err != nil {
				return "", errors.New("precompiled contract method not exists")
			} else {
				quotaRequired = method.GetQuota()
			}
		} else {
			quotaRequired, _ = util.IntrinsicGasCost(param.Data, false)
		}
	} else {
		return "", errors.New("block type not supported")
	}

	db, err := vm_context.NewVmContext(t.vite.Chain(), &param.SnapshotHash, &param.PrevHash, &param.SelfAddr)
	if err != nil {
		return "", err
	}
	if param.UsePledgeQuota {
		pledgeAmount := abi.GetPledgeBeneficialAmount(db, param.SelfAddr)
		quotaLeft, _, err := quota.CalcQuotaV2(db, param.SelfAddr, pledgeAmount, helper.Big0)
		if err != nil {
			return "", err
		}
		if quotaLeft >= quotaRequired {
			return "0", nil
		}
	}

	if !quota.CanPoW(db, param.SelfAddr) {
		return "", util.ErrCalcPoWTwice
	}
	// TODO optimize part use quota left
	d := quota.CalcPoWDifficulty(quotaRequired)
	return d.String(), nil
}

const (
	day           = 24 * time.Hour
	powTimesLimit = 10
)

// A single account is limited to send 10 tx with PoW in one day
func checkPoWLimit(c chain.Chain, addr types.Address, prevHash *types.Hash, snapshotHash types.Hash, nonce []byte) (bool, error) {
	if prevHash == nil || !isPoW(nonce) {
		return false, nil
	}
	powtimes := 1
	currentSb, err := c.GetSnapshotBlockByHash(&snapshotHash)
	if err != nil {
		return false, err
	}
	startTime := getTodayStartTime(currentSb.Timestamp, *c.GetGenesisSnapshotBlock().Timestamp)
	prevBlockHash := *prevHash
	for {
		prevBlock, err := c.GetAccountBlockByHash(&prevBlockHash)
		if err != nil {
			return false, err
		}
		if prevBlock == nil || prevBlock.Timestamp.Before(*startTime) {
			return false, nil
		}
		if isPoW(prevBlock.Nonce) {
			powtimes = powtimes + 1
			if powtimes > powTimesLimit {
				return true, nil
			}
		}
		prevBlockHash = prevBlock.PrevHash
	}
	return false, err
}

func getTodayStartTime(currentTime *time.Time, genesisTime time.Time) *time.Time {
	startTime := genesisTime.Add(currentTime.Sub(genesisTime).Round(day))
	return &startTime
}

func isPoW(nonce []byte) bool {
	return len(nonce) > 0
}
